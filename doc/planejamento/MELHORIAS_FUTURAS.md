# 🚀 Propostas de Melhorias & Roadmap Estratégico (Afya Canvas Hub)

Este documento compila uma nova seleção de melhorias pedagógicas, tecnológicas e de usabilidade para o ecossistema do **Afya Canvas Assistant**, baseadas na vivência docente do Professor Karan e nas diretrizes oficiais da Afya / Centro Universitário São Lucas Ji-Paraná (Resolução CONSEPE 005/2024 & NAPED 2026).

---

## 🧭 Pilares do Roadmap

```
           ┌───────────────────────────────────────────────┐
           │        AFYA CANVAS HUB - ROADMAP 2.0          │
           └──────┬──────────────┬──────────────┬──────────┘
                  │              │              │
        ┌─────────┴──────┐ ┌─────┴────────┐ ┌───┴──────────┐
        │ 1. Pedagogia & │ │ 2. Experiência│ │ 3. Automação │
        │    Avaliação   │ │   & Interface │ │   & Sistemas │
        └────────────────┘ └──────────────┘ └──────────────┘
```

---

## 1. 🧠 Inteligência Pedagógica & Avaliação Avançada

### 1.1. Integração Direta com Rubricas Oficiais do Canvas (SpeedGrader Rubrics)
- **Problema:** O Canvas permite vincular matrizes de avaliação por critérios (rubricas), mas preenchê-las manualmente para 50+ alunos é demorado.
- **Melhoria:** Extrair as rubricas configuradas na atividade (ex: *Lógica/Algoritmo: 40%*, *Estrutura de Dados/Ponteiros: 30%*, *Boas Práticas: 20%*, *Indentação/Comentários: 10%*) e preencher nota e comentário individual por critério via API do Canvas (`/api/v1/courses/:id/assignments/:id/submissions/:id/rubric_assessments`).
- **Impacto:** Transparência total para o estudante leigo, eliminando questionamentos sobre a composição da nota.

### 1.2. Detector de Código Gerado por IA Generativa (ChatGPT / GitHub Copilot)
- **Problema:** Estudantes iniciantes utilizam LLMs para resolver listas de exercícios e entregam códigos que fogem completamente do vocabulário ensinado no 2º ou 4º período.
- **Melhoria:** Cruzador estilístico que identifica anomalias:
  - Uso de bibliotecas ou funções avançadas não ministradas em aula (ex: `qsort`, `bsearch`, estruturas de STL em C básico).
  - Padrões de comentários excessivamente acadêmicos ou em inglês perfeito.
  - Alerta de "Complexidade Incompatível com o Nível da Turma".
- **Impacto:** Permite ao professor chamar o estudante para explicar a lógica em sala antes da nota ser homologada.

### 1.3. Gerador em Lote de Questões no Modelo ENADE (com Distratores Explicados)
- **Problema:** Criar questões conceituais contextualizadas com situação-problema e 4 distratores justificados (exigência do Art. 5º da Resolução CONSEPE 005/2024) exige muito tempo.
- **Melhoria:** Ferramenta MCP (`canvas_generate_enade_question`) que recebe o tema da aula (ex: *Árvores AVL*, *Ponteiros Duplos*, *Recursão*) e gera:
  - Texto-base com contexto real de mercado/engenharia.
  - Situação-problema clara.
  - Gabarito fundamentado + 4 alternativas incorretas com justificativa pedagógica do erro.
  - Exportação direta para Questionário do Canvas via API ou pacote QTI.

### 1.4. Alerta Regimental de Cumprimento de Prazos de Devolutiva (10 Dias)
- **Problema:** O Art. 16 § 2º do CONSEPE estipula prazo improrrogável de até 10 dias após a aplicação para devolutiva e entrega das notas/provas aos estudantes.
- **Melhoria:** Monitor automático que rastreia a data de aplicação/entrega e exibe cronômetro regressivo com alertas visuais:
  - 🟢 1 a 5 dias: No prazo regular.
  - 🟡 6 a 8 dias: Alerta moderado de fechamento de correções.
  - 🔴 9 a 10 dias: Alerta crítico regimental de devolução obrigatória.

---

## 2. 💻 Interface Web, Usabilidade & Comunicação

### 2.1. "Mail Merge" no Inbox do Canvas (Disparo em Lote Personalizado)
- **Problema:** Enviar mensagens individuais para 15 alunos em risco de evasão pelo Canvas consome muito tempo manual.
- **Melhoria:** Selecionar os alunos identificados no Radar de Evasão na tela web e clicar em *"Enviar Alerta de Busca Ativa"*. O sistema envia mensagens individuais no Inbox com interpolação de tags:
  > *"Olá, {primeiro_nome}! Notei que você não acessa nossa disciplina de {disciplina} há {dias_inativo} dias e ainda temos a atividade {tarefa_pendente} em aberto. Estou à disposição para tirar dúvidas no laboratório ou aqui pelo Canvas. Não deixe acumular!"*

### 2.2. Exportação de Relatórios Executivos em PDF e Excel para o NAPED / Coordenação
- **Problema:** Reuniões de colegiado e intervenções pedagógicas do NAPED demandam relatórios formais consolidados.
- **Melhoria:** Botão de exportação no painel web que gera:
  - **PDF Executivo Formatado:** Com cabeçalho da Afya / Centro Universitário São Lucas, resumo estatístico da turma e lista nominal dos alunos críticos.
  - **Planilha Excel (.xlsx):** Com abas separadas para *Notas por Atividade*, *Histograma de Rendimento* e *Radar de Inatividade*.

### 2.3. Dashboard Analítico Visual com Gráficos de Desempenho
- **Problema:** Dados em tabela são ótimos para conferência, mas gráficos facilitam a percepção rápida da turma.
- **Melhoria:** Inclusão de gráficos interativos (via Chart.js leve nativo) na aba web:
  - Distribuição de notas (curva de rendimento N1 e N2).
  - Linha do tempo de engajamento semanal da turma.
  - Gráfico de pizza de estudantes: Regulares vs. Atenção vs. Moderado vs. Crítico.

### 2.4. Histórico de Conversas e Sessões no Chat Web
- **Problema:** Ao recarregar a página do chat web, o histórico da conversa atual é reiniciado.
- **Melhoria:** Armazenamento local (IndexedDB ou SQLite em disco) com abas laterais de conversas anteriores (ex: *"Correção Lista 1"*, *"Radar de Evasão 4º Período"*).

---

## 3. ⚙️ Engenharia de Testes & Execução Segura

### 3.1. Runner Multi-Linguagem em Sandbox (Python, Java, TypeScript/JS)
- **Problema:** O avaliador automatizado atual foca prioritariamente em código C (`test_c_submissions.py`).
- **Melhoria:** Criar executores isolados com timeout estrito e restrição de recursos para outras disciplinas do curso de Ciência da Computação e Sistemas de Informação:
  - Python (`pytest` / análise estática `ruff`).
  - Java (`javac` + execução com timeout).
  - Algoritmos e Complexidade (medição de tempo de execução e casos de teste com vetores grandes).

### 3.2. Versionamento e Snapshot de Submissões dos Alunos no Git
- **Problema:** O SpeedGrader do Canvas às vezes sofre com sobrescrita de re-submissões tardias de alunos.
- **Melhoria:** Comando que cria um repositório Git local em `scratch/archive/` onde cada atividade vira uma branch e cada aluno gera um commit com timestamp exato da entrega, permitindo diffs instantâneos entre entregas sucessivas do mesmo aluno.

---

## 4. 🏢 Integrações Corporativas Afya (TOTVS RM & Sistemas Externos)

### 4.1. Cruzamento de Presença do Diário de Classe Oficial (TOTVS RM Educacional)
- **Problema:** As faltas e o controle do teto de 25% de ausências (critério de reprovação direta por frequência) ficam no Portal do Professor (TOTVS RM), enquanto o engajamento fica no Canvas.
- **Melhoria:** Módulo de importação/scraping do Diário de Classe que cruza faltas do RM com notas do Canvas:
  - Identifica o aluno com 68 pontos no Canvas, mas com 26% de faltas no RM (reprovação regimental por falta).
  - Identifica o aluno que não vai à aula presencial, mas faz entregas perfeitas no Canvas (caso para investigação de plágio/terceirização).

### 4.2. Bot de Notificações de Pendências (Telegram / WhatsApp / Discord)
- **Problema:** O professor precisa entrar no sistema para saber se novos alunos postaram atividades atrasadas ou enviaram dúvidas.
- **Melhoria:** Disparo de briefing matinal automático:
  - *"Bom dia, Professor Karan! Na turma de Estrutura de Dados (4º Período), 4 novas tarefas foram entregues e aguardam nota. Você tem 1 mensagem não lida no Inbox."*

---

## 📊 Matriz de Priorização Sugerida (Esforço x Valor Docente)

| Funcionalidade | Valor para o Professor | Esforço Técnico | Prioridade Recomendada |
| :--- | :---: | :---: | :---: |
| **Mail Merge no Inbox do Canvas** (Avisos individuais em lote) | ⭐⭐⭐⭐⭐ | Baixo (1-2 dias) | **Imediata (Sprint 1)** |
| **Exportação PDF / Excel para NAPED** (Relatórios com 1 clique) | ⭐⭐⭐⭐⭐ | Baixo (1-2 dias) | **Imediata (Sprint 1)** |
| **Alerta Regimental de 10 Dias CONSEPE** (Prazos de Devolutiva) | ⭐⭐⭐⭐ | Muito Baixo (1 dia) | **Imediata (Sprint 1)** |
| **Gerador de Questões ENADE com Distratores** | ⭐⭐⭐⭐⭐ | Médio (2-3 dias) | **Curto Prazo (Sprint 2)** |
| **Gráficos Visuais de Rendimento no Web** (Chart.js) | ⭐⭐⭐⭐ | Médio (2 dias) | **Curto Prazo (Sprint 2)** |
| **Detector de Código com Suspeita de IA** (ChatGPT/Copilot) | ⭐⭐⭐⭐ | Médio (3 dias) | **Médio Prazo (Sprint 3)** |
| **Runner Multi-Linguagem** (Python, Java) | ⭐⭐⭐ | Médio (3 dias) | **Médio Prazo (Sprint 3)** |
| **Cruzamento de Frequência TOTVS RM** | ⭐⭐⭐⭐⭐ | Alto (Pesquisa / Auth) | **Longo Prazo** |
