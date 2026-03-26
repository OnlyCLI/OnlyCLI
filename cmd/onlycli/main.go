package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/onlycli/onlycli/internal/codegen"
	"github.com/onlycli/onlycli/internal/parser"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "onlycli",
	Short: "Turn any OpenAPI spec into a native CLI. No MCP, no bloat.",
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "onlycli version %s (commit: %s, built: %s)\n", version, commit, date)
	},
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a CLI project from an OpenAPI spec",
	RunE:  runGenerate,
}

var (
	flagSpec   string
	flagName   string
	flagAuth   string
	flagOut    string
	flagModule string
)

func init() {
	rootCmd.Version = version
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringVar(&flagSpec, "spec", "", "Path or URL to OpenAPI spec (YAML/JSON)")
	generateCmd.Flags().StringVar(&flagName, "name", "", "Name of the generated CLI")
	generateCmd.Flags().StringVar(&flagAuth, "auth", "", "Auth type: bearer, basic, apikey (auto-detected from spec if omitted)")
	generateCmd.Flags().StringVar(&flagOut, "out", "", "Output directory for generated project")
	generateCmd.Flags().StringVar(&flagModule, "module", "", "Go module path (default: github.com/<name>/<name>-cli)")

	_ = generateCmd.MarkFlagRequired("spec")
	_ = generateCmd.MarkFlagRequired("name")
	_ = generateCmd.MarkFlagRequired("out")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	specBytes, err := loadSpec(flagSpec)
	if err != nil {
		return fmt.Errorf("failed to load spec: %w", err)
	}

	modulePath := flagModule
	if modulePath == "" {
		modulePath = fmt.Sprintf("github.com/%s/%s-cli", flagName, flagName)
	}

	envVar := strings.ToUpper(strings.ReplaceAll(flagName, "-", "_")) + "_TOKEN"

	spec, err := parser.Parse(specBytes, flagName, flagAuth, envVar, modulePath)
	if err != nil {
		return fmt.Errorf("failed to parse spec: %w", err)
	}

	gen, err := codegen.NewGenerator(spec, flagOut)
	if err != nil {
		return fmt.Errorf("failed to create generator: %w", err)
	}

	if err := gen.Generate(); err != nil {
		return fmt.Errorf("failed to generate CLI project: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Generated CLI project at %s\n", flagOut)
	fmt.Fprintf(cmd.OutOrStdout(), "  Groups: %d\n", len(spec.Groups))
	totalCmds := 0
	for _, g := range spec.Groups {
		totalCmds += len(g.Commands)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "  Commands: %d\n", totalCmds)
	fmt.Fprintf(cmd.OutOrStdout(), "  Auth: %s (env: %s)\n", spec.AuthType, spec.AuthEnvVar)
	fmt.Fprintf(cmd.OutOrStdout(), "\nTo build:\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  cd %s && go mod tidy && go build -o %s .\n", flagOut, flagName)

	return nil
}

func loadSpec(specPath string) ([]byte, error) {
	if strings.HasPrefix(specPath, "http://") || strings.HasPrefix(specPath, "https://") {
		parsedURL, err := url.Parse(specPath)
		if err != nil {
			return nil, fmt.Errorf("invalid spec URL: %w", err)
		}

		resp, err := http.Get(parsedURL.String())
		if err != nil {
			return nil, fmt.Errorf("failed to fetch spec from URL: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("HTTP %d fetching spec", resp.StatusCode)
		}

		return io.ReadAll(resp.Body)
	}

	return os.ReadFile(specPath)
}
