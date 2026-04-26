# ai-japanese-learning v1.2

当前版本已经从 `v1.1` 推进到 `v1.2`，新增可配置 AI 接入接口和各类 AI Prompt 模板，同时继续保留占位生成器作为 fallback。

已完成：

- Go 后端服务与 PostgreSQL 连接
- 用户注册、登录、退出、当前用户、JLPT 更新
- 文章创建、我的文章列表、文章详情、句子列表、重新处理
- 基础语言检测
- 非日语文章的占位翻译服务
- 文章句子拆分入库
- 内置文章库种子数据
- 文章阅读接口 `/api/reading/articles/{id}`
- 词典纯查库接口 `/api/dictionary/search`
- 词典查询接口 `/api/dictionary/lookup`
- 占位词条生成与词典入库
- 加入生词本与已加入状态查询
- 生词本列表、详情、状态修改、删除接口
- 生词上下文查询接口 `/api/vocabulary/{id}/context`
- 生词本页面与状态筛选
- 查词时的上下文句子或半句会随生词一起保存，作为该词的例句
- 同一个词从新上下文再次加入时，会刷新生词本中的例句上下文
- 挑战阅读题生成接口 `POST /api/reading/articles/{id}/challenge-questions`
- 挑战阅读题查询接口 `GET /api/reading/articles/{id}/challenge-questions`
- 挑战阅读答题接口 `POST /api/reading/questions/{id}/answer`
- 挑战阅读题缓存入库与答题记录保存
- 挑战阅读答题会校验当前用户是否可访问题目所属文章
- 挑战阅读题保存题型、JLPT、AI 模型和 Prompt 版本元数据，便于后续扩展阅读后测验
- 挑战阅读页面、进度展示、正误反馈、下一题流程
- 阅读后测验题生成接口 `POST /api/reading/articles/{id}/post-quiz`
- 阅读后测验题查询接口 `GET /api/reading/articles/{id}/post-quiz`
- 阅读后测验题缓存入库，与挑战阅读题按 `question_type` 隔离
- 阅读后测验页面、进度展示、正误反馈、下一题流程
- 今日待复习生词接口 `GET /api/review/due`
- 词汇复习题生成接口 `POST /api/review/questions`
- 词汇复习答题接口 `POST /api/review/answer`
- 词汇复习题缓存入库与复习记录保存
- 根据答题结果更新生词状态、熟练度、答对/答错次数、连续答对次数和下次复习时间
- 词汇复习页面、上下文展示、进度展示、正误反馈、下一题流程
- AI 缓存表 `ai_cache`
- AI 调用日志表 `ai_logs`
- 词典占位生成结果会写入 AI 缓存和日志
- 非日语文章占位翻译结果会写入 AI 缓存和日志
- 挑战阅读题、阅读后测验题、词汇复习题的占位生成结果会写入 AI 缓存和日志
- 词典生成结果入库前会校验必填字段、JLPT 等级和来源枚举
- 阅读后测验答题结果查询接口 `GET /api/reading/articles/{id}/post-quiz/results`
- 词汇复习记录查询接口 `GET /api/review/records`
- 健康检查接口 `GET /api/health`
- 显式词典生成接口 `POST /api/dictionary/generate`
- 学习记录页面：阅读后测验记录与词汇复习记录
- 新手引导页面和完成引导接口 `POST /api/profile/onboarding/complete`
- 基础学习统计接口 `GET /api/stats/learning`
- 统计概览页面：文章、生词、待复习、阅读答题、复习记录
- 可配置 AI provider 接口，支持 OpenAI、OpenAI Responses、Gemini、Anthropic、Azure OpenAI 和 New API
- AI Prompt 模板：词典生成、文章翻译、挑战阅读、阅读后测验、词汇复习
- 词典生成、文章翻译、词汇复习题已可优先调用真实 AI，失败时降级到占位生成
- 前端基础加载状态、错误提示、空状态和查词防重复请求
- Dockerfile、docker-compose 和 `.env.example` 基础部署配置
- MVP 测试数据 `seeds/002_mvp_seed_v09.sql`
- 静态前端页面：登录、注册、首页、个人中心、文章上传、文章详情、阅读模式、查词弹窗
- 静态前端 UI 重构：统一左侧导航、页面 Header、卡片、按钮、Badge、Toast、阅读页、查词弹窗、生词本、复习页和统计概览视觉系统

## 版本记录

- `v0.2`
  - 用户系统、JLPT 设置、文章上传、文章库、文章处理、句子拆分。
- `v0.3`
  - 文章阅读页。
  - 鼠标框选文本后延迟查词。
  - 词典查询，不存在时生成占位词条并入库。
  - 生词本添加和已加入状态展示。
- `v0.4`
  - 生词本列表页、详情区和状态筛选。
  - 生词状态更新、忽略、删除。
  - 查询生词时的上下文句子或半句会保存到生词本，作为例句展示。
  - 再次从新句子加入同一词时，会更新保存的例句上下文。
- `v0.5`
  - 按文章顺序生成挑战阅读题，并缓存到数据库。
  - 支持四选一答题、答题记录保存、答案解析展示。
  - 答题接口会校验当前用户只能回答自己可访问文章下的题目。
  - 题目表补充题型、JLPT、AI 模型和 Prompt 版本元数据。
  - 前端新增挑战阅读页面，支持进度、提交答案和下一题。
  - 挑战题句子中仍可选中文本继续查词。
- `v0.6`
  - 基于文章句子和词典条目生成阅读后测验题，并缓存到数据库。
  - 支持四选一测验、答题记录保存、答案解析展示。
  - 挑战阅读题和阅读后测验题按 `question_type` 分开缓存，互不覆盖。
  - 前端新增阅读后测验页面，支持进度、提交答案和下一题。
- `v0.7`
  - 查询用户生词本中到期的待复习生词。
  - 基于词典 `primary_meaning_zh` 生成词汇复习四选一题，并缓存到数据库。
  - 支持复习答题记录保存。
  - 根据简单记忆曲线更新生词状态、熟练度和下次复习时间。
  - 前端新增词汇复习页面，支持上下文、进度、提交答案和下一题。
- `v0.8`
  - 新增 AI 缓存表和 AI 调用日志表。
  - 封装基础 AI Service，统一生成 cache key、输入 hash、缓存读写和日志写入。
  - 词典占位生成接入 AI 缓存，避免同一输入重复生成。
  - 词典生成结果入库前校验必填字段、JLPT 枚举和 source 枚举。
- `v0.8-review`
  - 将非日语文章占位翻译、挑战阅读题、阅读后测验题、词汇复习题接入统一 AI 缓存和日志。
  - 新增阅读后测验答题结果查询接口。
  - 新增词汇复习记录查询接口。
- `v0.9`
  - 完成 MVP 流程联调收尾，强化前端加载状态、错误提示、空状态和查词防重复请求。
  - 新增健康检查接口，便于部署探活。
  - 新增 Docker / Compose / env 示例配置。
  - 新增 MVP 测试文章和词典 seed。
- `v1.0`
  - 补齐计划 API 草案中的 `POST /api/dictionary/generate`。
  - 前端新增学习记录页面，展示当前文章阅读后测验记录和最近词汇复习记录。
  - 将真实 AI 服务商、通用 JSON Schema、自动重试队列、文章标签、独立 onboarding 页面列为后续增强项。
- `v1.1`
  - 新增独立新手引导页和完成引导接口。
  - 新增基础学习统计接口和统计概览页面。
  - 将真实 AI 服务商、通用 JSON Schema、自动重试队列、文章标签、错题本和生产级迁移流程继续列为后续 1.x 增强。
- `v1.2`
  - 新增 `AIProvider` 接口和 OpenAI-compatible provider。
  - 新增 `AI_PROVIDER`、`AI_BASE_URL`、`AI_API_KEY`、`AI_MODEL` 配置。
  - 新增各类 AI Prompt 模板，统一要求 JSON 输出。
  - 词典生成、文章翻译、词汇复习题支持配置真实 AI 后优先调用；未配置或失败时继续使用占位逻辑。
  - 生词本能力保持完整：添加、列表、详情、状态筛选、状态修改、删除、上下文保存和复习联动。
- `v1.2-frontend-refactor`
  - 将静态前端重构为现代化单页学习工作台视觉：左侧固定导航、轻量顶部状态区、统一卡片与按钮设计系统。
  - 登录、注册、首页、文章列表、上传、文章详情、阅读模式、挑战阅读、阅读后测验、生词本、词汇复习、学习记录、统计概览和个人中心完成统一样式更新。
  - 阅读页强化日语正文排版、悬浮工具条、学习工具侧栏和小型词典卡查词体验。
  - 词汇复习页改为居中大卡片、顶部悬浮工具条和卡片式选项反馈。
  - 登录后隐藏登录/注册入口；阅读页改为展示完整正文，生词上下文在加入时按自然句边界截取。
  - 后端文章分句增强为尊重日文括号和引号，避免把「勇輝君、持ってきたぞ。」这类引语拆断。
  - 挑战阅读和阅读后测验切换为真实 AI 题目生成路径，未配置 AI 时返回明确提示，不再生成占位乱码选项。
  - 个人中心新增 AI 接入参数记录与服务器环境变量配置提示。
  - 保持原有后端 API 和静态前端单文件运行方式不变。
- `v1.2-dictionary-ai-flow`
  - 新增 `GET /api/dictionary/search?text=xxx` 纯查库接口，命中时只返回已有词条，不触发生成。
  - 阅读、挑战阅读和阅读后测验的鼠标划词弹窗改为先查本地词典，未命中再调用 `POST /api/dictionary/generate`。
  - AI 生成成功后词条会入库并立即显示原形、读音、罗马音、词性、JLPT、中文释义和保存例句，再可加入生词本。
  - 保留旧 `/api/dictionary/lookup` 的查不到即生成行为，兼容已有调用。
- `v1.2-ai-provider-console`
  - AI 接入从单一 OpenAI-compatible 扩展为 OpenAI、OpenAI Responses、Gemini、Anthropic、Azure OpenAI 和 New API。
  - 新增 `GET /api/ai/providers`、`GET /api/ai/config`、`PUT /api/ai/config`、`POST /api/ai/models`、`POST /api/ai/check`。
  - 个人中心支持供应商名称、API Key、API 地址、API Version、模型列表获取、模型选择、连接检测和保存启用。
  - 后端会按供应商类型自动补齐调用 endpoint 和模型列表 endpoint，用户输入 `/v1` 或完整 endpoint 时不会重复追加后缀。
  - 保存后的 AI 配置会在当前服务进程内即时生效，环境变量仍可作为启动默认值。
- `v1.2-study-flow-polish`
  - 进入挑战阅读时先切到挑战页并展示加载卡片，避免 AI 生成期间页面无反馈。
  - 生词本新增批量选择、批量删除、批量标记学习中、批量标记熟练。
  - `mastered` 状态明确作为“熟练/已经学会”，熟练词会从后续待复习队列中移出。
  - 词汇复习卡新增“标记熟练”按钮，点击后当前词直接移出复习队列并跳到下一词。
  - 新增 `POST /api/vocabulary/batch/status` 和 `POST /api/vocabulary/batch/delete`。

后续每次功能或结构改动，都需要同步更新 `README.md` 的版本记录和当前说明。

## 正式名称

项目正式名已改为 `ai-japanese-learning`，Go 模块名也已切换。
当前项目目录：

```text
D:\project\ai-japanese-learning
```

## 目录

```text
cmd/server             服务入口
internal/config        配置读取
internal/db            数据库初始化
internal/model         领域模型
internal/repository    数据访问层
internal/service       业务服务
internal/handler       HTTP 路由与处理器
internal/web/assets    静态前端页面
migrations             SQL 迁移
seeds                  初始化数据
```

## 默认配置

默认读取根目录的 `password.json`：

```json
{
  "pgsql": {
    "ip": "localhost:5432",
    "password": "123456"
  }
}
```

默认数据库：

```text
postgres://postgres:<password>@localhost:5432/japanese_learning?sslmode=disable
```

可覆盖环境变量：

- `DATABASE_URL`
- `APP_TOKEN_SECRET`
- `SERVER_ADDRESS`
- `PORT`
- `AI_PROVIDER`
- `AI_PROVIDER_NAME`
- `AI_BASE_URL`
- `AI_API_KEY`
- `AI_MODEL`
- `AI_API_VERSION`

AI 配置说明：

```text
AI_PROVIDER=placeholder
```

默认不调用外部 AI，继续使用项目内置占位生成器。

支持的 `AI_PROVIDER`：

- `openai`
- `openai-responses`
- `gemini`
- `anthropic`
- `azure-openai`
- `new-api`
- `placeholder`

如需接入 OpenAI Responses：

```text
AI_PROVIDER=openai-responses
AI_PROVIDER_NAME=OpenAI
AI_BASE_URL=https://api.openai.com
AI_API_KEY=<your-api-key>
AI_MODEL=gpt-4o-mini
```

如果 `AI_BASE_URL` 已经包含 `/v1` 或完整 endpoint，后端会按供应商类型自动补齐缺失部分并避免重复追加。例如 OpenAI Responses 输入 `https://api.openai.com`、`https://api.openai.com/v1` 或 `https://api.openai.com/v1/responses` 都会被规范化到对应调用地址。Azure OpenAI 可额外设置 `AI_API_VERSION`。

## 数据库初始化

先创建数据库：

```sql
CREATE DATABASE japanese_learning;
```

新环境直接执行：

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

如果已经跑过旧版本，再按尚未执行过的版本追加执行：

```bash
psql -U postgres -d japanese_learning -f migrations/002_articles_v02.sql
psql -U postgres -d japanese_learning -f migrations/003_challenge_reading_v05.sql
psql -U postgres -d japanese_learning -f migrations/004_challenge_metadata_v05_fix.sql
psql -U postgres -d japanese_learning -f migrations/005_post_reading_quiz_v06.sql
psql -U postgres -d japanese_learning -f migrations/006_vocabulary_review_v07.sql
psql -U postgres -d japanese_learning -f migrations/007_ai_cache_logs_v08.sql
psql -U postgres -d japanese_learning -f seeds/002_mvp_seed_v09.sql
```

## 运行

联网环境可先执行：

```bash
go mod tidy
```

若环境受限，可直接：

```bash
go build -mod=mod ./...
```

启动：

```bash
go run ./cmd/server
```

访问：

```text
http://localhost:8080
```

健康检查：

```text
http://localhost:8080/api/health
```

## Docker 运行

复制环境变量示例后按需修改：

```bash
copy .env.example .env
```

启动 PostgreSQL 和应用：

```bash
docker compose up --build
```

首次启动后需要在数据库容器中执行迁移和 seed，或用本机 `psql` 连接 compose 暴露的 PostgreSQL 后执行上面的初始化 SQL。

## 当前说明

- `v0.2` 的非日语转日语目前仍是占位翻译服务，用于先打通上传和处理流程。后续接入真实 AI 翻译后，只需要替换 `internal/service/translation_service.go`。
- `v0.3` 的词典生成目前也是占位实现，用于先打通阅读查词和生词本流程。后续接入真实 AI 词条生成后，主要替换 `internal/service/dictionary_service.go`。
- `v0.4` 的上下文例句目前来自查词时所在句子或就近半句，保存在生词本的 `source_sentence_text` 中，供后续复习和详情展示使用。
- `v0.5` 的挑战题生成和干扰项目前仍是占位算法实现，用于先打通挑战阅读与题目缓存流程。后续接入真实 AI 后，主要替换 `internal/service/challenge_service.go`。
- `v0.6` 的阅读后测验题目前复用占位题目生成逻辑，优先围绕文章句子中可匹配或可生成的词条出中文释义选择题。后续接入真实 AI 后，继续替换 `internal/service/challenge_service.go` 中的 post quiz 生成逻辑。
- `v0.7` 的词汇复习题目前复用占位干扰项生成逻辑，正确答案来自词典 `primary_meaning_zh`。后续接入真实 AI 后，主要替换 `internal/service/review_service.go` 中的复习题生成逻辑。
- `v0.8-review` 已把当前所有占位 AI 生成路径接入统一 AI 缓存和日志，包括翻译、词典、挑战阅读题、阅读后测验题和词汇复习题。后续接入真实 AI 时可以沿用 `internal/service/ai_service.go` 和 `ai_cache` / `ai_logs` 表。
- `v0.9` 仍是单体 MVP 形态，部署配置只负责启动应用和 PostgreSQL；迁移执行仍保留为显式步骤，避免应用启动时自动改库。
- `v1.0` 仍沿用当前占位 AI 生成器；真实 AI 服务商、通用 JSON Schema 校验和自动重试队列尚未接入。
- `v1.1` 完成新手路径和基础统计，但仍未接入真实 AI 服务商、通用 JSON Schema 校验、AI 自动重试队列、文章标签、错题本和生产级自动迁移。
- `v1.2` 已具备多供应商 AI 接口和 Prompt 模板；真实 AI 调用可通过环境变量或个人中心配置启用。

## v0.5 Review 记录

本轮 review 覆盖 `v0.1` 到 `v0.5` 的文档、迁移、后端服务、静态前端和本地 PostgreSQL 初始化流程。

已确认：

- `v0.1` 用户、资料、会话、JLPT、基础文章、句子、词典、生词表结构已落地。
- `v0.2` 文章上传、语言检测、占位翻译、文章处理和句子拆分流程已落地。
- `v0.3` 阅读页、鼠标框选延迟查词、词典占位生成、加入生词本和已加入状态查询已落地。
- `v0.4` 生词本列表、状态筛选、详情、上下文保存、状态修改和删除已落地。
- `v0.5` 挑战阅读题生成、缓存、答题记录、前端答题反馈和下一题流程已落地。

本轮修复：

- 修复 README 新环境初始化步骤遗漏 `002`、`003`、`004` 迁移的问题。
- 修复挑战阅读答题接口只按题目 ID 查询、缺少文章访问权限校验的问题。
- 为挑战题补充 `question_type`、`jlpt_level`、`ai_model`、`prompt_version` 字段，降低后续扩展阅读后测验时的迁移成本。
- 新增 `migrations/004_challenge_metadata_v05_fix.sql`，用于旧库无损补齐挑战题元数据字段。

验证结果：

- `go test ./...` 通过。
- 本地 PostgreSQL 已创建 `japanese_learning` 数据库并执行 `001` 到 `004` 迁移和 seed。
- 当前种子数据包含 2 篇文章、3 条文章句子、2 条词典条目。
- 本地服务可启动并通过 `http://localhost:8080` 访问。

## v0.6 验证记录

本轮实现阅读后测验模式，复用 `challenge_questions` 作为阅读题缓存表，并通过 `question_type` 区分挑战阅读和阅读后测验。

已验证：

- `go test ./...` 通过。
- 本地 PostgreSQL 已执行 `migrations/005_post_reading_quiz_v06.sql`。
- `GET /api/reading/articles/{id}/post-quiz` 可基于内置文章生成 `post_reading_quiz` 题目。
- `POST /api/reading/questions/{id}/answer` 可提交阅读后测验答案并保存答题记录。
- 本地 HTTP 烟测中，内置文章 `朝の散歩` 生成 2 道阅读后测验题，提交正确答案返回 `is_correct = true`。

## v0.7 验证记录

本轮实现词汇复习模式，新增复习题缓存表和复习记录表，并复用 `user_vocabulary` 中已有的复习字段更新记忆曲线。

已验证：

- `go test ./...` 通过。
- 本地 PostgreSQL 已执行 `migrations/006_vocabulary_review_v07.sql`。
- `GET /api/review/due` 可返回用户到期待复习生词和缓存复习题。
- `POST /api/review/answer` 可提交复习答案、保存复习记录并更新生词复习字段。
- 本地 HTTP 烟测中，测试用户添加 `散歩` 到生词本后生成 1 道复习题，提交正确答案返回 `is_correct = true`、`status = learning`、`consecutive_correct_count = 1`。

## v0.8 验证记录

本轮实现 AI 缓存、AI 调用日志和词典生成约束的基础能力，先接入词典占位生成路径。

已验证：

- `go test ./...` 通过。
- 本地 PostgreSQL 已执行 `migrations/007_ai_cache_logs_v08.sql`。
- 查询不存在的词条时，会生成占位词典结果、校验字段、写入 `ai_cache` 和 `ai_logs`，再写入 `dictionary_entries`。
- 再次查询同一文本时会优先命中 `dictionary_entries`，避免重复生成。
- 本地 HTTP 烟测中，测试词 `v08smokeword...` 返回 `generated = true`，并验证 `ai_cache` 和 `ai_logs` 中各写入 1 条对应记录。

## v0.8-review 验证记录

本轮补齐 0.1 到 0.8 迭代 review 后发现的缺口，重点是扩大 AI 缓存覆盖范围，并补上结果查询接口。

已验证：

- `go test ./...` 通过。
- 非日语文章占位翻译通过 `AIService` 写入 `ai_cache` 和 `ai_logs`。
- 挑战阅读题、阅读后测验题、词汇复习题生成通过 `AIService` 读写缓存和日志。
- `GET /api/reading/articles/{id}/post-quiz/results` 可返回当前用户在指定文章下的阅读后测验答题记录和题目信息。
- `GET /api/review/records` 可返回当前用户的词汇复习记录、题目、生词和词典信息。

## v0.9 验证记录

本轮完成 MVP 发布前联调收尾，重点处理前端交互稳定性、部署入口和测试数据。

已验证：

- `go test ./...` 通过。
- `GET /api/health` 返回 `status = ok` 和 `version = v0.9`。
- 前端全局 loading、请求超时提示、空列表状态和查词防重复请求已接入。
- Dockerfile、docker-compose、`.env.example` 已加入基础部署配置。
- `seeds/002_mvp_seed_v09.sql` 已加入 1 篇 MVP 内置文章和 2 条词典测试数据。

## v1.0 验证记录

本轮将计划 review 后确认的缺口整理进 `计划.md` 的 `Version 1.0`，并先落实可直接闭合的正式化功能。

已验证：

- `go test ./...` 通过。
- `node --check internal/web/assets/app.js` 通过。
- `POST /api/dictionary/generate` 可复用词典查询生成逻辑返回词条。
- 学习记录页可加载 `GET /api/reading/articles/{id}/post-quiz/results` 和 `GET /api/review/records`。

## v1.1 验证记录

本轮将项目完结状态中仍保留的后续增强整理为 `计划.md` 的 `Version 1.1`，并先落实新手路径和基础统计。

已验证：

- `go test ./...` 通过。
- `node --check internal/web/assets/app.js` 通过。
- `POST /api/profile/onboarding/complete` 可将新用户引导状态置为完成。
- `GET /api/stats/learning` 可返回文章、生词、待复习、阅读答题、词汇复习和生词状态统计。

## v1.2 验证记录

本轮新增真实 AI 接入基础设施和 Prompt 模板，并保持未配置 AI 时的本地可运行能力。

已验证：

- `go test ./...` 通过。
- `node --check internal/web/assets/app.js` 通过。
- 未配置 `AI_API_KEY` 时，词典生成、文章翻译、词汇复习仍可走占位 fallback。
- `GET /api/health` 返回 `version = v1.2`。
- 配置真实 AI provider 后会通过对应供应商接口请求 JSON 输出。

## v1.2 前端重构验证记录

本轮按 `前端重构.md` 对 `internal/web/assets` 静态前端做视觉和布局重构，保留原有 API 请求逻辑。

已验证：

- `node --check internal/web/assets/app.js` 通过。
- `GOCACHE=D:\project\ai-japanese-learning\.gocache go test ./...` 通过。
- 登录、注册、首页、文章列表、上传、文章详情、阅读模式、查词弹窗、生词本、词汇复习、学习记录、统计概览和个人中心仍使用原有 DOM id 与 API 入口。
- 阅读查词恢复为完整正文选词，保存生词时以前后句号和引号边界截取上下文。
- 挑战阅读与阅读后测验现在依赖已配置 AI Provider；旧占位题缓存会在下次生成时替换。

## v1.2 划词 AI 查词链路验证记录

本轮在前端新页面上适配查词弹窗流程：鼠标划词后先请求本地词典，未命中再调用 AI 词条生成接口，生成结果入库后直接显示并可加入生词本。

已验证：

- `node --check internal/web/assets/app.js` 通过。
- `GOCACHE=D:\project\ai-japanese-learning\.gocache go test ./...` 通过。
- `go build -o .\tmp\ai-japanese-learning.exe .\cmd\server` 通过。
- 本地烟测 `GET /api/dictionary/search` 首次返回 `found=false`，随后 `POST /api/dictionary/generate` 返回 `generated=true`，再次 `GET /api/dictionary/search` 返回 `found=true` 且 entry id 一致。

## v1.2 AI 供应商控制台验证记录

本轮重做 AI 接入配置能力，补齐多供应商、模型列表、连接检测和 endpoint 自动规范化。

已验证：

- `node --check internal/web/assets/app.js` 通过。
- `GOCACHE=D:\project\ai-japanese-learning\.gocache go test ./...` 通过。
- `GET /api/ai/providers` 可返回 OpenAI、OpenAI Responses、Gemini、Anthropic、Azure OpenAI 和 New API 定义。
- `GET /api/ai/config` 可返回当前脱敏配置和规范化后的调用地址、模型地址。
- `PUT /api/ai/config` 可在当前服务进程内保存并启用 AI 配置。

## v1.2 学习流程体验验证记录

本轮补齐挑战阅读加载反馈、生词本批量操作和复习中标记熟练。

已验证：

- `node --check internal/web/assets/app.js` 通过。
- `GOCACHE=D:\project\ai-japanese-learning\.gocache go test ./...` 通过。
- 熟练状态 `mastered` 已从 `GET /api/review/due` 的待复习查询中排除。
- 生词本批量接口 `POST /api/vocabulary/batch/status`、`POST /api/vocabulary/batch/delete` 可按当前用户限制批量操作。
