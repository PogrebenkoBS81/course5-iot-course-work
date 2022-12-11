package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"device-event-multiplier/internal/adapter"
	"device-event-multiplier/internal/application"
	infrastructure "device-event-multiplier/internal/infrastructute"
	"device-event-multiplier/pkg/kafka"

	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/sync/errgroup"
)

type Service struct {
	pool *pgxpool.Pool

	runners []Runner
}

func New() *Service {
	return new(Service)
}

func (s *Service) init() (err error) {
	config, err := infrastructure.GetConfig()
	if err != nil {
		return fmt.Errorf("get config: %w", err)
	}

	deviceProcessor := application.NewDeviceProcessor(config.MinVariance, config.MaxVariance, config.CloneNum)

	publisher, err := kafka.NewPublisher(config.KafkaAddr)
	if err != nil {
		return fmt.Errorf("new publisher: %w", err)
	}

	subscriber, err := kafka.NewSubscriber(config.KafkaAddr, config.PubSubSubscribeGroupName)
	if err != nil {
		return fmt.Errorf("new subscriber: %w", err)
	}

	kafkaAdapter, err := adapter.NewPubSubAdapter(
		deviceProcessor,
		publisher,
		subscriber,
		config.PubSubHandlerName,
		config.PubSubSubscribeTopic,
		config.PubSubPublishTopic,
	)
	if err != nil {
		return fmt.Errorf("new kafka adapter: %w", err)
	}
	s.runners = append(s.runners, kafkaAdapter)

	client := infrastructure.NewHTTPClient(config.HTTPInsecureSkipVerify, config.HTTPMillisecondTimeout)
	httpAdapter, err := adapter.NewHTTPAdapter(
		map[string]adapter.DeviceDataPublisher{
			"kafka": infrastructure.NewPubSubDeviceDataPublisher(publisher, config.PubSubPublishTopic),
			"http":  infrastructure.NewHttpDeviceDataPublisher(client, config.HTTPPublishPath),
		},
		deviceProcessor,
		config.HTTPPort,
	)
	if err != nil {
		return fmt.Errorf("new http adapter: %w", err)
	}
	s.runners = append(s.runners, httpAdapter)

	return
}

func (s *Service) Start() {
	if err := s.start(); err != nil {
		log.Fatalf("service: start: %v", err)
	}
}

func (s *Service) start() (err error) {
	root, cancel := context.WithCancel(context.Background())
	defer cancel()

	eg, ctx := errgroup.WithContext(root)

	if err = s.init(); err != nil {
		return fmt.Errorf("service: init: %w", err)
	}

	defer s.pool.Close()

	for _, runner := range s.runners {
		run(ctx, eg, runner)
	}

	errChan := make(chan error, 1)
	go func() { errChan <- eg.Wait() }()

	osChan := make(chan os.Signal, 1)
	signal.Notify(osChan, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	select {
	case <-osChan:
		log.Println("main: shutting down gracefully...")
		cancel()
		err = <-errChan // await for errgroup response.
	case err = <-errChan: // but this can happen earlier due to service error.
	}

	return
}

func run(ctx context.Context, eg *errgroup.Group, svc Runner) {
	eg.Go(func() error { return svc.Run(ctx) })
}
