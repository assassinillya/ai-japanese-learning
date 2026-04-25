ALTER TABLE challenge_questions
    ADD COLUMN IF NOT EXISTS question_type TEXT NOT NULL DEFAULT 'challenge_reading';

ALTER TABLE challenge_questions
    ADD COLUMN IF NOT EXISTS jlpt_level TEXT NOT NULL DEFAULT 'unknown';

ALTER TABLE challenge_questions
    ADD COLUMN IF NOT EXISTS ai_model TEXT;

ALTER TABLE challenge_questions
    ADD COLUMN IF NOT EXISTS prompt_version TEXT;

ALTER TABLE challenge_questions
    DROP CONSTRAINT IF EXISTS challenge_questions_question_type_check;

ALTER TABLE challenge_questions
    ADD CONSTRAINT challenge_questions_question_type_check
    CHECK (question_type IN ('challenge_reading', 'post_reading_quiz'));

ALTER TABLE challenge_questions
    DROP CONSTRAINT IF EXISTS challenge_questions_jlpt_level_check;

ALTER TABLE challenge_questions
    ADD CONSTRAINT challenge_questions_jlpt_level_check
    CHECK (jlpt_level IN ('N5', 'N4', 'N3', 'N2', 'N1', 'unknown'));
