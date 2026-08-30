package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "m2h [file]",
	Short: "Simple replacement for pandoc",
	Long: `Simple replacement for pandoc, which only converts markdown to html.

Converts the given file/stdin to html/htm and prints it to stdout.
Base64 images are always embedded into the generated html.

Attention: Paged.js is not supported when using mermaid diagrams !

Examples:

$ m2h test.md > test.htm

$ m2h -b test.md

$ m2h -p test-no-mermaid-inside.md > test-no-mermaid-inside.htm

$ m2h -l proto test.proto > test.htm

`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		// without a file name and without piped input there is nothing to
		// convert, so print the help instead of blocking on stdin
		if len(args) == 0 && isTerminal(os.Stdin) {
			_ = cmd.Help()
			return
		}

		convCmdFunc(cmd, args)
	},
}

// isTerminal reports whether file is attached to an interactive terminal.
// Redirections and pipes are not character devices, so they read as false.
func isTerminal(file *os.File) bool {

	fi, err := file.Stat()
	if err != nil {
		return false
	}

	return (fi.Mode() & os.ModeCharDevice) != 0
}

// Execute runs the root command. version is injected from main.version (set at
// build time via -ldflags) and exposed through the cobra `--version` flag.
func Execute(version string) {
	rootCmd.Version = version
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

func initConfig() {
}
