package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/Masih-Ghasri/RedWeb/services/payment-service/internal/core/domain"
	"github.com/Masih-Ghasri/RedWeb/services/payment-service/internal/core/ports"
	"github.com/google/uuid"
)

type PaymentService struct {
	repo        ports.PaymentRepository
	idempotency ports.IdempotencyRepository
	publisher   ports.EventPublisher
}

func NewPaymentService(repo ports.PaymentRepository, idem ports.IdempotencyRepository, pub ports.EventPublisher) *PaymentService {
	return &PaymentService{
		repo:        repo,
		idempotency: idem,
		publisher:   pub,
	}
}

func (s *PaymentService) ProcessPayment(ctx context.Context, cmd ports.ProcessPaymentCommand) error {
	lockKey := fmt.Sprintf("payment_lock:%s", cmd.OrderID)
	acquired, err := s.idempotency.LockIfNotExists(ctx, lockKey)
	if err != nil {
		return fmt.Errorf("idempotency check failed: %w", err)
	}
	if !acquired {
		log.Printf("Payment for order %s is already being processed or completed", cmd.OrderID)
		return nil
	}

	defer func() {
		if err != nil {
			_ = s.idempotency.Unlock(ctx, lockKey)
		}
	}()

	paymentID := uuid.New().String()
	payment, err := domain.NewPayment(paymentID, cmd.OrderID, cmd.Amount)
	if err != nil {
		return err
	}

	payment.MarkAsSuccess()

	if err = s.repo.Save(ctx, payment); err != nil {
		return fmt.Errorf("failed to save payment: %w", err)
	}

	event := domain.PaymentProcessedEvent{
		PaymentID: payment.ID,
		OrderID:   payment.OrderID,
		Status:    string(payment.Status),
	}
	payloadBytes, _ := json.Marshal(event)

	err = s.publisher.Publish(ctx, "order.exchange", "payment.processed", payloadBytes)
	if err != nil {
		log.Printf("Warning: Failed to publish PaymentProcessedEvent: %v", err)
	}

	log.Printf("Successfully processed payment for order %s", cmd.OrderID)
	return nil
}
