// Estado global da aplicação
const state = {
  user: null,
  courses: [],
  selectedCourseId: 'all',
  assignments: {}, // courseId -> array
  currentFilter: 'all',
  students: [],
  theme: localStorage.getItem('canvas_hub_theme') || 'dark',
  chatHistory: JSON.parse(localStorage.getItem('afya_chat_history') || '[]'),
  isGenerating: false,
  agentInfo: null
};

// Cache e estado da aba de Gráficos & Métricas (Analytics)
const analyticsCache = {};
const analyticsState = {
  selectedCourseId: null,
  summary: null,
  charts: {
    donut: null,
    speedgrader: null,
    histogram: null,
    scatter: null
  }
};

// Configura tema inicial
document.documentElement.setAttribute('data-theme', state.theme);

// Elementos DOM principais
const elements = {
  connectionStatus: document.getElementById('connectionStatus'),
  userProfile: document.getElementById('userProfile'),
  themeToggle: document.getElementById('themeToggle'),
  
  totalCoursesCount: document.getElementById('totalCoursesCount'),
  totalAssignmentsCount: document.getElementById('totalAssignmentsCount'),
  totalPendingCount: document.getElementById('totalPendingCount'),

  globalCourseSelect: document.getElementById('globalCourseSelect'),
  chatCourseSelect: document.getElementById('chatCourseSelect'),
  courseSelect: document.getElementById('courseSelect'),
  refreshDataBtn: document.getElementById('refreshDataBtn'),
  coursesCardsGrid: document.getElementById('coursesCardsGrid'),
  assignmentsList: document.getElementById('assignmentsList'),
  assignmentsTitle: document.getElementById('assignmentsTitle'),
  assignmentsSubtitle: document.getElementById('assignmentsSubtitle'),

  announcementForm: document.getElementById('announcementForm'),
  announcementCoursesList: document.getElementById('announcementCoursesList'),
  selectAllCoursesBtn: document.getElementById('selectAllCoursesBtn'),
  unselectAllCoursesBtn: document.getElementById('unselectAllCoursesBtn'),
  announcementTitle: document.getElementById('announcementTitle'),
  announcementMessage: document.getElementById('announcementMessage'),
  sendAnnouncementBtn: document.getElementById('sendAnnouncementBtn'),
  announcementHistoryCourseSelect: document.getElementById('announcementHistoryCourseSelect'),
  announcementsHistoryList: document.getElementById('announcementsHistoryList'),

  studentsCourseSelect: document.getElementById('studentsCourseSelect'),
  studentsSearchInput: document.getElementById('studentsSearchInput'),
  studentsCounter: document.getElementById('studentsCounter'),
  studentsTableBody: document.getElementById('studentsTableBody'),
  exportCsvBtn: document.getElementById('exportCsvBtn'),

  // Elementos do Assistente IA (Chat)
  chatMessages: document.getElementById('chatMessages'),
  chatThinking: document.getElementById('chatThinking'),
  thinkingStatusText: document.getElementById('thinkingStatusText'),
  chatActionCard: document.getElementById('chatActionCard'),
  chatForm: document.getElementById('chatForm'),
  chatInput: document.getElementById('chatInput'),
  btnSendChat: document.getElementById('btnSendChat'),
  btnNewChat: document.getElementById('btnNewChat'),
  aiProviderBadge: document.getElementById('aiProviderBadge'),
  aiModelName: document.getElementById('aiModelName'),
  aiCanvasStatus: document.getElementById('aiCanvasStatus'),
  chatMetaInfo: document.getElementById('chatMetaInfo'),
  chatResponseTime: document.getElementById('chatResponseTime'),
  chatActiveModel: document.getElementById('chatActiveModel'),

  toastContainer: document.getElementById('toastContainer')
};

// ==========================================
// Utilitários
// ==========================================
function showToast(message, type = 'success') {
  const toast = document.createElement('div');
  toast.className = `toast ${type}`;
  toast.innerHTML = `
    <span>${type === 'success' ? '✓' : '⚠️'}</span>
    <span>${message}</span>
  `;
  elements.toastContainer.appendChild(toast);
  setTimeout(() => {
    toast.style.opacity = '0';
    setTimeout(() => toast.remove(), 250);
  }, 4000);
}

function formatDate(isoString) {
  if (!isoString) return 'Sem data limite';
  const date = new Date(isoString);
  return new Intl.DateTimeFormat('pt-BR', {
    day: '2-digit',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit'
  }).format(date);
}

// ==========================================
// Alternador de Tema
// ==========================================
elements.themeToggle.addEventListener('click', () => {
  state.theme = state.theme === 'dark' ? 'light' : 'dark';
  document.documentElement.setAttribute('data-theme', state.theme);
  localStorage.setItem('canvas_hub_theme', state.theme);
});

// ==========================================
// Navegação em Abas
// ==========================================
document.querySelectorAll('.tab-btn').forEach(btn => {
  btn.addEventListener('click', () => {
    document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
    document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));

    btn.classList.add('active');
    const targetId = btn.getAttribute('data-tab');
    document.getElementById(targetId).classList.add('active');

    // Carregamento sob demanda para a aba de alunos se necessário
    if (targetId === 'students-tab' && elements.studentsCourseSelect.value) {
      loadStudentsForCourse(elements.studentsCourseSelect.value);
    }

    // Carregamento da aba de Gráficos e Analytics
    if (targetId === 'analytics-tab') {
      loadAnalyticsTab();
      setTimeout(() => {
        Object.values(analyticsState.charts).forEach(c => {
          if (c) {
            c.resize();
            c.update('none');
          }
        });
      }, 280);
    }
  });
});

// ==========================================
// Carregamento de Dados da API
// ==========================================
async function initApp() {
  try {
    elements.connectionStatus.innerHTML = `
      <span class="status-dot" style="background: #f59e0b; box-shadow: 0 0 8px #f59e0b;"></span>
      <span class="status-text" style="color: #f59e0b;">Conectando...</span>
    `;

    // 1. Carrega Perfil
    const userResp = await fetch('/api/me');
    if (!userResp.ok) throw new Error('Falha ao autenticar com Canvas API');
    state.user = await userResp.json();
    renderUserProfile(state.user);

    // 2. Carrega Cursos
    const coursesResp = await fetch('/api/courses');
    if (!coursesResp.ok) throw new Error('Falha ao obter lista de cursos');
    state.courses = await coursesResp.json();

    // Pré-seleciona a turma salva no localStorage, ou a primeira do semestre vigente, ou a primeira da lista
    const savedCourseId = localStorage.getItem('canvas_hub_active_course');
    const currentCourses = state.courses.filter(c => c.is_current_term);

    if (savedCourseId && (savedCourseId === 'all' || state.courses.some(c => String(c.id) === savedCourseId))) {
      state.selectedCourseId = savedCourseId;
    } else if (currentCourses.length > 0) {
      state.selectedCourseId = String(currentCourses[0].id);
    } else if (state.courses.length > 0) {
      state.selectedCourseId = String(state.courses[0].id);
    }

    // Atualiza status de conexão imediatamente
    elements.connectionStatus.innerHTML = `
      <span class="status-dot"></span>
      <span class="status-text">Conectado ao Canvas</span>
    `;

    // Renderiza cursos e popula todos os seletores sincronizados
    populateSelects();
    renderCourses();
    selectCourse(state.selectedCourseId);

    // Carrega detalhes de tarefas em segundo plano para não travar a interface
    loadAllAssignments();

  } catch (err) {
    console.error('Erro na inicialização:', err);
    elements.connectionStatus.innerHTML = `
      <span class="status-dot" style="background: #ef4444; box-shadow: 0 0 8px #ef4444;"></span>
      <span class="status-text" style="color: #ef4444;">Erro de Conexão</span>
    `;
    showToast('Não foi possível conectar com a API do Canvas: ' + err.message, 'error');
  }
}

function renderUserProfile(user) {
  const avatar = user.avatar_url 
    ? `<img src="${user.avatar_url}" alt="${user.name}" class="user-avatar">`
    : `<div class="avatar-skeleton"></div>`;

  elements.userProfile.innerHTML = `
    ${avatar}
    <div class="user-info">
      <span class="user-name" title="${user.name}">${user.name}</span>
      <span class="user-role">Professor(a)</span>
    </div>
  `;
}

function populateSelects() {
  const currentCourses = state.courses.filter(c => c.is_current_term);
  const pastCourses = state.courses.filter(c => !c.is_current_term);

  // Helper para preencher um select com optgroups
  const appendOptGroup = (selectEl, label, list) => {
    if (!selectEl || list.length === 0) return;
    const group = document.createElement('optgroup');
    group.label = label;
    list.forEach(course => {
      const opt = document.createElement('option');
      opt.value = String(course.id);
      const period = course.period ? ` (${course.period})` : '';
      opt.textContent = `${course.clean_name || course.name}${period}`;
      group.appendChild(opt);
    });
    selectEl.appendChild(group);
  };

  // 1. Seletor Global no Header
  if (elements.globalCourseSelect) {
    elements.globalCourseSelect.innerHTML = '<option value="all">Todas as Disciplinas (Visão Geral)</option>';
    appendOptGroup(elements.globalCourseSelect, '🟢 Semestre Vigente (Atual)', currentCourses);
    appendOptGroup(elements.globalCourseSelect, '⚪ Semestres Anteriores', pastCourses);
  }

  // 2. Seletor no Topo do Chat
  if (elements.chatCourseSelect) {
    elements.chatCourseSelect.innerHTML = '<option value="all">Todas as Disciplinas</option>';
    appendOptGroup(elements.chatCourseSelect, '🟢 Semestre Vigente (Atual)', currentCourses);
    appendOptGroup(elements.chatCourseSelect, '⚪ Semestres Anteriores', pastCourses);
  }

  // 3. Seletor de Cursos na aba 1
  if (elements.courseSelect) {
    elements.courseSelect.innerHTML = '<option value="all">Todas as Disciplinas</option>';
    appendOptGroup(elements.courseSelect, '🟢 Semestre Vigente (Atual)', currentCourses);
    appendOptGroup(elements.courseSelect, '⚪ Semestres Anteriores (Concluídos)', pastCourses);
  }

  // 4. Seletor de histórico de avisos na aba 2
  if (elements.announcementHistoryCourseSelect) {
    elements.announcementHistoryCourseSelect.innerHTML = '';
    appendOptGroup(elements.announcementHistoryCourseSelect, '🟢 Semestre Vigente', currentCourses);
    appendOptGroup(elements.announcementHistoryCourseSelect, '⚪ Semestres Anteriores', pastCourses);
  }

  // 5. Seletor de Alunos na aba 3
  if (elements.studentsCourseSelect) {
    elements.studentsCourseSelect.innerHTML = '';
    appendOptGroup(elements.studentsCourseSelect, '🟢 Semestre Vigente', currentCourses);
    appendOptGroup(elements.studentsCourseSelect, '⚪ Semestres Anteriores', pastCourses);
  }

  // 6. Seletor de Gráficos na aba 4 (com formatação rica e optgroups)
  if (typeof populateAnalyticsCourseSelect === 'function') {
    populateAnalyticsCourseSelect(true);
  }

  // Checkboxes de envio de avisos
  if (elements.announcementCoursesList) {
    elements.announcementCoursesList.innerHTML = '';
    state.courses.forEach(course => {
      const isCur = course.is_current_term;
      const checkItem = document.createElement('label');
      checkItem.className = 'checkbox-item';
      const periodText = course.period || (isCur ? 'Semestre Vigente' : 'Semestre Concluído');
      const studentsText = course.total_students ? ` • ${course.total_students} alunos` : '';
      checkItem.innerHTML = `
        <input type="checkbox" name="announcement_course" value="${course.id}" ${isCur ? 'checked' : ''}>
        <span>
          <strong>${isCur ? '🟢 ' : '⚪ '}${course.clean_name || course.name}</strong>
          <small style="color: var(--text-muted); font-size: 0.74rem;">${periodText}${studentsText}</small>
        </span>
      `;
      elements.announcementCoursesList.appendChild(checkItem);
    });
  }

  if (currentCourses.length > 0) {
    loadAnnouncementsForCourse(currentCourses[0].id);
  } else if (state.courses.length > 0) {
    loadAnnouncementsForCourse(state.courses[0].id);
  }
}

// Filtro de semestres ativo (padrão: "current")
state.termFilter = 'current';
state.assignmentSearchQuery = '';

function createCourseCard(course) {
  const card = document.createElement('div');
  const isCur = course.is_current_term;
  card.className = `course-card ${isCur ? 'current-card' : 'past-card'} ${state.selectedCourseId === String(course.id) ? 'selected' : ''}`;
  card.setAttribute('data-id', course.id);

  const students = course.total_students !== undefined ? course.total_students : 0;
  const courseAssigns = state.assignments[course.id] || [];
  let pendingCount = 0;
  courseAssigns.forEach(a => {
    pendingCount += (a.needs_grading_count || 0);
  });

  let pendingBadgeHtml = '';
  if (courseAssigns.length > 0) {
    if (pendingCount > 0) {
      pendingBadgeHtml = `<div class="card-pending-badge has-pending"><span>⚠️ Submissões pendentes:</span> <strong>${pendingCount}</strong></div>`;
    } else {
      pendingBadgeHtml = `<div class="card-pending-badge all-graded"><span>✓ Avaliações:</span> <strong>100% em dia</strong></div>`;
    }
  }

  card.innerHTML = `
    <div>
      <div class="card-top-bar">
        <span class="card-period-tag">${course.period || (isCur ? 'Semestre Vigente' : 'Concluído')}</span>
        <span class="card-id-tag">ID: ${course.id}</span>
      </div>
      <h3 class="course-clean-title">${course.clean_name || course.name}</h3>
      <div class="course-raw-name">${course.name}</div>
    </div>
    <div>
      <div class="card-metrics-grid">
        <div class="card-metric-pill">
          <span>👥</span> <span><strong>${students}</strong> alunos</span>
        </div>
        <div class="card-metric-pill">
          <span>📋</span> <span><strong>${courseAssigns.length}</strong> atividades</span>
        </div>
        ${pendingBadgeHtml}
      </div>
    </div>
  `;

  card.addEventListener('click', () => {
    selectCourse(course.id);
  });

  return card;
}

function renderCourses() {
  const currentCourses = state.courses.filter(c => c.is_current_term);
  const pastCourses = state.courses.filter(c => !c.is_current_term);

  // Atualiza contadores nas pills
  const pillCur = document.getElementById('pillCurrentCount');
  const pillPst = document.getElementById('pillPastCount');
  const pillAll = document.getElementById('pillAllCount');
  if (pillCur) pillCur.textContent = currentCourses.length;
  if (pillPst) pillPst.textContent = pastCourses.length;
  if (pillAll) pillAll.textContent = state.courses.length;
  if (elements.totalCoursesCount) elements.totalCoursesCount.textContent = currentCourses.length;

  const container = document.getElementById('coursesCardsGrid');
  if (!container) return;
  container.innerHTML = '';

  // Renderiza conforme o filtro de semestre selecionado
  if (state.termFilter === 'current') {
    const group = document.createElement('div');
    group.className = 'courses-group-section';
    const totalStudents = currentCourses.reduce((acc, c) => acc + (c.total_students || 0), 0);

    group.innerHTML = `
      <div class="courses-group-header">
        <div class="cgh-left">
          <span class="pill-dot"></span>
          <h3 class="cgh-title">Disciplinas do Semestre Vigente</h3>
          <span class="cgh-tag current">Em Andamento</span>
        </div>
        <div class="cgh-meta">${currentCourses.length} turmas • ${totalStudents} alunos matriculados</div>
      </div>
      <div class="courses-cards-grid" id="gridCurrentCourses"></div>
    `;
    container.appendChild(group);

    const grid = group.querySelector('#gridCurrentCourses');
    currentCourses.forEach(course => grid.appendChild(createCourseCard(course)));

  } else if (state.termFilter === 'past') {
    const group = document.createElement('div');
    group.className = 'courses-group-section';
    const totalStudents = pastCourses.reduce((acc, c) => acc + (c.total_students || 0), 0);

    group.innerHTML = `
      <div class="courses-group-header">
        <div class="cgh-left">
          <span class="pill-dot past"></span>
          <h3 class="cgh-title">Disciplinas de Semestres Anteriores</h3>
          <span class="cgh-tag past">Concluídas / Histórico</span>
        </div>
        <div class="cgh-meta">${pastCourses.length} turmas • ${totalStudents} alunos</div>
      </div>
      <div class="courses-cards-grid" id="gridPastCourses"></div>
    `;
    container.appendChild(group);

    const grid = group.querySelector('#gridPastCourses');
    pastCourses.forEach(course => grid.appendChild(createCourseCard(course)));

  } else {
    // "all": Mostra as duas seções claramente separadas
    if (currentCourses.length > 0) {
      const groupCur = document.createElement('div');
      groupCur.className = 'courses-group-section';
      const totalStudents = currentCourses.reduce((acc, c) => acc + (c.total_students || 0), 0);

      groupCur.innerHTML = `
        <div class="courses-group-header">
          <div class="cgh-left">
            <span class="pill-dot"></span>
            <h3 class="cgh-title">Semestre Vigente (Atual)</h3>
            <span class="cgh-tag current">Em Andamento</span>
          </div>
          <div class="cgh-meta">${currentCourses.length} turmas • ${totalStudents} alunos</div>
        </div>
        <div class="courses-cards-grid" id="gridAllCurrent"></div>
      `;
      container.appendChild(groupCur);
      const gridCur = groupCur.querySelector('#gridAllCurrent');
      currentCourses.forEach(course => gridCur.appendChild(createCourseCard(course)));
    }

    if (pastCourses.length > 0) {
      const groupPast = document.createElement('div');
      groupPast.className = 'courses-group-section';
      const totalStudents = pastCourses.reduce((acc, c) => acc + (c.total_students || 0), 0);

      groupPast.innerHTML = `
        <div class="courses-group-header" style="margin-top: 1rem;">
          <div class="cgh-left">
            <span class="pill-dot past"></span>
            <h3 class="cgh-title">Semestres Anteriores</h3>
            <span class="cgh-tag past">Concluídas</span>
          </div>
          <div class="cgh-meta">${pastCourses.length} turmas • ${totalStudents} alunos</div>
        </div>
        <div class="courses-cards-grid" id="gridAllPast"></div>
      `;
      container.appendChild(groupPast);
      const gridPast = groupPast.querySelector('#gridAllPast');
      pastCourses.forEach(course => gridPast.appendChild(createCourseCard(course)));
    }
  }

  updateCourseBanner();
}

function updateCourseBanner() {
  const scTitle = document.getElementById('scTitle');
  const scSubtitle = document.getElementById('scSubtitle');
  const scPeriodBadge = document.getElementById('scPeriodBadge');
  const scTermBadge = document.getElementById('scTermBadge');
  const scIdBadge = document.getElementById('scIdBadge');
  const btnOpenCanvasCourse = document.getElementById('btnOpenCanvasCourse');
  const btnAskAIAboutCourse = document.getElementById('btnAskAIAboutCourse');

  if (!scTitle) return;

  if (state.selectedCourseId === 'all') {
    scTitle.textContent = 'Todas as Disciplinas';
    scSubtitle.textContent = 'Visualizando tarefas consolidadas de todas as turmas cadastradas';
    if (scPeriodBadge) scPeriodBadge.style.display = 'none';
    if (scTermBadge) {
      scTermBadge.textContent = 'Visão Global';
      scTermBadge.className = 'sc-badge-term';
    }
    if (scIdBadge) scIdBadge.textContent = `${state.courses.length} turmas`;
    if (btnOpenCanvasCourse) btnOpenCanvasCourse.style.display = 'none';
    if (btnAskAIAboutCourse) {
      btnAskAIAboutCourse.onclick = () => {
        sendPromptToChat('Apresente um panorama completo de todas as pendências e atividades das minhas turmas ativas no Canvas LMS.');
      };
    }
  } else {
    const course = state.courses.find(c => String(c.id) === state.selectedCourseId);
    if (!course) return;

    scTitle.textContent = course.clean_name || course.name;
    scSubtitle.textContent = `${course.name} • ${course.total_students || 0} alunos matriculados`;
    if (scPeriodBadge) {
      scPeriodBadge.style.display = 'inline-flex';
      scPeriodBadge.textContent = course.period || 'Graduação';
    }
    if (scTermBadge) {
      scTermBadge.textContent = course.is_current_term ? '🟢 Semestre Atual' : '⚪ Semestre Anterior';
      scTermBadge.className = `sc-badge-term ${course.is_current_term ? 'current' : 'past'}`;
    }
    if (scIdBadge) scIdBadge.textContent = `ID Canvas: ${course.id}`;
    if (btnOpenCanvasCourse) {
      btnOpenCanvasCourse.style.display = 'inline-flex';
      btnOpenCanvasCourse.href = `https://afya.instructure.com/courses/${course.id}`;
    }
    if (btnAskAIAboutCourse) {
      btnAskAIAboutCourse.onclick = () => {
        sendPromptToChat(`Analise a situação da disciplina '${course.clean_name || course.name}' (ID ${course.id}) e liste as atividades e notas pendentes.`);
      };
    }
  }
}

function selectCourse(courseId) {
  if (!courseId) return;
  state.selectedCourseId = String(courseId);
  localStorage.setItem('canvas_hub_active_course', state.selectedCourseId);

  // Sincroniza o valor em todos os seletores da aplicação
  const syncSelects = [
    elements.globalCourseSelect,
    elements.chatCourseSelect,
    elements.courseSelect,
    elements.studentsCourseSelect,
    elements.announcementHistoryCourseSelect,
    document.getElementById('analyticsCourseSelect')
  ];

  syncSelects.forEach(sel => {
    if (sel && sel.value !== state.selectedCourseId) {
      const hasOption = Array.from(sel.querySelectorAll('option')).some(o => o.value === state.selectedCourseId);
      if (hasOption) {
        sel.value = state.selectedCourseId;
      }
    }
  });

  // Atualiza cartões de turmas
  document.querySelectorAll('.course-card').forEach(card => {
    if (card.getAttribute('data-id') === state.selectedCourseId) {
      card.classList.add('selected');
    } else {
      card.classList.remove('selected');
    }
  });

  updateCourseBanner();
  renderAssignmentsList();

  // Carrega dados contextuais sob demanda para a aba que estiver aberta
  if (state.selectedCourseId !== 'all') {
    const activeTab = document.querySelector('.tab-content.active');
    if (activeTab) {
      if (activeTab.id === 'students-tab') {
        loadStudentsForCourse(state.selectedCourseId);
      } else if (activeTab.id === 'analytics-tab' && typeof fetchAndRenderAnalytics === 'function') {
        analyticsState.selectedCourseId = state.selectedCourseId;
        fetchAndRenderAnalytics(state.selectedCourseId);
      } else if (activeTab.id === 'announcements-tab') {
        loadAnnouncementsForCourse(state.selectedCourseId);
      }
    }
  }
}

// Envia um prompt para o Assistente IA e troca automaticamente para a aba do chat
function sendPromptToChat(promptText) {
  // Alterna para a aba do chat
  document.querySelectorAll('.tab-btn').forEach(b => {
    if (b.getAttribute('data-tab') === 'chat-tab') b.classList.add('active');
    else b.classList.remove('active');
  });
  document.querySelectorAll('.tab-content').forEach(c => {
    if (c.id === 'chat-tab') c.classList.add('active');
    else c.classList.remove('active');
  });

  if (elements.chatInput) {
    elements.chatInput.value = promptText;
    elements.chatInput.focus();
    if (elements.chatForm) {
      elements.chatForm.dispatchEvent(new Event('submit'));
    }
  }
}

async function loadAllAssignments() {
  let totalAssignments = 0;
  let totalPending = 0;

  for (const course of state.courses) {
    try {
      const resp = await fetch(`/api/courses/${course.id}/assignments`);
      if (resp.ok) {
        const assignments = await resp.json();
        state.assignments[course.id] = assignments;

        // Atribui o nome do curso a cada tarefa para referência
        assignments.forEach(a => {
          a.course_clean_name = course.clean_name || course.name;
          a.course_period = course.period || '';
          a.is_current_term = course.is_current_term;
          totalPending += (a.needs_grading_count || 0);
        });

        totalAssignments += assignments.length;
      }
    } catch (e) {
      console.error(`Erro ao carregar tarefas do curso ${course.id}:`, e);
    }
  }

  elements.totalAssignmentsCount.textContent = totalAssignments;
  elements.totalPendingCount.textContent = totalPending;

  // Atualiza os cards com as métricas de tarefas carregadas
  renderCourses();
  renderAssignmentsList();
}

function renderAssignmentsList() {
  let list = [];
  if (state.selectedCourseId === 'all') {
    elements.assignmentsTitle.textContent = 'Atividades Consolidadas';
    elements.assignmentsSubtitle.textContent = 'Visualizando tarefas de todas as suas turmas';
    for (const courseId in state.assignments) {
      list.push(...(state.assignments[courseId] || []));
    }
  } else {
    const course = state.courses.find(c => String(c.id) === state.selectedCourseId);
    elements.assignmentsTitle.textContent = `Atividades: ${course ? (course.clean_name || course.name) : ''}`;
    elements.assignmentsSubtitle.textContent = course ? course.name : 'Tarefas cadastradas';
    list = state.assignments[state.selectedCourseId] || [];
  }

  // Contadores para os chips
  const totalCount = list.length;
  const pendingCount = list.filter(a => (a.needs_grading_count || 0) > 0).length;
  const doneCount = list.filter(a => (a.needs_grading_count || 0) === 0).length;

  const countAllEl = document.getElementById('countFilterAll');
  const countPendEl = document.getElementById('countFilterPending');
  const countDoneEl = document.getElementById('countFilterDone');
  if (countAllEl) countAllEl.textContent = totalCount;
  if (countPendEl) countPendEl.textContent = pendingCount;
  if (countDoneEl) countDoneEl.textContent = doneCount;

  // Filtro por busca de texto
  if (state.assignmentSearchQuery) {
    const q = state.assignmentSearchQuery.toLowerCase();
    list = list.filter(a => (a.name || '').toLowerCase().includes(q));
  }

  // Filtro por chip
  if (state.currentFilter === 'pending') {
    list = list.filter(a => (a.needs_grading_count || 0) > 0);
  } else if (state.currentFilter === 'done') {
    list = list.filter(a => (a.needs_grading_count || 0) === 0);
  }

  if (list.length === 0) {
    elements.assignmentsList.innerHTML = `
      <div class="empty-state">
        <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>
        <p>Nenhuma atividade encontrada para os critérios selecionados.</p>
      </div>
    `;
    return;
  }

  elements.assignmentsList.innerHTML = '';
  list.forEach(item => {
    const pending = item.needs_grading_count || 0;
    const isPending = pending > 0;
    const dueDate = formatDate(item.due_at);
    const canvasUrl = item.html_url || `https://afya.instructure.com/courses/${item.course_id}/assignments/${item.id}`;

    const div = document.createElement('div');
    div.className = 'assignment-item';

    const courseTagHtml = state.selectedCourseId === 'all' && item.course_clean_name
      ? `<div class="assignment-course-tag">${item.course_clean_name} ${item.course_period ? '• ' + item.course_period : ''}</div>`
      : '';

    const btnAIEvaluate = isPending
      ? `<button class="btn-ai-action" data-assign-id="${item.id}" data-course-id="${item.course_id}" data-name="${item.name}">
           <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="m12 3-1.912 5.813a2 2 0 0 1-1.275 1.275L3 12l5.813 1.912a2 2 0 0 1 1.275 1.275L12 21l1.912-5.813a2 2 0 0 1 1.275-1.275L21 12l-5.813-1.912a2 2 0 0 1-1.275-1.275L12 3Z"/></svg>
           Corrigir com IA
         </button>`
      : '';

    div.innerHTML = `
      <div class="assignment-info">
        ${courseTagHtml}
        <h4 class="assignment-title">${item.name}</h4>
        <div class="assignment-meta">
          <span>📅 Entrega: <strong>${dueDate}</strong></span>
          <span>🏆 Pontos: <strong>${item.points_possible !== null && item.points_possible !== undefined ? item.points_possible : 'N/A'}</strong></span>
          ${item.submission_types ? `<span>📝 Formato: ${item.submission_types.join(', ')}</span>` : ''}
        </div>
      </div>
      <div class="assignment-actions">
        <span class="grading-badge ${isPending ? 'pending' : 'done'}">
          ${isPending ? `⚠️ ${pending} para corrigir` : '✓ 100% Corrigidas'}
        </span>
        <button class="btn btn-secondary btn-view-code" data-assign-id="${item.id}" data-course-id="${item.course_id}" data-name="${item.name}" data-points="${item.points_possible || ''}" style="padding: 0.4rem 0.75rem; font-size: 0.8125rem;">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="16 18 22 12 16 6"></polyline><polyline points="8 6 2 12 8 18"></polyline></svg>
          Ver Códigos
        </button>
        ${btnAIEvaluate}
        <a href="${canvasUrl}" target="_blank" rel="noopener noreferrer" class="btn btn-secondary" style="padding: 0.4rem 0.75rem; font-size: 0.8125rem;">
          Ver no Canvas
        </a>
      </div>
    `;

    // Evento do botão Ver Códigos
    const viewCodeBtn = div.querySelector('.btn-view-code');
    if (viewCodeBtn) {
      viewCodeBtn.addEventListener('click', (e) => {
        e.stopPropagation();
        const aName = viewCodeBtn.getAttribute('data-name');
        const cId = viewCodeBtn.getAttribute('data-course-id');
        const aId = viewCodeBtn.getAttribute('data-assign-id');
        const pts = viewCodeBtn.getAttribute('data-points');
        openCodeViewerModal(cId, aId, aName, pts);
      });
    }

    // Evento do botão Corrigir com IA
    const aiBtn = div.querySelector('.btn-ai-action');
    if (aiBtn) {
      aiBtn.addEventListener('click', (e) => {
        e.stopPropagation();
        const aName = aiBtn.getAttribute('data-name');
        const cId = aiBtn.getAttribute('data-course-id');
        const aId = aiBtn.getAttribute('data-assign-id');
        sendPromptToChat(`Inicie a preparação e avaliação das submissões pendentes da atividade '${aName}' (ID ${aId}) da disciplina ID ${cId}.`);
      });
    }

    elements.assignmentsList.appendChild(div);
  });
}

// Listeners de Semestre (Pills)
document.querySelectorAll('.semester-pill').forEach(pill => {
  pill.addEventListener('click', () => {
    document.querySelectorAll('.semester-pill').forEach(p => p.classList.remove('active'));
    pill.classList.add('active');
    state.termFilter = pill.getAttribute('data-term');
    renderCourses();
  });
});

// Listener de busca em tempo real de atividades
const assignSearchInput = document.getElementById('assignmentSearchInput');
if (assignSearchInput) {
  assignSearchInput.addEventListener('input', (e) => {
    state.assignmentSearchQuery = e.target.value.trim();
    renderAssignmentsList();
  });
}

// Filtros de Atividades (Chips)
document.querySelectorAll('.filter-chip').forEach(chip => {
  chip.addEventListener('click', () => {
    document.querySelectorAll('.filter-chip').forEach(c => c.classList.remove('active'));
    chip.classList.add('active');
    state.currentFilter = chip.getAttribute('data-filter');
    renderAssignmentsList();
  });
});

elements.globalCourseSelect?.addEventListener('change', (e) => {
  selectCourse(e.target.value);
  showToast('Turma ativa alterada em todo o portal!', 'info');
});

elements.chatCourseSelect?.addEventListener('change', (e) => {
  selectCourse(e.target.value);
  showToast('Contexto do chat direcionado para a turma selecionada.', 'info');
});

elements.courseSelect?.addEventListener('change', (e) => {
  selectCourse(e.target.value);
});

elements.refreshDataBtn.addEventListener('click', async () => {
  elements.refreshDataBtn.disabled = true;
  elements.refreshDataBtn.innerHTML = 'Carregando...';
  await initApp();
  elements.refreshDataBtn.disabled = false;
  elements.refreshDataBtn.innerHTML = `
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"></polyline><polyline points="1 20 1 14 7 14"></polyline><path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"></path></svg>
    Atualizar
  `;
  showToast('Dados atualizados com sucesso!');
});

// ==========================================
// Aba 2: Comunicados (Announcements)
// ==========================================
const ANNOUNCEMENT_TEMPLATES = {
  reminder: {
    title: '[Lembrete de Prazo] Entrega da Atividade Prática no Canvas LMS',
    message: `Prezados alunos,\n\nLembramos que o prazo final para envio da atividade prática no Canvas LMS encerra-se em breve. Recomendamos que realizem o envio com antecedência para evitar eventuais instabilidades.\n\nCritérios e orientações:\n- Submissão exclusivamente pelo ambiente Canvas LMS;\n- Respeito aos padrões e requisitos definidos no enunciado.\n\nBons estudos a todos!`
  },
  exam: {
    title: '[Orientações Institucionais] Avaliação Presencial Escrita (ENADE)',
    message: `Prezados estudantes,\n\nConforme o calendário institucional e as diretrizes do CONSEPE, nossa avaliação presencial individual será aplicada no horário regular da aula.\n\nOrientações regimentais:\n- Prova estritamente individual e sem consulta;\n- Questões elaboradas no Modelo ENADE (texto-base contextualizado, situação-problema e distratores fundamentados);\n- Pontualidade obrigatória com tolerância máxima regimental.\n\nSucesso a todos!`
  },
  feedback: {
    title: '[Devolutiva de Avaliação] Notas e Feedbacks Disponibilizados',
    message: `Prezados alunos,\n\nAs notas e comentários pedagógicos da atividade foram publicados no Canvas LMS.\n\nConforme os regulamentos acadêmicos da Afya:\n- A devolutiva com gabarito comentado está aberta para conferência;\n- Eventual solicitação fundamentada de revisão pode ser solicitada em até 2 dias letivos.\n\nParabéns pelo empenho de todos!`
  },
  office: {
    title: '[Plantão de Dúvidas] Horário de Atendimento e Apoio Pedagógico',
    message: `Prezados alunos,\n\nInformamos que estarei disponível para atendimento individual e esclarecimento de dúvidas conceituais e práticas sobre os conteúdos ministrados.\n\nAproveitem o momento para sanar dificuldades nos tópicos da disciplina.\n\nAtenciosamente,\nProfessor Karan`
  }
};

// Templates rápidos
document.querySelectorAll('.template-chip').forEach(chip => {
  chip.addEventListener('click', () => {
    const key = chip.getAttribute('data-template');
    const template = ANNOUNCEMENT_TEMPLATES[key];
    if (template) {
      elements.announcementTitle.value = template.title;
      elements.announcementMessage.value = template.message;
      showToast('Modelo de comunicado aplicado com sucesso!');
    }
  });
});

function updateAudienceCounter() {
  const selectedCbs = Array.from(document.querySelectorAll('input[name="announcement_course"]:checked'));
  let totalStudents = 0;
  selectedCbs.forEach(cb => {
    const course = state.courses.find(c => String(c.id) === cb.value);
    if (course) totalStudents += (course.total_students || 0);
  });
  const counter = document.getElementById('audienceCounter');
  if (counter) {
    counter.textContent = `👥 Alcance: ${totalStudents} alunos (${selectedCbs.length} turma${selectedCbs.length === 1 ? '' : 's'})`;
  }
}

// Botões de seleção rápida de destinatários
const btnSelectCurCourses = document.getElementById('btnSelectCurrentTermCourses');
if (btnSelectCurCourses) {
  btnSelectCurCourses.addEventListener('click', () => {
    document.querySelectorAll('input[name="announcement_course"]').forEach(cb => {
      const course = state.courses.find(c => String(c.id) === cb.value);
      cb.checked = course ? course.is_current_term : false;
    });
    updateAudienceCounter();
  });
}

elements.selectAllCoursesBtn.addEventListener('click', () => {
  document.querySelectorAll('input[name="announcement_course"]').forEach(cb => cb.checked = true);
  updateAudienceCounter();
});

elements.unselectAllCoursesBtn.addEventListener('click', () => {
  document.querySelectorAll('input[name="announcement_course"]').forEach(cb => cb.checked = false);
  updateAudienceCounter();
});

// Listener nos checkboxes para recalcular audiência dinamicamente
elements.announcementCoursesList.addEventListener('change', () => {
  updateAudienceCounter();
});

// Refinar com IA
const btnEnhanceAnnouncementAI = document.getElementById('btnEnhanceAnnouncementAI');
if (btnEnhanceAnnouncementAI) {
  btnEnhanceAnnouncementAI.addEventListener('click', async () => {
    const text = elements.announcementMessage.value.trim();
    if (!text) {
      showToast('Digite uma mensagem primeiro para a IA refinar.', 'info');
      return;
    }

    btnEnhanceAnnouncementAI.disabled = true;
    btnEnhanceAnnouncementAI.innerHTML = `
      <div class="thinking-spinner" style="width: 12px; height: 12px; border-width: 1.5px;"></div>
      Refinando...
    `;

    try {
      const resp = await fetch('/api/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          message: `Atue como assistente acadêmico da Afya. Reescreva e aprimore o seguinte comunicado aos estudantes, garantindo tom formal, institucional, respeitoso, claro e sem exageros ou intimidade:\n\n${text}`,
          history: []
        })
      });

      if (resp.ok) {
        const data = await resp.json();
        if (data.reply) {
          elements.announcementMessage.value = data.reply;
          showToast('Comunicado refinado com sucesso pelo Assistente IA!');
        }
      } else {
        showToast('Não foi possível conectar ao Assistente IA.', 'error');
      }
    } catch (err) {
      showToast('Erro ao comunicar com a IA: ' + err.message, 'error');
    } finally {
      btnEnhanceAnnouncementAI.disabled = false;
      btnEnhanceAnnouncementAI.innerHTML = `
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="m12 3-1.912 5.813a2 2 0 0 1-1.275 1.275L3 12l5.813 1.912a2 2 0 0 1 1.275 1.275L12 21l1.912-5.813a2 2 0 0 1 1.275-1.275L21 12l-5.813-1.912a2 2 0 0 1-1.275-1.275L12 3Z"/></svg>
        Refinar Texto com IA
      `;
    }
  });
}

elements.announcementForm.addEventListener('submit', async (e) => {
  e.preventDefault();
  const selectedCheckboxes = Array.from(document.querySelectorAll('input[name="announcement_course"]:checked'));
  const courseIds = selectedCheckboxes.map(cb => cb.value);

  if (courseIds.length === 0) {
    showToast('Selecione pelo menos uma disciplina de destino!', 'error');
    return;
  }

  const title = elements.announcementTitle.value.trim();
  const message = elements.announcementMessage.value.trim();

  if (!title || !message) {
    showToast('Preencha título e mensagem.', 'error');
    return;
  }

  elements.sendAnnouncementBtn.disabled = true;
  elements.sendAnnouncementBtn.innerHTML = 'Enviando comunicado...';

  let successCount = 0;
  for (const courseId of courseIds) {
    try {
      const resp = await fetch(`/api/courses/${courseId}/announcements`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ title, message })
      });
      if (resp.ok) successCount++;
    } catch (err) {
      console.error(`Erro ao disparar para curso ${courseId}:`, err);
    }
  }

  elements.sendAnnouncementBtn.disabled = false;
  elements.sendAnnouncementBtn.innerHTML = `
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="22" y1="2" x2="11" y2="13"></line><polygon points="22 2 15 22 11 13 2 9 22 2"></polygon></svg>
    <span>Disparar Comunicado Oficial</span>
  `;

  if (successCount > 0) {
    showToast(`Comunicado publicado com sucesso em ${successCount} disciplina(s)!`);
    elements.announcementTitle.value = '';
    elements.announcementMessage.value = '';
    loadAnnouncementsForCourse(elements.announcementHistoryCourseSelect.value);
  } else {
    showToast('Ocorreu um erro ao publicar o comunicado.', 'error');
  }
});

elements.announcementHistoryCourseSelect.addEventListener('change', (e) => {
  selectCourse(e.target.value);
  loadAnnouncementsForCourse(e.target.value);
});

async function loadAnnouncementsForCourse(courseId) {
  if (!courseId) return;
  elements.announcementsHistoryList.innerHTML = '<div class="empty-state sm"><p>Carregando avisos...</p></div>';

  try {
    const resp = await fetch(`/api/courses/${courseId}/announcements`);
    if (!resp.ok) throw new Error('Falha ao obter avisos');
    const announcements = await resp.json();

    if (announcements.length === 0) {
      elements.announcementsHistoryList.innerHTML = `
        <div class="empty-state sm">
          <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path></svg>
          <p>Nenhum comunicado publicado nesta disciplina.</p>
        </div>
      `;
      return;
    }

    elements.announcementsHistoryList.innerHTML = '';
    announcements.forEach(a => {
      const item = document.createElement('div');
      item.className = 'history-item';
      const cleanMessage = (a.message || '').replace(/<[^>]*>?/gm, '').trim();
      const dateStr = formatDate(a.posted_at || a.created_at);
      item.innerHTML = `
        <div class="history-item-top">
          <h4 class="history-item-title">${a.title || 'Comunicado'}</h4>
          <span class="history-item-date">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"></circle><polyline points="12 6 12 12 16 14"></polyline></svg>
            ${dateStr}
          </span>
        </div>
        <p class="history-item-body">${cleanMessage || 'Sem conteúdo de texto'}</p>
        ${a.html_url ? `
          <div class="history-item-footer">
            <a href="${a.html_url}" target="_blank" rel="noopener noreferrer" class="history-item-link">
              Abrir no Canvas
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"></path><polyline points="15 3 21 3 21 9"></polyline><line x1="10" y1="14" x2="21" y2="3"></line></svg>
            </a>
          </div>
        ` : ''}
      `;
      elements.announcementsHistoryList.appendChild(item);
    });
  } catch (err) {
    elements.announcementsHistoryList.innerHTML = '<div class="empty-state sm"><p>Erro ao carregar histórico de avisos.</p></div>';
  }
}

// ==========================================
// Aba 3: Alunos & Exportação CSV
// ==========================================
elements.studentsCourseSelect.addEventListener('change', (e) => {
  selectCourse(e.target.value);
});

async function loadStudentsForCourse(courseId) {
  if (!courseId) return;
  elements.studentsTableBody.innerHTML = `
    <tr><td colspan="6" class="table-placeholder">Carregando alunos da turma...</td></tr>
  `;

  // Atualiza os mini cards analíticos da turma
  const course = state.courses.find(c => String(c.id) === String(courseId));
  const statTotal = document.getElementById('statTotalStudents');
  const statPeriod = document.getElementById('statPeriod');
  const statStatus = document.getElementById('statTermStatus');
  const statAssigns = document.getElementById('statTotalAssignments');

  if (course) {
    if (statTotal) statTotal.textContent = course.total_students || 0;
    if (statPeriod) statPeriod.textContent = course.period || 'Graduação';
    if (statStatus) statStatus.textContent = course.is_current_term ? '🟢 Semestre Atual' : '⚪ Concluído';
    const courseAssigns = state.assignments[course.id] || [];
    if (statAssigns) statAssigns.textContent = courseAssigns.length;
  }

  try {
    const resp = await fetch(`/api/courses/${courseId}/students`);
    if (!resp.ok) throw new Error('Falha ao obter lista de alunos');
    state.students = await resp.json();
    renderStudentsTable(state.students);
  } catch (err) {
    elements.studentsTableBody.innerHTML = `
      <tr><td colspan="6" class="table-placeholder">Erro ao carregar alunos: ${err.message}</td></tr>
    `;
  }
}

function renderStudentsTable(studentsList) {
  elements.studentsCounter.textContent = `${studentsList.length} aluno${studentsList.length === 1 ? '' : 's'} listado${studentsList.length === 1 ? '' : 's'}`;

  if (studentsList.length === 0) {
    elements.studentsTableBody.innerHTML = `
      <tr><td colspan="6" class="table-placeholder">Nenhum aluno encontrado para a busca.</td></tr>
    `;
    return;
  }

  const selectedCourseId = elements.studentsCourseSelect.value;
  elements.studentsTableBody.innerHTML = '';

  studentsList.forEach(student => {
    const tr = document.createElement('tr');
    
    // Iniciais do estudante para caso não haja foto
    const names = (student.name || student.sortable_name || 'A').trim().split(/\s+/);
    const initials = (names[0][0] + (names.length > 1 ? names[names.length - 1][0] : '')).toUpperCase();

    const avatarHtml = student.avatar_url 
      ? `<img src="${student.avatar_url}" alt="${student.name}" class="student-avatar-img">`
      : `<div class="student-avatar-initials">${initials}</div>`;

    tr.innerHTML = `
      <td style="width: 60px;">${avatarHtml}</td>
      <td>
        <div class="student-row-user">${student.name || student.sortable_name}</div>
      </td>
      <td>
        <button type="button" class="student-id-copy" data-id="${student.id}" title="Clique para copiar ID">
          ${student.id} 📋
        </button>
      </td>
      <td>
        <span style="font-size: 0.8125rem; color: var(--text-muted);">Estudante Regular</span>
      </td>
      <td>
        <span class="badge-enrollment-active">✓ Ativo</span>
      </td>
      <td style="text-align: right;">
        <button type="button" class="btn-row-action btn-ask-ai-student" data-id="${student.id}" data-name="${student.name || student.sortable_name}" title="Analisar submissões do aluno com IA">
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="m12 3-1.912 5.813a2 2 0 0 1-1.275 1.275L3 12l5.813 1.912a2 2 0 0 1 1.275 1.275L12 21l1.912-5.813a2 2 0 0 1 1.275-1.275L21 12l-5.813-1.912a2 2 0 0 1-1.275-1.275L12 3Z"/></svg>
          Avaliar no Chat
        </button>
      </td>
    `;

    // Evento de cópia de ID
    const copyBtn = tr.querySelector('.student-id-copy');
    if (copyBtn) {
      copyBtn.addEventListener('click', (e) => {
        e.stopPropagation();
        navigator.clipboard.writeText(student.id).then(() => {
          showToast(`ID ${student.id} copiado para a área de transferência!`);
        });
      });
    }

    // Evento de avaliar com IA
    const aiBtn = tr.querySelector('.btn-ask-ai-student');
    if (aiBtn) {
      aiBtn.addEventListener('click', (e) => {
        e.stopPropagation();
        const sName = aiBtn.getAttribute('data-name');
        const sId = aiBtn.getAttribute('data-id');
        sendPromptToChat(`Consulte e avalie as submissões e notas do estudante '${sName}' (ID ${sId}) na turma ID ${selectedCourseId}.`);
      });
    }

    elements.studentsTableBody.appendChild(tr);
  });
}

// Botão Copiar Lista de Nomes
const btnCopyStudentsNames = document.getElementById('btnCopyStudentsNames');
if (btnCopyStudentsNames) {
  btnCopyStudentsNames.addEventListener('click', () => {
    if (state.students.length === 0) {
      showToast('Nenhum aluno carregado para copiar.', 'info');
      return;
    }
    const namesList = state.students.map(s => s.name || s.sortable_name).join('\n');
    navigator.clipboard.writeText(namesList).then(() => {
      showToast(`Lista com ${state.students.length} nomes copiada com sucesso!`);
    });
  });
}

elements.studentsSearchInput.addEventListener('input', (e) => {
  const query = e.target.value.toLowerCase().trim();
  const filtered = state.students.filter(s => 
    (s.name && s.name.toLowerCase().includes(query)) ||
    (s.sortable_name && s.sortable_name.toLowerCase().includes(query)) ||
    String(s.id).includes(query)
  );
  renderStudentsTable(filtered);
});

elements.exportCsvBtn.addEventListener('click', () => {
  if (state.students.length === 0) {
    showToast('Nenhum aluno para exportar!', 'error');
    return;
  }

  const course = state.courses.find(c => String(c.id) === elements.studentsCourseSelect.value);
  const courseName = course ? (course.clean_name || course.name).replace(/[^a-zA-Z0-9_-]/g, '_') : 'alunos';

  let csvContent = "data:text/csv;charset=utf-8,ID_Canvas;Nome_Completo;Papel;Status\n";
  state.students.forEach(s => {
    const row = `"${s.id}";"${s.name || s.sortable_name}";"Estudante";"Ativo"`;
    csvContent += row + "\n";
  });

  const encodedUri = encodeURI(csvContent);
  const link = document.createElement('a');
  link.setAttribute('href', encodedUri);
  link.setAttribute('download', `Canvas_Alunos_${courseName}.csv`);
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);

  showToast('Arquivo CSV (formato Excel) exportado com sucesso!');
});

// Botão Limpar Conversa no topo do Chat
const btnClearChat = document.getElementById('btnClearChatHistory');
if (btnClearChat) {
  btnClearChat.addEventListener('click', () => {
    if (state.chatHistory.length === 0) {
      showToast('O chat já está limpo.', 'info');
      return;
    }
    state.chatHistory = [];
    localStorage.removeItem('afya_chat_history');
    if (elements.chatMessages) {
      elements.chatMessages.innerHTML = `
        <div class="chat-welcome-state">
          <div class="welcome-badge">Afya Canvas Assistant</div>
          <h3 class="welcome-title">Conversa reiniciada. Como posso apoiar você agora?</h3>
          <p class="welcome-subtitle">
            Estou conectado ao Canvas LMS e pronto para auditar avaliações, criar questionários e verificar suas disciplinas.
          </p>
        </div>
      `;
    }
    showToast('Histórico de conversa limpo!');
  });
}

// ==========================================
// Módulo do Assistente Conversacional IA (Chat)
// ==========================================

async function loadAgentInfo() {
  try {
    const res = await fetch('/api/agent/info');
    if (!res.ok) return;
    const info = await res.json();
    state.agentInfo = info;

    if (elements.aiProviderBadge) {
      const providerNames = {
        gemini: 'Google Gemini',
        openai: 'OpenAI',
        openrouter: 'OpenRouter',
        ollama: 'Ollama Local'
      };
      elements.aiProviderBadge.textContent = providerNames[info.provider] || (info.provider ? info.provider.toUpperCase() : 'IA');
    }
    if (elements.aiModelName) {
      elements.aiModelName.textContent = info.model || 'Modelo Padrão';
    }
    if (elements.aiCanvasStatus) {
      elements.aiCanvasStatus.textContent = info.canvas_connected ? 'Canvas LMS Conectado' : 'Canvas Desconectado';
      elements.aiCanvasStatus.style.color = info.canvas_connected ? '#10b981' : '#ef4444';
    }
  } catch (err) {
    console.error('Erro ao carregar informações do agente:', err);
  }
}

function parseMarkdown(text) {
  if (window.marked && typeof window.marked.parse === 'function') {
    try {
      return window.marked.parse(text);
    } catch (e) {
      console.warn('Erro no parser marked, usando fallback:', e);
    }
  }
  // Fallback simples
  let escaped = text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;');
  return escaped
    .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
    .replace(/\*(.*?)\*/g, '<em>$1</em>')
    .replace(/`([^`]+)`/g, '<code>$1</code>')
    .replace(/\n/g, '<br>');
}

// ==========================================
// Pós-processamento de UI/UX: Cópias, Tabelas e Badges
// ==========================================

function enhanceMessageContent(container) {
  if (!container) return;

  // 1. Processar blocos de código (<pre>) com botão Copiar
  const preElements = container.querySelectorAll('pre');
  preElements.forEach((pre) => {
    if (pre.parentElement && pre.parentElement.classList.contains('code-block-wrapper')) return;

    const wrapper = document.createElement('div');
    wrapper.className = 'code-block-wrapper';

    const codeEl = pre.querySelector('code');
    const codeText = codeEl ? codeEl.innerText : pre.innerText;

    // Detecta linguagem didática
    let lang = 'Código';
    if (codeEl && codeEl.className) {
      const classMatch = codeEl.className.match(/language-(\w+)/);
      if (classMatch) {
        lang = classMatch[1].toUpperCase();
      }
    }
    if (lang === 'C' || codeText.includes('#include') || codeText.includes('int main') || codeText.includes('printf(')) {
      lang = 'Linguagem C';
    } else if (codeText.trim().startsWith('{') || codeText.trim().startsWith('[')) {
      lang = 'JSON';
    }

    const header = document.createElement('div');
    header.className = 'code-block-header';
    header.innerHTML = `
      <span class="code-lang-tag">
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="16 18 22 12 16 6"></polyline><polyline points="8 6 2 12 8 18"></polyline></svg>
        ${lang}
      </span>
      <button class="btn-copy-code" type="button" title="Copiar código para área de transferência">
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path></svg>
        <span>Copiar</span>
      </button>
    `;

    const copyBtn = header.querySelector('.btn-copy-code');
    copyBtn.addEventListener('click', async (e) => {
      e.stopPropagation();
      try {
        await navigator.clipboard.writeText(codeText);
        copyBtn.classList.add('copied');
        copyBtn.querySelector('span').textContent = 'Copiado!';
        setTimeout(() => {
          copyBtn.classList.remove('copied');
          copyBtn.querySelector('span').textContent = 'Copiar';
        }, 2000);
      } catch (err) {
        console.error('Falha ao copiar:', err);
      }
    });

    pre.parentNode.insertBefore(wrapper, pre);
    wrapper.appendChild(header);
    wrapper.appendChild(pre);
  });

  // 2. Processar tabelas (<table>) com botão Copiar TSV / Excel
  const tables = container.querySelectorAll('table');
  tables.forEach((table) => {
    if (table.parentElement && table.parentElement.classList.contains('chat-table-wrapper')) return;

    const wrapper = document.createElement('div');
    wrapper.className = 'chat-table-wrapper';

    const headerBar = document.createElement('div');
    headerBar.className = 'chat-table-header-bar';
    headerBar.innerHTML = `
      <span class="table-title-hint">
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect><line x1="3" y1="9" x2="21" y2="9"></line><line x1="9" y1="21" x2="9" y2="9"></line></svg>
        Tabela de Dados
      </span>
      <button class="btn-copy-table" type="button" title="Copiar tabela (formato compatível com Excel e Sheets)">
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path></svg>
        <span>Copiar Tabela</span>
      </button>
    `;

    const copyBtn = headerBar.querySelector('.btn-copy-table');
    copyBtn.addEventListener('click', async (e) => {
      e.stopPropagation();
      try {
        const rows = Array.from(table.querySelectorAll('tr'));
        const tsv = rows.map(r => {
          const cells = Array.from(r.querySelectorAll('th, td'));
          return cells.map(c => c.innerText.trim().replace(/\t/g, ' ')).join('\t');
        }).join('\n');

        await navigator.clipboard.writeText(tsv);
        copyBtn.classList.add('copied');
        copyBtn.querySelector('span').textContent = 'Copiada!';
        setTimeout(() => {
          copyBtn.classList.remove('copied');
          copyBtn.querySelector('span').textContent = 'Copiar Tabela';
        }, 2000);
      } catch (err) {
        console.error('Falha ao copiar tabela:', err);
      }
    });

    table.parentNode.insertBefore(wrapper, table);
    wrapper.appendChild(headerBar);
    wrapper.appendChild(table);

    // Aplicar badges nas células
    applyBadgesToTable(table);
  });

  // 3. Destacar notas e riscos em blocos de texto
  applyBadgesToTextNodes(container);
}

function applyBadgesToTable(table) {
  if (!table) return;
  const rows = Array.from(table.querySelectorAll('tr'));
  if (rows.length === 0) return;

  const headerCells = Array.from(rows[0].querySelectorAll('th, td'));
  let gradeColIdx = -1;
  let riskColIdx = -1;
  let plagColIdx = -1;

  headerCells.forEach((th, idx) => {
    const txt = th.innerText.toLowerCase();
    if (txt.includes('nota') || txt.includes('pontos') || txt.includes('score')) gradeColIdx = idx;
    if (txt.includes('risco') || txt.includes('alerta') || txt.includes('status')) riskColIdx = idx;
    if (txt.includes('plágio') || txt.includes('plagio') || txt.includes('similaridade') || txt.includes('cópia')) plagColIdx = idx;
  });

  for (let i = 1; i < rows.length; i++) {
    const cells = Array.from(rows[i].querySelectorAll('td'));

    cells.forEach((cell, cellIdx) => {
      const originalText = cell.innerText.trim();
      const lower = originalText.toLowerCase();

      // Notas: Formato "X / Y"
      const gradeMatch = originalText.match(/^([0-9]{1,3}(?:\.[0-9]+)?)\s*(?:\/|\s*de\s*)\s*([0-9]{1,3}(?:\.[0-9]+)?)$/i);
      if (gradeMatch) {
        const val = parseFloat(gradeMatch[1]);
        const max = parseFloat(gradeMatch[2]);
        const pct = max > 0 ? (val / max) : 0;
        const isPass = pct >= 0.70;
        cell.innerHTML = `<span class="${isPass ? 'badge-grade-pass' : 'badge-grade-fail'}">${isPass ? '✓' : '⚠️'} ${val} / ${max}</span>`;
        return;
      }

      // Coluna identificada de nota apenas numérica
      if (cellIdx === gradeColIdx) {
        const numVal = parseFloat(originalText.replace(',', '.'));
        if (!isNaN(numVal)) {
          let isPass = false;
          if (numVal > 20) isPass = numVal >= 70;
          else if (numVal > 10) isPass = numVal >= 14;
          else isPass = numVal >= 7.0;
          cell.innerHTML = `<span class="${isPass ? 'badge-grade-pass' : 'badge-grade-fail'}">${isPass ? '✓' : '⚠️'} ${numVal}</span>`;
          return;
        }
      }

      // Risco de Evasão
      if (lower.includes('crítico') || lower.includes('critico')) {
        cell.innerHTML = `<span class="badge-risk-critical">🚨 ${originalText}</span>`;
        return;
      } else if (lower.includes('moderado')) {
        cell.innerHTML = `<span class="badge-risk-moderate">⚠️ ${originalText}</span>`;
        return;
      } else if (lower.includes('regular') || lower.includes('baixo risco') || lower.includes('sem risco')) {
        cell.innerHTML = `<span class="badge-risk-low">🟢 ${originalText}</span>`;
        return;
      }

      // Plágio e Similaridade
      if (lower === 'alto' || lower === 'alto risco' || lower.includes('similaridade alta')) {
        cell.innerHTML = `<span class="badge-plagiarism-high">⚠️ Alto</span>`;
        return;
      } else if (lower === 'médio' || lower === 'medio' || lower === 'médio risco' || lower.includes('similaridade média')) {
        cell.innerHTML = `<span class="badge-plagiarism-medium">⚡ Médio</span>`;
        return;
      } else if (lower === 'baixo' || lower === 'baixo risco' || lower.includes('similaridade baixa')) {
        cell.innerHTML = `<span class="badge-plagiarism-low">✓ Baixo</span>`;
        return;
      }

      // Percentuais de similaridade (ex: "85%", "25%")
      const pctMatch = originalText.match(/^([0-9]{1,3}(?:\.[0-9]+)?)\s*%$/);
      if (pctMatch && (cellIdx === plagColIdx || lower.includes('%'))) {
        const pctVal = parseFloat(pctMatch[1]);
        if (pctVal >= 60) {
          cell.innerHTML = `<span class="badge-plagiarism-high">⚠️ ${pctVal}%</span>`;
        } else if (pctVal >= 30) {
          cell.innerHTML = `<span class="badge-plagiarism-medium">⚡ ${pctVal}%</span>`;
        } else {
          cell.innerHTML = `<span class="badge-plagiarism-low">✓ ${pctVal}%</span>`;
        }
        return;
      }
    });
  }
}

function applyBadgesToTextNodes(container) {
  if (!container) return;
  const elementsToScan = container.querySelectorAll('p, li');
  elementsToScan.forEach(el => {
    if (el.closest('pre') || el.closest('code') || el.closest('table')) return;

    let html = el.innerHTML;
    let modified = false;

    // Destaca notas "X / 100", "X / 50", "X / 20"
    html = html.replace(/\b([0-9]{1,3}(?:\.[0-9]+)?)\s*\/\s*(100|50|30|20|10)\b/g, (match, valStr, maxStr) => {
      const val = parseFloat(valStr);
      const max = parseFloat(maxStr);
      const pct = max > 0 ? (val / max) : 0;
      const isPass = pct >= 0.70;
      modified = true;
      return `<span class="${isPass ? 'badge-grade-pass' : 'badge-grade-fail'}">${isPass ? '✓' : '⚠️'} ${val} / ${max}</span>`;
    });

    if (modified) {
      el.innerHTML = html;
    }
  });
}

function renderChatMessage(role, content, meta = null) {
  // Remove welcome state se presente
  const welcome = elements.chatMessages.querySelector('.chat-welcome-state');
  if (welcome) {
    welcome.remove();
  }

  const msgDiv = document.createElement('div');
  msgDiv.className = `chat-msg ${role}`;

  const avatar = document.createElement('div');
  avatar.className = 'msg-avatar';
  avatar.textContent = role === 'user' ? '👤' : '🤖';

  const bubble = document.createElement('div');
  bubble.className = 'msg-bubble';

  if (role === 'assistant') {
    bubble.innerHTML = parseMarkdown(content);
    enhanceMessageContent(bubble);

    // Se houver metadados de execução (tempo / modelo), exibe de forma sutil no rodapé da bolha
    if (meta && (meta.duration_ms || meta.model)) {
      const metaRow = document.createElement('div');
      metaRow.className = 'msg-bubble-meta';
      const timeStr = meta.duration_ms ? (meta.duration_ms / 1000).toFixed(1) + 's' : '';
      const modelStr = meta.model || '';
      metaRow.innerHTML = `
        ${timeStr ? `<span class="meta-item">⏱️ ${timeStr}</span>` : ''}
        ${timeStr && modelStr ? '<span class="meta-dot">•</span>' : ''}
        ${modelStr ? `<span class="meta-item">${modelStr}</span>` : ''}
      `;
      bubble.appendChild(metaRow);
    }
  } else {
    bubble.textContent = content;
  }

  msgDiv.appendChild(avatar);
  msgDiv.appendChild(bubble);
  elements.chatMessages.appendChild(msgDiv);

  // Scroll suave para o fim
  elements.chatMessages.scrollTop = elements.chatMessages.scrollHeight;
}

function renderActionCard(card) {
  if (!card) {
    elements.chatActionCard.style.display = 'none';
    elements.chatActionCard.innerHTML = '';
    return;
  }

  elements.chatActionCard.style.display = 'block';
  elements.chatActionCard.innerHTML = `
    <div class="grade-approval-card">
      <div class="approval-header">
        <div class="approval-title">
          <span>⚠️</span>
          <span>${card.title || 'Ponto de Parada Obrigatório: Aprovação de Notas'}</span>
        </div>
      </div>
      <p style="font-size: 0.875rem; color: var(--text-secondary); margin-bottom: 0.75rem;">
        ${card.description || 'Confira as notas e justificativas antes de gravar no Canvas LMS.'}
      </p>
      <div class="approval-actions">
        <button class="btn-approve-grades" id="btnConfirmGradeAction">
          🚀 Aprovar e Publicar no Canvas LMS
        </button>
        <button class="btn-reject-grades" id="btnDismissAction">
          ✏️ Descartar / Ajustar com o Agente
        </button>
      </div>
    </div>
  `;

  document.getElementById('btnConfirmGradeAction')?.addEventListener('click', () => {
    executeConfirmedAction(card);
  });

  document.getElementById('btnDismissAction')?.addEventListener('click', () => {
    elements.chatActionCard.style.display = 'none';
    elements.chatActionCard.innerHTML = '';
    showToast('Ação cancelada pelo professor.', 'error');
  });
}

async function executeConfirmedAction(card) {
  try {
    elements.thinkingStatusText.textContent = 'Gravando notas no Canvas LMS...';
    elements.chatThinking.style.display = 'flex';

    const res = await fetch('/api/chat/confirm', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(card.payload || {})
    });

    elements.chatThinking.style.display = 'none';
    elements.chatActionCard.style.display = 'none';

    if (!res.ok) {
      const err = await res.text();
      showToast(`Erro ao lançar notas: ${err}`, 'error');
      renderChatMessage('assistant', `❌ **Falha ao publicar notas no Canvas:**\n${err}`);
      return;
    }

    const result = await res.json();
    showToast('Notas e feedbacks publicados com sucesso no Canvas!');
    renderChatMessage('assistant', `✅ **Sucesso!** As avaliações e comentários foram gravados com êxito diretamente no Canvas LMS para os estudantes.`);
  } catch (err) {
    elements.chatThinking.style.display = 'none';
    showToast(`Erro de conexão: ${err.message}`, 'error');
  }
}

async function sendChatMessage(userText) {
  if (!userText || state.isGenerating) return;

  const text = userText.trim();
  if (!text) return;

  state.isGenerating = true;
  elements.btnSendChat.disabled = true;

  // Renderiza mensagem do usuário
  renderChatMessage('user', text);
  state.chatHistory.push({ role: 'user', content: text });
  localStorage.setItem('afya_chat_history', JSON.stringify(state.chatHistory));

  // Limpa o input e reseta altura
  elements.chatInput.value = '';
  elements.chatInput.style.height = 'auto';

  // Exibe a barra de thinking inferior
  elements.thinkingStatusText.textContent = 'Iniciando processamento no Canvas LMS...';
  elements.chatThinking.style.display = 'flex';
  elements.chatMessages.scrollTop = elements.chatMessages.scrollHeight;

  // Remove welcome state se presente
  const welcome = elements.chatMessages.querySelector('.chat-welcome-state');
  if (welcome) welcome.remove();

  // Cria bolha do assistente no DOM imediatamente para receber as etapas em streaming
  const msgDiv = document.createElement('div');
  msgDiv.className = 'chat-msg assistant';

  const avatar = document.createElement('div');
  avatar.className = 'msg-avatar';
  avatar.textContent = '🤖';

  const bubble = document.createElement('div');
  bubble.className = 'msg-bubble';

  // Container de progresso dos passos em tempo real
  const progressContainer = document.createElement('div');
  progressContainer.className = 'chat-stream-progress';
  progressContainer.innerHTML = `
    <div class="chat-stream-steps">
      <div class="stream-step in-progress">
        <span class="stream-step-spinner"></span>
        <span class="step-text">Analisando solicitação e consultando o Canvas LMS...</span>
      </div>
    </div>
  `;
  bubble.appendChild(progressContainer);

  const replyBody = document.createElement('div');
  replyBody.className = 'stream-reply-body';
  bubble.appendChild(replyBody);

  msgDiv.appendChild(avatar);
  msgDiv.appendChild(bubble);
  elements.chatMessages.appendChild(msgDiv);
  elements.chatMessages.scrollTop = elements.chatMessages.scrollHeight;

  const stepsList = progressContainer.querySelector('.chat-stream-steps');
  const recordedSteps = [];
  let isCompleted = false;

  const addOrUpdateStep = (stepText) => {
    if (!stepText) return;
    recordedSteps.push(stepText);

    // Marca todos os passos anteriores como concluídos
    const currentSteps = stepsList.querySelectorAll('.stream-step');
    currentSteps.forEach(s => {
      s.className = 'stream-step done';
      const sp = s.querySelector('.stream-step-spinner');
      if (sp) {
        const check = document.createElement('span');
        check.className = 'stream-step-check';
        check.textContent = '✓';
        sp.replaceWith(check);
      }
    });

    // Adiciona o novo passo em andamento
    const newStep = document.createElement('div');
    newStep.className = 'stream-step in-progress';
    newStep.innerHTML = `
      <span class="stream-step-spinner"></span>
      <span class="step-text">${stepText}</span>
    `;
    stepsList.appendChild(newStep);

    // Atualiza também a barra inferior
    elements.thinkingStatusText.textContent = stepText;
    elements.chatMessages.scrollTop = elements.chatMessages.scrollHeight;
  };

  try {
    const res = await fetch('/api/chat/stream', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        message: text,
        history: state.chatHistory.slice(0, -1)
      })
    });

    if (!res.ok) {
      const errText = await res.text();
      throw new Error(errText || `HTTP ${res.status}`);
    }

    const reader = res.body.getReader();
    const decoder = new TextDecoder('utf-8');
    let buffer = '';

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;

      buffer += decoder.decode(value, { stream: true });
      const parts = buffer.split('\n\n');
      buffer = parts.pop(); // Mantém o pedaço incompleto

      for (const part of parts) {
        if (!part.trim()) continue;

        let eventType = 'message';
        let dataStr = '';
        const lines = part.split('\n');
        for (const line of lines) {
          if (line.startsWith('event: ')) {
            eventType = line.slice(7).trim();
          } else if (line.startsWith('data: ')) {
            dataStr = line.slice(6).trim();
          }
        }

        if (dataStr) {
          try {
            const data = JSON.parse(dataStr);

            if (eventType === 'status' || eventType === 'tool') {
              if (data.text) {
                addOrUpdateStep(data.text);
              }
            } else if (eventType === 'card') {
              if (data.action_card) {
                renderActionCard(data.action_card);
              }
            } else if (eventType === 'done') {
              isCompleted = true;
              elements.chatThinking.style.display = 'none';

              // Transforma os passos em um bloco colapsável elegante
              if (recordedSteps.length > 0) {
                const details = document.createElement('details');
                details.className = 'stream-steps-details';
                details.innerHTML = `
                  <summary>⚡ Ações executadas (${recordedSteps.length})</summary>
                  <div class="chat-stream-steps">
                    ${recordedSteps.map(s => `
                      <div class="stream-step done">
                        <span class="stream-step-check">✓</span>
                        <span class="step-text">${s}</span>
                      </div>
                    `).join('')}
                  </div>
                `;
                progressContainer.replaceWith(details);
              } else {
                progressContainer.remove();
              }

              // Renderiza corpo da resposta final
              const replyText = data.text && data.text.trim() ? data.text : 'Ação concluída com sucesso no Canvas LMS.';
              replyBody.innerHTML = parseMarkdown(replyText);
              enhanceMessageContent(bubble);

              // Metadados de tempo e modelo
              if (data.duration_ms || data.model) {
                const metaRow = document.createElement('div');
                metaRow.className = 'msg-bubble-meta';
                const timeStr = data.duration_ms ? (data.duration_ms / 1000).toFixed(1) + 's' : '';
                const modelStr = data.model || '';
                metaRow.innerHTML = `
                  ${timeStr ? `<span class="meta-item">⏱️ ${timeStr}</span>` : ''}
                  ${timeStr && modelStr ? '<span class="meta-dot">•</span>' : ''}
                  ${modelStr ? `<span class="meta-item">${modelStr}</span>` : ''}
                `;
                bubble.appendChild(metaRow);

                if (elements.chatResponseTime) elements.chatResponseTime.textContent = timeStr;
                if (elements.chatActiveModel) elements.chatActiveModel.textContent = modelStr;
                if (elements.chatMetaInfo) elements.chatMetaInfo.style.display = 'inline-flex';
              }

              state.chatHistory.push({
                role: 'assistant',
                content: replyText,
                duration_ms: data.duration_ms,
                model: data.model
              });
              localStorage.setItem('afya_chat_history', JSON.stringify(state.chatHistory));
              elements.chatMessages.scrollTop = elements.chatMessages.scrollHeight;

            } else if (eventType === 'error') {
              isCompleted = true;
              elements.chatThinking.style.display = 'none';
              progressContainer.remove();
              replyBody.innerHTML = `⚠️ **Erro na operação:** ${data.text}\n\n<button type="button" class="btn-chat-retry" onclick="retryLastChatMessage()">🔄 Tentar novamente</button>`;
              enhanceMessageContent(bubble);
            }
          } catch (jsonErr) {
            console.error('Erro no parse do evento SSE:', jsonErr, dataStr);
          }
        }
      }
    }

    if (!isCompleted) {
      // Se fechou a conexão sem enviar 'done' explícito
      elements.chatThinking.style.display = 'none';
      if (progressContainer.parentNode) {
        progressContainer.remove();
      }
      if (!replyBody.innerHTML.trim()) {
        replyBody.innerHTML = `Concluí as ações solicitadas no Canvas LMS.`;
        enhanceMessageContent(bubble);
      }
    }

  } catch (err) {
    elements.chatThinking.style.display = 'none';
    if (progressContainer.parentNode) progressContainer.remove();
    replyBody.innerHTML = `⚠️ **Erro de conexão:** ${err.message}\n\n<button type="button" class="btn-chat-retry" onclick="retryLastChatMessage()">🔄 Tentar novamente</button>`;
    enhanceMessageContent(bubble);
  } finally {
    state.isGenerating = false;
    elements.btnSendChat.disabled = false;
    elements.chatInput.focus();
  }
}

// Reenvia a última mensagem do usuário no chat em caso de oscilação
window.retryLastChatMessage = function() {
  if (state.chatHistory && state.chatHistory.length > 0) {
    for (let i = state.chatHistory.length - 1; i >= 0; i--) {
      if (state.chatHistory[i].role === 'user') {
        const lastMsg = state.chatHistory[i].content;
        sendMessage(lastMsg);
        return;
      }
    }
  }
};

// Inicializa eventos do chat
function initChatEvents() {
  loadAgentInfo();

  // Envio via formulário
  elements.chatForm?.addEventListener('submit', (e) => {
    e.preventDefault();
    sendChatMessage(elements.chatInput.value);
  });

  // Auto-resize do textarea e envio com Enter
  elements.chatInput?.addEventListener('keydown', (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      elements.chatForm.dispatchEvent(new Event('submit'));
    }
  });

  elements.chatInput?.addEventListener('input', () => {
    elements.chatInput.style.height = 'auto';
    elements.chatInput.style.height = Math.min(elements.chatInput.scrollHeight, 140) + 'px';
  });

  // Botão de Nova Conversa
  elements.btnNewChat?.addEventListener('click', () => {
    if (state.chatHistory.length === 0) return;
    if (confirm('Deseja iniciar uma nova conversa e limpar o histórico atual?')) {
      state.chatHistory = [];
      localStorage.removeItem('afya_chat_history');
      elements.chatMessages.innerHTML = `
        <div class="chat-welcome-state">
          <div class="welcome-badge">Afya Canvas Assistant</div>
          <h3 class="welcome-title">Olá, Professor Karan! Como posso apoiar você hoje?</h3>
          <p class="welcome-subtitle">
            Nova conversa iniciada. Envie qualquer dúvida, solicitação de questionário ou comando de gerenciamento acadêmico.
          </p>
        </div>
      `;
      elements.chatActionCard.style.display = 'none';
      showToast('Nova sessão iniciada!');
    }
  });

  // Alternar visibilidade da Sidebar do Chat para foco total
  const btnToggleSidebar = document.getElementById('btnToggleChatSidebar');
  const chatLayout = document.querySelector('.chat-layout');
  const toggleText = document.getElementById('toggleSidebarText');

  // Recupera preferência do usuário do localStorage
  const savedSidebarState = localStorage.getItem('afya_chat_sidebar_collapsed');
  if (savedSidebarState === 'true' && chatLayout) {
    chatLayout.classList.add('sidebar-collapsed');
    if (toggleText) toggleText.textContent = 'Mostrar Lateral';
  }

  if (btnToggleSidebar && chatLayout) {
    btnToggleSidebar.addEventListener('click', () => {
      const isCollapsed = chatLayout.classList.toggle('sidebar-collapsed');
      if (toggleText) {
        toggleText.textContent = isCollapsed ? 'Mostrar Lateral' : 'Recolher Lateral';
      }
      localStorage.setItem('afya_chat_sidebar_collapsed', isCollapsed ? 'true' : 'false');
    });
  }

  // ==========================================
  // Redimensionamento Dinâmico da Barra Lateral (Drag-to-Resize)
  // ==========================================
  const chatSidebar = document.getElementById('chatSidebar');
  const chatResizer = document.getElementById('chatSidebarResizer');

  // Restaura largura salva da barra lateral
  const savedSidebarWidth = localStorage.getItem('afya_chat_sidebar_width');
  if (savedSidebarWidth && chatSidebar) {
    const parsedWidth = parseInt(savedSidebarWidth, 10);
    if (!isNaN(parsedWidth) && parsedWidth >= 240 && parsedWidth <= 750) {
      chatSidebar.style.width = `${parsedWidth}px`;
    }
  }

  if (chatResizer && chatSidebar && chatLayout) {
    let isResizing = false;
    let startX = 0;
    let startWidth = 0;

    const startResize = (clientX) => {
      isResizing = true;
      startX = clientX;
      startWidth = chatSidebar.getBoundingClientRect().width;
      chatResizer.classList.add('is-dragging');
      chatLayout.classList.add('is-resizing');
      document.body.style.cursor = 'col-resize';
    };

    const doResize = (clientX) => {
      if (!isResizing) return;
      const deltaX = clientX - startX;
      const newWidth = Math.min(Math.max(startWidth + deltaX, 240), 750);
      chatSidebar.style.width = `${newWidth}px`;
    };

    const stopResize = () => {
      if (!isResizing) return;
      isResizing = false;
      chatResizer.classList.remove('is-dragging');
      chatLayout.classList.remove('is-resizing');
      document.body.style.cursor = '';
      const finalWidth = parseInt(chatSidebar.style.width, 10);
      if (!isNaN(finalWidth)) {
        localStorage.setItem('afya_chat_sidebar_width', finalWidth);
      }
    };

    // Eventos de Mouse
    chatResizer.addEventListener('mousedown', (e) => {
      e.preventDefault();
      startResize(e.clientX);

      const onMouseMove = (moveEvent) => doResize(moveEvent.clientX);
      const onMouseUp = () => {
        stopResize();
        window.removeEventListener('mousemove', onMouseMove);
        window.removeEventListener('mouseup', onMouseUp);
      };

      window.addEventListener('mousemove', onMouseMove);
      window.addEventListener('mouseup', onMouseUp);
    });

    // Suporte para Touch (Tablets / Touchscreens)
    chatResizer.addEventListener('touchstart', (e) => {
      if (e.touches && e.touches.length > 0) {
        startResize(e.touches[0].clientX);

        const onTouchMove = (moveEvent) => {
          if (moveEvent.touches && moveEvent.touches.length > 0) {
            doResize(moveEvent.touches[0].clientX);
          }
        };
        const onTouchEnd = () => {
          stopResize();
          window.removeEventListener('touchmove', onTouchMove);
          window.removeEventListener('touchend', onTouchEnd);
        };

        window.addEventListener('touchmove', onTouchMove, { passive: true });
        window.addEventListener('touchend', onTouchEnd);
      }
    });
  }

  // Acordeão de Categorias da Barra Lateral do Chat
  document.querySelectorAll('.qp-category-header').forEach(header => {
    header.addEventListener('click', (e) => {
      e.preventDefault();
      const targetId = header.getAttribute('data-toggle');
      const group = document.getElementById(targetId) || header.closest('.qp-category-group');
      if (group) {
        group.classList.toggle('open');
      }
    });
  });

  // Botões de Quick Prompts
  document.querySelectorAll('.quick-prompt-btn').forEach(btn => {
    btn.addEventListener('click', () => {
      const promptText = btn.getAttribute('data-prompt');
      if (promptText) {
        sendChatMessage(promptText);
      }
    });
  });

  // Restaura histórico salvo se houver
  if (state.chatHistory && state.chatHistory.length > 0) {
    const welcome = elements.chatMessages.querySelector('.chat-welcome-state');
    if (welcome) welcome.remove();

    state.chatHistory.forEach(msg => {
      renderChatMessage(msg.role, msg.content, {
        duration_ms: msg.duration_ms,
        model: msg.model
      });
    });
  }
}

// Inicialização automática
initApp();
initChatEvents();

// ==========================================================================
// Módulo: Visualizador de Códigos dos Alunos & Testador Integrado
// ==========================================================================

const codeViewerState = {
  courseId: null,
  assignmentId: null,
  assignmentName: '',
  pointsPossible: null,
  submissions: [],
  currentIndex: 0,
  currentCode: '',
  errorLines: new Set()
};

// Elementos do Modal de Visualização de Código
const cvElements = {
  modal: document.getElementById('codeViewerModal'),
  assignmentTitle: document.getElementById('cvAssignmentTitle'),
  courseSubtitle: document.getElementById('cvCourseSubtitle'),
  studentSelect: document.getElementById('cvStudentSelect'),
  btnPrev: document.getElementById('cvBtnPrevStudent'),
  btnNext: document.getElementById('cvBtnNextStudent'),
  btnClose: document.getElementById('cvBtnClose'),
  studentAvatar: document.getElementById('cvStudentAvatar'),
  studentName: document.getElementById('cvStudentName'),
  statusTag: document.getElementById('cvStatusTag'),
  dateTag: document.getElementById('cvDateTag'),
  scoreTag: document.getElementById('cvScoreTag'),
  langBadge: document.getElementById('cvLangBadge'),
  btnCopyCode: document.getElementById('cvBtnCopyCode'),
  btnTestCode: document.getElementById('cvBtnTestCode'),
  btnAskAICode: document.getElementById('cvBtnAskAICode'),
  lineNumbers: document.getElementById('cvLineNumbers'),
  codeContent: document.getElementById('cvCodeContent'),
  codeContainer: document.getElementById('cvCodeContainer'),
  testResultsPanel: document.getElementById('cvTestResultsPanel'),
  trpTitle: document.getElementById('cvTrpTitle'),
  trpStatusText: document.getElementById('cvTrpStatusText'),
  trpBody: document.getElementById('cvTrpBody'),
  btnCloseResults: document.getElementById('cvBtnCloseResults'),
  attachmentsArea: document.getElementById('cvAttachmentsArea'),
  studentCounter: document.getElementById('cvStudentCounter')
};

// Abre o Modal e carrega as submissões
async function openCodeViewerModal(courseId, assignmentId, assignmentName, pointsPossible) {
  codeViewerState.courseId = courseId;
  codeViewerState.assignmentId = assignmentId;
  codeViewerState.assignmentName = assignmentName;
  codeViewerState.pointsPossible = pointsPossible;
  codeViewerState.submissions = [];
  codeViewerState.currentIndex = 0;
  codeViewerState.currentCode = '';
  codeViewerState.errorLines.clear();

  cvElements.assignmentTitle.textContent = assignmentName || 'Visualizador de Código';
  const course = state.courses.find(c => String(c.id) === String(courseId));
  cvElements.courseSubtitle.textContent = course ? (course.clean_name || course.name) : `Disciplina ID ${courseId}`;

  // Exibe o modal com estado inicial de carregamento
  cvElements.modal.style.display = 'flex';
  cvElements.studentSelect.innerHTML = '<option value="">Buscando entregas no Canvas...</option>';
  cvElements.lineNumbers.innerHTML = '<span>1</span>';
  cvElements.codeContent.innerHTML = '<span class="sh-comment">// Consultando submissões e códigos dos estudantes no Canvas LMS...</span>';
  cvElements.testResultsPanel.style.display = 'none';
  cvElements.attachmentsArea.innerHTML = '';
  cvElements.studentCounter.textContent = 'Carregando...';

  try {
    const res = await fetch(`/api/courses/${courseId}/assignments/${assignmentId}/submissions`);
    if (!res.ok) throw new Error(`Erro HTTP ${res.status}`);
    const list = await res.json();

    if (!Array.isArray(list) || list.length === 0) {
      cvElements.studentSelect.innerHTML = '<option value="">Nenhuma entrega registrada</option>';
      cvElements.codeContent.innerHTML = '<span class="sh-comment">// Nenhum aluno submeteu esta atividade ainda.</span>';
      cvElements.studentCounter.textContent = '0 alunos';
      return;
    }

    // Ordena: quem tem código primeiro, depois alfabético
    list.sort((a, b) => {
      const aHas = a.has_code || (a.clean_body && a.clean_body.trim().length > 0) ? 1 : 0;
      const bHas = b.has_code || (b.clean_body && b.clean_body.trim().length > 0) ? 1 : 0;
      if (bHas !== aHas) return bHas - aHas;
      return (a.user_name || '').localeCompare(b.user_name || '');
    });

    codeViewerState.submissions = list;

    // Popula o select de estudantes
    cvElements.studentSelect.innerHTML = '';
    list.forEach((sub, idx) => {
      const opt = document.createElement('option');
      opt.value = idx;
      const hasCodeIcon = sub.has_code || (sub.clean_body && sub.clean_body.trim().length > 0) ? '💻 ' : '📄 ';
      const statusIcon = sub.workflow_state === 'graded' ? '✓ ' : '⚠️ ';
      const scoreTxt = sub.grade ? ` [${sub.grade} pts]` : '';
      opt.textContent = `${statusIcon}${hasCodeIcon}${sub.user_name || `Aluno ID ${sub.user_id}`}${scoreTxt}`;
      cvElements.studentSelect.appendChild(opt);
    });

    // Renderiza a primeira submissão
    renderStudentSubmission(0);

  } catch (err) {
    console.error('Erro ao buscar submissões:', err);
    cvElements.studentSelect.innerHTML = '<option value="">Erro ao carregar</option>';
    cvElements.codeContent.innerHTML = `<span class="sh-comment">// Falha ao conectar ao Canvas: ${err.message}</span>`;
    showToast('Não foi possível carregar as submissões desta atividade.', 'error');
  }
}

// Fecha o modal
function closeCodeViewerModal() {
  cvElements.modal.style.display = 'none';
  cvElements.testResultsPanel.style.display = 'none';
  codeViewerState.errorLines.clear();
}

// Renderiza o aluno selecionado no visualizador
function renderStudentSubmission(index) {
  if (!codeViewerState.submissions || codeViewerState.submissions.length === 0) return;
  if (index < 0) index = 0;
  if (index >= codeViewerState.submissions.length) index = codeViewerState.submissions.length - 1;

  codeViewerState.currentIndex = index;
  codeViewerState.errorLines.clear();
  cvElements.studentSelect.value = index;

  const sub = codeViewerState.submissions[index];
  const uName = sub.user_name || `Estudante ID ${sub.user_id}`;

  // Avatar e Nome
  const initials = uName.split(' ').map(n => n[0]).slice(0, 2).join('').toUpperCase();
  cvElements.studentAvatar.textContent = initials || 'AL';
  cvElements.studentName.textContent = uName;

  // Status de Avaliação
  const isGraded = sub.workflow_state === 'graded';
  cvElements.statusTag.className = `cv-status-tag ${isGraded ? '' : 'pending'}`;
  cvElements.statusTag.textContent = isGraded ? '✓ Corrigida' : '⚠️ Aguardando Nota';

  // Data de Entrega (fuso de Brasília)
  cvElements.dateTag.textContent = sub.submitted_at ? `📅 ${formatDate(sub.submitted_at)}` : '📅 Sem data';

  // Nota
  const maxPts = codeViewerState.pointsPossible !== null && codeViewerState.pointsPossible !== undefined ? ` / ${codeViewerState.pointsPossible}` : '';
  cvElements.scoreTag.textContent = sub.grade ? `🏆 Nota: ${sub.grade}${maxPts}` : `🏆 Sem nota${maxPts}`;

  // Contador
  cvElements.studentCounter.textContent = `Aluno ${index + 1} de ${codeViewerState.submissions.length}`;

  // Reseta painel de diagnóstico
  cvElements.testResultsPanel.style.display = 'none';

  // Extração do Código
  let rawCode = '';
  if (sub.clean_body && sub.clean_body.trim().length > 0) {
    rawCode = sub.clean_body.trim();
  } else if (sub.body && sub.body.trim().length > 0) {
    rawCode = sub.body.trim();
  } else if (sub.url && sub.url.trim().length > 0) {
    rawCode = `// O estudante submeteu um link externo:\n// URL: ${sub.url}`;
  } else if (sub.attachments && sub.attachments.length > 0) {
    const attNames = sub.attachments.map(a => a.display_name || a.filename).join(', ');
    rawCode = `// Arquivo(s) anexado(s) pelo estudante: ${attNames}\n// Baixe o anexo nos botões de download no rodapé deste painel para inspecionar.`;
  } else {
    rawCode = `// Nenhuma entrega textual ou código foi encontrado para este estudante nesta atividade.`;
  }

  codeViewerState.currentCode = rawCode;

  // Renderiza Anexos se houver
  cvElements.attachmentsArea.innerHTML = '';
  if (sub.attachments && sub.attachments.length > 0) {
    sub.attachments.forEach(att => {
      const a = document.createElement('a');
      a.className = 'cv-attach-chip';
      a.href = att.url || '#';
      a.target = '_blank';
      a.rel = 'noopener noreferrer';
      a.innerHTML = `📎 ${att.display_name || att.filename || 'Anexo'}`;
      cvElements.attachmentsArea.appendChild(a);
    });
  }

  // Renderiza com Syntax Highlighting e Linhas
  renderCodeViewerLines(rawCode, codeViewerState.errorLines);
}

// Tokenizador e Syntax Highlighting para C / C++
function highlightCSyntax(line) {
  let escaped = line
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;');

  // Comentários de linha única (// ...)
  if (escaped.includes('//')) {
    const parts = escaped.split('//');
    const codePart = highlightCSyntaxTokens(parts[0]);
    return `${codePart}<span class="sh-comment">//${parts.slice(1).join('//')}</span>`;
  }

  return highlightCSyntaxTokens(escaped);
}

function highlightCSyntaxTokens(text) {
  // Strings ("..." ou '...')
  text = text.replace(/("(\\.|[^"\\])*")/g, '<span class="sh-string">$1</span>');
  text = text.replace(/('(\\.|[^'\\])*')/g, '<span class="sh-string">$1</span>');

  // Diretivas pré-processador (#include, #define...)
  text = text.replace(/^\s*(#(?:include|define|ifndef|ifdef|endif|pragma)[^\n<]*)(&lt;[^&]*&gt;)?/g, (match, p1, p2) => {
    return `<span class="sh-preprocessor">${p1}</span>${p2 ? `<span class="sh-string">${p2}</span>` : ''}`;
  });

  // Palavras-chave de controle
  const keywords = /\b(return|if|else|while|for|switch|case|default|break|continue|struct|typedef|sizeof|goto|do|const|static|extern|inline|volatile)\b/g;
  text = text.replace(keywords, '<span class="sh-keyword">$1</span>');

  // Tipos primitivos e estruturais de C
  const types = /\b(int|void|char|float|double|bool|size_t|long|short|unsigned|signed|FILE|uint8_t|uint16_t|uint32_t|int32_t|int64_t|NULL)\b/g;
  text = text.replace(types, '<span class="sh-type">$1</span>');

  // Números literais
  text = text.replace(/\b(0x[0-9a-fA-F]+|\d+(\.\d+)?)\b/g, '<span class="sh-number">$1</span>');

  // Chamadas de funções
  text = text.replace(/\b([a-zA-Z_][a-zA-Z0-9_]*)\s*(?=\()/g, '<span class="sh-function">$1</span>');

  return text;
}

// Renderiza as linhas numeradas e o código com destaque
function renderCodeViewerLines(rawCode, errorLinesSet) {
  const lines = rawCode.split('\n');
  cvElements.lineNumbers.innerHTML = '';
  cvElements.codeContent.innerHTML = '';

  const numFragment = document.createDocumentFragment();
  const codeFragment = document.createDocumentFragment();

  lines.forEach((lineText, idx) => {
    const lineNum = idx + 1;
    const isError = errorLinesSet && errorLinesSet.has(lineNum);

    // Gutter número de linha
    const numSpan = document.createElement('span');
    numSpan.textContent = lineNum;
    if (isError) numSpan.classList.add('line-highlight-error');
    numFragment.appendChild(numSpan);

    // Linha de código
    const lineDiv = document.createElement('div');
    lineDiv.id = `cv-line-${lineNum}`;
    lineDiv.innerHTML = highlightCSyntax(lineText) || '&nbsp;';
    if (isError) lineDiv.classList.add('line-highlight-error');
    codeFragment.appendChild(lineDiv);
  });

  cvElements.lineNumbers.appendChild(numFragment);
  cvElements.codeContent.appendChild(codeFragment);
}

// Testar Compilação / Sintaxe do Aluno
async function runCodeTestOnCurrentSubmission() {
  if (!codeViewerState.currentCode || codeViewerState.currentCode.trim().length === 0) {
    showToast('Não há código submetido para testar.', 'error');
    return;
  }

  const origBtnText = cvElements.btnTestCode.innerHTML;
  cvElements.btnTestCode.disabled = true;
  cvElements.btnTestCode.innerHTML = `
    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="thinking-spinner"><path d="M21 12a9 9 0 1 1-6.219-8.56"/></svg>
    <span>Compilando...</span>
  `;

  cvElements.testResultsPanel.style.display = 'block';
  cvElements.trpTitle.className = 'cv-trp-title';
  cvElements.trpStatusText.textContent = 'Executando análise sintática e teste com compilador gcc/clang...';
  cvElements.trpBody.innerHTML = '<div class="sh-comment">// Compilando código em ambiente sandbox isolado...</div>';

  try {
    const res = await fetch('/api/code/test', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        code: codeViewerState.currentCode,
        language: 'c'
      })
    });

    if (!res.ok) throw new Error(`Erro na compilação: HTTP ${res.status}`);
    const result = await res.json();

    codeViewerState.errorLines.clear();

    const isSuccess = result.status === 'success' || result.syntax_valid === true;

    if (isSuccess) {
      cvElements.trpTitle.className = 'cv-trp-title success';
      cvElements.trpStatusText.textContent = '✓ Código Compilado com Sucesso!';

      let detailsHtml = `
        <div class="cv-diag-item success">
          <span class="cv-diag-line">Sintaxe C99:</span>
          <span>Compilação concluída sem erros fatais.</span>
        </div>
      `;

      if (result.warnings && result.warnings.length > 0) {
        detailsHtml += `<div style="margin-top: 0.5rem; font-weight: 700; color: var(--color-warning);">⚠️ Avisos (Warnings) do Compilador:</div>`;
        result.warnings.forEach(w => {
          detailsHtml += `
            <div class="cv-diag-item warning">
              <span>${w.message || w}</span>
            </div>
          `;
        });
      }

      if (result.test_results && Array.isArray(result.test_results)) {
        detailsHtml += `<div style="margin-top: 0.5rem; font-weight: 700; color: var(--color-success);">🧪 Casos de Teste Executados:</div>`;
        result.test_results.forEach(tr => {
          detailsHtml += `
            <div class="cv-diag-item ${tr.passed ? 'success' : 'error'}">
              <span>${tr.passed ? '✓ Passou' : '✕ Falhou'}: ${tr.name || tr.description || 'Caso de teste'}</span>
            </div>
          `;
        });
      }

      cvElements.trpBody.innerHTML = detailsHtml;
      renderCodeViewerLines(codeViewerState.currentCode, codeViewerState.errorLines);

    } else {
      cvElements.trpTitle.className = 'cv-trp-title error';
      cvElements.trpStatusText.textContent = '✕ Falha de Compilação / Sintaxe Detectada';

      let errorsList = [];
      if (Array.isArray(result.errors) && result.errors.length > 0) {
        errorsList = result.errors;
      } else if (result.error) {
        errorsList.push({ message: result.error });
      } else if (result.output) {
        // Tenta extrair linhas e mensagens do output do gcc
        const errorRegex = /(?:eval_[a-zA-Z0-9_]+\.c|scratch\/[^\s:]+):(\d+):(?:\d+:)?\s*(error|warning):\s*([^\n]+)/g;
        let match;
        while ((match = errorRegex.exec(result.output)) !== null) {
          errorsList.push({
            line: parseInt(match[1]),
            type: match[2],
            message: match[3]
          });
        }
        if (errorsList.length === 0) {
          errorsList.push({ message: result.output });
        }
      }

      let errHtml = `<div style="margin-bottom: 0.5rem; font-weight: 700; color: var(--color-danger);">Erros apontados pelo compilador (clique para navegar até a linha):</div>`;

      errorsList.forEach(err => {
        const lineNum = err.line || (err.message && err.message.match(/linha\s+(\d+)/i) ? parseInt(err.message.match(/linha\s+(\d+)/i)[1]) : null);
        if (lineNum) codeViewerState.errorLines.add(lineNum);

        const lineTag = lineNum ? `<span class="cv-diag-line" data-goto-line="${lineNum}">[Linha ${lineNum}]</span>` : '';
        errHtml += `
          <div class="cv-diag-item error" data-goto-line="${lineNum || ''}">
            ${lineTag}
            <span>${err.message || 'Erro de sintaxe desconhecido'}</span>
          </div>
        `;
      });

      cvElements.trpBody.innerHTML = errHtml;

      // Re-renderiza com as linhas vermelhas destacadas
      renderCodeViewerLines(codeViewerState.currentCode, codeViewerState.errorLines);

      // Adiciona listener de clique para ir até a linha
      cvElements.trpBody.querySelectorAll('[data-goto-line]').forEach(el => {
        el.addEventListener('click', () => {
          const targetLine = parseInt(el.getAttribute('data-goto-line'));
          if (targetLine) {
            const lineEl = document.getElementById(`cv-line-${targetLine}`);
            if (lineEl) {
              lineEl.scrollIntoView({ behavior: 'smooth', block: 'center' });
            }
          }
        });
      });
    }

  } catch (err) {
    cvElements.trpTitle.className = 'cv-trp-title error';
    cvElements.trpStatusText.textContent = 'Erro ao se comunicar com o testador';
    cvElements.trpBody.innerHTML = `<div class="cv-diag-item error"><span>${err.message}</span></div>`;
  } finally {
    cvElements.btnTestCode.disabled = false;
    cvElements.btnTestCode.innerHTML = origBtnText;
  }
}

// Copiar código da submissão
function copyCurrentSubmissionCode() {
  if (!codeViewerState.currentCode) return;
  navigator.clipboard.writeText(codeViewerState.currentCode).then(() => {
    const origHtml = cvElements.btnCopyCode.innerHTML;
    cvElements.btnCopyCode.innerHTML = `
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12"></polyline></svg>
      <span>Copiado!</span>
    `;
    setTimeout(() => {
      cvElements.btnCopyCode.innerHTML = origHtml;
    }, 2000);
    showToast('Código copiado para a área de transferência!');
  }).catch(err => {
    console.error('Falha ao copiar:', err);
    showToast('Não foi possível copiar o código automaticamente.', 'error');
  });
}

// Enviar código atual diretamente para avaliação com IA
function askAICodeCurrentSubmission() {
  if (!codeViewerState.currentCode) return;
  const sub = codeViewerState.submissions[codeViewerState.currentIndex];
  const uName = sub ? sub.user_name : 'Estudante';
  const uId = sub ? sub.user_id : '-';

  closeCodeViewerModal();

  // Ativa a aba de chat
  const chatTabBtn = document.getElementById('tabBtnChat');
  if (chatTabBtn) chatTabBtn.click();

  const prompt = `Por favor, faça uma avaliação técnica e pedagógica (com tom neutro e institucional, conforme o regulamento da Afya) do código em C submetido pelo estudante ${uName} (ID Canvas: ${uId}) para a atividade '${codeViewerState.assignmentName}'.\n\nIdentifique se o código compila corretamente, verifique casos de borda e aponte cirurgicamente a linha de qualquer falha se houver:\n\n\`\`\`c\n${codeViewerState.currentCode}\n\`\`\``;

  sendChatMessage(prompt);
}

// Inicializa Eventos do Visualizador de Código
function initCodeViewerEvents() {
  // Fechar Modal
  cvElements.btnClose?.addEventListener('click', closeCodeViewerModal);
  cvElements.btnCloseResults?.addEventListener('click', () => {
    cvElements.testResultsPanel.style.display = 'none';
  });

  // Fechar clicando no backdrop escuro fora da caixa
  cvElements.modal?.addEventListener('click', (e) => {
    if (e.target === cvElements.modal) {
      closeCodeViewerModal();
    }
  });

  // Mudança no select de estudantes
  cvElements.studentSelect?.addEventListener('change', (e) => {
    const idx = parseInt(e.target.value);
    if (!isNaN(idx)) renderStudentSubmission(idx);
  });

  // Navegação Anterior / Próximo
  cvElements.btnPrev?.addEventListener('click', () => {
    if (codeViewerState.currentIndex > 0) {
      renderStudentSubmission(codeViewerState.currentIndex - 1);
    }
  });

  cvElements.btnNext?.addEventListener('click', () => {
    if (codeViewerState.currentIndex < codeViewerState.submissions.length - 1) {
      renderStudentSubmission(codeViewerState.currentIndex + 1);
    }
  });

  // Botões de Ação
  cvElements.btnCopyCode?.addEventListener('click', copyCurrentSubmissionCode);
  cvElements.btnTestCode?.addEventListener('click', runCodeTestOnCurrentSubmission);
  cvElements.btnAskAICode?.addEventListener('click', askAICodeCurrentSubmission);

  // Atalhos de teclado (Esc para fechar, Setas para navegar)
  window.addEventListener('keydown', (e) => {
    if (cvElements.modal && cvElements.modal.style.display === 'flex') {
      if (e.key === 'Escape') {
        closeCodeViewerModal();
      } else if (e.key === 'ArrowLeft') {
        if (codeViewerState.currentIndex > 0) {
          renderStudentSubmission(codeViewerState.currentIndex - 1);
        }
      } else if (e.key === 'ArrowRight') {
        if (codeViewerState.currentIndex < codeViewerState.submissions.length - 1) {
          renderStudentSubmission(codeViewerState.currentIndex + 1);
        }
      }
    }
  });
}

// Registra os eventos do visualizador
initCodeViewerEvents();

// ==========================================================================
// MÓDULO DE GRÁFICOS & LEARNING ANALYTICS (CHART.JS)
// ==========================================================================

// Destrói gráfico anterior com segurança para evitar leaks de memória e sobreposição
function destroyAnalyticsChart(key) {
  if (analyticsState.charts[key]) {
    try {
      analyticsState.charts[key].destroy();
    } catch (e) {
      console.warn('Erro ao destruir gráfico:', e);
    }
    analyticsState.charts[key] = null;
  }
}

// Formata o rótulo de cada opção do seletor para ser 100% descritivo
function formatAnalyticsCourseLabel(c) {
  const clean = c.clean_name || c.name || 'Disciplina';
  const period = c.period ? ` • ${c.period}` : '';
  let codeSnippet = '';
  const match = (c.name || '').match(/(?:-\s*)?(\d{5,6})\b/);
  if (match) {
    codeSnippet = ` - Turma ${match[1]}`;
  }
  const students = c.total_students ? ` (${c.total_students} alunos)` : '';
  return `${clean}${period}${codeSnippet}${students}`;
}

// Preenche o seletor com optgroups claros (Vigente e Anteriores)
function populateAnalyticsCourseSelect(forceRebuild = false) {
  const sel = document.getElementById('analyticsCourseSelect');
  if (!sel || !state.courses || state.courses.length === 0) return;

  // Se já tiver opções e não for rebuild forçado, apenas garante seleção correta
  if (!forceRebuild && sel.options.length > 1 && !sel.querySelector('option[value=""]')) {
    if (analyticsState.selectedCourseId && sel.value !== analyticsState.selectedCourseId) {
      sel.value = analyticsState.selectedCourseId;
    }
    return;
  }

  const currentVal = analyticsState.selectedCourseId || sel.value;
  sel.innerHTML = '';

  const currentCourses = state.courses.filter(c => c.is_current_term);
  const pastCourses = state.courses.filter(c => !c.is_current_term);

  const addOptGroup = (label, list) => {
    if (list.length === 0) return;
    const group = document.createElement('optgroup');
    group.label = label;
    list.forEach(c => {
      const opt = document.createElement('option');
      opt.value = String(c.id);
      opt.textContent = formatAnalyticsCourseLabel(c);
      group.appendChild(opt);
    });
    sel.appendChild(group);
  };

  addOptGroup('🟢 Semestre Vigente (Atual)', currentCourses);
  addOptGroup('⚪ Semestres Anteriores (Histórico)', pastCourses);

  // Seleciona a turma adequada
  if (currentVal && state.courses.some(c => String(c.id) === currentVal)) {
    sel.value = currentVal;
  } else if (state.selectedCourseId && state.selectedCourseId !== 'all' && state.courses.some(c => String(c.id) === state.selectedCourseId)) {
    sel.value = state.selectedCourseId;
  } else if (currentCourses.length > 0) {
    sel.value = String(currentCourses[0].id);
  } else if (state.courses.length > 0) {
    sel.value = String(state.courses[0].id);
  }

  analyticsState.selectedCourseId = sel.value;
}

// Ponto de entrada chamado ao abrir a aba de gráficos
async function loadAnalyticsTab() {
  populateAnalyticsCourseSelect(false);
  if (analyticsState.selectedCourseId) {
    await fetchAndRenderAnalytics(analyticsState.selectedCourseId);
  }
}

// Busca dados da API (ou usa cache) e renderiza KPIs e gráficos
async function fetchAndRenderAnalytics(courseId, forceRefresh = false) {
  if (!courseId) return;

  const container = document.querySelector('.analytics-container');
  const badge = document.getElementById('analyticsLoadingBadge');
  const updatedAt = document.getElementById('analyticsUpdatedAt');

  // Resposta instantânea via cache local
  if (!forceRefresh && analyticsCache[courseId]) {
    const cached = analyticsCache[courseId];
    analyticsState.summary = cached;
    renderAnalyticsKPIs(cached);
    renderApprovalDonut(cached);
    renderSpeedGraderBars(cached);
    renderGradesHistogram(cached);
    renderRiskScatter(cached);
    if (updatedAt) {
      updatedAt.textContent = 'Carregado (Cache rápido)';
    }
    return;
  }

  // Ativa feedback visual de carregamento
  if (container) container.classList.add('is-loading');
  if (badge) badge.style.display = 'inline-flex';
  if (updatedAt) updatedAt.textContent = 'Sincronizando com Canvas...';

  try {
    const resp = await fetch(`/api/analytics/summary?course_id=${encodeURIComponent(courseId)}`);
    if (!resp.ok) throw new Error(`Falha HTTP ${resp.status} ao obter dados`);

    const summary = await resp.json();
    analyticsCache[courseId] = summary;
    analyticsState.summary = summary;

    renderAnalyticsKPIs(summary);
    renderApprovalDonut(summary);
    renderSpeedGraderBars(summary);
    renderGradesHistogram(summary);
    renderRiskScatter(summary);

    if (updatedAt) {
      const now = new Date();
      const timeStr = `${String(now.getHours()).padStart(2, '0')}:${String(now.getMinutes()).padStart(2, '0')}`;
      updatedAt.textContent = `Atualizado às ${timeStr}`;
    }

  } catch (err) {
    console.error('Erro ao renderizar analytics:', err);
    showToast('Não foi possível carregar as métricas da disciplina.', 'error');
    if (updatedAt) updatedAt.textContent = 'Falha na conexão';
  } finally {
    if (container) container.classList.remove('is-loading');
    if (badge) badge.style.display = 'none';
  }
}

// Renderiza os banners superiores de KPIs
function renderAnalyticsKPIs(summary) {
  const total = summary.total_students || 1;
  const appPct = Math.round((summary.approved_direct_count / total) * 100);
  const finPct = Math.round((summary.final_exam_count / total) * 100);
  const riskPct = Math.round((summary.at_risk_count / total) * 100);

  // Banner de estágio pedagógico
  const stageBanner = document.getElementById('analyticsStageBanner');
  const stageText = document.getElementById('analyticsStageText');
  if (stageBanner && stageText) {
    if (summary.stage_description) {
      stageText.textContent = summary.stage_description;
      stageBanner.style.display = 'flex';
    } else {
      stageBanner.style.display = 'none';
    }
  }

  const kAvg = document.getElementById('kpiAvgScore');
  const kAvgSub = document.getElementById('kpiAvgScoreSub');
  const kAppR = document.getElementById('kpiApprovalRate');
  const kAppC = document.getElementById('kpiApprovalCount');
  const kFinR = document.getElementById('kpiFinalExamRate');
  const kFinC = document.getElementById('kpiFinalExamCount');
  const kRiskR = document.getElementById('kpiAtRiskRate');
  const kRiskC = document.getElementById('kpiAtRiskCount');

  // Média adaptativa às notas parciais
  if (kAvg) {
    if (summary.graded_students_count > 0) {
      if (summary.evaluated_points_possible > 0 && summary.evaluated_points_possible < 100) {
        kAvg.textContent = `${(summary.average_class_score || 0).toFixed(1)} / ${(summary.evaluated_points_possible || 0).toFixed(0)} pts`;
      } else {
        kAvg.textContent = `${(summary.average_class_score || 0).toFixed(1)} pts`;
      }
    } else {
      kAvg.textContent = 'Em Andamento';
    }
  }

  if (kAvgSub) {
    if (summary.graded_students_count > 0) {
      kAvgSub.textContent = `${summary.graded_students_count} de ${summary.total_students} avaliados (${summary.evaluated_tasks_count} tarefa(s))`;
    } else {
      kAvgSub.textContent = 'Atividades parciais em execução';
    }
  }

  if (kAppR) kAppR.textContent = `${appPct}%`;
  if (kAppC) kAppC.textContent = `${summary.approved_direct_count || 0} de ${summary.total_students || 0} em dia`;

  if (kFinR) kFinR.textContent = `${finPct}%`;
  if (kFinC) kFinC.textContent = `${summary.final_exam_count || 0} em acompanhamento`;

  if (kRiskR) kRiskR.textContent = `${riskPct}%`;
  if (kRiskC) kRiskC.textContent = `${summary.at_risk_count || 0} com alerta efetivo`;
}

// Gráfico 1: Donut Contextual & CONSEPE
function renderApprovalDonut(summary) {
  const canvas = document.getElementById('chartApprovalDonut');
  if (!canvas || typeof Chart === 'undefined') return;
  destroyAnalyticsChart('donut');

  const ctx = canvas.getContext('2d');
  const isDark = state.theme === 'dark';

  const total = summary.total_students || 1;
  const hasData = (summary.approved_direct_count + summary.final_exam_count + summary.at_risk_count) > 0;

  const dataValues = hasData
    ? [summary.approved_direct_count || 0, summary.final_exam_count || 0, summary.at_risk_count || 0]
    : [0, 0, 1];

  analyticsState.charts.donut = new Chart(ctx, {
    type: 'doughnut',
    data: {
      labels: ['Em Dia / Ritmo Adequado', 'Acompanhamento Parcial', 'Atenção Prioritária (Risco)'],
      datasets: [{
        data: dataValues,
        backgroundColor: hasData ? ['#10b981', '#f59e0b', '#ef4444'] : ['#475569', '#475569', '#475569'],
        borderColor: isDark ? '#1e293b' : '#ffffff',
        borderWidth: 3,
        hoverOffset: 6
      }]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          position: 'bottom',
          labels: {
            color: isDark ? '#cbd5e1' : '#334155',
            font: { family: 'Plus Jakarta Sans', size: 12, weight: '600' },
            padding: 16
          }
        },
        tooltip: {
          callbacks: {
            label: function(context) {
              const val = context.raw || 0;
              const pct = Math.round((val / total) * 100);
              return ` ${context.label}: ${val} alunos (${pct}%)`;
            }
          }
        }
      },
      cutout: '65%'
    }
  });
}

// Gráfico 2: SpeedGrader Barras Empilhadas
function renderSpeedGraderBars(summary) {
  const canvas = document.getElementById('chartSpeedGraderBars');
  if (!canvas || typeof Chart === 'undefined') return;
  destroyAnalyticsChart('speedgrader');

  const tasks = summary.assignments_status || [];
  const ctx = canvas.getContext('2d');
  const isDark = state.theme === 'dark';

  let labels = [];
  let graded = [];
  let pending = [];
  let unsubmitted = [];

  if (tasks.length > 0) {
    labels = tasks.map(t => t.title.length > 28 ? t.title.substring(0, 28) + '...' : t.title);
    graded = tasks.map(t => t.graded_count);
    pending = tasks.map(t => t.pending_count);
    unsubmitted = tasks.map(t => t.unsubmitted_count);
  } else {
    labels = ['Sem atividades com pendências'];
    graded = [0];
    pending = [0];
    unsubmitted = [0];
  }

  analyticsState.charts.speedgrader = new Chart(ctx, {
    type: 'bar',
    data: {
      labels: labels,
      datasets: [
        {
          label: 'Corrigidas e Lançadas',
          data: graded,
          backgroundColor: '#10b981',
          borderRadius: 4,
          maxBarThickness: 26
        },
        {
          label: 'Aguardando Professor',
          data: pending,
          backgroundColor: '#f97316',
          borderRadius: 4,
          maxBarThickness: 26
        },
        {
          label: 'Dentro do Prazo / Abertas',
          data: unsubmitted,
          backgroundColor: isDark ? '#475569' : '#94a3b8',
          borderRadius: 4,
          maxBarThickness: 26
        }
      ]
    },
    options: {
      indexAxis: 'y',
      responsive: true,
      maintainAspectRatio: false,
      scales: {
        x: {
          stacked: true,
          grid: { color: isDark ? 'rgba(255, 255, 255, 0.05)' : 'rgba(0, 0, 0, 0.05)' },
          ticks: { color: isDark ? '#94a3b8' : '#64748b' }
        },
        y: {
          stacked: true,
          grid: { display: false },
          ticks: { color: isDark ? '#cbd5e1' : '#334155', font: { weight: '600' } }
        }
      },
      plugins: {
        legend: {
          position: 'bottom',
          labels: { color: isDark ? '#cbd5e1' : '#334155', font: { family: 'Plus Jakarta Sans', size: 11 } }
        }
      }
    }
  });
}

// Gráfico 3: Histograma de Notas (Gauss)
function renderGradesHistogram(summary) {
  const canvas = document.getElementById('chartGradesHistogram');
  if (!canvas || typeof Chart === 'undefined') return;
  destroyAnalyticsChart('histogram');

  const bands = summary.grade_distribution || [];
  const labels = bands.map(b => b.label);
  const counts = bands.map(b => b.count);

  const ctx = canvas.getContext('2d');
  const isDark = state.theme === 'dark';

  analyticsState.charts.histogram = new Chart(ctx, {
    type: 'bar',
    data: {
      labels: labels,
      datasets: [{
        label: 'Quantidade de Alunos',
        data: counts,
        backgroundColor: [
          'rgba(239, 68, 68, 0.85)',
          'rgba(245, 158, 11, 0.85)',
          'rgba(59, 130, 246, 0.85)',
          'rgba(14, 165, 233, 0.85)',
          'rgba(16, 185, 129, 0.85)'
        ],
        borderRadius: 6,
        maxBarThickness: 44
      }]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      scales: {
        y: {
          beginAtZero: true,
          grid: { color: isDark ? 'rgba(255, 255, 255, 0.05)' : 'rgba(0, 0, 0, 0.05)' },
          ticks: { stepSize: 1, color: isDark ? '#94a3b8' : '#64748b' }
        },
        x: {
          grid: { display: false },
          ticks: { color: isDark ? '#cbd5e1' : '#334155', font: { weight: '600' } }
        }
      },
      plugins: {
        legend: { display: false }
      }
    }
  });
}

// Gráfico 4: Scatter Plot de Evasão (Inatividade vs Média)
function renderRiskScatter(summary) {
  const canvas = document.getElementById('chartRiskScatter');
  if (!canvas || typeof Chart === 'undefined') return;
  destroyAnalyticsChart('scatter');

  const students = summary.students_risk_plot || [];
  
  // Agrupa pontos idênticos para adicionar leve jitter visual para facilitar identificação
  const coordCount = {};
  const scatterData = students.map((s, idx) => {
    const isNever = s.days_inactive >= 900;
    const baseClampedX = isNever ? 35 : Math.min(s.days_inactive, 35);
    const key = `${baseClampedX}_${s.average_score}`;
    coordCount[key] = (coordCount[key] || 0) + 1;
    const offset = (coordCount[key] - 1) * 0.25;

    return {
      x: baseClampedX + (offset > 0 ? (offset % 2 === 0 ? offset : -offset) : 0),
      realDays: s.days_inactive,
      isNever: isNever,
      y: s.average_score,
      name: s.name,
      cat: s.risk_category
    };
  });

  const ctx = canvas.getContext('2d');
  const isDark = state.theme === 'dark';

  analyticsState.charts.scatter = new Chart(ctx, {
    type: 'scatter',
    data: {
      datasets: [{
        label: 'Estudantes',
        data: scatterData,
        backgroundColor: function(context) {
          const raw = context.raw;
          if (!raw) return '#10b981';
          if (raw.cat === 'critico') return '#ef4444';
          if (raw.cat === 'moderado') return '#f59e0b';
          return '#10b981';
        },
        pointRadius: 6,
        pointHoverRadius: 9
      }]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      scales: {
        x: {
          title: {
            display: true,
            text: 'Dias Sem Acesso ao Canvas LMS (até 35+ dias)',
            color: isDark ? '#94a3b8' : '#64748b',
            font: { weight: '600' }
          },
          min: 0,
          max: 36,
          grid: { color: isDark ? 'rgba(255, 255, 255, 0.05)' : 'rgba(0, 0, 0, 0.05)' },
          ticks: {
            color: isDark ? '#cbd5e1' : '#334155',
            callback: function(val) {
              return val >= 35 ? '35+ d' : val + 'd';
            }
          }
        },
        y: {
          title: {
            display: true,
            text: 'Nota Média Acumulada (0 a 100)',
            color: isDark ? '#94a3b8' : '#64748b',
            font: { weight: '600' }
          },
          min: 0,
          max: 100,
          grid: { color: isDark ? 'rgba(255, 255, 255, 0.05)' : 'rgba(0, 0, 0, 0.05)' },
          ticks: { color: isDark ? '#cbd5e1' : '#334155' }
        }
      },
      plugins: {
        legend: { display: false },
        tooltip: {
          callbacks: {
            label: function(context) {
              const raw = context.raw;
              const inatLabel = raw.isNever ? 'Sem registros de acesso' : `${raw.realDays}d inativo`;
              const scoreLabel = raw.y > 0 ? `Nota: ${raw.y} pts` : 'Sem avaliações lançadas';
              let statusLabel = '🟢 Em dia';
              if (raw.cat === 'critico') statusLabel = '🔴 Risco de evasão';
              else if (raw.cat === 'moderado') statusLabel = '🟡 Acompanhamento';
              return ` ${raw.name} • ${statusLabel} (${scoreLabel} | ${inatLabel})`;
            }
          }
        }
      }
    }
  });
}

// Eventos da toolbar de Analytics
document.getElementById('analyticsCourseSelect')?.addEventListener('change', async (e) => {
  const chosenCourseId = e.target.value;
  if (!chosenCourseId) return;

  analyticsState.selectedCourseId = chosenCourseId;
  state.selectedCourseId = chosenCourseId;
  localStorage.setItem('canvas_hub_active_course', chosenCourseId);

  // Sincroniza outros seletores se compatíveis
  if (elements.courseSelect) {
    const has = Array.from(elements.courseSelect.querySelectorAll('option')).some(o => o.value === chosenCourseId);
    if (has) elements.courseSelect.value = chosenCourseId;
  }
  if (elements.studentsCourseSelect) {
    const has = Array.from(elements.studentsCourseSelect.querySelectorAll('option')).some(o => o.value === chosenCourseId);
    if (has) elements.studentsCourseSelect.value = chosenCourseId;
  }

  // Atualiza cartões de curso silenciosamente
  document.querySelectorAll('.course-card').forEach(card => {
    if (card.getAttribute('data-id') === chosenCourseId) {
      card.classList.add('selected');
    } else {
      card.classList.remove('selected');
    }
  });

  await fetchAndRenderAnalytics(chosenCourseId);
});

document.getElementById('btnRefreshAnalytics')?.addEventListener('click', async () => {
  if (analyticsState.selectedCourseId) {
    showToast('Atualizando métricas da disciplina em tempo real...', 'info');
    await fetchAndRenderAnalytics(analyticsState.selectedCourseId, true);
  }
});

document.getElementById('btnExportAnalytics')?.addEventListener('click', () => {
  window.print();
});

// Redimensionamento responsivo seguro em resize de janela
let analyticsResizeDebounce = null;
window.addEventListener('resize', () => {
  clearTimeout(analyticsResizeDebounce);
  analyticsResizeDebounce = setTimeout(() => {
    const activeTab = document.querySelector('.tab-content.active');
    if (activeTab && activeTab.id === 'analytics-tab') {
      Object.values(analyticsState.charts).forEach(c => {
        if (c) c.resize();
      });
    }
  }, 150);
});



