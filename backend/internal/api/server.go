package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"personal-doctor/backend/internal/agent"
	"personal-doctor/backend/internal/domain"
	"personal-doctor/backend/internal/store"
)

type Server struct {
	store *store.Store
	agent *agent.DoctorAgent
}

func NewServer(store *store.Store, agent *agent.DoctorAgent) *Server {
	return &Server{store: store, agent: agent}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", s.health)
	mux.HandleFunc("GET /api/patients", s.listPatients)
	mux.HandleFunc("POST /api/patients", s.createPatient)
	mux.HandleFunc("GET /api/patients/{id}/records", s.listRecords)
	mux.HandleFunc("POST /api/patients/{id}/records", s.createRecord)
	mux.HandleFunc("GET /api/patients/{id}/messages", s.listMessages)
	mux.HandleFunc("POST /api/patients/{id}/chat", s.chat)
	return withCORS(mux)
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) listPatients(w http.ResponseWriter, r *http.Request) {
	patients, err := s.store.ListPatients(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, patients)
}

func (s *Server) createPatient(w http.ResponseWriter, r *http.Request) {
	var patient domain.Patient
	if err := decodeJSON(r, &patient); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.store.CreatePatient(r.Context(), patient)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) listRecords(w http.ResponseWriter, r *http.Request) {
	records, err := s.store.ListRecords(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, records)
}

func (s *Server) createRecord(w http.ResponseWriter, r *http.Request) {
	var record domain.MedicalRecord
	if err := decodeJSON(r, &record); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	record.PatientID = r.PathValue("id")
	created, err := s.store.CreateRecord(r.Context(), record)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) listMessages(w http.ResponseWriter, r *http.Request) {
	messages, err := s.store.ListMessages(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, messages)
}

func (s *Server) chat(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Message string `json:"message"`
	}
	if err := decodeJSON(r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	payload.Message = strings.TrimSpace(payload.Message)
	if payload.Message == "" {
		writeError(w, http.StatusBadRequest, errors.New("message is required"))
		return
	}

	patientID := r.PathValue("id")
	patient, err := s.store.GetPatient(r.Context(), patientID)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	records, err := s.store.ListRecords(r.Context(), patientID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	history, err := s.store.ListMessages(r.Context(), patientID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	userMessage, err := s.store.CreateMessage(r.Context(), domain.ChatMessage{
		PatientID: patientID,
		Role:      "user",
		Content:   payload.Message,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	reply, err := s.agent.Chat(r.Context(), domain.ChatInput{
		Patient: patient,
		Records: records,
		History: history,
		Message: payload.Message,
	})
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}

	assistantMessage, err := s.store.CreateMessage(r.Context(), domain.ChatMessage{
		PatientID: patientID,
		Role:      "assistant",
		Content:   reply,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]domain.ChatMessage{
		"user":      userMessage,
		"assistant": assistantMessage,
	})
}

func decodeJSON(r *http.Request, target any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Max-Age", "86400")
		w.Header().Set("Date", time.Now().UTC().Format(http.TimeFormat))
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
