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

    // Pré-seleciona por padrão a primeira turma ativa do semestre vigente
    const currentCourses = state.courses.filter(c => c.is_current_term);
    if (currentCourses.length > 0) {
      state.selectedCourseId = String(currentCourses[0].id);
    } else if (state.courses.length > 0) {
      state.selectedCourseId = String(state.courses[0].id);
    }

    // Atualiza status de conexão
    elements.connectionStatus.innerHTML = `
      <span class="status-dot"></span>
      <span class="status-text">Conectado ao Canvas</span>
    `;

    renderCourses();
    populateSelects();

    // Sincroniza o select com a turma pré-selecionada
    if (elements.courseSelect && state.selectedCourseId) {
      elements.courseSelect.value = state.selectedCourseId;
    }

    // Carrega tarefas em paralelo para calcular métricas
    await loadAllAssignments();

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
  // Limpa selects
  elements.courseSelect.innerHTML = '<option value="all">Todas as Disciplinas</option>';
  elements.announcementCoursesList.innerHTML = '';
  elements.announcementHistoryCourseSelect.innerHTML = '';
  elements.studentsCourseSelect.innerHTML = '';

  const currentCourses = state.courses.filter(c => c.is_current_term);
  const pastCourses = state.courses.filter(c => !c.is_current_term);

  // Helper para preencher um select agrupado
  const appendOptGroup = (selectEl, label, list) => {
    if (list.length === 0) return;
    const group = document.createElement('optgroup');
    group.label = label;
    list.forEach(course => {
      const opt = document.createElement('option');
      opt.value = course.id;
      opt.textContent = `${course.clean_name || course.name} (${course.period || course.id})`;
      group.appendChild(opt);
    });
    selectEl.appendChild(group);
  };

  // Select de Cursos na aba 1
  appendOptGroup(elements.courseSelect, '🟢 Semestre Vigente (Atual)', currentCourses);
  appendOptGroup(elements.courseSelect, '⚪ Semestres Anteriores (Concluídos)', pastCourses);

  // Select de histórico de avisos
  appendOptGroup(elements.announcementHistoryCourseSelect, '🟢 Semestre Vigente', currentCourses);
  appendOptGroup(elements.announcementHistoryCourseSelect, '⚪ Semestres Anteriores', pastCourses);

  // Select de Alunos na aba 3
  appendOptGroup(elements.studentsCourseSelect, '🟢 Semestre Vigente', currentCourses);
  appendOptGroup(elements.studentsCourseSelect, '⚪ Semestres Anteriores', pastCourses);

  // Checkboxes de envio de avisos (apenas ativas marcadas por padrão)
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
  state.selectedCourseId = String(courseId);
  elements.courseSelect.value = state.selectedCourseId;

  document.querySelectorAll('.course-card').forEach(card => {
    if (card.getAttribute('data-id') === state.selectedCourseId) {
      card.classList.add('selected');
    } else {
      card.classList.remove('selected');
    }
  });

  updateCourseBanner();
  renderAssignmentsList();
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
        ${btnAIEvaluate}
        <a href="${canvasUrl}" target="_blank" rel="noopener noreferrer" class="btn btn-secondary" style="padding: 0.4rem 0.75rem; font-size: 0.8125rem;">
          Ver no Canvas
        </a>
      </div>
    `;

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

elements.courseSelect.addEventListener('change', (e) => {
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
  loadStudentsForCourse(e.target.value);
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
      elements.aiProviderBadge.textContent = providerNames[info.provider] || info.provider.toUpperCase();
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

function renderChatMessage(role, content) {
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

  // Exibe o thinking
  elements.thinkingStatusText.textContent = 'O assistente está pensando e consultando o Canvas...';
  elements.chatThinking.style.display = 'flex';
  elements.chatMessages.scrollTop = elements.chatMessages.scrollHeight;

  try {
    const res = await fetch('/api/chat', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        message: text,
        history: state.chatHistory.slice(0, -1) // histórico anterior
      })
    });

    elements.chatThinking.style.display = 'none';

    if (!res.ok) {
      const errText = await res.text();
      renderChatMessage('assistant', `⚠️ **Erro na comunicação:** ${errText}`);
      return;
    }

    const data = await res.json();
    const reply = data.reply || 'Não obtive resposta do modelo de IA.';

    renderChatMessage('assistant', reply);
    state.chatHistory.push({ role: 'assistant', content: reply });
    localStorage.setItem('afya_chat_history', JSON.stringify(state.chatHistory));

    // Se houver card de ação sensível
    if (data.action_card) {
      renderActionCard(data.action_card);
    }
  } catch (err) {
    elements.chatThinking.style.display = 'none';
    renderChatMessage('assistant', `⚠️ **Erro de rede:** ${err.message}`);
  } finally {
    state.isGenerating = false;
    elements.btnSendChat.disabled = false;
    elements.chatInput.focus();
  }
}

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
      renderChatMessage(msg.role, msg.content);
    });
  }
}

// Inicialização automática
initApp();
initChatEvents();

