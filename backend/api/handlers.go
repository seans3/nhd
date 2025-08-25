package api

import (
	"encoding/json"
	"net/http"

	"github.com/seans3/nhd/backend/datastore"
	"github.com/seans3/nhd/backend/publisher"
	"github.com/seans3/nhd/proto/gen/go;nhd_report"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type API struct {
	DS *datastore.Client
	PS *publisher.Client
}

// Users
func (a *API) RegisterUser(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// Customers
func (a *API) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	var customer nhd_report.Customer
	if err := json.NewDecoder(r.Body).Decode(&customer); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	docRef, _, err := a.DS.CreateCustomer(r.Context(), &customer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"customer_id": docRef.ID})
}

func (a *API) GetCustomers(w http.ResponseWriter, r *http.Request) {
	customers, err := a.DS.GetCustomers(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(customers)
}

// Report Runs
func (a *API) CreateReportRun(w http.ResponseWriter, r *http.Request) {
	var reportRun nhd_report.ReportRun
	if err := json.NewDecoder(r.Body).Decode(&reportRun); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	reportRun.Status = nhd_report.ReportRun_PENDING
	reportRun.CreatedAt = timestamppb.Now()

	docRef, _, err := a.DS.CreateReportRun(r.Context(), &reportRun)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = a.PS.Publish(r.Context(), "nhd-report-requests", []byte(docRef.ID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"report_run_id": docRef.ID})
}

func (a *API) GetReportRuns(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (a *API) ResendReportEmail(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (a *API) UpdateReportCost(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (a *API) RecordReportPayment(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// Financials
func (a *API) GetFinancialsSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := a.DS.GetPaidReportsSummary(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(summary)
}