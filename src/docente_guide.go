package main

import (
	"fmt"
	"os"
	"strings"
)

// TeacherGuideSection representa uma seção do Guia do Docente no Canvas
type TeacherGuideSection struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Summary   string   `json:"summary"`
	KeyPoints []string `json:"key_points"`
	Content   string   `json:"content"`
}

// TeacherGuideResult encapsula o retorno estruturado do Guia do Docente
type TeacherGuideResult struct {
	Title           string                `json:"title"`
	Institution     string                `json:"institution"`
	FilteredTopic   string                `json:"filtered_topic,omitempty"`
	SearchQuery     string                `json:"search_query,omitempty"`
	TotalSections   int                   `json:"total_sections"`
	Sections        []TeacherGuideSection `json:"sections"`
	Checklist       []string              `json:"checklist"`
	Troubleshooting map[string]string     `json:"troubleshooting"`
	Markdown        string                `json:"markdown"`
}

// GetDefaultTeacherGuideSections retorna as seções estruturadas oficiais do Guia do Docente
func GetDefaultTeacherGuideSections() []TeacherGuideSection {
	return []TeacherGuideSection{
		{
			ID:      "tipologia",
			Title:   "Visão Geral e Tipologia de Disciplinas",
			Summary: "Autonomia pedagógica e regras de produção de conteúdo por modalidade.",
			KeyPoints: []string{
				"Presenciais (PR): Total autonomia pedagógica para criar, estruturar e publicar módulos e tarefas.",
				"Digitais/Híbridas (ON.A, ON.S, HB.S, HB): Conteúdo previamente padronizado; foco docente em mediação, prazos e correção.",
				"Vínculo Acadêmico: A carga horária e turmas refletem a sincronização com o RM Educacional.",
			},
			Content: "- **Autonomia Pedagógica**:\n  - **Presenciais (PR)**: O professor possui total autonomia para criar, organizar e publicar recursos, módulos e atividades.\n  - **Digitais/Híbridas (ON.A, ON.S, HB.S, HB)**: O conteúdo é previamente produzido e padronizado. O professor deve manter o alinhamento institucional e gerenciar interações, prazos e correções.\n- **Vínculo Acadêmico**: A carga de disciplinas reflete diretamente a integração com o RM Educacional. Pendências de listagem exigem validação com a coordenação/secretaria.",
		},
		{
			ID:      "acesso",
			Title:   "Acesso e Configurações Iniciais",
			Summary: "Credenciais de acesso, fuso horário e preferências de notificação.",
			KeyPoints: []string{
				"Acesso via SSO Microsoft institucional (e-mail Afya) ou credenciais locais (CPF / data de nascimento).",
				"Validar fuso horário (ex.: América/Manaus ou Porto Velho UTC-4 / Brasília UTC-3) e idioma para evitar confusão de prazos.",
				"Ativar notificações imediatas para novas submissões, mensagens da Caixa de Entrada e comentários de avaliação.",
			},
			Content: "- **Acesso**: Via SSO Microsoft institucional (e-mail Afya) ou credenciais locais (CPF no usuário, data de nascimento como senha inicial).\n- **Ajustes Críticos**: Validar fuso horário, idioma e ativar notificações imediatas para submissões, mensagens e comentários.",
		},
		{
			ID:      "modulos",
			Title:   "Organização Estrutural (Módulos, Páginas e Arquivos)",
			Summary: "Regras de publicação individual, página inicial e boas práticas de nomenclatura.",
			KeyPoints: []string{
				"Regra de Ouro da Visibilidade: Publicar um módulo NÃO publica seus itens internos. Páginas, tarefas e testes devem ser publicados individualmente.",
				"Página Inicial: Deve explicitar propósito da disciplina, orientações de navegação, principais prazos e canais de suporte.",
				"Nomenclatura Linear: Estruturação clara em trilha (ex.: Módulos numerados e identificados).",
				"LGPD e Arquivos: Proibido disponibilizar gabaritos prematuros, arquivos obsoletos ou dados pessoais de estudantes.",
			},
			Content: "- **Regra de Visibilidade**: Publicar um módulo não publica seus itens internos. Módulos, páginas, tarefas e testes devem ser publicados individualmente.\n- **Página Inicial**: Deve explicitar propósito da disciplina, orientações de navegação, principais prazos e canais de suporte.\n- **Nomenclatura**: Estruturação linear e clara (ex.: `Unidade 1`, `Semana 2`, `Módulo N: [Título]`).\n- **Arquivos**: Proibido subir gabaritos indevidos, arquivos obsoletos ou dados sensíveis/pessoais (LGPD).",
		},
		{
			ID:      "avaliacoes",
			Title:   "Avaliações, Prazos, Rubricas e SpeedGrader",
			Summary: "Configuração de atividades, SpeedGrader, retenção de notas no boletim e rubricas.",
			KeyPoints: []string{
				"Tarefas e Quizzes: Definir tipo de submissão, grupo de tarefas e prazos claros (data de entrega vs. data de disponibilidade).",
				"Protocolos de Quizzes: Exigem cumprimento de senhas, limite de tentativas e aplicação em laboratório quando formal.",
				"SpeedGrader: Feedbacks pedagógicos e construtivos, apontando erros exatos (evitar monosílabos como 'bom' ou 'incompleto').",
				"Boletim e Política de Postagem: Atividades só aparecem se pontuadas e publicadas. Use 'Política de Postagem Manual' para reter notas até o término da correção de toda a turma.",
			},
			Content: "- **Tarefas e Quizzes**:\n  - Sempre definir tipo de submissão, grupo de tarefas e prazos claros (data de entrega vs. data de disponibilidade).\n  - Quizzes formais exigem cumprimento dos protocolos institucionais (senhas, IP/laboratório, tentativas).\n- **Rubricas**: Vincular critérios de avaliação objetivos com pontuação alinhada ao valor da tarefa.\n- **SpeedGrader**:\n  - Fornecer feedbacks qualitativos e construtivos (evitar monosílabos como 'bom' ou 'incompleto').\n  - Anotações pontuais e uso de rubricas para padronização.\n- **Boletim (Notas) e Política de Postagem**:\n  - Atividades só aparecem no boletim se forem pontuadas e publicadas.\n  - Configurar **Política de Postagem Manual** para reter notas até a revisão completa da turma, evitando liberação prematura.\n  - Não alterar notas arbitrárias manualmente sem amparo dos fluxos institucionais da Afya.",
		},
		{
			ID:      "comunicacao",
			Title:   "Canais Oficiais de Comunicação",
			Summary: "Diferenciação pedagógica entre Avisos Coletivos e Caixa de Entrada individual.",
			KeyPoints: []string{
				"Avisos da Disciplina: Comunicação estritamente coletiva (lembretes de prazos, início de unidades, síntese semanal).",
				"Caixa de Entrada (Inbox): Comunicação direta, privada e personalizada com o estudante.",
			},
			Content: "- **Avisos da Disciplina**: Uso exclusivamente coletivo (início de unidades, lembretes de prazos, síntese da semana).\n- **Caixa de Entrada**: Comunicação direta e individualizada (dúvidas particulares, suporte personalizado).",
		},
	}
}

// GetTeacherChecklist retorna o checklist obrigatório pré-liberação
func GetTeacherChecklist() []string {
	return []string{
		"1. Página inicial contextualizada, funcional e com boas-vindas.",
		"2. Módulos organizados e com status Publicado.",
		"3. Itens internos publicados individualmente (páginas, tarefas, testes, fóruns).",
		"4. Prazos conferidos e refletidos corretamente no Calendário Acadêmico.",
		"5. Políticas de postagem de notas ajustadas (manual durante correção / automática após conclusão).",
		"6. Rubricas vinculadas e com pesos conferidos.",
		"7. Teste de conformidade realizado via 'Visualização do Estudante' (validando links, visibilidade de abas e experiência do usuário).",
	}
}

// GetTeacherTroubleshooting retorna as soluções para os problemas operacionais mais comuns
func GetTeacherTroubleshooting() map[string]string {
	return map[string]string{
		"modulo_invisivel":  "Módulo publicado, mas invisível ao aluno: Cheque se os itens internos estão publicados individualmente, se há data futura de liberação ou pré-requisitos bloqueantes configurados no módulo.",
		"atividade_sumiu":   "Atividade sumiu do boletim: Verifique se possui pontuação atribuída (> 0) e se está com o status Publicado.",
		"ocultar_notas":     "Ocultar notas antes do término da correção: Ative a Política de Postagem Manual na coluna da atividade no Boletim para reter notas até a liberação oficial.",
		"fuso_horario":      "Prazos com horário divergente: Verifique as configurações de fuso horário da conta e do curso (Rondônia segue UTC-4, Brasília UTC-3).",
		"itens_orfaos":      "Páginas ou quizzes órfãos: Todos os itens avaliativos devem ser organizados na trilha de Módulos para garantir rastreabilidade pedagógica.",
	}
}

// GetTeacherGuide consulta o Guia do Docente da Afya no Canvas
func GetTeacherGuide(topic string, search string) (*TeacherGuideResult, error) {
	// 1. Tenta carregar do arquivo físico em doc/institucional/GUIA_DO_DOCENTE_NO_CANVAS.md se existir
	filePath := "doc/institucional/GUIA_DO_DOCENTE_NO_CANVAS.md"
	var rawFileContent string
	if data, err := os.ReadFile(filePath); err == nil {
		rawFileContent = string(data)
	}

	allSections := GetDefaultTeacherGuideSections()
	checklist := GetTeacherChecklist()
	troubleshooting := GetTeacherTroubleshooting()

	var filtered []TeacherGuideSection
	topicLower := strings.ToLower(strings.TrimSpace(topic))
	searchLower := strings.ToLower(strings.TrimSpace(search))

	for _, sec := range allSections {
		// Filtro por tópico
		if topicLower != "" && topicLower != "all" && topicLower != "todos" {
			matchTopic := false
			switch topicLower {
			case "tipologia", "autonomia", "modalidades":
				matchTopic = sec.ID == "tipologia"
			case "acesso", "configuracao", "configuracoes":
				matchTopic = sec.ID == "acesso"
			case "modulos", "paginas", "arquivos", "estrutura":
				matchTopic = sec.ID == "modulos"
			case "avaliacoes", "speedgrader", "notas", "rubricas", "boletim":
				matchTopic = sec.ID == "avaliacoes"
			case "comunicacao", "avisos", "inbox":
				matchTopic = sec.ID == "comunicacao"
			default:
				if strings.Contains(sec.ID, topicLower) || strings.Contains(strings.ToLower(sec.Title), topicLower) {
					matchTopic = true
				}
			}
			if !matchTopic {
				continue
			}
		}

		// Filtro por busca textual
		if searchLower != "" {
			combined := strings.ToLower(sec.Title + " " + sec.Summary + " " + sec.Content + " " + strings.Join(sec.KeyPoints, " "))
			if !strings.Contains(combined, searchLower) {
				continue
			}
		}

		filtered = append(filtered, sec)
	}

	// Monta visualização em Markdown elegante
	var sb strings.Builder
	sb.WriteString("### 📘 Guia Oficial do Docente no Canvas LMS • Afya\n\n")

	if topicLower == "checklist" {
		sb.WriteString("#### ✅ Checklist Pré-Liberação (Validação Operacional)\n\n")
		for _, item := range checklist {
			sb.WriteString(fmt.Sprintf("- %s\n", item))
		}
	} else if topicLower == "troubleshooting" {
		sb.WriteString("#### 🛠️ Troubleshooting / Resolução de Problemas Frequentes\n\n")
		for _, v := range troubleshooting {
			sb.WriteString(fmt.Sprintf("• %s\n\n", v))
		}
	} else {
		for _, sec := range filtered {
			sb.WriteString(fmt.Sprintf("#### 📌 %s\n", sec.Title))
			sb.WriteString(fmt.Sprintf("_%s_\n\n", sec.Summary))
			for _, kp := range sec.KeyPoints {
				sb.WriteString(fmt.Sprintf("• %s\n", kp))
			}
			sb.WriteString("\n")
		}

		if topicLower == "" || topicLower == "all" || topicLower == "todos" {
			sb.WriteString("---\n\n#### ✅ Checklist Pré-Liberação\n")
			for _, item := range checklist {
				sb.WriteString(fmt.Sprintf("- %s\n", item))
			}
			sb.WriteString("\n---\n\n#### 🛠️ Problemas Frequentes & Soluções Rápidas\n")
			for _, v := range troubleshooting {
				sb.WriteString(fmt.Sprintf("• %s\n", v))
			}
		}
	}

	// Se o usuário procurou e não achou nas seções, mas achou no arquivo bruto, adiciona
	if len(filtered) == 0 && searchLower != "" && strings.Contains(strings.ToLower(rawFileContent), searchLower) {
		sb.WriteString(fmt.Sprintf("\n_Nota: O termo '%s' foi localizado no documento oficial completo:_\n\n", search))
		sb.WriteString(rawFileContent)
	}

	return &TeacherGuideResult{
		Title:           "Guia do Docente no Canvas LMS (Afya)",
		Institution:     "Afya Educação / Centro Universitário São Lucas",
		FilteredTopic:   topic,
		SearchQuery:     search,
		TotalSections:   len(filtered),
		Sections:        filtered,
		Checklist:       checklist,
		Troubleshooting: troubleshooting,
		Markdown:        sb.String(),
	}, nil
}
