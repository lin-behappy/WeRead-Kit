package weread

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestAndroidCredentialsValidate(t *testing.T) {
	valid := &AndroidCredentials{
		AccessToken: "at", DeviceID: "dev", UserAgent: "ua",
		AppVersion: "10.1.0.10167436", BaseAPI: "29", OSVersion: "10", ChannelID: "1",
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid credentials rejected: %v", err)
	}
	valid.AccessToken = ""
	if err := valid.Validate(); err == nil || !strings.Contains(err.Error(), "access_token") {
		t.Fatalf("missing access_token error = %v", err)
	}
}

func TestWeeklyExchangeAndroidUsesAndroidFingerprint(t *testing.T) {
	client := NewClient()
	var headers http.Header
	var payload map[string]any
	client.HTTP.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/weekly/exchange" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		headers = r.Header.Clone()
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("request JSON: %v", err)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body: io.NopCloser(strings.NewReader(`{
				"readdayAwards":[{
					"awardLevelId":13,"awardStatus":1,
					"awardChoices":[{"choiceType":1,"awardNum":2},{"choiceType":2,"awardNum":6}]
				}]
			}`)),
			Header: make(http.Header),
		}, nil
	})

	creds := &AndroidCredentials{
		Vid:          "443553272",
		AccessToken:  "at",
		RefreshToken: "rt",
		DeviceID:     "dev",
		UserAgent:    "WeRead/10.1.0 WRBrand/xiaomi Dalvik/2.1.0 (Linux; U; Android 10; Redmi Note 7 Pro Build/QKQ1.190915.002)",
		AppVersion:   "10.1.0.10167436",
		BaseAPI:      "29",
		OSVersion:    "10",
		ChannelID:    "1",
	}
	data, err := client.WeeklyExchangeAndroid(context.Background(), creds, 0, 0, 0)
	if err != nil {
		t.Fatalf("WeeklyExchangeAndroid: %v", err)
	}

	for k, want := range map[string]string{
		"baseapi":     "29",
		"appver":      "10.1.0.10167436",
		"basever":     "10.1.0.10167436",
		"osver":       "10",
		"channelId":   "1",
		"User-Agent":  creds.UserAgent,
		"vid":         "443553272",
		"accessToken": "at",
	} {
		if got := headers.Get(k); got != want {
			t.Errorf("header %s = %q, want %q", k, got, want)
		}
	}
	for k, want := range map[string]float64{
		"awardLevelId":    0,
		"awardChoiceType": 0,
		"isExchangeAward": 0,
		"isVisitReadGoal": 1,
		"unread":          0,
	} {
		if got, ok := payload[k].(float64); !ok || got != want {
			t.Errorf("payload %s = %#v, want %v", k, payload[k], want)
		}
	}
	if got := payload["pf"]; got != androidWeeklyPF {
		t.Errorf("payload pf = %#v, want %q", got, androidWeeklyPF)
	}

	var rw WeeklyRewards
	if err := json.Unmarshal(data, &rw); err != nil {
		t.Fatalf("response JSON: %v", err)
	}
	if len(rw.ReaddayAwards) != 1 || len(rw.ReaddayAwards[0].AwardChoices) != 2 {
		t.Fatalf("unexpected award response: %+v", rw.ReaddayAwards)
	}
}

func TestCheckBusinessCodeLowercaseAndroid(t *testing.T) {
	if err := checkBusinessCode([]byte(`{"errcode":-2012,"errmsg":"登录超时"}`)); err != ErrSessionExpired {
		t.Fatalf("errcode -2012 = %v, want ErrSessionExpired", err)
	}
}
