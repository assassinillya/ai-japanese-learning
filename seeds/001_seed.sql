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
VALUES
(
    '朝の散歩',
    'ja',
    '朝、私は近くの公園を散歩します。空気がきれいで、気分がよくなります。',
    '朝、私は近くの公園を散歩します。空気がきれいで、気分がよくなります。',
    '早上，我会在附近的公园散步。空气很清新，心情也会变好。',
    'N5',
    'builtin',
    FALSE,
    TRUE,
    'done',
    'Seeded builtin article.',
    2
),
(
    '日本のコンビニ',
    'ja',
    '日本のコンビニでは、おにぎりや飲み物だけでなく、公共料金の支払いもできます。',
    '日本のコンビニでは、おにぎりや飲み物だけでなく、公共料金の支払いもできます。',
    '在日本的便利店里，不仅可以买饭团和饮料，还可以缴纳公共事业费用。',
    'N4',
    'builtin',
    FALSE,
    TRUE,
    'done',
    'Seeded builtin article.',
    1
)
ON CONFLICT DO NOTHING;

INSERT INTO article_sentences (article_id, sentence_order, sentence_text, translation_zh)
SELECT id, 1, '朝、私は近くの公園を散歩します。', '早上，我会在附近的公园散步。'
FROM articles
WHERE title = '朝の散歩'
ON CONFLICT (article_id, sentence_order) DO NOTHING;

INSERT INTO article_sentences (article_id, sentence_order, sentence_text, translation_zh)
SELECT id, 2, '空気がきれいで、気分がよくなります。', '空气很清新，心情也会变好。'
FROM articles
WHERE title = '朝の散歩'
ON CONFLICT (article_id, sentence_order) DO NOTHING;

INSERT INTO article_sentences (article_id, sentence_order, sentence_text, translation_zh)
SELECT id, 1, '日本のコンビニでは、おにぎりや飲み物だけでなく、公共料金の支払いもできます。', '在日本的便利店里，不仅可以买饭团和饮料，还可以缴纳公共事业费用。'
FROM articles
WHERE title = '日本のコンビニ'
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
VALUES
(
    '散歩',
    '散歩',
    'さんぽ',
    'sanpo',
    'noun',
    '散步；步行。',
    '散步',
    'N5',
    '朝、公園で散歩します。',
    '早上在公园散步。',
    'builtin',
    TRUE,
    0.98
),
(
    '予約',
    '予約',
    'よやく',
    'yoyaku',
    'noun',
    '预约；预订。',
    '预约',
    'N4',
    'ホテルを予約しました。',
    '我预订了酒店。',
    'builtin',
    TRUE,
    0.98
)
ON CONFLICT DO NOTHING;
