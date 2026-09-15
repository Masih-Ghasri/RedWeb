package postgres

import (
	"github.com/Masih-Ghasri/RedWeb/services/order-service/internal/core/domain"
	"time"
)

type OrderModel struct {
	ID          string `gorm:"primaryKey"`
	CustomerID  string
	TotalAmount float64
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Items       []OrderItemModel `gorm:"foreignKey:OrderID"`
}

func (OrderModel) TableName() string { return "orders" }

type OrderItemModel struct {
	ID        uint `gorm:"primaryKey;autoIncrement"`
	OrderID   string
	ProductID string
	Quantity  int
	Price     float64
}

func (OrderItemModel) TableName() string { return "order_items" }

type OutboxEventModel struct {
	ID            string `gorm:"primaryKey"`
	AggregateType string
	AggregateID   string
	EventType     string
	Payload       []byte `gorm:"type:jsonb"`
	Status        string
	CreatedAt     time.Time
}

func (OutboxEventModel) TableName() string { return "outbox_events" }

// Mappers
func toOrderModel(o *domain.Order) *OrderModel {
	model := &OrderModel{
		ID:          o.ID,
		CustomerID:  o.CustomerID,
		TotalAmount: o.TotalAmount,
		Status:      string(o.Status),
		CreatedAt:   o.CreatedAt,
		UpdatedAt:   o.UpdatedAt,
	}
	for _, item := range o.Items {
		model.Items = append(model.Items, OrderItemModel{
			OrderID:   o.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		})
	}
	return model
}

func toOutboxModel(e *domain.OutboxEvent) *OutboxEventModel {
	return &OutboxEventModel{
		ID:            e.ID,
		AggregateType: e.AggregateType,
		AggregateID:   e.AggregateID,
		EventType:     e.EventType,
		Payload:       e.Payload,
		Status:        string(e.Status),
		CreatedAt:     e.CreatedAt,
	}
}
