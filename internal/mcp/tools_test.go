package mcp

import (
	"regexp"
	"strings"
	"testing"

	"github.com/shotah/google-health-mcp/internal/ghealth"
)

var toolNameRE = regexp.MustCompile(`^[a-z]+(_[a-z0-9]+)+$`)

func TestRegisteredToolNames(t *testing.T) {
	names := RegisteredToolNames()
	if len(names) != 8 {
		t.Fatalf("want 8 tools, got %d: %v", len(names), names)
	}
	for _, n := range names {
		if strings.HasPrefix(n, "ghealth_") || strings.HasPrefix(n, "health_") || strings.HasPrefix(n, "fitbit_") {
			t.Fatalf("tool %q must not start with host/provider prefix", n)
		}
		if !toolNameRE.MatchString(n) {
			t.Fatalf("tool %q bad pattern", n)
		}
	}
}

func TestSelfTest(t *testing.T) {
	if err := SelfTest(); err != nil {
		t.Fatal(err)
	}
}

func TestSleepGetStub(t *testing.T) {
	s := New(ghealth.NewStub())
	res, out, err := s.sleepGet(t.Context(), nil, sleepGetInput{Date: "2026-08-03"})
	if err != nil {
		t.Fatal(err)
	}
	if res != nil && res.IsError {
		t.Fatalf("%+v", res)
	}
	if out == nil {
		t.Fatal("nil out")
	}
}

func TestActivitiesGetRequiresID(t *testing.T) {
	s := New(ghealth.NewStub())
	res, _, err := s.activitiesGet(t.Context(), nil, activitiesGetInput{})
	if err != nil {
		t.Fatal(err)
	}
	if res == nil || !res.IsError {
		t.Fatal("expected error")
	}
}

func TestAccountGet(t *testing.T) {
	s := New(ghealth.NewStub())
	_, out, err := s.accountGet(t.Context(), nil, accountGetInput{})
	if err != nil {
		t.Fatal(err)
	}
	m, ok := out.(map[string]any)
	if !ok || m["provider"] != "Google Health API" {
		t.Fatalf("%#v", out)
	}
}
