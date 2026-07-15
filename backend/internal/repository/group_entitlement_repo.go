package repository

import (
	"context"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
)

type groupEntitlementRepository struct {
	client *dbent.Client
}

func NewGroupEntitlementRepository(client *dbent.Client) service.GroupEntitlementRepository {
	return &groupEntitlementRepository{client: client}
}

func (r *groupEntitlementRepository) Grant(ctx context.Context, userID, redeemCodeID, groupID, fallbackGroupID int64, startsAt, expiresAt time.Time) error {
	client := clientFromContext(ctx, r.client)
	_, err := client.ExecContext(ctx, `
		INSERT INTO user_group_entitlements
			(user_id, group_id, fallback_group_id, redeem_code_id, starts_at, expires_at, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, 'active', $5, $5)
	`, userID, groupID, fallbackGroupID, redeemCodeID, startsAt.UTC(), expiresAt.UTC())
	return err
}

func (r *groupEntitlementRepository) HasActive(ctx context.Context, userID, groupID int64, now time.Time) (bool, error) {
	client := clientFromContext(ctx, r.client)
	rows, err := client.QueryContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM user_group_entitlements
			WHERE user_id = $1 AND group_id = $2
			  AND status = 'active' AND expires_at > $3
		)
	`, userID, groupID, now.UTC())
	if err != nil {
		return false, err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		return false, nil
	}
	var active bool
	return active, rows.Scan(&active)
}

func (r *groupEntitlementRepository) ListActiveGroupExpiries(ctx context.Context, userID int64, now time.Time) (map[int64]time.Time, error) {
	client := clientFromContext(ctx, r.client)
	rows, err := client.QueryContext(ctx, `
		SELECT group_id, MAX(expires_at)
		FROM user_group_entitlements
		WHERE user_id = $1 AND status = 'active' AND expires_at > $2
		GROUP BY group_id
	`, userID, now.UTC())
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	expiries := make(map[int64]time.Time)
	for rows.Next() {
		var groupID int64
		var expiresAt time.Time
		if err := rows.Scan(&groupID, &expiresAt); err != nil {
			return nil, err
		}
		expiries[groupID] = expiresAt
	}
	return expiries, rows.Err()
}

func (r *groupEntitlementRepository) ExpireDue(ctx context.Context, now time.Time, limit int) ([]int64, int64, error) {
	if limit <= 0 {
		limit = 200
	}
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = tx.Rollback() }()
	client := tx.Client()

	rows, err := client.QueryContext(ctx, `
		SELECT id, user_id, group_id
		FROM user_group_entitlements
		WHERE status = 'active' AND expires_at <= $1
		ORDER BY expires_at, id
		LIMIT $2
		FOR UPDATE SKIP LOCKED
	`, now.UTC(), limit)
	if err != nil {
		return nil, 0, err
	}
	type pair struct{ userID, groupID int64 }
	ids := make([]int64, 0, limit)
	pairs := make(map[pair]struct{})
	for rows.Next() {
		var id, userID, groupID int64
		if err := rows.Scan(&id, &userID, &groupID); err != nil {
			_ = rows.Close()
			return nil, 0, err
		}
		ids = append(ids, id)
		pairs[pair{userID: userID, groupID: groupID}] = struct{}{}
	}
	if err := rows.Close(); err != nil {
		return nil, 0, err
	}
	if len(ids) == 0 {
		if err := tx.Commit(); err != nil {
			return nil, 0, err
		}
		return nil, 0, nil
	}

	if _, err := client.ExecContext(ctx, `
		UPDATE user_group_entitlements
		SET status = 'expired', expired_at = $1, updated_at = $1
		WHERE id = ANY($2)
	`, now.UTC(), pq.Array(ids)); err != nil {
		return nil, 0, err
	}

	affectedUsers := make(map[int64]struct{})
	var migrated int64
	for p := range pairs {
		checkRows, err := client.QueryContext(ctx, `
			SELECT
				EXISTS (SELECT 1 FROM user_allowed_groups WHERE user_id = $1 AND group_id = $2),
				EXISTS (SELECT 1 FROM user_group_entitlements WHERE user_id = $1 AND group_id = $2 AND status = 'active' AND expires_at > $3),
				(SELECT fallback_group_id FROM user_group_entitlements WHERE user_id = $1 AND group_id = $2 ORDER BY expires_at DESC, id DESC LIMIT 1)
		`, p.userID, p.groupID, now.UTC())
		if err != nil {
			return nil, 0, err
		}
		var permanent, stillActive bool
		var fallbackID int64
		if !checkRows.Next() {
			_ = checkRows.Close()
			continue
		}
		if err := checkRows.Scan(&permanent, &stillActive, &fallbackID); err != nil {
			_ = checkRows.Close()
			return nil, 0, err
		}
		_ = checkRows.Close()
		if permanent || stillActive {
			continue
		}
		result, err := client.ExecContext(ctx, `
			UPDATE api_keys SET group_id = $1, updated_at = $2
			WHERE user_id = $3 AND group_id = $4 AND deleted_at IS NULL
		`, fallbackID, now.UTC(), p.userID, p.groupID)
		if err != nil {
			return nil, 0, err
		}
		count, err := result.RowsAffected()
		if err != nil {
			return nil, 0, err
		}
		migrated += count
		affectedUsers[p.userID] = struct{}{}
	}

	if err := tx.Commit(); err != nil {
		return nil, 0, err
	}
	userIDs := make([]int64, 0, len(affectedUsers))
	for userID := range affectedUsers {
		userIDs = append(userIDs, userID)
	}
	return userIDs, migrated, nil
}
