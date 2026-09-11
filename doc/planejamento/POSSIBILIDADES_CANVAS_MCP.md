# 🚀 Possibilidades de Integração: Canvas LMS Afya + MCP & Agentes de IA

Este documento mapeia o universo de funcionalidades possíveis e recomendadas para expansão do nosso **MCP Server em Go** integrado à **API REST oficial do Canvas LMS da Instructure** (`https://afya.instructure.com`).

---

## 🧭 Sumário Executivo
O ecossistema do Canvas LMS expõe centenas de endpoints REST que cobrem praticamente tudo o que um professor faz no navegador. Ao conectar essa API a um servidor **MCP (Model Context Protocol)** e a um **Agente de IA**, o professor deixa de ser um "operador de cliques e formulários" para se tornar um orientador pedagógico de alto nível, com um assistente automatizado cuidando de todo o trabalho operacional.

Abaixo estão divididas as possibilidades entre o que **já temos**, o que **podemos fazer imediatamente** e as **ideias mais inovadoras**.

---

## 1. 🔍 Avaliação e Correção Automatizada (Core Pedagógico)

### O que já temos:
- ✅ Varredura de atividades pendentes com prazos e notas máximas.
- ✅ Extração higienizada de códigos colados no Canvas (`clean_body`).
- ✅ Download concorrente de anexos via Goroutines.
- ✅ Download e descompactação de repositórios do GitHub com geração de snippets mastigados para economia de tokens.
- ✅ Ponto de parada com aprovação humana obrigatória e validação de limites de nota.
- ✅ Lançamento de notas e feedbacks individuais e em lote (`canvas_submit_grades_batch`).

### 💡 O que seria excelente implementar:
1. **Detecção Inteligente de Plágio e Similaridade entre Alunos:**
   - O MCP compara todos os códigos da turma entre si (usando árvores sintáticas AST para C/Python e embeddings semânticos).
   - Gera um mapa de similaridade acusando cópias diretas ou tentativas de mascaramento (troca de nomes de variáveis, reorganização de loops).
2. **Harness de Testes Automatizados em Sandbox Local:**
   - O MCP compila o código do aluno e roda uma bateria de casos de teste ocultos com entradas e saídas esperadas.
   - Mede tempo de execução (análise de complexidade O(n) vs O(n²)) e checa vazamento de memória (Valgrind em C).
   - O feedback já sai com o relatório de testes: *"Seu algoritmo passou em 8 dos 10 casos de teste. Falhou no caso de borda com vetor vazio."*
3. **Preenchimento de Rubricas Oficiais do Canvas:**
   - Em tarefas que utilizam Rubricas de Avaliação estruturadas no Canvas, o MCP pode preencher cada critério individualmente (ex: "Estrutura do Código: 20/20", "Casos de Teste: 30/40") em vez de lançar apenas nota global.
4. **Avaliação de Artigos e Relatórios Técnicos em PDF:**
   - Leitura de PDFs anexados, checagem de formatação, presença de referências bibliográficas e coerência do raciocínio com resumo executivo para o professor.

---

## 2. 📊 Analytics de Turma & Prevenção de Evasão (Alunos em Risco)

A API do Canvas fornece dados ricos de engajamento através dos endpoints de **Analytics e Enrollments**.

### 💡 O que podemos fazer:
1. **Radar de Alunos em Risco (Early Warning System):**
   - Comando: *"Agente, quem está em risco de reprovação no 2º período?"*
   - O MCP cruza:
     - Alunos que não acessam o Canvas há mais de 10 dias.
     - Alunos com 2 ou mais atividades consecutivas não entregues.
     - Alunos com média acumulada abaixo do corte (ex: < 60%).
   - Gera um relatório visual apontando exatamente onde cada estudante está com dificuldades.
2. **Engajamento e Acesso a Materiais:**
   - Verificar quais alunos sequer baixaram os slides ou abriram a página da aula antes de uma entrega.
3. **Estatísticas Avançadas de Notas:**
   - Distribuição de notas por turma (histograma, desvio padrão, mediana), identificando quais atividades foram desproporcionalmente difíceis e demandam revisão de conteúdo.

---

## 3. ✍️ Criação e Gestão de Conteúdo (Geração Automatizada)

Em vez de preencher formulários manuais no Canvas para cadastrar aulas e tarefas:

### ✅ Funcionalidades Ativas no MCP (100% Implementadas):
1. **Criação de Atividades pelo Agente via Prompt (`canvas_create_assignment`):**
   - O professor descreve em linguagem natural a atividade e o MCP gera o enunciado formatado com HTML institucional, rubrica de avaliação e cadastra direto na API do Canvas via `POST /api/v1/courses/:id/assignments`.
2. **Criação Automática de Quizzes / Questionários (`canvas_create_quiz`):**
   - Criação de questionários completos com questões de múltipla escolha, preenchimento de lacunas de código (`fill_in_multiple_blanks_question`), associação de colunas (`matching_question`) e ordenação sequencial com justificativas pedagógicas pré-configuradas.
3. **Módulos de Aula e Páginas de Conteúdo (`canvas_create_module`, `canvas_add_module_item` e `canvas_create_page`):**
   - Criação e estruturação automática de Módulos temáticos semanais com páginas de teoria, questionários e atividades práticas vinculadas em uma única chamada.
4. **Ponderação e Gestão Dinâmica de Notas (`canvas_set_course_weighting`, `canvas_list_assignment_groups`, `canvas_create_assignment_group`):**
   - Gestão de pesos da disciplina (ex: 50% Atividades Contínuas vs 50% Prova Oficial). Permite adicionar quantas atividades forem necessárias durante o semestre sem quebrar a fórmula ou precisar recalcular pesos individuais.

## 4. 💬 Comunicação e Atendimento Institucional (Inbox do Canvas)

O Canvas possui um sistema de mensagens internas (`Conversations API`).

### 💡 O que podemos fazer:
1. **Disparo de Lembretes Proativos para Quem Não Entregou:**
   - Faltando 24 horas para o prazo de uma atividade importante:
   - O MCP detecta quais alunos ainda não submeteram e dispara uma mensagem individual pelo Canvas: *"Olá [Nome], lembre-se de que a atividade de Grafos encerra amanhã às 23h59. Qualquer dúvida, estou à disposição."*
2. **Triagem do Inbox do Professor:**
   - O MCP lista as mensagens não lidas enviadas pelos alunos no Canvas, resume os assuntos e já sugere minutas de respostas técnicas para o professor aprovar.
3. **Publicação de Anúncios da Turma:**
   - Publicar avisos de aula, erratas de listas de exercícios ou orientações de provas no mural oficial do curso em segundos.

---

## 5. 👥 Gestão de Grupos e Equipes

Como vimos na atividade de Git em Equipe:

### 💡 O que podemos fazer:
1. **Sincronização e Criação de Grupos no Canvas:**
   - O MCP pode ler as equipes declaradas pelos alunos nos repositórios GitHub ou planilhas e cadastrar automaticamente os grupos dentro do Canvas (`/api/v1/group_categories/:id/groups`), alocando cada estudante no seu respectivo grupo.
2. **Replicação Inteligente de Notas de Equipe:**
   - Identificar o líder que submeteu o trabalho e aplicar a mesma avaliação e feedback proporcional para todos os integrantes matriculados daquela equipe, mesmo que a atividade esteja configurada como nota individual.

---

## 6. 🌐 Integrações Externas (O Ecossistema Afya Completo)

1. **Ponte Canvas ↔ TOTVS RM (Portal do Professor Afya):**
   - O Canvas calcula notas contínuas, mas a nota oficial de fechamento de semestre vai para o TOTVS RM.
   - O MCP pode exportar uma planilha ou formato exato do TOTVS RM com notas de N1, N2 e faltas para upload imediato, economizando horas de digitação manual de notas no fim do semestre.
2. **Notificações em Tempo Real (Telegram / Discord / WhatsApp):**
   - Um bot conectado ao MCP avisa o professor:
     - *"Novo envio na atividade de Estrutura de Dados."*
     - *"Relatório diário: 8 novas entregas aguardando correção."*
3. **Dashboard Web Centralizado (com a flag `--web`):**
   - Visualização de todas as turmas em um painel unificado em tempo real, com filtros rápidos por status de pendência, prazo e gráficos de evolução.

---

## 📋 Resumo das Próximas Ideias Recomendadas

| Prioridade | Funcionalidade | Impacto para o Professor |
| :---: | :--- | :--- |
| ⭐⭐⭐ | **Detector de Similaridade/Plágio de Código** | Identifica cópias entre alunos instantaneamente antes de dar notas. |
| ⭐⭐⭐ | **Harness de Testes de Código Automatizado** | Roda baterias de testes em C/Python direto nos códigos submetidos. |
| ⭐⭐ | **Radar de Alunos em Risco** | Notifica antes das provas quem não está entregando nada ou sumiu do Canvas. |
| ⭐⭐ | **Criador de Atividades e Quizzes via IA** | Cadastra atividades e questionários no Canvas a partir de uma conversa. |
| ⭐ | **Exportação de Notas para o TOTVS RM** | Elimina digitação manual de notas no sistema institucional no final do período. |
