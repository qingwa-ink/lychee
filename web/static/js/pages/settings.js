// pages/settings.js — 个人设置：展示账号、修改密码（旧密码/验证码二选一）
(function () {
  if (!location.pathname.startsWith('/app/settings')) return;
  const emailEl = document.getElementById('settings-email');
  const form = document.getElementById('pw-form');
  const msg = document.getElementById('pw-msg');
  const submit = document.getElementById('pw-submit');
  const sendBtn = document.getElementById('pw-send-code');
  let email = '';

  function esc(s) {
    const d = document.createElement('div');
    d.textContent = s == null ? '' : String(s);
    return d.innerHTML;
  }

  // 拉取账号信息
  (async () => {
    try {
      const p = await window.Ly.api('/auth/profile');
      email = p.email || '';
      emailEl.textContent = '📧 ' + email;
    } catch (e) {
      emailEl.textContent = window.Ly.t('common.error');
    }
  })();

  // 切换验证方式：禁用另一组字段
  function syncMethod() {
    const method = form.method.value;
    form.querySelector('[name="old"]').disabled = method !== 'old';
    form.querySelector('[name="code"]').disabled = method !== 'code';
    sendBtn.disabled = method !== 'code';
  }
  form.querySelectorAll('input[name="method"]').forEach(r => r.addEventListener('change', syncMethod));
  syncMethod();

  // 发送验证码（type=change_password）
  if (sendBtn) {
    sendBtn.addEventListener('click', async () => {
      if (!email) { window.Ly.showMsg(msg, window.Ly.t('common.error'), true); return; }
      sendBtn.disabled = true;
      try {
        await window.Ly.api('/auth/send-code', { method: 'POST', body: { email, type: 'change_password' } });
        window.Ly.showMsg(msg, window.Ly.t('auth.send_code') + ' ✓');
        let cd = 60; const orig = sendBtn.textContent;
        const timer = setInterval(() => {
          cd -= 1;
          if (cd <= 0) { clearInterval(timer); sendBtn.textContent = orig; if (form.method.value === 'code') sendBtn.disabled = false; }
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
    const method = form.method.value;
    const newPassword = form.new.value;
    if (newPassword !== form.confirm.value) {
      window.Ly.showMsg(msg, window.Ly.t('auth.confirm_password'), true);
      return;
    }
    const body = { new_password: newPassword };
    if (method === 'old') {
      body.old_password = form.old.value;
      if (!body.old_password) { window.Ly.showMsg(msg, window.Ly.t('auth.old_password'), true); return; }
    } else {
      body.code = form.code.value.trim();
      if (!body.code) { window.Ly.showMsg(msg, window.Ly.t('auth.code'), true); return; }
    }
    submit.disabled = true;
    try {
      await window.Ly.api('/auth/password', { method: 'PUT', body });
      window.Ly.showMsg(msg, window.Ly.t('settings.pw_changed'));
      setTimeout(() => { window.Ly.clearTokens(); location.href = '/login'; }, 1500);
    } catch (err) {
      window.Ly.showMsg(msg, err.message || window.Ly.t('common.error'), true);
      submit.disabled = false;
    }
  });
})();
