package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
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
		Description: "Varre as disciplinas do professor e lista as atividades com submissões pendentes de correção (needs_grading_count > 0), identificando período e semestre.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"current_term_only": map[string]any{
					"type":        "boolean",
					"description": "Se verdadeiro (ou omitido), filtra para retornar apenas pendências das disciplinas do semestre atual/vigente. Se falso, inclui matérias anteriores.",
				},
			},
		},
	},
	{
		Name:        "canvas_list_courses",
		Description: "Lista de forma inteligente as disciplinas do docente no Canvas LMS, identificando automaticamente quais pertencem ao semestre atual (vigente) e quais são de semestres anteriores (concluídos), com períodos curriculares, total de alunos e status.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"term_filter": map[string]any{
					"type":        "string",
					"enum":        []string{"all", "current", "past"},
					"description": "Filtro de período letivo: 'current' (somente matérias do semestre atual/vigente), 'past' (somente matérias de semestres anteriores já concluídos) ou 'all' (todas as matérias cadastradas). Padrão: 'all'.",
				},
				"grouped": map[string]any{
					"type":        "boolean",
					"description": "Se verdadeiro, agrupa o resultado em duas listas organizadas: 'current_courses' (semestre atual) e 'past_courses' (semestres anteriores).",
				},
			},
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
	{
		Name:        "canvas_create_assignment",
		Description: "Cria uma nova atividade acadêmica no Canvas LMS diretamente com enunciado formatado em HTML, pontuação, prazos, tipos de entrega e regras de grupo.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"course_id": map[string]any{
					"type":        "string",
					"description": "ID numérico da disciplina no Canvas",
				},
				"name": map[string]any{
					"type":        "string",
					"description": "Título da atividade",
				},
				"description": map[string]any{
					"type":        "string",
					"description": "Enunciado e instruções da atividade formatados em HTML institucional",
				},
				"points_possible": map[string]any{
					"type":        "number",
					"description": "Pontuação máxima da atividade (ex: 100.0 ou 10.0)",
				},
				"submission_types": map[string]any{
					"type":        "array",
					"description": "Tipos de entrega aceitos (ex: ['online_url'], ['online_upload'], ['online_text_entry']). Padrão: ['online_url']",
					"items":       map[string]any{"type": "string"},
				},
				"due_at": map[string]any{
					"type":        "string",
					"description": "Data e hora de vencimento no padrão ISO UTC (ex: '2026-09-25T02:59:59Z')",
				},
				"unlock_at": map[string]any{
					"type":        "string",
					"description": "Data e hora de liberação da atividade (opcional)",
				},
				"lock_at": map[string]any{
					"type":        "string",
					"description": "Data e hora de bloqueio final para envios (opcional)",
				},
				"group_category_id": map[string]any{
					"type":        "string",
					"description": "ID do conjunto de grupos caso seja uma tarefa em equipe (opcional)",
				},
				"published": map[string]any{
					"type":        "boolean",
					"description": "Se verdadeiro, publica a atividade imediatamente para os alunos (padrão: true)",
				},
				"allowed_extensions": map[string]any{
					"type":        "array",
					"description": "Extensões permitidas em caso de upload de arquivos (ex: ['c', 'h', 'zip'])",
					"items":       map[string]any{"type": "string"},
				},
			},
			"required": []string{"course_id", "name", "description", "points_possible"},
		},
	},
	{
		Name:        "canvas_create_quiz",
		Description: "Cria um Questionário / Quiz completo no Canvas LMS com perguntas de múltipla escolha ou discursivas, alternativas, pesos de pontuação e feedbacks dos distratores em uma só chamada.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"course_id": map[string]any{
					"type":        "string",
					"description": "ID numérico da disciplina no Canvas",
				},
				"title": map[string]any{
					"type":        "string",
					"description": "Título do questionário",
				},
				"description": map[string]any{
					"type":        "string",
					"description": "Instruções do questionário em HTML",
				},
				"time_limit": map[string]any{
					"type":        "integer",
					"description": "Tempo limite em minutos para responder à prova/quiz (opcional)",
				},
				"shuffle_answers": map[string]any{
					"type":        "boolean",
					"description": "Embaralhar as opções de resposta para os estudantes (padrão: true)",
				},
				"allowed_attempts": map[string]any{
					"type":        "integer",
					"description": "Número de tentativas permitidas (padrão: 1)",
				},
				"due_at": map[string]any{
					"type":        "string",
					"description": "Prazo de entrega em formato ISO UTC",
				},
				"published": map[string]any{
					"type":        "boolean",
					"description": "Publicar o quiz imediatamente (padrão: true)",
				},
				"questions": map[string]any{
					"type":        "array",
					"description": "Lista de questões do questionário",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"title":           map[string]any{"type": "string", "description": "Título/Identificador da questão"},
							"text":            map[string]any{"type": "string", "description": "Enunciado da questão (HTML)"},
							"type":            map[string]any{"type": "string", "description": "Tipo da questão (ex: 'multiple_choice_question', 'true_false_question')"},
							"points_possible": map[string]any{"type": "number", "description": "Pontos desta questão (padrão: 10.0)"},
							"answers": map[string]any{
								"type": "array",
								"description": "Alternativas da questão (para múltipla escolha)",
								"items": map[string]any{
									"type": "object",
									"properties": map[string]any{
										"text":    map[string]any{"type": "string", "description": "Texto da alternativa"},
										"weight":  map[string]any{"type": "integer", "description": "100 para correta, 0 para incorreta"},
										"comment": map[string]any{"type": "string", "description": "Feedback explicativo para quem marcar esta alternativa"},
									},
									"required": []string{"text", "weight"},
								},
							},
						},
						"required": []string{"text", "answers"},
					},
				},
			},
			"required": []string{"course_id", "title"},
		},
	},
	{
		Name:        "canvas_create_module",
		Description: "Cria um novo Módulo semanal ou temático na disciplina do Canvas LMS para estruturar o plano de ensino.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"course_id": map[string]any{
					"type":        "string",
					"description": "ID numérico da disciplina no Canvas",
				},
				"name": map[string]any{
					"type":        "string",
					"description": "Nome do módulo (ex: 'Semana 4: Árvores Binárias de Busca')",
				},
				"position": map[string]any{
					"type":        "integer",
					"description": "Posição ordinal do módulo na lista do curso (opcional)",
				},
				"unlock_at": map[string]any{
					"type":        "string",
					"description": "Data de desbloqueio automático do módulo (opcional)",
				},
			},
			"required": []string{"course_id", "name"},
		},
	},
	{
		Name:        "canvas_add_module_item",
		Description: "Vincula uma tarefa, questionário, link externo ou página a um Módulo de aula existente.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"course_id": map[string]any{
					"type":        "string",
					"description": "ID numérico da disciplina no Canvas",
				},
				"module_id": map[string]any{
					"type":        "string",
					"description": "ID numérico do módulo",
				},
				"title": map[string]any{
					"type":        "string",
					"description": "Título de exibição do item no módulo",
				},
				"type": map[string]any{
					"type":        "string",
					"description": "Tipo do item: 'Assignment', 'Quiz', 'ExternalUrl', 'Page', 'SubHeader'",
				},
				"content_id": map[string]any{
					"type":        "string",
					"description": "ID numérico da atividade ou quiz correspondente (obrigatório se type for Assignment ou Quiz)",
				},
				"page_url": map[string]any{
					"type":        "string",
					"description": "URL slug da página Wiki no Canvas (obrigatório se type for Page)",
				},
				"external_url": map[string]any{
					"type":        "string",
					"description": "URL externa caso o tipo seja ExternalUrl",
				},
				"new_tab": map[string]any{
					"type":        "boolean",
					"description": "Abrir em nova aba (para ExternalUrl)",
				},
			},
			"required": []string{"course_id", "module_id", "type"},
		},
	},
	{
		Name:        "canvas_list_modules",
		Description: "Lista todos os módulos de aula e seus respectivos itens cadastrados em uma disciplina do Canvas LMS.",
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
		Name:        "canvas_create_page",
		Description: "Cria uma nova Página de Conteúdo/Teoria (Wiki Page) no Canvas LMS com formatação rica em HTML para aulas e materiais de apoio.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"course_id": map[string]any{
					"type":        "string",
					"description": "ID numérico da disciplina no Canvas",
				},
				"title": map[string]any{
					"type":        "string",
					"description": "Título da página",
				},
				"body": map[string]any{
					"type":        "string",
					"description": "Conteúdo da aula formatado em HTML",
				},
				"published": map[string]any{
					"type":        "boolean",
					"description": "Publicar a página imediatamente (padrão: true)",
				},
			},
			"required": []string{"course_id", "title", "body"},
		},
	},
	{
		Name:        "canvas_list_assignment_groups",
		Description: "Lista os grupos de notas/tarefas da disciplina com suas respectivas porcentagens de peso (ponderação) e atividades vinculadas.",
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
		Name:        "canvas_set_course_weighting",
		Description: "Ativa ou desativa a ponderação de notas final baseada em grupos de tarefas (ex: 50% atividades + 50% prova).",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"course_id": map[string]any{
					"type":        "string",
					"description": "ID numérico da disciplina no Canvas",
				},
				"enable_weights": map[string]any{
					"type":        "boolean",
					"description": "True para ativar cálculo ponderado por grupos, False para pontos corridos",
				},
			},
			"required": []string{"course_id", "enable_weights"},
		},
	},
	{
		Name:        "canvas_create_assignment_group",
		Description: "Cria um novo grupo de atividades com peso percentual (ex: 'Avaliação Oficial / Prova' com peso 50%).",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"course_id": map[string]any{
					"type":        "string",
					"description": "ID numérico da disciplina no Canvas",
				},
				"name": map[string]any{
					"type":        "string",
					"description": "Nome do grupo (ex: 'Atividades Práticas', 'Avaliação Oficial / Prova')",
				},
				"group_weight": map[string]any{
					"type":        "number",
					"description": "Peso percentual do grupo no cálculo da média final (ex: 50.0)",
				},
			},
			"required": []string{"course_id", "name", "group_weight"},
		},
	},
	{
		Name:        "canvas_update_assignment_group",
		Description: "Atualiza o nome e o peso percentual de um grupo de atividades existente.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"course_id": map[string]any{
					"type":        "string",
					"description": "ID numérico da disciplina no Canvas",
				},
				"group_id": map[string]any{
					"type":        "string",
					"description": "ID numérico do grupo de tarefas",
				},
				"name": map[string]any{
					"type":        "string",
					"description": "Novo nome do grupo",
				},
				"group_weight": map[string]any{
					"type":        "number",
					"description": "Novo peso percentual (ex: 50.0)",
				},
			},
			"required": []string{"course_id", "group_id", "name", "group_weight"},
		},
	},
	{
		Name:        "canvas_move_assignment_to_group",
		Description: "Move uma atividade ou quiz para um grupo de notas específico.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"course_id": map[string]any{
					"type":        "string",
					"description": "ID numérico da disciplina no Canvas",
				},
				"assignment_id": map[string]any{
					"type":        "string",
					"description": "ID numérico da atividade",
				},
				"group_id": map[string]any{
					"type":        "string",
					"description": "ID numérico do grupo de destino",
				},
			},
			"required": []string{"course_id", "assignment_id", "group_id"},
		},
	},
	{
		Name:        "canvas_get_institutional_rules",
		Description: "Consulta as regras e diretrizes institucionais da Afya / São Lucas Ji-Paraná (Resolução CONSEPE 005/2024 e Guia NAPED 2026), incluindo composição de notas N1/N2, prazos de devolutiva, revisão de prova e critérios de aprovação.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"modality": map[string]any{
					"type":        "string",
					"description": "Modalidade da disciplina: 'presencial_sem_tpi', 'presencial_com_tpi', 'hibrida_sem_tpi', 'online_assincrona', 'simplificado_50_50' (deixe vazio para listar todas)",
				},
			},
		},
	},
	{
		Name:        "canvas_setup_afya_grading_scheme",
		Description: "Aplica e configura automaticamente a matriz de avaliação oficial da Afya no Canvas LMS, ativando grupos ponderados e estruturando N1 e N2 em conformidade com o regimento.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"course_id": map[string]any{
					"type":        "string",
					"description": "ID numérico da disciplina no Canvas",
				},
				"modality": map[string]any{
					"type":        "string",
					"description": "Modalidade oficial da disciplina: 'presencial_sem_tpi', 'presencial_com_tpi', 'hibrida_sem_tpi', 'online_assincrona', 'simplificado_50_50'",
				},
			},
			"required": []string{"course_id", "modality"},
		},
	},
	{
		Name:        "canvas_detect_plagiarism",
		Description: "Varre e compara automaticamente todos os códigos submetidos pelos alunos em uma atividade, detectando similaridades, cópias literais e tentativas de mascaramento (troca de nomes de variáveis, alteração de comentários e reordenação de blocos), retornando percentual de similaridade, classificação de risco (ALTO/MÉDIO/BAIXO) e tabela formatada em Markdown.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"course_id": map[string]any{
					"type":        "string",
					"description": "ID numérico da disciplina no Canvas LMS.",
				},
				"assignment_id": map[string]any{
					"type":        "string",
					"description": "ID numérico da atividade no Canvas LMS.",
				},
				"similarity_threshold": map[string]any{
					"type":        "number",
					"description": "Limiar percentual mínimo de similaridade para reportar suspeita (padrão: 65). Valores >= 80 representam risco alto.",
				},
				"normalize_identifiers": map[string]any{
					"type":        "boolean",
					"description": "Se verdadeiro (padrão), normaliza nomes de variáveis e funções locais para desmascarar trocas de identificadores. Se falso, compara código bruto.",
				},
			},
			"required": []string{"course_id", "assignment_id"},
		},
	},
	{
		Name:        "canvas_list_inbox_messages",
		Description: "Lista conversas e mensagens diretas do Inbox do Canvas LMS com triagem de dúvidas de estudantes, contagem de não lidas, identificação de participantes, disciplina vinculada e tabela Markdown formatada.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"scope": map[string]any{
					"type":        "string",
					"description": "Filtro de visualização: 'unread' (padrão, para focar nas dúvidas pendentes de atendimento), 'inbox' (todas as recebidas), 'starred', 'sent', 'archived'",
				},
				"course_id": map[string]any{
					"type":        "string",
					"description": "ID numérico da disciplina para filtrar apenas mensagens enviadas por alunos daquela matéria específica (opcional)",
				},
				"limit": map[string]any{
					"type":        "integer",
					"description": "Quantidade máxima de conversas a retornar (padrão: 25)",
				},
			},
		},
	},
	{
		Name:        "canvas_get_inbox_conversation",
		Description: "Obtém o histórico completo e a íntegra de uma conversa no Canvas LMS (thread com todas as mensagens trocadas, autores, datas em BRT, links de anexos e resumo para elaboração de resposta pedagógica).",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"conversation_id": map[string]any{
					"type":        "string",
					"description": "ID numérico da conversa no Canvas LMS",
				},
			},
			"required": []string{"conversation_id"},
		},
	},
	{
		Name:        "canvas_reply_inbox_message",
		Description: "Envia uma resposta oficial e direta em uma conversa existente no Inbox do Canvas LMS. Utilize após sugerir a minuta pedagógica/técnica e obter a aprovação do professor.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"conversation_id": map[string]any{
					"type":        "string",
					"description": "ID numérico da conversa no Canvas LMS",
				},
				"message": map[string]any{
					"type":        "string",
					"description": "Conteúdo da resposta que será enviada para o aluno no Canvas",
				},
			},
			"required": []string{"conversation_id", "message"},
		},
	},
	{
		Name:        "canvas_send_inbox_message",
		Description: "Envia uma nova mensagem direta privada para um ou mais estudantes pelo Canvas LMS (inicia uma nova conversa no Inbox com assunto e contexto da disciplina).",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"recipient_ids": map[string]any{
					"type":        "array",
					"description": "Lista de IDs numéricos dos alunos destinatários no Canvas",
					"items":       map[string]any{"type": "string"},
				},
				"subject": map[string]any{
					"type":        "string",
					"description": "Assunto da mensagem",
				},
				"message": map[string]any{
					"type":        "string",
					"description": "Corpo da mensagem a ser enviada",
				},
				"course_id": map[string]any{
					"type":        "string",
					"description": "ID da disciplina para vincular o contexto da mensagem (opcional)",
				},
			},
			"required": []string{"recipient_ids", "subject", "message"},
		},
	},
	{
		Name:        "canvas_clear_cache",
		Description: "Limpa o cache em memória das requisições do Canvas (turmas, alunos, atividades). Permite forçar sincronização fresca com a API da Afya.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"course_id": map[string]any{
					"type":        "string",
					"description": "ID numérico da disciplina para limpar apenas os dados daquela turma específica (opcional, se omitido limpa todo o cache global)",
				},
			},
		},
	},
	{
		Name:        "canvas_get_cache_stats",
		Description: "Obtém as estatísticas de desempenho do cache em memória (hits, misses, total de requisições, porcentagem de acertos e tempo economizado).",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	},
	{
		Name:        "canvas_detect_at_risk_students",
		Description: "Painel de identificação precoce de estudantes em risco acadêmico ou de evasão, cruzando alunos inativos há mais de 10 dias, tarefas consecutivas zeradas/faltantes e média acumulada abaixo do corte da Afya (70 pontos).",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"course_id": map[string]any{
					"type":        "string",
					"description": "ID numérico ou nome/período da disciplina no Canvas LMS (ex: '162263', 'Estrutura de Dados', '4º Período', ou 'ativa'). Se omitido, analisa a turma ativa do semestre.",
				},
				"inactivity_days": map[string]any{
					"type":        "integer",
					"description": "Número de dias sem acesso à disciplina para sinalizar alerta de ausência (padrão: 10 dias).",
				},
				"grade_cutoff": map[string]any{
					"type":        "number",
					"description": "Nota de corte institucional para aprovação direta (padrão oficial da Afya: 70.0 pontos).",
				},
				"consecutive_threshold": map[string]any{
					"type":        "integer",
					"description": "Número de atividades consecutivas não entregues ou com nota zero (padrão: 2 atividades).",
				},
			},
		},
	},
	{
		Name:        "canvas_get_academic_calendar",
		Description: "Consulta o Calendário Acadêmico Oficial 2026.2 da Afya / Centro Universitário São Lucas Ji-Paraná, retornando marcos oficiais, semanas de prova N1 e N2, 2ª chamada, exames finais, prazos limites de lançamento de notas no Canvas, feriados e compensações de sábados letivos para Ciência da Computação / SHE.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"term": map[string]any{
					"type":        "string",
					"description": "Semestre acadêmico (padrão: '2026.2').",
				},
				"category": map[string]any{
					"type":        "string",
					"description": "Filtro por categoria: 'she' (Computação/Engenharias), 'n1', 'n2', 'feriado', 'sabado_letivo', 'exames', 'docentes', 'online', ou 'all' (todas as categorias).",
				},
				"month": map[string]any{
					"type":        "integer",
					"description": "Filtrar eventos de um mês específico (ex: 7 para Julho, 8 para Agosto, 9 para Setembro, etc. Omitir para o semestre todo).",
				},
				"search": map[string]any{
					"type":        "string",
					"description": "Termo de busca textual para filtrar eventos por título ou data.",
				},
				"only_she": map[string]any{
					"type":        "boolean",
					"description": "Se verdadeiro, restringe a busca apenas a eventos aplicáveis à área de Computação / SHE.",
				},
			},
		},
	},
	{
		Name:        "canvas_get_recommended_reading",
		Description: "Consulta a bibliografia oficial da disciplina (básica e complementar) na Minha Biblioteca da Afya, retornando indicações de leitura por tópico (ex: ponteiros, árvores, ordenação, tabelas hash, grafos), referências ABNT e links autenticados de acesso direto via LTI SSO.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"course_id": map[string]any{
					"type":        "string",
					"description": "ID numérico do curso, nome da disciplina ou período (ex: '4º Período', 'Estrutura de Dados', '160754'). Se omitido, detecta dinamicamente a turma ativa do professor no semestre corrente.",
				},
				"topic": map[string]any{
					"type":        "string",
					"description": "Tópico ou assunto da aula para filtrar capítulos e obras recomendadas (ex: 'ponteiros', 'arvores', 'ordenacao', 'hash', 'grafos', 'complexidade', 'c'). Omitir para listar todas as obras da ementa.",
				},
				"type": map[string]any{
					"type":        "string",
					"description": "Tipo de bibliografia: 'basica', 'complementar' ou 'todas' (padrão: 'todas').",
				},
			},
		},
	},
	{
		Name:        "canvas_clone_module",
		Description: "Clona e replica integralmente um módulo (com todas as suas páginas wiki, simuladores HTML, tarefas, quizzes e links da Minha Biblioteca) de uma turma de origem para uma turma de destino no Canvas LMS (ex: do 2º para o 4º período).",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"source_course_id": map[string]any{
					"type":        "string",
					"description": "ID numérico, nome da matéria ou período da turma de origem (ex: '2º Período' ou '160754').",
				},
				"dest_course_id": map[string]any{
					"type":        "string",
					"description": "ID numérico, nome da matéria ou período da turma de destino (ex: '4º Período' ou '162263').",
				},
				"module_id_or_name": map[string]any{
					"type":        "string",
					"description": "ID numérico ou nome do módulo a ser clonado (ex: 'Módulo: Controle de Versão Colaborativo (Git & GitHub)').",
				},
				"publish_after_clone": map[string]any{
					"type":        "boolean",
					"description": "Se verdadeiro (padrão), publica os itens clonados no curso de destino.",
				},
			},
			"required": []string{"source_course_id", "dest_course_id", "module_id_or_name"},
		},
	},
	{
		Name:        "canvas_export_course_blueprint",
		Description: "Exporta a matriz didática completa de uma disciplina (módulos, páginas wiki, simuladores, tarefas e questionários) em formato estruturado (JSON e Markdown) para reaproveitamento em futuros semestres (2027.1+) ou replicação em massa.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"course_id": map[string]any{
					"type":        "string",
					"description": "ID numérico, nome da matéria ou período da disciplina a exportar. Se omitido, utiliza a turma ativa do professor no semestre vigente.",
				},
				"output_format": map[string]any{
					"type":        "string",
					"description": "Formato do artefato gerado: 'both' (padrão), 'json' ou 'markdown'.",
				},
			},
		},
	},
	{
		Name:        "canvas_create_question_bank",
		Description: "Cria e organiza um banco de itens institucional por disciplina e competência curricular (DCNs / ENADE) para calibração de questões e sorteio em simulados.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"title": map[string]any{
					"type":        "string",
					"description": "Título do banco de itens (ex: 'Banco de Itens: Estruturas Lineares e Memória').",
				},
				"subject": map[string]any{
					"type":        "string",
					"description": "Disciplina associada (ex: 'Estrutura de Dados', 'Algoritmos').",
				},
				"competence": map[string]any{
					"type":        "string",
					"description": "Competência ou habilidade avaliada no padrão ENADE (ex: 'Alocação Dinâmica e Aritmética de Ponteiros').",
				},
				"description": map[string]any{
					"type":        "string",
					"description": "Descrição pedagógica do propósito deste banco de questões.",
				},
			},
			"required": []string{"title"},
		},
	},
	{
		Name:        "canvas_add_item_to_bank",
		Description: "Adiciona uma questão calibrada no modelo ENADE/CONSEPE a um banco de itens, contendo texto-base contextualizado, situação-problema, gabarito e distratores explicados com justificativa pedagógica imediata.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"bank_id": map[string]any{
					"type":        "string",
					"description": "ID do banco de itens onde a questão será inserida (ex: 'bank_banco_de_itens__estruturas_lineares_e_memoria').",
				},
				"context_text": map[string]any{
					"type":        "string",
					"description": "Texto-base contextualizado com a situação-problema técnica a ser analisada pelo estudante.",
				},
				"competence": map[string]any{
					"type":        "string",
					"description": "Competência específica avaliada neste item.",
				},
				"difficulty": map[string]any{
					"type":        "string",
					"description": "Nível de dificuldade: 'Fácil', 'Médio' ou 'Difícil'.",
				},
				"points": map[string]any{
					"type":        "number",
					"description": "Pontuação padrão da questão (padrão: 10.0).",
				},
				"answers": map[string]any{
					"type":        "array",
					"description": "Lista de alternativas (mínimo 2). Cada alternativa deve conter text, is_correct (boolean) e pedagogical_justification (justificativa de acerto ou explicação do erro do distrator).",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"text": map[string]any{"type": "string"},
							"is_correct": map[string]any{"type": "boolean"},
							"pedagogical_justification": map[string]any{"type": "string"},
						},
						"required": []string{"text", "is_correct"},
					},
				},
			},
			"required": []string{"bank_id", "context_text", "answers"},
		},
	},
	{
		Name:        "canvas_generate_mock_exam",
		Description: "Cria e publica um simulado ou prova formativa no Canvas LMS sorteando aleatoriamente N questões dos bancos de itens ENADE, injetando autocorreção e justificativas pedagógicas imediatas no SpeedGrader.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"course_id": map[string]any{
					"type":        "string",
					"description": "ID numérico, nome da matéria ou período da turma no Canvas. Se omitido, seleciona a turma ativa.",
				},
				"exam_title": map[string]any{
					"type":        "string",
					"description": "Título oficial do simulado/teste no Canvas (ex: 'Simulado Preparatório ENADE / N1').",
				},
				"bank_ids": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "Lista de IDs de bancos de itens para sorteio. Omitir para sortear de todos os bancos cadastrados da disciplina.",
				},
				"pick_count": map[string]any{
					"type":        "integer",
					"description": "Quantidade de questões a serem sorteadas para o simulado (padrão: 10).",
				},
				"points_per_question": map[string]any{
					"type":        "number",
					"description": "Pontos atribuídos a cada questão (padrão: 10.0).",
				},
				"time_limit_minutes": map[string]any{
					"type":        "integer",
					"description": "Tempo limite de realização em minutos (padrão: 60).",
				},
				"due_at": map[string]any{
					"type":        "string",
					"description": "Data e hora de entrega no formato ISO UTC (ex: '2026-09-30T23:59:59Z').",
				},
				"publish_now": map[string]any{
					"type":        "boolean",
					"description": "Se verdadeiro (padrão), publica o quiz imediatamente no Canvas.",
				},
				"target_module_name": map[string]any{
					"type":        "string",
					"description": "Nome ou ID do módulo onde o simulado deve ser inserido (ex: 'Simulados & Avaliações').",
				},
			},
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
		var args struct {
			CurrentTermOnly *bool `json:"current_term_only"`
		}
		if len(rawArgs) > 0 {
			_ = json.Unmarshal(rawArgs, &args)
		}

		pending, err := client.ListPendingAssignments()
		if err != nil {
			return nil, err
		}

		if args.CurrentTermOnly == nil || *args.CurrentTermOnly {
			var currentPending []PendingAssignment
			for _, p := range pending {
				if p.IsCurrentTerm {
					currentPending = append(currentPending, p)
				}
			}
			return currentPending, nil
		}
		return pending, nil

	case "canvas_list_courses":
		var args struct {
			TermFilter string `json:"term_filter"`
			Grouped    bool   `json:"grouped"`
		}
		if len(rawArgs) > 0 {
			_ = json.Unmarshal(rawArgs, &args)
		}

		courses, err := client.ListCourses()
		if err != nil {
			return nil, err
		}

		if args.Grouped {
			var currentCourses, pastCourses []map[string]any
			currentTermName := ""
			for _, crs := range courses {
				if isCur, ok := crs["is_current_term"].(bool); ok && isCur {
					currentCourses = append(currentCourses, crs)
					if currentTermName == "" {
						if tn, ok := crs["term_name"].(string); ok {
							currentTermName = tn
						}
					}
				} else {
					pastCourses = append(pastCourses, crs)
				}
			}
			return map[string]any{
				"current_term":    currentTermName,
				"total_current":   len(currentCourses),
				"current_courses": currentCourses,
				"total_past":      len(pastCourses),
				"past_courses":    pastCourses,
			}, nil
		}

		if args.TermFilter == "current" {
			var filtered []map[string]any
			for _, crs := range courses {
				if isCur, ok := crs["is_current_term"].(bool); ok && isCur {
					filtered = append(filtered, crs)
				}
			}
			return filtered, nil
		} else if args.TermFilter == "past" {
			var filtered []map[string]any
			for _, crs := range courses {
				if isCur, ok := crs["is_current_term"].(bool); !ok || !isCur {
					filtered = append(filtered, crs)
				}
			}
			return filtered, nil
		}

		return courses, nil

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

	case "canvas_create_assignment":
		var params CreateAssignmentParams
		if err := json.Unmarshal(rawArgs, &params); err != nil {
			return nil, err
		}
		return client.CreateAssignment(params)

	case "canvas_create_quiz":
		var params CreateQuizParams
		if err := json.Unmarshal(rawArgs, &params); err != nil {
			return nil, err
		}
		return client.CreateQuiz(params)

	case "canvas_create_module":
		var params CreateModuleParams
		if err := json.Unmarshal(rawArgs, &params); err != nil {
			return nil, err
		}
		return client.CreateModule(params)

	case "canvas_add_module_item":
		var params AddModuleItemParams
		if err := json.Unmarshal(rawArgs, &params); err != nil {
			return nil, err
		}
		return client.AddModuleItem(params)

	case "canvas_list_modules":
		var args struct {
			CourseID string `json:"course_id"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return nil, err
		}
		return client.ListModules(args.CourseID)

	case "canvas_create_page":
		var params CreatePageParams
		if err := json.Unmarshal(rawArgs, &params); err != nil {
			return nil, err
		}
		return client.CreatePage(params)

	case "canvas_list_assignment_groups":
		var args struct {
			CourseID string `json:"course_id"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return nil, err
		}
		return client.ListAssignmentGroups(args.CourseID)

	case "canvas_set_course_weighting":
		var args struct {
			CourseID      string `json:"course_id"`
			EnableWeights bool   `json:"enable_weights"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return nil, err
		}
		return client.SetCourseWeighting(args.CourseID, args.EnableWeights)

	case "canvas_create_assignment_group":
		var args struct {
			CourseID    string  `json:"course_id"`
			Name        string  `json:"name"`
			GroupWeight float64 `json:"group_weight"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return nil, err
		}
		return client.CreateAssignmentGroup(args.CourseID, args.Name, args.GroupWeight)

	case "canvas_update_assignment_group":
		var args struct {
			CourseID    string  `json:"course_id"`
			GroupID     string  `json:"group_id"`
			Name        string  `json:"name"`
			GroupWeight float64 `json:"group_weight"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return nil, err
		}
		return client.UpdateAssignmentGroup(args.CourseID, args.GroupID, args.Name, args.GroupWeight)

	case "canvas_move_assignment_to_group":
		var args struct {
			CourseID     string `json:"course_id"`
			AssignmentID string `json:"assignment_id"`
			GroupID      string `json:"group_id"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return nil, err
		}
		return client.MoveAssignmentToGroup(args.CourseID, args.AssignmentID, args.GroupID)

	case "canvas_get_institutional_rules":
		var args struct {
			Modality string `json:"modality"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return nil, err
		}
		return client.GetInstitutionalRules(args.Modality)

	case "canvas_setup_afya_grading_scheme":
		var args struct {
			CourseID string `json:"course_id"`
			Modality string `json:"modality"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return nil, err
		}
		return client.SetupAfyaGradingScheme(args.CourseID, args.Modality)

	case "canvas_detect_plagiarism":
		var args struct {
			CourseID             string   `json:"course_id"`
			AssignmentID         string   `json:"assignment_id"`
			SimilarityThreshold  *float64 `json:"similarity_threshold,omitempty"`
			NormalizeIdentifiers *bool    `json:"normalize_identifiers,omitempty"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return nil, err
		}
		if args.CourseID == "" || args.AssignmentID == "" {
			return nil, fmt.Errorf("parâmetros 'course_id' e 'assignment_id' são obrigatórios")
		}

		threshold := 65.0
		if args.SimilarityThreshold != nil && *args.SimilarityThreshold > 0 {
			threshold = *args.SimilarityThreshold
		}

		normalizeVars := true
		if args.NormalizeIdentifiers != nil {
			normalizeVars = *args.NormalizeIdentifiers
		}

		return client.CheckSubmissionsSimilarity(args.CourseID, args.AssignmentID, threshold, normalizeVars)

	case "canvas_list_inbox_messages":
		var args struct {
			Scope    string `json:"scope"`
			CourseID string `json:"course_id"`
			Limit    int    `json:"limit"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return nil, err
		}
		if args.Scope == "" {
			args.Scope = "unread"
		}
		return client.ListInboxConversations(args.Scope, args.CourseID, args.Limit)

	case "canvas_get_inbox_conversation":
		var args struct {
			ConversationID string `json:"conversation_id"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return nil, err
		}
		if args.ConversationID == "" {
			return nil, fmt.Errorf("parâmetro 'conversation_id' é obrigatório")
		}
		return client.GetInboxConversation(args.ConversationID)

	case "canvas_reply_inbox_message":
		var args struct {
			ConversationID string `json:"conversation_id"`
			Message        string `json:"message"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return nil, err
		}
		if args.ConversationID == "" || args.Message == "" {
			return nil, fmt.Errorf("parâmetros 'conversation_id' e 'message' são obrigatórios")
		}
		return client.ReplyInboxConversation(args.ConversationID, args.Message)

	case "canvas_send_inbox_message":
		var args struct {
			RecipientIDs []string `json:"recipient_ids"`
			Subject      string   `json:"subject"`
			Message      string   `json:"message"`
			CourseID     string   `json:"course_id"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return nil, err
		}
		if len(args.RecipientIDs) == 0 || args.Subject == "" || args.Message == "" {
			return nil, fmt.Errorf("parâmetros 'recipient_ids', 'subject' e 'message' são obrigatórios")
		}
		return client.SendInboxMessage(args.RecipientIDs, args.Subject, args.Message, args.CourseID)

	case "canvas_clear_cache":
		var args struct {
			CourseID string `json:"course_id"`
		}
		_ = json.Unmarshal(rawArgs, &args)

		if args.CourseID != "" {
			removed := client.ClearCourseCache(args.CourseID)
			return map[string]any{
				"status":           "ok",
				"message":          fmt.Sprintf("Cache da disciplina %s limpo com sucesso (%d chaves removidas)", args.CourseID, removed),
				"keys_removed":     removed,
				"target_course_id": args.CourseID,
			}, nil
		}

		removed := client.ClearCache()
		return map[string]any{
			"status":       "ok",
			"message":      fmt.Sprintf("Cache global em memória limpo com sucesso (%d itens removidos)", removed),
			"keys_removed": removed,
		}, nil

	case "canvas_get_cache_stats":
		stats := client.GetCacheStats()
		return map[string]any{
			"status":      "ok",
			"stats":       stats,
			"description": fmt.Sprintf("%d hits, %d misses (Taxa de acerto: %.1f%%, %d itens em memória)", stats.Hits, stats.Misses, stats.HitRate, stats.ItemCount),
		}, nil

	case "canvas_detect_at_risk_students":
		var args struct {
			CourseID             string   `json:"course_id"`
			InactivityDays       *int     `json:"inactivity_days"`
			GradeCutoff          *float64 `json:"grade_cutoff"`
			ConsecutiveThreshold *int     `json:"consecutive_threshold"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return nil, err
		}
		if strings.TrimSpace(args.CourseID) == "" {
			args.CourseID = "ativa"
		}

		inactDays := 10
		if args.InactivityDays != nil && *args.InactivityDays > 0 {
			inactDays = *args.InactivityDays
		}

		gradeCut := 70.0
		if args.GradeCutoff != nil && *args.GradeCutoff > 0 {
			gradeCut = *args.GradeCutoff
		}

		consecThresh := 2
		if args.ConsecutiveThreshold != nil && *args.ConsecutiveThreshold > 0 {
			consecThresh = *args.ConsecutiveThreshold
		}

		return client.DetectAtRiskStudents(args.CourseID, inactDays, gradeCut, consecThresh)

	case "canvas_get_academic_calendar":
		var args struct {
			Term     string `json:"term"`
			Category string `json:"category"`
			Month    int    `json:"month"`
			Search   string `json:"search"`
			OnlySHE  bool   `json:"only_she"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return nil, err
		}
		return client.GetAcademicCalendar(args.Term, args.Category, args.Month, args.Search, args.OnlySHE)

	case "canvas_get_recommended_reading":
		var args struct {
			CourseID string `json:"course_id"`
			Topic    string `json:"topic"`
			Type     string `json:"type"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return nil, err
		}
		return client.GetRecommendedReadings(args.CourseID, args.Topic, args.Type)

	case "canvas_clone_module":
		var args struct {
			SourceCourseID    string `json:"source_course_id"`
			DestCourseID      string `json:"dest_course_id"`
			ModuleIDOrName    string `json:"module_id_or_name"`
			PublishAfterClone *bool  `json:"publish_after_clone"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return nil, err
		}
		pub := true
		if args.PublishAfterClone != nil {
			pub = *args.PublishAfterClone
		}
		return client.CloneModule(args.SourceCourseID, args.DestCourseID, args.ModuleIDOrName, pub)

	case "canvas_export_course_blueprint":
		var args struct {
			CourseID     string `json:"course_id"`
			OutputFormat string `json:"output_format"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return nil, err
		}
		return client.ExportCourseBlueprint(args.CourseID, args.OutputFormat)

	case "canvas_create_question_bank":
		var args struct {
			Title       string `json:"title"`
			Subject     string `json:"subject"`
			Competence  string `json:"competence"`
			Description string `json:"description"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return nil, err
		}
		return client.CreateQuestionBank(args.Title, args.Subject, args.Competence, args.Description)

	case "canvas_add_item_to_bank":
		var args struct {
			BankID      string        `json:"bank_id"`
			ContextText string        `json:"context_text"`
			Competence  string        `json:"competence"`
			Difficulty  string        `json:"difficulty"`
			Points      float64       `json:"points"`
			Answers     []ENADEAnswer `json:"answers"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return nil, err
		}
		q := ENADEQuestion{
			ContextText: args.ContextText,
			Competence:  args.Competence,
			Difficulty:  args.Difficulty,
			Points:      args.Points,
			Answers:     args.Answers,
		}
		return client.AddItemToBank(args.BankID, q)

	case "canvas_generate_mock_exam":
		var args struct {
			CourseID          string   `json:"course_id"`
			ExamTitle         string   `json:"exam_title"`
			BankIDs           []string `json:"bank_ids"`
			PickCount         int      `json:"pick_count"`
			PointsPerQuestion float64  `json:"points_per_question"`
			TimeLimitMinutes  int      `json:"time_limit_minutes"`
			DueAt             string   `json:"due_at"`
			PublishNow        *bool    `json:"publish_now"`
			TargetModuleName  string   `json:"target_module_name"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return nil, err
		}
		pub := true
		if args.PublishNow != nil {
			pub = *args.PublishNow
		}
		return client.GenerateMockExam(args.CourseID, args.ExamTitle, args.BankIDs, args.PickCount, args.PointsPerQuestion, args.TimeLimitMinutes, args.DueAt, pub, args.TargetModuleName)

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
