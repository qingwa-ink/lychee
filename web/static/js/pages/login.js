// pages/login.js — 登录
(function () {
  const form = document.getElementById('auth-form');
  if (!form || form.dataset.action !== 'login') return;
  const msg = document.getElementById('auth-msg');
  const submit = document.getElementById('auth-submit');

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    const email = form.email.value.trim();
    const password = form.password.value;
    window.Ly.showMsg(msg, window.Ly.t('common.loading'));
    submit.disabled = true;
    try {
      const data = await window.Ly.api('/auth/login', {
        method: 'POST', body: { email, password }, auth: false,
      });
      window.Ly.setTokens(data.access_token, data.refresh_token);
      location.href = '/app/tasks';
    } catch (err) {
      window.Ly.showMsg(msg, err.message || window.Ly.t('common.error'), true);
      submit.disabled = false;
    }
  });
})();
