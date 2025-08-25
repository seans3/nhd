package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/seans3/nhd/backend/api"
	"github.com/seans3/nhd/backend/datastore"
	"github.com/seans3/nhd/backend/publisher"
)

func main() {
	ctx := context.Background()
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		log.Fatal("GOOGLE_CLOUD_PROJECT environment variable must be set")
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

	mux := http.NewServeMux()

	// User
	mux.HandleFunc("POST /users/register", apiHandler.RegisterUser)

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

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
