# Changelog

All notable changes to `@plugin/social-google` will be documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-05-20

### Added
- Initial release: Google Sign-In ID-token validator wrapping `google.golang.org/api/idtoken`.
- Go server adapter implementing the social `Provider` surface for `auth.SocialProvider` usage (see `manifest.toml`).
- Auto-registers via `init()` against `@plugin/social-google`.

### Vendor SDK pin
- `google.golang.org/api` v0.220.0
