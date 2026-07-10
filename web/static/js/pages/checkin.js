// pages/checkin.js — 打卡与健康：记录录入 / 每日目标 / 每日报告（完成度进度条）
(function () {
  if (!location.pathname.startsWith('/app/checkin')) return;
  const t = window.Ly.t;
  const TYPES = ['water', 'exercise', 'nap'];
  const UNITS = { water: 'ml', exercise: 'min', nap: 'min' };
  const typeLabel = (k) => t('checkin.' + k);

  const dateEl = document.getElementById('ci-date');
  const msg = document.getElementById('ci-msg');
  const reportEl = document.getElementById('ci-report');
  const reportEmpty = document.getElementById('ci-report-empty');
  const listEl = document.getElementById('ci-list');
  const listEmpty = document.getElementById('ci-list-empty');
  const tableEl = document.getElementById('ci-table');
  const goalsEl = document.getElementById('ci-goals');

  function esc(s) {
    const d = document.createElement('div');
    d.textContent = s == null ? '' : String(s);
    return d.innerHTML;
  }
  function todayISO() {
    const n = new Date();
    return n.getFullYear() + '-' + String(n.getMonth() + 1).padStart(2, '0') + '-' + String(n.getDate()).padStart(2, '0');
  }
  function fmt(n) { return Math.round(n * 100) / 100; }

  dateEl.value = todayISO();
  dateEl.addEventListener('change', () => loadAll(dateEl.value));
  document.getElementById('ci-today').addEventListener('click', () => { dateEl.value = todayISO(); loadAll(dateEl.value); });

  async function loadAll(date) {
    try {
      const [report, records, goals] = await Promise.all([
        window.Ly.api('/check-in/report?date=' + date),
        window.Ly.api('/check-in/records?date=' + date + '&page=1&page_size=50'),
        window.Ly.api('/check-in/goals'),
      ]);
      renderReport(report, goals);
      renderRecords(records);
      renderGoals(goals);
    } catch (e) {
      window.Ly.toast(e.message || t('common.error'), 2000, true);
    }
  }

  function renderReport(report, goalsResp) {
    const types = (report && report.types) || [];
    reportEl.innerHTML = '';
    if (!types.length) { reportEmpty.hidden = false; return; }
    reportEmpty.hidden = true;
    for (const tr of types) {
      const rate = tr.has_goal && tr.daily_target > 0 ? tr.achievement_rate : 0;
      const pct = Math.min(100, Math.round(rate * 100));
      const unit = tr.unit || UNITS[tr.type] || '';
      const div = document.createElement('div');
      div.className = 'ci-report-item';
      div.innerHTML =
        '<div class="ci-row">' +
          '<strong>' + esc(typeLabel(tr.type)) + '</strong>' +
          '<span class="ci-total">' + fmt(tr.total) + ' ' + esc(unit) +
            (tr.has_goal ? ' / ' + fmt(tr.daily_target) + ' ' + esc(unit) : '') + '</span>' +
        '</div>' +
        '<div class="ly-progress"><span style="width:' + pct + '%"></span></div>' +
        '<div class="muted small">' +
          (tr.has_goal
            ? t('checkin.achievement') + ' ' + Math.round(rate * 100) + '%'
            : t('checkin.no_goal')) +
        '</div>';
      reportEl.appendChild(div);
    }
  }

  function renderRecords(data) {
    const list = (data && data.list) || [];
    listEl.innerHTML = '';
    if (!list.length) { tableEl.hidden = true; listEmpty.hidden = false; return; }
    tableEl.hidden = false; listEmpty.hidden = true;
    for (const r of list) {
      const tr = document.createElement('tr');
      tr.innerHTML =
        '<td>' + esc(typeLabel(r.type)) + '</td>' +
        '<td>' + fmt(r.value) + ' ' + esc(r.unit || UNITS[r.type] || '') + '</td>' +
        '<td class="small muted">' + esc(r.record_date) + '</td>';
      listEl.appendChild(tr);
    }
  }

  function renderGoals(goalsResp) {
    const goals = (goalsResp && goalsResp.list) || [];
    const byType = {};
    for (const g of goals) byType[g.type] = g;
    goalsEl.innerHTML = '';
    for (const ty of TYPES) {
      const g = byType[ty];
      const cur = g ? g.daily_target : '';
      const unit = (g && g.unit) || UNITS[ty];
      const row = document.createElement('div');
      row.className = 'ci-goal-row';
      row.innerHTML =
        '<span class="ci-goal-label">' + esc(typeLabel(ty)) + ' (' + esc(unit) + ')</span>' +
        '<input type="number" class="ci-goal-input" min="0.1" step="0.1" value="' + (cur === '' ? '' : cur) + '" placeholder="' + esc(t('checkin.target')) + '">' +
        '<button type="button" class="small">' + t('checkin.set_target') + '</button>';
      row.querySelector('button').addEventListener('click', async () => {
        const v = parseFloat(row.querySelector('input').value);
        if (!(v > 0)) { window.Ly.toast(t('checkin.value'), 2000, true); return; }
        try {
          await window.Ly.api('/check-in/goals', { method: 'PUT', body: { type: ty, daily_target: v, unit } });
          window.Ly.toast(t('checkin.set_target') + ' ✓', 2000);
          await loadAll(dateEl.value);
        } catch (e) { window.Ly.toast(e.message || t('common.error'), 2000, true); }
      });
      goalsEl.appendChild(row);
    }
  }

  document.getElementById('ci-record-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const type = document.getElementById('ci-type').value;
    const value = parseFloat(document.getElementById('ci-value').value);
    if (!(value > 0)) { window.Ly.toast(t('checkin.value'), 2000, true); return; }
    const btn = document.getElementById('ci-submit');
    btn.disabled = true;
    try {
      await window.Ly.api('/check-in/records', {
        method: 'POST', body: { type, value, record_date: dateEl.value },
      });
      document.getElementById('ci-value').value = '';
      window.Ly.toast(t('checkin.recorded') + ' ✓', 2000);
      await loadAll(dateEl.value);
    } catch (err) {
      window.Ly.toast(err.message || t('common.error'), 2000, true);
    } finally {
      btn.disabled = false;
    }
  });

  loadAll(dateEl.value);
})();
