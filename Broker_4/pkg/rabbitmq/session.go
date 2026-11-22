package rabbitmq

import (
	"context"
	"errors"
	"net"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"gitlab.ozon.ru/university/broker-4/config"
	"go.uber.org/zap"
)

var ErrUnroutable = errors.New("message has no destination queue")

type Session struct {
	Connection *amqp.Connection
	Channel    *amqp.Channel
	returns    <-chan amqp.Return
}

type timedConnection struct{ net.Conn }

func (c timedConnection) Write(data []byte) (int, error) {
	if err := c.SetWriteDeadline(time.Now().Add(config.OperationTimeout)); err != nil {
		return 0, err
	}
	n, err := c.Conn.Write(data)
	return n, err
}

func Open(ctx context.Context, address string) (*Session, error) {
	var network net.Conn
	dialer := net.Dialer{Timeout: config.OperationTimeout}
	cfg := amqp.Config{Heartbeat: 5 * time.Second}
	cfg.Dial = func(networkType, address string) (net.Conn, error) {
		connection, err := dialer.DialContext(ctx, networkType, address)
		if err != nil {
			return nil, err
		}
		network = connection
		if err = connection.SetDeadline(time.Now().Add(config.OperationTimeout)); err != nil {
			closeErr := connection.Close()
			combinedErr := errors.Join(err, closeErr)

			return nil, combinedErr
		}
		return timedConnection{connection}, nil
	}

	connection, err := amqp.DialConfig(address, cfg)
	if err != nil {
		validationErr := errors.New("connect to rabbitmq failed; check broker and local configuration")

		return nil, validationErr
	}

	if err = network.SetDeadline(time.Time{}); err != nil {
		closeErr := connection.CloseDeadline(time.Now().Add(config.OperationTimeout))
		combinedErr := errors.Join(err, closeErr)

		return nil, combinedErr
	}

	session := &Session{Connection: connection}
	setupCtx, cancel := context.WithTimeout(ctx, config.OperationTimeout)
	defer cancel()
	stop := session.interruptOnCancel(setupCtx)
	defer stop()
	session.Channel, err = connection.Channel()
	if err == nil {
		err = session.Channel.Confirm(false)
	}
	if err != nil {
		closeErr := session.Close()
		combinedErr := errors.Join(err, closeErr)

		return nil, combinedErr
	}

	session.returns = session.Channel.NotifyReturn(make(chan amqp.Return, 1))
	return session, nil
}

func (s *Session) Publish(ctx context.Context, queue string, message amqp.Publishing) error {
	ctx, cancel := context.WithTimeout(ctx, config.OperationTimeout)
	defer cancel()
	confirmation, err := s.Channel.PublishWithDeferredConfirmWithContext(ctx, "", queue, true, false, message)
	if err != nil {
		return err
	}

	ack, err := confirmation.WaitContext(ctx)
	if err != nil {
		return err
	}
	if !ack {
		validationErr := errors.New("broker rejected publication")

		return validationErr
	}

	select {
	case <-s.returns:
		return ErrUnroutable
	default:
		return nil
	}
}

func (s *Session) Close() error {
	err := s.Connection.CloseDeadline(time.Now().Add(config.OperationTimeout))
	if errors.Is(err, amqp.ErrClosed) {
		return nil
	}
	return err
}

func (s *Session) interruptOnCancel(ctx context.Context) func() bool {
	stop := context.AfterFunc(ctx, func() {
		err := s.Connection.CloseDeadline(time.Now())
		if err != nil && !errors.Is(err, amqp.ErrClosed) {
			log := zap.NewExample()
			log.Warn("interrupted rabbitmq operation", zap.Error(err))
		}
	})
	return stop
}
