/* weread-kit Web UI：账号管理、扫码登录、阅读挑战、日志与推送设置。
   无框架，原生 fetch；图标来自 lucide.js 提供的 SVG 路径数据。 */

const ICONS = {"search": "<path d=\"m21 21-4.34-4.34\" /> <circle cx=\"11\" cy=\"11\" r=\"8\" />", "gift": "<rect x=\"3\" y=\"8\" width=\"18\" height=\"4\" rx=\"1\" /><path d=\"M12 8v13\" /><path d=\"M19 12v7a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2v-7\" /><path d=\"M7.5 8a2.5 2.5 0 0 1 0-5A4.8 8 0 0 1 12 8a4.8 8 0 0 1 4.5-5 2.5 2.5 0 0 1 0 5\" />", "chevron-down": "<path d=\"m6 9 6 6 6-6\" />", "x": "<path d=\"M18 6 6 18\" /> <path d=\"m6 6 12 12\" />", "refresh-cw": "<path d=\"M3 12a9 9 0 0 1 9-9 9.75 9.75 0 0 1 6.74 2.74L21 8\" /> <path d=\"M21 3v5h-5\" /> <path d=\"M21 12a9 9 0 0 1-9 9 9.75 9.75 0 0 1-6.74-2.74L3 16\" /> <path d=\"M8 16H3v5\" />", "book-open": "<path d=\"M12 5v16\" /> <path d=\"M20.001 19A2 2 0 0022 17V5a2 2 0 00-1.999-2L16 3.002A5 5 0 0012 5a5 5 0 00-4-2H4a2 2 0 00-2 2v12a2 2 0 001.999 2H8a5 5 0 014 2 5 5 0 014-2z\" />", "circle-check": "<circle cx=\"12\" cy=\"12\" r=\"10\" /> <path d=\"m16 9-5.5 5.5L8 12\" />", "settings-2": "<path d=\"M14 17H5\" /> <path d=\"M19 7h-9\" /> <circle cx=\"17\" cy=\"17\" r=\"3\" /> <circle cx=\"7\" cy=\"7\" r=\"3\" />", "plus": "<path d=\"M5 12h14\" /> <path d=\"M12 5v14\" />", "check": "<path d=\"M20 6 9 17l-5-5\" />", "save": "<path d=\"M15.2 3a2 2 0 0 1 1.4.6l3.8 3.8a2 2 0 0 1 .6 1.4V19a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2z\" /> <path d=\"M17 21v-7a1 1 0 0 0-1-1H8a1 1 0 0 0-1 1v7\" /> <path d=\"M7 3v4a1 1 0 0 0 1 1h7\" />", "square": "<rect width=\"18\" height=\"18\" x=\"3\" y=\"3\" rx=\"2\" />", "play": "<path d=\"M5 5a2 2 0 0 1 3.008-1.728l11.997 6.998a2 2 0 0 1 .003 3.458l-12 7A2 2 0 0 1 5 19z\" />", "send": "<path d=\"M14.536 21.686a.5.5 0 0 0 .937-.024l6.5-19a.496.496 0 0 0-.635-.635l-19 6.5a.5.5 0 0 0-.024.937l7.93 3.18a2 2 0 0 1 1.112 1.11z\" /> <path d=\"m21.854 2.147-10.94 10.939\" />", "trash": "<path d=\"M10 11v6\" /> <path d=\"M14 11v6\" /> <path d=\"M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6\" /> <path d=\"M3 6h18\" /> <path d=\"M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2\" />", "sun": "<circle cx=\"12\" cy=\"12\" r=\"4\" /> <path d=\"M12 2v2\" /> <path d=\"M12 20v2\" /> <path d=\"m4.93 4.93 1.41 1.41\" /> <path d=\"m17.66 17.66 1.41 1.41\" /> <path d=\"M2 12h2\" /> <path d=\"M20 12h2\" /> <path d=\"m6.34 17.66-1.41 1.41\" /> <path d=\"m19.07 4.93-1.41 1.41\" />", "scroll-text": "<path d=\"M15 12h-5\" /> <path d=\"M15 8h-5\" /> <path d=\"M19 17V5a2 2 0 0 0-2-2H4\" /> <path d=\"M8 21h12a2 2 0 0 0 2-2v-1a1 1 0 0 0-1-1H11a1 1 0 0 0-1 1v1a2 2 0 1 1-4 0V5a2 2 0 1 0-4 0v2a1 1 0 0 0 1 1h3\" />", "clock": "<circle cx=\"12\" cy=\"12\" r=\"10\" /> <path d=\"M12 6v6l4 2\" />", "user":'<path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2" /><circle cx="12" cy="7" r="4" />',
  "user": '<path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2" /><circle cx="12" cy="7" r="4" />',
  "users-round": "<path d=\"M18 21a8 8 0 0 0-16 0\" /> <circle cx=\"10\" cy=\"8\" r=\"5\" /> <path d=\"M22 20c0-3.37-2-6.5-4-8a5 5 0 0 0-.45-8.3\" />", "pause": "<rect x=\"14\" y=\"3\" width=\"5\" height=\"18\" rx=\"1\" /> <rect x=\"5\" y=\"3\" width=\"5\" height=\"18\" rx=\"1\" />", "monitor": "<rect width=\"20\" height=\"14\" x=\"2\" y=\"3\" rx=\"2\" /> <line x1=\"8\" x2=\"16\" y1=\"21\" y2=\"21\" /> <line x1=\"12\" x2=\"12\" y1=\"17\" y2=\"21\" />", "minus": "<path d=\"M5 12h14\" />", "moon": "<path d=\"M20.985 12.486a9 9 0 1 1-9.473-9.472c.405-.022.617.46.402.803a6 6 0 0 0 8.268 8.268c.344-.215.825-.004.803.401\" />", "menu": "<path d=\"M4 6h16\" /> <path d=\"M4 12h16\" /> <path d=\"M4 18h16\" />", "alert": "<path d=\"m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 20h16a2 2 0 0 0 1.73-2\" /> <path d=\"M12 9v4\" /> <path d=\"M12 17h.01\" />", "info": "<circle cx=\"12\" cy=\"12\" r=\"10\" /> <path d=\"M12 16v-4\" /> <path d=\"M12 8h.01\" />"};

const $ = (sel) => document.querySelector(sel);
const $$ = (sel) => document.querySelectorAll(sel);

async function checkAppVersion() {
  const el = $("#app-version");
  const pop = $("#version-popover");
  const status = $("#version-status");
  if (!el || !pop) return;

  if (!el.dataset.bound) {
    el.dataset.bound = "1";
    el.addEventListener("click", (e) => {
      e.preventDefault();
      e.stopPropagation();
      const open = pop.classList.toggle("hidden") === false;
      el.setAttribute("aria-expanded", String(open));
    });
    document.addEventListener("click", (e) => {
      if (!el.contains(e.target) && !pop.contains(e.target)) {
        pop.classList.add("hidden");
        el.setAttribute("aria-expanded", "false");
      }
    });
  }

  const render = (v) => {
    // 版本号以 Go 服务端 /api/version 返回值为准，前端不再维护副本。
    const current = `v${v.current_version || "未知"}`;
    el.textContent = current;
    el.title = v.has_update ? "发现新版本，点击查看" : "查看版本信息";
    el.classList.toggle("has-update", !!v.has_update);
    status?.classList.toggle("hidden", !v.has_update);
    if (status) status.textContent = v.has_update ? "有更新" : "";

    if (v.has_update) {
      const latest = escapeHtml(v.latest_version || "新版本");
      pop.innerHTML = `<strong>发现新版本 ${latest}</strong><p>当前版本 ${current}。建议更新以获得最新功能和修复。</p>${v.html_url ? `<a href="${escapeHtml(v.html_url)}" target="_blank" rel="noreferrer">查看发布说明 <span aria-hidden="true">→</span></a>` : ""}`;
    } else if (v.check_failed) {
      pop.innerHTML = `<strong>暂时无法检查更新</strong><p>网络不可用或更新服务暂时没有响应。</p><button type="button" class="version-retry" id="version-retry">重新检查</button>`;
      pop.querySelector("#version-retry")?.addEventListener("click", () => checkAppVersion());
    } else {
      pop.innerHTML = `<strong>已是最新版本</strong><p>当前版本 ${current}，暂时不需要更新。</p>`;
    }
  };

  try {
    const v = await fetch("/api/version").then((r) => r.json());
    render(v);
  } catch (_) {
    render({ current_version: el.textContent.replace(/^v/, ""), check_failed: true });
  }
}
setTimeout(checkAppVersion, 300);

/* ---------- 通用工具 ---------- */

function iconSvg(name) {
  const paths = ICONS[name];
  if (!paths) return "";
  return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">${paths}</svg>`;
}

// 把静态 HTML 里 data-icon 占位替换为 SVG 图标。
function renderIcons(root = document) {
  root.querySelectorAll("[data-icon]").forEach((el) => {
    el.innerHTML = iconSvg(el.dataset.icon);
  });
}

function iconEl(name) {
  const span = document.createElement("span");
  span.className = "nav-icon";
  span.innerHTML = iconSvg(name);
  return span;
}

async function api(path, opts = {}) {
  const resp = await fetch(path, { headers: { "content-type": "application/json" }, ...opts });
  const data = await resp.json().catch(() => ({}));
  if (resp.status === 401) { window.location.assign("/login"); throw new Error("登录已过期，请重新登录"); }
  if (!resp.ok) throw new Error(data.error || `HTTP ${resp.status}`);
  return data;
}

function fmtTime(unix) {
  if (!unix) return "—";
  return new Date(unix * 1000).toLocaleString("zh-CN", { hour12: false });
}
function fmtDay(unix) {
  if (!unix) return "—";
  return new Date(unix * 1000).toLocaleDateString("zh-CN");
}
function fmtRemain(seconds) {
  if (!seconds || seconds <= 0) return "—";
  const days = Math.floor(seconds / 86400);
  return days >= 1 ? `约 ${days} 天` : "不足 1 天";
}
function escapeHtml(s) {
  return String(s).replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]));
}

// 异步操作期间禁用按钮并显示加载指示，避免重复提交时界面毫无反馈。
function setBusy(btn, busy) {
  if (!btn) return;
  btn.disabled = busy;
  btn.classList.toggle("loading", busy);
}

// 加载占位骨架：屏幕阅读器仍读到"正在加载"，视觉上是脉冲色块，等待期间布局不跳动。
function skeletonHTML(lines = 3) {
  const rows = Array.from({ length: lines }, () => '<div class="skeleton-line" aria-hidden="true"></div>').join("");
  return `<div class="skeleton-lines"><span class="sr-only" role="status">正在加载…</span>${rows}</div>`;
}

// 卡片形状的骨架：高度贴近真实卡片，加载完成时下方内容不发生跳动。
function skeletonCardsHTML(count, height, cover = false) {
  const cards = Array.from({ length: count }, () => `<div class="skeleton-card" style="height:${height}px" aria-hidden="true"></div>`).join("");
  return `<div class="skeleton-cards${cover ? " covers" : ""}"><span class="sr-only" role="status">正在加载…</span>${cards}</div>`;
}

// 自定义确认框：返回 Promise<boolean>，替代原生 confirm，与界面风格保持一致。
function confirmAction({ title = "确认操作", message = "", confirmText = "确定", danger = true } = {}) {
  return new Promise((resolve) => {
    const dlg = $("#confirm-dialog");
    $("#confirm-title").textContent = title;
    $("#confirm-message").textContent = message;
    const ok = $("#confirm-ok");
    const cancel = $("#confirm-cancel");
    ok.textContent = confirmText;
    ok.className = danger ? "btn danger" : "btn primary";
    const finish = (result) => {
      ok.removeEventListener("click", onOk);
      cancel.removeEventListener("click", onCancel);
      dlg.removeEventListener("cancel", onEsc);
      if (dlg.open) dlg.close();
      resolve(result);
    };
    const onOk = () => finish(true);
    const onCancel = () => finish(false);
    const onEsc = (e) => { e.preventDefault(); finish(false); };
    ok.addEventListener("click", onOk);
    cancel.addEventListener("click", onCancel);
    dlg.addEventListener("cancel", onEsc);
    dlg.showModal();
  });
}

/* ---------- 视图切换 ---------- */

let currentView = "accounts";

function setView(view) {
  currentView = view;
  for (const name of ["accounts", "reading", "challenge", "logs", "settings"]) {
    $("#view-" + name).classList.toggle("hidden", view !== name);
  }
  $$(".nav-item").forEach((el) => {
    const active = el.dataset.view === view;
    el.classList.toggle("active", active);
    if (active) el.setAttribute("aria-current", "page");
    else el.removeAttribute("aria-current");
  });
}

/* ---------- 窄屏导航按钮 ---------- */

const sidebar = $(".sidebar");
function closeMobileNav() {
  sidebar.classList.remove("nav-open");
  $("#nav-toggle").setAttribute("aria-expanded", "false");
}
$("#nav-toggle").addEventListener("click", () => {
  const open = sidebar.classList.toggle("nav-open");
  $("#nav-toggle").setAttribute("aria-expanded", String(open));
});

/* ---------- 自绘下拉 ---------- */

// 用自绘弹层替换原生 select:select 本体隐藏但仍保存值、派发 change,
// 选项被外部重建时自动刷新弹层与按钮文案。
function enhanceSelect(sel) {
  if (sel.dataset.dd) return;
  sel.dataset.dd = "1";
  sel.setAttribute("aria-hidden", "true");
  sel.tabIndex = -1;
  const dd = document.createElement("div");
  dd.className = "dd";
  sel.before(dd);
  dd.append(sel);
  const toggle = document.createElement("button");
  toggle.type = "button";
  toggle.className = "dd-toggle";
  toggle.setAttribute("aria-haspopup", "listbox");
  toggle.setAttribute("aria-expanded", "false");
  toggle.setAttribute("aria-controls", `${sel.id}-menu`);
  toggle.innerHTML = '<span class="dd-label"></span><span class="nav-icon" data-icon="chevron-down"></span>';
  const menu = document.createElement("div");
  menu.className = "dd-menu";
  menu.id = `${sel.id}-menu`;
  menu.setAttribute("role", "listbox");
  dd.append(toggle, menu);
  renderIcons(toggle);

  const currentLabel = () => sel.options[sel.selectedIndex]?.textContent ?? "";
  function sync() {
    toggle.querySelector(".dd-label").textContent = currentLabel();
    [...menu.children].forEach((item, i) => {
      const active = !!sel.options[i]?.selected;
      item.classList.toggle("active", active);
      item.setAttribute("aria-selected", String(active));
    });
  }
  function renderMenu() {
    menu.textContent = "";
    for (const opt of sel.options) {
      const item = document.createElement("button");
      item.type = "button";
      item.className = "dd-item";
      item.setAttribute("role", "option");
      item.id = `${sel.id}-option-${opt.value || opt.index}`;
      item.tabIndex = -1;
      item.textContent = opt.textContent;
      item.addEventListener("click", () => {
        sel.value = opt.value;
        sync();
        close();
        toggle.focus();
        sel.dispatchEvent(new Event("change"));
      });
      item.addEventListener("keydown", (e) => {
        const items = [...menu.children];
        const index = items.indexOf(item);
        if (e.key === "ArrowDown" || e.key === "ArrowUp") {
          e.preventDefault();
          items[(index + (e.key === "ArrowDown" ? 1 : -1) + items.length) % items.length]?.focus();
        } else if (e.key === "Enter" || e.key === " ") {
          e.preventDefault();
          item.click();
        } else if (e.key === "Escape") {
          e.preventDefault();
          close();
          toggle.focus();
        }
      });
      menu.append(item);
    }
    sync();
  }
  function open() {
    renderMenu();
    dd.classList.add("open");
    toggle.setAttribute("aria-expanded", "true");
    menu.querySelector(".active")?.scrollIntoView({ block: "nearest" });
    menu.querySelector(".active")?.focus();
  }
  function close() {
    dd.classList.remove("open");
    toggle.setAttribute("aria-expanded", "false");
  }
  toggle.addEventListener("click", () => (dd.classList.contains("open") ? close() : open()));
  toggle.addEventListener("keydown", (e) => {
    if (["Enter", " ", "ArrowDown"].includes(e.key)) { e.preventDefault(); open(); }
  });
  // 外部代码程序化设置 sel.value 后派发 change，即可同步按钮文案。
  sel.addEventListener("change", sync);
  document.addEventListener("click", (e) => {
    if (!dd.contains(e.target)) close();
  });
  document.addEventListener("keydown", (e) => {
    if (e.key === "Escape") { close(); toggle.focus(); }
  });
  new MutationObserver(() => {
    sync();
    if (dd.classList.contains("open")) renderMenu();
  }).observe(sel, { childList: true });
  renderMenu();
}

// 开始时间下拉：30 分钟一档
const runatSel = $("#challenge-runat-input");
for (let h = 0; h < 24; h++) {
  for (const m of [0, 30]) {
    const opt = document.createElement("option");
    opt.value = opt.textContent = `${String(h).padStart(2, "0")}:${String(m).padStart(2, "0")}`;
    runatSel.append(opt);
  }
}
runatSel.value = "03:00";
document.querySelectorAll("select.select").forEach(enhanceSelect);

$$(".nav-item").forEach((el) => {
  el.addEventListener("click", (e) => {
    e.preventDefault();
    closeMobileNav();
    const view = el.dataset.view;
    setView(view);
    if (view === "reading") openReading();
    if (view === "challenge") openChallenge();
    if (view === "logs") openLogs();
    if (view === "settings") loadPushChannels();
  });
});

/* ---------- 主题切换 ---------- */

$("#theme-toggle").addEventListener("click", () => {
  const dark = document.documentElement.dataset.theme === "dark";
  const next = dark ? "light" : "dark";
  document.documentElement.dataset.theme = next;
  localStorage.setItem("weread-theme", next);
  syncThemeUI();
});

function syncThemeUI() {
  const dark = document.documentElement.dataset.theme === "dark";
  $("#theme-label").textContent = dark ? "浅色" : "深色";
  $("#theme-toggle [data-icon]").dataset.icon = dark ? "sun" : "moon";
  renderIcons($("#theme-toggle"));
}
syncThemeUI();

/* ---------- Toast ---------- */

let toastTimer = null;
function toast(msg, type = "info") {
  const el = $("#toast");
  el.textContent = "";
  const icon = document.createElement("span");
  icon.className = "nav-icon";
  icon.dataset.icon = type === "error" ? "alert" : type === "ok" ? "circle-check" : "info";
  const text = document.createElement("span");
  text.textContent = msg;
  el.append(icon, text);
  renderIcons(el);
  el.classList.remove("hidden", "ok", "error");
  el.classList.add(type);
  // popover 提到最顶层，原生弹窗打开时也能看到报错
  try { el.showPopover(); } catch {}
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => {
    el.classList.add("hidden");
    try { el.hidePopover(); } catch {}
  }, type === "error" ? 4200 : 3200);
}

/* ---------- 账号列表 ---------- */

let allAccounts = [];
let listRequest = 0;

async function loadAccounts() {
  const request = ++listRequest;
  setBusy($("#reload-btn"), true);
  try {
    allAccounts = await api("/api/accounts");
    if (request !== listRequest) return;
    renderAccounts();
  } catch (err) {
    if (request !== listRequest) return;
    showEmptyState(`无法加载账号：${err.message}`, "点击「刷新」重试");
    $("#list-summary").textContent = "加载失败";
  } finally {
    if (request === listRequest) setBusy($("#reload-btn"), false);
  }
}

function displayName(a) {
  return a.remark || a.name || a.vid;
}

function renderAccounts() {
  const query = $("#account-search").value.trim().toLocaleLowerCase();
  const accounts = allAccounts.filter((a) =>
    `${a.remark || ""} ${a.name || ""} ${a.vid}`.toLocaleLowerCase().includes(query)
  );
  $("#account-count").textContent = accounts.length;
  $("#account-count-label").textContent = query ? "搜索结果" : "全部账号";
  // 无账号时的空态由 #empty 承载：未添加任何账号，或搜索无匹配
  $("#empty").classList.toggle("hidden", accounts.length > 0);
  if (!accounts.length && query) {
    $("#empty-title").textContent = `没有找到「${query}」`;
    $("#empty-desc").textContent = "换个关键词试试";
  } else {
    $("#empty-title").textContent = "从连接第一个账号开始";
    $("#empty-desc").textContent = "用微信扫一扫，即可把阅读账号添加到这里。";
  }

  const list = $("#account-list");
  list.textContent = "";
  for (const a of accounts) {
    const row = document.createElement("div");
    row.className = "account-row";

    // 头像：优先微信头像，否则取显示名首字
    const avatar = document.createElement("div");
    avatar.className = "avatar";
    if (a.avatar) {
      const img = document.createElement("img");
      img.src = a.avatar;
      img.alt = "";
      img.loading = "lazy";
      avatar.append(img);
    } else {
      avatar.textContent = Array.from(displayName(a))[0] || "读";
    }

    // 信息列：可编辑昵称/备注 + 别名与 ID
    const info = document.createElement("div");
    info.className = "account-info";
    const name = document.createElement("input");
    name.className = "account-name";
    name.value = a.remark || a.name || a.vid;
    name.placeholder = "点击设置备注";
    name.setAttribute("aria-label", `账号 ${displayName(a)} 的备注`);
    name.maxLength = 100;
    name.title = "点击修改备注";
    name.addEventListener("keydown", (e) => {
      if (e.key === "Enter") name.blur();
      if (e.key === "Escape") {
        name.value = a.remark || a.name || a.vid;
        name.blur();
      }
    });
    name.addEventListener("change", async () => {
      try {
        await api(`/api/accounts/${encodeURIComponent(a.vid)}/remark`, {
          method: "PUT",
          body: JSON.stringify({ remark: name.value.trim() }),
        });
        a.remark = name.value.trim();
        toast("备注已保存", "ok");
      } catch (err) {
        name.value = a.remark || "";
        toast(`保存失败：${err.message}`, "error");
      }
    });
    const sub = document.createElement("div");
    sub.className = "account-sub";
    // 网页扫码的账号别名即 vid，此时只显示 ID；别名不同（CLI 创建）才额外标注。
    sub.textContent = `ID ${a.vid}`;
    sub.title = `本地别名 ${a.vid}(CLI --alias 用）`;
    info.append(name, sub);
    row.append(avatar, info);

    // 凭据状态 + 上次刷新时间
    const cred = document.createElement("div");
    cred.className = "account-meta";
    const badge = document.createElement("span");
    badge.className = "badge " + (a.has_access_token ? "ok" : "missing");
    badge.textContent = a.has_access_token ? "凭据有效" : "待刷新";
    badge.title = "表示本地是否保存凭据，不代表实时登录状态";
    const rotated = document.createElement("span");
    rotated.className = "account-time";
    rotated.textContent = fmtTime(a.rotated_at);
    cred.append(badge, rotated);

    // 操作：详情 / 刷新 / 删除
    const ops = document.createElement("div");
    ops.className = "ops";
    const detailBtn = button("详情", "btn", () => showDetails(a));
    detailBtn.setAttribute("aria-label", `查看账号 ${displayName(a)} 详情`);
    const refreshBtn = button("刷新", "btn", async () => {
      setBusy(refreshBtn, true);
      try {
        await api(`/api/accounts/${encodeURIComponent(a.vid)}/refresh`, { method: "POST" });
        toast(`已刷新 ${displayName(a)}`, "ok");
        await loadAccounts();
      } catch (err) {
        toast(`刷新失败：${err.message}`, "error");
        setBusy(refreshBtn, false);
      }
    });
    refreshBtn.setAttribute("aria-label", `刷新账号 ${displayName(a)}`);
    const delBtn = button("删除", "btn danger", async () => {
      const okDelete = await confirmAction({
        title: "删除账号",
        message: `确定删除账号「${displayName(a)}」？仅删除本地凭据，不影响微信读书账号。`,
        confirmText: "删除",
      });
      if (!okDelete) return;
      try {
        await api(`/api/accounts/${encodeURIComponent(a.vid)}`, { method: "DELETE" });
        toast("已删除", "ok");
        await loadAccounts();
      } catch (err) {
        toast(`删除失败：${err.message}`, "error");
      }
    });
    delBtn.setAttribute("aria-label", `删除账号 ${displayName(a)}`);
    ops.append(detailBtn, refreshBtn, delBtn);

    row.append(cred, ops);
    // 让操作区最右侧：info 占满剩余宽度
    info.style.flex = "1";
    list.append(row);
  }
}

function button(text, cls, onClick) {
  const b = document.createElement("button");
  b.textContent = text;
  b.className = cls;
  b.addEventListener("click", onClick);
  return b;
}

$("#account-search").addEventListener("input", renderAccounts);
$("#reload-btn").addEventListener("click", loadAccounts);

/* ---------- 扫码登录 ---------- */

let loginId = null;
let loginPollTimer = null;
let loginGeneration = 0;

function stopLoginPolling() {
  if (loginPollTimer) {
    clearInterval(loginPollTimer);
    loginPollTimer = null;
  }
}

const loginStatusText = {
  pending: "",
  scanned: "已扫码，请在微信中确认",
  success: "登录成功！",
  expired: "二维码已过期，请重新生成",
  declined: "你在微信中拒绝了授权",
  canceled: "登录已取消",
  error: "登录失败",
};

function setLoginStatus(status, errMsg) {
  const el = $("#login-status");
  // 注意用 in 判断：某些状态的文案就是空串，不能用真值回退，否则状态码会漏到界面。
  const label = status in loginStatusText ? loginStatusText[status] : status;
  // 错误详情与状态文案相同时只展示一次，避免重复。
  const duplicated = !errMsg || errMsg === label || label.includes(errMsg) || errMsg.includes(label);
  el.textContent = duplicated ? label : `${label}:${errMsg}`;
  el.className = "status" + (status === "success" ? " ok" : ["expired", "declined", "error", "canceled"].includes(status) ? " err" : "");
}

function resetLoginUI() {
  stopLoginPolling();
  loginId = null;
  $("#qr-area").classList.add("hidden");
  $("#login-status").textContent = "";
  $("#login-status").className = "status";
  $("#login-start").disabled = false;
}

function setStartBtnText(text) {
  $("#login-start-text").textContent = text;
}

async function startLogin() {
  const generation = ++loginGeneration;
  stopLoginPolling();
  $("#login-start").disabled = true;
  setStartBtnText("正在生成二维码…");
  try {
    const data = await api("/api/login", { method: "POST", body: "{}" });
    if (generation !== loginGeneration) return;
    loginId = data.id;
    $("#qr-area").classList.remove("hidden");
    $("#qr-img").src = `${data.qr_url}?t=${Date.now()}`;
    // 二维码已就绪，按钮常驻底部，随时可刷新换新码
    $("#login-start").disabled = false;
    setStartBtnText("刷新二维码");
    setLoginStatus("pending");
    loginPollTimer = setInterval(() => pollLogin(generation), 1500);
  } catch (err) {
    if (generation !== loginGeneration) return;
    toast(`发起登录失败：${err.message}`, "error");
    $("#login-start").disabled = false;
    setStartBtnText("重新生成二维码");
  }
}

async function pollLogin(generation) {
  if (!loginId) return;
  let data;
  try {
    data = await api(`/api/login/${loginId}`);
  } catch {
    return; // 单次轮询失败忽略，下一轮再试
  }
  if (generation !== loginGeneration) return;
  setLoginStatus(data.status, data.error);

  if (data.status === "success") {
    stopLoginPolling();
    toast(`账号「${data.account.vid}」登录成功`, "ok");
    setTimeout(() => {
      $("#login-dialog").close();
      loadAccounts();
    }, 900);
  } else if (["expired", "declined", "canceled", "error"].includes(data.status)) {
    stopLoginPolling();
    $("#login-start").classList.remove("hidden");
    $("#login-start").disabled = false;
    setStartBtnText("重新生成二维码");
  }
}

function openLogin() {
  resetLoginUI();
  $("#login-dialog").showModal();
  startLogin(); // 打开弹窗即自动生成二维码
}

$("#add-btn").addEventListener("click", openLogin);
$("#login-start").addEventListener("click", startLogin);
$("#login-close").addEventListener("click", () => $("#login-dialog").close());
$("#login-dialog").addEventListener("close", () => {
  resetLoginUI();
  loadAccounts();
});

/* ---------- 账号详情 ---------- */

let detailAccount = null;
let detailRequest = 0;

async function showDetails(account, force = false) {
  if (!account) return;
  const generation = ++detailRequest;
  detailAccount = account;
  $("#detail-title").textContent = displayName(account);
  $("#detail-content").innerHTML = force
    ? '<p class="empty">正在从微信读书更新…</p>'
    : skeletonHTML(4);
  if (!$("#detail-dialog").open) $("#detail-dialog").showModal();
  setBusy($("#detail-reload"), true);
  try {
    const data = await api(`/api/accounts/${encodeURIComponent(account.vid)}/details`, {
      method: force ? "POST" : "GET",
    });
    if (generation !== detailRequest) return;
    renderDetails(data);
  } catch (err) {
    if (generation !== detailRequest) return;
    $("#detail-content").innerHTML = `<p class="empty">加载失败：${escapeHtml(err.message)}</p>`;
  } finally {
    if (generation === detailRequest) setBusy($("#detail-reload"), false);
  }
}

const detailSectionNames = { shelf: "书架", user: "用户信息", card: "会员卡", session: "网页会话", profile: "资料缓存" };

function renderDetails(d) {
  const root = $("#detail-content");
  root.textContent = "";

  const failed = Object.entries(d.errors || {});
  if (failed.length) {
    const warn = document.createElement("p");
    warn.className = "detail-warn";
    warn.textContent = `部分数据未能获取（${failed.map(([k]) => detailSectionNames[k] || k).join("、")})，可点击「刷新数据」重试。`;
    root.append(warn);
  }

  const user = nested(d.user, "user");
  if (user) root.append(detailSection("用户信息", userCard(user)));
  if (d.card) root.append(detailSection("会员卡", memberCardBlock(d.card)));

  const books = Array.isArray(d.books) ? d.books : [];
  root.append(detailSection(`书架（${books.length} 本）`, shelfGrid(books, d.reading)));
}

function detailSection(title, el) {
  const sec = document.createElement("section");
  sec.className = "detail-section";
  const h = document.createElement("h3");
  h.textContent = title;
  sec.append(h, el);
  return sec;
}

// 微信读书接口返回有的把对象包在子字段里，有的直接平铺，这里做兼容。
function nested(obj, ...keys) {
  if (!obj || typeof obj !== "object") return null;
  for (const k of keys) {
    if (obj[k] && typeof obj[k] === "object" && !Array.isArray(obj[k])) return obj[k];
  }
  return obj;
}

function kvListEl(entries) {
  const wrap = document.createElement("div");
  wrap.className = "kv-list";
  for (const [k, v, cls] of entries) {
    const row = document.createElement("div");
    row.className = "kv-row";
    const key = document.createElement("span");
    key.className = "kv-key";
    key.textContent = k;
    const val = document.createElement("span");
    val.className = "kv-val" + (cls ? " " + cls : "");
    val.textContent = v;
    row.append(key, val);
    wrap.append(row);
  }
  return wrap;
}

function userCard(u) {
  const wrap = document.createElement("div");
  wrap.className = "user-head";
  if (u.avatar) {
    const img = document.createElement("img");
    img.className = "user-avatar";
    img.src = u.avatar;
    img.alt = "头像";
    wrap.append(img);
  }
  const info = document.createElement("div");
  const name = document.createElement("div");
  name.className = "user-name";
  name.textContent = (u.name || u.nickname || "微信读书用户").trim();
  const vid = document.createElement("div");
  vid.className = "user-vid";
  vid.textContent = `用户 ID:${u.userVid || u.vid || "—"}`;
  info.append(name, vid);
  wrap.append(info);
  return wrap;
}

// 会员卡只展示对用户有意义的四项，其余接口字段一律不展示。
function memberCardBlock(c) {
  return kvListEl([
    ["起始日期", fmtDay(c.startTime)],
    ["到期时间", fmtDay(c.expiredTime)],
    ["当前状态", c.expired ? "已过期" : "有效中", c.expired ? "off" : "on"],
    ["剩余时长", c.expired ? "—" : fmtRemain(c.remainTime)],
  ]);
}

function shelfGrid(books) {
  const wrap = document.createElement("div");
  wrap.className = "shelf-grid";
  if (!books.length) {
    wrap.innerHTML = '<p class="empty">书架暂无书籍</p>';
    return wrap;
  }
  for (const b of books) {
    const item = document.createElement("div");
    item.className = "shelf-item";
    if (b.cover) {
      const img = document.createElement("img");
      img.loading = "lazy";
      img.src = b.cover;
      img.alt = "";
      item.append(img);
    }
    const title = document.createElement("div");
    title.className = "shelf-title";
    title.textContent = b.title || b.bookId;
    title.title = b.title || b.bookId || "";
    const author = document.createElement("div");
    author.className = "shelf-author";
    author.textContent = b.author || "";
    item.append(title, author);
    wrap.append(item);
  }
  return wrap;
}

$("#detail-close").addEventListener("click", () => $("#detail-dialog").close());
$("#detail-reload").addEventListener("click", () => showDetails(detailAccount, true));

/* ---------- 阅读挑战 ---------- */

let challengeAlias = null;
let challengeBooks = [];
const challengePicked = new Set();
let challengeSavedBooks = [];
let challengeRunning = false;
let challengeRunTimer = null;

async function openChallenge() {
  setView("challenge");
  const accounts = allAccounts.length ? allAccounts : await api("/api/accounts").catch(() => []);
  const sel = $("#challenge-account");
  sel.textContent = "";
  if (!accounts.length) {
    // 一个账号都没有：隐藏整个配置面板，只显示引导
    $("#challenge-panel").classList.add("hidden");
    $("#official-challenge").classList.add("hidden");
    $("#challenge-empty").classList.remove("hidden");
    challengeAlias = null;
    return;
  }
  $("#challenge-panel").classList.remove("hidden");
  $("#challenge-empty").classList.add("hidden");
  $("#challenge-banner").classList.remove("hidden");
  // 官方挑战面板由数据决定是否展示:加载期间保持隐藏,避免"闪现再消失"造成抖动。
  $("#official-challenge").classList.add("hidden");
  if (!accounts.some((a) => a.vid === challengeAlias)) challengeAlias = accounts[0].vid;
  for (const a of accounts) {
    const opt = document.createElement("option");
    opt.value = a.vid;
    opt.textContent = displayName(a);
    sel.append(opt);
  }
  sel.value = challengeAlias;
  await loadChallenge();
}

async function loadChallenge() {
  if (!challengeAlias) return;
  $("#challenge-books").innerHTML = skeletonCardsHTML(18, 190, true);
  try {
    // GET details 走本地缓存，瞬时返回；书架与阅读配置一次拿全。
    const d = await api(`/api/accounts/${encodeURIComponent(challengeAlias)}/details`);
    renderChallenge(d);
  } catch (err) {
    $("#challenge-books").innerHTML = `<p class="empty">加载失败：${escapeHtml(err.message)}</p>`;
  }
  loadOfficialChallenge();
}

/* ---------- 官方挑战赛进度 ---------- */

async function loadOfficialChallenge() {
  if (!challengeAlias) return;
  const status = $("#official-challenge-status");
  const panel = $("#official-challenge");
  // 加载期间保持隐藏:未参加挑战时面板最终不展示,先渲染骨架再隐藏会让页面抖动一下。
  panel.classList.add("hidden");
  status.textContent = "—";
  try {
    const d = await api(`/api/accounts/${encodeURIComponent(challengeAlias)}/challenge`);
    renderOfficialChallenge(d.challenge || {});
  } catch (err) {
    // 面板属于辅助信息,加载失败时保持隐藏避免布局跳动;详情留给控制台。
    console.warn("官方挑战进度加载失败:", err.message);
  }
}

function renderOfficialChallenge(c) {
  const body = $("#official-challenge-body");
  const status = $("#official-challenge-status");
  const panel = $("#official-challenge");
  if (!c.id || !c.status) {
    // 未参加官方挑战赛：整个面板不展示。
    panel.classList.add("hidden");
    return;
  }
  panel.classList.remove("hidden");
  status.textContent = c.status === 1 ? "进行中" : c.status === 2 ? "已完成" : "已结束";
  const ch = c.challenge || {};
  const signedDays = (c.readDateList || []).length;
  const dayPct = Math.min(100, (signedDays / (ch.targetDay || 1)) * 100);
  const timePct = Math.min(100, ((c.readTime || 0) / (ch.targetTime || 1)) * 100);
  const now = c.currentTime || Math.floor(Date.now() / 1000);
  const remainDays = c.endTime ? Math.max(0, Math.ceil((c.endTime - now) / 86400)) : 0;

  const R = 52;
  const circ = (2 * Math.PI * R).toFixed(1);
  const ring = (pct, value, unit, label, target) => `
    <div class="challenge-metric">
      <div class="challenge-ring" role="img" aria-label="${label} ${value} ${unit}，目标 ${target} ${unit}，完成 ${Math.round(pct)}%">
        <svg viewBox="0 0 120 120" aria-hidden="true">
          <circle class="ring-track" cx="60" cy="60" r="${R}"></circle>
          <circle class="ring-fill" cx="60" cy="60" r="${R}" stroke-dasharray="${circ}"
            stroke-dashoffset="${(circ * (1 - pct / 100)).toFixed(1)}"${pct <= 0 ? ' style="opacity:0"' : ''}></circle>
        </svg>
        <div class="ring-center"><b>${value}<small>${unit}</small></b><span>${Math.round(pct)}% 已完成</span></div>
      </div>
      <div class="challenge-metric-label">${label}<span>目标 ${target} ${unit}</span></div>
    </div>`;

  body.innerHTML = `
    <div class="challenge-rings">
      ${ring(dayPct, signedDays, "天", "累计打卡", ch.targetDay)}
      ${ring(timePct, ((c.readTime || 0) / 3600).toFixed(1), "小时", "阅读时长", ch.targetTime / 3600)}
    </div>
    <div class="challenge-info">
      <div class="challenge-period">
        <span>挑战周期</span>
        <div><b>${ch.challengeDay} 天<span class="challenge-remaining">剩余 ${remainDays} 天</span></b>
          <p>${fmtDay(c.startTime)} — ${fmtDay(c.endTime)}</p></div>
      </div>
      <div><span>达标条件</span><b>累计阅读 ${fmtDuration(ch.targetTime)} · 打卡 ${ch.targetDay} 天</b></div>
      <div><span>报名费用</span><b>¥${(ch.price / 100).toFixed(0)}<small>报名即得体验卡 ${ch.initRewardCard} 天</small></b></div>
      <div><span>达标奖励</span><b>体验卡 ${ch.reachRewardCard} 天 或 书币 ${ch.reachRewardCoin} 个</b></div>
      <div><span>全站数据</span><b>${c.challengingCnt ?? 0} 人参赛 · ${c.succCnt ?? 0} 人成功</b></div>
    </div>`;
}

function renderChallenge(d) {
  challengeBooks = Array.isArray(d.books) ? d.books : [];
  const cfg = d.reading || {};
  challengePicked.clear();
  challengeSavedBooks = Array.isArray(cfg.book_ids) ? cfg.book_ids : [];
  for (const id of challengeSavedBooks) challengePicked.add(id);

  $("#challenge-state").textContent = cfg.enabled ? "已开启" : "未开启";
  $("#challenge-banner").classList.toggle("on", !!cfg.enabled);
  $("#challenge-banner").classList.toggle("off", !cfg.enabled);
  $("#challenge-runat").textContent = cfg.run_at || "—";
  $("#challenge-minutes").textContent = `${cfg.minutes || 30} 分钟`;
  $("#challenge-enabled").checked = !!cfg.enabled;
  const runatValue = cfg.run_at || "03:00";
  const runatInput = $("#challenge-runat-input");
  if (![...runatInput.options].some((o) => o.value === runatValue)) {
    // 已保存的时间不在 30 分钟档位：补一个选项，避免显示为空
    const opt = document.createElement("option");
    opt.value = opt.textContent = runatValue;
    runatInput.append(opt);
  }
  if (runatInput.value !== runatValue) {
    runatInput.value = runatValue;
    // 派发 change 让自绘下拉同步按钮文案
    runatInput.dispatchEvent(new Event("change"));
  }
  $("#challenge-minutes-input").value = cfg.minutes || 30;
  updateChallengeButtons(!!cfg.running);
  $("#challenge-laststatus").textContent = cfg.last_status
    ? `上次执行（${cfg.last_run_date || "—"}):${cfg.last_status}`
    : "尚未执行过";

  const grid = $("#challenge-books");
  grid.textContent = "";
  if (!challengeBooks.length) {
    grid.innerHTML = '<p class="empty">书架为空，先在详情里刷新数据。</p>';
    return;
  }
  for (const b of challengeBooks) {
    const card = document.createElement("div");
    card.className = "book-card" + (challengePicked.has(b.bookId) ? " selected" : "");
    card.title = b.title || b.bookId || "";
    if (b.cover) {
      const img = document.createElement("img");
      img.loading = "lazy";
      img.src = b.cover;
      img.alt = "";
      card.append(img);
    } else {
      const ph = document.createElement("div");
      ph.className = "book-card-placeholder";
      ph.textContent = "▤";
      card.append(ph);
    }
    const meta = document.createElement("div");
    meta.className = "book-meta";
    const bookTitle = b.title || "未命名书籍";
    const bookAuthor = b.author || b.authorName || "未知作者";
    meta.innerHTML = `<div class="book-title" title="${escapeHtml(bookTitle)}">${escapeHtml(bookTitle)}</div><div class="book-author">${escapeHtml(bookAuthor)}</div>`;
    card.append(meta);
    const check = document.createElement("span");
    check.className = "check";
    check.textContent = "✓";
    card.append(check);
    card.addEventListener("click", () => {
      if (challengePicked.has(b.bookId)) {
        challengePicked.delete(b.bookId);
        card.classList.remove("selected");
      } else {
        challengePicked.add(b.bookId);
        card.classList.add("selected");
      }
    });
    grid.append(card);
  }
}

// 「立即执行」与「停止阅读」是同一个按钮：按会话状态切换文案、图标与配色。
function updateChallengeButtons(running) {
  challengeRunning = !!running;
  const btn = $("#challenge-run");
  btn.classList.toggle("danger", challengeRunning);
  btn.innerHTML = challengeRunning
    ? `<span class="nav-icon">${iconSvg("square")}</span>停止阅读`
    : `<span class="nav-icon">${iconSvg("play")}</span>立即执行`;
}

$("#challenge-account").addEventListener("change", (e) => {
  challengeAlias = e.target.value;
  loadChallenge();
});

$("#challenge-save").addEventListener("click", async () => {
  if (!challengeAlias) return;
  if (!challengePicked.size) {
    toast("请至少选择一本书籍再保存", "error");
    return;
  }
  const btn = $("#challenge-save");
  setBusy(btn, true);
  try {
    await api(`/api/accounts/${encodeURIComponent(challengeAlias)}/reading`, {
      method: "POST",
      body: JSON.stringify({
        enabled: $("#challenge-enabled").checked,
        book_ids: Array.from(challengePicked),
        minutes: Number($("#challenge-minutes-input").value) || 30,
        run_at: $("#challenge-runat-input").value || "03:00",
      }),
    });
    $("#challenge-state").textContent = $("#challenge-enabled").checked ? "已开启" : "未开启";
    challengeSavedBooks = Array.from(challengePicked);
    $("#challenge-banner").classList.toggle("on", $("#challenge-enabled").checked);
    $("#challenge-banner").classList.toggle("off", !$("#challenge-enabled").checked);
    $("#challenge-runat").textContent = $("#challenge-runat-input").value || "03:00";
    $("#challenge-minutes").textContent = `${$("#challenge-minutes-input").value || 30} 分钟`;
    toast($("#challenge-enabled").checked ? "已保存，到点自动参与阅读挑战" : "配置已保存（未开启）", "ok");
  } catch (err) {
    toast(`保存失败：${err.message}`, "error");
  }
  setBusy(btn, false);
});

async function runChallengeNow() {
  if (!challengeAlias) return;
  if (!challengeSavedBooks.length) {
    toast("请先选择要阅读的书籍，并点「保存配置」", "error");
    return;
  }
  const btn = $("#challenge-run");
  setBusy(btn, true);
  try {
    const data = await api(`/api/accounts/${encodeURIComponent(challengeAlias)}/reading/run`, { method: "POST" });
    toast(data.resumed ? "继续上次未完成的阅读" : "阅读会话已启动，每 30 秒记 0.5 分钟", "ok");
  } catch (err) {
    toast(`启动失败：${err.message}`, "error");
    setBusy(btn, false);
    return;
  }
  setBusy(btn, false);
  updateChallengeButtons(true);
  let polls = 0;
  if (challengeRunTimer) clearInterval(challengeRunTimer);
  challengeRunTimer = setInterval(async () => {
    polls++;
    try {
      const cfg = await api(`/api/accounts/${encodeURIComponent(challengeAlias)}/reading`);
      updateChallengeButtons(!!cfg.running);
      // 会话是否结束以 running 为准；结束后按钮回到「立即执行」。
      if (!cfg.running) {
        $("#challenge-laststatus").textContent = cfg.last_status || "已结束";
        clearInterval(challengeRunTimer);
        return;
      }
      $("#challenge-laststatus").textContent = cfg.last_status || "阅读中…";
    } catch { /* 单次轮询失败忽略 */ }
    if (polls > 360) clearInterval(challengeRunTimer);
  }, 5000);
}

async function stopChallenge() {
  if (!challengeAlias) return;
  const okStop = await confirmAction({
    title: "停止阅读",
    message: "确定停止本次阅读会话？已上报的时长会保留，当日不再重跑。",
    confirmText: "停止",
  });
  if (!okStop) return;
  try {
    await api(`/api/accounts/${encodeURIComponent(challengeAlias)}/reading/stop`, { method: "POST" });
    $("#challenge-laststatus").textContent = "正在停止…";
  } catch (err) {
    toast(`停止失败：${err.message}`, "error");
  }
}

$("#challenge-run").addEventListener("click", () => (challengeRunning ? stopChallenge() : runChallengeNow()));
$("#challenge-goto-accounts").addEventListener("click", () => {
  setView("accounts");
  loadAccounts();
});

/* ---------- 我的阅读（周阅读奖励） ---------- */

let readingAlias = null;
let readingData = null;
let readingCard = null;
let readingPrefs = {};
let weeklyRequest = 0;

async function openReading() {
  setView("reading");
  const accounts = allAccounts.length ? allAccounts : await api("/api/accounts").catch(() => []);
  const sel = $("#reading-account");
  sel.textContent = "";
  if (!accounts.length) {
    $("#reading-empty").classList.remove("hidden");
    $$("#view-reading .panel, #reading-stats").forEach((el) => el.classList.add("hidden"));
    readingAlias = null;
    return;
  }
  $("#reading-empty").classList.add("hidden");
  $$("#view-reading .panel, #reading-stats").forEach((el) => el.classList.remove("hidden"));
  if (!accounts.some((a) => a.vid === readingAlias)) readingAlias = accounts[0].vid;
  for (const a of accounts) {
    const opt = document.createElement("option");
    opt.value = a.vid;
    opt.textContent = displayName(a);
    sel.append(opt);
  }
  sel.value = readingAlias;
  await loadWeekly();
}

function resetWeeklyUI() {
  // 加载前清空全部动态区域：切换账号后不能残留上一账号的数据。
  readingData = null;
  readingCard = null;
  readingPrefs = {};
  for (const id of ["reading-stat-time", "reading-stat-week-detail", "reading-stat-days",
    "reading-stat-month", "reading-stat-month-detail", "reading-stat-card", "reading-stat-card-sub"]) {
    $("#" + id).textContent = "—";
  }
  $("#reading-time-progress-label").textContent = "";
  $("#reading-day-progress-label").textContent = "";
  $("#reading-rules").textContent = "";
}

async function loadWeekly() {
  if (!readingAlias) return;
  const request = ++weeklyRequest;
  resetWeeklyUI();
  $("#reading-time-awards").innerHTML = skeletonCardsHTML(5, 148);
  $("#reading-day-awards").innerHTML = skeletonCardsHTML(3, 148);
  try {
    const d = await api(`/api/accounts/${encodeURIComponent(readingAlias)}/weekly`);
    if (request !== weeklyRequest) return;
    readingData = d.weekly || {};
    readingCard = d.card || null;
    readingPrefs = d.prefs || {};
    renderWeekly();
  } catch (err) {
    if (request !== weeklyRequest) return;
    const msg = `加载失败：${escapeHtml(err.message)}`;
    $("#reading-time-awards").innerHTML = `<p class="empty">${msg}</p>`;
    $("#reading-day-awards").innerHTML = `<p class="empty">${msg}</p>`;
  }
}

// 秒数 → 「X 小时 Y 分钟」；不足 1 分钟按秒展示。
function fmtDuration(sec) {
  sec = Math.max(0, Math.floor(sec || 0));
  if (sec === 0) return "0 分钟";
  const h = Math.floor(sec / 3600);
  const m = Math.floor((sec % 3600) / 60);
  if (h > 0) return m > 0 ? `${h} 小时 ${m} 分钟` : `${h} 小时`;
  return m > 0 ? `${m} 分钟` : `${sec} 秒`;
}

// 「读 30 分钟」「读 1 小时」「读 2 天」→ 秒数 / 天数，用于进度条。
function parseLevelGoal(desc) {
  const sec = desc && desc.match(/读 (\d+) 分钟/);
  if (sec) return { seconds: Number(sec[1]) * 60 };
  const hour = desc && desc.match(/读 (\d+) 小时/);
  if (hour) return { seconds: Number(hour[1]) * 3600 };
  const day = desc && desc.match(/读 (\d+) 天/);
  if (day) return { days: Number(day[1]) };
  return {};
}

// 奖品描述：优先用档位的 awardChoices 自行拼接（统一叫体验卡）,
// 没有明细时退回服务端原文并把"无限卡"替换为"体验卡"。
function choicesDesc(a) {
  const choices = a.awardChoices || [];
  if (choices.length) {
    return "可得 " + choices
      .map((c) => (c.choiceType === 1 ? `体验卡 ${c.awardNum} 天` : `书币 ${c.awardNum} 个`))
      .join(" 或 ");
  }
  return String(a.awardChoicesDesc || "").replaceAll("无限卡", "体验卡");
}

function weeklyCard(a, weekly, prefs) {
  const goal = parseLevelGoal(a.awardLevelDesc);
  const claimed = a.awardStatus === 2;
  const claimable = a.awardStatus === 1;
  const locked = a.awardStatus === 0;
  const prefChoice = Number(prefs[a.awardLevelId]) || 0;

  const card = document.createElement("div");
  card.className = "reward-card";

  const top = document.createElement("div");
  top.className = "reward-top";
  top.innerHTML = `<span class="reward-title">${escapeHtml(a.awardLevelDesc || "阅读奖励")}</span>`;
  const badge = document.createElement("span");
  badge.className = "award-badge";
  badge.textContent = a.awardStatusDesc || (claimed ? "已领取" : claimable ? "可领取" : "未达成");
  top.append(badge);
  card.append(top);

  const desc = document.createElement("div");
  desc.className = "reward-desc";
  desc.textContent = choicesDesc(a);
  card.append(desc);

  // 进度条：时长奖励按本周秒数，天数奖励按本周天数；右侧标注当前进度。
  // 达成/已领取的档位进度条保持 100% 展示。
  if (goal.seconds || goal.days) {
    const done = goal.seconds ? weekly.readingTime : weekly.readingDay;
    const need = goal.seconds || goal.days;
    const pct = Math.min(100, Math.round((done / need) * 100));
    let label = goal.seconds
      ? `${Math.floor(done / 60)} / ${need / 60} 分钟`
      : `${done} / ${need} 天`;
    if (goal.seconds && need >= 3600) {
      // 小时档位用小时计（"1 / 3 小时")，分钟档位用分钟计。
      label = `${Math.round((done / 3600) * 10) / 10} / ${need / 3600} 小时`;
    }
    const row = document.createElement("div");
    row.className = "reward-progress-row";
    row.innerHTML = `<div class="reward-progress"><i style="transform:scaleX(${(pct / 100).toFixed(4)})"></i></div>`;
    const tag = document.createElement("span");
    tag.className = "reward-progress-label";
    tag.textContent = label;
    row.append(tag);
    card.append(row);
  }

  // 双按钮：书币 / 体验卡。任何时候都可点击 —— 点击保存每周的自动领取设置，
  // 设置写入数据库；若当前正好可领取，保存后立即领取一次。选中项以描边色标识。
  const actions = document.createElement("div");
  actions.className = "reward-actions";
  for (const c of a.awardChoices || []) {
    const isCard = c.choiceType === 1;
    const name = isCard ? "体验卡" : "书币";
    const val = `${c.awardNum}${isCard ? " 天" : " 个"}`;
    const btn = document.createElement("button");
    btn.type = "button";
    btn.className = "reward-btn";
    btn.title = `设为每周自动领取：${name} ${val}`;
    if (claimable) {
      btn.classList.add("claimable");
      btn.textContent = `${name} ${val}`;
    } else {
      // 选中描边由保存的设置决定，与本周是否领取过无关：已领取的档位换选另
      // 一种后，描边移动到新选项上，表示下周起自动领取它。
      const selected = prefChoice === c.choiceType;
      const claimedThisWeek = claimed && a.awardChooseType === c.choiceType;
      btn.classList.add(selected ? "selected" : "plain");
      if (claimedThisWeek) {
        btn.classList.add("claimed");
        btn.textContent = `✓ ${name} ${val}`;
      } else {
        btn.textContent = `${name} ${val}`;
      }
    }
    btn.addEventListener("click", () => saveRewardChoice(a, c, btn));
    actions.append(btn);
  }
  card.append(actions);
  return card;
}

// 把本周每日阅读秒数渲染成迷你柱状图：柱高是当天时长，柱下标注星期，
// 今天用强调色标出。每根柱子可聚焦，悬停或键盘聚焦时显示具体时长。
// readDays 为周一到周日 7 项。
function renderWeekBars(root, readDays) {
  const labels = ["一", "二", "三", "四", "五", "六", "日"];
  const days = readDays.slice(-labels.length);
  const offset = labels.length - days.length;
  root.textContent = "";
  if (!days.length || days.every((s) => !s)) {
    root.textContent = "本周还没有阅读记录";
    return;
  }
  const max = Math.max(...days, 1);
  // 仅当拿到完整一周时才能把下标当作星期；今天是周几（0=周一）。
  const today = days.length === labels.length ? (new Date().getDay() + 6) % 7 : -1;
  const bars = document.createElement("div");
  bars.className = "week-bars";
  bars.setAttribute("role", "list");
  bars.setAttribute("aria-label", "本周每日阅读时长");
  root.append(bars);
  days.forEach((sec, i) => {
    const label = labels[offset + i];
    const value = sec > 0 ? fmtDuration(sec) : "没有阅读";
    const col = document.createElement("span");
    col.className = "week-bar";
    if (i === today) col.classList.add("today");
    col.setAttribute("role", "listitem");
    col.setAttribute("tabindex", "0");
    col.setAttribute("aria-label", `周${label}:${value}`);
    col.dataset.tip = `周${label} · ${value}`;
    const bar = document.createElement("i");
    bar.style.height = (sec > 0 ? Math.max(3, Math.round((sec / max) * 28)) : 2) + "px";
    const name = document.createElement("b");
    name.textContent = label;
    col.append(bar, name);
    bars.append(col);
  });
}

function renderWeekly() {
  const w = readingData || {};
  $("#reading-stat-time").textContent = fmtDuration(w.readingTime);
  renderWeekBars($("#reading-stat-week-detail"), (w.weekReadDaysDetail || {}).readDays || []);
  $("#reading-stat-days").textContent = `${w.readingDay ?? 0} 天`;
  $("#reading-stat-month").textContent = fmtDuration((w.monthReadDaysDetail || {}).readTimes);
  const monthDays = ((w.monthReadDaysDetail || {}).readDays || []).filter((s) => s > 0).length;
  $("#reading-stat-month-detail").textContent = `本月有 ${monthDays} 天读过`;
  // 无限卡剩余天数以详情缓存的会员卡信息为准（官方 memberCardSummary 接口）,
  // weekly 里的 infiniteCard 字段不反映真实余额，仅作兜底。
  const mc = readingCard || {};
  const ic = w.infiniteCard || {};
  if (mc.has) {
    $("#reading-stat-card").textContent = `${mc.days} 天`;
    $("#reading-stat-card-sub").textContent = `${fmtDay(mc.expired_at)} 到期`;
  } else if (ic.day > 0) {
    $("#reading-stat-card").textContent = `${ic.day} 天`;
    $("#reading-stat-card-sub").textContent = ic.paying ? "付费会员" : ic.cardType === "novice" ? "新手卡" : "免费体验";
  } else {
    $("#reading-stat-card").textContent = "无";
    $("#reading-stat-card-sub").textContent = "还没有可用的体验卡";
  }

  // 时长奖励按档位秒数从小到大排序，天数奖励同理。
  const byGoal = (a, b) => {
    const ga = parseLevelGoal(a.awardLevelDesc) || {};
    const gb = parseLevelGoal(b.awardLevelDesc) || {};
    return (ga.seconds || (ga.days || 0) * 86400) - (gb.seconds || (gb.days || 0) * 86400);
  };
  const timeGrid = $("#reading-time-awards");
  timeGrid.textContent = "";
  for (const a of [...(w.readtimeAwards || [])].sort(byGoal)) timeGrid.append(weeklyCard(a, w, readingPrefs));
  $("#reading-time-progress-label").innerHTML = `本周已读 <b>${escapeHtml(fmtDuration(w.readingTime))}</b>`;

  const dayGrid = $("#reading-day-awards");
  dayGrid.textContent = "";
  for (const a of [...(w.readdayAwards || [])].sort(byGoal)) dayGrid.append(weeklyCard(a, w, readingPrefs));
  $("#reading-day-progress-label").innerHTML = `本周已读 <b>${w.readingDay ?? 0} 天</b>`;

  const rules = $("#reading-rules");
  rules.textContent = "";
  for (const rule of w.rewardRules || []) {
    const li = document.createElement("li");
    li.textContent = rule;
    rules.append(li);
  }
}

/* ---------- 奖品选择（保存每周设置，由调度器自动领取） ---------- */

function choiceLabel(choiceType, num) {
  return choiceType === 1 ? `体验卡 ${num} 天` : `书币 ${num} 个`;
}

async function saveRewardChoice(a, c, btn) {
  setBusy(btn, true);
  try {
    const p = await api(`/api/accounts/${encodeURIComponent(readingAlias)}/weekly/pref`, {
      method: "POST",
      body: JSON.stringify({ award_level_id: a.awardLevelId, choice_type: c.choiceType }),
    });
    readingPrefs = p.prefs || {};
    toast(`设置已保存：将自动领取${choiceLabel(c.choiceType, c.awardNum)}`, "ok");
    renderWeekly();
  } catch (err) {
    toast(`设置保存失败：${err.message}`, "error");
    setBusy(btn, false);
  }
}

$("#reading-account").addEventListener("change", (e) => {
  readingAlias = e.target.value;
  loadWeekly();
});
$("#reading-reload").addEventListener("click", loadWeekly);
$("#reading-goto-accounts").addEventListener("click", () => {
  setView("accounts");
  loadAccounts();
});

/* ---------- 运行日志 ---------- */

let logsTimer = null;

async function openLogs() {
  setView("logs");
  const sel = $("#log-alias");
  const current = sel.value;
  sel.textContent = "";
  const all = document.createElement("option");
  all.value = "";
  all.textContent = "全部账号";
  sel.append(all);
  const accounts = allAccounts.length ? allAccounts : await api("/api/accounts").catch(() => []);
  for (const a of accounts) {
    const opt = document.createElement("option");
    opt.value = a.vid;
    opt.textContent = a.name || a.remark || a.vid;
    sel.append(opt);
  }
  sel.value = current;
  await loadLogs();
  if (!logsTimer) logsTimer = setInterval(() => { if (currentView === "logs") loadLogs(true); }, 10000);
}

async function loadLogs(silent = false) {
  const params = new URLSearchParams({ limit: "200" });
  if ($("#log-alias").value) params.set("vid", $("#log-alias").value);
  if ($("#log-level").value) params.set("level", $("#log-level").value);
  try {
    const logs = await api(`/api/logs?${params}`);
    $("#logs-rows").textContent = "";
    if (!logs.length) {
      $("#logs-rows").innerHTML = '<p class="empty">暂无日志</p>';
    } else {
      for (const e of logs) {
        const row = document.createElement("div");
        row.className = `log-row log-${e.level}`;
        const time = document.createElement("span");
        time.className = "log-time";
        time.dataset.label = "时间";
        time.textContent = fmtTime(e.ts);
        const level = document.createElement("span");
        level.className = "log-level";
        level.dataset.label = "级别";
        level.textContent = e.level;
        const source = document.createElement("span");
        source.className = "log-source";
        source.dataset.label = "来源";
        source.textContent = e.source || "-";
        const account = document.createElement("span");
        account.className = "log-alias";
        account.dataset.label = "账号";
        account.textContent = e.name || e.vid || "-";
        account.title = e.vid;
        const msg = document.createElement("span");
        msg.className = "log-msg";
        msg.dataset.label = "内容";
        msg.textContent = e.message;
        msg.title = e.message;
        row.append(time, level, source, account, msg);
        $("#logs-rows").append(row);
      }
      $("#logs-list").scrollTop = 0;
    }
    $("#logs-updated").textContent = `更新于 ${new Date().toLocaleTimeString("zh-CN", { hour12: false })}`;
  } catch (err) {
    if (!silent) toast(`加载日志失败：${err.message}`, "error");
  }
}

$("#logs-reload").addEventListener("click", () => loadLogs());
$("#log-alias").addEventListener("change", () => loadLogs());
$("#log-level").addEventListener("change", () => loadLogs());
$("#logs-clear").addEventListener("click", async () => {
  const okClear = await confirmAction({
    title: "清空日志",
    message: "确定清空全部日志？清空后无法恢复。",
    confirmText: "清空",
  });
  if (!okClear) return;
  try {
    await api("/api/logs", { method: "DELETE" });
    toast("日志已清空", "ok");
    loadLogs();
  } catch (err) {
    toast(`清空失败：${err.message}`, "error");
  }
});

/* ---------- 推送设置 ---------- */

const pushTypes = {
  bark: {
    label: "Bark", desc: "iOS 通知，免费、可自建服务端",
    fields: [["device_key", "Device Key", true], ["server", "服务端地址", false, "https://api.day.app"]],
  },
  telegram: {
    label: "Telegram", desc: "无条数限制，网络需可访问 Telegram",
    fields: [["bot_token", "Bot Token", true], ["chat_id", "Chat ID", true]],
  },
  serverchan: {
    label: "Server酱", desc: "消息直达微信，免费版每天限 5 条",
    fields: [["send_key", "SendKey", true]],
  },
  pushplus: {
    label: "pushplus", desc: "微信公众号推送，需关注公众号",
    fields: [["token", "Token", true]],
  },
};

async function loadPushChannels() {
  const channels = await api("/api/settings/push");
  const byType = {};
  for (const c of channels) byType[c.type] = c;

  let enabledCount = 0;
  const list = $("#push-list");
  list.textContent = "";

  for (const [type, meta] of Object.entries(pushTypes)) {
    const ch = byType[type] || { type, enabled: false, params: {} };
    if (ch.enabled) enabledCount++;
    const saved = ch.params || {};

    const card = document.createElement("div");
    card.className = "push-card";

    const head = document.createElement("div");
    head.className = "push-card-head";
    const info = document.createElement("div");
    const name = document.createElement("div");
    name.className = "push-card-name";
    name.textContent = meta.label;
    const desc = document.createElement("div");
    desc.className = "push-card-desc";
    desc.textContent = meta.desc;
    info.append(name, desc);

    // 开关：切换即保存
    const sw = document.createElement("label");
    sw.className = "switch";
    const swInput = document.createElement("input");
    swInput.type = "checkbox";
    swInput.checked = !!ch.enabled;
    swInput.setAttribute("aria-label", `启用 ${meta.label} 推送`);
    const slider = document.createElement("i");
    sw.append(swInput, slider);
    swInput.addEventListener("change", () => saveChannel(type, swInput.checked, collectParams(card), swInput));
    head.append(info, sw);
    card.append(head);

    const fields = document.createElement("div");
    fields.className = "push-fields";
    const inputs = {};
    for (const [key, label, required, def] of meta.fields) {
      const field = document.createElement("label");
      field.className = "push-field";
      const fieldLabel = document.createElement("span");
      fieldLabel.textContent = required ? `${label}（必填）` : label;
      const input = document.createElement("input");
      input.type = key.includes("token") || key.includes("key") ? "password" : "text";
      input.className = "push-param-input";
      input.value = saved[key] !== undefined ? saved[key] : (def || "");
      input.placeholder = label;
      input.setAttribute("aria-label", label);
      input.dataset.key = key;
      if (required) input.required = true;
      inputs[key] = input;
      field.append(fieldLabel, input);
      fields.append(field);
    }
    card.append(fields);

    const ops = document.createElement("div");
    ops.className = "push-card-ops";
    const saveBtn = button("保存参数", "btn", () => saveChannel(type, swInput.checked, collectParams(fields), saveBtn));
    const testBtn = button("发送测试", "btn", async () => {
      const params = collectParams(fields);
      setBusy(testBtn, true);
      try {
        await api(`/api/settings/push/${type}/test`, { method: "POST", body: JSON.stringify({ params }) });
        toast("测试消息已发送，请查收", "ok");
      } catch (err) {
        toast(`测试失败：${err.message}`, "error");
      }
      setBusy(testBtn, false);
    });
    ops.append(saveBtn, testBtn);
    card.append(ops);

    function collectParams(scope) {
      const out = {};
      scope.querySelectorAll(".push-param-input").forEach((i) => { out[i.dataset.key] = i.value.trim(); });
      return out;
    }

    list.append(card);
  }
  const summary = $("#push-summary");
  if (summary) summary.textContent = `${enabledCount}/${Object.keys(pushTypes).length} 个渠道已开启`;
}

async function saveChannel(type, enabled, params, btn) {
  setBusy(btn, true);
  try {
    await api("/api/settings/push", {
      method: "POST",
      body: JSON.stringify({ type, enabled, params }),
    });
    toast(enabled ? "已开启" : "配置已保存（未开启）", "ok");
  } catch (err) {
    toast(`保存失败：${err.message}`, "error");
  }
  setBusy(btn, false);
}

/* ---------- 启动 ---------- */

renderIcons();
loadAccounts();
