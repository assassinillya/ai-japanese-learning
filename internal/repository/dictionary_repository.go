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

func (r *DictionaryRepository) EnsureExampleTable(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS dictionary_examples (
			id BIGSERIAL PRIMARY KEY,
			dictionary_entry_id BIGINT NOT NULL REFERENCES dictionary_entries(id) ON DELETE CASCADE,
			example_sentence TEXT NOT NULL,
			example_translation_zh TEXT,
			source TEXT NOT NULL CHECK (source IN ('ai', 'admin', 'builtin')),
			ai_model TEXT,
			prompt_version TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		CREATE TABLE IF NOT EXISTS dictionary_forms (
			id BIGSERIAL PRIMARY KEY,
			dictionary_entry_id BIGINT NOT NULL REFERENCES dictionary_entries(id) ON DELETE CASCADE,
			form TEXT NOT NULL UNIQUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("ensure dictionary examples table: %w", err)
	}
	return nil
}

func (r *DictionaryRepository) FindByText(ctx context.Context, text string) (*model.DictionaryEntry, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT de.id, de.surface, de.lemma, de.reading, de.romaji, de.part_of_speech, de.meaning_zh, de.meaning_ja, de.meaning_en,
		       de.primary_meaning_zh, de.jlpt_level, de.example_sentence, de.example_translation_zh, de.conjugation_type,
		       de.is_common, de.source, de.verified, de.confidence_score::text, de.ai_model, de.prompt_version, de.created_at, de.updated_at
		FROM dictionary_entries de
		LEFT JOIN dictionary_forms df ON df.dictionary_entry_id = de.id
		WHERE de.surface = $1 OR de.lemma = $1 OR df.form = $1
		ORDER BY de.verified DESC, de.source = 'builtin' DESC, de.id ASC
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

func (r *DictionaryRepository) CreateForm(ctx context.Context, entryID int64, form string) error {
	form = strings.TrimSpace(form)
	if form == "" {
		return nil
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO dictionary_forms (dictionary_entry_id, form)
		VALUES ($1, $2)
		ON CONFLICT (form) DO UPDATE SET dictionary_entry_id = EXCLUDED.dictionary_entry_id
	`, entryID, form)
	if err != nil {
		return fmt.Errorf("create dictionary form: %w", err)
	}
	return nil
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

func (r *DictionaryRepository) ListExamples(ctx context.Context, entryID int64) ([]model.DictionaryExample, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, dictionary_entry_id, example_sentence, example_translation_zh, source, ai_model, prompt_version, created_at
		FROM dictionary_examples
		WHERE dictionary_entry_id = $1
		ORDER BY created_at ASC, id ASC
	`, entryID)
	if err != nil {
		return nil, fmt.Errorf("list dictionary examples: %w", err)
	}
	defer rows.Close()
	var examples []model.DictionaryExample
	for rows.Next() {
		example, err := scanDictionaryExample(rows)
		if err != nil {
			return nil, err
		}
		examples = append(examples, *example)
	}
	return examples, rows.Err()
}

func (r *DictionaryRepository) CreateExample(ctx context.Context, example *model.DictionaryExample) (*model.DictionaryExample, error) {
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO dictionary_examples (
			dictionary_entry_id, example_sentence, example_translation_zh, source, ai_model, prompt_version
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`, example.DictionaryEntryID, example.ExampleSentence, example.ExampleTranslationZH, example.Source, example.AIModel, example.PromptVersion).
		Scan(&example.ID, &example.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create dictionary example: %w", err)
	}
	return example, nil
}

func (r *DictionaryRepository) DeleteExample(ctx context.Context, exampleID int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM dictionary_examples WHERE id = $1`, exampleID)
	if err != nil {
		return fmt.Errorf("delete dictionary example: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected for delete dictionary example: %w", err)
	}
	if rows == 0 {
		return ErrDictionaryEntryNotFound
	}
	return nil
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

func scanDictionaryExample(scanner dictionaryEntryScanner) (*model.DictionaryExample, error) {
	var example model.DictionaryExample
	if err := scanner.Scan(
		&example.ID,
		&example.DictionaryEntryID,
		&example.ExampleSentence,
		&example.ExampleTranslationZH,
		&example.Source,
		&example.AIModel,
		&example.PromptVersion,
		&example.CreatedAt,
	); err != nil {
		return nil, err
	}
	return &example, nil
}
