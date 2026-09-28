async function loadProduct() {
  const slug = getQueryParam("slug");
  const container = qs("#product-detail");
  const alertBox = qs("#alert-box");

  if (!slug) {
    container.innerHTML = `<div class="empty-state">No product specified.</div>`;
    return;
  }

  try {
    const p = await api("/products/" + encodeURIComponent(slug));

    container.innerHTML = `
      <div class="grid" style="grid-template-columns: 1fr 1fr; gap: 32px; align-items: start;">
        <div class="card" style="padding:16px;">
          <img src="${p.image_url || "/img/placeholder.svg"}" alt="${escapeHtml(p.name)}"
               style="width:100%; border-radius:8px;" onerror="this.src='/img/placeholder.svg'" />
        </div>
        <div>
          <div class="muted">${escapeHtml(p.category_name || "")}</div>
          <h1 style="margin:6px 0;">${escapeHtml(p.name)}</h1>
          <div class="price" style="font-size:1.4rem;">${formatPrice(p.price)}</div>
          <p style="margin:16px 0; color:#4b5563;">${escapeHtml(p.description || "")}</p>
          <p class="muted">${p.stock > 0 ? p.stock + " in stock" : "Out of stock"}</p>
          <div style="display:flex; gap:10px; align-items:center; margin-top:16px;">
            <input type="number" id="qty-input" value="1" min="1" max="${p.stock}" style="width:70px; padding:8px; border:1px solid var(--color-border); border-radius:6px;" ${p.stock === 0 ? "disabled" : ""}/>
            <button class="btn" id="add-to-cart-btn" ${p.stock === 0 ? "disabled" : ""}>Add to Cart</button>
          </div>
        </div>
      </div>
    `;

    qs("#add-to-cart-btn")?.addEventListener("click", async () => {
      if (!Auth.isLoggedIn()) {
        window.location.href = "/login.html?next=/product.html?slug=" + encodeURIComponent(slug);
        return;
      }
      const quantity = parseInt(qs("#qty-input").value, 10) || 1;
      try {
        await api("/cart/items", { method: "POST", body: { product_id: p.id, quantity } });
        showAlert(alertBox, "Added to cart!", "success");
      } catch (e) {
        showAlert(alertBox, e.message);
      }
    });
  } catch (e) {
    container.innerHTML = `<div class="empty-state">Product not found.</div>`;
  }
}

document.addEventListener("DOMContentLoaded", loadProduct);
