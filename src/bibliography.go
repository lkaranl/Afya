package main

import (
	"fmt"
	"html"
	"net/url"
	"regexp"
	"strings"
)

// BookBibliography representa uma obra cadastrada dinamicamente na ementa da disciplina
type BookBibliography struct {
	ID          string   `json:"id"`
	Type        string   `json:"type"` // "basica" ou "complementar"
	Authors     string   `json:"authors"`
	Title       string   `json:"title"`
	Edition     string   `json:"edition,omitempty"`
	Publisher   string   `json:"publisher,omitempty"`
	Year        string   `json:"year,omitempty"`
	ISBN        string   `json:"isbn,omitempty"`
	URL         string   `json:"url"`
	Citation    string   `json:"citation"`
	Topics      []string `json:"topics"`
	KeyChapters string   `json:"key_chapters,omitempty"`
}

// TopicReading representa uma indicação de leitura contextualizada para um tema
type TopicReading struct {
	Topic           string           `json:"topic"`
	RecommendedBook BookBibliography `json:"book"`
	ReadingGoal     string           `json:"reading_goal"`
}

// BibliographyResult encapsula o retorno dinâmico de bibliografia da disciplina
type BibliographyResult struct {
	CourseID        string             `json:"course_id"`
	CourseName      string             `json:"course_name"`
	CleanName       string             `json:"clean_name"`
	Period          string             `json:"period"`
	TermName        string             `json:"term_name"`
	Institution     string             `json:"institution"`
	LibraryName     string             `json:"library_name"`
	LibraryLTIURL   string             `json:"library_lti_url"`
	GlobalSearchURL string             `json:"global_search_url"`
	SyllabusSource  string             `json:"syllabus_source"`
	SyllabusText    string             `json:"syllabus_text,omitempty"`
	TotalBooks      int                `json:"total_books"`
	Books           []BookBibliography `json:"books"`
	TopicReadings   []TopicReading     `json:"topic_readings,omitempty"`
	MarkdownTable   string             `json:"markdown_table"`
}

var (
	rxStrongTag   = regexp.MustCompile(`(?is)<strong[^>]*>(.*?)</strong>|<b[^>]*>(.*?)</b>`)
	rxEmTag       = regexp.MustCompile(`(?is)<em[^>]*>(.*?)</em>|<i[^>]*>(.*?)</i>`)
	rxAnchorTag   = regexp.MustCompile(`(?is)<a\s+[^>]*href=["'](https?://[^"']+)["'][^>]*>`)
	rxISBN        = regexp.MustCompile(`(?i)(?:ISBN(?:-1[03])?:?\s*)([0-9Xx\-]{10,17})`)
	rxLiTag       = regexp.MustCompile(`(?is)<li[^>]*>(.*?)</li>`)
	rxPTag        = regexp.MustCompile(`(?is)<p[^>]*>(.*?)</p>`)
	rxBasicSec    = regexp.MustCompile(`(?is)<h[1-6][^>]*>[^<]*b[áa]sica[^<]*</h[1-6]>`)
	rxComplSec    = regexp.MustCompile(`(?is)<h[1-6][^>]*>[^<]*complementar[^<]*</h[1-6]>`)
	rxSyllabusSec = regexp.MustCompile(`(?is)<h[1-6][^>]*>[^<]*ementa[^<]*</h[1-6]>\s*(?:<div[^>]*>)?\s*<p[^>]*>(.*?)</p>`)
)

// ExtractSyllabusText extrai o texto do parágrafo de ementa da página do Canvas
func ExtractSyllabusText(htmlContent string) string {
	if m := rxSyllabusSec.FindStringSubmatch(htmlContent); len(m) > 1 {
		return CleanCanvasHTML(m[1])
	}
	return ""
}

// ExtractBooksFromHTML faz o parsing dinâmico de blocos HTML de ementa (básica e complementar)
func ExtractBooksFromHTML(htmlContent string) []BookBibliography {
	if strings.TrimSpace(htmlContent) == "" {
		return nil
	}

	lower := strings.ToLower(htmlContent)
	var basicHTML, complHTML string

	locBasic := rxBasicSec.FindStringIndex(htmlContent)
	locCompl := rxComplSec.FindStringIndex(htmlContent)

	if locBasic != nil && locCompl != nil {
		if locBasic[0] < locCompl[0] {
			basicHTML = htmlContent[locBasic[1]:locCompl[0]]
			complHTML = htmlContent[locCompl[1]:]
		} else {
			complHTML = htmlContent[locCompl[1]:locBasic[0]]
			basicHTML = htmlContent[locBasic[1]:]
		}
	} else if locBasic != nil {
		basicHTML = htmlContent[locBasic[1]:]
	} else if locCompl != nil {
		complHTML = htmlContent[locCompl[1]:]
	} else {
		// Se não encontrou cabeçalhos h1-h6 específicos, busca por palavras "básica" ou "complementar" em texto
		idxB := strings.Index(lower, "básica")
		if idxB == -1 {
			idxB = strings.Index(lower, "basica")
		}
		idxC := strings.Index(lower, "complementar")

		if idxB != -1 && idxC != -1 {
			if idxB < idxC {
				basicHTML = htmlContent[idxB:idxC]
				complHTML = htmlContent[idxC:]
			} else {
				complHTML = htmlContent[idxC:idxB]
				basicHTML = htmlContent[idxB:]
			}
		} else {
			basicHTML = htmlContent
		}
	}

	var books []BookBibliography
	books = append(books, parseBookSection(basicHTML, "basica")...)
	books = append(books, parseBookSection(complHTML, "complementar")...)

	return books
}

func parseBookSection(sectionHTML string, bibType string) []BookBibliography {
	if strings.TrimSpace(sectionHTML) == "" {
		return nil
	}

	// Tenta extrair itens por <li>
	matches := rxLiTag.FindAllStringSubmatch(sectionHTML, -1)
	var rawItems []string

	if len(matches) > 0 {
		for _, m := range matches {
			if len(m) > 1 && strings.TrimSpace(m[1]) != "" {
				rawItems = append(rawItems, m[1])
			}
		}
	} else {
		// Se não houver <li>, tenta por <p>
		pMatches := rxPTag.FindAllStringSubmatch(sectionHTML, -1)
		for _, pm := range pMatches {
			if len(pm) > 1 && strings.TrimSpace(pm[1]) != "" {
				rawItems = append(rawItems, pm[1])
			}
		}
	}

	var list []BookBibliography
	for idx, itemHTML := range rawItems {
		cleanItem := CleanCanvasHTML(itemHTML)
		if len(cleanItem) < 10 {
			continue
		}

		// Extrai autores (strong/b)
		authors := ""
		if m := rxStrongTag.FindStringSubmatch(itemHTML); len(m) > 1 {
			authors = CleanCanvasHTML(m[1])
			if authors == "" && len(m) > 2 {
				authors = CleanCanvasHTML(m[2])
			}
		}
		authors = strings.TrimSuffix(strings.TrimSpace(authors), ".")

		// Extrai título (em/i)
		title := ""
		if m := rxEmTag.FindStringSubmatch(itemHTML); len(m) > 1 {
			title = CleanCanvasHTML(m[1])
			if title == "" && len(m) > 2 {
				title = CleanCanvasHTML(m[2])
			}
		}
		title = strings.TrimSuffix(strings.TrimSpace(title), ".")

		// Se não achou título em tags, tenta separar autores da citação
		if title == "" {
			parts := strings.Split(cleanItem, ".")
			if len(parts) > 1 {
				title = strings.TrimSpace(parts[1])
			} else {
				title = cleanItem
			}
		}

		// Extrai link da Minha Biblioteca ou externo
		bookURL := ""
		anchorMatches := rxAnchorTag.FindAllStringSubmatch(itemHTML, -1)
		for _, am := range anchorMatches {
			if len(am) > 1 {
				u := am[1]
				if strings.Contains(u, "minhabiblioteca") || bookURL == "" {
					bookURL = u
				}
			}
		}

		// Extrai ISBN
		isbn := ""
		if m := rxISBN.FindStringSubmatch(cleanItem); len(m) > 1 {
			isbn = strings.TrimSpace(m[1])
		}

		// Fallback de URL para busca direta pelo ISBN ou título caso não tenha link na ementa
		if bookURL == "" {
			if isbn != "" {
				bookURL = fmt.Sprintf("https://integrada.minhabiblioteca.com.br/#/books/%s", isbn)
			} else {
				bookURL = fmt.Sprintf("https://integrada.minhabiblioteca.com.br/#/search?query=%s", url.QueryEscape(title))
			}
		}

		// Gera tópicos/keywords dinâmicos para busca semântica a partir do título e citação
		keywords := extractKeywords(title + " " + cleanItem)

		bookID := fmt.Sprintf("%s_%d", bibType, idx+1)
		if authors != "" {
			firstAuthor := strings.Fields(authors)[0]
			bookID = fmt.Sprintf("%s_%s_%d", bibType, strings.ToLower(firstAuthor), idx+1)
		}

		list = append(list, BookBibliography{
			ID:       bookID,
			Type:     bibType,
			Authors:  authors,
			Title:    title,
			ISBN:     isbn,
			URL:      bookURL,
			Citation: cleanItem,
			Topics:   keywords,
		})
	}

	return list
}

func extractKeywords(text string) []string {
	text = strings.ToLower(text)
	text = strings.ReplaceAll(text, ",", " ")
	text = strings.ReplaceAll(text, ".", " ")
	text = strings.ReplaceAll(text, ":", " ")
	text = strings.ReplaceAll(text, ";", " ")
	text = strings.ReplaceAll(text, "(", " ")
	text = strings.ReplaceAll(text, ")", " ")
	text = strings.ReplaceAll(text, "/", " ")
	text = strings.ReplaceAll(text, "-", " ")

	stopWords := map[string]bool{
		"para": true, "com": true, "seus": true, "como": true, "sobre": true,
		"pelo": true, "pela": true, "pelos": true, "pelas": true, "numa": true,
		"mais": true, "suas": true, "esse": true, "essa": true,
		"livro": true, "ebook": true, "isbn": true, "editora": true, "edicao": true,
		"porto": true, "alegre": true, "paulo": true, "rio": true, "janeiro": true,
		"acessar": true, "biblioteca": true, "minha": true, "integrada": true,
	}

	tokens := strings.Fields(text)
	seen := make(map[string]bool)
	var keywords []string

	for _, tok := range tokens {
		tok = strings.TrimSpace(tok)
		if len(tok) >= 3 && !stopWords[tok] && !seen[tok] {
			seen[tok] = true
			keywords = append(keywords, tok)
		}
	}
	return keywords
}

// GetRecommendedReadings consulta a bibliografia da disciplina de forma 100% dinâmica via API do Canvas LMS
func (c *CanvasClient) GetRecommendedReadings(courseIDOrQuery string, topic string, bibType string) (*BibliographyResult, error) {
	// 1. Resolve dinamicamente a disciplina (ID numérico, nome da matéria ou período)
	resolvedCourseID := ""
	var course map[string]any

	if courseIDOrQuery != "" {
		cid, err := c.ResolveCourseID(courseIDOrQuery)
		if err != nil {
			return nil, err
		}
		resolvedCourseID = cid
	} else {
		// Tenta resolver automaticamente para a turma ativa do semestre
		courses, err := c.ListCourses()
		if err == nil {
			var activeCourses []map[string]any
			for _, crs := range courses {
				if isCur, ok := crs["is_current_term"].(bool); ok && isCur {
					activeCourses = append(activeCourses, crs)
				}
			}
			if len(activeCourses) == 1 {
				resolvedCourseID = getCourseIDStr(activeCourses[0])
			} else if len(activeCourses) > 1 {
				// Se há mais de uma turma ativa, seleciona a primeira como padrão para leitura
				resolvedCourseID = getCourseIDStr(activeCourses[0])
			}
		}
		if resolvedCourseID == "" {
			cid, err := c.ResolveCourseID("")
			if err != nil {
				return nil, err
			}
			resolvedCourseID = cid
		}
	}

	// 2. Obtém detalhes atualizados do curso no Canvas LMS
	var err error
	course, err = c.GetCourseDetails(resolvedCourseID)
	if err != nil {
		// Fallback para GetCourse básico se falhar
		courses, _ := c.ListCourses()
		for _, crs := range courses {
			if getCourseIDStr(crs) == resolvedCourseID {
				course = crs
				break
			}
		}
		if course == nil {
			return nil, fmt.Errorf("não foi possível obter informações do curso %s: %w", resolvedCourseID, err)
		}
	}

	courseName, _ := course["name"].(string)
	cleanName, _ := course["clean_name"].(string)
	if cleanName == "" {
		cleanName = courseName
	}
	period, _ := course["period"].(string)
	termObj, _ := course["term"].(map[string]any)
	termName, _ := termObj["name"].(string)
	if termName == "" {
		termName = "Semestre Vigente"
	}

	// 3. Procura dinamicamente a página de ementa / bibliografia do curso
	var allBooks []BookBibliography
	syllabusSource := "Não identificada"
	syllabusText := ""

	pages, err := c.GetCoursePages(resolvedCourseID)
	if err == nil && len(pages) > 0 {
		var targetPageURL string
		for _, p := range pages {
			pURL, _ := p["url"].(string)
			pTitle, _ := p["title"].(string)
			lowerStr := strings.ToLower(pURL + " " + pTitle)

			if strings.Contains(lowerStr, "ementa") ||
				strings.Contains(lowerStr, "bibliografia") ||
				strings.Contains(lowerStr, "syllabus") ||
				strings.Contains(lowerStr, "plano-de-ensino") {
				targetPageURL = pURL
				syllabusSource = fmt.Sprintf("Página Canvas: '%s'", pTitle)
				break
			}
		}

		if targetPageURL != "" {
			pageData, err := c.GetCoursePage(resolvedCourseID, targetPageURL)
			if err == nil {
				if bodyHTML, ok := pageData["body"].(string); ok && strings.TrimSpace(bodyHTML) != "" {
					allBooks = ExtractBooksFromHTML(bodyHTML)
					syllabusText = ExtractSyllabusText(bodyHTML)
				}
			}
		}
	}

	// Se não achou em páginas, tenta extrair de syllabus_body
	if len(allBooks) == 0 {
		if syllBody, ok := course["syllabus_body"].(string); ok && strings.TrimSpace(syllBody) != "" {
			allBooks = ExtractBooksFromHTML(syllBody)
			syllabusText = ExtractSyllabusText(syllBody)
			if len(allBooks) > 0 {
				syllabusSource = "Syllabus Oficial do Canvas"
			}
		}
	}

	// 4. Monta URLs de pesquisa global na Minha Biblioteca
	topicLower := strings.ToLower(strings.TrimSpace(topic))
	typeLower := strings.ToLower(strings.TrimSpace(bibType))

	var searchParam string
	if topicLower != "" && topicLower != "all" && topicLower != "todos" {
		searchParam = topic
	} else {
		searchParam = cleanName
	}
	globalSearchURL := fmt.Sprintf("https://integrada.minhabiblioteca.com.br/#/search?query=%s", url.QueryEscape(searchParam))

	// 5. Filtra por tipo (básica/complementar) e por tópico
	filteredBooks := make([]BookBibliography, 0)
	topicReadings := make([]TopicReading, 0)

	for _, b := range allBooks {
		// Filtro por tipo
		if typeLower != "" && typeLower != "todas" && typeLower != "all" {
			if b.Type != typeLower {
				continue
			}
		}

		// Filtro por tópico
		matchTopic := false
		if topicLower == "" || topicLower == "all" || topicLower == "todos" {
			matchTopic = true
		} else {
			titleLower := strings.ToLower(b.Title)
			authorsLower := strings.ToLower(b.Authors)
			citLower := strings.ToLower(b.Citation)

			if strings.Contains(titleLower, topicLower) ||
				strings.Contains(authorsLower, topicLower) ||
				strings.Contains(citLower, topicLower) {
				matchTopic = true
			} else {
				for _, kw := range b.Topics {
					if strings.Contains(kw, topicLower) || strings.Contains(topicLower, kw) {
						matchTopic = true
						break
					}
				}
			}

			// Heurística de correspondência pedagógica:
			// Se o tópico for ponteiros/memória/malloc/alocação e a obra foca em programação C/C++
			if !matchTopic {
				if (strings.Contains(topicLower, "ponteir") || strings.Contains(topicLower, "memori") || strings.Contains(topicLower, "alocac") || strings.Contains(topicLower, "malloc")) &&
					(strings.Contains(titleLower, " em c") || strings.Contains(titleLower, "em c++") || strings.Contains(citLower, " c.") || strings.Contains(citLower, " c,")) {
					matchTopic = true
				}
			}
		}

		if matchTopic {
			filteredBooks = append(filteredBooks, b)

			if topicLower != "" && topicLower != "all" && topicLower != "todos" {
				topicReadings = append(topicReadings, TopicReading{
					Topic:           topic,
					RecommendedBook: b,
					ReadingGoal:     fmt.Sprintf("Fundamentação conceitual e aprofundamento técnico em %s.", topic),
				})
			}
		}
	}

	// 6. Constrói tabela Markdown didática e elegante
	var md strings.Builder
	md.WriteString("### 📚 Acervo Oficial — Minha Biblioteca (Afya / São Lucas)\n")
	periodInfo := ""
	if period != "" {
		periodInfo = fmt.Sprintf(" (%s)", period)
	}
	md.WriteString(fmt.Sprintf("**Disciplina:** `%s`%s • **Termo:** `%s`\n", cleanName, periodInfo, termName))
	md.WriteString(fmt.Sprintf("**Fonte da Ementa:** %s • **Total de obras cadastradas no Canvas:** %d\n\n", syllabusSource, len(allBooks)))

	if topicLower != "" && topicLower != "all" && topicLower != "todos" {
		md.WriteString(fmt.Sprintf("🔍 **Filtro de Leitura Ativo:** `%s`\n\n", strings.ToUpper(topic)))
	}

	if len(allBooks) == 0 {
		md.WriteString("> [!WARNING]\n")
		md.WriteString("> Não foi localizada uma página com 'Ementa' ou 'Bibliografia' cadastrada no Canvas para esta disciplina.\n")
		md.WriteString(fmt.Sprintf("> Você pode consultar diretamente todo o acervo da Minha Biblioteca para **%s** através do link abaixo:\n", searchParam))
		md.WriteString(fmt.Sprintf("> 🔗 [Pesquisar '%s' no Acervo Global da Minha Biblioteca](%s)\n\n", html.EscapeString(searchParam), globalSearchURL))
	} else if len(filteredBooks) == 0 {
		md.WriteString(fmt.Sprintf("_Nenhuma obra na ementa cadastrada corresponde exatamente ao filtro `%s`._\n\n", topic))
		md.WriteString(fmt.Sprintf("📖 **Sugestão Dinâmica:** Você pode pesquisar todas as obras disponíveis na Minha Biblioteca sobre este tema:\n"))
		md.WriteString(fmt.Sprintf("👉 [Pesquisar '%s' na Minha Biblioteca da Afya](%s)\n\n", html.EscapeString(topic), globalSearchURL))
	} else {
		md.WriteString("| Tipo | Título da Obra | Autores / Referência | Acesso Direto |\n")
		md.WriteString("| :---: | :--- | :--- | :---: |\n")
		for _, b := range filteredBooks {
			typeBadge := "📘 Básica"
			if b.Type == "complementar" {
				typeBadge = "📗 Complem."
			}
			authorDisp := b.Authors
			if authorDisp == "" {
				authorDisp = b.Citation
			}
			md.WriteString(fmt.Sprintf("| %s | **%s** | %s | [Acessar E-book](%s) |\n",
				typeBadge, b.Title, authorDisp, b.URL))
		}
		md.WriteString("\n")
		if topicLower != "" && topicLower != "all" && topicLower != "todos" {
			md.WriteString(fmt.Sprintf("🔎 **Expandir Pesquisa:** [Buscar mais livros sobre '%s' na Minha Biblioteca](%s)\n\n", html.EscapeString(topic), globalSearchURL))
		}
	}

	md.WriteString("> [!TIP]\n")
	md.WriteString("> O acesso direto e autenticado dos estudantes aos e-books é garantido via **LTI Single Sign-On** no Canvas LMS da disciplina.\n")

	return &BibliographyResult{
		CourseID:        resolvedCourseID,
		CourseName:      courseName,
		CleanName:       cleanName,
		Period:          period,
		TermName:        termName,
		Institution:     "Centro Universitário São Lucas Ji-Paraná • Afya",
		LibraryName:     "Minha Biblioteca",
		LibraryLTIURL:   "https://minhabiblioteca.azurewebsites.net/MinhaBiblioteca/Canvas",
		GlobalSearchURL: globalSearchURL,
		SyllabusSource:  syllabusSource,
		SyllabusText:    syllabusText,
		TotalBooks:      len(allBooks),
		Books:           filteredBooks,
		TopicReadings:   topicReadings,
		MarkdownTable:   md.String(),
	}, nil
}
