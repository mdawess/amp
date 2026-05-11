package run

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type RunStatus string

const (
	RunStatusRunning  RunStatus = "running"
	RunStatusComplete RunStatus = "complete"
	RunStatusError    RunStatus = "error"
)

type RunState struct {
	Branch           string    `json:"branch"`
	Status           RunStatus `json:"status"`
	StartTime        time.Time `json:"start_time"`
	EndTime          time.Time `json:"end_time,omitempty"`
	Prompt           string    `json:"prompt"`
	TmuxSession      string    `json:"tmux_session"`
	LogPath          string    `json:"log_path"`
	CompletionSignal string    `json:"completion_signal,omitempty"`
	ExitCode         int       `json:"exit_code,omitempty"`
}

func SanitizeBranch(branch string) string {
	return strings.NewReplacer("/", "-", "\\", "-", ":", "-").Replace(branch)
}

func RunsDir(repoRoot string) string {
	return filepath.Join(repoRoot, ".amp", "runs")
}

func SaveRun(repoRoot string, state RunState) error {
	dir := RunsDir(repoRoot)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, SanitizeBranch(state.Branch)+".json"), data, 0o644)
}

func LoadRun(repoRoot, branch string) (RunState, error) {
	data, err := os.ReadFile(filepath.Join(RunsDir(repoRoot), SanitizeBranch(branch)+".json"))
	if err != nil {
		return RunState{}, err
	}
	var state RunState
	return state, json.Unmarshal(data, &state)
}

func ListRuns(repoRoot string) ([]RunState, error) {
	entries, err := os.ReadDir(RunsDir(repoRoot))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var states []RunState
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(RunsDir(repoRoot), e.Name()))
		if err != nil {
			continue
		}
		var s RunState
		if err := json.Unmarshal(data, &s); err != nil {
			continue
		}
		states = append(states, s)
	}
	return states, nil
}

func EnsureAmpDirGitignored(repoRoot string) error {
	path := filepath.Join(repoRoot, ".gitignore")
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if strings.Contains(string(data), ".amp/") {
		return nil
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	prefix := ""
	if len(data) > 0 && data[len(data)-1] != '\n' {
		prefix = "\n"
	}
	_, err = f.WriteString(prefix + ".amp/\n")
	return err
}
