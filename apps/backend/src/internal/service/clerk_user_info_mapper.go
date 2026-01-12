package service

import (
	"errors"
	"strings"
)

var (
	ErrClerkUserInfoInvalid = errors.New("invalid clerk user info")
)

// Webhook の生フィールドから ClerkUserInfo を構築します。
// Email は必須です（空の場合は error を返します）。
func NewClerkUserInfoFromWebhook(id, email, firstName, lastName string, imageURL *string) (ClerkUserInfo, error) {
	id = strings.TrimSpace(id)
	email = strings.TrimSpace(email)
	firstName = strings.TrimSpace(firstName)
	lastName = strings.TrimSpace(lastName)

	if id == "" || email == "" {
		return ClerkUserInfo{}, ErrClerkUserInfoInvalid
	}

	return ClerkUserInfo{
		ID:        id,
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		AvatarURL: trimStringPtr(imageURL),
	}, nil
}

// NewClerkUserInfoFromJWTClaims は JWT claims から ClerkUserInfo を構築します。
// - ID: claim["sub"]
// - Email: claim["email"]
// - FirstName: claim["first_name"]
// - LastName: claim["last_name"]
// - AvatarURL: claim["image_url"]
//
// Email は必須です（空の場合は error を返します）。
func NewClerkUserInfoFromJWTClaims(claims map[string]any) (ClerkUserInfo, error) {
	sub := trimStringClaim(claims, "sub")
	email := trimStringClaim(claims, "email")
	firstName := trimStringClaim(claims, "first_name")
	lastName := trimStringClaim(claims, "last_name")
	imageURL := trimStringClaimPtr(claims, "image_url")

	if sub == "" || email == "" {
		return ClerkUserInfo{}, ErrClerkUserInfoInvalid
	}

	return ClerkUserInfo{
		ID:        sub,
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		AvatarURL: imageURL,
	}, nil
}

func trimStringClaim(claims map[string]any, key string) string {
	v, ok := claims[key].(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(v)
}

func trimStringClaimPtr(claims map[string]any, key string) *string {
	s := trimStringClaim(claims, key)
	if s == "" {
		return nil
	}
	return &s
}

func trimStringPtr(s *string) *string {
	if s == nil {
		return nil
	}
	v := strings.TrimSpace(*s)
	if v == "" {
		return nil
	}
	return &v
}
