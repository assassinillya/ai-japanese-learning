CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    username TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_profiles (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    jlpt_level TEXT NOT NULL CHECK (jlpt_level IN ('N5', 'N4', 'N3', 'N2', 'N1')),
    onboarding_completed BOOLEAN NOT NULL DEFAULT FALSE,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS auth_sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS articles (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    title TEXT NOT NULL,
    original_language TEXT NOT NULL,
    original_content TEXT,
    japanese_content TEXT NOT NULL DEFAULT '',
    chinese_translation TEXT,
    jlpt_level TEXT NOT NULL CHECK (jlpt_level IN ('N5', 'N4', 'N3', 'N2', 'N1')),
    source_type TEXT NOT NULL CHECK (source_type IN ('builtin', 'user_uploaded', 'ai_translated')),
    is_ai_generated BOOLEAN NOT NULL DEFAULT FALSE,
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    translation_status TEXT NOT NULL DEFAULT 'pending' CHECK (translation_status IN ('pending', 'processing', 'done', 'failed')),
    processing_notes TEXT,
    sentence_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS article_sentences (
    id BIGSERIAL PRIMARY KEY,
    article_id BIGINT NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    sentence_order INTEGER NOT NULL,
    sentence_text TEXT NOT NULL,
    translation_zh TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (article_id, sentence_order)
);

CREATE TABLE IF NOT EXISTS dictionary_entries (
    id BIGSERIAL PRIMARY KEY,
    surface TEXT NOT NULL,
    lemma TEXT NOT NULL,
    reading TEXT NOT NULL,
    romaji TEXT,
    part_of_speech TEXT NOT NULL,
    meaning_zh TEXT NOT NULL,
    meaning_ja TEXT,
    meaning_en TEXT,
    primary_meaning_zh TEXT NOT NULL,
    jlpt_level TEXT NOT NULL CHECK (jlpt_level IN ('N5', 'N4', 'N3', 'N2', 'N1', 'unknown')),
    example_sentence TEXT,
    example_translation_zh TEXT,
    conjugation_type TEXT,
    is_common BOOLEAN NOT NULL DEFAULT TRUE,
    source TEXT NOT NULL CHECK (source IN ('builtin', 'ai', 'admin')),
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    confidence_score NUMERIC(4, 2) NOT NULL DEFAULT 0.80,
    ai_model TEXT,
    prompt_version TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_vocabulary (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    dictionary_entry_id BIGINT NOT NULL REFERENCES dictionary_entries(id) ON DELETE CASCADE,
    article_id BIGINT REFERENCES articles(id) ON DELETE SET NULL,
    source_sentence_id BIGINT REFERENCES article_sentences(id) ON DELETE SET NULL,
    selected_text TEXT NOT NULL,
    source_sentence_text TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('new', 'learning', 'reviewing', 'mastered', 'ignored')) DEFAULT 'new',
    familiarity INTEGER NOT NULL DEFAULT 0,
    correct_count INTEGER NOT NULL DEFAULT 0,
    wrong_count INTEGER NOT NULL DEFAULT 0,
    consecutive_correct_count INTEGER NOT NULL DEFAULT 0,
    added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_reviewed_at TIMESTAMPTZ,
    next_review_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, dictionary_entry_id)
);
