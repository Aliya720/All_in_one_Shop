async function loadSummary() {
  if (!Auth.isLoggedIn()) {
    window.location.href = "/login.html?next=/checkout.html";
    return;
  }

  try {
    const cart = await api("/cart");
    const summary = qs("#checkout-summary");

    if (!cart.items || cart.items.length === 0) {
      summary.innerHTML = `<p class="muted">Your cart is empty.</p>`;
      qs("#place-order-btn").disabled = true;
      return;
    }

    summary.innerHTML =
      cart.items
        .map(
          (i) => `
        <div class="summary-row">
          <span>${escapeHtml(i.name)} × ${i.quantity}</span>
          <span>${formatPrice(i.subtotal)}</span>
        </div>
      `
        )
        .join("") + `<div class="summary-row total"><span>Total</span><span>${formatPrice(cart.total)}</span></div>`;
  } catch (e) {
    showAlert(qs("#alert-box"), e.message);
  }
}

document.addEventListener("DOMContentLoaded", () => {
  loadSummary();

  qs("#checkout-form").addEventListener("submit", async (e) => {
    e.preventDefault();
    const alertBox = qs("#alert-box");
    const address = qs("#address").value.trim();
    const btn = qs("#place-order-btn");

    if (!address) {
      showAlert(alertBox, "Please enter a shipping address.");
      return;
    }

    btn.disabled = true;
    btn.textContent = "Placing order...";

    try {
      const order = await api("/orders/checkout", { method: "POST", body: { shipping_address: address } });
      window.location.href = "/orders.html?placed=" + order.id;
    } catch (err) {
      showAlert(alertBox, err.message);
      btn.disabled = false;
      btn.textContent = "Place Order";
    }
  });
});
