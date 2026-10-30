package input

import (
	"fmt"
	"strings"

	sliceUtils "github.com/edoardottt/cariddi/v2/internal/slice"
)

// StringSlice is a custom flag type for []string.
type StringSlice []string

func (s *StringSlice) String() string {
	return fmt.Sprint(*s)
}

// Set adds the comma-separated items of value, skipping the empty ones:
// an empty extension would match every URL without an extension.
func (s *StringSlice) Set(value string) error {
	parts := strings.Split(value, ",")
	for _, part := range parts {
		if item := strings.ToLower(strings.TrimSpace(part)); item != "" {
			*s = append(*s, item)
		}
	}

	// Never nil, even if value has no items (e.g. -ie ""): the
	// flag is set, so the default extensions must not be used.
	*s = sliceUtils.RemoveDuplicateValues(*s)

	return nil
}
