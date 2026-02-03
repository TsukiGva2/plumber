package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/getplumber/plumber/configuration"
	"github.com/getplumber/plumber/internal/defaultconfig"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	// config view flags
	configViewFile    string
	configViewNoColor bool

	// config generate flags
	configGenerateOutput string
	configGenerateForce  bool
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage Plumber configuration",
	Long:  `Commands for viewing and managing Plumber configuration files.`,
}

var configViewCmd = &cobra.Command{
	Use:   "view",
	Short: "Display the effective configuration",
	Long: `Display a clean, human-readable view of the effective configuration.

This command loads and parses the configuration file, then displays it
without comments, making it easy to see exactly what settings are active.

Booleans are colorized for quick scanning:
  - true  → green
  - false → red

Examples:
  # View the default .plumber.yaml
  plumber config view

  # View a specific config file
  plumber config view --config custom-plumber.yaml

  # View without colors (for piping or scripts)
  plumber config view --no-color
`,
	RunE: runConfigView,
}

var configGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a default .plumber.yaml configuration file",
	Long: `Generate a default .plumber.yaml configuration file.

This creates a configuration file with sensible defaults that you can
customize for your organization's compliance requirements.

The generated config includes:
- Container image tag policies (forbid 'latest', 'dev', etc.)
- Trusted registry whitelist
- Branch protection requirements

Examples:
  # Generate default config in current directory
  plumber config generate

  # Generate config with custom filename
  plumber config generate --output my-plumber-config.yaml

  # Overwrite existing file
  plumber config generate --force
`,
	RunE: runConfigGenerate,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configViewCmd)
	configCmd.AddCommand(configGenerateCmd)

	// config view flags
	configViewCmd.Flags().StringVarP(&configViewFile, "config", "c", ".plumber.yaml", "Path to configuration file")
	configViewCmd.Flags().BoolVar(&configViewNoColor, "no-color", false, "Disable colorized output")

	// config generate flags
	configGenerateCmd.Flags().StringVarP(&configGenerateOutput, "output", "o", ".plumber.yaml", "Output file path")
	configGenerateCmd.Flags().BoolVarP(&configGenerateForce, "force", "f", false, "Overwrite existing file")
}

func runConfigView(cmd *cobra.Command, args []string) error {
	// Suppress debug logs for clean output (unless verbose)
	if !verbose {
		logrus.SetLevel(logrus.WarnLevel)
	}

	// Determine if we should colorize output
	useColor := !configViewNoColor
	// Auto-detect: disable color if not a terminal (unless explicitly set)
	if !cmd.Flags().Changed("no-color") {
		if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) == 0 {
			useColor = false
		}
	}

	// Load the configuration
	config, _, err := configuration.LoadPlumberConfig(configViewFile)
	if err != nil {
		return err
	}

	// Marshal to clean YAML (this strips comments)
	cleanYAML, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to serialize configuration: %w", err)
	}

	// Convert to string for processing
	output := string(cleanYAML)

	// Colorize if enabled
	if useColor {
		output = colorizeBooleans(output)
	}

	fmt.Print(output)
	return nil
}

// colorizeBooleans replaces true/false with colorized versions
func colorizeBooleans(input string) string {
	// Match 'true' and 'false' as YAML boolean values (after : or as list items)
	// This regex ensures we only match actual boolean values, not substrings
	trueRegex := regexp.MustCompile(`(:\s*)true(\s*$)`)
	falseRegex := regexp.MustCompile(`(:\s*)false(\s*$)`)

	lines := strings.Split(input, "\n")
	for i, line := range lines {
		lines[i] = trueRegex.ReplaceAllString(line, fmt.Sprintf("${1}%strue%s${2}", colorGreen, colorReset))
		lines[i] = falseRegex.ReplaceAllString(lines[i], fmt.Sprintf("${1}%sfalse%s${2}", colorRed, colorReset))
	}

	return strings.Join(lines, "\n")
}

func runConfigGenerate(cmd *cobra.Command, args []string) error {
	// Check if file already exists
	if _, err := os.Stat(configGenerateOutput); err == nil {
		if !configGenerateForce {
			return fmt.Errorf("file %s already exists. Use --force to overwrite", configGenerateOutput)
		}
		fmt.Fprintf(os.Stderr, "Overwriting existing file: %s\n", configGenerateOutput)
	}

	// Get embedded default config
	configContent := defaultconfig.Get()

	// Write to file
	if err := os.WriteFile(configGenerateOutput, configContent, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("Generated %s\n", configGenerateOutput)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Review and customize the configuration for your needs")
	fmt.Println("  2. Export the GITLAB_TOKEN environment variable if you haven't already")
	fmt.Println("  3. Run: plumber analyze --gitlab-url <url> --project <path>")

	return nil
}
