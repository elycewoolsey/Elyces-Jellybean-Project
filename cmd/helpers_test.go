package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// setFlags applies flag values to a command. Bool flags accept "true"/"false".
func setFlags(t *testing.T, c *cobra.Command, flags map[string]string) {
	t.Helper()
	for k, v := range flags {
		if err := c.Flags().Set(k, v); err != nil {
			t.Fatalf("set flag %q=%q: %v", k, v, err)
		}
	}
}

// runWithStdin attaches a stdin reader to a command (no-op if stdin is empty).
func runWithStdin(c *cobra.Command, stdin string) {
	if stdin != "" {
		c.SetIn(strings.NewReader(stdin))
	}
}
