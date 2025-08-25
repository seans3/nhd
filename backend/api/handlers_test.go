package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/seans3/nhd/backend/mocks"
	"github.com/seans3/nhd/backend/proto/gen/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAPI_GetCustomers(t *testing.T) {
	// Create a mock datastore client
	mockDS := new(mocks.MockDatastoreClient)

	// Create the API handler with the mock client
	apiHandler := &API{
		DS: mockDS,
	}

	// Define the expected data that the mock should return
	expectedCustomers := []*nhd_report.Customer{
		{CustomerId: "cust1", FullName: "Alice"},
		{CustomerId: "cust2", FullName: "Bob"},
	}

	// Set up the expectation on the mock
	mockDS.On("GetCustomers", mock.Anything).Return(expectedCustomers, nil)

	// Create a request to pass to our handler
	req, err := http.NewRequest("GET", "/customers", nil)
	assert.NoError(t, err)

	// We use httptest.NewRecorder to record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(apiHandler.GetCustomers)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check the response body is what we expect.
	var actualCustomers []*nhd_report.Customer
	err = json.Unmarshal(rr.Body.Bytes(), &actualCustomers)
	assert.NoError(t, err)
	assert.Equal(t, len(expectedCustomers), len(actualCustomers))
	assert.Equal(t, expectedCustomers[0].FullName, actualCustomers[0].FullName)

	// Assert that the expectations were met
	mockDS.AssertExpectations(t)
}
