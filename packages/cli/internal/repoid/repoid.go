package repoid

import (
	"os/exec"
	"path/filepath"
	"strings"
)

func Detect(dir string) string {
	if remote, err := exec.Command("git", "-C", dir, "config", "--get", "remote.origin.url").Output(); err == nil {
		if url := strings.TrimSpace(string(remote)); url != "" {
			return normalizeRemote(url)
		}
	}

	if toplevel, err := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel").Output(); err == nil {
		if top := strings.TrimSpace(string(toplevel)); top != "" {
			return filepath.Base(top)
		}
	}

	return filepath.Base(dir)
}

func normalizeRemote(url string) string {
	url = strings.TrimPrefix(url, "git@")
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")

	if idx := strings.Index(url, ":"); idx != -1 && !strings.Contains(url[:idx], "/") {
		url = url[:idx] + "/" + url[idx+1:]
	}

	url = strings.TrimSuffix(url, ".git")

	return url
}
