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
        <a class="brand" href="/index.html">ShopSphere</a>
        <nav class="main-nav">
          <a href="/index.html">Products</a>
          <a href="/cart.html">Cart</a>
          ${loggedIn ? '<a href="/orders.html">Orders</a>' : ""}
          ${isAdmin ? '<a href="/admin.html">Admin</a>' : ""}
          ${
            loggedIn
              ? `<span class="muted">Hi, ${escapeHtml(user.name)}</span><a href="#" id="logout-link">Logout</a>`
              : '<a href="/login.html">Login</a><a href="/register.html">Register</a>'
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
