package schema

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// remoteCacheDir is where remote schemas are cached.
const remoteCacheDir = ".envguard/cache/schemas"

// remoteTimeout is the HTTP timeout for fetching remote schemas.
const remoteTimeout = 30 * time.Second

// fetchRemoteSchema downloads a schema from a remote URL.
// It caches the file locally for subsequent loads.
func fetchRemoteSchema(remoteURL string) (string, error) {
	u, err := url.Parse(remoteURL)
	if err != nil {
		return "", fmt.Errorf("invalid remote URL: %w", err)
	}

	cacheFile := cachePathForURL(u)

	// Check cache first
	if data, err := os.ReadFile(cacheFile); err == nil {
		return string(data), nil
	}

	// Fetch from remote
	client := &http.Client{Timeout: remoteTimeout}
	resp, err := client.Get(remoteURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch remote schema: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("remote schema returned status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read remote schema: %w", err)
	}

	// Cache the result
	if err := os.MkdirAll(filepath.Dir(cacheFile), 0755); err == nil {
		_ = os.WriteFile(cacheFile, data, 0644)
	}

	return string(data), nil
}

func cachePathForURL(u *url.URL) string {
	host := strings.ReplaceAll(u.Host, ":", "_")
	path := strings.ReplaceAll(u.Path, "/", "_")
	name := fmt.Sprintf("%s%s_%s", host, path, u.Scheme)
	cacheDir := filepath.Join(os.TempDir(), remoteCacheDir)
	return filepath.Join(cacheDir, name)
}

// ClearRemoteCache removes all cached remote schemas.
func ClearRemoteCache() error {
	cacheDir := filepath.Join(os.TempDir(), remoteCacheDir)
	return os.RemoveAll(cacheDir)
}
