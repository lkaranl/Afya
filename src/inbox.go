package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
)

// CanvasParticipant representa um participante em uma conversa do Canvas LMS
type CanvasParticipant struct {
	ID        any    `json:"id"`
	Name      string `json:"name"`
	FullName  string `json:"full_name"`
	AvatarURL string `json:"avatar_url"`
}

// GetIDString retorna o ID do participante como string
func (p CanvasParticipant) GetIDString() string {
	return fmt.Sprintf("%v", p.ID)
}

// CanvasMessageAttachment representa um arquivo anexado à mensagem
type CanvasMessageAttachment struct {
	ID          any    `json:"id"`
	DisplayName string `json:"display_name"`
	Filename    string `json:"filename"`
	URL         string `json:"url"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
}

// CanvasInboxMessage representa uma mensagem individual dentro da thread da conversa
type CanvasInboxMessage struct {
	ID           any                       `json:"id"`
	CreatedAt    string                    `json:"created_at"`
	CreatedAtBR  string                    `json:"created_at_br"`
	Body         string                    `json:"body"`
	AuthorID     any                       `json:"author_id"`
	AuthorName   string                    `json:"author_name"`
	Generated    bool                      `json:"generated"`
	Attachments  []CanvasMessageAttachment `json:"attachments"`
	Forwarded    bool                      `json:"forwarded"`
}

// CanvasConversation representa o resumo de uma conversa no Inbox do Canvas
type CanvasConversation struct {
	ID             any                 `json:"id"`
	Subject        string              `json:"subject"`
	WorkflowState  string              `json:"workflow_state"` // "unread", "read", "archived"
	LastMessage    string              `json:"last_message"`
	LastMessageAt  string              `json:"last_message_at"`
	LastMessageBR  string              `json:"last_message_br"`
	MessageCount   int                 `json:"message_count"`
	Starred        bool                `json:"starred"`
	ContextName    string              `json:"context_name"`
	Participants   []CanvasParticipant `json:"participants"`
	StudentNames   string              `json:"student_names"`
}

// GetIDString retorna o ID da conversa como string
func (c CanvasConversation) GetIDString() string {
	return fmt.Sprintf("%v", c.ID)
}

// InboxListResult é o resultado consolidado da listagem de mensagens com triagem
type InboxListResult struct {
	TotalConversations int                   `json:"total_conversations"`
	UnreadCount        int                   `json:"unread_count"`
	FilterScope        string                `json:"filter_scope"`
	FilterCourseID     string                `json:"filter_course_id,omitempty"`
	Conversations      []CanvasConversation `json:"conversations"`
	MarkdownTable      string                `json:"markdown_table"`
}

// ConversationDetailResult é o resultado detalhado de uma conversa com toda a thread
type ConversationDetailResult struct {
	ID             string                `json:"id"`
	Subject        string                `json:"subject"`
	WorkflowState  string                `json:"workflow_state"`
	ContextName    string                `json:"context_name"`
	MessageCount   int                   `json:"message_count"`
	Participants   []CanvasParticipant   `json:"participants"`
	Messages       []CanvasInboxMessage  `json:"messages"`
	MarkdownThread string                `json:"markdown_thread"`
}

// formatBRT converte data ISO UTC para horário de Brasília (UTC-3)
func formatBRT(isoDate string) string {
	if isoDate == "" || isoDate == "<nil>" {
		return "Sem data"
	}
	t, err := time.Parse(time.RFC3339, isoDate)
	if err != nil {
		return isoDate
	}
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		loc = time.FixedZone("BRT", -3*60*60)
	}
	return t.In(loc).Format("02/01/2006 às 15:04")
}

// ListInboxConversations busca as mensagens do Inbox com triagem de dúvidas
func (c *CanvasClient) ListInboxConversations(scope, courseID string, limit int) (*InboxListResult, error) {
	if limit <= 0 {
		limit = 25
	}

	endpoint := fmt.Sprintf("/api/v1/conversations?per_page=%d", limit)
	if scope != "" && scope != "inbox" {
		endpoint += fmt.Sprintf("&scope=%s", url.QueryEscape(scope))
	}
	if courseID != "" {
		endpoint += fmt.Sprintf("&filter=course_%s", url.QueryEscape(courseID))
	}

	data, status, err := c.Request("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("falha ao listar conversas do Canvas (HTTP %d): %w", status, err)
	}

	var rawList []map[string]any
	if err := json.Unmarshal(data, &rawList); err != nil {
		return nil, fmt.Errorf("falha ao interpretar resposta do Inbox do Canvas: %w", err)
	}

	var conversations []CanvasConversation
	unreadCount := 0

	for _, item := range rawList {
		lastMsg := fmt.Sprintf("%v", item["last_message"])
		if lastMsg == "" || lastMsg == "<nil>" {
			lastMsg = fmt.Sprintf("%v", item["last_authored_message"])
		}

		lastMsgAt := fmt.Sprintf("%v", item["last_message_at"])
		if lastMsgAt == "" || lastMsgAt == "<nil>" {
			lastMsgAt = fmt.Sprintf("%v", item["last_authored_message_at"])
		}

		conv := CanvasConversation{
			ID:            item["id"],
			Subject:       fmt.Sprintf("%v", item["subject"]),
			WorkflowState: fmt.Sprintf("%v", item["workflow_state"]),
			LastMessage:   lastMsg,
			LastMessageAt: lastMsgAt,
			ContextName:   fmt.Sprintf("%v", item["context_name"]),
		}

		if countVal, ok := item["message_count"].(float64); ok {
			conv.MessageCount = int(countVal)
		}
		if starredVal, ok := item["starred"].(bool); ok {
			conv.Starred = starredVal
		}

		conv.LastMessageBR = formatBRT(conv.LastMessageAt)
		if conv.WorkflowState == "unread" {
			unreadCount++
		}

		// Participantes
		if parts, ok := item["participants"].([]any); ok {
			var studentNames []string
			for _, p := range parts {
				if pMap, ok := p.(map[string]any); ok {
					participant := CanvasParticipant{
						ID:        pMap["id"],
						Name:      fmt.Sprintf("%v", pMap["name"]),
						FullName:  fmt.Sprintf("%v", pMap["full_name"]),
						AvatarURL: fmt.Sprintf("%v", pMap["avatar_url"]),
					}
					conv.Participants = append(conv.Participants, participant)
					if participant.Name != "" && participant.Name != "<nil>" {
						studentNames = append(studentNames, participant.Name)
					}
				}
			}
			conv.StudentNames = strings.Join(studentNames, ", ")
		}

		conversations = append(conversations, conv)
	}

	// Ordenar pelas mais recentes primeiro
	sort.Slice(conversations, func(i, j int) bool {
		return conversations[i].LastMessageAt > conversations[j].LastMessageAt
	})

	// Gerar tabela formatada em Markdown para o professor/agente
	var md strings.Builder
	md.WriteString(fmt.Sprintf("### 📬 Inbox do Canvas LMS: Triagem de Mensagens (%d conversas", len(conversations)))
	if unreadCount > 0 {
		md.WriteString(fmt.Sprintf(", **%d não lidas**", unreadCount))
	}
	md.WriteString(")\n\n")

	if len(conversations) == 0 {
		md.WriteString("Nenhuma mensagem encontrada para o filtro informado.\n")
	} else {
		md.WriteString("| ID Conversa | Remetente / Alunos | Assunto | Disciplina | Data (BRT) | Status |\n")
		md.WriteString("| :--- | :--- | :--- | :--- | :--- | :--- |\n")
		for _, conv := range conversations {
			statusIcon := "🟢 Lida"
			if conv.WorkflowState == "unread" {
				statusIcon = "🔴 **NÃO LIDA**"
			}
			subject := conv.Subject
			if subject == "" || subject == "<nil>" {
				subject = "(Sem assunto)"
			}
			if len(subject) > 35 {
				subject = subject[:32] + "..."
			}

			students := conv.StudentNames
			if len(students) > 30 {
				students = students[:27] + "..."
			}

			ctxName := conv.ContextName
			if ctxName == "<nil>" || ctxName == "" {
				ctxName = "Geral"
			}
			if len(ctxName) > 25 {
				ctxName = ctxName[:22] + "..."
			}

			md.WriteString(fmt.Sprintf("| `%s` | %s | %s | %s | %s | %s |\n",
				conv.GetIDString(), students, subject, ctxName, conv.LastMessageBR, statusIcon))
		}
	}

	result := &InboxListResult{
		TotalConversations: len(conversations),
		UnreadCount:        unreadCount,
		FilterScope:        scope,
		FilterCourseID:     courseID,
		Conversations:      conversations,
		MarkdownTable:      md.String(),
	}

	return result, nil
}

// GetInboxConversation obtém a íntegra de uma conversa e de todas as mensagens trocadas
func (c *CanvasClient) GetInboxConversation(conversationID string) (*ConversationDetailResult, error) {
	if conversationID == "" {
		return nil, fmt.Errorf("ID da conversa não fornecido")
	}

	endpoint := fmt.Sprintf("/api/v1/conversations/%s", url.PathEscape(conversationID))
	data, status, err := c.Request("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar conversa %s no Canvas (HTTP %d): %w", conversationID, status, err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("falha ao interpretar detalhes da conversa: %w", err)
	}

	result := &ConversationDetailResult{
		ID:            fmt.Sprintf("%v", raw["id"]),
		Subject:       fmt.Sprintf("%v", raw["subject"]),
		WorkflowState: fmt.Sprintf("%v", raw["workflow_state"]),
		ContextName:   fmt.Sprintf("%v", raw["context_name"]),
	}

	// Mapear participantes por ID para exibir os nomes dos autores
	participantMap := make(map[string]string)
	if parts, ok := raw["participants"].([]any); ok {
		for _, p := range parts {
			if pMap, ok := p.(map[string]any); ok {
				pID := fmt.Sprintf("%v", pMap["id"])
				pName := fmt.Sprintf("%v", pMap["name"])
				participantMap[pID] = pName
				result.Participants = append(result.Participants, CanvasParticipant{
					ID:        pID,
					Name:      pName,
					FullName:  fmt.Sprintf("%v", pMap["full_name"]),
					AvatarURL: fmt.Sprintf("%v", pMap["avatar_url"]),
				})
			}
		}
	}

	// Mapear mensagens da thread
	if msgs, ok := raw["messages"].([]any); ok {
		for _, m := range msgs {
			if mMap, ok := m.(map[string]any); ok {
				authorID := fmt.Sprintf("%v", mMap["author_id"])
				authorName := participantMap[authorID]
				if authorName == "" {
					authorName = "Desconhecido"
				}

				msg := CanvasInboxMessage{
					ID:          mMap["id"],
					CreatedAt:   fmt.Sprintf("%v", mMap["created_at"]),
					CreatedAtBR: formatBRT(fmt.Sprintf("%v", mMap["created_at"])),
					Body:        fmt.Sprintf("%v", mMap["body"]),
					AuthorID:    authorID,
					AuthorName:  authorName,
				}

				if gen, ok := mMap["generated"].(bool); ok {
					msg.Generated = gen
				}
				if fwd, ok := mMap["forwarded"].(bool); ok {
					msg.Forwarded = fwd
				}

				// Anexos da mensagem
				if atts, ok := mMap["attachments"].([]any); ok {
					for _, a := range atts {
						if aMap, ok := a.(map[string]any); ok {
							var attSize int64
							if sVal, ok := aMap["size"].(float64); ok {
								attSize = int64(sVal)
							}
							msg.Attachments = append(msg.Attachments, CanvasMessageAttachment{
								ID:          aMap["id"],
								DisplayName: fmt.Sprintf("%v", aMap["display_name"]),
								Filename:    fmt.Sprintf("%v", aMap["filename"]),
								URL:         fmt.Sprintf("%v", aMap["url"]),
								Size:        attSize,
								ContentType: fmt.Sprintf("%v", aMap["content_type"]),
							})
						}
					}
				}

				result.Messages = append(result.Messages, msg)
			}
		}
	}

	result.MessageCount = len(result.Messages)

	// Ordenar mensagens cronologicamente (da mais antiga para a mais recente)
	sort.Slice(result.Messages, func(i, j int) bool {
		return result.Messages[i].CreatedAt < result.Messages[j].CreatedAt
	})

	// Gerar visualização em Markdown da Thread completa
	var md strings.Builder
	md.WriteString(fmt.Sprintf("## 💬 Conversa do Canvas: %s\n", result.Subject))
	md.WriteString(fmt.Sprintf("**ID da Conversa:** `%s` | **Contexto:** %s | **Mensagens:** %d\n\n",
		result.ID, result.ContextName, result.MessageCount))

	md.WriteString("### 👥 Participantes:\n")
	for _, p := range result.Participants {
		md.WriteString(fmt.Sprintf("- **%s** (ID: `%s`)\n", p.Name, p.GetIDString()))
	}
	md.WriteString("\n---\n\n### 📜 Histórico de Mensagens:\n\n")

	for i, msg := range result.Messages {
		md.WriteString(fmt.Sprintf("#### Mensagem %d de %d — **%s** (%s)\n",
			i+1, len(result.Messages), msg.AuthorName, msg.CreatedAtBR))
		md.WriteString(fmt.Sprintf("> %s\n\n", strings.ReplaceAll(msg.Body, "\n", "\n> ")))

		if len(msg.Attachments) > 0 {
			md.WriteString("**📎 Anexos:**\n")
			for _, att := range msg.Attachments {
				md.WriteString(fmt.Sprintf("- [%s](%s) (%d bytes)\n", att.DisplayName, att.URL, att.Size))
			}
			md.WriteString("\n")
		}
	}

	result.MarkdownThread = md.String()
	return result, nil
}

// ReplyInboxConversation envia uma resposta em uma conversa existente
func (c *CanvasClient) ReplyInboxConversation(conversationID, message string) (any, error) {
	if conversationID == "" {
		return nil, fmt.Errorf("ID da conversa é obrigatório")
	}
	if strings.TrimSpace(message) == "" {
		return nil, fmt.Errorf("conteúdo da mensagem não pode ser vazio")
	}

	endpoint := fmt.Sprintf("/api/v1/conversations/%s/add_message", url.PathEscape(conversationID))
	payload := map[string]any{
		"body": message,
	}

	data, status, err := c.Request("POST", endpoint, payload)
	if err != nil {
		return nil, fmt.Errorf("falha ao enviar resposta no Canvas (HTTP %d): %w", status, err)
	}

	var resp any
	if err := json.Unmarshal(data, &resp); err != nil {
		return map[string]any{"status": "ok", "message": "Resposta enviada com sucesso"}, nil
	}

	return resp, nil
}

// SendInboxMessage envia uma nova mensagem direta privada para um ou mais alunos
func (c *CanvasClient) SendInboxMessage(recipientIDs []string, subject, message, courseID string) (any, error) {
	if len(recipientIDs) == 0 {
		return nil, fmt.Errorf("é necessário especificar ao menos um destinatário (recipient_ids)")
	}
	if strings.TrimSpace(subject) == "" {
		return nil, fmt.Errorf("assunto da mensagem é obrigatório")
	}
	if strings.TrimSpace(message) == "" {
		return nil, fmt.Errorf("corpo da mensagem é obrigatório")
	}

	endpoint := "/api/v1/conversations"
	payload := map[string]any{
		"recipients":         recipientIDs,
		"subject":            subject,
		"body":               message,
		"group_conversation": false,
	}

	if courseID != "" {
		payload["context_code"] = fmt.Sprintf("course_%s", courseID)
	}

	data, status, err := c.Request("POST", endpoint, payload)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar conversa no Canvas (HTTP %d): %w", status, err)
	}

	var resp any
	if err := json.Unmarshal(data, &resp); err != nil {
		return map[string]any{"status": "ok", "message": "Mensagem enviada com sucesso"}, nil
	}

	return resp, nil
}
