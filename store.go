package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Progress struct {
	Status             string   `json:"status"`
	Score              *float64 `json:"score,omitempty"`
	Notes              string   `json:"notes"`
	Evidence           string   `json:"evidence"`
	PracticalConfirmed bool     `json:"practicalConfirmed"`
	UpdatedAt          string   `json:"updatedAt"`
	CompletedAt        string   `json:"completedAt,omitempty"`
}
type Session struct {
	ID       string `json:"id"`
	ModuleID string `json:"moduleId"`
	Minutes  int    `json:"minutes"`
	Note     string `json:"note"`
	Date     string `json:"date"`
}
type State struct {
	Progress map[string]Progress `json:"progress"`
	Sessions []Session           `json:"sessions"`
}
type Server struct {
	mu         sync.Mutex
	curriculum Curriculum
	state      State
	path       string
	valid      map[string]Module
}

func validStatus(status string) bool {
	switch status {
	case "not_started", "in_progress", "ready_for_assessment", "passed", "needs_revision":
		return true
	}
	return false
}
func passingEvidence(p Progress) bool {
	return p.Score != nil && *p.Score >= 80 && *p.Score <= 100 && p.PracticalConfirmed && strings.TrimSpace(p.Evidence) != ""
}
func validateDate(date string) bool {
	d, err := time.Parse("2006-01-02", date)
	return err == nil && d.Format("2006-01-02") == date
}

func newServer(curriculumPath, dataPath string) (*Server, error) {
	content, err := os.ReadFile(curriculumPath)
	if err != nil {
		return nil, err
	}
	var cur Curriculum
	if err = json.Unmarshal(content, &cur); err != nil {
		return nil, fmt.Errorf("invalid curriculum: %w", err)
	}
	modules, err := cur.validate()
	if err != nil {
		return nil, err
	}
	s := &Server{curriculum: cur, path: dataPath, valid: modules, state: State{Progress: map[string]Progress{}, Sessions: []Session{}}}
	content, err = os.ReadFile(dataPath)
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(bytes.NewReader(content))
	dec.DisallowUnknownFields()
	var saved *State
	if err = dec.Decode(&saved); err != nil {
		return nil, fmt.Errorf("invalid progress data (original preserved): %w", err)
	}
	if saved == nil {
		return nil, fmt.Errorf("invalid progress data: expected an object")
	}
	s.state = *saved
	if err = dec.Decode(new(any)); err != io.EOF {
		return nil, fmt.Errorf("invalid progress data: trailing JSON")
	}
	if s.state.Progress == nil {
		s.state.Progress = map[string]Progress{}
	}
	if s.state.Sessions == nil {
		s.state.Sessions = []Session{}
	}
	for id, p := range s.state.Progress {
		if p.Status == "submitted" {
			p.Status = "ready_for_assessment"
		}
		if _, ok := modules[id]; !ok {
			return nil, fmt.Errorf("saved progress has unknown module %s", id)
		}
		if !validStatus(p.Status) || len(p.Notes) > 5000 || len(p.Evidence) > 5000 || (p.Score != nil && (*p.Score < 0 || *p.Score > 100)) {
			return nil, fmt.Errorf("invalid saved progress for %s", id)
		}
		for _, stamp := range []string{p.UpdatedAt, p.CompletedAt} {
			if stamp != "" {
				if _, err := time.Parse(time.RFC3339, stamp); err != nil {
					return nil, fmt.Errorf("invalid saved timestamp for %s", id)
				}
			}
		}
		// Old passes did not record practical review. Preserve their work for re-review.
		if p.Status == "passed" && (!passingEvidence(p) || p.CompletedAt == "") {
			p.Status = "needs_revision"
		}
		if p.Status != "passed" {
			p.CompletedAt = ""
		}
		s.state.Progress[id] = p
	}
	// Reconcile descendants of legacy passes without inventing assessment evidence.
	for changed := true; changed; {
		changed = false
		for id, p := range s.state.Progress {
			if p.Status != "passed" {
				continue
			}
			for _, dep := range modules[id].Prerequisites {
				if s.state.Progress[dep].Status != "passed" {
					p.Status, p.CompletedAt = "needs_revision", ""
					s.state.Progress[id] = p
					changed = true
					break
				}
			}
		}
	}
	seen := map[string]bool{}
	for _, session := range s.state.Sessions {
		_, exists := modules[session.ModuleID]
		if !exists || session.ID == "" || seen[session.ID] || session.Minutes < 1 || session.Minutes > 1440 || len(session.Note) > 2000 || !validateDate(session.Date) {
			return nil, fmt.Errorf("invalid saved study session %q", session.ID)
		}
		seen[session.ID] = true
	}
	return s, nil
}

// Called under mu. Memory changes only after the replacement has succeeded.
func (s *Server) save(next State) error {
	content, err := json.MarshalIndent(next, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(s.path)
	if err = os.MkdirAll(dir, 0750); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".progress-*.tmp") // mode 0600, exclusive unpredictable name
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(content); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err = os.Rename(f.Name(), s.path); err != nil {
		return err
	}
	s.state = next
	return nil
}
func (s *Server) copyState() State {
	next := State{Progress: make(map[string]Progress, len(s.state.Progress)), Sessions: append([]Session{}, s.state.Sessions...)}
	for id, p := range s.state.Progress {
		next.Progress[id] = p
	}
	return next
}
