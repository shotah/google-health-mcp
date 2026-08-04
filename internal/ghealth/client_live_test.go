package ghealth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func testClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return &Client{
		BaseURL:         srv.URL,
		HTTPClient:      srv.Client(),
		AccessToken:     "test-token",
		PreferWearables: true,
		mu:              &sync.Mutex{},
	}
}

func TestSleepGetReconcile(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/dataTypes/sleep/dataPoints:reconcile") {
			t.Fatalf("path %s", r.URL.Path)
		}
		if r.URL.Query().Get("dataSourceFamily") != DataSourceFamilyGoogleWearables {
			t.Fatalf("family %q", r.URL.Query().Get("dataSourceFamily"))
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Fatalf("auth %q", r.Header.Get("Authorization"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"dataPoints": []map[string]any{
				{
					"name": "users/me/dataTypes/sleep/dataPoints/abc",
					"sleep": map[string]any{
						"interval": map[string]any{
							"startTime": "2026-08-02T22:00:00Z",
							"endTime":   "2026-08-03T06:00:00Z",
						},
						"metadata": map[string]any{"main": true, "processed": true},
						"summary": map[string]any{
							"minutesAsleep":        "420",
							"minutesInSleepPeriod": "480",
							"minutesAwake":         "60",
						},
					},
				},
			},
		})
	})
	res, err := c.SleepGet(t.Context(), SleepGetRequest{Date: "2026-08-03"})
	if err != nil {
		t.Fatal(err)
	}
	if res.DurationMinutes != 420 {
		t.Fatalf("duration %#v", res)
	}
	if res.Efficiency < 87 || res.Efficiency > 88 {
		t.Fatalf("efficiency %v", res.Efficiency)
	}
}

func TestActivitiesListAndGet(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "dataPoints:reconcile"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"dataPoints": []map[string]any{
					{
						"name": "users/me/dataTypes/exercise/dataPoints/ex1",
						"exercise": map[string]any{
							"displayName":    "Walk",
							"exerciseType":   "WALKING",
							"activeDuration": "1800s",
							"interval":       map[string]any{"startTime": "2026-08-03T10:00:00Z"},
							"metricsSummary": map[string]any{"caloriesKcal": 120, "steps": "2500", "distanceMillimeters": 2_000_000},
						},
					},
				},
			})
		case strings.Contains(r.URL.Path, "/dataPoints/ex1"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"name": "users/me/dataTypes/exercise/dataPoints/ex1",
				"exercise": map[string]any{
					"displayName":    "Walk",
					"exerciseType":   "WALKING",
					"activeDuration": "1800s",
					"interval":       map[string]any{"startTime": "2026-08-03T10:00:00Z"},
					"metricsSummary": map[string]any{"caloriesKcal": 120},
				},
			})
		default:
			http.NotFound(w, r)
		}
	})
	list, err := c.ActivitiesList(t.Context(), ActivitiesListRequest{
		AfterDate:  "2026-08-01",
		BeforeDate: "2026-08-04",
		Limit:      10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if list.Count != 1 || list.Activities[0].Name != "Walk" {
		t.Fatalf("%#v", list)
	}
	if list.Activities[0].DurationMin != 30 {
		t.Fatalf("duration %v", list.Activities[0].DurationMin)
	}
	got, err := c.ActivitiesGet(t.Context(), "ex1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Walk" {
		t.Fatalf("%#v", got)
	}
}

func TestHeartRateAndHRV(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "daily-resting-heart-rate"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"dataPoints": []map[string]any{{
					"dailyRestingHeartRate": map[string]any{
						"date":           map[string]any{"year": 2026, "month": 8, "day": 3},
						"beatsPerMinute": "54",
					},
				}},
			})
		case strings.Contains(r.URL.Path, "daily-heart-rate-variability"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"dataPoints": []map[string]any{{
					"dailyHeartRateVariability": map[string]any{
						"date": map[string]any{"year": 2026, "month": 8, "day": 3},
						"averageHeartRateVariabilityMilliseconds": 42.5,
					},
				}},
			})
		default:
			http.NotFound(w, r)
		}
	})
	hr, err := c.HeartRateGet(t.Context(), HeartRateRequest{Date: "2026-08-03"})
	if err != nil {
		t.Fatal(err)
	}
	if hr.RestingHR != 54 {
		t.Fatalf("%#v", hr)
	}
	hrv, err := c.HRVGet(t.Context(), HRVRequest{Date: "2026-08-03"})
	if err != nil {
		t.Fatal(err)
	}
	if hrv.RMSSD != 42.5 {
		t.Fatalf("%#v", hrv)
	}
}

func TestWeightAndProfile(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/identity"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"legacyUserId": "FITBIT1",
				"healthUserId": "G123",
			})
		case strings.Contains(r.URL.Path, "dataTypes/weight"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"dataPoints": []map[string]any{{
					"name": "users/me/dataTypes/weight/dataPoints/w1",
					"weight": map[string]any{
						"weightGrams": 82000,
						"sampleTime": map[string]any{
							"physicalTime": "2026-08-03T08:00:00Z",
							"civilTime":    map[string]any{"date": map[string]any{"year": 2026, "month": 8, "day": 3}},
						},
					},
				}},
			})
		default:
			http.NotFound(w, r)
		}
	})
	wres, err := c.WeightGet(t.Context(), WeightRequest{Period: "1m"})
	if err != nil {
		t.Fatal(err)
	}
	if len(wres.Entries) != 1 {
		t.Fatalf("%#v", wres)
	}
	prof, err := c.ProfileGet(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if prof.FitbitUser != "FITBIT1" || prof.GoogleUser != "G123" {
		t.Fatalf("%#v", prof)
	}
}

func TestAPIError(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"nope"}`))
	})
	if _, err := c.SleepGet(t.Context(), SleepGetRequest{Date: "2026-08-03"}); err == nil {
		t.Fatal("expected error")
	}
}

func TestEmptyDailyResults(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"dataPoints": []any{}})
	})
	hr, err := c.HeartRateGet(t.Context(), HeartRateRequest{Date: "2026-08-03"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(hr.Note, "no daily") {
		t.Fatalf("%#v", hr)
	}
	hrv, err := c.HRVGet(t.Context(), HRVRequest{Date: "2026-08-03"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(hrv.Note, "no daily") {
		t.Fatalf("%#v", hrv)
	}
}

func TestAccountStatusLive(t *testing.T) {
	c2 := &Client{AccessToken: "tok", PreferWearables: true, mu: &sync.Mutex{}}
	st := c2.AccountStatus()
	if st["authenticated"] != true {
		t.Fatalf("%#v", st)
	}
	if st["implementation"] != "live health.googleapis.com/v4 reconcile" {
		t.Fatalf("%#v", st["implementation"])
	}
}
