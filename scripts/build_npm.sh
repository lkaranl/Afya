#!/bin/bash
set -e

echo "🚀 Compilando binários estritos e enxutos do MCP para distribuição npm..."
mkdir -p bin

LDFLAGS="-s -w"

echo "  📦 [1/5] macOS arm64 (Apple Silicon)..."
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$LDFLAGS" -o bin/afya-canvas-darwin-arm64 ./src

echo "  📦 [2/5] macOS x64 (Intel)..."
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$LDFLAGS" -o bin/afya-canvas-darwin-x64 ./src

echo "  📦 [3/5] Linux arm64..."
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$LDFLAGS" -o bin/afya-canvas-linux-arm64 ./src

echo "  📦 [4/5] Linux x64..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$LDFLAGS" -o bin/afya-canvas-linux-x64 ./src

echo "  📦 [5/5] Windows x64..."
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$LDFLAGS" -o bin/afya-canvas-win32-x64.exe ./src

echo "  🖥️  Compilando binário web local com painel visual e chat..."
go build -tags web -ldflags="$LDFLAGS" -o bin/afya-canvas-web ./src

chmod +x bin/afya-canvas-* bin/cli.js 2>/dev/null || true

echo ""
echo "✅ Todos os binários foram compilados com sucesso!"
ls -lh bin/
