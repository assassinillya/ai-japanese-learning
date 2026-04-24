package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"ai-japanese-learning/internal/model"
)

var ErrDictionaryEntryNotFound = errors.New("dictionary entry not found")

type DictionaryRepository struct {
	db *sql.DB
}

func NewDictionaryRepository(db *sql.DB) *DictionaryRepository {
	return &DictionaryRepository{db: db}
}

func (r *DictionaryRepository) FindByText(ctx context.Context, text string) (*model.DictionaryEntry, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, surface, lemma, reading, romaji, part_of_speech, meaning_zh, meaning_ja, meaning_en,
		       primary_meaning_zh, jlpt_level, example_sentence, example_translation_zh, conjugation_type,
		       is_common, source, verified, confidence_score::text, ai_model, prompt_version, created_at, updated_at
		FROM dictionary_entries
		WHERE surface = $1 OR lemma = $1
		ORDER BY verified DESC, source = 'builtin' DESC, id ASC
		LIMIT 1
	`, text)

	entry, err := scanDictionaryEntry(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDictionaryEntryNotFound
		}
		return nil, fmt.Errorf("find dictionary entry by text: %w", err)
	}
	return entry, nil
}

func (r *DictionaryRepository) GetByID(ctx context.Context, entryID int64) (*model.DictionaryEntry, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, surface, lemma, reading, romaji, part_of_speech, meaning_zh, meaning_ja, meaning_en,
		       primary_meaning_zh, jlpt_level, example_sentence, example_translation_zh, conjugation_type,
		       is_common, source, verified, confidence_score::text, ai_model, prompt_version, created_at, updated_at
		FROM dictionary_entries
		WHERE id = $1
	`, entryID)

	entry, err := scanDictionaryEntry(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDictionaryEntryNotFound
		}
		return nil, fmt.Errorf("get dictionary entry: %w", err)
	}
	return entry, nil
}

func (r *DictionaryRepository) Create(ctx context.Context, entry *model.DictionaryEntry) (*model.DictionaryEntry, error) {
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO dictionary_entries (
			surface, lemma, reading, romaji, part_of_speech, meaning_zh, meaning_ja, meaning_en,
			primary_meaning_zh, jlpt_level, example_sentence, example_translation_zh, conjugation_type,
			is_common, source, verified, confidence_score, ai_model, prompt_version
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
		RETURNING id, created_at, updated_at, confidence_score::text
	`,
		entry.Surface,
		entry.Lemma,
		entry.Reading,
		entry.Romaji,
		entry.PartOfSpeech,
		entry.MeaningZH,
		entry.MeaningJA,
		entry.MeaningEN,
		entry.PrimaryMeaningZH,
		entry.JLPTLevel,
		entry.ExampleSentence,
		entry.ExampleTranslationZH,
		entry.ConjugationType,
		entry.IsCommon,
		entry.Source,
		entry.Verified,
		entry.ConfidenceScore,
		entry.AIModel,
		entry.PromptVersion,
	).Scan(&entry.ID, &entry.CreatedAt, &entry.UpdatedAt, &entry.ConfidenceScore)
	if err != nil {
		return nil, fmt.Errorf("create dictionary entry: %w", err)
	}
	return entry, nil
}

type dictionaryEntryScanner interface {
	Scan(dest ...any) error
}

func scanDictionaryEntry(scanner dictionaryEntryScanner) (*model.DictionaryEntry, error) {
	var entry model.DictionaryEntry
	if err := scanner.Scan(
		&entry.ID,
		&entry.Surface,
		&entry.Lemma,
		&entry.Reading,
		&entry.Romaji,
		&entry.PartOfSpeech,
		&entry.MeaningZH,
		&entry.MeaningJA,
		&entry.MeaningEN,
		&entry.PrimaryMeaningZH,
		&entry.JLPTLevel,
		&entry.ExampleSentence,
		&entry.ExampleTranslationZH,
		&entry.ConjugationType,
		&entry.IsCommon,
		&entry.Source,
		&entry.Verified,
		&entry.ConfidenceScore,
		&entry.AIModel,
		&entry.PromptVersion,
		&entry.CreatedAt,
		&entry.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &entry, nil
}
