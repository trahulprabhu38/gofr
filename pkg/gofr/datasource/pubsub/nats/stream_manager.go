package nats

import (
	"context"
	"errors"

	"github.com/nats-io/nats.go/jetstream"
	"gofr.dev/pkg/gofr/datasource/pubsub"
)

type StreamManager struct {
	js     jetstream.JetStream
	logger pubsub.Logger
}

func NewStreamManager(js jetstream.JetStream, logger pubsub.Logger) *StreamManager {
	return &StreamManager{
		js:     js,
		logger: logger,
	}
}

func (sm *StreamManager) CreateStream(ctx context.Context, cfg StreamConfig) error {
	sm.logger.Debugf("creating stream %s", cfg.Stream)
	jsCfg := jetstream.StreamConfig{
		Name:     cfg.Stream,
		Subjects: cfg.Subjects,
	}

	_, err := sm.js.CreateStream(ctx, jsCfg)
	if err != nil {
		sm.logger.Errorf("failed to create stream: %v", err)
		return err
	}

	return nil
}

func (sm *StreamManager) DeleteStream(ctx context.Context, name string) error {
	sm.logger.Debugf("deleting stream %s", name)

	err := sm.js.DeleteStream(ctx, name)
	if err != nil {
		if errors.Is(err, jetstream.ErrStreamNotFound) {
			sm.logger.Debugf("stream %s not found, considering delete successful", name)
			return nil // If the stream doesn't exist, we consider it a success
		}
		sm.logger.Errorf("failed to delete stream %s: %v", name, err)
		return err
	}

	sm.logger.Debugf("successfully deleted stream %s", name)
	return nil
}

func (sm *StreamManager) CreateOrUpdateStream(ctx context.Context, cfg *jetstream.StreamConfig) (jetstream.Stream, error) {
	sm.logger.Debugf("creating or updating stream %s", cfg.Name)

	stream, err := sm.js.CreateOrUpdateStream(ctx, *cfg)
	if err != nil {
		sm.logger.Errorf("failed to create or update stream: %v", err)
		return nil, err
	}

	return stream, nil
}

func (sm *StreamManager) GetStream(ctx context.Context, name string) (jetstream.Stream, error) {
	sm.logger.Debugf("getting stream %s", name)

	stream, err := sm.js.Stream(ctx, name)
	if err != nil {
		if errors.Is(err, jetstream.ErrStreamNotFound) {
			sm.logger.Debugf("stream %s not found", name)
			return nil, err
		}
		sm.logger.Errorf("failed to get stream %s: %v", name, err)
		return nil, err
	}

	return stream, nil
}
