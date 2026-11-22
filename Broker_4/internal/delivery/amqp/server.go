package delivery

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"gitlab.ozon.ru/university/broker-4/config"
	"gitlab.ozon.ru/university/broker-4/internal/model"
	"gitlab.ozon.ru/university/broker-4/internal/protocol"
	"gitlab.ozon.ru/university/broker-4/internal/usecase"
	"gitlab.ozon.ru/university/broker-4/pkg/rabbitmq"
	"go.uber.org/zap"
)

type Server struct {
	session  *rabbitmq.Session
	executor *usecase.Executor
	security *security
	log      *zap.Logger
}

func New(session *rabbitmq.Session, executor *usecase.Executor, cfg config.Config, log *zap.Logger) *Server {
	server := &Server{session: session, executor: executor, security: newSecurity(cfg.APIKey, cfg.RateLimit), log: log}
	return server
}

func (s *Server) Run(ctx context.Context) error {
	if err := s.session.Declare(ctx); err != nil {
		return err
	}
	deliveries, err := s.session.Consume(ctx, rabbitmq.Requests, false)
	if err != nil {
		return err
	}
	s.log.Info("server ready", zap.String("queue", rabbitmq.Requests))
	for {
		select {
		case <-ctx.Done():
			return nil
		case message, ok := <-deliveries:
			if !ok {
				err = rabbitmq.Closed(ctx)
				return err
			}
			if err = s.handle(ctx, message); err != nil {
				if ctx.Err() != nil {
					return nil
				}
				return err
			}
		}
	}
}

func (s *Server) handle(ctx context.Context, message amqp.Delivery) error {
	request, requestErr := protocol.Decode(message.Body)
	if requestErr == nil && (message.CorrelationId != request.ID || !strings.HasPrefix(message.ReplyTo, "amq.gen-")) {
		requestErr = model.Protocol("invalid_message", "correlation_id must match id and reply_to must be a private reply queue")
	}
	if requestErr == nil && message.ContentType != "application/json" {
		requestErr = model.Protocol("invalid_message", "content_type must be application/json")
	}

	var headers amqp.Table
	if requestErr == nil {
		headers, requestErr = s.security.check(request.Auth)
	}
	var data any
	if requestErr == nil {
		data, requestErr = s.executor.Execute(ctx, s.security.scope, request.Command)
	}
	if ctx.Err() != nil {
		contextErr := ctx.Err()

		return contextErr
	}

	attempt := attempts(message.Headers)
	dead := false
	if requestErr != nil {
		domain, expected := errors.AsType[*model.Error](requestErr)
		if !expected {
			s.log.Error("temporary processing failure", zap.String("id", logID(request.ID)), zap.Int("attempt", attempt), zap.Error(requestErr))
			if attempt < rabbitmq.MaxAttempts {
				err := s.retry(ctx, message, attempt)
				return err
			}
			requestErr = model.Protocol("retry_exhausted", "processing failed after three attempts")
			dead = true
		}
		if expected {
			dead = domain.Code == "invalid_message" || domain.Code == "unsupported_version" || domain.Code == "unsupported_action"
			s.log.Warn("request rejected", zap.String("id", logID(request.ID)), zap.String("code", domain.Code))
		}
	}

	response := protocol.Reply(message.CorrelationId, data, requestErr)
	if err := s.reply(ctx, message.ReplyTo, response, headers); err != nil {
		return err
	}
	if dead {
		s.log.Warn("message sent to dlq", zap.String("id", logID(request.ID)), zap.Int("attempt", attempt))
		err := message.Reject(false)
		return err
	}
	s.log.Info("request completed", zap.String("id", logID(request.ID)), zap.String("status", response.Status), zap.Int("attempt", attempt))
	err := message.Ack(false)
	return err
}

func (s *Server) reply(ctx context.Context, queue string, response protocol.Response, headers amqp.Table) error {
	if !strings.HasPrefix(queue, "amq.gen-") {
		return nil
	}
	body, err := json.Marshal(response)
	if err != nil {
		return err
	}
	err = s.session.Publish(ctx, queue, amqp.Publishing{ContentType: "application/json", CorrelationId: response.CorrelationID, Body: body, Headers: headers})
	if errors.Is(err, rabbitmq.ErrUnroutable) {
		s.log.Warn("client reply queue no longer exists", zap.String("id", logID(response.CorrelationID)))
		return nil
	}
	return err
}

func (s *Server) retry(ctx context.Context, message amqp.Delivery, attempt int) error {
	publication := amqp.Publishing{ContentType: message.ContentType, DeliveryMode: amqp.Persistent,
		CorrelationId: message.CorrelationId, MessageId: message.MessageId, ReplyTo: message.ReplyTo,
		Body: message.Body, Headers: amqp.Table{"x-attempt": int32(attempt + 1)}, Timestamp: time.Now().UTC()}
	if err := s.session.Publish(ctx, rabbitmq.Retry, publication); err != nil {
		return err
	}
	err := message.Ack(false)
	return err
}

func attempts(headers amqp.Table) int {
	if attempt, ok := headers["x-attempt"].(int32); ok && attempt >= 1 && attempt <= rabbitmq.MaxAttempts {
		number := int(attempt)

		return number
	}
	return 1
}

func logID(value string) string {
	id, err := uuid.Parse(value)
	if err != nil {
		return "invalid"
	}
	result := id.String()
	return result
}
