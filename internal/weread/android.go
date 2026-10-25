package weread

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// 安卓渠道的领取端标识。安卓 token 必须配这个 pf,
// 不能跟随 WEREAD_PF 的 iOS 覆盖,否则服务端会按错误渠道处理。
const androidWeeklyPF = "wechat_wx-2001-android-100-weread"

// AndroidCredentials 是安卓客户端(非墨水屏)的独立会话凭据。
//
// 微信读书的周奖励按领取端记录奖品类型:墨水屏端 7 天档只提供体验卡,
// 安卓端才提供 6 书币。两套 token 不能混用,因此这里单独保存安卓凭据。
// 字段名与导入 API 的 JSON 保持一致,便于直接落库和回显。
type AndroidCredentials struct {
	// Vid 不写入 JSON,由调用方按账号路径填充。
	Vid string `json:"-"`

	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	DeviceID     string `json:"device_id"`
	DeviceName   string `json:"device_name,omitempty"`
	InstallID    string `json:"install_id,omitempty"`

	// 客户端指纹:服务端按 UA + 版本头识别渠道,缺一项就可能返回 -2012。
	UserAgent  string `json:"user_agent"`
	AppVersion string `json:"app_version"`
	BaseAPI    string `json:"baseapi"`
	OSVersion  string `json:"osver"`
	ChannelID  string `json:"channel_id"`
}

// Validate 检查调用 weekly/exchange 所必需的字段。
// refresh_token 不强制:当前安卓 refresh 接口拒绝 HTTP 续期,
// token 过期后需要重新从客户端导入,但保留它便于以后支持自动续期。
func (a *AndroidCredentials) Validate() error {
	if a == nil {
		return fmt.Errorf("安卓凭据为空")
	}
	var missing []string
	for _, f := range []struct {
		name  string
		value string
	}{
		{"access_token", a.AccessToken},
		{"device_id", a.DeviceID},
		{"user_agent", a.UserAgent},
		{"app_version", a.AppVersion},
		{"baseapi", a.BaseAPI},
		{"osver", a.OSVersion},
		{"channel_id", a.ChannelID},
	} {
		if strings.TrimSpace(f.value) == "" {
			missing = append(missing, f.name)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("安卓凭据缺少字段: %s", strings.Join(missing, ", "))
	}
	return nil
}

// headers 复刻安卓客户端的请求头。安卓 token 必须搭配同一渠道指纹,
// 换成墨水屏的 WRBrand/Onyx 或错误的 channelId 会直接 -2012。
func (a *AndroidCredentials) headers() map[string]string {
	return map[string]string{
		"baseapi":     a.BaseAPI,
		"appver":      a.AppVersion,
		"basever":     a.AppVersion,
		"osver":       a.OSVersion,
		"channelId":   a.ChannelID,
		"User-Agent":  a.UserAgent,
		"vid":         a.Vid,
		"accessToken": a.AccessToken,
	}
}

// WeeklyExchangeAndroid 用安卓会话查询/领取周奖励。
// 安卓端响应会同时列出体验卡和书币两个 choiceType,这是与墨水屏端的关键差异。
// 会话过期返回 ErrSessionExpired;安卓 token 目前无法 HTTP 续期,调用方应提示重新导入。
func (c *Client) WeeklyExchangeAndroid(ctx context.Context, creds *AndroidCredentials, awardLevelID, awardChooseType, isExchangeAward int) (json.RawMessage, error) {
	if err := creds.Validate(); err != nil {
		return nil, err
	}
	return c.weeklyExchange(ctx, creds.headers(), awardLevelID, awardChooseType, isExchangeAward, map[string]any{
		"isVisitReadGoal": 1,
		"unread":          0,
		"pf":              androidWeeklyPF,
	})
}
