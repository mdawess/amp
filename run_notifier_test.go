package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExpandNotificationTemplate(t *testing.T) {
	cases := []struct {
		command string
		payload NotifyPayload
		want    string
	}{
		{
			"echo {{branch}} {{status}} {{summary}}",
			NotifyPayload{Branch: "feat-x", Status: RunStatusComplete, Summary: "all done"},
			"echo feat-x complete all done",
		},
		{
			"notify {{branch}}",
			NotifyPayload{Branch: "fix/bug", Status: RunStatusError, Summary: ""},
			"notify fix/bug",
		},
		{
			"plain command",
			NotifyPayload{},
			"plain command",
		},
	}
	for _, c := range cases {
		got := expandNotificationTemplate(c.command, c.payload)
		if got != c.want {
			t.Errorf("expandNotificationTemplate(%q) = %q, want %q", c.command, got, c.want)
		}
	}
}

func TestSendNotification_empty(t *testing.T) {
	// Should not panic or error when command is empty
	sendNotification(context.Background(), "", NotifyPayload{Branch: "x", Status: RunStatusComplete})
}

func TestSendNotification_runs(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.txt")
	cmd := "echo {{branch}} > " + out
	sendNotification(context.Background(), cmd, NotifyPayload{Branch: "my-branch", Status: RunStatusComplete, Summary: "ok"})
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("output file not created: %v", err)
	}
	if !strings.Contains(string(data), "my-branch") {
		t.Errorf("expected branch in output, got %q", string(data))
	}
}
