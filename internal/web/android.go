// 安卓客户端凭据的导入/查看/移除 API。
// 安卓 token 与墨水屏 refreshToken 是两套独立会话,单独存放在 weread_account.android。
package web

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/27Aaron/weread-kit/internal/store"
	"github.com/27Aaron/weread-kit/internal/weread"
)

// handleAndroidCredential 处理 GET/POST/DELETE /api/accounts/{vid}/android。
// GET 只返回脱敏元数据,绝不回显 token。
func (s *Server) handleAndroidCredential(w http.ResponseWriter, r *http.Request) {
	vid := r.PathValue("vid")
	switch r.Method {
	case http.MethodGet:
		c, err := store.Load(s.db, vid)
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, http.StatusNotFound, err)
			return
		}
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err)
			return
		}
		resp := map[string]any{"configured": strings.TrimSpace(c.Android) != ""}
		if c.Android != "" {
			var a weread.AndroidCredentials
			if err := json.Unmarshal([]byte(c.Android), &a); err == nil {
				a.Vid = c.Vid
				resp["device_id"] = a.DeviceID
				resp["device_name"] = a.DeviceName
				resp["app_version"] = a.AppVersion
				resp["channel_id"] = a.ChannelID
				resp["has_refresh_token"] = a.RefreshToken != ""
				resp["valid"] = a.Validate() == nil
			} else {
				resp["valid"] = false
				resp["error"] = "安卓凭据 JSON 无法解析"
			}
		}
		writeJSON(w, http.StatusOK, resp)

	case http.MethodPost:
		var body weread.AndroidCredentials
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeErr(w, http.StatusBadRequest, errors.New("请求体不是有效的安卓凭据 JSON"))
			return
		}
		body.Vid = vid
		if err := body.Validate(); err != nil {
			writeErr(w, http.StatusBadRequest, err)
			return
		}
		raw, err := json.Marshal(body)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err)
			return
		}
		if err := store.SaveAndroid(s.db, vid, string(raw)); errors.Is(err, store.ErrNotFound) {
			writeErr(w, http.StatusNotFound, err)
			return
		} else if err != nil {
			writeErr(w, http.StatusInternalServerError, err)
			return
		}
		s.logf("info", "account", vid, "已导入安卓客户端凭据")
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "configured": true})

	case http.MethodDelete:
		if err := store.SaveAndroid(s.db, vid, ""); errors.Is(err, store.ErrNotFound) {
			writeErr(w, http.StatusNotFound, err)
			return
		} else if err != nil {
			writeErr(w, http.StatusInternalServerError, err)
			return
		}
		s.logf("info", "account", vid, "已移除安卓客户端凭据")
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "configured": false})

	default:
		w.Header().Set("allow", "GET, POST, DELETE")
		writeErr(w, http.StatusMethodNotAllowed, errors.New("不支持的请求方法"))
	}
}
