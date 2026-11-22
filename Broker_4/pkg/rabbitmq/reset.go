package rabbitmq

import (
	"context"
	"fmt"

	"gitlab.ozon.ru/university/broker-4/config"
)

// ResetQueues recreates the application queues after checking that consumers stopped.
// Publishers must also be stopped before calling it.
func (s *Session) ResetQueues(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, config.OperationTimeout)
	defer cancel()
	stop := s.interruptOnCancel(ctx)
	defer stop()

	if err := s.Declare(ctx); err != nil {
		return err
	}
	names := []string{Requests, Retry, DLQ}
	for _, name := range names {
		queue, err := s.Channel.QueueDeclarePassive(name, true, false, false, false, nil)
		if err != nil {
			return err
		}
		if queue.Consumers != 0 {
			return fmt.Errorf("stop the server and clients before resetting: queue %s has active consumers", name)
		}
	}

	// Remove the retry destination first so delayed messages cannot repopulate it.
	for _, name := range names {
		if _, err := s.Channel.QueueDelete(name, true, false, false); err != nil {
			return err
		}
	}

	err := s.Declare(ctx)
	return err
}
