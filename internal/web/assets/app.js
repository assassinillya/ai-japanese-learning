const state = {
  token: localStorage.getItem("access_token") || "",
  user: null,
  selectedArticleId: null,
  selectedVocabularyId: null,
  readingArticle: null,
  vocabularyFilter: "",
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
  review: {
    items: [],
    currentIndex: 0,
    selectedOption: "",
    answered: false,
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
  },
  pendingRequests: 0,
};

const views = document.querySelectorAll(".view");
const messageBox = document.getElementById("message-box");
const globalLoading = document.getElementById("global-loading");
const authStatus = document.getElementById("auth-status");
const homeGreeting = document.getElementById("home-greeting");
const profileSummary = document.getElementById("profile-summary");
const learningStats = document.getElementById("learning-stats");
const libraryList = document.getElementById("library-list");
const articleList = document.getElementById("article-list");
const articleDetail = document.getElementById("article-detail");
const sentenceList = document.getElementById("sentence-list");
const readingHeader = document.getElementById("reading-header");
const readingContent = document.getElementById("reading-content");
const challengeHeader = document.getElementById("challenge-header");
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
const reviewProgress = document.getElementById("review-progress");
const reviewQuestion = document.getElementById("review-question");
const reviewContext = document.getElementById("review-context");
const reviewOptions = document.getElementById("review-options");
const reviewFeedback = document.getElementById("review-feedback");
const postQuizResultsList = document.getElementById("post-quiz-results-list");
const reviewRecordsList = document.getElementById("review-records-list");
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
const openChallengeButton = document.getElementById("open-challenge-button");
const openPostQuizButton = document.getElementById("open-post-quiz-button");
const submitChallengeAnswerButton = document.getElementById("submit-challenge-answer-button");
const nextChallengeQuestionButton = document.getElementById("next-challenge-question-button");
const submitPostQuizAnswerButton = document.getElementById("submit-post-quiz-answer-button");
const nextPostQuizQuestionButton = document.getElementById("next-post-quiz-question-button");
const submitReviewAnswerButton = document.getElementById("submit-review-answer-button");
const nextReviewQuestionButton = document.getElementById("next-review-question-button");
const loadPostQuizResultsButton = document.getElementById("load-post-quiz-results-button");
const loadReviewRecordsButton = document.getElementById("load-review-records-button");
const completeOnboardingButton = document.getElementById("complete-onboarding-button");
const reprocessButton = document.getElementById("reprocess-button");

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
    if (view === "review" && state.user) {
      await loadReviewDue();
    }
    if (view === "records" && state.user) {
      await loadLearningRecords();
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
  const result = await request("/api/articles/upload", {
    method: "POST",
    body: JSON.stringify(payload),
    loadingMessage: "正在上传并处理文章...",
    timeoutMs: 60000,
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

openChallengeButton.addEventListener("click", async () => {
  if (!state.selectedArticleId) {
    setMessage("请先选择一篇文章");
    return;
  }
  await loadChallengeQuestions(state.selectedArticleId);
  showView("challenge");
});

openPostQuizButton.addEventListener("click", async () => {
  if (!state.selectedArticleId) {
    setMessage("请先选择一篇文章");
    return;
  }
  await loadPostQuizQuestions(state.selectedArticleId);
  showView("post-quiz");
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
  reviewFeedback.classList.remove("hidden");
  reviewFeedback.textContent = [
    result.data.is_correct ? "回答正确" : "回答错误",
    `正确选项：${result.data.correct_option}`,
    `正确答案：${result.data.correct_answer}`,
    `当前状态：${result.data.status}`,
    `下次复习：${formatDateTime(result.data.next_review_at)}`,
    `解析：${result.data.explanation}`,
  ].join("\n");
  await loadVocabularyList();
  renderReviewQuestion();
});

nextReviewQuestionButton.addEventListener("click", () => {
  if (state.review.currentIndex + 1 >= state.review.items.length) {
    setMessage("词汇复习已完成");
    return;
  }
  state.review.currentIndex += 1;
  state.review.selectedOption = "";
  state.review.answered = false;
  reviewFeedback.classList.add("hidden");
  reviewFeedback.textContent = "";
  renderReviewQuestion();
});

loadPostQuizResultsButton.addEventListener("click", async () => {
  await loadPostQuizResults();
});

loadReviewRecordsButton.addEventListener("click", async () => {
  await loadReviewRecords();
});

readingContent.addEventListener("mouseup", () => {
  scheduleLookupFromSelection();
});

challengeSentence.addEventListener("mouseup", () => {
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
    learningStats.textContent = "请先登录后查看学习统计。";
    libraryList.innerHTML = "";
    articleList.innerHTML = "";
    articleDetail.textContent = "请选择一篇文章。";
    sentenceList.innerHTML = "";
    readingHeader.textContent = "请选择一篇文章进入阅读。";
    readingContent.innerHTML = "";
    challengeHeader.textContent = "请选择一篇文章开始挑战阅读。";
    challengeCard.classList.add("hidden");
    postQuizHeader.textContent = "请选择一篇文章开始测验。";
    postQuizCard.classList.add("hidden");
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
  Promise.all([loadLibrary(), loadArticles(), loadVocabularyList()]);
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
  learningStats.textContent = [
    `我的文章：${stats.article_count}`,
    `生词总数：${stats.vocabulary_count}`,
    `今日待复习：${stats.due_vocabulary_count}`,
    `阅读答题：${stats.reading_attempt_count} 次（正确 ${stats.reading_correct_count} / 错误 ${stats.reading_wrong_count}）`,
    `词汇复习：${stats.review_record_count} 次（正确 ${stats.review_correct_count} / 错误 ${stats.review_wrong_count}）`,
    "",
    "生词状态：",
    `new：${statusCounts.new || 0}`,
    `learning：${statusCounts.learning || 0}`,
    `reviewing：${statusCounts.reviewing || 0}`,
    `mastered：${statusCounts.mastered || 0}`,
    `ignored：${statusCounts.ignored || 0}`,
  ].join("\n");
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
  articleList.innerHTML = `<li class="empty-state">正在加载我的文章...</li>`;
  const result = await request("/api/articles");
  if (!result.ok) {
    return;
  }
  const items = result.data.items || [];
  if (items.length === 0) {
    articleList.innerHTML = `<li class="empty-state">还没有上传文章。可以先去“上传文章”创建第一篇。</li>`;
    return;
  }
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
  if (!sentenceList.innerHTML) {
    sentenceList.innerHTML = `<li class="empty-state">当前文章还没有句子拆分结果，可以尝试重新处理。</li>`;
  }

  reprocessButton.disabled = article.source_type === "builtin";
  reprocessButton.title = article.source_type === "builtin" ? "内置文章无需重新处理" : "";
}

async function loadReadingArticle(articleId) {
  readingContent.innerHTML = `<div class="empty-state">正在加载阅读内容...</div>`;
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
  if (!readingContent.innerHTML) {
    readingContent.innerHTML = `<div class="empty-state">当前文章没有可阅读句子。</div>`;
  }
}

async function loadChallengeQuestions(articleId) {
  challengeHeader.textContent = "正在生成或加载挑战阅读题...";
  challengeCard.classList.add("hidden");
  const result = await request(`/api/reading/articles/${articleId}/challenge-questions`, { timeoutMs: 60000 });
  if (!result.ok) {
    return;
  }

  state.challenge.questions = result.data.items || [];
  state.challenge.currentIndex = 0;
  state.challenge.selectedOption = "";
  state.challenge.answered = false;
  hideLookupPopup();

  if (state.challenge.questions.length === 0) {
    challengeHeader.textContent = "当前文章还没有可用的挑战题。";
    challengeCard.classList.add("hidden");
    return;
  }

  challengeHeader.textContent = "挑战阅读会按文章顺序出题。你仍然可以在题干句子中选中文本查词。";
  challengeCard.classList.remove("hidden");
  challengeFeedback.classList.add("hidden");
  challengeFeedback.textContent = "";
  renderChallengeQuestion();
}

async function loadPostQuizQuestions(articleId) {
  postQuizHeader.textContent = "正在生成或加载阅读后测验题...";
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

  if (state.postQuiz.questions.length === 0) {
    postQuizHeader.textContent = "当前文章还没有可用的测验题。";
    postQuizCard.classList.add("hidden");
    return;
  }

  postQuizHeader.textContent = "阅读后测验会基于文章中的重点词汇出题。";
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
      const className = ["challenge-option", isCorrect ? "correct" : "", isIncorrect ? "incorrect" : ""].filter(Boolean).join(" ");
      return `
        <label class="${className}">
          <input type="radio" name="challenge-option" value="${key}" ${selected ? "checked" : ""} ${state.challenge.answered ? "disabled" : ""} />
          <span>${key}. ${escapeHTML(value)}</span>
        </label>
      `;
    })
    .join("");

  challengeOptions.querySelectorAll('input[name="challenge-option"]').forEach((input) => {
    input.addEventListener("change", () => {
      state.challenge.selectedOption = input.value;
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
      const className = ["challenge-option", isCorrect ? "correct" : "", isIncorrect ? "incorrect" : ""].filter(Boolean).join(" ");
      return `
        <label class="${className}">
          <input type="radio" name="post-quiz-option" value="${key}" ${selected ? "checked" : ""} ${state.postQuiz.answered ? "disabled" : ""} />
          <span>${key}. ${escapeHTML(value)}</span>
        </label>
      `;
    })
    .join("");

  postQuizOptions.querySelectorAll('input[name="post-quiz-option"]').forEach((input) => {
    input.addEventListener("change", () => {
      state.postQuiz.selectedOption = input.value;
    });
  });

  submitPostQuizAnswerButton.disabled = state.postQuiz.answered;
  nextPostQuizQuestionButton.disabled = !state.postQuiz.answered;
}

async function loadReviewDue() {
  reviewHeader.textContent = "正在加载今日待复习生词...";
  reviewCard.classList.add("hidden");
  const result = await request("/api/review/due", { timeoutMs: 60000 });
  if (!result.ok) {
    return;
  }

  state.review.items = result.data.items || [];
  state.review.currentIndex = 0;
  state.review.selectedOption = "";
  state.review.answered = false;

  if (state.review.items.length === 0) {
    reviewHeader.textContent = "当前没有到期需要复习的生词。";
    reviewCard.classList.add("hidden");
    return;
  }

  reviewHeader.textContent = `今日待复习：${state.review.items.length} 个`;
  reviewCard.classList.remove("hidden");
  reviewFeedback.classList.add("hidden");
  reviewFeedback.textContent = "";
  renderReviewQuestion();
}

function renderReviewQuestion() {
  const item = state.review.items[state.review.currentIndex];
  if (!item) {
    reviewCard.classList.add("hidden");
    return;
  }

  const question = item.question;
  reviewProgress.textContent = `第 ${state.review.currentIndex + 1} / ${state.review.items.length} 题`;
  reviewQuestion.textContent = `「${question.question_text}」的主要中文意思是？`;
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
      const className = ["challenge-option", isCorrect ? "correct" : "", isIncorrect ? "incorrect" : ""].filter(Boolean).join(" ");
      return `
        <label class="${className}">
          <input type="radio" name="review-option" value="${key}" ${selected ? "checked" : ""} ${state.review.answered ? "disabled" : ""} />
          <span>${key}. ${escapeHTML(value)}</span>
        </label>
      `;
    })
    .join("");

  reviewOptions.querySelectorAll('input[name="review-option"]').forEach((input) => {
    input.addEventListener("change", () => {
      state.review.selectedOption = input.value;
    });
  });

  submitReviewAnswerButton.disabled = state.review.answered;
  nextReviewQuestionButton.disabled = !state.review.answered;
}

async function loadVocabularyList() {
  const suffix = state.vocabularyFilter ? `?status=${encodeURIComponent(state.vocabularyFilter)}` : "";
  vocabularyList.innerHTML = `<li class="empty-state">正在加载生词本...</li>`;
  const result = await request(`/api/vocabulary${suffix}`);
  if (!result.ok) {
    return;
  }

  const items = result.data.items || [];
  if (items.length === 0) {
    vocabularyList.innerHTML = `<li class="empty-state">当前筛选下没有生词。阅读文章时框选词语即可加入。</li>`;
    return;
  }
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
  const withinReading = readingContent.contains(range.commonAncestorContainer);
  const withinChallenge = challengeSentence.contains(range.commonAncestorContainer);
  if (!text || (!withinReading && !withinChallenge)) {
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
  if (state.lookup.inFlightKey === lookupKey) {
    return;
  }
  state.lookup.currentText = selectionState.text;
  state.lookup.currentSentenceId = selectionState.sentenceId;
  state.lookup.currentSentenceText = selectionState.sentenceText;
  state.lookup.currentContextSnippet = selectionState.contextSnippet;

  if (state.lookup.lastLookupKey === lookupKey && state.lookup.currentEntry) {
    await refreshVocabularyButton(state.lookup.currentEntry.id);
    popupBody.textContent = formatDictionaryEntry(state.lookup.currentEntry, state.lookup.currentGenerated, state.lookup.currentContextSnippet);
    return;
  }

  state.lookup.inFlightKey = lookupKey;
  try {
    const lookupResult = await request(`/api/dictionary/lookup?text=${encodeURIComponent(selectionState.text)}`, {
      loadingMessage: "正在查词...",
      timeoutMs: 60000,
    });
    if (!lookupResult.ok) {
      popupBody.textContent = "词典查询失败，可以重新选择文本再试。";
      return;
    }

    state.lookup.currentEntry = lookupResult.data.entry;
    state.lookup.currentGenerated = lookupResult.data.generated;
    state.lookup.lastLookupKey = lookupKey;
    popupBody.textContent = formatDictionaryEntry(lookupResult.data.entry, lookupResult.data.generated, selectionState.contextSnippet);
    await refreshVocabularyButton(lookupResult.data.entry.id);
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
  const { loadingMessage, timeoutMs = 30000, headers = {}, ...fetchOptions } = options;
  const controller = new AbortController();
  const timeout = window.setTimeout(() => controller.abort(), timeoutMs);
  if (loadingMessage) {
    setMessage(loadingMessage);
  }
  setGlobalLoading(true);
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
      setMessage(data.error || `请求失败：${response.status}`);
      return { ok: false, data };
    }
    return { ok: true, data };
  } catch (error) {
    setMessage(error.name === "AbortError" ? "请求超时，请稍后重试" : `网络错误：${error.message}`);
    return { ok: false, data: null };
  } finally {
    window.clearTimeout(timeout);
    setGlobalLoading(false);
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

bootstrap();
