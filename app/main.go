package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ResponseData represents the JSON structure for GET /projeto-korp
type ResponseData struct {
	Nome    string `json:"nome"`
	Horario string `json:"horario"`
}

var (
	// Metric for request volume
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests processed by http-server-projeto-korp",
		},
		[]string{"path", "method", "status"},
	)

	// Metric for service availability gauge
	httpServerUp = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_server_up",
			Help: "Availability status of http-server-projeto-korp (1 = UP, 0 = DOWN)",
		},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpServerUp)
	httpServerUp.Set(1) // Set status to 1 (UP)
}

func projetoKorpHandler(w http.ResponseWriter, r *http.Request) {
	status := http.StatusOK

	if r.Method != http.MethodGet {
		status = http.StatusMethodNotAllowed
		httpRequestsTotal.WithLabelValues("/projeto-korp", r.Method, strconv.Itoa(status)).Inc()
		http.Error(w, "Method not allowed", status)
		return
	}

	response := ResponseData{
		Nome:    "Projeto Korp",
		Horario: time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}

	httpRequestsTotal.WithLabelValues("/projeto-korp", r.Method, strconv.Itoa(status)).Inc()
}

func main() {
	mux := http.NewServeMux()

	// Endpoints
	mux.HandleFunc("/projeto-korp", projetoKorpHandler)
	mux.Handle("/metrics", promhttp.Handler())

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("Starting http-server-projeto-korp on port 8080...")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}
