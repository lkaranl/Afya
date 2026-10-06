# ⚡ Estratégias Agressivas para Economia Extrema de Tokens

> **Objetivo:** Reduzir o consumo de tokens de entrada e saída (input/output) em mais de **80% a 90%** nos fluxos operacionais com o Canvas LMS, geração de materiais e avaliação de estudantes.

---

## 📊 Matriz Comparativa de Impacto

| # | Estratégia | Onde Atua | Complexidade | Redução Estimada |
| :-: | :--- | :--- | :-: | :-: |
| **1** | **Avaliação Determinística Local** | Correção de Códigos C | Média | **80% a 85%** de tokens em correções |
| **2** | **Poda Agressiva de Payloads no Go** | MCP Server (`./afya-canvas`) | Baixa | **70% a 90%** no tamanho de tools |
| **3** | **Detecção de Plágio 100% Offline** | Algoritmo Shingling/AST | Baixa | **100%** de economia na triagem |
| **4** | **Templates Parametrizados (Slot-Filling)** | Geração de Feedbacks | Baixa | **60% a 70%** dos tokens de output |
| **5** | **Persistência de Estado em Manifesto Local** | Gestão de Sessão | Baixa | Impede bola de neve em chats longos |
| **6** | **Injeção Modular de Regras (Just-in-Time)** | System Prompt / Skills | Média | **~6.000 tokens** por mensagem |
| **7** | **Diffs Cirúrgicos em Código** | Edição de HTML / Scripts | Baixa | **90%** em tarefas de desenvolvimento |

---

## 🔍 Detalhamento das 7 Estratégias

### 1. 🥇 Avaliação Determinística Local (Zero LLM para Códigos Perfeitos)
* **Diagnóstico do Desperdício:** Ao avaliar uma turma de 80 alunos, enviar todos os arquivos `.c` para o modelo de IA consome entre **40.000 e 80.000 tokens** por atividade, mesmo quando a maioria dos alunos fez um código 100% idêntico ao solicitado.
* **Solução Agressiva:**
  * Implementar um *test harness* determinístico local que:
    1. Compila o código com `gcc -Wall -Wextra -std=c11`;
    2. Executa casos de teste pré-definidos (entradas padrão e asserção de saídas);
    3. Verifica vazamentos básicos de memória (`-fsanitize=address` ou análise estática).
  * **Regra de Ouro:** Códigos que compilaram perfeitamente e passaram em 100% dos testes recebem nota integral e feedback padrão **sem que o código passe pelo modelo de IA**.
  * **O modelo só é acionado para exceções:** Apenas alunos com falhas de compilação, lógica ou dúvidas conceituais são encaminhados ao LLM para diagnóstico pedagógico cirúrgico.

---

### 2. ✂️ Poda Agressiva de Payloads & Cache L1 em SQLite no Go (Universal)
* **Diagnóstico do Desperdício:** A API do Canvas devolve centenas de campos por aluno que nunca usamos (URLs temporárias do S3, tokens, rubricas em branco), e listar 80 alunos matriculados gasta milhares de tokens desnecessários mesmo quando o agente só precisa de 1 ou 2 alunos.
* **Solução Implementada no Servidor Go (`src/sqlite_cache.go`):**
  * **Banco SQLite Integrado (`data/afya_cache.db`):** Armazena disciplinas e estudantes com WAL mode e isolamento seguro por `course_id`.
  * **Padrão Read-Through Transparente:**
    * Na primeira requisição de qualquer professor/turma, o Go busca na API do Canvas, normaliza e salva no SQLite.
    * A partir da segunda requisição, os dados são lidos instantaneamente do disco em **0,001s**, sem bater na rede e sem risco de falha de DNS.
  * **Busca Cirúrgica por Query (`canvas_list_students` com parâmetro `query`):**
    * Se o agente quiser saber os dados do aluno Wesley, ele chama `canvas_list_students(course_id: "...", query: "wesley")`.
    * O SQLite executa `WHERE search_name LIKE '%wesley%'` e devolve **apenas o aluno correspondente** (menos de 50 tokens, em vez de despejar a turma inteira de 80 alunos no contexto).
  * **Multi-docente & Universal:** Funciona para qualquer professor, turma ou semestre, isolando por ID sem colisões.

---

### 3. 🛡️ Detecção de Plágio 100% Offline (Zero Token em Comparações)
* **Diagnóstico do Desperdício:** Submeter pares de códigos para o modelo comparar similaridade é computacionalmente caro e consome dezenas de milhares de tokens ($N \times N$ pares).
* **Solução Agressiva:**
  * A detecção já roda nativamente no MCP em Go (`canvas_detect_plagiarism`), utilizando normalização canônica de identificadores e n-gramas (k-shingles com Jaccard/Dice).
  * **Regra de Ouro:** O modelo **nunca recebe** dados de alunos com similaridade normal (< 30%). Apenas pares com índice de suspeita elevado (> 75%) são injetados no contexto, já acompanhados da classificação de risco e das linhas coincidentes.

---

### 4. 📝 Templates Parametrizados de Feedback (Slot Filling)
* **Diagnóstico do Desperdício:** Pedir ao modelo para redigir textos longos do zero para cada aluno gasta tokens de saída (*output tokens*), que são tipicamente 3x mais caros e lentos.
* **Solução Agressiva:**
  * Criar uma matriz local de rubricas institucionais (`data/templates/feedbacks.json`):
    * `EXCELENTE`: *"Código compilado com sucesso. Lógica algorítmica correta, ponteiros manipulados de forma segura e convenções atendidas."*
    * `FALTA_DESALOCACAO`: *"Lógica correta, porém ausente a liberação da memória alocada dinamicamente via free() na função {funcao}."*
    * `ERRO_SINTAXE`: *"Falha de compilação identificada na linha {linha}: {erro_resumido}."*
  * O modelo gera apenas o código da rubrica ou a linha afetada (gastando 20 tokens em vez de 300 por aluno).

---

### 5. 💾 Persistência de Estado em Disco (Manifesto de Correção)
* **Diagnóstico do Desperdício:** Em tarefas longas com várias etapas, o histórico da conversa acumula tabelas e listas repetidas. A cada nova mensagem do professor ("pode lançar", "corrija o próximo"), todo o histórico anterior é reenviado.
* **Solução Agressiva:**
  * Persistir o progresso em um arquivo local leve (ex.: `scratch/progresso_avaliacao.json`).
  * Em vez de reenviar tabelas com dezenas de linhas no chat, o assistente atualiza o manifesto local e envia apenas uma síntese compacta de status.

---

### 6. 🎯 Injeção Modular de Regras (Just-in-Time Instructions)
* **Diagnóstico do Desperdício:** Manter um arquivo de instruções monolítico de mais de 30 KB (`AGENTS.md`) consome cerca de **8.000 tokens de system prompt a cada mensagem** enviada pelo usuário, mesmo para tarefas simples como "qual a data da prova?".
* **Solução Agressiva:**
  * Manter no arquivo central `AGENTS.md` estritamente o núcleo de identidade institucional, segurança e conduta do docente.
  * Transferir manuais operacionais extensos para **Skills Específicas** carregadas sob demanda:
    * Criação e formatação de slides $\rightarrow$ skill `afya-ui-standards`
    * Fluxo de correção SpeedGrader $\rightarrow$ skill `corretor-canvas`
    * Redação de artigos e relatos $\rightarrow$ documentação em `RelatoExperiencia/`

---

### 7. 🧱 Diffs Cirúrgicos em Código (Proibição de Reescrita Integral de Arquivos)
* **Diagnóstico do Desperdício:** Ler ou salvar arquivos HTML de 2.000 linhas (como `doc/sandbox_bst/index.html` ou apresentações de slides) consome até 10.000 tokens a cada ajuste de estilo ou texto.
* **Solução Agressiva:**
  * Utilizar exclusivamente a ferramenta `replace_file_content` com fatiamento cirúrgico de linhas (`StartLine` e `EndLine` estritamente limitados ao bloco que será alterado).
  * Inspecionar apenas trechos pontuais com `view_file` (linhas 40 a 80) em vez de carregar arquivos inteiros para o contexto.

---

## 📈 Impacto Real Estimado

```
Sessão Típica de Correção de Atividade (80 Alunos):
-------------------------------------------------------------
Sem Otimizações:   ~140.000 a 180.000 tokens (Risco de estourar contexto)
Com as 7 Regras:    ~12.000 a 18.000 tokens (Redução de até ~90%)
Tempo de Execução: Cai de 4 minutos para menos de 25 segundos
-------------------------------------------------------------
```
