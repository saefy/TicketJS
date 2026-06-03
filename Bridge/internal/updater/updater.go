// Package updater checks for new thermalik-bridge releases on GitHub
// and applies them via atomic binary replacement.
package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/blang/semver/v4"
	"github.com/minio/selfupdate"

	"thermalik-bridge/internal/version"
)

const (
	githubReleaseAPI = "https://api.github.com/repos/saefy/thermalik/releases/latest"
	httpTimeout      = 10 * time.Second
)

// UpdateInfo holds the result of a release check.
type UpdateInfo struct {
	Available bool
	Version   string
	AssetURL  string
	Checksum  string // hex-encoded SHA256
}

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// CheckLatest queries the GitHub Releases API and returns update info.
// Returns UpdateInfo{Available: false} when already on the latest version.
func CheckLatest() (UpdateInfo, error) {
	client := &http.Client{Timeout: httpTimeout}

	resp, err := client.Get(githubReleaseAPI)
	if err != nil {
		return UpdateInfo{}, fmt.Errorf("check latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return UpdateInfo{}, fmt.Errorf("github api returned %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return UpdateInfo{}, fmt.Errorf("decode release: %w", err)
	}

	remoteVersion, err := parseVersion(release.TagName)
	if err != nil {
		return UpdateInfo{}, fmt.Errorf("parse remote version %q: %w", release.TagName, err)
	}

	current := version.Version
	if current == "dev" {
		// Skip update checks for local dev builds.
		return UpdateInfo{}, nil
	}

	currentVersion, err := parseVersion(current)
	if err != nil {
		return UpdateInfo{}, fmt.Errorf("parse current version %q: %w", current, err)
	}

	if !remoteVersion.GT(currentVersion) {
		return UpdateInfo{Available: false}, nil
	}

	assetName := assetNameForPlatform()
	checksumName := "checksums.txt"

	var assetURL, checksumURL string
	for _, asset := range release.Assets {
		switch asset.Name {
		case assetName:
			assetURL = asset.BrowserDownloadURL
		case checksumName:
			checksumURL = asset.BrowserDownloadURL
		}
	}

	if assetURL == "" {
		return UpdateInfo{}, fmt.Errorf("no release asset found for %s", assetName)
	}

	checksum, err := fetchChecksum(client, checksumURL, assetName)
	if err != nil {
		return UpdateInfo{}, fmt.Errorf("fetch checksum: %w", err)
	}

	return UpdateInfo{
		Available: true,
		Version:   release.TagName,
		AssetURL:  assetURL,
		Checksum:  checksum,
	}, nil
}

// Apply downloads the release asset, verifies its SHA256 checksum,
// and atomically replaces the running binary.
func Apply(info UpdateInfo) error {
	if !info.Available {
		return fmt.Errorf("no update available")
	}

	client := &http.Client{Timeout: 5 * time.Minute}

	resp, err := client.Get(info.AssetURL)
	if err != nil {
		return fmt.Errorf("download release: %w", err)
	}
	defer resp.Body.Close()

	// Buffer the download so we can verify the checksum before applying.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read release body: %w", err)
	}

	if err := verifyChecksum(body, info.Checksum); err != nil {
		return err
	}

	if err := selfupdate.Apply(strings.NewReader(string(body)), selfupdate.Options{}); err != nil {
		if rerr := selfupdate.RollbackError(err); rerr != nil {
			return fmt.Errorf("apply update failed and rollback failed: %w", rerr)
		}

		return fmt.Errorf("apply update (rolled back): %w", err)
	}

	return nil
}

func parseVersion(tag string) (semver.Version, error) {
	return semver.ParseTolerant(strings.TrimPrefix(tag, "v"))
}

func assetNameForPlatform() string {
	os := runtime.GOOS
	arch := runtime.GOARCH

	ext := ".tar.gz"
	if os == "windows" {
		ext = ".zip"
	}

	return fmt.Sprintf("thermalik-bridge_%s_%s%s", os, arch, ext)
}

func fetchChecksum(client *http.Client, checksumURL, assetName string) (string, error) {
	if checksumURL == "" {
		return "", nil
	}

	resp, err := client.Get(checksumURL)
	if err != nil {
		return "", fmt.Errorf("download checksums.txt: %w", err)
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read checksums.txt: %w", err)
	}

	for _, line := range strings.Split(string(content), "\n") {
		parts := strings.Fields(line)
		if len(parts) == 2 && parts[1] == assetName {
			return parts[0], nil
		}
	}

	return "", fmt.Errorf("checksum not found for %s", assetName)
}

func verifyChecksum(data []byte, expected string) error {
	if expected == "" {
		return nil
	}

	hash := sha256.Sum256(data)
	got := hex.EncodeToString(hash[:])

	if got != expected {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expected, got)
	}

	return nil
}
