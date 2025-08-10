package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/abdultolba/nizam/internal/version"
	"github.com/spf13/cobra"
)

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func NewUpdateCmd() *cobra.Command {
	var checkOnly bool
	var prerelease bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update nizam to the latest version",
		Long: `Update nizam to the latest version from GitHub releases.

This command downloads and installs the latest stable release of nizam.
It will replace the current binary with the new version.`,
		Example: `  # Check for updates without installing
  nizam update --check

  # Update to latest stable version
  nizam update

  # Include prerelease versions
  nizam update --prerelease`,
		RunE: func(cmd *cobra.Command, args []string) error {
			currentVersion := version.Version()
			if currentVersion == "" || currentVersion == "dev" {
				currentVersion = "unknown"
			}

			fmt.Printf("Current version: %s\n", currentVersion)

			latest, err := getLatestRelease(prerelease)
			if err != nil {
				return fmt.Errorf("failed to check for updates: %w", err)
			}

			fmt.Printf("Latest version: %s\n", latest.TagName)

			if currentVersion != "unknown" && currentVersion == latest.TagName {
				fmt.Println("✔ You are already running the latest version")
				return nil
			}

			if checkOnly {
				if currentVersion == "unknown" {
					fmt.Println("! Cannot determine if update is needed (unknown current version)")
				} else {
					fmt.Println("📦 Update available")
				}
				return nil
			}

			// Find appropriate asset for current platform
			assetName := fmt.Sprintf("nizam_%s_%s", runtime.GOOS, runtime.GOARCH)
			var downloadURL string

			for _, asset := range latest.Assets {
				if asset.Name == assetName || asset.Name == assetName+".exe" {
					downloadURL = asset.BrowserDownloadURL
					break
				}
			}

			if downloadURL == "" {
				return fmt.Errorf("no binary found for %s/%s", runtime.GOOS, runtime.GOARCH)
			}

			fmt.Printf("Downloading %s...\n", latest.TagName)

			// Get current executable path
			execPath, err := os.Executable()
			if err != nil {
				return fmt.Errorf("failed to get executable path: %w", err)
			}

			// Download new version
			if err := downloadAndReplace(downloadURL, execPath); err != nil {
				return fmt.Errorf("failed to update: %w", err)
			}

			fmt.Printf("✔ Successfully updated to %s\n", latest.TagName)
			return nil
		},
	}

	cmd.Flags().BoolVar(&checkOnly, "check", false, "only check for updates, don't install")
	cmd.Flags().BoolVar(&prerelease, "prerelease", false, "include prerelease versions")

	return cmd
}

func getLatestRelease(includePrerelease bool) (*GitHubRelease, error) {
	url := "https://api.github.com/repos/abdultolba/nizam/releases"
	if !includePrerelease {
		url += "/latest"
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	if includePrerelease {
		var releases []GitHubRelease
		if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
			return nil, err
		}
		if len(releases) == 0 {
			return nil, fmt.Errorf("no releases found")
		}
		return &releases[0], nil
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}
	return &release, nil
}

func downloadAndReplace(downloadURL, execPath string) error {
	// Create temporary file
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "nizam_update")
	if runtime.GOOS == "windows" {
		tmpFile += ".exe"
	}

	// Download to temporary file
	resp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	out, err := os.Create(tmpFile)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	// Make executable
	if err := os.Chmod(tmpFile, 0755); err != nil {
		return err
	}

	// Replace current binary
	if runtime.GOOS == "windows" {
		// On Windows, move old file and copy new one
		oldPath := execPath + ".old"
		if err := os.Rename(execPath, oldPath); err != nil {
			return err
		}
		if err := os.Rename(tmpFile, execPath); err != nil {
			// Try to restore old file
			os.Rename(oldPath, execPath)
			return err
		}
		os.Remove(oldPath)
	} else {
		// On Unix-like systems, can overwrite directly
		if err := os.Rename(tmpFile, execPath); err != nil {
			return err
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(NewUpdateCmd())
}
