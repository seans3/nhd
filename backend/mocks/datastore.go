package mocks

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/seans3/nhd/backend/interfaces"
	"github.com/seans3/nhd/backend/proto/gen/go"
	"github.com/stretchr/testify/mock"
)

// Statically assert that our mock satisfies the interface.
var _ interfaces.Datastore = (*MockDatastoreClient)(nil)

// MockDatastoreClient is a mock implementation of the Datastore interface.
type MockDatastoreClient struct {
	mock.Mock
}

func (m *MockDatastoreClient) GetCustomers(ctx context.Context) ([]*nhd_report.Customer, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*nhd_report.Customer), args.Error(1)
}

func (m *MockDatastoreClient) CreateCustomer(ctx context.Context, customer *nhd_report.Customer) (*firestore.DocumentRef, *firestore.WriteResult, error) {
	args := m.Called(ctx, customer)
	return args.Get(0).(*firestore.DocumentRef), args.Get(1).(*firestore.WriteResult), args.Error(2)
}

func (m *MockDatastoreClient) CreateReportRun(ctx context.Context, reportRun *nhd_report.ReportRun) (*firestore.DocumentRef, *firestore.WriteResult, error) {
	args := m.Called(ctx, reportRun)
	return args.Get(0).(*firestore.DocumentRef), args.Get(1).(*firestore.WriteResult), args.Error(2)
}

func (m *MockDatastoreClient) GetReportRuns(ctx context.Context, paymentStatusFilter string) ([]*nhd_report.ReportRun, error) {
	args := m.Called(ctx, paymentStatusFilter)
	return args.Get(0).([]*nhd_report.ReportRun), args.Error(1)
}

func (m *MockDatastoreClient) UpdateReportCost(ctx context.Context, reportRunID string, newCost *nhd_report.ReportRun_ReportCost) error {
	args := m.Called(ctx, reportRunID, newCost)
	return args.Error(0)
}

func (m *MockDatastoreClient) RecordReportPayment(ctx context.Context, reportRunID string, payment *nhd_report.ReportRun_Payment) error {
	args := m.Called(ctx, reportRunID, payment)
	return args.Error(0)
}

func (m *MockDatastoreClient) GetPaidReportsSummary(ctx context.Context) (*interfaces.FinancialsSummary, error) {
	args := m.Called(ctx)
	return args.Get(0).(*interfaces.FinancialsSummary), args.Error(1)
}

func (m *MockDatastoreClient) GetUserByID(ctx context.Context, uid string) (*nhd_report.User, error) {
	args := m.Called(ctx, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*nhd_report.User), args.Error(1)
}

func (m *MockDatastoreClient) CreateUser(ctx context.Context, user *nhd_report.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}
