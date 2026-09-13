# 📱 Guia Definitivo de Formatação no Telegram Bot API

Este documento serve como referência rápida e completa para estilização de mensagens no Telegram Bot da Afya, detalhando tanto as regras oficiais da Telegram Bot API quanto os **Helpers Prontos em Go** já integrados ao projeto em [`src/telegram_formatter.go`](file:///Users/karan/Github/Afya/src/telegram_formatter.go).

---

## 🌟 1. Modos de Formatação Suportados pelo Telegram

O Telegram Bot API aceita 3 opções de `parse_mode`:

| Modo | Sintaxe de Exemplo | Vantagens | Desvantagens / Cuidados |
| :--- | :--- | :--- | :--- |
| **Markdown (Legado v1)** | `*negrito*`, `_itálico_` | Simples, poucos caracteres especiais | Não suporta sublinhado, tachado, spoilers ou citações nativas; underscores soltos (`user_id`) podem quebrar. |
| **MarkdownV2** | `*negrito*`, `__sublinhado__`, `~tachado~`, `\|\|spoiler\|\|` | Suporta todos os estilos modernos e blockquotes | **Crítico:** Exige escape de 18 caracteres (`_ * [ ] ( ) ~ ` > # + - = \| { } . ! \`). Qualquer ponto ou hífen esquecido resulta em Erro 400! |
| **HTML** | `<b>negrito</b>`, `<i>itálico</i>`, `<u>sublinhado</u>`, `<tg-spoiler>` | Extremamente estável; caracteres comuns (`.`, `-`, `!`, `(`, `)`) não precisam de escape. | Exige escape apenas de `<`, `>`, `&`. |

---

## 🛠️ 2. A Solução do Assistente: Formatação Automática e Mobile-First

Quando os modelos de IA (como Gemini ou Claude) geram respostas, eles normalmente usam **Markdown tradicional do GitHub**:
- Usam `**negrito**` (dois asteriscos) em vez de `*negrito*`.
- Criam tabelas com pipes `| col1 | col2 |` que quebram a visualização em smartphones.
- Usam marcadores de lista `* item` ou `- item`, que no Telegram podem ser interpretados como asteriscos de formatação abertos sem par.
- Usam citações `> texto`.

A função [`FormatMarkdownForTelegram(text)`](file:///Users/karan/Github/Afya/src/telegram_formatter.go) resolve tudo isso automaticamente nos bastidores antes de enviar a mensagem:
1. **Converte tabelas rígidas em Cards/Fichas visuais** com emojis temáticos e divisores finos (`──────────────────────`).
2. **Normaliza `**negrito**` para `*negrito*`**, garantindo renderização correta sem asteriscos soltos.
3. **Converte `* item` e `- item` para `• item`**, evitando erros de parsing no Telegram.
4. **Transforma `> citação` em citações elegantes com barra lateral (`┃ _texto_`)**.
5. **Estiliza títulos `#`, `##`, `###` com badges visuais** (`📌 *TÍTULO*`, `🔹 *Título*`, `🔸 *Título*`).
6. **Preserva blocos de código intactos** (como ```` ```c ... ``` ````).
7. **Mecanismo de Fallback Seguro:** Se mesmo assim o Telegram recusar qualquer entidade de formatação, a função de envio reenviar automaticamente como texto limpo, garantindo que o professor nunca fique sem resposta.

---

## 🎨 3. Tabela Comparativa de Sintaxes e Efeitos Visuais

| Efeito Visual | Sintaxe Telegram Markdown | Sintaxe Telegram HTML | Helper em Go Pronto |
| :--- | :--- | :--- | :--- |
| **Negrito** | `*texto*` | `<b>texto</b>` | `TgBold("texto")` |
| *Itálico* | `_texto_` | `<i>texto</i>` | `TgItalic("texto")` |
| ***Negrito + Itálico*** | `*_texto_*` | `<b><i>texto</i></b>` | `TgBoldItalic("texto")` |
| `Código Inline` | `` `código` `` | `<code>código</code>` | `TgCode("código")` |
| **Bloco de Código** | ```` ```c\nint x = 0;\n``` ```` | `<pre><code class="language-c">...</code></pre>` | `TgPre("int x = 0;", "c")` |
| [Link Clicável](https://canvas.afya.com.br) | `[Título](https://url)` | `<a href="https://url">Título</a>` | `TgLink("Título", "https://url")` |
| Spoiler (Oculto) | `\|\|texto oculto\|\|` | `<tg-spoiler>texto oculto</tg-spoiler>` | `TgSpoiler("texto oculto")` |
| Citação / Destaque | `┃ _texto da citação_` | `<blockquote>texto</blockquote>` | `TgQuote("texto")` |
| Divisor Fino | `──────────────────────` | `──────────────────────` | `TgDivider()` |
| Item de Lista | `• Elemento da lista` | `• Elemento da lista` | `TgBullet("Elemento")` |
| Cabeçalho Nível 1 | `📌 *TÍTULO PRINCIPAL*` | `📌 <b>TÍTULO PRINCIPAL</b>` | `TgHeader1("Título")` |
| Cabeçalho Nível 2 | `🔹 *Subtítulo de Seção*` | `🔹 <b>Subtítulo</b>` | `TgHeader2("Subtítulo")` |
| Cabeçalho Nível 3 | `🔸 *Tópico Específico*` | `🔸 <b>Tópico</b>` | `TgHeader3("Tópico")` |

---

## 💻 4. Como Usar os Helpers no Código Go

Todos os helpers estão disponíveis globalmente no pacote `main`:

```go
package main

import (
    "fmt"
)

func ExemploMontagemMensagem() string {
    // 1. Título principal
    msg := TgHeader1("Relatório de Avaliação") + "\n\n"

    // 2. Citação / Aviso pedagógico
    msg += TgQuote("Lembrete: Prazo de devolutiva institucional de 10 dias úteis.") + "\n\n"

    // 3. Montagem de Card Visual de Aluno
    campos := []TgCardField{
        {Label: "Nota", Value: "95 / 100", Icon: "📊"},
        {Label: "Feedback", Value: "Excelente uso de ponteiros e liberação com free().", Icon: "💬"},
        {Label: "Situação", Value: "Aprovado em N1", Icon: "✅"},
    }
    msg += TgCard("Lucas Gabriel", "👤", campos) + "\n"

    // 4. Divisor de celular
    msg += TgDivider() + "\n\n"

    // 5. Bloco de código de referência
    codigo := "int *ptr = malloc(sizeof(int));\n*ptr = 42;\nfree(ptr);"
    msg += TgHeader2("Gabarito de Referência") + "\n"
    msg += TgPre(codigo, "c") + "\n\n"

    // 6. Link de acesso
    msg += "Acesse a atividade completa no " + TgLink("Canvas LMS da Afya", "https://afya.instructure.com")

    return msg
}
```

---

## 📱 5. Como Fica a Renderização no Celular

Ao enviar uma tabela convencional Markdown, o Telegram converte para:

```text
📌 RELATÓRIO DE CORREÇÃO

👤 Lucas Gabriel
📊 Nota: 95 / 100
💬 Feedback: Excelente lógica de ponteiros.
──────────────────────
👤 Beatriz Lima
📊 Nota: 80 / 100
💬 Feedback: Faltou liberar memória com free().
──────────────────────
🔹 Código Exemplo
```c
int *p = malloc(sizeof(int));
free(p);
```
```

---

## 🔒 6. Boas Práticas e Segurança

1. **Evite caracteres soltos:** Nunca envie caracteres de controle não emparelhados (como um `*` solto no meio de uma frase).
2. **Use sempre os helpers para dados dinâmicos:** Funções como `TgCode()` ou `TgBold()` já tratam e higienizam o conteúdo para evitar erros de renderização.
3. **Limite de 4096 caracteres:** Lembre-se de que cada mensagem no Telegram tem teto de 4096 caracteres. Para respostas longas, use a função nativa `b.sendLongTextMessage(chatID, texto)`, que divide os blocos de forma inteligente sem cortar palavras ou códigos ao meio.
