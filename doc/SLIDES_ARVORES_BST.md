---
marp: true
theme: gaia
_class: lead
paginate: true
backgroundColor: #090d16
color: #f8fafc
style: |
  section {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    padding: 40px;
  }
  h1 { color: #38bdf8; }
  h2 { color: #94a3b8; }
  code { font-family: ui-monospace, Menlo, Consolas, monospace; background: #1e293b; color: #a7f3d0; padding: 2px 6px; border-radius: 4px; }
  pre { background: #0b1120; border: 1px solid #1e293b; border-radius: 8px; padding: 16px; }
---

# 🌳 Árvores Binárias de Busca (BST)
### Estruturas Hierárquicas, Eficiência O(log n) e Aplicações no Mundo Real

**Disciplina:** Estrutura de Dados (4º Período CC)  
**Docente:** Prof. Me. Karan Luciano  
**Instituição:** Afya São Lucas Ji-Paraná • Sala 59B  

---

## ⚖️ O Ponto de Partida: O Dilema da N1

Por que precisamos de árvores se já tínhamos vetores e listas?

* **Vetor com Busca Binária:**
  * ⚡ Busca ultra-rápida: `O(log n)`.
  * 🐢 Inserção/Remoção péssima: `O(n)` (precisa empurrar milhares de elementos para o lado).

* **Lista Encadeada Dinâmica:**
  * ⚡ Inserção flexível na Heap com ponteiros.
  * 🐢 Busca lenta: `O(n)` (impossível pular para o meio).

> 💡 **A Solução BST:** O melhor dos dois mundos! Busca E inserção em `O(log n)`.

---

## 💡 Busca Binária vs. Árvore Binária (Em 1 Minuto)

| Critério | 🔍 Busca Binária (N1) | 🌳 Árvore Binária - BST (N2) |
| :--- | :--- | :--- |
| **O que é?** | É um **ALGORITMO** de passos. | É uma **ESTRUTURA DE DADOS**. |
| **Memória** | Vetor contíguo rígido na memória. | Nós dinâmicos na Heap com ponteiros. |
| **Buscar** | ⚡ `O(log n)` | ⚡ `O(log n)` |
| **Inserir** | 🐢 `O(n)` (empurra elementos vizinhos) | ⚡ `O(log n)` (apenas aloca e conecta!) |

---

## 🌍 Onde as Árvores Rodam no Mundo Real?

* 🗄️ **Bancos de Dados (PostgreSQL, MySQL, SQLite):**
  * Índices **B-Tree** encontram 1 CPF entre 100 milhões de clientes em **1 ms**.
* 🌐 **Navegadores & Compiladores (Chrome, React, GCC):**
  * O **DOM do HTML** e a **AST** do compilador são árvores em memória.
* 📦 **Compactação de Arquivos (ZIP, JPEG, MP3):**
  * A **Árvore de Huffman** encolhe arquivos atribuindo códigos curtos a caracteres comuns.
* 🎮 **Jogos 3D (Unreal Engine, Unity):**
  * Árvores **BSP** descartam polígonos fora do campo de visão para garantir 120 FPS.

---

## 📐 Anatomia de uma Árvore

```
          [ 50 ]  <-- Raiz (Root)
         /      \
     [ 30 ]    [ 70 ]  <-- Nós Internos
     /    \    /    \
  [ 20 ] [ 40 ][ 60 ] [ 80 ]  <-- Folhas (Leaves)
```

* **👑 Raiz:** O nó do topo, sem pai. Ponto de partida obrigatório de qualquer rota.
* **🍃 Folhas:** Nós terminais sem filhos (`esq == NULL && dir == NULL`).
* **📏 Altura:** O caminho mais longo da raiz até a folha mais distante.
* **⭐ Regra de Ouro:** `Esquerda < Raiz < Direita`.

---

## 💻 A Estrutura do Nó em C

```c
typedef struct No {
    int valor;
    struct No* esq; // ponteiro para os menores
    struct No* dir; // ponteiro para os maiores
} No;

No* criarNo(int valor) {
    No* novo = (No*) malloc(sizeof(No));
    novo->valor = valor;
    novo->esq = NULL;
    novo->dir = NULL;
    return novo;
}
```

> Todo nó recém-criado nasce como folha (`esq` e `dir` começam em `NULL`).

---

## 🔄 A Inserção Recursiva

Toda inserção **sempre começa pela raiz** e desce até achar ponteiro livre:

```c
No* inserir(No* raiz, int valor) {
    // 1. Caso Base: encontrou posição vazia!
    if (raiz == NULL) return criarNo(valor);

    // 2. Se for menor, caminha para a ESQUERDA
    if (valor < raiz->valor)
        raiz->esq = inserir(raiz->esq, valor);
    // 3. Se for maior, caminha para a DIREITA
    else if (valor > raiz->valor)
        raiz->dir = inserir(raiz->dir, valor);

    return raiz;
}
```

---

## 🌟 Os 3 Percursos em Árvores

Como visitar todos os nós de uma estrutura não-linear?

1. **🌟 Em-Ordem (In-Order: Esq ➔ Raiz ➔ Dir):**
   * *O Efeito Mágico:* Imprime os elementos em **ordem crescente perfeita**!
2. **💾 Pré-Ordem (Pre-Order: Raiz ➔ Esq ➔ Dir):**
   * Ideal para clonar, copiar e salvar a árvore em arquivo.
3. **🧹 Pós-Ordem (Post-Order: Esq ➔ Dir ➔ Raiz):**
   * **Mandatório para o `free`:** desaloca os filhos antes de apagar o pai!

---

## 🧪 Micro-Desafio em Sala: Achar o Menor Valor

**Pergunta:** Onde mora o menor número de uma BST?  
**Raciocínio:** Sempre à esquerda! Não precisamos olhar para a direita nenhuma vez.

```c
int buscarMenor(No* raiz) {
    if (raiz == NULL) return -1;
    No* atual = raiz;
    while (atual->esq != NULL) {
        atual = atual->esq; // anda tudo para a esquerda
    }
    return atual->valor;
}
```

---

## 🎯 Conclusão & Próximos Passos da N2

* **Hoje dominamos:**
  * Fundamentos conceituais e conexão com o mundo real.
  * Estrutura do nó, alocação e inserção recursiva partindo da raiz.
  * Percursos e ordenação automática Em-Ordem.

* **Próximos Encontros:**
  * Remoção de nós em árvores (casos de 0, 1 e 2 filhos).
  * Árvores Balanceadas (Árvores AVL e Rotações).
  * Tabelas Hash (Espalhamento e Colisões).
