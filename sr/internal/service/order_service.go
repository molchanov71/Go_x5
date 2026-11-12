package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"example.com/order-service/internal/order"
)

type OrderService struct {
	repository order.Repository
}

func NewOrderService(repository order.Repository) *OrderService {
	return &OrderService{repository: repository}
}

func (s *OrderService) Create(ctx context.Context, input order.CreateOrderInput) (order.Order, error) {
	if len(input.Items) == 0 {
		return order.Order{}, fmt.Errorf("%w: items are required", order.ErrInvalidOrder)
	}
	var total int64
	for _, item := range input.Items {
		if strings.TrimSpace(item.Name) == "" || item.Quantity <= 0 || item.Price <= 0 {
			return order.Order{}, fmt.Errorf("%w: invalid item", order.ErrInvalidOrder)
		}
		total += item.Price * int64(item.Quantity)
	}
	now := time.Now().UTC()
	newOrder := order.Order{
		Items:      input.Items,
		TotalPrice: total,
		Status:     order.StatusNew,
		CreatedAt:  now,
		History: []order.StatusChange{
			{Status: order.StatusNew, ChangedAt: now},
		},
	}
	return s.repository.Create(ctx, newOrder)
}

func (s *OrderService) GetByID(ctx context.Context, id int64) (order.Order, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *OrderService) List(ctx context.Context) ([]order.Order, error) {
	return s.repository.List(ctx)
}

func (s *OrderService) UpdateStatus(ctx context.Context, id int64, status order.Status) (order.Order, error) {
	switch status {
	case order.StatusNew, order.StatusProcessing, order.StatusCompleted, order.StatusCancelled:
	default:
		return order.Order{}, fmt.Errorf("%w: %s", order.ErrInvalidStatus, status)
	}
	currentOrder, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return order.Order{}, err
	}
	allowed := map[order.Status]map[order.Status]bool{
		order.StatusNew: {
			order.StatusProcessing: true,
			order.StatusCancelled:  true,
		},
		order.StatusProcessing: {
			order.StatusCompleted: true,
			order.StatusCancelled: true,
		},
	}
	if !allowed[currentOrder.Status][status] {
		return order.Order{}, fmt.Errorf("%w: %s -> %s", order.ErrInvalidStatusTransition, currentOrder.Status, status)
	}
	now := time.Now().UTC()
	currentOrder.Status = status
	currentOrder.History = append(currentOrder.History, order.StatusChange{
		Status:    status,
		ChangedAt: now,
	})
	return s.repository.Update(ctx, currentOrder)
}

func (s *OrderService) History(ctx context.Context, id int64) ([]order.StatusChange, error) {
	foundOrder, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	history := make([]order.StatusChange, len(foundOrder.History))
	copy(history, foundOrder.History)
	return history, nil
}
