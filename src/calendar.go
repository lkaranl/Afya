package main

import (
	"fmt"
	"strings"
)

// AcademicEvent representa uma entrada oficial do Calendário Acadêmico
type AcademicEvent struct {
	Date      string `json:"date"`
	DayOfWeek string `json:"day_of_week"`
	Title     string `json:"title"`
	Category  string `json:"category"`
	Month     int    `json:"month"`
	IsSHE     bool   `json:"is_she"`
}

// SaturdayClass representa a compensação de um sábado letivo
type SaturdayClass struct {
	Date            string `json:"date"`
	ReplacesWeekday string `json:"replaces_weekday"`
	Description     string `json:"description"`
}

// KeyDeadlines resume as datas cruciais do semestre para o professor e coordenação
type KeyDeadlines struct {
	StartSemester       string `json:"start_semester"`
	N1EvaluationsSHE    string `json:"n1_evaluations_she"`
	N1GradesDeadline    string `json:"n1_grades_deadline"`
	N2EvaluationsSHE    string `json:"n2_evaluations_she"`
	SecondChanceSHE     string `json:"second_chance_she"`
	N2GradesDeadline    string `json:"n2_grades_deadline"`
	FinalExams          string `json:"final_exams"`
	SemesterClosing     string `json:"semester_closing"`
}

// AcademicCalendarResult encapsula a resposta completa da consulta ao calendário
type AcademicCalendarResult struct {
	Term            string          `json:"term"`
	Institution     string          `json:"institution"`
	KeyDeadlines    KeyDeadlines    `json:"key_deadlines"`
	SaturdayClasses []SaturdayClass `json:"saturday_classes"`
	TotalEvents     int             `json:"total_events"`
	FilteredEvents  []AcademicEvent `json:"filtered_events"`
	MarkdownTable   string          `json:"markdown_table"`
}

// GetOfficialCalendar2026_2 retorna a base canônica do semestre 2026.2
func GetOfficialCalendar2026_2() []AcademicEvent {
	return []AcademicEvent{
		// Julho
		{Date: "01 a 04/07/2026", DayOfWeek: "Qua - Sáb", Title: "Avaliações N2 - SHE (Encerramento 2026.1)", Category: "SHE", Month: 7, IsSHE: true},
		{Date: "06 e 07/07/2026", DayOfWeek: "Seg - Ter", Title: "Período de Avaliação de 2ª Chamada (2026.1)", Category: "SHE", Month: 7, IsSHE: true},
		{Date: "07/07/2026", DayOfWeek: "Terça", Title: "Divulgação de Notas N2 e 2ª Chamada (2026.1)", Category: "Docentes", Month: 7, IsSHE: true},
		{Date: "08 a 10/07/2026", DayOfWeek: "Qua - Sex", Title: "Período de Exames Finais (2026.1)", Category: "Geral", Month: 7, IsSHE: true},
		{Date: "10/07/2026", DayOfWeek: "Sexta", Title: "Entrega dos Diários e Fechamento do Semestre 2026.1", Category: "Docentes", Month: 7, IsSHE: true},
		{Date: "20/07/2026", DayOfWeek: "Segunda", Title: "Início do Internato - Medicina", Category: "Medicina", Month: 7, IsSHE: false},
		{Date: "23 e 24/07/2026", DayOfWeek: "Qui - Sex", Title: "Semana de Desenvolvimento Docente (SDD)", Category: "Docentes", Month: 7, IsSHE: true},
		{Date: "27/07/2026", DayOfWeek: "Segunda", Title: "Início do Semestre Letivo 2026.2 (Acolhimento Ingressantes e Veteranos)", Category: "Geral", Month: 7, IsSHE: true},
		{Date: "28/07/2026", DayOfWeek: "Terça", Title: "Início - Pesquisa de Qualidade de Vida do Estudante", Category: "Institucional", Month: 7, IsSHE: true},
		{Date: "29/07/2026", DayOfWeek: "Quarta", Title: "Reunião CONSUP / CONSEPE", Category: "Institucional", Month: 7, IsSHE: false},

		// Agosto
		{Date: "01 a 30/08/2026", DayOfWeek: "-", Title: "Prazo para Solicitação de Outorga de Grau 2026.2", Category: "Institucional", Month: 8, IsSHE: true},
		{Date: "02/08/2026", DayOfWeek: "Domingo", Title: "Simulado ENAMED 2026.2", Category: "Medicina", Month: 8, IsSHE: false},
		{Date: "04/08/2026", DayOfWeek: "Terça", Title: "Início das Disciplinas ONA’s e ONS’s", Category: "Online", Month: 8, IsSHE: true},
		{Date: "08/08/2026", DayOfWeek: "Sábado", Title: "Sábado Letivo - SHE e Medicina", Category: "Sábado Letivo", Month: 8, IsSHE: true},
		{Date: "10/08/2026", DayOfWeek: "Segunda", Title: "Divulgação dos Planos de Ensino", Category: "Docentes", Month: 8, IsSHE: true},
		{Date: "12/08/2026", DayOfWeek: "Quarta", Title: "Reunião CEP / CEUA", Category: "Institucional", Month: 8, IsSHE: false},
		{Date: "12/08/2026", DayOfWeek: "Quarta", Title: "Avaliação Integrada I - Internato (1ª rotação)", Category: "Medicina", Month: 8, IsSHE: false},
		{Date: "15/08/2026", DayOfWeek: "Sábado", Title: "Sábado Letivo Medicina e SHE (Horário de Terça-feira)", Category: "Sábado Letivo", Month: 8, IsSHE: true},
		{Date: "22/08/2026", DayOfWeek: "Sábado", Title: "Sábado Letivo Medicina e SHE (Horário de Quarta-feira)", Category: "Sábado Letivo", Month: 8, IsSHE: true},
		{Date: "24 a 27/08/2026", DayOfWeek: "Seg - Qui", Title: "Outorga de Grau Institucional 2026.1", Category: "Institucional", Month: 8, IsSHE: true},
		{Date: "28/08/2026", DayOfWeek: "Sexta", Title: "Outorga de Grau Institucional 2026.1 (Gabinete)", Category: "Institucional", Month: 8, IsSHE: true},
		{Date: "28/08/2026", DayOfWeek: "Sexta", Title: "Avaliação Integrada II - Internato (1ª rotação)", Category: "Medicina", Month: 8, IsSHE: false},
		{Date: "28/08/2026", DayOfWeek: "Sexta", Title: "Término - Pesquisa de Qualidade de Vida do Estudante", Category: "Institucional", Month: 8, IsSHE: true},
		{Date: "29/08/2026", DayOfWeek: "Sábado", Title: "Sábado Letivo Medicina e SHE (Horário de Segunda-feira)", Category: "Sábado Letivo", Month: 8, IsSHE: true},
		{Date: "31/08 a 03/09/2026", DayOfWeek: "Seg - Qui", Title: "Semana de Inovação e Empreendedorismo (PROPPEXI / SEBRAE)", Category: "Institucional", Month: 8, IsSHE: true},

		// Setembro
		{Date: "04/09/2026", DayOfWeek: "Sexta", Title: "2ª Chamada de Avaliação Integrada (Internato 1ª rotação)", Category: "Medicina", Month: 9, IsSHE: false},
		{Date: "05/09/2026", DayOfWeek: "Sábado", Title: "Sábado Letivo - Medicina", Category: "Medicina", Month: 9, IsSHE: false},
		{Date: "07/09/2026", DayOfWeek: "Segunda", Title: "Feriado Nacional — Independência do Brasil", Category: "Feriado", Month: 9, IsSHE: true},
		{Date: "12/09/2026", DayOfWeek: "Sábado", Title: "Sábado Letivo Medicina e SHE (Horário de Sexta-feira)", Category: "Sábado Letivo", Month: 9, IsSHE: true},
		{Date: "16 a 18/09/2026", DayOfWeek: "Qua - Sex", Title: "Afya Global Meeting / Avaliações N1 - Medicina", Category: "Medicina", Month: 9, IsSHE: false},
		{Date: "17 e 18/09/2026", DayOfWeek: "Qui - Sex", Title: "Início do Agendamento para Provas Presenciais (ONA’s e ONS’s)", Category: "Online", Month: 9, IsSHE: true},
		{Date: "19/09/2026", DayOfWeek: "Sábado", Title: "Sábado Letivo - SHE e Medicina", Category: "Sábado Letivo", Month: 9, IsSHE: true},
		{Date: "21/09/2026", DayOfWeek: "Segunda", Title: "TPI - Medicina", Category: "Medicina", Month: 9, IsSHE: false},
		{Date: "21 a 26/09/2026", DayOfWeek: "Seg - Sáb", Title: "Período de Avaliações N1 — SHE (Ciência da Computação e Engenharias)", Category: "SHE", Month: 9, IsSHE: true},
		{Date: "26/09/2026", DayOfWeek: "Sábado", Title: "Sábado Letivo Medicina e SHE (Horário de Segunda-feira)", Category: "Sábado Letivo", Month: 9, IsSHE: true},
		{Date: "30/09/2026", DayOfWeek: "Quarta", Title: "Avaliação Integrada I - Internato (2ª rotação)", Category: "Medicina", Month: 9, IsSHE: false},
		{Date: "30/09/2026", DayOfWeek: "Quarta", Title: "Reunião CONSUP / CONSEPE", Category: "Institucional", Month: 9, IsSHE: false},

		// Outubro
		{Date: "03/10/2026", DayOfWeek: "Sábado", Title: "Sábado Letivo Medicina e SHE (Horário de Quinta-feira)", Category: "Sábado Letivo", Month: 10, IsSHE: true},
		{Date: "03/10/2026", DayOfWeek: "Sábado", Title: "TPI - Fisioterapia", Category: "Saúde", Month: 10, IsSHE: false},
		{Date: "05/10/2026", DayOfWeek: "Segunda", Title: "Início da Avaliação Institucional (CPA)", Category: "Institucional", Month: 10, IsSHE: true},
		{Date: "05/10/2026", DayOfWeek: "Segunda", Title: "Prazo Final para Divulgação de Notas N1 no Canvas", Category: "Docentes", Month: 10, IsSHE: true},
		{Date: "07/10/2026", DayOfWeek: "Quarta", Title: "TPI - Direito", Category: "Direito", Month: 10, IsSHE: false},
		{Date: "08/10/2026", DayOfWeek: "Quinta", Title: "2ª Chamada TPI Medicina / 1ª Chamada PROUNI e FIES", Category: "Medicina", Month: 10, IsSHE: false},
		{Date: "12/10/2026", DayOfWeek: "Segunda", Title: "Feriado Nacional — Nossa Senhora Aparecida", Category: "Feriado", Month: 10, IsSHE: true},
		{Date: "13/10/2026", DayOfWeek: "Terça", Title: "Recesso Acadêmico — Antecipação do Dia dos Professores", Category: "Recesso", Month: 10, IsSHE: true},
		{Date: "14 e 15/10/2026", DayOfWeek: "Qua - Qui", Title: "Término da 1ª Parte das Disciplinas ON's / Início da 2ª Parte", Category: "Online", Month: 10, IsSHE: true},
		{Date: "14/10/2026", DayOfWeek: "Quarta", Title: "Avaliação Integrada II - Internato (2ª rotação)", Category: "Medicina", Month: 10, IsSHE: false},
		{Date: "16/10/2026", DayOfWeek: "Sexta", Title: "Término do Agendamento para Provas Presenciais (ONA’s e ONS’s)", Category: "Online", Month: 10, IsSHE: true},
		{Date: "16/10/2026", DayOfWeek: "Sexta", Title: "2ª Chamada do TPI - Medicina (PROUNI e FIES)", Category: "Medicina", Month: 10, IsSHE: false},
		{Date: "20/10/2026", DayOfWeek: "Terça", Title: "TPI - Enfermagem", Category: "Saúde", Month: 10, IsSHE: false},
		{Date: "21/10/2026", DayOfWeek: "Quarta", Title: "TPI - Psicologia", Category: "Saúde", Month: 10, IsSHE: false},
		{Date: "22, 26 a 29/10/2026", DayOfWeek: "-", Title: "N2 - Medicina", Category: "Medicina", Month: 10, IsSHE: false},
		{Date: "23/10/2026", DayOfWeek: "Sexta", Title: "Término da Avaliação Institucional (CPA)", Category: "Institucional", Month: 10, IsSHE: true},
		{Date: "24/10/2026", DayOfWeek: "Sábado", Title: "Sábado Letivo - SHE e Medicina", Category: "Sábado Letivo", Month: 10, IsSHE: true},
		{Date: "26 a 30/10/2026", DayOfWeek: "Seg - Sex", Title: "Semana de Saúde Integral", Category: "Institucional", Month: 10, IsSHE: true},
		{Date: "30/10/2026", DayOfWeek: "Sexta", Title: "Início do Simulado ONLINE para Disciplinas HB e ONA’s via Canvas", Category: "Online", Month: 10, IsSHE: true},
		{Date: "31/10/2026", DayOfWeek: "Sábado", Title: "Sábado Letivo Medicina e SHE (Horário de Terça-feira)", Category: "Sábado Letivo", Month: 10, IsSHE: true},

		// Novembro
		{Date: "02/11/2026", DayOfWeek: "Segunda", Title: "Feriado Nacional — Finados", Category: "Feriado", Month: 11, IsSHE: true},
		{Date: "03/11/2026", DayOfWeek: "Terça", Title: "Término do Simulado ONLINE para Disciplinas HB e ONA’s via Canvas", Category: "Online", Month: 11, IsSHE: true},
		{Date: "03 a 06/11/2026", DayOfWeek: "Ter - Sex", Title: "N2 - Medicina", Category: "Medicina", Month: 11, IsSHE: false},
		{Date: "04 a 23/11/2026", DayOfWeek: "Qua - Seg", Title: "Provas Presenciais das Disciplinas ONA’s e ONS’s", Category: "Online", Month: 11, IsSHE: true},
		{Date: "07/11/2026", DayOfWeek: "Sábado", Title: "Sábado Letivo Medicina e SHE (Horário de Sexta-feira)", Category: "Sábado Letivo", Month: 11, IsSHE: true},
		{Date: "09/11/2026", DayOfWeek: "Segunda", Title: "Final da 2ª Parte das Disciplinas ONA’s e ONS’s", Category: "Online", Month: 11, IsSHE: true},
		{Date: "10/11/2026", DayOfWeek: "Terça", Title: "2ª Chamada TPI - SHE", Category: "SHE", Month: 11, IsSHE: true},
		{Date: "11/11/2026", DayOfWeek: "Quarta", Title: "Avaliação Integrada I - Internato (3ª rotação)", Category: "Medicina", Month: 11, IsSHE: false},
		{Date: "14/11/2026", DayOfWeek: "Sábado", Title: "Sábado Letivo Medicina e SHE (Horário de Segunda-feira)", Category: "Sábado Letivo", Month: 11, IsSHE: true},
		{Date: "16 a 27/11/2026", DayOfWeek: "Seg - Sex", Title: "OSCE - Medicina", Category: "Medicina", Month: 11, IsSHE: false},
		{Date: "19/11/2026", DayOfWeek: "Quinta", Title: "9ª Socialização da Extensão Universitária", Category: "Institucional", Month: 11, IsSHE: true},
		{Date: "20/11/2026", DayOfWeek: "Sexta", Title: "Feriado Nacional — Dia da Consciência Negra", Category: "Feriado", Month: 11, IsSHE: true},
		{Date: "23 e 24/11/2026", DayOfWeek: "Seg - Ter", Title: "N3 - Medicina", Category: "Medicina", Month: 11, IsSHE: false},
		{Date: "23 a 28/11/2026", DayOfWeek: "Seg - Sáb", Title: "Período de Avaliações N2 — SHE (Ciência da Computação e Engenharias)", Category: "SHE", Month: 11, IsSHE: true},
		{Date: "25/11/2026", DayOfWeek: "Quarta", Title: "Avaliação Integrada II - Internato (3ª rotação)", Category: "Medicina", Month: 11, IsSHE: false},
		{Date: "28/11/2026", DayOfWeek: "Sábado", Title: "Sábado Letivo - Medicina", Category: "Medicina", Month: 11, IsSHE: false},
		{Date: "30/11/2026", DayOfWeek: "Segunda", Title: "2ª Chamada - Provas Presenciais das Disciplinas ONA’s e ONS’s", Category: "Online", Month: 11, IsSHE: true},
		{Date: "30/11 a 02/12/2026", DayOfWeek: "Seg - Qua", Title: "Período de Avaliação de 2ª Chamada — SHE", Category: "SHE", Month: 11, IsSHE: true},
		{Date: "30/11 a 04/12/2026", DayOfWeek: "Seg - Sex", Title: "Agendamento para Exame Final - Disciplinas ONA’s e ONS’s", Category: "Online", Month: 11, IsSHE: true},

		// Dezembro
		{Date: "01/12/2026", DayOfWeek: "Terça", Title: "2ª Chamada Avaliação Integrada (Internato)", Category: "Medicina", Month: 12, IsSHE: false},
		{Date: "02/12/2026", DayOfWeek: "Quarta", Title: "Prazo Final para Divulgação de Notas N2 no Canvas", Category: "Docentes", Month: 12, IsSHE: true},
		{Date: "05/12/2026", DayOfWeek: "Sábado", Title: "Reintegradora Internato - Medicina", Category: "Medicina", Month: 12, IsSHE: false},
		{Date: "05/12/2026", DayOfWeek: "Sábado", Title: "Sábado Letivo - Medicina", Category: "Medicina", Month: 12, IsSHE: false},
		{Date: "07 a 09/12/2026", DayOfWeek: "Seg - Qua", Title: "Período de Exames Finais (SHE e Geral)", Category: "Exames Finais", Month: 12, IsSHE: true},
		{Date: "07 a 09/12/2026", DayOfWeek: "Seg - Qua", Title: "Exame Final das Disciplinas ONA’s e ONS’s", Category: "Online", Month: 12, IsSHE: true},
		{Date: "11/12/2026", DayOfWeek: "Sexta", Title: "Entrega dos Diários e Fechamento Oficial do Semestre 2026.2", Category: "Docentes", Month: 12, IsSHE: true},
		{Date: "17/12/2026", DayOfWeek: "Quinta", Title: "Término do Internato - Medicina", Category: "Medicina", Month: 12, IsSHE: false},
	}
}

// GetOfficialSaturdayClasses2026_2 retorna os sábados de compensação letiva de SHE
func GetOfficialSaturdayClasses2026_2() []SaturdayClass {
	return []SaturdayClass{
		{Date: "15/08/2026", ReplacesWeekday: "Terça-feira", Description: "Sábado Letivo com horário de Terça-feira (Medicina e SHE)"},
		{Date: "22/08/2026", ReplacesWeekday: "Quarta-feira", Description: "Sábado Letivo com horário de Quarta-feira (Medicina e SHE)"},
		{Date: "29/08/2026", ReplacesWeekday: "Segunda-feira", Description: "Sábado Letivo com horário de Segunda-feira (Medicina e SHE)"},
		{Date: "12/09/2026", ReplacesWeekday: "Sexta-feira", Description: "Sábado Letivo com horário de Sexta-feira (Medicina e SHE)"},
		{Date: "26/09/2026", ReplacesWeekday: "Segunda-feira", Description: "Sábado Letivo com horário de Segunda-feira (Medicina e SHE)"},
		{Date: "03/10/2026", ReplacesWeekday: "Quinta-feira", Description: "Sábado Letivo com horário de Quinta-feira (Medicina e SHE)"},
		{Date: "31/10/2026", ReplacesWeekday: "Terça-feira", Description: "Sábado Letivo com horário de Terça-feira (Medicina e SHE)"},
		{Date: "07/11/2026", ReplacesWeekday: "Sexta-feira", Description: "Sábado Letivo com horário de Sexta-feira (Medicina e SHE)"},
		{Date: "14/11/2026", ReplacesWeekday: "Segunda-feira", Description: "Sábado Letivo com horário de Segunda-feira (Medicina e SHE)"},
	}
}

// GetAcademicCalendar implementa a ferramenta MCP para consulta estruturada do calendário
func (c *CanvasClient) GetAcademicCalendar(term string, category string, month int, search string, onlySHE bool) (*AcademicCalendarResult, error) {
	if term == "" {
		term = "2026.2"
	}

	allEvents := GetOfficialCalendar2026_2()
	saturdays := GetOfficialSaturdayClasses2026_2()

	var filtered []AcademicEvent
	catLower := strings.ToLower(strings.TrimSpace(category))
	searchLower := strings.ToLower(strings.TrimSpace(search))

	for _, ev := range allEvents {
		// Filtro por mês
		if month > 0 && ev.Month != month {
			continue
		}

		// Filtro por SHE
		if onlySHE && !ev.IsSHE {
			continue
		}

		// Filtro por categoria
		if catLower != "" && catLower != "all" && catLower != "todas" {
			matchCat := false
			evCatLower := strings.ToLower(ev.Category)
			switch catLower {
			case "she":
				matchCat = ev.IsSHE
			case "n1":
				matchCat = strings.Contains(strings.ToLower(ev.Title), "n1")
			case "n2":
				matchCat = strings.Contains(strings.ToLower(ev.Title), "n2")
			case "feriado", "feriados":
				matchCat = evCatLower == "feriado" || evCatLower == "recesso"
			case "sabado", "sabados", "sabado_letivo":
				matchCat = strings.Contains(evCatLower, "sábado")
			case "exames", "finais":
				matchCat = strings.Contains(strings.ToLower(ev.Title), "exame")
			default:
				matchCat = strings.Contains(evCatLower, catLower)
			}
			if !matchCat {
				continue
			}
		}

		// Filtro por busca de texto
		if searchLower != "" {
			inTitle := strings.Contains(strings.ToLower(ev.Title), searchLower)
			inDate := strings.Contains(strings.ToLower(ev.Date), searchLower)
			inCat := strings.Contains(strings.ToLower(ev.Category), searchLower)
			if !inTitle && !inDate && !inCat {
				continue
			}
		}

		filtered = append(filtered, ev)
	}

	// Gera tabela Markdown limpa
	var md strings.Builder
	md.WriteString(fmt.Sprintf("### 📅 Calendário Acadêmico Institucional — Semestre %s\n", term))
	md.WriteString("**Centro Universitário São Lucas Ji-Paraná • Afya Educacional**\n\n")

	if len(filtered) == 0 {
		md.WriteString("_Nenhum evento encontrado para os filtros informados._\n")
	} else {
		md.WriteString("| Data | Dia | Evento / Atividade | Categoria |\n")
		md.WriteString("| :---: | :---: | :--- | :--- |\n")
		for _, ev := range filtered {
			icon := "📌"
			if ev.Category == "Feriado" || ev.Category == "Recesso" {
				icon = "🏖️"
			} else if strings.Contains(ev.Title, "N1") || strings.Contains(ev.Title, "N2") {
				icon = "📝"
			} else if strings.Contains(ev.Title, "Exame") {
				icon = "🎯"
			} else if strings.Contains(ev.Category, "Sábado") {
				icon = "🗓️"
			}
			md.WriteString(fmt.Sprintf("| **%s** | %s | %s %s | %s |\n", ev.Date, ev.DayOfWeek, icon, ev.Title, ev.Category))
		}
	}

	keyDeadlines := KeyDeadlines{
		StartSemester:    "27/07/2026",
		N1EvaluationsSHE: "21/09/2026 a 26/09/2026",
		N1GradesDeadline: "05/10/2026",
		N2EvaluationsSHE: "23/11/2026 a 28/11/2026",
		SecondChanceSHE:  "30/11/2026 a 02/12/2026",
		N2GradesDeadline: "02/12/2026",
		FinalExams:       "07/12/2026 a 09/12/2026",
		SemesterClosing:  "11/12/2026",
	}

	return &AcademicCalendarResult{
		Term:            term,
		Institution:     "Centro Universitário São Lucas Ji-Paraná • Afya Educacional",
		KeyDeadlines:    keyDeadlines,
		SaturdayClasses: saturdays,
		TotalEvents:     len(allEvents),
		FilteredEvents:  filtered,
		MarkdownTable:   md.String(),
	}, nil
}
