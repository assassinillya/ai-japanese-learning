package service

import "unicode"

type LanguageService struct{}

func NewLanguageService() *LanguageService {
	return &LanguageService{}
}

func (s *LanguageService) Detect(content string) string {
	var japaneseRunes, cjkRunes, latinRunes int
	for _, r := range content {
		switch {
		case unicode.In(r, unicode.Hiragana, unicode.Katakana):
			japaneseRunes++
		case unicode.In(r, unicode.Han):
			cjkRunes++
		case unicode.In(r, unicode.Latin):
			latinRunes++
		}
	}

	switch {
	case japaneseRunes > 0:
		return "ja"
	case cjkRunes > latinRunes && cjkRunes > 0:
		return "zh"
	case latinRunes > 0:
		return "en"
	default:
		return "unknown"
	}
}
