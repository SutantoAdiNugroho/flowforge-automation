package scheduler

import (
	"context"
	"fmt"
	"log"
	"sync"

	"flowforge-automation-backend/pkg/model/dto"
	"flowforge-automation-backend/pkg/repository/workflow"
	"flowforge-automation-backend/pkg/service/run"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
)

type Service interface {
	Start()
	Stop()
	SyncWorkflows() error
}

type service struct {
	cron         *cron.Cron
	workflowRepo workflow.Repository
	runService   run.Service
	mu           sync.Mutex
	entries      map[uuid.UUID]cron.EntryID
}

func NewSchedulerService(workflowRepo workflow.Repository, runService run.Service) Service {
	return &service{
		cron:         cron.New(),
		workflowRepo: workflowRepo,
		runService:   runService,
		entries:      make(map[uuid.UUID]cron.EntryID),
	}
}

func (s *service) Start() {
	s.cron.Start()
	log.Println("scheduler service started")
	if err := s.SyncWorkflows(); err != nil {
		log.Printf("failed to initial sync workflows: %v", err)
	}
}

func (s *service) Stop() {
	s.cron.Stop()
	log.Println("scheduler service stopped")
}

func (s *service) SyncWorkflows() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	workflows, err := s.workflowRepo.ListActiveCronWorkflows(context.Background())

	if err != nil {
		return fmt.Errorf("failed to fetch cron workflows: %w", err)
	}
	for wfID, entryID := range s.entries {
		s.cron.Remove(entryID)
		delete(s.entries, wfID)
	}

	for _, wf := range workflows {
		if wf.CronExpression == "" {
			continue
		}

		wfID := wf.ID
		tenantID := wf.TenantID
		userID := wf.CreatedByID

		entryID, err := s.cron.AddFunc(wf.CronExpression, func() {
			log.Printf("triggering cron run for workflow %s (%s)", wf.Name, wf.ID)
			_, err := s.runService.TriggerRun(context.Background(), tenantID, userID, wfID, &dto.TriggerRunRequest{
				TriggeredBy: "cron",
				Inputs:      nil,
			})
			if err != nil {
				log.Printf("failed to trigger cron run for workflow %s: %v", wf.ID, err)
			}
		})

		if err != nil {
			log.Printf("failed to schedule workflow %s with expression %s: %v", wf.ID, wf.CronExpression, err)
			continue
		}

		s.entries[wf.ID] = entryID
	}

	log.Printf("scheduler synced: %d workflows active", len(s.entries))
	return nil
}
