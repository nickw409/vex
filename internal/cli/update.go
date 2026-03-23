package cli

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/nickw409/vex/internal/log"
	"github.com/nickw409/vex/internal/version"
	"github.com/spf13/cobra"
)

const (
	ghRepo           = "nickw409/vex"
	latestReleaseURL = "https://api.github.com/repos/" + ghRepo + "/releases/latest"
	goModule         = "github.com/nickw409/vex/cmd/vex"
)

type githubRelease struct {
	TagName string `json:"tag_name"`
}

func newUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update vex to the latest version",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate()
		},
	}
}

func runUpdate() error {
	current := version.Short()

	log.Info("current version: %s", current)
	log.Info("checking for updates...")

	latest, err := fetchLatestVersion()
	if err != nil {
		return fmt.Errorf("checking latest version: %w", err)
	}

	if latest == current {
		log.Info("already up to date")
		return nil
	}

	log.Info("updating %s -> %s", current, latest)

	if version.InstalledFromGo() {
		return updateViaGo(latest)
	}
	return updateViaBinary(latest)
}

func updateViaGo(tag string) error {
	target := goModule + "@" + tag
	log.Info("running go install %s", target)

	cmd := exec.Command("go", "install", target)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go install failed: %w", err)
	}

	log.Info("updated to %s via go install", tag)
	return nil
}

func updateViaBinary(tag string) error {
	binary, err := downloadRelease(tag)
	if err != nil {
		return fmt.Errorf("downloading release: %w", err)
	}

	if err := replaceBinary(binary); err != nil {
		return fmt.Errorf("replacing binary: %w", err)
	}

	log.Info("updated to %s", tag)
	return nil
}

func fetchLatestVersion() (string, error) {
	resp, err := http.Get(latestReleaseURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}

	if release.TagName == "" {
		return "", fmt.Errorf("no tag_name in response")
	}

	return release.TagName, nil
}

func downloadRelease(tag string) ([]byte, error) {
	vnum := strings.TrimPrefix(tag, "v")
	filename := fmt.Sprintf("vex_%s_%s_%s.tar.gz", vnum, runtime.GOOS, runtime.GOARCH)
	url := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", ghRepo, tag, filename)

	log.Info("downloading %s", url)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download returned %d", resp.StatusCode)
	}

	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("decompressing: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading tar: %w", err)
		}

		if filepath.Base(hdr.Name) == "vex" && hdr.Typeflag == tar.TypeReg {
			data, err := io.ReadAll(tr)
			if err != nil {
				return nil, fmt.Errorf("reading binary from tar: %w", err)
			}
			return data, nil
		}
	}

	return nil, fmt.Errorf("vex binary not found in archive")
}

func replaceBinary(newBinary []byte) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("finding current executable: %w", err)
	}

	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return fmt.Errorf("resolving symlinks: %w", err)
	}

	// Write new binary to a temp file in the same directory so rename is atomic.
	dir := filepath.Dir(exe)
	tmp, err := os.CreateTemp(dir, "vex-update-*")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(newBinary); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("writing temp file: %w", err)
	}
	tmp.Close()

	if err := os.Chmod(tmpPath, 0755); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("setting permissions: %w", err)
	}

	if err := os.Rename(tmpPath, exe); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("replacing binary: %w", err)
	}

	return nil
}
