package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// User representa um usuário/professor no banco de dados
type User struct {
	TelegramID   int64      `json:"telegram_id"`
	Username     string     `json:"username"`
	FirstName    string     `json:"first_name"`
	CanvasURL    string     `json:"canvas_url"`
	CanvasToken  string     `json:"-"` // Mantido apenas em memória após decriptografia
	PlanTier     string     `json:"plan_tier"` // "trial", "pro", "vip"
	RequestsUsed int        `json:"requests_used"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// Database gerencia a persistência em SQLite
type Database struct {
	db           *sql.DB
	encKey       []byte
	adminIDs     map[int64]bool
	trialMax     int
	historyLimit int
	mu           sync.RWMutex
}

// NewDatabase inicializa a conexão com o SQLite e aplica as migrações necessárias
func NewDatabase(dbPath string) (*Database, error) {
	if dbPath == "" {
		dbPath = os.Getenv("DATABASE_PATH")
		if dbPath == "" {
			dbPath = "data/afya_bot.db"
		}
	}

	// Garante que o diretório pai exista
	dir := filepath.Dir(dbPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("erro ao criar diretório do banco de dados '%s': %w", dir, err)
		}
	}

	db, err := sql.Open("sqlite", dbPath+"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, fmt.Errorf("falha ao abrir banco de dados SQLite '%s': %w", dbPath, err)
	}

	// Configuração do pool de conexões para SQLite com WAL
	db.SetMaxOpenConns(1) // SQLite opera melhor com 1 writer concorrente
	db.SetMaxIdleConns(1)

	// Prepara chave de criptografia AES-256 (32 bytes)
	rawKey := os.Getenv("DATABASE_ENCRYPTION_KEY")
	if rawKey == "" {
		log.Println("⚠️ [DATABASE] ATENÇÃO: 'DATABASE_ENCRYPTION_KEY' não configurada no .env. Utilizando chave derivada interna para desenvolvimento.")
		rawKey = "afya-canvas-secret-key-dev-salt-2026"
	}
	hash := sha256.Sum256([]byte(rawKey))
	encKey := hash[:]

	// Carrega lista de administradores/VIPs do .env
	adminStr := os.Getenv("ADMIN_TELEGRAM_IDS")
	if notifierID := os.Getenv("TELEGRAM_NOTIFIER_CHAT_ID"); notifierID != "" {
		if adminStr != "" {
			adminStr += "," + notifierID
		} else {
			adminStr = notifierID
		}
	}
	adminIDs := parseAdminIDs(adminStr)

	// Carrega limite de requisições de trial (padrão: 20)
	trialMax := 20
	if envTrial := os.Getenv("TRIAL_MAX_REQUESTS"); envTrial != "" {
		if v, err := strconv.Atoi(envTrial); err == nil && v > 0 {
			trialMax = v
		}
	}

	// Carrega limite de mensagens do histórico de contexto (padrão: 10 mensagens / 5 turnos)
	historyLimit := 10
	if envHist := os.Getenv("CHAT_HISTORY_LIMIT"); envHist != "" {
		if v, err := strconv.Atoi(envHist); err == nil && v > 0 {
			historyLimit = v
		}
	}

	database := &Database{
		db:           db,
		encKey:       encKey,
		adminIDs:     adminIDs,
		trialMax:     trialMax,
		historyLimit: historyLimit,
	}

	if err := database.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("erro ao executar migrações no banco: %w", err)
	}

	log.Printf("[DATABASE] Banco de dados SQLite inicializado em '%s' (Trial Max: %d requisições, Histórico Contexto: %d msgs, %d Admins VIPs configurados)", dbPath, trialMax, historyLimit, len(adminIDs))
	return database, nil
}

// Close fecha a conexão com o banco de dados
func (d *Database) Close() error {
	return d.db.Close()
}

// migrate cria as tabelas essenciais caso não existam
func (d *Database) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			telegram_id INTEGER PRIMARY KEY,
			username TEXT,
			first_name TEXT,
			canvas_url TEXT DEFAULT 'https://afya.instructure.com',
			canvas_token_encrypted BLOB,
			canvas_token_iv BLOB,
			plan_tier TEXT DEFAULT 'trial',
			requests_used INTEGER DEFAULT 0,
			expires_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS payments (
			id TEXT PRIMARY KEY,
			telegram_id INTEGER,
			amount_cents INTEGER,
			status TEXT DEFAULT 'pending',
			pix_copy_paste TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			paid_at DATETIME,
			FOREIGN KEY(telegram_id) REFERENCES users(telegram_id)
		);`,
		`CREATE TABLE IF NOT EXISTS chat_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			telegram_id INTEGER,
			role TEXT,
			content TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(telegram_id) REFERENCES users(telegram_id)
		);`,
		`CREATE INDEX IF NOT EXISTS idx_chat_history_user ON chat_history(telegram_id, id DESC);`,
	}

	for _, q := range queries {
		if _, err := d.db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

// parseAdminIDs faz o parse da lista separada por vírgula no .env
func parseAdminIDs(raw string) map[int64]bool {
	res := make(map[int64]bool)
	if raw == "" {
		return res
	}
	parts := strings.Split(raw, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if id, err := strconv.ParseInt(p, 10, 64); err == nil && id != 0 {
			res[id] = true
		}
	}
	return res
}

// IsAdmin verifica se um telegramID é admin/VIP perpétuo configurado
func (d *Database) IsAdmin(telegramID int64) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.adminIDs[telegramID]
}

// GetTrialMaxRequests retorna o limite atual de requisições de teste
func (d *Database) GetTrialMaxRequests() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.trialMax
}

// GetOrCreateUser localiza ou registra um novo usuário
func (d *Database) GetOrCreateUser(telegramID int64, username, firstName string) (*User, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	user, err := d.getUserLocked(telegramID)
	if err == nil && user != nil {
		// Atualiza apenas o @username do Telegram se tiver mudado (o nome do docente pertence ao Canvas LMS)
		if user.Username != username {
			_, _ = d.db.Exec("UPDATE users SET username = ?, updated_at = CURRENT_TIMESTAMP WHERE telegram_id = ?",
				username, telegramID)
			user.Username = username
		}
		return user, nil
	}

	// Define plano inicial: se estiver na lista de admins, já nasce VIP ilimitado
	planTier := "trial"
	if d.adminIDs[telegramID] {
		planTier = "vip"
	}

	defaultCanvasURL := os.Getenv("CANVAS_DEFAULT_BASE_URL")
	if defaultCanvasURL == "" {
		defaultCanvasURL = "https://afya.instructure.com"
	}

	now := time.Now().UTC()
	_, err = d.db.Exec(`
		INSERT INTO users (telegram_id, username, first_name, canvas_url, plan_tier, requests_used, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, 0, ?, ?)
	`, telegramID, username, firstName, defaultCanvasURL, planTier, now, now)
	if err != nil {
		return nil, fmt.Errorf("falha ao inserir novo usuário: %w", err)
	}

	return &User{
		TelegramID:   telegramID,
		Username:     username,
		FirstName:    firstName,
		CanvasURL:    defaultCanvasURL,
		PlanTier:     planTier,
		RequestsUsed: 0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// GetUser obtém os dados de um usuário pelo Telegram ID
func (d *Database) GetUser(telegramID int64) (*User, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.getUserLocked(telegramID)
}

func (d *Database) getUserLocked(telegramID int64) (*User, error) {
	query := `
		SELECT telegram_id, username, first_name, canvas_url, canvas_token_encrypted, canvas_token_iv,
		       plan_tier, requests_used, expires_at, created_at, updated_at
		FROM users WHERE telegram_id = ?
	`
	row := d.db.QueryRow(query, telegramID)

	var u User
	var encToken, iv []byte
	var expTime sql.NullTime

	err := row.Scan(
		&u.TelegramID, &u.Username, &u.FirstName, &u.CanvasURL,
		&encToken, &iv, &u.PlanTier, &u.RequestsUsed, &expTime,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if expTime.Valid {
		u.ExpiresAt = &expTime.Time
	}

	// Se o ID estiver na lista de admins do .env, assegura privilégio VIP
	if d.adminIDs[telegramID] {
		u.PlanTier = "vip"
	}

	// Descriptografa o token do Canvas se presente
	if len(encToken) > 0 && len(iv) > 0 {
		plainToken, err := d.decrypt(encToken, iv)
		if err != nil {
			log.Printf("[DATABASE] Erro ao decifrar token Canvas do usuário %d: %v", telegramID, err)
		} else {
			u.CanvasToken = plainToken
		}
	}

	return &u, nil
}

// UpdateCanvasCredentials atualiza o token do Canvas e opcionalmente a URL da instituição
func (d *Database) UpdateCanvasCredentials(telegramID int64, canvasURL, token string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	encToken, iv, err := d.encrypt(token)
	if err != nil {
		return fmt.Errorf("erro ao criptografar token: %w", err)
	}

	if canvasURL == "" {
		canvasURL = "https://afya.instructure.com"
	}

	_, err = d.db.Exec(`
		UPDATE users 
		SET canvas_url = ?, canvas_token_encrypted = ?, canvas_token_iv = ?, updated_at = CURRENT_TIMESTAMP
		WHERE telegram_id = ?
	`, canvasURL, encToken, iv, telegramID)
	return err
}

// UpdateUserName atualiza o nome acadêmico oficial do professor no banco
func (d *Database) UpdateUserName(telegramID int64, realName string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, err := d.db.Exec(`
		UPDATE users 
		SET first_name = ?, updated_at = CURRENT_TIMESTAMP
		WHERE telegram_id = ?
	`, realName, telegramID)
	return err
}

// CheckQuota avalia se o usuário tem permissão para processar requisições
func (d *Database) CheckQuota(telegramID int64) (hasAccess bool, remaining int, plan string, err error) {
	user, err := d.GetUser(telegramID)
	if err != nil {
		return false, 0, "", err
	}

	// Admins do .env ou usuários marcados como VIP têm acesso ilimitado
	if d.IsAdmin(telegramID) || user.PlanTier == "vip" {
		return true, -1, "vip", nil
	}

	// Plano Pro (Assinatura)
	if user.PlanTier == "pro" {
		if user.ExpiresAt == nil || user.ExpiresAt.After(time.Now()) {
			return true, -1, "pro", nil
		}
		// Assinatura expirou
		return false, 0, "expired", nil
	}

	// Plano Trial (Limite de requisições)
	maxReq := d.GetTrialMaxRequests()
	remaining = maxReq - user.RequestsUsed
	if remaining > 0 {
		return true, remaining, "trial", nil
	}

	return false, 0, "trial_ended", nil
}

// IncrementRequests incrementa a contagem de requisições utilizadas
func (d *Database) IncrementRequests(telegramID int64) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, err := d.db.Exec(`
		UPDATE users 
		SET requests_used = requests_used + 1, updated_at = CURRENT_TIMESTAMP
		WHERE telegram_id = ?
	`, telegramID)
	return err
}

// SetUserPlan define o plano de um usuário
func (d *Database) SetUserPlan(telegramID int64, planTier string, durationDays int) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	var expTime any = nil
	if durationDays > 0 {
		t := time.Now().AddDate(0, 0, durationDays)
		expTime = t
	}

	_, err := d.db.Exec(`
		UPDATE users 
		SET plan_tier = ?, expires_at = ?, updated_at = CURRENT_TIMESTAMP
		WHERE telegram_id = ?
	`, planTier, expTime, telegramID)
	return err
}

// -----------------------------------------------------------------------
// CRIPTOGRAFIA AES-256-GCM (Application-Level Envelope Encryption)
// -----------------------------------------------------------------------

func (d *Database) encrypt(plainText string) (cipherBytes, ivBytes []byte, err error) {
	block, err := aes.NewCipher(d.encKey)
	if err != nil {
		return nil, nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	iv := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, nil, err
	}

	ciphertext := gcm.Seal(nil, iv, []byte(plainText), nil)
	return ciphertext, iv, nil
}

func (d *Database) decrypt(cipherBytes, ivBytes []byte) (string, error) {
	block, err := aes.NewCipher(d.encKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	plainBytes, err := gcm.Open(nil, ivBytes, cipherBytes, nil)
	if err != nil {
		return "", fmt.Errorf("falha de autenticação/chave inválida no AES-GCM: %w", err)
	}

	return string(plainBytes), nil
}

// -----------------------------------------------------------------------
// GESTÃO DE HISTÓRICO DE CONVERSAS (MEMÓRIA CONTEXTUAL SLIDING WINDOW)
// -----------------------------------------------------------------------

// GetChatHistoryLimit retorna a quantidade configurada de mensagens para manter no contexto
func (d *Database) GetChatHistoryLimit() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.historyLimit
}

// SaveChatMessage registra uma mensagem no histórico persistente do usuário e
// remove silenciosamente mensagens antigas para manter o banco leve sem intervenção do usuário
func (d *Database) SaveChatMessage(telegramID int64, role, content string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, err := d.db.Exec(`
		INSERT INTO chat_history (telegram_id, role, content, created_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	`, telegramID, role, content)
	if err != nil {
		return err
	}

	// Auto-limpeza silenciosa nos bastidores: descarta mensagens excedentes a 50 turnos por usuário
	_, _ = d.db.Exec(`
		DELETE FROM chat_history 
		WHERE telegram_id = ? AND id NOT IN (
			SELECT id FROM chat_history 
			WHERE telegram_id = ? 
			ORDER BY id DESC 
			LIMIT 50
		)
	`, telegramID, telegramID)

	return nil
}

// GetChatHistory resgata as últimas N mensagens de um usuário em ordem cronológica.
// Se a última interação ocorreu há mais de 4 horas, a sessão expira silenciosamente de forma transparente.
func (d *Database) GetChatHistory(telegramID int64, limit int) ([]ChatMessage, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if limit <= 0 {
		limit = d.historyLimit
	}

	query := `
		SELECT role, content, created_at
		FROM chat_history
		WHERE telegram_id = ?
		ORDER BY id DESC
		LIMIT ?
	`
	rows, err := d.db.Query(query, telegramID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var descMsgs []ChatMessage
	firstRow := true
	for rows.Next() {
		var m ChatMessage
		var createdAt time.Time
		if err := rows.Scan(&m.Role, &m.Content, &createdAt); err != nil {
			return nil, err
		}
		// Se a conversa anterior foi há mais de 4 horas, inicia automaticamente uma sessão fresca
		if firstRow {
			firstRow = false
			if time.Since(createdAt) > 4*time.Hour {
				return nil, nil
			}
		}
		descMsgs = append(descMsgs, m)
	}

	// Inverte a ordem para cronológica (mais antiga -> mais recente)
	var cronoMsgs []ChatMessage
	for i := len(descMsgs) - 1; i >= 0; i-- {
		cronoMsgs = append(cronoMsgs, descMsgs[i])
	}

	return cronoMsgs, nil
}

// ClearChatHistory zera o histórico de conversas do usuário (reset de contexto)
func (d *Database) ClearChatHistory(telegramID int64) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, err := d.db.Exec("DELETE FROM chat_history WHERE telegram_id = ?", telegramID)
	return err
}
