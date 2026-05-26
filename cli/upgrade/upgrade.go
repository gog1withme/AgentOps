package upgrade

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gog1withme/AgentOps/cli/internal/paths"
	"github.com/gog1withme/AgentOps/cli/version"
)

const repo = "gog1withme/AgentOps"

type ReleaseInfo struct {
	CurrentVersion string
	LatestVersion  string
	UpdateAvailable bool
}

func Check() (*ReleaseInfo, error) {
	latest, err := fetchLatestVersion()
	if err != nil {
		return nil, err
	}
	current := strings.TrimPrefix(version.Version, "v")
	latest = strings.TrimPrefix(latest, "v")
	return &ReleaseInfo{
		CurrentVersion:  current,
		LatestVersion:   latest,
		UpdateAvailable: compareVersions(latest, current) > 0,
	}, nil
}

func Upgrade(force bool) error {
	info, err := Check()
	if err != nil {
		return err
	}
	if !info.UpdateAvailable && !force {
		return fmt.Errorf("already on latest version %s", info.CurrentVersion)
	}
	target := info.LatestVersion
	if force && !info.UpdateAvailable {
		target = info.CurrentVersion
	}

	platform, err := detectPlatform()
	if err != nil {
		return err
	}

	archive, binaryName, extractDir, err := downloadRelease(target, platform)
	if err != nil {
		return err
	}
	defer os.RemoveAll(extractDir)

	if err := verifyChecksum(extractDir, archive, target); err != nil {
		return err
	}

	exe, err := os.Executable()
	if err != nil {
		return err
	}
	srcBinary := filepath.Join(extractDir, binaryName)
	if _, err := os.Stat(srcBinary); err != nil {
		return fmt.Errorf("binary not found in release archive: %s", srcBinary)
	}

	if err := replaceBinary(exe, srcBinary); err != nil {
		return err
	}

	dashSrc := filepath.Join(extractDir, "dashboard", "out")
	if _, err := os.Stat(filepath.Join(dashSrc, "index.html")); err == nil {
		dashDest := filepath.Join(paths.AgentOpsDir(), "dashboard", "out")
		if err := os.RemoveAll(dashDest); err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(dashDest), 0o755); err != nil {
			return err
		}
		if err := copyDir(dashSrc, dashDest); err != nil {
			return err
		}
	}

	return nil
}

func fetchLatestVersion() (string, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/" + repo + "/releases/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github API returned %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	// lightweight parse without adding dependency
	const marker = `"tag_name":`
	idx := strings.Index(string(body), marker)
	if idx < 0 {
		return "", fmt.Errorf("tag_name not found in release response")
	}
	rest := string(body[idx+len(marker):])
	start := strings.Index(rest, `"`)
	if start < 0 {
		return "", fmt.Errorf("invalid tag_name format")
	}
	rest = rest[start+1:]
	end := strings.Index(rest, `"`)
	if end < 0 {
		return "", fmt.Errorf("invalid tag_name format")
	}
	return strings.TrimPrefix(rest[:end], "v"), nil
}

func detectPlatform() (string, error) {
	osName := runtime.GOOS
	arch := runtime.GOARCH
	switch osName {
	case "linux", "darwin", "windows":
	default:
		return "", fmt.Errorf("unsupported OS: %s", osName)
	}
	if osName == "windows" && arch == "arm64" {
		return "", fmt.Errorf("windows arm64 is not supported")
	}
	return osName + "_" + arch, nil
}

func downloadRelease(version, platform string) (archiveName, binaryName, extractDir string, err error) {
	osName := strings.Split(platform, "_")[0]
	arch := strings.Split(platform, "_")[1]
	binaryName = "agentops"
	if osName == "windows" {
		binaryName = "agentops.exe"
		archiveName = fmt.Sprintf("agentops_%s_%s_%s.zip", version, osName, arch)
	} else {
		archiveName = fmt.Sprintf("agentops_%s_%s_%s.tar.gz", version, osName, arch)
	}

	url := fmt.Sprintf("https://github.com/%s/releases/download/v%s/%s", repo, version, archiveName)
	extractDir, err = os.MkdirTemp("", "agentops-upgrade-*")
	if err != nil {
		return "", "", "", err
	}

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", "", "", fmt.Errorf("download failed: %s", resp.Status)
	}

	archivePath := filepath.Join(extractDir, archiveName)
	f, err := os.Create(archivePath)
	if err != nil {
		return "", "", "", err
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		return "", "", "", err
	}
	f.Close()

	if osName == "windows" {
		if err := unzip(archivePath, extractDir); err != nil {
			return "", "", "", err
		}
	} else {
		if err := untarGz(archivePath, extractDir); err != nil {
			return "", "", "", err
		}
	}
	return archiveName, binaryName, extractDir, nil
}

func verifyChecksum(extractDir, archiveName, version string) error {
	url := fmt.Sprintf("https://github.com/%s/releases/download/v%s/checksums.txt", repo, version)
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var expected string
	for _, line := range strings.Split(string(body), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasSuffix(line, " "+archiveName) {
			expected = strings.Fields(line)[0]
			break
		}
	}
	if expected == "" {
		return fmt.Errorf("checksum entry not found for %s", archiveName)
	}

	archivePath := filepath.Join(extractDir, archiveName)
	data, err := os.ReadFile(archivePath)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(data)
	actual := hex.EncodeToString(sum[:])
	if !strings.EqualFold(expected, actual) {
		return fmt.Errorf("checksum mismatch for %s", archiveName)
	}
	return nil
}

func replaceBinary(dest, src string) error {
	srcData, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.WriteFile(dest+".new", srcData, 0o755); err != nil {
		return err
	}
	if err := os.Rename(dest+".new", dest); err != nil {
		_ = os.Remove(dest + ".new")
		return err
	}
	return nil
}

func copyDir(src, dest string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dest, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode())
	})
}

func compareVersions(a, b string) int {
	pa := strings.Split(a, ".")
	pb := strings.Split(b, ".")
	for len(pa) < 3 {
		pa = append(pa, "0")
	}
	for len(pb) < 3 {
		pb = append(pb, "0")
	}
	for i := 0; i < 3; i++ {
		ai := parseVersionPart(pa[i])
		bi := parseVersionPart(pb[i])
		if ai > bi {
			return 1
		}
		if ai < bi {
			return -1
		}
	}
	return 0
}

func parseVersionPart(s string) int {
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			break
		}
		n = n*10 + int(r-'0')
	}
	return n
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		target := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid zip entry: %s", f.Name)
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.Create(target)
		if err != nil {
			rc.Close()
			return err
		}
		_, err = io.Copy(out, rc)
		out.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
