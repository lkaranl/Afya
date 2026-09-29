# Relatório de Avaliação Prévia (Rascunho): Atividade Prática de Git em Equipe
**Disciplina:** Estrutura de Dados — 4º Período (2026.2)  
**Atividade:** ATIVIDADE PRÁTICA DE GIT EM EQUIPE: SISTEMA DE CONTROLE DE ESTOQUE  
**Data da Avaliação:** 21/09/2026  
**Total de Envios Analisados nesta rodada:** 12 alunos (distribuídos em 6 equipes/repositórios)  
**Escala Oficial:** 0 a 100 pontos (Componente Prático de N1)

---

## 📌 Critérios Oficiais Utilizados na Correção
1. **Compilação e Execução em C (GCC C99 sem erros):**
   - Implementação das funções `listar_produtos`, `calcular_total` (com tributo de 10% e multiplicação correta de preço × quantidade), `aplicar_desconto` (5%) e `aplicar_juros` (8%).
   - Estrutura `Produto` com `id`, `nome`, `preco`, `quantidade`, `codigo_barras` e `categoria`.
2. **Uso do Git & GitHub em Equipe:**
   - Criação de branches separadas para cada funcionalidade (`feature-aluno-X`).
   - Abertura de Pull Requests e merge na `main`.
   - Histórico de commits identificando individualmente a contribuição de cada membro.

---

## 👥 Resumo Executivo das Equipes

| Equipe / Repositório | Integrantes Avaliados nesta Rodada | Compilação (GCC) | Branches & PRs | Nota Sugerida |
| :--- | :--- | :---: | :---: | :---: |
| **Equipe 1 (`Truxera/atividade-git`)** | Chrystian Manuel, Guilherme Pereira, Hilion de Paula, Luiz Gabriel, Renan de Lima | ✅ Sucesso (100%) | ✅ Excelente (6 PRs, merges limpos) | **100 / 100** |
| **Equipe 2 (`PedroMilhomens/...`)** | Pedro Milhomens | ✅ Sucesso (100%) | ⚠️ Branches mescladas pelo líder / Desconto divergente | **85 / 100** |
| **Equipe 3 (`ffeliper/atividade-git`)** | Felipe Henrique, Hauan Lins, Pedro Henrique Pereira | ✅ Sucesso (100%) | ⚠️ Branches 4 e 5 unificadas / `MAX_ITENS` mantido em 10 | **90 / 100** |
| **Equipe 4 (`ZeDaAreia/Aula`)** | Paulo Justus Iurk Neto | ❌ Falha (Erro de compilação na linha 67) | ✅ PRs criados por membros distintos | **65 / 100** |
| **Equipe 5 (`thi-fs7/...`)** | Pedro Henrique Xavier de Paula | ❌ Falha (Erro de compilação em `estoque.h:23`) | ✅ PRs criados por branches separadas | **65 / 100** |
| **Equipe 6 (`CamilaBrozeguine/...`)** | Ana Lígia de Souza Vale | ✅ Sucesso parcial (Apenas tarefas 1 e 2) | ⚠️ Faltaram tarefas 3, 4 e 5 | **50 / 100** |

---

## 🔍 Avaliação Detalhada por Aluno e Feedbacks Pedagógicos

### 🔹 Equipe 1 — Repositório `Truxera/atividade-git`
*Repositório completo com 22 commits, 6 Pull Requests integrados, 5 branches individuais (`feature-aluno-1` a `feature-aluno-5`) e compilação limpa com GCC C99.*

#### 1. Chrystian Manuel Fernandez Leon (ID Canvas: 301806)
- **Nota Sugerida:** **100 / 100**
- **Feedback Proposto:**
> Parabéns pela entrega! O repositório da equipe seguiu com rigor as boas práticas de desenvolvimento colaborativo com Git. Todas as branches individuais (`feature-aluno-1` a `feature-aluno-5`) foram criadas separadamente, com Pull Requests revisados e integrados à branch principal. O código em C compilou sem avisos e cumpriu integralmente todos os requisitos: estrutura de dados expandida com categoria e código de barras, cálculo do total em estoque com o tributo de 10% (preço × quantidade), e as opções de pagamento à vista (desconto de 5%) e a prazo (juros de 8%) funcionando perfeitamente no menu.

#### 2. Guilherme Pereira Godinho (ID Canvas: 312802)
- **Nota Sugerida:** **100 / 100**
- **Feedback Proposto:**
> *(Mesmo feedback da Equipe 1 — Desempenho integral atendido).*

#### 3. Hilion de Paula Hermsdorf (ID Canvas: 322055)
- **Nota Sugerida:** **100 / 100**
- **Feedback Proposto:**
> *(Mesmo feedback da Equipe 1 — Desempenho integral atendido).*

#### 4. Luiz Gabriel Alves da Silva (ID Canvas: 325045)
- **Nota Sugerida:** **100 / 100**
- **Feedback Proposto:**
> *(Mesmo feedback da Equipe 1 — Desempenho integral atendido).*

#### 5. Renan de Lima Souza (ID Canvas: 343620)
- **Nota Sugerida:** **100 / 100**
- **Feedback Proposto:**
> *(Mesmo feedback da Equipe 1 — Desempenho integral atendido).*

---

### 🔹 Equipe 2 — Repositório `PedroMilhomens/atividade-git-equipe-karan`

#### 6. Pedro Milhomens (ID Canvas: 313673)
- **Nota Sugerida:** **85 / 100**
- **Apontamento Técnico de Erros:**
  1. **Fórmula de Desconto em `main.c` (linha 35):** A função `aplicar_desconto` foi implementada retornando apenas o valor do abatimento (`return valor_total * 0.05;`), em vez de retornar o valor final a pagar com o desconto deduzido (`valor_total - (valor_total * 0.05)`). Dessa forma, ao selecionar a opção 3, o sistema exibe apenas os 5% e não o total a pagar.
  2. **Fluxo do Git / Branches:** Boa parte das integrações e commits das tarefas 1 a 5 foi realizada diretamente pelo próprio usuário Pedro Milhomens na branch principal, reduzindo a evidência da descentralização dos Pull Requests em equipe.
- **Feedback Proposto:**
> O projeto foi estruturado e compilou com sucesso em C, cobrindo a expansão dos produtos, cálculo de estoque tributado e menu interativo. Contudo, foram identificados dois pontos de atenção:
> 1. Na função `aplicar_desconto` (`main.c`, linha 35), o retorno foi implementado como `valor_total * 0.05`, retornando apenas a fração do desconto em vez do preço final à vista com o desconto aplicado (`valor_total * 0.95`).
> 2. No fluxo de Git, parte considerável das alterações foi commitada centralizadamente na branch principal, em detrimento do uso estrito de Pull Requests separados por cada membro da equipe.

---

### 🔹 Equipe 3 — Repositório `ffeliper/atividade-git`

#### 7. Felipe Henrique Mortari Ercolin (ID Canvas: 322054)
- **Nota Sugerida:** **90 / 100**
- **Apontamento Técnico de Erros:**
  1. **Capacidade Máxima em `estoque.h` (linha 4):** A Tarefa 1 exigia expandir a capacidade de produtos de `MAX_ITENS 10` para `MAX_ITENS 50`. O grupo manteve `MAX_ITENS 10`.
  2. **Fluxo de Branches Git:** O grupo unificou as tarefas 4 e 5 em uma única branch `feature-aluno-4-e-5`, descumprindo a orientação explícita do enunciado de manter 1 Tarefa = 1 Branch mesmo em grupos com menos membros.
- **Feedback Proposto:**
> A aplicação compilou perfeitamente e as operações matemáticas de cálculo de estoque com tributo (10%), desconto à vista (5%) e juros (8%) foram implementadas de maneira correta e modular. Como oportunidades de melhoria:
> 1. No arquivo `estoque.h` (linha 4), a capacidade de itens permaneceu como `MAX_ITENS 10`, quando o requisito da Tarefa 1 solicitava a expansão para 50 itens.
> 2. No fluxo do Git, as tarefas 4 e 5 foram agrupadas na mesma branch (`feature-aluno-4-e-5`), quando o regulamento solicitava branches estritamente separadas por funcionalidade para evitar conflitos de integração.

#### 8. Hauan Lins Gonçalves (ID Canvas: 328249)
- **Nota Sugerida:** **90 / 100**
- **Feedback Proposto:**
> *(Mesmo feedback da Equipe 3).*

#### 9. Pedro Henrique Pereira Guimarães (ID Canvas: 329079)
- **Nota Sugerida:** **90 / 100**
- **Feedback Proposto:**
> *(Mesmo feedback da Equipe 3).*

---

### 🔹 Equipe 4 — Repositório `ZeDaAreia/Aula`

#### 10. Paulo Justus Iurk Neto (ID Canvas: 92085)
- **Nota Sugerida:** **65 / 100**
- **Apontamento Técnico de Erros:**
  1. **Erro de Compilação (GCC) em `main.c` (linha 67):**
     ```c
     case 4: {
         float total = calcular_total(estoque, qtd); // ERRO: identificador 'qtd' não declarado
         printf("Total a prazo: R$ %.2f\n", aplicar_juros(total));
         break;
     }
     ```
     O compilador interrompe a compilação com `error: use of undeclared identifier 'qtd'`. A variável correta no escopo da `main` é `total_produtos`.
  2. **Cálculo de Estoque Não Tributado em `main.c` (linha 27):** A função `calcular_total` soma apenas o preço unitário `soma += lista[i].preco`, sem multiplicar pela quantidade (`lista[i].quantidade`) e sem aplicar a `TAXA_PADRAO` de 10%.
- **Feedback Proposto:**
> A equipe demonstrou boa articulação no Git, com histórico organizado de Pull Requests e branches individuais por membro. No entanto, o código submetido apresentou erro crítico de compilação:
> 1. Em `main.c` (linha 67), a chamada `calcular_total(estoque, qtd)` utiliza a variável inexistente `qtd`, impedindo a compilação do programa (`use of undeclared identifier 'qtd'`). O correto seria utilizar `total_produtos`.
> 2. A função `calcular_total` (`main.c`, linha 27) não multiplicou o preço pela quantidade de cada item e não aplicou a taxa tributária de 10% prevista na Tarefa 3.

---

### 🔹 Equipe 5 — Repositório `thi-fs7/Sistema-de-Controle-de-Estoque`

#### 11. Pedro Henrique Xavier de Paula Azevedo (ID Canvas: 339091)
- **Nota Sugerida:** **65 / 100**
- **Apontamento Técnico de Erros:**
  1. **Erro de Compilação (GCC) em `estoque.h` (linha 23):**
     O arquivo de cabeçalho `estoque.h` declarou e tentou acessar `lista[i].categoria`, porém o campo `categoria` não foi inserido na definição da `struct Produto` dentro do mesmo `estoque.h`. O compilador falha com: `error: no member named 'categoria' in 'Produto'`.
  2. **Cálculo do Estoque em `main.c` (linha 27):** A função `calcular_total` soma apenas o preço simples de cada produto, deixando de multiplicar pela quantidade em estoque e sem aplicar o tributo de 10%.
- **Feedback Proposto:**
> A equipe estruturou adequadamente os Pull Requests no GitHub, demonstrando compreensão do fluxo de trabalho colaborativo. Porém, houve falha crítica de integração no código C:
> 1. No arquivo `estoque.h` (linha 23), a função tenta imprimir `lista[i].categoria`, mas o campo `categoria` não foi declarado na `struct Produto`, gerando erro fatal de compilação (`no member named 'categoria' in 'Produto'`).
> 2. Em `main.c` (linha 27), a função `calcular_total` acumula apenas o preço unitário (`soma += lista[i].preco`), sem considerar a quantidade de itens e sem a incidência da `TAXA_PADRAO` de 10%.

---

### 🔹 Equipe 6 — Repositório `CamilaBrozeguine/...`

#### 12. Ana Lígia de Souza Vale (ID Canvas: 313910)
- **Nota Sugerida:** **50 / 100**
- **Apontamento Técnico de Erros:**
  1. **Entrega Incompleta de Tarefas:** O repositório possui apenas 4 commits e implementou apenas até a Tarefa 2 (`feature-aluno-2`).
  2. **Ausência das Tarefas 3, 4 e 5:**
     - Não foi feita a correção fiscal da taxa de 10% e o cálculo de quantidade × preço (Tarefa 3);
     - Não foi implementado o módulo de pagamento à vista com desconto de 5% (Tarefa 4);
     - Não foi implementado o módulo de pagamento a prazo com juros de 8% (Tarefa 5).
- **Feedback Proposto:**
> A equipe realizou a configuração inicial do Git e implementou adequadamente a inclusão de categorias e estoque mínimo (Tarefas 1 e 2). Entretanto, a entrega foi submetida de forma parcial:
> 1. Não foram implementadas as opções de pagamento à vista com desconto (Tarefa 4) e pagamento a prazo com juros (Tarefa 5).
> 2. A função `calcular_total` (`main.c`, linha 30) permaneceu com a lógica original que não multiplica o preço pela quantidade e não aplica a taxa tributária de 10% (Tarefa 3).
