package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
)

// RiskLevel define a gravidade do risco acadêmico do aluno
type RiskLevel string

const (
	RiskCritical RiskLevel = "CRITICO" // 🔴 3 fatores
	RiskModerate RiskLevel = "MODERADO" // 🟡 2 fatores
	RiskWarning  RiskLevel = "ATENCAO"  // 🟠 1 fator
	RiskRegular  RiskLevel = "REGULAR"  // 🟢 Em dia
)

// AtRiskStudent detalha o diagnóstico de um estudante na disciplina
type AtRiskStudent struct {
	UserID                string    `json:"user_id"`
	Name                  string    `json:"name"`
	Email                 string    `json:"email,omitempty"`
	LastActivityAt        string    `json:"last_activity_at"`
	LastActivityBR        string    `json:"last_activity_br"`
	DaysInactive          int       `json:"days_inactive"`
	TotalActivitySeconds  int64     `json:"total_activity_seconds"`
	CurrentScore          float64   `json:"current_score"`
	HasScore              bool      `json:"has_score"`
	ConsecutiveMissing    int       `json:"consecutive_missing_or_zero"`
	TotalMissingOrZero    int       `json:"total_missing_or_zero"`
	RiskLevel             RiskLevel `json:"risk_level"`
	RiskScore             int       `json:"risk_score"` // 0 a 3 fatores
	AlertFactors          []string  `json:"alert_factors"`
	SuggestedIntervention string    `json:"suggested_intervention"`
}

// CourseAtRiskSummary consolida as estatísticas da turma
type CourseAtRiskSummary struct {
	TotalStudents   int     `json:"total_students"`
	CriticalCount   int     `json:"critical_count"`
	ModerateCount   int     `json:"moderate_count"`
	WarningCount    int     `json:"warning_count"`
	RegularCount    int     `json:"regular_count"`
	ClassAverage    float64 `json:"class_average"`
	InactivityCutoff int    `json:"inactivity_cutoff_days"`
	GradeCutoff     float64 `json:"grade_cutoff"`
}

// CourseAtRiskReport é o relatório completo emitido para o professor
type CourseAtRiskReport struct {
	CourseID      string              `json:"course_id"`
	CourseName    string              `json:"course_name"`
	GeneratedAtBR string              `json:"generated_at_br"`
	Summary       CourseAtRiskSummary `json:"summary"`
	Students      []AtRiskStudent     `json:"students"`
	MarkdownTable string              `json:"markdown_table"`
}

// DetectAtRiskStudents cruza inatividade, notas abaixo da média de corte da Afya e tarefas zeradas/faltantes
func (c *CanvasClient) DetectAtRiskStudents(courseID string, inactivityDays int, gradeCutoff float64, consecutiveThreshold int) (*CourseAtRiskReport, error) {
	if courseID == "" {
		return nil, fmt.Errorf("ID da disciplina é obrigatório")
	}

	if inactivityDays <= 0 {
		inactivityDays = 10
	}
	if gradeCutoff <= 0 {
		gradeCutoff = 70.0 // Padrão institucional Afya
	}
	if consecutiveThreshold <= 0 {
		consecutiveThreshold = 2
	}

	// 1. Busca alunos com histórico de enrollment
	endpointUsers := fmt.Sprintf("/api/v1/courses/%s/users?enrollment_type[]=student&include[]=enrollments&per_page=100", url.PathEscape(courseID))
	dataUsers, status, err := c.Request("GET", endpointUsers, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar estudantes da disciplina (HTTP %d): %w", status, err)
	}

	var rawUsers []map[string]any
	if err := json.Unmarshal(dataUsers, &rawUsers); err != nil {
		return nil, fmt.Errorf("erro ao decodificar estudantes: %w", err)
	}

	// 2. Busca submissões de todos os alunos da disciplina
	endpointSubs := fmt.Sprintf("/api/v1/courses/%s/students/submissions?student_ids[]=all&per_page=100", url.PathEscape(courseID))
	dataSubs, _, _ := c.Request("GET", endpointSubs, nil)
	var rawSubs []map[string]any
	if len(dataSubs) > 0 {
		_ = json.Unmarshal(dataSubs, &rawSubs)
	}

	// Agrupar submissões por UserID
	subsByUser := make(map[string][]map[string]any)
	for _, sub := range rawSubs {
		uID := fmt.Sprintf("%v", sub["user_id"])
		subsByUser[uID] = append(subsByUser[uID], sub)
	}

	// Ordenar submissões de cada aluno por assignment_id para checar consecutivas
	for uID := range subsByUser {
		userSubs := subsByUser[uID]
		sort.Slice(userSubs, func(i, j int) bool {
			return fmt.Sprintf("%v", userSubs[i]["assignment_id"]) < fmt.Sprintf("%v", userSubs[j]["assignment_id"])
		})
		subsByUser[uID] = userSubs
	}

	now := time.Now()
	var atRiskList []AtRiskStudent
	summary := CourseAtRiskSummary{
		TotalStudents:    len(rawUsers),
		InactivityCutoff: inactivityDays,
		GradeCutoff:      gradeCutoff,
	}

	var totalScoreSum float64
	var scoredStudentsCount int

	for _, u := range rawUsers {
		uID := fmt.Sprintf("%v", u["id"])
		name := fmt.Sprintf("%v", u["name"])
		email := ""
		if eVal, ok := u["email"].(string); ok {
			email = eVal
		}

		student := AtRiskStudent{
			UserID:       uID,
			Name:         name,
			Email:        email,
			AlertFactors: []string{},
		}

		// Extrair dados do enrollment
		var lastActivityStr string
		var totalSecs int64
		var currentScore float64
		hasScore := false

		if enrolls, ok := u["enrollments"].([]any); ok && len(enrolls) > 0 {
			if eMap, ok := enrolls[0].(map[string]any); ok {
				if act, ok := eMap["last_activity_at"].(string); ok && act != "" {
					lastActivityStr = act
				}
				if actTime, ok := eMap["total_activity_time"].(float64); ok {
					totalSecs = int64(actTime)
				}
				if gMap, ok := eMap["grades"].(map[string]any); ok {
					if cScore, ok := gMap["current_score"].(float64); ok {
						currentScore = cScore
						hasScore = true
					}
				}
			}
		}

		student.LastActivityAt = lastActivityStr
		student.TotalActivitySeconds = totalSecs
		student.CurrentScore = currentScore
		student.HasScore = hasScore

		// Calcular dias inativo
		if lastActivityStr == "" {
			student.DaysInactive = 999 // Nunca acessou
			student.LastActivityBR = "Nunca acessou"
		} else {
			student.LastActivityBR = formatBRT(lastActivityStr)
			if parsed, err := time.Parse(time.RFC3339, lastActivityStr); err == nil {
				diff := now.Sub(parsed)
				student.DaysInactive = int(diff.Hours() / 24)
			}
		}

		// Analisar submissões
		userSubs := subsByUser[uID]
		totalMissing := 0
		maxConsecutive := 0
		currentConsecutive := 0

		for _, s := range userSubs {
			isMissing := false
			if mVal, ok := s["missing"].(bool); ok && mVal {
				isMissing = true
			}
			var scoreVal float64
			hasSubScore := false
			if sc, ok := s["score"].(float64); ok {
				scoreVal = sc
				hasSubScore = true
			}
			wfState := fmt.Sprintf("%v", s["workflow_state"])

			// É considerada zerada/faltante se missing, ou unsubmitted, ou score == 0 explicitamente avaliado
			if isMissing || wfState == "unsubmitted" || (hasSubScore && scoreVal == 0) {
				totalMissing++
				currentConsecutive++
				if currentConsecutive > maxConsecutive {
					maxConsecutive = currentConsecutive
				}
			} else {
				currentConsecutive = 0
			}
		}

		student.TotalMissingOrZero = totalMissing
		student.ConsecutiveMissing = maxConsecutive

		if hasScore {
			totalScoreSum += currentScore
			scoredStudentsCount++
		}

		// Avaliar Fatores de Alerta
		// Fator 1: Inatividade prolongada
		if student.DaysInactive >= inactivityDays {
			if student.DaysInactive >= 999 {
				student.AlertFactors = append(student.AlertFactors, "Sem registro de acesso ao Canvas")
			} else {
				student.AlertFactors = append(student.AlertFactors, fmt.Sprintf("%d dias sem acessar a matéria", student.DaysInactive))
			}
		}

		// Fator 2: Entregas consecutivas zeradas/faltantes
		if student.ConsecutiveMissing >= consecutiveThreshold {
			student.AlertFactors = append(student.AlertFactors, fmt.Sprintf("%d atividades consecutivas zeradas/faltantes", student.ConsecutiveMissing))
		}

		// Fator 3: Média abaixo do corte da Afya (70 pts)
		if !hasScore || student.CurrentScore < gradeCutoff {
			if !hasScore {
				student.AlertFactors = append(student.AlertFactors, "Sem nota lançada até o momento")
			} else {
				student.AlertFactors = append(student.AlertFactors, fmt.Sprintf("Média atual %.1f pts (abaixo do corte de %.0f pts)", student.CurrentScore, gradeCutoff))
			}
		}

		student.RiskScore = len(student.AlertFactors)

		// Classificar nível de risco
		switch student.RiskScore {
		case 3:
			student.RiskLevel = RiskCritical
			student.SuggestedIntervention = "🚨 Contato prioritário urgente (Coordenação / NAPED / Busca Ativa)"
			summary.CriticalCount++
		case 2:
			student.RiskLevel = RiskModerate
			student.SuggestedIntervention = "⚠️ Alerta individual pelo Inbox do Canvas e sondagem de dúvidas"
			summary.ModerateCount++
		case 1:
			student.RiskLevel = RiskWarning
			student.SuggestedIntervention = "🔍 Monitoramento em sala e incentivo para entrega das pendências"
			summary.WarningCount++
		default:
			student.RiskLevel = RiskRegular
			student.SuggestedIntervention = "✅ Desempenho regular e engajado"
			summary.RegularCount++
		}

		atRiskList = append(atRiskList, student)
	}

	if scoredStudentsCount > 0 {
		summary.ClassAverage = totalScoreSum / float64(scoredStudentsCount)
	}

	// Ordenar a lista: mais críticos primeiro, seguidos por menor nota e maior inatividade
	sort.Slice(atRiskList, func(i, j int) bool {
		if atRiskList[i].RiskScore != atRiskList[j].RiskScore {
			return atRiskList[i].RiskScore > atRiskList[j].RiskScore
		}
		if atRiskList[i].CurrentScore != atRiskList[j].CurrentScore {
			return atRiskList[i].CurrentScore < atRiskList[j].CurrentScore
		}
		return atRiskList[i].DaysInactive > atRiskList[j].DaysInactive
	})

	// Gerar Tabela Executiva em Markdown
	var md strings.Builder
	md.WriteString(fmt.Sprintf("## 📊 Painel de Identificação Precoce e Risco de Evasão\n"))
	md.WriteString(fmt.Sprintf("**Disciplina ID:** `%s` | **Data:** %s\n\n", courseID, formatBRT(now.Format(time.RFC3339))))

	md.WriteString("### 📈 Panorama Consolidado da Turma:\n")
	md.WriteString(fmt.Sprintf("- **Total de Estudantes Matriculados:** %d\n", summary.TotalStudents))
	md.WriteString(fmt.Sprintf("- 🔴 **Risco Crítico (3 fatores):** **%d alunos** (Inatividade $> %d$d + Tarefas zeradas $\\ge %d$ + Média $< %.0f$ pts)\n",
		summary.CriticalCount, inactivityDays, consecutiveThreshold, gradeCutoff))
	md.WriteString(fmt.Sprintf("- 🟡 **Risco Moderado (2 fatores):** **%d alunos**\n", summary.ModerateCount))
	md.WriteString(fmt.Sprintf("- 🟠 **Atenção (1 fator):** **%d alunos**\n", summary.WarningCount))
	md.WriteString(fmt.Sprintf("- 🟢 **Regulares / Em Dia:** **%d alunos**\n", summary.RegularCount))
	if scoredStudentsCount > 0 {
		md.WriteString(fmt.Sprintf("- **Média Geral da Turma:** **%.1f pontos**\n\n", summary.ClassAverage))
	} else {
		md.WriteString("\n")
	}

	md.WriteString("### 🚨 Estudantes que Exigem Intervenção Pedagógica:\n\n")
	md.WriteString("| Aluno | ID | Nível de Risco | Último Acesso | Faltas/Zeros | Média Atual | Fatores de Alerta |\n")
	md.WriteString("| :--- | :--- | :--- | :--- | :--- | :--- | :--- |\n")

	atRiskCount := 0
	for _, st := range atRiskList {
		// Exibir todos que possuem pelo menos 1 fator de alerta na tabela principal
		if st.RiskScore > 0 {
			atRiskCount++
			icon := "🔴 CRÍTICO"
			if st.RiskLevel == RiskModerate {
				icon = "🟡 MODERADO"
			} else if st.RiskLevel == RiskWarning {
				icon = "🟠 ATENÇÃO"
			}

			scoreStr := fmt.Sprintf("%.1f", st.CurrentScore)
			if !st.HasScore {
				scoreStr = "Sem nota"
			}

			factorsStr := strings.Join(st.AlertFactors, "; ")

			lastAct := st.LastActivityBR
			if st.DaysInactive > 0 && st.DaysInactive < 999 {
				lastAct = fmt.Sprintf("%s (%dd)", st.LastActivityBR, st.DaysInactive)
			}

			md.WriteString(fmt.Sprintf("| **%s** | `%s` | %s | %s | %d consecutivas | %s pts | %s |\n",
				st.Name, st.UserID, icon, lastAct, st.ConsecutiveMissing, scoreStr, factorsStr))
		}
	}

	if atRiskCount == 0 {
		md.WriteString("| *(Nenhum aluno em risco)* | - | 🟢 REGULAR | - | 0 | - | Turma 100% em dia com os critérios! |\n")
	}

	report := &CourseAtRiskReport{
		CourseID:      courseID,
		GeneratedAtBR: formatBRT(now.Format(time.RFC3339)),
		Summary:       summary,
		Students:      atRiskList,
		MarkdownTable: md.String(),
	}

	return report, nil
}
