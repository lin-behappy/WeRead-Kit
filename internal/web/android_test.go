package web

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/27Aaron/weread-kit/internal/store"
	"github.com/27Aaron/weread-kit/internal/weread"
)

type webRoundTripFunc func(*http.Request) (*http.Response, error)

func (f webRoundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestAndroidCredentialAPI(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()
	if err := store.Save(db, &store.Credential{Vid: "v1", RefreshToken: "rt", DeviceID: "dev"}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	s := &Server{db: db}

	call := func(method, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, "/api/accounts/v1/android", strings.NewReader(body))
		req.SetPathValue("vid", "v1")
		rec := httptest.NewRecorder()
		s.handleAndroidCredential(rec, req)
		return rec
	}

	rec := call(http.MethodPost, `{
		"access_token":"at","device_id":"dev","user_agent":"ua",
		"app_version":"10.1.0.10167436","baseapi":"29","osver":"10","channel_id":"1"
	}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST valid: got %d %s", rec.Code, rec.Body.String())
	}

	rec = call(http.MethodGet, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET: got %d %s", rec.Code, rec.Body.String())
	}
	var meta map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &meta); err != nil {
		t.Fatalf("GET JSON: %v", err)
	}
	if meta["configured"] != true || meta["valid"] != true {
		t.Fatalf("GET meta = %#v", meta)
	}
	if _, ok := meta["access_token"]; ok {
		t.Fatal("GET leaked access_token")
	}
	if _, ok := meta["refresh_token"]; ok {
		t.Fatal("GET leaked refresh_token")
	}

	rec = call(http.MethodPost, `{"access_token":"at"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("POST incomplete: got %d %s", rec.Code, rec.Body.String())
	}

	rec = call(http.MethodDelete, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("DELETE: got %d %s", rec.Code, rec.Body.String())
	}
	rec = call(http.MethodGet, "")
	if err := json.Unmarshal(rec.Body.Bytes(), &meta); err != nil {
		t.Fatalf("GET after delete JSON: %v", err)
	}
	if meta["configured"] != false {
		t.Fatalf("configured after delete = %#v", meta["configured"])
	}
}

// TestWeeklyClaimCoinRequiresAndroid 回归:书币领取不能回退墨水屏,
// 否则用户点“书币”会被悄悄改领体验卡。
func TestWeeklyClaimCoinRequiresAndroid(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()
	if err := store.Save(db, &store.Credential{Vid: "v1", RefreshToken: "rt", DeviceID: "dev"}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	s := &Server{db: db, client: weread.NewClient()}

	req := httptest.NewRequest(http.MethodPost, "/api/accounts/v1/weekly/claim",
		strings.NewReader(`{"award_level_id":13,"choice_type":2}`))
	req.SetPathValue("vid", "v1")
	rec := httptest.NewRecorder()
	s.handleWeeklyClaim(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("coin claim without android: got %d %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "未配置安卓客户端凭据") {
		t.Fatalf("unexpected error: %s", rec.Body.String())
	}
	if prefs := s.weeklyClaimPrefs("v1"); len(prefs) != 0 {
		t.Fatalf("failed coin claim saved prefs: %#v", prefs)
	}
}

// TestWeeklyClaimLowerLevelCoinUsesEink 验证只有 7 天档书币走安卓,
// 其他档位保持原来的墨水屏领取端。
func TestWeeklyClaimLowerLevelCoinUsesEink(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()
	if err := store.Save(db, &store.Credential{Vid: "v1", RefreshToken: "eink-rt", DeviceID: "eink-dev", AccessToken: "eink-at"}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	android := weread.AndroidCredentials{
		AccessToken: "android-at", DeviceID: "android-dev",
		UserAgent: "ua", AppVersion: "10.1.0.10167436", BaseAPI: "29", OSVersion: "10", ChannelID: "1",
	}
	raw, _ := json.Marshal(android)
	if err := store.SaveAndroid(db, "v1", string(raw)); err != nil {
		t.Fatalf("SaveAndroid: %v", err)
	}
	s := &Server{db: db, client: weread.NewClient()}
	var headers http.Header
	var payload map[string]any
	s.client.HTTP.Transport = webRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		headers = r.Header.Clone()
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("request JSON: %v", err)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"succ":1}`)),
			Header:     make(http.Header),
		}, nil
	})

	req := httptest.NewRequest(http.MethodPost, "/api/accounts/v1/weekly/claim",
		strings.NewReader(`{"award_level_id":5,"choice_type":2}`))
	req.SetPathValue("vid", "v1")
	rec := httptest.NewRecorder()
	s.handleWeeklyClaim(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("claim: got %d %s", rec.Code, rec.Body.String())
	}
	if got := headers.Get("baseapi"); got != "30" {
		t.Errorf("baseapi = %q, want 30 (e-ink)", got)
	}
	if got := headers.Get("accessToken"); got != "eink-at" {
		t.Errorf("accessToken = %q, want e-ink token", got)
	}
	if got := payload["awardLevelId"]; got != float64(5) {
		t.Errorf("awardLevelId = %#v", got)
	}
	if got := s.weeklyClaimPrefs("v1")["5"]; got != 2 {
		t.Errorf("pref 5 = %d, want 2", got)
	}
}

// TestAutoClaimUsesAndroidForCoins 验证调度器:查询和 7 天档书币领取都走安卓头,
// 不会用墨水屏 token 去领 choiceType=2。
func TestAutoClaimUsesAndroidForCoins(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()
	if err := store.Save(db, &store.Credential{Vid: "v1", RefreshToken: "eink-rt", DeviceID: "eink-dev", AccessToken: "eink-at"}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	android := weread.AndroidCredentials{
		AccessToken: "android-at", DeviceID: "android-dev",
		UserAgent:  "WeRead/10.1.0 WRBrand/xiaomi Dalvik/2.1.0 (Linux; U; Android 10; Redmi Note 7 Pro Build/QKQ1.190915.002)",
		AppVersion: "10.1.0.10167436", BaseAPI: "29", OSVersion: "10", ChannelID: "1",
	}
	raw, err := json.Marshal(android)
	if err != nil {
		t.Fatalf("marshal android: %v", err)
	}
	if err := store.SaveAndroid(db, "v1", string(raw)); err != nil {
		t.Fatalf("SaveAndroid: %v", err)
	}

	s := &Server{db: db, client: weread.NewClient()}
	if err := s.saveWeeklyClaimPref("v1", 13, weread.AwardChoiceBookCoin); err != nil {
		t.Fatalf("saveWeeklyClaimPref: %v", err)
	}

	type request struct {
		headers http.Header
		body    map[string]any
	}
	var calls []request
	s.client.HTTP.Transport = webRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		body, _ := io.ReadAll(r.Body)
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("request JSON: %v", err)
		}
		calls = append(calls, request{headers: r.Header.Clone(), body: payload})
		resp := `{"succ":1}`
		if payload["isExchangeAward"] == float64(0) {
			resp = `{"readdayAwards":[{
				"awardLevelId":13,"awardStatus":1,"awardLevelDesc":"读 7 天",
				"awardChoices":[{"choiceType":1,"awardNum":2},{"choiceType":2,"awardNum":6}]
			}]}`
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(resp)),
			Header:     make(http.Header),
		}, nil
	})

	s.autoClaimOnce(context.Background())
	if len(calls) != 2 {
		t.Fatalf("requests = %d, want 2 (query + claim)", len(calls))
	}
	for i, call := range calls {
		if got := call.headers.Get("baseapi"); got != "29" {
			t.Errorf("request %d baseapi = %q, want 29 (e-ink token used?)", i, got)
		}
		if got := call.headers.Get("accessToken"); got != "android-at" {
			t.Errorf("request %d accessToken = %q, want android token", i, got)
		}
	}
	claim := calls[1]
	if got := claim.body["awardLevelId"]; got != float64(13) {
		t.Errorf("claim awardLevelId = %#v", got)
	}
	if got := claim.body["awardChoiceType"]; got != float64(2) {
		t.Errorf("claim awardChoiceType = %#v, want 2", got)
	}
	if got := claim.body["pf"]; got != "wechat_wx-2001-android-100-weread" {
		t.Errorf("claim pf = %#v", got)
	}
}

// TestAutoClaimLowerLevelCoinUsesEink 验证调度器不会把其他档位的书币也改走安卓。
func TestAutoClaimLowerLevelCoinUsesEink(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()
	if err := store.Save(db, &store.Credential{Vid: "v1", RefreshToken: "eink-rt", DeviceID: "eink-dev", AccessToken: "eink-at"}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	android := weread.AndroidCredentials{
		AccessToken: "android-at", DeviceID: "android-dev",
		UserAgent: "ua", AppVersion: "10.1.0.10167436", BaseAPI: "29", OSVersion: "10", ChannelID: "1",
	}
	raw, _ := json.Marshal(android)
	if err := store.SaveAndroid(db, "v1", string(raw)); err != nil {
		t.Fatalf("SaveAndroid: %v", err)
	}
	s := &Server{db: db, client: weread.NewClient()}
	if err := s.saveWeeklyClaimPref("v1", 5, weread.AwardChoiceBookCoin); err != nil {
		t.Fatalf("saveWeeklyClaimPref: %v", err)
	}

	type request struct {
		headers http.Header
		body    map[string]any
	}
	var calls []request
	s.client.HTTP.Transport = webRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		body, _ := io.ReadAll(r.Body)
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("request JSON: %v", err)
		}
		calls = append(calls, request{headers: r.Header.Clone(), body: payload})
		resp := `{"succ":1}`
		if payload["isExchangeAward"] == float64(0) {
			resp = `{"readtimeAwards":[{
				"awardLevelId":5,"awardStatus":1,"awardLevelDesc":"读 1 小时",
				"awardChoices":[{"choiceType":1,"awardNum":1},{"choiceType":2,"awardNum":1}]
			}]}`
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(resp)),
			Header:     make(http.Header),
		}, nil
	})

	s.autoClaimOnce(context.Background())
	if len(calls) != 2 {
		t.Fatalf("requests = %d, want 2 (query + claim)", len(calls))
	}
	claim := calls[1]
	if got := claim.headers.Get("baseapi"); got != "30" {
		t.Errorf("claim baseapi = %q, want 30 (e-ink)", got)
	}
	if got := claim.headers.Get("accessToken"); got != "eink-at" {
		t.Errorf("claim accessToken = %q, want e-ink token", got)
	}
	if got := claim.body["awardLevelId"]; got != float64(5) {
		t.Errorf("claim awardLevelId = %#v", got)
	}
	if got := claim.body["awardChoiceType"]; got != float64(2) {
		t.Errorf("claim awardChoiceType = %#v", got)
	}
}
