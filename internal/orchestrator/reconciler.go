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

		pod, err := s.k8s.GetPod(ctx, inst.K8sPodName)
		if err != nil {
			// Pod might have been deleted
			s.store.UpdateInstanceStatus(ctx, inst.ID, StatusFailed, "", "pod not found")
			continue
		}

		newStatus := PodStatus(pod)
		if newStatus != inst.Status {
			errMsg := ""
			if newStatus == StatusFailed {
				errMsg = pod.Status.Message
				if errMsg == "" && len(pod.Status.ContainerStatuses) > 0 {
					cs := pod.Status.ContainerStatuses[0]
					if cs.State.Terminated != nil {
						errMsg = cs.State.Terminated.Reason
					}
				}
			}
			s.store.UpdateInstanceStatus(ctx, inst.ID, newStatus, "", errMsg)
			log.Printf("reconcile: instance %s status %s → %s", inst.ID, inst.Status, newStatus)
		}
	}
}
