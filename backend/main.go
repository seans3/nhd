package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/seans3/nhd/backend/api"
	"github.comcom/seans3/nhd/backend/datastore"
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

	// User
	http.HandleFunc("/users/register", apiHandler.RegisterUser)

	// Customers
	http.HandleFunc("/customers", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			apiHandler.CreateCustomer(w, r)
		case http.MethodGet:
			apiHandler.GetCustomers(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Report Runs
	http.HandleFunc("/report-runs", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			apiHandler.CreateReportRun(w, r)
		case http.MethodGet:
			apiHandler.GetReportRuns(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/report-runs/{id}/resend-email", apiHandler.ResendReportEmail)
	http.HandleFunc("/report-runs/{id}/cost", apiHandler.UpdateReportCost)
	http.HandleFunc("/report-runs/{id}/payment", apiHandler.RecordReportPayment)

	// Financials
	http.HandleFunc("/financials/summary", apiHandler.GetFinancialsSummary)

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
