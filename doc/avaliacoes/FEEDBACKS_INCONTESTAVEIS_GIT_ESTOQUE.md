# Feedbacks Pedagógicos Incontestáveis (Notas < 100) — Atividade Git & Estoque

Este documento reúne os textos exatos e cirúrgicos de devolutiva para os estudantes que obtiveram pontuação inferior a 100, apontando o arquivo, a linha exata, o trecho de código e o confronto direto com o regulamento do enunciado.

---

### 1. Equipe `ffeliper` — Nota: 90 / 100
**Alunos afetados:**
- Felipe Henrique Mortari Ercolin (ID: 322054)
- Hauan Lins Gonçalves (ID: 328249)
- Pedro Henrique Pereira Guimarães (ID: 329079)

**Texto do Feedback no Canvas:**
> A aplicação em C compilou com sucesso e as operações de cálculo de estoque tributado, desconto e juros funcionaram adequadamente na execução do programa. Contudo, a avaliação considerou dois critérios expressos no enunciado que não foram atendidos:
>
> 1. **Violação da Regra de Git para Grupos de 3 Integrantes (Desconto de 5 pontos):**
>    O enunciado estabelece expressamente na seção *"⚠️ Adaptação para Grupos Menores (4 ou 3 Alunos)"*:
>    `"cada tarefa deve ser desenvolvida em uma branch separada (1 Tarefa = 1 Branch). Nunca junte duas tarefas na mesma branch!"`
>    Para grupos de 3 alunos, o Aluno 3 deveria obrigatoriamente abrir duas branches distintas: `feature-aluno-4` e `feature-aluno-5`. A equipe agrupou ambas as implementações em uma única branch e Pull Request (`feature-aluno-4-e-5`), descumprindo a diretriz de isolamento de branches do projeto.
>
> 2. **Requisitos de Constantes Não Cumpridos em `estoque.h` (Desconto de 5 pontos):**
>    - **Capacidade do Estoque (Tarefa 1):** O enunciado exigia: *"Em estoque.h, altere a constante MAX_ITENS de 10 para 50"*. O arquivo `estoque.h` (linha 4) permaneceu definido como `#define MAX_ITENS 10`.
>    - **Estoque Mínimo (Tarefa 2):** O enunciado solicitava: *"Adicione a constante #define ESTOQUE_MINIMO 5 logo abaixo de MAX_ITENS"*. Essa constante não foi declarada no cabeçalho.

---

### 2. Equipe `PedroMilhomens` — Nota: 85 / 100
**Alunos afetados:**
- Danilo Chaves Tiago Junior (ID: 318445)
- Pedro Milhomens (ID: 313673)
- Weverton Henrique da Silva Cordeiro (ID: 314053)

**Texto do Feedback no Canvas:**
> O projeto foi estruturado em C e compilou sem erros no GCC. No entanto, houve duas inconsistências técnicas em relação aos requisitos solicitados:
>
> 1. **Erro de Regra de Negócio na Função `aplicar_desconto` (Desconto de 10 pontos):**
>    No arquivo `main.c` (linhas 34 a 36), a função foi implementada da seguinte forma:
>    ```c
>    float aplicar_desconto(float valor_total) {
>        return valor_total * 0.05;
>    }
>    ```
>    A função retorna apenas o **valor do abatimento** (5%) e não o **valor final a pagar com o desconto aplicado**. Como consequência, ao selecionar a opção 3 no menu ("Exibir total com desconto a vista"), se o total for R$ 200,00, o programa exibe incorretamente `Total: R$ 10,00`, em vez de `R$ 190,00`. O correto seria: `return valor_total * (1 - TAXA_DESCONTO);` ou `return valor_total - (valor_total * 0.05);`.
>
> 2. **Fluxo de Trabalho Descentralizado do Git (Desconto de 5 pontos):**
>    A maior parte das modificações das Tarefas 1, 3 e 4 foi commitada diretamente na branch principal (`main`) pelo líder da equipe, sem a criação formal das branches correspondentes e sem a abertura/revisão dos Pull Requests individuais por cada integrante, conforme exigido nas diretrizes de trabalho colaborativo em equipe.

---

### 3. Equipe `thi-fs7` — Nota: 65 / 100
**Alunos afetados:**
- Daniel Vitor Maciel do Carmo (ID: 320096)
- Diego Victor Pereira de Melo (ID: 316435)
- Pedro Henrique Xavier de Paula Azevedo (ID: 339091)
- Thiago Henrique Fonseca Silveira (ID: 324578)

**Texto do Feedback no Canvas:**
> A equipe organizou adequadamente as branches e Pull Requests no GitHub para um grupo de 4 integrantes. Contudo, o código entregue na branch principal apresentou uma falha crítica que **impede a compilação do programa**:
>
> 1. **Erro Fatal de Compilação no GCC (Desconto de 30 pontos):**
>    No arquivo `estoque.h` (linhas 19 a 26), foi inserida a implementação da função `listar_produtos`:
>    ```c
>    void listar_produtos(Produto lista[], int total) {
>        ...
>        printf("ID: %d | Categoria: %s | Código: %s | Nome: %s | Preco: R$ %.2f | Qtd: %d\n",
>               lista[i].id, lista[i].categoria, lista[i].codigo_barras, ...);
>    }
>    ```
>    O código tenta acessar o campo `lista[i].categoria`. Entretanto, na definição da `struct Produto` (linhas 10 a 16 do mesmo arquivo `estoque.h`), o campo `categoria` **não foi declarado**:
>    ```c
>    typedef struct {
>        int id;
>        char codigo_barras[20]; 
>        char nome[30];
>        float preco;
>        int quantidade;
>    } Produto; // Faltou declarar: char categoria[20];
>    ```
>    Ao tentar compilar com o comando padrão da disciplina (`gcc main.c`), o compilador aborta imediatamente com o erro:
>    `error: no member named 'categoria' in 'Produto'`.
>
> 2. **Duplicidade de Código e Formatação (Desconto de 5 pontos):**
>    A função `listar_produtos` foi mantida também dentro de `main.c` (linha 15) com a versão antiga sem quebra de linha (`\n`), gerando inconsistência estrutural entre o cabeçalho e o arquivo principal.

---

### 4. Equipe `CamilaBrozeguine` — Nota: 50 / 100
**Alunas afetadas:**
- Ana Lígia de Souza Vale (ID: 313910)
- Bruna Michelly Dias Martins (ID: 251846)
- Camila Brozeguine Dernei (ID: 326476)
- Fernanda Santana Leandro (ID: 308416)

**Texto do Feedback no Canvas:**
> A equipe realizou a configuração do repositório e integrou as Tarefas 1 e 2 via branches e Pull Requests. No entanto, a entrega foi submetida de forma **parcial e incompleta**, deixando de atender a mais da metade dos requisitos avaliativos da atividade:
>
> 1. **Ausência da Tarefa 3 — Correção Fiscal e Tributos (Desconto de 15 pontos):**
>    - Em `estoque.h` (linha 6), a constante permaneceu com a alíquota antiga `#define TAXA_PADRAO 0.05` (o enunciado exigia a atualização para `0.10`).
>    - Em `main.c` (linha 25), a função `calcular_total` permaneceu com o cálculo original que soma apenas os preços unitários (`soma += lista[i].preco`), sem multiplicar pela quantidade (`preco * quantidade`) e sem aplicar a taxa tributária sobre o total.
>
> 2. **Ausência da Tarefa 4 — Módulo de Pagamento à Vista (Desconto de 15 pontos):**
>    - Não foi criada a constante `TAXA_DESCONTO 0.05`.
>    - Não foi implementada a função `aplicar_desconto`.
>    - Não foi adicionada a opção `3 - Exibir total com desconto a vista` no menu e nem o respectivo `case 3:` no `switch`.
>
> 3. **Ausência da Tarefa 5 — Módulo de Pagamento a Prazo (Desconto de 15 pontos):**
>    - Não foi criada a constante `TAXA_JUROS 0.08`.
>    - Não foi implementada a função `aplicar_juros`.
>    - Não foi adicionada a opção `4 - Exibir total a prazo (com juros)` no menu e nem o respectivo `case 4:` no `switch`.
>
> 4. **Git e Branches Faltantes (Desconto de 5 pontos):**
>    O repositório possui apenas 4 commits e parou na branch `feature-aluno-2`, deixando de criar as branches individuais `feature-aluno-3`, `feature-aluno-4` e `feature-aluno-5` e seus respectivos Pull Requests.

---

### 5. Equipe `mariastrelow` — Nota: 20 / 100
**Aluna afetada:**
- Brenda Emanuelly Dias Brunel (ID: 363018)

**Texto do Feedback no Canvas:**
> A pontuação atribuída (20 pontos) refere-se exclusivamente à criação do repositório no GitHub e envio do link dentro do prazo regulamentar.
>
> Em relação ao desenvolvimento técnico da atividade prática, **nenhum dos requisitos solicitados foi implementado**:
> 1. **Código Estático / Sem Implementações:** O repositório possui apenas 1 commit ("Primeiro commit") que contém exatamente o template inicial fornecido pelo professor, mantendo inclusive os comentários de erro originais (`// BUG: esqueceram de imprimir o ID...`).
> 2. **Nenhuma das 5 Tarefas Realizada:** Não foram implementados o código de barras, a categorização com estoque mínimo, a correção do cálculo tributário de 10%, o módulo de desconto à vista (5%) e o módulo de juros a prazo (8%).
> 3. **Ausência de Prática com Git:** Não foi criada nenhuma branch funcional (`feature-aluno-X`) e nenhum Pull Request foi aberto ou mesclado na branch principal.
