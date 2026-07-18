package repository

import (
	"context"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
)

type membershipRepository struct {
	client *dbent.Client
}

func NewMembershipRepository(client *dbent.Client) service.MembershipRepository {
	return &membershipRepository{client: client}
}

func (r *membershipRepository) GrantOrExtend(ctx context.Context, userID, redeemCodeID int64, validityDays int, now time.Time) (time.Time, error) {
	client := clientFromContext(ctx, r.client)
	rows, err := client.QueryContext(ctx, `
		INSERT INTO user_memberships
			(user_id, expires_at, status, last_redeem_code_id, created_at, updated_at)
		VALUES (
			$1,
			$2::timestamptz + ($3::integer * INTERVAL '1 day'),
			'active',
			$4,
			$2::timestamptz,
			$2::timestamptz
		)
		ON CONFLICT (user_id) DO UPDATE SET
			expires_at = GREATEST(user_memberships.expires_at, $2::timestamptz)
				+ ($3::integer * INTERVAL '1 day'),
			status = 'active',
			last_redeem_code_id = $4,
			updated_at = $2::timestamptz
		RETURNING expires_at
	`, userID, now.UTC(), validityDays, redeemCodeID)
	if err != nil {
		return time.Time{}, err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		return time.Time{}, rows.Err()
	}
	var expiresAt time.Time
	return expiresAt, rows.Scan(&expiresAt)
}

func (r *membershipRepository) GetActiveExpiry(ctx context.Context, userID int64, now time.Time) (*time.Time, error) {
	client := clientFromContext(ctx, r.client)
	rows, err := client.QueryContext(ctx, `
		SELECT expires_at FROM user_memberships
		WHERE user_id = $1 AND status = 'active' AND expires_at > $2
	`, userID, now.UTC())
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		return nil, rows.Err()
	}
	var expiresAt time.Time
	if err := rows.Scan(&expiresAt); err != nil {
		return nil, err
	}
	return &expiresAt, nil
}

func (r *membershipRepository) ExpireDue(ctx context.Context, now time.Time, limit int) ([]int64, int64, error) {
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
		SELECT user_id FROM user_memberships
		WHERE status = 'active' AND expires_at <= $1
		ORDER BY expires_at, user_id
		LIMIT $2 FOR UPDATE SKIP LOCKED
	`, now.UTC(), limit)
	if err != nil {
		return nil, 0, err
	}
	userIDs := make([]int64, 0, limit)
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			_ = rows.Close()
			return nil, 0, err
		}
		userIDs = append(userIDs, userID)
	}
	if err := rows.Close(); err != nil {
		return nil, 0, err
	}
	if len(userIDs) == 0 {
		return nil, 0, tx.Commit()
	}
	if _, err := client.ExecContext(ctx, `
		UPDATE user_memberships SET status = 'expired', updated_at = $1
		WHERE user_id = ANY($2)
	`, now.UTC(), pq.Array(userIDs)); err != nil {
		return nil, 0, err
	}
	result, err := client.ExecContext(ctx, `
		UPDATE api_keys AS k
		SET group_id = g.member_fallback_group_id, updated_at = $1
		FROM groups AS g
		WHERE k.user_id = ANY($2)
		  AND k.group_id = g.id
		  AND g.is_member_group = TRUE
		  AND g.member_fallback_group_id IS NOT NULL
		  AND k.deleted_at IS NULL
	`, now.UTC(), pq.Array(userIDs))
	if err != nil {
		return nil, 0, err
	}
	migrated, err := result.RowsAffected()
	if err != nil {
		return nil, 0, err
	}
	if err := tx.Commit(); err != nil {
		return nil, 0, err
	}
	return userIDs, migrated, nil
}
