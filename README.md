# lazuli-plugin-social-google

Lazuli `@plugin/social-google` adapter. Validates Google Sign-In **ID
tokens** server-side via [google.golang.org/api/idtoken][idtoken].

## Status

- **Go server adapter** — wraps `google.golang.org/api/idtoken` for
  JWKS-based ID-token validation. Implements the `Provider` contract
  defined locally (until `lazuli.dev/runtime/lazuli/auth/social.Provider`
  lands in framework, per hostpoint-complete-roadmap §2.5).
  Auto-registers via `init()` against `@plugin/social-google`.
- **TS web / mobile client** — none. ID-token capture happens client-side
  via Google's native SDK (Expo `expo-auth-session` for mobile, GIS
  JS one-tap for web); the JWT is posted to the server, which calls
  this adapter to verify it. No TS face needed.

## Two Google auth flows — which is this?

Google's OAuth surface has two server-relevant flows:

| Flow | Who initiates | Server's job | This plugin |
|---|---|---|---|
| Authz code → token exchange | Web redirect with `?code=` | Exchange code, fetch userinfo | ❌ Lives in `lazuli.dev/runtime/lazuli/auth/oauth_google.go` |
| ID-token validation | Native SDK (mobile / one-tap) hands a JWT | Verify JWT signature + aud + exp | ✅ This plugin |

If you only need browser-redirect login, the in-tree
`runtime/go/lazuli/auth/oauth_google.go` already handles it. This plugin
is for the Sign-In-with-Google native flow Hostpoint needs for Play Store
submission and one-tap web sign-in.

## Usage

In `lazurite.toml`:

```toml
[plugins]
"@plugin/social-google" = { module = "github.com/lazuli-lang/lazuli-plugin-social-google", version = "v0.1.0" }
```

In `.lzi` (once `auth/social.Provider` framework contract lands):

```lzi
registry
  integrations
    google_signin: SocialProvider
      adapter @plugin/social-google
```

Set `GOOGLE_CLIENT_ID` at runtime to the **web client ID** (the same
value the Expo or web app uses when initiating the Google Sign-In SDK).
The adapter reads it once at process startup; missing key →
`ValidateIDToken` returns `ErrUnconfigured`.

### Go API

```go
import sg "github.com/lazuli-lang/lazuli-plugin-social-google"

provider := lazuli.LookupAdapter(sg.AdapterRef).(sg.Provider)
claims, err := provider.ValidateIDToken(ctx, idTokenFromClient)
if err != nil {
    switch {
    case errors.Is(err, sg.ErrTokenExpired):     // re-auth
    case errors.Is(err, sg.ErrAudienceMismatch): // wrong client id
    case errors.Is(err, sg.ErrSignatureInvalid): // forged or tampered
    case errors.Is(err, sg.ErrTokenInvalid):     // generic invalid
    default:                                     // network / 5xx
    }
}
// claims.Subject = stable Google user id ("sub")
// claims.Email + claims.EmailVerified
// claims.Name + claims.Picture
// claims.Issuer = "https://accounts.google.com"
```

## What `idtoken.Validate` actually does

The wire-thin call `idtoken.Validate(ctx, token, clientID)`:

1. Fetches Google's JWKS from `https://www.googleapis.com/oauth2/v3/certs`
   (cached internally with HTTP cache-control respect).
2. Parses the JWT header, finds the matching `kid`.
3. Verifies the RSA signature.
4. Checks `aud == clientID`.
5. Checks `exp` is in the future.
6. Returns a `*Payload` with `Subject`, `Issuer`, `Audience`, `Expires`,
   and a `Claims` map for the rest (`email`, `email_verified`, `name`,
   `picture`, …).

This adapter does NONE of that itself — it only classifies the upstream
error string into a typed sentinel (`ErrTokenExpired`,
`ErrAudienceMismatch`, `ErrSignatureInvalid`, `ErrTokenInvalid`) and
shapes the payload into the canonical `UserClaims` struct.

## Wire-thin discipline

~116 LOC of import + call. The plugin does NOT implement JWT parsing,
RSA verification, or JWKS caching — `idtoken` handles all of that. If
this file grows past ~120 LOC, audit for sneak-in business logic.

## License

MIT — see [LICENSE](LICENSE).

[idtoken]: https://pkg.go.dev/google.golang.org/api/idtoken
