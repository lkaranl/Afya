package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
)

//go:embed public/*
var embeddedPublic embed.FS

func runWebServer(client *CanvasClient, port string) {
	mux := http.NewServeMux()
	engine := NewAgentEngine(client)

	// Endpoints do Assistente Conversacional IA (Chat)
	mux.HandleFunc("GET /api/agent/info", func(w http.ResponseWriter, r *http.Request) {
		info := map[string]any{
			"provider":         engine.Provider,
			"model":            engine.Model,
			"has_key":          engine.APIKey != "" || engine.Provider == "ollama",
			"canvas_connected": client.Token != "",
		}
		sendWebJSON(w, info, nil)
	})

	mux.HandleFunc("POST /api/chat", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Message string        `json:"message"`
			History []ChatMessage `json:"history"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "JSON inválido", http.StatusBadRequest)
			return
		}
		res, err := engine.Chat(r.Context(), req.Message, req.History)
		sendWebJSON(w, res, err)
	})

	mux.HandleFunc("POST /api/chat/confirm", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			CourseID     string        `json:"course_id"`
			AssignmentID string        `json:"assignment_id"`
			Grades       []GradeEntry `json:"grades"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "JSON inválido", http.StatusBadRequest)
			return
		}
		res, err := client.SubmitGradesBatch(req.CourseID, req.AssignmentID, req.Grades)
		sendWebJSON(w, res, err)
	})

	// Endpoints da API para o frontend
	mux.HandleFunc("GET /api/me", func(w http.ResponseWriter, r *http.Request) {
		res, err := client.GetUserProfile()
		sendWebJSON(w, res, err)
	})

	mux.HandleFunc("GET /api/courses", func(w http.ResponseWriter, r *http.Request) {
		res, err := client.ListCourses()
		sendWebJSON(w, res, err)
	})

	mux.HandleFunc("GET /api/courses/{id}/assignments", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		res, err := client.ListAssignments(id)
		sendWebJSON(w, res, err)
	})

	mux.HandleFunc("GET /api/courses/{id}/announcements", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		res, err := client.ListAnnouncements(id)
		sendWebJSON(w, res, err)
	})

	mux.HandleFunc("POST /api/courses/{id}/announcements", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var body struct {
			Title   string `json:"title"`
			Message string `json:"message"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "JSON inválido", http.StatusBadRequest)
			return
		}
		res, err := client.PostAnnouncement(id, body.Title, body.Message)
		sendWebJSON(w, res, err)
	})

	mux.HandleFunc("GET /api/courses/{id}/students", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		res, err := client.ListStudents(id)
		sendWebJSON(w, res, err)
	})

	// Arquivos estáticos da interface web
	var staticFS http.FileSystem
	if fi, err := os.Stat("public"); err == nil && fi.IsDir() {
		staticFS = http.Dir("public")
	} else {
		sub, err := fs.Sub(embeddedPublic, "public")
		if err != nil {
			log.Fatalf("Erro ao carregar arquivos estáticos embutidos: %v", err)
		}
		staticFS = http.FS(sub)
	}

	fileServer := http.FileServer(staticFS)
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		fileServer.ServeHTTP(w, r)
	})

	// Middleware de CORS
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		mux.ServeHTTP(w, r)
	})

	addr := fmt.Sprintf(":%s", port)
	log.Printf("[Afya Canvas Hub - Web] Painel web rodando na porta %s (acesse http://localhost:%s)", port, port)
	log.Printf("[Afya Canvas Hub - Web] Canvas Base URL: %s", client.BaseURL)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Erro ao iniciar servidor web: %v", err)
	}
}

func sendWebJSON(w http.ResponseWriter, data any, err error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}
