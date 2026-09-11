package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      any           `json:"id,omitempty"`
	Result  any           `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type MCPTool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InputSchema any    `json:"inputSchema"`
}

type MCPToolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type MCPToolCallResult struct {
	Content []MCPContentItem `json:"content"`
	IsError bool             `json:"isError,omitempty"`
}

type MCPContentItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

var mcpTools = []MCPTool{
	{
		Name:        "canvas_list_pending_assignments",
		Description: "Varre todas as disciplinas ativas do professor e retorna a lista de tarefas que possuem submissões de alunos aguardando correção (com prazos formatados em português, notas máximas e contagem de pendências).",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	},
	{
		Name:        "canvas_list_courses",
		Description: "Lista todas as disciplinas ativas do docente no Canvas LMS com seus respectivos IDs, nomes e total de alunos matriculados.",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	},
	{
		Name:        "canvas_list_assignments",
		Description: "Lista todas as tarefas/atividades cadastradas em uma disciplina específica com prazos, pontuação máxima e contagem de entregas pendentes.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"course_id": map[string]any{
					"type":        "string",
					"description": "ID numérico da disciplina no Canvas (ex: '162263')",
				},
			},
			"required": []string{"course_id"},
		},
	},
	{
		Name:        "canvas_get_assignment",
		Description: "Obtém os detalhes completos, enunciado oficial, rubrica e regras de uma atividade específica no Canvas LMS.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"course_id": map[string]any{
					"type":        "string",
					"description": "ID numérico da disciplina no Canvas",
				},
				"assignment_id": map[string]any{
					"type":        "string",
					"description": "ID numérico da atividade no Canvas (ex: '1374884')",
				},
			},
			"required": []string{"course_id", "assignment_id"},
		},
	},
	{
		Name:        "canvas_list_students",
		Description: "Lista todos os alunos matriculados em uma disciplina com IDs do Canvas, nomes completos e nomes ordenáveis.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"course_id": map[string]any{
					"type":        "string",
					"description": "ID numérico da disciplina no Canvas",
				},
			},
			"required": []string{"course_id"},
		},
	},
	{
		Name:        "canvas_get_submissions",
		Description: "Obtém as submissões dos estudantes para uma atividade com nome do aluno, código digitado higienizado sem tags HTML (clean_body), links de anexos e data formatada no padrão brasileiro.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"course_id": map[string]any{
					"type":        "string",
					"description": "ID numérico da disciplina no Canvas",
				},
				"assignment_id": map[string]any{
					"type":        "string",
					"description": "ID numérico da atividade no Canvas",
				},
				"only_pending": map[string]any{
					"type":        "boolean",
					"description": "Se verdadeiro, retorna apenas as submissões que ainda não receberam nota (padrão: true)",
				},
			},
			"required": []string{"course_id", "assignment_id"},
		},
	},
	{
		Name:        "canvas_get_grading_status",
		Description: "Diagnóstico e panorama consolidado de correções de uma disciplina ou atividade específica (percentual concluído, atividades 100% corrigidas, parciais, pendentes e lista de alunos sem nota). Substitui scripts externos de auditoria.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"course_id": map[string]any{
					"type":        "string",
					"description": "ID numérico da disciplina no Canvas (ex: '162263')",
				},
				"assignment_id": map[string]any{
					"type":        "string",
					"description": "ID numérico da atividade (opcional; se omitido, faz o panorama completo de todas as atividades do curso)",
				},
			},
			"required": []string{"course_id"},
		},
	},
	{
		Name:        "canvas_prepare_assignment",
		Description: "Prepara em lote de forma concorrente todo o ambiente de avaliação: baixa o enunciado oficial, extrai os códigos C/Python colados para arquivos locais e baixa todos os anexos em paralelo com Goroutines. Retorna o manifesto pronto.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"course_id": map[string]any{
					"type":        "string",
					"description": "ID numérico da disciplina no Canvas",
				},
				"assignment_id": map[string]any{
					"type":        "string",
					"description": "ID numérico da atividade no Canvas",
				},
				"output_dir": map[string]any{
					"type":        "string",
					"description": "Diretório de saída (padrão: 'scratch')",
				},
				"only_pending": map[string]any{
					"type":        "boolean",
					"description": "Se verdadeiro, prepara apenas as submissões pendentes de nota (padrão: true)",
				},
			},
			"required": []string{"course_id", "assignment_id"},
		},
	},
	{
		Name:        "canvas_validate_grades",
		Description: "Valida notas e feedbacks contra limites da atividade e lista de matriculados, gerando a tabela Markdown formatada para o ponto de parada obrigatório de aprovação humana.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"course_id": map[string]any{
					"type":        "string",
					"description": "ID numérico da disciplina no Canvas",
				},
				"assignment_id": map[string]any{
					"type":        "string",
					"description": "ID numérico da atividade no Canvas",
				},
				"grades": map[string]any{
					"type":        "array",
					"description": "Lista de notas e comentários sugeridos para os alunos",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"user_id": map[string]any{"type": "string"},
							"grade":   map[string]any{"type": "string"},
							"comment": map[string]any{"type": "string"},
						},
						"required": []string{"user_id", "grade"},
					},
				},
			},
			"required": []string{"course_id", "assignment_id", "grades"},
		},
	},
	{
		Name:        "canvas_unpack_zip",
		Description: "Descompacta e organiza automaticamente arquivos ZIP de submissões baixados do Canvas SpeedGrader mapeando para os alunos da disciplina.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"zip_path": map[string]any{
					"type":        "string",
					"description": "Caminho do arquivo ZIP local (ex: 'scratch/envios.zip')",
				},
				"course_id": map[string]any{
					"type":        "string",
					"description": "ID numérico da disciplina no Canvas (opcional)",
				},
				"output_dir": map[string]any{
					"type":        "string",
					"description": "Diretório de saída (padrão: 'scratch/unpacked')",
				},
			},
			"required": []string{"zip_path"},
		},
	},
	{
		Name:        "canvas_download_attachment",
		Description: "Baixa um arquivo enviado pelo aluno como anexo (PDF, .c, .py, .zip, etc.) diretamente para o disco local.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"download_url": map[string]any{
					"type":        "string",
					"description": "URL de download do anexo",
				},
				"destination_path": map[string]any{
					"type":        "string",
					"description": "Caminho local onde o arquivo deve ser salvo",
				},
			},
			"required": []string{"download_url", "destination_path"},
		},
	},
	{
		Name:        "canvas_fetch_github_repo",
		Description: "Baixa um repositório público do GitHub enviado pelo aluno, descompacta localmente, cataloga os arquivos de código-fonte e extrai o trecho do arquivo principal (main/código fonte) mastigado para economizar tokens.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"github_url": map[string]any{
					"type":        "string",
					"description": "URL do repositório no GitHub (ex: 'https://github.com/usuario/projeto')",
				},
				"destination_dir": map[string]any{
					"type":        "string",
					"description": "Diretório local para extração (opcional; padrão: 'scratch/github_repos/<owner>_<repo>')",
				},
			},
			"required": []string{"github_url"},
		},
	},
	{
		Name:        "canvas_submit_grade",
		Description: "Lança uma nota e um feedback/comentário para a entrega de um aluno específico no Canvas LMS.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"course_id": map[string]any{
					"type":        "string",
					"description": "ID numérico da disciplina no Canvas",
				},
				"assignment_id": map[string]any{
					"type":        "string",
					"description": "ID numérico da atividade no Canvas",
				},
				"user_id": map[string]any{
					"type":        "string",
					"description": "ID numérico do estudante no Canvas (ex: '310790')",
				},
				"grade": map[string]any{
					"type":        "string",
					"description": "Nota atribuída (ex: '100', '8.5', '90')",
				},
				"comment": map[string]any{
					"type":        "string",
					"description": "Feedback ou comentário de correção detalhado para o estudante",
				},
			},
			"required": []string{"course_id", "assignment_id", "user_id", "grade"},
		},
	},
	{
		Name:        "canvas_submit_grades_batch",
		Description: "Lança notas e feedbacks em lote para múltiplos alunos de uma só vez em uma atividade do Canvas LMS.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"course_id": map[string]any{
					"type":        "string",
					"description": "ID numérico da disciplina no Canvas",
				},
				"assignment_id": map[string]any{
					"type":        "string",
					"description": "ID numérico da atividade no Canvas",
				},
				"grades": map[string]any{
					"type":        "array",
					"description": "Lista de notas e feedbacks dos alunos",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"user_id": map[string]any{"type": "string", "description": "ID do aluno no Canvas"},
							"grade":   map[string]any{"type": "string", "description": "Nota atribuída"},
							"comment": map[string]any{"type": "string", "description": "Feedback explicativo"},
						},
						"required": []string{"user_id", "grade"},
					},
				},
			},
			"required": []string{"course_id", "assignment_id", "grades"},
		},
	},
}

func runMCPServer(client *CanvasClient) {
	logger := log.New(os.Stderr, "[Canvas MCP] ", log.LstdFlags)
	logger.Println("Servidor MCP iniciado aguardando conexões via stdio...")

	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			logger.Printf("Erro na leitura de stdin: %v", err)
			continue
		}

		line = []byte(string(line))
		if len(line) == 0 || string(line) == "\n" || string(line) == "\r\n" {
			continue
		}

		var req JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			logger.Printf("Erro ao decodificar JSON-RPC: %v", err)
			sendMCPError(writer, nil, -32700, "Parse error")
			continue
		}

		handleMCPRequest(writer, logger, client, &req)
	}
}

func handleMCPRequest(w *bufio.Writer, logger *log.Logger, client *CanvasClient, req *JSONRPCRequest) {
	switch req.Method {
	case "initialize":
		result := map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]any{
				"tools": map[string]any{},
			},
			"serverInfo": map[string]any{
				"name":    "afya-canvas-mcp",
				"version": "1.1.0",
			},
		}
		sendMCPResponse(w, req.ID, result)

	case "notifications/initialized":
		logger.Println("Cliente MCP confirmou inicialização.")

	case "ping":
		sendMCPResponse(w, req.ID, map[string]any{})

	case "tools/list":
		sendMCPResponse(w, req.ID, map[string]any{
			"tools": mcpTools,
		})

	case "tools/call":
		var callParams MCPToolCallParams
		if err := json.Unmarshal(req.Params, &callParams); err != nil {
			sendMCPError(w, req.ID, -32602, "Parâmetros de tools/call inválidos")
			return
		}

		toolResult, err := executeMCPTool(client, callParams.Name, callParams.Arguments)
		if err != nil {
			sendMCPResponse(w, req.ID, MCPToolCallResult{
				Content: []MCPContentItem{
					{Type: "text", Text: fmt.Sprintf("Erro ao executar ferramenta %s: %v", callParams.Name, err)},
				},
				IsError: true,
			})
			return
		}

		resultJSON, _ := json.MarshalIndent(toolResult, "", "  ")
		sendMCPResponse(w, req.ID, MCPToolCallResult{
			Content: []MCPContentItem{
				{Type: "text", Text: string(resultJSON)},
			},
		})

	default:
		sendMCPError(w, req.ID, -32601, fmt.Sprintf("Método '%s' não encontrado", req.Method))
	}
}

func executeMCPTool(client *CanvasClient, name string, rawArgs json.RawMessage) (any, error) {
	switch name {
	case "canvas_list_pending_assignments":
		return client.ListPendingAssignments()

	case "canvas_list_courses":
		return client.ListCourses()

	case "canvas_list_assignments":
		var args struct {
			CourseID string `json:"course_id"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return nil, err
		}
		return client.ListAssignments(args.CourseID)

	case "canvas_get_assignment":
		var args struct {
			CourseID     string `json:"course_id"`
			AssignmentID string `json:"assignment_id"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return nil, err
		}
		return client.GetAssignment(args.CourseID, args.AssignmentID)

	case "canvas_list_students":
		var args struct {
			CourseID string `json:"course_id"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return nil, err
		}
		return client.ListStudents(args.CourseID)

	case "canvas_get_submissions":
		var args struct {
			CourseID     string `json:"course_id"`
			AssignmentID string `json:"assignment_id"`
			OnlyPending  *bool  `json:"only_pending"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return nil, err
		}
		onlyPending := true
		if args.OnlyPending != nil {
			onlyPending = *args.OnlyPending
		}
		return client.GetSubmissionsDetails(args.CourseID, args.AssignmentID, onlyPending)

	case "canvas_get_grading_status":
		var args struct {
			CourseID     string `json:"course_id"`
			AssignmentID string `json:"assignment_id"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return nil, err
		}
		return client.GetGradingStatus(args.CourseID, args.AssignmentID)

	case "canvas_prepare_assignment":
		var args struct {
			CourseID     string `json:"course_id"`
			AssignmentID string `json:"assignment_id"`
			OutputDir    string `json:"output_dir"`
			OnlyPending  *bool  `json:"only_pending"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return nil, err
		}
		onlyPending := true
		if args.OnlyPending != nil {
			onlyPending = *args.OnlyPending
		}
		return client.PrepareAssignment(args.CourseID, args.AssignmentID, args.OutputDir, onlyPending)

	case "canvas_validate_grades":
		var args struct {
			CourseID     string       `json:"course_id"`
			AssignmentID string       `json:"assignment_id"`
			Grades       []GradeEntry `json:"grades"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return nil, err
		}
		return client.ValidateGrades(args.CourseID, args.AssignmentID, args.Grades)

	case "canvas_unpack_zip":
		var args struct {
			ZipPath   string `json:"zip_path"`
			CourseID  string `json:"course_id"`
			OutputDir string `json:"output_dir"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return nil, err
		}
		return client.UnpackSubmissionsZip(args.ZipPath, args.CourseID, args.OutputDir)

	case "canvas_download_attachment":
		var args struct {
			DownloadURL     string `json:"download_url"`
			DestinationPath string `json:"destination_path"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return nil, err
		}
		savedPath, err := client.DownloadAttachment(args.DownloadURL, args.DestinationPath)
		if err != nil {
			return nil, err
		}
		return map[string]string{
			"status":     "success",
			"saved_path": savedPath,
		}, nil

	case "canvas_fetch_github_repo":
		var args struct {
			GitHubURL      string `json:"github_url"`
			DestinationDir string `json:"destination_dir"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return nil, err
		}
		return client.FetchGitHubRepo(args.GitHubURL, args.DestinationDir)

	case "canvas_submit_grade":
		var args struct {
			CourseID     string `json:"course_id"`
			AssignmentID string `json:"assignment_id"`
			UserID       string `json:"user_id"`
			Grade        string `json:"grade"`
			Comment      string `json:"comment"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return nil, err
		}
		return client.SubmitGrade(args.CourseID, args.AssignmentID, args.UserID, args.Grade, args.Comment)

	case "canvas_submit_grades_batch":
		var args struct {
			CourseID     string       `json:"course_id"`
			AssignmentID string       `json:"assignment_id"`
			Grades       []GradeEntry `json:"grades"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return nil, err
		}
		return client.SubmitGradesBatch(args.CourseID, args.AssignmentID, args.Grades)

	default:
		return nil, fmt.Errorf("ferramenta desconhecida: %s", name)
	}
}

func sendMCPResponse(w *bufio.Writer, id any, result any) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	bytes, _ := json.Marshal(resp)
	w.Write(bytes)
	w.WriteByte('\n')
	w.Flush()
}

func sendMCPError(w *bufio.Writer, id any, code int, message string) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
		},
	}
	bytes, _ := json.Marshal(resp)
	w.Write(bytes)
	w.WriteByte('\n')
	w.Flush()
}
