package setting

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"slices"
	"strings"
)

var ErrNoSourceProvidedValue = errors.New("no source provided a value")

type Sources []Source

// ResolveSetting takes the value from the first source that carries one, in this order:
// FrontendSource, any other source, the environment, FallbackSource, the setting's default.
// Two sources of the same kind keep the order they were given in. It returns
// ErrNoSourceProvidedValue when every source stayed empty.
func (sources Sources) ResolveSetting[T any](ctx context.Context, setting Setting[T], extraSources ...Source) (value T, err error) {
	value, _, err = sources.ResolveSettingWithSource(ctx, setting, extraSources...)
	return
}

func (sources Sources) ResolveSettingWithSource[T any](ctx context.Context, setting Setting[T], extraSources ...Source) (value T, source Source, err error) {
	var emptySources Sources
	defer func() {
		if err != nil {
			errs := []error{err}
			for _, source := range emptySources {
				switch source.(type) {
				case DefaultSource, FallbackSource:
					// Both derive a value from other state, so no person can set them.
					continue
				}
				if describeSource := source.Describe(setting.EnvKey()); describeSource != "" {
					errs = append(errs, fmt.Errorf("try setting %s", describeSource))
				}
			}
			err = errors.Join(errs...)
		}
	}()

	var frontendSources, otherSources, fallbackSources Sources
	for _, source := range slices.Concat(sources, extraSources) {
		if _, isFrontend := source.(FrontendSource); isFrontend {
			frontendSources = append(frontendSources, source)
		} else if _, isFallback := source.(FallbackSource); isFallback {
			fallbackSources = append(fallbackSources, source)
		} else {
			otherSources = append(otherSources, source)
		}
	}

	sourceCandidates := slices.Concat(
		frontendSources,
		otherSources,
		Sources{setting.Env},
		fallbackSources,
		setting.Default.asSourcesIfPresent(),
	)

	for i, candidate := range sourceCandidates {
		if candidate == nil {
			continue
		}
		description := candidate.Describe(setting.EnvKey())
		var text string
		text, err = candidate.Lookup(ctx, setting.EnvKey())
		if err != nil {
			return value, nil, fmt.Errorf("value from source '%s' could not be looked up: %w", description, err)
		}
		text = strings.TrimSpace(text)
		if text == "" {
			emptySources = append(emptySources, candidate)
			continue
		}
		value, err = setting.Parse(text)
		if err != nil {
			return value, nil, fmt.Errorf("value from source '%s' could not be parsed: %w", description, err)
		}
		if _, fromEnv := candidate.(EnvKey); fromEnv {
			fallbackSources.warnOfOverriddenStoredValues(ctx, setting.EnvKey(), description, text)
		}
		if i >= len(frontendSources) && i < len(frontendSources)+len(otherSources) {
			setting.warnOfOverriddenEnv(ctx, description, text, value)
		}
		return value, candidate, nil
	}
	return value, nil, fmt.Errorf("%w for %s", ErrNoSourceProvidedValue, setting.EnvKey())
}

func (sources Sources) warnOfOverriddenStoredValues(ctx context.Context, key, description, text string) {
	for _, source := range sources {
		if fallback, ok := source.(FallbackSource); ok && fallback.Stored {
			stored, err := fallback.Lookup(ctx, key)
			if stored = strings.TrimSpace(stored); err == nil && stored != "" && stored != text {
				slog.WarnContext(ctx, fmt.Sprintf("%s sets '%s', which overrides '%s' from the %s", description, text, stored, fallback.Describe(key)))
			}
		}
	}
}

func (s Setting[T]) warnOfOverriddenEnv(ctx context.Context, description, text string, value T) {
	envText, err := s.Env.Lookup(ctx, s.EnvKey())
	if envText = strings.TrimSpace(envText); err != nil || envText == "" {
		return
	}
	if envValue, parseErr := s.Parse(envText); parseErr == nil && reflect.DeepEqual(envValue, value) {
		return
	}
	slog.WarnContext(ctx, fmt.Sprintf("%s sets '%s', which overrides '%s' from %s", description, text, envText, s.Env.Describe(s.EnvKey())))
}
