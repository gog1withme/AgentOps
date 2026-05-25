package collector

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gog1withme/AgentOps/cli/internal/platform"
	"github.com/gog1withme/AgentOps/cli/store"
	"github.com/gog1withme/AgentOps/schema"
	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
)

type FSWatcher struct {
	collector *Collector
	watcher   *fsnotify.Watcher
	workDir   string
	contents  map[string]string
}

func NewFSWatcher(c *Collector, workDir string) (*FSWatcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	fw := &FSWatcher{
		collector: c,
		watcher:   w,
		workDir:   workDir,
		contents:  make(map[string]string),
	}
	return fw, nil
}

func (fw *FSWatcher) Start() error {
	if err := fw.watcher.Add(fw.workDir); err != nil {
		return err
	}
	go fw.loop()
	return nil
}

func (fw *FSWatcher) loop() {
	for {
		select {
		case ev, ok := <-fw.watcher.Events:
			if !ok {
				return
			}
			if ev.Op&(fsnotify.Write|fsnotify.Create) != 0 {
				fw.handleFile(ev.Name)
			}
		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			log.Warn().Err(err).Msg("fs watcher error")
		}
	}
}

func (fw *FSWatcher) handleFile(path string) {
	if strings.Contains(path, ".git") || strings.Contains(path, "node_modules") {
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	newContent := string(data)
	oldContent := fw.contents[path]
	fw.contents[path] = newContent
	diff := computeSimpleDiff(oldContent, newContent)
	if diff == "" && oldContent != "" {
		return
	}
	agent := platform.AttributeFileAgent(path)
	rel, _ := filepath.Rel(fw.workDir, path)
	if rel == "." {
		rel = path
	}
	added, removed := countDiffLines(diff)
	eventID := store.NewEventID()
	now := time.Now()
	fw.collector.Ingest(schema.Event{
		ID:        eventID,
		SessionID: fw.collector.cfg.SessionID,
		Timestamp: now,
		Source:    agent,
		Type:      schema.EventFileEdit,
		FilePath:  rel,
		FileDiff:  diff,
	})
	_ = fw.collector.store.WriteAttribution(&schema.Attribution{
		ID:           store.NewEventID(),
		SessionID:    fw.collector.cfg.SessionID,
		Timestamp:    now,
		FilePath:     rel,
		Agent:        agent,
		LinesAdded:   added,
		LinesRemoved: removed,
		Diff:         diff,
		EventID:      eventID,
	})
}

func computeSimpleDiff(old, new string) string {
	if old == new {
		return ""
	}
	if old == "" {
		return "+ " + strings.ReplaceAll(new, "\n", "\n+ ")
	}
	return "--- previous\n+++ current\n" + new
}

func countDiffLines(diff string) (added, removed int) {
	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			added++
		}
		if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			removed++
		}
	}
	return
}

func (fw *FSWatcher) Close() error {
	return fw.watcher.Close()
}
