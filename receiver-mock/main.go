package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/Shopify/sarama"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
)

var httpMessagesCount, pubSubMessagesCount uint64

func NewSubscriber(kafkaAddress, group string) (message.Subscriber, error) {
	subscriberConfig := kafka.DefaultSaramaSubscriberConfig()
	subscriberConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	subscriber, err := kafka.NewSubscriber(
		kafka.SubscriberConfig{
			Brokers:               []string{kafkaAddress},
			Unmarshaler:           kafka.DefaultMarshaler{},
			OverwriteSaramaConfig: subscriberConfig,
			ConsumerGroup:         group,
		},
		watermill.NewStdLogger(false, false),
	)
	if err != nil {
		return nil, fmt.Errorf("new subscriber: %w", err)
	}

	return subscriber, nil
}

func start() {
	wg := new(sync.WaitGroup)
	root, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg.Add(2)
	log.Println("Starting http receiver...")
	go HandleHttpReceiver(root, wg)
	log.Println("Starting pub sub receiver...")
	go HandlePubSubReceiver(root, wg)

	osChan := make(chan os.Signal, 1)
	signal.Notify(osChan, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	select {
	case <-osChan:
		log.Println("main: shutting down gracefully...")
		cancel()
	}

	wg.Wait()
	// No extra routines here, so no need in atomic.Load...()
	log.Printf(
		"all tasks exited successfully, total messages recieved: %d (http: %d; pubsub: %d) \n",
		httpMessagesCount+pubSubMessagesCount,
		httpMessagesCount,
		pubSubMessagesCount,
	)

	return
}

func HandleHttpReceiver(ctx context.Context, wg *sync.WaitGroup) {
	srv := &http.Server{Addr: ":8081"}

	http.HandleFunc("/process", func(w http.ResponseWriter, r *http.Request) {
		all, _ := ioutil.ReadAll(r.Body)
		atomic.AddUint64(&httpMessagesCount, 1)

		log.Println("received http message: ", string(all))
	})

	go func() {
		defer wg.Done() // let main know we are done cleaning up

		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	select {
	case <-ctx.Done():
		log.Println("HTTP: context is closed")

		err := srv.Shutdown(context.TODO())
		if err != nil {
			log.Println("HTTP: Error closing server:", err)
		}
	}
}

func HandlePubSubReceiver(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	subscriber, err := NewSubscriber("kafka:9092", "device_data_reader_group")
	if err != nil {
		log.Fatal(err)
	}

	messagesChan, err := subscriber.Subscribe(ctx, "device_data")
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("PUBSUB: context is closed")

			return
		case msg, ok := <-messagesChan:
			if !ok {
				log.Println("messagesChan is closed")
				return
			}

			msg.Ack()

			atomic.AddUint64(&pubSubMessagesCount, 1)
			log.Println("received pub sub message: ", string(msg.Payload))
		}
	}
}

func main() {
	log.Println("Starting receiver...")
	start()
}
