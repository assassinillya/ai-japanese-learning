package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

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

func (r *DictionaryRepository) ListAll(ctx context.Context) ([]model.DictionaryEntry, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, surface, lemma, reading, romaji, part_of_speech, meaning_zh, meaning_ja, meaning_en,
		       primary_meaning_zh, jlpt_level, example_sentence, example_translation_zh, conjugation_type,
		       is_common, source, verified, confidence_score::text, ai_model, prompt_version, created_at, updated_at
		FROM dictionary_entries
		ORDER BY verified DESC, source = 'builtin' DESC, id ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list dictionary entries: %w", err)
	}
	defer rows.Close()

	var entries []model.DictionaryEntry
	for rows.Next() {
		entry, err := scanDictionaryEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, *entry)
	}
	return entries, rows.Err()
}

func (r *DictionaryRepository) ListDistractors(ctx context.Context, excludeID int64, partOfSpeech string, limit int) ([]model.DictionaryEntry, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, surface, lemma, reading, romaji, part_of_speech, meaning_zh, meaning_ja, meaning_en,
		       primary_meaning_zh, jlpt_level, example_sentence, example_translation_zh, conjugation_type,
		       is_common, source, verified, confidence_score::text, ai_model, prompt_version, created_at, updated_at
		FROM dictionary_entries
		WHERE id <> $1
		  AND ($2 = '' OR part_of_speech = $2 OR part_of_speech = 'unknown')
		ORDER BY verified DESC, source = 'builtin' DESC, id ASC
		LIMIT $3
	`, excludeID, strings.TrimSpace(partOfSpeech), limit)
	if err != nil {
		return nil, fmt.Errorf("list distractor dictionary entries: %w", err)
	}
	defer rows.Close()

	var entries []model.DictionaryEntry
	for rows.Next() {
		entry, err := scanDictionaryEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, *entry)
	}
	return entries, rows.Err()
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
