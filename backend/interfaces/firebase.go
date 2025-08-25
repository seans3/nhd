package interfaces

import (
	"context"

	"firebase.google.com/go/v4/auth"
)

// FirebaseAuth is an interface for the Firebase Auth client to allow for mocking.
type FirebaseAuth interface {
	VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error)
	// Add other methods you use from the auth.Client here
}
