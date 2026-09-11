package main

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// PlagiarismPair representa um par de alunos com suspeita de similaridade/plágio
type PlagiarismPair struct {
	StudentA             string   `json:"student_a"`
	StudentAID           string   `json:"student_a_id"`
	StudentB             string   `json:"student_b"`
	StudentBID           string   `json:"student_b_id"`
	SimilarityPercentage float64  `json:"similarity_percentage"`
	PlagiarismRisk       string   `json:"plagiarism_risk"` // "ALTO", "MÉDIO", "BAIXO"
	TechniquesDetected   []string `json:"techniques_detected"`
	SharedTokenCount     int      `json:"shared_token_count"`
	TotalTokensA         int      `json:"total_tokens_a"`
	TotalTokensB         int      `json:"total_tokens_b"`
}

// PlagiarismCheckResult consolida a auditoria de plágio de uma atividade inteira
type PlagiarismCheckResult struct {
	CourseID            string           `json:"course_id"`
	AssignmentID        string           `json:"assignment_id"`
	AssignmentName      string           `json:"assignment_name,omitempty"`
	TotalSubmissions    int              `json:"total_submissions"`
	TotalEvaluatedCodes int              `json:"total_evaluated_codes"`
	Threshold           float64          `json:"threshold"`
	SuspectPairsCount   int              `json:"suspect_pairs_count"`
	SuspectPairs        []PlagiarismPair `json:"suspect_pairs"`
	MarkdownReport      string           `json:"markdown_report"`
}

// cKeywords conjunto de palavras-chave protegidas em C/C++ que definem estrutura
var cKeywords = map[string]bool{
	"if": true, "else": true, "for": true, "while": true, "do": true,
	"switch": true, "case": true, "default": true, "break": true, "continue": true,
	"return": true, "goto": true, "struct": true, "typedef": true, "union": true,
	"enum": true, "sizeof": true, "const": true, "static": true, "extern": true,
	"inline": true, "volatile": true, "int": true, "char": true, "float": true,
	"double": true, "void": true, "bool": true, "size_t": true, "long": true,
	"short": true, "unsigned": true, "signed": true, "NULL": true,
}

// removeComments remove comentários de linha única (//) e em bloco (/* */)
func removeComments(code string) string {
	// Remove bloco /* ... */
	reBlock := regexp.MustCompile(`/\*[\s\S]*?\*/`)
	code = reBlock.ReplaceAllString(code, " ")

	// Remove linha // ...
	reLine := regexp.MustCompile(`//[^\n]*`)
	code = reLine.ReplaceAllString(code, " ")

	return code
}

// tokenizeAndCanonicalize processa o código fonte gerando uma sequência de tokens normalizados
func tokenizeAndCanonicalize(rawCode string, normalizeIdentifiers bool) []string {
	clean := removeComments(rawCode)

	// Remove diretivas de pré-processador
	reDirectives := regexp.MustCompile(`(?m)^\s*#(?:include|define|ifndef|ifdef|endif|pragma)[^\n]*`)
	clean = reDirectives.ReplaceAllString(clean, " ")

	// Normaliza literais de string e char para token genérico
	reStrings := regexp.MustCompile(`"(\\.|[^"\\])*"|'(\\.|[^'\\])*'`)
	clean = reStrings.ReplaceAllString(clean, " STR_LIT ")

	// Normaliza números literais
	reNumbers := regexp.MustCompile(`\b\d+(\.\d+)?\b|\b0x[0-9a-fA-F]+\b`)
	clean = reNumbers.ReplaceAllString(clean, " NUM_LIT ")

	// Divide em tokens alfanuméricos e símbolos de controle
	var rawTokens []string
	var cur strings.Builder

	for _, r := range clean {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			cur.WriteRune(r)
		} else {
			if cur.Len() > 0 {
				rawTokens = append(rawTokens, cur.String())
				cur.Reset()
			}
			if strings.ContainsRune("{}()[];,+-*/%=><!&|^~?:.", r) {
				rawTokens = append(rawTokens, string(r))
			}
		}
	}
	if cur.Len() > 0 {
		rawTokens = append(rawTokens, cur.String())
	}

	if !normalizeIdentifiers {
		return rawTokens
	}

	// Normalização Canônica: mapeia identificadores não-reservados para V1, V2, V3...
	identMap := make(map[string]string)
	counter := 1
	canonicalTokens := make([]string, len(rawTokens))

	for i, t := range rawTokens {
		if cKeywords[t] || strings.ContainsAny(t, "{}()[];,+-*/%=><!&|^~?:.") || t == "STR_LIT" || t == "NUM_LIT" {
			canonicalTokens[i] = t
		} else {
			// É identificador local (variável ou função do aluno)
			v, exists := identMap[t]
			if !exists {
				v = fmt.Sprintf("V%d", counter)
				counter++
				identMap[t] = v
			}
			canonicalTokens[i] = v
		}
	}

	return canonicalTokens
}

// generateNGrams cria conjunto de n-gramas (k-shingles) de tokens
func generateNGrams(tokens []string, k int) map[string]int {
	ngrams := make(map[string]int)
	if len(tokens) < k {
		if len(tokens) > 0 {
			ngrams[strings.Join(tokens, " ")] = 1
		}
		return ngrams
	}
	for i := 0; i <= len(tokens)-k; i++ {
		key := strings.Join(tokens[i:i+k], " ")
		ngrams[key]++
	}
	return ngrams
}

// calculateSimilarity calcula o coeficiente de similaridade de Dice entre dois conjuntos de n-gramas
func calculateSimilarity(ngA, ngB map[string]int) (float64, int) {
	if len(ngA) == 0 || len(ngB) == 0 {
		return 0.0, 0
	}

	intersection := 0
	totalA := 0
	for k, countA := range ngA {
		totalA += countA
		if countB, ok := ngB[k]; ok {
			intersection += int(math.Min(float64(countA), float64(countB)))
		}
	}

	totalB := 0
	for _, countB := range ngB {
		totalB += countB
	}

	if totalA+totalB == 0 {
		return 0.0, 0
	}

	dice := (2.0 * float64(intersection)) / float64(totalA+totalB)
	return dice * 100.0, intersection
}

// CheckSubmissionsSimilarity executa a varredura e comparação de todos os códigos de uma tarefa
func (c *CanvasClient) CheckSubmissionsSimilarity(courseID, assignmentID string, threshold float64, normalizeIdentifiers bool) (*PlagiarismCheckResult, error) {
	if resolved, err := c.ResolveCourseID(courseID); err == nil && resolved != "" {
		courseID = resolved
	}
	if threshold <= 0 {
		threshold = 65.0 // Padrão 65% de similaridade
	}

	subs, err := c.GetSubmissionsDetails(courseID, assignmentID, false)
	if err != nil {
		return nil, fmt.Errorf("falha ao obter submissões da atividade: %w", err)
	}

	assignName := ""
	if assignObj, err := c.GetAssignment(courseID, assignmentID); err == nil && assignObj != nil {
		if aMap, ok := assignObj.(map[string]any); ok {
			if n, ok := aMap["name"].(string); ok {
				assignName = n
			}
		}
	}

	type studentCode struct {
		id          string
		name        string
		rawCode     string
		tokensRaw   []string
		tokensCanon []string
		ngramsRaw   map[string]int
		ngramsCanon map[string]int
	}

	var validCodes []studentCode

	for _, sub := range subs {
		code := strings.TrimSpace(sub.CleanBody)
		if code == "" && sub.Body != "" {
			code = strings.TrimSpace(CleanCanvasHTML(sub.Body))
		}

		// Filtra submissões que têm menos de 20 caracteres úteis
		if len(code) < 20 {
			continue
		}

		tokRaw := tokenizeAndCanonicalize(code, false)
		tokCanon := tokenizeAndCanonicalize(code, true)

		if len(tokRaw) < 8 {
			continue
		}

		validCodes = append(validCodes, studentCode{
			id:          sub.UserID,
			name:        sub.UserName,
			rawCode:     code,
			tokensRaw:   tokRaw,
			tokensCanon: tokCanon,
			ngramsRaw:   generateNGrams(tokRaw, 4),
			ngramsCanon: generateNGrams(tokCanon, 4),
		})
	}

	var pairs []PlagiarismPair

	// Comparação par a par (O(N^2 / 2))
	for i := 0; i < len(validCodes); i++ {
		for j := i + 1; j < len(validCodes); j++ {
			sA := validCodes[i]
			sB := validCodes[j]

			// Calcula similaridade com identificadores canônicos (detecta renomeação)
			simCanon, sharedCanon := calculateSimilarity(sA.ngramsCanon, sB.ngramsCanon)

			// Calcula similaridade de texto bruto (detecta cópia literal)
			simRaw, _ := calculateSimilarity(sA.ngramsRaw, sB.ngramsRaw)

			effectiveSim := math.Max(simCanon, simRaw)

			if effectiveSim >= threshold {
				var techniques []string

				if simRaw >= 90.0 {
					techniques = append(techniques, "Código quase idêntico (cópia literal)")
				} else if simCanon >= 80.0 && simRaw < 60.0 {
					techniques = append(techniques, "Renomeação ostensiva de variáveis/identificadores")
				} else if simCanon >= 75.0 {
					techniques = append(techniques, "Estrutura algorítmica e fluxo idênticos")
				} else {
					techniques = append(techniques, "Trechos e blocos de código compartilhados")
				}

				risk := "BAIXO"
				if effectiveSim >= 80.0 {
					risk = "ALTO"
				} else if effectiveSim >= 65.0 {
					risk = "MÉDIO"
				}

				pairs = append(pairs, PlagiarismPair{
					StudentA:             sA.name,
					StudentAID:           sA.id,
					StudentB:             sB.name,
					StudentBID:           sB.id,
					SimilarityPercentage: math.Round(effectiveSim*10) / 10,
					PlagiarismRisk:       risk,
					TechniquesDetected:   techniques,
					SharedTokenCount:     sharedCanon,
					TotalTokensA:         len(sA.tokensCanon),
					TotalTokensB:         len(sB.tokensCanon),
				})
			}
		}
	}

	// Ordena da maior similaridade para a menor
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].SimilarityPercentage > pairs[j].SimilarityPercentage
	})

	// Gera relatório em Markdown
	var sb strings.Builder
	sb.WriteString("### 🚨 Relatório de Similaridade e Detecção de Plágio (Afya)\n\n")
	if assignName != "" {
		sb.WriteString(fmt.Sprintf("**Atividade:** %s (ID %s)\n", assignName, assignmentID))
	} else {
		sb.WriteString(fmt.Sprintf("**Atividade ID:** %s\n", assignmentID))
	}
	sb.WriteString(fmt.Sprintf("- **Total de Submissões:** %d\n", len(subs)))
	sb.WriteString(fmt.Sprintf("- **Códigos Analisados:** %d\n", len(validCodes)))
	sb.WriteString(fmt.Sprintf("- **Limiar de Suspeita (Threshold):** %.1f%%\n", threshold))
	sb.WriteString(fmt.Sprintf("- **Pares Suspeitos Encontrados:** %d\n\n", len(pairs)))

	if len(pairs) == 0 {
		sb.WriteString("✅ **Nenhuma suspeita de plágio encontrada.** Todos os códigos apresentaram variações naturais e originais.\n")
	} else {
		sb.WriteString("| Aluno A (ID) | Aluno B (ID) | Similaridade | Risco | Técnicas Detectadas |\n")
		sb.WriteString("| :--- | :--- | :---: | :---: | :--- |\n")
		for _, p := range pairs {
			iconRisk := "🟢"
			if p.PlagiarismRisk == "ALTO" {
				iconRisk = "🔴"
			} else if p.PlagiarismRisk == "MÉDIO" {
				iconRisk = "🟡"
			}
			techStr := strings.Join(p.TechniquesDetected, "; ")
			sb.WriteString(fmt.Sprintf("| %s (%s) | %s (%s) | **%.1f%%** | %s %s | %s |\n",
				p.StudentA, p.StudentAID, p.StudentB, p.StudentBID, p.SimilarityPercentage, iconRisk, p.PlagiarismRisk, techStr))
		}
		sb.WriteString("\n> [!NOTE]\n")
		sb.WriteString("> Códigos classificados com risco **ALTO** indicam que a lógica, o encadeamento de comandos e os blocos de controle coincidem quase que integralmente, variando principalmente nomes de variáveis ou comentários.\n")
	}

	return &PlagiarismCheckResult{
		CourseID:            courseID,
		AssignmentID:        assignmentID,
		AssignmentName:      assignName,
		TotalSubmissions:    len(subs),
		TotalEvaluatedCodes: len(validCodes),
		Threshold:           threshold,
		SuspectPairsCount:   len(pairs),
		SuspectPairs:        pairs,
		MarkdownReport:      sb.String(),
	}, nil
}
