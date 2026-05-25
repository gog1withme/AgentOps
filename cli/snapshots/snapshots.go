package snapshots

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gog1withme/AgentOps/cli/internal/paths"
	"github.com/gog1withme/AgentOps/cli/store"
	"github.com/gog1withme/AgentOps/schema"
	"github.com/oklog/ulid/v2"
)

type Manager struct {
	sessionID string
	workDir   string
	store     *store.Store
	dir       string
	counter   int
	ticker    *time.Ticker
	stop      chan struct{}
}

type manifest struct {
	ID         string    `json:"id"`
	Timestamp  time.Time `json:"timestamp"`
	GitHead    string    `json:"git_head"`
	PatchFile  string    `json:"patch_file"`
	Untracked  []string  `json:"untracked"`
	Trigger    string    `json:"trigger"`
}

func InitSession(sessionID, workDir string, st *store.Store) (*Manager, error) {
	dir := filepath.Join(paths.SnapshotsDir(), sessionID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	m := &Manager{
		sessionID: sessionID,
		workDir:   workDir,
		store:     st,
		dir:       dir,
		stop:      make(chan struct{}),
	}
	if os.Getenv("AGENTOPS_NO_SNAPSHOTS") == "1" {
		return m, nil
	}
	if _, err := m.TakeSnapshot("initial"); err != nil {
		return m, err
	}
	m.ticker = time.NewTicker(60 * time.Second)
	go m.periodic()
	return m, nil
}

func (m *Manager) periodic() {
	for {
		select {
		case <-m.stop:
			return
		case <-m.ticker.C:
			_, _ = m.TakeSnapshot("periodic")
		}
	}
}

func (m *Manager) OnFileEdit() {
	if os.Getenv("AGENTOPS_NO_SNAPSHOTS") == "1" {
		return
	}
	_, _ = m.TakeSnapshot("file_edit")
}

func (m *Manager) TakeSnapshot(trigger string) (string, error) {
	m.counter++
	id := fmt.Sprintf("snapshot_%03d_%s", m.counter, ulid.Make().String()[:8])
	snapDir := filepath.Join(m.dir, id)
	if err := os.MkdirAll(snapDir, 0o755); err != nil {
		return "", err
	}
	head := gitOutput(m.workDir, "rev-parse", "HEAD")
	patchFile := filepath.Join(snapDir, "changes.patch")
	diff := gitOutput(m.workDir, "diff", "HEAD")
	if err := os.WriteFile(patchFile, []byte(diff), 0o644); err != nil {
		return "", err
	}
	untracked := gitOutput(m.workDir, "ls-files", "--others", "--exclude-standard")
	var copied []string
	var size int64
	for _, f := range strings.Split(untracked, "\n") {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		src := filepath.Join(m.workDir, f)
		dst := filepath.Join(snapDir, "untracked", f)
		if err := copyFile(src, dst); err == nil {
			copied = append(copied, f)
			if st, err := os.Stat(dst); err == nil {
				size += st.Size()
			}
		}
	}
	if st, err := os.Stat(patchFile); err == nil {
		size += st.Size()
	}
	man := manifest{
		ID:        id,
		Timestamp: time.Now(),
		GitHead:   strings.TrimSpace(head),
		PatchFile: "changes.patch",
		Untracked: copied,
		Trigger:   trigger,
	}
	meta, _ := json.MarshalIndent(man, "", "  ")
	_ = os.WriteFile(filepath.Join(snapDir, "manifest.json"), meta, 0o644)
	snap := &schema.Snapshot{
		ID:           id,
		SessionID:    m.sessionID,
		Timestamp:    man.Timestamp,
		Trigger:      trigger,
		SnapshotPath: snapDir,
		FileCount:    len(copied) + 1,
		SizeBytes:    size,
	}
	_ = m.store.WriteSnapshot(snap)
	return id, nil
}

func (m *Manager) Close() {
	if m.ticker != nil {
		m.ticker.Stop()
	}
	close(m.stop)
}

func Restore(st *store.Store, sessionID, snapshotID, workDir string, dryRun bool) error {
	snaps, err := st.ListSnapshots(sessionID)
	if err != nil {
		return err
	}
	var target *schema.Snapshot
	for i := range snaps {
		if snaps[i].ID == snapshotID || strings.HasPrefix(snaps[i].ID, snapshotID) {
			target = &snaps[i]
			break
		}
	}
	if target == nil {
		return fmt.Errorf("snapshot not found: %s", snapshotID)
	}
	metaPath := filepath.Join(target.SnapshotPath, "manifest.json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return err
	}
	var man manifest
	if err := json.Unmarshal(data, &man); err != nil {
		return err
	}
	if dryRun {
		fmt.Printf("Would restore session %s to snapshot %s (%s)\n", sessionID, target.ID, man.Timestamp.Format(time.RFC3339))
		return nil
	}
	_ = gitRun(workDir, "checkout", "HEAD", "--", ".")
	patch := filepath.Join(target.SnapshotPath, man.PatchFile)
	_ = gitRun(workDir, "apply", patch)
	for _, f := range man.Untracked {
		src := filepath.Join(target.SnapshotPath, "untracked", f)
		dst := filepath.Join(workDir, f)
		_ = os.MkdirAll(filepath.Dir(dst), 0o755)
		_ = copyFile(src, dst)
	}
	fmt.Printf("✓ Restored to %s. Snapshot %s applied.\n", man.Timestamp.Format(time.RFC3339), target.ID)
	return nil
}

func RestoreAt(st *store.Store, sessionID string, ts time.Time, workDir string, dryRun bool) error {
	snaps, err := st.ListSnapshots(sessionID)
	if err != nil {
		return err
	}
	var target *schema.Snapshot
	for i := range snaps {
		if snaps[i].Timestamp.Before(ts) || snaps[i].Timestamp.Equal(ts) {
			target = &snaps[i]
			break
		}
	}
	if target == nil && len(snaps) > 0 {
		target = &snaps[len(snaps)-1]
	}
	if target == nil {
		return fmt.Errorf("no snapshot before %s", ts.Format(time.RFC3339))
	}
	return Restore(st, sessionID, target.ID, workDir, dryRun)
}

func gitOutput(dir string, args ...string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, _ := cmd.Output()
	return string(out)
}

func gitRun(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
