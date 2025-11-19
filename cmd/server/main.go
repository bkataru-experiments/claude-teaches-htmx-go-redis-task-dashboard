package main

import (
	"html/template"
	"log"
	"net/http"

	"task-dashboard/internal/handlers"
	"task-dashboard/internal/middleware"
	"task-dashboard/internal/store"
)

func main() {
	// Initialize Redis store
	redisStore := store.NewRedisStore("host.docker.internal:6379")

	// Load templates
	templates := template.Must(template.ParseGlob("templates/*.html"))
	template.Must(templates.ParseGlob("templates/components/*.html"))

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(redisStore, templates)
	taskHandler := handlers.NewTaskHandler(redisStore, templates)
	dashboardHandler := handlers.NewDashboardHandler(redisStore, templates)

	// Create router
	mux := http.NewServeMux()

	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Public routes
	mux.HandleFunc("GET /login", authHandler.ShowLogin)
	mux.HandleFunc("GET /register", authHandler.ShowRegister)
	mux.HandleFunc("POST /login", authHandler.Login)
	mux.HandleFunc("POST /register", authHandler.Register)

	// Protected routes
	authMux := http.NewServeMux()
	authMux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	})
	authMux.HandleFunc("GET /dashboard", dashboardHandler.ShowDashboard)
	authMux.HandleFunc("GET /tasks", taskHandler.ShowTasks)
	authMux.HandleFunc("GET /tasks/new", taskHandler.ShowTaskForm)
	authMux.HandleFunc("GET /tasks/{id}/edit", taskHandler.ShowTaskForm)
	authMux.HandleFunc("GET /logout", authHandler.Logout)

	// API routes
	authMux.HandleFunc("GET /api/stats", dashboardHandler.GetStats)
	authMux.HandleFunc("GET /api/tasks", taskHandler.ListTasks)
	authMux.HandleFunc("POST /api/tasks", taskHandler.CreateTask)
	authMux.HandleFunc("PUT /api/tasks/{id}", taskHandler.UpdateTask)
	authMux.HandleFunc("DELETE /api/tasks/{id}", taskHandler.DeleteTask)

	// Wrap protected routes with auth middleware
	mux.Handle("/", middleware.AuthMiddleware(redisStore)(authMux))

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
