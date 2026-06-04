package graph

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/tuankhanhvo/pulseops/internal/incidents"
	"github.com/tuankhanhvo/pulseops/internal/oncall"
	"github.com/tuankhanhvo/pulseops/pkg/auth"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func code(err *gqlerror.Error) string {
	if err == nil || err.Extensions == nil {
		return ""
	}
	value, _ := err.Extensions["code"].(string)
	return value
}

func TestMapError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode string
	}{
		{"nil", nil, ""},
		{"unauthenticated", auth.ErrUnauthenticated, codeUnauthenticated},
		{"unauthorized", auth.ErrUnauthorized, codeForbidden},
		{"invalid input", fmt.Errorf("%w: name is required", ErrInvalidInput), codeBadUserInput},
		{"incident not found", incidents.ErrIncidentNotFound, codeBadUserInput},
		{"invalid transition", incidents.ErrInvalidStateTransition, codeBadUserInput},
		{"oncall interval", oncall.ErrInvalidInterval, codeBadUserInput},
		{"unknown", errors.New("boom"), codeInternal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapError(tt.err)
			if tt.wantCode == "" {
				if got != nil {
					t.Fatalf("mapError(%v) = %v, want nil", tt.err, got)
				}
				return
			}
			if code(got) != tt.wantCode {
				t.Fatalf("mapError(%v) code = %q, want %q", tt.err, code(got), tt.wantCode)
			}
		})
	}
}

func TestMapErrorDoesNotLeakInternals(t *testing.T) {
	raw := errors.New("mongo: connection refused to db at 10.1.2.3:27017")
	mapped := mapError(raw)

	if code(mapped) != codeInternal {
		t.Fatalf("code = %q, want %q", code(mapped), codeInternal)
	}
	if strings.Contains(mapped.Message, "mongo") || strings.Contains(mapped.Message, "10.1.2.3") {
		t.Fatalf("internal error leaked to client message: %q", mapped.Message)
	}
}

func TestErrorPresenterPassesThroughNativeErrors(t *testing.T) {
	present := NewErrorPresenter(nil)

	native := gqlError(codeBadUserInput, "field must not be null")
	if got := present(context.Background(), native); got != native {
		t.Fatalf("native gqlerror should pass through unchanged")
	}

	sanitised := present(context.Background(), errors.New("raw mongo failure"))
	if code(sanitised) != codeInternal || strings.Contains(sanitised.Message, "mongo") {
		t.Fatalf("resolver error not sanitised: code=%q msg=%q", code(sanitised), sanitised.Message)
	}
}

func TestValidators(t *testing.T) {
	if err := requireNonEmpty("name", "  "); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("blank string should be invalid input, got %v", err)
	}
	if err := requireNonEmpty("name", "ok"); err != nil {
		t.Fatalf("non-empty string should pass, got %v", err)
	}
	if err := validateRotation(nil); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("empty rotation should be invalid input, got %v", err)
	}

	now := time.Now()
	earlier := now.Add(-time.Hour)
	if err := validateTimeRange(&now, &earlier); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("inverted range should be invalid input, got %v", err)
	}
	if err := validateTimeRange(&earlier, &now); err != nil {
		t.Fatalf("valid range should pass, got %v", err)
	}
	if err := validateTimeRange(nil, nil); err != nil {
		t.Fatalf("nil range should pass, got %v", err)
	}
}
