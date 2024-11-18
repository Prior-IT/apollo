package server_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prior-it/apollo/config"
	"github.com/prior-it/apollo/server"
	"github.com/stretchr/testify/assert"
)

type State struct{}

func (s State) Close(_ context.Context) {}

func runTest(cfg *config.Config, testfunc func(ctx context.Context)) {
	s := server.New(State{}, cfg)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(""))

	s.FeatureFlagMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		testfunc(r.Context())
	})).ServeHTTP(recorder, req)
}

func TestFeatureFlagMiddleware(t *testing.T) {
	t.Run("ok: feature flags are followed", func(t *testing.T) {
		t.Parallel()
		runTest(
			&config.Config{
				Features: config.FeaturesConfig{
					Flags: map[string]bool{
						"FEATURE1": true,
						"FEATURE2": false,
					},
				},
			},
			func(ctx context.Context) {
				assert.True(t, server.HasFeature(ctx, "FEATURE1"))
				assert.False(t, server.HasFeature(ctx, "FEATURE2"))
				assert.False(
					t,
					server.HasFeature(ctx, "FEATURE3"),
					"A feature that does not exist should return false",
				)
			},
		)
	})

	t.Run("ok: enable all overrides specific flags", func(t *testing.T) {
		t.Parallel()
		runTest(
			&config.Config{
				Features: config.FeaturesConfig{
					EnableAll: true,
					Flags: map[string]bool{
						"FEATURE":      true,
						"OTHERFEATURE": false,
					},
				},
			},
			func(ctx context.Context) {
				assert.True(t, server.HasFeature(ctx, "FEATURE"))
				assert.True(t, server.HasFeature(ctx, "OTHERFEATURE"))
			},
		)
	})

	t.Run("ok: enable all overrides unspecified flags", func(t *testing.T) {
		t.Parallel()
		runTest(
			&config.Config{
				Features: config.FeaturesConfig{
					EnableAll: true,
					Flags:     map[string]bool{},
				},
			},
			func(ctx context.Context) {
				assert.True(t, server.HasFeature(ctx, "SOMEFEATURE"))
				assert.True(t, server.HasFeature(ctx, "SOMEOTHERFEATURE"))
			},
		)
	})

	t.Run(
		"ok: all feature flags return false if the feature flag middleware was not added",
		func(t *testing.T) {
			t.Parallel()
			cfg := &config.Config{
				Features: config.FeaturesConfig{
					EnableAll: true,
					Flags: map[string]bool{
						"NOFEATURE1": true,
						"NOFEATURE2": true,
					},
				},
			}

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(""))
			s := server.New(State{}, cfg)
			s.ContextMiddleware( // Use context middleware to test in an environment where cfg does exist in the context
				http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
					ctx := r.Context()
					assert.False(t, server.HasFeature(ctx, "NOFEATURE1"))
					assert.False(t, server.HasFeature(ctx, "NOFEATURE2"))
				}),
			).
				ServeHTTP(recorder, req)
		},
	)
}
