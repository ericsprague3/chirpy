package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeAndValidateJWT(t *testing.T) {
	userId := uuid.New()
	secret := "super-secret-test-key"

	t.Run("Valid JWT", func(t *testing.T) {
		token, err := MakeJWT(userId, secret, time.Hour)
		if err != nil {
			t.Fatalf("expected no error creation JWT, got %v", err)
		}

		parsedId, err := ValidateJWT(token, secret)
		if err != nil {
			t.Fatalf("expected no error validating JWT, got %v", err)
		}

		if parsedId != userId {
			t.Errorf("expected user ID %v, got %v", userId, parsedId)
		}
	})

	t.Run("Expired JWT", func(t *testing.T) {
		token, err := MakeJWT(userId, secret, -time.Minute)
		if err != nil {
			t.Fatalf("expected no error creating JWT, got %v", err)
		}

		_, err = ValidateJWT(token, secret)
		if err == nil {
			t.Fatalf("expected error for expired JWT, got nil")
		}
	})

	t.Run("Wrong Secret", func(t *testing.T) {
		token, err := MakeJWT(userId, secret, time.Hour)
		if err != nil {
			t.Fatalf("expected no error creating JWT, got %v", err)
		}

		_, err = ValidateJWT(token, "wrong-secret-key")
		if err == nil {
			t.Fatalf("expected error for wrong secret, got nil")
		}
	})
}

func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name      string
		headers   http.Header
		wantToken string
		wantErr   bool
	}{
		{
			name:      "Valid Bearer Token",
			headers:   http.Header{"Authorization": []string{"Bearer my-secret-token"}},
			wantToken: "my-secret-token",
			wantErr:   false,
		},
		{
			name:      "Valid lowercase bearer with extra spaces",
			headers:   http.Header{"Authorization": []string{"  bearer   my-secret-token  "}},
			wantToken: "my-secret-token",
			wantErr:   false,
		},
		{
			name:      "Missing Authorization header",
			headers:   http.Header{},
			wantToken: "",
			wantErr:   true,
		},
		{
			name:      "Empty Authorization header",
			headers:   http.Header{"Authorization": []string{""}},
			wantToken: "",
			wantErr:   true,
		},
		{
			name:      "Malformed scheme (Basic instead of Bearer)",
			headers:   http.Header{"Authorization": []string{"Basic my-secret-token"}},
			wantToken: "",
			wantErr:   true,
		},
		{
			name:      "Bearer prefix with missing token",
			headers:   http.Header{"Authorization": []string{"Bearer "}},
			wantToken: "",
			wantErr:   true,
		},
		{
			name:      "Too many parts in header",
			headers:   http.Header{"Authorization": []string{"Bearer token1 token2"}},
			wantToken: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotToken, err := GetBearerToken(tt.headers)

			if (err != nil) != tt.wantErr {
				t.Fatalf("GetBearerToken() error = %v, wantErr %v", err, tt.wantErr)
			}

			if gotToken != tt.wantToken {
				t.Errorf("GetBearerToken() = %q, want %q", gotToken, tt.wantToken)
			}
		})
	}
}
