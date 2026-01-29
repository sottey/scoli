package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const taskFiltersFileName = "task-sets.json"

type TaskFiltersResponse struct {
	Filters TaskFilters `json:"filters"`
	Notice  string      `json:"notice,omitempty"`
}

type TaskFilters struct {
	Version int          `json:"version"`
	Filters []TaskFilter `json:"filters"`
}

type TaskFilter struct {
	ID         string              `json:"id"`
	Name       string              `json:"name"`
	Tags       []string            `json:"tags,omitempty"`
	Mentions   []string            `json:"mentions,omitempty"`
	Projects   []string            `json:"projects,omitempty"`
	Due        *TaskFilterDue      `json:"due,omitempty"`
	Priority   *TaskFilterPriority `json:"priority,omitempty"`
	Completed  *bool               `json:"completed,omitempty"`
	Text       string              `json:"text,omitempty"`
	PathPrefix string              `json:"pathPrefix,omitempty"`
}

type TaskFilterDue struct {
	From string `json:"from,omitempty"`
	To   string `json:"to,omitempty"`
}

type TaskFilterPriority struct {
	Min *int `json:"min,omitempty"`
	Max *int `json:"max,omitempty"`
}

func (s *Server) handleTaskFiltersGet(w http.ResponseWriter, r *http.Request) {
	filters, notice, err := s.loadTaskFilters()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to load task filters")
		return
	}
	resp := TaskFiltersResponse{Filters: filters}
	if notice != "" {
		resp.Notice = notice
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleTaskFiltersUpdate(w http.ResponseWriter, r *http.Request) {
	payload, err := decodeJSON[TaskFilters](r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	if err := validateTaskFilters(payload); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.saveTaskFilters(payload); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to save task filters")
		return
	}
	writeJSON(w, http.StatusOK, TaskFiltersResponse{Filters: payload})
}

func (s *Server) taskFiltersFilePath() string {
	return filepath.Join(s.notesDir, taskFiltersFileName)
}

func (s *Server) loadTaskFilters() (TaskFilters, string, error) {
	path := s.taskFiltersFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			filters := TaskFilters{Version: 1, Filters: []TaskFilter{}}
			if err := s.saveTaskFilters(filters); err != nil {
				return filters, "", err
			}
			s.logger.Info("task filters created", "path", path)
			return filters, "Created task-sets.json", nil
		}
		return TaskFilters{}, "", err
	}

	var filters TaskFilters
	if err := json.Unmarshal(data, &filters); err != nil {
		return TaskFilters{}, "", err
	}
	if filters.Version == 0 {
		filters.Version = 1
	}
	if filters.Filters == nil {
		filters.Filters = []TaskFilter{}
	}
	if err := validateTaskFilters(filters); err != nil {
		return TaskFilters{}, "", err
	}
	return filters, "", nil
}

func (s *Server) saveTaskFilters(filters TaskFilters) error {
	data, err := json.MarshalIndent(filters, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.taskFiltersFilePath(), data, 0o644)
}

func validateTaskFilters(filters TaskFilters) error {
	if filters.Version < 1 {
		return errors.New("version must be >= 1")
	}
	ids := make(map[string]bool)
	names := make(map[string]bool)
	for _, filter := range filters.Filters {
		id := strings.TrimSpace(filter.ID)
		name := strings.TrimSpace(filter.Name)
		if id == "" {
			return errors.New("filter id is required")
		}
		if name == "" {
			return errors.New("filter name is required")
		}
		idKey := strings.ToLower(id)
		nameKey := strings.ToLower(name)
		if ids[idKey] {
			return errors.New("filter id must be unique")
		}
		if names[nameKey] {
			return errors.New("filter name must be unique")
		}
		ids[idKey] = true
		names[nameKey] = true
	}

	return nil
}
