// auth.js — 前端路由守卫（页面外壳无服务端鉴权，由这里 + API JWT 中间件共同保护）
(function () {
  function guard() {
    const p = location.pathname;
    const authed = window.Ly.isLoggedIn();
    // 未登录访问 /app/* → 去登录
    if (p.startsWith('/app/') && !authed) {
      window.Ly.goLogin();
      return;
    }
    // 已登录访问登录页/首页 → 进任务页
    if ((p === '/login' || p === '/') && authed) {
      location.href = '/app/tasks';
    }
  }
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', guard);
  } else {
    guard();
  }
})();
