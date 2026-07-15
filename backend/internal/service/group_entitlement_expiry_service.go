package service

import (
	"context"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
)

type GroupEntitlementExpiryService struct {
	repo        GroupEntitlementRepository
	invalidator APIKeyAuthCacheInvalidator
	interval    time.Duration
	batch       int
	startOnce   sync.Once
	stopOnce    sync.Once
	stopCh      chan struct{}
}

func NewGroupEntitlementExpiryService(repo GroupEntitlementRepository, invalidator APIKeyAuthCacheInvalidator) *GroupEntitlementExpiryService {
	return &GroupEntitlementExpiryService{
		repo: repo, invalidator: invalidator,
		interval: time.Minute, batch: 200, stopCh: make(chan struct{}),
	}
}

func (s *GroupEntitlementExpiryService) Start() {
	if s == nil || s.repo == nil {
		return
	}
	s.startOnce.Do(func() {
		logger.LegacyPrintf("service.group_entitlement_expiry", "[GroupEntitlementExpiry] started interval=%s batch=%d", s.interval, s.batch)
		go s.runLoop()
	})
}

func (s *GroupEntitlementExpiryService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() { close(s.stopCh) })
}

func (s *GroupEntitlementExpiryService) runLoop() {
	ticker := time.NewTicker(s.interval)
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

func (s *GroupEntitlementExpiryService) processOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	userIDs, migrated, err := s.repo.ExpireDue(ctx, time.Now().UTC(), s.batch)
	if err != nil {
		logger.LegacyPrintf("service.group_entitlement_expiry", "[GroupEntitlementExpiry] processing failed err=%v", err)
		return
	}
	for _, userID := range userIDs {
		if s.invalidator != nil {
			s.invalidator.InvalidateAuthCacheByUserID(ctx, userID)
		}
	}
	if migrated > 0 {
		logger.LegacyPrintf("service.group_entitlement_expiry", "[GroupEntitlementExpiry] migrated api keys count=%d users=%d", migrated, len(userIDs))
	}
}

func ProvideGroupEntitlementExpiryService(repo GroupEntitlementRepository, invalidator APIKeyAuthCacheInvalidator) *GroupEntitlementExpiryService {
	svc := NewGroupEntitlementExpiryService(repo, invalidator)
	svc.Start()
	return svc
}
