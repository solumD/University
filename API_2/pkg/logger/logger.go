package logger

import "go.uber.org/zap"

func New(level string) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{"stdout"}
	cfg.ErrorOutputPaths = []string{"stdout"}

	if err := cfg.Level.UnmarshalText([]byte(level)); err != nil {
		return nil, err
	}

	log, err := cfg.Build()

	return log, err
}
