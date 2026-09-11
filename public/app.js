// Estado global da aplicação
const state = {
  user: null,
  courses: [],
  selectedCourseId: 'all',
  assignments: {}, // courseId -> array
  currentFilter: 'all',
  students: [],
  theme: localStorage.getItem('canvas_hub_theme') || 'dark'
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

    // Atualiza status de conexão
    elements.connectionStatus.innerHTML = `
      <span class="status-dot"></span>
      <span class="status-text">Conectado ao Canvas</span>
    `;

    renderCourses();
    populateSelects();

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

  state.courses.forEach(course => {
    // Select de Cursos na aba 1
    const opt = document.createElement('option');
    opt.value = course.id;
    opt.textContent = course.name;
    elements.courseSelect.appendChild(opt);

    // Checkboxes na aba 2
    const checkItem = document.createElement('label');
    checkItem.className = 'checkbox-item';
    checkItem.innerHTML = `
      <input type="checkbox" name="announcement_course" value="${course.id}" checked>
      <span>${course.name}</span>
    `;
    elements.announcementCoursesList.appendChild(checkItem);

    // Select de histórico de avisos
    const optHist = document.createElement('option');
    optHist.value = course.id;
    optHist.textContent = course.name;
    elements.announcementHistoryCourseSelect.appendChild(optHist);

    // Select de Alunos na aba 3
    const optStud = document.createElement('option');
    optStud.value = course.id;
    optStud.textContent = course.name;
    elements.studentsCourseSelect.appendChild(optStud);
  });

  if (state.courses.length > 0) {
    loadAnnouncementsForCourse(state.courses[0].id);
  }
}

function renderCourses() {
  elements.totalCoursesCount.textContent = state.courses.length;
  elements.coursesCardsGrid.innerHTML = '';

  state.courses.forEach(course => {
    const card = document.createElement('div');
    card.className = `course-card ${state.selectedCourseId === String(course.id) ? 'selected' : ''}`;
    card.setAttribute('data-id', course.id);

    const students = course.total_students !== undefined ? course.total_students : '-';

    card.innerHTML = `
      <div>
        <span class="course-badge">ID: ${course.id}</span>
        <h3 class="course-title">${course.name}</h3>
      </div>
      <div class="course-footer">
        <span class="course-students-count">
          👥 ${students} alunos
        </span>
        <span class="course-term">${course.course_code ? course.course_code.slice(0, 18) + '...' : ''}</span>
      </div>
    `;

    card.addEventListener('click', () => {
      selectCourse(course.id);
    });

    elements.coursesCardsGrid.appendChild(card);
  });
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

  renderAssignmentsList();
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

        totalAssignments += assignments.length;
        assignments.forEach(a => {
          totalPending += (a.needs_grading_count || 0);
        });
      }
    } catch (e) {
      console.error(`Erro ao carregar tarefas do curso ${course.id}:`, e);
    }
  }

  elements.totalAssignmentsCount.textContent = totalAssignments;
  elements.totalPendingCount.textContent = totalPending;

  renderAssignmentsList();
}

function renderAssignmentsList() {
  let list = [];
  if (state.selectedCourseId === 'all') {
    elements.assignmentsTitle.textContent = 'Atividades de Todas as Disciplinas';
    elements.assignmentsSubtitle.textContent = 'Consolidado de todas as suas disciplinas ativas';
    for (const courseId in state.assignments) {
      list.push(...(state.assignments[courseId] || []));
    }
  } else {
    const course = state.courses.find(c => String(c.id) === state.selectedCourseId);
    elements.assignmentsTitle.textContent = `Atividades: ${course ? course.name : ''}`;
    elements.assignmentsSubtitle.textContent = `Listagem de tarefas cadastradas nesta turma`;
    list = state.assignments[state.selectedCourseId] || [];
  }

  // Aplica filtro de pendências
  if (state.currentFilter === 'pending') {
    list = list.filter(a => (a.needs_grading_count || 0) > 0);
  }

  if (list.length === 0) {
    elements.assignmentsList.innerHTML = `
      <div class="empty-state">
        <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>
        <p>Nenhuma atividade encontrada com o filtro selecionado.</p>
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
    div.innerHTML = `
      <div class="assignment-info">
        <h4 class="assignment-title">${item.name}</h4>
        <div class="assignment-meta">
          <span>📅 Entrega: ${dueDate}</span>
          <span>🏆 Pontos: ${item.points_possible !== null && item.points_possible !== undefined ? item.points_possible : 'N/A'}</span>
          ${item.submission_types ? `<span>📝 Tipo: ${item.submission_types.join(', ')}</span>` : ''}
        </div>
      </div>
      <div class="assignment-actions">
        <span class="grading-badge ${isPending ? 'pending' : 'done'}">
          ${isPending ? `⚠️ ${pending} para corrigir` : '✓ Todas corrigidas'}
        </span>
        <a href="${canvasUrl}" target="_blank" rel="noopener noreferrer" class="btn btn-secondary" style="padding: 0.4rem 0.75rem; font-size: 0.8125rem;">
          Abrir no Canvas
        </a>
      </div>
    `;
    elements.assignmentsList.appendChild(div);
  });
}

// Filtros de Atividades
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
elements.selectAllCoursesBtn.addEventListener('click', () => {
  document.querySelectorAll('input[name="announcement_course"]').forEach(cb => cb.checked = true);
});

elements.unselectAllCoursesBtn.addEventListener('click', () => {
  document.querySelectorAll('input[name="announcement_course"]').forEach(cb => cb.checked = false);
});

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
    <span>Disparar Comunicado</span>
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
      elements.announcementsHistoryList.innerHTML = '<div class="empty-state sm"><p>Nenhum comunicado recente nesta disciplina.</p></div>';
      return;
    }

    elements.announcementsHistoryList.innerHTML = '';
    announcements.forEach(a => {
      const item = document.createElement('div');
      item.className = 'history-item';
      const cleanMessage = (a.message || '').replace(/<[^>]*>?/gm, '');
      item.innerHTML = `
        <h4 class="history-item-title">${a.title}</h4>
        <div class="history-item-date">${formatDate(a.posted_at || a.created_at)}</div>
        <p class="history-item-body">${cleanMessage || 'Sem conteúdo de texto'}</p>
      `;
      elements.announcementsHistoryList.appendChild(item);
    });
  } catch (err) {
    elements.announcementsHistoryList.innerHTML = '<div class="empty-state sm"><p>Erro ao carregar avisos.</p></div>';
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
    <tr><td colspan="4" class="table-placeholder">Carregando alunos...</td></tr>
  `;

  try {
    const resp = await fetch(`/api/courses/${courseId}/students`);
    if (!resp.ok) throw new Error('Falha ao obter lista de alunos');
    state.students = await resp.json();
    renderStudentsTable(state.students);
  } catch (err) {
    elements.studentsTableBody.innerHTML = `
      <tr><td colspan="4" class="table-placeholder">Erro ao carregar alunos: ${err.message}</td></tr>
    `;
  }
}

function renderStudentsTable(studentsList) {
  elements.studentsCounter.textContent = `${studentsList.length} alunos`;

  if (studentsList.length === 0) {
    elements.studentsTableBody.innerHTML = `
      <tr><td colspan="4" class="table-placeholder">Nenhum aluno encontrado nesta disciplina.</td></tr>
    `;
    return;
  }

  elements.studentsTableBody.innerHTML = '';
  studentsList.forEach(student => {
    const tr = document.createElement('tr');
    const avatar = student.avatar_url 
      ? `<img src="${student.avatar_url}" alt="" class="student-avatar-img">`
      : `<div class="student-avatar-img" style="display:flex;align-items:center;justify-content:center;font-size:12px;">👤</div>`;

    tr.innerHTML = `
      <td style="width: 60px;">${avatar}</td>
      <td><div class="student-row-user">${student.name || student.sortable_name}</div></td>
      <td><code>${student.id}</code></td>
      <td><span class="grading-badge done">Matriculado</span></td>
    `;
    elements.studentsTableBody.appendChild(tr);
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
  const courseName = course ? course.name.replace(/[^a-zA-Z0-9_-]/g, '_') : 'alunos';

  let csvContent = "data:text/csv;charset=utf-8,ID_Canvas;Nome_Completo;Status\n";
  state.students.forEach(s => {
    const row = `"${s.id}";"${s.name || s.sortable_name}";"Matriculado"`;
    csvContent += row + "\n";
  });

  const encodedUri = encodeURI(csvContent);
  const link = document.createElement('a');
  link.setAttribute('href', encodedUri);
  link.setAttribute('download', `Canvas_Alunos_${courseName}.csv`);
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);

  showToast('Arquivo CSV exportado com sucesso!');
});

// Inicialização automática
initApp();
