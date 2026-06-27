package cmd

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// confirm prints prompt and returns true if the user answers yes (Y or YES,
// case-insensitive). Any other input, including empty input or EOF, is no.
func confirm(cmd *cobra.Command, prompt string) bool {
	fmt.Fprint(cmd.OutOrStdout(), prompt)
	scanner := bufio.NewScanner(cmd.InOrStdin())
	scanner.Scan()
	answer := strings.ToUpper(strings.TrimSpace(scanner.Text()))
	return answer == "Y" || answer == "YES"
}
