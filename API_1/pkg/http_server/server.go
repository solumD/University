package httpserver

import (
	"context"
	"errors"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type Server struct {
	http *http.Server
	log  *zap.Logger
}

func New(address string, handler http.Handler, log *zap.Logger) *Server {
	server := &Server{
		http: &http.Server{
			Addr:              address,
			Handler:           handler,
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       10 * time.Second,
			WriteTimeout:      15 * time.Second,
			IdleTimeout:       60 * time.Second,
			MaxHeaderBytes:    16 << 10,
		},
		log: log,
	}

	return server
}

func (s *Server) Run(ctx context.Context) error {
	result := make(chan error, 1)

	go func() {
		err := s.http.ListenAndServe()
		result <- err
	}()

	select {
	case err := <-result:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.http.Shutdown(shutdownCtx); err != nil {
		closeErr := s.http.Close()
		err = errors.Join(err, closeErr)

		return err
	}

	if err := <-result; !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	s.log.Info("server stopped")

	return nil
}
