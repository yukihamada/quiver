package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

type UpdateChecker struct {
	currentVersion string
	updateURL      string
	logger         *logrus.Logger
}

type Release struct {
	TagName    string `json:"tag_name"`
	Name       string `json:"name"`
	Body       string `json:"body"`
	PreRelease bool   `json:"prerelease"`
	Draft      bool   `json:"draft"`
	Assets     []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func NewUpdateChecker(currentVersion string, logger *logrus.Logger) *UpdateChecker {
	return &UpdateChecker{
		currentVersion: currentVersion,
		updateURL:      "https://api.github.com/repos/yukihamada/quiver/releases/latest",
		logger:         logger,
	}
}

func (u *UpdateChecker) CheckForUpdates(ctx context.Context) (*Release, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", u.updateURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release info: %w", err)
	}

	if release.Draft || release.PreRelease {
		return nil, nil // Skip draft/prerelease versions
	}

	if release.TagName <= u.currentVersion {
		return nil, nil // No update available
	}

	return &release, nil
}

func (u *UpdateChecker) DownloadAndInstall(ctx context.Context, release *Release) error {
	// Find the appropriate installer for the current platform
	var installerURL string
	installerName := ""

	for _, asset := range release.Assets {
		if runtime.GOOS == "darwin" && filepath.Ext(asset.Name) == ".pkg" {
			installerURL = asset.BrowserDownloadURL
			installerName = asset.Name
			break
		}
	}

	if installerURL == "" {
		return fmt.Errorf("no installer found for platform %s", runtime.GOOS)
	}

	u.logger.WithFields(logrus.Fields{
		"version":   release.TagName,
		"installer": installerName,
	}).Info("Downloading update")

	// Download to temporary directory
	tmpDir := os.TempDir()
	installerPath := filepath.Join(tmpDir, installerName)

	if err := u.downloadFile(ctx, installerURL, installerPath); err != nil {
		return fmt.Errorf("failed to download installer: %w", err)
	}

	u.logger.Info("Download complete, installing update")

	// Install the update
	if err := u.installUpdate(installerPath); err != nil {
		return fmt.Errorf("failed to install update: %w", err)
	}

	u.logger.Info("Update installed successfully")
	return nil
}

func (u *UpdateChecker) downloadFile(ctx context.Context, url, filepath string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func (u *UpdateChecker) installUpdate(installerPath string) error {
	switch runtime.GOOS {
	case "darwin":
		// Use installer command to install PKG
		cmd := exec.Command("sudo", "installer", "-pkg", installerPath, "-target", "/")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	default:
		return fmt.Errorf("auto-update not supported on %s", runtime.GOOS)
	}
}

func (u *UpdateChecker) StartAutoUpdateCheck(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			u.checkAndUpdate(ctx)
		}
	}
}

func (u *UpdateChecker) checkAndUpdate(ctx context.Context) {
	release, err := u.CheckForUpdates(ctx)
	if err != nil {
		u.logger.WithError(err).Warn("Failed to check for updates")
		return
	}

	if release == nil {
		u.logger.Debug("No updates available")
		return
	}

	u.logger.WithField("version", release.TagName).Info("Update available, downloading...")

	if err := u.DownloadAndInstall(ctx, release); err != nil {
		u.logger.WithError(err).Error("Failed to install update")
		return
	}

	u.logger.Info("Update installed, please restart the application")
}