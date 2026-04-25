const state = {
  token: localStorage.getItem("access_token") || "",
  user: null,
  selectedArticleId: null,
  selectedVocabularyId: null,
  readingArticle: null,
  vocabularyFilter: "",
  lookup: {
    timer: null,
    currentText: "",
    currentSentenceId: null,
    currentSentenceText: "",
    currentContextSnippet: "",
    currentEntry: null,
    currentGenerated: false,
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
const vocabularyList = document.getElementById("vocabulary-list");
const vocabularyDetail = document.getElementById("vocabulary-detail");
const vocabularyFilterForm = document.getElementById("vocabulary-filter-form");
const vocabularyStatusButtons = document.querySelectorAll("[data-vocabulary-status]");
const deleteVocabularyButton = document.getElementById("delete-vocabulary-button");
const openVocabularyArticleButton = document.getElementById("open-vocabulary-article-button");
const popup = document.getElementById("lookup-popup");
const popupCard = popup.querySelector(".lookup-popup-card");
const popupTitle = document.getElementById("lookup-popup-title");
const popupBody = document.getElementById("lookup-popup-body");
const addVocabularyButton = document.getElementById("add-vocabulary-button");
const openReadingButton = document.getElementById("open-reading-button");
const reprocessButton = document.getElementById("reprocess-button");

document.querySelectorAll("[data-view]").forEach((button) => {
  button.addEventListener("click", async () => {
    const view = button.dataset.view;
    if (view === "vocabulary" && state.user) {
      await loadVocabularyList();
    }
    showView(view);
  });
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
  state.selectedVocabularyId = null;
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
    source_sentence_text: state.lookup.currentContextSnippet || state.lookup.currentSentenceText,
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
  if (state.selectedVocabularyId) {
    await loadVocabularyList();
  }
  setMessage(result.data.created ? "已加入生词本，当前查询上下文已作为例句保存" : "该词已在生词本中");
});

vocabularyFilterForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  state.vocabularyFilter = new FormData(event.currentTarget).get("status") || "";
  await loadVocabularyList();
});

vocabularyStatusButtons.forEach((button) => {
  button.addEventListener("click", async () => {
    if (!state.selectedVocabularyId) {
      setMessage("请先选择一个生词");
      return;
    }
    const result = await request(`/api/vocabulary/${state.selectedVocabularyId}/status`, {
      method: "PUT",
      body: JSON.stringify({ status: button.dataset.vocabularyStatus }),
    });
    if (!result.ok) {
      return;
    }
    await Promise.all([loadVocabularyList(), loadVocabularyDetail(state.selectedVocabularyId)]);
    setMessage(`生词状态已更新为 ${button.dataset.vocabularyStatus}`);
  });
});

deleteVocabularyButton.addEventListener("click", async () => {
  if (!state.selectedVocabularyId) {
    setMessage("请先选择一个生词");
    return;
  }
  const vocabularyId = state.selectedVocabularyId;
  const result = await request(`/api/vocabulary/${vocabularyId}`, {
    method: "DELETE",
  });
  if (!result.ok) {
    return;
  }
  state.selectedVocabularyId = null;
  await loadVocabularyList();
  vocabularyDetail.textContent = "请选择一个生词查看详情。";
  openVocabularyArticleButton.disabled = true;
  setMessage("生词已删除");
});

openVocabularyArticleButton.addEventListener("click", async () => {
  if (!state.selectedVocabularyId) {
    setMessage("请先选择一个生词");
    return;
  }
  const result = await request(`/api/vocabulary/${state.selectedVocabularyId}/context`);
  if (!result.ok) {
    return;
  }
  if (!result.data.article_id) {
    setMessage("这个生词没有关联来源文章");
    return;
  }
  await loadArticleDetail(result.data.article_id);
  showView("detail");
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
      await Promise.all([loadLibrary(), loadArticles(), loadVocabularyList()]);
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
    vocabularyList.innerHTML = "";
    vocabularyDetail.textContent = "请选择一个生词查看详情。";
    openVocabularyArticleButton.disabled = true;
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
  vocabularyFilterForm.elements.status.value = state.vocabularyFilter;
  if (!state.selectedVocabularyId) {
    openVocabularyArticleButton.disabled = true;
  }
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
  Promise.all([loadLibrary(), loadArticles(), loadVocabularyList()]);
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
    "提示：在下方正文中选中文本以查词。加入生词本时会把当前上下文句子或半句一起保存为例句。",
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

async function loadVocabularyList() {
  const suffix = state.vocabularyFilter ? `?status=${encodeURIComponent(state.vocabularyFilter)}` : "";
  const result = await request(`/api/vocabulary${suffix}`);
  if (!result.ok) {
    return;
  }

  const items = result.data.items || [];
  vocabularyList.innerHTML = items
    .map(
      (detail) => `
        <li>
          <button class="link-button vocabulary-item" data-vocabulary-id="${detail.item.id}">
            <span><strong>${escapeHTML(detail.dictionary_entry.surface)}</strong> <span class="tag">${escapeHTML(detail.item.status)}</span></span>
            <span class="meta">${escapeHTML(detail.dictionary_entry.primary_meaning_zh)} / ${escapeHTML(detail.dictionary_entry.jlpt_level)}</span>
            <span class="meta">${escapeHTML(detail.example_sentence || detail.item.source_sentence_text || "-")}</span>
          </button>
        </li>
      `,
    )
    .join("");

  vocabularyList.querySelectorAll("[data-vocabulary-id]").forEach((button) => {
    button.addEventListener("click", async () => {
      const vocabularyId = Number(button.dataset.vocabularyId);
      await loadVocabularyDetail(vocabularyId);
    });
  });
}

async function loadVocabularyDetail(vocabularyId) {
  const result = await request(`/api/vocabulary/${vocabularyId}`);
  if (!result.ok) {
    return;
  }

  state.selectedVocabularyId = vocabularyId;
  const detail = result.data;
  vocabularyDetail.textContent = [
    `词形：${detail.dictionary_entry.surface}`,
    `原形：${detail.dictionary_entry.lemma}`,
    `读音：${detail.dictionary_entry.reading || "-"}`,
    `中文释义：${detail.dictionary_entry.meaning_zh}`,
    `主要释义：${detail.dictionary_entry.primary_meaning_zh}`,
    `当前状态：${detail.item.status}`,
    `来源文章：${detail.article_title || "-"}`,
    `查询原文：${detail.item.selected_text}`,
    "",
    "保存的例句：",
    detail.example_sentence || detail.item.source_sentence_text || "-",
    "",
    "词典自带例句：",
    detail.dictionary_entry.example_sentence || "-",
  ].join("\n");
  openVocabularyArticleButton.disabled = !detail.item.article_id;
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

  const sentenceText = sentenceElement.dataset.sentenceText || sentenceElement.textContent.trim();
  const rect = range.getBoundingClientRect();
  return {
    text,
    rect,
    sentenceId: Number(sentenceElement.dataset.sentenceId),
    sentenceText,
    contextSnippet: extractContextSnippet(sentenceText, text),
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

  const lookupKey = `${selectionState.text}:${selectionState.sentenceId}:${selectionState.contextSnippet}`;
  state.lookup.currentText = selectionState.text;
  state.lookup.currentSentenceId = selectionState.sentenceId;
  state.lookup.currentSentenceText = selectionState.sentenceText;
  state.lookup.currentContextSnippet = selectionState.contextSnippet;

  if (state.lookup.lastLookupKey === lookupKey && state.lookup.currentEntry) {
    await refreshVocabularyButton(state.lookup.currentEntry.id);
    popupBody.textContent = formatDictionaryEntry(state.lookup.currentEntry, state.lookup.currentGenerated, state.lookup.currentContextSnippet);
    return;
  }

  const lookupResult = await request(`/api/dictionary/lookup?text=${encodeURIComponent(selectionState.text)}`);
  if (!lookupResult.ok) {
    popupBody.textContent = "词典查询失败。";
    return;
  }

  state.lookup.currentEntry = lookupResult.data.entry;
  state.lookup.currentGenerated = lookupResult.data.generated;
  state.lookup.lastLookupKey = lookupKey;
  popupBody.textContent = formatDictionaryEntry(lookupResult.data.entry, lookupResult.data.generated, selectionState.contextSnippet);
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

function formatDictionaryEntry(entry, generated, contextSnippet) {
  return [
    `词形：${entry.surface}`,
    `原形：${entry.lemma}`,
    `读音：${entry.reading || "-"}`,
    `罗马音：${entry.romaji || "-"}`,
    `词性：${entry.part_of_speech}`,
    `中文释义：${entry.meaning_zh}`,
    `主要释义：${entry.primary_meaning_zh}`,
    `JLPT：${entry.jlpt_level}`,
    `本次保存例句：${contextSnippet || "-"}`,
    `词典例句：${entry.example_sentence || "-"}`,
    generated ? "说明：当前为占位 AI 词条，后续可替换为真实模型生成结果。" : "",
  ]
    .filter(Boolean)
    .join("\n");
}

function extractContextSnippet(sentenceText, selectedText) {
  const text = String(sentenceText || "").replace(/\s+/g, " ").trim();
  const needle = String(selectedText || "").trim();
  if (!text || !needle) {
    return text;
  }
  if (text.length <= 36) {
    return text;
  }

  const index = text.indexOf(needle);
  if (index < 0) {
    return text;
  }

  const delimiters = new Set(["。", "！", "？", "，", "、", ",", ";", "；"]);
  let start = 0;
  for (let i = index - 1; i >= 0; i -= 1) {
    if (delimiters.has(text[i])) {
      start = i + 1;
      break;
    }
  }

  let end = text.length;
  for (let i = index + needle.length; i < text.length; i += 1) {
    if (delimiters.has(text[i])) {
      end = i + 1;
      break;
    }
  }

  const snippet = text.slice(start, end).trim();
  if (snippet.length >= 8 && snippet.length < text.length) {
    return snippet;
  }
  return text;
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
