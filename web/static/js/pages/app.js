// pages/app.js — /app/* 外壳：拉取 profile 展示问候（真实内容由后续阶段各页 JS 渲染）
(function () {
  if (!location.pathname.startsWith('/app/')) return;
  const root = document.getElementById('app-root');
  const placeholder = document.getElementById('app-placeholder');

  function esc(s) {
    const d = document.createElement('div');
    d.textContent = s == null ? '' : String(s);
    return d.innerHTML;
  }
  function pageKey(p) {
    return ({ tasks: 'nav.tasks', phrases: 'nav.phrases', checkin: 'nav.checkin', logs: 'nav.logs', settings: 'nav.settings' })[p] || 'app.name';
  }

  (async () => {
    try {
      const profile = await window.Ly.api('/auth/profile');
      if (placeholder) placeholder.remove();
      root.innerHTML =
        '<p>👋 ' + esc(profile.email) + '</p>' +
        '<p class="muted">' + esc(window.Ly.t(pageKey(window.__PAGE__))) + ' · ' + esc(window.Ly.t('common.loading')) + '</p>';
    } catch (err) {
      if (root) root.innerHTML = '<p class="error">' + esc(err.message) + '</p>';
    }
  })();
})();
