package skills

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/everscribe/cli/internal/types"
)

// DefaultManifestsURL is the public CDN that hosts the skill catalog
// and SKILL.md bodies. Mirrors the spec's "Code organization" section:
// the monorepo's internal/ui/manifests/ package serves these files,
// CI publishes them here.
const DefaultManifestsURL = "https://everscribe.io"

const manifestsOverrideEnv = "EVERSCRIBE_MANIFESTS_URL_OVERRIDE"

// resolveManifestsURL honors an undocumented env override so tests
// can point the fetcher at a httptest.Server. Same pattern as the
// authenticated client's EVERSCRIBE_API_URL_OVERRIDE.
func resolveManifestsURL() string {
	if v := os.Getenv(manifestsOverrideEnv); v != "" {
		return v
	}
	return DefaultManifestsURL
}

// Fetcher is a thin HTTP client for the skills CDN. No auth, no state —
// the manifests are public. A small struct (rather than free functions)
// so the base URL and underlying *http.Client are easy to swap in tests.
type Fetcher struct {
	httpClient *http.Client
	baseURL    string
}

// FetcherOption customizes a Fetcher.
type FetcherOption func(*Fetcher)

// WithHTTPClient overrides the underlying *http.Client (test seam).
func WithHTTPClient(h *http.Client) FetcherOption {
	return func(f *Fetcher) { f.httpClient = h }
}

// WithBaseURL overrides the resolved CDN base URL (test seam).
func WithBaseURL(u string) FetcherOption {
	return func(f *Fetcher) { f.baseURL = u }
}

// NewFetcher returns a Fetcher pointed at the production CDN unless
// EVERSCRIBE_MANIFESTS_URL_OVERRIDE is set.
func NewFetcher(opts ...FetcherOption) *Fetcher {
	f := &Fetcher{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    resolveManifestsURL(),
	}
	for _, o := range opts {
		o(f)
	}
	return f
}

// Catalog fetches and decodes the published skills.json. Returns a
// wrapped error on non-2xx status or invalid JSON so the CLI can
// surface a clear message rather than the raw HTTP failure.
func (f *Fetcher) Catalog(ctx context.Context) (*types.SkillCatalog, error) {
	url := f.baseURL + "/manifests/skills.json"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch skills catalog: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fetch skills catalog: HTTP %d", resp.StatusCode)
	}
	var c types.SkillCatalog
	if err := json.NewDecoder(resp.Body).Decode(&c); err != nil {
		return nil, fmt.Errorf("decode skills catalog: %w", err)
	}
	return &c, nil
}

// SkillBody fetches the raw bytes at a skill's `source` URL. We don't
// validate the markdown shape here — Claude Code reads the file on
// next invocation, and surfaces its own error if SKILL.md is malformed.
// Pass the full URL from Skill.Source (it's an absolute URL in the
// catalog, so we don't need to combine it with baseURL).
func (f *Fetcher) SkillBody(ctx context.Context, sourceURL string) ([]byte, error) {
	if sourceURL == "" {
		return nil, errors.New("skill source URL is empty")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch skill body: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fetch skill body %s: HTTP %d", sourceURL, resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read skill body: %w", err)
	}
	return body, nil
}
