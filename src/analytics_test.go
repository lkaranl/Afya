package main

import (
	"strings"
	"testing"
)

func TestCourseAnalyticsCalculationAndCard(t *testing.T) {
	summary := &CourseAnalyticsSummary{
		CourseID:            160754,
		CourseName:          "Estrutura de Dados",
		Period:              "4º Período",
		TotalStudents:       30,
		AverageClassScore:   76.5,
		ApprovedDirectCount: 22,
		FinalExamCount:      5,
		AtRiskCount:         3,
		AssignmentsStatus: []AnalyticsAssignmentStatus{
			{
				AssignmentID:     101,
				Title:            "Atividade 1: Ponteiros em C",
				GradedCount:      24,
				PendingCount:     6,
				UnsubmittedCount: 0,
				TotalStudents:    30,
			},
		},
		GradeDistribution: []AnalyticsGradeBand{
			{Label: "0 - 20 pts", Count: 1},
			{Label: "21 - 40 pts", Count: 2},
			{Label: "41 - 60 pts", Count: 5},
			{Label: "61 - 80 pts", Count: 12},
			{Label: "81 - 100 pts", Count: 10},
		},
		StudentsRiskPlot: []AnalyticsStudentRisk{
			{
				StudentID:    1,
				Name:         "Aluno Alerta",
				DaysInactive: 14,
				AverageScore: 32.0,
				RiskCategory: "critico",
			},
		},
		GeneratedAt: "13/09/2026 às 16:00",
	}

	// 1. Testa a geração do card visual de texto para o Telegram
	cardText := GenerateTelegramAnalyticsCard(summary)

	if !strings.Contains(strings.ToUpper(cardText), "ESTRUTURA DE DADOS") {
		t.Errorf("Esperava nome da disciplina no card, obteve:\n%s", cardText)
	}

	if !strings.Contains(cardText, "Aprovação Direta (≥70)") {
		t.Errorf("Esperava seção de Aprovação Direta no card")
	}

	if !strings.Contains(cardText, "Fila do SpeedGrader:* 6 avaliações") {
		t.Errorf("Esperava contagem de pendências no SpeedGrader")
	}

	if !strings.Contains(cardText, "Radar de Evasão NAPED:* 1 estudantes") {
		t.Errorf("Esperava contagem de estudantes em risco")
	}

	// Valida barras em blocos Unicode
	if !strings.Contains(cardText, "[███") {
		t.Errorf("Esperava representação gráfica de barras em Unicode")
	}
}
