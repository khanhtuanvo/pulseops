package graph

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/tuankhanhvo/pulseops/internal/incidents"
	"github.com/tuankhanhvo/pulseops/internal/oncall"
	"github.com/tuankhanhvo/pulseops/pkg/auth"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.uber.org/zap"
)

// GraphQL error codes surfaced to clients via the `extensions.code` field.
const (
	codeUnauthenticated = "UNAUTHENTICATED"
	codeForbidden       = "FORBIDDEN"
	codeBadUserInput    = "BAD_USER_INPUT"
	codeInternal        = "INTERNAL_ERROR"
)

// ErrInvalidInput marks client input-validation failures. Wrap it with a
// human-readable, safe message; that message is surfaced to the client.
var ErrInvalidInput = errors.New("invalid input")

func gqlError(code, message string) *gqlerror.Error {
	return &gqlerror.Error{
		Message:    message,
		Extensions: map[string]interface{}{"code": code},
	}
}

// mapError converts a resolver-origin error into a safe GraphQL error. Unknown
// errors collapse to a generic INTERNAL_ERROR so MongoDB/Go internals never
// reach the client.
func mapError(err error) *gqlerror.Error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, auth.ErrUnauthenticated):
		return gqlError(codeUnauthenticated, "authentication required")
	case errors.Is(err, auth.ErrUnauthorized):
		return gqlError(codeForbidden, "you do not have permission to perform this action")
	case errors.Is(err, ErrInvalidInput):
		return gqlError(codeBadUserInput, err.Error())
	case errors.Is(err, incidents.ErrIncidentNotFound):
		return gqlError(codeBadUserInput, "incident not found")
	case errors.Is(err, incidents.ErrInvalidStateTransition):
		return gqlError(codeBadUserInput, "invalid incident state transition")
	case errors.Is(err, oncall.ErrEmptyRotation),
		errors.Is(err, oncall.ErrInvalidInterval),
		errors.Is(err, oncall.ErrOverrideTooLong),
		errors.Is(err, oncall.ErrInvalidOverrideTime):
		return gqlError(codeBadUserInput, err.Error())
	default:
		return gqlError(codeInternal, "internal server error")
	}
}

// NewErrorPresenter returns a gqlgen error presenter that sanitises every
// resolver error and logs the underlying cause of internal errors server-side.
// gqlgen-native errors (query parsing, validation, argument coercion) are
// already client-safe and pass through unchanged.
func NewErrorPresenter(logger *zap.Logger) graphql.ErrorPresenterFunc {
	if logger == nil {
		logger = zap.NewNop()
	}

	return func(ctx context.Context, e error) *gqlerror.Error {
		var native *gqlerror.Error
		if errors.As(e, &native) {
			return native
		}

		mapped := mapError(e)
		if mapped == nil {
			return nil
		}
		if mapped.Extensions["code"] == codeInternal {
			logger.Error("graphql internal error", zap.Error(e), zap.Any("path", graphql.GetPath(ctx)))
		}
		mapped.Path = graphql.GetPath(ctx)
		return mapped
	}
}

// NewRecoverFunc converts a resolver panic into a sanitised internal error
// instead of leaking a stack trace to the client.
func NewRecoverFunc(logger *zap.Logger) graphql.RecoverFunc {
	if logger == nil {
		logger = zap.NewNop()
	}

	return func(ctx context.Context, err interface{}) error {
		logger.Error("graphql panic recovered", zap.Any("panic", err), zap.Any("path", graphql.GetPath(ctx)))
		return gqlError(codeInternal, "internal server error")
	}
}

// requireNonEmpty returns a BAD_USER_INPUT error when a required string is blank.
func requireNonEmpty(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%w: %s is required", ErrInvalidInput, field)
	}

	return nil
}

// validateTimeRange rejects an inverted time window.
func validateTimeRange(from, to *time.Time) error {
	if from != nil && to != nil && from.After(*to) {
		return fmt.Errorf("%w: 'from' must be before 'to'", ErrInvalidInput)
	}

	return nil
}

// validateRotation rejects an empty on-call rotation.
func validateRotation(rotation []string) error {
	if len(rotation) == 0 {
		return fmt.Errorf("%w: rotation must include at least one member", ErrInvalidInput)
	}

	return nil
}
