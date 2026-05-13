package natsx

import (
	"context"
	"errors"
	"log/slog"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

const (
	StreamName     = "FB_EVENTS"
	StreamSubjects = "fb.>"
)

// EnsureStream creates the FB_EVENTS stream if it does not exist.
// It does not fail if the stream already exists with compatible config.
func EnsureStream(ctx context.Context, nc *nats.Conn) (jetstream.JetStream, error) {
	js, err := jetstream.New(nc)
	if err != nil {
		return nil, err
	}

	_, err = js.CreateStream(ctx, jetstream.StreamConfig{
		Name:      StreamName,
		Subjects:  []string{StreamSubjects},
		Storage:   jetstream.FileStorage,
		Retention: jetstream.LimitsPolicy,
	})
	if err != nil {
		var apiErr *jetstream.APIError
		if errors.As(err, &apiErr) && apiErr.ErrorCode == jetstream.JSErrCodeStreamNameInUse {
			slog.Info("NATS stream already exists", "stream", StreamName)
			stream, streamErr := js.Stream(ctx, StreamName)
			if streamErr != nil {
				return nil, streamErr
			}
			_ = stream
			return js, nil
		}
		return nil, err
	}

	slog.Info("NATS stream created", "stream", StreamName)
	return js, nil
}

// Connect connects to NATS with default options.
func Connect(url string) (*nats.Conn, error) {
	return nats.Connect(url,
		nats.Name("fb-market-watcher"),
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(-1),
	)
}
