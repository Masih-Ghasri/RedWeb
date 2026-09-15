package domain

import "time"

type PaymentStatus string

const (
	PaymentStatusPending PaymentStatus = "PENDING"
	PaymentStatusSuccess PaymentStatus = "SUCCESS"
	PaymentStatusFailed  PaymentStatus = "FAILED"
)

type Payment struct {
	ID        string
	OrderID   string
	Amount    float64
	Status    PaymentStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewPayment Business Rules for Payment
func NewPayment(id, orderID string, amount float64) (*Payment, error) {
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}

	return &Payment{
		ID:        id,
		OrderID:   orderID,
		Amount:    amount,
		Status:    PaymentStatusPending,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}, nil
}

func (p *Payment) MarkAsSuccess() {
	p.Status = PaymentStatusSuccess
	p.UpdatedAt = time.Now().UTC()
}

func (p *Payment) MarkAsFailed() {
	p.Status = PaymentStatusFailed
	p.UpdatedAt = time.Now().UTC()
}
