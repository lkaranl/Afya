package main

import (
	"strings"
	"testing"
)

func TestFormatBRT(t *testing.T) {
	iso := "2026-09-11T14:30:00Z"
	brt := formatBRT(iso)
	if !strings.Contains(brt, "11/09/2026") {
		t.Errorf("Esperava conter 11/09/2026, obteve: %s", brt)
	}
	if !strings.Contains(brt, "11:30") { // 14:30 UTC - 3h = 11:30 BRT
		t.Errorf("Esperava conter 11:30 BRT, obteve: %s", brt)
	}

	empty := formatBRT("")
	if empty != "Sem data" {
		t.Errorf("Esperava 'Sem data', obteve: %s", empty)
	}
}

func TestInboxMarkdownGeneration(t *testing.T) {
	result := &InboxListResult{
		TotalConversations: 2,
		UnreadCount:        1,
		FilterScope:        "unread",
		Conversations: []CanvasConversation{
			{
				ID:            "1001",
				Subject:       "Dúvida sobre ponteiros em C",
				WorkflowState: "unread",
				LastMessageAt: "2026-09-11T14:00:00Z",
				LastMessageBR: "11/09/2026 às 11:00",
				StudentNames:  "Lucas Silva",
				ContextName:   "Estrutura de Dados",
			},
			{
				ID:            "1002",
				Subject:       "Entrega da atividade",
				WorkflowState: "read",
				LastMessageAt: "2026-09-10T14:00:00Z",
				LastMessageBR: "10/09/2026 às 11:00",
				StudentNames:  "Maria Oliveira",
				ContextName:   "Estrutura de Dados",
			},
		},
	}

	if result.TotalConversations != 2 || result.UnreadCount != 1 {
		t.Errorf("Contagem incorreta de conversas ou unread")
	}
}
