// pages/phrases.js — 常用语管理（CRUD + 分页）
(function () {
  if (!location.pathname.startsWith('/app/phrases')) return;
  const PAGE_SIZE = 10;
  const listEl = document.getElementById('phrase-list');
  const emptyEl = document.getElementById('phrase-empty');
  const pagerEl = document.getElementById('phrase-pager');
  const msg = document.getElementById('phrase-msg');
  const form = document.getElementById('phrase-form');
  const input = document.getElementById('phrase-input');
  const submitBtn = document.getElementById('phrase-submit');
  const tableEl = document.getElementById('phrase-table');
  const dlg = document.getElementById('phrase-edit-dlg');
  const editInput = document.getElementById('phrase-edit-input');
  const editCancel = document.getElementById('phrase-edit-cancel');
  let editingId = null;
  let page = 1;

  function esc(s) {
    const d = document.createElement('div');
    d.textContent = s == null ? '' : String(s);
    return d.innerHTML;
  }

  async function load(p) {
    page = p || 1;
    try {
      const data = await window.Ly.api('/phrases?page=' + page + '&page_size=' + PAGE_SIZE);
      render(data);
    } catch (err) {
      window.Ly.showMsg(msg, err.message || window.Ly.t('common.error'), true);
    }
  }

  function render(data) {
    const list = (data && data.list) || [];
    const total = (data && data.total) || 0;
    listEl.innerHTML = '';
    if (!list.length) {
      tableEl.hidden = true;
      emptyEl.hidden = false;
      pagerEl.innerHTML = '';
      return;
    }
    tableEl.hidden = false;
    emptyEl.hidden = true;
    for (const item of list) {
      const tr = document.createElement('tr');
      tr.innerHTML =
        '<td class="phrase-cell">' + esc(item.content) + '</td>' +
        '<td class="muted small">' + esc((item.updated_at || '').replace('T', ' ').slice(0, 16)) + '</td>' +
        '<td><button class="copy-btn">' + window.Ly.t('common.copy') + '</button> <button class="edit-btn">' + window.Ly.t('common.edit') + '</button> <button class="del-btn danger">' + window.Ly.t('common.delete') + '</button></td>';
      tr.querySelector('.copy-btn').addEventListener('click', () => copyToClipboard(item.content));
      tr.querySelector('.edit-btn').addEventListener('click', () => openEdit(item));
      tr.querySelector('.del-btn').addEventListener('click', () => del(item));
      listEl.appendChild(tr);
    }
    renderPager(total);
  }

  async function copyToClipboard(text) {
    try {
      await navigator.clipboard.writeText(text);
      window.Ly.toast(window.Ly.t('common.copied'));
    } catch (err) {
      window.Ly.toast(window.Ly.t('common.copy_failed'), true);
    }
  }

  function renderPager(total) {
    const pages = Math.max(1, Math.ceil(total / PAGE_SIZE));
    if (pages <= 1) { pagerEl.innerHTML = ''; return; }
    pagerEl.innerHTML =
      '<button id="pg-prev"' + (page <= 1 ? ' disabled' : '') + '>‹ ' + window.Ly.t('common.prev') + '</button> ' +
      '<span>' + page + ' / ' + pages + '</span> ' +
      '<button id="pg-next"' + (page >= pages ? ' disabled' : '') + '>' + window.Ly.t('common.next') + ' ›</button>';
    document.getElementById('pg-prev').addEventListener('click', () => load(page - 1));
    document.getElementById('pg-next').addEventListener('click', () => load(page + 1));
  }

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    const content = input.value.trim();
    if (!content) return;
    submitBtn.disabled = true;
    try {
      await window.Ly.api('/phrases', { method: 'POST', body: { content } });
      input.value = '';
      window.Ly.toast(window.Ly.t('common.saved'));
      await load(1);
    } catch (err) {
      window.Ly.toast(err.message || window.Ly.t('common.error'), true);
    } finally {
      submitBtn.disabled = false;
    }
  });

  function openEdit(item) {
    editingId = item.id;
    editInput.value = item.content;
    dlg.showModal();
  }
  editCancel.addEventListener('click', () => { dlg.close(); editingId = null; });
  document.getElementById('phrase-edit-form').addEventListener('submit', async (e) => {
    // dialog 表单 submit 即"保存"
    if (!editingId) return;
    e.preventDefault();
    const content = editInput.value.trim();
    if (!content) return;
    try {
      await window.Ly.api('/phrases/' + editingId, { method: 'PUT', body: { content } });
      dlg.close();
      editingId = null;
      window.Ly.toast(window.Ly.t('common.saved'));
      await load(page);
    } catch (err) {
      dlg.close();
      window.Ly.toast(err.message || window.Ly.t('common.error'), true);
    }
  });

  async function del(item) {
    if (!confirm(window.Ly.t('common.confirm_delete'))) return;
    try {
      await window.Ly.api('/phrases/' + item.id, { method: 'DELETE' });
      window.Ly.toast(window.Ly.t('common.deleted'));
      await load(page);
    } catch (err) {
      window.Ly.toast(err.message || window.Ly.t('common.error'), true);
    }
  }

  load(1);
})();
