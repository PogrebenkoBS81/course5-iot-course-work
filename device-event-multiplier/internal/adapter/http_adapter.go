package adapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"device-event-multiplier/internal/domain"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// HTTPAdapter - test server structure.
type HTTPAdapter struct {
	http *http.Server
}

// NewHTTPAdapter - returns new API service instance.
func NewHTTPAdapter(
	publishers map[string]DeviceDataPublisher,
	deviceDataProcessor DeviceDataProcessor,
	httpPort int,
) (*HTTPAdapter, error) {
	s := &HTTPAdapter{
		http: &http.Server{
			Addr:         ":" + strconv.Itoa(httpPort),
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
	}

	// router and routes.
	r := mux.NewRouter()
	for route, publisher := range publishers {
		r.HandleFunc(
			fmt.Sprintf("/start/%s", route),
			handlerStart(
				publisher,
				deviceDataProcessor,
			),
		)
	}

	r.HandleFunc("/health", handlerHealth)
	r.HandleFunc("/readiness", handlerReadiness)

	s.http.Handler = handlers.LoggingHandler(log.Writer(), r)

	return s, nil
}

// Run - start listening.
func (s *HTTPAdapter) Run(ctx context.Context) error {
	log.Println("Starting server...")
	stop := make(chan error, 1)

	go func() {
		if err := s.http.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			stop <- err
		}

		log.Println("HTTP-server: graceful shutdown complete.")
		stop <- nil
	}()

	select {
	case <-ctx.Done():
		return s.shutdown()
	case err := <-stop:
		return err
	}
}

// shutdown - stops the server.
func (s *HTTPAdapter) shutdown() error {
	log.Println("Shutting down...")
	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	// shutdown a server.
	if err := s.http.Shutdown(ctx); err != nil {
		log.Println("Shutdown:", err)
	}

	return nil
}

func handlerStart(
	publisher DeviceDataPublisher,
	deviceDataProcessor DeviceDataProcessor,
) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		log.Printf("Starting cloning with publisher: %T", publisher)
		dd := new(domain.DeviceData)

		if err := json.NewDecoder(request.Body).Decode(dd); err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			log.Println("handlerPOST:", err)

			return
		}

		if err := publisher.PublishDeviceData(deviceDataProcessor.CloneDeviceData(dd)); err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			log.Println("Publish:", err)

			return
		}
	}
}

// handlerHealth - kubernetes health check.
func handlerHealth(_ http.ResponseWriter, _ *http.Request) {}

// handlerReadiness - kubernetes readiness check.
func handlerReadiness(_ http.ResponseWriter, _ *http.Request) {}
