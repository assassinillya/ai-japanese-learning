const state = {
  token: localStorage.getItem("access_token") || "",
  user: null,
  selectedArticleId: null,
};

const views = document.querySelectorAll(".view");
const messageBox = document.getElementById("message-box");
const authStatus = document.getElementById("auth-status");
const homeGreeting = document.getElementById("home-greeting");
const profileSummary = document.getElementById("profile-summary");
const libraryList = document.getElementById("library-list");
const articleList = document.getElementById("article-list");
const articleDetail = document.getElementById("article-detail");
const sentenceList = document.getElementById("sentence-list");

document.querySelectorAll("[data-view]").forEach((button) => {
  button.addEventListener("click", () => showView(button.dataset.view));
});

document.getElementById("register-form").addEventListener("submit", async (event) => {
  event.preventDefault();
  const payload = Object.fromEntries(new FormData(event.currentTarget).entries());
  const result = await request("/api/auth/register", {
    method: "POST",
    body: JSON.stringify(payload),
  });
  handleAuthResult(result, "注册成功");
});

document.getElementById("login-form").addEventListener("submit", async (event) => {
  event.preventDefault();
  const payload = Object.fromEntries(new FormData(event.currentTarget).entries());
  const result = await request("/api/auth/login", {
    method: "POST",
    body: JSON.stringify(payload),
  });
  handleAuthResult(result, "登录成功");
});

document.getElementById("logout-button").addEventListener("click", async () => {
  if (!state.token) {
    setMessage("当前未登录");
    return;
  }
  await request("/api/auth/logout", { method: "POST" });
  localStorage.removeItem("access_token");
  state.token = "";
  state.user = null;
  state.selectedArticleId = null;
  renderUser();
  setMessage("已退出登录");
});

document.getElementById("jlpt-form").addEventListener("submit", async (event) => {
  event.preventDefault();
  const payload = Object.fromEntries(new FormData(event.currentTarget).entries());
  const result = await request("/api/profile/jlpt-level", {
    method: "PUT",
    body: JSON.stringify(payload),
  });
  if (result.ok) {
    state.user = result.data;
    renderUser();
    setMessage("JLPT 等级已更新");
  }
});

document.getElementById("article-form").addEventListener("submit", async (event) => {
  event.preventDefault();
  const payload = Object.fromEntries(new FormData(event.currentTarget).entries());
  const result = await request("/api/articles", {
    method: "POST",
    body: JSON.stringify(payload),
  });
  if (!result.ok) {
    return;
  }
  state.selectedArticleId = result.data.id;
  await Promise.all([loadArticles(), loadArticleDetail(result.data.id)]);
  showView("detail");
  setMessage("文章已创建并处理完成");
  event.currentTarget.reset();
});

document.getElementById("reprocess-button").addEventListener("click", async () => {
  if (!state.selectedArticleId) {
    setMessage("请先选择一篇文章");
    return;
  }
  const result = await request(`/api/articles/${state.selectedArticleId}/process`, {
    method: "POST",
  });
  if (!result.ok) {
    return;
  }
  await loadArticleDetail(state.selectedArticleId);
  await loadArticles();
  setMessage("文章已重新处理");
});

async function bootstrap() {
  if (state.token) {
    const me = await request("/api/auth/me");
    if (me.ok) {
      state.user = me.data;
      await Promise.all([loadLibrary(), loadArticles()]);
    } else {
      localStorage.removeItem("access_token");
      state.token = "";
    }
  }
  renderUser();
}

function showView(name) {
  views.forEach((view) => view.classList.toggle("active", view.id === `view-${name}`));
}

function renderUser() {
  if (!state.user) {
    authStatus.textContent = "未登录";
    homeGreeting.textContent = "登录后可查看文章库、上传文章并进入处理流程。";
    profileSummary.textContent = "尚未加载资料。";
    libraryList.innerHTML = "";
    articleList.innerHTML = "";
    articleDetail.textContent = "请选择一篇文章。";
    sentenceList.innerHTML = "";
    showView("login");
    return;
  }

  authStatus.textContent = `已登录：${state.user.username}（${state.user.email}）`;
  homeGreeting.textContent = `欢迎回来，${state.user.username}。当前 JLPT：${state.user.jlpt_level}`;
  profileSummary.textContent = [
    `用户名：${state.user.username}`,
    `邮箱：${state.user.email}`,
    `JLPT：${state.user.jlpt_level}`,
    `首次引导完成：${state.user.onboarding_completed ? "是" : "否"}`,
  ].join("\n");
  document.querySelector('#jlpt-form select[name="jlpt_level"]').value = state.user.jlpt_level;
  showView("home");
}

function handleAuthResult(result, successMessage) {
  if (!result.ok) {
    return;
  }
  state.token = result.data.access_token;
  state.user = result.data.user;
  localStorage.setItem("access_token", state.token);
  renderUser();
  Promise.all([loadLibrary(), loadArticles()]);
  setMessage(successMessage);
}

async function loadLibrary() {
  const result = await request("/api/articles/library");
  if (!result.ok) {
    return;
  }
  const items = result.data.items || [];
  libraryList.innerHTML = items
    .map(
      (article) => `
        <li>
          <button class="link-button" data-article-id="${article.id}">
            ${escapeHTML(article.title)}
          </button>
          <span class="meta">${article.jlpt_level} / ${article.translation_status} / ${article.sentence_count} 句</span>
        </li>
      `,
    )
    .join("");

  bindArticleSelection(libraryList);
}

async function loadArticles() {
  const result = await request("/api/articles");
  if (!result.ok) {
    return;
  }
  const items = result.data.items || [];
  articleList.innerHTML = items
    .map(
      (article) => `
        <li>
          <button class="link-button" data-article-id="${article.id}">
            ${escapeHTML(article.title)}
          </button>
          <span class="meta">${article.original_language} / ${article.translation_status} / ${article.sentence_count} 句</span>
        </li>
      `,
    )
    .join("");

  bindArticleSelection(articleList);
}

async function loadArticleDetail(articleId) {
  const [articleResult, sentenceResult] = await Promise.all([
    request(`/api/articles/${articleId}`),
    request(`/api/articles/${articleId}/sentences`),
  ]);
  if (!articleResult.ok || !sentenceResult.ok) {
    return;
  }

  const article = articleResult.data;
  articleDetail.textContent = [
    `标题：${article.title}`,
    `原文语言：${article.original_language}`,
    `JLPT：${article.jlpt_level}`,
    `处理状态：${article.translation_status}`,
    `来源类型：${article.source_type}`,
    `句子数量：${article.sentence_count}`,
    `处理说明：${article.processing_notes || "-"}`,
    `原文预览：${article.original_content || "-"}`,
    `中文翻译：${article.chinese_translation || "-"}`,
    "",
    "日语内容：",
    article.japanese_content,
  ].join("\n");

  sentenceList.innerHTML = (sentenceResult.data.items || [])
    .map((sentence) => `<li>${escapeHTML(sentence.sentence_text)}</li>`)
    .join("");

  const reprocessButton = document.getElementById("reprocess-button");
  reprocessButton.disabled = article.source_type === "builtin";
  reprocessButton.title = article.source_type === "builtin" ? "内置文章无需重新处理" : "";
}

function bindArticleSelection(container) {
  container.querySelectorAll("[data-article-id]").forEach((button) => {
    button.addEventListener("click", async () => {
      const articleId = Number(button.dataset.articleId);
      state.selectedArticleId = articleId;
      await loadArticleDetail(articleId);
      showView("detail");
    });
  });
}

async function request(url, options = {}) {
  try {
    const response = await fetch(url, {
      headers: {
        "Content-Type": "application/json",
        ...(state.token ? { Authorization: `Bearer ${state.token}` } : {}),
        ...(options.headers || {}),
      },
      ...options,
    });
    const data = await response.json().catch(() => ({}));
    if (!response.ok) {
      setMessage(data.error || `请求失败：${response.status}`);
      return { ok: false, data };
    }
    return { ok: true, data };
  } catch (error) {
    setMessage(`网络错误：${error.message}`);
    return { ok: false, data: null };
  }
}

function setMessage(message) {
  messageBox.textContent = message;
}

function escapeHTML(input) {
  return String(input)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}

bootstrap();
