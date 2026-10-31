package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testServer(t *testing.T) *Server {
	t.Helper()
	s := &Server{path: filepath.Join(t.TempDir(), "progress.json"), state: State{Progress: map[string]Progress{}, Sessions: []Session{}}, valid: map[string]Module{}}
	s.valid["A"] = Module{ID: "A"}
	s.valid["B"] = Module{ID: "B", Prerequisites: []string{"A"}}
	return s
}
func TestProgressPrerequisitesAndPersistence(t *testing.T) {
	s := testServer(t)
	request := func(id, body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPut, "/api/progress/"+id, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		s.api(w, r)
		return w
	}
	if r := request("B", `{"status":"passed","score":90,"notes":"ok","evidence":"calculations/report.md","practicalConfirmed":true}`); r.Code != 409 {
		t.Fatalf("expected prerequisite conflict, got %d", r.Code)
	}
	if r := request("A", `{"status":"passed","score":70,"notes":"ok","evidence":"calculations/report.md","practicalConfirmed":true}`); r.Code != 400 {
		t.Fatalf("expected low-score rejection, got %d", r.Code)
	}
	if r := request("A", `{"status":"passed","score":90,"notes":"work verified","evidence":"calculations/report.md","practicalConfirmed":true}`); r.Code != 200 {
		t.Fatalf("expected success, got %d %s", r.Code, r.Body.String())
	}
	if r := request("B", `{"status":"passed","score":85,"notes":"ok","evidence":"calculations/report.md","practicalConfirmed":true}`); r.Code != 200 {
		t.Fatalf("expected success, got %d %s", r.Code, r.Body.String())
	}
	content, err := os.ReadFile(s.path)
	if err != nil {
		t.Fatal(err)
	}
	var saved State
	if err = json.Unmarshal(content, &saved); err != nil {
		t.Fatal(err)
	}
	if saved.Progress["B"].Status != "passed" {
		t.Fatalf("progress not persisted: %#v", saved)
	}
}
func TestStudySessionValidation(t *testing.T) {
	s := testServer(t)
	for _, tc := range []struct {
		body string
		want int
	}{{`{"moduleId":"unknown","minutes":15}`, 400}, {`{"moduleId":"A","minutes":0}`, 400}, {`{"moduleId":"A","minutes":25,"date":"2026-02-30"}`, 400}, {`{"moduleId":"A","minutes":25,"date":"2026-10-30","note":"worked problems"}`, 201}} {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/sessions", strings.NewReader(tc.body))
		r.Header.Set("Content-Type", "application/json")
		s.api(w, r)
		if w.Code != tc.want {
			t.Errorf("for %s expected %d got %d %s", tc.body, tc.want, w.Code, w.Body.String())
		}
	}
	if len(s.state.Sessions) != 1 {
		t.Fatalf("expected one session, got %d", len(s.state.Sessions))
	}
}
