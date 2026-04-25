CREATE TABLE IF NOT EXISTS challenge_questions (
    id BIGSERIAL PRIMARY KEY,
    article_id BIGINT NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    sentence_id BIGINT NOT NULL REFERENCES article_sentences(id) ON DELETE CASCADE,
    question_order INTEGER NOT NULL,
    sentence_text TEXT NOT NULL,
    masked_sentence TEXT NOT NULL,
    correct_entry_id BIGINT NOT NULL REFERENCES dictionary_entries(id) ON DELETE CASCADE,
    correct_answer_text TEXT NOT NULL,
    option_a TEXT NOT NULL,
    option_b TEXT NOT NULL,
    option_c TEXT NOT NULL,
    option_d TEXT NOT NULL,
    correct_option TEXT NOT NULL CHECK (correct_option IN ('A', 'B', 'C', 'D')),
    explanation TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (article_id, question_order)
);

CREATE TABLE IF NOT EXISTS challenge_question_attempts (
    id BIGSERIAL PRIMARY KEY,
    question_id BIGINT NOT NULL REFERENCES challenge_questions(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    selected_option TEXT NOT NULL CHECK (selected_option IN ('A', 'B', 'C', 'D')),
    is_correct BOOLEAN NOT NULL,
    answered_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
