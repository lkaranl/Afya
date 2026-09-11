package main

import (
	"regexp"
	"strings"
)

// Constantes para delimitação semântica de dados não-confiáveis de estudantes
const (
	UntrustedInputTagOpen  = "<untrusted_student_input role=\"data_only\">"
	UntrustedInputTagClose = "</untrusted_student_input>"
)

// Padrões conhecidos de injeção de prompt indireta em códigos e textos
var promptInjectionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\[\s*(instrução|instrucao|instruction|system|sistema)\s*(do\s*sistema|instruction)?\s*\]`),
	regexp.MustCompile(`(?i)(ignore|esqueça|desconsidere)\s+(os\s+critérios|as\s+instruções|todas\s+as\s+regras|previous\s+instructions)`),
	regexp.MustCompile(`(?i)(atribua|dê|dar|give|assign)\s+(nota\s+máxima|nota\s+100|100\/100|full\s+marks|100\s*pontos)`),
	regexp.MustCompile(`(?i)(você\s+agora\s+é|you\s+are\s+now|developer\s+mode|modo\s+desenvolvedor)`),
	regexp.MustCompile(`(?i)(disregard\s+grading|bypass\s+evaluation|overwrite\s+grade)`),
	regexp.MustCompile(`(?i)(SYSTEM\s*:\s*|SYSTEM_PROMPT|AI_INSTRUCTION)`),
}

// WrapUntrustedInput engloba qualquer texto ou código fornecido por estudante
// dentro de tags explícitas de não-confiança para o LLM.
func WrapUntrustedInput(content string) string {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return ""
	}
	if strings.Contains(content, "<untrusted_student_input") {
		return content
	}
	return UntrustedInputTagOpen + "\n" + content + "\n" + UntrustedInputTagClose
}

// DetectPromptInjectionAttempts verifica se há padrões de manipulação de prompt no texto
func DetectPromptInjectionAttempts(content string) []string {
	var matches []string
	for _, pattern := range promptInjectionPatterns {
		if loc := pattern.FindString(content); loc != "" {
			matches = append(matches, strings.TrimSpace(loc))
		}
	}
	return matches
}

// SanitizeToolPayloadForAI aplica isolamento semântico aos dados de estudantes antes de enviá-los ao LLM
func SanitizeToolPayloadForAI(toolName string, payload any) any {
	if payload == nil {
		return nil
	}

	switch toolName {
	case "canvas_get_submissions":
		if subs, ok := payload.([]SubmissionDetail); ok {
			sanitized := make([]SubmissionDetail, len(subs))
			for i, s := range subs {
				sanitized[i] = s
				if s.CleanBody != "" {
					sanitized[i].CleanBody = WrapUntrustedInput(s.CleanBody)
				}
				if s.Body != "" {
					sanitized[i].Body = WrapUntrustedInput(s.Body)
				}
			}
			return sanitized
		}

	case "canvas_prepare_assignment":
		if prep, ok := payload.(*PrepareResult); ok {
			cp := *prep
			sanitizedItems := make([]PreparedItem, len(prep.Items))
			for i, it := range prep.Items {
				sanitizedItems[i] = it
				if it.CodeSnippet != "" {
					sanitizedItems[i].CodeSnippet = WrapUntrustedInput(it.CodeSnippet)
				}
			}
			cp.Items = sanitizedItems
			return &cp
		}

	case "canvas_get_inbox_conversation":
		if thread, ok := payload.(*ConversationDetailResult); ok {
			cp := *thread
			sanitizedMsgs := make([]CanvasInboxMessage, len(thread.Messages))
			for i, m := range thread.Messages {
				sanitizedMsgs[i] = m
				sanitizedMsgs[i].Body = WrapUntrustedInput(m.Body)
			}
			cp.Messages = sanitizedMsgs
			return &cp
		}
	}

	return payload
}
