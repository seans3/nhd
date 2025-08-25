package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	firebase "firebase.google.com/go/v4"
	"github.com/seans3/nhd/backend/api"
	"github.com/seans3/nhd/backend/datastore"
	"github.com/seans3/nhd/backend/health"
	"github.com/seans3/nhd/backend/metrics"
	"github.com/seans3/nhd/backend/middleware"
	"github.com/seans3/nhd/backend/publisher"
)

// Define constants for the rate limiter and timeout.
const (
	DefaultRateLimitRPS   = 10.0
	DefaultRateLimitBurst = 20
	DefaultRequestTimeout = 30 * time.Second
)

func main() {
	// Define command-line flags for configuration.
	rps := flag.Float64("ratelimit.rps", DefaultRateLimitRPS, "Requests per second for the rate limiter")
	burst := flag.Int("ratelimit.burst", DefaultRateLimitBurst, "Burst size for the rate limiter")
	timeout := flag.Duration("server.timeout", DefaultRequestTimeout, "Request timeout duration")
	flag.Parse()

	ctx := context.Background()
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		log.Fatal("GOOGLE_CLOUD_PROJECT environment variable must be set")
	}

	// Initialize Firebase Admin SDK
	firebaseApp, err := firebase.NewApp(ctx, nil)
	if err != nil {
		log.Fatalf("error initializing Firebase app: %v\n", err)
	}
	firebaseAuth, err := firebaseApp.Auth(ctx)
	if err != nil {
		log.Fatalf("error getting Firebase Auth client: %v\n", err)
	}

	dsClient, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create datastore client: %v", err)
	}
	defer dsClient.Close()

	psClient, err := publisher.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create publisher client: %v", err)
	}
	defer psClient.Close()

	apiHandler := &api.API{
		DS: dsClient,
		PS: psClient,
	}

	authClient := &middleware.AuthClient{
		Firebase: firebaseAuth,
		DS:       dsClient,
	}

	metricsHandler := metrics.NewMetricsHandler()
	readyzHandler := &health.ReadyzHandler{DS: dsClient}

	// --- Create protected sub-routers ---
	apiMux := http.NewServeMux()
	// User
	apiMux.HandleFunc("POST /customers", apiHandler.CreateCustomer)
	apiMux.HandleFunc("GET /customers", apiHandler.GetCustomers)
	// Report Runs
	apiMux.HandleFunc("POST /report-runs", apiHandler.CreateReportRun)
	apiMux.HandleFunc("GET /report-runs", apiHandler.GetReportRuns)
	apiMux.HandleFunc("POST /report-runs/{id}/resend-email", apiHandler.ResendReportEmail)
	// Financials
	apiMux.HandleFunc("GET /financials/summary", apiHandler.GetFinancialsSummary)

	adminMux := http.NewServeMux()
	// User Management
	adminMux.HandleFunc("POST /users/register", apiHandler.RegisterUser)
	// Financial Management
	adminMux.HandleFunc("PUT /report-runs/{id}/cost", apiHandler.UpdateReportCost)
	adminMux.HandleFunc("POST /report-runs/{id}/payment", apiHandler.RecordReportPayment)

	// --- Register all routes ---
	mux := http.NewServeMux()
	// Health and Metrics Probes (public)
	mux.HandleFunc("GET /healthz", health.HealthzHandler)
	mux.Handle("GET /readyz", readyzHandler)
	mux.HandleFunc("GET /metrics", metricsHandler.Handler)
	// Standard authenticated API routes
	mux.Handle("/api/", http.StripPrefix("/api", authClient.VerifyAuthToken(apiMux)))
	// Admin-only API routes
	mux.Handle("/admin/", http.StripPrefix("/admin", authClient.RequireAdmin(adminMux)))

	// Create the rate limiting middleware with the configured values.
	rateLimitMiddleware := middleware.RateLimit(*rps, *burst)

	// Wrap the entire mux with all middleware
	var finalMux http.Handler = mux
	finalMux = middleware.Timeout(finalMux, *timeout)
	finalMux = middleware.Recover(finalMux) // Recover from panics
	finalMux = rateLimitMiddleware(finalMux)
	finalMux = metricsHandler.Middleware(finalMux)
	finalMux = middleware.Logging(finalMux)

	log.Printf("Starting server on :8080 with rate limit of %.2f rps and a burst of %d", *rps, *burst)
	if err := http.ListenAndServe(":8080", finalMux); err != nil {
		log.Fatal(err)
	}
}
