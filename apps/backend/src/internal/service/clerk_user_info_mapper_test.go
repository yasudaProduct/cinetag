package service

import (
	"testing"
)

func TestNewClerkUserInfoFromWebhook(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		id        string
		email     string
		firstName string
		lastName  string
		imageURL  *string
		wantErr   bool
		wantInfo  ClerkUserInfo
	}{
		{
			name:      "valid input with all fields",
			id:        "user_123",
			email:     "test@example.com",
			firstName: "John",
			lastName:  "Doe",
			imageURL:  strPtr("https://example.com/avatar.png"),
			wantErr:   false,
			wantInfo: ClerkUserInfo{
				ID:        "user_123",
				Email:     "test@example.com",
				FirstName: "John",
				LastName:  "Doe",
				AvatarURL: strPtr("https://example.com/avatar.png"),
			},
		},
		{
			name:      "valid input without optional fields",
			id:        "user_456",
			email:     "minimal@example.com",
			firstName: "",
			lastName:  "",
			imageURL:  nil,
			wantErr:   false,
			wantInfo: ClerkUserInfo{
				ID:        "user_456",
				Email:     "minimal@example.com",
				FirstName: "",
				LastName:  "",
				AvatarURL: nil,
			},
		},
		{
			name:      "valid input with whitespace trimmed",
			id:        "  user_789  ",
			email:     "  whitespace@example.com  ",
			firstName: "  Jane  ",
			lastName:  "  Smith  ",
			imageURL:  strPtr("  https://example.com/photo.png  "),
			wantErr:   false,
			wantInfo: ClerkUserInfo{
				ID:        "user_789",
				Email:     "whitespace@example.com",
				FirstName: "Jane",
				LastName:  "Smith",
				AvatarURL: strPtr("https://example.com/photo.png"),
			},
		},
		{
			name:      "empty id returns error",
			id:        "",
			email:     "test@example.com",
			firstName: "John",
			lastName:  "Doe",
			imageURL:  nil,
			wantErr:   true,
		},
		{
			name:      "whitespace only id returns error",
			id:        "   ",
			email:     "test@example.com",
			firstName: "John",
			lastName:  "Doe",
			imageURL:  nil,
			wantErr:   true,
		},
		{
			name:      "empty email returns error",
			id:        "user_123",
			email:     "",
			firstName: "John",
			lastName:  "Doe",
			imageURL:  nil,
			wantErr:   true,
		},
		{
			name:      "whitespace only email returns error",
			id:        "user_123",
			email:     "   ",
			firstName: "John",
			lastName:  "Doe",
			imageURL:  nil,
			wantErr:   true,
		},
		{
			name:      "both id and email empty returns error",
			id:        "",
			email:     "",
			firstName: "",
			lastName:  "",
			imageURL:  nil,
			wantErr:   true,
		},
		{
			name:      "empty imageURL pointer becomes nil",
			id:        "user_123",
			email:     "test@example.com",
			firstName: "",
			lastName:  "",
			imageURL:  strPtr(""),
			wantErr:   false,
			wantInfo: ClerkUserInfo{
				ID:        "user_123",
				Email:     "test@example.com",
				FirstName: "",
				LastName:  "",
				AvatarURL: nil,
			},
		},
		{
			name:      "whitespace only imageURL becomes nil",
			id:        "user_123",
			email:     "test@example.com",
			firstName: "",
			lastName:  "",
			imageURL:  strPtr("   "),
			wantErr:   false,
			wantInfo: ClerkUserInfo{
				ID:        "user_123",
				Email:     "test@example.com",
				FirstName: "",
				LastName:  "",
				AvatarURL: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewClerkUserInfoFromWebhook(tt.id, tt.email, tt.firstName, tt.lastName, tt.imageURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClerkUserInfoFromWebhook() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err != ErrClerkUserInfoInvalid {
					t.Errorf("NewClerkUserInfoFromWebhook() error = %v, want ErrClerkUserInfoInvalid", err)
				}
				return
			}
			if got.ID != tt.wantInfo.ID {
				t.Errorf("ID = %v, want %v", got.ID, tt.wantInfo.ID)
			}
			if got.Email != tt.wantInfo.Email {
				t.Errorf("Email = %v, want %v", got.Email, tt.wantInfo.Email)
			}
			if got.FirstName != tt.wantInfo.FirstName {
				t.Errorf("FirstName = %v, want %v", got.FirstName, tt.wantInfo.FirstName)
			}
			if got.LastName != tt.wantInfo.LastName {
				t.Errorf("LastName = %v, want %v", got.LastName, tt.wantInfo.LastName)
			}
			if !strPtrEqual(got.AvatarURL, tt.wantInfo.AvatarURL) {
				t.Errorf("AvatarURL = %v, want %v", strPtrVal(got.AvatarURL), strPtrVal(tt.wantInfo.AvatarURL))
			}
		})
	}
}

func TestNewClerkUserInfoFromJWTClaims(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		claims   map[string]any
		wantErr  bool
		wantInfo ClerkUserInfo
	}{
		{
			name: "valid claims with all fields",
			claims: map[string]any{
				"sub":        "user_abc",
				"email":      "jwt@example.com",
				"first_name": "Alice",
				"last_name":  "Wonderland",
				"image_url":  "https://example.com/alice.png",
			},
			wantErr: false,
			wantInfo: ClerkUserInfo{
				ID:        "user_abc",
				Email:     "jwt@example.com",
				FirstName: "Alice",
				LastName:  "Wonderland",
				AvatarURL: strPtr("https://example.com/alice.png"),
			},
		},
		{
			name: "valid claims with minimal fields",
			claims: map[string]any{
				"sub":   "user_xyz",
				"email": "minimal@example.com",
			},
			wantErr: false,
			wantInfo: ClerkUserInfo{
				ID:        "user_xyz",
				Email:     "minimal@example.com",
				FirstName: "",
				LastName:  "",
				AvatarURL: nil,
			},
		},
		{
			name: "claims with whitespace trimmed",
			claims: map[string]any{
				"sub":        "  user_trim  ",
				"email":      "  trim@example.com  ",
				"first_name": "  Bob  ",
				"last_name":  "  Builder  ",
				"image_url":  "  https://example.com/bob.png  ",
			},
			wantErr: false,
			wantInfo: ClerkUserInfo{
				ID:        "user_trim",
				Email:     "trim@example.com",
				FirstName: "Bob",
				LastName:  "Builder",
				AvatarURL: strPtr("https://example.com/bob.png"),
			},
		},
		{
			name: "missing sub returns error",
			claims: map[string]any{
				"email": "nosub@example.com",
			},
			wantErr: true,
		},
		{
			name: "empty sub returns error",
			claims: map[string]any{
				"sub":   "",
				"email": "empty@example.com",
			},
			wantErr: true,
		},
		{
			name: "whitespace only sub returns error",
			claims: map[string]any{
				"sub":   "   ",
				"email": "ws@example.com",
			},
			wantErr: true,
		},
		{
			name: "missing email returns error",
			claims: map[string]any{
				"sub": "user_noemail",
			},
			wantErr: true,
		},
		{
			name: "empty email returns error",
			claims: map[string]any{
				"sub":   "user_123",
				"email": "",
			},
			wantErr: true,
		},
		{
			name: "whitespace only email returns error",
			claims: map[string]any{
				"sub":   "user_123",
				"email": "   ",
			},
			wantErr: true,
		},
		{
			name:    "empty claims returns error",
			claims:  map[string]any{},
			wantErr: true,
		},
		{
			name:    "nil claims returns error",
			claims:  nil,
			wantErr: true,
		},
		{
			name: "non-string sub returns error",
			claims: map[string]any{
				"sub":   123,
				"email": "test@example.com",
			},
			wantErr: true,
		},
		{
			name: "non-string email returns error",
			claims: map[string]any{
				"sub":   "user_123",
				"email": 123,
			},
			wantErr: true,
		},
		{
			name: "empty image_url becomes nil",
			claims: map[string]any{
				"sub":       "user_123",
				"email":     "test@example.com",
				"image_url": "",
			},
			wantErr: false,
			wantInfo: ClerkUserInfo{
				ID:        "user_123",
				Email:     "test@example.com",
				FirstName: "",
				LastName:  "",
				AvatarURL: nil,
			},
		},
		{
			name: "whitespace only image_url becomes nil",
			claims: map[string]any{
				"sub":       "user_123",
				"email":     "test@example.com",
				"image_url": "   ",
			},
			wantErr: false,
			wantInfo: ClerkUserInfo{
				ID:        "user_123",
				Email:     "test@example.com",
				FirstName: "",
				LastName:  "",
				AvatarURL: nil,
			},
		},
		{
			name: "non-string first_name is ignored",
			claims: map[string]any{
				"sub":        "user_123",
				"email":      "test@example.com",
				"first_name": 123,
			},
			wantErr: false,
			wantInfo: ClerkUserInfo{
				ID:        "user_123",
				Email:     "test@example.com",
				FirstName: "",
				LastName:  "",
				AvatarURL: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewClerkUserInfoFromJWTClaims(tt.claims)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClerkUserInfoFromJWTClaims() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err != ErrClerkUserInfoInvalid {
					t.Errorf("NewClerkUserInfoFromJWTClaims() error = %v, want ErrClerkUserInfoInvalid", err)
				}
				return
			}
			if got.ID != tt.wantInfo.ID {
				t.Errorf("ID = %v, want %v", got.ID, tt.wantInfo.ID)
			}
			if got.Email != tt.wantInfo.Email {
				t.Errorf("Email = %v, want %v", got.Email, tt.wantInfo.Email)
			}
			if got.FirstName != tt.wantInfo.FirstName {
				t.Errorf("FirstName = %v, want %v", got.FirstName, tt.wantInfo.FirstName)
			}
			if got.LastName != tt.wantInfo.LastName {
				t.Errorf("LastName = %v, want %v", got.LastName, tt.wantInfo.LastName)
			}
			if !strPtrEqual(got.AvatarURL, tt.wantInfo.AvatarURL) {
				t.Errorf("AvatarURL = %v, want %v", strPtrVal(got.AvatarURL), strPtrVal(tt.wantInfo.AvatarURL))
			}
		})
	}
}

// Helper functions for tests

func strPtr(s string) *string {
	return &s
}

func strPtrEqual(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func strPtrVal(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}
