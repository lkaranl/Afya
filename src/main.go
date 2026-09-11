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

	if *webMode {
		runWebServer(client, port)
	} else {
		// Modo padrão: Servidor MCP via stdio para Agentes de IA
		log.SetOutput(os.Stderr)
		runMCPServer(client)
	}
}
