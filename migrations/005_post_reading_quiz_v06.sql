ALTER TABLE challenge_questions
    DROP CONSTRAINT IF EXISTS challenge_questions_article_id_question_order_key;

ALTER TABLE challenge_questions
    DROP CONSTRAINT IF EXISTS challenge_questions_article_id_question_type_question_order_key;

ALTER TABLE challenge_questions
    ADD CONSTRAINT challenge_questions_article_id_question_type_question_order_key
    UNIQUE (article_id, question_type, question_order);
