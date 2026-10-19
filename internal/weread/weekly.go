// 周阅读奖励:阅读时长/天数折算成积分,可兑换体验卡天数或书币。
// weekly/exchange 同一端点承载查询与领取,靠 isExchangeAward 参数区分。
package weread

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// weeklyPF 是奖励领取接口(i.weread.qq.com/weekly/exchange)的平台标识。
// 服务端按它记录奖励的领取端,两端的真值格式并不对称:
//   - 安卓:wechat_wx-2001-android-100-weread
//   - iOS: weread_wx-2001-iap-2001-iphone
//
// 默认用安卓串,避免奖励被记为 iOS 端(iOS 渠道币)后在安卓客户端无法使用。
// 服务端只在真正领取(isExchangeAward=1)时校验,传其它值会报 errcode -2003;
// 可用环境变量 WEREAD_PF 覆盖(如需恢复 iOS 行为)。
var weeklyPF = envOr("WEREAD_PF", "wechat_wx-2001-android-100-weread")

// envOr 返回环境变量的值,为空时回落到默认值。
func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// 奖励档位状态。
const (
	AwardStatusLocked    = 0 // 未达成
	AwardStatusClaimable = 1 // 可领取
	AwardStatusClaimed   = 2 // 已领取
)

// 奖品类型(awardChoices[].choiceType 与领取时的 awardChooseType)。
const (
	AwardChoiceCard     = 1 // 体验卡天数
	AwardChoiceBookCoin = 2 // 书币
)

// WeeklyAwardChoice 是一档奖励下的可选奖品。
type WeeklyAwardChoice struct {
	ChoiceType int `json:"choiceType"` // 1=体验卡天数 2=书币
	AwardNum   int `json:"awardNum"`   // 数量:天数或书币个数
	CanChoice  int `json:"canChoice"`
}

// WeeklyAward 是一档周奖励。
type WeeklyAward struct {
	AwardLevelID     int                 `json:"awardLevelId"`
	AwardChooseType  int                 `json:"awardChooseType"` // 已领取时记录当时选的奖品类型
	AwardStatus      int                 `json:"awardStatus"`
	AwardStatusDesc  string              `json:"awardStatusDesc"`  // 领取 / 已领取 / 差120分钟 / 差1天
	AwardLevelDesc   string              `json:"awardLevelDesc"`   // 读 30 分钟 / 读 2 天
	AwardChoicesDesc string              `json:"awardChoicesDesc"` // 可得 1 天体验卡或 1 书币(服务端原文用"无限卡")
	AwardChoices     []WeeklyAwardChoice `json:"awardChoices"`
}

// WeeklyPeriodDetail 是按天展开的阅读秒数(本周 7 项 / 本月 30 项)。
type WeeklyPeriodDetail struct {
	ReadDays  []int `json:"readDays"`
	ReadTimes int   `json:"readTimes"`
}

// WeeklyRewards 是 weekly/exchange 查询模式的响应(节选前端需要的字段)。
type WeeklyRewards struct {
	ReadingTime       int           `json:"readingTime"` // 本周阅读秒数
	ReadingDay        int           `json:"readingDay"`  // 本周阅读天数
	IsMCardVip        int           `json:"isMCardVip"`
	ValidDayMinSecond int           `json:"validDayMinSecond"`
	ReadtimeAwards    []WeeklyAward `json:"readtimeAwards"`
	ReaddayAwards     []WeeklyAward `json:"readdayAwards"`
	ReadgoalAwards    []WeeklyAward `json:"readgoalAwards"`
	InfiniteCard      struct {
		Day      int    `json:"day"`
		Paying   int    `json:"paying"`
		ItemID   string `json:"itemId"`
		CardType string `json:"cardType"`
	} `json:"infiniteCard"`
	WeekReadDaysDetail   WeeklyPeriodDetail `json:"weekReadDaysDetail"`
	MonthReadDaysDetail  WeeklyPeriodDetail `json:"monthReadDaysDetail"`
	ReadTimeGears        []int              `json:"readTimeGears"`
	RewardRules          []string           `json:"rewardRules"`
	AwardStatusOuterDesc string             `json:"awardStatusOuterDesc"`
}

// ParseWeeklyRewards 解析查询模式的原始响应。
func ParseWeeklyRewards(data json.RawMessage) (*WeeklyRewards, error) {
	var rw WeeklyRewards
	if err := json.Unmarshal(data, &rw); err != nil {
		return nil, fmt.Errorf("周阅读奖励响应解析失败: %w", err)
	}
	return &rw, nil
}

// ChallengeDetail 查询官方阅读挑战赛的当前进度详情(报名中的那一场)。
// 未报名任何挑战时返回 status=0 的空壳响应,由调用方自行判断。
// 会话过期返回 ErrSessionExpired,由调用方续期后重试。
func (c *Client) ChallengeDetail(ctx context.Context, creds *Credentials) (json.RawMessage, error) {
	headers := versionHeaders()
	for k, v := range authHeaders(creds) {
		headers[k] = v
	}
	data, status, err := c.getJSON(ctx, BaseURL+"/challenge/detail?scene=1", headers)
	if err != nil {
		return nil, fmt.Errorf("挑战赛详情请求失败: %w", err)
	}
	if status == http.StatusUnauthorized {
		return nil, ErrSessionExpired
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("挑战赛详情请求失败: HTTP %d", status)
	}
	if err := checkBusinessCode(data); err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// WeeklyExchange 查询或领取周阅读奖励。
// isExchangeAward=0 为查询(其余参数填 0);=1 时 awardLevelID 与 awardChooseType
// 指定领取哪一档的哪种奖品(1=体验卡天数 2=书币)。
// 注意:领取参数名是 awardChoiceType,与响应里的 awardChooseType 拼写不一致,
// 实测用后者会被服务端以 errcode -2417 拒绝;pf 必填,缺失报 -2003。
// 返回原始 JSON;会话过期返回 ErrSessionExpired,由调用方续期后重试。
func (c *Client) WeeklyExchange(ctx context.Context, creds *Credentials, awardLevelID, awardChooseType, isExchangeAward int) (json.RawMessage, error) {
	headers := versionHeaders()
	for k, v := range authHeaders(creds) {
		headers[k] = v
	}
	payload := map[string]any{
		"awardLevelId":    awardLevelID,
		"awardChoiceType": awardChooseType,
		"isExchangeAward": isExchangeAward,
		"pf":              weeklyPF,
	}
	data, status, err := c.postJSON(ctx, BaseURL+"/weekly/exchange", headers, payload)
	if err != nil {
		return nil, fmt.Errorf("周阅读奖励请求失败: %w", err)
	}
	if status == http.StatusUnauthorized {
		return nil, ErrSessionExpired
	}
	if status != http.StatusOK {
		// 非 200 的响应体里通常带 errcode/errmsg(如 -2417/-2003),一并带回方便排查。
		var env struct {
			ErrCode json.Number `json:"errcode"`
			ErrMsg  string      `json:"errmsg"`
		}
		if json.Unmarshal(data, &env) == nil && env.ErrCode.String() != "" {
			return nil, fmt.Errorf("周阅读奖励请求失败: HTTP %d errcode=%s errmsg=%q", status, env.ErrCode.String(), env.ErrMsg)
		}
		return nil, fmt.Errorf("周阅读奖励请求失败: HTTP %d", status)
	}
	if err := checkBusinessCode(data); err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}
