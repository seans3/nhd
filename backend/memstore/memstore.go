package memstore

import (
	"context"
	"fmt"
	"sync"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"github.com/seans3/nhd/backend/interfaces"
	"github.com/seans3/nhd/backend/proto/gen/go"
)

// Statically assert that our client satisfies the interface.
var _ interfaces.Datastore = (*Client)(nil)

// Client is a thread-safe, in-memory implementation of the Datastore interface.
type Client struct {
	mu        sync.RWMutex
	users     map[string]*nhd_report.User
	customers map[string]*nhd_report.Customer
	reports   map[string]*nhd_report.ReportRun
}

// NewClient creates a new in-memory datastore client.
func NewClient() *Client {
	return &Client{
		users:     make(map[string]*nhd_report.User),
		customers: make(map[string]*nhd_report.Customer),
		reports:   make(map[string]*nhd_report.ReportRun),
	}
}

// --- User Methods ---

func (c *Client) GetUserByID(ctx context.Context, uid string) (*nhd_report.User, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	user, ok := c.users[uid]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (c *Client) CreateUser(ctx context.Context, user *nhd_report.User) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if user.UserId == "" {
		return fmt.Errorf("user id cannot be empty")
	}
	c.users[user.UserId] = user
	return nil
}

// --- Customer Methods ---

func (c *Client) GetCustomers(ctx context.Context) ([]*nhd_report.Customer, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	customers := make([]*nhd_report.Customer, 0, len(c.customers))
	for _, customer := range c.customers {
		customers = append(customers, customer)
	}
	return customers, nil
}

func (c *Client) CreateCustomer(ctx context.Context, customer *nhd_report.Customer) (*firestore.DocumentRef, *firestore.WriteResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	newID := uuid.New().String()
	customer.CustomerId = newID
	c.customers[newID] = customer
	return &firestore.DocumentRef{ID: newID}, nil, nil
}

// --- Report Run Methods ---

func (c *Client) CreateReportRun(ctx context.Context, reportRun *nhd_report.ReportRun) (*firestore.DocumentRef, *firestore.WriteResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	newID := uuid.New().String()
	reportRun.ReportRunId = newID
	c.reports[newID] = reportRun
	return &firestore.DocumentRef{ID: newID}, nil, nil
}

func (c *Client) GetReportRuns(ctx context.Context, paymentStatusFilter string) ([]*nhd_report.ReportRun, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	reports := make([]*nhd_report.ReportRun, 0, len(c.reports))
	for _, report := range c.reports {
		if paymentStatusFilter != "" {
			if report.PaymentDetails != nil && report.PaymentDetails.Status.String() == paymentStatusFilter {
				reports = append(reports, report)
			}
		} else {
			reports = append(reports, report)
		}
	}
	return reports, nil
}

func (c *Client) UpdateReportCost(ctx context.Context, reportRunID string, newCost *nhd_report.ReportRun_ReportCost) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	report, ok := c.reports[reportRunID]
	if !ok {
		return fmt.Errorf("report not found")
	}
	report.CostHistory = append(report.CostHistory, newCost)
	return nil
}

func (c *Client) RecordReportPayment(ctx context.Context, reportRunID string, payment *nhd_report.ReportRun_Payment) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	report, ok := c.reports[reportRunID]
	if !ok {
		return fmt.Errorf("report not found")
	}
	report.PaymentDetails = payment
	return nil
}

func (c *Client) GetPaidReportsSummary(ctx context.Context) (*interfaces.FinancialsSummary, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	summary := &interfaces.FinancialsSummary{}
	var totalRevenue float64

	for _, report := range c.reports {
		if report.PaymentDetails != nil && report.PaymentDetails.Status == nhd_report.ReportRun_Payment_PAID {
			totalRevenue += report.PaymentDetails.AmountPaid
			paidReport := interfaces.PaidReportInfo{
				CustomerName:    "Customer " + report.CustomerId, // Placeholder
				PropertyAddress: "Address for " + report.PropertyAddressId, // Placeholder
				AmountPaid:      report.PaymentDetails.AmountPaid,
				PaidAt:          report.PaymentDetails.PaidAt.AsTime().Format("2006-01-02"),
			}
			summary.PaidReports = append(summary.PaidReports, paidReport)
		}
	}
	summary.TotalRevenue = totalRevenue
	return summary, nil
}
