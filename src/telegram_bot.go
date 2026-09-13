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
	"strconv"
	"strings"
	"sync"
	"time"
)

// TelegramBot gerencia a interação com a Bot API do Telegram via Long Polling
type TelegramBot struct {
	token      string
	apiURL     string
	db         *Database
	engine     *AgentEngine
	httpClient *http.Client
	stopChan   chan struct{}
	wg         sync.WaitGroup

	// Cache em memória de submissões pendentes de aprovação por usuário
	pendingGradesMu sync.Mutex
	pendingGrades   map[int64]*PendingGradeAction
}

// PendingGradeAction armazena os dados de notas aguardando o clique de confirmação do professor
type PendingGradeAction struct {
	CourseID     string
	AssignmentID string
	Grades       []GradeEntry
	CreatedAt    time.Time
}

// Modelos da API do Telegram
type tgUpdate struct {
	UpdateID      int64             `json:"update_id"`
	Message       *tgMessage        `json:"message,omitempty"`
	CallbackQuery *tgCallbackQuery  `json:"callback_query,omitempty"`
}

type tgMessage struct {
	MessageID int64   `json:"message_id"`
	From      *tgUser `json:"from,omitempty"`
	Chat      tgChat  `json:"chat"`
	Text      string  `json:"text,omitempty"`
	Date      int64   `json:"date"`
}

type tgUser struct {
	ID        int64  `json:"id"`
	IsBot     bool   `json:"is_bot"`
	FirstName string `json:"first_name"`
	Username  string `json:"username,omitempty"`
}

type tgChat struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}

type tgCallbackQuery struct {
	ID      string     `json:"id"`
	From    tgUser     `json:"from"`
	Message *tgMessage `json:"message,omitempty"`
	Data    string     `json:"data"`
}

type tgInlineKeyboardMarkup struct {
	InlineKeyboard [][]tgInlineKeyboardButton `json:"inline_keyboard"`
}

type tgInlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data,omitempty"`
	URL          string `json:"url,omitempty"`
}

// NewTelegramBot instancia o bot do Telegram
func NewTelegramBot(token string, db *Database, engine *AgentEngine) *TelegramBot {
	return &TelegramBot{
		token:      token,
		apiURL:     fmt.Sprintf("https://api.telegram.org/bot%s", token),
		db:         db,
		engine:     engine,
		httpClient: &http.Client{Timeout: 60 * time.Second},
		stopChan:   make(chan struct{}),
		pendingGrades: make(map[int64]*PendingGradeAction),
	}
}

// Start inicia o loop de Long Polling do bot
func (b *TelegramBot) Start() error {
	// Valida o token chamando getMe
	me, err := b.getMe()
	if err != nil {
		return fmt.Errorf("falha ao conectar na API do Telegram: %w", err)
	}

	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "data/afya_bot.db"
	}

	fmt.Println("=====================================================")
	fmt.Println("🚀 AFYA CANVAS ASSISTANT - TELEGRAM BOT ONLINE")
	fmt.Println("=====================================================")
	fmt.Printf("🤖 Bot Conectado: @%s (%s, ID: %d)\n", me.Username, me.FirstName, me.ID)
	fmt.Printf("🗄️  Banco SQLite: %s (Pronto e Criptografado)\n", dbPath)
	fmt.Printf("👑 Admins VIPs: %d cadastrados\n", len(b.db.adminIDs))
	fmt.Printf("🎁 Trial Máximo: %d requisições\n", b.db.GetTrialMaxRequests())
	fmt.Println("📡 Status: Long Polling Ativo. Aguardando mensagens no Telegram...")
	fmt.Println("=====================================================")

	// Notifica proativamente os administradores cadastrados avisando que o servidor está online
	go func() {
		for adminID := range b.db.adminIDs {
			b.sendTextMessage(adminID, "🟢 *Afya Canvas Assistant Online!*\n\nO servidor foi iniciado com sucesso e já está pronto para receber suas instruções.\nEnvie /start ou qualquer comando pedagógico para começar.")
		}
	}()

	// Registra os comandos oficiais na Telegram Bot API para autocomplete e botão de Menu
	if err := b.registerCommands(); err != nil {
		log.Printf("[TELEGRAM] Aviso: não foi possível registrar comandos automáticos: %v", err)
	}

	b.wg.Add(1)
	go b.pollingLoop()
	return nil
}

type tgBotCommand struct {
	Command     string `json:"command"`
	Description string `json:"description"`
}

func (b *TelegramBot) registerCommands() error {
	commands := []tgBotCommand{
		{Command: "start", Description: "Início e boas-vindas do assistente"},
		{Command: "graficos", Description: "Painel de métricas e gráficos de rendimento"},
		{Command: "status", Description: "Consulta plano, requisições e Canvas"},
		{Command: "conectar_canvas", Description: "Conecta seu token de acesso ao Canvas"},
		{Command: "assinar", Description: "Detalhes do Plano Pro ilimitado via PIX"},
		{Command: "ajuda", Description: "Guia de comandos e exemplos práticos"},
	}

	payload := map[string]any{
		"commands": commands,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/setMyCommands", b.apiURL)
	resp, err := b.httpClient.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// Stop finaliza graciosamente o bot
func (b *TelegramBot) Stop() {
	close(b.stopChan)
	b.wg.Wait()
	log.Println("🛑 [TELEGRAM] Bot finalizado.")
}

func (b *TelegramBot) pollingLoop() {
	defer b.wg.Done()

	var offset int64 = 0
	for {
		select {
		case <-b.stopChan:
			return
		default:
			updates, err := b.getUpdates(offset, 30)
			if err != nil {
				errStr := err.Error()
				// Timeouts transitórios normais de conexões ociosas em polling são ignorados silenciosamente
				if strings.Contains(errStr, "Client.Timeout") || strings.Contains(errStr, "deadline exceeded") || strings.Contains(errStr, "reset by peer") {
					time.Sleep(1 * time.Second)
					continue
				}
				log.Printf("[TELEGRAM] Reconexão no polling de updates: %v", err)
				time.Sleep(2 * time.Second)
				continue
			}

			for _, u := range updates {
				if u.UpdateID >= offset {
					offset = u.UpdateID + 1
				}

				if u.Message != nil {
					go b.handleMessage(u.Message)
				} else if u.CallbackQuery != nil {
					go b.handleCallbackQuery(u.CallbackQuery)
				}
			}
		}
	}
}

// -----------------------------------------------------------------------
// TRATAMENTO DE MENSAGENS E COMANDOS
// -----------------------------------------------------------------------

func (b *TelegramBot) handleMessage(msg *tgMessage) {
	if msg.From == nil || msg.From.IsBot {
		return
	}

	text := strings.TrimSpace(msg.Text)
	if text == "" {
		return
	}

	telegramID := msg.From.ID
	username := msg.From.Username
	firstName := msg.From.FirstName

	// Registra ou obtém o usuário no banco
	user, err := b.db.GetOrCreateUser(telegramID, username, firstName)
	if err != nil {
		log.Printf("[TELEGRAM] Erro ao obter usuário %d: %v", telegramID, err)
		b.sendTextMessage(msg.Chat.ID, "⚠️ Ocorreu um erro interno ao carregar seu perfil. Tente novamente em instantes.")
		return
	}

	// Se o usuário for Admin e ainda não tiver token ou estiver com nickname, autovincula com o Canvas oficial
	if b.db.IsAdmin(telegramID) && (user.CanvasToken == "" || user.FirstName == username || user.FirstName == "lNaraKl") {
		envToken := os.Getenv("TOKEN")
		envURL := os.Getenv("CANVAS_BASE_URL")
		if envToken != "" {
			if envURL == "" {
				envURL = "https://afya.instructure.com"
			}
			testClient := NewCanvasClient(envURL, envToken)
			if profile, err := testClient.GetUserProfile(); err == nil {
				if pMap, ok := profile.(map[string]any); ok {
					if u, ok := pMap["name"].(string); ok && u != "" {
						user.FirstName = u
						_ = b.db.UpdateUserName(telegramID, u)
					}
				}
			} else {
				user.FirstName = "Karan"
				_ = b.db.UpdateUserName(telegramID, "Karan")
			}
			_ = b.db.UpdateCanvasCredentials(telegramID, envURL, envToken)
			user.CanvasToken = envToken
			user.CanvasURL = envURL
		}
	}

	// Comandos com barra
	if strings.HasPrefix(text, "/") {
		parts := strings.Fields(text)
		cmd := strings.ToLower(parts[0])

		switch cmd {
		case "/start":
			b.cmdStart(msg.Chat.ID, user)
			return
		case "/graficos", "/dashboard", "/metricas":
			b.cmdAnalytics(msg.Chat.ID, user, parts[1:])
			return
		case "/ajuda", "/help":
			b.cmdHelp(msg.Chat.ID, user)
			return
		case "/status", "/perfil":
			b.cmdStatus(msg.Chat.ID, user)
			return
		case "/conectar_canvas", "/conectar", "/token":
			b.cmdConnectCanvas(msg.Chat.ID, user, parts[1:])
			return
		case "/assinar", "/upgrade", "/pix":
			b.cmdSubscribe(msg.Chat.ID, user)
			return
		case "/ping":
			b.sendTextMessage(msg.Chat.ID, "🏓 *Pong!*\n\nBot conectado à API do Telegram e banco de dados SQLite operando normalmente.")
			return
		case "/novo", "/limpar", "/reset":
			_ = b.db.ClearChatHistory(user.TelegramID)
			b.sendTextMessage(msg.Chat.ID, "Tudo bem! Pode me dizer o que precisa agora.")
			return
		default:
			// Comando não reconhecido, continua para IA se for o caso
		}
	}

	// Mensagem de instrução natural para o Agente Pedagógico
	b.processAgentInstruction(msg.Chat.ID, user, text)
}

// /start
func (b *TelegramBot) cmdStart(chatID int64, user *User) {
	isVIP := b.db.IsAdmin(user.TelegramID) || user.PlanTier == "vip"
	
	var planBadge string
	if isVIP {
		planBadge = "⭐️ *Acesso VIP Ilimitado (Administrador)*"
	} else if user.PlanTier == "pro" {
		planBadge = "💎 *Plano Professor Pro Ativo*"
	} else {
		maxTrial := b.db.GetTrialMaxRequests()
		rem := maxTrial - user.RequestsUsed
		if rem < 0 {
			rem = 0
		}
		planBadge = fmt.Sprintf("🎁 *Período de Teste Gratuito:* %d/%d requisições restantes", rem, maxTrial)
	}

	docenteGreeting := "Professor(a)"
	if user.FirstName != "" && user.FirstName != user.Username && user.FirstName != "lNaraKl" {
		docenteGreeting = fmt.Sprintf("Professor(a) *%s*", user.FirstName)
	}

	welcomeMsg := fmt.Sprintf(
		"👋 Olá, %s!\n\n"+
			"Sou seu *Assistente Pedagógico do Canvas LMS*.\n"+
			"Posso auditar pendências, avaliar códigos e respostas no padrão ENADE, criar bancos de itens, responder dúvidas de alunos no Inbox e lançar notas com rapidez e segurança.\n\n"+
			"%s\n\n",
		docenteGreeting, planBadge,
	)

	if user.CanvasToken == "" {
		welcomeMsg += "⚠️ *Etapa Inicial Necessária:*\n"+
			"Para que eu possa interagir com suas turmas, preciso do seu token de acesso da sua instituição de ensino.\n\n"+
			"📖 *Como gerar em 30 segundos:*\n"+
			"1. Acesse o Canvas da sua faculdade no navegador;\n"+
			"2. Clique na sua foto de perfil (*Conta*) > *Configurações*;\n"+
			"3. Role até 'Tokens de acesso aprovados' e clique em *+ Novo token de acesso*;\n"+
			"4. Copie o token gerado e envie para mim assim:\n\n"+
			"`/conectar_canvas SEU_TOKEN_AQUI`\n\n"+
			"_(Seus dados são salvos com criptografia militar AES-256-GCM)_"
	} else {
		welcomeMsg += "✅ *Seu Canvas já está conectado!*\n"+
			"Você pode me enviar qualquer instrução em linguagem natural, por exemplo:\n"+
			"• _\"O que tenho para corrigir nas minhas turmas ativas?\"_\n"+
			"• _\"Crie uma atividade sobre Recursão no módulo correspondente\"_\n"+
			"• _\"Verifique as mensagens não lidas no meu Inbox\"_\n\n"+
			"Para mais opções, digite /ajuda."
	}

	b.sendTextMessage(chatID, welcomeMsg)
}

// /conectar_canvas <token> [canvas_url]
func (b *TelegramBot) cmdConnectCanvas(chatID int64, user *User, args []string) {
	if len(args) == 0 {
		b.sendTextMessage(chatID, "⚠️ *Formato incorreto.*\nEnvie o comando no formato:\n`/conectar_canvas SEU_TOKEN_AQUI`\n\nSe sua faculdade utilizar um endereço diferente de `afya.instructure.com`, informe a URL:\n`/conectar_canvas SEU_TOKEN https://canvas.suafaculdade.edu.br`")
		return
	}

	token := strings.TrimSpace(args[0])
	canvasURL := os.Getenv("CANVAS_DEFAULT_BASE_URL")
	if canvasURL == "" {
		canvasURL = "https://afya.instructure.com"
	}

	if len(args) >= 2 {
		canvasURL = strings.TrimRight(strings.TrimSpace(args[1]), "/")
	}

	// Testa a conexão antes de salvar
	testClient := NewCanvasClient(canvasURL, token)
	profile, err := testClient.GetUserProfile()
	if err != nil {
		b.sendTextMessage(chatID, fmt.Sprintf("❌ *Falha ao validar token no Canvas:*\n%s\n\nVerifique se o token foi copiado corretamente e sem espaços extras.", err.Error()))
		return
	}

	// Salva no banco criptografado
	err = b.db.UpdateCanvasCredentials(user.TelegramID, canvasURL, token)
	if err != nil {
		b.sendTextMessage(chatID, "❌ Ocorreu um erro ao salvar suas credenciais criptografadas.")
		return
	}

	userName := "Docente"
	if pMap, ok := profile.(map[string]any); ok {
		if u, ok := pMap["name"].(string); ok && u != "" {
			userName = u
		}
	}
	_ = b.db.UpdateUserName(user.TelegramID, userName)
	user.FirstName = userName

	b.sendTextMessage(chatID, fmt.Sprintf(
		"🎉 *Canvas Conectado com Sucesso!*\n\n"+
			"👤 *Professor(a):* %s\n"+
			"🏛 *Instituição:* %s\n\n"+
			"Agora já estou pronto para automatizar suas tarefas letivas! Experimente perguntar:\n"+
			"_\"Quais tarefas estão com correções pendentes?\"_",
		userName, canvasURL,
	))
}

// /status
func (b *TelegramBot) cmdStatus(chatID int64, user *User) {
	isVIP := b.db.IsAdmin(user.TelegramID) || user.PlanTier == "vip"

	var statusText string
	if isVIP {
		statusText = "⭐️ *Plano:* VIP Perpétuo (Acesso Livre Ilimitado)\n"
	} else if user.PlanTier == "pro" {
		exp := "Vitalício"
		if user.ExpiresAt != nil {
			exp = user.ExpiresAt.Format("02/01/2006")
		}
		statusText = fmt.Sprintf("💎 *Plano:* Professor Pro (Ativo até %s)\n", exp)
	} else {
		maxTrial := b.db.GetTrialMaxRequests()
		rem := maxTrial - user.RequestsUsed
		if rem < 0 {
			rem = 0
		}
		statusText = fmt.Sprintf("🎁 *Plano:* Período de Teste Gratuito\n📊 *Requisições restantes:* %d de %d\n", rem, maxTrial)
	}

	canvasStatus := "🔴 Não conectado (use /conectar_canvas)"
	if user.CanvasToken != "" {
		canvasStatus = fmt.Sprintf("🟢 Conectado (%s)", user.CanvasURL)
	}

	docenteName := "Professor(a)"
	if user.FirstName != "" && user.FirstName != user.Username && user.FirstName != "lNaraKl" {
		docenteName = user.FirstName
	}

	msg := fmt.Sprintf(
		"📋 *Status da sua Conta:*\n\n"+
			"👤 *Docente:* %s (@%s)\n"+
			"%s"+
			"⚡️ *Requisições realizadas:* %d\n"+
			"🔗 *Canvas LMS:* %s\n",
		docenteName, user.Username, statusText, user.RequestsUsed, canvasStatus,
	)

	var keyboard *tgInlineKeyboardMarkup
	if !isVIP && user.PlanTier != "pro" {
		keyboard = &tgInlineKeyboardMarkup{
			InlineKeyboard: [][]tgInlineKeyboardButton{
				{
					{Text: "💎 Assinar Plano Pro (Ilimitado)", CallbackData: "open_subscribe"},
				},
			},
		}
	}

	b.sendMessage(chatID, msg, keyboard)
}

// /ajuda
func (b *TelegramBot) cmdHelp(chatID int64, user *User) {
	helpText := "🤖 *Guia de Comandos e Instruções:*\n\n" +
		"🔹 *Comandos de Sistema:*\n" +
		"• /start - Início e boas-vindas\n" +
		"• /status - Consulta plano e cota de requisições\n" +
		"• /conectar_canvas - Cadastra seu token do Canvas LMS\n" +
		"• /assinar - Detalhes do Plano Pro ilimitado via PIX\n" +
		"• /ajuda - Este menu de orientação\n\n" +
		"🔹 *Exemplos de Instruções para Enviar:*\n" +
		"• _\"O que tenho para corrigir hoje?\"_\n" +
		"• _\"Analise o status de entrega da turma de Estrutura de Dados\"_\n" +
		"• _\"Crie um simulado ENADE com 5 questões de Git no módulo 2\"_\n" +
		"• _\"Verifique se há dúvidas de alunos pendentes no meu Inbox\"_\n" +
		"• _\"Identifique alunos com risco de evasão nas turmas atuais\"_\n\n" +
		"⚖️ *Segurança das Notas:*\n" +
		"Toda nota e feedback gerado pelo assistente é apresentado primeiro em uma tabela para sua revisão e **só é lançada no Canvas após seu clique no botão de aprovação**."

	b.sendTextMessage(chatID, helpText)
}

// /assinar
func (b *TelegramBot) cmdSubscribe(chatID int64, user *User) {
	priceCents := 4900
	if envPrice := os.Getenv("SUBSCRIPTION_PRICE_CENTS"); envPrice != "" {
		if p, err := strconv.Atoi(envPrice); err == nil && p > 0 {
			priceCents = p
		}
	}
	priceBRL := fmt.Sprintf("R$ %.2f", float64(priceCents)/100.0)

	pixKey := os.Getenv("PAYMENT_PIX_KEY")
	if pixKey == "" {
		pixKey = "financeiro@afyacanvas.com.br"
	}

	msg := fmt.Sprintf(
		"💎 *Assinatura do Plano Professor Pro*\n\n"+
			"Tenha correção em lote ilimitada, radar preditivo de evasão de alunos, geração de questões no padrão ENADE e respostas automáticas no Inbox diretamente pelo Telegram!\n\n"+
			"💰 *Valor:* %s / mês\n\n"+
			"🔑 *Chave PIX:* `%s`\n\n"+
			"📌 *Instruções:*\n"+
			"1. Realize a transferência PIX no valor de %s;\n"+
			"2. O acesso é liberado automaticamente ou envie o comprovante para nosso suporte.",
		priceBRL, pixKey, priceBRL,
	)

	keyboard := &tgInlineKeyboardMarkup{
		InlineKeyboard: [][]tgInlineKeyboardButton{
			{
				{Text: "📋 Copiar Chave PIX", CallbackData: "copy_pix"},
				{Text: "💬 Falar com Suporte", URL: "https://t.me/prof_karan"},
			},
		},
	}

	b.sendMessage(chatID, msg, keyboard)
}

// /graficos
func (b *TelegramBot) cmdAnalytics(chatID int64, user *User, args []string) {
	if user.CanvasToken == "" {
		b.sendTextMessage(chatID, "⚠️ *Canvas não conectado!*\nConecte primeiro seu token de acesso:\n`/conectar_canvas SEU_TOKEN_AQUI`")
		return
	}

	b.sendChatAction(chatID, "typing")
	client := NewCanvasClient(user.CanvasURL, user.CanvasToken)

	courses, err := client.ListCourses()
	if err != nil || len(courses) == 0 {
		b.sendTextMessage(chatID, "⚠️ Nenhuma disciplina ativa encontrada no seu Canvas.")
		return
	}

	var targetCourseID string
	if len(args) > 0 {
		query := strings.ToLower(strings.Join(args, " "))
		for _, c := range courses {
			cName := strings.ToLower(fmt.Sprintf("%v", c["name"]))
			cID := fmt.Sprintf("%v", c["id"])
			if strings.Contains(cName, query) || cID == query {
				targetCourseID = cID
				break
			}
		}
	}

	if targetCourseID == "" {
		targetCourseID = fmt.Sprintf("%v", courses[0]["id"])
	}

	summary, err := BuildCourseAnalytics(client, targetCourseID)
	if err != nil {
		b.sendTextMessage(chatID, fmt.Sprintf("⚠️ Falha ao calcular métricas da disciplina:\n%s", err.Error()))
		return
	}

	text := GenerateTelegramAnalyticsCard(summary)

	webURL := os.Getenv("WEB_DASHBOARD_URL")
	if webURL == "" {
		port := os.Getenv("PORT")
		if port == "" {
			port = "3000"
		}
		webURL = fmt.Sprintf("http://localhost:%s/#analytics-tab", port)
	}

	keyboard := &tgInlineKeyboardMarkup{
		InlineKeyboard: [][]tgInlineKeyboardButton{
			{
				{Text: "🌐 Abrir Gráficos Interativos na Web", URL: webURL},
				{Text: "🔄 Atualizar", CallbackData: fmt.Sprintf("refresh_analytics_%s", targetCourseID)},
			},
		},
	}

	b.sendMessage(chatID, text, keyboard)
}

// -----------------------------------------------------------------------
// PROCESSAMENTO INTELIGENTE DE INSTRUÇÕES (AGENT ENGINE + STREAMING)
// -----------------------------------------------------------------------

func (b *TelegramBot) processAgentInstruction(chatID int64, user *User, instruction string) {
	// 1. Verifica se o Canvas está conectado
	if user.CanvasToken == "" {
		b.sendTextMessage(chatID, "⚠️ *Canvas não conectado!*\nPor favor, conecte seu token de acesso primeiro usando:\n`/conectar_canvas SEU_TOKEN`\n\nDigite /ajuda para saber como gerar seu token.")
		return
	}

	// 2. Valida Cota
	hasAccess, remaining, plan, err := b.db.CheckQuota(user.TelegramID)
	if err != nil || !hasAccess {
		msg := "⚠️ *Limite de Requisições Atingido!*\n\nVocê utilizou todas as suas requisições gratuitas do período de teste.\nPara continuar economizando horas nas suas correções com requisições ilimitadas, assine o *Plano Professor Pro*."
		keyboard := &tgInlineKeyboardMarkup{
			InlineKeyboard: [][]tgInlineKeyboardButton{
				{
					{Text: "💎 Assinar Plano Pro (PIX)", CallbackData: "open_subscribe"},
				},
			},
		}
		b.sendMessage(chatID, msg, keyboard)
		return
	}

	// 3. Envia mensagem inicial de status
	statusMsgID, err := b.sendTextMessageWithID(chatID, "⏳ _Iniciando consulta ao Canvas LMS..._")
	if err != nil {
		statusMsgID = 0
	}

	// Emite ação "typing" periódica em background
	stopTyping := make(chan struct{})
	go func() {
		b.sendChatAction(chatID, "typing")
		ticker := time.NewTicker(4 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-stopTyping:
				return
			case <-ticker.C:
				b.sendChatAction(chatID, "typing")
			}
		}
	}()

	// Instancia cliente Canvas exclusivo do usuário
	userClient := NewCanvasClient(user.CanvasURL, user.CanvasToken)

	// Contexto com timeout de 3 minutos para operações longas
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	var recordedSteps []string
	var actionCard *AgentActionCard
	var finalReply string
	var streamErr error
	lastStepUpdate := time.Now()

	// Callback de streaming para atualizar o Telegram em tempo real
	streamCallback := func(evt StreamEvent) {
		if evt.Type == "status" || evt.Type == "tool" {
			if evt.Text != "" {
				recordedSteps = append(recordedSteps, evt.Text)
				// Atualiza a mensagem a cada 1.5s no máximo para respeitar limites do Telegram
				if statusMsgID != 0 && time.Since(lastStepUpdate) > 1500*time.Millisecond {
					lastStepUpdate = time.Now()
					_ = b.editMessageText(chatID, statusMsgID, fmt.Sprintf("⏳ _%s_", evt.Text))
				}
			}
		} else if evt.Type == "card" {
			actionCard = evt.ActionCard
		} else if evt.Type == "done" {
			finalReply = evt.Text
		} else if evt.Type == "error" {
			streamErr = fmt.Errorf("%s", evt.Text)
		}
	}

	// Cria instância de engine configurada para o cliente do usuário
	userEngine := &AgentEngine{
		Provider:       b.engine.Provider,
		Model:          b.engine.Model,
		SecondaryModel: b.engine.SecondaryModel,
		APIKey:         b.engine.APIKey,
		BaseURL:        b.engine.BaseURL,
		Client:         userClient,
		HTTPClient:     b.httpClient,
	}

	// Resgata o histórico recente de mensagens (memória com sliding window parametrizada)
	chatHistory, _ := b.db.GetChatHistory(user.TelegramID, b.db.GetChatHistoryLimit())

	// Executa agente com streaming aplicando diretriz visual para celular no Telegram
	telegramInstruction := instruction + "\n\n[Diretriz de Formatação para Telegram: Apresente listagens, notas e relatórios em formato de fichas/cards concisos com emojis e divisores de linha '───', em vez de tabelas markdown com pipes '|' que quebram no celular.]"
	chatErr := userEngine.ChatStream(ctx, telegramInstruction, chatHistory, streamCallback)
	close(stopTyping)

	if chatErr == nil && streamErr != nil {
		chatErr = streamErr
	}

	// Remove ou limpa a mensagem de status inicial
	if statusMsgID != 0 {
		b.deleteMessage(chatID, statusMsgID)
	}

	if chatErr != nil {
		log.Printf("[TELEGRAM] Erro no AgentEngine para usuário %d: %v", user.TelegramID, chatErr)
		b.sendTextMessage(chatID, fmt.Sprintf("⚠️ *Ocorreu um erro no processamento:*\n%s", chatErr.Error()))
		return
	}

	// Incrementa contagem de requisições utilizadas
	_ = b.db.IncrementRequests(user.TelegramID)

	// Grava o turno da conversa na memória persistente do usuário
	_ = b.db.SaveChatMessage(user.TelegramID, "user", instruction)
	_ = b.db.SaveChatMessage(user.TelegramID, "assistant", finalReply)

	// Converte Markdown convencional (inclusive tabelas) para cards modernos do Telegram
	replyText := FormatMarkdownForTelegram(finalReply)
	if strings.TrimSpace(replyText) == "" {
		replyText = "Processamento concluído no Canvas LMS com sucesso."
	}

	// Avisa com discrição apenas se o teste gratuito estiver quase no fim
	if plan == "trial" && remaining <= 2 && remaining > 0 {
		replyText += fmt.Sprintf("\n\n_(Nota: restam apenas %d requisições do seu teste gratuito. Digite /assinar para continuar sem interrupções.)_", remaining)
	}

	// Envia a resposta final (separa em blocos se exceder 4000 caracteres)
	b.sendLongTextMessage(chatID, replyText)

	// 5. Se tiver card de aprovação de notas, exibe teclado interativo
	if actionCard != nil {
		b.presentGradeApprovalCard(chatID, user.TelegramID, actionCard)
	}
}

// Apresenta o teclado interativo de aprovação de notas
func (b *TelegramBot) presentGradeApprovalCard(chatID, telegramID int64, card *AgentActionCard) {
	// Faz o unmarshal do payload para obter os dados
	payloadBytes, err := json.Marshal(card.Payload)
	if err != nil {
		return
	}

	var payloadData struct {
		CourseID     string        `json:"course_id"`
		AssignmentID string        `json:"assignment_id"`
		Grades       []GradeEntry `json:"grades"`
	}
	if err := json.Unmarshal(payloadBytes, &payloadData); err != nil {
		return
	}

	b.pendingGradesMu.Lock()
	b.pendingGrades[telegramID] = &PendingGradeAction{
		CourseID:     payloadData.CourseID,
		AssignmentID: payloadData.AssignmentID,
		Grades:       payloadData.Grades,
		CreatedAt:    time.Now(),
	}
	b.pendingGradesMu.Unlock()

	msg := "🛑 *PONTO DE PARADA PARA APROVAÇÃO HUMANA*\n\n" +
		"A tabela de notas e feedbacks acima foi preparada pelo assistente.\n" +
		"Deseja aprovar e gravar oficialmente as notas no SpeedGrader do Canvas?"

	keyboard := &tgInlineKeyboardMarkup{
		InlineKeyboard: [][]tgInlineKeyboardButton{
			{
				{Text: "✅ Aprovar e Publicar no Canvas", CallbackData: "approve_grades"},
				{Text: "❌ Cancelar", CallbackData: "cancel_grades"},
			},
		},
	}

	b.sendMessage(chatID, msg, keyboard)
}

// -----------------------------------------------------------------------
// CALLBACK QUERIES (BOTÕES INLINE)
// -----------------------------------------------------------------------

func (b *TelegramBot) handleCallbackQuery(cb *tgCallbackQuery) {
	data := cb.Data
	telegramID := cb.From.ID
	chatID := cb.From.ID
	if cb.Message != nil {
		chatID = cb.Message.Chat.ID
	}

	b.answerCallbackQuery(cb.ID, "")

	switch data {
	case "open_subscribe":
		user, _ := b.db.GetUser(telegramID)
		if user != nil {
			b.cmdSubscribe(chatID, user)
		}
	case "copy_pix":
		pixKey := os.Getenv("PAYMENT_PIX_KEY")
		if pixKey == "" {
			pixKey = "financeiro@afyacanvas.com.br"
		}
		b.sendTextMessage(chatID, fmt.Sprintf("🔑 *Chave PIX para cópia:*\n`%s`", pixKey))

	case "cancel_grades":
		b.pendingGradesMu.Lock()
		delete(b.pendingGrades, telegramID)
		b.pendingGradesMu.Unlock()
		if cb.Message != nil {
			_ = b.editMessageText(chatID, cb.Message.MessageID, "❌ Publicação de notas cancelada. Nenhuma alteração foi realizada no Canvas LMS.")
		}

	case "approve_grades":
		b.pendingGradesMu.Lock()
		action, exists := b.pendingGrades[telegramID]
		if exists {
			delete(b.pendingGrades, telegramID)
		}
		b.pendingGradesMu.Unlock()

		if !exists || action == nil {
			b.sendTextMessage(chatID, "⚠️ Nenhuma operação de notas pendente de aprovação encontrada ou a sessão expirou.")
			return
		}

		user, err := b.db.GetUser(telegramID)
		if err != nil || user.CanvasToken == "" {
			b.sendTextMessage(chatID, "❌ Erro ao resgatar credenciais do Canvas para envio.")
			return
		}

		if cb.Message != nil {
			_ = b.editMessageText(chatID, cb.Message.MessageID, "⏳ Gravando notas e feedbacks no Canvas SpeedGrader...")
		}

		client := NewCanvasClient(user.CanvasURL, user.CanvasToken)
		res, err := client.SubmitGradesBatch(action.CourseID, action.AssignmentID, action.Grades)
		if err != nil {
			b.sendTextMessage(chatID, fmt.Sprintf("❌ Falha ao publicar notas no Canvas: %s", err.Error()))
			return
		}

		successCount := len(action.Grades)
		if resMap, ok := res.(map[string]any); ok {
			if cnt, ok := resMap["grades_submitted"].(int); ok {
				successCount = cnt
			}
		}

		confirmMsg := fmt.Sprintf(
			"🎉 *Notas Publicadas com Sucesso!*\n\n"+
				"✅ Total de alunos avaliados: *%d*\n"+
				"Notas e comentários pedagógicos já estão visíveis no SpeedGrader da disciplina.",
			successCount,
		)
		if cb.Message != nil {
			_ = b.editMessageText(chatID, cb.Message.MessageID, confirmMsg)
		} else {
			b.sendTextMessage(chatID, confirmMsg)
		}

	default:
		if strings.HasPrefix(cb.Data, "refresh_analytics_") {
			courseID := strings.TrimPrefix(cb.Data, "refresh_analytics_")
			user, err := b.db.GetUser(telegramID)
			if err == nil && user != nil && user.CanvasToken != "" {
				client := NewCanvasClient(user.CanvasURL, user.CanvasToken)
				if summary, err := BuildCourseAnalytics(client, courseID); err == nil {
					text := GenerateTelegramAnalyticsCard(summary)
					if cb.Message != nil {
						webURL := os.Getenv("WEB_DASHBOARD_URL")
						if webURL == "" {
							port := os.Getenv("PORT")
							if port == "" {
								port = "3000"
							}
							webURL = fmt.Sprintf("http://localhost:%s/#analytics-tab", port)
						}
						keyboard := &tgInlineKeyboardMarkup{
							InlineKeyboard: [][]tgInlineKeyboardButton{
								{
									{Text: "🌐 Abrir Gráficos Interativos na Web", URL: webURL},
									{Text: "🔄 Atualizar", CallbackData: fmt.Sprintf("refresh_analytics_%s", courseID)},
								},
							},
						}
						_ = b.editMessageTextAndMarkup(chatID, cb.Message.MessageID, text, keyboard)
					}
				}
			}
		}
	}
}

// -----------------------------------------------------------------------
// MÉTODOS HTTP DE BAIXO NÍVEL DA API TELEGRAM
// -----------------------------------------------------------------------

func (b *TelegramBot) getMe() (*tgUser, error) {
	url := fmt.Sprintf("%s/getMe", b.apiURL)
	resp, err := b.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var res struct {
		OK     bool    `json:"ok"`
		Result *tgUser `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}
	if !res.OK || res.Result == nil {
		return nil, fmt.Errorf("resposta inválida do getMe")
	}
	return res.Result, nil
}

func (b *TelegramBot) getUpdates(offset int64, timeout int) ([]tgUpdate, error) {
	url := fmt.Sprintf("%s/getUpdates?offset=%d&timeout=%d", b.apiURL, offset, timeout)
	resp, err := b.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var res struct {
		OK     bool       `json:"ok"`
		Result []tgUpdate `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}
	return res.Result, nil
}

func (b *TelegramBot) sendTextMessage(chatID int64, text string) {
	b.sendMessage(chatID, text, nil)
}

func (b *TelegramBot) sendTextMessageWithID(chatID int64, text string) (int64, error) {
	payload := map[string]any{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "Markdown",
	}
	body, _ := json.Marshal(payload)
	resp, err := b.httpClient.Post(fmt.Sprintf("%s/sendMessage", b.apiURL), "application/json", bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var res struct {
		OK     bool      `json:"ok"`
		Result tgMessage `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err == nil && res.OK {
		return res.Result.MessageID, nil
	}
	return 0, fmt.Errorf("falha ao enviar")
}

func (b *TelegramBot) sendMessage(chatID int64, text string, keyboard *tgInlineKeyboardMarkup) {
	payload := map[string]any{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "Markdown",
	}
	if keyboard != nil {
		payload["reply_markup"] = keyboard
	}

	body, _ := json.Marshal(payload)
	resp, err := b.httpClient.Post(fmt.Sprintf("%s/sendMessage", b.apiURL), "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("[TELEGRAM] Erro ao enviar mensagem: %v", err)
		return
	}
	defer resp.Body.Close()

	// Se falhar o Markdown do Telegram, tenta reenviar como texto simples
	if resp.StatusCode != http.StatusOK {
		delete(payload, "parse_mode")
		plainBody, _ := json.Marshal(payload)
		_, _ = b.httpClient.Post(fmt.Sprintf("%s/sendMessage", b.apiURL), "application/json", bytes.NewReader(plainBody))
	}
}

func (b *TelegramBot) sendLongTextMessage(chatID int64, text string) {
	const maxLen = 3800 // Limite de segurança abaixo de 4096
	if len(text) <= maxLen {
		b.sendTextMessage(chatID, text)
		return
	}

	// Divide em partes
	lines := strings.Split(text, "\n")
	var currentChunk strings.Builder

	for _, line := range lines {
		if currentChunk.Len()+len(line)+1 > maxLen {
			b.sendTextMessage(chatID, currentChunk.String())
			currentChunk.Reset()
		}
		currentChunk.WriteString(line)
		currentChunk.WriteString("\n")
	}

	if currentChunk.Len() > 0 {
		b.sendTextMessage(chatID, currentChunk.String())
	}
}

func (b *TelegramBot) editMessageText(chatID int64, messageID int64, newText string) error {
	return b.editMessageTextAndMarkup(chatID, messageID, newText, nil)
}

func (b *TelegramBot) editMessageTextAndMarkup(chatID int64, messageID int64, newText string, keyboard *tgInlineKeyboardMarkup) error {
	payload := map[string]any{
		"chat_id":    chatID,
		"message_id": messageID,
		"text":       newText,
		"parse_mode": "Markdown",
	}
	if keyboard != nil {
		payload["reply_markup"] = keyboard
	}
	body, _ := json.Marshal(payload)
	resp, err := b.httpClient.Post(fmt.Sprintf("%s/editMessageText", b.apiURL), "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (b *TelegramBot) deleteMessage(chatID int64, messageID int64) {
	url := fmt.Sprintf("%s/deleteMessage?chat_id=%d&message_id=%d", b.apiURL, chatID, messageID)
	resp, err := b.httpClient.Get(url)
	if err == nil && resp != nil {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

func (b *TelegramBot) sendChatAction(chatID int64, action string) {
	url := fmt.Sprintf("%s/sendChatAction?chat_id=%d&action=%s", b.apiURL, chatID, action)
	resp, err := b.httpClient.Get(url)
	if err == nil && resp != nil {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

func (b *TelegramBot) answerCallbackQuery(callbackQueryID string, text string) {
	payload := map[string]any{
		"callback_query_id": callbackQueryID,
	}
	if text != "" {
		payload["text"] = text
	}
	body, _ := json.Marshal(payload)
	resp, err := b.httpClient.Post(fmt.Sprintf("%s/answerCallbackQuery", b.apiURL), "application/json", bytes.NewReader(body))
	if err == nil && resp != nil {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}
