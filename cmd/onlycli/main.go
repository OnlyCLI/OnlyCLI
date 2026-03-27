package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/onlycli/onlycli/internal/codegen"
	"github.com/onlycli/onlycli/internal/model"
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

var skillmdCmd = &cobra.Command{
	Use:   "skillmd",
	Short: "Generate a SKILL.md agent-facing summary from an OpenAPI spec",
	RunE:  runSkillmd,
}

var (
	flagSpec      string
	flagName      string
	flagAuth      string
	flagOut       string
	flagModule    string
	flagSkillSpec string
	flagSkillName string
	flagSkillOut  string
	flagSkillMax  int
)

func init() {
	rootCmd.Version = version
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(skillmdCmd)

	generateCmd.Flags().StringVar(&flagSpec, "spec", "", "Path or URL to OpenAPI spec (YAML/JSON)")
	generateCmd.Flags().StringVar(&flagName, "name", "", "Name of the generated CLI")
	generateCmd.Flags().StringVar(&flagAuth, "auth", "", "Auth type: bearer, basic, apikey (auto-detected from spec if omitted)")
	generateCmd.Flags().StringVar(&flagOut, "out", "", "Output directory for generated project")
	generateCmd.Flags().StringVar(&flagModule, "module", "", "Go module path (default: github.com/<name>/<name>-cli)")

	_ = generateCmd.MarkFlagRequired("spec")
	_ = generateCmd.MarkFlagRequired("name")
	_ = generateCmd.MarkFlagRequired("out")

	skillmdCmd.Flags().StringVar(&flagSkillSpec, "spec", "", "Path or URL to OpenAPI spec (YAML/JSON)")
	skillmdCmd.Flags().StringVar(&flagSkillName, "name", "", "CLI binary name")
	skillmdCmd.Flags().StringVar(&flagSkillOut, "out", "", "Output file path (default: stdout)")
	skillmdCmd.Flags().IntVar(&flagSkillMax, "max-commands", 40, "Max commands to include in the summary")

	_ = skillmdCmd.MarkFlagRequired("spec")
	_ = skillmdCmd.MarkFlagRequired("name")
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

func runSkillmd(cmd *cobra.Command, args []string) error {
	specBytes, err := loadSpec(flagSkillSpec)
	if err != nil {
		return fmt.Errorf("failed to load spec: %w", err)
	}

	envVar := strings.ToUpper(strings.ReplaceAll(flagSkillName, "-", "_")) + "_TOKEN"
	modulePath := fmt.Sprintf("github.com/%s/%s-cli", flagSkillName, flagSkillName)

	spec, err := parser.Parse(specBytes, flagSkillName, "", envVar, modulePath)
	if err != nil {
		return fmt.Errorf("failed to parse spec: %w", err)
	}

	var out *os.File
	if flagSkillOut != "" {
		out, err = os.Create(flagSkillOut)
		if err != nil {
			return fmt.Errorf("create output file: %w", err)
		}
		defer out.Close()
	} else {
		out = os.Stdout
	}

	fmt.Fprintf(out, "# %s CLI\n\n", flagSkillName)
	if spec.Description != "" {
		fmt.Fprintf(out, "%s\n\n", spec.Description)
	}
	fmt.Fprintf(out, "## Quick reference\n\n")
	fmt.Fprintf(out, "- Binary: `%s`\n", flagSkillName)
	fmt.Fprintf(out, "- Auth: `%s` (env: `%s`)\n", spec.AuthType, spec.AuthEnvVar)
	if spec.BaseURL != "" {
		fmt.Fprintf(out, "- Base URL: `%s`\n", spec.BaseURL)
	}
	fmt.Fprintf(out, "- Global flags: `--format`, `--transform`, `--page-limit`, `--stream`, `--verbose`, `--dry-run`\n")
	fmt.Fprintf(out, "- Output formats: `json`, `pretty`, `yaml`, `jsonl`, `table`, `csv`, `raw`\n")
	fmt.Fprintf(out, "\n## Commands\n\n")

	cmdCount := 0
	for _, group := range spec.Groups {
		if cmdCount >= flagSkillMax {
			break
		}
		fmt.Fprintf(out, "### %s\n\n", group.Name)
		for _, c := range group.Commands {
			if cmdCount >= flagSkillMax {
				break
			}
			desc := c.Description
			if desc == "" {
				desc = c.FullName
			}
			if len(desc) > 80 {
				desc = desc[:77] + "..."
			}
			fmt.Fprintf(out, "- `%s %s %s`", flagSkillName, group.Name, c.Name)

			var flags []string
			for _, p := range c.Parameters {
				if p.Required {
					flags = append(flags, "--"+p.Name)
				}
			}
			if len(flags) > 0 {
				fmt.Fprintf(out, " %s", strings.Join(flags, " "))
			}
			if c.HasBody {
				fmt.Fprintf(out, " [--data]")
			}
			fmt.Fprintf(out, " -- %s\n", desc)
			cmdCount++
		}
		fmt.Fprintln(out)
	}

	if cmdCount < totalCommands(spec) {
		fmt.Fprintf(out, "_Showing %d of %d commands. Run `%s --help` for the full list._\n\n", cmdCount, totalCommands(spec), flagSkillName)
	}

	fmt.Fprintf(out, "## Examples\n\n")
	fmt.Fprintf(out, "```bash\n")
	fmt.Fprintf(out, "# List resources\n")
	fmt.Fprintf(out, "%s %s --format table\n\n", flagSkillName, exampleListCmd(spec))
	fmt.Fprintf(out, "# Get details with GJSON transform\n")
	fmt.Fprintf(out, "%s %s --transform 'name' --format pretty\n", flagSkillName, exampleGetCmd(spec))
	fmt.Fprintf(out, "```\n")

	if flagSkillOut != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Generated SKILL.md at %s (%d commands)\n", flagSkillOut, cmdCount)
	}

	return nil
}

func totalCommands(spec *model.APISpec) int {
	n := 0
	for _, g := range spec.Groups {
		n += len(g.Commands)
	}
	return n
}

func exampleListCmd(spec *model.APISpec) string {
	for _, g := range spec.Groups {
		for _, c := range g.Commands {
			if strings.HasPrefix(c.Name, "list") && c.Method == "GET" {
				return g.Name + " " + c.Name
			}
		}
	}
	if len(spec.Groups) > 0 && len(spec.Groups[0].Commands) > 0 {
		return spec.Groups[0].Name + " " + spec.Groups[0].Commands[0].Name
	}
	return "<group> <command>"
}

func exampleGetCmd(spec *model.APISpec) string {
	for _, g := range spec.Groups {
		for _, c := range g.Commands {
			if strings.HasPrefix(c.Name, "get") && c.Method == "GET" {
				return g.Name + " " + c.Name
			}
		}
	}
	return exampleListCmd(spec)
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
