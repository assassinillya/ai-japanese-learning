# ai-japanese-learning 项目汇总

## 项目简介

`ai-japanese-learning` 是一个面向中文用户的 AI 日语阅读与生词复习网站。项目目标是让用户通过自己感兴趣的文章学习日语，在阅读中查词、积累生词，并通过简单记忆曲线进行复习。

当前版本：`v1.4`

当前定位：已完成 MVP 闭环和基础学习体验。项目可以完成注册、文章阅读、查词、生词本、复习、学习记录和统计概览等核心流程。

## 核心学习流程

```text
注册 / 登录
  ↓
完成新手引导并确认 JLPT 等级
  ↓
上传文章或选择内置文章
  ↓
进入阅读模式
  ↓
框选词语查词
  ↓
加入生词本并保存上下文
  ↓
进行词汇复习
  ↓
查看学习记录和统计概览
```

## 已实现功能

### 用户与资料

- 用户注册、登录、退出。
- 当前用户信息查询。
- JLPT 等级选择与修改，支持 `N5` 到 `N1`。
- 新手引导页面。
- 完成新手引导接口：`POST /api/profile/onboarding/complete`。
- 用户数据隔离：生词本、答题记录、复习记录均绑定用户。
- 学习内容公共化：文章、词典、挑战阅读题、阅读后测验题、复习题和 AI 例句存入本地公共库，用户学习时优先复用已有数据。

### 文章与阅读

- 用户上传文章。
- 内置文章库。
- 我的文章列表、文章详情、句子列表。
- 文章重新处理。
- 基础语言检测。
- 非日语文章会通过占位翻译服务转成日语。
- 文章内容会拆分成句子并保存。
- 阅读页面支持选择文章后进入阅读。

### 鼠标框选查词

- 阅读页面和挑战阅读题干支持鼠标框选文本。
- 停留约 0.5 秒后弹出查词窗口。
- 词典存在时直接展示词条。
- 词典不存在时生成占位 AI 词条并入库。
- 查词结果包括：
  - 词形
  - 原形
  - 读音
  - 罗马音
  - 词性
  - 中文释义
  - 主要中文释义
  - JLPT 等级
  - 例句
- 查词时有请求超时、错误提示、重复请求保护和文本长度限制。

### 词典系统

- 词典查询接口：`GET /api/dictionary/lookup?text=xxx`。
- 显式词典生成接口：`POST /api/dictionary/generate`。
- 词典详情接口：`GET /api/dictionary/{id}`。
- 词典字段入库前会做基础完整性校验。
- 词条包含来源、可信度、是否审核、AI 模型、Prompt 版本等元数据。

### 生词本

- 从查词弹窗加入生词本。
- 生词本列表。
- 生词详情。
- AI 例句管理：每次生成 1 句，最多 3 句，可删除后继续生成。
- 变形词会通过隐藏索引映射到原型词条，生词本中以原型词条复用。
- 生词状态筛选。
- 修改生词状态。
- 删除生词。
- 批量删除、批量标记学习中、批量标记熟练。
- 打开来源文章。
- 生词会保存上下文，包括来源文章、来源句子、用户当时框选文本。
- 同一用户重复添加同一词时不会重复创建，必要时会刷新上下文。

生词状态：

```text
new
learning
reviewing
mastered
ignored
```

`mastered` 表示熟练/已经学会，后续不会出现在今日待复习队列中。

### 挑战阅读

- 挑战阅读题生成接口：`POST /api/reading/articles/{id}/challenge-questions`。
- 挑战阅读题查询接口：`GET /api/reading/articles/{id}/challenge-questions`。
- 挑战阅读答题接口：`POST /api/reading/questions/{id}/answer`。
- 题目会按文章顺序生成。
- 题目、选项、答案解释会缓存入库。
- 答题记录会保存。
- 答题接口会校验当前用户是否可访问题目所属文章。
- 前端支持进度、四选一、正误反馈、答案解释和下一题。

### 阅读后测验

- 阅读后测验题生成接口：`POST /api/reading/articles/{id}/post-quiz`。
- 阅读后测验题查询接口：`GET /api/reading/articles/{id}/post-quiz`。
- 阅读后测验答题复用阅读答题接口：`POST /api/reading/questions/{id}/answer`。
- 阅读后测验结果查询接口：`GET /api/reading/articles/{id}/post-quiz/results`。
- 前端支持测验答题、正误反馈、答案解释和学习记录展示。

### 词汇复习

- 今日待复习接口：`GET /api/review/due`。
- 词汇复习题生成接口：`POST /api/review/questions`。
- 词汇复习答题接口：`POST /api/review/answer`。
- 词汇复习记录查询接口：`GET /api/review/records`。
- 复习题来自用户自己的生词本。
- 正确答案来自词典 `primary_meaning_zh`。
- 生词加入生词本后会后台异步预生成复习题；打开生词本时会补生成缺失题目。
- 答题后保存复习记录。
- 复习中可直接标记熟练，标记后当前词会移出本轮复习并跳到下一词。
- 答题后更新：
  - 状态
  - 熟练度
  - 答对次数
  - 答错次数
  - 连续答对次数
  - 下次复习时间

当前记忆曲线规则：

```text
答错：10 分钟后复习，状态回到 learning，连续答对清零
第 1 次答对：1 天后
连续 2 次答对：3 天后
连续 3 次答对：7 天后，进入 reviewing
连续 4 次答对：15 天后，保持 reviewing
连续 5 次答对：30 天后，进入 mastered
```

### 学习记录与统计

- 学习记录页面：
  - 当前文章阅读后测验记录。
  - 最近词汇复习记录。
- 统计概览页面：
  - 我的文章数量。
  - 生词总数。
  - 今日待复习数量。
  - 阅读答题次数、正确数、错误数。
  - 词汇复习次数、正确数、错误数。
  - 各生词状态数量。

统计接口：

```http
GET /api/stats/learning
```

### AI 缓存与日志

- AI 缓存表：`ai_cache`。
- AI 调用日志表：`ai_logs`。
- 可配置 AI provider 接口。
- 支持 OpenAI、OpenAI Responses、Gemini、Anthropic、Azure OpenAI 和 New API。
- 个人中心可填写供应商名称、API 地址、API Key、API Version，并获取模型列表、选择模型、检测连接、保存启用。
- AI 配置保存到用户账号，登录后的 AI 调用默认使用当前用户自己的配置。
- 后端会按供应商类型自动补齐调用 endpoint 和模型列表 endpoint，输入 `/v1` 或完整 endpoint 时不会重复追加后缀。
- 未配置 `AI_API_KEY` 时，继续使用本地占位生成器。
- 已提供 Prompt 模板：
  - 词典生成。
  - 文章翻译。
  - 挑战阅读题。
  - 阅读后测验题。
  - 词汇复习题。
- 词典生成、文章翻译、词汇复习题支持真实 AI 优先，失败时 fallback 到占位逻辑。
- 当前占位 AI 路径已接入缓存和日志：
  - 非日语文章翻译。
  - 词典生成。
  - 挑战阅读题生成。
  - 阅读后测验题生成。
  - 词汇复习题生成。
- 缓存使用 task type、input hash、model name、prompt version 生成 cache key。

## 前端页面

当前是静态前端页面，位于：

```text
internal/web/assets
```

已实现页面：

- 登录
- 注册
- 新手引导
- 首页
- 我的文章
- 上传文章
- 文章详情
- 阅读模式
- 挑战阅读
- 阅读后测验
- 生词本
- 词汇复习
- 学习记录
- 统计概览
- 个人中心

前端已包含：

- 登录态保存。
- 全局 loading。
- 基础错误提示。
- 空状态展示。
- 请求超时提示。
- 查词重复请求保护。

## 主要 API

### 健康检查

```http
GET /api/health
```

### 认证

```http
POST /api/auth/register
POST /api/auth/login
POST /api/auth/logout
GET  /api/auth/me
```

### 用户资料

```http
GET  /api/profile
PUT  /api/profile/jlpt-level
POST /api/profile/onboarding/complete
```

### 文章

```http
GET  /api/articles/library
GET  /api/articles
POST /api/articles
POST /api/articles/upload
GET  /api/articles/{id}
POST /api/articles/{id}/process
GET  /api/articles/{id}/sentences
```

### 阅读

```http
GET  /api/reading/articles/{id}
POST /api/reading/articles/{id}/challenge-questions
GET  /api/reading/articles/{id}/challenge-questions
POST /api/reading/articles/{id}/post-quiz
GET  /api/reading/articles/{id}/post-quiz
GET  /api/reading/articles/{id}/post-quiz/results
POST /api/reading/questions/{id}/answer
```

### 词典

```http
GET  /api/dictionary/lookup?text=xxx
POST /api/dictionary/generate
GET  /api/dictionary/{id}
```

### 生词本

```http
GET    /api/vocabulary
POST   /api/vocabulary
GET    /api/vocabulary/check?dictionary_entry_id=xxx
GET    /api/vocabulary/{id}
GET    /api/vocabulary/{id}/context
PUT    /api/vocabulary/{id}/status
DELETE /api/vocabulary/{id}
```

### 词汇复习

```http
GET  /api/review/due
POST /api/review/questions
POST /api/review/answer
GET  /api/review/records
```

### 统计

```http
GET /api/stats/learning
```

## 本地运行

### 1. 准备数据库

创建 PostgreSQL 数据库：

```sql
CREATE DATABASE japanese_learning;
```

执行迁移和 seed：

```bash
psql -U postgres -d japanese_learning -f migrations/001_init.sql
psql -U postgres -d japanese_learning -f migrations/002_articles_v02.sql
psql -U postgres -d japanese_learning -f migrations/003_challenge_reading_v05.sql
psql -U postgres -d japanese_learning -f migrations/004_challenge_metadata_v05_fix.sql
psql -U postgres -d japanese_learning -f migrations/005_post_reading_quiz_v06.sql
psql -U postgres -d japanese_learning -f migrations/006_vocabulary_review_v07.sql
psql -U postgres -d japanese_learning -f migrations/007_ai_cache_logs_v08.sql
psql -U postgres -d japanese_learning -f seeds/001_seed.sql
psql -U postgres -d japanese_learning -f seeds/002_mvp_seed_v09.sql
```

### 2. 配置数据库连接

默认读取根目录 `password.json`：

```json
{
  "pgsql": {
    "ip": "localhost:5432",
    "password": "123456"
  }
}
```

也可以使用环境变量覆盖：

```text
DATABASE_URL
APP_TOKEN_SECRET
SERVER_ADDRESS
PORT
AI_PROVIDER
AI_PROVIDER_NAME
AI_BASE_URL
AI_API_KEY
AI_MODEL
AI_API_VERSION
```

默认 AI 配置为：

```text
AI_PROVIDER=placeholder
```

接入 OpenAI Responses 时：

```text
AI_PROVIDER=openai-responses
AI_PROVIDER_NAME=OpenAI
AI_BASE_URL=https://api.openai.com
AI_API_KEY=<your-api-key>
AI_MODEL=gpt-4o-mini
```

### 3. 启动服务

```bash
go run ./cmd/server
```

默认访问：

```text
http://localhost:8080
```

健康检查：

```text
http://localhost:8080/api/health
```

## Docker 运行

复制环境变量示例：

```bash
copy .env.example .env
```

启动：

```bash
docker compose up --build
```

注意：Docker Compose 会启动应用和 PostgreSQL，但迁移和 seed 仍需要手动执行。

## 当前限制

当前项目仍保留以下后续增强项：

- 已具备多供应商 AI 接口，真实调用可通过环境变量或个人中心配置启用。
- 挑战阅读和阅读后测验 Prompt 已准备好，但当前生成路径仍以稳定的本地占位算法为主。
- 尚未实现通用 JSON Schema 校验层。
- 尚未实现 AI 失败自动重试队列。
- 尚未实现文章标签和标签筛选。
- 尚未实现错题本和学习趋势图。
- 尚未实现生产级自动迁移和运维发布流程。

## 适合的使用状态

当前项目适合作为：

- 日语学习网站 MVP。
- Go + PostgreSQL 全栈学习项目。
- AI 学习产品原型。
- 后续接入真实 AI 服务商前的业务闭环基线。

如果作为正式商业产品继续推进，下一步建议优先接入真实 AI provider、补齐 JSON Schema 校验、增加错题本和生产级部署流程。
