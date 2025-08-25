package interfaces

import (
	"context"
)

// Publisher is an interface for the pub/sub client to allow for mocking.
type Publisher interface {
	Publish(ctx context.Context, topicID string, data []byte) (string, error)
}
