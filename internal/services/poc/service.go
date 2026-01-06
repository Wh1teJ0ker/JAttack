package poc

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	nuclei "github.com/projectdiscovery/nuclei/v3/lib"
	"github.com/projectdiscovery/nuclei/v3/pkg/output"
)

type PocService struct {
	ctx            context.Context
	dataDir        string
	libDir         string
	mu             sync.Mutex
	userPythonPath string
}

func NewPocService(dataDir string) *PocService {
	return &PocService{
		dataDir: dataDir,
		libDir:  filepath.Join(dataDir, "lib"),
	}
}

func (s *PocService) Startup(ctx context.Context) {
	s.ctx = ctx
}

// SetPythonPath allows user to set a custom python path
func (s *PocService) SetPythonPath(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.userPythonPath = path
}

// GetPythonPath tries to find a valid python executable
func (s *PocService) GetPythonPath() (string, error) {
	s.mu.Lock()
	userPath := s.userPythonPath
	s.mu.Unlock()

	// 1. Try user config
	if userPath != "" {
		return userPath, nil
	}

	// 2. Try python3
	path, err := exec.LookPath("python3")
	if err == nil {
		return path, nil
	}

	// 3. Try python
	path, err = exec.LookPath("python")
	if err == nil {
		return path, nil
	}

	return "", fmt.Errorf("python environment not found")
}

// RunPython executes a python script content
func (s *PocService) RunPython(scriptContent string) (string, error) {
	pythonPath, err := s.GetPythonPath()
	if err != nil {
		return "", fmt.Errorf("python environment check failed: %v", err)
	}

	// Create temp file
	tmpFile, err := os.CreateTemp("", "poc-*.py")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err = tmpFile.WriteString(scriptContent); err != nil {
		return "", fmt.Errorf("failed to write script: %v", err)
	}
	tmpFile.Close()

	cmd := exec.Command(pythonPath, tmpFile.Name())
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("execution failed: %v, output: %s", err, string(output))
	}

	return string(output), nil
}

// RunNuclei executes nuclei with given target and template content using the embedded library
func (s *PocService) RunNuclei(target string, templateContent string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create temp template file
	tmpFile, err := os.CreateTemp("", "nuclei-*.yaml")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err = tmpFile.WriteString(templateContent); err != nil {
		return "", fmt.Errorf("failed to write template: %v", err)
	}
	tmpFile.Close()

	// Initialize Nuclei Engine
	// We configure it to scan the target with the specific template file
	ne, err := nuclei.NewNucleiEngine(
		nuclei.WithTemplatesOrWorkflows(nuclei.TemplateSources{
			Templates: []string{tmpFile.Name()},
		}),
	)
	if err != nil {
		return "", fmt.Errorf("failed to init nuclei engine: %v", err)
	}

	// Load Target
	ne.LoadTargets([]string{target}, false)

	var resultsBuilder strings.Builder
	resultsBuilder.WriteString(fmt.Sprintf("Scanning %s with provided template...\n\n", target))

	// Execute with callback
	// Note: Nuclei engine execution might capture stdout/stderr, so we rely on the callback for results
	// but we also want to capture logs if possible.
	err = ne.ExecuteWithCallback(func(event *output.ResultEvent) {
		// Basic formatting
		resultsBuilder.WriteString(fmt.Sprintf("[Matched] %s\n", event.TemplateID))
		resultsBuilder.WriteString(fmt.Sprintf("Name: %s\n", event.Info.Name))
		resultsBuilder.WriteString(fmt.Sprintf("Severity: %s\n", event.Info.SeverityHolder))
		resultsBuilder.WriteString(fmt.Sprintf("Host: %s\n", event.Host))
		resultsBuilder.WriteString(fmt.Sprintf("Matched: %s\n", event.Matched))
		if len(event.ExtractedResults) > 0 {
			resultsBuilder.WriteString(fmt.Sprintf("Extracted: %v\n", event.ExtractedResults))
		}
		resultsBuilder.WriteString("\n")
	})

	if err != nil {
		return resultsBuilder.String(), fmt.Errorf("nuclei execution failed: %v", err)
	}

	finalOutput := resultsBuilder.String()
	if finalOutput == fmt.Sprintf("Scanning %s with provided template...\n\n", target) {
		finalOutput += "No vulnerabilities found or no match."
	}

	return finalOutput, nil
}
