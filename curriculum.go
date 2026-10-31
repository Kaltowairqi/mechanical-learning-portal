package main

import (
	"fmt"
	"regexp"
)

type Module struct {
	ID                    string   `json:"id"`
	Title                 string   `json:"title"`
	Hours                 int      `json:"hours"`
	Resources             string   `json:"resources"`
	ResourcesVerification string   `json:"resourcesVerification"`
	Criteria              string   `json:"criteria"`
	Prerequisites         []string `json:"prerequisites"`
	Assessment            string   `json:"assessment"`
}
type Stage struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Hours       int      `json:"hours"`
	Modules     []Module `json:"modules"`
}
type Curriculum struct {
	Title       string  `json:"title"`
	Version     int     `json:"version"`
	Description string  `json:"description"`
	Stages      []Stage `json:"stages"`
}

var identifier = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]*$`)

// Validate the content graph before it can affect saved learning records.
func (c Curriculum) validate() (map[string]Module, error) {
	modules := map[string]Module{}
	stages := map[string]bool{}
	if c.Title == "" || c.Version < 1 || len(c.Stages) == 0 {
		return nil, fmt.Errorf("curriculum needs a title, version and stages")
	}
	for _, stage := range c.Stages {
		if !identifier.MatchString(stage.ID) || stages[stage.ID] || stage.Name == "" || len(stage.Modules) == 0 {
			return nil, fmt.Errorf("invalid or duplicate stage %q", stage.ID)
		}
		stages[stage.ID] = true
		hours := 0
		for _, m := range stage.Modules {
			if !identifier.MatchString(m.ID) || m.Title == "" || m.Hours <= 0 || m.Criteria == "" || m.Assessment == "" || m.Resources == "" {
				return nil, fmt.Errorf("invalid module %q", m.ID)
			}
			if _, ok := modules[m.ID]; ok {
				return nil, fmt.Errorf("duplicate module %s", m.ID)
			}
			if m.ResourcesVerification != "requires_verification" && m.ResourcesVerification != "verified" {
				return nil, fmt.Errorf("module %s needs a resource verification status", m.ID)
			}
			modules[m.ID] = m
			hours += m.Hours
		}
		if hours != stage.Hours {
			return nil, fmt.Errorf("stage %s hours do not match its modules", stage.ID)
		}
	}
	visiting, visited := map[string]bool{}, map[string]bool{}
	var visit func(string) error
	visit = func(id string) error {
		if visiting[id] {
			return fmt.Errorf("prerequisite cycle at %s", id)
		}
		if visited[id] {
			return nil
		}
		visiting[id] = true
		seen := map[string]bool{}
		for _, dep := range modules[id].Prerequisites {
			if _, ok := modules[dep]; !ok {
				return fmt.Errorf("module %s has unknown prerequisite %s", id, dep)
			}
			if seen[dep] {
				return fmt.Errorf("module %s repeats prerequisite %s", id, dep)
			}
			seen[dep] = true
			if err := visit(dep); err != nil {
				return err
			}
		}
		visiting[id] = false
		visited[id] = true
		return nil
	}
	for id := range modules {
		if err := visit(id); err != nil {
			return nil, err
		}
	}
	return modules, nil
}
