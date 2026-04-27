const state = {
  token: localStorage.getItem("access_token") || "",
  user: null,
  selectedArticleId: null,
  selectedVocabularyId: null,
  selectedVocabularyIds: new Set(),
  readingArticle: null,
  vocabularyFilter: "",
  vocabularySearch: "",
  articles: [],
  publicArticles: [],
  articlePage: 1,
  publicArticlePage: 1,
  publicJLPTFilter: "",
  challenge: {
    questions: [],
    currentIndex: 0,
    selectedOption: "",
    answered: false,
  },
  postQuiz: {
    questions: [],
    currentIndex: 0,
    selectedOption: "",
    answered: false,
  },
  readingQuizModalIndex: 0,
  review: {
    items: [],
    currentIndex: 0,
    selectedOption: "",
    answered: false,
    baseItems: [],
    completedCount: 0,
    correctCount: 0,
    wrongCount: 0,
  },
  lookup: {
    timer: null,
    currentText: "",
    currentSentenceId: null,
    currentSentenceText: "",
    currentContextSnippet: "",
    currentEntry: null,
    currentGenerated: false,
    lastLookupKey: "",
    inFlightKey: "",
    drag: null,
  },
  pendingRequests: 0,
  historyReady: false,
  questionGeneration: {
    articleInit: "unknown",
    challenge: "unknown",
    postQuiz: "unknown",
  },
};

const ARTICLE_PAGE_SIZE = 6;

const views = document.querySelectorAll(".view");
const messageBox = document.getElementById("message-box");
const globalLoading = document.getElementById("global-loading");
const authStatus = document.getElementById("auth-status");
const authEntryActions = document.getElementById("auth-entry-actions");
const homeGreeting = document.getElementById("home-greeting");
const profileSummary = document.getElementById("profile-summary");
const learningStats = document.getElementById("learning-stats");
const libraryList = document.getElementById("library-list");
const articleList = document.getElementById("article-list");
const publicArticleList = document.getElementById("public-article-list");
const myArticlePagination = document.getElementById("my-article-pagination");
const publicArticlePagination = document.getElementById("public-article-pagination");
const publicJLPTFilter = document.getElementById("public-jlpt-filter");
const articleDetail = document.getElementById("article-detail");
const sentenceList = document.getElementById("sentence-list");
const readingHeader = document.getElementById("reading-header");
const readingContent = document.getElementById("reading-content");
const challengeHeader = document.getElementById("challenge-header");
const challengeLoading = document.getElementById("challenge-loading");
const challengeCard = document.getElementById("challenge-card");
const challengeProgress = document.getElementById("challenge-progress");
const challengeSentence = document.getElementById("challenge-sentence");
const challengeOptions = document.getElementById("challenge-options");
const challengeFeedback = document.getElementById("challenge-feedback");
const postQuizHeader = document.getElementById("post-quiz-header");
const postQuizCard = document.getElementById("post-quiz-card");
const postQuizProgress = document.getElementById("post-quiz-progress");
const postQuizQuestion = document.getElementById("post-quiz-question");
const postQuizSource = document.getElementById("post-quiz-source");
const postQuizOptions = document.getElementById("post-quiz-options");
const postQuizFeedback = document.getElementById("post-quiz-feedback");
const reviewHeader = document.getElementById("review-header");
const reviewCard = document.getElementById("review-card");
const reviewCompletePanel = document.getElementById("review-complete-panel");
const reviewCompleteSummary = document.getElementById("review-complete-summary");
const extraReviewForm = document.getElementById("extra-review-form");
const reviewProgress = document.getElementById("review-progress");
const reviewQuestion = document.getElementById("review-question");
const reviewContext = document.getElementById("review-context");
const reviewOptions = document.getElementById("review-options");
const reviewFeedback = document.getElementById("review-feedback");
const postQuizResultsList = document.getElementById("post-quiz-results-list");
const reviewRecordsList = document.getElementById("review-records-list");
const vocabularyList = document.getElementById("vocabulary-list");
const vocabularyDetail = document.getElementById("vocabulary-detail");
const vocabularySRSSummary = document.getElementById("vocabulary-srs-summary");
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
const openChallengeButton = document.getElementById("open-challenge-button");
const openPostQuizButton = document.getElementById("open-post-quiz-button");
const openChallengeToolbarButton = document.getElementById("open-challenge-button-toolbar");
const openPostQuizToolbarButton = document.getElementById("open-post-quiz-button-toolbar");
const keyVocabularyList = document.getElementById("key-vocabulary-list");
const readingComprehensionList = document.getElementById("reading-comprehension-list");
const regenerateKeyVocabularyButton = document.getElementById("regenerate-key-vocabulary-button");
const appendReadingQuizButton = document.getElementById("append-reading-quiz-button");
const readingQuizModal = document.getElementById("reading-quiz-modal");
const readingQuizModalBody = document.getElementById("reading-quiz-modal-body");
const readingQuizCloseButton = document.getElementById("reading-quiz-close-button");
const questionGenerationPanel = document.getElementById("question-generation-panel");
const articleInitGenerationBar = document.getElementById("article-init-generation-bar");
const articleInitGenerationStatus = document.getElementById("article-init-generation-status");
const challengeGenerationBar = document.getElementById("challenge-generation-bar");
const challengeGenerationStatus = document.getElementById("challenge-generation-status");
const postQuizGenerationBar = document.getElementById("post-quiz-generation-bar");
const postQuizGenerationStatus = document.getElementById("post-quiz-generation-status");
const submitChallengeAnswerButton = document.getElementById("submit-challenge-answer-button");
const nextChallengeQuestionButton = document.getElementById("next-challenge-question-button");
const submitPostQuizAnswerButton = document.getElementById("submit-post-quiz-answer-button");
const nextPostQuizQuestionButton = document.getElementById("next-post-quiz-question-button");
const submitReviewAnswerButton = document.getElementById("submit-review-answer-button");
const masterReviewWordButton = document.getElementById("master-review-word-button");
const nextReviewQuestionButton = document.getElementById("next-review-question-button");
const loadPostQuizResultsButton = document.getElementById("load-post-quiz-results-button");
const loadReviewRecordsButton = document.getElementById("load-review-records-button");
const completeOnboardingButton = document.getElementById("complete-onboarding-button");
const reprocessButton = document.getElementById("reprocess-button");
const aiConfigForm = document.getElementById("ai-config-form");
const aiProviderSelect = document.getElementById("ai-provider-select");
const aiModelSelect = document.getElementById("ai-model-select");
const aiLoadModelsButton = document.getElementById("ai-load-models-button");
const aiCheckButton = document.getElementById("ai-check-button");
const aiConfigStatus = document.getElementById("ai-config-status");
const vocabularySelectAll = document.getElementById("vocabulary-select-all");
const vocabularySelectedCount = document.getElementById("vocabulary-selected-count");
const batchMasterVocabularyButton = document.getElementById("batch-master-vocabulary-button");
const batchLearningVocabularyButton = document.getElementById("batch-learning-vocabulary-button");
const batchDeleteVocabularyButton = document.getElementById("batch-delete-vocabulary-button");

document.querySelectorAll("[data-view]").forEach((button) => {
  button.addEventListener("click", async () => {
    const view = button.dataset.view;
    if (!state.user && !["login", "register"].includes(view)) {
      setMessage("请先登录或注册");
      showView("login");
      return;
    }
    if (view === "stats" && state.user) {
      await loadLearningStats();
    }
    if (view === "vocabulary" && state.user) {
      await loadVocabularyList();
    }
    if (view === "articles" && state.user) {
      await Promise.all([loadArticles(), loadPublicArticles()]);
    }
    if (view === "review" && state.user) {
      await loadReviewDue();
    }
    if (view === "records" && state.user) {
      await loadLearningRecords();
    }
    if (view === "reading" && state.user && state.selectedArticleId) {
      await loadReadingArticle(state.selectedArticleId);
    }
    if (view === "post-quiz" && state.user && state.selectedArticleId) {
      showView(view);
      await loadPostQuizQuestions(state.selectedArticleId);
    }
    if (view === "profile" && state.user) {
      await loadAIConfig();
    }
    showView(view);
  });
});

completeOnboardingButton.addEventListener("click", async () => {
  const result = await request("/api/profile/onboarding/complete", {
    method: "POST",
    loadingMessage: "正在完成新手引导...",
  });
  if (!result.ok) {
    return;
  }
  state.user = result.data;
  renderUser();
  setMessage("新手引导已完成");
});

document.getElementById("register-form").addEventListener("submit", async (event) => {
  event.preventDefault();
  const payload = Object.fromEntries(new FormData(event.currentTarget).entries());
  const result = await request("/api/auth/register", {
    method: "POST",
    body: JSON.stringify(payload),
    loadingMessage: "正在注册账号...",
  });
  handleAuthResult(result, "注册成功");
});

document.getElementById("login-form").addEventListener("submit", async (event) => {
  event.preventDefault();
  const payload = Object.fromEntries(new FormData(event.currentTarget).entries());
  const result = await request("/api/auth/login", {
    method: "POST",
    body: JSON.stringify(payload),
    loadingMessage: "正在登录...",
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

if (aiConfigForm) {
  aiProviderSelect?.addEventListener("change", () => {
    applyAIProviderDefaults(aiProviderSelect.value);
  });
  aiLoadModelsButton?.addEventListener("click", async () => {
    await loadAIModels();
  });
  aiCheckButton?.addEventListener("click", async () => {
    await checkAIProvider();
  });
  aiConfigForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    await saveAIConfig();
  });
}

async function loadAIConfig() {
  if (!aiConfigForm) {
    return;
  }
  const result = await request("/api/ai/config", {
    loadingMessage: "正在加载 AI 配置...",
  });
  if (!result.ok) {
    return;
  }
  renderAIConfig(result.data);
}

function renderAIConfig(status) {
  if (!aiConfigForm || !status) {
    return;
  }
  setAIFormValue("provider", status.provider || "openai");
  setAIFormValue("provider_name", status.provider_name || "");
  setAIFormValue("base_url", status.base_url || "");
  setAIFormValue("api_key", "");
  setAIFormValue("api_version", status.api_version || "");
  setAIModelOptions(status.model ? [status.model] : [], status.model || "");
  const apiKeyInput = aiConfigForm.elements.api_key;
  if (apiKeyInput) {
    apiKeyInput.placeholder = status.api_key_saved ? "已保存，留空则继续使用原 API Key" : "sk-...";
  }
  aiConfigStatus.innerHTML = formatAIStatus(status, status.configured ? "当前后端 AI 已配置。" : "当前后端未启用 AI，保存配置后生效。");
}

async function loadAIModels() {
  const payload = collectAIConfigPayload();
  const result = await request("/api/ai/models", {
    method: "POST",
    body: JSON.stringify(payload),
    loadingMessage: "正在获取模型列表...",
    timeoutMs: 60000,
  });
  if (!result.ok) {
    if (result.data?.status) {
      aiConfigStatus.innerHTML = formatAIStatus(result.data.status, result.data.error || "获取模型列表失败。");
    }
    return;
  }
  setAIModelOptions(result.data.items || [], payload.model);
  aiConfigStatus.innerHTML = formatAIStatus(result.data.status, `已获取 ${result.data.items?.length || 0} 个模型。`);
  setMessage("模型列表已更新");
}

async function checkAIProvider() {
  const payload = collectAIConfigPayload();
  const result = await request("/api/ai/check", {
    method: "POST",
    body: JSON.stringify(payload),
    loadingMessage: "正在检测 AI 连接...",
    timeoutMs: 60000,
  });
  if (!result.ok) {
    if (result.data?.status) {
      aiConfigStatus.innerHTML = formatAIStatus(result.data.status, result.data.error || "AI 连接检测失败。");
    }
    return;
  }
  aiConfigStatus.innerHTML = formatAIStatus(result.data.status, "AI 连接检测通过。");
  setMessage("AI 连接检测通过");
}

async function saveAIConfig() {
  const payload = collectAIConfigPayload();
  const result = await request("/api/ai/config", {
    method: "PUT",
    body: JSON.stringify(payload),
    loadingMessage: "正在保存 AI 配置...",
  });
  if (!result.ok) {
    return;
  }
  aiConfigStatus.innerHTML = formatAIStatus(result.data.status, "AI 配置已保存并启用。");
  setMessage("AI 配置已保存并启用");
}

function collectAIConfigPayload() {
  const payload = Object.fromEntries(new FormData(aiConfigForm).entries());
  const selectedModel = payload.model || "";
  const manualModel = (payload.model_manual || "").trim();
  return {
    provider: payload.provider,
    provider_name: payload.provider_name,
    base_url: payload.base_url,
    api_key: payload.api_key,
    api_version: payload.api_version,
    model: manualModel || selectedModel,
  };
}

function setAIFormValue(name, value) {
  if (aiConfigForm?.elements[name]) {
    aiConfigForm.elements[name].value = value || "";
  }
}

function setAIModelOptions(models, selected) {
  if (!aiModelSelect) {
    return;
  }
  const uniqueModels = [...new Set((models || []).filter(Boolean))];
  if (selected && !uniqueModels.includes(selected)) {
    uniqueModels.unshift(selected);
  }
  aiModelSelect.innerHTML = uniqueModels.length
    ? uniqueModels.map((model) => `<option value="${escapeHTMLAttribute(model)}">${escapeHTML(model)}</option>`).join("")
    : `<option value="">未获取到模型，可使用手动模型名</option>`;
  aiModelSelect.value = selected && uniqueModels.includes(selected) ? selected : (uniqueModels[0] || "");
}

function applyAIProviderDefaults(provider) {
  const defaults = {
    openai: { name: "OpenAI", baseURL: "https://api.openai.com", model: "gpt-4o-mini", apiVersion: "" },
    "openai-responses": { name: "OpenAI Responses", baseURL: "https://api.openai.com", model: "gpt-4o-mini", apiVersion: "" },
    gemini: { name: "Gemini", baseURL: "https://generativelanguage.googleapis.com", model: "gemini-1.5-flash", apiVersion: "" },
    anthropic: { name: "Anthropic", baseURL: "https://api.anthropic.com", model: "claude-3-5-haiku-latest", apiVersion: "" },
    "azure-openai": { name: "Azure OpenAI", baseURL: "https://{resource}.openai.azure.com", model: "", apiVersion: "2024-10-21" },
    "new-api": { name: "New API", baseURL: "", model: "gpt-4o-mini", apiVersion: "" },
  };
  const current = defaults[provider] || defaults.openai;
  setAIFormValue("provider_name", current.name);
  setAIFormValue("base_url", current.baseURL);
  setAIFormValue("api_version", current.apiVersion);
  setAIModelOptions(current.model ? [current.model] : [], current.model);
  setAIFormValue("model_manual", "");
}

function formatAIStatus(status, message) {
  if (!status) {
    return escapeHTML(message || "-");
  }
  const keyStatus = status.api_key_saved ? "已保存，留空保存不会覆盖" : "未保存";
  return `
    <strong>${escapeHTML(message || "-")}</strong>
    <div class="meta">供应商：${escapeHTML(status.provider_name || status.provider || "-")} · 类型：${escapeHTML(status.provider || "-")} · 模型：${escapeHTML(status.model || "-")}</div>
    <div class="meta">API Key：${escapeHTML(keyStatus)}</div>
    <div class="meta">调用地址：${escapeHTML(status.endpoint || "-")}</div>
    <div class="meta">模型地址：${escapeHTML(status.models_endpoint || "-")}</div>
  `;
}

document.getElementById("jlpt-form").addEventListener("submit", async (event) => {
  event.preventDefault();
  const payload = Object.fromEntries(new FormData(event.currentTarget).entries());
  const result = await request("/api/profile/jlpt-level", {
    method: "PUT",
    body: JSON.stringify(payload),
    loadingMessage: "正在保存 JLPT 等级...",
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
  showQuestionGenerationPanel(true);
  setQuestionGenerationStatus("articleInit", "generating");
  setQuestionGenerationStatus("challenge", "unknown");
  setQuestionGenerationStatus("postQuiz", "unknown");
  state.selectedArticleId = null;
  state.readingArticle = null;
  hideLookupPopup();
  readingHeader.textContent = "文章正在上传并初始化...";
  readingContent.innerHTML = `<div class="empty-state">正在处理文章内容，完成后会自动进入阅读。</div>`;
  if (keyVocabularyList) {
    keyVocabularyList.innerHTML = `<span class="meta">等待文章初始化完成后生成 JLPT 考点重点。</span>`;
  }
  if (readingComprehensionList) {
    readingComprehensionList.innerHTML = `<span class="meta">等待文章初始化完成后生成阅读理解题。</span>`;
  }
  showView("reading");
  const result = await request("/api/articles/upload", {
    method: "POST",
    body: JSON.stringify(payload),
    loadingMessage: "正在上传并处理文章...",
    timeoutMs: 60000,
  });
  if (!result.ok) {
    setQuestionGenerationStatus("articleInit", "failed");
    return;
  }
  setQuestionGenerationStatus("articleInit", "ready");
  event.currentTarget.reset();
  state.selectedArticleId = result.data.id;
  await loadReadingArticle(result.data.id);
  void Promise.all([loadArticles(), loadPublicArticles()]);
  setMessage("文章已创建并处理完成");
});

reprocessButton.addEventListener("click", async () => {
  if (!state.selectedArticleId) {
    setMessage("请先选择一篇文章");
    return;
  }
  const result = await request(`/api/articles/${state.selectedArticleId}/process`, {
    method: "POST",
    loadingMessage: "正在重新处理文章...",
    timeoutMs: 60000,
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

openChallengeButton?.addEventListener("click", async () => {
  if (!state.selectedArticleId) {
    setMessage("请先选择一篇文章");
    return;
  }
  if (state.questionGeneration.challenge !== "ready") {
    setMessage("重点词推荐还没有生成完成，请等待进度条完成。");
    return;
  }
  showView("reading");
  await loadReadingArticle(state.selectedArticleId);
});

openPostQuizButton?.addEventListener("click", async () => {
  if (!state.selectedArticleId) {
    setMessage("请先选择一篇文章");
    return;
  }
  if (state.questionGeneration.postQuiz !== "ready") {
    setMessage("阅读理解题还没有生成完成，请等待进度条完成。");
    return;
  }
  await loadPostQuizQuestions(state.selectedArticleId);
  showView("reading");
  openReadingQuizModal(0);
});

readingQuizCloseButton?.addEventListener("click", () => {
  closeReadingQuizModal();
});

readingQuizModal?.addEventListener("mousedown", (event) => {
  if (event.target === readingQuizModal) {
    closeReadingQuizModal();
  }
});

regenerateKeyVocabularyButton?.addEventListener("click", async () => {
  if (!state.selectedArticleId) {
    setMessage("请先选择一篇文章");
    return;
  }
  regenerateKeyVocabularyButton.disabled = true;
  setQuestionGenerationStatus("challenge", "generating");
  if (keyVocabularyList) {
    keyVocabularyList.innerHTML = `<span class="meta">正在重新按 JLPT 考点分析重点词汇和语法...</span>`;
  }
  await generateQuestionSet(state.selectedArticleId, "challenge", { refresh: true });
  regenerateKeyVocabularyButton.disabled = false;
});

appendReadingQuizButton?.addEventListener("click", async () => {
  if (!state.selectedArticleId) {
    setMessage("请先选择一篇文章");
    return;
  }
  appendReadingQuizButton.disabled = true;
  setQuestionGenerationStatus("postQuiz", "generating");
  if (readingComprehensionList) {
    readingComprehensionList.insertAdjacentHTML("beforeend", `<span class="meta" data-quiz-appending>正在追加 JLPT 阅读理解题...</span>`);
  }
  await generateQuestionSet(state.selectedArticleId, "postQuiz", { append: true });
  document.querySelector("[data-quiz-appending]")?.remove();
  appendReadingQuizButton.disabled = false;
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
    loadingMessage: "正在加入生词本...",
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
  const form = new FormData(event.currentTarget);
  state.vocabularyFilter = form.get("status") || "";
  state.vocabularySearch = form.get("q") || "";
  await loadVocabularyList();
});

publicJLPTFilter?.addEventListener("click", async (event) => {
  const button = event.target.closest("[data-public-jlpt]");
  if (!button) {
    return;
  }
  state.publicJLPTFilter = button.dataset.publicJlpt || "";
  state.publicArticlePage = 1;
  publicJLPTFilter.querySelectorAll("[data-public-jlpt]").forEach((item) => item.classList.toggle("active", item === button));
  renderPublicArticles();
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
      loadingMessage: "正在更新生词状态...",
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
  if (!window.confirm("确定要删除这个生词吗？")) {
    return;
  }
  const vocabularyId = state.selectedVocabularyId;
  const result = await request(`/api/vocabulary/${vocabularyId}`, {
    method: "DELETE",
    loadingMessage: "正在删除生词...",
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

vocabularySelectAll?.addEventListener("change", () => {
  const checked = vocabularySelectAll.checked;
  document.querySelectorAll(".vocabulary-select-checkbox").forEach((checkbox) => {
    checkbox.checked = checked;
    const id = Number(checkbox.dataset.vocabularyId);
    if (checked) {
      state.selectedVocabularyIds.add(id);
    } else {
      state.selectedVocabularyIds.delete(id);
    }
  });
  updateVocabularyBatchState();
});

batchMasterVocabularyButton?.addEventListener("click", async () => {
  await batchUpdateVocabularyStatus("mastered", "熟悉");
});

batchLearningVocabularyButton?.addEventListener("click", async () => {
  await batchUpdateVocabularyStatus("learning", "学习中");
});

batchDeleteVocabularyButton?.addEventListener("click", async () => {
  const ids = getSelectedVocabularyIds();
  if (ids.length === 0) {
    setMessage("请先选择要删除的生词");
    return;
  }
  if (!window.confirm(`确定要删除选中的 ${ids.length} 个生词吗？`)) {
    return;
  }
  const result = await request("/api/vocabulary/batch/delete", {
    method: "POST",
    body: JSON.stringify({ vocabulary_ids: ids }),
    loadingMessage: "正在批量删除生词...",
  });
  if (!result.ok) {
    return;
  }
  state.selectedVocabularyIds.clear();
  state.selectedVocabularyId = null;
  vocabularyDetail.textContent = "请选择一个生词查看详情。";
  await loadVocabularyList();
  setMessage(`已删除 ${result.data.deleted || 0} 个生词`);
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

submitChallengeAnswerButton.addEventListener("click", async () => {
  const question = state.challenge.questions[state.challenge.currentIndex];
  if (!question) {
    setMessage("当前没有可作答的题目");
    return;
  }
  if (!state.challenge.selectedOption) {
    setMessage("请先选择一个选项");
    return;
  }
  if (state.challenge.answered) {
    setMessage("这题已经提交过了");
    return;
  }

  const result = await request(`/api/reading/questions/${question.id}/answer`, {
    method: "POST",
    body: JSON.stringify({ selected_option: state.challenge.selectedOption }),
    loadingMessage: "正在提交答案...",
  });
  if (!result.ok) {
    return;
  }

  state.challenge.answered = true;
  challengeFeedback.classList.remove("hidden");
  challengeFeedback.textContent = [
    result.data.is_correct ? "回答正确" : "回答错误",
    `正确选项：${result.data.correct_option}`,
    `正确答案：${result.data.correct_answer_text}`,
    `解析：${result.data.explanation}`,
  ].join("\n");
  renderChallengeQuestion();
});

nextChallengeQuestionButton.addEventListener("click", () => {
  if (state.challenge.currentIndex + 1 >= state.challenge.questions.length) {
    setMessage("挑战阅读已完成");
    return;
  }
  state.challenge.currentIndex += 1;
  state.challenge.selectedOption = "";
  state.challenge.answered = false;
  challengeFeedback.classList.add("hidden");
  challengeFeedback.textContent = "";
  renderChallengeQuestion();
});

submitPostQuizAnswerButton.addEventListener("click", async () => {
  const question = state.postQuiz.questions[state.postQuiz.currentIndex];
  if (!question) {
    setMessage("当前没有可作答的测验题");
    return;
  }
  if (!state.postQuiz.selectedOption) {
    setMessage("请先选择一个选项");
    return;
  }
  if (state.postQuiz.answered) {
    setMessage("这题已经提交过了");
    return;
  }

  const result = await request(`/api/reading/questions/${question.id}/answer`, {
    method: "POST",
    body: JSON.stringify({ selected_option: state.postQuiz.selectedOption }),
    loadingMessage: "正在提交测验答案...",
  });
  if (!result.ok) {
    return;
  }

  state.postQuiz.answered = true;
  postQuizFeedback.classList.remove("hidden");
  postQuizFeedback.textContent = [
    result.data.is_correct ? "回答正确" : "回答错误",
    `正确选项：${result.data.correct_option}`,
    `正确答案：${result.data.correct_answer_text}`,
    `解析：${result.data.explanation}`,
  ].join("\n");
  renderPostQuizQuestion();
});

nextPostQuizQuestionButton.addEventListener("click", () => {
  if (state.postQuiz.currentIndex + 1 >= state.postQuiz.questions.length) {
    setMessage("阅读后测验已完成");
    return;
  }
  state.postQuiz.currentIndex += 1;
  state.postQuiz.selectedOption = "";
  state.postQuiz.answered = false;
  postQuizFeedback.classList.add("hidden");
  postQuizFeedback.textContent = "";
  renderPostQuizQuestion();
});

submitReviewAnswerButton.addEventListener("click", async () => {
  await submitCurrentReviewAnswer();
});

async function submitCurrentReviewAnswer() {
  const item = state.review.items[state.review.currentIndex];
  if (!item) {
    setMessage("当前没有可作答的复习题");
    return;
  }
  if (!state.review.selectedOption) {
    setMessage("请先选择一个选项");
    return;
  }
  if (state.review.answered) {
    setMessage("这题已经提交过了");
    return;
  }

  const result = await request("/api/review/answer", {
    method: "POST",
    body: JSON.stringify({
      user_vocabulary_id: item.user_vocabulary.id,
      review_question_id: item.question.id,
      selected_option: state.review.selectedOption,
    }),
    loadingMessage: "正在提交复习答案...",
  });
  if (!result.ok) {
    return;
  }

  state.review.answered = true;
  state.review.completedCount += 1;
  if (result.data.is_correct) {
    state.review.correctCount += 1;
  } else {
    state.review.wrongCount += 1;
    insertExtraWrongReview(item);
  }
  reviewFeedback.classList.remove("hidden");
  const progressNote = result.data.is_correct
    ? Number(result.data.familiarity_delta || 0) > 0
      ? `熟练度 +${result.data.familiarity_delta}%`
      : "这个词今天熟练度涨幅已到 40%，继续学习不会再增加熟练度。"
    : "";
  reviewFeedback.textContent = [
    result.data.is_correct ? "回答正确" : "回答错误",
    `正确选项：${result.data.correct_option}`,
    `正确答案：${result.data.correct_answer}`,
    progressNote,
    `当前状态：${vocabularyStatusLabel(result.data.status)} · 熟练度 ${result.data.proficiency ?? result.data.familiarity ?? 0}%`,
    `下次复习：${result.data.status === "mastered" ? "已熟悉，不再进入每日复习" : formatDateTime(result.data.next_review_at)}`,
    `解析：${result.data.explanation}`,
  ].join("\n");
  await loadVocabularyList();
  renderReviewQuestion();
}

function insertExtraWrongReview(item) {
  const sameWordPending = state.review.items
    .slice(state.review.currentIndex + 1)
    .filter((candidate) => candidate.user_vocabulary.id === item.user_vocabulary.id).length;
  const targetTotal = Math.min(4, Math.max(Number(item.planned_rounds || 1), sameWordPending + 2));
  const currentTotal = sameWordPending + 1;
  if (currentTotal >= targetTotal) {
    return;
  }
  const nextRound = currentTotal + 1;
  const insertAt = Math.min(state.review.items.length, state.review.currentIndex + 6 + sameWordPending * 6);
  state.review.items.splice(insertAt, 0, {
    ...item,
    review_round: nextRound,
    planned_rounds: targetTotal,
    question_loaded_for_turn: false,
    question_refreshing: false,
  });
}

nextReviewQuestionButton.addEventListener("click", () => {
  moveToNextReviewQuestion();
});

masterReviewWordButton?.addEventListener("click", async () => {
  const item = state.review.items[state.review.currentIndex];
  if (!item) {
    setMessage("当前没有可标记的复习词");
    return;
  }
  const result = await request(`/api/vocabulary/${item.user_vocabulary.id}/status`, {
    method: "PUT",
    body: JSON.stringify({ status: "mastered" }),
    loadingMessage: "正在标记熟悉...",
  });
  if (!result.ok) {
    return;
  }
  state.review.items.splice(state.review.currentIndex, 1);
  if (state.review.currentIndex >= state.review.items.length) {
    state.review.currentIndex = Math.max(0, state.review.items.length - 1);
  }
  state.review.selectedOption = "";
  state.review.answered = false;
  reviewFeedback.classList.add("hidden");
  reviewFeedback.textContent = "";
  await loadVocabularyList();
  setMessage("已标记熟悉，后续复习会跳过这个词");
  if (state.review.items.length === 0) {
    reviewHeader.textContent = "今日复习完成，熟悉词已移出复习队列。";
    showReviewCompletePanel();
    return;
  }
  renderReviewQuestion();
});

loadPostQuizResultsButton.addEventListener("click", async () => {
  await loadPostQuizResults();
});

loadReviewRecordsButton.addEventListener("click", async () => {
  await loadReviewRecords();
});

extraReviewForm?.addEventListener("submit", async (event) => {
  event.preventDefault();
  const limit = Number(new FormData(event.currentTarget).get("limit") || 10);
  await loadReviewDue(Number.isFinite(limit) ? Math.max(1, Math.min(50, limit)) : 10, true);
});

readingContent.addEventListener("mouseup", () => {
  scheduleLookupFromSelection();
});

challengeSentence?.addEventListener("mouseup", () => {
  scheduleLookupFromSelection();
});

postQuizQuestion.addEventListener("mouseup", () => {
  scheduleLookupFromSelection();
});

document.addEventListener("mousedown", (event) => {
  if (!popupCard.contains(event.target)) {
    hideLookupPopup();
  }
});

popupTitle.addEventListener("mousedown", (event) => {
  if (popup.classList.contains("hidden")) {
    return;
  }
  const rect = popupCard.getBoundingClientRect();
  state.lookup.drag = {
    offsetX: event.clientX - rect.left,
    offsetY: event.clientY - rect.top,
  };
  popupCard.classList.add("dragging");
  event.preventDefault();
});

document.addEventListener("mousemove", (event) => {
  if (!state.lookup.drag) {
    return;
  }
  const width = popupCard.offsetWidth || 360;
  const height = popupCard.offsetHeight || 260;
  const left = clamp(event.clientX - state.lookup.drag.offsetX, 12, window.innerWidth - width - 12);
  const top = clamp(event.clientY - state.lookup.drag.offsetY, 12, window.innerHeight - height - 12);
  popupCard.style.left = `${left}px`;
  popupCard.style.top = `${top}px`;
});

document.addEventListener("mouseup", () => {
  state.lookup.drag = null;
  popupCard.classList.remove("dragging");
});

document.addEventListener("selectionchange", () => {
  const selection = window.getSelection();
  if (!selection || selection.isCollapsed) {
    clearPendingLookup();
  }
});

window.addEventListener("popstate", (event) => {
  const view = event.state?.view || "home";
  showView(view, { push: false });
});

async function bootstrap() {
  if (state.token) {
    const me = await request("/api/auth/me");
    if (me.ok) {
      state.user = me.data;
      await Promise.all([loadLibrary(), loadArticles(), loadPublicArticles(), loadVocabularyList()]);
    } else {
      localStorage.removeItem("access_token");
      state.token = "";
    }
  }
  renderUser();
}

function showView(name) {
  views.forEach((view) => view.classList.toggle("active", view.id === `view-${name}`));
  document.querySelectorAll("[data-view]").forEach((button) => {
    button.classList.toggle("active", button.dataset.view === name && button.classList.contains("nav-item"));
  });
  if (!state.historyReady) {
    window.history.replaceState({ view: name }, "", window.location.href);
    state.historyReady = true;
    return;
  }
  if (arguments[1]?.push === false) {
    return;
  }
  if (window.history.state?.view !== name) {
    window.history.pushState({ view: name }, "", window.location.href);
  }
}

function renderUser() {
  if (!state.user) {
    authStatus.textContent = "未登录";
    authEntryActions?.classList.remove("hidden");
    document.getElementById("logout-button").classList.add("hidden");
    homeGreeting.textContent = "登录后可查看文章库、上传文章并进入处理流程。";
    profileSummary.textContent = "尚未加载资料。";
    learningStats.textContent = "请先登录后查看学习统计。";
    libraryList.innerHTML = "";
    articleList.innerHTML = "";
    if (publicArticleList) {
      publicArticleList.innerHTML = "";
    }
    articleDetail.textContent = "请选择一篇文章。";
    sentenceList.innerHTML = "";
    readingHeader.textContent = "请选择一篇文章进入阅读。";
    readingContent.innerHTML = "";
    if (challengeHeader) {
      challengeHeader.textContent = "请选择一篇文章开始挑战阅读。";
    }
    challengeCard?.classList.add("hidden");
    postQuizHeader.textContent = "请选择一篇文章开始测验。";
    postQuizCard.classList.add("hidden");
    if (keyVocabularyList) {
      keyVocabularyList.innerHTML = `<span class="meta">等待 AI 分析文章。</span>`;
    }
    if (readingComprehensionList) {
      readingComprehensionList.innerHTML = `<span class="meta">等待 AI 生成题目。</span>`;
    }
    reviewHeader.textContent = "加载今日待复习生词。";
    reviewCard.classList.add("hidden");
    postQuizResultsList.innerHTML = "";
    reviewRecordsList.innerHTML = "";
    vocabularyList.innerHTML = "";
    vocabularyDetail.textContent = "请选择一个生词查看详情。";
    openVocabularyArticleButton.disabled = true;
    showView("login");
    return;
  }

  authStatus.textContent = `已登录：${state.user.username}（${state.user.email}）`;
  authEntryActions?.classList.add("hidden");
  document.getElementById("logout-button").classList.remove("hidden");
  homeGreeting.textContent = `欢迎回来，${state.user.username}。当前 JLPT：${state.user.jlpt_level}`;
  document.querySelectorAll("[data-user-jlpt]").forEach((node) => {
    node.textContent = state.user.jlpt_level || "N5";
  });
  profileSummary.innerHTML = `
    <div class="article-card">
      <span class="article-card-title">${escapeHTML(state.user.username)}</span>
      <span class="meta">${escapeHTML(state.user.email)}</span>
      <span class="badge badge-jlpt">${escapeHTML(state.user.jlpt_level || "-")}</span>
      <span class="meta">新手引导：${state.user.onboarding_completed ? "已完成" : "未完成"}</span>
    </div>
  `;
  document.querySelector('#jlpt-form select[name="jlpt_level"]').value = state.user.jlpt_level;
  vocabularyFilterForm.elements.status.value = state.vocabularyFilter;
  if (vocabularyFilterForm.elements.q) {
    vocabularyFilterForm.elements.q.value = state.vocabularySearch;
  }
  if (!state.selectedVocabularyId) {
    openVocabularyArticleButton.disabled = true;
  }
  showView(state.user.onboarding_completed ? "home" : "onboarding");
}

function handleAuthResult(result, successMessage) {
  if (!result.ok) {
    return;
  }
  state.token = result.data.access_token;
  state.user = result.data.user;
  localStorage.setItem("access_token", state.token);
  renderUser();
  Promise.all([loadLibrary(), loadArticles(), loadPublicArticles(), loadVocabularyList()]);
  setMessage(successMessage);
}

async function loadLearningRecords() {
  postQuizResultsList.innerHTML = `<li class="empty-state">选择文章后可查看阅读后测验记录。</li>`;
  await loadReviewRecords();
  if (state.selectedArticleId) {
    await loadPostQuizResults();
  }
}

async function loadLearningStats() {
  learningStats.textContent = "正在加载学习统计。";
  const result = await request("/api/stats/learning");
  if (!result.ok) {
    return;
  }

  const stats = result.data;
  const statusCounts = stats.vocabulary_status_counts || {};
  const readingRate = percent(stats.reading_correct_count, stats.reading_attempt_count);
  const reviewRate = percent(stats.review_correct_count, stats.review_record_count);
  const totalStatus = Object.values(statusCounts).reduce((sum, value) => sum + Number(value || 0), 0) || 1;
  learningStats.innerHTML = `
    <div class="article-grid">
      ${statCard("我的文章", stats.article_count)}
      ${statCard("生词总数", stats.vocabulary_count)}
      ${statCard("今日待复习", stats.due_vocabulary_count)}
      ${statCard("阅读正确率", readingRate)}
      ${statCard("复习正确率", reviewRate)}
      ${statCard("复习次数", stats.review_record_count)}
    </div>
    <div class="card" style="margin-top:18px">
      <h3>生词状态分布</h3>
      ${["new", "learning", "mastered"].map((key) => {
        const count = key === "learning" ? Number(statusCounts.learning || 0) + Number(statusCounts.reviewing || 0) : statusCounts[key] || 0;
        const width = Math.max(4, Math.round((count / totalStatus) * 100));
        return `<div class="stat-row"><span class="tag status-${key}">${vocabularyStatusLabel(key)}</span><div class="progress-bar"><span style="width:${width}%"></span></div><strong>${count}</strong></div>`;
      }).join("")}
    </div>
    <div class="card" style="margin-top:18px">
      <h3>学习建议</h3>
      <p class="muted">${stats.due_vocabulary_count > 0 ? "今日有到期生词，建议先完成词汇复习。" : "今天的到期生词已清空，可以继续阅读添加新词。"}</p>
    </div>
  `;
}

async function loadPostQuizResults() {
  if (!state.selectedArticleId) {
    postQuizResultsList.innerHTML = `<li class="empty-state">请先在文章详情中选择一篇文章。</li>`;
    return;
  }

  postQuizResultsList.innerHTML = `<li class="empty-state">正在加载阅读后测验记录...</li>`;
  const result = await request(`/api/reading/articles/${state.selectedArticleId}/post-quiz/results`);
  if (!result.ok) {
    return;
  }

  const items = result.data.items || [];
  if (items.length === 0) {
    postQuizResultsList.innerHTML = `<li class="empty-state">当前文章还没有阅读后测验答题记录。</li>`;
    return;
  }

  postQuizResultsList.innerHTML = items
    .map((item) => {
      const question = item.question;
      const attempt = item.attempt;
      return `
        <li>
          <div class="record-item">
            <strong>${attempt.is_correct ? "正确" : "错误"} · ${escapeHTML(question.correct_answer_text)}</strong>
            <span class="meta">选择：${escapeHTML(attempt.selected_option)} / 正确：${escapeHTML(question.correct_option)} / ${formatDateTime(attempt.answered_at)}</span>
            <span class="meta">${escapeHTML(question.masked_sentence)}</span>
          </div>
        </li>
      `;
    })
    .join("");
}

async function loadReviewRecords() {
  reviewRecordsList.innerHTML = `<li class="empty-state">正在加载词汇复习记录...</li>`;
  const result = await request("/api/review/records?limit=20");
  if (!result.ok) {
    return;
  }

  const items = result.data.items || [];
  if (items.length === 0) {
    reviewRecordsList.innerHTML = `<li class="empty-state">还没有词汇复习记录。</li>`;
    return;
  }

  reviewRecordsList.innerHTML = items
    .map((item) => {
      const record = item.record;
      const entry = item.dictionary_entry;
      const question = item.question;
      return `
        <li>
          <div class="record-item">
            <strong>${escapeHTML(entry.surface)} · ${record.is_correct ? "正确" : "错误"}</strong>
            <span class="meta">选择：${escapeHTML(record.selected_option)} / 正确：${escapeHTML(question.correct_option)} / ${formatDateTime(record.reviewed_at)}</span>
            <span class="meta">${escapeHTML(entry.primary_meaning_zh)} · ${escapeHTML(item.context_sentence || "-")}</span>
          </div>
        </li>
      `;
    })
    .join("");
}

async function loadLibrary() {
  libraryList.innerHTML = `<li class="empty-state">正在加载内置文章...</li>`;
  const result = await request("/api/articles/library");
  if (!result.ok) {
    return;
  }
  const items = result.data.items || [];
  if (items.length === 0) {
    libraryList.innerHTML = `<li class="empty-state">暂无内置文章。</li>`;
    return;
  }
  libraryList.innerHTML = items
    .map(
      (article) => `
        <li>
          <button class="article-card" data-article-id="${article.id}">
            <span class="article-card-title">${escapeHTML(article.title)}</span>
            <span><span class="badge badge-jlpt">${escapeHTML(article.jlpt_level || "-")}</span> <span class="tag">${escapeHTML(article.translation_status || "-")}</span></span>
            <span class="meta">${article.sentence_count || 0} 句 · 内置文章</span>
          </button>
        </li>
      `,
    )
    .join("");

  bindArticleSelection(libraryList);
}

async function loadArticles() {
  articleList.innerHTML = `<li class="empty-state">正在加载我的文章...</li>`;
  const result = await request("/api/articles");
  if (!result.ok) {
    return;
  }
  state.articles = result.data.items || [];
  state.articlePage = clampPage(state.articlePage, state.articles.length);
  renderMyArticles();
}

function renderMyArticles() {
  const items = paginateItems(state.articles, state.articlePage);
  if (items.length === 0) {
    articleList.innerHTML = `<li class="empty-state">还没有上传文章。可以先去“上传文章”创建第一篇。</li>`;
    renderPagination(myArticlePagination, 0, state.articlePage, () => {});
    return;
  }
  articleList.innerHTML = items
    .map(
      (article) => `
        <li>
          <button class="article-card" data-article-id="${article.id}">
            <span class="article-card-title">${escapeHTML(article.title)}</span>
            <span><span class="tag">${escapeHTML(article.source_type || "mine")}</span> <span class="tag">${escapeHTML(article.translation_status || "-")}</span></span>
            <span class="meta">${escapeHTML(article.original_language || "-")} · ${article.sentence_count || 0} 句 · ${formatDateTime(article.updated_at || article.created_at)}</span>
            <span class="meta">${escapeHTML(article.chinese_translation || "点击查看详情、阅读或重新处理。")}</span>
          </button>
        </li>
      `,
    )
    .join("");

  bindArticleSelection(articleList);
  renderPagination(myArticlePagination, state.articles.length, state.articlePage, (page) => {
    state.articlePage = page;
    renderMyArticles();
  });
}

async function loadPublicArticles() {
  if (!publicArticleList) {
    return;
  }
  publicArticleList.innerHTML = `<li class="empty-state">正在加载公共文章...</li>`;
  const result = await request("/api/articles/public");
  if (!result.ok) {
    return;
  }
  state.publicArticles = result.data.items || [];
  state.publicArticlePage = clampPage(state.publicArticlePage, filteredPublicArticles().length);
  renderPublicArticles();
}

function renderPublicArticles() {
  if (!publicArticleList) {
    return;
  }
  const filteredItems = filteredPublicArticles();
  const items = paginateItems(filteredItems, state.publicArticlePage);
  if (items.length === 0) {
    publicArticleList.innerHTML = `<li class="empty-state">当前筛选下没有公共文章。</li>`;
    renderPagination(publicArticlePagination, 0, state.publicArticlePage, () => {});
    return;
  }
  publicArticleList.innerHTML = items
    .map(
      (article) => `
        <li>
          <button class="article-card" data-article-id="${article.id}">
            <span class="article-card-title">${escapeHTML(article.title)}</span>
            <span><span class="tag">${escapeHTML(article.source_type || "-")}</span> <span class="badge badge-jlpt">${escapeHTML(article.jlpt_level || "-")}</span></span>
            <span class="meta">${article.sentence_count || 0} 句 · ${formatDateTime(article.updated_at || article.created_at)}</span>
            <span class="meta">${escapeHTML(article.chinese_translation || "这是一篇可直接学习的公共日语阅读文章。")}</span>
          </button>
        </li>
      `,
    )
    .join("");
  bindArticleSelection(publicArticleList);
  renderPagination(publicArticlePagination, filteredItems.length, state.publicArticlePage, (page) => {
    state.publicArticlePage = page;
    renderPublicArticles();
  });
}

function filteredPublicArticles() {
  const level = state.publicJLPTFilter;
  return level ? state.publicArticles.filter((article) => article.jlpt_level === level) : state.publicArticles;
}

function paginateItems(items, page) {
  const start = (page - 1) * ARTICLE_PAGE_SIZE;
  return items.slice(start, start + ARTICLE_PAGE_SIZE);
}

function clampPage(page, itemCount) {
  return Math.max(1, Math.min(page || 1, Math.max(1, Math.ceil(itemCount / ARTICLE_PAGE_SIZE))));
}

function renderPagination(container, itemCount, currentPage, onPageChange) {
  if (!container) {
    return;
  }
  const totalPages = Math.ceil(itemCount / ARTICLE_PAGE_SIZE);
  if (totalPages <= 1) {
    container.innerHTML = "";
    return;
  }
  container.innerHTML = `
    <button class="btn btn-ghost compact" type="button" data-page="prev" ${currentPage <= 1 ? "disabled" : ""}>上一页</button>
    <span class="meta">第 ${currentPage} / ${totalPages} 页</span>
    <button class="btn btn-ghost compact" type="button" data-page="next" ${currentPage >= totalPages ? "disabled" : ""}>下一页</button>
  `;
  container.querySelectorAll("[data-page]").forEach((button) => {
    button.addEventListener("click", () => {
      const nextPage = button.dataset.page === "prev" ? currentPage - 1 : currentPage + 1;
      onPageChange(clampPage(nextPage, itemCount));
    });
  });
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
  articleDetail.innerHTML = `
    <div class="article-card">
      <span class="article-card-title">${escapeHTML(article.title)}</span>
      <span>
        <span class="badge badge-jlpt">${escapeHTML(article.jlpt_level || "-")}</span>
        <span class="tag">${escapeHTML(article.translation_status || "-")}</span>
        <span class="tag">${escapeHTML(article.source_type || "-")}</span>
      </span>
      <span class="meta">原文语言：${escapeHTML(article.original_language || "-")} · 句子数量：${article.sentence_count || 0}</span>
      <span class="meta">处理说明：${escapeHTML(article.processing_notes || "-")}</span>
    </div>
    <div class="summary" style="margin-top:14px"><strong>中文翻译</strong><br>${escapeHTML(article.chinese_translation || "-")}</div>
    <div class="summary" style="margin-top:14px"><strong>日语内容</strong><br>${escapeHTML(article.japanese_content || article.original_content || "-")}</div>
  `;

  sentenceList.innerHTML = (sentenceResult.data.items || [])
    .map((sentence, index) => `<li><span class="tag">${index + 1}</span><span>${escapeHTML(sentence.sentence_text)}</span></li>`)
    .join("");
  if (!sentenceList.innerHTML) {
    sentenceList.innerHTML = `<li class="empty-state">当前文章还没有句子拆分结果，可以尝试重新处理。</li>`;
  }
  sentenceList.classList.add("hidden");

  reprocessButton.disabled = article.source_type === "builtin";
  reprocessButton.title = article.source_type === "builtin" ? "内置文章无需重新处理" : "";
  await refreshQuestionGenerationReadiness(articleId);
}

async function loadReadingArticle(articleId) {
  readingContent.innerHTML = `<div class="empty-state">正在加载阅读内容...</div>`;
  const result = await request(`/api/reading/articles/${articleId}`);
  if (!result.ok) {
    return;
  }

  const { article } = result.data;
  state.selectedArticleId = article.id;
  state.readingArticle = article;
  hideLookupPopup();
  readingHeader.innerHTML = `${escapeHTML(article.title)} <span class="badge badge-jlpt">${escapeHTML(article.jlpt_level || "-")}</span>`;

  const text = article.japanese_content || article.original_content || "";
  readingContent.innerHTML = renderReadingArticleText(text);
  if (!readingContent.innerHTML) {
    readingContent.innerHTML = `<div class="empty-state">当前文章没有可阅读句子。</div>`;
  }
  await refreshQuestionGenerationReadiness(articleId);
}

async function refreshQuestionGenerationReadiness(articleId) {
  if (!articleId) {
    return;
  }
  showQuestionGenerationPanel(true);
  setQuestionGenerationStatus("articleInit", "ready");
  setQuestionGenerationStatus("challenge", "checking");
  setQuestionGenerationStatus("postQuiz", "checking");
  const [challengeResult, postQuizResult] = await Promise.all([
    request(`/api/reading/articles/${articleId}/challenge-questions`, { silent: true }),
    request(`/api/reading/articles/${articleId}/post-quiz`, { silent: true }),
  ]);
  setQuestionGenerationStatus("challenge", challengeResult.ok && (challengeResult.data.items || []).length > 0 ? "ready" : "pending");
  setQuestionGenerationStatus("postQuiz", postQuizResult.ok && (postQuizResult.data.items || []).length > 0 ? "ready" : "pending");
  if (challengeResult.ok) {
    state.challenge.questions = challengeResult.data.items || [];
    renderKeyVocabularyRecommendations();
  }
  if (postQuizResult.ok) {
    state.postQuiz.questions = postQuizResult.data.items || [];
    renderReadingComprehensionList();
  }
  if (!challengeResult.ok || (challengeResult.data.items || []).length === 0) {
    setQuestionGenerationStatus("challenge", "generating");
    if (keyVocabularyList) {
      keyVocabularyList.innerHTML = `<span class="meta">尚未生成 JLPT 考点重点，正在自动生成...</span>`;
    }
    void generateQuestionSet(articleId, "challenge");
  }
  if (!postQuizResult.ok || (postQuizResult.data.items || []).length === 0) {
    setQuestionGenerationStatus("postQuiz", "generating");
    if (readingComprehensionList) {
      readingComprehensionList.innerHTML = `<span class="meta">尚未生成阅读理解题，正在自动生成...</span>`;
    }
    void generateQuestionSet(articleId, "postQuiz");
  }
  hideQuestionGenerationPanelIfReady();
}

function startArticleQuestionGeneration(articleId) {
  if (!articleId) {
    return;
  }
  showQuestionGenerationPanel(true);
  state.challenge.questions = [];
  state.postQuiz.questions = [];
  if (keyVocabularyList) {
    keyVocabularyList.innerHTML = `<span class="meta">AI 正在按 JLPT 考点分析重点词汇和语法，生成后会自动显示在文章底部。</span>`;
  }
  if (readingComprehensionList) {
    readingComprehensionList.innerHTML = `<span class="meta">AI 正在生成阅读理解题，生成后会自动显示。</span>`;
  }
  setQuestionGenerationStatus("challenge", "generating");
  setQuestionGenerationStatus("postQuiz", "generating");
  void generateQuestionSet(articleId, "challenge");
  void generateQuestionSet(articleId, "postQuiz");
}

async function generateQuestionSet(articleId, type, options = {}) {
  const query = new URLSearchParams();
  if (options.refresh) {
    query.set("refresh", "1");
  }
  if (options.append) {
    query.set("append", "1");
  }
  const baseEndpoint =
    type === "challenge"
      ? `/api/reading/articles/${articleId}/challenge-questions`
      : `/api/reading/articles/${articleId}/post-quiz`;
  const endpoint = query.toString() ? `${baseEndpoint}?${query.toString()}` : baseEndpoint;
  const result = await request(endpoint, {
    method: "POST",
    timeoutMs: 120000,
    silent: true,
  });
  if (state.selectedArticleId !== articleId) {
    return;
  }
  setQuestionGenerationStatus(type, result.ok && (result.data.items || []).length > 0 ? "ready" : "failed");
  if (result.ok) {
    if (type === "challenge") {
      state.challenge.questions = result.data.items || [];
      renderKeyVocabularyRecommendations();
    } else {
      state.postQuiz.questions = result.data.items || [];
      renderReadingComprehensionList();
    }
  }
  hideQuestionGenerationPanelIfReady();
}

function showQuestionGenerationPanel(visible) {
  questionGenerationPanel?.classList.toggle("hidden", !visible);
}

function hideQuestionGenerationPanelIfReady() {
  if (state.questionGeneration.articleInit !== "generating" && state.questionGeneration.challenge === "ready" && state.questionGeneration.postQuiz === "ready") {
    window.setTimeout(() => showQuestionGenerationPanel(false), 800);
  }
}

function setQuestionGenerationStatus(type, status) {
  state.questionGeneration[type] = status;
  const bar = type === "articleInit" ? articleInitGenerationBar : type === "challenge" ? challengeGenerationBar : postQuizGenerationBar;
  const label = type === "articleInit" ? articleInitGenerationStatus : type === "challenge" ? challengeGenerationStatus : postQuizGenerationStatus;
  const statusMeta = {
    checking: ["检测中", 35],
    pending: ["未生成", 0],
    generating: ["生成中", 65],
    ready: ["已完成", 100],
    failed: ["生成失败", 0],
    unknown: ["等待生成", 0],
  }[status] || ["等待生成", 0];
  if (bar) {
    bar.style.width = `${statusMeta[1]}%`;
  }
  if (label) {
    label.textContent = statusMeta[0];
  }
  updateQuestionActionButtons();
}

function updateQuestionActionButtons() {
  const postQuizReady = state.questionGeneration.postQuiz === "ready";
  [openChallengeButton, openChallengeToolbarButton].forEach((button) => {
    if (button) {
      button.disabled = true;
      button.title = "挑战阅读已改为阅读页右侧的重点词汇和语法推荐";
    }
  });
  [openPostQuizButton, openPostQuizToolbarButton].forEach((button) => {
    if (button) {
      button.disabled = !postQuizReady;
      button.title = postQuizReady ? "" : "阅读理解题生成完成后会显示在右侧";
    }
  });
}

function renderKeyVocabularyRecommendations() {
  if (!keyVocabularyList) {
    return;
  }
  const items = state.challenge.questions || [];
  if (items.length === 0) {
    keyVocabularyList.innerHTML = `<span class="meta">暂无 JLPT 考点重点词汇或语法推荐。</span>`;
    return;
  }
  const vocabularyItems = items.filter((item) => recommendationKind(item) !== "grammar").slice(0, 5);
  const grammarItems = items.filter((item) => recommendationKind(item) === "grammar").slice(0, 5);
  const renderGroup = (title, groupItems, emptyText) => `
    <div class="recommendation-group">
      <span class="meta strong-meta">${title}</span>
      ${
        groupItems.length === 0
          ? `<span class="meta">${emptyText}</span>`
          : groupItems
              .map(
                (item) => `
        <div class="key-vocab-card">
          <strong>${escapeHTML(item.correct_answer_text || item.masked_sentence || "-")}</strong>
          <div class="key-vocab-meta">
            <span class="badge badge-jlpt">${escapeHTML(item.option_a || item.jlpt_level || "unknown")}</span>
            <span class="tag">频次 ${escapeHTML(item.option_b || "-")}</span>
            <span class="tag">考点重要度 ${escapeHTML(item.option_c || "-")}</span>
            <span class="tag">${recommendationKind(item) === "grammar" ? "文法" : "词汇"}</span>
          </div>
          <span class="meta">${escapeHTML(item.explanation || item.option_d || "")}</span>
          <button class="btn btn-secondary compact" data-add-key-vocab="${item.id}">加入生词本</button>
        </div>
      `,
              )
              .join("")
      }
    </div>
  `;
  keyVocabularyList.innerHTML = [
    renderGroup("重点词汇", vocabularyItems, "暂无重点词汇推荐。"),
    renderGroup("重点语法 / 固定用法", grammarItems, "暂无重点语法推荐。"),
  ].join("");
  keyVocabularyList.querySelectorAll("[data-add-key-vocab]").forEach((button) => {
    button.addEventListener("click", async () => {
      const item = items.find((candidate) => String(candidate.id) === button.dataset.addKeyVocab);
      if (!item) {
        return;
      }
      const result = await request("/api/vocabulary", {
        method: "POST",
        body: JSON.stringify({
          dictionary_entry_id: item.correct_entry_id,
          article_id: state.selectedArticleId,
          source_sentence_id: item.sentence_id,
          selected_text: item.correct_answer_text || item.masked_sentence,
          source_sentence_text: item.sentence_text,
        }),
        loadingMessage: "正在加入生词本...",
      });
      if (result.ok) {
        button.disabled = true;
        button.textContent = "已加入";
        setMessage(result.data.created ? "重点词已加入生词本" : "该词已在生词本中");
      }
    });
  });
}

function recommendationKind(item) {
  const raw = `${item.option_d || ""} ${item.masked_sentence || ""} ${item.correct_answer_text || ""}`.toLowerCase();
  if (raw.includes("grammar") || raw.includes("文法") || raw.includes("语法") || raw.includes("固定用法")) {
    return "grammar";
  }
  return "vocabulary";
}

function renderReadingComprehensionList() {
  if (!readingComprehensionList) {
    return;
  }
  const items = state.postQuiz.questions || [];
  if (items.length === 0) {
    readingComprehensionList.innerHTML = `<span class="meta">暂无阅读理解题。</span>`;
    return;
  }
  readingComprehensionList.innerHTML = items
    .map((item, index) => {
      const selected = getCachedReadingQuizOption(item.id);
      return `
        <button class="reading-comprehension-item ${selected ? "answered" : ""}" data-reading-quiz-index="${index}">
          <strong>第 ${index + 1} 题</strong>
          <span>${escapeHTML(item.masked_sentence)}</span>
          ${selected ? `<span class="meta">已选择：${escapeHTML(selected)}</span>` : `<span class="meta">点击作答</span>`}
        </button>
      `;
    })
    .join("");
  readingComprehensionList.querySelectorAll("[data-reading-quiz-index]").forEach((button) => {
    button.addEventListener("click", () => {
      openReadingQuizModal(Number(button.dataset.readingQuizIndex));
    });
  });
}

function openReadingQuizModal(index) {
  const question = state.postQuiz.questions[index];
  if (!question || !readingQuizModal || !readingQuizModalBody) {
    return;
  }
  state.readingQuizModalIndex = index;
  const selectedOption = getCachedReadingQuizOption(question.id);
  const submitted = getCachedReadingQuizSubmitted(question.id);
  const options = [
    ["A", question.option_a],
    ["B", question.option_b],
    ["C", question.option_c],
    ["D", question.option_d],
  ];
  readingQuizModalBody.innerHTML = `
    <div class="progress-label">阅读理解 · 第 ${index + 1} / ${state.postQuiz.questions.length} 题</div>
    <h3>${escapeHTML(question.masked_sentence)}</h3>
    <p class="muted">${escapeHTML(question.sentence_text || "")}</p>
    <div class="challenge-options">
      ${options
        .map(([key, value]) => {
          const selected = selectedOption === key;
          const isCorrect = submitted && question.correct_option === key;
          const isIncorrect = submitted && selected && question.correct_option !== key;
          const className = ["challenge-option", "review-option", selected ? "selected" : "", isCorrect ? "correct" : "", isIncorrect ? "incorrect" : ""].filter(Boolean).join(" ");
          return `
            <label class="${className}">
              <input type="radio" name="reading-quiz-option" value="${key}" ${selected ? "checked" : ""} ${submitted ? "disabled" : ""} />
              <span class="option-key">${key}</span>
              <span class="option-value">${escapeHTML(value)}</span>
            </label>
          `;
        })
        .join("")}
    </div>
    <div id="reading-quiz-modal-feedback" class="summary ${submitted ? "" : "hidden"}">
      ${submitted ? renderReadingQuizFeedback(question, selectedOption) : ""}
    </div>
    <div class="detail-actions">
      <button id="reading-quiz-submit-button" class="btn btn-primary" ${!selectedOption || submitted ? "disabled" : ""}>提交答案</button>
    </div>
  `;
  readingQuizModal.classList.remove("hidden");
  readingQuizModalBody.querySelectorAll('input[name="reading-quiz-option"]').forEach((input) => {
    input.addEventListener("change", () => {
      setCachedReadingQuizOption(question.id, input.value);
      openReadingQuizModal(index);
    });
  });
  document.getElementById("reading-quiz-submit-button")?.addEventListener("click", async () => {
    const option = getCachedReadingQuizOption(question.id);
    if (!option) {
      setMessage("请先选择一个选项");
      return;
    }
    const result = await request(`/api/reading/questions/${question.id}/answer`, {
      method: "POST",
      body: JSON.stringify({ selected_option: option }),
      loadingMessage: "正在提交阅读理解答案...",
    });
    if (!result.ok) {
      return;
    }
    setCachedReadingQuizSubmitted(question.id, result.data);
    openReadingQuizModal(index);
    renderReadingComprehensionList();
  });
}

function closeReadingQuizModal() {
  readingQuizModal?.classList.add("hidden");
  renderReadingComprehensionList();
}

function readingQuizStorageKey(questionID, field) {
  return `reading_quiz:${state.selectedArticleId || "article"}:${questionID}:${field}`;
}

function getCachedReadingQuizOption(questionID) {
  return localStorage.getItem(readingQuizStorageKey(questionID, "option")) || "";
}

function setCachedReadingQuizOption(questionID, option) {
  localStorage.setItem(readingQuizStorageKey(questionID, "option"), option);
}

function getCachedReadingQuizSubmitted(questionID) {
  const raw = localStorage.getItem(readingQuizStorageKey(questionID, "submitted"));
  if (!raw) {
    return null;
  }
  try {
    return JSON.parse(raw);
  } catch {
    return null;
  }
}

function setCachedReadingQuizSubmitted(questionID, result) {
  localStorage.setItem(readingQuizStorageKey(questionID, "submitted"), JSON.stringify(result));
}

function renderReadingQuizFeedback(question, selectedOption) {
  const submitted = getCachedReadingQuizSubmitted(question.id);
  const isCorrect = submitted ? submitted.is_correct : question.correct_option === selectedOption;
  return [
    isCorrect ? "回答正确" : "回答错误",
    `正确选项：${escapeHTML(question.correct_option)}`,
    `正确答案：${escapeHTML(question.correct_answer_text)}`,
    `解析：${escapeHTML(question.explanation || "-")}`,
  ].join("\n");
}

async function loadChallengeQuestions(articleId) {
  if (challengeHeader) {
    challengeHeader.textContent = "正在加载重点词汇和语法推荐...";
  }
  challengeLoading?.classList.remove("hidden");
  challengeCard?.classList.add("hidden");
  const result = await request(`/api/reading/articles/${articleId}/challenge-questions`, { timeoutMs: 60000 });
  challengeLoading?.classList.add("hidden");
  if (!result.ok) {
    if (keyVocabularyList) {
      keyVocabularyList.innerHTML = `<span class="meta">重点词推荐加载失败，请检查 AI 配置。</span>`;
    }
    return;
  }

  state.challenge.questions = result.data.items || [];
  renderKeyVocabularyRecommendations();
}

async function loadPostQuizQuestions(articleId) {
  postQuizHeader.textContent = "正在加载阅读后测验题...";
  postQuizCard.classList.add("hidden");
  const result = await request(`/api/reading/articles/${articleId}/post-quiz`, { timeoutMs: 60000 });
  if (!result.ok) {
    return;
  }

  state.postQuiz.questions = result.data.items || [];
  state.postQuiz.currentIndex = 0;
  state.postQuiz.selectedOption = "";
  state.postQuiz.answered = false;
  hideLookupPopup();
  renderReadingComprehensionList();

  if (state.postQuiz.questions.length === 0) {
    postQuizHeader.textContent = "当前文章还没有可用的阅读理解题，请等待上传后的题目生成任务完成。";
    postQuizCard.classList.add("hidden");
    return;
  }

  postQuizHeader.textContent = "阅读理解题会基于文章主旨、细节和句间关系出题。";
  postQuizCard.classList.remove("hidden");
  postQuizFeedback.classList.add("hidden");
  postQuizFeedback.textContent = "";
  renderPostQuizQuestion();
}

function renderChallengeQuestion() {
  const question = state.challenge.questions[state.challenge.currentIndex];
  if (!question) {
    challengeCard.classList.add("hidden");
    return;
  }

  challengeProgress.textContent = `第 ${state.challenge.currentIndex + 1} / ${state.challenge.questions.length} 题`;
  challengeProgress.nextElementSibling?.querySelector("span")?.style.setProperty("width", `${Math.round(((state.challenge.currentIndex + 1) / state.challenge.questions.length) * 100)}%`);
  challengeSentence.dataset.sentenceId = question.sentence_id;
  challengeSentence.dataset.sentenceText = question.sentence_text;
  challengeSentence.textContent = question.masked_sentence;

  const options = [
    ["A", question.option_a],
    ["B", question.option_b],
    ["C", question.option_c],
    ["D", question.option_d],
  ];
  challengeOptions.innerHTML = options
    .map(([key, value]) => {
      const selected = state.challenge.selectedOption === key;
      const isCorrect = state.challenge.answered && question.correct_option === key;
      const isIncorrect = state.challenge.answered && selected && question.correct_option !== key;
      const className = ["challenge-option", "review-option", selected ? "selected" : "", isCorrect ? "correct" : "", isIncorrect ? "incorrect" : ""].filter(Boolean).join(" ");
      return `
        <label class="${className}">
          <input type="radio" name="challenge-option" value="${key}" ${selected ? "checked" : ""} ${state.challenge.answered ? "disabled" : ""} />
          <span class="option-key">${key}</span>
          <span class="option-value">${escapeHTML(value)}</span>
        </label>
      `;
    })
    .join("");

  challengeOptions.querySelectorAll('input[name="challenge-option"]').forEach((input) => {
    input.addEventListener("change", () => {
      state.challenge.selectedOption = input.value;
      renderChallengeQuestion();
    });
  });

  submitChallengeAnswerButton.disabled = state.challenge.answered;
  nextChallengeQuestionButton.disabled = !state.challenge.answered;
}

function renderPostQuizQuestion() {
  const question = state.postQuiz.questions[state.postQuiz.currentIndex];
  if (!question) {
    postQuizCard.classList.add("hidden");
    return;
  }

  postQuizProgress.textContent = `第 ${state.postQuiz.currentIndex + 1} / ${state.postQuiz.questions.length} 题`;
  postQuizProgress.nextElementSibling?.querySelector("span")?.style.setProperty("width", `${Math.round(((state.postQuiz.currentIndex + 1) / state.postQuiz.questions.length) * 100)}%`);
  postQuizQuestion.dataset.sentenceId = question.sentence_id;
  postQuizQuestion.dataset.sentenceText = question.sentence_text;
  postQuizQuestion.textContent = question.masked_sentence;
  postQuizSource.textContent = `原句：${question.sentence_text}`;

  const options = [
    ["A", question.option_a],
    ["B", question.option_b],
    ["C", question.option_c],
    ["D", question.option_d],
  ];
  postQuizOptions.innerHTML = options
    .map(([key, value]) => {
      const selected = state.postQuiz.selectedOption === key;
      const isCorrect = state.postQuiz.answered && question.correct_option === key;
      const isIncorrect = state.postQuiz.answered && selected && question.correct_option !== key;
      const className = ["challenge-option", "review-option", selected ? "selected" : "", isCorrect ? "correct" : "", isIncorrect ? "incorrect" : ""].filter(Boolean).join(" ");
      return `
        <label class="${className}">
          <input type="radio" name="post-quiz-option" value="${key}" ${selected ? "checked" : ""} ${state.postQuiz.answered ? "disabled" : ""} />
          <span class="option-key">${key}</span>
          <span class="option-value">${escapeHTML(value)}</span>
        </label>
      `;
    })
    .join("");

  postQuizOptions.querySelectorAll('input[name="post-quiz-option"]').forEach((input) => {
    input.addEventListener("change", () => {
      state.postQuiz.selectedOption = input.value;
      renderPostQuizQuestion();
    });
  });

  submitPostQuizAnswerButton.disabled = state.postQuiz.answered;
  nextPostQuizQuestionButton.disabled = !state.postQuiz.answered;
}

async function loadReviewDue(limit = 20, extra = false) {
  reviewHeader.textContent = extra ? "正在加载追加学习生词..." : "正在加载今日待复习生词...";
  reviewCard.classList.add("hidden");
  reviewCompletePanel?.classList.add("hidden");
  const params = new URLSearchParams({ limit: String(limit) });
  if (extra) {
    params.set("extra", "1");
  }
  const result = await request(`/api/review/due?${params.toString()}`, { timeoutMs: 60000 });
  if (!result.ok) {
    return;
  }

  state.review.baseItems = result.data.items || [];
  state.review.items = buildDailyReviewQueue(state.review.baseItems);
  state.review.currentIndex = 0;
  state.review.selectedOption = "";
  state.review.answered = false;
  state.review.completedCount = 0;
  state.review.correctCount = 0;
  state.review.wrongCount = 0;

  if (state.review.items.length === 0) {
    const emptyMessage = extra ? "当前没有更多可追加学习的生词。" : "当前没有到期需要复习的生词。";
    reviewHeader.textContent = emptyMessage;
    reviewCard.classList.add("hidden");
    showReviewCompletePanel(emptyMessage);
    return;
  }

  reviewHeader.textContent = `${extra ? "追加学习" : "今日任务"}：${state.review.baseItems.length} 个词 · ${state.review.items.length} 轮`;
  const allProgressCapped = state.review.baseItems.every((item) => Number(item.today_progress_gain || 0) >= 40);
  if (extra && allProgressCapped) {
    setMessage("这些词今天熟练度涨幅都已达到 40%，继续学习会保留练习记录，但不会再增加熟练度。");
  }
  reviewCard.classList.remove("hidden");
  reviewFeedback.classList.add("hidden");
  reviewFeedback.textContent = "";
  renderReviewQuestion();
}

function buildDailyReviewQueue(items) {
  const firstRound = [];
  const laterRounds = [];
  items.forEach((item) => {
    const rounds = plannedReviewRounds(item);
    firstRound.push({ ...item, review_round: 1, planned_rounds: rounds, question_loaded_for_turn: false });
    for (let round = 2; round <= rounds; round += 1) {
      laterRounds.push({ ...item, review_round: round, planned_rounds: rounds, question_loaded_for_turn: false });
    }
  });
  return interleaveReviewRounds(firstRound, laterRounds);
}

function plannedReviewRounds(item) {
  const status = visibleVocabularyStatus(item.user_vocabulary.status);
  const wrongToday = Number(item.today_wrong_count || 0);
  const consecutive = Number(item.user_vocabulary.consecutive_correct_count || 0);
  if (status === "new") return 3;
  if (wrongToday > 0) return 4;
  if (consecutive >= 3) return 1;
  if (consecutive >= 1) return 2;
  return 2;
}

function interleaveReviewRounds(firstRound, laterRounds) {
  const queue = [...firstRound];
  laterRounds.forEach((item, index) => {
    const offset = item.review_round === 2 ? 5 : item.review_round === 3 ? 15 : 25;
    const position = Math.min(queue.length, index + offset);
    queue.splice(position, 0, item);
  });
  return queue;
}

function renderReviewQuestion() {
  const item = state.review.items[state.review.currentIndex];
  if (!item) {
    reviewCard.classList.add("hidden");
    return;
  }
  refreshReviewQuestionForTurn(item);

  const question = item.question;
  reviewProgress.textContent = `第 ${state.review.currentIndex + 1} / ${state.review.items.length} 轮 · ${escapeHTML(item.dictionary.surface)} 第 ${item.review_round || 1} 轮`;
  reviewQuestion.textContent = question.question_text;
  reviewContext.textContent = item.context_sentence ? `上下文：${item.context_sentence}` : "";

  const options = [
    ["A", question.option_a],
    ["B", question.option_b],
    ["C", question.option_c],
    ["D", question.option_d],
  ];
  reviewOptions.innerHTML = options
    .map(([key, value]) => {
      const selected = state.review.selectedOption === key;
      const isCorrect = state.review.answered && question.correct_option === key;
      const isIncorrect = state.review.answered && selected && question.correct_option !== key;
      const className = ["challenge-option", "review-option", selected ? "selected" : "", isCorrect ? "correct" : "", isIncorrect ? "incorrect" : ""].filter(Boolean).join(" ");
      return `
        <label class="${className}">
          <input type="radio" name="review-option" value="${key}" ${selected ? "checked" : ""} ${state.review.answered ? "disabled" : ""} />
          <span class="option-key">${key}</span>
          <span class="option-value">${escapeHTML(value)}</span>
        </label>
      `;
    })
    .join("");

  reviewOptions.querySelectorAll('input[name="review-option"]').forEach((input) => {
    input.addEventListener("change", async () => {
      state.review.selectedOption = input.value;
      renderReviewQuestion();
      if (!state.review.answered) {
        await submitCurrentReviewAnswer();
      }
    });
  });

  submitReviewAnswerButton.disabled = state.review.answered;
  nextReviewQuestionButton.disabled = !state.review.answered;
  masterReviewWordButton.disabled = false;
}

async function refreshReviewQuestionForTurn(item) {
  if (!item || item.question_loaded_for_turn || item.question_refreshing || state.review.answered) {
    return;
  }
  item.question_refreshing = true;
  const result = await request(`/api/review/question?dictionary_entry_id=${encodeURIComponent(item.dictionary_entry.id)}`, { timeoutMs: 60000 });
  item.question_refreshing = false;
  item.question_loaded_for_turn = true;
  if (!result.ok || !result.data.question) {
    return;
  }
  item.question = result.data.question;
  if (state.review.items[state.review.currentIndex] === item && !state.review.answered) {
    renderReviewQuestion();
  }
}

function moveToNextReviewQuestion() {
  if (state.review.currentIndex + 1 >= state.review.items.length) {
    showReviewCompletePanel();
    return;
  }
  state.review.currentIndex += 1;
  state.review.selectedOption = "";
  state.review.answered = false;
  reviewFeedback.classList.add("hidden");
  reviewFeedback.textContent = "";
  renderReviewQuestion();
}

function showReviewCompletePanel(emptyMessage = "当前没有到期任务，可以输入数量继续学习更多词。") {
  reviewCard.classList.add("hidden");
  reviewCompletePanel?.classList.remove("hidden");
  const total = state.review.completedCount || 0;
  const correct = state.review.correctCount || 0;
  const wrong = state.review.wrongCount || 0;
  const rate = total ? Math.round((correct / total) * 100) : 0;
  reviewHeader.textContent = "今日任务已完成";
  if (reviewCompleteSummary) {
    reviewCompleteSummary.textContent = total
      ? `本次完成 ${total} 轮，答对 ${correct} 轮，答错 ${wrong} 轮，正确率 ${rate}%。`
      : emptyMessage;
  }
}

async function batchUpdateVocabularyStatus(status, label) {
  const ids = getSelectedVocabularyIds();
  if (ids.length === 0) {
    setMessage("请先选择要操作的生词");
    return;
  }
  const result = await request("/api/vocabulary/batch/status", {
    method: "POST",
    body: JSON.stringify({ vocabulary_ids: ids, status }),
    loadingMessage: `正在批量标记为${label}...`,
  });
  if (!result.ok) {
    return;
  }
  state.selectedVocabularyIds.clear();
  await loadVocabularyList();
  if (state.selectedVocabularyId) {
    await loadVocabularyDetail(state.selectedVocabularyId).catch(() => {});
  }
  setMessage(`已将 ${result.data.updated || 0} 个生词标记为${label}`);
}

function getSelectedVocabularyIds() {
  return [...state.selectedVocabularyIds].filter((id) => Number.isFinite(id) && id > 0);
}

function updateVocabularyBatchState() {
  const count = state.selectedVocabularyIds.size;
  if (vocabularySelectedCount) {
    vocabularySelectedCount.textContent = `已选择 ${count} 个`;
  }
  [batchMasterVocabularyButton, batchLearningVocabularyButton, batchDeleteVocabularyButton].forEach((button) => {
    if (button) {
      button.disabled = count === 0;
    }
  });
  if (vocabularySelectAll) {
    const checkboxes = [...document.querySelectorAll(".vocabulary-select-checkbox")];
    vocabularySelectAll.checked = checkboxes.length > 0 && checkboxes.every((checkbox) => checkbox.checked);
    vocabularySelectAll.indeterminate = checkboxes.some((checkbox) => checkbox.checked) && !vocabularySelectAll.checked;
  }
}

function vocabularyStatusLabel(status) {
  const labels = {
    new: "新词",
    learning: "学习中",
    reviewing: "学习中",
    mastered: "熟悉",
    ignored: "忽略",
  };
  return escapeHTML(labels[status] || status || "-");
}

function visibleVocabularyStatus(status) {
  return status === "reviewing" ? "learning" : status;
}

function proficiencyPercent(item) {
  const value = Number(item?.familiarity || 0);
  if (value <= 5) {
    return Math.max(0, Math.min(100, value * 20));
  }
  return Math.max(0, Math.min(100, value));
}

function renderProficiencyBar(value) {
  return `
    <div class="proficiency-row">
      <span>熟练度</span>
      <div class="proficiency-bar"><span style="width:${value}%"></span></div>
      <strong>${value}%</strong>
    </div>
  `;
}

function renderVocabularySRSSummary(items) {
  if (!vocabularySRSSummary) {
    return;
  }
  const summary = items.reduce(
    (acc, detail) => {
      const status = visibleVocabularyStatus(detail.item.status);
      if (status === "new") acc.newCount++;
      if (status === "learning") acc.learningCount++;
      if (status === "mastered") acc.masteredCount++;
      if (status !== "mastered" && status !== "ignored" && new Date(detail.item.next_review_at) <= new Date()) acc.dueCount++;
      return acc;
    },
    { newCount: 0, learningCount: 0, masteredCount: 0, dueCount: 0 },
  );
  vocabularySRSSummary.innerHTML = `
    <div><strong>${summary.dueCount}</strong><span>今日待学/复习</span></div>
    <div><strong>${summary.newCount}</strong><span>新词</span></div>
    <div><strong>${summary.learningCount}</strong><span>学习中</span></div>
    <div><strong>${summary.masteredCount}</strong><span>熟悉</span></div>
  `;
}

async function loadVocabularyList() {
  const params = new URLSearchParams();
  if (state.vocabularyFilter) {
    params.set("status", state.vocabularyFilter);
  }
  if (state.vocabularySearch) {
    params.set("q", state.vocabularySearch);
  }
  const suffix = params.toString() ? `?${params.toString()}` : "";
  vocabularyList.innerHTML = `<li class="empty-state">正在加载生词本...</li>`;
  await request("/api/review/prewarm", {
    method: "POST",
    timeoutMs: 60000,
  });
  const result = await request(`/api/vocabulary${suffix}`);
  if (!result.ok) {
    return;
  }

  const items = result.data.items || [];
  renderVocabularySRSSummary(items);
  state.selectedVocabularyIds = new Set([...state.selectedVocabularyIds].filter((id) => items.some((detail) => detail.item.id === id)));
  updateVocabularyBatchState();
  if (items.length === 0) {
    vocabularyList.innerHTML = `<li class="empty-state">当前筛选或搜索下没有生词。阅读文章时框选词语即可加入。</li>`;
    return;
  }
  vocabularyList.innerHTML = items
    .map(
      (detail) => `
        <li>
          <div class="vocabulary-row">
            <label class="vocabulary-check">
              <input class="vocabulary-select-checkbox" type="checkbox" data-vocabulary-id="${detail.item.id}" ${state.selectedVocabularyIds.has(detail.item.id) ? "checked" : ""} />
            </label>
            <button class="link-button vocabulary-item" data-vocabulary-id="${detail.item.id}">
              <span><strong class="vocab-surface">${escapeHTML(detail.dictionary_entry.surface)}</strong> <span class="tag status-${escapeHTML(visibleVocabularyStatus(detail.item.status))}">${vocabularyStatusLabel(detail.item.status)}</span></span>
              <span class="meta">${escapeHTML(detail.dictionary_entry.reading || "-")} · ${escapeHTML(detail.dictionary_entry.romaji || "-")} · ${escapeHTML(detail.dictionary_entry.jlpt_level)}</span>
              <span>${escapeHTML(detail.dictionary_entry.primary_meaning_zh)}</span>
              ${renderProficiencyBar(proficiencyPercent(detail.item))}
              <span class="meta">${escapeHTML(detail.example_sentence || detail.item.source_sentence_text || "-")}</span>
            </button>
          </div>
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
  vocabularyList.querySelectorAll(".vocabulary-select-checkbox").forEach((checkbox) => {
    checkbox.addEventListener("change", () => {
      const id = Number(checkbox.dataset.vocabularyId);
      if (checkbox.checked) {
        state.selectedVocabularyIds.add(id);
      } else {
        state.selectedVocabularyIds.delete(id);
      }
      updateVocabularyBatchState();
    });
  });
  updateVocabularyBatchState();
}

async function loadVocabularyDetail(vocabularyId) {
  const result = await request(`/api/vocabulary/${vocabularyId}`);
  if (!result.ok) {
    return;
  }

  state.selectedVocabularyId = vocabularyId;
  const detail = result.data;
  const proficiency = proficiencyPercent(detail.item);
  vocabularyDetail.innerHTML = `
    <div class="vocab-surface">${escapeHTML(detail.dictionary_entry.surface)}</div>
    <p class="meta">${escapeHTML(detail.dictionary_entry.lemma || "-")} · ${escapeHTML(detail.dictionary_entry.reading || "-")} · ${escapeHTML(detail.dictionary_entry.romaji || "-")}</p>
    <p><span class="tag">${escapeHTML(detail.dictionary_entry.part_of_speech || "-")}</span> <span class="badge badge-jlpt">${escapeHTML(detail.dictionary_entry.jlpt_level || "-")}</span> <span class="tag status-${escapeHTML(visibleVocabularyStatus(detail.item.status))}">${vocabularyStatusLabel(detail.item.status)}</span></p>
    ${renderProficiencyBar(proficiency)}
    <p class="meta">正确 ${detail.item.correct_count || 0} 次 · 错误 ${detail.item.wrong_count || 0} 次 · 连续正确 ${detail.item.consecutive_correct_count || 0} 次</p>
    <p class="meta">下次复习：${visibleVocabularyStatus(detail.item.status) === "mastered" ? "已熟悉，不再进入每日复习" : formatDateTime(detail.item.next_review_at)}</p>
    <p><strong>中文释义</strong><br>${escapeHTML(detail.dictionary_entry.meaning_zh || detail.dictionary_entry.primary_meaning_zh || "-")}</p>
    <p><strong>上下文</strong><br>${escapeHTML(detail.example_sentence || detail.item.source_sentence_text || "-")}</p>
    <p class="meta">来源文章：${escapeHTML(detail.article_title || "-")} · 查询原文：${escapeHTML(detail.item.selected_text || "-")}</p>
    <p class="meta">词典例句：${escapeHTML(detail.dictionary_entry.example_sentence || "-")}</p>
    <div class="dictionary-examples-panel">
      <div class="detail-actions">
        <strong>AI 例句</strong>
        <button class="btn btn-secondary compact" data-generate-example="${detail.dictionary_entry.id}" type="button">AI 生成例句</button>
      </div>
      <div id="dictionary-example-list" class="dictionary-example-list">正在加载例句...</div>
    </div>
  `;
  openVocabularyArticleButton.disabled = !detail.item.article_id;
  vocabularyDetail.querySelector("[data-generate-example]")?.addEventListener("click", async () => {
    await generateDictionaryExample(detail.dictionary_entry.id);
  });
  await loadDictionaryExamples(detail.dictionary_entry.id);
}

async function loadDictionaryExamples(entryId) {
  const container = document.getElementById("dictionary-example-list");
  if (!container) {
    return;
  }
  const result = await request(`/api/dictionary/${entryId}/examples`);
  if (!result.ok) {
    container.textContent = "例句加载失败。";
    return;
  }
  renderDictionaryExamples(result.data.items || [], entryId);
}

function renderDictionaryExamples(items, entryId) {
  const container = document.getElementById("dictionary-example-list");
  if (!container) {
    return;
  }
  if (items.length === 0) {
    container.innerHTML = `<div class="empty-state">还没有额外例句。每次点击可生成 1 句，最多 3 句。</div>`;
    return;
  }
  container.innerHTML = items
    .map(
      (item, index) => `
        <div class="dictionary-example-item">
          <span class="tag">${index + 1}/3</span>
          <strong>${escapeHTML(item.example_sentence)}</strong>
          <span class="meta">${escapeHTML(item.example_translation_zh || "-")}</span>
          <button class="btn btn-ghost compact" data-delete-example="${item.id}" type="button">删除</button>
        </div>
      `,
    )
    .join("");
  container.querySelectorAll("[data-delete-example]").forEach((button) => {
    button.addEventListener("click", async () => {
      const result = await request(`/api/dictionary/examples/${button.dataset.deleteExample}`, {
        method: "DELETE",
        loadingMessage: "正在删除例句...",
      });
      if (!result.ok) {
        return;
      }
      await loadDictionaryExamples(entryId);
      setMessage("例句已删除");
    });
  });
}

async function generateDictionaryExample(entryId) {
  const result = await request("/api/dictionary/examples/generate", {
    method: "POST",
    body: JSON.stringify({ dictionary_entry_id: entryId }),
    loadingMessage: "正在生成 AI 例句...",
    timeoutMs: 60000,
  });
  if (!result.ok) {
    return;
  }
  await loadDictionaryExamples(entryId);
  setMessage("已生成 1 句 AI 例句");
}

function bindArticleSelection(container) {
  container.querySelectorAll("[data-article-id]").forEach((button) => {
    button.addEventListener("click", async () => {
      const articleId = Number(button.dataset.articleId);
      state.selectedArticleId = articleId;
      showView("reading");
      await loadReadingArticle(articleId);
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
  const withinReading = readingContent.contains(range.commonAncestorContainer);
  const withinChallenge = challengeSentence.contains(range.commonAncestorContainer);
  const withinPostQuiz = postQuizQuestion.contains(range.commonAncestorContainer);
  if (!text || (!withinReading && !withinChallenge && !withinPostQuiz)) {
    return null;
  }
  if (text.length > 40) {
    setMessage("查词文本过长，请选择一个词或短语");
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
    sentenceId: Number(sentenceElement.dataset.sentenceId) || null,
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

function renderReadingArticleText(text) {
  const normalized = String(text || "").replace(/\r\n/g, "\n").trim();
  if (!normalized) {
    return "";
  }
  return normalized
    .split(/\n{2,}/)
    .map((paragraph) => paragraph.trim())
    .filter(Boolean)
    .map(
      (paragraph) => `
        <p class="reading-sentence" data-sentence-id="" data-sentence-text="${escapeHTMLAttribute(paragraph)}">
          <span class="reading-text">${escapeHTML(paragraph)}</span>
        </p>
      `,
    )
    .join("");
}

async function lookupSelection(selectionState) {
  positionPopup(selectionState.rect);
  popup.classList.remove("hidden");
  popupTitle.textContent = `查词：${selectionState.text}`;
  popupBody.innerHTML = renderLookupStatus("正在查询本地词典...");
  addVocabularyButton.disabled = true;
  addVocabularyButton.textContent = "加入生词本";

  const lookupKey = `${selectionState.text}:${selectionState.sentenceId}:${selectionState.contextSnippet}`;
  if (state.lookup.inFlightKey === lookupKey) {
    return;
  }
  state.lookup.currentText = selectionState.text;
  state.lookup.currentSentenceId = selectionState.sentenceId;
  state.lookup.currentSentenceText = selectionState.sentenceText;
  state.lookup.currentContextSnippet = selectionState.contextSnippet;

  if (state.lookup.lastLookupKey === lookupKey && state.lookup.currentEntry) {
    await refreshVocabularyButton(state.lookup.currentEntry.id);
    popupBody.innerHTML = formatDictionaryEntry(state.lookup.currentEntry, state.lookup.currentGenerated, state.lookup.currentContextSnippet);
    return;
  }

  state.lookup.inFlightKey = lookupKey;
  try {
    const searchResult = await request(`/api/dictionary/search?text=${encodeURIComponent(selectionState.text)}`, {
      loadingMessage: "正在查询本地词典...",
    });
    if (!searchResult.ok) {
      popupBody.textContent = "词典查询失败，可以重新选择文本再试。";
      return;
    }

    let entry = searchResult.data.entry;
    let generated = false;
    if (!searchResult.data.found) {
      popupBody.innerHTML = renderLookupStatus("本地词典未命中，正在调用 AI 生成释义、词性和例句...");
      const generateResult = await request("/api/dictionary/generate", {
        method: "POST",
        body: JSON.stringify({
          text: selectionState.text,
          context: selectionState.contextSnippet || selectionState.sentenceText,
        }),
        loadingMessage: "正在调用 AI 生成词条...",
        timeoutMs: 60000,
      });
      if (!generateResult.ok) {
        popupBody.textContent = "AI 生成词条失败，可以检查 AI 配置后重试。";
        return;
      }
      entry = generateResult.data.entry;
      generated = generateResult.data.generated;
    }

    state.lookup.currentEntry = entry;
    state.lookup.currentGenerated = generated;
    state.lookup.lastLookupKey = lookupKey;
    popupBody.innerHTML = formatDictionaryEntry(entry, generated, selectionState.contextSnippet);
    await refreshVocabularyButton(entry.id);
  } finally {
    state.lookup.inFlightKey = "";
  }
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
  const width = popupCard.offsetWidth || 380;
  const height = popupCard.offsetHeight || 280;
  const spaceBelow = window.innerHeight - rect.bottom;
  const spaceAbove = rect.top;
  const shouldOpenAbove = spaceBelow < height + 24 && spaceAbove > spaceBelow;
  const top = shouldOpenAbove ? rect.top - height - 12 : rect.bottom + 12;
  const left = rect.left + width > window.innerWidth ? window.innerWidth - width - 12 : rect.left;
  popupCard.style.top = `${clamp(top, 12, window.innerHeight - height - 12)}px`;
  popupCard.style.left = `${clamp(left, 12, window.innerWidth - width - 12)}px`;
}

function hideLookupPopup() {
  clearPendingLookup();
  popup.classList.add("hidden");
}

function formatDictionaryEntry(entry, generated, contextSnippet) {
  return `
    <div class="dictionary-row"><strong>原形</strong><span>${escapeHTML(entry.lemma || "-")}</span></div>
    <div class="dictionary-row"><strong>读音</strong><span>${escapeHTML(entry.reading || "-")} / ${escapeHTML(entry.romaji || "-")}</span></div>
    <div class="dictionary-row"><strong>分类</strong><span class="tag">${escapeHTML(dictionaryEntryKindLabel(entry.part_of_speech))}</span> <span class="tag">${escapeHTML(entry.part_of_speech || "-")}</span> <span class="badge badge-jlpt">${escapeHTML(entry.jlpt_level || "-")}</span></div>
    <div><strong>中文释义</strong><br>${escapeHTML(entry.meaning_zh || entry.primary_meaning_zh || "-")}</div>
    <div><strong>保存例句</strong><br>${escapeHTML(contextSnippet || "-")}</div>
    <div class="meta">${generated ? "本地未命中，已由 AI 生成并写入词典。" : "来自本地词典或已缓存词条。"}</div>
  `;
}

function dictionaryEntryKindLabel(partOfSpeech) {
  const value = String(partOfSpeech || "").toLowerCase();
  if (value === "grammar" || value.includes("grammar") || value.includes("文法") || value.includes("语法")) {
    return "文法";
  }
  return "单词";
}

function renderLookupStatus(message) {
  return `<div class="lookup-status"><span class="loading-spinner"></span><span>${escapeHTML(message)}</span></div>`;
}

function statCard(label, value) {
  return `<div class="card stat-card"><span class="meta">${escapeHTML(label)}</span><strong style="font-size:28px">${escapeHTML(value ?? "-")}</strong></div>`;
}

function percent(correct, total) {
  const denominator = Number(total || 0);
  if (!denominator) {
    return "0%";
  }
  return `${Math.round((Number(correct || 0) / denominator) * 100)}%`;
}

function extractContextSnippet(sentenceText, selectedText) {
  const text = String(sentenceText || "").replace(/\s+/g, " ").trim();
  const needle = String(selectedText || "").trim();
  if (!text || !needle) {
    return text;
  }
  if (text.length <= 80) {
    return text;
  }

  const index = text.indexOf(needle);
  if (index < 0) {
    return text;
  }
  const chars = Array.from(text);
  const selectedStart = Array.from(text.slice(0, index)).length;
  const selectedEnd = selectedStart + Array.from(needle).length;
  const start = findContextStart(chars, selectedStart);
  const end = findContextEnd(chars, selectedEnd);
  return chars.slice(start, end).join("").trim() || text;
}

function findContextStart(chars, selectedStart) {
  let quoteDepth = 0;
  for (let i = selectedStart - 1; i >= 0; i -= 1) {
    const ch = chars[i];
    if (isClosingQuote(ch)) {
      quoteDepth += 1;
      continue;
    }
    if (isOpeningQuote(ch) && quoteDepth > 0) {
      quoteDepth -= 1;
      continue;
    }
    if (quoteDepth === 0 && isSentenceTerminator(ch)) {
      return i + 1;
    }
  }
  return 0;
}

function findContextEnd(chars, selectedEnd) {
  let quoteDepth = 0;
  let pendingQuotedEnd = false;
  for (let i = selectedEnd; i < chars.length; i += 1) {
    const ch = chars[i];
    if (isOpeningQuote(ch)) {
      quoteDepth += 1;
      continue;
    }
    if (isSentenceTerminator(ch)) {
      if (quoteDepth > 0) {
        pendingQuotedEnd = true;
        continue;
      }
      return i + 1;
    }
    if (isClosingQuote(ch)) {
      if (quoteDepth > 0) {
        quoteDepth -= 1;
      }
      if (pendingQuotedEnd && quoteDepth === 0) {
        return i + 1;
      }
    }
  }
  return chars.length;
}

function isSentenceTerminator(ch) {
  return ["。", "！", "？", "!", "?"].includes(ch);
}

function isOpeningQuote(ch) {
  return ["「", "『", "“", "‘", "（", "(", "《"].includes(ch);
}

function isClosingQuote(ch) {
  return ["」", "』", "”", "’", "）", ")", "》"].includes(ch);
}

async function request(url, options = {}) {
  const { loadingMessage, timeoutMs = 30000, headers = {}, silent = false, ...fetchOptions } = options;
  const controller = new AbortController();
  const timeout = window.setTimeout(() => controller.abort(), timeoutMs);
  if (loadingMessage && !silent) {
    setMessage(loadingMessage);
  }
  if (!silent) {
    setGlobalLoading(true);
  }
  try {
    const response = await fetch(url, {
      headers: {
        "Content-Type": "application/json",
        ...(state.token ? { Authorization: `Bearer ${state.token}` } : {}),
        ...headers,
      },
      ...fetchOptions,
      signal: controller.signal,
    });
    const data = await response.json().catch(() => ({}));
    if (!response.ok) {
      if (!silent) {
        setMessage(data.error || `请求失败：${response.status}`);
      }
      return { ok: false, data };
    }
    return { ok: true, data };
  } catch (error) {
    if (!silent) {
      setMessage(error.name === "AbortError" ? "请求超时，请稍后重试" : `网络错误：${error.message}`);
    }
    return { ok: false, data: null };
  } finally {
    window.clearTimeout(timeout);
    if (!silent) {
      setGlobalLoading(false);
    }
  }
}

function setMessage(message) {
  messageBox.textContent = message;
}

function setGlobalLoading(active) {
  state.pendingRequests += active ? 1 : -1;
  if (state.pendingRequests < 0) {
    state.pendingRequests = 0;
  }
  globalLoading.classList.toggle("hidden", state.pendingRequests === 0);
  document.body.toggleAttribute("aria-busy", state.pendingRequests > 0);
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

function formatDateTime(value) {
  if (!value) {
    return "-";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return String(value);
  }
  return date.toLocaleString();
}

function clamp(value, min, max) {
  return Math.min(Math.max(value, min), Math.max(min, max));
}

bootstrap();
