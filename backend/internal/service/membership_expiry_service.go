package service

import (
	"context"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
)

type MembershipExpiryService struct {
	repo        MembershipRepository
	invalidator APIKeyAuthCacheInvalidator
	startOnce   sync.Once
	stopOnce    sync.Once
	stopCh      chan struct{}
}

func NewMembershipExpiryService(repo MembershipRepository, invalidator APIKeyAuthCacheInvalidator) *MembershipExpiryService {
	return &MembershipExpiryService{repo: repo, invalidator: invalidator, stopCh: make(chan struct{})}
}

func (s *MembershipExpiryService) Start() {
	if s == nil || s.repo == nil {
		return
	}
	s.startOnce.Do(func() { go s.runLoop() })
}

func (s *MembershipExpiryService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() { close(s.stopCh) })
}

func (s *MembershipExpiryService) runLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	s.processOnce()
	for {
		select {
		case <-ticker.C:
			s.processOnce()
		case <-s.stopCh:
			return
		}
	}
}

func (s *MembershipExpiryService) processOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	userIDs, migrated, err := s.repo.ExpireDue(ctx, time.Now().UTC(), 200)
	if err != nil {
		logger.LegacyPrintf("service.membership_expiry", "[MembershipExpiry] processing failed err=%v", err)
		return
	}
	for _, userID := range userIDs {
		if s.invalidator != nil {
			s.invalidator.InvalidateAuthCacheByUserID(ctx, userID)
		}
	}
	if migrated > 0 {
		logger.LegacyPrintf("service.membership_expiry", "[MembershipExpiry] migrated api keys count=%d users=%d", migrated, len(userIDs))
	}
}

func ProvideMembershipExpiryService(repo MembershipRepository, invalidator APIKeyAuthCacheInvalidator) *MembershipExpiryService {
	svc := NewMembershipExpiryService(repo, invalidator)
	svc.Start()
	return svc
}
