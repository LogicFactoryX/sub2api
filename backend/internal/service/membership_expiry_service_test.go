package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type membershipRepoStub struct {
	userIDs   []int64
	migrated  int64
	expiresAt *time.Time
}

func (s *membershipRepoStub) GrantOrExtend(context.Context, int64, int64, int, time.Time) (time.Time, error) {
	return time.Time{}, nil
}
func (s *membershipRepoStub) GetActiveExpiry(context.Context, int64, time.Time) (*time.Time, error) {
	return s.expiresAt, nil
}
func (s *membershipRepoStub) ExpireDue(context.Context, time.Time, int) ([]int64, int64, error) {
	return s.userIDs, s.migrated, nil
}

func TestAPIKeyServiceRequiresMembershipForMemberGroup(t *testing.T) {
	expiresAt := time.Now().Add(time.Hour)
	svc := &APIKeyService{membershipRepo: &membershipRepoStub{expiresAt: &expiresAt}}
	require.True(t, svc.canUserBindGroup(context.Background(), &User{ID: 7}, &Group{ID: 9, IsMemberGroup: true}))

	svc.membershipRepo = &membershipRepoStub{}
	require.False(t, svc.canUserBindGroup(context.Background(), &User{ID: 7}, &Group{ID: 9, IsMemberGroup: true}))
}

type membershipInvalidatorStub struct{ users []int64 }

func (*membershipInvalidatorStub) InvalidateAuthCacheByKey(context.Context, string) {}
func (s *membershipInvalidatorStub) InvalidateAuthCacheByUserID(_ context.Context, userID int64) {
	s.users = append(s.users, userID)
}
func (*membershipInvalidatorStub) InvalidateAuthCacheByGroupID(context.Context, int64) {}

func TestMembershipExpiryServiceProcessOnceInvalidatesExpiredUsers(t *testing.T) {
	repo := &membershipRepoStub{userIDs: []int64{11, 22}, migrated: 3}
	invalidator := &membershipInvalidatorStub{}
	svc := NewMembershipExpiryService(repo, invalidator)
	svc.processOnce()
	require.Equal(t, []int64{11, 22}, invalidator.users)
}
