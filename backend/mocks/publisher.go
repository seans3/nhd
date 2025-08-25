package mocks

import (
	"context"

	"github.com/seans3/nhd/backend/interfaces"
	"github.com/stretchr/testify/mock"
)

// Statically assert that our mock satisfies the interface.
var _ interfaces.Publisher = (*MockPublisherClient)(nil)

// MockPublisherClient is a mock implementation of the Publisher interface.
type MockPublisherClient struct {
	mock.Mock
}

func (m *MockPublisherClient) Publish(ctx context.Context, topicID string, data []byte) (string, error) {
	args := m.Called(ctx, topicID, data)
	return args.String(0), args.Error(1)
}
