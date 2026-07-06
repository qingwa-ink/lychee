// pages/register.js — 注册（发送验证码 + 注册）
(function () {
  const form = document.getElementById('auth-form');
  if (!form || form.dataset.action !== 'register') return;
  const msg = document.getElementById('auth-msg');
  const submit = document.getElementById('auth-submit');
  const sendBtn = document.getElementById('send-code-btn');

  // 发送验证码（按钮倒计时 60s，顺带规避 1s 限流）
  if (sendBtn) {
    let timer = null;
    sendBtn.addEventListener('click', async () => {
      const email = form.email.value.trim();
      if (!email) { window.Ly.showMsg(msg, window.Ly.t('auth.email'), true); return; }
      sendBtn.disabled = true;
      try {
        await window.Ly.api('/auth/send-code', { method: 'POST', body: { email, type: 'register' }, auth: false });
        window.Ly.showMsg(msg, window.Ly.t('auth.send_code') + ' ✓');
        let cd = 60; const orig = sendBtn.textContent;
        timer = setInterval(() => {
          cd -= 1;
          if (cd <= 0) { clearInterval(timer); sendBtn.textContent = orig; sendBtn.disabled = false; }
          else { sendBtn.textContent = orig + '(' + cd + ')'; }
        }, 1000);
      } catch (err) {
        window.Ly.showMsg(msg, err.message, true);
        sendBtn.disabled = false;
      }
    });
  }

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    const email = form.email.value.trim();
    const password = form.password.value;
    const code = form.code.value.trim();
    if (password !== form.confirm.value) {
      window.Ly.showMsg(msg, window.Ly.t('auth.confirm_password'), true);
      return;
    }
    submit.disabled = true;
    try {
      await window.Ly.api('/auth/register', { method: 'POST', body: { email, password, code }, auth: false });
      window.Ly.showMsg(msg, window.Ly.t('auth.register_done'));
      setTimeout(() => { location.href = '/login'; }, 1200);
    } catch (err) {
      window.Ly.showMsg(msg, err.message || window.Ly.t('common.error'), true);
      submit.disabled = false;
    }
  });
})();
