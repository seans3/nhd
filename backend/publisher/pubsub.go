package publisher

import (
	"context"
	"log"

	"cloud.google.com/go/pubsub"
	"github.com/seans3/nhd/backend/interfaces"
)

// Statically assert that our client satisfies the interface.
var _ interfaces.Publisher = (*Client)(nil)

type Client struct {
	*pubsub.Client
}

func NewClient(ctx context.Context, projectID string) (*Client, error) {
	psClient, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	log.Println("Pub/Sub client initialized")
	return &Client{psClient}, nil
}

func (c *Client) Publish(ctx context.Context, topicID string, data []byte) (string, error) {
	topic := c.Topic(topicID)
	result := topic.Publish(ctx, &pubsub.Message{
		Data: data,
	})
	return result.Get(ctx)
}
