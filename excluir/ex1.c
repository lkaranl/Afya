#include <stdio.h>
#include <stdlib.h>

typedef struct No {
  int valor;
  struct No *esq;
  struct No *dir;
} No;

No *criarNo(int valor) {
  No *novo = (No *)malloc(sizeof(No));
  novo->valor = valor;
  novo->esq = NULL;
  novo->dir = NULL;
  return novo;
}

No *inserir(No *raiz, int valor) {
  if (raiz == NULL)
    return criarNo(valor);
  if (valor < raiz->valor)
    raiz->esq = inserir(raiz->esq, valor);
  else if (valor > raiz->valor)
    raiz->dir = inserir(raiz->dir, valor);
  return raiz;
}

// ==========================================
// FUNÇÃO DO EXERCÍCIO 1:
// ==========================================
int somarValores(No *raiz) {
  // Caso Base: se o galho é vazio, a soma é zero
  if (raiz == NULL) {
    return 0;
  }
  // Caso Recursivo: meu valor + soma da esquerda + soma da direita
  return raiz->valor + somarValores(raiz->esq) + somarValores(raiz->dir);
}

int main(void) {
  No *raiz = NULL;
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

  return 0;
}
