package input

import (
	"fmt"
	"strings"

	sliceUtils "github.com/edoardottt/cariddi/internal/slice"
)

// StringSlice is a custom flag type for []string
type StringSlice []string

func (s *StringSlice) String() string {
	return fmt.Sprint(*s)
}

func (s *StringSlice) Set(value string) error {
	parts := strings.Split(value, ",")
	for _, part := range parts {
		*s = append(*s, strings.ToLower(strings.TrimSpace(part)))
	}

	*s = sliceUtils.RemoveDuplicateValues(*s)
	fmt.Println(s)

	return nil
}
