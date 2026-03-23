package templates

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTemplate(t *testing.T) {
	tests := []struct {
		name      string
		stack     string
		wantNil   bool
		wantStack Stack
	}{
		{
			name:      "React template",
			stack:     "react",
			wantNil:   false,
			wantStack: StackReact,
		},
		{
			name:      "React with TS suffix",
			stack:     "react+ts",
			wantNil:   false,
			wantStack: StackReact,
		},
		{
			name:      "Vue template",
			stack:     "vue",
			wantNil:   false,
			wantStack: StackVue,
		},
		{
			name:      "Angular template",
			stack:     "angular",
			wantNil:   false,
			wantStack: StackAngular,
		},
		{
			name:      "Angular alias",
			stack:     "ng",
			wantNil:   false,
			wantStack: StackAngular,
		},
		{
			name:      "Next.js template",
			stack:     "nextjs",
			wantNil:   false,
			wantStack: StackNextJS,
		},
		{
			name:      "Next.js with dot",
			stack:     "next.js",
			wantNil:   false,
			wantStack: StackNextJS,
		},
		{
			name:      "Go Chi template",
			stack:     "go-chi",
			wantNil:   false,
			wantStack: StackGoChi,
		},
		{
			name:      "Go Chi alias",
			stack:     "chi",
			wantNil:   false,
			wantStack: StackGoChi,
		},
		{
			name:      "Go Echo template",
			stack:     "go-echo",
			wantNil:   false,
			wantStack: StackGoEcho,
		},
		{
			name:      "Go Gin template",
			stack:     "go-gin",
			wantNil:   false,
			wantStack: StackGoGin,
		},
		{
			name:      "Go Gin alias",
			stack:     "gin",
			wantNil:   false,
			wantStack: StackGoGin,
		},
		{
			name:      "Python FastAPI",
			stack:     "python-fastapi",
			wantNil:   false,
			wantStack: StackPythonFastAPI,
		},
		{
			name:      "FastAPI alias",
			stack:     "fastapi",
			wantNil:   false,
			wantStack: StackPythonFastAPI,
		},
		{
			name:      "Python Django",
			stack:     "python-django",
			wantNil:   false,
			wantStack: StackPythonDjango,
		},
		{
			name:      "Django alias",
			stack:     "django",
			wantNil:   false,
			wantStack: StackPythonDjango,
		},
		{
			name:      "Node Express",
			stack:     "node-express",
			wantNil:   false,
			wantStack: StackNodeExpress,
		},
		{
			name:      "Express alias",
			stack:     "express",
			wantNil:   false,
			wantStack: StackNodeExpress,
		},
		{
			name:      "Node alias",
			stack:     "node",
			wantNil:   false,
			wantStack: StackNodeExpress,
		},
		{
			name:      "Unknown stack",
			stack:     "unknown-stack",
			wantNil:   true,
			wantStack: "",
		},
		{
			name:      "Empty string",
			stack:     "",
			wantNil:   true,
			wantStack: "",
		},
		{
			name:      "Case insensitive - REACT",
			stack:     "REACT",
			wantNil:   false,
			wantStack: StackReact,
		},
		{
			name:      "Case insensitive - Vue",
			stack:     "Vue",
			wantNil:   false,
			wantStack: StackVue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetTemplate(tt.stack)

			if tt.wantNil {
				assert.Nil(t, got)
			} else {
				require.NotNil(t, got)
				assert.Equal(t, tt.wantStack, got.Stack)
			}
		})
	}
}

func TestDetectStack(t *testing.T) {
	tests := []struct {
		name     string
		idea     string
		expected Stack
	}{
		{
			name:     "React keywords",
			idea:     "I want to build a React app with TypeScript and Vite",
			expected: StackReact,
		},
		{
			name:     "Vue keywords",
			idea:     "Creating a Vue.js application with Composition API",
			expected: StackVue,
		},
		{
			name:     "Angular keywords",
			idea:     "Need an Angular project with standalone components",
			expected: StackAngular,
		},
		{
			name:     "Next.js keywords",
			idea:     "Building a Next.js app with SSR capabilities",
			expected: StackNextJS,
		},
		{
			name:     "Next.js App Router",
			idea:     "Next.js with App Router and Server Components",
			expected: StackNextJS,
		},
		{
			name:     "Go Chi",
			idea:     "Creating a Go API with Chi router",
			expected: StackGoChi,
		},
		{
			name:     "Go Echo",
			idea:     "Building with Go Echo framework",
			expected: StackGoEcho,
		},
		{
			name:     "Go Gin",
			idea:     "Go Gin REST API",
			expected: StackGoGin,
		},
		{
			name:     "Python FastAPI",
			idea:     "FastAPI Python async API",
			expected: StackPythonFastAPI,
		},
		{
			name:     "Django",
			idea:     "Django project with ORM",
			expected: StackPythonDjango,
		},
		{
			name:     "Express",
			idea:     "Node.js Express backend",
			expected: StackNodeExpress,
		},
		{
			name:     "No keywords - defaults to Express",
			idea:     "A simple web project",
			expected: StackNodeExpress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectStack(tt.idea)
			assert.Equal(t, string(tt.expected), got)
		})
	}
}

func TestGetAllStacks(t *testing.T) {
	stacks := GetAllStacks()

	assert.Len(t, stacks, 10)

	expectedStacks := []Stack{
		StackReact,
		StackVue,
		StackAngular,
		StackNextJS,
		StackGoChi,
		StackGoEcho,
		StackGoGin,
		StackPythonFastAPI,
		StackPythonDjango,
		StackNodeExpress,
	}

	for _, expected := range expectedStacks {
		found := false
		for _, s := range stacks {
			if s == expected {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected stack %s not found in GetAllStacks()", expected)
	}
}

func TestTemplateStructure(t *testing.T) {
	// Verify that all templates have required fields
	registry := getDefaultRegistry()

	for stack, template := range registry {
		t.Run(string(stack), func(t *testing.T) {
			require.NotNil(t, template, "Template for %s should not be nil", stack)
			assert.NotEmpty(t, template.Name, "Name should not be empty")
			assert.NotEmpty(t, template.Description, "Description should not be empty")
			assert.NotEmpty(t, template.Folders, "Folders should not be empty")
			assert.NotEmpty(t, template.Conventions.Files.Pattern, "Files naming pattern should not be empty")
			assert.NotEmpty(t, template.Testing.Framework, "Testing framework should not be empty")
		})
	}
}

func TestReactTemplateSkills(t *testing.T) {
	template := reactTemplate()

	require.NotEmpty(t, template.Skills)

	// Check that SDD skills are included
	skillNames := make([]string, len(template.Skills))
	for i, s := range template.Skills {
		skillNames[i] = s.Name
	}

	expectedSkills := []string{"sdd-init", "sdd-spec", "sdd-design", "sdd-tasks", "sdd-apply"}
	for _, expected := range expectedSkills {
		found := false
		for _, name := range skillNames {
			if name == expected {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected skill %s not found in React template", expected)
	}
}

func TestGoTemplateSkills(t *testing.T) {
	t.Run("Chi template includes go-testing", func(t *testing.T) {
		template := goChiTemplate()
		require.NotEmpty(t, template.Skills)

		found := false
		for _, s := range template.Skills {
			if s.Name == "go-testing" {
				found = true
				assert.True(t, s.Required, "go-testing should be required")
				break
			}
		}
		assert.True(t, found, "go-testing skill should be included in Go Chi template")
	})

	t.Run("Gin template includes go-testing", func(t *testing.T) {
		template := goGinTemplate()
		require.NotEmpty(t, template.Skills)

		found := false
		for _, s := range template.Skills {
			if s.Name == "go-testing" {
				found = true
				break
			}
		}
		assert.True(t, found, "go-testing skill should be included in Go Gin template")
	})

	t.Run("Echo template includes go-testing", func(t *testing.T) {
		template := goEchoTemplate()
		require.NotEmpty(t, template.Skills)

		found := false
		for _, s := range template.Skills {
			if s.Name == "go-testing" {
				found = true
				break
			}
		}
		assert.True(t, found, "go-testing skill should be included in Go Echo template")
	})
}

func TestNamingConventions(t *testing.T) {
	// Test React conventions
	template := reactTemplate()

	assert.Equal(t, "kebab-case", template.Conventions.Files.Pattern)
	assert.Equal(t, "PascalCase", template.Conventions.Components.Pattern)
	assert.Equal(t, "camelCase", template.Conventions.Functions.Pattern)
	assert.Equal(t, "SCREAMING_SNAKE_CASE", template.Conventions.Constants.Pattern)
}

func TestTestingPatterns(t *testing.T) {
	// Test React testing
	react := reactTemplate()
	assert.Equal(t, "Vitest + React Testing Library", react.Testing.Framework)
	assert.Equal(t, "tests/unit/", react.Testing.Location)
	assert.NotEmpty(t, react.Testing.Utilities)

	// Test Go testing
	goChi := goChiTemplate()
	assert.Equal(t, "testing package + testify", goChi.Testing.Framework)
	assert.Contains(t, goChi.Testing.Utilities, "github.com/stretchr/testify")

	// Test Python testing
	python := pythonFastAPITemplate()
	assert.Equal(t, "pytest + pytest-asyncio", python.Testing.Framework)
	assert.Equal(t, "tests/", python.Testing.Location)
}
