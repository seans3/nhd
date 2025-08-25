package interfaces

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/seans3/nhd/backend/proto/gen/go"
)

// FinancialsSummary holds the aggregated financial data.
type FinancialsSummary struct {
	TotalRevenue float64        `json:"total_revenue"`
	PaidReports  []PaidReportInfo `json:"paid_reports"`
}

// PaidReportInfo holds data for a single paid report.
type PaidReportInfo struct {
	CustomerName      string  `json:"customer_name"`
	PropertyAddress   string  `json:"property_address"`
	AmountPaid        float64 `json:"amount_paid"`
	PaidAt            string  `json:"paid_at"`
}

// Datastore is an interface for the datastore client to allow for mocking.
type Datastore interface {
	GetCustomers(ctx context.Context) ([]*nhd_report.Customer, error)
	CreateCustomer(ctx context.Context, customer *nhd_report.Customer) (*firestore.DocumentRef, *firestore.WriteResult, error)
	CreateReportRun(ctx context.Context, reportRun *nhd_report.ReportRun) (*firestore.DocumentRef, *firestore.WriteResult, error)
	GetReportRuns(ctx context.Context, paymentStatusFilter string) ([]*nhd_report.ReportRun, error)
	UpdateReportCost(ctx context.Context, reportRunID string, newCost *nhd_report.ReportRun_ReportCost) error
	RecordReportPayment(ctx context.Context, reportRunID string, payment *nhd_report.ReportRun_Payment) error
	GetPaidReportsSummary(ctx context.Context) (*FinancialsSummary, error)
	GetUserByID(ctx context.Context, uid string) (*nhd_report.User, error)
	CreateUser(ctx context.Context, user *nhd_report.User) error
}
