package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

const repo = "mdawess/amp"

type UpdateCmd struct{}

func (u *UpdateCmd) Run() error {
	tag, err := latestTag()
	if err != nil {
		return fmt.Errorf("fetch latest release: %w", err)
	}

	asset := fmt.Sprintf("amp-%s-%s", runtime.GOOS, runtime.GOARCH)
	url := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", repo, tag, asset)

	fmt.Printf("Downloading %s/%s...\n", tag, asset)
	body, err := download(url)
	if err != nil {
		return fmt.Errorf("download %s: %w", url, err)
	}
	defer body.Close()

	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locate current binary: %w", err)
	}
	self, err = filepath.EvalSymlinks(self)
	if err != nil {
		return fmt.Errorf("resolve symlink: %w", err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(self), "amp-update-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer func() { os.Remove(tmpName) }()

	if _, err := io.Copy(tmp, body); err != nil {
		tmp.Close()
		return fmt.Errorf("write download: %w", err)
	}
	tmp.Close()

	if err := os.Chmod(tmpName, 0o755); err != nil {
		return err
	}
	if err := os.Rename(tmpName, self); err != nil {
		return fmt.Errorf("replace binary: %w", err)
	}

	fmt.Printf("Updated to %s\n", tag)
	return nil
}

func latestTag() (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	resp, err := http.Get(url) //nolint:noctx
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var payload struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if payload.TagName == "" {
		return "", fmt.Errorf("empty tag in response")
	}
	return payload.TagName, nil
}

func download(url string) (io.ReadCloser, error) {
	resp, err := http.Get(url) //nolint:noctx
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return resp.Body, nil
}
