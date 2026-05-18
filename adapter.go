// Package socialgoogle is the Lazuli @plugin/social-google adapter.
// Validates Google Sign-In ID tokens via google.golang.org/api/idtoken
// (JWKS fetch + RSA signature + audience + expiry — all upstream).
//
// This is NOT the authz-code-exchange flow; that lives in
// lazuli.dev/runtime/lazuli/auth/oauth_google.go. ID-token validation
// is the leg the client takes after Google's native SDK hands it a JWT.
//
// Configuration: GOOGLE_CLIENT_ID (the web client ID used by the Expo
// or web app). Missing at init → ValidateIDToken returns ErrUnconfigured.
package socialgoogle

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"google.golang.org/api/idtoken"

	"lazuli.dev/runtime/lazuli"
)

const AdapterRef = "@plugin/social-google"

var (
	ErrUnconfigured     = errors.New("social-google: GOOGLE_CLIENT_ID not set")
	ErrTokenExpired     = errors.New("social-google: id token expired")
	ErrAudienceMismatch = errors.New("social-google: audience does not match GOOGLE_CLIENT_ID")
	ErrSignatureInvalid = errors.New("social-google: id token signature invalid")
	ErrTokenInvalid     = errors.New("social-google: id token invalid")
)

// Provider is the Lazuli auth/social contract. Defined here until
// lazuli.dev/runtime/lazuli/auth/social lands the framework type.
type Provider interface {
	ValidateIDToken(ctx context.Context, idToken string) (*UserClaims, error)
}

// UserClaims is the canonical post-validation payload. Fields map
// directly to Google's ID-token claims.
type UserClaims struct {
	Subject       string
	Email         string
	EmailVerified bool
	Name          string
	Picture       string
	Issuer        string
}

type Adapter struct {
	clientID string
	err      error
}

var _ Provider = (*Adapter)(nil)

func init() {
	lazuli.RegisterAdapter(AdapterRef, newAdapter())
}

func newAdapter() *Adapter {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	if clientID == "" {
		return &Adapter{err: ErrUnconfigured}
	}
	return &Adapter{clientID: clientID}
}

func (a *Adapter) ValidateIDToken(ctx context.Context, token string) (*UserClaims, error) {
	if a.err != nil {
		return nil, a.err
	}
	payload, err := idtoken.Validate(ctx, token, a.clientID)
	if err != nil {
		return nil, classifyError(err)
	}
	return &UserClaims{
		Subject:       payload.Subject,
		Issuer:        payload.Issuer,
		Email:         stringClaim(payload.Claims, "email"),
		EmailVerified: boolClaim(payload.Claims, "email_verified"),
		Name:          stringClaim(payload.Claims, "name"),
		Picture:       stringClaim(payload.Claims, "picture"),
	}, nil
}

// classifyError maps idtoken's freeform error strings to typed
// sentinels so callers can branch with errors.Is.
func classifyError(err error) error {
	if err == nil {
		return nil
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "expired"):
		return fmt.Errorf("%w: %v", ErrTokenExpired, err)
	case strings.Contains(msg, "audience"), strings.Contains(msg, "aud "):
		return fmt.Errorf("%w: %v", ErrAudienceMismatch, err)
	case strings.Contains(msg, "signature"), strings.Contains(msg, "verify"):
		return fmt.Errorf("%w: %v", ErrSignatureInvalid, err)
	default:
		return fmt.Errorf("%w: %v", ErrTokenInvalid, err)
	}
}

func stringClaim(claims map[string]interface{}, key string) string {
	v, _ := claims[key].(string)
	return v
}

func boolClaim(claims map[string]interface{}, key string) bool {
	v, _ := claims[key].(bool)
	return v
}
