document.addEventListener("DOMContentLoaded", () => {
  const loginForm = qs("#login-form");
  const registerForm = qs("#register-form");

  if (loginForm) {
    loginForm.addEventListener("submit", async (e) => {
      e.preventDefault();
      const alertBox = qs("#alert-box");
      const email = qs("#email").value.trim();
      const password = qs("#password").value;

      try {
        const result = await api("/auth/login", { method: "POST", body: { email, password } });
        Auth.setSession(result.token, result.user);
        const next = getQueryParam("next") || appUrl("index.html");
        window.location.href = next;
      } catch (err) {
        showAlert(alertBox, err.message);
      }
    });
  }

  if (registerForm) {
    registerForm.addEventListener("submit", async (e) => {
      e.preventDefault();
      const alertBox = qs("#alert-box");
      const name = qs("#name").value.trim();
      const email = qs("#email").value.trim();
      const password = qs("#password").value;

      try {
        const result = await api("/auth/register", { method: "POST", body: { name, email, password } });
        Auth.setSession(result.token, result.user);
        window.location.href = appUrl("index.html");
      } catch (err) {
        showAlert(alertBox, err.message);
      }
    });
  }
});
