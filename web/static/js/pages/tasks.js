// pages/tasks.js — 任务看板：分组树 + 任务列表（筛选/排序/分页）+ 编辑弹框（常用语插入/复制）
(function () {
  if (!location.pathname.startsWith('/app/tasks')) return;
  const PAGE_SIZE = 10;
  const t = window.Ly.t;
  const msg = () => document.getElementById('task-msg');

  let selectedGroupId = null;   // null = 全部任务
  let selectedGroupName = '';
  let filters = { status: '', priority: '', sort: 'created_at', order: 'desc' };
  let page = 1;
  let phraseCache = null;

  function esc(s) {
    const d = document.createElement('div');
    d.textContent = s == null ? '' : String(s);
    return d.innerHTML;
  }
  function showMsg(m, isErr) { window.Ly.showMsg(msg(), m, isErr); }

  // ---------- 分组树 ----------
  async function loadTree(selectId) {
    try {
      const data = await window.Ly.api('/task-groups');
      const tree = (data && data.tree) || [];
      renderTree(tree);
      // 维持/初始化选中
      if (selectId != null) {
        selectedGroupId = selectId;
      } else if (selectedGroupId == null && tree.length) {
        selectedGroupId = null; // 默认停在“全部”，不擅自选中
      }
      syncGroupUI();
      await loadTasks();
    } catch (e) {
      showMsg(e.message || t('common.error'), true);
    }
  }

  function renderTree(tree) {
    const root = document.getElementById('group-tree');
    root.innerHTML = '';
    // 虚拟“全部任务”节点
    const allLi = document.createElement('li');
    allLi.innerHTML = '<a href="#" class="group-item' + (selectedGroupId == null ? ' active' : '') + '" data-id="">🗂️ ' + esc(t('task.all')) + '</a>';
    allLi.querySelector('a').addEventListener('click', (e) => { e.preventDefault(); selectGroup(null, t('task.all')); });
    root.appendChild(allLi);
    for (const node of tree) root.appendChild(renderNode(node, 0));
  }

  function renderNode(node, depth) {
    const li = document.createElement('li');
    const a = document.createElement('a');
    a.href = '#';
    a.className = 'group-item' + (selectedGroupId === node.id ? ' active' : '');
    a.style.paddingLeft = (0.6 + depth * 1) + 'rem';
    a.dataset.id = node.id;
    a.textContent = '📁 ' + node.name;
    a.addEventListener('click', (e) => { e.preventDefault(); selectGroup(node.id, node.name); });
    li.appendChild(a);
    if (node.children && node.children.length) {
      const ul = document.createElement('ul');
      for (const ch of node.children) ul.appendChild(renderNode(ch, depth + 1));
      li.appendChild(ul);
    }
    return li;
  }

  function selectGroup(id, name) {
    selectedGroupId = id;
    selectedGroupName = id == null ? t('task.all') : name;
    syncGroupUI();
    page = 1;
    loadTasks();
  }

  function syncGroupUI() {
    document.getElementById('current-group-name').textContent = selectedGroupName || t('task.all');
    document.getElementById('group-actions').hidden = selectedGroupId == null;
    document.getElementById('new-task-btn').disabled = selectedGroupId == null;
    document.getElementById('new-task-btn').title = selectedGroupId == null ? t('task.no_group') : '';
    // 高亮
    document.querySelectorAll('.group-item').forEach(a => {
      const id = a.dataset.id;
      const cur = id === '' ? null : Number(id);
      a.classList.toggle('active', cur === selectedGroupId);
    });
  }

  // 分组增删改
  document.getElementById('add-root-group').addEventListener('click', async () => {
    const name = prompt(t('task.add_group') + ':');
    if (!name || !name.trim()) return;
    try {
      await window.Ly.api('/task-groups', { method: 'POST', body: { name: name.trim() } });
      await loadTree();
    } catch (e) { showMsg(e.message || t('common.error'), true); }
  });
  document.getElementById('add-sub-group').addEventListener('click', async () => {
    if (selectedGroupId == null) return;
    const name = prompt(t('task.add_subgroup') + ':');
    if (!name || !name.trim()) return;
    try {
      await window.Ly.api('/task-groups', { method: 'POST', body: { parent_id: selectedGroupId, name: name.trim() } });
      await loadTree(selectedGroupId);
    } catch (e) { showMsg(e.message || t('common.error'), true); }
  });
  document.getElementById('rename-group').addEventListener('click', async () => {
    if (selectedGroupId == null) return;
    const name = prompt(t('common.edit') + ':', selectedGroupName);
    if (name == null || !name.trim()) return;
    try {
      await window.Ly.api('/task-groups/' + selectedGroupId, { method: 'PUT', body: { name: name.trim() } });
      selectedGroupName = name.trim();
      await loadTree(selectedGroupId);
    } catch (e) { showMsg(e.message || t('common.error'), true); }
  });
  document.getElementById('del-group').addEventListener('click', async () => {
    if (selectedGroupId == null) return;
    if (!confirm(t('task.confirm_delete_group'))) return;
    try {
      await window.Ly.api('/task-groups/' + selectedGroupId, { method: 'DELETE' });
      selectedGroupId = null;
      await loadTree();
    } catch (e) { showMsg(e.message || t('common.error'), true); }
  });

  // ---------- 任务列表 ----------
  async function loadTasks() {
    const params = new URLSearchParams({
      page: String(page), page_size: String(PAGE_SIZE),
      sort: filters.sort, order: filters.order,
    });
    if (selectedGroupId != null) params.set('group_id', selectedGroupId);
    if (filters.status) params.set('status', filters.status);
    if (filters.priority !== '') params.set('priority', filters.priority);
    try {
      const data = await window.Ly.api('/tasks?' + params.toString());
      renderTasks(data);
    } catch (e) {
      showMsg(e.message || t('common.error'), true);
    }
  }

  function statusLabel(s) {
    return ({ editing: t('task.status_editing'), pending: t('task.status_pending'), completed: t('task.status_completed') })[s] || s;
  }
  function statusClass(s) { return 'st st-' + s; }
  function dueText(iso) {
    if (!iso) return '';
    return (iso + '').slice(0, 10);
  }

  function renderTasks(data) {
    const list = (data && data.list) || [];
    const total = (data && data.total) || 0;
    const listEl = document.getElementById('task-list');
    const emptyEl = document.getElementById('task-empty');
    listEl.innerHTML = '';
    if (!list.length) {
      emptyEl.hidden = false;
      renderPager(total);
      return;
    }
    emptyEl.hidden = true;
    for (const task of list) {
      const card = document.createElement('article');
      card.className = 'task-card';
      const pri = task.priority;
      card.innerHTML =
        '<div class="task-row">' +
          '<span class="pri pri-' + pri + '">P' + pri + '</span>' +
          '<span class="' + statusClass(task.status) + '">' + esc(statusLabel(task.status)) + '</span>' +
          (task.due_date ? '<span class="due">⏰ ' + esc(dueText(task.due_date)) + '</span>' : '') +
          '<span class="task-spacer"></span>' +
          '<button class="copy-task small">' + t('common.copy') + '</button>' +
          '<button class="edit-task small">' + t('common.edit') + '</button>' +
          '<button class="del-task small danger">' + t('common.delete') + '</button>' +
        '</div>' +
        '<div class="task-content">' + esc(task.content) + '</div>';
      card.querySelector('.copy-task').addEventListener('click', () => copyText(task.content));
      card.querySelector('.edit-task').addEventListener('click', () => openEdit(task));
      card.querySelector('.del-task').addEventListener('click', () => delTask(task));
      listEl.appendChild(card);
    }
    renderPager(total);
  }

  function renderPager(total) {
    const pager = document.getElementById('task-pager');
    const pages = Math.max(1, Math.ceil(total / PAGE_SIZE));
    if (pages <= 1) { pager.innerHTML = ''; return; }
    pager.innerHTML =
      '<button id="tp-prev"' + (page <= 1 ? ' disabled' : '') + '>‹ ' + t('common.prev') + '</button> ' +
      '<span>' + page + ' / ' + pages + '</span> ' +
      '<button id="tp-next"' + (page >= pages ? ' disabled' : '') + '>' + t('common.next') + ' ›</button>';
    document.getElementById('tp-prev').addEventListener('click', () => { page--; loadTasks(); });
    document.getElementById('tp-next').addEventListener('click', () => { page++; loadTasks(); });
  }

  // 筛选/排序联动
  ['f-status', 'f-priority', 'f-sort', 'f-order'].forEach(id => {
    document.getElementById(id).addEventListener('change', () => {
      filters.status = document.getElementById('f-status').value;
      filters.priority = document.getElementById('f-priority').value;
      filters.sort = document.getElementById('f-sort').value;
      filters.order = document.getElementById('f-order').value;
      page = 1;
      loadTasks();
    });
  });

  // ---------- 编辑弹框 ----------
  const dlg = document.getElementById('task-edit-dlg');
  const fId = document.getElementById('task-id');
  const fContent = document.getElementById('task-content');
  const fPriority = document.getElementById('task-priority');
  const fStatus = document.getElementById('task-status');
  const fDue = document.getElementById('task-due');
  const fPhrase = document.getElementById('task-phrase-insert');

  async function ensurePhrases() {
    if (phraseCache) return phraseCache;
    try {
      const data = await window.Ly.api('/phrases?page=1&page_size=100');
      phraseCache = (data && data.list) || [];
    } catch (e) { phraseCache = []; }
    return phraseCache;
  }
  async function fillPhraseOptions() {
    const list = await ensurePhrases();
    fPhrase.innerHTML = '<option value="">' + t('task.insert_phrase') + '</option>';
    for (const p of list) {
      const o = document.createElement('option');
      o.value = p.content;
      o.textContent = p.content.length > 30 ? p.content.slice(0, 30) + '…' : p.content;
      fPhrase.appendChild(o);
    }
  }

  // 在 textarea 光标处插入
  fPhrase.addEventListener('change', () => {
    const val = fPhrase.value;
    if (!val) return;
    const s = fContent.selectionStart || 0;
    const e = fContent.selectionEnd || 0;
    fContent.value = fContent.value.slice(0, s) + val + fContent.value.slice(e);
    const pos = s + val.length;
    fContent.focus();
    fContent.setSelectionRange(pos, pos);
    fPhrase.value = '';
  });

  document.getElementById('task-copy').addEventListener('click', () => copyText(fContent.value));

  function copyText(text) {
    const ok = () => window.Ly.toast(t('task.copy_done'), 2000);
    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(text).then(ok, () => {});
    } else {
      ok();
    }
  }

  document.getElementById('new-task-btn').addEventListener('click', () => openEdit(null));

  function openEdit(task) {
    // task==null → 新建；需选中分组
    if (!task && selectedGroupId == null) { showMsg(t('task.no_group'), true); return; }
    fillPhraseOptions();
    document.getElementById('task-edit-title').textContent = task ? t('task.edit_task') : t('task.new_task');
    fId.value = task ? task.id : '';
    fContent.value = task ? task.content : '';
    fPriority.value = String(task ? task.priority : 3);
    fStatus.value = task ? task.status : 'editing';
    fDue.value = task && task.due_date ? dueText(task.due_date) : '';
    dlg.showModal();
  }
  document.getElementById('task-cancel').addEventListener('click', () => dlg.close());

  document.getElementById('task-edit-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const content = fContent.value.trim();
    if (!content) { showMsg(t('task.content'), true); return; }
    const due = fDue.value ? fDue.value + 'T00:00:00+08:00' : null;
    const id = fId.value;
    try {
      if (id) {
        await window.Ly.api('/tasks/' + id, {
          method: 'PUT',
          body: { content, priority: Number(fPriority.value), status: fStatus.value, due_date: due },
        });
      } else {
        await window.Ly.api('/tasks', {
          method: 'POST',
          body: { group_id: selectedGroupId, content, priority: Number(fPriority.value), status: fStatus.value, due_date: due },
        });
      }
      dlg.close();
      await loadTasks();
    } catch (err) {
      showMsg(err.message || t('common.error'), true);
    }
  });

  async function delTask(task) {
    if (!confirm(t('task.confirm_delete_task'))) return;
    try {
      await window.Ly.api('/tasks/' + task.id, { method: 'DELETE' });
      await loadTasks();
    } catch (e) { showMsg(e.message || t('common.error'), true); }
  }

  // 初始化
  loadTree();
})();
