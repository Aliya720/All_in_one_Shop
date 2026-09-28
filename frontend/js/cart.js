async function loadCart() {
  const alertBox = qs("#alert-box");

  if (!Auth.isLoggedIn()) {
    window.location.href = "/login.html?next=/cart.html";
    return;
  }

  try {
    const cart = await api("/cart");
    renderCart(cart);
  } catch (e) {
    showAlert(alertBox, e.message);
  }
}

function renderCart(cart) {
  const container = qs("#cart-items");
  const checkoutBtn = qs("#checkout-btn");

  if (!cart.items || cart.items.length === 0) {
    container.innerHTML = `<div class="empty-state">Your cart is empty. <a href="/index.html">Browse products</a></div>`;
    qs("#summary-count").textContent = "0";
    qs("#summary-total").textContent = formatPrice(0);
    checkoutBtn.classList.add("btn-disabled");
    checkoutBtn.style.pointerEvents = "none";
    checkoutBtn.style.opacity = "0.5";
    return;
  }

  checkoutBtn.style.pointerEvents = "";
  checkoutBtn.style.opacity = "";

  container.innerHTML = cart.items
    .map(
      (item) => `
      <div class="cart-row" data-product-id="${item.product_id}">
        <img src="${item.image_url || "/img/placeholder.svg"}" alt="${escapeHtml(item.name)}" onerror="this.src='/img/placeholder.svg'" />
        <div class="grow">
          <div><a href="/product.html?slug=${encodeURIComponent(item.slug)}">${escapeHtml(item.name)}</a></div>
          <div class="muted">${formatPrice(item.price)} each</div>
        </div>
        <div class="qty-control">
          <button class="qty-decrease">-</button>
          <span class="qty-value">${item.quantity}</span>
          <button class="qty-increase">+</button>
        </div>
        <div class="price" style="width:110px; text-align:right;">${formatPrice(item.subtotal)}</div>
        <button class="btn secondary remove-item">Remove</button>
      </div>
    `
    )
    .join("");

  qs("#summary-count").textContent = cart.items.reduce((sum, i) => sum + i.quantity, 0);
  qs("#summary-total").textContent = formatPrice(cart.total);

  qsa(".cart-row").forEach((row) => {
    const productId = parseInt(row.dataset.productId, 10);
    const qtyValueEl = row.querySelector(".qty-value");

    row.querySelector(".qty-increase").addEventListener("click", async () => {
      const newQty = parseInt(qtyValueEl.textContent, 10) + 1;
      await updateQuantity(productId, newQty);
    });

    row.querySelector(".qty-decrease").addEventListener("click", async () => {
      const newQty = parseInt(qtyValueEl.textContent, 10) - 1;
      if (newQty < 1) return;
      await updateQuantity(productId, newQty);
    });

    row.querySelector(".remove-item").addEventListener("click", async () => {
      try {
        const cart = await api("/cart/items/" + productId, { method: "DELETE" });
        renderCart(cart);
      } catch (e) {
        showAlert(qs("#alert-box"), e.message);
      }
    });
  });
}

async function updateQuantity(productId, quantity) {
  try {
    const cart = await api("/cart/items/" + productId, { method: "PUT", body: { quantity } });
    renderCart(cart);
  } catch (e) {
    showAlert(qs("#alert-box"), e.message);
  }
}

document.addEventListener("DOMContentLoaded", loadCart);
