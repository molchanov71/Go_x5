package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"example.com/order-service/internal/order"
	"example.com/order-service/internal/service"
	"github.com/stretchr/testify/require"
)

var errUnexpectedRepositoryCall = errors.New("unexpected repository call")

type repositoryMock struct {
	createFn  func(context.Context, order.Order) (order.Order, error)
	getByIDFn func(context.Context, int64) (order.Order, error)
	listFn    func(context.Context) ([]order.Order, error)
	updateFn  func(context.Context, order.Order) (order.Order, error)
}

func (m *repositoryMock) Create(ctx context.Context, value order.Order) (order.Order, error) {
	if m.createFn == nil {
		return order.Order{}, errUnexpectedRepositoryCall
	}
	return m.createFn(ctx, value)
}

func (m *repositoryMock) GetByID(ctx context.Context, id int64) (order.Order, error) {
	if m.getByIDFn == nil {
		return order.Order{}, errUnexpectedRepositoryCall
	}
	return m.getByIDFn(ctx, id)
}

func (m *repositoryMock) List(ctx context.Context) ([]order.Order, error) {
	if m.listFn == nil {
		return nil, errUnexpectedRepositoryCall
	}
	return m.listFn(ctx)
}

func (m *repositoryMock) Update(ctx context.Context, value order.Order) (order.Order, error) {
	if m.updateFn == nil {
		return order.Order{}, errUnexpectedRepositoryCall
	}
	return m.updateFn(ctx, value)
}

func TestOrderService_Create(t *testing.T) {
	input := order.CreateOrderInput{
		Items: []order.Item{
			{Name: "Keyboard", Quantity: 2, Price: 500000},
			{Name: "Mouse", Quantity: 3, Price: 10000},
		},
	}

	repository := &repositoryMock{}
	repository.createFn = func(_ context.Context, value order.Order) (order.Order, error) {
		value.ID = 1
		return value, nil
	}

	result, err := service.NewOrderService(repository).Create(context.Background(), input)

	require.NoError(t, err)
	require.Equal(t, int64(1), result.ID)
	require.Equal(t, input.Items, result.Items)
	require.Equal(t, int64(1030000), result.TotalPrice)
	require.Equal(t, order.StatusNew, result.Status)
	require.False(t, result.CreatedAt.IsZero())
	require.Len(t, result.History, 1)
	require.Equal(t, order.StatusNew, result.History[0].Status)
	require.Equal(t, result.CreatedAt, result.History[0].ChangedAt)
}

func TestOrderService_Create_Validation(t *testing.T) {
	tests := []struct {
		name  string
		input order.CreateOrderInput
	}{
		{name: "nil items", input: order.CreateOrderInput{}},
		{name: "empty items", input: order.CreateOrderInput{Items: []order.Item{}}},
		{
			name: "empty name",
			input: order.CreateOrderInput{Items: []order.Item{
				{Name: "", Quantity: 1, Price: 100},
			}},
		},
		{
			name: "whitespace name",
			input: order.CreateOrderInput{Items: []order.Item{
				{Name: "   ", Quantity: 1, Price: 100},
			}},
		},
		{
			name: "zero quantity",
			input: order.CreateOrderInput{Items: []order.Item{
				{Name: "Keyboard", Quantity: 0, Price: 100},
			}},
		},
		{
			name: "negative quantity",
			input: order.CreateOrderInput{Items: []order.Item{
				{Name: "Keyboard", Quantity: -1, Price: 100},
			}},
		},
		{
			name: "zero price",
			input: order.CreateOrderInput{Items: []order.Item{
				{Name: "Keyboard", Quantity: 1, Price: 0},
			}},
		},
		{
			name: "negative price",
			input: order.CreateOrderInput{Items: []order.Item{
				{Name: "Keyboard", Quantity: 1, Price: -100},
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.NewOrderService(&repositoryMock{}).Create(context.Background(), tt.input)

			require.ErrorIs(t, err, order.ErrInvalidOrder)
			require.Empty(t, result)
		})
	}
}

func TestOrderService_Create_RepositoryError(t *testing.T) {
	wantErr := errors.New("create failed")
	repository := &repositoryMock{
		createFn: func(context.Context, order.Order) (order.Order, error) {
			return order.Order{}, wantErr
		},
	}
	input := order.CreateOrderInput{Items: []order.Item{
		{Name: "Keyboard", Quantity: 1, Price: 100},
	}}

	_, err := service.NewOrderService(repository).Create(context.Background(), input)

	require.ErrorIs(t, err, wantErr)
}

func TestOrderService_GetByID(t *testing.T) {
	storedOrder := order.Order{ID: 1, Status: order.StatusNew}
	tests := []struct {
		name          string
		id            int64
		storedOrder   order.Order
		repositoryErr error
		wantErr       error
	}{
		{name: "existing order", id: 1, storedOrder: storedOrder},
		{name: "missing order", id: 999, repositoryErr: order.ErrOrderNotFound, wantErr: order.ErrOrderNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := &repositoryMock{
				getByIDFn: func(_ context.Context, id int64) (order.Order, error) {
					require.Equal(t, tt.id, id)
					return tt.storedOrder, tt.repositoryErr
				},
			}

			result, err := service.NewOrderService(repository).GetByID(context.Background(), tt.id)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.storedOrder, result)
		})
	}
}

func TestOrderService_List(t *testing.T) {
	tests := []struct {
		name          string
		storedOrders  []order.Order
		repositoryErr error
		wantErr       error
	}{
		{name: "empty list", storedOrders: []order.Order{}},
		{name: "orders", storedOrders: []order.Order{{ID: 1}, {ID: 2}}},
		{name: "repository error", repositoryErr: errors.New("list failed"), wantErr: errors.New("list failed")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := &repositoryMock{
				listFn: func(context.Context) ([]order.Order, error) {
					return tt.storedOrders, tt.repositoryErr
				},
			}

			result, err := service.NewOrderService(repository).List(context.Background())

			if tt.wantErr != nil {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.storedOrders, result)
		})
	}
}

func TestOrderService_UpdateStatus_AllowedTransitions(t *testing.T) {
	tests := []struct {
		name          string
		currentStatus order.Status
		newStatus     order.Status
	}{
		{name: "new to processing", currentStatus: order.StatusNew, newStatus: order.StatusProcessing},
		{name: "new to cancelled", currentStatus: order.StatusNew, newStatus: order.StatusCancelled},
		{name: "processing to completed", currentStatus: order.StatusProcessing, newStatus: order.StatusCompleted},
		{name: "processing to cancelled", currentStatus: order.StatusProcessing, newStatus: order.StatusCancelled},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createdAt := time.Now().UTC().Add(-time.Hour)
			currentOrder := order.Order{
				ID:     1,
				Status: tt.currentStatus,
				History: []order.StatusChange{
					{Status: order.StatusNew, ChangedAt: createdAt},
				},
			}
			repository := &repositoryMock{
				getByIDFn: func(context.Context, int64) (order.Order, error) {
					return currentOrder, nil
				},
			}
			repository.updateFn = func(_ context.Context, value order.Order) (order.Order, error) {
				return value, nil
			}

			before := time.Now().UTC()
			result, err := service.NewOrderService(repository).UpdateStatus(context.Background(), 1, tt.newStatus)
			after := time.Now().UTC()

			require.NoError(t, err)
			require.Equal(t, tt.newStatus, result.Status)
			require.Len(t, result.History, len(currentOrder.History)+1)
			lastChange := result.History[len(result.History)-1]
			require.Equal(t, tt.newStatus, lastChange.Status)
			require.False(t, lastChange.ChangedAt.Before(before))
			require.False(t, lastChange.ChangedAt.After(after))
		})
	}
}

func TestOrderService_UpdateStatus_InvalidTransitions(t *testing.T) {
	tests := []struct {
		name          string
		currentStatus order.Status
		newStatus     order.Status
	}{
		{name: "new to completed", currentStatus: order.StatusNew, newStatus: order.StatusCompleted},
		{name: "new to new", currentStatus: order.StatusNew, newStatus: order.StatusNew},
		{name: "processing to new", currentStatus: order.StatusProcessing, newStatus: order.StatusNew},
		{name: "processing to processing", currentStatus: order.StatusProcessing, newStatus: order.StatusProcessing},
		{name: "completed to processing", currentStatus: order.StatusCompleted, newStatus: order.StatusProcessing},
		{name: "completed to cancelled", currentStatus: order.StatusCompleted, newStatus: order.StatusCancelled},
		{name: "cancelled to new", currentStatus: order.StatusCancelled, newStatus: order.StatusNew},
		{name: "cancelled to processing", currentStatus: order.StatusCancelled, newStatus: order.StatusProcessing},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := &repositoryMock{
				getByIDFn: func(context.Context, int64) (order.Order, error) {
					return order.Order{ID: 1, Status: tt.currentStatus}, nil
				},
			}

			_, err := service.NewOrderService(repository).UpdateStatus(context.Background(), 1, tt.newStatus)

			require.ErrorIs(t, err, order.ErrInvalidStatusTransition)
		})
	}
}

func TestOrderService_UpdateStatus_InvalidStatus(t *testing.T) {
	for _, status := range []order.Status{"", "unknown"} {
		t.Run(string(status), func(t *testing.T) {
			_, err := service.NewOrderService(&repositoryMock{}).UpdateStatus(context.Background(), 1, status)

			require.ErrorIs(t, err, order.ErrInvalidStatus)
		})
	}
}

func TestOrderService_UpdateStatus_NotFound(t *testing.T) {
	repository := &repositoryMock{
		getByIDFn: func(context.Context, int64) (order.Order, error) {
			return order.Order{}, order.ErrOrderNotFound
		},
	}

	_, err := service.NewOrderService(repository).UpdateStatus(context.Background(), 999, order.StatusProcessing)

	require.ErrorIs(t, err, order.ErrOrderNotFound)
}

func TestOrderService_UpdateStatus_RepositoryError(t *testing.T) {
	wantErr := errors.New("update failed")
	repository := &repositoryMock{
		getByIDFn: func(context.Context, int64) (order.Order, error) {
			return order.Order{ID: 1, Status: order.StatusNew}, nil
		},
		updateFn: func(context.Context, order.Order) (order.Order, error) {
			return order.Order{}, wantErr
		},
	}

	_, err := service.NewOrderService(repository).UpdateStatus(context.Background(), 1, order.StatusProcessing)

	require.ErrorIs(t, err, wantErr)
}

func TestOrderService_History(t *testing.T) {
	createdAt := time.Date(2026, time.January, 1, 10, 0, 0, 0, time.UTC)
	storedOrder := order.Order{
		ID: 1,
		History: []order.StatusChange{
			{Status: order.StatusNew, ChangedAt: createdAt},
			{Status: order.StatusProcessing, ChangedAt: createdAt.Add(time.Minute)},
			{Status: order.StatusCompleted, ChangedAt: createdAt.Add(2 * time.Minute)},
		},
	}
	repository := &repositoryMock{
		getByIDFn: func(context.Context, int64) (order.Order, error) {
			return storedOrder, nil
		},
	}

	history, err := service.NewOrderService(repository).History(context.Background(), 1)

	require.NoError(t, err)
	require.Equal(t, storedOrder.History, history)
	for i := 1; i < len(history); i++ {
		require.False(t, history[i].ChangedAt.Before(history[i-1].ChangedAt))
	}

	history[0].Status = order.StatusCancelled
	require.Equal(t, order.StatusNew, storedOrder.History[0].Status)
}

func TestOrderService_History_NotFound(t *testing.T) {
	repository := &repositoryMock{
		getByIDFn: func(context.Context, int64) (order.Order, error) {
			return order.Order{}, order.ErrOrderNotFound
		},
	}

	history, err := service.NewOrderService(repository).History(context.Background(), 999)

	require.ErrorIs(t, err, order.ErrOrderNotFound)
	require.Nil(t, history)
}
