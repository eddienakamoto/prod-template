package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

var requestCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests processed, labeled by status code, method, and endpoint.",
	},
	[]string{"status", "method", "endpoint"},
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	pgconnStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		"localhost",
		"5432",
		"genome",
		"genome",
		"tester",
		"disable")
	pgconn, err := pgxpool.Connect(context.Background(), pgconnStr)
	if err != nil {
		fmt.Printf("failed to connect to postgres: %v\n", err)
		return
	}
	defer pgconn.Close()
	if err := pgconn.Ping(context.Background()); err != nil {
		fmt.Printf("failed to ping postgres: %v\n", err)
		return
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		choice := rand.IntN(3) + 1
		if choice == 1 {
			requestCounter.WithLabelValues("400", r.Method, r.URL.Path).Inc()
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Bad Request"))

			pgconn.Exec(r.Context(), `INSERT INTO dummy(description) VALUES($1)`, fmt.Sprintf("Bad Request: %d at %v", choice, time.Now()))
			pgconn.Exec(r.Context(), `INSERT INTO again(description) VALUES($1)`, fmt.Sprintf("Bad Request: %d at %v", choice, time.Now()))

			logger.Error("Request failed",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("status", "400"))
			return
		}

		requestCounter.WithLabelValues("200", r.Method, r.URL.Path).Inc()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, world!"))

		pgconn.Exec(r.Context(), `INSERT INTO dummy(description) VALUES($1)`, fmt.Sprintf("Good Request: %d at %v", choice, time.Now()))
		pgconn.Exec(r.Context(), `INSERT INTO again(description) VALUES($1)`, fmt.Sprintf("Good Request: %d at %v", choice, time.Now()))

		logger.Info("Request processed",
			zap.String("secret", os.Getenv("MY_SECRET")),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("status", "200"))
	})

	http.Handle("/metrics", promhttp.Handler())

	prometheus.MustRegister(requestCounter)

	logger.Info("Starting server on :8080...")
	http.ListenAndServe(fmt.Sprintf("%s:%s", os.Getenv("HOST"), os.Getenv("PORT")), nil)
}
