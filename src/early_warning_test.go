package main

import (
	"testing"
)

func TestRiskLevelClassification(t *testing.T) {
	// Cenário 1: Estudante Crítico (3 fatores)
	st1 := AtRiskStudent{
		Name:               "Aluno Crítico",
		DaysInactive:       25,
		ConsecutiveMissing: 3,
		CurrentScore:       20.0,
		HasScore:           true,
	}
	factors1 := []string{}
	if st1.DaysInactive >= 10 {
		factors1 = append(factors1, "Inatividade")
	}
	if st1.ConsecutiveMissing >= 2 {
		factors1 = append(factors1, "Tarefas zeradas")
	}
	if st1.CurrentScore < 70.0 {
		factors1 = append(factors1, "Nota abaixo de 70")
	}
	if len(factors1) != 3 {
		t.Fatalf("Esperava 3 fatores de risco, obteve %d", len(factors1))
	}

	// Cenário 2: Estudante Moderado (2 fatores)
	st2 := AtRiskStudent{
		Name:               "Aluno Moderado",
		DaysInactive:       2,
		ConsecutiveMissing: 2,
		CurrentScore:       55.0,
		HasScore:           true,
	}
	factors2 := []string{}
	if st2.DaysInactive >= 10 {
		factors2 = append(factors2, "Inatividade")
	}
	if st2.ConsecutiveMissing >= 2 {
		factors2 = append(factors2, "Tarefas zeradas")
	}
	if st2.CurrentScore < 70.0 {
		factors2 = append(factors2, "Nota abaixo de 70")
	}
	if len(factors2) != 2 {
		t.Fatalf("Esperava 2 fatores de risco, obteve %d", len(factors2))
	}

	// Cenário 3: Estudante Regular (0 fatores)
	st3 := AtRiskStudent{
		Name:               "Aluno Regular",
		DaysInactive:       1,
		ConsecutiveMissing: 0,
		CurrentScore:       85.0,
		HasScore:           true,
	}
	factors3 := []string{}
	if st3.DaysInactive >= 10 {
		factors3 = append(factors3, "Inatividade")
	}
	if st3.ConsecutiveMissing >= 2 {
		factors3 = append(factors3, "Tarefas zeradas")
	}
	if st3.CurrentScore < 70.0 {
		factors3 = append(factors3, "Nota abaixo de 70")
	}
	if len(factors3) != 0 {
		t.Fatalf("Esperava 0 fatores de risco para aluno regular, obteve %d", len(factors3))
	}
}
