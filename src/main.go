package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"strings"
)

func loadEnvFile() {
	file, err := os.Open(".env")
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			if os.Getenv(key) == "" {
				os.Setenv(key, val)
			}
		}
	}
}

func main() {
	loadEnvFile()

	webMode := flag.Bool("web", false, "Executa o painel visual web no navegador em vez do servidor MCP")
	telegramMode := flag.Bool("telegram", false, "Executa o bot comercial do Telegram via Long Polling")
	portFlag := flag.String("port", "", "Porta HTTP para o painel web (sobrescreve variável PORT do .env)")
	flag.Parse()

	token := os.Getenv("TOKEN")
	baseURL := os.Getenv("CANVAS_BASE_URL")
	if baseURL == "" {
		baseURL = "https://afya.instructure.com"
	}

	port := *portFlag
	if port == "" {
		port = os.Getenv("PORT")
		if port == "" {
			port = "3000"
		}
	}

	client := NewCanvasClient(baseURL, token)

	// Se o modo Telegram foi acionado
	if *telegramMode {
		tgToken := os.Getenv("TELEGRAM_BOT_TOKEN")
		if tgToken == "" {
			tgToken = os.Getenv("TELEGRAM_NOTIFIER_BOT_TOKEN")
		}
		if tgToken == "" {
			log.Fatalf("❌ [TELEGRAM] 'TELEGRAM_BOT_TOKEN' ou 'TELEGRAM_NOTIFIER_BOT_TOKEN' não configurado no arquivo .env.")
		}

		db, err := NewDatabase("")
		if err != nil {
			log.Fatalf("❌ [DATABASE] Falha ao inicializar banco de dados SQLite: %v", err)
		}
		defer db.Close()

		engine := NewAgentEngine(client)
		bot := NewTelegramBot(tgToken, db, engine)

		if err := bot.Start(); err != nil {
			log.Fatalf("❌ [TELEGRAM] Erro ao iniciar bot do Telegram: %v", err)
		}
		defer bot.Stop()

		// Se também solicitou a web (--telegram --web), roda a web concorrentemente
		if *webMode {
			log.Printf("🚀 [SISTEMA] Executando Web Dashboard (porta %s) e Bot do Telegram simultaneamente.", port)
			runWebServer(client, port)
			return
		}

		// Caso contrário, aguarda sinal de encerramento do sistema operacional
		log.Println("🤖 [SISTEMA] Afya Canvas Assistant rodando exclusivamente via Telegram Bot. Pressione Ctrl+C para encerrar.")
		select {}
	}

	if *webMode {
		runWebServer(client, port)
	} else {
		// Modo padrão: Servidor MCP via stdio para Agentes de IA
		log.SetOutput(os.Stderr)
		runMCPServer(client)
	}
}
