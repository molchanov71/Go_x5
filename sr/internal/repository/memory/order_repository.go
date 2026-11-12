package memory

import (
	"context"

	"example.com/order-service/internal/order"
)

type OrderRepository struct {
	items  []order.Order
	nextID int64
}

func NewOrderRepository() *OrderRepository {
	return &OrderRepository{
		items:  make([]order.Order, 0),
		nextID: 1,
	}
}

func (r *OrderRepository) Create(_ context.Context, value order.Order) (order.Order, error) {
	value.ID = r.nextID
	r.nextID++
	r.items = append(r.items, value)
	return value, nil
}

func (r *OrderRepository) GetByID(_ context.Context, id int64) (order.Order, error) {
	for _, item := range r.items {
		if item.ID == id {
			return item, nil
		}
	}
	return order.Order{}, order.ErrOrderNotFound
}

func (r *OrderRepository) List(_ context.Context) ([]order.Order, error) {
	orders := make([]order.Order, len(r.items))
	copy(orders, r.items)
	return orders, nil
}

func (r *OrderRepository) Update(_ context.Context, value order.Order) (order.Order, error) {
	for i, item := range r.items {
		if item.ID == value.ID {
			r.items[i] = value
			return value, nil
		}
	}
	return order.Order{}, order.ErrOrderNotFound
}
