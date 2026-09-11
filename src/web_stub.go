//go:build !web

package main

import (
	"fmt"
	"os"
)

// runWebServer neste modo stub informa que o binário é exclusivamente o servidor MCP
func runWebServer(client *CanvasClient, port string) {
	fmt.Fprintln(os.Stderr, "================================================================================")
	fmt.Fprintln(os.Stderr, "  [Afya Canvas MCP] Binário Exclusivo de Servidor MCP (Model Context Protocol)")
	fmt.Fprintln(os.Stderr, "================================================================================")
	fmt.Fprintln(os.Stderr, "Este pacote foi compilado exclusivamente para integração MCP (stdio) com Agentes de IA")
	fmt.Fprintln(os.Stderr, "(Claude Desktop, Cursor, Antigravity, etc.).")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Para utilizar o painel visual web, execute a versão completa compilada com suporte web.")
	os.Exit(1)
}
