package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode"

	_ "modernc.org/sqlite"
)

// normalizeSearchText remove acentos e caracteres especiais para busca insensível no SQLite
func normalizeSearchText(text string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(text) {
		switch r {
		case 'á', 'à', 'ã', 'â', 'ä':
			b.WriteRune('a')
		case 'é', 'è', 'ê', 'ë':
			b.WriteRune('e')
		case 'í', 'ì', 'î', 'ï':
			b.WriteRune('i')
		case 'ó', 'ò', 'õ', 'ô', 'ö':
			b.WriteRune('o')
		case 'ú', 'ù', 'û', 'ü':
			b.WriteRune('u')
		case 'ç':
			b.WriteRune('c')
		default:
			if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
				b.WriteRune(r)
			}
		}
	}
	return strings.TrimSpace(b.String())
}

// SQLiteCache implementa cache persistente em disco de alto desempenho com SQLite WAL
type SQLiteCache struct {
	db     *sql.DB
	dbPath string
	mu     sync.RWMutex
}

// NewSQLiteCache inicializa o banco SQLite de cache local
func NewSQLiteCache(dbPath string) (*SQLiteCache, error) {
	if dbPath == "" {
		dbPath = os.Getenv("CANVAS_DB_PATH")
		if dbPath == "" {
			dbPath = "data/afya_cache.db"
		}
	}

	dir := filepath.Dir(dbPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("erro ao criar diretório para sqlite cache '%s': %w", dir, err)
		}
	}

	// Conexão com WAL mode e timeout para máxima concorrência segura
	db, err := sql.Open("sqlite", dbPath+"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)")
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar no SQLite de cache '%s': %w", dbPath, err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	sc := &SQLiteCache{
		db:     db,
		dbPath: dbPath,
	}

	if err := sc.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("erro ao aplicar migrações no SQLite cache: %w", err)
	}

	return sc, nil
}

// migrate cria as tabelas essenciais para o cache L1
func (sc *SQLiteCache) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS cache_key_value (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			expires_at DATETIME,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE INDEX IF NOT EXISTS idx_cache_key_expires ON cache_key_value(expires_at);`,

		`CREATE TABLE IF NOT EXISTS cached_students (
			user_id TEXT NOT NULL,
			course_id TEXT NOT NULL,
			name TEXT NOT NULL,
			search_name TEXT NOT NULL,
			email TEXT,
			login_id TEXT,
			raw_json TEXT NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (user_id, course_id)
		);`,
		`CREATE INDEX IF NOT EXISTS idx_students_course ON cached_students(course_id);`,
		`CREATE INDEX IF NOT EXISTS idx_students_search ON cached_students(course_id, search_name);`,

		`CREATE TABLE IF NOT EXISTS cached_courses (
			course_id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			clean_name TEXT,
			period TEXT,
			term_status TEXT,
			is_current_term INTEGER DEFAULT 0,
			raw_json TEXT NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
	}

	for _, q := range queries {
		if _, err := sc.db.Exec(q); err != nil {
			return fmt.Errorf("erro executando query '%s': %w", q, err)
		}
	}
	return nil
}

// Get obtém um valor em string JSON do cache se não estiver expirado
func (sc *SQLiteCache) Get(key string) (string, bool) {
	if sc == nil || sc.db == nil {
		return "", false
	}
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	var value string
	var expiresAt sql.NullTime

	query := `SELECT value, expires_at FROM cache_key_value WHERE key = ?`
	err := sc.db.QueryRow(query, key).Scan(&value, &expiresAt)
	if err != nil {
		return "", false
	}

	if expiresAt.Valid && time.Now().After(expiresAt.Time) {
		// Expirado
		go sc.Delete(key)
		return "", false
	}

	return value, true
}

// Set salva um valor no cache com TTL
func (sc *SQLiteCache) Set(key string, value string, ttl time.Duration) error {
	if sc == nil || sc.db == nil {
		return nil
	}
	sc.mu.Lock()
	defer sc.mu.Unlock()

	var expiresAt *time.Time
	if ttl > 0 {
		t := time.Now().Add(ttl)
		expiresAt = &t
	}

	query := `INSERT INTO cache_key_value (key, value, expires_at, updated_at) 
	          VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	          ON CONFLICT(key) DO UPDATE SET value=excluded.value, expires_at=excluded.expires_at, updated_at=CURRENT_TIMESTAMP;`
	_, err := sc.db.Exec(query, key, value, expiresAt)
	return err
}

// Delete remove uma chave específica
func (sc *SQLiteCache) Delete(key string) error {
	if sc == nil || sc.db == nil {
		return nil
	}
	sc.mu.Lock()
	defer sc.mu.Unlock()

	_, err := sc.db.Exec(`DELETE FROM cache_key_value WHERE key = ?`, key)
	return err
}

// DeletePrefix remove chaves que iniciam com um determinado prefixo
func (sc *SQLiteCache) DeletePrefix(prefix string) (int, error) {
	if sc == nil || sc.db == nil {
		return 0, nil
	}
	sc.mu.Lock()
	defer sc.mu.Unlock()

	res, err := sc.db.Exec(`DELETE FROM cache_key_value WHERE key LIKE ?`, prefix+"%")
	if err != nil {
		return 0, err
	}
	rows, _ := res.RowsAffected()
	return int(rows), nil
}

// SaveStudents grava a lista de alunos de uma turma de forma indexada
func (sc *SQLiteCache) SaveStudents(courseID string, rawStudents []any) error {
	if sc == nil || sc.db == nil {
		return nil
	}
	sc.mu.Lock()
	defer sc.mu.Unlock()

	tx, err := sc.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT INTO cached_students (user_id, course_id, name, search_name, email, login_id, raw_json, updated_at)
	                         VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	                         ON CONFLICT(user_id, course_id) DO UPDATE SET
	                         name=excluded.name, search_name=excluded.search_name, email=excluded.email,
	                         login_id=excluded.login_id, raw_json=excluded.raw_json, updated_at=CURRENT_TIMESTAMP;`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, item := range rawStudents {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		uid := fmt.Sprintf("%v", m["id"])
		name := fmt.Sprintf("%v", m["name"])
		if uid == "" || name == "" {
			continue
		}
		email := ""
		if em, ok := m["email"].(string); ok {
			email = em
		}
		loginID := ""
		if l, ok := m["login_id"].(string); ok {
			loginID = l
		}

		rawBytes, _ := json.Marshal(m)
		searchName := normalizeSearchText(name)

		if _, err := stmt.Exec(uid, courseID, name, searchName, email, loginID, string(rawBytes)); err != nil {
			log.Printf("[SQLITE] Erro ao gravar aluno %s: %v", name, err)
		}
	}

	return tx.Commit()
}

// GetStudents retorna todos os alunos salvos para uma disciplina
func (sc *SQLiteCache) GetStudents(courseID string) ([]map[string]any, bool) {
	if sc == nil || sc.db == nil {
		return nil, false
	}
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	rows, err := sc.db.Query(`SELECT raw_json FROM cached_students WHERE course_id = ? ORDER BY name ASC`, courseID)
	if err != nil {
		return nil, false
	}
	defer rows.Close()

	var result []map[string]any
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err == nil {
			var m map[string]any
			if err := json.Unmarshal([]byte(raw), &m); err == nil {
				result = append(result, m)
			}
		}
	}

	if len(result) == 0 {
		return nil, false
	}
	return result, true
}

// SearchStudents busca alunos por parte do nome de forma ultra-rápida no SQLite
func (sc *SQLiteCache) SearchStudents(courseID string, query string) ([]map[string]any, error) {
	if sc == nil || sc.db == nil {
		return nil, nil
	}
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	normQuery := "%" + normalizeSearchText(query) + "%"

	sqlQuery := `SELECT raw_json FROM cached_students WHERE search_name LIKE ?`
	args := []any{normQuery}
	if courseID != "" {
		sqlQuery += ` AND course_id = ?`
		args = append(args, courseID)
	}
	sqlQuery += ` ORDER BY name ASC LIMIT 20;`

	rows, err := sc.db.Query(sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]any
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err == nil {
			var m map[string]any
			if err := json.Unmarshal([]byte(raw), &m); err == nil {
				result = append(result, m)
			}
		}
	}
	return result, nil
}

// Clear apaga todos os registros de cache
func (sc *SQLiteCache) Clear() error {
	if sc == nil || sc.db == nil {
		return nil
	}
	sc.mu.Lock()
	defer sc.mu.Unlock()

	_, err1 := sc.db.Exec(`DELETE FROM cache_key_value;`)
	_, err2 := sc.db.Exec(`DELETE FROM cached_students;`)
	_, err3 := sc.db.Exec(`DELETE FROM cached_courses;`)
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return err3
}

// Close fecha o banco de dados
func (sc *SQLiteCache) Close() error {
	if sc == nil || sc.db == nil {
		return nil
	}
	return sc.db.Close()
}
