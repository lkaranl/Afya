//go:build web

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// ChatMessage representa uma mensagem no histórico da conversa
type ChatMessage struct {
	Role       string     `json:"role"` // "user", "assistant", "system", "tool"
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	Name       string     `json:"name,omitempty"`
}

// ToolCall representa uma requisição de invocação de ferramenta pelo modelo
type ToolCall struct {
	ID       string `json:"id,omitempty"`
	Type     string `json:"type,omitempty"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// AgentActionCard representa uma ação sensível que requer aprovação do professor (ex: notas)
type AgentActionCard struct {
	Type         string `json:"type"` // "grade_approval", "info"
	Title        string `json:"title"`
	Description  string `json:"description"`
	CourseID     string `json:"course_id,omitempty"`
	AssignmentID string `json:"assignment_id,omitempty"`
	Payload      any    `json:"payload,omitempty"`
}

// AgentChatResponse é a resposta estruturada para o frontend
type AgentChatResponse struct {
	Reply        string           `json:"reply"`
	ToolExecuted []string         `json:"tools_executed,omitempty"`
	ActionCard   *AgentActionCard `json:"action_card,omitempty"`
	DurationMs   int64            `json:"duration_ms,omitempty"`
	Model        string           `json:"model,omitempty"`
}

// AgentEngine gerencia as chamadas LLM e a execução de ferramentas
type AgentEngine struct {
	Provider   string // "gemini", "openai", "ollama"
	APIKey     string
	Model      string
	BaseURL    string
	Client     *CanvasClient
	HTTPClient *http.Client
}

// NewAgentEngine inicializa o motor com base nas variáveis do .env
func NewAgentEngine(client *CanvasClient) *AgentEngine {
	provider := strings.ToLower(strings.TrimSpace(os.Getenv("AI_PROVIDER")))
	apiKey := strings.TrimSpace(os.Getenv("AI_API_KEY"))

	// Detecta OpenRouter se a chave começar com sk-or- ou se OPENROUTER_API_KEY existir
	if provider == "" {
		if os.Getenv("OPENROUTER_API_KEY") != "" || strings.HasPrefix(apiKey, "sk-or-") {
			provider = "openrouter"
		} else if os.Getenv("GEMINI_API_KEY") != "" {
			provider = "gemini"
		} else if os.Getenv("OPENAI_API_KEY") != "" {
			provider = "openai"
		} else if apiKey != "" {
			provider = "gemini" // padrão
		} else {
			provider = "gemini"
		}
	}

	if apiKey == "" {
		if provider == "openrouter" {
			apiKey = os.Getenv("OPENROUTER_API_KEY")
		} else if provider == "gemini" {
			apiKey = os.Getenv("GEMINI_API_KEY")
		} else {
			apiKey = os.Getenv("OPENAI_API_KEY")
		}
	}

	baseURL := strings.TrimSpace(os.Getenv("AI_BASE_URL"))
	if provider == "openrouter" && baseURL == "" {
		baseURL = "https://openrouter.ai/api/v1"
	}

	model := strings.TrimSpace(os.Getenv("AI_MODEL"))
	if model == "" {
		if provider == "openrouter" {
			model = "google/gemini-2.0-flash-001"
		} else if provider == "gemini" {
			model = "gemini-2.0-flash"
		} else {
			model = "gpt-4o-mini"
		}
	}

	return &AgentEngine{
		Provider: provider,
		APIKey:   apiKey,
		Model:    model,
		BaseURL:  baseURL,
		Client:   client,
		HTTPClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// SystemPrompt oficial compilado com as diretrizes do AGENTS.md e CONSEPE Afya
const AgentSystemPrompt = `Você é o Assistente Pedagógico e de Gestão Acadêmica oficial do Professor no Canvas LMS da Afya / Centro Universitário São Lucas Ji-Paraná.

SUA MISSÃO:
Ajudar o professor a gerenciar disciplinas, verificar atividades pendentes de nota, criar questionários, páginas de módulos e auditar o sistema de avaliação com praticidade, rapidez e elegância.

REGRAS INSTITUCIONAIS AFYA (CONSEPE 005/2024 & NAPED 2026):
1. Disciplinas Presenciais sem TPI (Ciência da Computação, Engenharias):
   - N1: Prova Teórica Individual (30 pts, modelo ENADE) + Atividades Práticas (20 pts) = 50 pts.
   - N2: Prova Teórica Individual (30 pts, modelo ENADE) + Atividades Práticas (20 pts) = 50 pts.
   - Total Semestral: 100 pontos.
2. Critérios de Aprovação:
   - Aprovação Direta: Nota Semestral >= 70 pontos e Frequência >= 75%.
   - Exame Final: Nota Semestral entre 40 e 69 pontos (abaixo de 40 é reprovação direta). Média final pós-exame: (Nota + Exame)/2 >= 60 pontos.
3. Prazos Regimentais:
   - Devolutiva e discussão de provas em sala: máximo de até 10 dias após a aplicação (Art. 16, § 2º).
   - Revisão de prova pelo aluno: até 2 dias letivos após devolutiva em sala.

DISCIPLINAS DO PROFESSOR E GESTÃO DINÂMICA DE SEMESTRES:
- Nunca assuma IDs, códigos ou disciplinas fixas. Sempre consulte a API via ferramentas MCP (canvas_list_courses ou canvas_list_pending_assignments).
- O MCP analisa automaticamente as datas de vigência e os termos oficiais do Canvas LMS e classifica cada disciplina:
  - is_current_term = true (term_status = "atual"): disciplinas em andamento no semestre letivo vigente.
  - is_current_term = false (term_status = "anterior"): disciplinas de semestres anteriores já concluídos (histórico).
- Comportamento padrão: Quando o professor solicitar matérias, pendências de correção ou relatórios sem especificar um semestre passado, priorize e foque sempre nas turmas com is_current_term = true. Utilize canvas_list_courses com term_filter='current' ou grouped=true.

REGRA DE OURO DE USABILIDADE - NUNCA PEÇA IDs NUMÉRICOS AO PROFESSOR:
- NUNCA, SOB HIPÓTESE ALGUMA, pergunte ao professor qual é o "ID" numérico de uma disciplina ou atividade. O professor é um docente humano e JAMAIS sabe ou precisa saber IDs numéricos de banco de dados.
- O professor sempre se refere às disciplinas pelo NOME (ex: "Estrutura de Dados"), pelo PERÍODO (ex: "4º Período", "2º Período") ou simplesmente por "turma ativa" / "turmas deste semestre".
- QUANDO O PROFESSOR PEDIR UMA AÇÃO (como radar de evasão, plágio, status de notas, listar atividades ou pendências):
  1. Chame IMEDIATAMENTE a ferramenta canvas_list_courses(term_filter: "current") nos bastidores para ver as matérias ativas.
  2. Se houver apenas 1 turma ativa, execute a ação diretamente nela sem fazer perguntas desnecessárias.
  3. Se houver mais de 1 turma ativa (por exemplo: "Estrutura de Dados - 4º Período" e "Estrutura de Dados - 2º Período"):
     - Apresente educadamente as opções encontradas usando os NOMES e PERÍODOS didáticos (ex: "Professor, identifiquei duas turmas ativas neste semestre: 1. Estrutura de Dados (4º Período) e 2. Estrutura de Dados (2º Período). Em qual delas deseja que eu execute o radar de evasão, ou deseja que eu analise ambas?").
     - NUNCA diga "me informe o ID da disciplina".
  4. O sistema aceita tanto o ID descoberto quanto o nome/período da disciplina (ex: "Estrutura de Dados 4º Período" ou "4º Período") no parâmetro 'course_id', resolvendo automaticamente.

TOM DOS FEEDBACKS E RESPOSTAS:
- Seja sempre respeitoso, formal, técnico, objetivo e didático.
- Evite bajulação, adjetivação afetuosa ou qualquer intimidade pessoal. O tom deve ser estritamente institucional.
- Explique conceitos técnicos de forma clara para estudantes iniciantes.
- Ao formatar tabelas e notas, utilize Markdown impecável.

PONTO DE PARADA PARA NOTAS:
- Você tem ferramentas para validar notas (canvas_validate_grades). NUNCA publique notas no Canvas LMS sem antes apresentar a tabela detalhada para conferência e aprovação do professor.`

// Chat processa a mensagem do professor, executa as ferramentas do Canvas necessárias e retorna a resposta final
func (e *AgentEngine) Chat(ctx context.Context, userMsg string, history []ChatMessage) (*AgentChatResponse, error) {
	start := time.Now()

	if e.APIKey == "" && e.Provider != "ollama" {
		return &AgentChatResponse{
			Reply:      "⚠️ **Chave de API de IA não configurada!**\n\nPara conversar com o agente, adicione sua chave no arquivo `.env`:\n```env\nAI_PROVIDER=gemini\nAI_API_KEY=sua_chave_do_google_aqui\nAI_MODEL=gemini-2.0-flash\n```\n*(Se preferir OpenAI, configure `AI_PROVIDER=openai` e `AI_API_KEY=sk-...`)*",
			DurationMs: time.Since(start).Milliseconds(),
			Model:      e.Model,
		}, nil
	}

	var res *AgentChatResponse
	var err error

	if e.Provider == "gemini" {
		res, err = e.chatGemini(ctx, userMsg, history)
	} else {
		res, err = e.chatOpenAI(ctx, userMsg, history)
	}

	if res != nil {
		res.DurationMs = time.Since(start).Milliseconds()
		if res.Model == "" {
			res.Model = e.Model
		}
	}
	return res, err
}

// -----------------------------------------------------------------------
// INTEGRAÇÃO COM GOOGLE GEMINI (REST API v1beta)
// -----------------------------------------------------------------------

func (e *AgentEngine) chatGemini(ctx context.Context, userMsg string, history []ChatMessage) (*AgentChatResponse, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", e.Model, e.APIKey)

	// Prepara tools no formato Gemini
	geminiTools := []map[string]any{
		{
			"functionDeclarations": e.getGeminiFunctionDeclarations(),
		},
	}

	// Prepara contents
	var contents []map[string]any
	for _, h := range history {
		role := "user"
		if h.Role == "assistant" {
			role = "model"
		}
		contents = append(contents, map[string]any{
			"role": role,
			"parts": []map[string]any{
				{"text": h.Content},
			},
		})
	}
	// Mensagem atual
	contents = append(contents, map[string]any{
		"role": "user",
		"parts": []map[string]any{
			{"text": userMsg},
		},
	})

	var toolsExecuted []string
	var actionCard *AgentActionCard

	// Loop de Tool Calling (até 5 iterações de reflexão autônoma)
	for iter := 0; iter < 5; iter++ {
		reqBody := map[string]any{
			"systemInstruction": map[string]any{
				"parts": []map[string]any{
					{"text": AgentSystemPrompt},
				},
			},
			"contents": contents,
			"tools":    geminiTools,
		}

		bodyBytes, _ := json.Marshal(reqBody)
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := e.HTTPClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("falha ao comunicar com Google Gemini: %w", err)
		}
		respBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("erro Gemini HTTP %d: %s", resp.StatusCode, string(respBytes))
		}

		var geminiResp struct {
			Candidates []struct {
				Content struct {
					Role  string `json:"role"`
					Parts []struct {
						Text         string `json:"text,omitempty"`
						FunctionCall *struct {
							Name string         `json:"name"`
							Args map[string]any `json:"args"`
						} `json:"functionCall,omitempty"`
					} `json:"parts"`
				} `json:"content"`
			} `json:"candidates"`
		}

		if err := json.Unmarshal(respBytes, &geminiResp); err != nil {
			return nil, fmt.Errorf("erro ao decodificar resposta Gemini: %w", err)
		}

		if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
			return &AgentChatResponse{Reply: "Não recebi resposta do modelo de IA."}, nil
		}

		cand := geminiResp.Candidates[0]
		var modelText string
		var funcCalls []*struct {
			Name string         `json:"name"`
			Args map[string]any `json:"args"`
		}

		for _, p := range cand.Content.Parts {
			if p.Text != "" {
				modelText += p.Text
			}
			if p.FunctionCall != nil {
				funcCalls = append(funcCalls, p.FunctionCall)
			}
		}

		// Se o modelo não chamou ferramentas, encerrou a resposta
		if len(funcCalls) == 0 {
			return &AgentChatResponse{
				Reply:        modelText,
				ToolExecuted: toolsExecuted,
				ActionCard:   actionCard,
			}, nil
		}

		// Adiciona a resposta com FunctionCall ao histórico
		contents = append(contents, map[string]any{
			"role":  "model",
			"parts": cand.Content.Parts,
		})

		// Executa cada ferramenta chamada
		var funcResponseParts []map[string]any
		for _, fc := range funcCalls {
			toolsExecuted = append(toolsExecuted, fc.Name)
			log.Printf("[Agent Engine] Executando ferramenta: %s com args: %v", fc.Name, fc.Args)

			rawArgs, _ := json.Marshal(fc.Args)
			toolResult, execErr := executeMCPTool(e.Client, fc.Name, rawArgs)

			var resPayload any
			if execErr != nil {
				resPayload = map[string]any{"error": execErr.Error()}
			} else {
				resPayload = toolResult
			}

			// Se a ferramenta for de validação de notas, prepara o card de ação para aprovação humana
			if fc.Name == "canvas_validate_grades" {
				actionCard = &AgentActionCard{
					Type:        "grade_approval",
					Title:       "Conferência e Aprovação de Notas",
					Description: "Revise a distribuição sugerida abaixo antes de publicar no Canvas LMS.",
					Payload:     toolResult,
				}
			}

			funcResponseParts = append(funcResponseParts, map[string]any{
				"functionResponse": map[string]any{
					"name":     fc.Name,
					"response": resPayload,
				},
			})
		}

		contents = append(contents, map[string]any{
			"role":  "function",
			"parts": funcResponseParts,
		})
	}

	return &AgentChatResponse{
		Reply:        "O processamento atingiu o limite de etapas de raciocínio. Por favor, tente simplificar sua solicitação.",
		ToolExecuted: toolsExecuted,
	}, nil
}

// -----------------------------------------------------------------------
// INTEGRAÇÃO COM OPENAI / OLLAMA / COMPATÍVEIS (REST API v1)
// -----------------------------------------------------------------------

func (e *AgentEngine) chatOpenAI(ctx context.Context, userMsg string, history []ChatMessage) (*AgentChatResponse, error) {
	baseURL := e.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	endpoint := fmt.Sprintf("%s/chat/completions", strings.TrimRight(baseURL, "/"))

	// Monta mensagens
	var msgs []map[string]any
	msgs = append(msgs, map[string]any{
		"role":    "system",
		"content": AgentSystemPrompt,
	})

	for _, h := range history {
		m := map[string]any{
			"role":    h.Role,
			"content": h.Content,
		}
		if len(h.ToolCalls) > 0 {
			m["tool_calls"] = h.ToolCalls
		}
		if h.ToolCallID != "" {
			m["tool_call_id"] = h.ToolCallID
		}
		if h.Name != "" {
			m["name"] = h.Name
		}
		msgs = append(msgs, m)
	}

	msgs = append(msgs, map[string]any{
		"role":    "user",
		"content": userMsg,
	})

	openaiTools := e.getOpenAITools()
	var toolsExecuted []string
	var actionCard *AgentActionCard

	for iter := 0; iter < 5; iter++ {
		reqBody := map[string]any{
			"model":    e.Model,
			"messages": msgs,
			"tools":    openaiTools,
		}

		bodyBytes, _ := json.Marshal(reqBody)
		req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		if e.APIKey != "" {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", e.APIKey))
		}
		if e.Provider == "openrouter" || strings.Contains(endpoint, "openrouter.ai") {
			req.Header.Set("HTTP-Referer", "http://localhost:3000")
			req.Header.Set("X-Title", "Afya Canvas Assistant")
		}

		resp, err := e.HTTPClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("falha ao comunicar com provedor OpenAI: %w", err)
		}
		respBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("erro OpenAI/OpenRouter HTTP %d: %s", resp.StatusCode, string(respBytes))
		}

		if len(bytes.TrimSpace(respBytes)) == 0 {
			return nil, fmt.Errorf("o modelo '%s' retornou uma resposta vazia. O provedor no OpenRouter pode estar temporariamente indisponível ou sobrecarregado. Tente outro modelo (ex: 'nex-agi/nex-n2.5-mini:free' ou 'google/gemini-2.0-flash-001')", e.Model)
		}

		var openAIResp struct {
			Choices []struct {
				Message struct {
					Role      string     `json:"role"`
					Content   *string    `json:"content"`
					ToolCalls []ToolCall `json:"tool_calls"`
				} `json:"message"`
			} `json:"choices"`
			Error *struct {
				Message string `json:"message"`
				Code    any    `json:"code"`
			} `json:"error,omitempty"`
		}

		if err := json.Unmarshal(respBytes, &openAIResp); err != nil {
			return nil, fmt.Errorf("erro ao decodificar resposta do OpenRouter: %w (conteúdo recebido: %s)", err, string(respBytes))
		}

		if openAIResp.Error != nil {
			return nil, fmt.Errorf("erro retornado pelo OpenRouter: %s", openAIResp.Error.Message)
		}

		if len(openAIResp.Choices) == 0 {
			return &AgentChatResponse{Reply: "Não recebi opções de resposta do modelo."}, nil
		}

		choice := openAIResp.Choices[0]
		replyContent := ""
		if choice.Message.Content != nil {
			replyContent = *choice.Message.Content
		}

		if len(choice.Message.ToolCalls) == 0 {
			return &AgentChatResponse{
				Reply:        replyContent,
				ToolExecuted: toolsExecuted,
				ActionCard:   actionCard,
			}, nil
		}

		// Adiciona a intenção da tool ao histórico com content seguro
		assistantMsg := map[string]any{
			"role":       "assistant",
			"tool_calls": choice.Message.ToolCalls,
		}
		if choice.Message.Content != nil {
			assistantMsg["content"] = *choice.Message.Content
		} else {
			assistantMsg["content"] = nil
		}
		msgs = append(msgs, assistantMsg)

		// Executa cada tool call
		for _, tc := range choice.Message.ToolCalls {
			toolsExecuted = append(toolsExecuted, tc.Function.Name)
			log.Printf("[Agent Engine] Executando ferramenta OpenAI: %s com args: %s", tc.Function.Name, tc.Function.Arguments)

			toolResult, execErr := executeMCPTool(e.Client, tc.Function.Name, []byte(tc.Function.Arguments))
			var resText string
			if execErr != nil {
				resText = fmt.Sprintf(`{"error": %q}`, execErr.Error())
			} else {
				resBytes, _ := json.Marshal(toolResult)
				resText = string(resBytes)
			}

			if tc.Function.Name == "canvas_validate_grades" {
				actionCard = &AgentActionCard{
					Type:        "grade_approval",
					Title:       "Conferência e Aprovação de Notas",
					Description: "Revise a distribuição sugerida abaixo antes de publicar no Canvas LMS.",
					Payload:     toolResult,
				}
			}

			msgs = append(msgs, map[string]any{
				"role":         "tool",
				"tool_call_id": tc.ID,
				"name":         tc.Function.Name,
				"content":      resText,
			})
		}
	}

	return &AgentChatResponse{
		Reply:        "O limite de etapas de raciocínio foi atingido.",
		ToolExecuted: toolsExecuted,
	}, nil
}

// -----------------------------------------------------------------------
// MAPEAMENTO DAS FERRAMENTAS DO MCP PARA OS SCHEMAS DE FUNCTION CALLING
// -----------------------------------------------------------------------

func (e *AgentEngine) getOpenAITools() []map[string]any {
	var tools []map[string]any
	for _, t := range mcpTools {
		tools = append(tools, map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        t.Name,
				"description": t.Description,
				"parameters":  t.InputSchema,
			},
		})
	}
	return tools
}

func (e *AgentEngine) getGeminiFunctionDeclarations() []map[string]any {
	var decls []map[string]any
	for _, t := range mcpTools {
		decls = append(decls, map[string]any{
			"name":        t.Name,
			"description": t.Description,
			"parameters":  t.InputSchema,
		})
	}
	return decls
}
