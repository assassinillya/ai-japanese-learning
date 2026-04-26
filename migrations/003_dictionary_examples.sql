CREATE TABLE IF NOT EXISTS dictionary_examples (
    id BIGSERIAL PRIMARY KEY,
    dictionary_entry_id BIGINT NOT NULL REFERENCES dictionary_entries(id) ON DELETE CASCADE,
    example_sentence TEXT NOT NULL,
    example_translation_zh TEXT,
    source TEXT NOT NULL CHECK (source IN ('ai', 'admin', 'builtin')),
    ai_model TEXT,
    prompt_version TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
