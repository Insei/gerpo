package task

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Handler exposes the Service over a small JSON REST surface.
//
// Routes:
//   POST   /tasks          — create
//   GET    /tasks          — list (?page, ?size, ?done=true|false)
//   GET    /tasks/{id}     — fetch one
//   PATCH  /tasks/{id}     — partial update
//   DELETE /tasks/{id}     — delete
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Register wires the handler onto a standard net/http mux using Go 1.22+
// pattern routing.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /tasks", h.create)
	mux.HandleFunc("GET /tasks", h.list)
	mux.HandleFunc("GET /tasks/{id}", h.get)
	mux.HandleFunc("PATCH /tasks/{id}", h.update)
	mux.HandleFunc("DELETE /tasks/{id}", h.delete)
}

// taskDTO decouples the HTTP shape from the domain Task. Keeps timestamps
// optional and avoids leaking internal column names if they ever diverge.
type taskDTO struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Done        bool       `json:"done"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

func toDTO(t *Task) taskDTO {
	return taskDTO{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Done:        t.Done,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

type createRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type updateRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Done        *bool   `json:"done,omitempty"`
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if strings.TrimSpace(req.Title) == "" {
		writeErr(w, http.StatusBadRequest, "title is required")
		return
	}
	t := &Task{Title: req.Title, Description: req.Description}
	if err := h.svc.Create(r.Context(), t); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, toDTO(t))
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	var p ListParams
	if s := q.Get("page"); s != "" {
		v, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "invalid page")
			return
		}
		p.Page = v
	}
	if s := q.Get("size"); s != "" {
		v, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "invalid size")
			return
		}
		p.Size = v
	}
	if s := q.Get("done"); s != "" {
		v, err := strconv.ParseBool(s)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "invalid done")
			return
		}
		p.Done = &v
	}

	tasks, err := h.svc.List(r.Context(), p)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]taskDTO, 0, len(tasks))
	for _, t := range tasks {
		out = append(out, toDTO(t))
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	t, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeServiceErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toDTO(t))
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	// Fetch-then-apply: simplest way to emulate PATCH without duplicating
	// the update column list. Cache dedupes the read inside a single request.
	current, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeServiceErr(w, err)
		return
	}
	var req updateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Title != nil {
		current.Title = *req.Title
	}
	if req.Description != nil {
		current.Description = *req.Description
	}
	if req.Done != nil {
		current.Done = *req.Done
	}
	if err := h.svc.Update(r.Context(), current); err != nil {
		writeServiceErr(w, err)
		return
	}
	// Re-read to pick up the trigger-driven UpdatedAt.
	fresh, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeServiceErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toDTO(fresh))
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		writeServiceErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	raw := r.PathValue("id")
	id, err := uuid.Parse(raw)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return uuid.Nil, false
	}
	return id, true
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func writeServiceErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		writeErr(w, http.StatusNotFound, err.Error())
	default:
		writeErr(w, http.StatusInternalServerError, err.Error())
	}
}
