// pages/logs.js — 操作日志 / 登录日志 + Chart.js 柱状图（日/小时切换）
(function () {
  if (!location.pathname.startsWith('/app/logs')) return;
  const t = window.Ly.t;
  const PAGE_SIZE = 10;
  const msg = () => document.getElementById('log-msg');

  function esc(s) {
    const d = document.createElement('div');
    d.textContent = s == null ? '' : String(s);
    return d.innerHTML;
  }
  function fmtDT(s) { return (s + '').replace('T', ' ').slice(0, 16); }
  function daysAgoISO(n) {
    const d = new Date(Date.now() - n * 864e5);
    return d.getFullYear() + '-' + String(d.getMonth() + 1).padStart(2, '0') + '-' + String(d.getDate()).padStart(2, '0');
  }

  // ---------- 图表 ----------
  let chart = null;
  const dimEl = document.getElementById('log-dim');
  const cStart = document.getElementById('log-start');
  const cEnd = document.getElementById('log-end');
  cStart.value = daysAgoISO(6);
  cEnd.value = daysAgoISO(0);

  function initChart() {
    const ctx = document.getElementById('log-chart').getContext('2d');
    chart = new Chart(ctx, {
      type: 'bar',
      data: { labels: [], datasets: [{ label: t('logs.chart_title'), data: [], backgroundColor: '#0070f3' }] },
      options: {
        responsive: true,
        plugins: { legend: { display: false } },
        scales: { y: { beginAtZero: true, ticks: { precision: 0 } } },
      },
    });
  }

  async function loadChart() {
    const params = new URLSearchParams({ dimension: dimEl.value });
    if (cStart.value) params.set('start', cStart.value);
    if (cEnd.value) params.set('end', cEnd.value);
    try {
      const data = await window.Ly.api('/logs/operations/report?' + params.toString());
      const buckets = (data && data.buckets) || [];
      chart.data.labels = buckets.map(b => b.bucket);
      chart.data.datasets[0].data = buckets.map(b => b.count);
      chart.update();
    } catch (e) {
      window.Ly.showMsg(msg(), e.message || t('common.error'), true);
    }
  }
  document.getElementById('log-redraw').addEventListener('click', loadChart);
  dimEl.addEventListener('change', loadChart);

  // ---------- 操作历史 ----------
  let opPage = 1;
  const opCat = document.getElementById('op-cat');
  const opStart = document.getElementById('op-start');
  const opEnd = document.getElementById('op-end');

  async function loadOps(p) {
    opPage = p || 1;
    const params = new URLSearchParams({ page: String(opPage), page_size: String(PAGE_SIZE) });
    if (opCat.value) params.set('category', opCat.value);
    if (opStart.value) params.set('start', opStart.value);
    if (opEnd.value) params.set('end', opEnd.value);
    try {
      const data = await window.Ly.api('/logs/operations?' + params.toString());
      renderOps(data);
    } catch (e) { window.Ly.showMsg(msg(), e.message || t('common.error'), true); }
  }
  function renderOps(data) {
    const list = (data && data.list) || [];
    const total = (data && data.total) || 0;
    const tbody = document.getElementById('op-list');
    const empty = document.getElementById('op-empty');
    const table = document.getElementById('op-table');
    tbody.innerHTML = '';
    if (!list.length) { table.hidden = true; empty.hidden = false; renderPager('op-pager', total, opPage, loadOps); return; }
    table.hidden = false; empty.hidden = true;
    for (const r of list) {
      const tr = document.createElement('tr');
      tr.innerHTML =
        '<td class="small muted">' + esc(fmtDT(r.created_at)) + '</td>' +
        '<td>' + esc(r.category) + '</td>' +
        '<td>' + esc(r.action || (r.method + ' ')) + '</td>' +
        '<td class="small">' + esc(r.path) + '</td>' +
        '<td class="small">' + esc(r.ip) + '</td>';
      tbody.appendChild(tr);
    }
    renderPager('op-pager', total, opPage, loadOps);
  }
  document.getElementById('op-search').addEventListener('click', () => loadOps(1));

  // ---------- 登录日志 ----------
  let lgPage = 1;
  async function loadLogins(p) {
    lgPage = p || 1;
    const params = new URLSearchParams({ page: String(lgPage), page_size: String(PAGE_SIZE) });
    try {
      const data = await window.Ly.api('/logs/logins?' + params.toString());
      renderLogins(data);
    } catch (e) { window.Ly.showMsg(msg(), e.message || t('common.error'), true); }
  }
  function renderLogins(data) {
    const list = (data && data.list) || [];
    const total = (data && data.total) || 0;
    const tbody = document.getElementById('lg-list');
    const empty = document.getElementById('lg-empty');
    const table = document.getElementById('lg-table');
    tbody.innerHTML = '';
    if (!list.length) { table.hidden = true; empty.hidden = false; renderPager('lg-pager', total, lgPage, loadLogins); return; }
    table.hidden = false; empty.hidden = true;
    for (const r of list) {
      const tr = document.createElement('tr');
      tr.innerHTML =
        '<td class="small muted">' + esc(fmtDT(r.created_at)) + '</td>' +
        '<td class="small">' + esc(r.ip) + '</td>' +
        '<td class="small muted ua-cell">' + esc((r.ua || '').slice(0, 80)) + '</td>';
      tbody.appendChild(tr);
    }
    renderPager('lg-pager', total, lgPage, loadLogins);
  }

  function renderPager(id, total, current, loader) {
    const pager = document.getElementById(id);
    const pages = Math.max(1, Math.ceil(total / PAGE_SIZE));
    if (pages <= 1) { pager.innerHTML = ''; return; }
    pager.innerHTML =
      '<button data-d="-1"' + (current <= 1 ? ' disabled' : '') + '>‹ ' + t('common.prev') + '</button> ' +
      '<span>' + current + ' / ' + pages + '</span> ' +
      '<button data-d="1"' + (current >= pages ? ' disabled' : '') + '>' + t('common.next') + ' ›</button>';
    pager.querySelectorAll('button').forEach(b => {
      b.addEventListener('click', () => loader(current + Number(b.dataset.d)));
    });
  }

  // 初始化
  initChart();
  loadChart();
  loadOps(1);
  loadLogins(1);
})();
