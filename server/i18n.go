package server

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/invopop/ctxi18n"
	"github.com/invopop/ctxi18n/i18n"
)

// This will panic if APP_FALLBACKLANG has not been set or if no bundle could be found for the
// fallback language.
func (s *Server[state]) WithI18n(fs fs.FS) *Server[state] {
	lang := i18n.Code(s.cfg.App.FallbackLang)
	if len(lang) == 0 {
		panic("You need to set a fallbacklang in the project config before calling WithI18n!")
	}
	if err := ctxi18n.LoadWithDefault(fs, lang); err != nil {
		panic(err)
	}
	ctxi18n.DefaultLocale = lang

	return s
}

// DetectLanguage is middleware that automatically tries to detect a user's language by looking at
// the request headers. If the detected language is not found, it will fallback to the configure
// fallback language.
// This will return an error only if no bundle could be found for the configured fallback language.
func DetectLanguage[state any](apollo *Apollo, _ state) (context.Context, error) {
	ctx := apollo.Context()

	// Skip this middleware if no language was set
	if len(apollo.Cfg.App.FallbackLang) == 0 {
		return ctx, nil
	}

	// NOTE: We need to reset this on every request because you shouldn't use singletons for
	// things like this but I'm also too lazy to fix ctxi18n right now - robin
	ctxi18n.DefaultLocale = i18n.Code(apollo.Cfg.App.FallbackLang)

	lang := apollo.Request.Header.Get("Accept-Language")

	// This is here for logging, ctxi18n should already use the fallback if lang is empty
	if len(lang) == 0 {
		lang = apollo.Cfg.App.FallbackLang
	}

	apollo.LogField("lang", slog.StringValue(lang))

	ctx, err := ctxi18n.WithLocale(ctx, lang)
	if errors.Is(err, ctxi18n.ErrMissingLocale) {
		err = fmt.Errorf(
			"No language bundle found for the fallback language %q: %w",
			apollo.Cfg.App.FallbackLang,
			err,
		)
	}

	return ctx, err
}
