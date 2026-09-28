let currentPage = 1;

async function loadCategories() {
  try {
    const categories = await api("/categories");
    const select = qs("#category-select");
    (categories || []).forEach((c) => {
      const opt = document.createElement("option");
      opt.value = c.slug;
      opt.textContent = c.name;
      select.appendChild(opt);
    });
  } catch (e) {
    // Non-fatal: category filter just stays empty.
  }
}

function buildQuery() {
  const params = new URLSearchParams();
  const search = qs("#search-input").value.trim();
  const category = qs("#category-select").value;
  const minPrice = qs("#min-price").value;
  const maxPrice = qs("#max-price").value;
  const sort = qs("#sort-select").value;

  if (search) params.set("search", search);
  if (category) params.set("category", category);
  if (minPrice) params.set("min_price", minPrice);
  if (maxPrice) params.set("max_price", maxPrice);
  if (sort) params.set("sort", sort);
  params.set("page", currentPage);
  params.set("limit", 12);

  return params.toString();
}

function renderProducts(result) {
  const grid = qs("#product-grid");
  const alertBox = qs("#alert-box");
  alertBox.innerHTML = "";

  if (!result.items || result.items.length === 0) {
    grid.innerHTML = `<div class="empty-state">No products match your filters.</div>`;
    qs("#pagination").innerHTML = "";
    return;
  }

  grid.innerHTML = result.items
    .map(
      (p) => `
      <a class="card product-card" href="${appUrl("product.html?slug=" + encodeURIComponent(p.slug))}">
        <img class="thumb" src="${appUrl(p.image_url || "/img/placeholder.svg")}" alt="${escapeHtml(p.name)}" onerror="this.src='${appUrl("/img/placeholder.svg")}'" />
        <h3>${escapeHtml(p.name)}</h3>
        <div class="muted">${escapeHtml(p.category_name || "")}</div>
        <div class="price">${formatPrice(p.price)}</div>
        ${p.stock === 0 ? '<div class="muted">Out of stock</div>' : ""}
      </a>
    `
    )
    .join("");

  renderPagination(result);
}

function renderPagination(result) {
  const pagination = qs("#pagination");
  if (result.total_pages <= 1) {
    pagination.innerHTML = "";
    return;
  }

  let html = "";
  for (let i = 1; i <= result.total_pages; i++) {
    html += `<button class="${i === result.page ? "active" : ""}" data-page="${i}">${i}</button>`;
  }
  pagination.innerHTML = html;

  qsa("#pagination button").forEach((btn) => {
    btn.addEventListener("click", () => {
      currentPage = parseInt(btn.dataset.page, 10);
      loadProducts();
    });
  });
}

async function loadProducts() {
  try {
    const query = buildQuery();
    const result = await api("/products?" + query);
    renderProducts(result);
  } catch (e) {
    showAlert(qs("#alert-box"), e.message);
  }
}

function debounce(fn, delay) {
  let timer;
  return (...args) => {
    clearTimeout(timer);
    timer = setTimeout(() => fn(...args), delay);
  };
}

document.addEventListener("DOMContentLoaded", () => {
  loadCategories();
  loadProducts();

  const refresh = debounce(() => {
    currentPage = 1;
    loadProducts();
  }, 350);

  ["search-input", "min-price", "max-price"].forEach((id) => {
    qs("#" + id).addEventListener("input", refresh);
  });
  ["category-select", "sort-select"].forEach((id) => {
    qs("#" + id).addEventListener("change", () => {
      currentPage = 1;
      loadProducts();
    });
  });

  qs("#clear-filters").addEventListener("click", () => {
    qs("#search-input").value = "";
    qs("#category-select").value = "";
    qs("#min-price").value = "";
    qs("#max-price").value = "";
    qs("#sort-select").value = "newest";
    currentPage = 1;
    loadProducts();
  });
});
