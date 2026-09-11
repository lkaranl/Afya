# Afya Canvas MCP & Hub

Servidor **MCP (Model Context Protocol)** e painel web de alta performance desenvolvido em **Go (Golang)** para integração do Canvas LMS da Afya com **Agentes de Inteligência Artificial**.

---

## 🎯 O que esta ferramenta faz

Permite que o **Agente de IA** atue como seu assistente pedagógico de ponta a ponta:

1. **Varredura Automática:** O agente localiza automaticamente quais tarefas de todas as suas turmas possuem submissões de alunos aguardando correção e informa os prazos.
2. **Coleta Direta da API:** O agente obtém o código/texto submetido pelos alunos (ou faz o download de anexos como `.c`, `.py`, `.pdf`, `.zip`) diretamente da API do Canvas, **sem que você precise baixar ou enviar arquivos ZIP manualmente**.
3. **Avaliação Inteligente:** O agente lê o enunciado oficial do Canvas, avalia cada entrega (inclusive executando testes de código quando aplicável) e elabora nota e feedback construtivo individual.
4. **Ponto de Controle Humano:** O agente exibe uma tabela de revisão para você validar as notas propostas.
5. **Lançamento Automático de Notas:** Com o seu "OK", o agente publica todas as notas e feedbacks diretamente no livro de notas do Canvas LMS em lote.
6. **Painel Web (Opcional):** Um dashboard visual moderno continua disponível para consulta rápida no navegador com a flag `--web`.

---

## 🛠️ Ferramentas Disponíveis no MCP Server (`tools`)

| Ferramenta | Descrição |
| :--- | :--- |
| `canvas_prepare_assignment` | **Prepara em lote e concorrente**: Baixa enunciado, extrai código digitado sem HTML, clona/baixa repositórios GitHub entregues pelos alunos (extraindo snippets) e baixa anexos concorrentemente para `scratch/`. |
| `canvas_fetch_github_repo` | **Download e resumo inteligente de repositório GitHub**: Baixa repositório público do GitHub enviado pelo aluno, descompacta localmente, extrai zips internos recursivamente se houver, cataloga códigos e gera snippet de até 80 linhas / 3.500 caracteres mastigado para economia de tokens. |
| `canvas_get_grading_status` | **Diagnóstico consolidado**: Retorna status das correções da turma (% concluído, atividades 100% corrigidas, parciais, pendentes e lista de alunos sem nota). |
| `canvas_validate_grades` | **Validador de notas e gerador de tabela**: Valida regras de negócio e gera tabela em Markdown formatada para aprovação humana. |
| `canvas_unpack_zip` | **Descompactador inteligente**: Extrai arquivos ZIP do Canvas SpeedGrader e organiza por aluno. |
| `canvas_list_pending_assignments` | Varre todas as turmas do professor e retorna atividades com submissões pendentes de correção (`needs_grading_count > 0`), com prazos e notas máximas. |
| `canvas_get_submissions` | Puxa as submissões dos alunos para uma tarefa (traz `clean_body` limpo de HTML e data em horário de Brasília). |
| `canvas_download_attachment` | Baixa arquivos anexos individuais enviados pelos alunos para o disco local para análise ou testes. |
| `canvas_get_assignment` | Obtém o enunciado completo, rubrica e regras cadastradas na tarefa do Canvas. |
| `canvas_submit_grades_batch` | Publica notas e feedbacks para toda a turma de uma vez só no Canvas. |
| `canvas_submit_grade` | Lança nota e feedback para um aluno específico. |
| `canvas_list_courses` | Lista as disciplinas ativas do professor (com IDs e nomes). |
| `canvas_list_assignments` | Lista tarefas de uma disciplina com prazos e pendências. |
| `canvas_list_students` | Lista os alunos matriculados para conferência de nomes e IDs. |

---

## 💬 Como Usar no Dia a Dia com o Agente

Com o MCP Server configurado, você não precisa se preocupar com aspectos técnicos. Basta conversar naturalmente com o Agente:

### Exemplo 1: Verificar o que tem para corrigir
> **Professor:** *"Quais atividades ainda preciso corrigir?"*
> 
> **Agente:** Consulta o Canvas em tempo real e responde:
> *"Professor, você tem 5 atividades com entregas pendentes:*
> *1. **Implementação ponteiro** (Estrutura de Dados) — Prazo: 18/08 — **26 aguardando nota***
> *2. **Desafio de Lógica: Sistema de Fretes** (Estrutura de Dados) — Prazo: 01/09 — **34 aguardando nota***
> *3. **Competição Ordenação** — Prazo: 25/08 — **25 aguardando nota***
> *Qual delas deseja corrigir agora?"*

### Exemplo 2: Pedir para corrigir a atividade
> **Professor:** *"Pode corrigir a atividade de Implementação de ponteiro."*
> 
> **Agente:**
> 1. Puxa o enunciado e as submissões dos 26 alunos direto da API.
> 2. Avalia os códigos contra as regras de negócio.
> 3. Gera uma tabela para sua revisão:
> 
> | Aluno | ID | Nota Sugerida / 100 | Resumo do Feedback |
> | :--- | :--- | :--- | :--- |
> | João Silva | 122365 | 95 | Lógica de busca binária correta com ponteiro; faltou verificar ponteiro nulo. |
> 
> *"Professor, deseja ajustar alguma nota ou posso publicar no Canvas?"*
> 
> **Professor:** *"Pode publicar."*
> 
> **Agente:** Envia todas as notas e feedbacks em lote para o Canvas via `canvas_submit_grades_batch` e confirma a conclusão!

---

## ⚙️ Variáveis de Ambiente (.env)

Certifique-se de que o arquivo `.env` na raiz do projeto contenha:
```env
TOKEN=seu_token_aqui
CANVAS_BASE_URL=https://afya.instructure.com
PORT=3000
```

---

## 📦 Instalação e Uso via NPM / NPX (Recomendado)

Você e outros professores ou agentes podem rodar o servidor diretamente sem precisar ter o compilador Go instalado:

```json
{
  "mcpServers": {
    "canvas": {
      "command": "npx",
      "args": ["-y", "afya-canvas"],
      "env": {
        "TOKEN": "seu_token_canvas_aqui",
        "CANVAS_BASE_URL": "https://afya.instructure.com"
      }
    }
  }
}
```

Para abrir o painel web no navegador via NPX:
```bash
npx -y afya-canvas --web
```

---

## 🚀 Execução Local (Código-fonte Go)

### 1. Modo Padrão: Servidor MCP (para o Agente de IA)
Ao executar sem argumentos, o binário opera como servidor MCP JSON-RPC sobre `stdio`:
```bash
./afya-canvas
```

### 2. Modo Painel Web (Bônus Visual)
Para abrir o painel visual no navegador:
```bash
go run . --web
# ou:
./afya-canvas --web
```
Acesse no navegador: `http://localhost:3000`

---

## 📄 Instruções para Novos Agentes

Para garantir que qualquer novo agente de IA entenda exatamente o fluxo pedagógico e as regras do professor sem perder o contexto, consulte o arquivo [`AGENT.md`](AGENT.md).
