# 🌳 Banco de Exercícios Práticos: Árvores Binárias de Busca (BST)
**Disciplina:** Estrutura de Dados • 4º Período — Ciência da Computação  
**Docente:** Prof. Karan  
**Nível:** Introdutório / Guiado em Sala (Passo a Passo)  
**Formato:** Cada exercício contém o **código C completo e autocontido (do `#include` ao `main()`)**, pronto para copiar, colar no VS Code e rodar imediatamente no projetor.

---

## 📋 Sumário dos Exercícios
1. [Exercício 1: Somar todos os elementos da Árvore](#-exercício-1-somar-todos-os-elementos-da-árvore)
2. [Exercício 2: Contar a quantidade de Folhas](#-exercício-2-contar-a-quantidade-de-folhas)
3. [Exercício 3: Busca Booleana (Existe ou Não Existe?)](#-exercício-3-busca-booleana-existe-ou-não-existe)
4. [Exercício 4: Imprimir apenas os Valores Maiores que X](#-exercício-4-imprimir-apenas-os-valores-maiores-que-x)
5. [Exercício 5: Inspecionar Filhos Diretos da Raiz](#-exercício-5-inspecionar-filhos-diretos-da-raiz)
6. [Exercício 6: Rastreio Manual no Caderno (Sem Computador)](#-exercício-6-rastreio-manual-no-caderno-sem-computador)

---

## 🧮 Exercício 1: Somar todos os elementos da Árvore

### 🎯 Enunciado
Escreva uma função recursiva `int somarValores(No* raiz)` que percorre toda a árvore e retorna a **soma aritmética de todos os números** armazenados nela. Se a árvore estiver vazia (`NULL`), a soma é `0`.

### 💡 Raciocínio Didático para Explicar em Sala
> *"Como calcular a conta de duas mesas? O valor do nó atual (`raiz->valor`) + a soma da mesa da esquerda (`somarValores(raiz->esq)`) + a soma da mesa da direita (`somarValores(raiz->dir)`)."*

### 💻 Código Completo Pronto para Rodar (Copiar e Colar)
```c
#include <stdio.h>
#include <stdlib.h>

typedef struct No {
    int valor;
    struct No* esq;
    struct No* dir;
} No;

No* criarNo(int valor) {
    No* novo = (No*) malloc(sizeof(No));
    novo->valor = valor;
    novo->esq = NULL;
    novo->dir = NULL;
    return novo;
}

No* inserir(No* raiz, int valor) {
    if (raiz == NULL) return criarNo(valor);
    if (valor < raiz->valor) raiz->esq = inserir(raiz->esq, valor);
    else if (valor > raiz->valor) raiz->dir = inserir(raiz->dir, valor);
    return raiz;
}

// ==========================================
// FUNÇÃO DO EXERCÍCIO 1:
// ==========================================
int somarValores(No* raiz) {
    // Caso Base: se o galho é vazio, a soma é zero
    if (raiz == NULL) {
        return 0;
    }
    // Caso Recursivo: meu valor + soma da esquerda + soma da direita
    return raiz->valor + somarValores(raiz->esq) + somarValores(raiz->dir);
}

int main(void) {
    No* raiz = NULL;
    // Inserindo os valores: 50, 30, 70, 20
    raiz = inserir(raiz, 50);
    raiz = inserir(raiz, 30);
    raiz = inserir(raiz, 70);
    raiz = inserir(raiz, 20);

    // Soma esperada: 50 + 30 + 70 + 20 = 170
    int total = somarValores(raiz);

    printf("=== EXERCICIO 1: SOMA TOTAL ===\n");
    printf("Valores na arvore: [50, 30, 70, 20]\n");
    printf("Soma calculada: %d\n", total);
    printf("Resultado esperado: 170\n");

    return 0;
}
```

---

## 🍃 Exercício 2: Contar a quantidade de Folhas

### 🎯 Enunciado
Escreva uma função `int contarFolhas(No* raiz)` que conta apenas os **nós folhas** da árvore.  
*Regra de Ouro:* Um nó só é considerado folha quando ele **não possui nenhum filho** (`esq == NULL && dir == NULL`).

### 💡 Raciocínio Didático para Explicar em Sala
> *"Se o nó for NULL, retorna 0. Se os dois ponteiros dele forem NULL, achamos uma folha (retorna 1). Se tiver pelo menos um filho, continua somando as folhas da esquerda e da direita."*

### 💻 Código Completo Pronto para Rodar (Copiar e Colar)
```c
#include <stdio.h>
#include <stdlib.h>

typedef struct No {
    int valor;
    struct No* esq;
    struct No* dir;
} No;

No* criarNo(int valor) {
    No* novo = (No*) malloc(sizeof(No));
    novo->valor = valor;
    novo->esq = NULL;
    novo->dir = NULL;
    return novo;
}

No* inserir(No* raiz, int valor) {
    if (raiz == NULL) return criarNo(valor);
    if (valor < raiz->valor) raiz->esq = inserir(raiz->esq, valor);
    else if (valor > raiz->valor) raiz->dir = inserir(raiz->dir, valor);
    return raiz;
}

// ==========================================
// FUNÇÃO DO EXERCÍCIO 2:
// ==========================================
int contarFolhas(No* raiz) {
    if (raiz == NULL) {
        return 0;
    }
    // Condição de folha: sem filho na esquerda E sem filho na direita
    if (raiz->esq == NULL && raiz->dir == NULL) {
        return 1;
    }
    // Caso não seja folha, soma as folhas encontradas nos ramos
    return contarFolhas(raiz->esq) + contarFolhas(raiz->dir);
}

int main(void) {
    No* raiz = NULL;
    /*
             [ 50 ]
            /              [ 25 ]    [ 75 ]
        /         [ 10 ] [ 35 ]
     
     Folhas desta árvore: [10], [35] e [75] (Total = 3 folhas)
    */
    raiz = inserir(raiz, 50);
    raiz = inserir(raiz, 25);
    raiz = inserir(raiz, 75);
    raiz = inserir(raiz, 10);
    raiz = inserir(raiz, 35);

    int totalFolhas = contarFolhas(raiz);

    printf("=== EXERCICIO 2: CONTAR FOLHAS ===\n");
    printf("Total de folhas encontradas: %d\n", totalFolhas);
    printf("Resultado esperado: 3 folhas (nos 10, 35 e 75)\n");

    return 0;
}
```

---

## 🔍 Exercício 3: Busca Booleana (Existe ou Não Existe?)

### 🎯 Enunciado
Escreva uma função `int existe(No* raiz, int chave)` que retorna **`1` (verdadeiro)** se o número estiver na árvore ou **`0` (falso)** se ele não for encontrado.

### 💡 Raciocínio Didático para Explicar em Sala
> *"Se chegou em NULL, acabou a árvore e não achamos (retorna 0). Se o nó atual tem a chave, achamos (retorna 1). Se a chave for menor, busca só na esquerda; se for maior, busca só na direita."*

### 💻 Código Completo Pronto para Rodar (Copiar e Colar)
```c
#include <stdio.h>
#include <stdlib.h>

typedef struct No {
    int valor;
    struct No* esq;
    struct No* dir;
} No;

No* criarNo(int valor) {
    No* novo = (No*) malloc(sizeof(No));
    novo->valor = valor;
    novo->esq = NULL;
    novo->dir = NULL;
    return novo;
}

No* inserir(No* raiz, int valor) {
    if (raiz == NULL) return criarNo(valor);
    if (valor < raiz->valor) raiz->esq = inserir(raiz->esq, valor);
    else if (valor > raiz->valor) raiz->dir = inserir(raiz->dir, valor);
    return raiz;
}

// ==========================================
// FUNÇÃO DO EXERCÍCIO 3:
// ==========================================
int existe(No* raiz, int chave) {
    // Caso Base 1: árvore vazia ou chegou ao fim sem achar
    if (raiz == NULL) {
        return 0; // Falso
    }
    // Caso Base 2: achou o valor procurado!
    if (raiz->valor == chave) {
        return 1; // Verdadeiro
    }
    // Passo Recursivo: decide o lado do salto O(log n)
    if (chave < raiz->valor) {
        return existe(raiz->esq, chave);
    } else {
        return existe(raiz->dir, chave);
    }
}

int main(void) {
    No* raiz = NULL;
    raiz = inserir(raiz, 40);
    raiz = inserir(raiz, 20);
    raiz = inserir(raiz, 60);
    raiz = inserir(raiz, 10);
    raiz = inserir(raiz, 30);

    printf("=== EXERCICIO 3: BUSCA BOOLEANA ===\n");
    printf("Valores na arvore: [40, 20, 60, 10, 30]\n\n");

    int teste1 = 30;
    int teste2 = 99;

    printf("Buscando %d: %s (Esperado: SIM)\n", teste1, existe(raiz, teste1) ? "SIM (1)" : "NAO (0)");
    printf("Buscando %d: %s (Esperado: NAO)\n", teste2, existe(raiz, teste2) ? "SIM (1)" : "NAO (0)");

    return 0;
}
```

---

## 📢 Exercício 4: Imprimir apenas os Valores Maiores que X

### 🎯 Enunciado
Escreva uma função `void imprimirMaioresQue(No* raiz, int limite)` que percorre a árvore e imprime na tela **somente os números maiores** que o valor de corte `limite`, em ordem crescente.

### 💡 Raciocínio Didático para Explicar em Sala
> *"Aproveitamos a mágica do percurso Em-Ordem (*In-Order*), que já visita os números em ordem crescente natural. Basta colocar um `if (raiz->valor > limite)` antes de imprimir!"*

### 💻 Código Completo Pronto para Rodar (Copiar e Colar)
```c
#include <stdio.h>
#include <stdlib.h>

typedef struct No {
    int valor;
    struct No* esq;
    struct No* dir;
} No;

No* criarNo(int valor) {
    No* novo = (No*) malloc(sizeof(No));
    novo->valor = valor;
    novo->esq = NULL;
    novo->dir = NULL;
    return novo;
}

No* inserir(No* raiz, int valor) {
    if (raiz == NULL) return criarNo(valor);
    if (valor < raiz->valor) raiz->esq = inserir(raiz->esq, valor);
    else if (valor > raiz->valor) raiz->dir = inserir(raiz->dir, valor);
    return raiz;
}

// ==========================================
// FUNÇÃO DO EXERCÍCIO 4:
// ==========================================
void imprimirMaioresQue(No* raiz, int limite) {
    if (raiz != NULL) {
        // 1. Visita primeiro os menores à esquerda
        imprimirMaioresQue(raiz->esq, limite);

        // 2. Filtra o valor do nó atual
        if (raiz->valor > limite) {
            printf("%d ", raiz->valor);
        }

        // 3. Visita os maiores à direita
        imprimirMaioresQue(raiz->dir, limite);
    }
}

int main(void) {
    No* raiz = NULL;
    int notas[] = {50, 20, 80, 10, 30, 70, 90};
    int n = sizeof(notas) / sizeof(notas[0]);

    for (int i = 0; i < n; i++) {
        raiz = inserir(raiz, notas[i]);
    }

    printf("=== EXERCICIO 4: FILTRAR ELEMENTOS ===\n");
    printf("Todos os elementos: [10, 20, 30, 50, 70, 80, 90]\n");

    int limite = 50;
    printf("Valores maiores que %d: ", limite);
    imprimirMaioresQue(raiz, limite);
    printf("\nResultado esperado: 70 80 90\n");

    return 0;
}
```

---

## 👑 Exercício 5: Inspecionar Filhos Diretos da Raiz

### 🎯 Enunciado
Escreva uma função `void inspecionarFilhosDaRaiz(No* raiz)` que imprime na tela quem é o filho imediato da esquerda e quem é o filho imediato da direita da raiz principal.

### 💡 Raciocínio Didático para Explicar em Sala
> *"Excelente para perder o medo de acessar membros de ponteiros com a seta (`->`). Não precisa de loops nem de recursão: apenas checa se `raiz->esq != NULL` e lê o valor."*

### 💻 Código Completo Pronto para Rodar (Copiar e Colar)
```c
#include <stdio.h>
#include <stdlib.h>

typedef struct No {
    int valor;
    struct No* esq;
    struct No* dir;
} No;

No* criarNo(int valor) {
    No* novo = (No*) malloc(sizeof(No));
    novo->valor = valor;
    novo->esq = NULL;
    novo->dir = NULL;
    return novo;
}

No* inserir(No* raiz, int valor) {
    if (raiz == NULL) return criarNo(valor);
    if (valor < raiz->valor) raiz->esq = inserir(raiz->esq, valor);
    else if (valor > raiz->valor) raiz->dir = inserir(raiz->dir, valor);
    return raiz;
}

// ==========================================
// FUNÇÃO DO EXERCÍCIO 5:
// ==========================================
void inspecionarFilhosDaRaiz(No* raiz) {
    if (raiz == NULL) {
        printf("A arvore esta vazia!\n");
        return;
    }

    printf("Nó Raiz: %d\n", raiz->valor);

    if (raiz->esq != NULL) {
        printf("  <- Filho Imediato da Esquerda: %d\n", raiz->esq->valor);
    } else {
        printf("  <- Filho Imediato da Esquerda: [VAZIO / NULL]\n");
    }

    if (raiz->dir != NULL) {
        printf("  -> Filho Imediato da Direita:  %d\n", raiz->dir->valor);
    } else {
        printf("  -> Filho Imediato da Direita:  [VAZIO / NULL]\n");
    }
}

int main(void) {
    No* raiz = NULL;
    // Inserindo raiz (50) e seus dois filhos diretos
    raiz = inserir(raiz, 50);
    raiz = inserir(raiz, 20); // filho esquerdo de 50
    raiz = inserir(raiz, 80); // filho direito de 50

    printf("=== EXERCICIO 5: INSPEÇÃO DIRETA DA RAIZ ===\n");
    inspecionarFilhosDaRaiz(raiz);

    return 0;
}
```

---

## ✏️ Exercício 6: Rastreio Manual no Caderno (Sem Computador)

### 🎯 Enunciado
Peça para os alunos fecharem os notebooks por 3 minutos e resolverem no papel:
Dada a sequência de inserção:
```
[ 50,  20,  80,  10,  30,  70,  90 ]
```

1. **Desenhe a árvore no caderno.**
2. **Escreva qual será a saída do percurso Em-Ordem (*In-Order*).**
3. **Quantas folhas essa árvore possui?**

### 📝 Gabarito do Professor:
1. **Estrutura Visual:**
   ```
            [ 50 ]
           /             [ 20 ]    [ 80 ]
       /    \    /        [ 10 ] [ 30 ][ 70 ][ 90 ]
   ```
2. **Saída Em-Ordem:** `10 20 30 50 70 80 90` *(ordem crescente automática)*.
3. **Folhas:** São **4 folhas**: `10`, `30`, `70` e `90`.
