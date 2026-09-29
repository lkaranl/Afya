# Feedbacks Equilibrados e Diretos (Sem Exagero) — Atividade Git & Estoque

### 1. Equipe `ffeliper` — Nota: 90 / 100
**Alunos:** Felipe Henrique Mortari Ercolin, Hauan Lins Gonçalves, Pedro Henrique Pereira Guimarães
> Boa entrega! O programa compilou e as operações funcionaram bem no terminal. 
> Apenas dois detalhes para ficarem atentos nas próximas entregas:
> • Na divisão do Git para grupos de 3, as tarefas 4 e 5 foram colocadas na mesma branch (`feature-aluno-4-e-5`), quando o enunciado pedia uma branch individual para cada tarefa.
> • No `estoque.h`, faltou atualizar a capacidade de itens para 50 (`MAX_ITENS 50`) e definir o `ESTOQUE_MINIMO 5`.

---

### 2. Equipe `PedroMilhomens` — Nota: 85 / 100
**Alunos:** Danilo Chaves Tiago Junior, Pedro Milhomens, Weverton Henrique da Silva Cordeiro
> O projeto foi bem estruturado e compilou sem erros. 
> Dois pontos de atenção observados:
> • Na função `aplicar_desconto` (`main.c`), o retorno foi `valor_total * 0.05`. Isso calcula apenas o valor do desconto (o abatimento) e não o preço final com desconto deduzido. O correto seria retornar `valor_total - (valor_total * 0.05)`.
> • No fluxo do Git, boa parte dos commits foi feita diretamente na branch principal (`main`), sem passar pelo fluxo de branches e Pull Requests individuais por integrante.

---

### 3. Equipe `thi-fs7` — Nota: 65 / 100
**Alunos:** Daniel Vitor Maciel do Carmo, Diego Victor Pereira de Melo, Pedro Henrique Xavier de Paula Azevedo, Thiago Henrique Fonseca Silveira
> A equipe organizou bem as branches e Pull Requests no GitHub. 
> Contudo, o código final não compilou devido a um detalhe de integração:
> • No arquivo `estoque.h`, a função `listar_produtos` tenta imprimir `lista[i].categoria`, mas o campo `categoria` não foi incluído na `struct Produto` dentro do próprio `estoque.h`. Isso fez o GCC acusar erro de membro inexistente e impedir a geração do executável.

---

### 4. Equipe `CamilaBrozeguine` — Nota: 50 / 100
**Alunas:** Ana Lígia de Souza Vale, Bruna Michelly Dias Martins, Camila Brozeguine Dernei, Fernanda Santana Leandro
> A equipe iniciou bem o versionamento e entregou as duas primeiras tarefas (código de barras e categorias). 
> Porém, a entrega ficou incompleta em relação aos demais requisitos do enunciado:
> • Faltaram as Tarefas 3, 4 e 5: a correção fiscal da taxa de 10% no cálculo do total, o módulo de desconto à vista e o módulo de pagamento com juros a prazo.

---

### 5. Equipe `mariastrelow` — Nota: 20 / 100
**Aluna:** Brenda Emanuelly Dias Brunel
> O repositório foi criado e entregue no prazo.
> No entanto, o código enviado contém apenas o arquivo inicial da aula, sem a realização das 5 tarefas práticas de código e sem as branches do Git solicitadas no enunciado.

---

### 6. Equipes com Nota 100 (`Truxera`, `victrmldy`, `RodrigoCarm`, `ZeDaAreia`)
> Excelente trabalho! O repositório seguiu perfeitamente o fluxo de Git colaborativo com branches e Pull Requests, e o código em C compilou sem erros, atendendo a todos os requisitos do enunciado. Parabéns à equipe!
