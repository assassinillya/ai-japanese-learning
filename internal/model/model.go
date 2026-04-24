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
