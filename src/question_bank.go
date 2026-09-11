package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ENADEAnswer representa uma alternativa de questão com justificativa pedagógica
type ENADEAnswer struct {
	Letter                   string `json:"letter"` // "A", "B", "C", "D", "E"
	Text                     string `json:"text"`
	IsCorrect                bool   `json:"is_correct"`
	PedagogicalJustification string `json:"pedagogical_justification"` // Explica por que é correta ou por que o distrator está errado
}

// ENADEQuestion representa um item calibrado no modelo ENADE/CONSEPE
type ENADEQuestion struct {
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	ContextText string        `json:"context_text"` // Texto-base contextualizado e situação-problema
	Subject     string        `json:"subject"`      // Disciplina (ex: "Estrutura de Dados")
	Competence  string        `json:"competence"`   // Competência / DCN (ex: "Ponteiros e Alocação de Memória")
	Difficulty  string        `json:"difficulty"`   // "Fácil", "Médio", "Difícil"
	Points      float64       `json:"points"`       // Pontuação padrão
	Answers     []ENADEAnswer `json:"answers"`
}

// QuestionBank representa um banco de itens institucional
type QuestionBank struct {
	ID             string          `json:"id"`
	Title          string          `json:"title"`
	Subject        string          `json:"subject"`
	Competence     string          `json:"competence"`
	Description    string          `json:"description"`
	CreatedAt      string          `json:"created_at"`
	UpdatedAt      string          `json:"updated_at"`
	TotalQuestions int             `json:"total_questions"`
	Questions      []ENADEQuestion `json:"questions"`
}

// MockExamResult consolida o resultado da geração de um simulado ou prova no Canvas
type MockExamResult struct {
	QuizID          string          `json:"quiz_id"`
	CourseID        string          `json:"course_id"`
	CourseName      string          `json:"course_name"`
	Title           string          `json:"title"`
	TotalQuestions  int             `json:"total_questions"`
	PointsPossible  float64         `json:"points_possible"`
	TimeLimit       int             `json:"time_limit"`
	DueAtBR         string          `json:"due_at_br"`
	QuizURL         string          `json:"quiz_url"`
	SpeedGraderURL  string          `json:"speed_grader_url"`
	Questions       []ENADEQuestion `json:"questions"`
	MarkdownSummary string          `json:"markdown_summary"`
}

func getQuestionBanksDir() string {
	dir := filepath.Join("doc", "question_banks")
	_ = os.MkdirAll(dir, 0755)
	return dir
}

// CreateQuestionBank cria um novo banco de itens estruturado por competência
func (c *CanvasClient) CreateQuestionBank(title, subject, competence, description string) (*QuestionBank, error) {
	if strings.TrimSpace(title) == "" {
		return nil, fmt.Errorf("o título do banco de questões é obrigatório")
	}

	banksDir := getQuestionBanksDir()
	slug := strings.ToLower(strings.TrimSpace(title))
	slug = strings.ReplaceAll(slug, " ", "_")
	slug = strings.ReplaceAll(slug, "/", "_")
	slug = strings.ReplaceAll(slug, "-", "_")

	bankID := fmt.Sprintf("bank_%s", slug)
	filePath := filepath.Join(banksDir, bankID+".json")

	now := time.Now().Format(time.RFC3339)
	bank := QuestionBank{
		ID:             bankID,
		Title:          title,
		Subject:        subject,
		Competence:     competence,
		Description:    description,
		CreatedAt:      now,
		UpdatedAt:      now,
		TotalQuestions: 0,
		Questions:      []ENADEQuestion{},
	}

	// Se já existir, apenas carrega para manter itens prévios
	if data, err := os.ReadFile(filePath); err == nil {
		_ = json.Unmarshal(data, &bank)
		bank.Title = title
		if subject != "" {
			bank.Subject = subject
		}
		if competence != "" {
			bank.Competence = competence
		}
		if description != "" {
			bank.Description = description
		}
		bank.UpdatedAt = now
	}

	bytes, err := json.MarshalIndent(bank, "", "  ")
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(filePath, bytes, 0644); err != nil {
		return nil, fmt.Errorf("erro ao salvar banco de questões: %w", err)
	}

	return &bank, nil
}

// AddItemToBank adiciona uma questão calibrada no modelo ENADE ao banco indicado
func (c *CanvasClient) AddItemToBank(bankID string, q ENADEQuestion) (*ENADEQuestion, error) {
	if strings.TrimSpace(bankID) == "" {
		return nil, fmt.Errorf("bank_id é obrigatório")
	}
	if strings.TrimSpace(q.ContextText) == "" {
		return nil, fmt.Errorf("o texto-base/enunciado da questão é obrigatório")
	}
	if len(q.Answers) < 2 {
		return nil, fmt.Errorf("a questão deve conter pelo menos 2 alternativas")
	}

	banksDir := getQuestionBanksDir()
	targetFile := filepath.Join(banksDir, bankID+".json")
	if !strings.HasSuffix(bankID, ".json") {
		targetFile = filepath.Join(banksDir, bankID+".json")
	}

	data, err := os.ReadFile(targetFile)
	if err != nil {
		return nil, fmt.Errorf("banco de questões '%s' não encontrado em %s: %w", bankID, targetFile, err)
	}

	var bank QuestionBank
	if err := json.Unmarshal(data, &bank); err != nil {
		return nil, fmt.Errorf("erro ao decodificar banco de questões: %w", err)
	}

	// Atribui ID e metadados à questão
	if q.ID == "" {
		q.ID = fmt.Sprintf("enade_%s_q%d", bank.ID, len(bank.Questions)+1)
	}
	if q.Title == "" {
		q.Title = fmt.Sprintf("Item %d: %s", len(bank.Questions)+1, bank.Competence)
	}
	if q.Subject == "" {
		q.Subject = bank.Subject
	}
	if q.Competence == "" {
		q.Competence = bank.Competence
	}
	if q.Difficulty == "" {
		q.Difficulty = "Médio"
	}
	if q.Points == 0 {
		q.Points = 10.0
	}

	// Garante letras para as alternativas
	letters := []string{"A", "B", "C", "D", "E"}
	hasCorrect := false
	for idx := range q.Answers {
		if idx < len(letters) {
			q.Answers[idx].Letter = letters[idx]
		}
		if q.Answers[idx].IsCorrect {
			hasCorrect = true
		}
	}
	if !hasCorrect {
		q.Answers[0].IsCorrect = true // Garante pelo menos uma correta
	}

	bank.Questions = append(bank.Questions, q)
	bank.TotalQuestions = len(bank.Questions)
	bank.UpdatedAt = time.Now().Format(time.RFC3339)

	updatedBytes, err := json.MarshalIndent(bank, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(targetFile, updatedBytes, 0644); err != nil {
		return nil, fmt.Errorf("falha ao gravar questão no banco: %w", err)
	}

	return &q, nil
}

// ListQuestionBanks lista todos os bancos de questões cadastrados
func (c *CanvasClient) ListQuestionBanks(subjectFilter string) ([]QuestionBank, error) {
	banksDir := getQuestionBanksDir()
	files, err := os.ReadDir(banksDir)
	if err != nil {
		return nil, err
	}

	var list []QuestionBank
	subFilterLower := strings.ToLower(strings.TrimSpace(subjectFilter))

	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".json") {
			path := filepath.Join(banksDir, f.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}

			var b QuestionBank
			if err := json.Unmarshal(data, &b); err == nil {
				if subFilterLower == "" || strings.Contains(strings.ToLower(b.Subject), subFilterLower) || strings.Contains(strings.ToLower(b.Title), subFilterLower) {
					list = append(list, b)
				}
			}
		}
	}
	return list, nil
}

// GenerateMockExam cria um simulado oficial no Canvas LMS sorteando questões dos bancos de itens
func (c *CanvasClient) GenerateMockExam(courseIDOrQuery string, examTitle string, bankIDs []string, pickCount int, pointsPerQuestion float64, timeLimitMinutes int, dueAt string, publishNow bool, targetModuleIDOrName string) (*MockExamResult, error) {
	// 1. Resolve o curso
	courseID, err := c.ResolveCourseID(courseIDOrQuery)
	if err != nil {
		return nil, err
	}

	course, _ := c.GetCourseDetails(courseID)
	courseName, _ := course["name"].(string)

	if examTitle == "" {
		examTitle = fmt.Sprintf("Simulado Oficial ENADE / N1 — %s", time.Now().Format("02/01/2006"))
	}
	if pickCount <= 0 {
		pickCount = 10
	}
	if pointsPerQuestion <= 0 {
		pointsPerQuestion = 10.0
	}
	if timeLimitMinutes <= 0 {
		timeLimitMinutes = 60
	}

	// 2. Coleta o acervo de questões disponíveis
	var pool []ENADEQuestion
	allBanks, err := c.ListQuestionBanks("")
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar bancos de questões: %w", err)
	}

	selectedBankIDs := make(map[string]bool)
	for _, bID := range bankIDs {
		selectedBankIDs[strings.TrimSpace(bID)] = true
	}

	for _, b := range allBanks {
		if len(bankIDs) == 0 || selectedBankIDs[b.ID] || selectedBankIDs[b.Title] {
			pool = append(pool, b.Questions...)
		}
	}

	// Se não houver questões no pool, cria automaticamente um acervo demonstrativo calibrado de Estrutura de Dados
	if len(pool) == 0 {
		defaultBank, _ := c.CreateQuestionBank("Banco de Itens: Estrutura de Dados & Ponteiros", "Estrutura de Dados", "Estruturas Lineares e Memória", "Itens no modelo ENADE da disciplina")
		seedQuestions := getSeedENADEQuestions()
		for _, sq := range seedQuestions {
			_, _ = c.AddItemToBank(defaultBank.ID, sq)
			pool = append(pool, sq)
		}
	}

	// 3. Sorteia aleatoriamente N questões sem repetição
	rand.Seed(time.Now().UnixNano())
	shuffled := make([]ENADEQuestion, len(pool))
	copy(shuffled, pool)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	numToPick := pickCount
	if numToPick > len(shuffled) {
		numToPick = len(shuffled)
	}
	selectedQuestions := shuffled[:numToPick]

	// 4. Converte as questões para o formato Canvas QuizQuestion
	var canvasQuestions []QuizQuestion
	totalPoints := 0.0

	for idx, sq := range selectedQuestions {
		qTitle := fmt.Sprintf("Questão %d: %s (%s)", idx+1, sq.Competence, sq.Difficulty)

		// Formata o HTML do texto da questão com texto-base e situação-problema
		var qTextHTML strings.Builder
		qTextHTML.WriteString(fmt.Sprintf("<div style='font-family: sans-serif; line-height: 1.6;'>"))
		qTextHTML.WriteString(fmt.Sprintf("<p style='background: #f1f5f9; padding: 10px 14px; border-left: 4px solid #2563eb; border-radius: 4px; font-size: 13px; color: #475569;'>"))
		qTextHTML.WriteString(fmt.Sprintf("<b>Competência Avaliada:</b> %s • <b>Nível:</b> %s</p>", sq.Competence, sq.Difficulty))
		qTextHTML.WriteString(fmt.Sprintf("<div style='font-size: 15px; color: #1e293b; margin: 16px 0;'>%s</div>", sq.ContextText))
		qTextHTML.WriteString(fmt.Sprintf("</div>"))

		var answers []QuizAnswer
		for _, ans := range sq.Answers {
			weight := 0
			if ans.IsCorrect {
				weight = 100
			}

			// Constrói justificativa pedagógica imediata
			comment := ans.PedagogicalJustification
			if comment == "" {
				if ans.IsCorrect {
					comment = "✅ Alternativa correta com base nos conceitos técnicos da disciplina."
				} else {
					comment = "❌ Distrator incorreto: revise os conceitos fundamentais da aula."
				}
			}

			answers = append(answers, QuizAnswer{
				Text:    ans.Text,
				Weight:  weight,
				Comment: comment,
			})
		}

		canvasQuestions = append(canvasQuestions, QuizQuestion{
			Title:          qTitle,
			Text:           qTextHTML.String(),
			Type:           "multiple_choice_question",
			PointsPossible: pointsPerQuestion,
			Answers:        answers,
		})
		totalPoints += pointsPerQuestion
	}

	// 5. Cria o Quiz oficial no Canvas LMS
	pub := publishNow
	descriptionHTML := fmt.Sprintf("<p><b>%s</b></p><p>Avaliação formativa no Modelo ENADE elaborada pelo Professor Me. Karan Luciano. Sorteio aleatório de %d itens calibrados por competência curricular.</p>", examTitle, numToPick)

	createParams := CreateQuizParams{
		CourseID:        courseID,
		Title:           examTitle,
		Description:     descriptionHTML,
		QuizType:        "assignment",
		TimeLimit:       timeLimitMinutes,
		AllowedAttempts: 1,
		DueAt:           dueAt,
		Published:       &pub,
		Questions:       canvasQuestions,
	}

	quizResRaw, err := c.CreateQuiz(createParams)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar simulado no Canvas: %w", err)
	}

	quizRes, _ := quizResRaw.(map[string]any)
	quizID := ""
	if qid, ok := quizRes["id"]; ok && qid != nil {
		switch v := qid.(type) {
		case float64:
			quizID = strconv.FormatInt(int64(v), 10)
		case int:
			quizID = strconv.Itoa(v)
		case string:
			quizID = v
		default:
			quizID = fmt.Sprintf("%v", v)
		}
	}
	quizURL, _ := quizRes["html_url"].(string)
	if (quizID == "" || quizID == "<nil>") && quizURL != "" {
		parts := strings.Split(quizURL, "/quizzes/")
		if len(parts) > 1 {
			quizID = parts[1]
		}
	}

	speedGraderURL, _ := quizRes["speed_grader_url"].(string)
	if speedGraderURL == "" {
		if assignID, ok := quizRes["assignment_id"]; ok && assignID != nil {
			speedGraderURL = fmt.Sprintf("https://afya.instructure.com/courses/%s/gradebook/speed_grader?assignment_id=%v", courseID, assignID)
		}
	}

	// 6. Vincula ao módulo de destino caso informado
	if targetModuleIDOrName != "" && quizID != "" {
		modulesRaw, _ := c.ListModules(courseID)
		var mList []map[string]any
		mb, _ := json.Marshal(modulesRaw)
		_ = json.Unmarshal(mb, &mList)

		var targetModID string
		targetQuery := strings.ToLower(strings.TrimSpace(targetModuleIDOrName))
		for _, m := range mList {
			mID := fmt.Sprintf("%v", m["id"])
			mName, _ := m["name"].(string)
			if mID == targetModuleIDOrName || strings.Contains(strings.ToLower(mName), targetQuery) {
				targetModID = mID
				break
			}
		}

		if targetModID != "" {
			_, _ = c.AddModuleItem(AddModuleItemParams{
				CourseID:  courseID,
				ModuleID:  targetModID,
				Title:     examTitle,
				Type:      "Quiz",
				ContentID: quizID,
			})
		}
	}

	// 7. Monta Markdown pedagógico
	var md strings.Builder
	md.WriteString("### 📝 Simulado / Prova ENADE Criado com Sucesso no Canvas\n\n")
	md.WriteString(fmt.Sprintf("- **Título da Avaliação:** `%s`\n", examTitle))
	md.WriteString(fmt.Sprintf("- **Disciplina:** %s (`ID %s`)\n", courseName, courseID))
	md.WriteString(fmt.Sprintf("- **Total de Questões Sorteadas:** %d itens calibrados\n", numToPick))
	md.WriteString(fmt.Sprintf("- **Pontuação Total:** `%.1f pontos` (%.1f pts por item)\n", totalPoints, pointsPerQuestion))
	md.WriteString(fmt.Sprintf("- **Tempo Limite:** `%d minutos` • **Data/Prazo:** %s\n", timeLimitMinutes, FormatBRDateTime(dueAt)))
	md.WriteString(fmt.Sprintf("- **Links Diretos:** [Acessar Simulado](%s) • [SpeedGrader](%s)\n\n", quizURL, speedGraderURL))

	md.WriteString("| Item | Competência Avaliada | Dificuldade | Pontos |\n")
	md.WriteString("| :---: | :--- | :---: | :---: |\n")
	for idx, q := range selectedQuestions {
		md.WriteString(fmt.Sprintf("| %d | **%s** | %s | %.1f pts |\n", idx+1, q.Competence, q.Difficulty, pointsPerQuestion))
	}

	md.WriteString("\n> [!TIP]\n")
	md.WriteString("> Todas as questões foram inseridas no Canvas com **autocorreção imediata** e **justificativas pedagógicas completas para cada distrator**, promovendo feedback reflexivo imediato para os estudantes.\n")

	return &MockExamResult{
		QuizID:          quizID,
		CourseID:        courseID,
		CourseName:      courseName,
		Title:           examTitle,
		TotalQuestions:  numToPick,
		PointsPossible:  totalPoints,
		TimeLimit:       timeLimitMinutes,
		DueAtBR:         FormatBRDateTime(dueAt),
		QuizURL:         quizURL,
		SpeedGraderURL:  speedGraderURL,
		Questions:       selectedQuestions,
		MarkdownSummary: md.String(),
	}, nil
}

// getSeedENADEQuestions fornece itens de partida com alto rigor pedagógico no padrão ENADE
func getSeedENADEQuestions() []ENADEQuestion {
	return []ENADEQuestion{
		{
			ID:          "enade_ed_01",
			Title:       "Aritmética de Ponteiros e Memória RAM",
			ContextText: "Considere um sistema embarcado crítico desenvolvido em linguagem C onde a eficiência de acesso à memória é mandatória. Um desenvolvedor declara o vetor `int v[5] = {10, 20, 30, 40, 50};` em uma arquitetura de 32 bits (onde `sizeof(int) == 4` bytes). Sabendo que o endereço inicial do vetor é `0x1000`, analise a operação `*(v + 3)` e sua relação com o endereço de memória manipulado.",
			Subject:     "Estrutura de Dados",
			Competence:  "Ponteiros e Aritmética de Endereços",
			Difficulty:  "Médio",
			Points:      10.0,
			Answers: []ENADEAnswer{
				{
					Letter:                   "A",
					Text:                     "A expressão avalia para o valor 40, acessando o endereço físico 0x100C.",
					IsCorrect:                true,
					PedagogicalJustification: "Correto. O identificador 'v' decai para ponteiro base (0x1000). O deslocamento '+ 3' incrementa 3 * sizeof(int) = 12 bytes (0xC em hexadecimal), resultando em 0x100C, cujo conteúdo desreferenciado é 40.",
				},
				{
					Letter:                   "B",
					Text:                     "A expressão avalia para o valor 30, acessando o endereço físico 0x1003.",
					IsCorrect:                false,
					PedagogicalJustification: "Incorreto. Em aritmética de ponteiros, a adição soma unidades do tipo apontado (4 bytes), e não bytes individuais brutos (+3). Além disso, o índice 3 corresponde ao quarto elemento (40).",
				},
				{
					Letter:                   "C",
					Text:                     "A expressão causa falha de segmentação (Segmentation Fault) por violação de limite.",
					IsCorrect:                false,
					PedagogicalJustification: "Incorreto. O vetor possui tamanho 5 (índices 0 a 4). O índice 3 é perfeitamente válido e reside dentro dos limites alocados na pilha.",
				},
				{
					Letter:                   "D",
					Text:                     "A expressão retorna o ponteiro para o elemento 3 sem acessar seu valor.",
					IsCorrect:                false,
					PedagogicalJustification: "Incorreto. O operador de desreferenciação '*' antes dos parênteses acessa diretamente o valor contido no endereço apontado.",
				},
			},
		},
		{
			ID:          "enade_ed_02",
			Title:       "Vazamento de Memória e Gerenciamento Dinâmico (Memory Leak)",
			ContextText: "Durante a execução de um servidor de monitoramento em tempo real escrito em C, observou-se que o consumo de memória RAM do processo cresce linearmente até o sistema operacional acionar o OOM-Killer (Out of Memory Killer). A análise do código revelou uma função executada em loop contendo: `Node* n = (Node*) malloc(sizeof(Node)); /* processamento */ n = NULL;`.",
			Subject:     "Estrutura de Dados",
			Competence:  "Alocação Dinâmica e Ciclo de Vida da Heap",
			Difficulty:  "Fácil",
			Points:      10.0,
			Answers: []ENADEAnswer{
				{
					Letter:                   "A",
					Text:                     "Ocorre vazamento de memória (memory leak), pois a referência na stack é perdida sem que a memória alocada na heap seja liberada com free(n).",
					IsCorrect:                true,
					PedagogicalJustification: "Correto. Atribuir NULL ao ponteiro local apenas sobrescreve o endereço salvo na stack, tornando o bloco de memória na heap inacessível e irrecuperável até o término do processo.",
				},
				{
					Letter:                   "B",
					Text:                     "A memória é liberada automaticamente pelo coletor de lixo (Garbage Collector) nativo da linguagem C ao receber NULL.",
					IsCorrect:                false,
					PedagogicalJustification: "Incorreto. A linguagem C padrão não possui Garbage Collector. A liberação de blocos da Heap deve ser feita explicitamente pelo programador via free().",
				},
				{
					Letter:                   "C",
					Text:                     "O código causa falha de compilação, pois um ponteiro alocado com malloc não pode receber NULL.",
					IsCorrect:                false,
					PedagogicalJustification: "Incorreto. A atribuição n = NULL é sintaticamente e tipologicamente válida em C, porém logicamente desastrosa sem o free() prévio.",
				},
				{
					Letter:                   "D",
					Text:                     "O ponteiro se torna uma referência pendente (dangling pointer) que causa corrupção imediata da pilha.",
					IsCorrect:                false,
					PedagogicalJustification: "Incorreto. 'Dangling pointer' ocorre quando o ponteiro continua apontando para um bloco já liberado. Ao receber NULL, ele é anulado, gerando leak e não dangling.",
				},
			},
		},
		{
			ID:          "enade_ed_03",
			Title:       "Complexidade Assintótica de Árvores Binárias de Busca (BST)",
			ContextText: "Uma Árvore Binária de Busca (BST) foi construída inserindo sequencialmente chaves ordenadas de forma estritamente crescente: [10, 20, 30, 40, 50, 60, 70]. Em seguida, uma operação de busca pelo elemento 70 foi requisitada. Avalie a complexidade assintótica de tempo desta operação no pior caso.",
			Subject:     "Estrutura de Dados",
			Competence:  "Árvores Binárias e Balanceamento AVL",
			Difficulty:  "Médio",
			Points:      10.0,
			Answers: []ENADEAnswer{
				{
					Letter:                   "A",
					Text:                     "A complexidade é O(n), pois a inserção ordenada degenerou a árvore em uma lista encadeada simples de altura n.",
					IsCorrect:                true,
					PedagogicalJustification: "Correto. Sem mecanismos de autobalanceamento (como AVL ou Rubro-Negra), a inserção de chaves ordenadas faz com que cada novo nó seja inserido sempre à direita, resultando em altura h = n e tempo de busca linear O(n).",
				},
				{
					Letter:                   "B",
					Text:                     "A complexidade permanece O(log n), pois árvores binárias garantem busca logarítmica independentemente da ordem de inserção.",
					IsCorrect:                false,
					PedagogicalJustification: "Incorreto. A garantia O(log n) aplica-se apenas a árvores balanceadas. Em árvores BST degeneradas, a busca degrada para O(n).",
				},
				{
					Letter:                   "C",
					Text:                     "A complexidade é O(1), pois o elemento 70 é o maior e fica alocado diretamente na raiz da árvore.",
					IsCorrect:                false,
					PedagogicalJustification: "Incorreto. O elemento 70 é o último nó à direita da árvore degenerada, exigindo percorrer todos os nós anteriores da raiz até a folha.",
				},
				{
					Letter:                   "D",
					Text:                     "A operação é impossível, pois árvores binárias de busca não aceitam inserção ordenada de elementos.",
					IsCorrect:                false,
					PedagogicalJustification: "Incorreto. A inserção ordenada é perfeitamente aceita pelas regras da BST, apenas produz uma estrutura desbalanceada.",
				},
			},
		},
		{
			ID:          "enade_ed_04",
			Title:       "Colisões em Tabelas Hash e Complexidade O(1)",
			ContextText: "Em uma aplicação de alta performance para consulta de prontuários de pacientes da Afya, adotou-se uma Tabela Hash com tratamento de colisões por encadeamento externo (separate chaining). Considere que a função de espalhamento utilizada seja deficiente e mapeie todas as chaves inseridas para o mesmo bucket (índice 0) da tabela.",
			Subject:     "Estrutura de Dados",
			Competence:  "Tabelas Hash e Funções de Espalhamento",
			Difficulty:  "Difícil",
			Points:      10.0,
			Answers: []ENADEAnswer{
				{
					Letter:                   "A",
					Text:                     "A complexidade de busca no pior caso degrada de O(1) para O(n), equivalente à busca linear em uma lista encadeada.",
					IsCorrect:                true,
					PedagogicalJustification: "Correto. No encadeamento externo, elementos com o mesmo hash colidem na mesma lista encadeada. Se todas as chaves caem no mesmo bucket, a tabela se comporta exatamente como uma lista linear, exigindo até n comparações.",
				},
				{
					Letter:                   "B",
					Text:                     "A tabela hash interrompe a execução e lança um estouro de pilha (Stack Overflow).",
					IsCorrect:                false,
					PedagogicalJustification: "Incorreto. Encadeamento externo aloca novos nós na heap, não havendo consumo anormal da stack para provocar stack overflow.",
				},
				{
					Letter:                   "C",
					Text:                     "A complexidade de busca se mantém O(1), pois a tabela hash compensa colisões via re-hashing automático.",
					IsCorrect:                false,
					PedagogicalJustification: "Incorreto. Se a função de hash sempre direciona para o mesmo bucket, a lista daquele bucket crescerá com tamanho n, quebrando a garantia de tempo constante O(1).",
				},
				{
					Letter:                   "D",
					Text:                     "Os registros anteriores são sobrescritos e perdidos irremediavelmente a cada colisão.",
					IsCorrect:                false,
					PedagogicalJustification: "Incorreto. No tratamento por encadeamento separado, as colisões são inseridas na lista encadeada encadeada ao bucket, sem sobrescrever dados prévios.",
				},
			},
		},
	}
}
