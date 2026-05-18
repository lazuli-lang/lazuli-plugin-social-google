package socialgoogle

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

func TestAdapterSatisfiesProvider(t *testing.T) {
	var _ Provider = (*Adapter)(nil)
}

func TestAdapterUnconfiguredReturnsError(t *testing.T) {
	a := &Adapter{err: ErrUnconfigured}
	_, err := a.ValidateIDToken(context.Background(), "any-token")
	if !errors.Is(err, ErrUnconfigured) {
		t.Errorf("ValidateIDToken: want ErrUnconfigured, got %v", err)
	}
}

func TestNewAdapterMissingClientID(t *testing.T) {
	t.Setenv("GOOGLE_CLIENT_ID", "")
	a := newAdapter()
	if !errors.Is(a.err, ErrUnconfigured) {
		t.Errorf("newAdapter without env: want ErrUnconfigured, got %v", a.err)
	}
}

func TestNewAdapterWithClientID(t *testing.T) {
	t.Setenv("GOOGLE_CLIENT_ID", "fake-client.apps.googleusercontent.com")
	a := newAdapter()
	if a.err != nil {
		t.Errorf("newAdapter with env: want nil err, got %v", a.err)
	}
	if a.clientID != "fake-client.apps.googleusercontent.com" {
		t.Errorf("newAdapter: clientID got %q", a.clientID)
	}
}

func TestClassifyError(t *testing.T) {
	cases := []struct {
		name string
		raw  error
		want error
	}{
		{"nil", nil, nil},
		{"expired", fmt.Errorf("idtoken: token expired"), ErrTokenExpired},
		{"audience", fmt.Errorf("idtoken: audience provided does not match aud claim"), ErrAudienceMismatch},
		{"aud-short", fmt.Errorf("idtoken: aud value mismatch"), ErrAudienceMismatch},
		{"signature", fmt.Errorf("idtoken: invalid signature"), ErrSignatureInvalid},
		{"verify", fmt.Errorf("idtoken: could not verify token"), ErrSignatureInvalid},
		{"other", fmt.Errorf("idtoken: malformed jwt"), ErrTokenInvalid},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := classifyError(c.raw)
			if c.want == nil {
				if got != nil {
					t.Errorf("want nil, got %v", got)
				}
				return
			}
			if !errors.Is(got, c.want) {
				t.Errorf("classifyError(%v): want errors.Is %v, got %v", c.raw, c.want, got)
			}
		})
	}
}

func TestStringClaim(t *testing.T) {
	claims := map[string]interface{}{"name": "Ada", "n": 42}
	if got := stringClaim(claims, "name"); got != "Ada" {
		t.Errorf("string hit: got %q", got)
	}
	if got := stringClaim(claims, "n"); got != "" {
		t.Errorf("type mismatch: got %q", got)
	}
	if got := stringClaim(claims, "missing"); got != "" {
		t.Errorf("missing: got %q", got)
	}
}

func TestBoolClaim(t *testing.T) {
	claims := map[string]interface{}{"verified": true, "x": "yes"}
	if !boolClaim(claims, "verified") {
		t.Errorf("bool hit: want true")
	}
	if boolClaim(claims, "x") {
		t.Errorf("type mismatch: want false")
	}
	if boolClaim(claims, "missing") {
		t.Errorf("missing: want false")
	}
}
