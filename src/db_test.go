package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDatabaseUsersAndCrypto(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_afya.db")

	os.Setenv("DATABASE_ENCRYPTION_KEY", "chave-super-secreta-para-testes-unitarios-32chars!!")
	os.Setenv("ADMIN_TELEGRAM_IDS", "999888,777666")
	os.Setenv("TRIAL_MAX_REQUESTS", "3")

	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Erro ao inicializar banco de dados de teste: %v", err)
	}
	defer db.Close()

	// 1. Testa criação de usuário comum (Trial)
	userCommon, err := db.GetOrCreateUser(12345, "prof_silva", "Carlos")
	if err != nil {
		t.Fatalf("Erro ao criar usuário comum: %v", err)
	}
	if userCommon.PlanTier != "trial" {
		t.Errorf("Esperava plan_tier='trial', obteve '%s'", userCommon.PlanTier)
	}

	// 2. Testa criação de usuário Admin / VIP
	userVIP, err := db.GetOrCreateUser(999888, "prof_karan", "Karan")
	if err != nil {
		t.Fatalf("Erro ao criar usuário VIP: %v", err)
	}
	if userVIP.PlanTier != "vip" {
		t.Errorf("Esperava plan_tier='vip', obteve '%s'", userVIP.PlanTier)
	}
	if !db.IsAdmin(999888) {
		t.Errorf("db.IsAdmin(999888) deveria ser true")
	}

	// 3. Testa Criptografia e Decriptografia de Token do Canvas
	rawToken := "2094~SecretCanvasTokenABC123XYZ"
	err = db.UpdateCanvasCredentials(12345, "https://afya.instructure.com", rawToken)
	if err != nil {
		t.Fatalf("Erro ao atualizar credenciais do Canvas: %v", err)
	}

	fetchedUser, err := db.GetUser(12345)
	if err != nil {
		t.Fatalf("Erro ao buscar usuário: %v", err)
	}
	if fetchedUser.CanvasToken != rawToken {
		t.Errorf("Esperava token recuperado '%s', obteve '%s'", rawToken, fetchedUser.CanvasToken)
	}

	// 4. Testa Cotas do Trial (limite 3)
	// Requisição 1
	hasAccess, remaining, plan, err := db.CheckQuota(12345)
	if err != nil || !hasAccess || remaining != 3 || plan != "trial" {
		t.Errorf("CheckQuota req 1 falhou: access=%v, remaining=%d, plan=%s", hasAccess, remaining, plan)
	}
	_ = db.IncrementRequests(12345)

	// Requisição 2
	hasAccess, remaining, _, _ = db.CheckQuota(12345)
	if !hasAccess || remaining != 2 {
		t.Errorf("CheckQuota req 2 falhou: access=%v, remaining=%d", hasAccess, remaining)
	}
	_ = db.IncrementRequests(12345)

	// Requisição 3
	hasAccess, remaining, _, _ = db.CheckQuota(12345)
	if !hasAccess || remaining != 1 {
		t.Errorf("CheckQuota req 3 falhou: access=%v, remaining=%d", hasAccess, remaining)
	}
	_ = db.IncrementRequests(12345)

	// Requisição 4 (Esgotado)
	hasAccess, remaining, plan, _ = db.CheckQuota(12345)
	if hasAccess || remaining != 0 || plan != "trial_ended" {
		t.Errorf("CheckQuota após esgotar falhou: access=%v, remaining=%d, plan=%s (esperava false, 0, trial_ended)", hasAccess, remaining, plan)
	}

	// 5. Usuário VIP nunca esgota cota
	vipAccess, vipRemaining, vipPlan, _ := db.CheckQuota(999888)
	if !vipAccess || vipRemaining != -1 || vipPlan != "vip" {
		t.Errorf("VIP CheckQuota falhou: access=%v, remaining=%d, plan=%s", vipAccess, vipRemaining, vipPlan)
	}

	// 6. Promove usuário comum para Pro (30 dias)
	err = db.SetUserPlan(12345, "pro", 30)
	if err != nil {
		t.Fatalf("Erro ao promover para Pro: %v", err)
	}

	proAccess, proRemaining, proPlan, _ := db.CheckQuota(12345)
	if !proAccess || proRemaining != -1 || proPlan != "pro" {
		t.Errorf("Pro CheckQuota falhou: access=%v, remaining=%d, plan=%s", proAccess, proRemaining, proPlan)
	}

	// Testa expiração do Pro
	pastDate := time.Now().AddDate(0, 0, -1)
	_, _ = db.db.Exec("UPDATE users SET expires_at = ? WHERE telegram_id = ?", pastDate, 12345)
	expAccess, _, expPlan, _ := db.CheckQuota(12345)
	if expAccess || expPlan != "expired" {
		t.Errorf("Esperava expiração do Pro: access=%v, plan=%s", expAccess, expPlan)
	}

	// 7. Testa Histórico de Chat Persistente e Memória Contextual
	_ = db.SaveChatMessage(12345, "user", "Pergunta 1")
	_ = db.SaveChatMessage(12345, "assistant", "Resposta 1")
	_ = db.SaveChatMessage(12345, "user", "Pergunta 2")
	_ = db.SaveChatMessage(12345, "assistant", "Resposta 2")

	// Resgata com limite de 2 mensagens mais recentes
	hist, err := db.GetChatHistory(12345, 2)
	if err != nil {
		t.Fatalf("Erro ao resgatar histórico: %v", err)
	}
	if len(hist) != 2 {
		t.Fatalf("Esperava 2 mensagens no histórico, obteve %d", len(hist))
	}
	if hist[0].Content != "Pergunta 2" || hist[1].Content != "Resposta 2" {
		t.Errorf("Histórico em ordem cronológica incorreto: %+v", hist)
	}

	// Testa limpeza de memória /reset
	err = db.ClearChatHistory(12345)
	if err != nil {
		t.Fatalf("Erro ao limpar histórico: %v", err)
	}
	clearedHist, _ := db.GetChatHistory(12345, 10)
	if len(clearedHist) != 0 {
		t.Errorf("Esperava histórico vazio após ClearChatHistory, obteve %d", len(clearedHist))
	}
}
