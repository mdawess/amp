package main

import (
	"context"
	"log"
	"strings"
)

type NotifyPayload struct {
	Branch  string
	Status  RunStatus
	Summary string
}

func sendNotification(ctx context.Context, command string, p NotifyPayload) {
	if command == "" {
		return
	}
	expanded := expandNotificationTemplate(command, p)
	result := runHook(ctx, "", Hook{Command: expanded})
	if result.Err != nil {
		log.Printf("notification hook failed: %v (stderr: %s)", result.Err, result.Stderr)
	}
}

func expandNotificationTemplate(command string, p NotifyPayload) string {
	return strings.NewReplacer(
		"{{branch}}", p.Branch,
		"{{status}}", string(p.Status),
		"{{summary}}", p.Summary,
	).Replace(command)
}
