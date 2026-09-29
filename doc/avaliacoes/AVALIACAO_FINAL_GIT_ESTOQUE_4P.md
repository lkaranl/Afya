# Relatório Oficial de Avaliação com Adaptação por Quantidade de Integrantes
**Disciplina:** Estrutura de Dados — 4º Período (2026.2)  
**Atividade:** ATIVIDADE PRÁTICA DE GIT EM EQUIPE: SISTEMA DE CONTROLE DE ESTOQUE  
**Data da Avaliação:** 21/09/2026 (Pós-prazo oficial das 18:59)  
**Total de Alunos Matriculados:** 56  
**Total de Submissões Avaliadas:** 31 estudantes (distribuídos em 9 repositórios/equipes)  
**Escala:** 0 a 100 pontos

---

## 📌 Regras de Adaptação por Quantidade de Integrantes (Conforme Enunciado Oficial)

O enunciado da atividade estabelece regras de ouro específicas para dimensionamento da equipe:

1. **Equipes de 5 Alunos (Padrão):**
   - 1 Aluno por Tarefa (Aluno 1: Tarefa 1, Aluno 2: Tarefa 2, Aluno 3: Tarefa 3, Aluno 4: Tarefa 4, Aluno 5: Tarefa 5).
   - 5 branches individuais (`feature-aluno-1` a `feature-aluno-5`) e 5 Pull Requests separados.
2. **Equipes de 4 Alunos:**
   - **Aluno 1:** Tarefa 1 (`feature-aluno-1`).
   - **Aluno 2:** Tarefa 2 (`feature-aluno-2`).
   - **Aluno 3:** Tarefa 3 (`feature-aluno-3`).
   - **Aluno 4:** Assume Tarefas 4 e 5, **obrigatoriamente criando 2 branches separadas** (`feature-aluno-4` e `feature-aluno-5`) e **2 Pull Requests distintos**. Nunca juntar as duas tarefas na mesma branch!
3. **Equipes de 3 Alunos:**
   - **Aluno 1:** Tarefa 1 (`feature-aluno-1`) e Tarefa 3 (`feature-aluno-3`).
   - **Aluno 2:** Tarefa 2 (`feature-aluno-2`).
   - **Aluno 3:** Tarefa 4 (`feature-aluno-4`) e Tarefa 5 (`feature-aluno-5`).
   - *Obrigatoriamente 1 branch individual para cada tarefa desenvolvida*.

---

## 👥 Análise Detalhada das 9 Equipes à Luz da Regra de Integrantes

### 1. Equipe `Truxera` (5 Integrantes) — NOTA 100
- **Integrantes:** Chrystian Manuel, Guilherme Pereira, Hilion de Paula, Luiz Gabriel, Renan de Lima.
- **Auditoria de Integrantes:** 5 alunos, 5 branches separadas (`feature-aluno-1` a `feature-aluno-5`), 6 Pull Requests mesclados na `main`.
- **Código C:** Compilação limpa com GCC C99. Funções `calcular_total` (com 10% e preço × quantidade), `aplicar_desconto` e `aplicar_juros` 100% corretas.
- **Resultado:** **100 / 100**.

---

### 2. Equipe `victrmldy` (5 Integrantes) — NOTA 100
- **Integrantes:** Carlos Victor, João Pedro Souza, Lucas Clameirich, Pedro Paulo Soares, Ryan Freitas.
- **Auditoria de Integrantes:** 5 alunos, 25 commits, branches e PRs individuais para cada membro.
- **Código C:** Compilação limpa sem erros. Todas as 5 tarefas implementadas com precisão matemática.
- **Resultado:** **100 / 100**.

---

### 3. Equipe `RodrigoCarm` (3 Integrantes) — NOTA 100 (Exemplo Perfeito de Grupo de 3)
- **Integrantes:** João Guilherme Fernandes, Pedro Henrique Guilherme, Rodrigo Vasconcelos.
- **Auditoria de Integrantes:** 
  - Rodrigo (Aluno 1) assumiu Tarefa 1 (`feature-aluno-1`) e Tarefa 3 (`feature-aluno-3`);
  - Pedro H (Aluno 2) assumiu Tarefa 2 (`feature-aluno-2`);
  - João Guilherme (Aluno 3) assumiu Tarefa 4 (`feature-aluno-4`) e Tarefa 5 (`feature-aluno-5`).
  - **Seguiu 100% à risca a regra do enunciado:** 5 branches separadas e 5 Pull Requests individuais, mesmo sendo apenas 3 alunos!
- **Código C:** Compilação limpa com GCC C99 e lógica impecável.
- **Resultado:** **100 / 100**.

---

### 4. Equipe `ZeDaAreia` (3 Integrantes) — NOTA 100
- **Integrantes:** Gabriel Figueiredo, Kauan Henrique, Paulo Justus Iurk Neto.
- **Auditoria de Integrantes:** Grupo de 3 integrantes. Gabriel assumiu Tarefa 3, Paulo Iurk assumiu Tarefa 2, David Alves / ZeDaAreia assumiram Tarefa 4 e 5.
- **Código C:** Compilação limpa. A falha da linha 67 (`qtd`) foi corrigida às 16h antes do encerramento do prazo, com cálculo de tributo regularizado.
- **Resultado:** **100 / 100**.

---

### 5. Equipe `ffeliper` (3 Integrantes) — NOTA 90
- **Integrantes:** Felipe Henrique, Hauan Lins, Pedro Henrique Pereira.
- **Auditoria de Integrantes:** Grupo de 3 integrantes.
  - **Violação da Regra de Grupos Menores:** O enunciado determinava expressamente: *"Caso o grupo não tenha exatamente 5 integrantes, a regra de ouro se mantém: cada tarefa deve ser desenvolvida em uma branch separada (1 Tarefa = 1 Branch). Nunca junte duas tarefas na mesma branch!"*. O grupo unificou as tarefas 4 e 5 na mesma branch `feature-aluno-4-e-5` em um único PR.
  - **Requisito Não Atendido:** O cabeçalho permaneceu com `MAX_ITENS 10` e sem a constante `ESTOQUE_MINIMO 5` (Tarefa 1 e 2).
- **Código C:** Compilou com sucesso e as fórmulas matemáticas funcionam.
- **Resultado:** **90 / 100**.

---

### 6. Equipe `PedroMilhomens` (3 Integrantes) — NOTA 85
- **Integrantes:** Danilo Chaves, Pedro Milhomens, Weverton Henrique.
- **Auditoria de Integrantes:** Grupo de 3 integrantes.
  - **Violação da Regra de Branches:** O aluno Pedro Milhomens concentrou a maior parte dos commits das tarefas 1, 3 e 4 diretamente na branch principal, reduzindo o uso dos PRs descentralizados previstos para 3 membros.
  - **Falha de Lógica:** Na função `aplicar_desconto` (`main.c`, linha 35), retornou apenas o valor do desconto (`valor_total * 0.05`) em vez do valor total a pagar com desconto (`valor_total * 0.95`).
- **Código C:** Compila com sucesso.
- **Resultado:** **85 / 100**.

---

### 7. Equipe `thi-fs7` (4 Integrantes) — NOTA 65
- **Integrantes:** Daniel Vitor, Diego Victor, Pedro Henrique Xavier, Thiago Henrique.
- **Auditoria de Integrantes:** Grupo de 4 integrantes.
  - **Divisão de Tarefas:** Thiago (Tarefa 1), Diego Victor (Tarefa 2), Pedro Henrique (Tarefas 4 e 5 com branches separadas `feature-aluno-4` e `feature-aluno-5` — parabéns por essa divisão correta).
  - **Falha Crítica de Compilação:** No arquivo `estoque.h` (linha 23), foi inserida uma cópia da função `listar_produtos` que tenta imprimir `lista[i].categoria`. Contudo, o campo `categoria` não foi declarado na `struct Produto` dentro do cabeçalho, travando a compilação do GCC (`error: no member named 'categoria' in 'Produto'`).
  - **Cálculo da Média:** Em `main.c` (linha 19), a função original foi mantida duplicada sem quebra de linha.
- **Resultado:** **65 / 100**.

---

### 8. Equipe `CamilaBrozeguine` (4 Integrantes) — NOTA 50
- **Integrantes:** Ana Lígia, Bruna Michelly, Camila Brozeguine, Fernanda Santana.
- **Auditoria de Integrantes:** Grupo de 4 integrantes.
  - **Entrega Incompleta de Tarefas:** O grupo realizou apenas 4 commits e parou na Tarefa 2. Não foram realizadas as Tarefas 3, 4 e 5 (faltaram 3 das 5 branches obrigatórias, sem implementação de tributos de 10%, desconto à vista de 5% e juros de 8%).
- **Resultado:** **50 / 100**.

---

### 9. Equipe `mariastrelow` (1 Integrante) — NOTA 20
- **Integrante:** Brenda Emanuelly Dias Brunel.
- **Auditoria de Integrantes:** Aluna isolada. Submeteu repositório com 1 único commit template com o código inicial da aula, sem nenhuma das 5 tarefas implementadas e sem nenhuma branch criada.
- **Resultado:** **20 / 100**.

---

## 📋 Tabela Final Conferida para Lançamento Oficial (31 Alunos)

| ID Aluno | Nome do Estudante | Equipe | Qtd Integrantes | Nota | Motivo Resumido |
| :---: | :--- | :---: | :---: | :---: | :--- |
| **301806** | Chrystian Manuel Fernandez Leon | `Truxera` | 5 | **100** | Padrão 5 membros: 5 branches, 6 PRs, C 100% funcional. |
| **312802** | Guilherme Pereira Godinho | `Truxera` | 5 | **100** | Padrão 5 membros: 5 branches, 6 PRs, C 100% funcional. |
| **322055** | Hilion de Paula Hermsdorf | `Truxera` | 5 | **100** | Padrão 5 membros: 5 branches, 6 PRs, C 100% funcional. |
| **325045** | Luiz Gabriel Alves da Silva | `Truxera` | 5 | **100** | Padrão 5 membros: 5 branches, 6 PRs, C 100% funcional. |
| **343620** | Renan de Lima Souza | `Truxera` | 5 | **100** | Padrão 5 membros: 5 branches, 6 PRs, C 100% funcional. |
| **317335** | Carlos Victor Estevam Lenk | `victrmldy` | 5 | **100** | Padrão 5 membros: 5 branches, PRs individuais, 100% funcional. |
| **312804** | João Pedro Souza Santos | `victrmldy` | 5 | **100** | Padrão 5 membros: 5 branches, PRs individuais, 100% funcional. |
| **310560** | Lucas Clameirich da Silveira | `victrmldy` | 5 | **100** | Padrão 5 membros: 5 branches, PRs individuais, 100% funcional. |
| **306878** | Pedro Paulo Soares | `victrmldy` | 5 | **100** | Padrão 5 membros: 5 branches, PRs individuais, 100% funcional. |
| **320091** | Ryan Freitas Lima | `victrmldy` | 5 | **100** | Padrão 5 membros: 5 branches, PRs individuais, 100% funcional. |
| **122365** | João Guilherme Fernandes Rosa | `RodrigoCarm` | 3 | **100** | Grupo de 3 exemplar: 5 branches separadas conforme enunciado. |
| **123828** | Pedro Henrique Guilherme Pereira | `RodrigoCarm` | 3 | **100** | Grupo de 3 exemplar: 5 branches separadas conforme enunciado. |
| **89497** | Rodrigo Vasconcelos do Carmo | `RodrigoCarm` | 3 | **100** | Grupo de 3 exemplar: 5 branches separadas conforme enunciado. |
| **105164** | Gabriel do Nascimento Figueiredo | `ZeDaAreia` | 3 | **100** | Grupo de 3: corrigiu bug da linha 67 antes das 18h; compila 100%. |
| **122935** | Kauan Henrique Barbosa Peres | `ZeDaAreia` | 3 | **100** | Grupo de 3: corrigiu bug da linha 67 antes das 18h; compila 100%. |
| **92085** | Paulo Justus Iurk Neto | `ZeDaAreia` | 3 | **100** | Grupo de 3: corrigiu bug da linha 67 antes das 18h; compila 100%. |
| **322054** | Felipe Henrique Mortari Ercolin | `ffeliper` | 3 | **90** | Grupo de 3: unificou tarefas 4 e 5 na mesma branch (descumpriu regra). |
| **328249** | Hauan Lins Gonçalves | `ffeliper` | 3 | **90** | Grupo de 3: unificou tarefas 4 e 5 na mesma branch (descumpriu regra). |
| **329079** | Pedro Henrique Pereira Guimarães | `ffeliper` | 3 | **90** | Grupo de 3: unificou tarefas 4 e 5 na mesma branch (descumpriu regra). |
| **318445** | Danilo Chaves Tiago Junior | `PedroMilhomens` | 3 | **85** | Grupo de 3: desconto retorna só 5%; mesclagem sem branches de tarefas. |
| **313673** | Pedro Milhomens | `PedroMilhomens` | 3 | **85** | Grupo de 3: desconto retorna só 5%; mesclagem sem branches de tarefas. |
| **314053** | Weverton Henrique da Silva Cordeiro | `PedroMilhomens` | 3 | **85** | Grupo de 3: desconto retorna só 5%; mesclagem sem branches de tarefas. |
| **320096** | Daniel Vitor Maciel do Carmo | `thi-fs7` | 4 | **65** | Grupo de 4: erro fatal em estoque.h:23 (categoria ausente na struct). |
| **316435** | Diego Victor Pereira de Melo | `thi-fs7` | 4 | **65** | Grupo de 4: erro fatal em estoque.h:23 (categoria ausente na struct). |
| **339091** | Pedro Henrique Xavier de Paula | `thi-fs7` | 4 | **65** | Grupo de 4: erro fatal em estoque.h:23 (categoria ausente na struct). |
| **324578** | Thiago Henrique Fonseca Silveira | `thi-fs7` | 4 | **65** | Grupo de 4: erro fatal em estoque.h:23 (categoria ausente na struct). |
| **313910** | Ana Lígia de Souza Vale | `CamilaBrozeguine` | 4 | **50** | Grupo de 4: incompleto, só entregou tarefas 1 e 2 (sem 3, 4 e 5). |
| **251846** | Bruna Michelly Dias Martins | `CamilaBrozeguine` | 4 | **50** | Grupo de 4: incompleto, só entregou tarefas 1 e 2 (sem 3, 4 e 5). |
| **326476** | Camila Brozeguine Dernei | `CamilaBrozeguine` | 4 | **50** | Grupo de 4: incompleto, só entregou tarefas 1 e 2 (sem 3, 4 e 5). |
| **308416** | Fernanda Santana Leandro | `CamilaBrozeguine` | 4 | **50** | Grupo de 4: incompleto, só entregou tarefas 1 e 2 (sem 3, 4 e 5). |
| **363018** | Brenda Emanuelly Dias Brunel | `mariastrelow` | 1 | **20** | Aluna individual: 1 único commit com o template sem tarefas. |
