//go:build web

package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
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

	mux.HandleFunc("GET /api/courses/{id}/at-risk", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		inactDays := 10
		if dStr := r.URL.Query().Get("inactivity_days"); dStr != "" {
			if dVal, err := strconv.Atoi(dStr); err == nil && dVal > 0 {
				inactDays = dVal
			}
		}
		gradeCut := 70.0
		if gStr := r.URL.Query().Get("grade_cutoff"); gStr != "" {
			if gVal, err := strconv.ParseFloat(gStr, 64); err == nil && gVal > 0 {
				gradeCut = gVal
			}
		}
		consecThresh := 2
		if cStr := r.URL.Query().Get("consecutive_threshold"); cStr != "" {
			if cVal, err := strconv.Atoi(cStr); err == nil && cVal > 0 {
				consecThresh = cVal
			}
		}
		res, err := client.DetectAtRiskStudents(id, inactDays, gradeCut, consecThresh)
		sendWebJSON(w, res, err)
	})

	// Endpoints de Submissões e Visualização de Códigos dos Alunos
	mux.HandleFunc("GET /api/courses/{id}/assignments/{assignment_id}/submissions", func(w http.ResponseWriter, r *http.Request) {
		courseID := r.PathValue("id")
		assignID := r.PathValue("assignment_id")
		res, err := client.GetSubmissionsDetails(courseID, assignID, false)
		sendWebJSON(w, res, err)
	})

	mux.HandleFunc("GET /api/courses/{id}/assignments/{assignment_id}/plagiarism", func(w http.ResponseWriter, r *http.Request) {
		courseID := r.PathValue("id")
		assignID := r.PathValue("assignment_id")
		threshold := 65.0
		if tStr := r.URL.Query().Get("threshold"); tStr != "" {
			if tVal, err := strconv.ParseFloat(tStr, 64); err == nil && tVal > 0 {
				threshold = tVal
			}
		}
		res, err := client.CheckSubmissionsSimilarity(courseID, assignID, threshold, true)
		sendWebJSON(w, res, err)
	})

	mux.HandleFunc("POST /api/code/test", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Code             string `json:"code"`
			Language         string `json:"language"`
			ExpectedFunction string `json:"expected_function,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "JSON inválido", http.StatusBadRequest)
			return
		}

		if strings.TrimSpace(req.Code) == "" {
			http.Error(w, "Código vazio", http.StatusBadRequest)
			return
		}

		// Cria pasta scratch se não existir
		_ = os.MkdirAll("scratch", 0755)
		tmpFile, err := os.CreateTemp("scratch", "eval_*.c")
		if err != nil {
			sendWebJSON(w, nil, fmt.Errorf("erro ao criar arquivo temporário: %v", err))
			return
		}
		tmpPath := tmpFile.Name()
		defer os.Remove(tmpPath)

		if _, err := tmpFile.WriteString(req.Code); err != nil {
			tmpFile.Close()
			sendWebJSON(w, nil, fmt.Errorf("erro ao salvar código: %v", err))
			return
		}
		tmpFile.Close()

		// Invoca scripts/test_c_submissions.py
		cmdArgs := []string{"scripts/test_c_submissions.py", tmpPath}
		if req.ExpectedFunction != "" {
			cmdArgs = append(cmdArgs, req.ExpectedFunction)
		}

		cmd := exec.Command("python3", cmdArgs...)
		out, _ := cmd.CombinedOutput()

		var parsedResult any
		if jsonErr := json.Unmarshal(out, &parsedResult); jsonErr == nil {
			sendWebJSON(w, parsedResult, nil)
		} else {
			sendWebJSON(w, map[string]any{
				"status": "raw_output",
				"output": string(out),
			}, nil)
		}
	})

	// Endpoints do Inbox de Mensagens do Canvas
	mux.HandleFunc("GET /api/inbox", func(w http.ResponseWriter, r *http.Request) {
		scope := r.URL.Query().Get("scope")
		courseID := r.URL.Query().Get("course_id")
		limitStr := r.URL.Query().Get("limit")
		limit := 25
		if limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
				limit = l
			}
		}
		res, err := client.ListInboxConversations(scope, courseID, limit)
		sendWebJSON(w, res, err)
	})

	mux.HandleFunc("GET /api/inbox/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		res, err := client.GetInboxConversation(id)
		sendWebJSON(w, res, err)
	})

	mux.HandleFunc("POST /api/inbox/{id}/reply", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var body struct {
			Message string `json:"message"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "JSON inválido", http.StatusBadRequest)
			return
		}
		res, err := client.ReplyInboxConversation(id, body.Message)
		sendWebJSON(w, res, err)
	})

	mux.HandleFunc("POST /api/inbox", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			RecipientIDs []string `json:"recipient_ids"`
			Subject      string   `json:"subject"`
			Message      string   `json:"message"`
			CourseID     string   `json:"course_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "JSON inválido", http.StatusBadRequest)
			return
		}
		res, err := client.SendInboxMessage(body.RecipientIDs, body.Subject, body.Message, body.CourseID)
		sendWebJSON(w, res, err)
	})

	// Endpoints de Gerenciamento de Cache
	mux.HandleFunc("GET /api/cache/stats", func(w http.ResponseWriter, r *http.Request) {
		stats := client.GetCacheStats()
		sendWebJSON(w, stats, nil)
	})

	mux.HandleFunc("POST /api/cache/clear", func(w http.ResponseWriter, r *http.Request) {
		courseID := r.URL.Query().Get("course_id")
		if courseID != "" {
			removed := client.ClearCourseCache(courseID)
			sendWebJSON(w, map[string]any{
				"status":       "ok",
				"keys_removed": removed,
				"course_id":    courseID,
			}, nil)
			return
		}
		removed := client.ClearCache()
		sendWebJSON(w, map[string]any{
			"status":       "ok",
			"keys_removed": removed,
		}, nil)
	})

	// Arquivos estáticos da interface web
	var staticFS http.FileSystem
	if fi, err := os.Stat("src/public"); err == nil && fi.IsDir() {
		staticFS = http.Dir("src/public")
	} else if fi, err := os.Stat("public"); err == nil && fi.IsDir() {
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
