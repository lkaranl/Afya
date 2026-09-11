package main

import (
	"testing"
	"time"
)

func TestResolveCourseIDMocked(t *testing.T) {
	// Cria mock de disciplinas ativas do professor
	mockCourses := []map[string]any{
		{
			"id":              "11050",
			"name":            "ESTRUTURA DE DADOS - SÃO LUCAS JI-PARANÁ",
			"clean_name":      "Estrutura de Dados",
			"period":          "4º Período",
			"course_code":     "ED-4P-2026.2",
			"is_current_term": true,
		},
		{
			"id":              "11060",
			"name":            "ESTRUTURA DE DADOS - ENGENHARIA",
			"clean_name":      "Estrutura de Dados",
			"period":          "2º Período",
			"course_code":     "ED-2P-2026.2",
			"is_current_term": true,
		},
		{
			"id":              "9900",
			"name":            "PROGRAMAÇÃO ORIENTADA A OBJETOS",
			"clean_name":      "Programação Orientada a Objetos",
			"period":          "3º Período",
			"course_code":     "POO-2025.2",
			"is_current_term": false, // anterior
		},
	}

	client := &CanvasClient{
		Cache: NewMemoryCache(5 * time.Minute),
	}

	// Injeta cursos com chave 'courses' no cache
	client.Cache.Set("courses", mockCourses, 5*time.Minute)

	testCases := []struct {
		input       string
		expectedID  string
		expectError bool
	}{
		// 1. O usuário escolhe opção "1"
		{"1", "11050", false},
		{"opção 1", "11050", false},
		{"2", "11060", false},
		{"opção 2", "11060", false},

		// 2. A IA passa o nome didático sintetizado com período entre parênteses
		{"ESTRUTURA DE DADOS (4º Período)", "11050", false},
		{"Estrutura de Dados (2º Período)", "11060", false},

		// 3. A IA ou usuário passa com hífen ou espaço
		{"Estrutura de Dados - 4º Período", "11050", false},
		{"Estrutura de Dados 2º Período", "11060", false},

		// 4. Apenas o período
		{"4º Período", "11050", false},
		{"2º Período", "11060", false},
		{"4", "11050", false},
		{"2", "11060", false},

		// 5. ID direto real
		{"11050", "11050", false},
		{"11060", "11060", false},
	}

	for _, tc := range testCases {
		res, err := client.ResolveCourseID(tc.input)
		if tc.expectError {
			if err == nil {
				t.Errorf("Para input %q esperava erro, mas obteve %q", tc.input, res)
			}
		} else {
			if err != nil {
				t.Errorf("Para input %q erro inesperado: %v", tc.input, err)
			}
			if res != tc.expectedID {
				t.Errorf("Para input %q esperado ID %q, mas obteve %q", tc.input, tc.expectedID, res)
			}
		}
	}
}
