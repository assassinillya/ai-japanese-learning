INSERT INTO articles (
    title,
    original_language,
    original_content,
    japanese_content,
    chinese_translation,
    jlpt_level,
    source_type,
    is_ai_generated,
    is_verified,
    translation_status,
    processing_notes,
    sentence_count
)
SELECT
    '駅での会話',
    'ja',
    'すみません、駅はどこですか。まっすぐ行って、右に曲がってください。ありがとうございます。',
    'すみません、駅はどこですか。まっすぐ行って、右に曲がってください。ありがとうございます。',
    '不好意思，车站在哪里？请直走，然后向右转。谢谢。',
    'N5',
    'builtin',
    FALSE,
    TRUE,
    'done',
    'Seeded MVP article for v0.9 smoke testing.',
    3
WHERE NOT EXISTS (
    SELECT 1 FROM articles WHERE title = '駅での会話' AND source_type = 'builtin'
);

INSERT INTO article_sentences (article_id, sentence_order, sentence_text, translation_zh)
SELECT id, 1, 'すみません、駅はどこですか。', '不好意思，车站在哪里？'
FROM articles
WHERE title = '駅での会話'
ON CONFLICT (article_id, sentence_order) DO NOTHING;

INSERT INTO article_sentences (article_id, sentence_order, sentence_text, translation_zh)
SELECT id, 2, 'まっすぐ行って、右に曲がってください。', '请直走，然后向右转。'
FROM articles
WHERE title = '駅での会話'
ON CONFLICT (article_id, sentence_order) DO NOTHING;

INSERT INTO article_sentences (article_id, sentence_order, sentence_text, translation_zh)
SELECT id, 3, 'ありがとうございます。', '谢谢。'
FROM articles
WHERE title = '駅での会話'
ON CONFLICT (article_id, sentence_order) DO NOTHING;

INSERT INTO dictionary_entries (
    surface,
    lemma,
    reading,
    romaji,
    part_of_speech,
    meaning_zh,
    primary_meaning_zh,
    jlpt_level,
    example_sentence,
    example_translation_zh,
    source,
    verified,
    confidence_score
)
SELECT
    '駅',
    '駅',
    'えき',
    'eki',
    'noun',
    '车站；火车站。',
    '车站',
    'N5',
    '駅はどこですか。',
    '车站在哪里？',
    'builtin',
    TRUE,
    0.98
WHERE NOT EXISTS (
    SELECT 1 FROM dictionary_entries WHERE surface = '駅' AND reading = 'えき'
);

INSERT INTO dictionary_entries (
    surface,
    lemma,
    reading,
    romaji,
    part_of_speech,
    meaning_zh,
    primary_meaning_zh,
    jlpt_level,
    example_sentence,
    example_translation_zh,
    source,
    verified,
    confidence_score
)
SELECT
    '曲がる',
    '曲がる',
    'まがる',
    'magaru',
    'verb',
    '转弯；弯曲。',
    '转弯',
    'N5',
    '右に曲がってください。',
    '请向右转。',
    'builtin',
    TRUE,
    0.97
WHERE NOT EXISTS (
    SELECT 1 FROM dictionary_entries WHERE surface = '曲がる' AND reading = 'まがる'
);
