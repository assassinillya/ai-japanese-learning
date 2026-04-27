ALTER TABLE vocabulary_review_questions
ADD COLUMN IF NOT EXISTS question_order INT NOT NULL DEFAULT 1;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'vocabulary_review_questions_dictionary_entry_id_key'
    ) THEN
        ALTER TABLE vocabulary_review_questions
        DROP CONSTRAINT vocabulary_review_questions_dictionary_entry_id_key;
    END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS idx_vocabulary_review_questions_entry_order
ON vocabulary_review_questions(dictionary_entry_id, question_order);

ALTER TABLE vocabulary_review_records
ADD COLUMN IF NOT EXISTS familiarity_delta INT NOT NULL DEFAULT 0;
