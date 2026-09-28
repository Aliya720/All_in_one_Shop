async function loadOrders() {
  if (!Auth.isLoggedIn()) {
    window.location.href = "/login.html?next=/orders.html";
    return;
  }

  const container = qs("#orders-list");
  const alertBox = qs("#alert-box");
  const placedId = getQueryParam("placed");

  if (placedId) {
    showAlert(alertBox, `Order #${placedId} placed successfully!`, "success");
  }

  try {
    const result = await api("/orders");
    const orders = result.items || [];

    if (orders.length === 0) {
      container.innerHTML = `<div class="empty-state">You haven't placed any orders yet. <a href="/index.html">Start shopping</a></div>`;
      return;
    }

    container.innerHTML = orders
      .map(
        (o) => `
      <div class="card" style="padding:16px; margin-bottom:14px;">
        <div style="display:flex; justify-content:space-between; align-items:center;">
          <div>
            <strong>Order #${o.id}</strong>
            <span class="badge ${o.status}">${o.status}</span>
          </div>
          <div class="price">${formatPrice(o.total)}</div>
        </div>
        <div class="muted" style="margin-top:6px;">Placed on ${new Date(o.created_at).toLocaleString()}</div>
        <div class="muted">Shipping to: ${escapeHtml(o.shipping_address)}</div>
      </div>
    `
      )
      .join("");
  } catch (e) {
    showAlert(alertBox, e.message);
  }
}

document.addEventListener("DOMContentLoaded", loadOrders);
