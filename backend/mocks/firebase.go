package mocks

import (
	"context"

	"firebase.google.com/go/v4/auth"
	"github.com/seans3/nhd/backend/interfaces"
	"github.com/stretchr/testify/mock"
)

// Statically assert that our mock satisfies the interface.
var _ interfaces.FirebaseAuth = (*MockFirebaseAuth)(nil)

// MockFirebaseAuth is a mock implementation of the FirebaseAuth interface.
type MockFirebaseAuth struct {
	mock.Mock
}

func (m *MockFirebaseAuth) VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error) {
	args := m.Called(ctx, idToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.Token), args.Error(1)
}
