package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"example.com/simple-queue/queue"
)

func EnqueueHandler(q *queue.Queue) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var msg queue.Message
		if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}

		if msg.GroupID == "" {
			http.Error(w, "missing groupId", http.StatusBadRequest)
			return
		}

		id := q.Enqueue(msg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{"id": id})
	}
}

func DequeueHandler(q *queue.Queue) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		msg, ok := q.Dequeue()
		if !ok {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(msg)
	}
}

func AckHandler(q *queue.Queue) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}

		if q.Ack(id) {
			w.WriteHeader(http.StatusOK)
		} else {
			http.Error(w, "message not found in flight", http.StatusNotFound)
		}
	}
}

func DLQHandler(q *queue.Queue) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := 100
		if l := r.URL.Query().Get("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
				limit = parsed
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(q.DLQList(limit))
	}
}

func DLQSizeHandler(q *queue.Queue) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{"size": q.DLQSize()})
	}
}

func SizeHandler(q *queue.Queue) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"size":     q.Size(),
			"inFlight": q.InFlightSize(),
			"shards":   q.ShardSizes(),
		})
	}
}
