package orchestrator

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *Server) handleListTemplates(w http.ResponseWriter, r *http.Request) {
	templates, err := s.store.ListTemplates(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list templates")
		log.Printf("list templates: %v", err)
		return
	}
	writeJSON(w, http.StatusOK, templates)
}

func (s *Server) handleGetTemplate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	tmpl, err := s.store.GetTemplate(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "template not found")
		return
	}
	writeJSON(w, http.StatusOK, tmpl)
}

func (s *Server) handleDeploy(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	var req DeployRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.TemplateID == "" {
		writeError(w, http.StatusBadRequest, "template_id is required")
		return
	}

	// Get template
	tmpl, err := s.store.GetTemplate(r.Context(), req.TemplateID)
	if err != nil {
		writeError(w, http.StatusNotFound, "template not found")
		return
	}

	// Check required integrations
	missing, err := s.store.CheckUserIntegrations(r.Context(), userID, tmpl.RequiredIntegrations)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check integrations")
		log.Printf("check integrations: %v", err)
		return
	}
	if len(missing) > 0 {
		writeError(w, http.StatusPreconditionFailed,
			fmt.Sprintf("missing required integrations: %v", missing))
		return
	}

	// Create instance record
	configOverrides := req.ConfigOverrides
	if configOverrides == nil {
		configOverrides = json.RawMessage("{}")
	}
	inst, err := s.store.CreateInstance(r.Context(), AgentInstance{
		UserID:          userID,
		TemplateID:      req.TemplateID,
		K8sNamespace:    s.cfg.KubeNamespace,
		ConfigOverrides: configOverrides,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create instance")
		log.Printf("create instance: %v", err)
		return
	}

	// Resolve credentials
	creds, err := s.creds.ResolveForUser(r.Context(), userID)
	if err != nil {
		s.store.UpdateInstanceStatus(r.Context(), inst.ID, StatusFailed, "", "credential resolution failed: "+err.Error())
		writeError(w, http.StatusInternalServerError, "failed to resolve credentials")
		log.Printf("resolve credentials: %v", err)
		return
	}

	// Build pod spec
	pod := BuildPodSpec(PodSpecParams{
		InstanceID:  inst.ID,
		UserID:      userID,
		Template:    tmpl,
		Namespace:   s.cfg.KubeNamespace,
		Credentials: creds,
		Config:      s.cfg,
	})

	// Create pod
	created, err := s.k8s.CreatePod(r.Context(), pod)
	if err != nil {
		s.store.UpdateInstanceStatus(r.Context(), inst.ID, StatusFailed, "", "pod creation failed: "+err.Error())
		writeError(w, http.StatusInternalServerError, "failed to create pod")
		log.Printf("create pod: %v", err)
		return
	}

	// Update instance with pod name and status
	s.store.UpdateInstanceStatus(r.Context(), inst.ID, StatusCreating, created.Name, "")

	// Re-fetch to get updated fields
	inst, _ = s.store.GetInstance(r.Context(), inst.ID)
	writeJSON(w, http.StatusCreated, inst)
}

func (s *Server) handleListInstances(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	instances, err := s.store.ListInstances(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list instances")
		log.Printf("list instances: %v", err)
		return
	}
	if instances == nil {
		instances = []AgentInstance{}
	}
	writeJSON(w, http.StatusOK, instances)
}

func (s *Server) handleGetInstance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	inst, err := s.store.GetInstance(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "instance not found")
		return
	}

	// Verify ownership
	if inst.UserID != getUserID(r) {
		writeError(w, http.StatusNotFound, "instance not found")
		return
	}

	writeJSON(w, http.StatusOK, inst)
}

func (s *Server) handleGetLogs(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	inst, err := s.store.GetInstance(r.Context(), id)
	if err != nil || inst.UserID != getUserID(r) {
		writeError(w, http.StatusNotFound, "instance not found")
		return
	}

	if inst.K8sPodName == "" {
		writeError(w, http.StatusPreconditionFailed, "pod not yet created")
		return
	}

	stream, err := s.k8s.GetPodLogs(r.Context(), inst.K8sPodName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get logs")
		log.Printf("get logs: %v", err)
		return
	}
	defer stream.Close()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	buf := make([]byte, 4096)
	for {
		n, err := stream.Read(buf)
		if n > 0 {
			fmt.Fprintf(w, "data: %s\n\n", buf[:n])
			flusher.Flush()
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
	}
}

func (s *Server) handleStopAgent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	inst, err := s.store.GetInstance(r.Context(), id)
	if err != nil || inst.UserID != getUserID(r) {
		writeError(w, http.StatusNotFound, "instance not found")
		return
	}

	if inst.Status != StatusRunning && inst.Status != StatusCreating {
		writeError(w, http.StatusConflict, "agent is not running")
		return
	}

	s.store.UpdateInstanceStatus(r.Context(), id, StatusStopping, "", "")

	if inst.K8sPodName != "" {
		if err := s.k8s.DeletePod(r.Context(), inst.K8sPodName); err != nil {
			log.Printf("delete pod %s: %v", inst.K8sPodName, err)
		}
	}

	s.store.UpdateInstanceStatus(r.Context(), id, StatusStopped, "", "")

	inst, _ = s.store.GetInstance(r.Context(), id)
	writeJSON(w, http.StatusOK, inst)
}

func (s *Server) handleDeleteInstance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	inst, err := s.store.GetInstance(r.Context(), id)
	if err != nil || inst.UserID != getUserID(r) {
		writeError(w, http.StatusNotFound, "instance not found")
		return
	}

	if err := s.store.DeleteInstance(r.Context(), id); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
