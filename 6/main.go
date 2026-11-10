package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func accessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
		}
		next.ServeHTTP(sw, r)
		requestID, _ := r.Context().Value(requestIDKey).(string)
		slog.Info("http request",
			"request_id", requestID,
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

type contextKey string

func randomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func newRequestID() string {
	val, _ := randomHex(16)
	return val
}

const requestIDKey contextKey = "request_id"

func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := newRequestID()

		w.Header().Set("X-Request-ID", id)

		ctx := context.WithValue(r.Context(), requestIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			requestID, _ := r.Context().Value(requestIDKey).(string)
			if value := recover(); value != nil {
				slog.Error("panic recovered",
					"request_id", requestID,
					"method", r.Method,
					"path", r.URL.Path,
				)

				writeJSON(w, http.StatusInternalServerError, map[string]string{
					"error": "internal server error",
				})
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func createOrder(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Customer string `json:"customer"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "неверный формат запроса",
			})
			return
		}
		customer := strings.TrimSpace(input.Customer)
		if customer == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "пустое имя покупателя",
			})
			return
		}
		order := store.Create(customer)
		w.Header().Set("Location", fmt.Sprintf("/orders/%d", order.ID))
		writeJSON(w, http.StatusCreated, order)
	}
}

func getOrderByID(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderIdStr := chi.URLParam(r, "id")
		if orderIdStr == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "нет параметра id",
			})
			return
		}
		orderID, err := strconv.Atoi(orderIdStr)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "не удалось распознать параметр id",
			})
			return
		}
		if orderID < 1 {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "id не может быть отрицательным",
			})
			return
		}
		order, found := store.ByID(orderID)
		if !found {
			writeJSON(w, http.StatusNotFound, map[string]string{
				"error": "заказ не найден",
			})
			return
		}
		writeJSON(w, http.StatusOK, order)
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		log.Printf("encode response: %v", err)
		status = http.StatusInternalServerError
		data = []byte(`{"error":"internal server error"}`)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(append(data, '\n')); err != nil {
		log.Printf("write response: %v", err)
	}
}

func main() {
	store := NewStore()
	r := chi.NewRouter()
	r.Use(requestID)
	r.Use(accessLog)
	r.Use(recovery)
	r.Get("/health", health)
	r.Post("/orders", createOrder(store))
	r.Get("/orders/{id}", getOrderByID(store))
	log.Fatal(http.ListenAndServe(":8080", r))
}
