package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func call(s *Server, method, path, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.api(w, r)
	return w
}

const passBody = `{"status":"passed","score":80,"evidence":"reports/calculation.md","practicalConfirmed":true}`

func TestAssessmentRulesAndStatuses(t *testing.T) {
	s := testServer(t)
	for _, tc := range []struct {
		body string
		want int
	}{
		{`{"status":"passed","score":100}`, 400},
		{`{"status":"passed","score":100,"evidence":"report"}`, 400},
		{`{"status":"passed","score":100,"practicalConfirmed":true,"evidence":"  "}`, 400},
		{`{"status":"passed","score":79.9,"evidence":"report","practicalConfirmed":true}`, 400},
		{`{"status":"in_progress","score":101}`, 400},
		{`{"status":"invented"}`, 400},
		{`{"status":"not_started"}`, 200},
		{`{"status":"in_progress"}`, 200},
		{`{"status":"submitted"}`, 200},
		{`{"status":"ready_for_assessment"}`, 200},
		{`{"status":"needs_revision"}`, 200},
		{passBody, 200},
	} {
		r := call(s, "PUT", "/api/progress/A", tc.body)
		if r.Code != tc.want {
			t.Errorf("%s: want %d got %d %s", tc.body, tc.want, r.Code, r.Body.String())
		}
	}
	completed := s.state.Progress["A"].CompletedAt
	if completed == "" {
		t.Fatal("completion date missing")
	}
	call(s, "PUT", "/api/progress/A", passBody)
	if s.state.Progress["A"].CompletedAt != completed {
		t.Fatal("editing pass changed completion date")
	}
	if r := call(s, "PUT", "/api/progress/B", passBody); r.Code != 200 {
		t.Fatal(r.Body.String())
	}
	if r := call(s, "PUT", "/api/progress/A", `{"status":"needs_revision"}`); r.Code != 409 {
		t.Fatal("allowed inconsistent passed dependency graph")
	}
	call(s, "PUT", "/api/progress/B", `{"status":"needs_revision"}`)
	if r := call(s, "PUT", "/api/progress/A", `{"status":"needs_revision"}`); r.Code != 200 {
		t.Fatal(r.Body.String())
	}
	if s.state.Progress["A"].CompletedAt != "" {
		t.Fatal("revision retains active completion")
	}
}

func TestRejectMalformedAndCrossOriginRequests(t *testing.T) {
	s := testServer(t)
	for _, body := range []string{
		`{"status":"in_progress","extra":1}`, `{"status":"in_progress"} {}`,
		`{"status":`, `{"status":"passed","score":NaN}`, `{"status":"in_progress","notes":"` + strings.Repeat("x", 65536) + `"}`,
		`{"status":"in_progress","notes":"` + strings.Repeat("x", 5001) + `"}`,
	} {
		if r := call(s, "PUT", "/api/progress/A", body); r.Code != 400 {
			t.Errorf("wanted 400, got %d", r.Code)
		}
	}
	for _, tc := range []struct {
		origin, site, content string
		want                  int
	}{
		{"https://evil.example", "", "application/json", 403},
		{"null", "", "application/json", 403},
		{"http://example.com", "cross-site", "application/json", 403},
		{"http://example.com", "same-origin", "text/plain", 400},
		{"http://example.com", "same-origin", "application/json", 200},
	} {
		r := httptest.NewRequest("PUT", "http://example.com/api/progress/A", strings.NewReader(`{"status":"in_progress"}`))
		r.Header.Set("Origin", tc.origin)
		r.Header.Set("Sec-Fetch-Site", tc.site)
		r.Header.Set("Content-Type", tc.content)
		w := httptest.NewRecorder()
		s.handler().ServeHTTP(w, r)
		if w.Code != tc.want {
			t.Errorf("%+v got %d %s", tc, w.Code, w.Body.String())
		}
	}
	if r := call(s, "PUT", "/api/progress/unknown", passBody); r.Code != 404 {
		t.Fatal("unknown module accepted")
	}
}

func TestFailedWriteDoesNotChangeState(t *testing.T) {
	s := testServer(t)
	// A regular file cannot serve as a parent directory; reliable even when run as root.
	blocker := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(blocker, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	s.path = filepath.Join(blocker, "progress.json")
	if r := call(s, "PUT", "/api/progress/A", passBody); r.Code != 500 {
		t.Fatal(r.Code)
	}
	if len(s.state.Progress) != 0 {
		t.Fatal("failed write inserted phantom progress")
	}
	if r := call(s, "POST", "/api/sessions", `{"moduleId":"A","minutes":30}`); r.Code != 500 {
		t.Fatal(r.Code)
	}
	if len(s.state.Sessions) != 0 {
		t.Fatal("failed session write changed memory")
	}
}

func TestRestartAndExport(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "progress.json")
	s, err := newServer("curriculum/helicopter.json", path)
	if err != nil {
		t.Fatal(err)
	}
	if r := call(s, "PUT", "/api/progress/S1M1", passBody); r.Code != 200 {
		t.Fatal(r.Body.String())
	}
	if r := call(s, "POST", "/api/sessions", `{"moduleId":"S1M1","minutes":45,"date":"2026-10-30","note":"<script>literal notes</script>"}`); r.Code != 201 {
		t.Fatal(r.Body.String())
	}
	reloaded, err := newServer("curriculum/helicopter.json", path)
	if err != nil {
		t.Fatal(err)
	}
	if len(reloaded.state.Sessions) != 1 || reloaded.state.Progress["S1M1"].Status != "passed" || reloaded.state.Progress["S1M1"].CompletedAt == "" {
		t.Fatal("restart lost assessment or session")
	}
	r := call(reloaded, "GET", "/api/export", "")
	if r.Code != 200 || !strings.Contains(r.Header().Get("Content-Disposition"), "attachment;") {
		t.Fatal("export is not downloadable")
	}
	var exported State
	if err := json.Unmarshal(r.Body.Bytes(), &exported); err != nil || len(exported.Sessions) != 1 {
		t.Fatal("invalid export", err)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatal("progress file permissions", err)
	}
	files, _ := filepath.Glob(filepath.Join(filepath.Dir(path), ".progress-*.tmp"))
	if len(files) != 0 {
		t.Fatal("temporary files leaked")
	}
}

func TestLegacyProgressAndCorruption(t *testing.T) {
	path := filepath.Join(t.TempDir(), "progress.json")
	legacy := `{"progress":{"S1M1":{"status":"passed","score":95,"notes":"prior report","updatedAt":"2026-10-29T01:02:03Z"},"S1M2":{"status":"submitted","notes":"ready"}},"sessions":null}`
	os.WriteFile(path, []byte(legacy), 0600)
	s, err := newServer("curriculum/helicopter.json", path)
	if err != nil {
		t.Fatal(err)
	}
	if s.state.Progress["S1M1"].Status != "needs_revision" || s.state.Progress["S1M1"].Notes != "prior report" || s.state.Progress["S1M2"].Status != "ready_for_assessment" {
		t.Fatal("legacy migration lost work")
	}
	original, _ := os.ReadFile(path)
	if string(original) != legacy {
		t.Fatal("startup silently rewrote original")
	}
	for _, bad := range []string{`null`, `{`, `{} {}`, `{"progress":{"unknown":{"status":"passed"}}}`, `{"progress":{"S1M1":{"status":"bogus"}}}`, `{"sessions":[{"id":"x","moduleId":"S1M1","minutes":0,"date":"2026-10-30"}]}`, `{"unknown":true}`} {
		os.WriteFile(path, []byte(bad), 0600)
		if _, err := newServer("curriculum/helicopter.json", path); err == nil {
			t.Errorf("accepted invalid state %s", bad)
		}
		original, _ = os.ReadFile(path)
		if string(original) != bad {
			t.Fatal("corrupt state overwritten")
		}
	}
}

func TestConcurrentStudyWrites(t *testing.T) {
	s, err := newServer("curriculum/helicopter.json", filepath.Join(t.TempDir(), "progress.json"))
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			r := call(s, "POST", "/api/sessions", fmt.Sprintf(`{"moduleId":"S1M1","minutes":%d}`, i+1))
			if r.Code != 201 {
				t.Errorf("write: %d", r.Code)
			}
			call(s, "GET", "/api/state", "")
		}(i)
	}
	wg.Wait()
	loaded, err := newServer("curriculum/helicopter.json", s.path)
	if err != nil || len(loaded.state.Sessions) != 20 {
		t.Fatal("lost concurrent writes", err)
	}
}

func TestCurriculumIntegrity(t *testing.T) {
	load := func() Curriculum {
		b, err := os.ReadFile("curriculum/helicopter.json")
		if err != nil {
			t.Fatal(err)
		}
		var c Curriculum
		if err = json.Unmarshal(b, &c); err != nil {
			t.Fatal(err)
		}
		return c
	}
	c := load()
	modules, err := c.validate()
	if err != nil || len(modules) != 42 || len(c.Stages) != 8 {
		t.Fatal("baseline content changed unexpectedly", err)
	}
	hours := 0
	for _, stage := range c.Stages {
		hours += stage.Hours
	}
	if hours != 3200 {
		t.Fatal("unexpected workload")
	}
	for name, mutate := range map[string]func(*Curriculum){
		"duplicate module":     func(c *Curriculum) { c.Stages[0].Modules[1].ID = c.Stages[0].Modules[0].ID },
		"unknown prerequisite": func(c *Curriculum) { c.Stages[0].Modules[0].Prerequisites = []string{"absent"} },
		"cycle":                func(c *Curriculum) { c.Stages[0].Modules[0].Prerequisites = []string{"S1M2"} },
		"incorrect hours":      func(c *Curriculum) { c.Stages[0].Hours++ },
		"untracked references": func(c *Curriculum) { c.Stages[0].Modules[0].ResourcesVerification = "" },
	} {
		t.Run(name, func(t *testing.T) {
			c := load()
			mutate(&c)
			if _, err := c.validate(); err == nil {
				t.Fatal("invalid curriculum accepted")
			}
		})
	}
}

func TestStaticDashboardAndCurriculumAccess(t *testing.T) {
	s, err := newServer("curriculum/helicopter.json", filepath.Join(t.TempDir(), "progress.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/", "/app.js", "/style.css", "/api/curriculum", "/api/state"} {
		r := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		s.handler().ServeHTTP(w, r)
		if w.Code != 200 || w.Header().Get("X-Content-Type-Options") != "nosniff" {
			t.Errorf("%s: %d", path, w.Code)
		}
	}
	r := httptest.NewRequest("GET", "/curriculum/helicopter.json", nil)
	w := httptest.NewRecorder()
	s.handler().ServeHTTP(w, r)
	if w.Code != 404 {
		t.Fatal("filesystem outside static exposed")
	}
}
