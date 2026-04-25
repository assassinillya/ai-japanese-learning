CREATE TABLE IF NOT EXISTS vocabulary_review_questions (
    id BIGSERIAL PRIMARY KEY,
    dictionary_entry_id BIGINT NOT NULL REFERENCES dictionary_entries(id) ON DELETE CASCADE,
    question_text TEXT NOT NULL,
    correct_answer TEXT NOT NULL,
    option_a TEXT NOT NULL,
    option_b TEXT NOT NULL,
    option_c TEXT NOT NULL,
    option_d TEXT NOT NULL,
    correct_option TEXT NOT NULL CHECK (correct_option IN ('A', 'B', 'C', 'D')),
    explanation_zh TEXT NOT NULL,
    ai_model TEXT,
    prompt_version TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (dictionary_entry_id)
);

CREATE TABLE IF NOT EXISTS vocabulary_review_records (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_vocabulary_id BIGINT NOT NULL REFERENCES user_vocabulary(id) ON DELETE CASCADE,
    review_question_id BIGINT NOT NULL REFERENCES vocabulary_review_questions(id) ON DELETE CASCADE,
    selected_option TEXT NOT NULL CHECK (selected_option IN ('A', 'B', 'C', 'D')),
    is_correct BOOLEAN NOT NULL,
    reviewed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
