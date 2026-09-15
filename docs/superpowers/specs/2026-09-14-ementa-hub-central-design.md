# Design Spec: Hub Central da Ementa de Estrutura de Dados (Canvas LMS)

- **Data:** 14/09/2026
- **Disciplina:** Estrutura de Dados (Ciência da Computação - Semestre 2026/2)
- **Docente:** Professor Karan
- **Instituição:** Afya / Centro Universitário São Lucas Ji-Paraná

---

## 1. Visão Geral e Propósito
Transformar a aba **Ementa (*Syllabus*)** da disciplina em um painel interativo, acolhedor e visualmente profissional (estilo *Hub Central*), utilizando HTML5 limpo com estilização CSS inline compatível com o sanitizador do Canvas LMS.

O objetivo é centralizar as informações pedagógicas cruciais do semestre, eliminar dúvidas recorrentes de alunos sobre composição de notas, prazos e regras institucionais, e fornecer acesso rápido com 1 clique aos livros digitais da ementa e ferramentas práticas.

---

## 2. Estrutura Visual dos Componentes (Seções)

### 2.1. Banner de Cabeçalho Institucional & Boas-Vindas
- Identidade visual com gradiente sóbrio nas cores institucionais (tons de azul marinho e ciano/teal).
- Título da disciplina, período letivo (2026/2) e identificação do docente.
- Orientações de comunicação: atendimento via Canvas Inbox e esclarecimento de dúvidas pedagógicas.

### 2.2. Mapa da Trilha de Aprendizagem (Linha do Tempo em Cards)
Organizado em cards visuais com ícones e marcadores sequenciais:
1. **Módulo 1:** Fundamentos de Memória RAM, Ponteiros e Alocação Dinâmica (`malloc`/`free`).
2. **Módulo 2:** Estruturas de Dados Lineares (Listas Simplesmente/Duplamente Encadeadas, Pilhas e Filas).
3. **Módulo 3:** Algoritmos de Ordenação Avançados e Complexidade de Algoritmos (MergeSort por Divisão e Conquista).
4. **Módulo 4:** Controle de Versão Colaborativo e Boas Práticas (Git & GitHub em Equipe).
5. **Módulo 5:** Estruturas Não-Lineares (Árvores Binárias de Busca - BST e Introdução a Grafos).

### 2.3. Painel de Composição de Notas e Critérios de Aprovação (Resolução CONSEPE 005/2024)
- **Etapa N1 (50 Pontos):**
  - Prova Escrita Individual: 30 pontos (sem consulta, modelo ENADE).
  - Atividades Teóricas e Práticas em Sala/Laboratório: 20 pontos.
- **Etapa N2 (50 Pontos):**
  - Prova Escrita Individual: 30 pontos (sem consulta, modelo ENADE).
  - Atividades Teóricas e Práticas em Sala/Laboratório: 20 pontos.
- **Total Semestral:** 100 pontos.
- **Critérios de Sucesso:**
  - Aprovação Direta: Nota Semestral $\ge$ 70 pontos e Frequência $\ge$ 75%.
  - Exame Final: Nota Semestral entre 40 e 69 pontos e Frequência $\ge$ 75%.
  - Reprovação Direta: Nota $<$ 40 pontos ou Frequência $<$ 75%.

### 2.4. Estante Digital: Bibliografia Oficial Integrada à Minha Biblioteca
Botões de acesso direto e seguro com redirecionamento de 1 clique:
- **Bibliografia Básica 1:** *Szwarcfiter & Markenzon — Estruturas de Dados e seus Algoritmos* (Editora LTC / Minha Biblioteca).
- **Bibliografia Básica 2:** *Adam Drozdek — Estrutura de Dados e Algoritmos em C++* (Cengage Learning / Minha Biblioteca).
- **Bibliografia Complementar:** *Scott Chacon & Ben Straub — Pro Git Oficial* (Versão em Português).

### 2.5. Central de Ferramentas & Práticas
- Atalho para o simulador interativo de alocação de ponteiros na memória RAM.
- Atalho para a trilha completa de Módulos da disciplina no Canvas LMS.

---

## 3. Compatibilidade Técnica Canvas LMS
- Uso exclusivo de tags HTML5 seguras aceitas pelo parser do Canvas: `div`, `table`, `tr`, `td`, `h2`, `h3`, `p`, `span`, `a`, `ul`, `li`.
- Estilização exclusivamente inline (`style="..."`), garantindo layout responsivo em celulares e computadores sem depender de folhas de estilo externas suscetíveis a bloqueio.
- Cores de alto contraste e acessibilidade (WCAG AA).
