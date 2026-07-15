package service

import (
	"context"
	"time"
)

type GroupEntitlementRepository interface {
	Grant(ctx context.Context, userID, redeemCodeID, groupID, fallbackGroupID int64, startsAt, expiresAt time.Time) error
	HasActive(ctx context.Context, userID, groupID int64, now time.Time) (bool, error)
	ListActiveGroupExpiries(ctx context.Context, userID int64, now time.Time) (map[int64]time.Time, error)
	ExpireDue(ctx context.Context, now time.Time, limit int) ([]int64, int64, error)
}
