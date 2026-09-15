package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Masih-Ghasri/RedWeb/services/payment-service/internal/core/domain"
	"github.com/Masih-Ghasri/RedWeb/services/payment-service/internal/core/ports"
	"github.com/google/uuid"
)

type PaymentService struct {
	repo        ports.PaymentRepository
	idempotency ports.IdempotencyRepository
	publisher   ports.EventPublisher
	gateway     ports.PaymentGateway
}

func NewPaymentService(repo ports.PaymentRepository, idem ports.IdempotencyRepository, pub ports.EventPublisher, gw ports.PaymentGateway) *PaymentService {
	return &PaymentService{
		repo:        repo,
		idempotency: idem,
		publisher:   pub,
		gateway:     gw,
	}
}

func (s *PaymentService) ProcessPayment(ctx context.Context, cmd ports.ProcessPaymentCommand) error {
	lockKey := fmt.Sprintf("payment_lock:%s", cmd.OrderID)
	acquired, err := s.idempotency.LockIfNotExists(ctx, lockKey)
	if err != nil {
		return fmt.Errorf("idempotency check failed: %w", err)
	}
	if !acquired {
		log.Printf("Payment for order %s is already processed", cmd.OrderID)
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

	bankCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	isSuccess, err := s.gateway.Charge(bankCtx, payment.OrderID, payment.Amount)
	if err != nil {
		return fmt.Errorf("bank gateway error: %w", err)
	}

	if isSuccess {
		payment.MarkAsSuccess()
	} else {
		payment.MarkAsFailed()
	}

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
		log.Printf("Warning: Failed to publish event: %v", err)
	}

	log.Printf("Processed payment for order %s. Bank Status: %s", cmd.OrderID, payment.Status)
	return nil
}
