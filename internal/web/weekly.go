// 「我的阅读」页 API:周阅读奖励的查询、领取与自动领取预约。
package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/27Aaron/weread-kit/internal/store"
	"github.com/27Aaron/weread-kit/internal/weread"
)

// weeklyClaimPrefKey 是账号奖励预约在 weread_setting 表中的键前缀,
// 完整键为 weeklyClaimPrefKey+vid,值为 {"档位ID": 奖品类型} 的 JSON。
const weeklyClaimPrefKey = "weekly_claim_pref:"

// handleWeeklyRewards 返回指定账号的周阅读奖励、阅读统计与领取预约。
// GET /api/accounts/{vid}/weekly
func (s *Server) handleWeeklyRewards(w http.ResponseWriter, r *http.Request) {
	vid := r.PathValue("vid")
	q, err := s.weeklyQuery(r.Context(), vid)
	if err != nil {
		writeWeeklyErr(w, err)
		return
	}
	resp := map[string]any{
		"vid":    vid,
		"weekly": json.RawMessage(q.Data),
		// weekly/exchange 的 infiniteCard 字段不反映真实卡状态,
		// 无限卡剩余天数以详情缓存的会员卡信息为准。
		"card":               memberCardView(q.Cred.Card),
		"prefs":              s.weeklyClaimPrefs(vid),
		"channel":            q.Channel,
		"android_configured": strings.TrimSpace(q.Cred.Android) != "",
	}
	if q.AndroidErr != nil {
		resp["android_error"] = q.AndroidErr.Error()
	}
	writeJSON(w, http.StatusOK, resp)
}

// handleWeeklyClaim 领取一档奖励。body: {"award_level_id":5,"choice_type":1}
// choice_type: 1=体验卡天数 2=书币。领取成功后同时记下预约,以后每周自动领取同款。
// POST /api/accounts/{vid}/weekly/claim
func (s *Server) handleWeeklyClaim(w http.ResponseWriter, r *http.Request) {
	vid := r.PathValue("vid")
	var body struct {
		AwardLevelID int `json:"award_level_id"`
		ChoiceType   int `json:"choice_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, errors.New("请求体不是 JSON"))
		return
	}
	if body.AwardLevelID <= 0 || (body.ChoiceType != weread.AwardChoiceCard && body.ChoiceType != weread.AwardChoiceBookCoin) {
		writeErr(w, http.StatusBadRequest, errors.New("参数无效:需要 award_level_id 与 choice_type(1=体验卡,2=书币)"))
		return
	}
	// 只有 7 天档（level 13）的书币必须走安卓渠道:墨水屏端该档没有
	// choiceType=2,若回退到墨水屏会以 -2417 失败,绝不能悄悄改领体验卡。
	// 其他档位的书币继续走墨水屏,保持原有领取端。
	var data json.RawMessage
	var err error
	if body.AwardLevelID == 13 && body.ChoiceType == weread.AwardChoiceBookCoin {
		data, _, err = s.weeklyCallAndroid(r.Context(), vid, func(creds *weread.AndroidCredentials) (json.RawMessage, error) {
			return s.client.WeeklyExchangeAndroid(r.Context(), creds, body.AwardLevelID, body.ChoiceType, 1)
		})
	} else {
		data, _, err = s.weeklyCall(r.Context(), vid, func(creds *weread.Credentials) (json.RawMessage, error) {
			return s.client.WeeklyExchange(r.Context(), creds, body.AwardLevelID, body.ChoiceType, 1)
		})
	}
	if err != nil {
		writeWeeklyErr(w, err)
		return
	}
	// 领取即选中:之后每周到达可领取时刻由调度器自动领同款。
	if err := s.saveWeeklyClaimPref(vid, body.AwardLevelID, body.ChoiceType); err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	s.logf("info", "reading", vid, "保存周奖励领取设置:档位 %d → %s(达成后自动领取)", body.AwardLevelID, weeklyChoiceName(body.ChoiceType))
	writeJSON(w, http.StatusOK, map[string]any{
		"vid":    vid,
		"result": json.RawMessage(data),
	})
}

// handleWeeklyPref 仅保存/修改预约,不立即领取(用于未达成的档位)。
// POST /api/accounts/{vid}/weekly/pref
func (s *Server) handleWeeklyPref(w http.ResponseWriter, r *http.Request) {
	vid := r.PathValue("vid")
	var body struct {
		AwardLevelID int `json:"award_level_id"`
		ChoiceType   int `json:"choice_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, errors.New("请求体不是 JSON"))
		return
	}
	if body.AwardLevelID <= 0 || (body.ChoiceType != weread.AwardChoiceCard && body.ChoiceType != weread.AwardChoiceBookCoin) {
		writeErr(w, http.StatusBadRequest, errors.New("参数无效:需要 award_level_id 与 choice_type(1=体验卡,2=书币)"))
		return
	}
	if _, err := store.Load(s.db, vid); errors.Is(err, store.ErrNotFound) {
		writeErr(w, http.StatusNotFound, err)
		return
	} else if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	if err := s.saveWeeklyClaimPref(vid, body.AwardLevelID, body.ChoiceType); err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	s.logf("info", "reading", vid, "保存周奖励领取设置:档位 %d → %s(达成后自动领取)", body.AwardLevelID, weeklyChoiceName(body.ChoiceType))
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "prefs": s.weeklyClaimPrefs(vid)})
}

// weeklyClaimPrefs 读取账号的奖励预约;无预约时返回空 map。
func (s *Server) weeklyClaimPrefs(vid string) map[string]int {
	v, err := store.GetSetting(s.db, weeklyClaimPrefKey+vid)
	if err != nil || v == "" {
		return map[string]int{}
	}
	prefs := map[string]int{}
	if json.Unmarshal([]byte(v), &prefs) != nil {
		return map[string]int{}
	}
	return prefs
}

func (s *Server) saveWeeklyClaimPref(vid string, level, choice int) error {
	prefs := s.weeklyClaimPrefs(vid)
	prefs[strconv.Itoa(level)] = choice
	b, err := json.Marshal(prefs)
	if err != nil {
		return err
	}
	return store.SetSetting(s.db, weeklyClaimPrefKey+vid, string(b))
}

// StartWeeklyClaimScheduler 启动自动领取循环:启动 30 秒后首轮,此后每 5 分钟
// 扫描一遍有预约的账号,奖励一旦达成(status=1)即按预约自动领取并记日志。
func (s *Server) StartWeeklyClaimScheduler(ctx context.Context) {
	go func() {
		select {
		case <-ctx.Done():
			return
		case <-time.After(30 * time.Second):
		}
		s.autoClaimOnce(ctx)
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.autoClaimOnce(ctx)
			}
		}
	}()
}

// autoClaimOnce 执行一轮自动领取。
func (s *Server) autoClaimOnce(ctx context.Context) {
	accounts, err := store.List(s.db)
	if err != nil {
		return
	}
	for _, acc := range accounts {
		if ctx.Err() != nil {
			return
		}
		prefs := s.weeklyClaimPrefs(acc.Vid)
		if len(prefs) == 0 {
			continue
		}
		q, err := s.weeklyQuery(ctx, acc.Vid)
		if err != nil {
			s.warnWeeklyOnce(acc.Vid, fmt.Sprintf("周奖励查询失败: %v", err))
			continue // 查询失败(凭据失效等)跳过,下轮再试
		}
		var rw weread.WeeklyRewards
		if json.Unmarshal(q.Data, &rw) != nil {
			continue
		}
		awards := append(append(rw.ReadtimeAwards, rw.ReaddayAwards...), rw.ReadgoalAwards...)
		for _, a := range awards {
			if a.AwardStatus != weread.AwardStatusClaimable {
				continue
			}
			choice, ok := prefs[strconv.Itoa(a.AwardLevelID)]
			if !ok {
				continue
			}
			var claimErr error
			if a.AwardLevelID == 13 && choice == weread.AwardChoiceBookCoin {
				// 7 天档书币必须走安卓通道。查询已回退到墨水屏时,响应里根本没有
				// choiceType=2,不能拿墨水屏 token 去试,更不能改领体验卡。
				if q.Channel != "android" {
					detail := ""
					if q.AndroidErr != nil {
						detail = ": " + q.AndroidErr.Error()
					}
					s.warnWeeklyOnce(acc.Vid, fmt.Sprintf("自动领取跳过(%s → 书币): 安卓通道不可用%s", a.AwardLevelDesc, detail))
					continue
				}
				_, _, claimErr = s.weeklyCallAndroid(ctx, acc.Vid, func(creds *weread.AndroidCredentials) (json.RawMessage, error) {
					return s.client.WeeklyExchangeAndroid(ctx, creds, a.AwardLevelID, choice, 1)
				})
			} else {
				_, _, claimErr = s.weeklyCall(ctx, acc.Vid, func(creds *weread.Credentials) (json.RawMessage, error) {
					return s.client.WeeklyExchange(ctx, creds, a.AwardLevelID, choice, 1)
				})
			}
			if claimErr != nil {
				s.warnWeeklyOnce(acc.Vid, fmt.Sprintf("自动领取失败(%s → %s): %v", a.AwardLevelDesc, weeklyChoiceName(choice), claimErr))
				continue
			}
			s.logf("info", "reading", acc.Vid, "自动领取周阅读奖励:%s → %s", a.AwardLevelDesc, weeklyChoiceName(choice))
		}
	}
}

func weeklyChoiceName(choice int) string {
	if choice == weread.AwardChoiceCard {
		return "体验卡"
	}
	return "书币"
}

// weeklyCall 加载账号凭据执行 op;会话过期时自动续期、落库并重试一次。
// 返回原始响应与加载到的账号(含详情缓存,供计算无限卡剩余状态)。
func (s *Server) weeklyCall(ctx context.Context, vid string, op func(*weread.Credentials) (json.RawMessage, error)) (json.RawMessage, *store.Credential, error) {
	c, err := store.Load(s.db, vid)
	if errors.Is(err, store.ErrNotFound) {
		return nil, nil, err
	}
	if err != nil {
		return nil, nil, err
	}
	creds := toWeread(c)
	data, err := op(creds)
	if errors.Is(err, weread.ErrSessionExpired) {
		next, rerr := s.client.Refresh(ctx, creds)
		if rerr != nil {
			return nil, nil, fmt.Errorf("会话已过期且续期失败: %w", rerr)
		}
		if err := store.Save(s.db, toStore(vid, next)); err != nil {
			return nil, nil, err
		}
		if c, err = store.Load(s.db, vid); err != nil {
			return nil, nil, err
		}
		data, err = op(next)
	}
	if err != nil {
		return nil, nil, err
	}
	return data, c, nil
}

// errAndroidUnavailable 表示该账号还没导入安卓客户端凭据。
var errAndroidUnavailable = errors.New("未配置安卓客户端凭据")

// androidCredential 解析并校验账号的安卓凭据。
func androidCredential(c *store.Credential) (*weread.AndroidCredentials, error) {
	if c == nil || strings.TrimSpace(c.Android) == "" {
		return nil, errAndroidUnavailable
	}
	var a weread.AndroidCredentials
	if err := json.Unmarshal([]byte(c.Android), &a); err != nil {
		return nil, fmt.Errorf("安卓凭据格式错误: %w", err)
	}
	a.Vid = c.Vid
	if err := a.Validate(); err != nil {
		return nil, err
	}
	return &a, nil
}

// weeklyCallAndroid 加载安卓凭据执行 op。安卓 refresh 目前不可用,
// 会话过期时直接返回可操作的错误,由用户重新导入 token。
func (s *Server) weeklyCallAndroid(ctx context.Context, vid string, op func(*weread.AndroidCredentials) (json.RawMessage, error)) (json.RawMessage, *store.Credential, error) {
	c, err := store.Load(s.db, vid)
	if errors.Is(err, store.ErrNotFound) {
		return nil, nil, err
	}
	if err != nil {
		return nil, nil, err
	}
	a, err := androidCredential(c)
	if err != nil {
		if errors.Is(err, errAndroidUnavailable) {
			return nil, nil, fmt.Errorf("%w；7 天档书币只能通过安卓客户端领取", err)
		}
		return nil, nil, err
	}
	data, err := op(a)
	if errors.Is(err, weread.ErrSessionExpired) {
		return nil, nil, fmt.Errorf("安卓会话已过期，请在微信读书安卓客户端重新登录并导入新的 token: %w", err)
	}
	if err != nil {
		return nil, nil, err
	}
	return data, c, nil
}

// weeklyQueryResult 是周奖励查询结果及实际使用的通道。
type weeklyQueryResult struct {
	Data       json.RawMessage
	Cred       *store.Credential
	Channel    string // android / eink
	AndroidErr error  // 安卓通道尝试失败的原因;未配置或成功时为 nil
}

// weeklyQuery 优先用安卓通道查询(响应里才有 7 天档书币选项),失败时回退墨水屏。
// 回退只用于展示和其他档位领取;7 天档书币领取仍必须显式走安卓通道。
func (s *Server) weeklyQuery(ctx context.Context, vid string) (*weeklyQueryResult, error) {
	c, err := store.Load(s.db, vid)
	if errors.Is(err, store.ErrNotFound) {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	var androidErr error
	if a, aerr := androidCredential(c); aerr == nil {
		data, err := s.client.WeeklyExchangeAndroid(ctx, a, 0, 0, 0)
		if err == nil {
			return &weeklyQueryResult{Data: data, Cred: c, Channel: "android"}, nil
		}
		androidErr = err
		if errors.Is(err, weread.ErrSessionExpired) {
			androidErr = fmt.Errorf("安卓会话已过期，请在微信读书安卓客户端重新登录并导入新的 token: %w", err)
		}
	} else if !errors.Is(aerr, errAndroidUnavailable) {
		androidErr = aerr
	}
	data, c, err := s.weeklyCall(ctx, vid, func(creds *weread.Credentials) (json.RawMessage, error) {
		return s.client.WeeklyExchange(ctx, creds, 0, 0, 0)
	})
	if err != nil {
		if androidErr != nil {
			return nil, fmt.Errorf("安卓通道失败(%v)；墨水屏通道也失败: %w", androidErr, err)
		}
		return nil, err
	}
	return &weeklyQueryResult{Data: data, Cred: c, Channel: "eink", AndroidErr: androidErr}, nil
}

// warnWeeklyOnce 对同一账号每天最多记录一条周奖励告警,
// 避免调度器每 5 分钟重复刷屏。
func (s *Server) warnWeeklyOnce(vid, msg string) {
	key := "weekly_warn:" + vid
	today := time.Now().Format("2006-01-02")
	if v, _ := store.GetSetting(s.db, key); v == today {
		return
	}
	_ = store.SetSetting(s.db, key, today)
	s.logf("warn", "reading", vid, "%s", msg)
}

// weeklyCardView 是体验卡的剩余状态,由详情缓存的会员卡信息计算。
type weeklyCardView struct {
	Has       bool  `json:"has"`  // 是否有未过期的体验卡
	Days      int   `json:"days"` // 剩余整天数(向下取整)
	RemainSec int64 `json:"remain_seconds"`
	ExpiredAt int64 `json:"expired_at"` // 到期时间戳(秒)
}

// memberCardView 解析详情缓存的会员卡 JSON({expired,expiredTime,remainTime,startTime}),
// 缓存为空或解析失败时返回零值(has=false),前端退回 weekly 响应里的无限卡字段。
func memberCardView(raw string) weeklyCardView {
	var mc struct {
		Expired     int   `json:"expired"`
		ExpiredTime int64 `json:"expiredTime"`
		RemainTime  int64 `json:"remainTime"`
	}
	if raw == "" || json.Unmarshal([]byte(raw), &mc) != nil {
		return weeklyCardView{}
	}
	has := mc.Expired == 0 && mc.RemainTime > 0
	return weeklyCardView{
		Has:       has,
		Days:      int(mc.RemainTime / 86400),
		RemainSec: mc.RemainTime,
		ExpiredAt: mc.ExpiredTime,
	}
}

// handleChallengeDetail 返回指定账号的官方挑战赛进度详情。
// GET /api/accounts/{vid}/challenge
func (s *Server) handleChallengeDetail(w http.ResponseWriter, r *http.Request) {
	vid := r.PathValue("vid")
	data, _, err := s.weeklyCall(r.Context(), vid, func(creds *weread.Credentials) (json.RawMessage, error) {
		return s.client.ChallengeDetail(r.Context(), creds)
	})
	if err != nil {
		writeWeeklyErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"vid":       vid,
		"challenge": json.RawMessage(data),
	})
}

func writeWeeklyErr(w http.ResponseWriter, err error) {
	if errors.Is(err, store.ErrNotFound) {
		writeErr(w, http.StatusNotFound, err)
		return
	}
	writeErr(w, http.StatusBadGateway, err)
}
