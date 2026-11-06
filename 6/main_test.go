package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

// До реализации health этот тест должен быть красным.
func TestHealth(t *testing.T) {
	r := chi.NewRouter()
	r.Get("/health", health)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200", rec.Code)
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("Content-Type=%q", rec.Header().Get("Content-Type"))
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Status != "ok" {
		t.Fatalf("status in body=%q", body.Status)
	}
}

func TestCreateOrder(t *testing.T) {
	store := NewStore()
	r := chi.NewRouter()
	r.Post("/orders", createOrder(store))
	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`{"customer":"Анна"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body = %s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}
	var order Order
	if err := json.Unmarshal(rec.Body.Bytes(), &order); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if order.ID <= 0 {
		t.Fatalf("id = %d, want positive", order.ID)
	}
	if order.Customer != "Анна" {
		t.Errorf("customer = %q, want Анна", order.Customer)
	}
	if order.Status != "new" {
		t.Errorf("status = %q, want new", order.Status)
	}
}

func TestReadCreatedOrder(t *testing.T) {
	store := NewStore()
	r := chi.NewRouter()
	r.Post("/orders", createOrder(store))
	r.Get("/orders/{id}", getOrderByID(store))
	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`{"customer":"Анна"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	var order Order
	if err := json.Unmarshal(rec.Body.Bytes(), &order); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/orders/%d", order.ID), nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", got)
	}
	var order1 Order
	if err := json.Unmarshal(rec.Body.Bytes(), &order1); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if order != order1 {
		t.Fatalf("order.ID = %d, order.Customer = %s, order.Status = %s; want order.ID = %d, order.Customer = %s, order.Status = %s", order1.ID, order1.Customer, order1.Status, order.ID, order.Customer, order.Status)
	}
}

func TestReadUnknownOrder(t *testing.T) {
	store := NewStore()
	r := chi.NewRouter()
	r.Get("/orders/{id}", getOrderByID(store))
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/orders/1234"), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	CheckError(t, rec, http.StatusNotFound, "заказ не найден")
}

func TestReadBadID(t *testing.T) {
	store := NewStore()
	r := chi.NewRouter()
	r.Get("/orders/{id}", getOrderByID(store))
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/orders/abc"), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	CheckError(t, rec, http.StatusBadRequest, "не удалось распознать параметр id")
}

func CheckError(t *testing.T, rec *httptest.ResponseRecorder, errCode int, errText string) {
	if rec.Code != errCode {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, errCode, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", got)
	}
	var body struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Error != errText {
		t.Fatalf("body.Error = %q, want: %q", body.Error, errText)
	}
}

func TestCreateNoBody(t *testing.T) {
	store := NewStore()
	r := chi.NewRouter()
	r.Post("/orders", createOrder(store))
	req := httptest.NewRequest(http.MethodPost, "/orders", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	CheckError(t, rec, http.StatusBadRequest, "неверный формат запроса")
}

func TestCreateBrokenJSON(t *testing.T) {
	store := NewStore()
	r := chi.NewRouter()
	r.Post("/orders", createOrder(store))
	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`{"customer":`))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	CheckError(t, rec, http.StatusBadRequest, "неверный формат запроса")
}

func TestCreateBadCustomer(t *testing.T) {
	store := NewStore()
	r := chi.NewRouter()
	r.Post("/orders", createOrder(store))
	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`{"customer": 1}`))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	CheckError(t, rec, http.StatusBadRequest, "неверный формат запроса")
}

func TestEmptyName(t *testing.T) {
	store := NewStore()
	r := chi.NewRouter()
	r.Post("/orders", createOrder(store))
	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`{"customer": ""}`))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	CheckError(t, rec, http.StatusBadRequest, "пустое имя покупателя")
}

func TestRecovery(t *testing.T) {
	r := chi.NewRouter()
	r.Use(accessLog)
	r.Use(recovery)

	r.Get("/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", got)
	}

	var body struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Error != "internal server error" {
		t.Errorf("error = %q, want internal server error", body.Error)
	}
}
