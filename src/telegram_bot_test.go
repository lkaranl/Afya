package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestTelegramBotMockFlow(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_tg.db")

	os.Setenv("ADMIN_TELEGRAM_IDS", "1001")
	os.Setenv("TRIAL_MAX_REQUESTS", "2")

	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Erro ao criar banco: %v", err)
	}
	defer db.Close()

	var sentMessagesMu sync.Mutex
	var sentMessages []map[string]any

	// Servidor Mock da API do Telegram
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if strings.HasSuffix(r.URL.Path, "/getMe") {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"ok": true,
				"result": map[string]any{
					"id":         999,
					"is_bot":     true,
					"first_name": "AfyaBotMock",
					"username":   "afya_bot_mock",
				},
			})
			return
		}

		if strings.HasSuffix(r.URL.Path, "/sendMessage") {
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			sentMessagesMu.Lock()
			sentMessages = append(sentMessages, body)
			sentMessagesMu.Unlock()

			_ = json.NewEncoder(w).Encode(map[string]any{
				"ok": true,
				"result": map[string]any{
					"message_id": 42,
					"chat":       map[string]any{"id": body["chat_id"]},
					"text":       body["text"],
				},
			})
			return
		}

		if strings.HasSuffix(r.URL.Path, "/answerCallbackQuery") {
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "result": true})
			return
		}

		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer mockServer.Close()

	engine := &AgentEngine{}
	bot := NewTelegramBot("fake-token", db, engine)
	bot.apiURL = mockServer.URL // Redireciona chamadas para o mock

	// 1. Testa /start para usuário novo (Trial)
	msgUser := &tgMessage{
		MessageID: 1,
		From:      &tgUser{ID: 2002, FirstName: "João", Username: "prof_joao"},
		Chat:      tgChat{ID: 2002, Type: "private"},
		Text:      "/start",
	}
	bot.handleMessage(msgUser)

	sentMessagesMu.Lock()
	if len(sentMessages) == 0 {
		t.Fatalf("Esperava mensagem enviada após /start")
	}
	firstMsg := sentMessages[0]["text"].(string)
	sentMessagesMu.Unlock()

	if !strings.Contains(firstMsg, "Período de Teste Gratuito") {
		t.Errorf("Esperava menção ao período de teste, obteve: %s", firstMsg)
	}

	// 2. Testa /start para usuário Admin VIP
	msgAdmin := &tgMessage{
		MessageID: 2,
		From:      &tgUser{ID: 1001, FirstName: "Karan", Username: "prof_karan"},
		Chat:      tgChat{ID: 1001, Type: "private"},
		Text:      "/start",
	}
	bot.handleMessage(msgAdmin)

	sentMessagesMu.Lock()
	adminMsg := sentMessages[len(sentMessages)-1]["text"].(string)
	sentMessagesMu.Unlock()

	if !strings.Contains(adminMsg, "Acesso VIP Ilimitado") {
		t.Errorf("Esperava badge VIP para admin 1001, obteve: %s", adminMsg)
	}

	// 3. Testa /status
	msgStatus := &tgMessage{
		MessageID: 3,
		From:      &tgUser{ID: 2002, FirstName: "João", Username: "prof_joao"},
		Chat:      tgChat{ID: 2002, Type: "private"},
		Text:      "/status",
	}
	bot.handleMessage(msgStatus)

	sentMessagesMu.Lock()
	statusReply := sentMessages[len(sentMessages)-1]["text"].(string)
	sentMessagesMu.Unlock()

	if !strings.Contains(statusReply, "Status da sua Conta") {
		t.Errorf("Esperava cabeçalho de status, obteve: %s", statusReply)
	}

	// 4. Testa /ajuda
	msgHelp := &tgMessage{
		MessageID: 4,
		From:      &tgUser{ID: 2002, FirstName: "João"},
		Chat:      tgChat{ID: 2002, Type: "private"},
		Text:      "/ajuda",
	}
	bot.handleMessage(msgHelp)

	sentMessagesMu.Lock()
	helpReply := sentMessages[len(sentMessages)-1]["text"].(string)
	sentMessagesMu.Unlock()

	if !strings.Contains(helpReply, "Guia de Comandos") {
		t.Errorf("Esperava guia de comandos, obteve: %s", helpReply)
	}

	// 5. Testa /assinar
	msgSub := &tgMessage{
		MessageID: 5,
		From:      &tgUser{ID: 2002, FirstName: "João"},
		Chat:      tgChat{ID: 2002, Type: "private"},
		Text:      "/assinar",
	}
	bot.handleMessage(msgSub)

	sentMessagesMu.Lock()
	subReply := sentMessages[len(sentMessages)-1]["text"].(string)
	sentMessagesMu.Unlock()

	if !strings.Contains(subReply, "Assinatura do Plano Professor Pro") {
		t.Errorf("Esperava detalhes do Plano Pro, obteve: %s", subReply)
	}
}
