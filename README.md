# ai-japanese-learning v0.8

当前版本已经从 `v0.7` 推进到 `v0.8`，实现了 AI 缓存、AI 调用日志和词典生成约束的基础版本。

已完成：

- Go 后端服务与 PostgreSQL 连接
- 用户注册、登录、退出、当前用户、JLPT 更新
- 文章创建、我的文章列表、文章详情、句子列表、重新处理
- 基础语言检测
- 非日语文章的占位翻译服务
- 文章句子拆分入库
- 内置文章库种子数据
- 文章阅读接口 `/api/reading/articles/{id}`
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
- 词典生成结果入库前会校验必填字段、JLPT 等级和来源枚举
- 静态前端页面：登录、注册、首页、个人中心、文章上传、文章详情、阅读模式、查词弹窗

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
- `PORT`

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
```

如果已经跑过旧版本，再按尚未执行过的版本追加执行：

```bash
psql -U postgres -d japanese_learning -f migrations/002_articles_v02.sql
psql -U postgres -d japanese_learning -f migrations/003_challenge_reading_v05.sql
psql -U postgres -d japanese_learning -f migrations/004_challenge_metadata_v05_fix.sql
psql -U postgres -d japanese_learning -f migrations/005_post_reading_quiz_v06.sql
psql -U postgres -d japanese_learning -f migrations/006_vocabulary_review_v07.sql
psql -U postgres -d japanese_learning -f migrations/007_ai_cache_logs_v08.sql
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

## 当前说明

- `v0.2` 的非日语转日语目前仍是占位翻译服务，用于先打通上传和处理流程。后续接入真实 AI 翻译后，只需要替换 `internal/service/translation_service.go`。
- `v0.3` 的词典生成目前也是占位实现，用于先打通阅读查词和生词本流程。后续接入真实 AI 词条生成后，主要替换 `internal/service/dictionary_service.go`。
- `v0.4` 的上下文例句目前来自查词时所在句子或就近半句，保存在生词本的 `source_sentence_text` 中，供后续复习和详情展示使用。
- `v0.5` 的挑战题生成和干扰项目前仍是占位算法实现，用于先打通挑战阅读与题目缓存流程。后续接入真实 AI 后，主要替换 `internal/service/challenge_service.go`。
- `v0.6` 的阅读后测验题目前复用占位题目生成逻辑，优先围绕文章句子中可匹配或可生成的词条出中文释义选择题。后续接入真实 AI 后，继续替换 `internal/service/challenge_service.go` 中的 post quiz 生成逻辑。
- `v0.7` 的词汇复习题目前复用占位干扰项生成逻辑，正确答案来自词典 `primary_meaning_zh`。后续接入真实 AI 后，主要替换 `internal/service/review_service.go` 中的复习题生成逻辑。
- `v0.8` 先把词典占位生成接入统一 AI 缓存和日志，后续接入真实 AI 时可以沿用 `internal/service/ai_service.go` 和 `ai_cache` / `ai_logs` 表。

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
