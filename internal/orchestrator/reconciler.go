package orchestrator

import (
	"context"
	"log"
	"time"
)

// StartReconciler launches a background goroutine that periodically syncs
// pod status from K8s back to the database.
func (s *Server) StartReconciler(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.reconcile(ctx)
			}
		}
	}()
}

func (s *Server) reconcile(ctx context.Context) {
	instances, err := s.store.ListRunningInstances(ctx)
	if err != nil {
		log.Printf("reconcile: list running instances: %v", err)
		return
	}

	for _, inst := range instances {
		if inst.K8sPodName == "" {
			continue
		}

		newStatus, err := s.runtime.ContainerStatus(ctx, inst.K8sPodName)
		if err != nil {
			// Container might have been deleted externally.
			if uerr := s.store.UpdateInstanceStatus(ctx, inst.ID, StatusFailed, "", "container not found"); uerr != nil {
				log.Printf("reconcile: update instance %s status: %v", inst.ID, uerr)
			}
			continue
		}

		if newStatus != inst.Status {
			errMsg := ""
			if newStatus == StatusFailed {
				errMsg = "container exited with non-zero status"
			}
			if uerr := s.store.UpdateInstanceStatus(ctx, inst.ID, newStatus, "", errMsg); uerr != nil {
				log.Printf("reconcile: update instance %s status: %v", inst.ID, uerr)
			}
			log.Printf("reconcile: instance %s status %s → %s", inst.ID, inst.Status, newStatus)
		}
	}
}
