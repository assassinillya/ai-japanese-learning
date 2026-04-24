ALTER TABLE articles
    ALTER COLUMN japanese_content SET DEFAULT '';

ALTER TABLE articles
    ADD COLUMN IF NOT EXISTS translation_status TEXT NOT NULL DEFAULT 'pending';

ALTER TABLE articles
    ADD COLUMN IF NOT EXISTS processing_notes TEXT;

ALTER TABLE articles
    ADD COLUMN IF NOT EXISTS sentence_count INTEGER NOT NULL DEFAULT 0;

ALTER TABLE articles
    DROP CONSTRAINT IF EXISTS articles_translation_status_check;

ALTER TABLE articles
    ADD CONSTRAINT articles_translation_status_check
    CHECK (translation_status IN ('pending', 'processing', 'done', 'failed'));

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'article_sentences_article_id_sentence_order_key'
    ) THEN
        ALTER TABLE article_sentences
            ADD CONSTRAINT article_sentences_article_id_sentence_order_key
            UNIQUE (article_id, sentence_order);
    END IF;
END $$;
