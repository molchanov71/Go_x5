package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"example.com/order-service/internal/order"
	"example.com/order-service/internal/utils"
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
)

type OrderService interface {
	Create(ctx context.Context, input order.CreateOrderInput) (order.Order, error)
	GetByID(ctx context.Context, id int64) (order.Order, error)
	List(ctx context.Context) ([]order.Order, error)
	UpdateStatus(ctx context.Context, id int64, status order.Status) (order.Order, error)
	History(ctx context.Context, id int64) ([]order.StatusChange, error)
}

type Handler struct {
	service OrderService
	logger  *slog.Logger
}

func NewRouter(service OrderService, logger *slog.Logger) http.Handler {
	h := &Handler{service: service, logger: logger}
	router := chi.NewRouter()
	router.Use(requestLoggingMiddleware(logger))
	router.Use(recoveryMiddleware(logger))

	router.Get("/health", h.health)
	router.Route("/api/v1/orders", func(r chi.Router) {
		r.Post("/", h.createOrder)
		r.Get("/", h.listOrders)
		r.Get("/{id}", h.getOrderByID)
		r.Patch("/{id}/status", h.updateOrderStatus)
		r.Get("/{id}/history", h.getOrderHistory)
	})
	router.Get("/swagger/*", httpSwagger.WrapHandler)

	return router
}

// status godoc
// @Summary Вернуть статус OK
// @Tags orders
// @Success 200 {object} utils.StatusResponse
// @Router /health [get]
func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, utils.StatusResponse{Status: "ok"})
}

// createOrder godoc
// @Summary Создать заказ
// @Tags orders
// @Param in body order.CreateOrderInput true "Имя покупателя"
// @Success 201 {object} order.Order
// @Header 201 {string} Location "/api/v1/orders/{id}"
// @Failure 400 {object} utils.ErrorResponse
// @Failure 413 {object} utils.ErrorResponse
// @Failure 415 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Produce json
// @Router /api/v1/orders [post]
func (h *Handler) createOrder(w http.ResponseWriter, r *http.Request) {
	var in *order.CreateOrderInput
	if err := utils.DecodeJSON(w, r, &in); err != nil {
		utils.WriteError(w, err.Status, err.Message, err.Code)
		return
	}
	if in == nil {
		utils.WriteError(w, 400, "body must be a JSON object", "INVALID_JSON")
		return
	}
	orders, err := h.service.Create(r.Context(), *in)
	if err != nil {
		if errors.Is(err, order.ErrInvalidOrder) {
			utils.WriteError(w, http.StatusBadRequest, "invalid order", "VALIDATION_ERROR")
			return
		}
		h.logger.Error("failed to create order", slog.String("error", err.Error()))
		utils.WriteError(w, http.StatusInternalServerError, "internal server error", "INTERNAL_ERROR")
		return
	}
	w.Header().Set("Location", "/api/v1/orders/"+strconv.FormatInt(orders.ID, 10))
	utils.WriteJSON(w, http.StatusCreated, orders)
}

// listOrders godoc
// @Summary Вернуть все заказы
// @Tags orders
// @Success 200 {object} []order.Order
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Produce json
// @Router /api/v1/orders [get]
func (h *Handler) listOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.service.List(r.Context())
	if err != nil {
		h.logger.Error("failed to list orders", slog.String("error", err.Error()))
		utils.WriteError(w, http.StatusInternalServerError, "internal server error", "INTERNAL_ERROR")
		return
	}
	utils.WriteJSON(w, http.StatusOK, orders)
}

// getOrderByID godoc
// @Summary Вернуть заказ
// @Tags orders
// @Param id path int true "ID заказа" minimum(1)
// @Success 200 {object} order.Order
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Produce json
// @Router /api/v1/orders/{id} [get]
func (h *Handler) getOrderByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid id", "INVALID_ID")
		return
	}
	foundOrder, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, order.ErrOrderNotFound) {
			utils.WriteError(w, http.StatusNotFound, "order not found", "NOT_FOUND")
			return
		}
		h.logger.Error("failed to get order", slog.String("error", err.Error()))
		utils.WriteError(w, http.StatusInternalServerError, "internal server error", "INTERNAL_ERROR")
		return
	}
	utils.WriteJSON(w, http.StatusOK, foundOrder)
}

// updateOrderStatus godoc
// @Summary Обновить статус заказа
// @Tags orders
// @Param id path int true "ID заказа"
// @Param in body order.UpdateStatusInput true "Новый статус заказа"
// @Success 200 {object} order.Order
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 409 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Produce json
// @Router /api/v1/orders/{id}/status [patch]
func (h *Handler) updateOrderStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid id", "INVALID_ID")
		return
	}
	var in *order.UpdateStatusInput
	if inputErr := utils.DecodeJSON(w, r, &in); inputErr != nil {
		utils.WriteError(w, inputErr.Status, inputErr.Message, inputErr.Code)
		return
	}
	if in == nil {
		utils.WriteError(w, 400, "body must be a JSON object", "INVALID_JSON")
		return
	}
	updatedOrder, err := h.service.UpdateStatus(r.Context(), id, in.Status)
	if err != nil {
		switch {
		case errors.Is(err, order.ErrOrderNotFound):
			utils.WriteError(w, http.StatusNotFound, "order not found", "NOT_FOUND")
		case errors.Is(err, order.ErrInvalidStatus):
			utils.WriteError(w, http.StatusBadRequest, "invalid status", "VALIDATION_ERROR")
		case errors.Is(err, order.ErrInvalidStatusTransition):
			utils.WriteError(w, http.StatusConflict, "status transition is not allowed", "INVALID_STATUS_TRANSITION")
		default:
			h.logger.Error("failed to update order status", slog.String("error", err.Error()))
			utils.WriteError(w, http.StatusInternalServerError, "internal server error", "INTERNAL_ERROR")
		}
		return
	}
	utils.WriteJSON(w, http.StatusOK, updatedOrder)
}

// getOrderHistory godoc
// @Summary Вернуть историю изменений статуса заказа
// @Tags orders
// @Param id path int true "ID заказа"
// @Success 200 {object} []order.StatusChange
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Produce json
// @Router /api/v1/orders/{id}/history [get]
func (h *Handler) getOrderHistory(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid id", "INVALID_ID")
		return
	}
	history, err := h.service.History(r.Context(), id)
	if err != nil {
		if errors.Is(err, order.ErrOrderNotFound) {
			utils.WriteError(w, http.StatusNotFound, "order not found", "NOT_FOUND")
			return
		}
		h.logger.Error("failed to get order", slog.String("error", err.Error()))
		utils.WriteError(w, http.StatusInternalServerError, "internal server error", "INTERNAL_ERROR")
		return
	}
	utils.WriteJSON(w, http.StatusOK, history)
}

type errorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
