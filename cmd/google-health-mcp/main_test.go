package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	mcpserver "github.com/shotah/google-health-mcp/internal/mcp"
)

func TestRunVersion(t *testing.T) {
	version = "v0.0.0-test"
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	code := run([]string{"--version"})
	_ = w.Close()
	os.Stdout = old
	out, _ := io.ReadAll(r)
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !bytes.Contains(out, []byte("v0.0.0-test")) {
		t.Fatalf("got %q", out)
	}
}

func TestRunSelfTest(t *testing.T) {
	if code := run([]string{"--self-test"}); code != 0 {
		t.Fatalf("exit %d", code)
	}
}

func TestRunAuthStub(t *testing.T) {
	if code := run([]string{"auth"}); code != 1 {
		t.Fatalf("want exit 1 for unimplemented auth, got %d", code)
	}
}

func TestRunBadFlag(t *testing.T) {
	if code := run([]string{"--not-a-real-flag"}); code != 2 {
		t.Fatalf("want 2, got %d", code)
	}
}

func TestRunServerError(t *testing.T) {
	old := runServer
	runServer = func(context.Context, *mcpserver.Server) error {
		return context.Canceled
	}
	t.Cleanup(func() { runServer = old })
	if code := run(nil); code != 1 {
		t.Fatalf("want 1, got %d", code)
	}
}
