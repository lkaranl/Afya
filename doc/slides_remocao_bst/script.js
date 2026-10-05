// =============================================================
// AFYA - ESTRUTURA DE DADOS • PROF. KARAN
// CONTROLADOR INTERATIVO DOS SLIDES DE REMOÇÃO EM BST & PERCURSOS
// =============================================================

// 1. GESTÃO E NAVEGAÇÃO DOS SLIDES
const slides = document.querySelectorAll('.slide');
const counter = document.getElementById('counter');
const progressBar = document.getElementById('progress-bar');
let current = 0;

function updateSlide() {
  slides.forEach((s, idx) => s.classList.toggle('active', idx === current));
  if (counter) counter.innerText = `${current + 1} / ${slides.length}`;
  if (progressBar) progressBar.style.width = `${((current + 1) / slides.length) * 100}%`;
}

function nextSlide() {
  if (current < slides.length - 1) {
    current++;
    updateSlide();
  }
}

function prevSlide() {
  if (current > 0) {
    current--;
    updateSlide();
  }
}

function toggleFullscreen() {
  if (!document.fullscreenElement) {
    document.documentElement.requestFullscreen().catch(() => {});
  } else {
    document.exitFullscreen().catch(() => {});
  }
}

window.addEventListener('keydown', (e) => {
  if (e.key === 'ArrowRight' || e.key === ' ' || e.key === 'PageDown') {
    e.preventDefault();
    nextSlide();
  } else if (e.key === 'ArrowLeft' || e.key === 'PageUp') {
    e.preventDefault();
    prevSlide();
  } else if (e.key === 'f' || e.key === 'F') {
    toggleFullscreen();
  }
});

updateSlide();

// -------------------------------------------------------------
// 2. INTERATIVIDADE: CASO 1 (NÓ FOLHA)
// -------------------------------------------------------------
function animarCaso1() {
  const node = document.getElementById('node-case1');
  const edge = document.getElementById('edge-case1');
  const status = document.getElementById('status-case1');

  if (node && edge && status) {
    node.style.opacity = '0';
    node.style.transform = 'scale(0.2)';
    edge.style.opacity = '0';
    status.innerHTML = '✅ <strong style="color:#a7f3d0">Sucesso!</strong> O nó folha 20 foi liberado com <code>free()</code> e o ponteiro <code>raiz->esq</code> do pai 30 agora é <code>NULL</code>.';
  }
}

function resetarCaso1() {
  const node = document.getElementById('node-case1');
  const edge = document.getElementById('edge-case1');
  const status = document.getElementById('status-case1');

  if (node && edge && status) {
    node.style.opacity = '1';
    node.style.transform = 'none';
    edge.style.opacity = '1';
    status.innerText = 'Clique no botão acima para ver o nó folha ser liberado e o ponteiro virar NULL.';
  }
}

// -------------------------------------------------------------
// 3. INTERATIVIDADE: CASO 2 (1 FILHO - ADOÇÃO)
// -------------------------------------------------------------
function animarCaso2() {
  const pai = document.getElementById('node-case2-pai');
  const neto = document.getElementById('node-case2-neto');
  const edgePai = document.getElementById('edge-case2-pai');
  const status = document.getElementById('status-case2');

  if (pai && neto && edgePai && status) {
    pai.style.opacity = '0';
    edgePai.style.opacity = '0';
    // O nó 30 estava em (140, 105). O neto 40 se move para as coordenadas exatas do pai:
    neto.setAttribute('transform', 'translate(140, 105)');
    status.innerHTML = '👪 <strong style="color:#fbbf24">Adoção Concluída!</strong> O nó <strong>30</strong> foi liberado da memória e o neto <strong>40</strong> subiu exatamente para o lugar dele, conectado diretamente ao avô <strong>50</strong>!';
  }
}

function resetarCaso2() {
  const pai = document.getElementById('node-case2-pai');
  const neto = document.getElementById('node-case2-neto');
  const edgePai = document.getElementById('edge-case2-pai');
  const status = document.getElementById('status-case2');

  if (pai && neto && edgePai && status) {
    pai.style.opacity = '1';
    edgePai.style.opacity = '1';
    // Retorna para a posição original de neto em (190, 175)
    neto.setAttribute('transform', 'translate(190, 175)');
    status.innerText = 'O nó 30 tem apenas o filho 40. Clique para ver a ponte de adoção direta.';
  }
}

// -------------------------------------------------------------
// 4. INTERATIVIDADE: CASO 3 (2 FILHOS - SUCESSÃO)
// -------------------------------------------------------------
function animarCaso3() {
  const textRaiz = document.getElementById('text-case3-raiz');
  const circleRaiz = document.getElementById('circle-case3-raiz');
  const nodeSuc = document.getElementById('node-case3-suc');
  const edgeSuc = document.getElementById('edge-case3-suc');
  const status = document.getElementById('status-case3');

  if (textRaiz && circleRaiz && nodeSuc && edgeSuc && status) {
    circleRaiz.style.fill = '#059669';
    circleRaiz.style.stroke = '#34d399';
    textRaiz.innerText = '60';

    nodeSuc.style.opacity = '0';
    edgeSuc.style.opacity = '0';

    status.innerHTML = '👑 <strong style="color:#a7f3d0">Sucessão ao Trono Concluída!</strong> O valor 60 assumiu a raiz e a folha original 60 foi liberada na base.';
  }
}

function resetarCaso3() {
  const textRaiz = document.getElementById('text-case3-raiz');
  const circleRaiz = document.getElementById('circle-case3-raiz');
  const nodeSuc = document.getElementById('node-case3-suc');
  const edgeSuc = document.getElementById('edge-case3-suc');
  const status = document.getElementById('status-case3');

  if (textRaiz && circleRaiz && nodeSuc && edgeSuc && status) {
    circleRaiz.style.fill = '#1e293b';
    circleRaiz.style.stroke = '#38bdf8';
    textRaiz.innerText = '50';

    nodeSuc.style.opacity = '1';
    edgeSuc.style.opacity = '1';
    status.innerText = 'Clique no botão acima para ver o 60 assumir a raiz e sua cópia na base ser liberada.';
  }
}

// -------------------------------------------------------------
// 5. INTERATIVIDADE: PERCURSOS ANIMADOS EM PROFUNDIDADE (DFS)
// -------------------------------------------------------------
let animandoPercurso = false;

function animarPercurso(tipo) {
  if (animandoPercurso) return;
  animandoPercurso = true;

  const ids = ['pnode-50', 'pnode-30', 'pnode-70', 'pnode-20', 'pnode-40', 'pnode-60', 'pnode-80'];
  ids.forEach(id => {
    const el = document.getElementById(id);
    if (el) el.classList.remove('node-active-visit', 'node-dim');
  });

  let ordem = [];
  let nome = '';
  if (tipo === 'pre') {
    ordem = [50, 30, 20, 40, 70, 60, 80];
    nome = 'Pré-Ordem (Raiz-Esq-Dir)';
  } else if (tipo === 'em') {
    ordem = [20, 30, 40, 50, 60, 70, 80];
    nome = 'Em-Ordem (Esq-Raiz-Dir)';
  } else if (tipo === 'pos') {
    ordem = [20, 40, 30, 60, 80, 70, 50];
    nome = 'Pós-Ordem (Esq-Dir-Raiz)';
  }

  const status = document.getElementById('status-percursos');
  if (status) status.innerText = `${nome}: `;

  let idx = 0;
  function proximo() {
    if (idx < ordem.length) {
      const val = ordem[idx];
      const node = document.getElementById(`pnode-${val}`);
      if (node) node.classList.add('node-active-visit');

      if (status) status.innerText += (idx === 0 ? '' : ' ➔ ') + val;

      setTimeout(() => {
        if (node) {
          node.classList.remove('node-active-visit');
          node.classList.add('node-dim');
        }
        idx++;
        proximo();
      }, 650);
    } else {
      animandoPercurso = false;
      if (status) status.innerHTML += ' <span style="color:#4ade80; font-weight:800;">[CONCLUÍDO]</span>';
      setTimeout(() => {
        ids.forEach(id => {
          const el = document.getElementById(id);
          if (el) el.classList.remove('node-dim');
        });
      }, 1500);
    }
  }
  proximo();
}

// -------------------------------------------------------------
// 6. INTERATIVIDADE: QUIZ DE FIXAÇÃO
// -------------------------------------------------------------
function responderQuiz(opcao) {
  const status = document.getElementById('status-quiz');
  if (!status) return;

  if (opcao === 2) {
    status.innerHTML = '🎉 <span style="color: #4ade80;">RESPOSTA CORRETA! (B)</span> O 60 é o menor da subárvore direita (sucessor imediato)!';
    status.style.borderColor = '#10b981';
  } else if (opcao === 1) {
    status.innerHTML = '❌ <span style="color: #f43f5e;">Incorreto:</span> O 80 é o maior da direita (não o menor). O sucessor é o 60!';
    status.style.borderColor = '#f43f5e';
  } else {
    status.innerHTML = '❌ <span style="color: #f43f5e;">Incorreto:</span> O 50 é o pai ancestral, não pode ser sucessor do 70!';
    status.style.borderColor = '#f43f5e';
  }
}
