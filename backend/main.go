package main

import (
	"context"
	"log"
	"net/http"
	"os"

	firebase "firebase.google.com/go/v4"
	"github.com/seans3/nhd/backend/api"
	"github.com/seans3/nhd/backend/datastore"
	"github.com/seans3/nhd/backend/middleware"
	"github.com/seans3/nhd/backend/publisher"
)

func main() {
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

	mux := http.NewServeMux()

	// User
	mux.Handle("POST /users/register", authClient.RequireAdmin(http.HandlerFunc(apiHandler.RegisterUser)))

	// Customers
	mux.HandleFunc("POST /customers", apiHandler.CreateCustomer)
	mux.HandleFunc("GET /customers", apiHandler.GetCustomers)

	// Report Runs
	mux.HandleFunc("POST /report-runs", apiHandler.CreateReportRun)
	mux.HandleFunc("GET /report-runs", apiHandler.GetReportRuns)
	mux.HandleFunc("POST /report-runs/{id}/resend-email", apiHandler.ResendReportEmail)
	mux.HandleFunc("PUT /report-runs/{id}/cost", apiHandler.UpdateReportCost)
	mux.HandleFunc("POST /report-runs/{id}/payment", apiHandler.RecordReportPayment)

	// Financials
	mux.HandleFunc("GET /financials/summary", apiHandler.GetFinancialsSummary)

	// Wrap the entire mux with the logging middleware
	loggedMux := middleware.Logging(mux)

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", loggedMux); err != nil {
		log.Fatal(err)
	}
}
