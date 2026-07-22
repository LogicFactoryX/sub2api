package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
)

type modelPlazaRepository struct {
	db *sql.DB
}

func NewModelPlazaRepository(db *sql.DB) service.ModelPlazaRepository {
	return &modelPlazaRepository{db: db}
}

const modelPlazaSelect = `
	SELECT e.id, e.model_name, e.display_name, e.description, e.group_id,
	       e.tags, e.sort_order, e.enabled, e.created_at, e.updated_at,
	       g.name, g.platform, g.status, g.rate_multiplier,
	       g.is_exclusive, g.is_member_group
	FROM model_plaza_entries e
	JOIN groups g ON g.id = e.group_id`

func (r *modelPlazaRepository) List(ctx context.Context, publicOnly bool) ([]service.ModelPlazaEntry, error) {
	query := modelPlazaSelect
	if publicOnly {
		query += ` WHERE e.enabled = TRUE AND g.deleted_at IS NULL AND g.status = 'active'`
	}
	query += ` ORDER BY e.sort_order ASC, e.id ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]service.ModelPlazaEntry, 0)
	for rows.Next() {
		entry, err := scanModelPlazaEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, *entry)
	}
	return entries, rows.Err()
}

func (r *modelPlazaRepository) GetByID(ctx context.Context, id int64) (*service.ModelPlazaEntry, error) {
	entry, err := scanModelPlazaEntry(r.db.QueryRowContext(ctx, modelPlazaSelect+` WHERE e.id = $1`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrModelPlazaEntryNotFound
	}
	return entry, err
}

func (r *modelPlazaRepository) Create(ctx context.Context, entry *service.ModelPlazaEntry) error {
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO model_plaza_entries
			(model_name, display_name, description, group_id, tags, sort_order, enabled)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`,
		entry.ModelName, entry.DisplayName, entry.Description, entry.GroupID,
		pq.Array(entry.Tags), entry.SortOrder, entry.Enabled,
	).Scan(&entry.ID, &entry.CreatedAt, &entry.UpdatedAt)
	return mapModelPlazaWriteError(err)
}

func (r *modelPlazaRepository) Update(ctx context.Context, entry *service.ModelPlazaEntry) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE model_plaza_entries
		SET model_name = $2, display_name = $3, description = $4, group_id = $5,
		    tags = $6, sort_order = $7, enabled = $8, updated_at = NOW()
		WHERE id = $1`,
		entry.ID, entry.ModelName, entry.DisplayName, entry.Description, entry.GroupID,
		pq.Array(entry.Tags), entry.SortOrder, entry.Enabled,
	)
	if err = mapModelPlazaWriteError(err); err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return service.ErrModelPlazaEntryNotFound
	}
	return nil
}

func (r *modelPlazaRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM model_plaza_entries WHERE id = $1`, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return service.ErrModelPlazaEntryNotFound
	}
	return nil
}

type modelPlazaScanner interface {
	Scan(dest ...any) error
}

func scanModelPlazaEntry(scanner modelPlazaScanner) (*service.ModelPlazaEntry, error) {
	var entry service.ModelPlazaEntry
	if err := scanner.Scan(
		&entry.ID, &entry.ModelName, &entry.DisplayName, &entry.Description, &entry.GroupID,
		pq.Array(&entry.Tags), &entry.SortOrder, &entry.Enabled, &entry.CreatedAt, &entry.UpdatedAt,
		&entry.GroupName, &entry.Platform, &entry.GroupStatus, &entry.RateMultiplier,
		&entry.IsExclusive, &entry.IsMemberGroup,
	); err != nil {
		return nil, err
	}
	if entry.Tags == nil {
		entry.Tags = []string{}
	}
	return &entry, nil
}

func mapModelPlazaWriteError(err error) error {
	if err == nil {
		return nil
	}
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case "23505":
			return service.ErrModelPlazaEntryExists
		case "23503":
			return service.ErrGroupNotFound
		}
	}
	return fmt.Errorf("write model plaza entry: %w", err)
}
