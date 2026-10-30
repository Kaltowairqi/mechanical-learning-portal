package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func send(w http.ResponseWriter, code int, x any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(x)
}
func errResp(w http.ResponseWriter, code int, msg string) {
	send(w, code, map[string]string{"error": msg})
}
func decode(w http.ResponseWriter, r *http.Request, value any) error {
	media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || media != "application/json" {
		return errors.New("Content-Type must be application/json")
	}
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err = dec.Decode(value); err != nil {
		return err
	}
	if err = dec.Decode(new(any)); err != io.EOF {
		return errors.New("unexpected trailing JSON")
	}
	return nil
}
func (s *Server) api(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	// Browser requests must come from this portal's origin. No CORS is enabled.
	if origin := r.Header.Get("Origin"); origin != "" {
		u, err := url.Parse(origin)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || !strings.EqualFold(u.Host, r.Host) || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
			errResp(w, 403, "cross-origin API request denied")
			return
		}
	}
	if r.Header.Get("Sec-Fetch-Site") == "cross-site" {
		errResp(w, 403, "cross-site API request denied")
		return
	}
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/api/curriculum":
		send(w, 200, s.curriculum)
	case r.Method == http.MethodGet && (r.URL.Path == "/api/state" || r.URL.Path == "/api/export"):
		s.mu.Lock()
		defer s.mu.Unlock()
		if r.URL.Path == "/api/export" {
			w.Header().Set("Content-Disposition", "attachment; filename=helicopter-progress.json")
		}
		send(w, 200, s.state)
	case r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/api/progress/"):
		id := strings.TrimPrefix(r.URL.Path, "/api/progress/")
		mod, ok := s.valid[id]
		if !ok {
			errResp(w, 404, "unknown module")
			return
		}
		var req struct {
			Status             string   `json:"status"`
			Score              *float64 `json:"score"`
			Notes              string   `json:"notes"`
			Evidence           string   `json:"evidence"`
			PracticalConfirmed bool     `json:"practicalConfirmed"`
		}
		if err := decode(w, r, &req); err != nil {
			errResp(w, 400, err.Error())
			return
		}
		if req.Status == "submitted" {
			req.Status = "ready_for_assessment"
		}
		p := Progress{Status: req.Status, Score: req.Score, Notes: req.Notes, Evidence: req.Evidence, PracticalConfirmed: req.PracticalConfirmed, UpdatedAt: time.Now().UTC().Format(time.RFC3339)}
		if !validStatus(p.Status) {
			errResp(w, 400, "invalid status")
			return
		}
		if len(p.Notes) > 5000 || len(p.Evidence) > 5000 {
			errResp(w, 400, "notes or evidence too long (5000 bytes maximum)")
			return
		}
		if p.Score != nil && (*p.Score < 0 || *p.Score > 100) {
			errResp(w, 400, "score must be 0..100")
			return
		}
		if p.Status == "passed" && !passingEvidence(p) {
			errResp(w, 400, "passing requires score >=80, evidence references and confirmation of mandatory practical work")
			return
		}
		s.mu.Lock()
		defer s.mu.Unlock()
		if p.Status == "passed" {
			for _, dep := range mod.Prerequisites {
				if s.state.Progress[dep].Status != "passed" {
					errResp(w, 409, "pass prerequisite "+dep+" first")
					return
				}
			}
			p.CompletedAt = s.state.Progress[id].CompletedAt
			if p.CompletedAt == "" {
				p.CompletedAt = p.UpdatedAt
			}
		} else if s.state.Progress[id].Status == "passed" {
			for other, m := range s.valid {
				if s.state.Progress[other].Status != "passed" {
					continue
				}
				for _, dep := range m.Prerequisites {
					if dep == id {
						errResp(w, 409, "revise passed dependent "+other+" first")
						return
					}
				}
			}
		}
		next := s.copyState()
		next.Progress[id] = p
		if err := s.save(next); err != nil {
			errResp(w, 500, "failed saving progress; previous state retained")
			return
		}
		send(w, 200, p)
	case r.Method == http.MethodPost && r.URL.Path == "/api/sessions":
		var req struct {
			ModuleID string `json:"moduleId"`
			Minutes  int    `json:"minutes"`
			Note     string `json:"note"`
			Date     string `json:"date"`
		}
		if err := decode(w, r, &req); err != nil {
			errResp(w, 400, err.Error())
			return
		}
		if _, ok := s.valid[req.ModuleID]; !ok {
			errResp(w, 400, "unknown module")
			return
		}
		if req.Minutes < 1 || req.Minutes > 1440 || len(req.Note) > 2000 {
			errResp(w, 400, "invalid session duration or note")
			return
		}
		if req.Date == "" {
			req.Date = time.Now().Format("2006-01-02")
		}
		if !validateDate(req.Date) {
			errResp(w, 400, "date must be YYYY-MM-DD")
			return
		}
		var token [16]byte
		if _, err := rand.Read(token[:]); err != nil {
			errResp(w, 500, "failed creating session ID")
			return
		}
		session := Session{ID: hex.EncodeToString(token[:]), ModuleID: req.ModuleID, Minutes: req.Minutes, Note: req.Note, Date: req.Date}
		s.mu.Lock()
		defer s.mu.Unlock()
		next := s.copyState()
		next.Sessions = append(next.Sessions, session)
		if err := s.save(next); err != nil {
			errResp(w, 500, "failed saving session; previous state retained")
			return
		}
		send(w, 201, session)
	default:
		errResp(w, 404, "not found")
	}
}

func (s *Server) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/", s.api)
	mux.Handle("/", http.FileServer(http.Dir("static")))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; base-uri 'none'; object-src 'none'; frame-ancestors 'none'; form-action 'self'")
		mux.ServeHTTP(w, r)
	})
}
