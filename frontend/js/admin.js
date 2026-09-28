let allCategories = [];

function guardAdmin() {
  if (!Auth.isLoggedIn()) {
    window.location.href = appUrl("login.html?next=" + encodeURIComponent(appUrl("admin.html")));
    return false;
  }
  if (!Auth.isAdmin()) {
    document.querySelector("main").innerHTML = `<div class="empty-state">You do not have access to this page.</div>`;
    return false;
  }
  return true;
}

function showModal(id) {
  qs("#modal-overlay").style.display = "block";
  qs("#" + id).style.display = "block";
}
function hideModals() {
  qs("#modal-overlay").style.display = "none";
  qs("#product-modal").style.display = "none";
  qs("#category-modal").style.display = "none";
}

// ---------- Stats ----------
async function loadStats() {
  try {
    const stats = await api("/admin/stats");
    qs("#stats-grid").innerHTML = `
      <div class="card stat-card"><div class="value">${stats.total_orders}</div><div class="label">Total Orders</div></div>
      <div class="card stat-card"><div class="value">${formatPrice(stats.total_revenue)}</div><div class="label">Revenue</div></div>
      <div class="card stat-card"><div class="value">${stats.total_products}</div><div class="label">Products</div></div>
      <div class="card stat-card"><div class="value">${stats.total_users}</div><div class="label">Users</div></div>
    `;
  } catch (e) {
    showAlert(qs("#alert-box"), e.message);
  }
}

// ---------- Categories ----------
async function loadCategories() {
  try {
    allCategories = await api("/categories");
    const tbody = qs("#categories-table tbody");
    tbody.innerHTML = allCategories
      .map(
        (c) => `
      <tr>
        <td>${escapeHtml(c.name)}</td>
        <td class="muted">${escapeHtml(c.slug)}</td>
        <td>${escapeHtml(c.description || "")}</td>
        <td>
          <button class="btn secondary edit-category" data-id="${c.id}">Edit</button>
          <button class="btn danger delete-category" data-id="${c.id}">Delete</button>
        </td>
      </tr>
    `
      )
      .join("");

    qsa(".edit-category").forEach((btn) =>
      btn.addEventListener("click", () => openCategoryModal(parseInt(btn.dataset.id, 10)))
    );
    qsa(".delete-category").forEach((btn) =>
      btn.addEventListener("click", () => deleteCategory(parseInt(btn.dataset.id, 10)))
    );

    const categorySelect = qs("#p-category");
    categorySelect.innerHTML = '<option value="">None</option>' + allCategories.map((c) => `<option value="${c.id}">${escapeHtml(c.name)}</option>`).join("");
  } catch (e) {
    showAlert(qs("#alert-box"), e.message);
  }
}

function openCategoryModal(id) {
  const form = qs("#category-form");
  form.reset();
  qs("#category-id").value = "";
  qs("#category-modal-title").textContent = "New Category";

  if (id) {
    const category = allCategories.find((c) => c.id === id);
    if (category) {
      qs("#category-id").value = category.id;
      qs("#c-name").value = category.name;
      qs("#c-description").value = category.description || "";
      qs("#category-modal-title").textContent = "Edit Category";
    }
  }
  showModal("category-modal");
}

async function deleteCategory(id) {
  if (!confirm("Delete this category?")) return;
  try {
    await api("/admin/categories/" + id, { method: "DELETE" });
    loadCategories();
  } catch (e) {
    showAlert(qs("#alert-box"), e.message);
  }
}

// ---------- Products ----------
async function loadProductsAdmin() {
  try {
    const result = await api("/admin/products?limit=100");
    const tbody = qs("#products-table tbody");
    tbody.innerHTML = result.items
      .map(
        (p) => `
      <tr>
        <td><img src="${appUrl(p.image_url || "/img/placeholder.svg")}" style="width:40px;height:40px;object-fit:cover;border-radius:4px;" onerror="this.src='${appUrl("/img/placeholder.svg")}'"/></td>
        <td>${escapeHtml(p.name)}</td>
        <td class="muted">${escapeHtml(p.category_name || "-")}</td>
        <td>${formatPrice(p.price)}</td>
        <td>${p.stock}</td>
        <td>${p.is_active ? "Yes" : "No"}</td>
        <td>
          <button class="btn secondary edit-product" data-id="${p.id}">Edit</button>
          <button class="btn danger delete-product" data-id="${p.id}">Delete</button>
        </td>
      </tr>
    `
      )
      .join("");

    tbody.dataset.products = JSON.stringify(result.items);

    qsa(".edit-product").forEach((btn) =>
      btn.addEventListener("click", () => openProductModal(parseInt(btn.dataset.id, 10), result.items))
    );
    qsa(".delete-product").forEach((btn) =>
      btn.addEventListener("click", () => deleteProduct(parseInt(btn.dataset.id, 10)))
    );
  } catch (e) {
    showAlert(qs("#alert-box"), e.message);
  }
}

function openProductModal(id, items) {
  const form = qs("#product-form");
  form.reset();
  qs("#product-id").value = "";
  qs("#product-modal-title").textContent = "New Product";
  qs("#p-active").checked = true;

  if (id) {
    const product = items.find((p) => p.id === id);
    if (product) {
      qs("#product-id").value = product.id;
      qs("#p-name").value = product.name;
      qs("#p-description").value = product.description || "";
      qs("#p-price").value = product.price;
      qs("#p-stock").value = product.stock;
      qs("#p-sku").value = product.sku || "";
      qs("#p-category").value = product.category_id || "";
      qs("#p-active").checked = product.is_active;
      qs("#product-modal-title").textContent = "Edit Product";
    }
  }
  showModal("product-modal");
}

async function deleteProduct(id) {
  if (!confirm("Delete this product? This cannot be undone.")) return;
  try {
    await api("/admin/products/" + id, { method: "DELETE" });
    loadProductsAdmin();
  } catch (e) {
    showAlert(qs("#alert-box"), e.message);
  }
}

async function saveProduct(e) {
  e.preventDefault();
  const alertBox = qs("#alert-box");
  const id = qs("#product-id").value;

  const payload = {
    name: qs("#p-name").value.trim(),
    description: qs("#p-description").value,
    price: parseFloat(qs("#p-price").value) || 0,
    stock: parseInt(qs("#p-stock").value, 10) || 0,
    sku: qs("#p-sku").value.trim(),
    category_id: qs("#p-category").value ? parseInt(qs("#p-category").value, 10) : null,
    image_url: "",
    is_active: qs("#p-active").checked,
  };

  try {
    let product;
    if (id) {
      product = await api("/admin/products/" + id, { method: "PUT", body: payload });
    } else {
      product = await api("/admin/products", { method: "POST", body: payload });
    }

    const imageFile = qs("#p-image").files[0];
    if (imageFile) {
      const formData = new FormData();
      formData.append("image", imageFile);
      await api(`/admin/products/${product.id}/image`, { method: "POST", body: formData });
    }

    hideModals();
    loadProductsAdmin();
    loadStats();
  } catch (err) {
    showAlert(alertBox, err.message);
  }
}

async function saveCategory(e) {
  e.preventDefault();
  const alertBox = qs("#alert-box");
  const id = qs("#category-id").value;
  const payload = {
    name: qs("#c-name").value.trim(),
    description: qs("#c-description").value,
  };

  try {
    if (id) {
      await api("/admin/categories/" + id, { method: "PUT", body: payload });
    } else {
      await api("/admin/categories", { method: "POST", body: payload });
    }
    hideModals();
    loadCategories();
  } catch (err) {
    showAlert(alertBox, err.message);
  }
}

// ---------- Orders ----------
const ORDER_STATUSES = ["pending", "paid", "shipped", "delivered", "cancelled"];

async function loadOrdersAdmin() {
  try {
    const result = await api("/admin/orders?limit=100");
    const tbody = qs("#orders-table tbody");
    tbody.innerHTML = result.items
      .map(
        (o) => `
      <tr>
        <td>#${o.id}</td>
        <td>User ${o.user_id}</td>
        <td>${formatPrice(o.total)}</td>
        <td>
          <select class="status-select" data-id="${o.id}">
            ${ORDER_STATUSES.map((s) => `<option value="${s}" ${s === o.status ? "selected" : ""}>${s}</option>`).join("")}
          </select>
        </td>
        <td class="muted">${new Date(o.created_at).toLocaleDateString()}</td>
        <td><button class="btn secondary save-status" data-id="${o.id}">Update</button></td>
      </tr>
    `
      )
      .join("");

    qsa(".save-status").forEach((btn) => {
      btn.addEventListener("click", async () => {
        const id = btn.dataset.id;
        const select = qs(`.status-select[data-id="${id}"]`);
        try {
          await api(`/admin/orders/${id}/status`, { method: "PUT", body: { status: select.value } });
          showAlert(qs("#alert-box"), `Order #${id} updated to ${select.value}.`, "success");
        } catch (e) {
          showAlert(qs("#alert-box"), e.message);
        }
      });
    });
  } catch (e) {
    showAlert(qs("#alert-box"), e.message);
  }
}

// ---------- Users ----------
async function loadUsersAdmin() {
  try {
    const result = await api("/admin/users?limit=100");
    const tbody = qs("#users-table tbody");
    tbody.innerHTML = result.items
      .map(
        (u) => `
      <tr>
        <td>${u.id}</td>
        <td>${escapeHtml(u.name)}</td>
        <td>${escapeHtml(u.email)}</td>
        <td><span class="badge">${u.role}</span></td>
        <td class="muted">${new Date(u.created_at).toLocaleDateString()}</td>
      </tr>
    `
      )
      .join("");
  } catch (e) {
    showAlert(qs("#alert-box"), e.message);
  }
}

// ---------- Tabs ----------
function setupTabs() {
  qsa(".tab-btn").forEach((btn) => {
    btn.addEventListener("click", () => {
      qsa(".tab-btn").forEach((b) => b.classList.remove("active"));
      btn.classList.add("active");
      qsa(".tab-panel").forEach((panel) => (panel.style.display = "none"));
      qs("#tab-" + btn.dataset.tab).style.display = "block";

      if (btn.dataset.tab === "orders") loadOrdersAdmin();
      if (btn.dataset.tab === "users") loadUsersAdmin();
    });
  });
}

document.addEventListener("DOMContentLoaded", () => {
  if (!guardAdmin()) return;

  setupTabs();
  loadStats();
  loadCategories().then(loadProductsAdmin);

  qs("#new-product-btn").addEventListener("click", () => openProductModal(null, []));
  qs("#new-category-btn").addEventListener("click", () => openCategoryModal(null));
  qs("#product-modal-cancel").addEventListener("click", hideModals);
  qs("#category-modal-cancel").addEventListener("click", hideModals);
  qs("#modal-overlay").addEventListener("click", hideModals);
  qs("#product-form").addEventListener("submit", saveProduct);
  qs("#category-form").addEventListener("submit", saveCategory);
});
