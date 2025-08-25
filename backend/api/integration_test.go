package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"cloud.google.com/go/firestore"
	"github.com/seans3/nhd/backend/interfaces"
	"github.com/seans3/nhd/backend/mocks"
	"github.com/seans3/nhd/backend/proto/gen/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// setupTestServer initializes a new test server with mock dependencies.
// It returns the server, the mock datastore, the mock publisher, and a cleanup function.
func setupTestServer() (*httptest.Server, *mocks.MockDatastoreClient, *mocks.MockPublisherClient, func()) {
	mockDS := new(mocks.MockDatastoreClient)
	mockPS := new(mocks.MockPublisherClient)

	apiHandler := &API{
		DS: mockDS,
		PS: mockPS,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /customers", apiHandler.GetCustomers)
	mux.HandleFunc("POST /customers", apiHandler.CreateCustomer)
	mux.HandleFunc("POST /report-runs", apiHandler.CreateReportRun)
	mux.HandleFunc("GET /report-runs", apiHandler.GetReportRuns)
	mux.HandleFunc("PUT /report-runs/{id}/cost", apiHandler.UpdateReportCost)
	mux.HandleFunc("POST /report-runs/{id}/payment", apiHandler.RecordReportPayment)
	mux.HandleFunc("GET /financials/summary", apiHandler.GetFinancialsSummary)

	server := httptest.NewServer(mux)
	cleanup := func() {
		server.Close()
	}

	return server, mockDS, mockPS, cleanup
}

func TestIntegration_GetCustomers(t *testing.T) {
	server, mockDS, _, cleanup := setupTestServer()
	defer cleanup()

	// Define the expected data
	expectedCustomers := []*nhd_report.Customer{
		{CustomerId: "cust1", FullName: "Alice Integration"},
		{CustomerId: "cust2", FullName: "Bob Integration"},
	}

	// Set up the mock expectation
	mockDS.On("GetCustomers", mock.Anything).Return(expectedCustomers, nil)

	// Make a real HTTP request to the test server
	resp, err := http.Get(server.URL + "/customers")
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Verify the response
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var actualCustomers []*nhd_report.Customer
	err = json.NewDecoder(resp.Body).Decode(&actualCustomers)
	assert.NoError(t, err)
	assert.Equal(t, len(expectedCustomers), len(actualCustomers))
	assert.Equal(t, "Alice Integration", actualCustomers[0].FullName)

	// Verify that the mock was called
	mockDS.AssertExpectations(t)
}

func TestIntegration_CreateCustomer(t *testing.T) {
	server, mockDS, _, cleanup := setupTestServer()
	defer cleanup()

	// Prepare request body
	customerToCreate := nhd_report.Customer{FullName: "Charlie Integration", Email: "charlie@it.com"}
	body, err := json.Marshal(customerToCreate)
	assert.NoError(t, err)

	// Set up the mock expectation
	mockDocRef := &firestore.DocumentRef{ID: "new-cust-id"}
	mockDS.On("CreateCustomer", mock.Anything, mock.AnythingOfType("*nhd_report.Customer")).Return(mockDocRef, (*firestore.WriteResult)(nil), nil)

	// Make a real HTTP request
	resp, err := http.Post(server.URL+"/customers", "application/json", bytes.NewBuffer(body))
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Verify the response
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var result map[string]string
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	assert.Equal(t, "new-cust-id", result["customer_id"])

	// Verify that the mock was called
	mockDS.AssertExpectations(t)
}

func TestIntegration_CreateReportRun(t *testing.T) {
	server, mockDS, mockPS, cleanup := setupTestServer()
	defer cleanup()

	reportRunJSON := `{"customer_id":"cust123","property_address_id":"addr123"}`
	body := bytes.NewBufferString(reportRunJSON)

	mockDocRef := &firestore.DocumentRef{ID: "new-report-id"}
	mockDS.On("CreateReportRun", mock.Anything, mock.AnythingOfType("*nhd_report.ReportRun")).Return(mockDocRef, (*firestore.WriteResult)(nil), nil)
	mockPS.On("Publish", mock.Anything, "nhd-report-requests", []byte(mockDocRef.ID)).Return("pub-msg-id", nil)

	resp, err := http.Post(server.URL+"/report-runs", "application/json", body)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var result map[string]string
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	assert.Equal(t, "new-report-id", result["report_run_id"])

	mockDS.AssertExpectations(t)
	mockPS.AssertExpectations(t)
}

func TestIntegration_GetReportRuns(t *testing.T) {
	server, mockDS, _, cleanup := setupTestServer()
	defer cleanup()

	expectedReports := []*nhd_report.ReportRun{
		{ReportRunId: "run1"},
		{ReportRunId: "run2"},
	}
	mockDS.On("GetReportRuns", mock.Anything, "").Return(expectedReports, nil)

	resp, err := http.Get(server.URL + "/report-runs")
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var actualReports []*nhd_report.ReportRun
	err = json.NewDecoder(resp.Body).Decode(&actualReports)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(actualReports))
	assert.Equal(t, "run1", actualReports[0].ReportRunId)

	mockDS.AssertExpectations(t)
}

func TestIntegration_UpdateReportCost(t *testing.T) {
	server, mockDS, _, cleanup := setupTestServer()
	defer cleanup()

	costJSON := `{"amount":123.45}`
	body := bytes.NewBufferString(costJSON)

	mockDS.On("UpdateReportCost", mock.Anything, "run123", mock.AnythingOfType("*nhd_report.ReportRun_ReportCost")).Return(nil)

	req, err := http.NewRequest("PUT", server.URL+"/report-runs/run123/cost", body)
	assert.NoError(t, err)
	
	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockDS.AssertExpectations(t)
}

func TestIntegration_RecordReportPayment(t *testing.T) {
	server, mockDS, _, cleanup := setupTestServer()
	defer cleanup()

	paymentJSON := `{"amount_paid":123.45}`
	body := bytes.NewBufferString(paymentJSON)

	mockDS.On("RecordReportPayment", mock.Anything, "run123", mock.AnythingOfType("*nhd_report.ReportRun_Payment")).Return(nil)

	resp, err := http.Post(server.URL+"/report-runs/run123/payment", "application/json", body)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockDS.AssertExpectations(t)
}

func TestIntegration_GetFinancialsSummary(t *testing.T) {
	server, mockDS, _, cleanup := setupTestServer()
	defer cleanup()

	expectedSummary := &interfaces.FinancialsSummary{
		TotalRevenue: 100.0,
		PaidReports:  []interfaces.PaidReportInfo{{CustomerName: "Test Cust"}},
	}
	mockDS.On("GetPaidReportsSummary", mock.Anything).Return(expectedSummary, nil)

	resp, err := http.Get(server.URL + "/financials/summary")
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var actualSummary interfaces.FinancialsSummary
	err = json.NewDecoder(resp.Body).Decode(&actualSummary)
	assert.NoError(t, err)
	assert.Equal(t, expectedSummary.TotalRevenue, actualSummary.TotalRevenue)
	assert.Equal(t, len(expectedSummary.PaidReports), len(actualSummary.PaidReports))

	mockDS.AssertExpectations(t)
}
