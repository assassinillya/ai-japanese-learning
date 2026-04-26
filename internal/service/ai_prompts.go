package service

import (
	"encoding/json"
	"fmt"

	"ai-japanese-learning/internal/model"
)

const aiPromptVersionV12 = "v1.2"

func promptDictionaryEntry(text string) AIPrompt {
	return AIPrompt{
		System: "你是日语词典生成器。只返回合法 JSON，不要 Markdown，不要额外解释。字段必须完整；不确定 JLPT 时返回 unknown；source 必须为 ai；verified 必须为 false；confidence_score 用字符串小数。",
		User: fmt.Sprintf(`为查询文本生成日语学习词典词条。

查询文本：%s

返回 JSON 格式：
{
  "surface": "用户查询的表层形式",
  "lemma": "标准词条或原形",
  "reading": "假名读音",
  "romaji": "罗马音",
  "part_of_speech": "noun/verb/adjective/adverb/expression/unknown",
  "meaning_zh": "中文释义，可以较完整",
  "meaning_ja": "日文释义，可为空字符串",
  "meaning_en": "英文释义，可为空字符串",
  "primary_meaning_zh": "主要中文意思，用于复习答案",
  "jlpt_level": "N5/N4/N3/N2/N1/unknown",
  "example_sentence": "自然日语例句",
  "example_translation_zh": "例句中文翻译",
  "conjugation_type": "活用类型，可为空字符串",
  "is_common": true,
  "source": "ai",
  "verified": false,
  "confidence_score": "0.80"
}`, text),
	}
}

func promptDictionaryExample(entry model.DictionaryEntry, existing []model.DictionaryExample) AIPrompt {
	raw, _ := json.Marshal(existing)
	return AIPrompt{
		System: "你是日语例句生成器。只返回合法 JSON，不要 Markdown。例句必须自然、简短，适合中文用户学习。",
		User: fmt.Sprintf(`请为下面日语词条生成 1 个新的日语例句，避免与已有例句重复。

词条：
- surface: %s
- lemma: %s
- reading: %s
- part_of_speech: %s
- meaning_zh: %s
- jlpt_level: %s

已有例句 JSON：
%s

返回 JSON 格式：
{
  "example_sentence": "包含该词的自然日语例句",
  "example_translation_zh": "例句中文翻译"
}`, entry.Surface, entry.Lemma, entry.Reading, entry.PartOfSpeech, entry.MeaningZH, entry.JLPTLevel, string(raw)),
	}
}

func promptArticleTranslation(language, content string, level model.JLPTLevel) AIPrompt {
	return AIPrompt{
		System: "你是面向中文用户的日语学习文章改写/翻译器。只返回合法 JSON，不要 Markdown，不要额外解释。",
		User: fmt.Sprintf(`将下面文章翻译或改写为适合 JLPT %s 学习者阅读的自然日语。

原文语言：%s
原文：
%s

返回 JSON 格式：
{
  "japanese_content": "日语文章，按自然句子组织",
  "source_type": "ai_translated",
  "is_ai_generated": true,
  "note": "简短处理说明"
}`, level, language, content),
	}
}

func promptChallengeQuestions(request challengeQuestionCacheRequest) AIPrompt {
	raw, _ := json.Marshal(request)
	return AIPrompt{
		System: "你是日语阅读选择题生成器。只返回合法 JSON，不要 Markdown。题目必须基于输入句子，选项必须 4 个且只有一个正确答案。",
		User: fmt.Sprintf(`请为挑战阅读生成挖空选择题，难度匹配 JLPT %s。干扰项尽量同词性、同难度。

输入 JSON：
%s

返回 JSON 格式：
{
  "items": [
    {
      "sentence_id": 1,
      "sentence_text": "原句",
      "masked_sentence": "挖空后的句子",
      "correct_answer_text": "正确词形",
      "option_a": "选项A",
      "option_b": "选项B",
      "option_c": "选项C",
      "option_d": "选项D",
      "correct_option": "A",
      "explanation": "中文解析"
    }
  ]
}`, request.JLPTLevel, string(raw)),
	}
}

func promptPostQuizQuestions(request challengeQuestionCacheRequest) AIPrompt {
	raw, _ := json.Marshal(request)
	return AIPrompt{
		System: "你是日语阅读后测验生成器。只返回合法 JSON，不要 Markdown。题目围绕文章中的重点词汇或句意理解。",
		User: fmt.Sprintf(`请为阅读后测验生成中文释义四选一题，难度匹配 JLPT %s。

输入 JSON：
%s

返回 JSON 格式：
{
  "items": [
    {
      "sentence_id": 1,
      "sentence_text": "原句",
      "masked_sentence": "题干",
      "correct_answer_text": "正确中文答案",
      "option_a": "选项A",
      "option_b": "选项B",
      "option_c": "选项C",
      "option_d": "选项D",
      "correct_option": "A",
      "explanation": "中文解析"
    }
  ]
}`, request.JLPTLevel, string(raw)),
	}
}

func promptReviewQuestion(entry model.DictionaryEntry) AIPrompt {
	raw, _ := json.Marshal(entry)
	return AIPrompt{
		System: "你是日语词汇复习题生成器。只返回合法 JSON，不要 Markdown。正确答案必须等于词典 primary_meaning_zh。",
		User: fmt.Sprintf(`请为这个词典条目生成中文释义四选一复习题。

词典条目 JSON：
%s

返回 JSON 格式：
{
  "question_text": "目标单词",
  "correct_answer": "必须等于 primary_meaning_zh",
  "option_a": "选项A",
  "option_b": "选项B",
  "option_c": "选项C",
  "option_d": "选项D",
  "correct_option": "A",
  "explanation_zh": "中文解析"
}`, string(raw)),
	}
}
