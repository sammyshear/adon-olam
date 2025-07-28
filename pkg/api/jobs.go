package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type JobStatus struct {
	State   string `json:"state"`
	Message string `json:"message,omitempty"`
	JobURL  string `json:"jobUrl,omitempty"`
}

var RequestStatus map[string]JobStatus

func storeStatus(requestID string, status JobStatus) {
	if RequestStatus == nil {
		RequestStatus = make(map[string]JobStatus)
	}
	RequestStatus[requestID] = status
}

func getStatus(requestID string) (JobStatus, error) {
	if v, ok := RequestStatus[requestID]; ok {
		return v, nil
	}

	return JobStatus{}, fmt.Errorf("request id not found")
}

func generateRequestID() string {
	return strings.ReplaceAll(uuid.NewString(), "-", "")
}

func JobStatusHandler(w http.ResponseWriter, r *http.Request) {
	requestID := r.PathValue("requestID")

	status, err := getStatus(requestID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(status.Message))
}

func JobStatusTicker(w http.ResponseWriter, r *http.Request) {
	requestID := r.PathValue("requestID")

	status, err := getStatus(requestID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if status.State == "COMPLETED" || status.State == "ERRORED" {
		w.Header().Add("HX-Trigger", "done")
		return
	}

	w.WriteHeader(http.StatusOK)
}
