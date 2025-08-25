package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/seans3/nhd/backend/interfaces"
)

type AuthClient struct {
	Firebase interfaces.FirebaseAuth
	DS       interfaces.Datastore
}

// ContextKey is a custom type for context keys to avoid collisions.
type ContextKey string

const UserIDKey ContextKey = "userID"

func (ac *AuthClient) VerifyAuthToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		idToken := tokenParts[1]
		token, err := ac.Firebase.VerifyIDToken(r.Context(), idToken)
		if err != nil {
			log.Printf("Error verifying ID token: %v\n", err)
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Add user ID to the context
		ctx := context.WithValue(r.Context(), UserIDKey, token.UID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (ac *AuthClient) RequireAdmin(next http.Handler) http.Handler {
	return ac.VerifyAuthToken(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid, ok := r.Context().Value(UserIDKey).(string)
		if !ok {
			// This should not happen if VerifyFirebaseToken runs first
			http.Error(w, "User ID not found in context", http.StatusInternalServerError)
			return
		}

		user, err := ac.DS.GetUserByID(r.Context(), uid)
		if err != nil {
			http.Error(w, "Could not retrieve user profile", http.StatusInternalServerError)
			return
		}

		if user.Permissions == nil || !user.Permissions.IsAdmin {
			http.Error(w, "Admin privileges required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	}))
}
