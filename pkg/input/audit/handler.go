package audit

import (
	"encoding/json"
	"io"
	"net/http"

	"k8s.io/apiserver/pkg/apis/audit"
)

type response struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

func (a *Audit) handler(wr http.ResponseWriter, r *http.Request) {
	wr.Header()["Content-Type"] = []string{"application/json"}
	output := json.NewEncoder(wr)
	log := a.log.WithValues(
		"method", r.Method,
		"remoteAddr", r.RemoteAddr,
		"URI", r.RequestURI,
		"headers", r.Header)
	trace := log.V(10)
	trace.Info("Received Request")

	if r.Body == nil {
		trace.Info("received nil body")
		wr.WriteHeader(http.StatusBadRequest)
		_ = output.Encode(response{
			Error: "received nil body",
		})
		return
	}
	defer r.Body.Close()
	eventList := new(audit.EventList)
	event, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error(err, "while reading audit event body")
		wr.WriteHeader(http.StatusInternalServerError)
		_ = output.Encode(response{
			Error: "error reading HTTP body: " + err.Error(),
		})
		return
	}
	err = json.Unmarshal(event, eventList)
	if err != nil {
		log.Error(err, "while decoding event list")
		wr.WriteHeader(http.StatusInternalServerError)
		_ = output.Encode(response{
			Error: "error decoding audit events: " + err.Error(),
		})
		return
	}
	if eventList.Kind != "EventList" {
		log.Error(err, "Received an HTTP request that was not an EventList",
			"kind", eventList.Kind,
		)
		wr.WriteHeader(http.StatusBadRequest)
		_ = output.Encode(response{
			Error: "Received an HTTP request that was not an EventList",
		})
		return
	}
	var eventsToProcess []*audit.Event
	for _, e := range eventList.Items {
		eventsToProcess = append(eventsToProcess, &e)
	}
	err = a.processFunc(r.Context(), eventsToProcess)
	if err != nil {
		log.Error(err, "While processing audit events")
		wr.WriteHeader(http.StatusInternalServerError)
		_ = output.Encode(response{
			Error: "While processing audit event: " + err.Error(),
		})
		return
	}
	trace.Info("finished processing")
	wr.WriteHeader(http.StatusOK)
	_ = output.Encode(response{Message: "OK"})
}
