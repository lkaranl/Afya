// =========================================================================
// AFYA • LABORATÓRIO VIRTUAL DE ESTRUTURA DE DADOS (PROF. KARAN)
// MOTOR INTERATIVO C-SANDBOX PARA ÁRVORES BINÁRIAS DE BUSCA (BST)
// =========================================================================

// --- CÓDIGO TEMPLATE C PADRÃO ---
const C_TEMPLATES = {
  balanced: `// =====================================================================
// ÁRVORE BINÁRIA DE BUSCA (BST) EM C • LABORATÓRIO VIRTUAL AFYA
// Prof. Karan • Estrutura de Dados
// =====================================================================

#include <stdio.h>
#include <stdlib.h>

typedef struct No {
    int conteudo;
    struct No* esq;
    struct No* dir;
} No;

No* criarNo(int valor) {
    No* novo = (No*)malloc(sizeof(No));
    novo->conteudo = valor;
    novo->esq = NULL;
    novo->dir = NULL;
    return novo;
}

No* inserir(No* raiz, int valor) {
    if (raiz == NULL) return criarNo(valor);
    if (valor < raiz->conteudo)
        raiz->esq = inserir(raiz->esq, valor);
    else if (valor > raiz->conteudo)
        raiz->dir = inserir(raiz->dir, valor);
    return raiz;
}

No* menorNo(No* raiz) {
    No* atual = raiz;
    while (atual && atual->esq != NULL)
        atual = atual->esq;
    return atual;
}

No* remover(No* raiz, int valor) {
    if (raiz == NULL) return NULL;

    if (valor < raiz->conteudo)
        raiz->esq = remover(raiz->esq, valor);
    else if (valor > raiz->conteudo)
        raiz->dir = remover(raiz->dir, valor);
    else {
        // Caso 1: Folha ou Caso 2: Só filho direito
        if (raiz->esq == NULL) {
            No* temp = raiz->dir;
            free(raiz);
            return temp;
        }
        // Caso 2: Só filho esquerdo
        else if (raiz->dir == NULL) {
            No* temp = raiz->esq;
            free(raiz);
            return temp;
        }

        // Caso 3: Dois filhos (Sucessor imediato)
        No* temp = menorNo(raiz->dir);
        raiz->conteudo = temp->conteudo;
        raiz->dir = remover(raiz->dir, temp->conteudo);
    }
    return raiz;
}

int main(void) {
    No* raiz = NULL;

    // INSERÇÃO INICIAL DOS NÓS (BST BALANCEADA DE 15 NÓS - 4 NÍVEIS)
    raiz = inserir(raiz, 50); // Raiz
    raiz = inserir(raiz, 25); // Nível 1
    raiz = inserir(raiz, 75);
    raiz = inserir(raiz, 15); // Nível 2
    raiz = inserir(raiz, 35);
    raiz = inserir(raiz, 65);
    raiz = inserir(raiz, 85);
    raiz = inserir(raiz, 10); // Nível 3 (Folhas)
    raiz = inserir(raiz, 20);
    raiz = inserir(raiz, 30);
    raiz = inserir(raiz, 40);
    raiz = inserir(raiz, 60);
    raiz = inserir(raiz, 70);
    raiz = inserir(raiz, 80);
    raiz = inserir(raiz, 90);

    return 0;
}
`,

  gamingLeaderboard: `// =====================================================================
// DUELO DE BUSCA: MODO TRADICIONAL (VETOR) vs ÁRVORE BINÁRIA (BST)
// Prof. Karan • Estrutura de Dados
// =====================================================================

#include <stdio.h>
#include <stdlib.h>

// ---------------------------------------------------------------------
// 1. MODO TRADICIONAL: BUSCA LINEAR EM VETOR ORDENADO (O(n))
// ---------------------------------------------------------------------
int buscaVetor(int vetor[], int n, int valor) {
    for (int i = 0; i < n; i++) {
        if (vetor[i] == valor) {
            return i; // Encontrado no índice i
        }
    }
    return -1; // Não encontrado
}

// ---------------------------------------------------------------------
// 2. MODO ÁRVORE BINÁRIA DE BUSCA (BST COM PONTEIROS - O(log n))
// ---------------------------------------------------------------------
typedef struct No {
    int score;             // Pontuação do jogador
    struct No* esq;
    struct No* dir;
} No;

No* criarNo(int valor) {
    No* novo = (No*)malloc(sizeof(No));
    novo->score = valor;
    novo->esq = NULL;
    novo->dir = NULL;
    return novo;
}

No* inserir(No* raiz, int valor) {
    if (raiz == NULL) return criarNo(valor);
    if (valor < raiz->score)
        raiz->esq = inserir(raiz->esq, valor);
    else if (valor > raiz->score)
        raiz->dir = inserir(raiz->dir, valor);
    return raiz;
}

No* buscarBST(No* raiz, int valor) {
    if (raiz == NULL) return NULL;
    if (valor == raiz->score)
        return raiz; // Encontrado no nó atual
    if (valor < raiz->score)
        return buscarBST(raiz->esq, valor);
    else
        return buscarBST(raiz->dir, valor);
}

int main(void) {
    // Array do Vetor Tradicional (15 Scores Ordenados):
    int vetorScores[] = {600, 1200, 1800, 2500, 3100, 3700, 4300, 5000, 5500, 6200, 6800, 7500, 8100, 8800, 9500};
    int totalScores = 15;

    // Árvore Binária do Servidor (BST Balanceada de 15 Scores):
    No* ranking = NULL;
    ranking = inserir(ranking, 5000); // Raiz
    ranking = inserir(ranking, 2500); // Nível 1
    ranking = inserir(ranking, 7500);
    ranking = inserir(ranking, 1200); // Nível 2
    ranking = inserir(ranking, 3700);
    ranking = inserir(ranking, 6200);
    ranking = inserir(ranking, 8800);
    ranking = inserir(ranking, 600);  // Nível 3 (Folhas)
    ranking = inserir(ranking, 1800);
    ranking = inserir(ranking, 3100);
    ranking = inserir(ranking, 4300);
    ranking = inserir(ranking, 5500);
    ranking = inserir(ranking, 6800);
    ranking = inserir(ranking, 8100);
    ranking = inserir(ranking, 9500);

    // Testes de Busca (Pior Caso: 15 testes no Vetor vs 4 na BST):
    int alvo = 9500;
    int posVetor = buscaVetor(vetorScores, totalScores, alvo);
    No* noBST = buscarBST(ranking, alvo);

    return 0;
}
`
};

// --- ESTRUTURA DE DADOS BST INTERNA ---
class BSTNode {
  constructor(value) {
    this.value = value;
    this.left = null;
    this.right = null;
    this.x = 0;
    this.y = 0;
    this.depth = 0;
  }
}

class BSTSimulator {
  constructor() {
    this.root = null;
  }

  clear() {
    this.root = null;
  }

  insert(val) {
    this.root = this._insertNode(this.root, val);
  }

  _insertNode(node, val) {
    if (!node) return new BSTNode(val);
    if (val < node.value) {
      node.left = this._insertNode(node.left, val);
    } else if (val > node.value) {
      node.right = this._insertNode(node.right, val);
    }
    return node;
  }

  minNode(node) {
    let curr = node;
    while (curr && curr.left) curr = curr.left;
    return curr;
  }

  remove(val) {
    this.root = this._removeNode(this.root, val);
  }

  _removeNode(node, val) {
    if (!node) return null;

    if (val < node.value) {
      node.left = this._removeNode(node.left, val);
    } else if (val > node.value) {
      node.right = this._removeNode(node.right, val);
    } else {
      // Caso 1 e 2
      if (!node.left) return node.right;
      if (!node.right) return node.left;

      // Caso 3: 2 filhos
      const temp = this.minNode(node.right);
      node.value = temp.value;
      node.right = this._removeNode(node.right, temp.value);
    }
    return node;
  }

  getHeight(node = this.root) {
    if (!node) return 0;
    return 1 + Math.max(this.getHeight(node.left), this.getHeight(node.right));
  }

  countNodes(node = this.root) {
    if (!node) return 0;
    return 1 + this.countNodes(node.left) + this.countNodes(node.right);
  }

  countLeaves(node = this.root) {
    if (!node) return 0;
    if (!node.left && !node.right) return 1;
    return this.countLeaves(node.left) + this.countLeaves(node.right);
  }

  inOrder(node = this.root, acc = []) {
    if (!node) return acc;
    this.inOrder(node.left, acc);
    acc.push(node.value);
    this.inOrder(node.right, acc);
    return acc;
  }

  preOrder(node = this.root, acc = []) {
    if (!node) return acc;
    acc.push(node.value);
    this.preOrder(node.left, acc);
    this.preOrder(node.right, acc);
    return acc;
  }

  postOrder(node = this.root, acc = []) {
    if (!node) return acc;
    this.postOrder(node.left, acc);
    this.postOrder(node.right, acc);
    acc.push(node.value);
    return acc;
  }
}

// Instância do simulador
const bst = new BSTSimulator();

// Elementos da UI
const codeEditor = document.getElementById('code-editor');
const highlightingContent = document.getElementById('highlighting-content');
const highlighting = document.getElementById('highlighting');
const lineNumbers = document.getElementById('line-numbers');
const presetSelect = document.getElementById('preset-select');
const svgCanvas = document.getElementById('tree-canvas');
const statNodes = document.getElementById('stat-nodes');
const statHeight = document.getElementById('stat-height');
const statLeaves = document.getElementById('stat-leaves');
const statRoot = document.getElementById('stat-root');
const travInOrder = document.getElementById('trav-inorder');
const travPreOrder = document.getElementById('trav-preorder');
const travPostOrder = document.getElementById('trav-postorder');
const emptyState = document.getElementById('empty-state');
const quickInput = document.getElementById('quick-input');
const hoverTooltip = document.getElementById('node-hover-tooltip');
const explanationCard = document.getElementById('node-explanation-card');
const slowmoToggle = document.getElementById('slowmo-toggle');
const slowmoSpeed = document.getElementById('slowmo-speed');
const debuggerBanner = document.getElementById('debugger-banner');
const debuggerText = document.getElementById('debugger-text');
const heapDrawer = document.getElementById('heap-drawer');
const btnHeapToggle = document.getElementById('btn-heap-toggle');
const statHealth = document.getElementById('stat-health');


// =========================================================================
// RASTREAMENTO PEDAGÓGICO DE INSERÇÃO (HOVER: POR QUE FOI ADICIONADO AQUI?)
// =========================================================================
function rastrearDecisaoBST(targetVal) {
  if (!bst.root) return null;

  if (bst.root.value === targetVal) {
    return {
      isRoot: true,
      passos: [
        '🌱 <strong>Raiz da Árvore</strong>: Foi o primeiro nó inserido (ou tornou-se a raiz atual).',
        '📍 Alocado no endereço base da árvore (<code>raiz != NULL</code>).',
        `⚖️ Todo novo nó inserido será comparado com <strong>${targetVal}</strong> para descer à esquerda ou direita.`
      ],
      conclusao: 'Nó Raiz: Ponto de partida de qualquer busca ou inserção na BST.'
    };
  }

  const passos = [];
  let curr = bst.root;
  let pai = null;
  let direcao = '';

  while (curr && curr.value !== targetVal) {
    pai = curr;
    if (targetVal < curr.value) {
      passos.push(`• <strong>${targetVal} &lt; ${curr.value}</strong> ➔ Desceu para a <em>subárvore esquerda</em>`);
      curr = curr.left;
      direcao = 'esquerda (esq)';
    } else if (targetVal > curr.value) {
      passos.push(`• <strong>${targetVal} &gt; ${curr.value}</strong> ➔ Desceu para a <em>subárvore direita</em>`);
      curr = curr.right;
      direcao = 'direita (dir)';
    }
  }

  if (pai) {
    const ponteiro = direcao.startsWith('esq') ? 'esq' : 'dir';
    passos.push(`• Encontrou ponteiro <code>NULL</code> à <strong>${direcao}</strong> do nó <strong>${pai.value}</strong>`);
    passos.push(`• Alocou <code>malloc(sizeof(No))</code> e conectou: <code>${pai.value}-&gt;${ponteiro} = novo</code>`);
  }

  return {
    isRoot: false,
    passos,
    conclusao: `Posicionado como filho à ${direcao} do nó ${pai ? pai.value : 'raiz'}.`
  };
}

// =========================================================================
// EXPLICAÇÃO PEDAGÓGICA AO CLICAR NO NÓ (O QUE ACONTECE NA MEMÓRIA?)
// =========================================================================
function exibirExplicacaoPedagogica(node) {
  if (!explanationCard || !node) return;

  if (hoverTooltip) {
    hoverTooltip.classList.remove('active');
  }

  const isRoot = node === bst.root;
  const isLeaf = !node.left && !node.right;
  const hasOneChild = (node.left && !node.right) || (!node.left && node.right);
  const hasTwoChildren = node.left && node.right;

  let badgeClass = '';
  let badgeText = '';
  let titulo = isRoot ? `Nó ${node.value} (Raiz)` : `Nó ${node.value}`;
  let htmlCorpo = '';

  if (isLeaf) {
    badgeClass = 'leaf';
    badgeText = 'Caso 1: Nó Folha';
    htmlCorpo = `
      <p style="margin-bottom:8px;"><strong>Por que este nó é uma Folha?</strong><br>
      Porque não possui filhos nem à esquerda nem à direita (<code>esq == NULL</code> e <code>dir == NULL</code>).</p>

      <p style="margin-bottom:8px;"><strong>O que acontece ao remover?</strong><br>
      É o caso mais simples da BST. Nenhum outro nó depende dele. O ponteiro do seu pai é ajustado para <code>NULL</code> e sua memória é liberada diretamente com <code>free()</code>.</p>

      <div style="background:#090d16; border:1px solid #10b981; padding:8px 10px; border-radius:6px; font-family:var(--font-code); font-size:12px; margin-top:8px; color:#34d399;">
        // Caso 1: Folha isolada<br>
        free(atual);<br>
        return NULL; // Pai passa a apontar para NULL
      </div>
    `;
  } else if (hasOneChild) {
    badgeClass = 'one-child';
    badgeText = 'Caso 2: 1 Filho';
    const filho = node.left ? node.left : node.right;
    const ladoFilho = node.left ? 'esquerdo' : 'direito';
    const ptrFilho = node.left ? 'esq' : 'dir';
    htmlCorpo = `
      <p style="margin-bottom:8px;"><strong>Por que este nó cai no Caso 2?</strong><br>
      Porque possui apenas um único filho descendente: o nó <strong>${filho.value}</strong> (lado ${ladoFilho}).</p>

      <p style="margin-bottom:8px;"><strong>O que acontece ao remover? (Efeito Bypass / Pular Nó)</strong><br>
      O pai deste nó assume diretamente a tutela do "neto" (o nó <strong>${filho.value}</strong>). O nó <strong>${node.value}</strong> é liberado com <code>free()</code> e a subárvore permanece conectada.</p>

      <div style="background:#090d16; border:1px solid #fbbf24; padding:8px 10px; border-radius:6px; font-family:var(--font-code); font-size:12px; margin-top:8px; color:#fde68a;">
        // Caso 2: Pula o nó e liga o neto ao pai<br>
        No* temp = raiz-&gt;${ptrFilho};<br>
        free(raiz);<br>
        return temp; // Pai herda diretamente o neto!
      </div>
    `;
  } else if (hasTwoChildren) {
    badgeClass = 'two-children';
    badgeText = isRoot ? 'Caso 3: Raiz com 2 Filhos' : 'Caso 3: 2 Filhos';
    const sucessor = bst.minNode(node.right);
    htmlCorpo = `
      <p style="margin-bottom:8px;"><strong>Por que este nó cai no Caso 3?</strong><br>
      Porque possui <strong>dois filhos</strong> (subárvore esquerda com valores menores e direita com valores maiores). Se o nó fosse simplesmente apagado, a árvore seria partida ao meio!</p>

      <p style="margin-bottom:8px;"><strong>O que acontece ao remover? (Sucessor Imediato)</strong><br>
      1. Procuramos o <strong>menor elemento da subárvore direita</strong> (sucessor em-ordem), que é o nó <strong>${sucessor.value}</strong>.<br>
      2. O valor <strong>${sucessor.value}</strong> é copiado para este nó.<br>
      3. Removemos recursivamente o nó <strong>${sucessor.value}</strong> original da subárvore direita (que terá no máximo 1 filho, caindo no Caso 1 ou 2).</p>

      <div style="background:#090d16; border:1px solid #c084fc; padding:8px 10px; border-radius:6px; font-family:var(--font-code); font-size:12px; margin-top:8px; color:#e9d5ff;">
        // Caso 3: Substituição pelo Menor da Direita<br>
        No* temp = menorNo(raiz-&gt;dir); // Sucessor = ${sucessor.value}<br>
        raiz-&gt;conteudo = temp-&gt;conteudo;<br>
        raiz-&gt;dir = remover(raiz-&gt;dir, temp-&gt;conteudo);
      </div>
    `;
  }

  explanationCard.innerHTML = `
    <div class="exp-header">
      <div class="exp-title">
        <span>🔍</span>
        <span>${titulo}</span>
        <span class="exp-badge ${badgeClass}">${badgeText}</span>
      </div>
      <button class="exp-btn-close" onclick="fecharExplicacao()" title="Fechar">✕</button>
    </div>
    <div class="exp-body">
      ${htmlCorpo}
    </div>
    <button class="btn-exp-action" onclick="executarRemocaoDoCard(${node.value})">
      🗑️ Remover ${node.value} no Código C
    </button>
  `;

  explanationCard.classList.add('active');
}

function fecharExplicacao() {
  if (explanationCard) {
    explanationCard.classList.remove('active');
  }
}

function executarRemocaoDoCard(val) {
  removerNoRapido(val);
  fecharExplicacao();
}

// =========================================================================
// 1. SIMULADOR DE MEMÓRIA RAM HEAP (MALLOC & FREE)
// =========================================================================
const heapMemory = {
  allocations: new Map(), // value -> { address, value, status }
  logs: [],
  baseAddr: 0x10A0
};

function getOrAllocHeapAddress(val) {
  if (!heapMemory.allocations.has(val)) {
    const addr = '0x' + (heapMemory.baseAddr + heapMemory.allocations.size * 24).toString(16).toUpperCase();
    heapMemory.allocations.set(val, {
      address: addr,
      value: val,
      status: 'active'
    });
    adicionarLogHeap(`malloc(sizeof(No)) ➔ Alocado nó com valor ${val} no endereço ${addr} (24 bytes)`, 'alloc');
  }
  return heapMemory.allocations.get(val);
}

function adicionarLogHeap(msg, type = 'alloc') {
  const agora = new Date();
  const timeStr = agora.toTimeString().split(' ')[0];
  heapMemory.logs.unshift({ msg, type, time: timeStr });
  if (heapMemory.logs.length > 25) heapMemory.logs.pop();
  renderizarLogsHeap();
}

function renderizarLogsHeap() {
  const container = document.getElementById('heap-logs-list');
  if (!container) return;
  container.innerHTML = heapMemory.logs.map(l =>
    `<div class="heap-log-entry ${l.type}">[${l.time}] ${l.msg}</div>`
  ).join('');
}

function renderizarTabelaHeap() {
  const tbody = document.getElementById('heap-table-body');
  const countEl = document.getElementById('heap-bytes-count');
  if (!tbody) return;

  tbody.innerHTML = '';
  const nosVivos = new Map();
  function coletarVivos(node) {
    if (!node) return;
    nosVivos.set(node.value, node);
    coletarVivos(node.left);
    coletarVivos(node.right);
  }
  coletarVivos(bst.root);

  let activeCount = 0;

  // Garante que todos os nós vivos têm endereço
  nosVivos.forEach((node, val) => {
    getOrAllocHeapAddress(val);
  });

  heapMemory.allocations.forEach((item, val) => {
    const isLive = nosVivos.has(val);
    if (isLive) {
      activeCount++;
      item.status = 'active';
    } else {
      if (item.status === 'active') {
        item.status = 'freed';
        adicionarLogHeap(`free(${item.address}) ➔ Nó ${val} desalocado da Heap pelo SO`, 'free');
      }
    }

    const node = nosVivos.get(val);
    const leftAddr = (node && node.left) ? (heapMemory.allocations.get(node.left.value)?.address || 'NULL') : 'NULL';
    const rightAddr = (node && node.right) ? (heapMemory.allocations.get(node.right.value)?.address || 'NULL') : 'NULL';

    const tr = document.createElement('tr');
    tr.innerHTML = `
      <td style="color:#38bdf8; font-weight:700;">${item.address}</td>
      <td><strong>${item.value}</strong></td>
      <td style="color:${leftAddr === 'NULL' ? '#64748b' : '#34d399'}; font-weight:${leftAddr === 'NULL' ? 'normal' : '700'};">${leftAddr}</td>
      <td style="color:${rightAddr === 'NULL' ? '#64748b' : '#34d399'}; font-weight:${rightAddr === 'NULL' ? 'normal' : '700'};">${rightAddr}</td>
      <td>
        <span class="heap-status-tag ${item.status === 'active' ? 'active' : 'freed'}">
          ${item.status === 'active' ? 'malloc OK' : 'free()'}
        </span>
      </td>
    `;
    tbody.appendChild(tr);
  });

  if (countEl) {
    countEl.innerText = `Nós Vivos: ${activeCount} (${activeCount * 24} bytes alocados)`;
  }
}

function toggleHeapDrawer() {
  const drawer = document.getElementById('heap-drawer');
  if (drawer) {
    drawer.classList.toggle('open');
    if (drawer.classList.contains('open')) {
      renderizarTabelaHeap();
      renderizarLogsHeap();
    }
  }
}

// =========================================================================
// 2. MEDIDOR DE SAÚDE DA ÁRVORE (O(log n) vs O(n))
// =========================================================================
function atualizarMedidorSaude(n, h) {
  const el = document.getElementById('stat-health');
  if (!el) return;

  if (n === 0) {
    el.className = 'health-pill balanced';
    el.innerHTML = '🌱 Vazia';
    el.title = 'A árvore está vazia (raiz == NULL)';
    return;
  }

  // Altura ideal para árvore binária equilibrada: floor(log2(n)) + 1
  const hIdeal = Math.floor(Math.log2(n)) + 1;

  if (n <= 3 || h <= hIdeal + 1) {
    el.className = 'health-pill balanced';
    el.innerHTML = `🟢 O(log n) Equilibrada (h=${h})`;
    el.title = `Excelente! A altura h=${h} está no nível ideal (${hIdeal}). Busca super rápida O(log n).`;
  } else if (h >= n - 1 && n >= 4) {
    el.className = 'health-pill degenerate';
    el.innerHTML = `🔴 O(n) Degenerada! (h=${h})`;
    el.title = `Cuidado crítico! A altura h=${h} é quase igual ao total de nós n=${n}. A árvore virou uma lista encadeada simples O(n)!`;
  } else {
    el.className = 'health-pill unbalanced';
    el.innerHTML = `🟡 Desbalanceada (h=${h})`;
    el.title = `Atenção: A altura h=${h} está acima do ideal (${hIdeal}). Ramos desiguais tornam a busca menos eficiente.`;
  }
}

// =========================================================================
// 3. ANIMADOR INTERATIVO DE PERCURSOS (EM-ORDEM, PRÉ-ORDEM, PÓS-ORDEM)
// =========================================================================
let animacaoPercursoAtiva = false;

function animarPercurso(tipo) {
  if (animacaoPercursoAtiva || !bst.root) return;

  const btnInorder = document.getElementById('btn-anim-inorder');
  const btnPreorder = document.getElementById('btn-anim-preorder');
  const btnPostorder = document.getElementById('btn-anim-postorder');
  const runningBadge = document.getElementById('trav-running-badge');
  const statusText = document.getElementById('trav-status-text');

  animacaoPercursoAtiva = true;
  if (btnInorder) btnInorder.disabled = true;
  if (btnPreorder) btnPreorder.disabled = true;
  if (btnPostorder) btnPostorder.disabled = true;
  if (runningBadge) {
    runningBadge.innerText = 'Executando Animação...';
    runningBadge.classList.add('running');
  }

  // Montar fila didática de passos da recursão
  const eventos = [];

  function simularRecursao(node) {
    if (!node) return;

    if (tipo === 'preorder') {
      eventos.push({ nodeVal: node.value, action: 'printf', msg: `Visita Raiz: <code>printf("%d ", ${node.value});</code> ➔ Imprime imediatamente!` });
      if (node.left) {
        eventos.push({ nodeVal: node.value, action: 'desce_esq', msg: `Desce recursivamente para subárvore esquerda de ${node.value}...` });
        simularRecursao(node.left);
      }
      if (node.right) {
        eventos.push({ nodeVal: node.value, action: 'desce_dir', msg: `Desce recursivamente para subárvore direita de ${node.value}...` });
        simularRecursao(node.right);
      }
    } else if (tipo === 'inorder') {
      if (node.left) {
        eventos.push({ nodeVal: node.value, action: 'desce_esq', msg: `1º Explora toda a subárvore esquerda de ${node.value}...` });
        simularRecursao(node.left);
      }
      eventos.push({ nodeVal: node.value, action: 'printf', msg: `2º Visita Raiz: <code>printf("%d ", ${node.value});</code> ➔ Imprime em ordem crescente!` });
      if (node.right) {
        eventos.push({ nodeVal: node.value, action: 'desce_dir', msg: `3º Explora toda a subárvore direita de ${node.value}...` });
        simularRecursao(node.right);
      }
    } else if (tipo === 'postorder') {
      if (node.left) {
        eventos.push({ nodeVal: node.value, action: 'desce_esq', msg: `1º Explora subárvore esquerda de ${node.value}...` });
        simularRecursao(node.left);
      }
      if (node.right) {
        eventos.push({ nodeVal: node.value, action: 'desce_dir', msg: `2º Explora subárvore direita de ${node.value}...` });
        simularRecursao(node.right);
      }
      eventos.push({ nodeVal: node.value, action: 'printf', msg: `3º Visita Raiz: <code>printf("%d ", ${node.value});</code> ➔ Imprime após ambos os filhos!` });
    }
  }

  simularRecursao(bst.root);

  let idx = 0;
  const delay = 750;

  function proximoPasso() {
    document.querySelectorAll('.svg-node-group').forEach(el => {
      el.classList.remove('node-animating-visit', 'node-animating-printf');
    });

    if (idx >= eventos.length) {
      animacaoPercursoAtiva = false;
      if (btnInorder) btnInorder.disabled = false;
      if (btnPreorder) btnPreorder.disabled = false;
      if (btnPostorder) btnPostorder.disabled = false;
      if (runningBadge) {
        runningBadge.innerText = 'Pronto';
        runningBadge.classList.remove('running');
      }
      if (statusText) {
        statusText.innerHTML = `✓ <strong>Percurso ${tipo.toUpperCase()} finalizado!</strong> Todos os nós foram impressos na ordem exata.`;
      }
      return;
    }

    const ev = eventos[idx];
    if (statusText) statusText.innerHTML = ev.msg;

    const svgNode = document.querySelector(`.svg-node-group[data-val="${ev.nodeVal}"]`);
    if (svgNode) {
      if (ev.action === 'printf') {
        svgNode.classList.add('node-animating-printf');
      } else {
        svgNode.classList.add('node-animating-visit');
      }
    }

    idx++;
    setTimeout(proximoPasso, delay);
  }

  proximoPasso();
}

// =========================================================================
// 4. MODO CÂMERA LENTA (DEBUGGER VISUAL PASSO A PASSO)
// =========================================================================
let debuggerEmExecucao = false;

function fecharDebuggerBanner() {
  const banner = document.getElementById('debugger-banner');
  if (banner) banner.classList.remove('active');
}

function getNodeSvgCoords(val) {
  const el = document.querySelector(`.svg-node-group[data-val="${val}"]`);
  if (!el) return null;
  const transform = el.getAttribute('transform');
  const m = transform.match(/translate\(\s*([-\d.]+)\s*,\s*([-\d.]+)\s*\)/);
  if (m) {
    return { x: parseFloat(m[1]), y: parseFloat(m[2]) };
  }
  return null;
}

function criarOuAtualizarTraveler(valor, x, y) {
  let traveler = document.getElementById('traveler-node');
  if (!traveler) {
    traveler = document.createElementNS('http://www.w3.org/2000/svg', 'g');
    traveler.setAttribute('id', 'traveler-node');
    traveler.setAttribute('class', 'svg-traveler-node');

    const circle = document.createElementNS('http://www.w3.org/2000/svg', 'circle');
    circle.setAttribute('class', 'svg-traveler-circle');
    circle.setAttribute('r', 21);
    traveler.appendChild(circle);

    const text = document.createElementNS('http://www.w3.org/2000/svg', 'text');
    text.setAttribute('class', 'svg-traveler-text');
    text.textContent = valor;
    traveler.appendChild(text);

    const badge = document.createElementNS('http://www.w3.org/2000/svg', 'text');
    badge.setAttribute('class', 'svg-traveler-badge');
    badge.setAttribute('y', -30);
    badge.textContent = 'NOVO';
    traveler.appendChild(badge);

    svgCanvas.appendChild(traveler);
  } else {
    const textEl = traveler.querySelector('.svg-traveler-text');
    const badgeEl = traveler.querySelector('.svg-traveler-badge');
    if (textEl) textEl.textContent = valor;
    if (badgeEl) badgeEl.textContent = 'NOVO';
    traveler.classList.remove('allocated');
  }

  traveler.setAttribute('transform', `translate(${x}, ${y})`);
  return traveler;
}

function removerTraveler() {
  const traveler = document.getElementById('traveler-node');
  if (traveler && traveler.parentNode) {
    traveler.parentNode.removeChild(traveler);
  }
}

function executarDebuggerPassoAPasso(operacao, valor, onComplete) {
  if (debuggerEmExecucao) return;
  debuggerEmExecucao = true;

  const banner = document.getElementById('debugger-banner');
  const dbgText = document.getElementById('debugger-text');
  const speedEl = document.getElementById('slowmo-speed');
  const stepDelay = parseInt(speedEl ? speedEl.value : '1100', 10);

  if (banner) banner.classList.add('active');

  const passos = [];

  if (operacao === 'inserir') {
    if (!bst.root) {
      passos.push({
        nodeVal: null,
        x: 350,
        y: 60,
        isFinal: true,
        msg: `Árvore vazia (raiz == NULL) ➔ Alocando novo nó <strong>${valor}</strong> como a RAIZ inicial da árvore!`
      });
    } else {
      const rootCoords = getNodeSvgCoords(bst.root.value) || { x: 350, y: 50 };
      // Ponto de entrada: ligeiramente acima da raiz
      passos.push({
        nodeVal: null,
        x: rootCoords.x,
        y: Math.max(20, rootCoords.y - 45),
        msg: `Iniciando inserção do valor <strong>${valor}</strong> na BST... descendo para comparar com a raiz!`
      });

      let curr = bst.root;
      let lastCoords = rootCoords;

      while (curr) {
        const coords = getNodeSvgCoords(curr.value) || lastCoords;
        lastCoords = coords;

        if (valor === curr.value) {
          passos.push({
            nodeVal: curr.value,
            x: coords.x,
            y: coords.y,
            msg: `O valor <strong>${valor}</strong> já existe na árvore! BSTs típicas não duplicam chaves.`
          });
          break;
        } else if (valor < curr.value) {
          passos.push({
            nodeVal: curr.value,
            x: coords.x,
            y: coords.y,
            msg: `Comparando: <strong>${valor} &lt; ${curr.value}</strong> (${valor} é menor!) ➔ Desce pela subárvore esquerda (raiz-&gt;esq)`
          });
          if (!curr.left) {
            passos.push({
              nodeVal: curr.value,
              x: Math.max(30, coords.x - 55),
              y: coords.y + 70,
              isFinal: true,
              msg: `Ponteiro esquerdo de ${curr.value} é <strong>NULL</strong>! Alocando com <code>malloc(sizeof(No))</code> e conectando: <code>${curr.value}-&gt;esq = novo</code>!`
            });
            break;
          }
          curr = curr.left;
        } else {
          passos.push({
            nodeVal: curr.value,
            x: coords.x,
            y: coords.y,
            msg: `Comparando: <strong>${valor} &gt; ${curr.value}</strong> (${valor} é maior!) ➔ Desce pela subárvore direita (raiz-&gt;dir)`
          });
          if (!curr.right) {
            passos.push({
              nodeVal: curr.value,
              x: coords.x + 55,
              y: coords.y + 70,
              isFinal: true,
              msg: `Ponteiro direito de ${curr.value} é <strong>NULL</strong>! Alocando com <code>malloc(sizeof(No))</code> e conectando: <code>${curr.value}-&gt;dir = novo</code>!`
            });
            break;
          }
          curr = curr.right;
        }
      }
    }

    // Criar traveler na posição inicial
    criarOuAtualizarTraveler(valor, passos[0].x, passos[0].y);
  } else {
    // Operação de Remoção
    passos.push({
      nodeVal: valor,
      msg: `Iniciando busca e remoção do nó <strong>${valor}</strong>... verificando estrutura na memória.`
    });
  }

  let stepIdx = 0;
  function rodarPasso() {
    document.querySelectorAll('.svg-node-group').forEach(el => {
      el.classList.remove('node-animating-visit', 'node-animating-printf');
    });

    if (stepIdx >= passos.length) {
      debuggerEmExecucao = false;
      if (dbgText) {
        dbgText.innerHTML = `✓ ${operacao === 'inserir' ? 'Inserção' : 'Remoção'} de <strong>${valor}</strong> concluída no código C!`;
      }
      setTimeout(() => {
        removerTraveler();
        if (banner) banner.classList.remove('active');
      }, 1000);

      if (onComplete) onComplete();
      return;
    }

    const p = passos[stepIdx];
    if (dbgText) dbgText.innerHTML = p.msg;

    if (operacao === 'inserir') {
      const traveler = criarOuAtualizarTraveler(valor, p.x, p.y);
      if (p.isFinal) {
        traveler.classList.add('allocated');
        const badge = traveler.querySelector('.svg-traveler-badge');
        if (badge) badge.textContent = 'malloc OK';
      }
    }

    if (p.nodeVal !== null) {
      const svgNode = document.querySelector(`.svg-node-group[data-val="${p.nodeVal}"]`);
      if (svgNode) {
        if (operacao === 'remover') {
          svgNode.classList.add('node-highlight-remove');
        } else {
          svgNode.classList.add('node-animating-visit');
        }
      }
    }

    stepIdx++;
    setTimeout(rodarPasso, stepDelay);
  }

  rodarPasso();
}



// =========================================================================
// SYNTAX HIGHLIGHTING ENGINE EM TEMPO REAL (PADRÃO AFYA)
// =========================================================================
function highlightC(source) {
  if (!source) return '';

  const tokens = [];
  function saveToken(raw, cls) {
    const idx = tokens.length;
    tokens.push({ raw, cls });
    return `___TOKEN_${idx}___`;
  }

  // 1. Comentários (linha e bloco)
  let s = source.replace(/\/\/[^\n]*/g, m => saveToken(m, 'c-cm'));
  s = s.replace(/\/\*[\s\S]*?\*\//g, m => saveToken(m, 'c-cm'));

  // 2. Includes de bibliotecas com <...>
  s = s.replace(/#include\s+<[^>]+>/g, m => saveToken(m, 'c-lib'));

  // 3. Strings literais
  s = s.replace(/"[^"\\]*(\\.[^"\\]*)*"/g, m => saveToken(m, 'c-str'));

  // 4. Escapar HTML em volta
  s = s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');

  // 5. Palavras-chave e tipos fundamentais
  const kwRegex = /\b(int|void|return|if|else|while|for|struct|typedef|char|float|double)\b/g;
  s = s.replace(kwRegex, '<span class="c-kw">$1</span>');

  // 6. Tipos de estruturas (No, No*)
  s = s.replace(/\bNo\b/g, '<span class="c-type">No</span>');

  // 7. Constantes e Números
  s = s.replace(/\b(NULL|EXIT_SUCCESS|EXIT_FAILURE)\b/g, '<span class="c-num">$1</span>');
  s = s.replace(/\b(-?\d+)\b/g, '<span class="c-num">$1</span>');

  // 8. Funções específicas de Árvores BST
  const fnRegex = /\b(inserir|remover|criarNo|menorNo|main)\b/g;
  s = s.replace(fnRegex, '<span class="c-fn">$1</span>');

  // 9. Chamadas de Biblioteca C
  const libRegex = /\b(malloc|free|printf|sizeof|perror|exit)\b/g;
  s = s.replace(libRegex, '<span class="c-lib">$1</span>');

  // 10. Restaurar tokens protegidos
  tokens.forEach((t, idx) => {
    let esc = t.raw.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
    s = s.replace(`___TOKEN_${idx}___`, `<span class="${t.cls}">${esc}</span>`);
  });

  // Garantir que nova linha final não quebre o alinhamento
  if (source[source.length - 1] === '\n') {
    s += ' ';
  }

  return s;
}

// Sincronizar scroll e highlight
function syncEditorView() {
  if (codeEditor && highlightingContent) {
    highlightingContent.innerHTML = highlightC(codeEditor.value);
  }
  if (codeEditor && highlighting) {
    highlighting.scrollTop = codeEditor.scrollTop;
    highlighting.scrollLeft = codeEditor.scrollLeft;
  }
  if (codeEditor && lineNumbers) {
    lineNumbers.scrollTop = codeEditor.scrollTop;
  }
}

// --- ATUALIZADOR DE NÚMERO DE LINHAS ---
function updateLineNumbers() {
  if (!codeEditor || !lineNumbers) return;
  const lines = codeEditor.value.split('\n').length;
  let nums = '';
  for (let i = 1; i <= lines; i++) {
    nums += i + '\n';
  }
  lineNumbers.innerText = nums;
}

// --- PARSER DE CÓDIGO C EM TEMPO REAL ---
function parseCCodeAndExecute() {
  if (!codeEditor) return;
  const code = codeEditor.value;
  bst.clear();

  // Procurar por chamadas de inserir(raiz, X) ou remover(raiz, X) (suporta 'raiz', 'ranking', etc.)
  const callRegex = /(?:[a-zA-Z_]\w*\s*=\s*)?(inserir|remover)\s*\(\s*[a-zA-Z_]\w*\s*,\s*(-?\d+)\s*\)/g;
  let match;
  let hasCalls = false;
  while ((match = callRegex.exec(code)) !== null) {
    hasCalls = true;
    const op = match[1];
    const val = parseInt(match[2], 10);
    if (!isNaN(val)) {
      if (op === 'inserir') {
        bst.insert(val);
      } else if (op === 'remover') {
        bst.remove(val);
      }
    }
  }

  // Se NÃO houver chamadas de inserir/remover, permite inicialização por array específico para árvore (NUNCA vetorScores):
  if (!hasCalls) {
    const arrayMatch = code.match(/int\s+(?:valores|elementos|dados|chaves)\[\s*\]\s*=\s*\{([^}]+)\}/i);
    if (arrayMatch && arrayMatch[1]) {
      const rawNums = arrayMatch[1].split(',');
      rawNums.forEach(n => {
        const parsed = parseInt(n.trim(), 10);
        if (!isNaN(parsed)) bst.insert(parsed);
      });
    }
  }

  renderTree();
  updateStatsAndTraversals();
  syncEditorView();
}

function atualizarTudo() {
  updateLineNumbers();
  parseCCodeAndExecute();
  syncEditorView();
}

// --- RENDERIZADOR SVG DINÂMICO E ROBUSTO DA ÁRVORE ---
function calculatePositions(root) {
  if (!root) return { minX: 0, maxX: 700, minY: 0, maxY: 400 };

  // 1. Atribui profundidade (depth)
  function assignDepth(node, depth) {
    if (!node) return;
    node.depth = depth;
    assignDepth(node.left, depth + 1);
    assignDepth(node.right, depth + 1);
  }
  assignDepth(root, 0);

  // 2. Ordem horizontal baseada no percurso Em-Ordem (in-order)
  // Garante que o filho esquerdo fique SEMPRE à esquerda e filho direito à direita
  let col = 0;
  const count = bst.countNodes(root);
  const xSpacing = count > 10 ? 74 : 95; // Espaçamento harmônico para árvore grande
  const ySpacing = count > 10 ? 82 : 88; // Espaçamento vertical entre níveis

  function assignX(node) {
    if (!node) return;
    assignX(node.left);
    node.x = col * xSpacing;
    col++;
    assignX(node.right);
  }
  assignX(root);

  // 3. Centralização estética de nós pais com 2 filhos
  function centerParents(node) {
    if (!node) return;
    centerParents(node.left);
    centerParents(node.right);
    if (node.left && node.right) {
      node.x = (node.left.x + node.right.x) / 2;
    }
  }
  centerParents(root);

  // 4. Atribuição de coordenadas Y
  function assignY(node) {
    if (!node) return;
    node.y = 55 + node.depth * ySpacing;
    assignY(node.left);
    assignY(node.right);
  }
  assignY(root);

  // 5. Bounding box real da árvore
  let minX = Infinity, maxX = -Infinity;
  let minY = Infinity, maxY = -Infinity;

  function findBounds(node) {
    if (!node) return;
    if (node.x < minX) minX = node.x;
    if (node.x > maxX) maxX = node.x;
    if (node.y < minY) minY = node.y;
    if (node.y > maxY) maxY = node.y;
    findBounds(node.left);
    findBounds(node.right);
  }
  findBounds(root);

  return { minX, maxX, minY, maxY };
}

function renderTree() {
  if (!svgCanvas) return;
  svgCanvas.innerHTML = '';

  if (!bst.root) {
    if (emptyState) emptyState.style.display = 'flex';
    return;
  }
  if (emptyState) emptyState.style.display = 'none';

  // Calcular coordenadas
  const bounds = calculatePositions(bst.root);

  // Margens confortáveis para visual amplo e imponente
  const paddingX = 50;
  const paddingTop = 45;
  const paddingBottom = 55;

  const vbX = bounds.minX - paddingX;
  const vbY = bounds.minY - paddingTop;
  const vbW = Math.max(500, (bounds.maxX - bounds.minX) + paddingX * 2);
  const vbH = Math.max(300, (bounds.maxY - bounds.minY) + paddingTop + paddingBottom);

  svgCanvas.setAttribute('viewBox', `${vbX} ${vbY} ${vbW} ${vbH}`);

  // 1. Desenhar Arestas (Edges)
  function drawEdges(node) {
    if (!node) return;
    if (node.left) {
      const line = document.createElementNS('http://www.w3.org/2000/svg', 'line');
      line.setAttribute('x1', node.x);
      line.setAttribute('y1', node.y);
      line.setAttribute('x2', node.left.x);
      line.setAttribute('y2', node.left.y);
      line.setAttribute('class', 'svg-edge');
      svgCanvas.appendChild(line);
      drawEdges(node.left);
    }
    if (node.right) {
      const line = document.createElementNS('http://www.w3.org/2000/svg', 'line');
      line.setAttribute('x1', node.x);
      line.setAttribute('y1', node.y);
      line.setAttribute('x2', node.right.x);
      line.setAttribute('y2', node.right.y);
      line.setAttribute('class', 'svg-edge');
      svgCanvas.appendChild(line);
      drawEdges(node.right);
    }
  }
  drawEdges(bst.root);

  // 2. Desenhar Nós (Nodes)
  function drawNodes(node) {
    if (!node) return;

    const g = document.createElementNS('http://www.w3.org/2000/svg', 'g');
    g.setAttribute('class', 'svg-node-group');
    g.setAttribute('transform', `translate(${node.x}, ${node.y})`);
    g.setAttribute('data-val', node.value);

    const isRoot = node === bst.root;
    const isLeaf = !node.left && !node.right;

    if (isRoot) g.classList.add('node-is-root');
    else if (isLeaf) g.classList.add('node-is-leaf');

    const circle = document.createElementNS('http://www.w3.org/2000/svg', 'circle');
    circle.setAttribute('r', isRoot ? 25 : 21.5);
    circle.setAttribute('class', 'svg-node-circle');
    g.appendChild(circle);

    const text = document.createElementNS('http://www.w3.org/2000/svg', 'text');
    text.setAttribute('class', 'svg-node-text');
    text.textContent = node.value;
    g.appendChild(text);

    // Badge textual superior ou inferior
    if (isRoot) {
      const badge = document.createElementNS('http://www.w3.org/2000/svg', 'text');
      badge.setAttribute('class', 'svg-node-badge');
      badge.setAttribute('y', -32);
      badge.setAttribute('fill', '#38bdf8');
      badge.textContent = '👑 RAIZ';
      g.appendChild(badge);
    } else if (isLeaf) {
      const badge = document.createElementNS('http://www.w3.org/2000/svg', 'text');
      badge.setAttribute('class', 'svg-node-badge');
      badge.setAttribute('y', 34);
      badge.setAttribute('fill', '#34d399');
      badge.textContent = 'FOLHA';
      g.appendChild(badge);
    }

    // Evento Hover: Explicar por que o nó foi adicionado nesta posição
    g.addEventListener('mouseenter', () => {
      const decisao = rastrearDecisaoBST(node.value);
      if (!decisao || !hoverTooltip) return;

      const circleRect = circle.getBoundingClientRect();
      const parentRect = svgCanvas.parentElement.getBoundingClientRect();

      const posX = circleRect.left - parentRect.left + circleRect.width / 2;
      const posY = circleRect.top - parentRect.top;

      let htmlSteps = decisao.passos.map(p => `<div class="tooltip-step">${p}</div>`).join('');

      hoverTooltip.innerHTML = `
        <div class="tooltip-title">
          <span>🧠</span>
          <span>Por que o nó <strong>${node.value}</strong> foi adicionado aqui?</span>
        </div>
        ${htmlSteps}
        <div class="tooltip-footer">
          ✓ ${decisao.conclusao}
        </div>
      `;

      // Posicionamento inteligente no container relativo
      const parentW = parentRect.width;
      const tooltipW = 310;
      let left = posX - tooltipW / 2;
      if (left < 10) left = 10;
      if (left + tooltipW > parentW - 10) left = parentW - tooltipW - 10;

      let top = posY - 130;
      if (top < 10) top = posY + 35; // se estiver muito perto do topo, mostra abaixo do nó

      hoverTooltip.style.left = `${left}px`;
      hoverTooltip.style.top = `${top}px`;
      hoverTooltip.classList.add('active');
    });

    g.addEventListener('mouseleave', () => {
      if (hoverTooltip) {
        hoverTooltip.classList.remove('active');
      }
    });

    // Clique interativo no nó para abrir explicação pedagógica detalhada
    g.addEventListener('click', (e) => {
      e.stopPropagation();
      exibirExplicacaoPedagogica(node);
    });

    svgCanvas.appendChild(g);

    drawNodes(node.left);
    drawNodes(node.right);
  }
  drawNodes(bst.root);
}

// --- ATUALIZAÇÃO DE TELEMETRIA E PERCURSOS ---
function updateStatsAndTraversals() {
  const nodes = bst.countNodes();
  const height = bst.getHeight();
  const leaves = bst.countLeaves();
  const rootVal = bst.root ? bst.root.value : 'NULL';

  if (statNodes) statNodes.innerText = nodes;
  if (statHeight) statHeight.innerText = height;
  if (statLeaves) statLeaves.innerText = leaves;
  if (statRoot) statRoot.innerText = rootVal;

  const inOrder = bst.inOrder().join(' ➔ ');
  const preOrder = bst.preOrder().join(' ➔ ');
  const postOrder = bst.postOrder().join(' ➔ ');

  if (travInOrder) travInOrder.innerText = inOrder ? inOrder : '(árvore vazia)';
  if (travPreOrder) travPreOrder.innerText = preOrder ? preOrder : '(árvore vazia)';
  if (travPostOrder) travPostOrder.innerText = postOrder ? postOrder : '(árvore vazia)';

  // 1. Atualizar diagnóstico assintótico de eficiência
  atualizarMedidorSaude(nodes, height);

  // 2. Sincronizar simulação de memória RAM Heap
  renderizarTabelaHeap();
}

// =========================================================================
// INSERÇÃO E REMOÇÃO COM INDENTAÇÃO ESTRITA DE 4 ESPAÇOS
// =========================================================================
function adicionarComandoAoCodigo(operacao, valor) {
  if (isNaN(valor)) return;
  let code = codeEditor.value;
  const varName = (code.includes('ranking =') || code.includes('No* ranking')) ? 'ranking' : 'raiz';
  const novaLinha = `    ${varName} = ${operacao}(${varName}, ${valor});`;

  // Localiza a linha do return 0; substituindo preservando rigorosamente 4 espaços
  const returnRegex = /^[ \t]*return\s+0\s*;/m;
  if (returnRegex.test(code)) {
    codeEditor.value = code.replace(returnRegex, `${novaLinha}\n    return 0;`);
  } else {
    // Se não houver return 0;, insere antes da última chave '}' da main
    const lastBrace = code.lastIndexOf('}');
    if (lastBrace !== -1) {
      codeEditor.value = code.substring(0, lastBrace) + `${novaLinha}\n` + code.substring(lastBrace);
    } else {
      codeEditor.value += `\n${novaLinha}`;
    }
  }

  if (quickInput) quickInput.value = '';
  atualizarTudo();
}

// =========================================================================
// DESTAQUE DINÂMICO DE LINHAS NO EDITOR C (CÂMERA LENTA INTERATIVA)
// =========================================================================
let linhaAtivaAtual = null;
let corAtivaAtual = 'amber';

function destacarLinhaCodigo(linhaNum, corCls = 'amber') {
  const highlighter = document.getElementById('code-line-highlighter');
  if (!highlighter || !codeEditor) return;

  linhaAtivaAtual = linhaNum;
  corAtivaAtual = corCls;
  const lineHeight = 21.6; // 13.5px * 1.6
  const lineTopInContent = 14 + (linhaNum - 1) * lineHeight;
  const editorH = codeEditor.clientHeight;

  // Se a linha estiver fora da visão do editor, rola suavemente para focar nela
  if (lineTopInContent < codeEditor.scrollTop || lineTopInContent > codeEditor.scrollTop + editorH - 45) {
    codeEditor.scrollTop = Math.max(0, lineTopInContent - editorH / 2);
    syncEditorView();
  }

  const topOffset = lineTopInContent - codeEditor.scrollTop;
  highlighter.style.top = `${topOffset}px`;
  highlighter.className = `code-line-highlighter active color-${corCls}`;
}

function limparDestaqueLinhaCodigo() {
  const highlighter = document.getElementById('code-line-highlighter');
  if (highlighter) {
    highlighter.classList.remove('active');
  }
  linhaAtivaAtual = null;
}

// =========================================================================
// DUELO DE BUSCA: MODO TRADICIONAL (VETOR) vs MODO ÁRVORE (BST)
// =========================================================================
const DUEL_DATA = {
  vector: [600, 1200, 1800, 2500, 3100, 3700, 4300, 5000, 5500, 6200, 6800, 7500, 8100, 8800, 9500],
  target: 9500,
  emExecucao: false
};

function atualizarAlvoDuelo(val) {
  DUEL_DATA.target = parseInt(val, 10);
  renderizarCelulasVetor();
  const liveText = document.getElementById('duel-live-text');
  if (liveText) {
    liveText.innerHTML = `Alvo definido para o score <strong>${DUEL_DATA.target}</strong>. Escolha um dos modos acima e dê Play!`;
  }
  limparDestaqueLinhaCodigo();
  document.querySelectorAll('.svg-node-group').forEach(el => {
    el.classList.remove('node-animating-visit', 'node-animating-printf');
  });
}

function renderizarCelulasVetor() {
  const container = document.getElementById('duel-vector-cells');
  if (!container) return;
  container.innerHTML = DUEL_DATA.vector.map((val, idx) =>
    `<div class="duel-cell-item" id="vec-cell-${idx}" data-val="${val}">[${idx}]: ${val}</div>`
  ).join('');
}

function rodarDuelo(modo) {
  if (DUEL_DATA.emExecucao) return;
  DUEL_DATA.emExecucao = true;

  const btnTrad = document.getElementById('btn-play-trad');
  const btnBst = document.getElementById('btn-play-bst');
  const liveText = document.getElementById('duel-live-text');
  const targetVal = DUEL_DATA.target;

  if (btnTrad) btnTrad.disabled = true;
  if (btnBst) btnBst.disabled = true;

  renderizarCelulasVetor();
  limparDestaqueLinhaCodigo();
  document.querySelectorAll('.svg-node-group').forEach(el => {
    el.classList.remove('node-animating-visit', 'node-animating-printf');
  });

  if (modo === 'vetor') {
    // --- MODO TRADICIONAL: BUSCA LINEAR EM VETOR (O(n)) ---
    destacarLinhaCodigo(12, 'cyan'); // int buscaVetor(int vetor[], int n, int valor)
    let i = 0;
    const n = DUEL_DATA.vector.length;
    let comparacoes = 0;
    const delay = 650;

    function passoVetor() {
      document.querySelectorAll('.duel-cell-item').forEach(c => c.classList.remove('testing'));

      if (i >= n) {
        destacarLinhaCodigo(18, 'rose'); // return -1;
        if (liveText) liveText.innerHTML = `❌ Score ${targetVal} não encontrado no vetor! (Total: ${comparacoes} comparações no laço for)`;
        concluirDueloVetor(comparacoes, false);
        return;
      }

      destacarLinhaCodigo(13, 'amber'); // for (int i = 0; i < n; i++)
      const cell = document.getElementById(`vec-cell-${i}`);
      if (cell) cell.classList.add('testing');

      setTimeout(() => {
        comparacoes++;
        destacarLinhaCodigo(14, 'amber'); // if (vetor[i] == valor)
        const currentVal = DUEL_DATA.vector[i];

        if (liveText) {
          liveText.innerHTML = `🐢 <strong>vetor[${i}] = ${currentVal}</strong>: testando <code>vetor[${i}] == ${targetVal}</code>... (Comparação #${comparacoes})`;
        }

        if (currentVal === targetVal) {
          if (cell) {
            cell.classList.remove('testing');
            cell.classList.add('found');
          }
          destacarLinhaCodigo(15, 'green'); // return i; (em verde)
          if (liveText) {
            liveText.innerHTML = `🎉 <strong>Encontrado no índice [${i}]!</strong> Total: <strong>${comparacoes} comparações</strong> no vetor.`;
          }
          concluirDueloVetor(comparacoes, true);
        } else {
          i++;
          setTimeout(passoVetor, delay);
        }
      }, delay / 2);
    }

    setTimeout(passoVetor, delay / 2);

  } else if (modo === 'bst') {
    // --- MODO ÁRVORE BINÁRIA: BUSCA EM BST (O(log n)) ---
    destacarLinhaCodigo(47, 'cyan'); // No* buscarBST(No* raiz, int valor)
    let curr = bst.root;
    let comparacoes = 0;
    const delay = 850;

    function passoBST() {
      document.querySelectorAll('.svg-node-group').forEach(el => {
        el.classList.remove('node-animating-visit', 'node-animating-printf');
      });

      if (!curr) {
        destacarLinhaCodigo(48, 'rose'); // if (raiz == NULL) return NULL;
        if (liveText) liveText.innerHTML = `❌ Score ${targetVal} não encontrado na BST! (Total: ${comparacoes} comparações)`;
        concluirDueloBST(comparacoes, false);
        return;
      }

      comparacoes++;
      const svgNode = document.querySelector(`.svg-node-group[data-val="${curr.value}"]`);
      if (svgNode) svgNode.classList.add('node-animating-visit');

      destacarLinhaCodigo(49, 'amber'); // if (valor == raiz->score)

      if (curr.value === targetVal) {
        destacarLinhaCodigo(50, 'green'); // return raiz; (em verde)
        if (svgNode) {
          svgNode.classList.remove('node-animating-visit');
          svgNode.classList.add('node-animating-printf');
        }
        if (liveText) {
          liveText.innerHTML = `⚡ <strong>Encontrado no nó ${curr.value}!</strong> Apenas <strong>${comparacoes} comparações</strong> na BST!`;
        }
        concluirDueloBST(comparacoes, true);
      } else if (targetVal < curr.value) {
        destacarLinhaCodigo(51, 'cyan'); // if (valor < raiz->score)
        if (liveText) {
          liveText.innerHTML = `⚡ <strong>${targetVal} &lt; ${curr.value}</strong>: descendo para subárvore ESQUERDA (raiz->esq)... (Comparação #${comparacoes})`;
        }
        setTimeout(() => {
          destacarLinhaCodigo(52, 'purple'); // return buscarBST(raiz->esq, valor); (em roxo)
          curr = curr.left;
          setTimeout(passoBST, delay);
        }, delay / 2);
      } else {
        destacarLinhaCodigo(51, 'cyan');
        if (liveText) {
          liveText.innerHTML = `⚡ <strong>${targetVal} &gt; ${curr.value}</strong>: descendo para subárvore DIREITA (raiz->dir)... (Comparação #${comparacoes})`;
        }
        setTimeout(() => {
          destacarLinhaCodigo(54, 'purple'); // return buscarBST(raiz->dir, valor); (em roxo)
          curr = curr.right;
          setTimeout(passoBST, delay);
        }, delay / 2);
      }
    }

    setTimeout(passoBST, delay / 2);
  }
}

function concluirDueloVetor(comps, found) {
  DUEL_DATA.emExecucao = false;
  const btnTrad = document.getElementById('btn-play-trad');
  const btnBst = document.getElementById('btn-play-bst');
  if (btnTrad) btnTrad.disabled = false;
  if (btnBst) btnBst.disabled = false;

  const resComps = document.getElementById('res-trad-comps');
  const resTime = document.getElementById('res-trad-time');
  const resCost = document.getElementById('res-trad-cost');

  const tempoMicros = (comps * 1.8).toFixed(1);
  if (resComps) resComps.innerText = `${comps} comparações`;
  if (resTime) resTime.innerText = `~${tempoMicros} μs`;
  if (resCost) resCost.innerText = `Exige mover até ${Math.max(0, 15 - comps)} elementos na RAM`;
}

function concluirDueloBST(comps, found) {
  DUEL_DATA.emExecucao = false;
  const btnTrad = document.getElementById('btn-play-trad');
  const btnBst = document.getElementById('btn-play-bst');
  if (btnTrad) btnTrad.disabled = false;
  if (btnBst) btnBst.disabled = false;

  const resComps = document.getElementById('res-bst-comps');
  const resTime = document.getElementById('res-bst-time');
  const resCost = document.getElementById('res-bst-cost');

  const tempoMicros = (comps * 0.9).toFixed(1);
  if (resComps) resComps.innerText = `${comps} comparações`;
  if (resTime) resTime.innerText = `~${tempoMicros} μs (Mais Rápido!)`;
  if (resCost) resCost.innerText = `0 elementos movidos (1 malloc apenas)`;
}

function inserirNoRapido(val) {
  const num = val !== undefined ? val : parseInt(quickInput ? quickInput.value : '', 10);
  if (isNaN(num)) return;

  if (slowmoToggle && slowmoToggle.checked) {
    executarDebuggerPassoAPasso('inserir', num, () => {
      adicionarComandoAoCodigo('inserir', num);
    });
  } else {
    adicionarComandoAoCodigo('inserir', num);
  }
}

function removerNoRapido(val) {
  const num = val !== undefined ? val : parseInt(quickInput ? quickInput.value : '', 10);
  if (isNaN(num)) return;

  if (slowmoToggle && slowmoToggle.checked) {
    executarDebuggerPassoAPasso('remover', num, () => {
      adicionarComandoAoCodigo('remover', num);
    });
  } else {
    adicionarComandoAoCodigo('remover', num);
  }
}

function resetarArvore() {
  fecharDebuggerBanner();
  heapMemory.allocations.clear();
  heapMemory.logs = [];
  adicionarLogHeap('Memória RAM Heap reinicializada com sucesso', 'alloc');

  if (presetSelect) {
    carregarPreset(presetSelect.value);
  } else {
    carregarPreset('balanced');
  }
}

// --- CARREGAMENTO DE PRESETS ---
function carregarPreset(presetKey) {
  const template = C_TEMPLATES[presetKey] || C_TEMPLATES.balanced;
  const duelPanel = document.getElementById('leaderboard-duel-panel');
  const quickBar = document.getElementById('quick-controls-bar');
  const travTray = document.getElementById('traversal-tray-panel');

  if (presetKey === 'gamingLeaderboard') {
    // Modo Duelo: esconde controles gerais e percursos, ativa painel de duelo
    if (quickBar) quickBar.style.display = 'none';
    if (travTray) travTray.style.display = 'none';
    if (duelPanel) {
      duelPanel.classList.add('active');
      renderizarCelulasVetor();
    }
  } else {
    // Modo Geral (Árvore Equilibrada): reativa controles e percursos, esconde duelo
    if (quickBar) quickBar.style.display = 'flex';
    if (travTray) travTray.style.display = 'flex';
    if (duelPanel) duelPanel.classList.remove('active');
    limparDestaqueLinhaCodigo();
  }

  if (codeEditor) {
    codeEditor.value = template;
    atualizarTudo();
  }
}

if (presetSelect) {
  presetSelect.addEventListener('change', (e) => {
    carregarPreset(e.target.value);
  });
}

// --- EVENT LISTENERS DO EDITOR ---
if (codeEditor) {
  codeEditor.addEventListener('input', () => {
    atualizarTudo();
  });

  codeEditor.addEventListener('scroll', () => {
    syncEditorView();
    if (linhaAtivaAtual !== null) {
      destacarLinhaCodigo(linhaAtivaAtual, corAtivaAtual);
    }
  });

  // Teclado: Suporte a Tab (4 espaços) e Enter com Auto-Indent inteligente
  codeEditor.addEventListener('keydown', (e) => {
    if (e.key === 'Tab') {
      e.preventDefault();
      const start = codeEditor.selectionStart;
      const end = codeEditor.selectionEnd;
      codeEditor.value = codeEditor.value.substring(0, start) + '    ' + codeEditor.value.substring(end);
      codeEditor.selectionStart = codeEditor.selectionEnd = start + 4;
      atualizarTudo();
    } else if (e.key === 'Enter') {
      // Auto-indent: herdar indentação da linha anterior
      const start = codeEditor.selectionStart;
      const lineStart = codeEditor.value.lastIndexOf('\n', start - 1) + 1;
      const currentLine = codeEditor.value.substring(lineStart, start);
      const match = currentLine.match(/^[ \t]+/);
      const indent = match ? match[0] : '';
      const extraIndent = currentLine.trim().endsWith('{') ? '    ' : '';

      e.preventDefault();
      const textToInsert = '\n' + indent + extraIndent;
      codeEditor.value = codeEditor.value.substring(0, start) + textToInsert + codeEditor.value.substring(start);
      codeEditor.selectionStart = codeEditor.selectionEnd = start + textToInsert.length;
      atualizarTudo();
    }
  });
}

if (quickInput) {
  quickInput.addEventListener('keydown', (e) => {
    if (e.key === 'Enter') {
      inserirNoRapido();
    }
  });
}

if (svgCanvas) {
  svgCanvas.addEventListener('click', (e) => {
    if (e.target === svgCanvas) {
      fecharExplicacao();
    }
  });
}

// Inicialização
carregarPreset('balanced');

