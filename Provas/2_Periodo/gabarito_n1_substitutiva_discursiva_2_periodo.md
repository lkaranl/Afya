# Gabarito e Espelho de Correção — Avaliação N1 Substitutiva (Discursiva Avançada)
**Disciplina:** Estrutura de Dados &bull; Ciência da Computação (2º Período) &bull; Afya São Lucas Ji-Paraná  
**Professor:** Me. Karan  
**Valor Total:** 30,0 Pontos (4 Questões Discursivas de 7,5 pontos cada)  
**Conteúdo Programático:** 100% embasado nos módulos ministrados na etapa N1 (Módulo 1: Ponteiros e Memória RAM; Módulo Bônus: Lógica Proposicional e Tabela Verdade; Módulo Bônus: Git & GitHub).

---

## QUESTÃO 01 (7,5 pts) — Rastreamento Avançado de Ponteiros, Vetores e Apontamento Múltiplo na RAM

### Subitem a) (4,0 pts) — Teste de Mesa Minucioso e Saída no Console

- **Evolução do Teste de Mesa instrução por instrução:**
  1. `int vet[4] = {5, 15, 25, 35};`  
     Vetor inicial: `vet[0]=5`, `vet[1]=15`, `vet[2]=25`, `vet[3]=35`.
  2. `int *p1 = &vet[0];`  
     `p1` armazena o endereço de `vet[0]`.
  3. `int *p2 = &vet[2];`  
     `p2` armazena o endereço de `vet[2]`.
  4. `int *p3 = p1;`  
     `p3` recebe o mesmo endereço de `p1`, apontando também para `vet[0]`.
  5. `*p3 = *p2 - *p1;`  
     `*p2` vale 25 e `*p1` vale 5. Operação: $25 - 5 = 20$.  
     Como `*p3` aponta para `vet[0]`, atualiza: `vet[0] = 20`.  
     *Estado atual do vetor:* `{20, 15, 25, 35}`.
  6. `p1 = p1 + 1;`  
     Aritmética de ponteiro: `p1` avança para o elemento seguinte, passando a apontar para `&vet[1]`.
  7. `*p1 = (*p3) * 2;`  
     `*p3` aponta para `vet[0]`, cujo valor atualizado é 20.  
     `*p1` aponta para `vet[1]`. Operação: $20 \times 2 = 40$.  
     Atualiza: `vet[1] = 40`.  
     *Estado atual do vetor:* `{20, 40, 25, 35}`.
  8. `p3 = p2;`  
     `p3` passa a apontar para o mesmo endereço de `p2` (`&vet[2]`).
  9. `*p2 = *(p1 - 1) + *(p3);`  
     `p1` aponta para `&vet[1]`, logo `(p1 - 1)` aponta para `&vet[0]`, cujo valor é 20.  
     `*(p3)` aponta para `vet[2]`, cujo valor atual é 25.  
     Operação: $20 + 25 = 45$.  
     Como `*p2` aponta para `vet[2]`, atualiza: `vet[2] = 45`.  
     (Nota: como `p3` também aponta para `vet[2]`, agora `*p3` também reflete 45).  
     *Estado atual do vetor:* `{20, 40, 45, 35}`.
  10. `p2 = &vet[3];`  
      `p2` passa a apontar para o último elemento (`&vet[3]`).
  11. `*p2 = *p1 + *p3;`  
      `*p1` aponta para `vet[1]` (valor 40).  
      `*p3` aponta para `vet[2]` (valor 45).  
      Operação: $40 + 45 = 85$.  
      Como `*p2` aponta para `vet[3]`, atualiza: `vet[3] = 85`.  
      *Estado final do vetor:* `{20, 40, 45, 85}`.

- **Saída Exata no Console:**
  ```text
  vet: [20, 40, 45, 85]
  ```

- **Critérios de Correção do Subitem a:**
  - Demonstração do teste de mesa com rastreamento dos endereços e valores intermediários: **2,5 pts**.
  - Saída exata impressa pelo `printf`: **1,5 pt** (0,375 pt por posição correta no array).

---

### Subitem b) (3,5 pts) — Análise Conceitual de Atribuição e Limites de Memória

1. **Diferença entre `p3 = p2;` e `*p3 = *p2;` (1,75 pt):**
   - **`p3 = p2;` (Atribuição de Endereços):** Modifica o registrador/célula de memória da variável ponteiro `p3`, copiando o endereço que está guardado em `p2`. A partir dessa linha, ambos passam a referenciar a mesma gaveta na RAM (`vet[2]`). O conteúdo da posição apontada permanece inalterado.
   - **`*p3 = *p2;` (Atribuição de Conteúdo / Desreferenciação):** Não altera para onde `p3` aponta. O operador `*` acessa o valor inteiro guardado no endereço apontado por `p2` e o grava na célula de memória apontada por `p3`. Ou seja, modifica o dado útil na RAM mantendo os ponteiros em seus endereços originais.

2. **Impacto de `p1 = p1 + 5;` seguido de `*p1 = 10;` (1,75 pt):**
   - O array `vet` possui tamanho 4 (índices válidos de 0 a 3). Ao somar 5, `p1` é deslocado para uma região de memória fora do bloco reservado para o vetor na Stack (equivalente a uma posição fictícia `vet[5]`).
   - Ao executar `*p1 = 10;`, o programa tenta gravar dados em uma área de memória não alocada para aquela variável.
   - **Consequências em nível de sistema/hardware:**
     - **Corrupção de Memória:** Pode sobrescrever outras variáveis locais da função ou registradores de controle da pilha (como o *stack frame* ou endereço de retorno).
     - **Violação de Segmento (*Segmentation Fault*):** Se o endereço ultrapassar o segmento autorizado do processo, a Unidade de Gerenciamento de Memória (MMU) do processador detecta o acesso ilegal e o sistema operacional aborta o programa imediatamente.
     - Caracteriza o perigo clássico de **ponteiro solto / invasão de memória (*buffer overflow*)**.

---

## QUESTÃO 02 (7,5 pts) — Escopo de Funções, Passagem por Referência Cruzada e Pilha (Stack)

### Subitem a) (4,0 pts) — Rastreamento do Teste de Mesa e Saídas no Console

- **Valores Iniciais na `main()`:**
  `x = 4`, `y = 6`, `z = 8`, `w = 2`.

#### 1. Rastreamento da Chamada 1: `calcular(&x, y, &z, &w);`
- **Mapeamento de parâmetros na Stack de `calcular`:**
  - `a = &x` (ponteiro para `x` da `main`)
  - `b = y = 6` (cópia por valor de `y`)
  - `c = &z` (ponteiro para `z` da `main`)
  - `d = &w` (ponteiro para `w` da `main`)
- **Instruções da função:**
  - `temp = *a + b = 4 + 6 = 10;`
  - `*a = *c * 2 = 8 * 2 = 16;` $\implies$ atualiza `x = 16` na `main`.
  - `*c = temp - b = 10 - 6 = 4;` $\implies$ atualiza `z = 4` na `main`.
  - `b = b * 3 = 6 * 3 = 18;` $\implies$ altera apenas a variável local `b` na Stack da função.
  - `*d = *a + *c + b = 16 + 4 + 18 = 38;` $\implies$ atualiza `w = 38` na `main`.
- **Valores na `main()` após Chamada 1:**
  - `x = 16`, `y = 6` (inalterado, pois foi passado por valor), `z = 4`, `w = 38`.
- **Saída C1:**
  ```text
  C1: x=16, y=6, z=4, w=38
  ```

#### 2. Rastreamento da Chamada 2 (Cruzada): `calcular(&x, x, &y, &z);`
- **Valores na `main()` imediatamente antes da chamada:**
  `x = 16`, `y = 6`, `z = 4`, `w = 38`.
- **Mapeamento de parâmetros na Stack de `calcular`:**
  - `a = &x` (ponteiro para `x` da `main`)
  - `b = x = 16` (cópia por valor do valor atual de `x`)
  - `c = &y` (ponteiro para `y` da `main`)
  - `d = &z` (ponteiro para `z` da `main`)
- **Instruções da função:**
  - `temp = *a + b = 16 + 16 = 32;`
  - `*a = *c * 2`: `*c` aponta para `y` (que vale 6). Logo: $6 \times 2 = 12 \implies$ atualiza `x = 12` na `main`.
  - `*c = temp - b = 32 - 16 = 16;` $\implies$ `*c` aponta para `y`, logo atualiza `y = 16` na `main`.
  - `b = b * 3 = 16 * 3 = 48;` $\implies$ altera apenas a variável local `b` na Stack.
  - `*d = *a + *c + b = 12 + 16 + 48 = 76;` $\implies$ `*d` aponta para `z`, logo atualiza `z = 76` na `main`.
- **Valores na `main()` após Chamada 2:**
  - `x = 12`, `y = 16`, `z = 76`, `w = 38` (não foi envolvido na Chamada 2, permaneceu 38).
- **Saída C2:**
  ```text
  C2: x=12, y=16, z=76, w=38
  ```

- **Critérios de Correção do Subitem a:**
  - Teste de mesa da Chamada 1: **1,0 pt**.
  - Saída exata de `C1`: **1,0 pt**.
  - Teste de mesa da Chamada 2: **1,0 pt**.
  - Saída exata de `C2`: **1,0 pt**.

---

### Subitem b) (3,5 pts) — Isolamento de Parâmetros na Stack e Passagem por Valor vs. Referência

- **Comportamento da Pilha de Execução (Stack):**
  - No momento em que `calcular(&x, x, &y, &z)` é chamada, o compilador cria um novo registro de ativação (*stack frame*) para a função.
  - Para o segundo parâmetro `b`, o valor atual de `x` (16) é **duplicado/copiado** para uma nova posição de memória local da pilha pertencente a `b`.
  - Para o primeiro parâmetro `a`, a pilha recebe apenas o endereço de memória de `x` (`&x`).
- **Por que a alteração `*a = *c * 2;` não afetou `b` durante a mesma execução:**
  - `*a` modifica diretamente a variável original `x` no quadro de memória da `main()`.
  - No entanto, `b` é uma variável completamente independente que reside em outro endereço na Stack da função. Modificar o endereço apontado por `*a` não altera a cópia local `b`, que continua com seu valor armazenado (16).
  - Isso demonstra o princípio fundamental do encapsulamento e isolamento de escopo por valor na linguagem C.

---

## QUESTÃO 03 (7,5 pts) — Lógica Proposicional, Tabela Verdade de 8 Linhas e Tautologia

### Subitem a) (4,5 pts) — Tabela Verdade Completa com 8 Linhas ($2^3 = 8$)

Proposição: $[ (P \to Q) \land (Q \to R) ] \to (P \to R)$

| Linha | P | Q | R | P &rarr; Q | Q &rarr; R | (P &rarr; Q) &and; (Q &rarr; R) | P &rarr; R | Coluna Final |
| :---: | :-: | :-: | :-: | :---: | :---: | :---: | :---: | :---: |
| **1** | V | V | V | **V** | **V** | **V** | **V** | **V** |
| **2** | V | V | F | **V** | **F** | **F** | **F** | **V** |
| **3** | V | F | V | **F** | **V** | **F** | **V** | **V** |
| **4** | V | F | F | **F** | **V** | **F** | **F** | **V** |
| **5** | F | V | V | **V** | **V** | **V** | **V** | **V** |
| **6** | F | V | F | **V** | **F** | **F** | **V** | **V** |
| **7** | F | F | V | **V** | **V** | **V** | **V** | **V** |
| **8** | F | F | F | **V** | **V** | **V** | **V** | **V** |

- **Detalhamento das Regras Aplicadas:**
  - Condicional ($\to$): resulta em **F** unicamente quando o antecedente é Verdadeiro e o consequente é Falso ($V \to F = F$). Nos demais casos, resulta em **V**.
  - Conjunção ($\land$): resulta em **V** unicamente quando ambos os operandos são Verdadeiros.
  - Implicação Final: como a premissa conjunta só é Verdadeira nas linhas 1, 5, 7 e 8, e em todas essas linhas a conclusão $P \to R$ também é Verdadeira ($V \to V = V$); e nas linhas 2, 3, 4 e 6 a premissa é Falsa ($F \to \dots = V$), todas as 8 linhas da coluna final resultam em **V**.

- **Critérios de Correção do Subitem a:**
  - Enumeração correta das 8 combinações de $P, Q, R$: **1,0 pt**.
  - Preenchimento correto das 4 colunas intermediárias: **2,0 pts** (0,5 pt por coluna).
  - Preenchimento correto da coluna final: **1,5 pt**.

---

### Subitem b) (3,0 pts) — Classificação Formal e Conectivos em C

1. **Classificação Formal (1,5 pt):**
   - A proposição é uma **Tautologia**.
   - **Justificativa:** Uma fórmula lógica é classificada como tautológica quando seu valor lógico é estritamente **Verdadeiro** para todas as atribuições de verdade possíveis de suas variáveis constituintes (todas as 8 linhas da tabela verdade resultam em V). Representa a validade formal do princípio da transitividade lógica (Silogismo Hipotético).

2. **Representação dos Conectivos em C e Equivalência da Condicional (1,5 pt):**
   - **Operadores em C:**
     - Conjunção ($\land$): operador `&&`
     - Disjunção ($\lor$): operador `||`
     - Negação ($\neg$): operador `!`
   - **Equivalência $P \to Q \equiv \text{!P || Q}$:**
     - A condicional $P \to Q$ só é Falsa no caso em que $P$ é Verdadeiro e $Q$ é Falso ($V \to F = F$). Em todos os outros casos, é Verdadeira.
     - A expressão `!P || Q` produz rigorosamente a mesma tabela-verdade:
       - Se $P$ for Falso ($F$), `!P` torna-se Verdadeiro ($V$), e pela regra do `||` o resultado é imediatamente Verdadeiro independentemente de $Q$.
       - Se $P$ for Verdadeiro ($V$), `!P` é Falso ($F$), logo o resultado da expressão dependerá unicamente de $Q$ ser Verdadeiro.
     - Como possuem a mesma tabela de valoração semântica, são logicamente equivalentes.

---

## QUESTÃO 04 (7,5 pts) — Controle de Versão, Ciclo de Vida do Git e Resolução de Conflito em Código C

### Subitem a) (3,5 pts) — Análise da Saída do `git status`, Áreas Locais e Comandos

1. **Localização dos arquivos nas Áreas Locais do Git (1,5 pt):**
   - **`operacoes.c`:** Encontra-se na **Staging Area (Index / Área de Preparação)**, pois a saída do terminal indica claramente a seção `Changes to be committed:` com o arquivo listado em verde como `modified: operacoes.c`.
   - **`teste.c`:** Encontra-se exclusivamente no **Working Directory (Diretório de Trabalho)**, pois a saída do terminal lista-o sob a seção `Untracked files:`, indicando que o Git ainda não monitora nem incluiu o arquivo no índice.

2. **Diferença Conceitual e Comandos de Terminal (2,0 pts):**
   - **Diferença entre *Staged* e *Untracked* (1,0 pt):**
     - **Staged (Changes to be committed):** O arquivo já teve suas modificações preparadas e indexadas através de `git add`. O Git já tirou um snapshot temporário e essas alterações farão parte do próximo commit a ser registrado no histórico.
     - **Untracked files:** Arquivo novo presente no diretório de trabalho físico que nunca foi adicionado ao versionamento. O Git não rastreia suas mudanças nem criará histórico para ele até que seja explicitamente adicionado.
   - **Comandos exatos de terminal (1,0 pt):**
     - (I) Para consolidar `operacoes.c` no repositório local:
       ```bash
       git commit -m "Ajusta operacoes"
       ```
     - (II) Para começar a rastrear `teste.c`:
       ```bash
       git add teste.c
       ```

---

### Subitem b) (4,0 pts) — Anatomia dos Marcadores, Unificação de Código C e Fluxo de Push

1. **Significado Técnico dos Marcadores de Conflito (1,5 pt):**
   - **`<<<<<<< HEAD` (0,5 pt):** Delimita o início do bloco com as modificações locais que já existiam no branch de trabalho atual na máquina do desenvolvedor (feitas antes do `git pull`).
   - **`=======` (0,5 pt):** Linha divisória neutra injetada pelo Git que separa as alterações locais das alterações recebidas do repositório remoto.
   - **`>>>>>>> 8f3c4e2a1b9d` (0,5 pt):** Delimita o término do bloco de alterações concorrentes vindas do commit remoto (com o respectivo hash do commit que causou a divergência).

2. **Código C Final Unificado e Sequência de Comandos no Terminal (2,5 pts):**
   - **Código C Final Limpo da função `processar` (1,5 pt):**
     O acadêmico deve remover obrigatoriamente todas as linhas com os marcadores de conflito (`<<<<<<<`, `=======`, `>>>>>>>`) e unificar as regras de negócio:
     ```c
     int processar(int valor) {
         if (valor < 0) {
             return 0;
         }
         return valor * 3 + 10;
     }
     ```
     *(Também é aceita a sintaxe compacta com operador ternário: `return (valor < 0) ? 0 : valor * 3 + 10;`).*
   - **Sequência exata de comandos no terminal (1,0 pt):**
     1. Adicionar o arquivo corrigido na Staging Area para avisar ao Git que o conflito foi resolvido:
        ```bash
        git add operacoes.c
        ```
     2. Concluir o commit de merge:
        ```bash
        git commit -m "Resolve conflito de mesclagem na funcao processar"
        ```
     3. Enviar a versão final unificada ao GitHub:
        ```bash
        git push origin main
        ```
        *(ou simplesmente `git push`).*

---

## 📊 Matriz de Pontuação e Resumo Pedagógico

| Questão | Tema Central (Módulo da Disciplina) | Peso Total | Subitem a | Subitem b |
| :---: | :--- | :---: | :---: | :---: |
| **01** | Ponteiros, Aritmética e Vetores na RAM (Módulo 1) | **7,50** | 4,00 pts | 3,50 pts |
| **02** | Escopo, Passagem por Referência/Valor e Stack (Módulo 1) | **7,50** | 4,00 pts | 3,50 pts |
| **03** | Lógica Proposicional, Tabela Verdade e C (Módulo Bônus) | **7,50** | 4,50 pts | 3,00 pts |
| **04** | Git & GitHub: 3 Áreas, Push e Conflitos (Módulo Bônus) | **7,50** | 3,50 pts | 4,00 pts |
| **TOTAL** | **Avaliação N1 Substitutiva Discursiva** | **30,00** | **16,00 pts** | **14,00 pts** |
