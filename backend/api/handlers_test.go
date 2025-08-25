package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cloud.google.com/go/firestore"
	"github.com/seans3/nhd/backend/middleware"
	"github.com/seans3/nhd/backend/mocks"
	"github.com/seans3/nhd/backend/proto/gen/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAPI_CreateCustomer(t *testing.T) {
	mockDS := new(mocks.MockDatastoreClient)
	apiHandler := &API{DS: mockDS}

	customerJSON := `{"full_name":"Test User","email":"test@example.com"}`
	req, err := http.NewRequest("POST", "/customers", strings.NewReader(customerJSON))
	assert.NoError(t, err)

	// Add the user ID to the request context to simulate an authenticated user
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "test-user-id")
	req = req.WithContext(ctx)

	mockDocRef := &firestore.DocumentRef{ID: "test-id"}
	mockDS.On("CreateCustomer", mock.Anything, mock.AnythingOfType("*nhd_report.Customer")).Return(mockDocRef, (*firestore.WriteResult)(nil), nil)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(apiHandler.CreateCustomer)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Contains(t, rr.Body.String(), `"customer_id":"test-id"`)
	mockDS.AssertExpectations(t)
}

func TestAPI_GetReportRuns(t *testing.T) {
	mockDS := new(mocks.MockDatastoreClient)
	apiHandler := &API{DS: mockDS}

	expectedReports := []*nhd_report.ReportRun{
		{ReportRunId: "run1"},
		{ReportRunId: "run2"},
	}

	mockDS.On("GetReportRuns", mock.Anything, "").Return(expectedReports, nil)

	req, err := http.NewRequest("GET", "/report-runs", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(apiHandler.GetReportRuns)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var actualReports []*nhd_report.ReportRun
	err = json.Unmarshal(rr.Body.Bytes(), &actualReports)
	assert.NoError(t, err)
	assert.Equal(t, len(expectedReports), len(actualReports))
	mockDS.AssertExpectations(t)
}

func TestAPI_UpdateReportCost(t *testing.T) {
	mockDS := new(mocks.MockDatastoreClient)
	apiHandler := &API{DS: mockDS}

	costJSON := `{"amount":99.99,"currency":"USD"}`
	req, err := http.NewRequest("PUT", "/report-runs/run123/cost", strings.NewReader(costJSON))
	assert.NoError(t, err)
	req.SetPathValue("id", "run123") // Set path value for Go 1.22+ mux

	mockDS.On("UpdateReportCost", mock.Anything, "run123", mock.AnythingOfType("*nhd_report.ReportRun_ReportCost")).Return(nil)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(apiHandler.UpdateReportCost)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockDS.AssertExpectations(t)
}

func TestAPI_RecordReportPayment(t *testing.T) {
	mockDS := new(mocks.MockDatastoreClient)
	apiHandler := &API{DS: mockDS}

	paymentJSON := `{"amount_paid":99.99,"currency":"USD"}`
	req, err := http.NewRequest("POST", "/report-runs/run123/payment", strings.NewReader(paymentJSON))
	assert.NoError(t, err)
	req.SetPathValue("id", "run123")

	mockDS.On("RecordReportPayment", mock.Anything, "run123", mock.AnythingOfType("*nhd_report.ReportRun_Payment")).Return(nil)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(apiHandler.RecordReportPayment)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockDS.AssertExpectations(t)
}


