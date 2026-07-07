// core.js — API 封装（token + 静默刷新 + 4290）、i18n、通用工具
(function () {
  const TOK = { access: 'access_token', refresh: 'refresh_token' };

  function getToken(k) { return localStorage.getItem(k) || ''; }
  function setTokens(access, refresh) {
    if (access) localStorage.setItem(TOK.access, access);
    if (refresh) localStorage.setItem(TOK.refresh, refresh);
  }
  function clearTokens() {
    localStorage.removeItem(TOK.access);
    localStorage.removeItem(TOK.refresh);
  }
  function isLoggedIn() { return !!getToken(TOK.access); }

  // i18n：取注入的文案表
  function t(key) {
    const m = window.__I18N__ || {};
    return m[key] || key;
  }

  function goLogin() {
    clearTokens();
    if (!location.pathname.startsWith('/login')) location.href = '/login';
  }

  function err(code, message, ratelimited) {
    const e = new Error(message || '');
    e.code = code; e.message = message; e.ratelimited = !!ratelimited;
    return e;
  }

  // 原始 fetch + 统一 envelope 解析
  async function rawFetch(path, method, body, withAuth) {
    const headers = { 'Content-Type': 'application/json' };
    if (withAuth) {
      const at = getToken(TOK.access);
      if (at) headers['Authorization'] = 'Bearer ' + at;
    }
    const res = await fetch('/api/v1' + path, {
      method, headers,
      body: body == null ? undefined : JSON.stringify(body),
    });
    let json = null;
    try { json = await res.json(); } catch (e) { /* 非 JSON 响应 */ }
    return { status: res.status, json };
  }

  // 用 refresh token 换新双 token；失败清空并跳登录
  async function refreshTokens() {
    const rt = getToken(TOK.refresh);
    if (!rt) { goLogin(); throw err(0, 'no refresh token'); }
    const { json } = await rawFetch('/auth/refresh', 'POST', { refresh_token: rt }, false);
    if (json && json.code === 0 && json.data) {
      setTokens(json.data.access_token, json.data.refresh_token);
      return true;
    }
    goLogin();
    throw err(json ? json.code : 0, json ? json.message : 'refresh failed');
  }

  // 对外 API：成功 resolve(data)，失败 reject({code,message,ratelimited})
  async function api(path, opts) {
    opts = opts || {};
    const method = opts.method || 'GET';
    const withAuth = opts.auth !== false;
    let { json } = await rawFetch(path, method, opts.body, withAuth);
    if (!json) throw err(0, 'Network error');

    // token 过期（4011）：静默刷新后重试一次
    if (withAuth && json.code === 4011) {
      try {
        await refreshTokens();
        const retry = await rawFetch(path, method, opts.body, true);
        json = retry.json;
        if (!json) throw err(0, 'Network error');
      } catch (e) {
        throw err(4011, t('auth.session_expired'));
      }
    }

    if (json.code === 0) return json.data;

    // 鉴权请求 4010：跳登录（登录页本身 auth=false，不受影响）
    if (withAuth && json.code === 4010) {
      goLogin();
      throw err(4010, json.message);
    }
    if (json.code === 4290) throw err(4290, t('auth.ratelimited'), true);
    throw err(json.code, json.message);
  }

  function showMsg(el, message, isErr) {
    if (!el) return;
    el.textContent = message;
    el.hidden = false;
    el.classList.toggle('error', !!isErr);
    el.classList.toggle('ok', !isErr);
  }

  // 轻量 toast：屏幕底部居中浮层，ms 毫秒后自动消失（默认 2s）。连续触发会重置计时并覆盖文案。
  let toastTimer = null;
  function toast(message, ms) {
    ms = ms || 2000;
    let el = document.getElementById('ly-toast');
    if (!el) {
      el = document.createElement('div');
      el.id = 'ly-toast';
      el.className = 'ly-toast';
      document.body.appendChild(el);
    }
    el.textContent = message;
    el.classList.add('show');
    if (toastTimer) clearTimeout(toastTimer);
    toastTimer = setTimeout(() => el.classList.remove('show'), ms);
  }

  async function switchLocale(loc) {
    try {
      await api('/locale', { method: 'PUT', body: { locale: loc }, auth: isLoggedIn() });
    } catch (e) { /* 即便失败也刷新页面 */ }
    location.reload();
    return false;
  }

  async function logout() {
    try { await api('/auth/logout', { method: 'POST' }); } catch (e) {}
    clearTokens();
    location.href = '/login';
    return false;
  }

  window.Ly = {
    api, t, showMsg, toast, setTokens, clearTokens, getToken,
    isLoggedIn, goLogin, switchLocale, logout, TOK,
  };
  // 布局里的内联 onclick（登出、语种切换）调用的是全局裸函数，故额外暴露到 window。
  window.logout = logout;
  window.switchLocale = switchLocale;
  window.toast = toast;
})();
