package service

import (
	"context"
	"time"
)

type MembershipRepository interface {
	GrantOrExtend(ctx context.Context, userID, redeemCodeID int64, validityDays int, now time.Time) (time.Time, error)
	GetActiveExpiry(ctx context.Context, userID int64, now time.Time) (*time.Time, error)
	ExpireDue(ctx context.Context, now time.Time, limit int) ([]int64, int64, error)
}
