CREATE TABLE IF NOT EXISTS ai_cache (
    id BIGSERIAL PRIMARY KEY,
    cache_key TEXT NOT NULL UNIQUE,
    task_type TEXT NOT NULL,
    input_hash TEXT NOT NULL,
    request_json JSONB NOT NULL,
    response_json JSONB NOT NULL,
    model_name TEXT NOT NULL,
    prompt_version TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS ai_cache_task_type_input_hash_idx
    ON ai_cache (task_type, input_hash);

CREATE TABLE IF NOT EXISTS ai_logs (
    id BIGSERIAL PRIMARY KEY,
    task_type TEXT NOT NULL,
    request_json JSONB NOT NULL,
    response_json JSONB,
    status TEXT NOT NULL CHECK (status IN ('success', 'failed')),
    error_message TEXT,
    model_name TEXT NOT NULL,
    prompt_version TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS ai_logs_task_type_created_at_idx
    ON ai_logs (task_type, created_at DESC);
