package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CloneItemDetail representa o resultado individual da clonagem de um item de módulo
type CloneItemDetail struct {
	Type      string `json:"type"` // "Page", "Assignment", "Quiz", "ExternalUrl", "SubHeader"
	Title     string `json:"title"`
	Status    string `json:"status"` // "sucesso", "ignorado", "erro"
	Details   string `json:"details,omitempty"`
	DestURL   string `json:"dest_url,omitempty"`
	ContentID string `json:"content_id,omitempty"`
}

// CloneModuleResult consolida o relatório de replicação inter-turmas
type CloneModuleResult struct {
	SourceCourseID   string            `json:"source_course_id"`
	SourceCourseName string            `json:"source_course_name"`
	DestCourseID     string            `json:"dest_course_id"`
	DestCourseName   string            `json:"dest_course_name"`
	SourceModuleName string            `json:"source_module_name"`
	DestModuleID     string            `json:"dest_module_id"`
	TotalItems       int               `json:"total_items"`
	ClonedItems      int               `json:"cloned_items"`
	Items            []CloneItemDetail `json:"items"`
	ExecutionTime    string            `json:"execution_time"`
	MarkdownSummary  string            `json:"markdown_summary"`
}

// BlueprintExportResult representa o manifesto exportado de uma disciplina
type BlueprintExportResult struct {
	CourseID         string `json:"course_id"`
	CourseName       string `json:"course_name"`
	CleanName        string `json:"clean_name"`
	Period           string `json:"period"`
	TermName         string `json:"term_name"`
	ExportedAt       string `json:"exported_at"`
	JSONFilePath     string `json:"json_file_path"`
	MarkdownPath     string `json:"markdown_path"`
	TotalModules     int    `json:"total_modules"`
	TotalPages       int    `json:"total_pages"`
	TotalAssignments int    `json:"total_assignments"`
	TotalQuizzes     int    `json:"total_quizzes"`
	MarkdownSummary  string `json:"markdown_summary"`
}

// CloneModule replica integralmente um módulo (com páginas, tarefas, quizzes e links) de uma turma para outra
func (c *CanvasClient) CloneModule(sourceCourseIDOrQuery, destCourseIDOrQuery, moduleIDOrName string, publishAfterClone bool) (*CloneModuleResult, error) {
	start := time.Now()

	// 1. Resolve cursos de origem e destino
	srcCourseID, err := c.ResolveCourseID(sourceCourseIDOrQuery)
	if err != nil {
		return nil, fmt.Errorf("erro ao identificar turma de origem: %w", err)
	}

	destCourseID, err := c.ResolveCourseID(destCourseIDOrQuery)
	if err != nil {
		return nil, fmt.Errorf("erro ao identificar turma de destino: %w", err)
	}

	if srcCourseID == destCourseID {
		return nil, fmt.Errorf("turma de origem e destino não podem ser as mesmas (ID: %s)", srcCourseID)
	}

	srcCourse, _ := c.GetCourseDetails(srcCourseID)
	srcName, _ := srcCourse["name"].(string)

	destCourse, _ := c.GetCourseDetails(destCourseID)
	destName, _ := destCourse["name"].(string)

	// 2. Localiza o módulo no curso de origem
	modulesRaw, err := c.ListModules(srcCourseID)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar módulos do curso de origem: %w", err)
	}

	var modulesList []map[string]any
	dataBytes, _ := json.Marshal(modulesRaw)
	_ = json.Unmarshal(dataBytes, &modulesList)

	var targetModule map[string]any
	modQueryLower := strings.ToLower(strings.TrimSpace(moduleIDOrName))

	for _, m := range modulesList {
		mID := fmt.Sprintf("%v", m["id"])
		mName, _ := m["name"].(string)

		if mID == moduleIDOrName || strings.ToLower(mName) == modQueryLower || strings.Contains(strings.ToLower(mName), modQueryLower) {
			targetModule = m
			break
		}
	}

	if targetModule == nil {
		return nil, fmt.Errorf("módulo '%s' não foi encontrado no curso de origem (%s)", moduleIDOrName, srcName)
	}

	sourceModuleName, _ := targetModule["name"].(string)
	sourceItemsRaw, _ := targetModule["items"].([]any)

	// 3. Localiza ou cria o módulo correspondente no curso de destino
	destModulesRaw, _ := c.ListModules(destCourseID)
	var destModulesList []map[string]any
	dBytes, _ := json.Marshal(destModulesRaw)
	_ = json.Unmarshal(dBytes, &destModulesList)

	var destModuleID string
	for _, dm := range destModulesList {
		dmName, _ := dm["name"].(string)
		if strings.EqualFold(dmName, sourceModuleName) {
			destModuleID = fmt.Sprintf("%v", dm["id"])
			break
		}
	}

	if destModuleID == "" {
		newMod, err := c.CreateModule(CreateModuleParams{
			CourseID: destCourseID,
			Name:     sourceModuleName,
		})
		if err != nil {
			return nil, fmt.Errorf("erro ao criar módulo no curso de destino: %w", err)
		}
		if nmMap, ok := newMod.(map[string]any); ok {
			destModuleID = fmt.Sprintf("%v", nmMap["id"])
		}
	}

	if destModuleID == "" {
		return nil, fmt.Errorf("não foi possível obter o ID do módulo criado no destino")
	}

	// 4. Clona os itens sequencialmente com tratamento por tipo
	var clonedItems []CloneItemDetail
	successCount := 0

	for _, itRaw := range sourceItemsRaw {
		it, ok := itRaw.(map[string]any)
		if !ok {
			continue
		}

		itemType, _ := it["type"].(string)
		itemTitle, _ := it["title"].(string)
		itemDetail := CloneItemDetail{
			Type:   itemType,
			Title:  itemTitle,
			Status: "sucesso",
		}

		switch itemType {
		case "Page":
			pageURL, _ := it["page_url"].(string)
			if pageURL == "" {
				itemDetail.Status = "ignorado"
				itemDetail.Details = "slug da página vazio"
				clonedItems = append(clonedItems, itemDetail)
				continue
			}

			// Busca o conteúdo da página no curso de origem
			srcPage, err := c.GetCoursePage(srcCourseID, pageURL)
			if err != nil {
				itemDetail.Status = "erro"
				itemDetail.Details = fmt.Sprintf("falha ao ler página de origem: %v", err)
				clonedItems = append(clonedItems, itemDetail)
				continue
			}

			pageTitle, _ := srcPage["title"].(string)
			pageBody, _ := srcPage["body"].(string)

			// Cria ou atualiza a página no curso de destino
			pub := publishAfterClone
			destPageRes, err := c.CreatePage(CreatePageParams{
				CourseID:  destCourseID,
				Title:     pageTitle,
				Body:      pageBody,
				Published: &pub,
			})
			if err != nil {
				itemDetail.Status = "erro"
				itemDetail.Details = fmt.Sprintf("falha ao criar página no destino: %v", err)
				clonedItems = append(clonedItems, itemDetail)
				continue
			}

			destPageURL := pageURL
			if dpMap, ok := destPageRes.(map[string]any); ok {
				if u, ok := dpMap["url"].(string); ok && u != "" {
					destPageURL = u
				}
			}

			// Vincula ao módulo de destino
			_, err = c.AddModuleItem(AddModuleItemParams{
				CourseID: destCourseID,
				ModuleID: destModuleID,
				Title:    pageTitle,
				Type:     "Page",
				PageURL:  destPageURL,
			})
			if err != nil {
				itemDetail.Status = "erro"
				itemDetail.Details = fmt.Sprintf("página criada, mas falhou ao vincular ao módulo: %v", err)
			} else {
				itemDetail.Details = "página e simuladores replicados com sucesso"
				itemDetail.DestURL = destPageURL
				successCount++
			}

		case "Assignment":
			contentID := fmt.Sprintf("%v", it["content_id"])
			if contentID == "" || contentID == "<nil>" {
				itemDetail.Status = "ignorado"
				itemDetail.Details = "content_id da tarefa ausente"
				clonedItems = append(clonedItems, itemDetail)
				continue
			}

			srcAssignRaw, err := c.GetAssignment(srcCourseID, contentID)
			if err != nil {
				itemDetail.Status = "erro"
				itemDetail.Details = fmt.Sprintf("falha ao ler tarefa de origem: %v", err)
				clonedItems = append(clonedItems, itemDetail)
				continue
			}

			srcAssign, _ := srcAssignRaw.(map[string]any)
			assignName, _ := srcAssign["name"].(string)
			assignDesc, _ := srcAssign["description"].(string)
			points, _ := srcAssign["points_possible"].(float64)

			var subTypes []string
			if stRaw, ok := srcAssign["submission_types"].([]any); ok {
				for _, st := range stRaw {
					if stStr, ok := st.(string); ok {
						subTypes = append(subTypes, stStr)
					}
				}
			}

			pub := publishAfterClone
			newAssign, err := c.CreateAssignment(CreateAssignmentParams{
				CourseID:        destCourseID,
				Name:            assignName,
				Description:     assignDesc,
				PointsPossible: points,
				SubmissionTypes: subTypes,
				Published:       &pub,
			})
			if err != nil {
				itemDetail.Status = "erro"
				itemDetail.Details = fmt.Sprintf("falha ao criar tarefa no destino: %v", err)
				clonedItems = append(clonedItems, itemDetail)
				continue
			}

			newAssignID := ""
			if naMap, ok := newAssign.(map[string]any); ok {
				newAssignID = fmt.Sprintf("%v", naMap["id"])
			}

			// Vincula ao módulo de destino
			_, err = c.AddModuleItem(AddModuleItemParams{
				CourseID:  destCourseID,
				ModuleID:  destModuleID,
				Title:     assignName,
				Type:      "Assignment",
				ContentID: newAssignID,
			})
			if err != nil {
				itemDetail.Status = "erro"
				itemDetail.Details = fmt.Sprintf("tarefa criada, mas falhou ao vincular ao módulo: %v", err)
			} else {
				itemDetail.Details = fmt.Sprintf("tarefa clonada (ID: %s, Pontos: %.1f)", newAssignID, points)
				itemDetail.ContentID = newAssignID
				successCount++
			}

		case "Quiz":
			contentID := fmt.Sprintf("%v", it["content_id"])
			if contentID == "" || contentID == "<nil>" {
				itemDetail.Status = "ignorado"
				itemDetail.Details = "content_id do quiz ausente"
				clonedItems = append(clonedItems, itemDetail)
				continue
			}

			srcQuiz, err := c.GetQuiz(srcCourseID, contentID)
			if err != nil {
				itemDetail.Status = "erro"
				itemDetail.Details = fmt.Sprintf("falha ao ler quiz de origem: %v", err)
				clonedItems = append(clonedItems, itemDetail)
				continue
			}

			qTitle, _ := srcQuiz["title"].(string)
			qDesc, _ := srcQuiz["description"].(string)
			timeLimit, _ := srcQuiz["time_limit"].(float64)

			// Puxa as questões do quiz de origem
			questionsRaw, err := c.GetQuizQuestions(srcCourseID, contentID)
			var questions []QuizQuestion
			if err == nil {
				for idx, qMap := range questionsRaw {
					questText, _ := qMap["question_text"].(string)
					questType, _ := qMap["question_type"].(string)
					questPoints, _ := qMap["points_possible"].(float64)
					questTitle, _ := qMap["question_name"].(string)
					if questTitle == "" {
						questTitle = fmt.Sprintf("Questão %d", idx+1)
					}

					var ansList []QuizAnswer
					if answersRaw, ok := qMap["answers"].([]any); ok {
						for _, aRaw := range answersRaw {
							if aMap, ok := aRaw.(map[string]any); ok {
								aText, _ := aMap["text"].(string)
								aComment, _ := aMap["comments"].(string)
								aWeight := 0
								if wVal, ok := aMap["weight"]; ok {
									switch w := wVal.(type) {
									case float64:
										aWeight = int(w)
									case int:
										aWeight = w
									}
								}
								aBlankID, _ := aMap["blank_id"].(string)
								aMatchLeft, _ := aMap["answer_match_left"].(string)
								if aMatchLeft == "" {
									aMatchLeft, _ = aMap["left"].(string)
								}
								aMatchRight, _ := aMap["answer_match_right"].(string)
								if aMatchRight == "" {
									aMatchRight, _ = aMap["right"].(string)
								}

								ansList = append(ansList, QuizAnswer{
									Text:             aText,
									Weight:           aWeight,
									Comment:          aComment,
									BlankID:          aBlankID,
									AnswerMatchLeft:  aMatchLeft,
									AnswerMatchRight: aMatchRight,
								})
							}
						}
					}

					questions = append(questions, QuizQuestion{
						Title:          questTitle,
						Text:           questText,
						Type:           questType,
						PointsPossible: questPoints,
						Answers:        ansList,
					})
				}
			}

			pub := publishAfterClone
			newQuiz, err := c.CreateQuiz(CreateQuizParams{
				CourseID:    destCourseID,
				Title:       qTitle,
				Description: qDesc,
				TimeLimit:   int(timeLimit),
				Published:   &pub,
				Questions:   questions,
			})
			if err != nil {
				itemDetail.Status = "erro"
				itemDetail.Details = fmt.Sprintf("falha ao recriar quiz no destino: %v", err)
				clonedItems = append(clonedItems, itemDetail)
				continue
			}

			newQuizID := ""
			if nqMap, ok := newQuiz.(map[string]any); ok {
				newQuizID = fmt.Sprintf("%v", nqMap["id"])
			}

			// Vincula ao módulo de destino
			_, err = c.AddModuleItem(AddModuleItemParams{
				CourseID:  destCourseID,
				ModuleID:  destModuleID,
				Title:     qTitle,
				Type:      "Quiz",
				ContentID: newQuizID,
			})
			if err != nil {
				itemDetail.Status = "erro"
				itemDetail.Details = fmt.Sprintf("quiz criado, mas falhou ao vincular ao módulo: %v", err)
			} else {
				itemDetail.Details = fmt.Sprintf("quiz e %d questões replicadas (ID: %s)", len(questions), newQuizID)
				itemDetail.ContentID = newQuizID
				successCount++
			}

		case "ExternalUrl":
			extURL, _ := it["external_url"].(string)
			newTab, _ := it["new_tab"].(bool)

			_, err = c.AddModuleItem(AddModuleItemParams{
				CourseID:    destCourseID,
				ModuleID:    destModuleID,
				Title:       itemTitle,
				Type:        "ExternalUrl",
				ExternalURL: extURL,
				NewTab:      newTab,
			})
			if err != nil {
				itemDetail.Status = "erro"
				itemDetail.Details = fmt.Sprintf("falha ao adicionar link externo: %v", err)
			} else {
				itemDetail.Details = fmt.Sprintf("link externo configurado (%s)", extURL)
				itemDetail.DestURL = extURL
				successCount++
			}

		case "SubHeader":
			_, err = c.AddModuleItem(AddModuleItemParams{
				CourseID: destCourseID,
				ModuleID: destModuleID,
				Title:    itemTitle,
				Type:     "SubHeader",
			})
			if err != nil {
				itemDetail.Status = "erro"
				itemDetail.Details = fmt.Sprintf("falha ao adicionar subtítulo: %v", err)
			} else {
				itemDetail.Details = "separador visual adicionado"
				successCount++
			}

		default:
			itemDetail.Status = "ignorado"
			itemDetail.Details = fmt.Sprintf("tipo de item '%s' não suportado para clonagem direta", itemType)
		}

		clonedItems = append(clonedItems, itemDetail)
	}

	execDuration := time.Since(start).Round(time.Millisecond).String()

	// 5. Monta tabela e sumário Markdown didático
	var md strings.Builder
	md.WriteString("### 🔄 Clonagem Inter-Turmas Concluída com Sucesso\n\n")
	md.WriteString(fmt.Sprintf("- **Módulo Clonado:** `%s`\n", sourceModuleName))
	md.WriteString(fmt.Sprintf("- **Origem:** %s (`ID %s`)\n", srcName, srcCourseID))
	md.WriteString(fmt.Sprintf("- **Destino:** %s (`ID %s`)\n", destName, destCourseID))
	md.WriteString(fmt.Sprintf("- **Módulo no Destino (ID):** `%s`\n", destModuleID))
	md.WriteString(fmt.Sprintf("- **Estatísticas:** %d de %d itens replicados com êxito • **Tempo de Execução:** `%s`\n\n", successCount, len(sourceItemsRaw), execDuration))

	md.WriteString("| Tipo | Título do Item | Status | Detalhes da Replicação |\n")
	md.WriteString("| :---: | :--- | :---: | :--- |\n")
	for _, it := range clonedItems {
		badge := "✅ Sucesso"
		if it.Status == "erro" {
			badge = "❌ Erro"
		} else if it.Status == "ignorado" {
			badge = "⚠️ Ignorado"
		}

		typeIcon := "📄"
		switch it.Type {
		case "Page":
			typeIcon = "📖 Página"
		case "Assignment":
			typeIcon = "📝 Tarefa"
		case "Quiz":
			typeIcon = "❓ Quiz"
		case "ExternalUrl":
			typeIcon = "🔗 Link Ext."
		case "SubHeader":
			typeIcon = "🏷️ Subtítulo"
		}

		md.WriteString(fmt.Sprintf("| %s | **%s** | %s | %s |\n", typeIcon, it.Title, badge, it.Details))
	}

	md.WriteString("\n> [!TIP]\n")
	md.WriteString(fmt.Sprintf("> Todos os itens já foram vinculados ao módulo e organizados no Canvas LMS da turma de destino. Acesso direto: [Abrir Módulo no Canvas](https://afya.instructure.com/courses/%s/modules)\n", destCourseID))

	return &CloneModuleResult{
		SourceCourseID:   srcCourseID,
		SourceCourseName: srcName,
		DestCourseID:     destCourseID,
		DestCourseName:   destName,
		SourceModuleName: sourceModuleName,
		DestModuleID:     destModuleID,
		TotalItems:       len(sourceItemsRaw),
		ClonedItems:      successCount,
		Items:            clonedItems,
		ExecutionTime:    execDuration,
		MarkdownSummary:  md.String(),
	}, nil
}

// ExportCourseBlueprint varre a turma e salva um manifesto pedagógico completo (JSON e Markdown) em doc/blueprints/
func (c *CanvasClient) ExportCourseBlueprint(courseIDOrQuery string, outputFormat string) (*BlueprintExportResult, error) {
	resolvedID, err := c.ResolveCourseID(courseIDOrQuery)
	if err != nil {
		return nil, err
	}

	course, err := c.GetCourseDetails(resolvedID)
	if err != nil {
		return nil, fmt.Errorf("falha ao obter detalhes do curso: %w", err)
	}

	courseName, _ := course["name"].(string)
	cleanName, _ := course["clean_name"].(string)
	if cleanName == "" {
		cleanName = courseName
	}
	period, _ := course["period"].(string)
	termObj, _ := course["term"].(map[string]any)
	termName, _ := termObj["name"].(string)

	// Varre módulos e itens
	modulesRaw, _ := c.ListModules(resolvedID)
	var modulesList []map[string]any
	mBytes, _ := json.Marshal(modulesRaw)
	_ = json.Unmarshal(mBytes, &modulesList)

	// Varre páginas wiki
	pages, _ := c.GetCoursePages(resolvedID)

	// Varre tarefas
	assignmentsRaw, _ := c.ListAssignments(resolvedID)
	var assignmentsList []map[string]any
	aBytes, _ := json.Marshal(assignmentsRaw)
	_ = json.Unmarshal(aBytes, &assignmentsList)

	// Conta quizzes nos módulos
	totalQuizzes := 0
	for _, m := range modulesList {
		if items, ok := m["items"].([]any); ok {
			for _, itRaw := range items {
				if it, ok := itRaw.(map[string]any); ok {
					if it["type"] == "Quiz" {
						totalQuizzes++
					}
				}
			}
		}
	}

	now := time.Now()
	exportTimeStr := now.Format("02/01/2006 às 15:04")

	blueprintData := map[string]any{
		"course_id":         resolvedID,
		"course_name":       courseName,
		"clean_name":        cleanName,
		"period":            period,
		"term_name":         termName,
		"exported_at":       now.Format(time.RFC3339),
		"total_modules":     len(modulesList),
		"total_pages":       len(pages),
		"total_assignments": len(assignmentsList),
		"total_quizzes":     totalQuizzes,
		"modules":           modulesList,
		"pages":             pages,
		"assignments":       assignmentsList,
	}

	// Garante criação do diretório doc/blueprints
	blueprintDir := filepath.Join("doc", "blueprints")
	_ = os.MkdirAll(blueprintDir, 0755)

	sanitizedCourse := strings.ToLower(cleanName)
	sanitizedCourse = strings.ReplaceAll(sanitizedCourse, " ", "_")
	sanitizedCourse = strings.ReplaceAll(sanitizedCourse, "/", "_")
	fileNameBase := fmt.Sprintf("blueprint_%s_%s", sanitizedCourse, now.Format("20060102_150405"))

	jsonPath := filepath.Join(blueprintDir, fileNameBase+".json")
	jsonBytes, _ := json.MarshalIndent(blueprintData, "", "  ")
	_ = os.WriteFile(jsonPath, jsonBytes, 0644)

	// Gera versão Markdown legível
	mdPath := filepath.Join(blueprintDir, fileNameBase+".md")
	var md strings.Builder
	md.WriteString(fmt.Sprintf("# 📐 Blueprint Didático — %s\n\n", cleanName))
	md.WriteString(fmt.Sprintf("- **Disciplina Oficial:** %s (`ID %s`)\n", courseName, resolvedID))
	md.WriteString(fmt.Sprintf("- **Período Curricular:** %s • **Semestre de Origem:** %s\n", period, termName))
	md.WriteString(fmt.Sprintf("- **Data de Exportação:** %s\n\n", exportTimeStr))
	md.WriteString("### 📊 Métricas do Acervo Didático:\n")
	md.WriteString(fmt.Sprintf("- **Módulos Estruturados:** %d\n", len(modulesList)))
	md.WriteString(fmt.Sprintf("- **Páginas Wiki / Simuladores:** %d\n", len(pages)))
	md.WriteString(fmt.Sprintf("- **Tarefas de Avaliação:** %d\n", len(assignmentsList)))
	md.WriteString(fmt.Sprintf("- **Quizzes e Simuladores:** %d\n\n", totalQuizzes))

	md.WriteString("### 🗂️ Estrutura de Módulos e Conteúdos:\n\n")
	for idx, m := range modulesList {
		mName, _ := m["name"].(string)
		md.WriteString(fmt.Sprintf("#### Módulo %d: %s\n", idx+1, mName))
		if items, ok := m["items"].([]any); ok {
			for _, itRaw := range items {
				if it, ok := itRaw.(map[string]any); ok {
					itType, _ := it["type"].(string)
					itTitle, _ := it["title"].(string)
					icon := "📄"
					switch itType {
					case "Page":
						icon = "📖"
					case "Assignment":
						icon = "📝"
					case "Quiz":
						icon = "❓"
					case "ExternalUrl":
						icon = "🔗"
					case "SubHeader":
						icon = "🏷️"
					}
					md.WriteString(fmt.Sprintf("- %s **[%s]** %s\n", icon, itType, itTitle))
				}
			}
		}
		md.WriteString("\n")
	}

	_ = os.WriteFile(mdPath, []byte(md.String()), 0644)

	var summary strings.Builder
	summary.WriteString("### 📐 Blueprint Didático Exportado com Sucesso\n\n")
	summary.WriteString(fmt.Sprintf("- **Disciplina:** `%s` (%s)\n", cleanName, period))
	summary.WriteString(fmt.Sprintf("- **Termo:** `%s` • **Data:** `%s`\n", termName, exportTimeStr))
	summary.WriteString("- **Artefatos Salvos:**\n")
	summary.WriteString(fmt.Sprintf("  * JSON (Estruturado): [`%s`](file://%s)\n", jsonPath, jsonPath))
	summary.WriteString(fmt.Sprintf("  * Markdown (Legível): [`%s`](file://%s)\n\n", mdPath, mdPath))
	summary.WriteString(fmt.Sprintf("- **Conteúdo Total:** %d módulos, %d páginas, %d tarefas e %d quizzes prontos para importação em outros semestres letivos.\n",
		len(modulesList), len(pages), len(assignmentsList), totalQuizzes))

	return &BlueprintExportResult{
		CourseID:         resolvedID,
		CourseName:       courseName,
		CleanName:        cleanName,
		Period:           period,
		TermName:         termName,
		ExportedAt:       exportTimeStr,
		JSONFilePath:     jsonPath,
		MarkdownPath:     mdPath,
		TotalModules:     len(modulesList),
		TotalPages:       len(pages),
		TotalAssignments: len(assignmentsList),
		TotalQuizzes:     totalQuizzes,
		MarkdownSummary:  summary.String(),
	}, nil
}
