// pages/forgot.js — 忘记密码（发送验证码 + 重置）
(function () {
  const form = document.getElementById('auth-form');
  if (!form || form.dataset.action !== 'forgot') return;
  const msg = document.getElementById('auth-msg');
  const submit = document.getElementById('auth-submit');
  const sendBtn = document.getElementById('send-code-btn');

  if (sendBtn) {
    sendBtn.addEventListener('click', async () => {
      const email = form.email.value.trim();
      if (!email) { window.Ly.showMsg(msg, window.Ly.t('auth.email'), true); return; }
      sendBtn.disabled = true;
      try {
        await window.Ly.api('/auth/send-code', { method: 'POST', body: { email, type: 'forgot' }, auth: false });
        window.Ly.showMsg(msg, window.Ly.t('auth.send_code') + ' ✓');
        let cd = 60; const orig = sendBtn.textContent;
        const timer = setInterval(() => {
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
    const code = form.code.value.trim();
    const new_password = form.new_password.value;
    submit.disabled = true;
    try {
      await window.Ly.api('/auth/forgot-password', {
        method: 'POST', body: { email, code, new_password }, auth: false,
      });
      window.Ly.showMsg(msg, window.Ly.t('auth.reset_done'));
      setTimeout(() => { location.href = '/login'; }, 1500);
    } catch (err) {
      window.Ly.showMsg(msg, err.message || window.Ly.t('common.error'), true);
      submit.disabled = false;
    }
  });
})();
