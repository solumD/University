package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"gitlab.ozon.ru/university/broker-4/config"
	"gitlab.ozon.ru/university/broker-4/internal/protocol"
	"gitlab.ozon.ru/university/broker-4/pkg/rabbitmq"
)

func Request(ctx context.Context, cfg config.Config, request protocol.Request, timeout time.Duration, output, diagnostics io.Writer) (err error) {
	body, err := json.Marshal(request)
	if err != nil {
		return err
	}
	if _, err = protocol.Decode(body); err != nil {
		return err
	}
	session, err := rabbitmq.Open(ctx, cfg.RabbitURL)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, session.Close()) }()
	if err = session.Declare(ctx); err != nil {
		return err
	}
	queue, err := session.ReplyQueue(ctx)
	if err != nil {
		return err
	}
	deliveries, err := session.Consume(ctx, queue, true)
	if err != nil {
		return err
	}

	if _, err = fmt.Fprintln(diagnostics, "request_id:", request.ID); err != nil {
		return err
	}
	err = session.Publish(ctx, rabbitmq.Requests, amqp.Publishing{
		ContentType: "application/json", DeliveryMode: amqp.Persistent,
		MessageId: request.ID, CorrelationId: request.ID, ReplyTo: queue,
		Timestamp: time.Now().UTC(), Body: body,
	})
	if err != nil {
		return err
	}

	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	err = awaitReply(waitCtx, deliveries, request.ID, output, diagnostics)
	return err
}

func awaitReply(ctx context.Context, deliveries <-chan amqp.Delivery, id string, output, diagnostics io.Writer) error {
	for {
		select {
		case <-ctx.Done():
			validationErr := errors.New("response was not received; operation may have completed, retry with the same id")

			return validationErr
		case message, ok := <-deliveries:
			if !ok {
				validationErr := errors.New("reply connection closed; retry with the same id")

				return validationErr
			}
			if message.CorrelationId != id {
				if err := message.Ack(false); err != nil {
					return err
				}
				continue
			}

			var response protocol.Response
			decoder := json.NewDecoder(bytes.NewReader(message.Body))
			decoder.UseNumber()
			if err := decoder.Decode(&response); err != nil {
				return err
			}
			if response.CorrelationID != id {
				validationErr := errors.New("response correlation_id does not match request")

				return validationErr
			}
			pretty := json.NewEncoder(output)
			pretty.SetIndent("", "  ")
			if err := pretty.Encode(response); err != nil {
				return err
			}
			if _, err := fmt.Fprintln(diagnostics, "headers:", message.Headers); err != nil {
				return err
			}
			if err := message.Ack(false); err != nil {
				return err
			}
			if response.Status != "ok" {
				validationErr := errors.New("server returned an error; see JSON response")

				return validationErr
			}
			return nil
		}
	}
}
