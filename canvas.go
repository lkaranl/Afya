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
// GESTÃO E CRIAÇÃO DE CONTEÚDO (ATIVIDADES, QUIZZES E MÓDULOS)
// -------------------------------------------------------------

type CreateAssignmentParams struct {
	CourseID          string   `json:"course_id"`
	Name              string   `json:"name"`
	Description       string   `json:"description"`                 // Enunciado em HTML
	PointsPossible    float64  `json:"points_possible"`             // Pontos (ex: 100.0)
	SubmissionTypes   []string `json:"submission_types,omitempty"`  // ["online_url"], ["online_upload"], etc.
	DueAt             string   `json:"due_at,omitempty"`            // Prazo (ISO UTC, ex: "2026-09-25T02:59:59Z")
	UnlockAt          string   `json:"unlock_at,omitempty"`         // Data de abertura
	LockAt            string   `json:"lock_at,omitempty"`           // Data de encerramento
	GroupCategoryID   string   `json:"group_category_id,omitempty"`
	Published         *bool    `json:"published,omitempty"`         // default: true
	AllowedExtensions []string `json:"allowed_extensions,omitempty"` // ex: ["c", "h", "zip"]
}

func (c *CanvasClient) CreateAssignment(p CreateAssignmentParams) (any, error) {
	if p.CourseID == "" || p.Name == "" {
		return nil, fmt.Errorf("course_id e name são obrigatórios para criar atividade")
	}

	endpoint := fmt.Sprintf("/api/v1/courses/%s/assignments", url.PathEscape(p.CourseID))
	assignMap := map[string]any{
		"name":            p.Name,
		"description":     p.Description,
		"points_possible": p.PointsPossible,
	}

	if len(p.SubmissionTypes) > 0 {
		assignMap["submission_types"] = p.SubmissionTypes
	} else {
		assignMap["submission_types"] = []string{"online_url"}
	}

	if p.DueAt != "" {
		assignMap["due_at"] = p.DueAt
	}
	if p.UnlockAt != "" {
		assignMap["unlock_at"] = p.UnlockAt
	}
	if p.LockAt != "" {
		assignMap["lock_at"] = p.LockAt
	}
	if p.GroupCategoryID != "" {
		assignMap["group_category_id"] = p.GroupCategoryID
	}
	if p.Published != nil {
		assignMap["published"] = *p.Published
	} else {
		assignMap["published"] = true
	}
	if len(p.AllowedExtensions) > 0 {
		assignMap["allowed_extensions"] = p.AllowedExtensions
	}

	payload := map[string]any{
		"assignment": assignMap,
	}

	data, _, err := c.Request("POST", endpoint, payload)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar atividade no Canvas: %w", err)
	}

	var result any
	err = json.Unmarshal(data, &result)
	return result, err
}

type QuizAnswer struct {
	Text             string `json:"text,omitempty"`
	Weight           int    `json:"weight,omitempty"` // 100 para correta, 0 para incorreta
	Comment          string `json:"comment,omitempty"`
	BlankID          string `json:"blank_id,omitempty"`           // Para fill_in_multiple_blanks_question
	AnswerMatchLeft  string `json:"answer_match_left,omitempty"`  // Para matching_question
	AnswerMatchRight string `json:"answer_match_right,omitempty"` // Para matching_question
}

type QuizQuestion struct {
	Title          string       `json:"title,omitempty"`
	Text           string       `json:"text"`                      // Enunciado da questão (HTML)
	Type           string       `json:"type,omitempty"`            // default: "multiple_choice_question"
	PointsPossible float64      `json:"points_possible,omitempty"` // default: 10
	Answers        []QuizAnswer `json:"answers"`
}

type CreateQuizParams struct {
	CourseID        string         `json:"course_id"`
	Title           string         `json:"title"`
	Description     string         `json:"description,omitempty"`
	QuizType        string         `json:"quiz_type,omitempty"`  // default: "assignment"
	TimeLimit       int            `json:"time_limit,omitempty"` // minutos
	ShuffleAnswers  *bool          `json:"shuffle_answers,omitempty"`
	AllowedAttempts int            `json:"allowed_attempts,omitempty"`
	DueAt           string         `json:"due_at,omitempty"`
	Published       *bool          `json:"published,omitempty"`
	Questions       []QuizQuestion `json:"questions,omitempty"`
}

func (c *CanvasClient) CreateQuiz(p CreateQuizParams) (any, error) {
	if p.CourseID == "" || p.Title == "" {
		return nil, fmt.Errorf("course_id e title são obrigatórios para criar quiz")
	}

	endpoint := fmt.Sprintf("/api/v1/courses/%s/quizzes", url.PathEscape(p.CourseID))
	quizMap := map[string]any{
		"title":       p.Title,
		"description": p.Description,
	}

	if p.QuizType != "" {
		quizMap["quiz_type"] = p.QuizType
	} else {
		quizMap["quiz_type"] = "assignment"
	}

	if p.TimeLimit > 0 {
		quizMap["time_limit"] = p.TimeLimit
	}
	if p.ShuffleAnswers != nil {
		quizMap["shuffle_answers"] = *p.ShuffleAnswers
	} else {
		quizMap["shuffle_answers"] = true
	}
	if p.AllowedAttempts > 0 {
		quizMap["allowed_attempts"] = p.AllowedAttempts
	}
	if p.DueAt != "" {
		quizMap["due_at"] = p.DueAt
	}
	if p.Published != nil {
		quizMap["published"] = *p.Published
	} else {
		quizMap["published"] = true
	}

	payload := map[string]any{
		"quiz": quizMap,
	}

	data, _, err := c.Request("POST", endpoint, payload)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar quiz no Canvas: %w", err)
	}

	var quizRes map[string]any
	if err := json.Unmarshal(data, &quizRes); err != nil {
		return nil, err
	}

	quizIDRaw, ok := quizRes["id"]
	if !ok {
		return quizRes, nil
	}
	quizID := fmt.Sprintf("%v", quizIDRaw)

	createdQuestions := 0
	var questionErrors []string

	for idx, q := range p.Questions {
		qTitle := q.Title
		if qTitle == "" {
			qTitle = fmt.Sprintf("Questão %d", idx+1)
		}
		qType := q.Type
		if qType == "" {
			qType = "multiple_choice_question"
		}
		qPoints := q.PointsPossible
		if qPoints == 0 {
			qPoints = 10.0
		}

		var answersList []map[string]any
		for _, a := range q.Answers {
			ansMap := map[string]any{}
			if a.AnswerMatchLeft != "" || a.AnswerMatchRight != "" {
				ansMap["answer_match_left"] = a.AnswerMatchLeft
				ansMap["answer_match_right"] = a.AnswerMatchRight
			} else {
				ansMap["answer_text"] = a.Text
				ansMap["answer_weight"] = a.Weight
				if a.BlankID != "" {
					ansMap["blank_id"] = a.BlankID
				}
				if a.Comment != "" {
					ansMap["answer_comment"] = a.Comment
				}
			}
			answersList = append(answersList, ansMap)
		}

		qPayload := map[string]any{
			"question": map[string]any{
				"question_name":   qTitle,
				"question_text":   q.Text,
				"question_type":   qType,
				"points_possible": qPoints,
				"answers":         answersList,
			},
		}

		qEndpoint := fmt.Sprintf("/api/v1/courses/%s/quizzes/%s/questions", url.PathEscape(p.CourseID), url.PathEscape(quizID))
		_, _, qErr := c.Request("POST", qEndpoint, qPayload)
		if qErr != nil {
			questionErrors = append(questionErrors, fmt.Sprintf("Erro na questão %d (%s): %v", idx+1, qTitle, qErr))
		} else {
			createdQuestions++
		}
	}

	return map[string]any{
		"quiz_id":                 quizID,
		"title":                   quizRes["title"],
		"html_url":                quizRes["html_url"],
		"published":               quizRes["published"],
		"total_questions_created": createdQuestions,
		"questions_errors":        questionErrors,
	}, nil
}

type CreateModuleParams struct {
	CourseID string                `json:"course_id"`
	Name     string                `json:"name"`
	Position int                   `json:"position,omitempty"`
	UnlockAt string                `json:"unlock_at,omitempty"`
	Items    []AddModuleItemParams `json:"items,omitempty"` // Cria e vincula itens automaticamente
}

func (c *CanvasClient) CreateModule(p CreateModuleParams) (any, error) {
	if p.CourseID == "" || p.Name == "" {
		return nil, fmt.Errorf("course_id e name são obrigatórios para criar módulo")
	}

	endpoint := fmt.Sprintf("/api/v1/courses/%s/modules", url.PathEscape(p.CourseID))
	modMap := map[string]any{
		"name": p.Name,
	}
	if p.Position > 0 {
		modMap["position"] = p.Position
	}
	if p.UnlockAt != "" {
		modMap["unlock_at"] = p.UnlockAt
	}

	payload := map[string]any{
		"module": modMap,
	}

	data, _, err := c.Request("POST", endpoint, payload)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar módulo no Canvas: %w", err)
	}

	var modRes map[string]any
	if err := json.Unmarshal(data, &modRes); err != nil {
		return nil, err
	}

	modIDRaw, ok := modRes["id"]
	if !ok {
		return modRes, nil
	}
	modID := fmt.Sprintf("%v", modIDRaw)

	var createdItems []any
	var itemErrors []string

	for _, item := range p.Items {
		item.CourseID = p.CourseID
		item.ModuleID = modID
		itemRes, itemErr := c.AddModuleItem(item)
		if itemErr != nil {
			itemErrors = append(itemErrors, fmt.Sprintf("Erro ao adicionar item %s: %v", item.Title, itemErr))
		} else {
			createdItems = append(createdItems, itemRes)
		}
	}

	modRes["items_created"] = createdItems
	if len(itemErrors) > 0 {
		modRes["items_errors"] = itemErrors
	}

	return modRes, nil
}

type CreatePageParams struct {
	CourseID  string `json:"course_id"`
	Title     string `json:"title"`
	Body      string `json:"body"` // HTML da página
	Published *bool  `json:"published,omitempty"`
}

func (c *CanvasClient) CreatePage(p CreatePageParams) (any, error) {
	if p.CourseID == "" || p.Title == "" {
		return nil, fmt.Errorf("course_id e title são obrigatórios para criar página")
	}

	endpoint := fmt.Sprintf("/api/v1/courses/%s/pages", url.PathEscape(p.CourseID))
	pageMap := map[string]any{
		"title": p.Title,
		"body":  p.Body,
	}
	if p.Published != nil {
		pageMap["published"] = *p.Published
	} else {
		pageMap["published"] = true
	}

	payload := map[string]any{
		"wiki_page": pageMap,
	}

	data, _, err := c.Request("POST", endpoint, payload)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar página no Canvas: %w", err)
	}

	var result any
	err = json.Unmarshal(data, &result)
	return result, err
}

type AddModuleItemParams struct {
	CourseID    string `json:"course_id"`
	ModuleID    string `json:"module_id"`
	Title       string `json:"title,omitempty"`
	Type        string `json:"type"` // "Assignment", "Quiz", "ExternalUrl", "Page", "SubHeader"
	ContentID   string `json:"content_id,omitempty"`
	PageURL     string `json:"page_url,omitempty"`   // URL slug da página (obrigatório se type for Page)
	ExternalURL string `json:"external_url,omitempty"`
	Position    int    `json:"position,omitempty"`
	NewTab      bool   `json:"new_tab,omitempty"`
}

func (c *CanvasClient) AddModuleItem(p AddModuleItemParams) (any, error) {
	if p.CourseID == "" || p.ModuleID == "" || p.Type == "" {
		return nil, fmt.Errorf("course_id, module_id e type são obrigatórios para adicionar item ao módulo")
	}

	endpoint := fmt.Sprintf("/api/v1/courses/%s/modules/%s/items", url.PathEscape(p.CourseID), url.PathEscape(p.ModuleID))
	itemMap := map[string]any{
		"type": p.Type,
	}
	if p.Title != "" {
		itemMap["title"] = p.Title
	}
	if p.ContentID != "" {
		itemMap["content_id"] = p.ContentID
	}
	if p.PageURL != "" {
		itemMap["page_url"] = p.PageURL
	}
	if p.ExternalURL != "" {
		itemMap["external_url"] = p.ExternalURL
		itemMap["new_tab"] = p.NewTab
	}
	if p.Position > 0 {
		itemMap["position"] = p.Position
	}

	payload := map[string]any{
		"module_item": itemMap,
	}

	data, _, err := c.Request("POST", endpoint, payload)
	if err != nil {
		return nil, fmt.Errorf("erro ao adicionar item ao módulo no Canvas: %w", err)
	}

	var result any
	err = json.Unmarshal(data, &result)
	return result, err
}

func (c *CanvasClient) ListModules(courseID string) (any, error) {
	endpoint := fmt.Sprintf("/api/v1/courses/%s/modules?include[]=items&per_page=50", url.PathEscape(courseID))
	data, _, err := c.Request("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	var result any
	err = json.Unmarshal(data, &result)
	return result, err
}

// -------------------------------------------------------------
// GESTÃO DE GRUPOS DE NOTAS PONDERADAS (ASSIGNMENT GROUPS)
// -------------------------------------------------------------

type AssignmentGroupItem struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	GroupWeight float64 `json:"group_weight"`
	Position    int     `json:"position,omitempty"`
	Assignments []any   `json:"assignments,omitempty"`
}

type SetupGradingSchemeParams struct {
	CourseID      string `json:"course_id"`
	EnableWeights bool   `json:"enable_weights"`
}

// ListAssignmentGroups obtém todos os grupos de tarefas e suas ponderações
func (c *CanvasClient) ListAssignmentGroups(courseID string) (any, error) {
	endpoint := fmt.Sprintf("/api/v1/courses/%s/assignment_groups?include[]=assignments", url.PathEscape(courseID))
	data, _, err := c.Request("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	var result any
	err = json.Unmarshal(data, &result)
	return result, err
}

// SetCourseWeighting ativa ou desativa o cálculo ponderado por grupos no curso
func (c *CanvasClient) SetCourseWeighting(courseID string, enable bool) (any, error) {
	endpoint := fmt.Sprintf("/api/v1/courses/%s", url.PathEscape(courseID))
	payload := map[string]any{
		"course": map[string]any{
			"apply_assignment_group_weights": enable,
		},
	}
	data, _, err := c.Request("PUT", endpoint, payload)
	if err != nil {
		return nil, err
	}
	var result any
	err = json.Unmarshal(data, &result)
	return result, err
}

// CreateAssignmentGroup cria um novo grupo de atividades com peso específico
func (c *CanvasClient) CreateAssignmentGroup(courseID, name string, weight float64) (any, error) {
	endpoint := fmt.Sprintf("/api/v1/courses/%s/assignment_groups", url.PathEscape(courseID))
	payload := map[string]any{
		"name":         name,
		"group_weight": weight,
	}
	data, _, err := c.Request("POST", endpoint, payload)
	if err != nil {
		return nil, err
	}
	var result any
	err = json.Unmarshal(data, &result)
	return result, err
}

// UpdateAssignmentGroup atualiza o nome e peso de um grupo de atividades
func (c *CanvasClient) UpdateAssignmentGroup(courseID, groupID, name string, weight float64) (any, error) {
	endpoint := fmt.Sprintf("/api/v1/courses/%s/assignment_groups/%s", url.PathEscape(courseID), url.PathEscape(groupID))
	payload := map[string]any{
		"name":         name,
		"group_weight": weight,
	}
	data, _, err := c.Request("PUT", endpoint, payload)
	if err != nil {
		return nil, err
	}
	var result any
	err = json.Unmarshal(data, &result)
	return result, err
}

// MoveAssignmentToGroup transfere uma atividade para um grupo de notas específico
func (c *CanvasClient) MoveAssignmentToGroup(courseID, assignmentID, groupID string) (any, error) {
	endpoint := fmt.Sprintf("/api/v1/courses/%s/assignments/%s", url.PathEscape(courseID), url.PathEscape(assignmentID))
	payload := map[string]any{
		"assignment": map[string]any{
			"assignment_group_id": groupID,
		},
	}
	data, _, err := c.Request("PUT", endpoint, payload)
	if err != nil {
		return nil, err
	}
	var result any
	err = json.Unmarshal(data, &result)
	return result, err
}

// -------------------------------------------------------------
// REGRAS E DIRETRIZES INSTITUCIONAIS AFYA (CONSEPE & NAPED 2026)
// -------------------------------------------------------------

type AfyaGroupRule struct {
	Name   string  `json:"name"`
	Weight float64 `json:"weight"`
	Points float64 `json:"points"`
	Notes  string  `json:"notes"`
}

type AfyaModalityRules struct {
	ModalityCode     string            `json:"modality_code"`
	ModalityName     string            `json:"modality_name"`
	Description      string            `json:"description"`
	Groups           []AfyaGroupRule   `json:"groups"`
	PassingGrade     float64           `json:"passing_grade"`
	MinAttendance    float64           `json:"min_attendance_percent"`
	ExamFinalRange   string            `json:"exam_final_range"`
	ExamFinalFormula string            `json:"exam_final_formula"`
	OfficialDeadlines map[string]string `json:"official_deadlines"`
}

// GetInstitutionalRules retorna a matriz de notas e critérios oficiais da Afya / São Lucas 2026
func (c *CanvasClient) GetInstitutionalRules(modality string) (any, error) {
	deadlines := map[string]string{
		"devolutiva_correcao_sala": "Até 10 dias após a aplicação da prova/atividade (Art. 16, § 2º).",
		"revisao_de_prova_aluno":   "Até 2 dias letivos após a devolutiva em sala (Art. 5º, § 1º e Art. 17, § 3º).",
		"prazo_professor_revisao":  "Até 7 dias após receber a notificação da Coordenação (Art. 17, § 4º).",
		"segunda_chamada_pedido":   "Até 72 horas após aplicação com atestado/justificativa legal (Art. 19).",
	}

	catalog := map[string]AfyaModalityRules{
		"presencial_sem_tpi": {
			ModalityCode: "PR_SEM_TPI",
			ModalityName: "Presencial - Cursos sem TPI (Ciência da Computação, Engenharias, etc.)",
			Description:  "Matriz 2022 e seguintes. N1 (50 pts) + N2 (50 pts) = 100 pontos totais.",
			Groups: []AfyaGroupRule{
				{Name: "N1 - Prova Escrita Individual", Weight: 30.0, Points: 30.0, Notes: "Sem consulta, modelo ENADE"},
				{Name: "N1 - Atividades Teóricas/Práticas", Weight: 20.0, Points: 20.0, Notes: "Trabalhos, exercícios práticos de programação"},
				{Name: "N2 - Prova Escrita Individual", Weight: 30.0, Points: 30.0, Notes: "Sem consulta, modelo ENADE"},
				{Name: "N2 - Atividades Teóricas/Práticas", Weight: 20.0, Points: 20.0, Notes: "Trabalhos, exercícios práticos de programação"},
			},
			PassingGrade:     70.0,
			MinAttendance:    75.0,
			ExamFinalRange:   "De 40.0 a 69.0 pontos (abaixo de 40.0 é reprovação direta sem exame)",
			ExamFinalFormula: "(Nota Semestral + Exame Final) / 2 >= 60.0 pontos",
			OfficialDeadlines: deadlines,
		},
		"presencial_com_tpi": {
			ModalityCode: "PR_COM_TPI",
			ModalityName: "Presencial - Cursos com TPI (Direito, Enfermagem, Fisioterapia, Psicologia)",
			Description:  "Matriz 2022 e seguintes. Inclui Teste de Progresso Institucional (TPI).",
			Groups: []AfyaGroupRule{
				{Name: "N1 - Prova Escrita Individual", Weight: 30.0, Points: 30.0, Notes: "Sem consulta, modelo ENADE"},
				{Name: "N1 - Atividades Teóricas/Práticas", Weight: 20.0, Points: 20.0, Notes: "Atividades em sala e práticas"},
				{Name: "N2 - Prova Escrita Individual", Weight: 20.0, Points: 20.0, Notes: "Sem consulta, modelo ENADE"},
				{Name: "N2 - Atividades Teóricas/Práticas", Weight: 20.0, Points: 20.0, Notes: "Atividades em sala e práticas"},
				{Name: "N2 - Teste de Progresso Institucional (TPI)", Weight: 10.0, Points: 10.0, Notes: "Aplicação institucional presencial"},
			},
			PassingGrade:     70.0,
			MinAttendance:    75.0,
			ExamFinalRange:   "De 40.0 a 69.0 pontos",
			ExamFinalFormula: "(Nota Semestral + Exame Final) / 2 >= 60.0 pontos",
			OfficialDeadlines: deadlines,
		},
		"hibrida_sem_tpi": {
			ModalityCode: "HB_SEM_TPI",
			ModalityName: "Híbrida (HB) - Cursos sem TPI",
			Description:  "Combina aulas presenciais com atividades online no Canvas.",
			Groups: []AfyaGroupRule{
				{Name: "N1 - Prova Escrita Presencial", Weight: 30.0, Points: 30.0, Notes: "Presencial, sem consulta"},
				{Name: "N1 - Atividade Teórica/Prática", Weight: 15.0, Points: 15.0, Notes: "Elaborada pelo professor"},
				{Name: "N2 - Prova Escrita Presencial", Weight: 30.0, Points: 30.0, Notes: "Presencial, sem consulta"},
				{Name: "N2 - Simulado Revisional Canvas", Weight: 10.0, Points: 10.0, Notes: "Autocorreção no Canvas, 1 tentativa"},
				{Name: "N2 - Atividade Presencial", Weight: 10.0, Points: 10.0, Notes: "Elaborada pelo professor"},
				{Name: "N2 - Atividade AVA Canvas", Weight: 5.0, Points: 5.0, Notes: "Escolhida pelo professor no Canvas"},
			},
			PassingGrade:     70.0,
			MinAttendance:    75.0,
			ExamFinalRange:   "De 40.0 a 69.0 pontos",
			ExamFinalFormula: "(Nota Semestral + Exame Final) / 2 >= 60.0 pontos",
			OfficialDeadlines: deadlines,
		},
		"online_assincrona": {
			ModalityCode: "ON_A",
			ModalityName: "Online Assíncrona (100% Online)",
			Description:  "Conteúdo integral no Canvas, sem aulas ao vivo. Prova em laboratório da IES.",
			Groups: []AfyaGroupRule{
				{Name: "N1 - Roteiro de Atividade 1", Weight: 25.0, Points: 25.0, Notes: "Envio pelo Canvas"},
				{Name: "N1 - Roteiro de Atividade 2", Weight: 25.0, Points: 25.0, Notes: "Envio pelo Canvas"},
				{Name: "N2 - Simulado para Avaliação Final", Weight: 10.0, Points: 10.0, Notes: "Autocorreção Canvas, tentativa única"},
				{Name: "N2 - Prova Presencial no Laboratório", Weight: 40.0, Points: 40.0, Notes: "20 questões via Canvas em laboratório"},
			},
			PassingGrade:     70.0,
			MinAttendance:    75.0,
			ExamFinalRange:   "De 40.0 a 69.0 pontos",
			ExamFinalFormula: "(Nota Semestral + Exame Final) / 2 >= 60.0 pontos",
			OfficialDeadlines: deadlines,
		},
		"simplificado_50_50": {
			ModalityCode: "SIMPLIFICADO_50_50",
			ModalityName: "Modelo Ponderado Contínuo (50% Atividades / 50% Prova)",
			Description:  "Ideal para turmas em andamento onde surgem múltiplas atividades práticas contínuas.",
			Groups: []AfyaGroupRule{
				{Name: "Atividades e Trabalhos Práticos", Weight: 50.0, Points: 100.0, Notes: "Compreende todas as listas, quizzes e projetos do semestre"},
				{Name: "Avaliação Oficial / Prova", Weight: 50.0, Points: 100.0, Notes: "Prova individual semestral"},
			},
			PassingGrade:     70.0,
			MinAttendance:    75.0,
			ExamFinalRange:   "De 40.0 a 69.0 pontos",
			ExamFinalFormula: "(Nota Semestral + Exame Final) / 2 >= 60.0 pontos",
			OfficialDeadlines: deadlines,
		},
	}

	if modality != "" {
		if rule, ok := catalog[strings.ToLower(strings.TrimSpace(modality))]; ok {
			return rule, nil
		}
		return nil, fmt.Errorf("modalidade desconhecida '%s'. Opções válidas: presencial_sem_tpi, presencial_com_tpi, hibrida_sem_tpi, online_assincrona, simplificado_50_50", modality)
	}

	return catalog, nil
}

// SetupAfyaGradingScheme configura os grupos ponderados do Canvas em conformidade com as regras institucionais
func (c *CanvasClient) SetupAfyaGradingScheme(courseID, modality string) (any, error) {
	rulesRaw, err := c.GetInstitutionalRules(modality)
	if err != nil {
		return nil, err
	}
	rule := rulesRaw.(AfyaModalityRules)

	// 1. Ativar ponderação no curso
	if _, err := c.SetCourseWeighting(courseID, true); err != nil {
		return nil, fmt.Errorf("falha ao ativar ponderação no curso: %w", err)
	}

	// 2. Listar grupos atuais
	groupsData, _, err := c.Request("GET", fmt.Sprintf("/api/v1/courses/%s/assignment_groups", url.PathEscape(courseID)), nil)
	if err != nil {
		return nil, fmt.Errorf("falha ao obter grupos atuais: %w", err)
	}
	var existingGroups []map[string]any
	json.Unmarshal(groupsData, &existingGroups)

	createdOrUpdated := []string{}

	// Reutilizar o primeiro grupo existente para o primeiro item da regra
	for i, gRule := range rule.Groups {
		if i < len(existingGroups) {
			targetID := fmt.Sprintf("%v", existingGroups[i]["id"])
			_, upErr := c.UpdateAssignmentGroup(courseID, targetID, gRule.Name, gRule.Weight)
			if upErr == nil {
				createdOrUpdated = append(createdOrUpdated, fmt.Sprintf("Grupo ID %s atualizado para: '%s' (Peso: %.1f%%)", targetID, gRule.Name, gRule.Weight))
			}
		} else {
			res, crErr := c.CreateAssignmentGroup(courseID, gRule.Name, gRule.Weight)
			if crErr == nil {
				createdOrUpdated = append(createdOrUpdated, fmt.Sprintf("Novo grupo criado: '%s' (Peso: %.1f%%)", gRule.Name, gRule.Weight))
				_ = res
			}
		}
	}

	return map[string]any{
		"course_id":          courseID,
		"modality_applied":   rule.ModalityName,
		"weights_enabled":    true,
		"actions_performed":  createdOrUpdated,
		"passing_grade":      rule.PassingGrade,
		"official_deadlines": rule.OfficialDeadlines,
	}, nil
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
