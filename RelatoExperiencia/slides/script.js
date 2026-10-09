/**
 * AFYA SLIDE DECK ENGINE - 12º FÓRUM RONDONIENSE DE PESQUISA
 * Prof. Karan Luciano Silva
 */

document.addEventListener('DOMContentLoaded', () => {
  const slides = document.querySelectorAll('.slide');
  const progressBar = document.getElementById('progress-bar');
  const counterCurrent = document.getElementById('counter-current');
  const counterTotal = document.getElementById('counter-total');
  const btnPrev = document.getElementById('btn-prev');
  const btnNext = document.getElementById('btn-next');
  const btnFullscreen = document.getElementById('btn-fullscreen');
  const btnMap = document.getElementById('btn-map');
  const drawer = document.getElementById('modal-drawer');
  const btnCloseDrawer = document.getElementById('btn-close-drawer');
  const drawerGrid = document.getElementById('drawer-grid');
  const timerDisplay = document.getElementById('timer-display');
  const timerBox = document.getElementById('timer-box');

  let currentSlide = 0;
  const totalSlides = slides.length;

  counterTotal.textContent = String(totalSlides).padStart(2, '0');

  // TIMER DO ORADOR (10 MINUTOS)
  let timerSeconds = 10 * 60; // 10 min
  let timerRunning = false;
  let timerInterval = null;

  function updateTimerUI() {
    const min = Math.floor(timerSeconds / 60);
    const sec = timerSeconds % 60;
    timerDisplay.textContent = `${String(min).padStart(2, '0')}:${String(sec).padStart(2, '0')}`;
    
    if (timerSeconds <= 120 && timerSeconds > 0) {
      timerDisplay.style.color = '#f59e0b'; // Amarelo (2 min)
    } else if (timerSeconds <= 0) {
      timerDisplay.style.color = '#ef4444'; // Vermelho (tempo esgotado)
    } else {
      timerDisplay.style.color = '#38bdf8';
    }
  }

  function toggleTimer() {
    if (timerRunning) {
      clearInterval(timerInterval);
      timerRunning = false;
      timerBox.style.borderColor = 'rgba(51, 65, 85, 0.6)';
    } else {
      timerRunning = true;
      timerBox.style.borderColor = '#10b981';
      timerInterval = setInterval(() => {
        if (timerSeconds > 0) {
          timerSeconds--;
          updateTimerUI();
        } else {
          clearInterval(timerInterval);
          timerRunning = false;
        }
      }, 1000);
    }
  }

  if (timerBox) {
    timerBox.addEventListener('click', toggleTimer);
  }

  // ATUALIZAÇÃO DO SLIDE
  function showSlide(index) {
    if (index < 0 || index >= totalSlides) return;

    slides.forEach((s, idx) => {
      s.classList.toggle('active', idx === index);
    });

    currentSlide = index;

    // Atualiza barra de progresso
    const progress = ((index + 1) / totalSlides) * 100;
    progressBar.style.width = `${progress}%`;

    // Atualiza contador
    counterCurrent.textContent = String(index + 1).padStart(2, '0');

    // Atualiza estados dos botões
    btnPrev.disabled = index === 0;
    btnNext.disabled = index === totalSlides - 1;

    // Salva no LocalStorage
    try {
      localStorage.setItem('afya_slide_idx', index);
    } catch(e) {}

    updateDrawerActive();
  }

  function nextSlide() {
    if (currentSlide < totalSlides - 1) {
      showSlide(currentSlide + 1);
    }
  }

  function prevSlide() {
    if (currentSlide > 0) {
      showSlide(currentSlide - 1);
    }
  }

  btnPrev.addEventListener('click', prevSlide);
  btnNext.addEventListener('click', nextSlide);

  // TELA CHEIA (FULLSCREEN)
  function toggleFullscreen() {
    if (!document.fullscreenElement) {
      document.documentElement.requestFullscreen().catch(err => {
        console.error(err);
      });
    } else {
      if (document.exitFullscreen) {
        document.exitFullscreen();
      }
    }
  }

  if (btnFullscreen) {
    btnFullscreen.addEventListener('click', toggleFullscreen);
  }

  // GERAÇÃO DO ÍNDICE DINÂMICO NO DRAWER
  slides.forEach((slide, idx) => {
    const titleEl = slide.querySelector('.slide-title');
    const badgeEl = slide.querySelector('.badge');
    const titleText = titleEl ? titleEl.innerText.replace(/\n/g, ' ') : `Slide ${idx + 1}`;
    const badgeText = badgeEl ? badgeEl.innerText : `ETAPA ${idx + 1}`;

    const item = document.createElement('div');
    item.className = 'drawer-item';
    item.dataset.index = idx;
    item.innerHTML = `
      <div class="drawer-num">${String(idx + 1).padStart(2, '0')} • ${badgeText}</div>
      <div class="drawer-name">${titleText}</div>
    `;

    item.addEventListener('click', () => {
      showSlide(idx);
      closeDrawer();
    });

    drawerGrid.appendChild(item);
  });

  function updateDrawerActive() {
    const items = drawerGrid.querySelectorAll('.drawer-item');
    items.forEach((it, idx) => {
      it.classList.toggle('current', idx === currentSlide);
    });
  }

  function openDrawer() {
    drawer.classList.add('active');
    updateDrawerActive();
  }

  function closeDrawer() {
    drawer.classList.remove('active');
  }

  if (btnMap) btnMap.addEventListener('click', openDrawer);
  if (btnCloseDrawer) btnCloseDrawer.addEventListener('click', closeDrawer);
  drawer.addEventListener('click', (e) => {
    if (e.target === drawer) closeDrawer();
  });

  // NAVEGAÇÃO POR TECLADO
  document.addEventListener('keydown', (e) => {
    if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA') return;

    switch (e.key) {
      case 'ArrowRight':
      case 'PageDown':
      case ' ':
      case 'l':
        e.preventDefault();
        nextSlide();
        break;

      case 'ArrowLeft':
      case 'PageUp':
      case 'Backspace':
      case 'h':
        e.preventDefault();
        prevSlide();
        break;

      case 'Home':
        e.preventDefault();
        showSlide(0);
        break;

      case 'End':
        e.preventDefault();
        showSlide(totalSlides - 1);
        break;

      case 'f':
      case 'F':
        toggleFullscreen();
        break;

      case 't':
      case 'T':
        toggleTimer();
        break;

      case 'm':
      case 'M':
      case 'Escape':
        if (drawer.classList.contains('active')) {
          closeDrawer();
        } else if (e.key !== 'Escape') {
          openDrawer();
        }
        break;
    }
  });

  // SUPORTE A SWIPE TOUCH EM SMARTPHONES E TABLETS
  let touchStartX = 0;
  let touchEndX = 0;

  document.addEventListener('touchstart', (e) => {
    touchStartX = e.changedTouches[0].screenX;
  }, false);

  document.addEventListener('touchend', (e) => {
    touchEndX = e.changedTouches[0].screenX;
    handleGesture();
  }, false);

  function handleGesture() {
    const delta = touchEndX - touchStartX;
    if (Math.abs(delta) > 50) {
      if (delta < 0) {
        nextSlide();
      } else {
        prevSlide();
      }
    }
  }

  // RECUPERA SLIDE SALVO OU INICIA NO PRIMEIRO
  let savedIdx = 0;
  try {
    const val = localStorage.getItem('afya_slide_idx');
    if (val !== null) savedIdx = parseInt(val, 10);
    if (isNaN(savedIdx) || savedIdx < 0 || savedIdx >= totalSlides) savedIdx = 0;
  } catch(e) {}

  showSlide(savedIdx);
  updateTimerUI();
});
