package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/ghazlabs/wa-scheduler/internal/core"
	"gopkg.in/validator.v2"
)

const (
	tableSchedule = "messages"
)

type StorageConfig struct {
	DB *sql.DB `validate:"nonnil"`
}

type Storage struct {
	StorageConfig
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

	if _, err := s.DB.Exec(query); err != nil {
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

	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	messages := make([]core.Message, 0)
	for rows.Next() {
		var msg core.Message
		var recipientNumbers string
		var sentAt sql.NullInt64
		var reason sql.NullString

		err := rows.Scan(
			&msg.ID,
			&msg.Content,
			&recipientNumbers,
			&msg.ScheduledSendingAt,
			&sentAt,
			&msg.RetriedCount,
			&msg.Status,
			&reason,
			&msg.CreatedAt,
			&msg.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		msg.RecipientNumbers = strings.Split(recipientNumbers, ",")
		if sentAt.Valid {
			msg.SentAt = &sentAt.Int64
		}
		if reason.Valid {
			msg.Reason = &reason.String
		}

		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
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

	row := s.DB.QueryRowContext(ctx, query, id)

	var msg core.Message
	var recipientNumbers string
	var sentAt sql.NullInt64
	var reason sql.NullString

	err := row.Scan(
		&msg.ID,
		&msg.Content,
		&recipientNumbers,
		&msg.ScheduledSendingAt,
		&sentAt,
		&msg.RetriedCount,
		&msg.Status,
		&reason,
		&msg.CreatedAt,
		&msg.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	msg.RecipientNumbers = strings.Split(recipientNumbers, ",")
	if sentAt.Valid {
		msg.SentAt = &sentAt.Int64
	}
	if reason.Valid {
		msg.Reason = &reason.String
	}

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
		VALUES (?, ?, ?, ?, ?, strftime('%%s','now'), strftime('%%s','now'))
	`, tableSchedule)

	_, err := s.DB.ExecContext(ctx, query,
		message.ID,
		message.Content,
		message.ScheduledSendingAt,
		strings.Join(message.RecipientNumbers, ","),
		message.Status,
	)
	if err != nil {
		return fmt.Errorf("failed to insert message: %w", err)
	}
	return nil
}

func (s *Storage) UpdateMessage(ctx context.Context, message core.Message) error {
	query := fmt.Sprintf(`
		UPDATE %s
			SET scheduled_sending_at = ?,
				sent_at = ?,
				retried_count = ?,
				status = ?,
				reason = ?,
				updated_at = strftime('%%s','now')
		WHERE id = ?
	`, tableSchedule)

	_, err := s.DB.ExecContext(ctx, query,
		message.ScheduledSendingAt,
		message.SentAt,
		message.RetriedCount,
		message.Status,
		message.Reason,
		message.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update message: %w", err)
	}
	return nil
}
