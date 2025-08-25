package datastore

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
	"github.com/seans3/nhd/proto/gen/go;nhd_report"
	"google.golang.org/api/iterator"
)

type Client struct {
	*firestore.Client
}

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

func NewClient(ctx context.Context, projectID string) (*Client, error) {
	fsClient, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	log.Println("Firestore client initialized")
	return &Client{fsClient}, nil
}

func (c *Client) CreateCustomer(ctx context.Context, customer *nhd_report.Customer) (*firestore.DocumentRef, *firestore.WriteResult, error) {
	return c.Collection("customers").Add(ctx, customer)
}

func (c *Client) GetCustomers(ctx context.Context) ([]*nhd_report.Customer, error) {
	var customers []*nhd_report.Customer
	iter := c.Collection("customers").Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var customer nhd_report.Customer
		if err := doc.DataTo(&customer); err != nil {
			return nil, err
		}
		customers = append(customers, &customer)
	}
	return customers, nil
}

func (c *Client) CreateReportRun(ctx context.Context, reportRun *nhd_report.ReportRun) (*firestore.DocumentRef, *firestore.WriteResult, error) {
	return c.Collection("report_runs").Add(ctx, reportRun)
}

func (c *Client) GetReportRuns(ctx context.Context, paymentStatusFilter string) ([]*nhd_report.ReportRun, error) {
	var reportRuns []*nhd_report.ReportRun
	
	query := c.Collection("report_runs").Query
	
	// Apply filter if one is provided
	if paymentStatusFilter != "" {
		// Note: This requires a composite index in Firestore on `payment_details.status`
		query = query.Where("payment_details.status", "==", paymentStatusFilter)
	}

	iter := query.Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var reportRun nhd_report.ReportRun
		if err := doc.DataTo(&reportRun); err != nil {
			log.Printf("Failed to unmarshal report run: %v", err)
			continue
		}
		reportRuns = append(reportRuns, &reportRun)
	}
	return reportRuns, nil
}

func (c *Client) UpdateReportCost(ctx context.Context, reportRunID string, newCost *nhd_report.ReportRun_ReportCost) error {
	reportRunRef := c.Collection("report_runs").Doc(reportRunID)

	return c.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		return tx.Update(reportRunRef, []firestore.Update{
			{Path: "cost_history", Value: firestore.ArrayUnion(newCost)},
		})
	})
}

func (c *Client) RecordReportPayment(ctx context.Context, reportRunID string, payment *nhd_report.ReportRun_Payment) error {
	reportRunRef := c.Collection("report_runs").Doc(reportRunID)
	return c.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		return tx.Update(reportRunRef, []firestore.Update{
			{Path: "payment_details", Value: payment},
		})
	})
}

func (c *Client) GetPaidReportsSummary(ctx context.Context) (*FinancialsSummary, error) {
	summary := &FinancialsSummary{}
	var totalRevenue float64

	// Query for paid reports
	iter := c.Collection("report_runs").Where("payment_details.status", "==", nhd_report.ReportRun_Payment_PAID).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var reportRun nhd_report.ReportRun
		if err := doc.DataTo(&reportRun); err != nil {
			// Log the error but continue if possible
			log.Printf("Failed to unmarshal report run: %v", err)
			continue
		}

		if reportRun.PaymentDetails != nil {
			totalRevenue += reportRun.PaymentDetails.AmountPaid

			// For simplicity, we are not fetching customer and address details in this example.
			// In a real application, you would fetch the customer and property address documents
			// using the IDs from the reportRun to get the full name and address string.
			paidReport := PaidReportInfo{
				CustomerName:      "Customer " + reportRun.CustomerId, // Placeholder
				PropertyAddress:   "Address for " + reportRun.PropertyAddressId, // Placeholder
				AmountPaid:        reportRun.PaymentDetails.AmountPaid,
				PaidAt:            reportRun.PaymentDetails.PaidAt.AsTime().Format("2006-01-02"),
			}
			summary.PaidReports = append(summary.PaidReports, paidReport)
		}
	}

	summary.TotalRevenue = totalRevenue
	return summary, nil
}
