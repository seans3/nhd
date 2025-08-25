package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"firebase.google.com/go/v4/auth"
	"github.com/seans3/nhd/backend/interfaces"
	"github.com/seans3/nhd/backend/memstore"
	"github.com/seans3/nhd/backend/mocks"
	"github.com/seans3/nhd/backend/middleware"
	"github.com/seans3/nhd/backend/proto/gen/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// setupIntegrationTestServer initializes a new test server with an in-memory datastore
// and a mock publisher/auth client.
func setupIntegrationTestServer() (*httptest.Server, interfaces.Datastore, *mocks.MockPublisherClient, *mocks.MockFirebaseAuth, func()) {
	memDS := memstore.NewClient()
	mockPS := new(mocks.MockPublisherClient)
	mockAuth := new(mocks.MockFirebaseAuth)

	apiHandler := &API{
		DS: memDS,
		PS: mockPS,
	}

	authClient := &middleware.AuthClient{
		Firebase: mockAuth,
		DS:       memDS,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /customers", apiHandler.GetCustomers)
	mux.HandleFunc("POST /customers", apiHandler.CreateCustomer)
	mux.HandleFunc("POST /report-runs", apiHandler.CreateReportRun)
	mux.HandleFunc("GET /report-runs", apiHandler.GetReportRuns)
	mux.HandleFunc("PUT /report-runs/{id}/cost", apiHandler.UpdateReportCost)
	mux.HandleFunc("POST /report-runs/{id}/payment", apiHandler.RecordReportPayment)
	mux.HandleFunc("GET /financials/summary", apiHandler.GetFinancialsSummary)
	mux.Handle("POST /users/register", authClient.RequireAdmin(http.HandlerFunc(apiHandler.RegisterUser)))

	server := httptest.NewServer(mux)
	cleanup := func() {
		server.Close()
	}

	return server, memDS, mockPS, mockAuth, cleanup
}

func TestIntegration_CreateAndGetCustomers(t *testing.T) {
	server, _, _, _, cleanup := setupIntegrationTestServer()
	defer cleanup()

	// 1. Create a new customer
	customerToCreate := `{"full_name":"Charlie Integration","email":"charlie@it.com"}`
	resp, err := http.Post(server.URL+"/customers", "application/json", bytes.NewBufferString(customerToCreate))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	defer resp.Body.Close()

	var result map[string]string
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	newCustomerID := result["customer_id"]
	assert.NotEmpty(t, newCustomerID)

	// 2. Get all customers and verify the new one is there
	resp, err = http.Get(server.URL + "/customers")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()

	var customers []*nhd_report.Customer
	err = json.NewDecoder(resp.Body).Decode(&customers)
	assert.NoError(t, err)
	assert.Len(t, customers, 1)
	assert.Equal(t, "Charlie Integration", customers[0].FullName)
	assert.Equal(t, newCustomerID, customers[0].CustomerId)
}

func TestIntegration_FullReportLifecycle(t *testing.T) {
	server, memDS, mockPS, _, cleanup := setupIntegrationTestServer()
	defer cleanup()

	// 1. Create a customer first
	customer := &nhd_report.Customer{FullName: "Test Customer"}
	docRef, _, err := memDS.CreateCustomer(context.Background(), customer)
	assert.NoError(t, err)
	customerID := docRef.ID

	// 2. Create a report run for that customer
	reportRunJSON := `{"customer_id":"` + customerID + `","property_address_id":"addr123"}`
	mockPS.On("Publish", mock.Anything, "nhd-report-requests", mock.Anything).Return("pub-msg-id", nil)

	resp, err := http.Post(server.URL+"/report-runs", "application/json", bytes.NewBufferString(reportRunJSON))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	defer resp.Body.Close()

	var createResult map[string]string
	err = json.NewDecoder(resp.Body).Decode(&createResult)
	assert.NoError(t, err)
	reportID := createResult["report_run_id"]
	assert.NotEmpty(t, reportID)

	// 4. Get all reports and verify the new one is there
	resp, err = http.Get(server.URL + "/report-runs")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()

	var reports []*nhd_report.ReportRun
	err = json.NewDecoder(resp.Body).Decode(&reports)
	assert.NoError(t, err)
	assert.Len(t, reports, 1)
	assert.Equal(t, reportID, reports[0].ReportRunId)

	// 5. Update the cost of the report
	costJSON := `{"amount":123.45, "currency": "USD"}`
	req, err := http.NewRequest("PUT", server.URL+"/report-runs/"+reportID+"/cost", bytes.NewBufferString(costJSON))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{}
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()

	// 6. Verify the cost was updated in the datastore
	updatedReport, err := memDS.GetReportRuns(context.Background(), "")
	assert.NoError(t, err)
	assert.Len(t, updatedReport, 1)
	assert.Len(t, updatedReport[0].CostHistory, 1)
	assert.Equal(t, 123.45, updatedReport[0].CostHistory[0].Amount)

	// 7. Record a payment for the report
	paymentJSON := `{"amount_paid":123.45, "status": 2}` // PAID = 2
	resp, err = http.Post(server.URL+"/report-runs/"+reportID+"/payment", "application/json", bytes.NewBufferString(paymentJSON))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()

	// 8. Get the financials summary and verify the payment is included
	resp, err = http.Get(server.URL + "/financials/summary")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()

	var summary interfaces.FinancialsSummary
	err = json.NewDecoder(resp.Body).Decode(&summary)
	assert.NoError(t, err)
	assert.Equal(t, 123.45, summary.TotalRevenue)
	assert.Len(t, summary.PaidReports, 1)
	assert.Equal(t, "Customer "+customerID, summary.PaidReports[0].CustomerName)
}

func TestIntegration_RegisterUser_FullFlow(t *testing.T) {
	server, memDS, _, mockAuth, cleanup := setupIntegrationTestServer()
	defer cleanup()

	// --- Part 1: Test Forbidden ---
	// 1. Setup Mocks for a non-admin user
	mockAuth.On("VerifyIDToken", mock.Anything, "valid-non-admin-token").Return(&auth.Token{UID: "non-admin-uid"}, nil)
	nonAdminUser := &nhd_report.User{UserId: "non-admin-uid", Permissions: &nhd_report.Permissions{IsAdmin: false}}
	err := memDS.CreateUser(context.Background(), nonAdminUser)
	assert.NoError(t, err)

	// 2. Prepare and make the request
	newUserJSON := `{"email":"newuser@example.com","password":"password","full_name":"New User"}`
	req, err := http.NewRequest("POST", server.URL+"/users/register", bytes.NewBufferString(newUserJSON))
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer valid-non-admin-token")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// 3. Assert Forbidden
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)

	// --- Part 2: Test Success ---
	// 1. Setup Mocks for an admin user
	mockAuth.On("VerifyIDToken", mock.Anything, "valid-admin-token").Return(&auth.Token{UID: "admin-uid"}, nil)
	adminUser := &nhd_report.User{UserId: "admin-uid", Permissions: &nhd_report.Permissions{IsAdmin: true}}
	err = memDS.CreateUser(context.Background(), adminUser)
	assert.NoError(t, err)

	// 2. Prepare and make the request again, this time as an admin
	req, err = http.NewRequest("POST", server.URL+"/users/register", bytes.NewBufferString(newUserJSON))
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer valid-admin-token")
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// 3. Assert Success
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}