package ghealth

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

const defaultBaseURL = "https://health.googleapis.com"

// CoreScopes are the readonly Google Health OAuth scopes for MVP tools.
var CoreScopes = []string{
	"https://www.googleapis.com/auth/googlehealth.sleep.readonly",
	"https://www.googleapis.com/auth/googlehealth.activity_and_fitness.readonly",
	"https://www.googleapis.com/auth/googlehealth.health_metrics_and_measurements.readonly",
	"https://www.googleapis.com/auth/googlehealth.profile.readonly",
}

// Client talks to Google Health API (health.googleapis.com/v4).
// Live data methods are stubs until OAuth + reconcile are wired (TODO.md).
type Client struct {
	ClientID     string
	ClientSecret string
	TokenPath    string
	BaseURL      string
	HTTPClient   *http.Client
	AccessToken  string // set after load; empty until auth
}

// MaybeFromEnv always returns a client; tokens may be missing.
func MaybeFromEnv() *Client {
	base := strings.TrimSpace(os.Getenv("GOOGLE_HEALTH_BASE_URL"))
	if base == "" {
		base = defaultBaseURL
	}
	return &Client{
		ClientID:     strings.TrimSpace(os.Getenv("GOOGLE_HEALTH_CLIENT_ID")),
		ClientSecret: strings.TrimSpace(os.Getenv("GOOGLE_HEALTH_CLIENT_SECRET")),
		TokenPath:    strings.TrimSpace(os.Getenv("GOOGLE_HEALTH_TOKEN_PATH")),
		BaseURL:      strings.TrimRight(base, "/"),
		HTTPClient:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) requireAuth() error {
	if c == nil {
		return fmt.Errorf("google health client not configured")
	}
	if strings.TrimSpace(c.AccessToken) == "" {
		return fmt.Errorf("Google Health auth required — run google-health-mcp auth (or gantry auth ghealth)")
	}
	return nil
}

// dataPointsURL builds .../v4/users/me/dataTypes/{type}/dataPoints[:suffix]
func (c *Client) dataPointsURL(dataType, suffix string) string {
	u := c.BaseURL + "/v4/users/me/dataTypes/" + dataType + "/dataPoints"
	if suffix != "" {
		u += ":" + suffix
	}
	return u
}

func (c *Client) SleepGet(_ context.Context, req SleepGetRequest) (*SleepResult, error) {
	if err := c.requireAuth(); err != nil {
		return nil, err
	}
	_ = c.dataPointsURL(DataTypeSleep, "reconcile")
	return nil, fmt.Errorf("ghealth SleepGet(%s) not implemented yet — see TODO.md", req.Date)
}

func (c *Client) ActivitiesList(_ context.Context, _ ActivitiesListRequest) (*ActivitiesListResult, error) {
	if err := c.requireAuth(); err != nil {
		return nil, err
	}
	_ = c.dataPointsURL(DataTypeExercise, "reconcile")
	return nil, fmt.Errorf("ghealth ActivitiesList not implemented yet — see TODO.md")
}

func (c *Client) ActivitiesGet(_ context.Context, id string) (*Activity, error) {
	if err := c.requireAuth(); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("ghealth ActivitiesGet(%s) not implemented yet — see TODO.md", id)
}

func (c *Client) HeartRateGet(_ context.Context, req HeartRateRequest) (*HeartRateResult, error) {
	if err := c.requireAuth(); err != nil {
		return nil, err
	}
	_ = c.dataPointsURL(DataTypeDailyRestingHeartRate, "reconcile")
	return nil, fmt.Errorf("ghealth HeartRateGet(%s) not implemented yet — see TODO.md", req.Date)
}

func (c *Client) HRVGet(_ context.Context, req HRVRequest) (*HRVResult, error) {
	if err := c.requireAuth(); err != nil {
		return nil, err
	}
	_ = c.dataPointsURL(DataTypeDailyHeartRateVariability, "reconcile")
	return nil, fmt.Errorf("ghealth HRVGet(%s) not implemented yet — see TODO.md", req.Date)
}

func (c *Client) WeightGet(_ context.Context, _ WeightRequest) (*WeightResult, error) {
	if err := c.requireAuth(); err != nil {
		return nil, err
	}
	_ = c.dataPointsURL(DataTypeWeight, "reconcile")
	return nil, fmt.Errorf("ghealth WeightGet not implemented yet — see TODO.md")
}

func (c *Client) ProfileGet(_ context.Context) (*Profile, error) {
	if err := c.requireAuth(); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("ghealth ProfileGet not implemented yet — see TODO.md")
}

// AccountStatus describes auth without calling the API (when possible).
func (c *Client) AccountStatus() map[string]any {
	authed := c != nil && strings.TrimSpace(c.AccessToken) != ""
	tokenPath := ""
	base := defaultBaseURL
	if c != nil {
		tokenPath = c.TokenPath
		if c.BaseURL != "" {
			base = c.BaseURL
		}
	}
	return map[string]any{
		"provider":       "Google Health API",
		"base_url":       base,
		"authenticated":  authed,
		"client_id_set":  c != nil && c.ClientID != "",
		"token_path":     tokenPath,
		"scopes":         CoreScopes,
		"docs_url":       "https://developers.google.com/health/about",
		"implementation": "stub — OAuth + reconcile pending (TODO.md)",
		"boot_note":      "Missing tokens do not prevent MCP initialize; data tools error until auth.",
		"host_id":        "ghealth",
	}
}
