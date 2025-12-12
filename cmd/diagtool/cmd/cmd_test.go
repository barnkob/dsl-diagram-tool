package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// Helper to create a fresh root command for testing
func newTestRootCmd() *cobra.Command {
	// Reset global flags
	outputFile = ""
	outputFormat = "svg"
	themeID = 0
	darkMode = false
	sketchMode = false
	padding = 100
	noCenter = false
	verbose = false

	// Create fresh commands
	testRoot := &cobra.Command{
		Use:           "diagtool",
		Short:         "DSL Diagram Tool - Render D2 diagrams to various formats",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	testRoot.AddCommand(renderCmd)
	testRoot.AddCommand(validateCmd)
	testRoot.AddCommand(versionCmd)

	return testRoot
}

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}
}

func TestDefaultOptions(t *testing.T) {
	if outputFormat != "svg" {
		t.Errorf("Default output format should be svg, got %s", outputFormat)
	}
	if padding != 100 {
		t.Errorf("Default padding should be 100, got %d", padding)
	}
}

func TestRenderCommand_RequiresInput(t *testing.T) {
	cmd := newTestRootCmd()
	cmd.SetArgs([]string{"render"})
	err := cmd.Execute()

	if err == nil {
		t.Error("render command should require input file")
	}
}

func TestRenderCommand_FileNotFound(t *testing.T) {
	cmd := newTestRootCmd()
	cmd.SetArgs([]string{"render", "nonexistent-file.d2"})
	err := cmd.Execute()

	if err == nil {
		t.Error("render command should fail for non-existent file")
	}
	if err != nil && !strings.Contains(err.Error(), "failed to read") {
		t.Errorf("Expected 'failed to read' error, got: %v", err)
	}
}

func TestRenderCommand_InvalidFormat(t *testing.T) {
	// Create temp file
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "test.d2")
	os.WriteFile(inputFile, []byte("a -> b"), 0644)

	cmd := newTestRootCmd()
	cmd.SetArgs([]string{"render", inputFile, "-f", "invalid"})
	err := cmd.Execute()

	if err == nil {
		t.Error("render command should fail for invalid format")
	}
	if err != nil && !strings.Contains(err.Error(), "unsupported output format") {
		t.Errorf("Expected 'unsupported output format' error, got: %v", err)
	}
}

func TestRenderCommand_SVGOutput(t *testing.T) {
	// Create temp file
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "test.d2")
	outputFilePath := filepath.Join(tmpDir, "output.svg")

	os.WriteFile(inputFile, []byte("server -> database: connects"), 0644)

	cmd := newTestRootCmd()
	cmd.SetArgs([]string{"render", inputFile, "-o", outputFilePath})
	err := cmd.Execute()

	if err != nil {
		t.Fatalf("render command failed: %v", err)
	}

	// Check output file exists
	if _, err := os.Stat(outputFilePath); os.IsNotExist(err) {
		t.Fatal("Output file was not created")
	}

	// Check it's SVG
	content, _ := os.ReadFile(outputFilePath)
	if !strings.Contains(string(content), "<svg") {
		t.Error("Output should contain SVG markup")
	}
}

func TestRenderCommand_WithSketch(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "test.d2")
	outputFilePath := filepath.Join(tmpDir, "sketch.svg")

	os.WriteFile(inputFile, []byte("a -> b"), 0644)

	cmd := newTestRootCmd()
	cmd.SetArgs([]string{"render", inputFile, "-o", outputFilePath, "--sketch"})
	err := cmd.Execute()

	if err != nil {
		t.Fatalf("render with sketch failed: %v", err)
	}

	if _, err := os.Stat(outputFilePath); os.IsNotExist(err) {
		t.Error("Sketch output file was not created")
	}
}

func TestRenderCommand_WithTheme(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "test.d2")
	outputFilePath := filepath.Join(tmpDir, "themed.svg")

	os.WriteFile(inputFile, []byte("x -> y"), 0644)

	cmd := newTestRootCmd()
	cmd.SetArgs([]string{"render", inputFile, "-o", outputFilePath, "-t", "3"})
	err := cmd.Execute()

	if err != nil {
		t.Fatalf("render with theme failed: %v", err)
	}

	if _, err := os.Stat(outputFilePath); os.IsNotExist(err) {
		t.Error("Themed output file was not created")
	}
}

func TestRenderCommand_PNGExport(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "test.d2")
	outputFilePath := filepath.Join(tmpDir, "test.png")

	os.WriteFile(inputFile, []byte("a -> b"), 0644)

	cmd := newTestRootCmd()
	cmd.SetArgs([]string{"render", inputFile, "-f", "png", "-o", outputFilePath})
	err := cmd.Execute()

	if err != nil {
		t.Fatalf("PNG render failed: %v", err)
	}

	// Check output file exists
	if _, err := os.Stat(outputFilePath); os.IsNotExist(err) {
		t.Fatal("PNG output file was not created")
	}

	// Check it's actually a PNG (magic bytes: 0x89 PNG)
	content, _ := os.ReadFile(outputFilePath)
	if len(content) < 8 {
		t.Fatal("PNG file is too small")
	}
	if content[0] != 0x89 || content[1] != 'P' || content[2] != 'N' || content[3] != 'G' {
		t.Error("Output is not a valid PNG file (incorrect magic bytes)")
	}

	t.Logf("PNG export successful: %d bytes", len(content))
}

func TestValidateCommand_RequiresInput(t *testing.T) {
	cmd := newTestRootCmd()
	cmd.SetArgs([]string{"validate"})
	err := cmd.Execute()

	if err == nil {
		t.Error("validate command should require input file")
	}
}

func TestValidateCommand_FileNotFound(t *testing.T) {
	cmd := newTestRootCmd()
	cmd.SetArgs([]string{"validate", "nonexistent.d2"})
	err := cmd.Execute()

	if err == nil {
		t.Error("validate command should fail for non-existent file")
	}
}

func TestValidateCommand_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "valid.d2")

	os.WriteFile(inputFile, []byte("server -> database\ndatabase -> cache"), 0644)

	cmd := newTestRootCmd()
	cmd.SetArgs([]string{"validate", inputFile})
	err := cmd.Execute()

	if err != nil {
		t.Fatalf("validate should succeed for valid file: %v", err)
	}
}

func TestValidateCommand_InvalidSyntax(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "invalid.d2")

	// This is invalid D2 syntax
	os.WriteFile(inputFile, []byte("a -> -> b"), 0644)

	cmd := newTestRootCmd()
	cmd.SetArgs([]string{"validate", inputFile})
	err := cmd.Execute()

	if err == nil {
		t.Error("validate should fail for invalid syntax")
	}
}

// Integration tests with example files
func TestRenderCommand_ExampleFiles(t *testing.T) {
	examplesDir := "../../../examples"

	files, err := filepath.Glob(filepath.Join(examplesDir, "*.d2"))
	if err != nil {
		t.Fatalf("Failed to glob examples: %v", err)
	}

	if len(files) == 0 {
		t.Skip("No example files found")
	}

	tmpDir := t.TempDir()

	for _, file := range files {
		// Skip macOS metadata files
		if strings.HasPrefix(filepath.Base(file), "._") {
			continue
		}

		t.Run(filepath.Base(file), func(t *testing.T) {
			outputFilePath := filepath.Join(tmpDir, strings.TrimSuffix(filepath.Base(file), ".d2")+".svg")

			cmd := newTestRootCmd()
			cmd.SetArgs([]string{"render", file, "-o", outputFilePath})
			err := cmd.Execute()

			if err != nil {
				t.Fatalf("Render failed for %s: %v", file, err)
			}

			// Check output exists and contains SVG
			content, err := os.ReadFile(outputFilePath)
			if err != nil {
				t.Fatalf("Failed to read output: %v", err)
			}
			if !strings.Contains(string(content), "<svg") {
				t.Error("Output is not SVG")
			}

			t.Logf("Rendered %s: %d bytes", filepath.Base(file), len(content))
		})
	}
}

func TestValidateCommand_ExampleFiles(t *testing.T) {
	examplesDir := "../../../examples"

	files, err := filepath.Glob(filepath.Join(examplesDir, "*.d2"))
	if err != nil {
		t.Fatalf("Failed to glob examples: %v", err)
	}

	if len(files) == 0 {
		t.Skip("No example files found")
	}

	for _, file := range files {
		// Skip macOS metadata files
		if strings.HasPrefix(filepath.Base(file), "._") {
			continue
		}

		t.Run(filepath.Base(file), func(t *testing.T) {
			cmd := newTestRootCmd()
			cmd.SetArgs([]string{"validate", file})
			err := cmd.Execute()

			if err != nil {
				t.Fatalf("Validate failed for %s: %v", file, err)
			}
		})
	}
}
