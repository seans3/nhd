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
