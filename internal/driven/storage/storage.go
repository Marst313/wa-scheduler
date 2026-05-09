package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/ghazlabs/wa-scheduler/internal/core"
	"github.com/jmoiron/sqlx"
	"gopkg.in/validator.v2"
)

const (
	tableSchedule = "messages"
)

type StorageConfig struct {
	DB *sqlx.DB `validate:"nonnil"`
}

type Storage struct {
	StorageConfig
}

type messageRow struct {
	ID                 string             `db:"id"`
	Content            string             `db:"content"`
	RecipientNumbers   string             `db:"recipient_numbers"`
	ScheduledSendingAt int64              `db:"scheduled_sending_at"`
	SentAt             sql.NullInt64      `db:"sent_at"`
	RetriedCount       int                `db:"retried_count"`
	Status             core.MessageStatus `db:"status"`
	Reason             sql.NullString     `db:"reason"`
	CreatedAt          int64              `db:"created_at"`
	UpdatedAt          int64              `db:"updated_at"`
}

func (r messageRow) toCoreMessage() core.Message {
	msg := core.Message{
		ID:                 r.ID,
		Content:            r.Content,
		RecipientNumbers:   strings.Split(r.RecipientNumbers, ","),
		ScheduledSendingAt: r.ScheduledSendingAt,
		RetriedCount:       r.RetriedCount,
		Status:             r.Status,
		CreatedAt:          r.CreatedAt,
		UpdatedAt:          r.UpdatedAt,
	}

	if r.SentAt.Valid {
		sentAt := r.SentAt.Int64
		msg.SentAt = &sentAt
	}

	if r.Reason.Valid {
		reason := r.Reason.String
		msg.Reason = &reason
	}

	return msg
}

func NewStorage(cfg StorageConfig) (*Storage, error) {
	if err := validator.Validate(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	s := &Storage{
		StorageConfig: cfg,
	}

	if err := s.ensureSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return s, nil
}

func (s *Storage) ensureSchema() error {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		id TEXT PRIMARY KEY,
		content TEXT NOT NULL,
		recipient_numbers TEXT NOT NULL,
		scheduled_sending_at INTEGER,
		sent_at INTEGER,
		retried_count INTEGER DEFAULT 0,
		status TEXT,
		reason TEXT DEFAULT NULL,
		created_at INTEGER NOT NULL DEFAULT (strftime('%%s','now')),
		updated_at INTEGER NOT NULL DEFAULT (strftime('%%s','now'))
	);`, tableSchedule)

	if _, err := s.DB.ExecContext(context.Background(), query); err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	return nil
}

func (s *Storage) GetAllMessages(ctx context.Context, input core.GetAllMessagesInput) ([]core.Message, error) {
	query := fmt.Sprintf(`SELECT
		id,
		content,
		recipient_numbers,
		scheduled_sending_at,
		sent_at,
		retried_count,
		status,
		reason,
		created_at,
		updated_at
	FROM %s`, tableSchedule)

	var args []interface{}
	var conditions []string

	if input.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, input.Status)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY created_at DESC"

	var rows []messageRow
	if err := s.DB.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}

	messages := make([]core.Message, 0, len(rows))
	for _, row := range rows {
		messages = append(messages, row.toCoreMessage())
	}

	return messages, nil
}

func (s *Storage) GetMessage(ctx context.Context, id string) (*core.Message, error) {
	query := fmt.Sprintf(`SELECT
		id,
		content,
		recipient_numbers,
		scheduled_sending_at,
		sent_at,
		retried_count,
		status,
		reason,
		created_at,
		updated_at
	FROM %s WHERE id = ?`, tableSchedule)

	var row messageRow
	err := s.DB.GetContext(ctx, &row, query, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	msg := row.toCoreMessage()
	return &msg, nil
}

func (s *Storage) SaveMessage(ctx context.Context, message core.Message) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (
			id,
			content,
			scheduled_sending_at,
			recipient_numbers,
			status,
			created_at,
			updated_at
		)
		VALUES (:id, :content, :scheduled_sending_at, :recipient_numbers, :status, strftime('%%s','now'), strftime('%%s','now'))
	`, tableSchedule)

	params := map[string]interface{}{
		"id":                   message.ID,
		"content":              message.Content,
		"scheduled_sending_at": message.ScheduledSendingAt,
		"recipient_numbers":    strings.Join(message.RecipientNumbers, ","),
		"status":               message.Status,
	}

	_, err := s.DB.NamedExecContext(ctx, query, params)
	if err != nil {
		return fmt.Errorf("failed to insert message: %w", err)
	}
	return nil
}

func (s *Storage) UpdateMessage(ctx context.Context, message core.Message) error {
	query := fmt.Sprintf(`
		UPDATE %s
			SET scheduled_sending_at = :scheduled_sending_at,
				sent_at = :sent_at,
				retried_count = :retried_count,
				status = :status,
				reason = :reason,
				updated_at = strftime('%%s','now')
		WHERE id = :id
	`, tableSchedule)

	params := map[string]interface{}{
		"id":                   message.ID,
		"scheduled_sending_at": message.ScheduledSendingAt,
		"sent_at":              message.SentAt,
		"retried_count":        message.RetriedCount,
		"status":               message.Status,
		"reason":               message.Reason,
	}

	_, err := s.DB.NamedExecContext(ctx, query, params)
	if err != nil {
		return fmt.Errorf("failed to update message: %w", err)
	}
	return nil
}
