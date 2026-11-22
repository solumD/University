package rabbitmq

import (
	"context"
	"errors"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"gitlab.ozon.ru/university/broker-4/config"
)

const (
	Requests    = "api.requests"
	Retry       = "api.retry"
	DLQ         = "api.dlq"
	MaxAttempts = 3
	RetryDelay  = 2 * time.Second
)

func (s *Session) Declare(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, config.OperationTimeout)
	defer cancel()
	stop := s.interruptOnCancel(ctx)
	defer stop()
	queues := []struct {
		name string
		args amqp.Table
	}{
		{DLQ, nil},
		{Requests, amqp.Table{"x-dead-letter-exchange": "", "x-dead-letter-routing-key": DLQ}},
		{Retry, amqp.Table{"x-message-ttl": int32(RetryDelay.Milliseconds()), "x-dead-letter-exchange": "", "x-dead-letter-routing-key": Requests}},
	}
	for _, queue := range queues {
		if _, err := s.Channel.QueueDeclare(queue.name, true, false, false, false, queue.args); err != nil {
			return err
		}
	}
	return nil
}

func (s *Session) Consume(ctx context.Context, queue string, exclusive bool) (<-chan amqp.Delivery, error) {
	ctx, cancel := context.WithTimeout(ctx, config.OperationTimeout)
	defer cancel()
	stop := s.interruptOnCancel(ctx)
	defer stop()
	if err := s.Channel.Qos(1, 0, false); err != nil {
		return nil, err
	}
	deliveries, err := s.Channel.Consume(queue, "", false, exclusive, false, false, nil)
	return deliveries, err
}

func (s *Session) ReplyQueue(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, config.OperationTimeout)
	defer cancel()
	stop := s.interruptOnCancel(ctx)
	defer stop()
	queue, err := s.Channel.QueueDeclare("", false, true, true, false, nil)
	return queue.Name, err
}

func Closed(ctx context.Context) error {
	if ctx.Err() != nil {
		return nil
	}
	err := errors.New("rabbitmq connection closed; restart the application")
	return err
}
