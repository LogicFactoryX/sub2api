//go:build integration

package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestMembershipRepositoryGrantOrExtend(t *testing.T) {
	ctx := context.Background()
	tx := testEntTx(t)
	txCtx := dbent.NewTxContext(ctx, tx)
	client := tx.Client()

	user, err := client.User.Create().
		SetEmail(fmt.Sprintf("membership-redeem-%d@example.com", time.Now().UnixNano())).
		SetPasswordHash("test-password-hash").
		Save(txCtx)
	require.NoError(t, err)

	createCode := func(code string) *dbent.RedeemCode {
		t.Helper()
		created, createErr := client.RedeemCode.Create().
			SetCode(code).
			SetType(service.RedeemTypeMembership).
			SetStatus(service.StatusUnused).
			SetValue(0).
			SetValidityDays(7).
			Save(txCtx)
		require.NoError(t, createErr)
		return created
	}

	repo := NewMembershipRepository(client)
	now := time.Now().UTC().Truncate(time.Microsecond)
	firstCode := createCode(fmt.Sprintf("member-first-%d", time.Now().UnixNano()))
	firstExpiry, err := repo.GrantOrExtend(txCtx, user.ID, firstCode.ID, 7, now)
	require.NoError(t, err)
	require.WithinDuration(t, now.Add(7*24*time.Hour), firstExpiry, time.Second)

	secondCode := createCode(fmt.Sprintf("member-next-%d", time.Now().UnixNano()))
	secondExpiry, err := repo.GrantOrExtend(txCtx, user.ID, secondCode.ID, 3, now.Add(time.Hour))
	require.NoError(t, err)
	require.WithinDuration(t, firstExpiry.Add(3*24*time.Hour), secondExpiry, time.Second)
}
