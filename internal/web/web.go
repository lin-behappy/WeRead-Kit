// Package web 提供账号管理与阅读任务 API。
package web

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/skip2/go-qrcode"

	"github.com/27Aaron/weread-kit/internal/store"
	"github.com/27Aaron/weread-kit/internal/weread"
)

//go:embed static
var staticFiles embed.FS

// Server 是 Web UI 的 HTTP 服务。
type Server struct {
	db                 *sql.DB
	client             *weread.Client
	logins             *loginManager
	username, password string
	sessionKey         []byte
	farmMu             sync.Mutex
	farms              map[string]*farmSessionHandle
	farmWG             sync.WaitGroup
	closing            bool
	ctx                context.Context
	cancel             context.CancelFunc
}

func New(db *sql.DB) *Server {
	client := weread.NewClient()
	ctx, cancel := context.WithCancel(context.Background())
	key, err := loadOrCreateSessionKey(db)
	if err != nil {
		panic(err)
	}
	return &Server{ctx: ctx, cancel: cancel, sessionKey: []byte(key), farms: map[string]*farmSessionHandle{}, db: db, client: client, logins: newLoginManager(db, client), username: os.Getenv("WEREAD_USERNAME"), password: os.Getenv("WEREAD_PASSWORD")}
}

func (s *Server) Handler() http.Handler {
	sub, err := fs.Sub(staticFiles, "static")
	if err != nil {
		panic("web: embedded static missing: " + err.Error())
	}
	mux := http.NewServeMux()
	// 静态资源禁用浏览器缓存:改版后强刷不再是必要操作(本地应用,无性能顾虑)。
	mux.HandleFunc("GET /login", s.handleLoginPage)
	mux.HandleFunc("GET /api/version", s.handleVersion)
	mux.HandleFunc("GET /healthz", s.handleHealth)
	mux.HandleFunc("GET /readyz", s.handleReady)
	mux.HandleFunc("POST /login", s.handleLoginPost)
	mux.Handle("GET /", http.FileServerFS(sub))
	mux.HandleFunc("GET /api/accounts", s.handleListAccounts)
	mux.HandleFunc("DELETE /api/accounts/{vid}", s.handleDeleteAccount)
	mux.HandleFunc("PUT /api/accounts/{vid}/remark", s.handleUpdateRemark)
	mux.HandleFunc("POST /api/accounts/{vid}/refresh", s.handleRefreshAccount)
	mux.HandleFunc("POST /api/accounts/{vid}/profile", s.handleSyncProfile)
	mux.HandleFunc("GET /api/accounts/{vid}/details", s.handleAccountDetails)
	mux.HandleFunc("POST /api/accounts/{vid}/details", s.handleAccountDetails)
	mux.HandleFunc("GET /api/accounts/{vid}/weekly", s.handleWeeklyRewards)
	mux.HandleFunc("POST /api/accounts/{vid}/weekly/claim", s.handleWeeklyClaim)
	mux.HandleFunc("POST /api/accounts/{vid}/weekly/pref", s.handleWeeklyPref)
	mux.HandleFunc("GET /api/accounts/{vid}/challenge", s.handleChallengeDetail)
	mux.HandleFunc("GET /api/accounts/{vid}/reading", s.handleReadingConfig)
	mux.HandleFunc("POST /api/accounts/{vid}/reading", s.handleReadingConfig)
	mux.HandleFunc("POST /api/accounts/{vid}/reading/run", s.handleReadingRun)
	mux.HandleFunc("POST /api/accounts/{vid}/reading/stop", s.handleReadingStop)
	mux.HandleFunc("GET /api/logs", s.handleLogs)
	mux.HandleFunc("DELETE /api/logs", s.handleLogs)
	mux.HandleFunc("GET /api/settings/push", s.handlePushChannels)
	mux.HandleFunc("POST /api/settings/push", s.handlePushChannels)
	mux.HandleFunc("POST /api/settings/push/{type}/test", s.handlePushTest)
	mux.HandleFunc("GET /api/accounts/{vid}/token", s.handleAccountToken)
	mux.HandleFunc("POST /api/login", s.handleStartLogin)
	mux.HandleFunc("GET /api/login/{id}", s.handleLoginStatus)
	mux.HandleFunc("GET /api/login/{id}/qr.png", s.handleLoginQR)
	return noStore(http.NewCrossOriginProtection().Handler(s.auth(http.MaxBytesHandler(mux, 1<<20))))
}

func (s *Server) handleLoginPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "text/html; charset=utf-8")
	errMsg := ""
	if r.URL.Query().Get("error") == "1" {
		errMsg = `<div class="error" role="alert">账号或密码错误，请重试</div>`
	}
	html := `<!doctype html><html lang="zh-CN">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>登录 · 微信读书</title>
<link rel="icon" type="image/png" href="/favicon.png">
<script>
  // 与主界面一致：优先用户手动选择的主题，其次跟随系统。
  document.documentElement.dataset.theme =
    localStorage.getItem("weread-theme") ||
    (matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light");
</script>
<link rel="stylesheet" href="https://gw.alipayobjects.com/os/k/jinkai/style.css" crossorigin="anonymous">
<style>
:root{color-scheme:light;--bg:#f6f4ef;--surface:#fffdf9;--text:#2a2620;--text-dim:#746e65;--border:#e6e1d5;--border-strong:#948a77;--accent:#b45309;--accent-strong:#8a3e06;--accent-soft:#f7ead7;--err:#b91c1c;--shadow:0 16px 45px rgba(76,59,36,.12);--font-sans:"TsangerJinKai02",system-ui,sans-serif}
html[data-theme="dark"]{color-scheme:dark;--bg:#171512;--surface:#201d19;--text:#eae5da;--text-dim:#a39c8f;--border:#383329;--border-strong:#736a5a;--accent:#e0a458;--accent-strong:#f0b869;--accent-soft:#332a1c;--err:#f87171;--shadow:0 16px 45px rgba(0,0,0,.35)}
*{box-sizing:border-box}
body{margin:0;min-height:100vh;display:grid;place-items:center;padding:24px;background:var(--bg);color:var(--text);font:15px/1.5 var(--font-sans)}
.card{width:min(100%,390px);padding:40px;border:1px solid var(--border);border-radius:16px;background:var(--surface);box-shadow:var(--shadow)}
.brand{text-align:center;margin-bottom:28px}
.brand img{width:54px;height:54px;border-radius:14px}
.brand h1{font-size:22px;margin:14px 0 6px}
.brand p{color:var(--text-dim);margin:0;font-size:13px}
label{display:block;font-size:13px;font-weight:600;margin:16px 0 7px}
input{display:block;width:100%;padding:12px 13px;border:1px solid var(--border-strong);border-radius:9px;background:var(--surface);color:inherit;font:inherit}
input:focus-visible{outline:2px solid var(--accent);outline-offset:1px;border-color:var(--accent)}
.pw-wrap{position:relative}
.pw-wrap input{padding-right:66px}
.pw-toggle{position:absolute;top:50%;right:6px;transform:translateY(-50%);padding:7px 9px;border:0;border-radius:6px;background:transparent;color:var(--text-dim);font:inherit;font-size:12px;cursor:pointer}
.pw-toggle:hover{color:var(--accent);background:var(--accent-soft)}
.pw-toggle:focus-visible{outline:2px solid var(--accent);outline-offset:1px}
button[type=submit]{width:100%;margin-top:24px;padding:12px;border:0;border-radius:9px;background:var(--accent);color:#fff;font:600 15px inherit;cursor:pointer}
button[type=submit]:hover{background:var(--accent-strong)}
button[type=submit]:focus-visible{outline:2px solid var(--accent);outline-offset:2px}
.error{padding:10px 12px;border-radius:8px;background:color-mix(in srgb,var(--err) 12%,transparent);color:var(--err);font-size:13px;margin-bottom:12px}
</style>
</head>
<body>
<main class="card">
  <div class="brand">
    <img src="/favicon.png" alt="">
    <h1>欢迎回来</h1>
    <p>登录后管理你的微信读书账号</p>
  </div>` + errMsg + `
  <form method="post">
    <label for="username">账号</label>
    <input id="username" name="username" autocomplete="username" placeholder="请输入登录账号" required>
    <label for="password">密码</label>
    <div class="pw-wrap">
      <input id="password" name="password" type="password" autocomplete="current-password" placeholder="请输入登录密码" required>
      <button type="button" id="pw-toggle" class="pw-toggle" aria-label="显示密码" aria-pressed="false">显示</button>
    </div>
    <button type="submit">登录</button>
  </form>
</main>
<script>
  const pw = document.getElementById("password");
  const pwToggle = document.getElementById("pw-toggle");
  pwToggle.addEventListener("click", () => {
    const show = pw.type === "password";
    pw.type = show ? "text" : "password";
    pwToggle.textContent = show ? "隐藏" : "显示";
    pwToggle.setAttribute("aria-pressed", String(show));
    pwToggle.setAttribute("aria-label", show ? "隐藏密码" : "显示密码");
  });
</script>
</body>
</html>`
	w.Write([]byte(html))
}

// noStore 为所有静态资源响应附加禁用缓存的头。
func noStore(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("cache-control", "no-store")
		w.Header().Set("x-content-type-options", "nosniff")
		w.Header().Set("x-frame-options", "DENY")
		w.Header().Set("referrer-policy", "same-origin")
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("content-type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	if err := s.db.PingContext(r.Context()); err != nil {
		writeErr(w, http.StatusServiceUnavailable, errors.New("数据库不可用"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

type accountView struct {
	Name             string `json:"name"`
	Avatar           string `json:"avatar"`
	UserVid          string `json:"user_vid"`
	ProfileUpdatedAt int64  `json:"profile_updated_at"`
	Remark           string `json:"remark"`
	Vid              string `json:"vid"`
	DeviceID         string `json:"device_id"`
	HasAccessToken   bool   `json:"has_access_token"`
	RotatedAt        int64  `json:"rotated_at"`
	CreatedAt        int64  `json:"created_at"`
}

func toView(c *store.Credential) accountView {
	return accountView{
		Name:             c.Name,
		Avatar:           c.Avatar,
		UserVid:          c.UserVid,
		ProfileUpdatedAt: c.ProfileUpdatedAt,
		Remark:           c.Remark,
		Vid:              c.Vid,
		DeviceID:         c.DeviceID,
		HasAccessToken:   c.AccessToken != "",
		RotatedAt:        c.RotatedAt.Unix(),
		CreatedAt:        c.CreatedAt.Unix(),
	}
}

func (s *Server) handleListAccounts(w http.ResponseWriter, r *http.Request) {
	list, err := store.List(s.db)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	views := make([]accountView, 0, len(list))
	for _, c := range list {
		views = append(views, toView(c))
	}
	writeJSON(w, http.StatusOK, views)
}

func (s *Server) handleDeleteAccount(w http.ResponseWriter, r *http.Request) {
	s.farmMu.Lock()
	defer s.farmMu.Unlock()
	if _, running := s.farms[r.PathValue("vid")]; running {
		writeErr(w, http.StatusConflict, errors.New("请先停止阅读，再删除账号"))
		return
	}
	err := store.Delete(s.db, r.PathValue("vid"))
	if errors.Is(err, store.ErrNotFound) {
		writeErr(w, http.StatusNotFound, err)
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleUpdateRemark(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Remark string `json:"remark"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, errors.New("请求体不是 JSON"))
		return
	}
	if len([]rune(body.Remark)) > 100 {
		writeErr(w, http.StatusBadRequest, errors.New("备注过长(最多 100 字)"))
		return
	}
	err := store.UpdateRemark(s.db, r.PathValue("vid"), body.Remark)
	if errors.Is(err, store.ErrNotFound) {
		writeErr(w, http.StatusNotFound, err)
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleRefreshAccount(w http.ResponseWriter, r *http.Request) {
	vid := r.PathValue("vid")
	c, err := store.Load(s.db, vid)
	if errors.Is(err, store.ErrNotFound) {
		writeErr(w, http.StatusNotFound, err)
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	refreshed, err := s.client.Refresh(r.Context(), toWeread(c))
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	if err := store.Save(s.db, toStore(vid, refreshed)); err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	updated, err := store.Load(s.db, vid)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, toView(updated))
}

func (s *Server) handleAccountToken(w http.ResponseWriter, r *http.Request) {
	vid := r.PathValue("vid")
	c, err := store.Load(s.db, vid)
	if errors.Is(err, store.ErrNotFound) {
		writeErr(w, http.StatusNotFound, err)
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	maxAge := 24 * time.Hour
	if v := r.URL.Query().Get("max_age"); v != "" {
		if parsed, err := time.ParseDuration(v); err == nil && parsed > 0 {
			maxAge = parsed
		}
	}
	if r.URL.Query().Get("force") == "1" || time.Since(c.RotatedAt) > maxAge {
		refreshed, err := s.client.Refresh(r.Context(), toWeread(c))
		if err != nil {
			writeErr(w, http.StatusBadGateway, err)
			return
		}
		if err := store.Save(s.db, toStore(vid, refreshed)); err != nil {
			writeErr(w, http.StatusInternalServerError, err)
			return
		}
		if c, err = store.Load(s.db, vid); err != nil {
			writeErr(w, http.StatusInternalServerError, err)
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"access_token": c.AccessToken,
		"vid":          c.Vid,
		"rotated_at":   c.RotatedAt.Unix(),
	})
}

// ---- 扫码登录会话 ----

func (s *Server) handleStartLogin(w http.ResponseWriter, r *http.Request) {
	id, err := s.logins.start()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	// 等二维码就绪再返回:前端拿到 qr_url 时 PNG 必定存在,
	// 票据阶段的即时失败也在这里直接报给用户。
	if err := s.logins.waitQR(r.Context(), id, 8*time.Second); err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"id":     id,
		"qr_url": "/api/login/" + id + "/qr.png",
	})
}

func (s *Server) handleLoginStatus(w http.ResponseWriter, r *http.Request) {
	sess, ok := s.logins.get(r.PathValue("id"))
	if !ok {
		writeErr(w, http.StatusNotFound, errors.New("登录会话不存在或已过期"))
		return
	}
	resp := map[string]any{"status": sess.status}
	if sess.errMsg != "" {
		resp["error"] = sess.errMsg
	}
	if sess.status == "success" {
		resp["account"] = map[string]string{"vid": sess.creds.Vid}
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleLoginQR(w http.ResponseWriter, r *http.Request) {
	sess, ok := s.logins.get(r.PathValue("id"))
	if !ok || sess.confirmURL == "" {
		http.NotFound(w, r)
		return
	}
	png, err := qrcode.Encode(sess.confirmURL, qrcode.Medium, 360)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("content-type", "image/png")
	w.Header().Set("cache-control", "no-store")
	_, _ = w.Write(png)
}

func toWeread(c *store.Credential) *weread.Credentials {
	return &weread.Credentials{
		Vid:          c.Vid,
		AccessToken:  c.AccessToken,
		RefreshToken: c.RefreshToken,
		DeviceID:     c.DeviceID,
	}
}

func toStore(vid string, c *weread.Credentials) *store.Credential {
	return &store.Credential{
		Vid:          c.Vid,
		RefreshToken: c.RefreshToken,
		DeviceID:     c.DeviceID,
		AccessToken:  c.AccessToken,
	}
}
