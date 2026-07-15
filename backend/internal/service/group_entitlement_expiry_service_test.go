package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type groupEntitlementRepoStub struct {
	userIDs  []int64
	migrated int64
	limit    int
	active   bool
}

func (s *groupEntitlementRepoStub) Grant(context.Context, int64, int64, int64, int64, time.Time, time.Time) error {
	return nil
}
func (s *groupEntitlementRepoStub) HasActive(context.Context, int64, int64, time.Time) (bool, error) {
	return s.active, nil
}

func TestAPIKeyServiceAllowsActiveGroupCardEntitlement(t *testing.T) {
	svc := &APIKeyService{groupEntitlementRepo: &groupEntitlementRepoStub{active: true}}
	allowed := svc.canUserBindGroup(context.Background(), &User{ID: 7}, &Group{ID: 9, IsExclusive: true})
	require.True(t, allowed)
}
func (s *groupEntitlementRepoStub) ListActiveGroupExpiries(context.Context, int64, time.Time) (map[int64]time.Time, error) {
	return nil, nil
}
func (s *groupEntitlementRepoStub) ExpireDue(_ context.Context, _ time.Time, limit int) ([]int64, int64, error) {
	s.limit = limit
	return s.userIDs, s.migrated, nil
}

type groupEntitlementInvalidatorStub struct{ users []int64 }

func (*groupEntitlementInvalidatorStub) InvalidateAuthCacheByKey(context.Context, string) {}
func (s *groupEntitlementInvalidatorStub) InvalidateAuthCacheByUserID(_ context.Context, userID int64) {
	s.users = append(s.users, userID)
}
func (*groupEntitlementInvalidatorStub) InvalidateAuthCacheByGroupID(context.Context, int64) {}

func TestGroupEntitlementExpiryServiceProcessOnceInvalidatesMigratedUsers(t *testing.T) {
	repo := &groupEntitlementRepoStub{userIDs: []int64{11, 22}, migrated: 3}
	invalidator := &groupEntitlementInvalidatorStub{}
	svc := NewGroupEntitlementExpiryService(repo, invalidator)
	svc.batch = 17

	svc.processOnce()

	require.Equal(t, 17, repo.limit)
	require.Equal(t, []int64{11, 22}, invalidator.users)
}
