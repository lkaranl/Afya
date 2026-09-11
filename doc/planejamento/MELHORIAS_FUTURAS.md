# 🚀 Melhorias Futuras & Roadmap (Afya Canvas Hub)

Este documento registra ideias, oportunidades de integração e melhorias arquiteturais identificadas para as próximas versões do assistente e da plataforma.

---

## 📡 No Radar: Integração com o Portal do Professor (TOTVS RM Educacional)

### 🎯 Objetivo
Permitir a visualização, auditoria e cruzamento de **faltas e presença do diário de classe oficial** da Afya diretamente no ecossistema do assistente, integrando o ambiente de notas (Canvas LMS) com o sistema acadêmico institucional (TOTVS RM).

### 🔍 Contexto Técnico Identificado
- **Sistema Acadêmico:** O Portal do Professor da Afya é sustentado pelo ERP **TOTVS RM (Linha RM / TOTVS Educacional)**.
- **Armazenamento de Frequência:** As faltas e presenças oficiais ficam nas tabelas corporativas `SFREQUENCIA` (frequência diária) e `SNOTAETAPA` (totalizadores por etapa letiva).
- **Desafio de API:** Diferente do Canvas LMS (que disponibiliza geração de tokens pessoais via perfil do usuário), as APIs REST e Web Services (DataServers) do TOTVS RM são controladas pela TI central corporativa da Afya, sem liberação de credenciais de integração pública para docentes.

### 🛠️ Estratégias de Implementação Propostas
1. **Automação Web / Web Scraping Autenticado (Caminho Recomendado):**
   - Criação de um módulo em Python/Go que realiza login automatizado com as credenciais do professor na URL do Portal do Professor da unidade.
   - Extração estruturada do diário de classe (aulas ministradas, presenças e faltas por aluno).
   - Disponibilização desses dados como uma nova ferramenta MCP: `canvas_get_attendance` ou `portal_get_attendance`.
2. **Parser de Diário de Classe (Importação de Planilhas):**
   - Suporte à importação manual de relatórios em `.csv` ou `.xlsx` exportados do Portal do Professor.
   - Cruzamento automático dos nomes e matrículas do Portal com os IDs e e-mails do Canvas.
3. **Consumo de APIs Internas (caso haja acesso):**
   - Mapeamento das chamadas REST que a interface web do portal realiza no front-end para replicação de requisições de sessão.

---

## 📋 Outras Melhorias em Backlog

- [ ] **Geração de Relatórios Pedagógicos Automatizados:** Geração de gráficos e resumos de rendimento por turma para reuniões de coordenação pedagógica (identificação precoce de alunos com risco de evasão ou baixo desempenho).
- [ ] **Integração de Notificações:** Envio de alertas ou resumos diários de pendências via bot de Telegram/Discord ou e-mail.
- [ ] **Suporte a Múltiplas Linguagens no Testador Automático:** Expandir o `scripts/test_c_submissions.py` para suportar testes automáticos e execução protegida em Python, Java e JavaScript.
