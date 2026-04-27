package service

import (
	"encoding/json"
	"fmt"

	"ai-japanese-learning/internal/model"
)

const aiPromptVersionV12 = "v1.3"

func promptDictionaryEntry(text string, contextText string) AIPrompt {
	return AIPrompt{
		System: "你是日语词典生成器。只返回合法 JSON，不要 Markdown，不要额外解释。字段必须完整。你必须结合上下文判断用户真正想查的完整日语词、固定用法或文法表达；如果用户只划中词的一部分，要返回上下文中的完整词/表达。动词、形容词或助动词变形的 lemma 必须返回词典形/原型；固定用法或文法表达的 part_of_speech 必须返回 grammar，例如 そういえば；普通词汇按实际词性返回。surface 也优先返回原型；不确定 JLPT 时返回 unknown；source 必须为 ai；verified 必须为 false；confidence_score 用字符串小数。",
		User: fmt.Sprintf(`为查询文本生成日语学习词典词条。

查询文本：%s
上下文：%s

要求：
- 如果查询文本只是完整表达的一部分，请根据上下文返回完整词或完整惯用表达。
- 例如上下文是「そういえば」而用户只划中「いえ」，应返回「そういえば」而不是「いえ」。
- 请判断它是文法/固定用法还是普通单词，例如「そういえば」标为 grammar，「広い」标为 adjective。
- surface 和 lemma 应尽量是适合加入生词本和复习的标准词形。

返回 JSON 格式：
{
  "surface": "用户查询的表层形式",
  "lemma": "标准词条或原形",
  "reading": "假名读音",
  "romaji": "罗马音",
  "part_of_speech": "noun/verb/adjective/adverb/expression/grammar/unknown",
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
}`, text, contextText),
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
  "chinese_summary": "用简体中文一句话概括文章内容，30到80字",
  "source_type": "ai_translated",
  "is_ai_generated": true,
  "note": "用简体中文简短说明本次处理"
}`, level, language, content),
	}
}

func promptArticleSummaryZH(title, content string) AIPrompt {
	return AIPrompt{
		System: "你是日语学习文章简介生成器。只返回合法 JSON，不要 Markdown。简介面向中文用户，必须简短、自然、说明文章大概内容。",
		User: fmt.Sprintf(`请为下面文章生成简体中文简介。

标题：%s
文章：
%s

返回 JSON 格式：
{
  "chinese_summary": "30到80字，概括文章讲了什么，不要出现 ja.20、文件名、处理流水号等技术信息"
}`, title, content),
	}
}

func promptChallengeQuestions(request challengeQuestionCacheRequest) AIPrompt {
	raw, _ := json.Marshal(request)
	return AIPrompt{
		System: "你是 JLPT 日语阅读考点分析器。只返回合法 JSON，不要 Markdown。你不出题，只根据文章推荐最可能作为 JLPT 考点的重点词汇、固定用法和语法。",
		User: fmt.Sprintf(`请以 JLPT 考点分析的方式，从文章中选出 3 到 5 个最值得学习、最可能在 JLPT 阅读或词汇语法题中考到的重点词汇或短语，并选出 3 到 5 个最可能考试的重点语法、句型或固定用法。两类都要按“JLPT 考点重要度”从高到低排序。

选择规则：
- 优先选择 JLPT 高频词、常见考点词、易混词、固定搭配、接续表达、句型和文法。
- 不要选择过于简单、只在本文偶然出现、没有考试价值的片段。
- explanation 必须说明它为什么是 JLPT 考点，以及在本文中的意思或用法。

输入 JSON：
%s

返回 JSON 格式：
{
  "items": [
    {
      "sentence_id": 1,
      "sentence_text": "包含该词的原句",
      "masked_sentence": "推荐词汇",
      "correct_answer_text": "推荐词汇的词典形或标准表达",
      "option_a": "JLPT 等级，例如 N3",
      "option_b": "文章内出现频次，例如 1",
      "option_c": "JLPT 考点重要度：高/中/低",
      "option_d": "类型：vocabulary 或 grammar",
      "correct_option": "A",
      "explanation": "中文释义、用法说明和 JLPT 考点理由"
    }
  ]
}`, string(raw)),
	}
}

func promptPostQuizQuestions(request challengeQuestionCacheRequest) AIPrompt {
	raw, _ := json.Marshal(request)
	return AIPrompt{
		System: "你是 JLPT 日语阅读理解题命题老师。只返回合法 JSON，不要 Markdown。必须基于整篇文章命题，题型遵循 JLPT 阅读理解，不要出填空题、挖空题、单词释义题或语法选择题。题干、选项和正确答案必须使用自然日语；解析必须使用简体中文，方便中文用户复盘。",
		User: fmt.Sprintf(`请基于整篇文章生成 3 到 5 道 JLPT 阅读理解四选一题，难度匹配 JLPT %s。

命题规则：
- 题目必须像 JLPT 阅读理解题，围绕文章内容提问，例如「ラジャさんは先週何をしましたか。」。
- 可以出主旨题、细节题、原因题、指代题、作者意图题、态度/立意题、段落关系题、推论题。
- 禁止把原句挖空，禁止要求选择一个词填入空格，禁止单纯考单词中文意思。
- 必须阅读 full_text 整篇文章后出题，不要只根据单句局部出题。
- sentence_text 填写支持答案的关键句、段落或证据；explanation 用简体中文说明依据。
- 如果 existing_questions 里已有题干，追加题目时不要重复相同考点或相同题干。
- masked_sentence、correct_answer_text、option_a、option_b、option_c、option_d 全部必须是日语。
- explanation 必须是简体中文解析，说明正确答案对应的原文依据和干扰项问题。
- 选项必须是内容理解答案，不要只给单词或短语。

输入 JSON：
%s

返回 JSON 格式：
{
  "items": [
    {
      "sentence_id": 1,
      "sentence_text": "支持答案的原文关键句或段落",
      "masked_sentence": "阅读理解题干，不要挖空",
      "correct_answer_text": "正确选项文本",
      "option_a": "日本語の選択肢A",
      "option_b": "日本語の選択肢B",
      "option_c": "日本語の選択肢C",
      "option_d": "日本語の選択肢D",
      "correct_option": "A",
      "explanation": "中文解析"
    }
  ]
}`, request.JLPTLevel, string(raw)),
	}
}

func promptReviewQuestion(entry model.DictionaryEntry, questionOrder int) AIPrompt {
	raw, _ := json.Marshal(entry)
	return AIPrompt{
		System: "你是日语词汇复习题生成器。只返回合法 JSON，不要 Markdown。正确答案必须等于词典 primary_meaning_zh。",
		User: fmt.Sprintf(`请为这个词典条目生成第 %d 道中文释义四选一复习题。

要求：
- 同一个词会生成 3 道题，请让第 %d 道题的题干表达和干扰项角度与其他题不同。
- 题型仍然是识别目标词的中文释义，不要改变正确答案。

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
}`, questionOrder, questionOrder, string(raw)),
	}
}
