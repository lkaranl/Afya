package main

import (
	"strings"
	"testing"
)

func TestFormatMarkdownForTelegram(t *testing.T) {
	inputMarkdown := `# Relatório de Correção

**Atenção:** Abaixo estão as notas dos alunos avaliados:

> Devolutiva obrigatória em até 10 dias conforme Resolução CONSEPE.

* Item 1: Critério de ponteiros
- Item 2: Liberação de memória

| Aluno | Nota | Feedback |
| :--- | :--- | :--- |
| Lucas Gabriel | 95 / 100 | Excelente lógica de ponteiros. |
| Beatriz Lima | 80 / 100 | Faltou liberar memória com free(). |

## Código Exemplo
` + "```c\nint *p = malloc(sizeof(int));\nfree(p);\n```" + `

### Próximos Passos
Por favor, confirme o lançamento das notas.`

	formatted := FormatMarkdownForTelegram(inputMarkdown)

	// Valida cabeçalho estilizado
	if !strings.Contains(formatted, "📌 *RELATÓRIO DE CORREÇÃO*") {
		t.Errorf("Esperava cabeçalho principal formatado, obteve:\n%s", formatted)
	}

	// Valida conversão de **Atenção:** para *Atenção:*
	if !strings.Contains(formatted, "*Atenção:*") {
		t.Errorf("Esperava conversão de ** para *, obteve:\n%s", formatted)
	}

	// Valida citação estilizada com barra lateral
	if !strings.Contains(formatted, "┃ _Devolutiva obrigatória em até 10 dias conforme Resolução CONSEPE._") {
		t.Errorf("Esperava citação formatada com barra lateral, obteve:\n%s", formatted)
	}

	// Valida marcadores convertidos para bullet
	if !strings.Contains(formatted, "• Item 1: Critério de ponteiros") || !strings.Contains(formatted, "• Item 2: Liberação de memória") {
		t.Errorf("Esperava listas convertidas para bullets, obteve:\n%s", formatted)
	}

	// Valida que a tabela foi convertida em cards
	if strings.Contains(formatted, "| Lucas Gabriel |") {
		t.Errorf("A tabela com pipes não deveria mais existir no resultado!")
	}

	if !strings.Contains(formatted, "👤 *Lucas Gabriel*") {
		t.Errorf("Esperava card de Lucas Gabriel com ícone 👤")
	}

	if !strings.Contains(formatted, "📊 *Nota:* 95 / 100") {
		t.Errorf("Esperava linha de nota formatada")
	}

	// Valida divisor entre cards
	if !strings.Contains(formatted, "──────────────────────") {
		t.Errorf("Esperava divisor entre os cartões dos alunos")
	}

	// Valida bloco de código preservado
	if !strings.Contains(formatted, "```c\nint *p = malloc(sizeof(int));\nfree(p);\n```") {
		t.Errorf("Esperava bloco de código C preservado intacto")
	}

	// Valida subtítulo nível 2 e 3
	if !strings.Contains(formatted, "🔹 *Código Exemplo*") {
		t.Errorf("Esperava subtítulo h2 formatado com 🔹")
	}
	if !strings.Contains(formatted, "🔸 *Próximos Passos*") {
		t.Errorf("Esperava subtítulo h3 formatado com 🔸")
	}
}

func TestTelegramFormattingHelpers(t *testing.T) {
	// Teste de Negrito
	if res := TgBold("Karan"); res != "*Karan*" {
		t.Errorf("TgBold falhou: %s", res)
	}

	// Teste de Itálico
	if res := TgItalic("Didática"); res != "_Didática_" {
		t.Errorf("TgItalic falhou: %s", res)
	}

	// Teste de Negrito Itálico
	if res := TgBoldItalic("Importante"); res != "*_Importante_*" {
		t.Errorf("TgBoldItalic falhou: %s", res)
	}

	// Teste de Código Inline
	if res := TgCode("int x = 10;"); res != "`int x = 10;`" {
		t.Errorf("TgCode falhou: %s", res)
	}

	// Teste de Bloco de Código
	if res := TgPre("return 0;", "c"); res != "```c\nreturn 0;\n```" {
		t.Errorf("TgPre falhou:\n%s", res)
	}

	// Teste de Link
	if res := TgLink("Canvas LMS", "https://afya.instructure.com"); res != "[Canvas LMS](https://afya.instructure.com)" {
		t.Errorf("TgLink falhou: %s", res)
	}

	// Teste de Spoiler
	if res := TgSpoiler("Resposta Secreta"); res != "||Resposta Secreta||" {
		t.Errorf("TgSpoiler falhou: %s", res)
	}

	// Teste de Quote
	if res := TgQuote("Linha 1\nLinha 2"); !strings.Contains(res, "┃ _Linha 1_") || !strings.Contains(res, "┃ _Linha 2_") {
		t.Errorf("TgQuote falhou: %s", res)
	}

	// Teste de Card Completo
	fields := []TgCardField{
		{Label: "Nota", Value: "90", Icon: "📊"},
		{Label: "Situação", Value: "Aprovado", Icon: "✅"},
	}
	card := TgCard("João Silva", "👤", fields)
	if !strings.Contains(card, "👤 *João Silva*") || !strings.Contains(card, "📊 *Nota:* 90") {
		t.Errorf("TgCard falhou:\n%s", card)
	}
}
