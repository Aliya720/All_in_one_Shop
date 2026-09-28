/**
 * Thin fetch wrapper for the e-commerce API.
 *
 * Assumes the frontend is served from the same origin as the API under
 * /api/ (this is how the bundled Nginx config and Docker Compose setup
 * are wired). If you host the frontend elsewhere, set window.API_BASE
 * before this script loads, e.g.:
 *   <script>window.API_BASE = "https://shop.example.com/api";</script>
 */
const APP_BASE = new URL(".", window.location.href).pathname.replace(/\/$/, "");

function appUrl(path = "") {
  if (/^(?:[a-z]+:)?\/\//i.test(path) || path.startsWith("data:")) return path;
  return `${APP_BASE}/${String(path).replace(/^\/+/, "")}`;
}

const API_BASE = window.API_BASE || appUrl("api");
const TOKEN_KEY = "ecommerce_token";
const USER_KEY = "ecommerce_user";

const Auth = {
  getToken() {
    return localStorage.getItem(TOKEN_KEY);
  },
  getUser() {
    const raw = localStorage.getItem(USER_KEY);
    return raw ? JSON.parse(raw) : null;
  },
  setSession(token, user) {
    localStorage.setItem(TOKEN_KEY, token);
    localStorage.setItem(USER_KEY, JSON.stringify(user));
  },
  clearSession() {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
  },
  isLoggedIn() {
    return !!this.getToken();
  },
  isAdmin() {
    const u = this.getUser();
    return !!u && u.role === "admin";
  },
  logout() {
    this.clearSession();
    window.location.href = appUrl("login.html");
  },
};

/**
 * api(path, options) -> parsed JSON body.
 * Automatically attaches the Bearer token when logged in, and throws an
 * Error with the server's error message on non-2xx responses.
 */
async function api(path, options = {}) {
  const headers = Object.assign(
    {},
    options.body && !(options.body instanceof FormData) ? { "Content-Type": "application/json" } : {},
    options.headers || {}
  );

  const token = Auth.getToken();
  if (token) {
    headers["Authorization"] = "Bearer " + token;
  }

  const res = await fetch(API_BASE + path, {
    ...options,
    headers,
    body:
      options.body && !(options.body instanceof FormData)
        ? JSON.stringify(options.body)
        : options.body,
  });

  let data = null;
  const text = await res.text();
  if (text) {
    try {
      data = JSON.parse(text);
    } catch (e) {
      data = null;
    }
  }

  if (!res.ok) {
    const message = (data && data.error) || `Request failed (${res.status})`;
    const err = new Error(message);
    err.status = res.status;
    throw err;
  }

  return data;
}

function formatPrice(value) {
  const n = Number(value) || 0;
  return "₹" + n.toLocaleString("en-IN", { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}

function qs(selector, root = document) {
  return root.querySelector(selector);
}

function qsa(selector, root = document) {
  return Array.from(root.querySelectorAll(selector));
}

function showAlert(container, message, type = "error") {
  container.innerHTML = `<div class="alert ${type}">${escapeHtml(message)}</div>`;
}

function escapeHtml(str) {
  const div = document.createElement("div");
  div.textContent = str;
  return div.innerHTML;
}

function getQueryParam(name) {
  return new URLSearchParams(window.location.search).get(name);
}
