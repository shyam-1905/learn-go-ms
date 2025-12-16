package service

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

// TemplateService handles rendering email templates
type TemplateService struct {
	templates map[string]*template.Template
}

// NewTemplateService creates a new template service
// templatesDir: path to the templates directory
func NewTemplateService(templatesDir string) (*TemplateService, error) {
	ts := &TemplateService{
		templates: make(map[string]*template.Template),
	}

	// Load all templates
	templateFiles := []string{
		"expense_created.html",
		"expense_updated.html",
		"receipt_uploaded.html",
		"receipt_linked.html",
		"user_registered.html",
	}

	for _, filename := range templateFiles {
		filepath := filepath.Join(templatesDir, filename)

		// Check if file exists
		if _, err := os.Stat(filepath); os.IsNotExist(err) {
			// Template doesn't exist, create a default one
			if err := ts.createDefaultTemplate(filepath, filename); err != nil {
				return nil, fmt.Errorf("failed to create default template %s: %w", filename, err)
			}
		}

		// Parse template
		tmpl, err := template.ParseFiles(filepath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse template %s: %w", filename, err)
		}

		// Store template (use base name without extension as key)
		key := filename[:len(filename)-5] // Remove .html extension
		ts.templates[key] = tmpl
	}

	return ts, nil
}

// Render renders a template with the given data
func (ts *TemplateService) Render(templateName string, data interface{}) (string, error) {
	tmpl, ok := ts.templates[templateName]
	if !ok {
		return "", fmt.Errorf("template %s not found", templateName)
	}

	// Render template to string
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render template: %w", err)
	}

	return buf.String(), nil
}

// createDefaultTemplate creates a default template if it doesn't exist
func (ts *TemplateService) createDefaultTemplate(filepath, filename string) error {
	// Create default HTML template
	defaultHTML := `<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<title>Notification</title>
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
		.container { max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background-color: #4CAF50; color: white; padding: 20px; text-align: center; }
		.content { padding: 20px; background-color: #f9f9f9; }
		.footer { text-align: center; padding: 20px; color: #666; font-size: 12px; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>Expense Tracker Notification</h1>
		</div>
		<div class="content">
			{{.Content}}
		</div>
		<div class="footer">
			<p>This is an automated notification from Expense Tracker.</p>
		</div>
	</div>
</body>
</html>`

	return os.WriteFile(filepath, []byte(defaultHTML), 0644)
}
