package main

import (
	"fmt"
	"regexp"
	"strings"
)

// -----------------------------------------------------------------------
// HELPERS DE FORMATAÇÃO PRONTOS PARA USAR NO TELEGRAM (MARKDOWN / HTML)
// -----------------------------------------------------------------------

// TgBold formata o texto em negrito
func TgBold(text string) string {
	clean := strings.TrimSpace(text)
	if clean == "" {
		return ""
	}
	return "*" + clean + "*"
}

// TgItalic formata o texto em itálico
func TgItalic(text string) string {
	clean := strings.TrimSpace(text)
	if clean == "" {
		return ""
	}
	return "_" + clean + "_"
}

// TgBoldItalic formata em negrito e itálico combinados
func TgBoldItalic(text string) string {
	clean := strings.TrimSpace(text)
	if clean == "" {
		return ""
	}
	return "*_" + clean + "_*"
}

// TgCode formata um trecho como código inline de largura fixa
func TgCode(code string) string {
	clean := strings.TrimSpace(code)
	if clean == "" {
		return ""
	}
	// Se já contiver backtick, evita quebrar a formatação
	clean = strings.ReplaceAll(clean, "`", "'")
	return "`" + clean + "`"
}

// TgPre formata um bloco de código de múltiplas linhas com indicação de linguagem
func TgPre(code, lang string) string {
	clean := strings.Trim(code, "\n")
	if lang != "" {
		return "```" + lang + "\n" + clean + "\n```"
	}
	return "```\n" + clean + "\n```"
}

// TgLink cria um hiperlink clicável com texto âncora
func TgLink(text, url string) string {
	return fmt.Sprintf("[%s](%s)", strings.TrimSpace(text), strings.TrimSpace(url))
}

// TgSpoiler formata como spoiler (oculto até o usuário tocar/clicar)
func TgSpoiler(text string) string {
	clean := strings.TrimSpace(text)
	if clean == "" {
		return ""
	}
	return "||" + clean + "||"
}

// TgQuote formata uma citação com barra lateral estilizada e itálico
func TgQuote(text string) string {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	var quoted []string
	for _, l := range lines {
		quoted = append(quoted, "┃ _"+strings.TrimSpace(l)+"_")
	}
	return strings.Join(quoted, "\n")
}

// TgBullet formata um item de lista com ponto elegante
func TgBullet(text string) string {
	return "• " + strings.TrimSpace(text)
}

// TgDivider retorna a linha divisora fina padronizada para celular
func TgDivider() string {
	return "──────────────────────"
}

// TgHeader1 cria um cabeçalho principal de primeiro nível com ícone e destaque
func TgHeader1(title string) string {
	return "📌 *" + strings.ToUpper(strings.TrimSpace(title)) + "*"
}

// TgHeader2 cria um cabeçalho de seção de segundo nível
func TgHeader2(title string) string {
	return "🔹 *" + strings.TrimSpace(title) + "*"
}

// TgHeader3 cria um cabeçalho de subseção de terceiro nível
func TgHeader3(title string) string {
	return "🔸 *" + strings.TrimSpace(title) + "*"
}

// TgCardField representa um par de rótulo e valor dentro de um card visual
type TgCardField struct {
	Label string
	Value string
	Icon  string
}

// TgCard monta um cartão visual elegante pronto com título, ícone e campos alinhados
func TgCard(title, icon string, fields []TgCardField) string {
	var b strings.Builder
	if icon == "" {
		icon = "📋"
	}
	if title != "" {
		b.WriteString(icon + " *" + strings.TrimSpace(title) + "*\n")
	}
	for _, f := range fields {
		fIcon := f.Icon
		if fIcon == "" {
			fIcon = getFieldIcon(f.Label)
		}
		val := strings.TrimSpace(f.Value)
		if val == "" {
			val = "-"
		}
		b.WriteString(fmt.Sprintf("%s *%s:* %s\n", fIcon, strings.TrimSpace(f.Label), val))
	}
	return strings.TrimRight(b.String(), "\n")
}

// -----------------------------------------------------------------------
// CONVERSOR UNIVERSAL DE MARKDOWN PARA TELEGRAM MOBILE-FIRST
// -----------------------------------------------------------------------

// FormatMarkdownForTelegram adapta o Markdown convencional para um formato mobile-first moderno no Telegram,
// convertendo tabelas rígidas em cards visuais com divisores, corrigindo negritos duplos (** -> *)
// e normalizando listas para evitar erros de sintaxe na API do Telegram.
func FormatMarkdownForTelegram(text string) string {
	if strings.TrimSpace(text) == "" {
		return text
	}

	// 1. Converte tabelas Markdown em Cards Elegantes
	text = convertMarkdownTablesToCards(text)

	// 2. Protege blocos de código para não alterar sintaxe interna
	codeBlockRegex := regexp.MustCompile("(?s)```.*?```")
	var codeBlocks []string
	text = codeBlockRegex.ReplaceAllStringFunc(text, func(match string) string {
		idx := len(codeBlocks)
		codeBlocks = append(codeBlocks, match)
		return fmt.Sprintf("@@TG_CODE_BLOCK_%d@@", idx)
	})

	// 3. Normaliza negrito de Markdown comum (**texto**) para o Telegram (*texto*)
	// Evita que '**' soltos confundam o parser do Telegram
	boldRegex := regexp.MustCompile(`\*\*([^\*]+?)\*\*`)
	text = boldRegex.ReplaceAllString(text, "*$1*")

	// 4. Converte tachado ~~texto~~ para formato limpo
	strikethroughRegex := regexp.MustCompile(`~~([^~]+?)~~`)
	text = strikethroughRegex.ReplaceAllString(text, "~$1~")

	// 5. Linha a linha: cabeçalhos (#), listas (* e -) e citações (>)
	lines := strings.Split(text, "\n")
	var resultLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Cabeçalhos
		if strings.HasPrefix(trimmed, "### ") {
			content := strings.TrimPrefix(trimmed, "### ")
			resultLines = append(resultLines, TgHeader3(content))
		} else if strings.HasPrefix(trimmed, "## ") {
			content := strings.TrimPrefix(trimmed, "## ")
			resultLines = append(resultLines, TgHeader2(content))
		} else if strings.HasPrefix(trimmed, "# ") {
			content := strings.TrimPrefix(trimmed, "# ")
			resultLines = append(resultLines, TgHeader1(content))
		} else if strings.HasPrefix(trimmed, "> ") {
			// Citações de bloco (blockquote estilizado)
			quoteContent := strings.TrimPrefix(trimmed, "> ")
			resultLines = append(resultLines, "┃ _"+quoteContent+"_")
		} else if strings.HasPrefix(trimmed, "* ") || strings.HasPrefix(trimmed, "- ") {
			// Converte marcadores '* ' e '- ' para '• ' para evitar que o '*' seja interpretado como início de negrito
			bulletContent := strings.TrimPrefix(trimmed, "* ")
			bulletContent = strings.TrimPrefix(bulletContent, "- ")
			resultLines = append(resultLines, "• "+bulletContent)
		} else {
			resultLines = append(resultLines, line)
		}
	}

	res := strings.Join(resultLines, "\n")

	// 6. Restaura blocos de código intactos
	for i, cb := range codeBlocks {
		placeholder := fmt.Sprintf("@@TG_CODE_BLOCK_%d@@", i)
		res = strings.ReplaceAll(res, placeholder, cb)
	}

	// 7. Remove quebras de linha excessivas
	res = regexp.MustCompile(`\n{3,}`).ReplaceAllString(res, "\n\n")
	return strings.TrimSpace(res)
}

// convertMarkdownTablesToCards identifica blocos de tabelas Markdown (| col1 | col2 |)
// e os converte em blocos de cartões/fichas com separadores finos.
func convertMarkdownTablesToCards(text string) string {
	lines := strings.Split(text, "\n")
	var out []string

	inTable := false
	var tableHeader []string
	var tableRows [][]string

	isTableSep := func(line string) bool {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "|") || !strings.HasSuffix(trimmed, "|") {
			return false
		}
		matched, _ := regexp.MatchString(`^\|(\s*:?-+:?\s*\|)+$`, trimmed)
		return matched
	}

	isTableRow := func(line string) bool {
		trimmed := strings.TrimSpace(line)
		return strings.HasPrefix(trimmed, "|") && strings.HasSuffix(trimmed, "|") && len(strings.Split(trimmed, "|")) >= 3
	}

	splitRow := func(line string) []string {
		trimmed := strings.TrimSpace(line)
		trimmed = strings.TrimPrefix(trimmed, "|")
		trimmed = strings.TrimSuffix(trimmed, "|")
		parts := strings.Split(trimmed, "|")
		var clean []string
		for _, p := range parts {
			clean = append(clean, strings.TrimSpace(p))
		}
		return clean
	}

	flushTable := func() {
		if len(tableHeader) == 0 || len(tableRows) == 0 {
			inTable = false
			tableHeader = nil
			tableRows = nil
			return
		}

		var cards []string
		for i, row := range tableRows {
			var cardLines []string
			mainTitle := ""
			mainIcon := "📋"

			for colIdx, colVal := range row {
				if colIdx >= len(tableHeader) {
					continue
				}
				hName := strings.ToLower(tableHeader[colIdx])
				val := strings.TrimSpace(colVal)
				if val == "" {
					val = "-"
				}

				icon := getFieldIcon(hName)

				if mainTitle == "" && (strings.Contains(hName, "aluno") || strings.Contains(hName, "estudante") || strings.Contains(hName, "disciplina") || strings.Contains(hName, "tarefa") || strings.Contains(hName, "título") || strings.Contains(hName, "remetente")) {
					mainTitle = val
					mainIcon = icon
				} else {
					cardLines = append(cardLines, icon+" *"+tableHeader[colIdx]+":* "+val)
				}
			}

			var cardBuilder strings.Builder
			if mainTitle != "" {
				cardBuilder.WriteString(mainIcon + " *" + mainTitle + "*\n")
			}
			for _, cl := range cardLines {
				cardBuilder.WriteString(cl + "\n")
			}

			// Divisor fino entre cartões
			if i < len(tableRows)-1 {
				cardBuilder.WriteString(TgDivider())
			}
			cards = append(cards, strings.TrimRight(cardBuilder.String(), "\n"))
		}

		out = append(out, strings.Join(cards, "\n"))
		inTable = false
		tableHeader = nil
		tableRows = nil
	}

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		if isTableRow(line) {
			if !inTable {
				if i+1 < len(lines) && isTableSep(lines[i+1]) {
					inTable = true
					tableHeader = splitRow(line)
					i++ // Pula o separador
					continue
				} else {
					out = append(out, line)
					continue
				}
			} else {
				tableRows = append(tableRows, splitRow(line))
				continue
			}
		} else {
			if inTable {
				flushTable()
			}
			out = append(out, line)
		}
	}

	if inTable {
		flushTable()
	}

	return strings.Join(out, "\n")
}

// getFieldIcon seleciona um emoji amigável e moderno para a coluna
func getFieldIcon(headerName string) string {
	h := strings.ToLower(headerName)
	switch {
	case strings.Contains(h, "aluno") || strings.Contains(h, "estudante") || strings.Contains(h, "remetente") || strings.Contains(h, "nome"):
		return "👤"
	case strings.Contains(h, "nota") || strings.Contains(h, "pontuação") || strings.Contains(h, "pontos") || strings.Contains(h, "score"):
		return "📊"
	case strings.Contains(h, "feedback") || strings.Contains(h, "comentário") || strings.Contains(h, "assunto") || strings.Contains(h, "mensagem"):
		return "💬"
	case strings.Contains(h, "prazo") || strings.Contains(h, "data") || strings.Contains(h, "vencimento") || strings.Contains(h, "entrega"):
		return "📅"
	case strings.Contains(h, "disciplina") || strings.Contains(h, "turma") || strings.Contains(h, "curso"):
		return "🏛"
	case strings.Contains(h, "status") || strings.Contains(h, "situação"):
		return "⚡️"
	case strings.Contains(h, "pendente") || strings.Contains(h, "aguardando"):
		return "⏳"
	case strings.Contains(h, "risco") || strings.Contains(h, "alerta"):
		return "⚠️"
	case strings.Contains(h, "id"):
		return "🆔"
	default:
		return "▫️"
	}
}
