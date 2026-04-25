package model

import "time"

type JLPTLevel string

const (
	JLPTN5 JLPTLevel = "N5"
	JLPTN4 JLPTLevel = "N4"
	JLPTN3 JLPTLevel = "N3"
	JLPTN2 JLPTLevel = "N2"
	JLPTN1 JLPTLevel = "N1"
)

func IsValidJLPT(level JLPTLevel) bool {
	switch level {
	case JLPTN5, JLPTN4, JLPTN3, JLPTN2, JLPTN1:
		return true
	default:
		return false
	}
}

type User struct {
	ID                  int64      `json:"id"`
	Email               string     `json:"email"`
	Username            string     `json:"username"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
	JLPTLevel           JLPTLevel  `json:"jlpt_level"`
	OnboardingCompleted bool       `json:"onboarding_completed"`
	LastLoginAt         *time.Time `json:"last_login_at,omitempty"`
}

type TranslationStatus string

const (
	TranslationPending    TranslationStatus = "pending"
	TranslationProcessing TranslationStatus = "processing"
	TranslationDone       TranslationStatus = "done"
	TranslationFailed     TranslationStatus = "failed"
)

type AuthSession struct {
	ID        int64
	UserID    int64
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type Article struct {
	ID                 int64             `json:"id"`
	UserID             *int64            `json:"user_id,omitempty"`
	Title              string            `json:"title"`
	OriginalLanguage   string            `json:"original_language"`
	OriginalContent    *string           `json:"original_content,omitempty"`
	JapaneseContent    string            `json:"japanese_content"`
	ChineseTranslation *string           `json:"chinese_translation,omitempty"`
	JLPTLevel          JLPTLevel         `json:"jlpt_level"`
	SourceType         string            `json:"source_type"`
	IsAIGenerated      bool              `json:"is_ai_generated"`
	IsVerified         bool              `json:"is_verified"`
	TranslationStatus  TranslationStatus `json:"translation_status"`
	ProcessingNotes    *string           `json:"processing_notes,omitempty"`
	SentenceCount      int               `json:"sentence_count"`
	CreatedAt          time.Time         `json:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at"`
}

type ArticleSentence struct {
	ID            int64     `json:"id"`
	ArticleID     int64     `json:"article_id"`
	SentenceOrder int       `json:"sentence_order"`
	SentenceText  string    `json:"sentence_text"`
	TranslationZH *string   `json:"translation_zh,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type DictionaryEntry struct {
	ID                   int64     `json:"id"`
	Surface              string    `json:"surface"`
	Lemma                string    `json:"lemma"`
	Reading              string    `json:"reading"`
	Romaji               *string   `json:"romaji,omitempty"`
	PartOfSpeech         string    `json:"part_of_speech"`
	MeaningZH            string    `json:"meaning_zh"`
	MeaningJA            *string   `json:"meaning_ja,omitempty"`
	MeaningEN            *string   `json:"meaning_en,omitempty"`
	PrimaryMeaningZH     string    `json:"primary_meaning_zh"`
	JLPTLevel            string    `json:"jlpt_level"`
	ExampleSentence      *string   `json:"example_sentence,omitempty"`
	ExampleTranslationZH *string   `json:"example_translation_zh,omitempty"`
	ConjugationType      *string   `json:"conjugation_type,omitempty"`
	IsCommon             bool      `json:"is_common"`
	Source               string    `json:"source"`
	Verified             bool      `json:"verified"`
	ConfidenceScore      string    `json:"confidence_score"`
	AIModel              *string   `json:"ai_model,omitempty"`
	PromptVersion        *string   `json:"prompt_version,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type VocabularyStatus string

const (
	VocabularyNew       VocabularyStatus = "new"
	VocabularyLearning  VocabularyStatus = "learning"
	VocabularyReviewing VocabularyStatus = "reviewing"
	VocabularyMastered  VocabularyStatus = "mastered"
	VocabularyIgnored   VocabularyStatus = "ignored"
)

type UserVocabulary struct {
	ID                      int64            `json:"id"`
	UserID                  int64            `json:"user_id"`
	DictionaryEntryID       int64            `json:"dictionary_entry_id"`
	ArticleID               *int64           `json:"article_id,omitempty"`
	SourceSentenceID        *int64           `json:"source_sentence_id,omitempty"`
	SelectedText            string           `json:"selected_text"`
	SourceSentenceText      string           `json:"source_sentence_text"`
	Status                  VocabularyStatus `json:"status"`
	Familiarity             int              `json:"familiarity"`
	CorrectCount            int              `json:"correct_count"`
	WrongCount              int              `json:"wrong_count"`
	ConsecutiveCorrectCount int              `json:"consecutive_correct_count"`
	AddedAt                 time.Time        `json:"added_at"`
	LastReviewedAt          *time.Time       `json:"last_reviewed_at,omitempty"`
	NextReviewAt            time.Time        `json:"next_review_at"`
	CreatedAt               time.Time        `json:"created_at"`
	UpdatedAt               time.Time        `json:"updated_at"`
}

type VocabularyDetail struct {
	Item            UserVocabulary  `json:"item"`
	DictionaryEntry DictionaryEntry `json:"dictionary_entry"`
	ArticleTitle    *string         `json:"article_title,omitempty"`
	ExampleSentence string          `json:"example_sentence"`
}

type ChallengeQuestion struct {
	ID                int64     `json:"id"`
	ArticleID         int64     `json:"article_id"`
	SentenceID        int64     `json:"sentence_id"`
	QuestionType      string    `json:"question_type"`
	QuestionOrder     int       `json:"question_order"`
	SentenceText      string    `json:"sentence_text"`
	MaskedSentence    string    `json:"masked_sentence"`
	CorrectEntryID    int64     `json:"correct_entry_id"`
	CorrectAnswerText string    `json:"correct_answer_text"`
	OptionA           string    `json:"option_a"`
	OptionB           string    `json:"option_b"`
	OptionC           string    `json:"option_c"`
	OptionD           string    `json:"option_d"`
	CorrectOption     string    `json:"correct_option"`
	Explanation       string    `json:"explanation"`
	JLPTLevel         string    `json:"jlpt_level"`
	AIModel           *string   `json:"ai_model,omitempty"`
	PromptVersion     *string   `json:"prompt_version,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

type ChallengeQuestionAttempt struct {
	ID             int64     `json:"id"`
	QuestionID     int64     `json:"question_id"`
	UserID         int64     `json:"user_id"`
	SelectedOption string    `json:"selected_option"`
	IsCorrect      bool      `json:"is_correct"`
	AnsweredAt     time.Time `json:"answered_at"`
}

type VocabularyReviewQuestion struct {
	ID                int64     `json:"id"`
	DictionaryEntryID int64     `json:"dictionary_entry_id"`
	QuestionText      string    `json:"question_text"`
	CorrectAnswer     string    `json:"correct_answer"`
	OptionA           string    `json:"option_a"`
	OptionB           string    `json:"option_b"`
	OptionC           string    `json:"option_c"`
	OptionD           string    `json:"option_d"`
	CorrectOption     string    `json:"correct_option"`
	ExplanationZH     string    `json:"explanation_zh"`
	AIModel           *string   `json:"ai_model,omitempty"`
	PromptVersion     *string   `json:"prompt_version,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

type VocabularyReviewItem struct {
	UserVocabulary  UserVocabulary           `json:"user_vocabulary"`
	Dictionary      DictionaryEntry          `json:"dictionary_entry"`
	Question        VocabularyReviewQuestion `json:"question"`
	ArticleTitle    *string                  `json:"article_title,omitempty"`
	ContextSentence string                   `json:"context_sentence"`
}

type VocabularyReviewRecord struct {
	ID               int64     `json:"id"`
	UserID           int64     `json:"user_id"`
	UserVocabularyID int64     `json:"user_vocabulary_id"`
	ReviewQuestionID int64     `json:"review_question_id"`
	SelectedOption   string    `json:"selected_option"`
	IsCorrect        bool      `json:"is_correct"`
	ReviewedAt       time.Time `json:"reviewed_at"`
}
