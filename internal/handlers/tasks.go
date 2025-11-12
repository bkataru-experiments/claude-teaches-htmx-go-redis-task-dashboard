package handlers

import (
	"html/template"
	"net/http"
	"time"

	"task-dashboard/internal/middleware"
	"task-dashboard/internal/models"
	"task-dashboard/internal/store"

	"github.com/google/uuid"
)

type TaskHandler struct {
	store     *store.RedisStore
	templates *template.Template
}

func NewTaskHandler(store *store.RedisStore, templates *template.Template) *TaskHandler {
	return &TaskHandler{
		store:     store,
		templates: templates,
	}
}

func (h *TaskHandler) ShowTasks(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	status := r.URL.Query().Get("status")

	tasks, err := h.store.GetUserTasks(r.Context(), userID, status)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if err := h.templates.ExecuteTemplate(w, "tasks.html", map[string]interface{}{
		"Tasks":  tasks,
		"Status": status,
	}); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
}

func (h *TaskHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	status := r.URL.Query().Get("status")

	tasks, err := h.store.GetUserTasks(r.Context(), userID, status)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if err := h.templates.ExecuteTemplate(w, "task-list.html", map[string]interface{}{
		"Tasks": tasks,
	}); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	task := &models.Task{
		ID:          uuid.New().String(),
		UserID:      userID,
		Title:       r.FormValue("title"),
		Description: r.FormValue("description"),
		Status:      "pending",
		Priority:    r.FormValue("priority"),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.store.CreateTask(r.Context(), task); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// Return updated task list
	tasks, err := h.store.GetUserTasks(r.Context(), userID, "")
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if err := h.templates.ExecuteTemplate(w, "task-list.html", map[string]interface{}{
		"Tasks": tasks,
	}); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	taskID := r.PathValue("id")

	task, err := h.store.GetTask(r.Context(), taskID)
	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	if title := r.FormValue("title"); title != "" {
		task.Title = title
	}
	if desc := r.FormValue("description"); desc != "" {
		task.Description = desc
	}
	if status := r.FormValue("status"); status != "" {
		task.Status = status
	}
	if priority := r.FormValue("priority"); priority != "" {
		task.Priority = priority
	}

	if err := h.store.UpdateTask(r.Context(), task); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if err := h.templates.ExecuteTemplate(w, "task-item.html", task); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	taskID := r.PathValue("id")

	if err := h.store.DeleteTask(r.Context(), userID, taskID); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *TaskHandler) ShowTaskForm(w http.ResponseWriter, r *http.Request) {
	taskID := r.PathValue("id")

	if taskID == "" {
		// New task form
		if err := h.templates.ExecuteTemplate(w, "task-form.html", nil); err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		return
	}

	task, err := h.store.GetTask(r.Context(), taskID)
	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	if err := h.templates.ExecuteTemplate(w, "task-form.html", task); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
}
