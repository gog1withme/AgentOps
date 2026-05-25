// Command generate-qa renders versioned Developer Q&A markdown from docs/qa/source/*.json.
//
// Usage:
//
//	go run scripts/generate-qa.go          # regenerate docs/qa/*.md
//	go run scripts/generate-qa.go --check  # fail if output differs from committed files
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	sourceDir = "docs/qa/source"
	outDir    = "docs/qa"
)

var releaseRE = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

type answer struct {
	Release string `json:"release"`
	Status  string `json:"status"`
	Date    string `json:"date"`
	Body    string `json:"body"`
}

type question struct {
	ID       string   `json:"id"`
	Slug     string   `json:"slug"`
	Title    string   `json:"title"`
	Category string   `json:"category"`
	Tags     []string `json:"tags"`
	Answers  []answer `json:"answers"`
}

func main() {
	check := flag.Bool("check", false, "verify generated markdown matches committed output")
	flag.Parse()

	root, err := findRepoRoot()
	if err != nil {
		fatal(err)
	}

	questions, err := loadQuestions(filepath.Join(root, sourceDir))
	if err != nil {
		fatal(err)
	}

	outputs, err := renderAll(root, questions)
	if err != nil {
		fatal(err)
	}

	if *check {
		for path, content := range outputs {
			existing, err := os.ReadFile(path)
			if err != nil {
				fatal(fmt.Errorf("%s: %w (run go run scripts/generate-qa.go)", path, err))
			}
			if string(existing) != content {
				fatal(fmt.Errorf("%s is out of date (run go run scripts/generate-qa.go)", path))
			}
		}
		fmt.Println("Q&A docs are up to date")
		return
	}

	for path, content := range outputs {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			fatal(err)
		}
	}
	fmt.Printf("Generated %d Q&A file(s)\n", len(outputs))
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find repo root (go.mod)")
		}
		dir = parent
	}
}

func loadQuestions(dir string) ([]question, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read source dir: %w", err)
	}

	var questions []question
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var q question
		if err := json.Unmarshal(data, &q); err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		if err := validateQuestion(q, path); err != nil {
			return nil, err
		}
		questions = append(questions, q)
	}
	if len(questions) == 0 {
		return nil, fmt.Errorf("no question files found in %s", dir)
	}

	sort.Slice(questions, func(i, j int) bool {
		if questions[i].Category != questions[j].Category {
			return questions[i].Category < questions[j].Category
		}
		return questions[i].Title < questions[j].Title
	})
	return questions, nil
}

func validateQuestion(q question, path string) error {
	prefix := path + ": "
	if q.ID == "" {
		return fmt.Errorf("%sid is required", prefix)
	}
	if q.Slug == "" {
		return fmt.Errorf("%sslug is required", prefix)
	}
	if q.ID != q.Slug {
		return fmt.Errorf("%sid and slug must match (%q != %q)", prefix, q.ID, q.Slug)
	}
	if q.Title == "" {
		return fmt.Errorf("%stitle is required", prefix)
	}
	if q.Category == "" {
		return fmt.Errorf("%scategory is required", prefix)
	}
	if len(q.Answers) == 0 {
		return fmt.Errorf("%sat least one answer is required", prefix)
	}

	current := 0
	for i, a := range q.Answers {
		if !releaseRE.MatchString(a.Release) {
			return fmt.Errorf("%sanswers[%d].release must be semver (e.g. 1.0.0)", prefix, i)
		}
		switch a.Status {
		case "current", "superseded":
		default:
			return fmt.Errorf("%sanswers[%d].status must be current or superseded", prefix, i)
		}
		if a.Status == "current" {
			current++
		}
		if strings.TrimSpace(a.Body) == "" {
			return fmt.Errorf("%sanswers[%d].body is required", prefix, i)
		}
	}
	if current != 1 {
		return fmt.Errorf("%sexactly one answer must have status current (found %d)", prefix, current)
	}
	return nil
}

func renderAll(root string, questions []question) (map[string]string, error) {
	outputs := make(map[string]string)

	index := renderIndex(questions)
	outputs[filepath.Join(root, outDir, "README.md")] = index

	for _, q := range questions {
		page := renderQuestion(q)
		path := filepath.Join(root, outDir, q.Category, q.Slug+".md")
		outputs[path] = page
	}
	return outputs, nil
}

func renderIndex(questions []question) string {
	byCategory := make(map[string][]question)
	categories := make([]string, 0)
	for _, q := range questions {
		if _, ok := byCategory[q.Category]; !ok {
			categories = append(categories, q.Category)
		}
		byCategory[q.Category] = append(byCategory[q.Category], q)
	}
	sort.Strings(categories)

	var b strings.Builder
	b.WriteString("# Developer Q&A\n\n")
	b.WriteString("Technical answers about how AgentOps works — tagged by release version. ")
	b.WriteString("When behavior changes, a new answer is added and older answers remain visible with strikethrough.\n\n")
	b.WriteString("See [MAINTAIN.md](MAINTAIN.md) for how to update Q&A on each release.\n\n")

	for _, cat := range categories {
		items := byCategory[cat]
		sort.Slice(items, func(i, j int) bool {
			return items[i].Title < items[j].Title
		})
		b.WriteString("## ")
		b.WriteString(titleCase(cat))
		b.WriteString("\n\n")
		for _, q := range items {
			current := currentAnswer(q)
			b.WriteString("- [")
			b.WriteString(q.Title)
			b.WriteString("](")
			b.WriteString(q.Category)
			b.WriteString("/")
			b.WriteString(q.Slug)
			b.WriteString(".md) — `v")
			b.WriteString(current.Release)
			b.WriteString("`\n")
		}
		b.WriteString("\n")
	}
	return b.String()
}

func renderQuestion(q question) string {
	answers := append([]answer(nil), q.Answers...)
	sort.Slice(answers, func(i, j int) bool {
		return compareRelease(answers[i].Release, answers[j].Release) > 0
	})

	var b strings.Builder
	b.WriteString("# ")
	b.WriteString(q.Title)
	b.WriteString("\n\n")
	b.WriteString("[← Developer Q&A](../README.md)\n\n")

	tagParts := make([]string, 0, len(q.Tags)+1)
	tagParts = append(tagParts, "`"+q.Category+"`")
	for _, t := range q.Tags {
		tagParts = append(tagParts, "`"+t+"`")
	}
	b.WriteString(strings.Join(tagParts, " · "))
	b.WriteString("\n\n")

	for i, a := range answers {
		if i > 0 {
			b.WriteString("\n---\n\n")
		}
		if a.Status == "current" {
			b.WriteString("**Current answer — v")
		} else {
			b.WriteString("**Superseded answer — v")
		}
		b.WriteString(a.Release)
		b.WriteString("**")
		if a.Date != "" {
			b.WriteString(" · ")
			b.WriteString(a.Date)
		}
		b.WriteString("\n\n")

		if a.Status == "current" {
			b.WriteString(strings.TrimSpace(a.Body))
			b.WriteString("\n")
		} else {
			b.WriteString(strikethroughBody(a.Body))
		}
	}
	return b.String()
}

func strikethroughBody(body string) string {
	body = strings.TrimSpace(body)
	paragraphs := strings.Split(body, "\n\n")
	var out []string
	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		lines := strings.Split(p, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			out = append(out, "~~"+line+"~~")
		}
		out = append(out, "")
	}
	return strings.TrimRight(strings.Join(out, "\n"), "\n") + "\n"
}

func currentAnswer(q question) answer {
	for _, a := range q.Answers {
		if a.Status == "current" {
			return a
		}
	}
	return q.Answers[0]
}

func compareRelease(a, b string) int {
	ap := parseRelease(a)
	bp := parseRelease(b)
	for i := 0; i < 3; i++ {
		if ap[i] != bp[i] {
			return ap[i] - bp[i]
		}
	}
	return 0
}

func parseRelease(s string) [3]int {
	var out [3]int
	parts := strings.Split(s, ".")
	for i := 0; i < len(parts) && i < 3; i++ {
		fmt.Sscanf(parts[i], "%d", &out[i])
	}
	return out
}

func titleCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + strings.ReplaceAll(s[1:], "-", " ")
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "generate-qa: %v\n", err)
	os.Exit(1)
}
