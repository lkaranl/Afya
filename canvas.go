package main

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	htmlBreakRegex    = regexp.MustCompile(`(?i)<br\s*/?>`)
	htmlBlockRegex    = regexp.MustCompile(`(?i)</(p|div|li|h[1-6]|tr)>`)
	htmlTagRegex      = regexp.MustCompile(`<[^>]+>`)
	sanitizeNameRegex = regexp.MustCompile(`[^\w\.-]`)
	githubURLRegex    = regexp.MustCompile(`(?i)(?:https?://)?(?:www\.)?github\.com/([a-zA-Z0-9_.-]+)/([a-zA-Z0-9_.-]+)`)
	sourceFileExts    = map[string]bool{
		".c": true, ".h": true, ".cpp": true, ".hpp": true,
		".py": true, ".rs": true, ".java": true, ".go": true,
		".js": true, ".ts": true, ".html": true, ".css": true,
		".txt": true, ".md": true, ".json": true, ".sh": true,
	}
)

// CleanCanvasHTML higieniza o HTML do Canvas mantendo operadores matemáticos e quebras de linha
func CleanCanvasHTML(rawHTML string) string {
	if rawHTML == "" {
		return ""
	}
	text := htmlBreakRegex.ReplaceAllString(rawHTML, "\n")
	text = htmlBlockRegex.ReplaceAllString(text, "\n")
	text = htmlTagRegex.ReplaceAllString(text, "")
	text = html.UnescapeString(text)
	text = strings.ReplaceAll(text, "\u00a0", " ")
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	lines := strings.Split(text, "\n")
	var cleaned []string
	for _, l := range lines {
		cleaned = append(cleaned, strings.TrimRight(l, " \t"))
	}
	return strings.TrimSpace(strings.Join(cleaned, "\n"))
}

// FormatBRDateTime converte timestamp ISO UTC para o horário de Brasília (UTC-3)
func FormatBRDateTime(isoStr string) string {
	if isoStr == "" {
		return "Sem prazo definido"
	}
	t, err := time.Parse(time.RFC3339, isoStr)
	if err != nil {
		return isoStr
	}
	loc := time.FixedZone("BRT", -3*3600)
	tBR := t.In(loc)
	return tBR.Format("02/01/2006 às 15:04")
}

func detectHasCode(body string) bool {
	lower := strings.ToLower(body)
	return strings.Contains(body, "#include") ||
		strings.Contains(body, "int main") ||
		strings.Contains(body, "void ") ||
		strings.Contains(body, "def ") ||
		strings.Contains(lower, "public class") ||
		strings.Contains(lower, "import ") ||
		strings.Contains(body, "printf(") ||
		strings.Contains(body, "scanf(") ||
		strings.Contains(body, "malloc(") ||
		strings.Contains(body, "struct ")
}

type CanvasClient struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

func NewCanvasClient(baseURL, token string) *CanvasClient {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		Resolver: &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{Timeout: 3 * time.Second}
				conn, err := d.DialContext(ctx, "udp", "8.8.8.8:53")
				if err != nil {
					return d.DialContext(ctx, "udp", "1.1.1.1:53")
				}
				return conn, nil
			},
		},
	}
	transport := &http.Transport{
		DialContext:         dialer.DialContext,
		ForceAttemptHTTP2:   true,
		MaxIdleConns:        100,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	return &CanvasClient{
		BaseURL:    strings.TrimRight(baseURL, "/"),
		Token:      token,
		HTTPClient: &http.Client{
			Transport: transport,
			Timeout:   45 * time.Second,
		},
	}
}

func (c *CanvasClient) Request(method, endpoint string, body any) ([]byte, int, error) {
	if c.Token == "" {
		return nil, http.StatusUnauthorized, fmt.Errorf("TOKEN do Canvas não configurado no .env")
	}

	targetURL := fmt.Sprintf("%s%s", c.BaseURL, endpoint)
	var bodyReader io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("erro ao serializar corpo da requisição: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBytes)
	}

	req, err := http.NewRequest(method, targetURL, bodyReader)
	if err != nil {
		return nil, 0, fmt.Errorf("erro ao criar requisição HTTP: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("Accept", "application/json+canvas-string-ids, application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("erro ao comunicar com Canvas LMS: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("erro ao ler resposta do Canvas: %w", err)
	}

	if resp.StatusCode >= 400 {
		return respBytes, resp.StatusCode, fmt.Errorf("Canvas API erro HTTP %d: %s", resp.StatusCode, string(respBytes))
	}

	return respBytes, resp.StatusCode, nil
}

func (c *CanvasClient) GetUserProfile() (any, error) {
	data, _, err := c.Request("GET", "/api/v1/users/self", nil)
	if err != nil {
		return nil, err
	}
	var result any
	err = json.Unmarshal(data, &result)
	return result, err
}

func (c *CanvasClient) ListCourses() (any, error) {
	data, _, err := c.Request("GET", "/api/v1/courses?per_page=50&include[]=total_students&include[]=term", nil)
	if err != nil {
		return nil, err
	}
	var result any
	err = json.Unmarshal(data, &result)
	return result, err
}

func (c *CanvasClient) ListAssignments(courseID string) (any, error) {
	endpoint := fmt.Sprintf("/api/v1/courses/%s/assignments?per_page=100&order_by=due_at", url.PathEscape(courseID))
	data, _, err := c.Request("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	var result any
	err = json.Unmarshal(data, &result)
	return result, err
}

func (c *CanvasClient) GetAssignment(courseID, assignmentID string) (any, error) {
	endpoint := fmt.Sprintf("/api/v1/courses/%s/assignments/%s", url.PathEscape(courseID), url.PathEscape(assignmentID))
	data, _, err := c.Request("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	var result any
	err = json.Unmarshal(data, &result)
	return result, err
}

func (c *CanvasClient) ListStudents(courseID string) (any, error) {
	endpoint := fmt.Sprintf("/api/v1/courses/%s/users?enrollment_type[]=student&per_page=100", url.PathEscape(courseID))
	data, _, err := c.Request("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	var result any
	err = json.Unmarshal(data, &result)
	return result, err
}

type PendingAssignment struct {
	CourseID          string  `json:"course_id"`
	CourseName        string  `json:"course_name"`
	AssignmentID      string  `json:"assignment_id"`
	AssignmentName    string  `json:"assignment_name"`
	DueAt             string  `json:"due_at"`
	DueAtBR           string  `json:"due_at_br"`
	PointsPossible    float64 `json:"points_possible"`
	NeedsGradingCount int     `json:"needs_grading_count"`
	HTMLURL           string  `json:"html_url"`
}

// ListPendingAssignments varre todas as disciplinas do professor e retorna as atividades que têm tarefas aguardando correção
func (c *CanvasClient) ListPendingAssignments() ([]PendingAssignment, error) {
	coursesData, _, err := c.Request("GET", "/api/v1/courses?per_page=50", nil)
	if err != nil {
		return nil, err
	}

	var courses []map[string]any
	if err := json.Unmarshal(coursesData, &courses); err != nil {
		return nil, err
	}

	var pending []PendingAssignment
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, crs := range courses {
		courseID := fmt.Sprintf("%v", crs["id"])
		courseName := fmt.Sprintf("%v", crs["name"])

		wg.Add(1)
		go func(cID, cName string) {
			defer wg.Done()
			endpoint := fmt.Sprintf("/api/v1/courses/%s/assignments?per_page=50", url.PathEscape(cID))
			assignData, _, err := c.Request("GET", endpoint, nil)
			if err != nil {
				return
			}

			var assignments []map[string]any
			if err := json.Unmarshal(assignData, &assignments); err != nil {
				return
			}

			for _, a := range assignments {
				gradingCount := 0
				if val, ok := a["needs_grading_count"].(float64); ok {
					gradingCount = int(val)
				}

				if gradingCount > 0 {
					var pts float64
					if p, ok := a["points_possible"].(float64); ok {
						pts = p
					}
					due := ""
					if d, ok := a["due_at"].(string); ok {
						due = d
					}
					htmlURL := ""
					if u, ok := a["html_url"].(string); ok {
						htmlURL = u
					}

					mu.Lock()
					pending = append(pending, PendingAssignment{
						CourseID:          cID,
						CourseName:        cName,
						AssignmentID:      fmt.Sprintf("%v", a["id"]),
						AssignmentName:    fmt.Sprintf("%v", a["name"]),
						DueAt:             due,
						DueAtBR:           FormatBRDateTime(due),
						PointsPossible:    pts,
						NeedsGradingCount: gradingCount,
						HTMLURL:           htmlURL,
					})
					mu.Unlock()
				}
			}
		}(courseID, courseName)
	}

	wg.Wait()
	return pending, nil
}

type SubmissionDetail struct {
	UserID         string           `json:"user_id"`
	UserName       string           `json:"user_name"`
	WorkflowState  string           `json:"workflow_state"`
	Grade          string           `json:"grade,omitempty"`
	SubmittedAt    string           `json:"submitted_at"`
	SubmittedAtBR  string           `json:"submitted_at_br"`
	SubmissionType string           `json:"submission_type"`
	URL            string           `json:"url,omitempty"`
	Body           string           `json:"body,omitempty"`
	CleanBody      string           `json:"clean_body,omitempty"`
	HasCode        bool             `json:"has_code"`
	Attachments    []map[string]any `json:"attachments,omitempty"`
}

// GetSubmissionsDetails busca todas as submissões com dados do usuário, anexos e código higienizado
func (c *CanvasClient) GetSubmissionsDetails(courseID, assignmentID string, onlyPending bool) ([]SubmissionDetail, error) {
	endpoint := fmt.Sprintf("/api/v1/courses/%s/assignments/%s/submissions?include[]=user&include[]=attachments&per_page=100",
		url.PathEscape(courseID), url.PathEscape(assignmentID))

	data, _, err := c.Request("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var rawList []map[string]any
	if err := json.Unmarshal(data, &rawList); err != nil {
		return nil, err
	}

	var results []SubmissionDetail
	for _, item := range rawList {
		workflowState := fmt.Sprintf("%v", item["workflow_state"])
		if onlyPending && workflowState == "graded" {
			continue
		}

		userName := ""
		if userObj, ok := item["user"].(map[string]any); ok {
			userName = fmt.Sprintf("%v", userObj["name"])
		}

		subType := ""
		if st, ok := item["submission_type"].(string); ok {
			subType = st
		}

		submittedAt := ""
		if sa, ok := item["submitted_at"].(string); ok {
			submittedAt = sa
		}

		bodyText := ""
		if b, ok := item["body"].(string); ok {
			bodyText = b
		}

		cleanBody := CleanCanvasHTML(bodyText)
		hasCode := detectHasCode(cleanBody)

		var atts []map[string]any
		if rawAtts, ok := item["attachments"].([]any); ok {
			for _, a := range rawAtts {
				if aMap, ok := a.(map[string]any); ok {
					atts = append(atts, aMap)
				}
			}
		}

		gradeStr := ""
		if g := item["grade"]; g != nil {
			gradeStr = fmt.Sprintf("%v", g)
		}

		urlStr := ""
		if u, ok := item["url"].(string); ok {
			urlStr = u
		}

		results = append(results, SubmissionDetail{
			UserID:         fmt.Sprintf("%v", item["user_id"]),
			UserName:       userName,
			WorkflowState:  workflowState,
			Grade:          gradeStr,
			SubmittedAt:    submittedAt,
			SubmittedAtBR:  FormatBRDateTime(submittedAt),
			SubmissionType: subType,
			URL:            urlStr,
			Body:           bodyText,
			CleanBody:      cleanBody,
			HasCode:        hasCode,
			Attachments:    atts,
		})
	}

	return results, nil
}

// DownloadAttachment baixa um arquivo anexo da submissão para o disco local
func (c *CanvasClient) DownloadAttachment(downloadURL, destinationPath string) (string, error) {
	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("falha ao baixar anexo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("servidor respondeu com status %d ao baixar arquivo", resp.StatusCode)
	}

	if err := os.MkdirAll(filepath.Dir(destinationPath), 0755); err != nil {
		return "", fmt.Errorf("erro ao criar diretório de destino: %w", err)
	}

	outFile, err := os.Create(destinationPath)
	if err != nil {
		return "", fmt.Errorf("erro ao criar arquivo local: %w", err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("erro ao salvar conteúdo no arquivo: %w", err)
	}

	return destinationPath, nil
}

func (c *CanvasClient) SubmitGrade(courseID, assignmentID, userID, grade, comment string) (any, error) {
	endpoint := fmt.Sprintf("/api/v1/courses/%s/assignments/%s/submissions/%s",
		url.PathEscape(courseID), url.PathEscape(assignmentID), url.PathEscape(userID))

	payload := map[string]any{
		"submission": map[string]string{
			"posted_grade": grade,
		},
	}
	if comment != "" {
		payload["comment"] = map[string]string{
			"text_comment": comment,
		}
	}

	data, _, err := c.Request("PUT", endpoint, payload)
	if err != nil {
		return nil, err
	}
	var result any
	err = json.Unmarshal(data, &result)
	return result, err
}

type GradeEntry struct {
	UserID  string `json:"user_id"`
	Grade   string `json:"grade"`
	Comment string `json:"comment,omitempty"`
}

func (c *CanvasClient) SubmitGradesBatch(courseID, assignmentID string, grades []GradeEntry) (any, error) {
	endpoint := fmt.Sprintf("/api/v1/courses/%s/assignments/%s/submissions/update_grades",
		url.PathEscape(courseID), url.PathEscape(assignmentID))

	gradeData := make(map[string]any)
	for _, entry := range grades {
		item := map[string]string{
			"posted_grade": entry.Grade,
		}
		if entry.Comment != "" {
			item["text_comment"] = entry.Comment
		}
		gradeData[entry.UserID] = item
	}

	payload := map[string]any{
		"grade_data": gradeData,
	}

	data, _, err := c.Request("POST", endpoint, payload)
	if err != nil {
		return nil, err
	}
	var result any
	err = json.Unmarshal(data, &result)
	return result, err
}

func (c *CanvasClient) ListAnnouncements(courseID string) (any, error) {
	endpoint := fmt.Sprintf("/api/v1/courses/%s/discussion_topics?only_announcements=true&per_page=30", url.PathEscape(courseID))
	data, _, err := c.Request("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	var result any
	err = json.Unmarshal(data, &result)
	return result, err
}

func (c *CanvasClient) PostAnnouncement(courseID, title, message string) (any, error) {
	endpoint := fmt.Sprintf("/api/v1/courses/%s/discussion_topics", url.PathEscape(courseID))
	payload := map[string]any{
		"title":           title,
		"message":         message,
		"is_announcement": true,
		"published":       true,
	}
	data, _, err := c.Request("POST", endpoint, payload)
	if err != nil {
		return nil, err
	}
	var result any
	err = json.Unmarshal(data, &result)
	return result, err
}

// -------------------------------------------------------------
// NOVAS FERRAMENTAS DE ALTO NÍVEL MCP
// -------------------------------------------------------------

type StudentGradingInfo struct {
	UserID        string `json:"user_id"`
	UserName      string `json:"user_name"`
	Grade         string `json:"grade,omitempty"`
	WorkflowState string `json:"workflow_state"`
	SubmittedAtBR string `json:"submitted_at_br"`
}

type AssignmentGradingStatus struct {
	AssignmentID     string               `json:"assignment_id"`
	AssignmentName   string               `json:"assignment_name"`
	DueAtBR          string               `json:"due_at_br"`
	PointsPossible   float64              `json:"points_possible"`
	TotalSubmissions int                  `json:"total_submissions"`
	GradedCount      int                  `json:"graded_count"`
	PendingCount     int                  `json:"pending_count"`
	UnsubmittedCount int                  `json:"unsubmitted_count"`
	PercentGraded    float64              `json:"percent_graded"`
	StatusCategory   string               `json:"status_category"` // CONCLUIDA, PARCIAL, PENDENTE, SEM_ENTREGAS
	PendingStudents  []StudentGradingInfo `json:"pending_students,omitempty"`
	GradedStudents   []StudentGradingInfo `json:"graded_students,omitempty"`
}

type CourseGradingStatus struct {
	CourseID    string                    `json:"course_id"`
	Summary     map[string]int            `json:"summary"`
	Assignments []AssignmentGradingStatus `json:"assignments"`
}

// GetGradingStatus analisa o status de correção de um curso ou de uma atividade específica
func (c *CanvasClient) GetGradingStatus(courseID, assignmentID string) (any, error) {
	if assignmentID != "" {
		assignData, err := c.GetAssignment(courseID, assignmentID)
		if err != nil {
			return nil, err
		}
		assignMap, _ := assignData.(map[string]any)

		subs, err := c.GetSubmissionsDetails(courseID, assignmentID, false)
		if err != nil {
			return nil, err
		}

		status := c.calculateAssignmentStatus(assignMap, subs, true)
		return status, nil
	}

	// Análise do curso inteiro
	assignmentsData, err := c.ListAssignments(courseID)
	if err != nil {
		return nil, err
	}

	assignmentsList, ok := assignmentsData.([]any)
	if !ok {
		return nil, fmt.Errorf("resposta inesperada ao listar tarefas do curso")
	}

	var statusList []AssignmentGradingStatus
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 6) // Limitar concorrência para 6 requisições simultâneas

	for _, aItem := range assignmentsList {
		aMap, ok := aItem.(map[string]any)
		if !ok {
			continue
		}
		aID := fmt.Sprintf("%v", aMap["id"])

		wg.Add(1)
		go func(aid string, amap map[string]any) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			subs, err := c.GetSubmissionsDetails(courseID, aid, false)
			if err != nil {
				return
			}
			st := c.calculateAssignmentStatus(amap, subs, false)

			mu.Lock()
			statusList = append(statusList, st)
			mu.Unlock()
		}(aID, aMap)
	}

	wg.Wait()

	sort.Slice(statusList, func(i, j int) bool {
		return statusList[i].AssignmentName < statusList[j].AssignmentName
	})

	summary := map[string]int{
		"concluidas":   0,
		"parciais":     0,
		"pendentes":    0,
		"sem_entregas": 0,
	}

	for _, s := range statusList {
		switch s.StatusCategory {
		case "CONCLUIDA":
			summary["concluidas"]++
		case "PARCIAL":
			summary["parciais"]++
		case "PENDENTE":
			summary["pendentes"]++
		case "SEM_ENTREGAS":
			summary["sem_entregas"]++
		}
	}

	return CourseGradingStatus{
		CourseID:    courseID,
		Summary:     summary,
		Assignments: statusList,
	}, nil
}

func (c *CanvasClient) calculateAssignmentStatus(aMap map[string]any, subs []SubmissionDetail, includeDetails bool) AssignmentGradingStatus {
	aID := fmt.Sprintf("%v", aMap["id"])
	name := fmt.Sprintf("%v", aMap["name"])
	dueStr := ""
	if d, ok := aMap["due_at"].(string); ok {
		dueStr = d
	}
	var pts float64
	if p, ok := aMap["points_possible"].(float64); ok {
		pts = p
	}

	var graded []StudentGradingInfo
	var pending []StudentGradingInfo
	unsubmittedCount := 0

	for _, s := range subs {
		info := StudentGradingInfo{
			UserID:        s.UserID,
			UserName:      s.UserName,
			Grade:         s.Grade,
			WorkflowState: s.WorkflowState,
			SubmittedAtBR: s.SubmittedAtBR,
		}

		if s.WorkflowState == "graded" || s.Grade != "" {
			graded = append(graded, info)
		} else if s.WorkflowState == "submitted" || s.SubmittedAt != "" {
			pending = append(pending, info)
		} else {
			unsubmittedCount++
		}
	}

	total := len(subs)
	gradedCount := len(graded)
	pendingCount := len(pending)

	percent := 0.0
	if total > 0 {
		percent = (float64(gradedCount) / float64(total)) * 100
	}

	category := "SEM_ENTREGAS"
	if pendingCount == 0 && gradedCount > 0 {
		category = "CONCLUIDA"
	} else if gradedCount > 0 && pendingCount > 0 {
		category = "PARCIAL"
	} else if gradedCount == 0 && pendingCount > 0 {
		category = "PENDENTE"
	}

	res := AssignmentGradingStatus{
		AssignmentID:     aID,
		AssignmentName:   name,
		DueAtBR:          FormatBRDateTime(dueStr),
		PointsPossible:   pts,
		TotalSubmissions: total,
		GradedCount:      gradedCount,
		PendingCount:     pendingCount,
		UnsubmittedCount: unsubmittedCount,
		PercentGraded:    percent,
		StatusCategory:   category,
	}

	if includeDetails {
		res.PendingStudents = pending
		res.GradedStudents = graded
	}

	return res
}

type GitHubRepoResult struct {
	GitHubURL   string   `json:"github_url"`
	Owner       string   `json:"owner"`
	Repo        string   `json:"repo"`
	LocalDir    string   `json:"local_dir"`
	CodeFiles   []string `json:"code_files"`
	PrimaryFile string   `json:"primary_file,omitempty"`
	Snippet     string   `json:"snippet,omitempty"`
	TotalFiles  int      `json:"total_files"`
}

// ExtractGitHubRepo extrai a URL canônica e as partes do repositório a partir de qualquer texto ou link
func ExtractGitHubRepo(rawText string) (cleanURL, owner, repo string) {
	if rawText == "" {
		return "", "", ""
	}
	matches := githubURLRegex.FindStringSubmatch(rawText)
	if len(matches) < 3 {
		return "", "", ""
	}
	owner = matches[1]
	repo = strings.TrimSuffix(matches[2], ".git")
	cleanURL = fmt.Sprintf("https://github.com/%s/%s", owner, repo)
	return cleanURL, owner, repo
}

// FetchGitHubRepo baixa e descompacta concorrentemente o repositório público do GitHub, catalogando códigos e extraindo snippet
func (c *CanvasClient) FetchGitHubRepo(rawURL, destDir string) (*GitHubRepoResult, error) {
	cleanURL, owner, repo := ExtractGitHubRepo(rawURL)
	if cleanURL == "" {
		return nil, fmt.Errorf("URL do GitHub inválida: %s", rawURL)
	}

	if destDir == "" {
		destDir = filepath.Join("scratch", "github_repos", fmt.Sprintf("%s_%s", owner, repo))
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return nil, fmt.Errorf("erro ao criar pasta destino %s: %w", destDir, err)
	}

	archiveURL := fmt.Sprintf("https://github.com/%s/%s/archive/HEAD.zip", owner, repo)
	req, err := http.NewRequest("GET", archiveURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "AfyaCanvasBot/1.0")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao baixar repositório %s: %w", cleanURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		altURL := fmt.Sprintf("https://github.com/%s/%s/archive/refs/heads/main.zip", owner, repo)
		altReq, _ := http.NewRequest("GET", altURL, nil)
		altReq.Header.Set("User-Agent", "AfyaCanvasBot/1.0")
		resp, err = c.HTTPClient.Do(altReq)
		if err != nil || resp.StatusCode != http.StatusOK {
			if resp != nil {
				resp.Body.Close()
			}
			altURL2 := fmt.Sprintf("https://github.com/%s/%s/archive/refs/heads/master.zip", owner, repo)
			altReq2, _ := http.NewRequest("GET", altURL2, nil)
			altReq2.Header.Set("User-Agent", "AfyaCanvasBot/1.0")
			resp, err = c.HTTPClient.Do(altReq2)
			if err != nil || resp.StatusCode != http.StatusOK {
				if resp != nil {
					resp.Body.Close()
				}
				return nil, fmt.Errorf("não foi possível baixar o repositório GitHub %s (HTTP %d)", cleanURL, resp.StatusCode)
			}
		}
	}

	zipBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler zip do repositório: %w", err)
	}

	zipReader, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir zip do repositório: %w", err)
	}

	totalFiles := 0
	var codeFiles []string

	for _, file := range zipReader.File {
		parts := strings.Split(file.Name, "/")
		if len(parts) <= 1 {
			continue
		}
		relPath := strings.Join(parts[1:], "/")
		if relPath == "" {
			continue
		}

		if strings.HasPrefix(relPath, ".git") || strings.Contains(relPath, "/.git") ||
			strings.Contains(relPath, "node_modules/") || strings.Contains(relPath, "target/") ||
			strings.Contains(relPath, "__pycache__/") {
			continue
		}

		targetPath := filepath.Join(destDir, relPath)

		if file.FileInfo().IsDir() {
			_ = os.MkdirAll(targetPath, 0755)
			continue
		}

		_ = os.MkdirAll(filepath.Dir(targetPath), 0755)
		outFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			continue
		}

		rc, err := file.Open()
		if err != nil {
			outFile.Close()
			continue
		}

		_, _ = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()

		totalFiles++

		ext := strings.ToLower(filepath.Ext(relPath))
		if sourceFileExts[ext] {
			codeFiles = append(codeFiles, targetPath)
		}
	}

	// Se não encontrou código-fonte ou só encontrou docs (.md/.txt), verificar se o aluno comitou um .zip dentro do repositório
	hasRealCode := false
	for _, f := range codeFiles {
		ext := strings.ToLower(filepath.Ext(f))
		if ext != ".md" && ext != ".txt" {
			hasRealCode = true
			break
		}
	}

	if !hasRealCode {
		_ = filepath.Walk(destDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			if strings.ToLower(filepath.Ext(path)) == ".zip" {
				unpackedZipDir := filepath.Join(filepath.Dir(path), "unpacked_"+sanitizeNameRegex.ReplaceAllString(strings.TrimSuffix(info.Name(), ".zip"), "_"))
				zr, zErr := zip.OpenReader(path)
				if zErr == nil {
					defer zr.Close()
					for _, zf := range zr.File {
						if zf.FileInfo().IsDir() {
							continue
						}
						// Ignorar lixos de compilação ou arquivos ocultos
						if strings.Contains(zf.Name, "__MACOSX") || strings.Contains(zf.Name, "/target/") || strings.HasPrefix(filepath.Base(zf.Name), ".") {
							continue
						}
						zExt := strings.ToLower(filepath.Ext(zf.Name))
						if sourceFileExts[zExt] {
							zDest := filepath.Join(unpackedZipDir, zf.Name)
							_ = os.MkdirAll(filepath.Dir(zDest), 0755)
							if zOut, zOutErr := os.OpenFile(zDest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, zf.Mode()); zOutErr == nil {
								if zrc, zrcErr := zf.Open(); zrcErr == nil {
									_, _ = io.Copy(zOut, zrc)
									zrc.Close()
								}
								zOut.Close()
								codeFiles = append(codeFiles, zDest)
								totalFiles++
							}
						}
					}
				}
			}
			return nil
		})
	}

	primaryFile := ""
	for _, f := range codeFiles {
		base := strings.ToLower(filepath.Base(f))
		if strings.HasPrefix(base, "main.") || base == "app.py" || base == "index.js" ||
			strings.HasPrefix(base, "sort") || strings.HasPrefix(base, "ordenacao") {
			primaryFile = f
			break
		}
	}
	if primaryFile == "" && len(codeFiles) > 0 {
		for _, f := range codeFiles {
			ext := strings.ToLower(filepath.Ext(f))
			if ext == ".c" || ext == ".py" || ext == ".rs" || ext == ".cpp" || ext == ".java" {
				primaryFile = f
				break
			}
		}
		if primaryFile == "" {
			primaryFile = codeFiles[0]
		}
	}

	snippet := ""
	if primaryFile != "" {
		if content, err := os.ReadFile(primaryFile); err == nil {
			lines := strings.Split(string(content), "\n")
			maxLines := 80
			if len(lines) < maxLines {
				maxLines = len(lines)
			}
			snippet = strings.Join(lines[:maxLines], "\n")
			if len(snippet) > 3500 {
				snippet = snippet[:3500] + "\n... [trecho truncado pelo MCP para economia de tokens]"
			}
		}
	}

	return &GitHubRepoResult{
		GitHubURL:   cleanURL,
		Owner:       owner,
		Repo:        repo,
		LocalDir:    destDir,
		CodeFiles:   codeFiles,
		PrimaryFile: primaryFile,
		Snippet:     snippet,
		TotalFiles:  totalFiles,
	}, nil
}

type PreparedItem struct {
	UserID          string   `json:"user_id"`
	UserName        string   `json:"user_name"`
	SubmissionType  string   `json:"submission_type"`
	LocalCodeFile   string   `json:"local_code_file,omitempty"`
	DownloadedFiles []string `json:"downloaded_files,omitempty"`
	GitHubURL       string   `json:"github_url,omitempty"`
	LocalRepoDir    string   `json:"local_repo_dir,omitempty"`
	RepoCodeFiles   []string `json:"repo_code_files,omitempty"`
	PrimaryFile     string   `json:"primary_file,omitempty"`
	CodeSnippet     string   `json:"code_snippet,omitempty"`
	HasCode         bool     `json:"has_code"`
	WorkflowState   string   `json:"workflow_state"`
	Grade           string   `json:"grade,omitempty"`
	SubmittedAtBR   string   `json:"submitted_at_br"`
}

type PrepareResult struct {
	CourseID                  string         `json:"course_id"`
	AssignmentID              string         `json:"assignment_id"`
	AssignmentName            string         `json:"assignment_name"`
	TotalSubmissions          int            `json:"total_submissions"`
	SavedCodeFilesCount       int            `json:"saved_code_files_count"`
	DownloadedAttachmentCount int            `json:"downloaded_attachment_count"`
	DownloadedReposCount      int            `json:"downloaded_repos_count"`
	CodeDirectory             string         `json:"code_directory"`
	AttachmentsDirectory      string         `json:"attachments_directory"`
	GitHubReposDirectory      string         `json:"github_repos_directory,omitempty"`
	ManifestPath              string         `json:"manifest_path"`
	AssignmentDetailsPath     string         `json:"assignment_details_path"`
	Items                     []PreparedItem `json:"items"`
	Errors                    []string       `json:"errors,omitempty"`
}

// PrepareAssignment baixa enunciado, prepara códigos colados, repositórios GitHub e anexos concorrentemente
func (c *CanvasClient) PrepareAssignment(courseID, assignmentID, outputDir string, onlyPending bool) (*PrepareResult, error) {
	if outputDir == "" {
		outputDir = "scratch"
	}

	assignData, err := c.GetAssignment(courseID, assignmentID)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter detalhes da tarefa: %w", err)
	}
	assignMap, _ := assignData.(map[string]any)
	assignName := fmt.Sprintf("%v", assignMap["name"])

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("erro ao criar pasta destino: %w", err)
	}

	// 1. Salvar assignment.json
	assignJSONPath := filepath.Join(outputDir, "assignment.json")
	assignBytes, _ := json.MarshalIndent(assignData, "", "  ")
	_ = os.WriteFile(assignJSONPath, assignBytes, 0644)

	// 2. Buscar submissões
	submissions, err := c.GetSubmissionsDetails(courseID, assignmentID, onlyPending)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar submissões: %w", err)
	}

	attachmentsDir := filepath.Join(outputDir, "attachments")
	codeDir := filepath.Join(outputDir, "submissions_code")
	githubDir := filepath.Join(outputDir, "github_repos")
	_ = os.MkdirAll(attachmentsDir, 0755)
	_ = os.MkdirAll(codeDir, 0755)

	type downloadTask struct {
		url      string
		destPath string
		userIdx  int
		fileName string
	}

	type gitHubTask struct {
		url     string
		destDir string
		userIdx int
	}

	var downloadQueue []downloadTask
	var githubQueue []gitHubTask
	var items []PreparedItem
	savedCodeCount := 0

	for idx, s := range submissions {
		cleanName := sanitizeNameRegex.ReplaceAllString(s.UserName, "_")
		pItem := PreparedItem{
			UserID:         s.UserID,
			UserName:       s.UserName,
			SubmissionType: s.SubmissionType,
			HasCode:        s.HasCode,
			WorkflowState:  s.WorkflowState,
			Grade:          s.Grade,
			SubmittedAtBR:  s.SubmittedAtBR,
		}

		// A. Código colado no body
		if s.CleanBody != "" {
			ext := ".c"
			if strings.Contains(s.CleanBody, "def ") || (strings.Contains(s.CleanBody, "import ") && !strings.Contains(s.CleanBody, "#include")) {
				ext = ".py"
			}
			codeFileName := fmt.Sprintf("%s_%s%s", s.UserID, cleanName, ext)
			codeFilePath := filepath.Join(codeDir, codeFileName)
			if err := os.WriteFile(codeFilePath, []byte(s.CleanBody), 0644); err == nil {
				pItem.LocalCodeFile = codeFilePath
				savedCodeCount++
			}
		}

		// B. Anexos
		for _, att := range s.Attachments {
			attURL, _ := att["url"].(string)
			rawFilename, _ := att["filename"].(string)
			if rawFilename == "" {
				rawFilename = "anexo"
			}
			cleanFileName := sanitizeNameRegex.ReplaceAllString(rawFilename, "_")
			localFileName := fmt.Sprintf("%s_%s", s.UserID, cleanFileName)
			localDestPath := filepath.Join(attachmentsDir, localFileName)

			downloadQueue = append(downloadQueue, downloadTask{
				url:      attURL,
				destPath: localDestPath,
				userIdx:  idx,
				fileName: rawFilename,
			})
		}

		// C. Detecção inteligente de GitHub (em URL ou Body)
		ghURL, _, _ := ExtractGitHubRepo(s.URL)
		if ghURL == "" && s.CleanBody != "" {
			ghURL, _, _ = ExtractGitHubRepo(s.CleanBody)
		}
		if ghURL == "" && s.Body != "" {
			ghURL, _, _ = ExtractGitHubRepo(s.Body)
		}
		if ghURL != "" {
			pItem.GitHubURL = ghURL
			repoDestDir := filepath.Join(githubDir, fmt.Sprintf("%s_%s", s.UserID, cleanName))
			githubQueue = append(githubQueue, gitHubTask{
				url:     ghURL,
				destDir: repoDestDir,
				userIdx: idx,
			})
		}

		items = append(items, pItem)
	}

	var errorsList []string
	var mu sync.Mutex
	downloadedCount := 0

	// 3. Download concorrente de anexos com worker pool de 8 goroutines
	if len(downloadQueue) > 0 {
		var wg sync.WaitGroup
		taskChan := make(chan downloadTask, len(downloadQueue))

		workerCount := 8
		if len(downloadQueue) < workerCount {
			workerCount = len(downloadQueue)
		}

		for w := 0; w < workerCount; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for task := range taskChan {
					if task.url == "" {
						continue
					}
					_, err := c.DownloadAttachment(task.url, task.destPath)
					mu.Lock()
					if err != nil {
						errorsList = append(errorsList, fmt.Sprintf("Erro ao baixar anexo %s: %v", task.fileName, err))
					} else {
						downloadedCount++
						items[task.userIdx].DownloadedFiles = append(items[task.userIdx].DownloadedFiles, task.destPath)
					}
					mu.Unlock()
				}
			}()
		}

		for _, task := range downloadQueue {
			taskChan <- task
		}
		close(taskChan)
		wg.Wait()
	}

	// 4. Download concorrente de repositórios GitHub com worker pool de 4 goroutines
	downloadedReposCount := 0
	if len(githubQueue) > 0 {
		_ = os.MkdirAll(githubDir, 0755)
		var ghWg sync.WaitGroup
		ghChan := make(chan gitHubTask, len(githubQueue))

		ghWorkerCount := 4
		if len(githubQueue) < ghWorkerCount {
			ghWorkerCount = len(githubQueue)
		}

		for w := 0; w < ghWorkerCount; w++ {
			ghWg.Add(1)
			go func() {
				defer ghWg.Done()
				for task := range ghChan {
					repoRes, err := c.FetchGitHubRepo(task.url, task.destDir)
					mu.Lock()
					if err != nil {
						errorsList = append(errorsList, fmt.Sprintf("Erro ao baixar repositório GitHub %s: %v", task.url, err))
					} else {
						downloadedReposCount++
						items[task.userIdx].LocalRepoDir = repoRes.LocalDir
						items[task.userIdx].RepoCodeFiles = repoRes.CodeFiles
						items[task.userIdx].PrimaryFile = repoRes.PrimaryFile
						items[task.userIdx].CodeSnippet = repoRes.Snippet
						items[task.userIdx].HasCode = true
					}
					mu.Unlock()
				}
			}()
		}

		for _, task := range githubQueue {
			ghChan <- task
		}
		close(ghChan)
		ghWg.Wait()
	}

	// 5. Salvar prepared_submissions.json
	manifestPath := filepath.Join(outputDir, "prepared_submissions.json")
	result := &PrepareResult{
		CourseID:                  courseID,
		AssignmentID:              assignmentID,
		AssignmentName:            assignName,
		TotalSubmissions:          len(submissions),
		SavedCodeFilesCount:       savedCodeCount,
		DownloadedAttachmentCount: downloadedCount,
		DownloadedReposCount:      downloadedReposCount,
		CodeDirectory:             codeDir,
		AttachmentsDirectory:      attachmentsDir,
		GitHubReposDirectory:      githubDir,
		ManifestPath:              manifestPath,
		AssignmentDetailsPath:     assignJSONPath,
		Items:                     items,
		Errors:                    errorsList,
	}

	manifestBytes, _ := json.MarshalIndent(result, "", "  ")
	_ = os.WriteFile(manifestPath, manifestBytes, 0644)

	return result, nil
}

type ValidateGradesResult struct {
	Valid          bool           `json:"valid"`
	PointsPossible float64        `json:"points_possible"`
	TotalSubmitted int            `json:"total_submitted"`
	Errors         []string       `json:"errors,omitempty"`
	Warnings       []string       `json:"warnings,omitempty"`
	Stats          map[string]any `json:"stats"`
	ReviewMarkdown string         `json:"review_markdown"`
}

// ValidateGrades valida notas contra a pontuação da tarefa, checa alunos matriculados e gera a tabela Markdown
func (c *CanvasClient) ValidateGrades(courseID, assignmentID string, grades []GradeEntry) (*ValidateGradesResult, error) {
	assignData, err := c.GetAssignment(courseID, assignmentID)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter tarefa: %w", err)
	}
	assignMap, _ := assignData.(map[string]any)
	assignName := fmt.Sprintf("%v", assignMap["name"])
	ptsPossible := 0.0
	if p, ok := assignMap["points_possible"].(float64); ok {
		ptsPossible = p
	}

	// Buscar alunos para validar IDs e recuperar nomes
	studentsData, err := c.ListStudents(courseID)
	studentMap := make(map[string]string)
	if err == nil {
		if sList, ok := studentsData.([]any); ok {
			for _, s := range sList {
				if sm, ok := s.(map[string]any); ok {
					sID := fmt.Sprintf("%v", sm["id"])
					sName := fmt.Sprintf("%v", sm["name"])
					studentMap[sID] = sName
				}
			}
		}
	}

	var errorsList []string
	var warningsList []string
	var validGrades []float64

	var mdLines []string
	mdLines = append(mdLines, fmt.Sprintf("### Revisão de Notas: %s (ID: %s)", assignName, assignmentID))
	mdLines = append(mdLines, fmt.Sprintf("**Pontuação Máxima:** %.1f pontos | **Total de Alunos Avaliados:** %d\n", ptsPossible, len(grades)))
	mdLines = append(mdLines, "| Aluno | ID | Nota Sugerida / Máxima | Resumo do Feedback |")
	mdLines = append(mdLines, "| :--- | :--- | :--- | :--- |")

	for _, g := range grades {
		name := studentMap[g.UserID]
		if name == "" {
			name = "Aluno Desconhecido"
			warningsList = append(warningsList, fmt.Sprintf("User ID %s não encontrado na lista oficial de matriculados", g.UserID))
		}

		gVal, err := strconv.ParseFloat(g.Grade, 64)
		if err != nil {
			errorsList = append(errorsList, fmt.Sprintf("Nota inválida para aluno %s (%s): %s", name, g.UserID, g.Grade))
		} else {
			if ptsPossible > 0 && gVal > ptsPossible {
				errorsList = append(errorsList, fmt.Sprintf("Nota %.1f excede a pontuação máxima (%.1f) para %s", gVal, ptsPossible, name))
			}
			if gVal < 0 {
				errorsList = append(errorsList, fmt.Sprintf("Nota %.1f não pode ser negativa para %s", gVal, name))
			}
			validGrades = append(validGrades, gVal)
		}

		cleanComment := strings.ReplaceAll(g.Comment, "\n", " ")
		if len(cleanComment) > 120 {
			cleanComment = cleanComment[:117] + "..."
		}
		mdLines = append(mdLines, fmt.Sprintf("| %s | %s | %s / %.0f | %s |", name, g.UserID, g.Grade, ptsPossible, cleanComment))
	}

	stats := make(map[string]any)
	if len(validGrades) > 0 {
		sum := 0.0
		min := validGrades[0]
		max := validGrades[0]
		for _, v := range validGrades {
			sum += v
			if v < min {
				min = v
			}
			if v > max {
				max = v
			}
		}
		stats["media"] = sum / float64(len(validGrades))
		stats["menor_nota"] = min
		stats["maior_nota"] = max
		stats["avaliados"] = len(validGrades)
	}

	return &ValidateGradesResult{
		Valid:          len(errorsList) == 0,
		PointsPossible: ptsPossible,
		TotalSubmitted: len(grades),
		Errors:         errorsList,
		Warnings:       warningsList,
		Stats:          stats,
		ReviewMarkdown: strings.Join(mdLines, "\n"),
	}, nil
}

type UnpackedFileInfo struct {
	UserID       string `json:"user_id,omitempty"`
	StudentName  string `json:"student_name,omitempty"`
	OriginalName string `json:"original_name"`
	ExtractedTo  string `json:"extracted_to"`
}

// UnpackSubmissionsZip descompacta arquivos ZIP exportados do Canvas SpeedGrader mapeando para os alunos
func (c *CanvasClient) UnpackSubmissionsZip(zipPath, courseID, outputDir string) ([]UnpackedFileInfo, error) {
	if outputDir == "" {
		outputDir = "scratch/unpacked"
	}

	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("falha ao abrir ZIP: %w", err)
	}
	defer reader.Close()

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("erro ao criar pasta destino: %w", err)
	}

	var extracted []UnpackedFileInfo

	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}

		cleanFileName := filepath.Base(file.Name)
		destPath := filepath.Join(outputDir, cleanFileName)

		rc, err := file.Open()
		if err != nil {
			continue
		}

		outFile, err := os.Create(destPath)
		if err != nil {
			rc.Close()
			continue
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err == nil {
			extracted = append(extracted, UnpackedFileInfo{
				OriginalName: file.Name,
				ExtractedTo:  destPath,
			})
		}
	}

	return extracted, nil
}
