/**
 * Renders the shared site header/footer into any page that includes a
 * <div id="site-header"></div> / <div id="site-footer"></div>, and keeps
 * the nav links in sync with login state.
 */
function renderLayout() {
  const header = document.getElementById("site-header");
  const footer = document.getElementById("site-footer");

  if (header) {
    const loggedIn = Auth.isLoggedIn();
    const isAdmin = Auth.isAdmin();
    const user = Auth.getUser();

    header.innerHTML = `
      <div class="header-inner">
        <a class="brand" href="${appUrl("index.html")}">ShopSphere</a>
        <nav class="main-nav">
          <a href="${appUrl("index.html")}">Products</a>
          <a href="${appUrl("cart.html")}">Cart</a>
          ${loggedIn ? `<a href="${appUrl("orders.html")}">Orders</a>` : ""}
          ${isAdmin ? `<a href="${appUrl("admin.html")}">Admin</a>` : ""}
          ${
            loggedIn
              ? `<span class="muted">Hi, ${escapeHtml(user.name)}</span><a href="#" id="logout-link">Logout</a>`
              : `<a href="${appUrl("login.html")}">Login</a><a href="${appUrl("register.html")}">Register</a>`
          }
        </nav>
      </div>
    `;

    const logoutLink = document.getElementById("logout-link");
    if (logoutLink) {
      logoutLink.addEventListener("click", (e) => {
        e.preventDefault();
        Auth.logout();
      });
    }
  }

  if (footer) {
    footer.innerHTML = `<p>ShopSphere &mdash; demo storefront built with Go, PostgreSQL &amp; vanilla JS.</p>`;
  }
}

document.addEventListener("DOMContentLoaded", renderLayout);
