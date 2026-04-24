const state = {
  token: localStorage.getItem("access_token") || "",
  user: null,
  selectedArticleId: null,
  readingArticle: null,
  lookup: {
    timer: null,
    currentText: "",
    currentSentenceId: null,
    currentSentenceText: "",
    currentEntry: null,
    lastLookupKey: "",
  },
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
const readingHeader = document.getElementById("reading-header");
const readingContent = document.getElementById("reading-content");
const popup = document.getElementById("lookup-popup");
const popupCard = popup.querySelector(".lookup-popup-card");
const popupTitle = document.getElementById("lookup-popup-title");
const popupBody = document.getElementById("lookup-popup-body");
const addVocabularyButton = document.getElementById("add-vocabulary-button");
const openReadingButton = document.getElementById("open-reading-button");
const reprocessButton = document.getElementById("reprocess-button");

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
  state.readingArticle = null;
  hideLookupPopup();
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
  const result = await request("/api/articles/upload", {
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

reprocessButton.addEventListener("click", async () => {
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
  await Promise.all([loadArticleDetail(state.selectedArticleId), loadArticles()]);
  setMessage("文章已重新处理");
});

openReadingButton.addEventListener("click", async () => {
  if (!state.selectedArticleId) {
    setMessage("请先选择一篇文章");
    return;
  }
  await loadReadingArticle(state.selectedArticleId);
  showView("reading");
});

addVocabularyButton.addEventListener("click", async () => {
  if (!state.lookup.currentEntry) {
    return;
  }
  const payload = {
    dictionary_entry_id: state.lookup.currentEntry.id,
    article_id: state.selectedArticleId,
    source_sentence_id: state.lookup.currentSentenceId,
    selected_text: state.lookup.currentText,
    source_sentence_text: state.lookup.currentSentenceText,
  };
  const result = await request("/api/vocabulary", {
    method: "POST",
    body: JSON.stringify(payload),
  });
  if (!result.ok) {
    return;
  }
  addVocabularyButton.disabled = true;
  addVocabularyButton.textContent = "已加入生词本";
  setMessage(result.data.created ? "已加入生词本" : "该词已在生词本中");
});

readingContent.addEventListener("mouseup", () => {
  scheduleLookupFromSelection();
});

document.addEventListener("mousedown", (event) => {
  if (!popupCard.contains(event.target)) {
    hideLookupPopup();
  }
});

document.addEventListener("selectionchange", () => {
  const selection = window.getSelection();
  if (!selection || selection.isCollapsed) {
    clearPendingLookup();
  }
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
    readingHeader.textContent = "请选择一篇文章进入阅读。";
    readingContent.innerHTML = "";
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
  state.selectedArticleId = article.id;
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

  reprocessButton.disabled = article.source_type === "builtin";
  reprocessButton.title = article.source_type === "builtin" ? "内置文章无需重新处理" : "";
}

async function loadReadingArticle(articleId) {
  const result = await request(`/api/reading/articles/${articleId}`);
  if (!result.ok) {
    return;
  }

  const { article, sentences } = result.data;
  state.selectedArticleId = article.id;
  state.readingArticle = article;
  hideLookupPopup();
  readingHeader.textContent = [
    `标题：${article.title}`,
    `JLPT：${article.jlpt_level}`,
    `处理状态：${article.translation_status}`,
    "提示：在下方正文中选中文本以查词。",
  ].join("\n");

  readingContent.innerHTML = (sentences || [])
    .map(
      (sentence) => `
        <p class="reading-sentence" data-sentence-id="${sentence.id}" data-sentence-text="${escapeHTMLAttribute(sentence.sentence_text)}">
          <span class="reading-text">${escapeHTML(sentence.sentence_text)}</span>
        </p>
      `,
    )
    .join("");
}

function bindArticleSelection(container) {
  container.querySelectorAll("[data-article-id]").forEach((button) => {
    button.addEventListener("click", async () => {
      const articleId = Number(button.dataset.articleId);
      await loadArticleDetail(articleId);
      showView("detail");
    });
  });
}

function scheduleLookupFromSelection() {
  clearPendingLookup();
  const selectionState = getSelectionState();
  if (!selectionState) {
    hideLookupPopup();
    return;
  }

  state.lookup.timer = window.setTimeout(() => {
    lookupSelection(selectionState);
  }, 500);
}

function clearPendingLookup() {
  if (state.lookup.timer) {
    window.clearTimeout(state.lookup.timer);
    state.lookup.timer = null;
  }
}

function getSelectionState() {
  const selection = window.getSelection();
  if (!selection || selection.isCollapsed || selection.rangeCount === 0) {
    return null;
  }

  const range = selection.getRangeAt(0);
  const text = selection.toString().trim();
  if (!text || !readingContent.contains(range.commonAncestorContainer)) {
    return null;
  }

  const sentenceElement = findSentenceElement(range.commonAncestorContainer);
  if (!sentenceElement) {
    return null;
  }

  const rect = range.getBoundingClientRect();
  return {
    text,
    rect,
    sentenceId: Number(sentenceElement.dataset.sentenceId),
    sentenceText: sentenceElement.dataset.sentenceText || sentenceElement.textContent.trim(),
  };
}

function findSentenceElement(node) {
  let current = node.nodeType === Node.ELEMENT_NODE ? node : node.parentElement;
  while (current) {
    if (current.classList && current.classList.contains("reading-sentence")) {
      return current;
    }
    current = current.parentElement;
  }
  return null;
}

async function lookupSelection(selectionState) {
  positionPopup(selectionState.rect);
  popup.classList.remove("hidden");
  popupTitle.textContent = `查词：${selectionState.text}`;
  popupBody.textContent = "正在查询词典...";
  addVocabularyButton.disabled = true;
  addVocabularyButton.textContent = "加入生词本";

  const lookupKey = `${selectionState.text}:${selectionState.sentenceId}`;
  state.lookup.currentText = selectionState.text;
  state.lookup.currentSentenceId = selectionState.sentenceId;
  state.lookup.currentSentenceText = selectionState.sentenceText;

  if (state.lookup.lastLookupKey === lookupKey && state.lookup.currentEntry) {
    await refreshVocabularyButton(state.lookup.currentEntry.id);
    popupBody.textContent = formatDictionaryEntry(state.lookup.currentEntry, false);
    return;
  }

  const lookupResult = await request(`/api/dictionary/lookup?text=${encodeURIComponent(selectionState.text)}`);
  if (!lookupResult.ok) {
    popupBody.textContent = "词典查询失败。";
    return;
  }

  state.lookup.currentEntry = lookupResult.data.entry;
  state.lookup.lastLookupKey = lookupKey;
  popupBody.textContent = formatDictionaryEntry(lookupResult.data.entry, lookupResult.data.generated);
  await refreshVocabularyButton(lookupResult.data.entry.id);
}

async function refreshVocabularyButton(entryId) {
  const result = await request(`/api/vocabulary/check?dictionary_entry_id=${entryId}`);
  if (!result.ok) {
    addVocabularyButton.disabled = false;
    addVocabularyButton.textContent = "加入生词本";
    return;
  }

  if (result.data.added) {
    addVocabularyButton.disabled = true;
    addVocabularyButton.textContent = "已加入生词本";
  } else {
    addVocabularyButton.disabled = false;
    addVocabularyButton.textContent = "加入生词本";
  }
}

function positionPopup(rect) {
  const top = Math.min(window.innerHeight - 220, rect.bottom + 12);
  const left = Math.min(window.innerWidth - 380, rect.left);
  popupCard.style.top = `${Math.max(12, top)}px`;
  popupCard.style.left = `${Math.max(12, left)}px`;
}

function hideLookupPopup() {
  clearPendingLookup();
  popup.classList.add("hidden");
}

function formatDictionaryEntry(entry, generated) {
  return [
    `词形：${entry.surface}`,
    `原形：${entry.lemma}`,
    `读音：${entry.reading || "-"}`,
    `罗马音：${entry.romaji || "-"}`,
    `词性：${entry.part_of_speech}`,
    `中文释义：${entry.meaning_zh}`,
    `主要释义：${entry.primary_meaning_zh}`,
    `JLPT：${entry.jlpt_level}`,
    `例句：${entry.example_sentence || "-"}`,
    `例句翻译：${entry.example_translation_zh || "-"}`,
    generated ? "说明：当前为占位 AI 词条，后续可替换为真实模型生成结果。" : "",
  ]
    .filter(Boolean)
    .join("\n");
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

function escapeHTMLAttribute(input) {
  return escapeHTML(input).replaceAll("\n", " ");
}

bootstrap();
