// Package weread 实现微信读书移动端(墨水屏客户端)私有协议:
// 终端扫码登录、refreshToken 换新。
package weread

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"sync"
	"time"
)

const (
	BaseURL = "https://i.weread.qq.com"

	// 微信读书注册的微信 AppID 与 OAuth scope,扫码授权走的就是这套参数。
	wxAppID  = "wxab9b71ad2b90ff34"
	wxScope  = "snsapi_userinfo,snsapi_timeline,snsapi_friend"
	nonceStr = "weread"

	deviceName = "BOOX"
	deviceType = 3

	// 微信读书 2.1.2 安卓客户端跑在 Onyx BOOX 墨水屏上的完整设备指纹。
	// 服务端按 UA + 版本头组合识别客户端,改动任何一项都可能被风控,
	// 因此与参考实现(weread-omni)保持逐字一致。
	userAgent  = "WeRead/2.1.2 WRBrand/Onyx wr_eink Dalvik/2.1.0 (Linux; U; Android 11; BOOX Build/onyx)"
	appVersion = "2.1.2.10245900"

	maxResponseBytes = 16 << 20
)

// Credentials 是登录与续期产出的凭据四元组,与 SQLite 里持久化的字段一一对应。
type Credentials struct {
	Vid          string
	AccessToken  string
	RefreshToken string
	DeviceID     string
}

// Client 是协议客户端,可注入自定义 *http.Client 便于测试。
type Client struct {
	HTTP *http.Client

	refreshMu sync.Mutex
	refreshes map[string]*refreshCall
}

// refreshCall 表示同一账号正在进行的一次凭据刷新。
// 多个请求同时发现会话过期时,后到的请求复用前一个请求的结果,避免重复消费 refreshToken。
type refreshCall struct {
	source *Credentials
	done   chan struct{}
	next   *Credentials
	err    error
}

func NewClient() *Client {
	return &Client{HTTP: &http.Client{
		Timeout: 70 * time.Second,
		// 协议里出现重定向都按异常处理,不跟随。
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}, refreshes: make(map[string]*refreshCall)}
}

func versionHeaders() map[string]string {
	return map[string]string{
		"baseapi":    "30",
		"appver":     appVersion,
		"basever":    appVersion,
		"osver":      "11",
		"channelId":  "900",
		"User-Agent": userAgent,
	}
}

// authHeaders 是业务接口的鉴权头:vid + accessToken 直接平铺在请求头上,
// 等价于网页版的 Cookie。
func authHeaders(creds *Credentials) map[string]string {
	return map[string]string{
		"vid":         creds.Vid,
		"accessToken": creds.AccessToken,
	}
}

// checkBusinessCode 检查业务响应包络:errCode -2012 映射为会话过期,
// 其它非 0 值转为错误。非 JSON 或无 errCode 的响应视为通过。
func checkBusinessCode(body []byte) error {
	var envelope struct {
		ErrCode json.Number `json:"errCode"`
		ErrMsg  string      `json:"errMsg"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil
	}
	if envelope.ErrCode.String() == "-2012" {
		return ErrSessionExpired
	}
	if n, err := envelope.ErrCode.Int64(); err == nil && n != 0 {
		return fmt.Errorf("errCode=%s errMsg=%q", envelope.ErrCode.String(), envelope.ErrMsg)
	}
	return nil
}

// ErrSessionExpired 表示服务端判定会话失效(errCode -2012 / HTTP 401),
// 调用方应刷新凭据后重试一次。
var ErrSessionExpired = errors.New("会话已过期")

// sign 复刻服务端校验的防篡改签名:SHA-256(timestamp + deviceId + random) 的十六进制小写。
func sign(timestampMs int64, deviceID string, random int) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%d%s%d", timestampMs, deviceID, random)))
	return hex.EncodeToString(sum[:])
}

func (c *Client) do(ctx context.Context, method, rawURL string, headers map[string]string, body []byte) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, method, rawURL, bytes.NewReader(body))
	if err != nil {
		return nil, 0, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	if body != nil {
		req.Header.Set("content-type", "application/json; charset=UTF-8")
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return data, resp.StatusCode, err
	}
	return data, resp.StatusCode, nil
}

func (c *Client) getJSON(ctx context.Context, rawURL string, headers map[string]string) ([]byte, int, error) {
	return c.do(ctx, http.MethodGet, rawURL, headers, nil)
}

func (c *Client) postJSON(ctx context.Context, rawURL string, headers map[string]string, payload any) ([]byte, int, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, err
	}
	return c.do(ctx, http.MethodPost, rawURL, headers, body)
}

// rawString 兼容 vid 这类服务端可能回字符串或整数的字段。
func rawString(raw json.RawMessage) string {
	s := string(raw)
	if s == "null" {
		return ""
	}
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		var unquoted string
		if err := json.Unmarshal(raw, &unquoted); err == nil {
			return unquoted
		}
	}
	return s
}

// randIntn 返回 [0, n) 的密码学安全随机数。
func randIntn(n int) int {
	v, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		// 熵源不可用属致命错误,直接终止。
		panic("weread: crypto/rand unavailable: " + err.Error())
	}
	return int(v.Int64())
}

func randomDigits(n int) string {
	out := make([]byte, n)
	for i := range out {
		out[i] = byte('0' + randIntn(10))
	}
	return string(out)
}

// NewDeviceID 生成 e-ink 设备 ID:eink334691225 前缀 + 63 位随机数的 19 位十进制。
func NewDeviceID() string {
	v, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 63))
	if err != nil {
		panic("weread: crypto/rand unavailable: " + err.Error())
	}
	return fmt.Sprintf("eink334691225%019d", v.Int64())
}

func newInstallID() string {
	return "eink31" + randomDigits(26)
}

func unixMilli() int64 { return time.Now().UnixMilli() }
