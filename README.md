# ai-japanese-learning v0.5

当前版本已经从 `v0.4` 推进到 `v0.5`，实现了挑战阅读模式主流程。

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
- 挑战阅读页面、进度展示、正误反馈、下一题流程
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
  - 前端新增挑战阅读页面，支持进度、提交答案和下一题。
  - 挑战题句子中仍可选中文本继续查词。

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
psql -U postgres -d japanese_learning -f seeds/001_seed.sql
```

如果已经跑过 `v0.1`，再追加执行：

```bash
psql -U postgres -d japanese_learning -f migrations/002_articles_v02.sql
psql -U postgres -d japanese_learning -f migrations/003_challenge_reading_v05.sql
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
