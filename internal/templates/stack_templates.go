package templates

// reactTemplate returns the React + TypeScript template.
func reactTemplate() *Template {
	return &Template{
		Name:        "React + TypeScript",
		Stack:       StackReact,
		Description: "React 18+ con TypeScript, Vite, y patrones modernos de arquitectura",
		Skills: []TemplateSkill{
			{
				Name:        "sdd-init",
				Path:        "~/.config/opencode/skills/sdd-init/SKILL.md",
				Required:    false,
				Description: "Initialize Spec-Driven Development context",
				Aliases:     []string{"sdd init", "openspec init"},
			},
			{
				Name:        "sdd-spec",
				Path:        "~/.config/opencode/skills/sdd-spec/SKILL.md",
				Required:    false,
				Description: "Write specifications with requirements and scenarios",
				Aliases:     []string{"sdd spec", "specs"},
			},
			{
				Name:        "sdd-design",
				Path:        "~/.config/opencode/skills/sdd-design/SKILL.md",
				Required:    false,
				Description: "Create technical design document",
				Aliases:     []string{"sdd design", "architecture"},
			},
			{
				Name:        "sdd-tasks",
				Path:        "~/.config/opencode/skills/sdd-tasks/SKILL.md",
				Required:    false,
				Description: "Break down change into implementation tasks",
				Aliases:     []string{"sdd tasks", "task breakdown"},
			},
			{
				Name:        "sdd-apply",
				Path:        "~/.config/opencode/skills/sdd-apply/SKILL.md",
				Required:    false,
				Description: "Implement tasks from change",
				Aliases:     []string{"sdd apply", "implement"},
			},
		},
		Folders: []FolderConfig{
			{Path: "src", Purpose: "Application source code"},
			{Path: "src/components", Purpose: "Reusable UI components"},
			{Path: "src/components/ui", Purpose: "Base UI primitives (Button, Input, Card)"},
			{Path: "src/components/features", Purpose: "Feature-specific components"},
			{Path: "src/components/layout", Purpose: "Layout components (Header, Sidebar, Footer)"},
			{Path: "src/hooks", Purpose: "Custom React hooks"},
			{Path: "src/contexts", Purpose: "React context providers"},
			{Path: "src/services", Purpose: "API clients and external services"},
			{Path: "src/utils", Purpose: "Utility functions"},
			{Path: "src/types", Purpose: "TypeScript type definitions"},
			{Path: "src/pages", Purpose: "Page components (routing destinations)"},
			{Path: "src/assets", Purpose: "Static assets (images, fonts, icons)"},
			{Path: "src/styles", Purpose: "Global styles and CSS variables"},
			{Path: "tests", Purpose: "Test files"},
			{Path: "tests/unit", Purpose: "Unit tests"},
			{Path: "tests/integration", Purpose: "Integration tests"},
			{Path: "tests/e2e", Purpose: "End-to-end tests"},
			{Path: "docs", Purpose: "Project documentation"},
			{Path: ".agent", Purpose: "Agent configuration and skills"},
		},
		Files: []FileConfig{
			{
				Path:         "AGENTS.md",
				TemplateName: "agents.md",
				SkipIfExists: true,
			},
			{
				Path:         "src/types/index.ts",
				Content:      "// Core types for the application\n\nexport interface AppConfig {\n  apiUrl: string;\n  environment: 'development' | 'staging' | 'production';\n}\n\nexport interface ApiResponse<T> {\n  data: T;\n  message?: string;\n  status: number;\n}\n\nexport type LoadingState = 'idle' | 'loading' | 'succeeded' | 'failed';\n",
				SkipIfExists: true,
			},
			{
				Path:         "tsconfig.json",
				Content:      "{\n  \"compilerOptions\": {\n    \"target\": \"ES2020\",\n    \"useDefineForClassFields\": true,\n    \"lib\": [\"ES2020\", \"DOM\", \"DOM.Iterable\"],\n    \"module\": \"ESNext\",\n    \"skipLibCheck\": true,\n    \"moduleResolution\": \"bundler\",\n    \"allowImportingTsExtensions\": true,\n    \"resolveJsonModule\": true,\n    \"isolatedModules\": true,\n    \"noEmit\": true,\n    \"jsx\": \"react-jsx\",\n    \"strict\": true,\n    \"noUnusedLocals\": true,\n    \"noUnusedParameters\": true,\n    \"noFallthroughCasesInSwitch\": true,\n    \"baseUrl\": \".\",\n    \"paths\": {\n      \"@/*\": [\"src/*\"],\n      \"@components/*\": [\"src/components/*\"],\n      \"@hooks/*\": [\"src/hooks/*\"],\n      \"@services/*\": [\"src/services/*\"],\n      \"@utils/*\": [\"src/utils/*\"],\n      \"@types/*\": [\"src/types/*\"]\n    }\n  },\n  \"include\": [\"src\"],\n  \"references\": [{ \"path\": \"./tsconfig.node.json\" }]\n}\n",
				SkipIfExists: true,
			},
			{
				Path:         ".env.example",
				Content:      "# API Configuration\nVITE_API_URL=http://localhost:3000/api\nVITE_APP_ENV=development\n\n# Feature Flags\nVITE_ENABLE_ANALYTICS=false\nVITE_ENABLE_DEBUG=false\n",
				SkipIfExists: true,
			},
		},
		Conventions: NamingConventions{
			Files: NamingRule{
				Pattern:     "kebab-case",
				Example:     "user-profile.tsx",
				Description: "Componentes y archivos en kebab-case",
			},
			Components: NamingRule{
				Pattern:     "PascalCase",
				Example:     "UserProfile, ProductCard",
				Description: "Componentes React en PascalCase",
			},
			Functions: NamingRule{
				Pattern:     "camelCase",
				Example:     "fetchUserData, calculateTotal",
				Description: "Funciones y hooks en camelCase",
			},
			Types: NamingRule{
				Pattern:     "PascalCase + Sufijo",
				Example:     "UserData, ApiResponse, LoadingState",
				Description: "Types e interfaces en PascalCase",
			},
			Constants: NamingRule{
				Pattern:     "SCREAMING_SNAKE_CASE",
				Example:     "MAX_RETRY_COUNT, API_TIMEOUT",
				Description: "Constantes globales en SCREAMING_SNAKE_CASE",
			},
			Tests: NamingRule{
				Pattern:     "nombre.test.ts(x)",
				Example:     "userProfile.test.tsx, calculateTotal.spec.ts",
				Description: "Tests con .test.ts(x) o .spec.ts(x)",
			},
		},
		Testing: TestingPatterns{
			Framework:    "Vitest + React Testing Library",
			Location:     "tests/unit/",
			Naming:       "{nombre}.test.ts(x)",
			SetupFile:    "tests/setup.ts",
			Utilities:    []string{"@testing-library/react", "@testing-library/jest-dom", "vitest"},
			CoverageTool: "vitest coverage",
		},
	}
}

// vueTemplate returns the Vue + TypeScript template.
func vueTemplate() *Template {
	return &Template{
		Name:        "Vue + TypeScript",
		Stack:       StackVue,
		Description: "Vue 3 con Composition API, TypeScript, y Pinia",
		Skills: []TemplateSkill{
			{
				Name:        "sdd-init",
				Path:        "~/.config/opencode/skills/sdd-init/SKILL.md",
				Required:    false,
				Description: "Initialize Spec-Driven Development context",
			},
			{
				Name:        "sdd-spec",
				Path:        "~/.config/opencode/skills/sdd-spec/SKILL.md",
				Required:    false,
				Description: "Write specifications with requirements and scenarios",
			},
			{
				Name:        "sdd-design",
				Path:        "~/.config/opencode/skills/sdd-design/SKILL.md",
				Required:    false,
				Description: "Create technical design document",
			},
		},
		Folders: []FolderConfig{
			{Path: "src", Purpose: "Application source code"},
			{Path: "src/components", Purpose: "Reusable Vue components"},
			{Path: "src/components/ui", Purpose: "Base UI primitives"},
			{Path: "src/components/features", Purpose: "Feature-specific components"},
			{Path: "src/components/layout", Purpose: "Layout components"},
			{Path: "src/composables", Purpose: "Vue composables (composition API hooks)"},
			{Path: "src/stores", Purpose: "Pinia stores"},
			{Path: "src/types", Purpose: "TypeScript type definitions"},
			{Path: "src/services", Purpose: "API clients"},
			{Path: "src/utils", Purpose: "Utility functions"},
			{Path: "src/views", Purpose: "Page views"},
			{Path: "src/router", Purpose: "Vue Router configuration"},
			{Path: "src/assets", Purpose: "Static assets"},
			{Path: "tests", Purpose: "Test files"},
			{Path: "tests/unit", Purpose: "Unit tests with Vitest"},
			{Path: "tests/e2e", Purpose: "E2E tests with Playwright"},
			{Path: "docs", Purpose: "Project documentation"},
		},
		Files: []FileConfig{
			{
				Path:         "AGENTS.md",
				TemplateName: "agents.md",
				SkipIfExists: true,
			},
			{
				Path:         "src/types/index.ts",
				Content:      "// Core types for the application\n\nimport type { Ref } from 'vue';\n\nexport interface AppConfig {\n  apiUrl: string;\n  environment: 'development' | 'staging' | 'production';\n}\n\nexport interface ApiResponse<T> {\n  data: T;\n  message?: string;\n  status: number;\n}\n\nexport type MaybeRef<T> = T | Ref<T>;\n\nexport type LoadingState = 'idle' | 'loading' | 'succeeded' | 'failed';\n",
				SkipIfExists: true,
			},
		},
		Conventions: NamingConventions{
			Files: NamingRule{
				Pattern:     "PascalCase (components) / kebab-case (other)",
				Example:     "UserProfile.vue, api-client.ts",
				Description: "Componentes en PascalCase, otros archivos en kebab-case",
			},
			Components: NamingRule{
				Pattern:     "PascalCase",
				Example:     "UserProfile, ProductCard",
				Description: "Componentes Vue en PascalCase",
			},
			Functions: NamingRule{
				Pattern:     "camelCase",
				Example:     "useUserData, fetchProducts",
				Description: "Composables y funciones en camelCase",
			},
			Types: NamingRule{
				Pattern:     "PascalCase",
				Example:     "UserData, ProductItem",
				Description: "Types e interfaces en PascalCase",
			},
			Constants: NamingRule{
				Pattern:     "SCREAMING_SNAKE_CASE",
				Example:     "MAX_ITEMS, API_BASE_URL",
				Description: "Constantes en SCREAMING_SNAKE_CASE",
			},
			Tests: NamingRule{
				Pattern:     "nombre.spec.ts",
				Example:     "useUserData.spec.ts, UserProfile.spec.ts",
				Description: "Tests con .spec.ts",
			},
		},
		Testing: TestingPatterns{
			Framework:    "Vitest + Vue Test Utils",
			Location:     "tests/unit/",
			Naming:       "{nombre}.spec.ts",
			SetupFile:    "tests/setup.ts",
			Utilities:    []string{"@vue/test-utils", "@testing-library/vue", "vitest"},
			CoverageTool: "vitest coverage",
		},
	}
}

// angularTemplate returns the Angular + TypeScript template.
func angularTemplate() *Template {
	return &Template{
		Name:        "Angular + TypeScript",
		Stack:       StackAngular,
		Description: "Angular 17+ con standalone components, signals, y RxJS",
		Skills: []TemplateSkill{
			{
				Name:        "sdd-init",
				Path:        "~/.config/opencode/skills/sdd-init/SKILL.md",
				Required:    false,
				Description: "Initialize Spec-Driven Development context",
			},
			{
				Name:        "sdd-spec",
				Path:        "~/.config/opencode/skills/sdd-spec/SKILL.md",
				Required:    false,
				Description: "Write specifications with requirements and scenarios",
			},
		},
		Folders: []FolderConfig{
			{Path: "src/app", Purpose: "Application source code"},
			{Path: "src/app/core", Purpose: "Singleton services, guards, interceptors"},
			{Path: "src/app/shared", Purpose: "Shared components, directives, pipes"},
			{Path: "src/app/features", Purpose: "Feature modules"},
			{Path: "src/app/layout", Purpose: "Layout components"},
			{Path: "src/app/models", Purpose: "TypeScript interfaces and types"},
			{Path: "src/app/services", Purpose: "API services"},
			{Path: "src/app/store", Purpose: "State management (NgRx or signals)"},
			{Path: "src/assets", Purpose: "Static assets"},
			{Path: "src/styles", Purpose: "Global styles"},
			{Path: "e2e", Purpose: "E2E tests"},
			{Path: "docs", Purpose: "Project documentation"},
		},
		Files: []FileConfig{
			{
				Path:         "AGENTS.md",
				TemplateName: "agents.md",
				SkipIfExists: true,
			},
		},
		Conventions: NamingConventions{
			Files: NamingRule{
				Pattern:     "kebab-case",
				Example:     "user-profile.component.ts",
				Description: "Archivos en kebab-case",
			},
			Components: NamingRule{
				Pattern:     "PascalCase + sufijo",
				Example:     "UserProfileComponent, ProductCardComponent",
				Description: "Componentes en PascalCase con sufijo",
			},
			Functions: NamingRule{
				Pattern:     "camelCase",
				Example:     "fetchUserData, calculateTotal",
				Description: "Metodos y funciones en camelCase",
			},
			Types: NamingRule{
				Pattern:     "PascalCase",
				Example:     "User, Product, ApiResponse",
				Description: "Interfaces y tipos en PascalCase",
			},
			Constants: NamingRule{
				Pattern:     "SCREAMING_SNAKE_CASE",
				Example:     "MAX_RETRY_COUNT, API_TIMEOUT",
				Description: "Constantes en SCREAMING_SNAKE_CASE",
			},
			Tests: NamingRule{
				Pattern:     "nombre.spec.ts",
				Example:     "user-profile.component.spec.ts",
				Description: "Tests con .spec.ts",
			},
		},
		Testing: TestingPatterns{
			Framework:    "Jasmine + Karma / Jest",
			Location:     "src/app/**/*.spec.ts",
			Naming:       "{nombre}.spec.ts",
			Utilities:    []string{"@testing-library/angular", "jasmine", "karma"},
			CoverageTool: "jest --coverage",
		},
	}
}

// nextJSTemplate returns the Next.js template.
func nextJSTemplate() *Template {
	return &Template{
		Name:        "Next.js",
		Stack:       StackNextJS,
		Description: "Next.js 14+ con App Router, Server Components, y TypeScript",
		Skills: []TemplateSkill{
			{
				Name:        "sdd-init",
				Path:        "~/.config/opencode/skills/sdd-init/SKILL.md",
				Required:    false,
				Description: "Initialize Spec-Driven Development context",
			},
			{
				Name:        "sdd-spec",
				Path:        "~/.config/opencode/skills/sdd-spec/SKILL.md",
				Required:    false,
				Description: "Write specifications with requirements and scenarios",
			},
			{
				Name:        "sdd-design",
				Path:        "~/.config/opencode/skills/sdd-design/SKILL.md",
				Required:    false,
				Description: "Create technical design document",
			},
		},
		Folders: []FolderConfig{
			{Path: "src/app", Purpose: "App Router pages and layouts"},
			{Path: "src/app/(routes)", Purpose: "Route groups"},
			{Path: "src/components", Purpose: "React components"},
			{Path: "src/components/ui", Purpose: "Base UI primitives"},
			{Path: "src/components/features", Purpose: "Feature-specific components"},
			{Path: "src/components/layout", Purpose: "Layout components"},
			{Path: "src/lib", Purpose: "Utilities and external integrations"},
			{Path: "src/hooks", Purpose: "Custom React hooks"},
			{Path: "src/types", Purpose: "TypeScript type definitions"},
			{Path: "src/services", Purpose: "API clients"},
			{Path: "src/actions", Purpose: "Server actions"},
			{Path: "src/store", Purpose: "State management"},
			{Path: "src/styles", Purpose: "Global styles"},
			{Path: "public", Purpose: "Static public assets"},
			{Path: "tests", Purpose: "Test files"},
			{Path: "tests/unit", Purpose: "Unit tests"},
			{Path: "tests/e2e", Purpose: "E2E tests with Playwright"},
			{Path: "docs", Purpose: "Project documentation"},
		},
		Files: []FileConfig{
			{
				Path:         "AGENTS.md",
				TemplateName: "agents.md",
				SkipIfExists: true,
			},
			{
				Path:         "src/types/index.ts",
				Content:      "// Core types for the Next.js application\n\nexport interface AppConfig {\n  apiUrl: string;\n  environment: 'development' | 'staging' | 'production';\n}\n\nexport interface ApiResponse<T> {\n  data: T;\n  message?: string;\n  status: number;\n}\n\n// Page props types for App Router\nexport interface PageProps {\n  params: Record<string, string>;\n  searchParams: Record<string, string | string[] | undefined>;\n}\n",
				SkipIfExists: true,
			},
		},
		Conventions: NamingConventions{
			Files: NamingRule{
				Pattern:     "kebab-case",
				Example:     "user-profile.tsx",
				Description: "Archivos en kebab-case",
			},
			Components: NamingRule{
				Pattern:     "PascalCase",
				Example:     "UserProfile, ProductCard",
				Description: "Componentes en PascalCase",
			},
			Functions: NamingRule{
				Pattern:     "camelCase",
				Example:     "fetchUserData, getServerData",
				Description: "Funciones en camelCase",
			},
			Types: NamingRule{
				Pattern:     "PascalCase",
				Example:     "User, Product, ApiResponse",
				Description: "Types en PascalCase",
			},
			Constants: NamingRule{
				Pattern:     "SCREAMING_SNAKE_CASE",
				Example:     "MAX_ITEMS, API_TIMEOUT",
				Description: "Constantes en SCREAMING_SNAKE_CASE",
			},
			Tests: NamingRule{
				Pattern:     "nombre.test.ts(x) / nombre.spec.ts",
				Example:     "user-profile.test.tsx",
				Description: "Tests con .test.ts(x) o .spec.ts",
			},
		},
		Testing: TestingPatterns{
			Framework:    "Vitest + React Testing Library / Playwright",
			Location:     "tests/",
			Naming:       "{nombre}.test.ts(x)",
			SetupFile:    "tests/setup.ts",
			Utilities:    []string{"@testing-library/react", "@testing-library/jest-dom", "vitest", "playwright"},
			CoverageTool: "vitest coverage",
		},
	}
}

// goChiTemplate returns the Go + Chi template.
func goChiTemplate() *Template {
	return &Template{
		Name:        "Go + Chi",
		Stack:       StackGoChi,
		Description: "Go con Chi router, arquitectura limpia, y testing",
		Skills: []TemplateSkill{
			{
				Name:        "go-testing",
				Path:        "~/.config/opencode/skills/go-testing/SKILL.md",
				Required:    true,
				Description: "Go testing patterns including Bubbletea TUI testing",
			},
			{
				Name:        "sdd-init",
				Path:        "~/.config/opencode/skills/sdd-init/SKILL.md",
				Required:    false,
				Description: "Initialize Spec-Driven Development context",
			},
			{
				Name:        "sdd-spec",
				Path:        "~/.config/opencode/skills/sdd-spec/SKILL.md",
				Required:    false,
				Description: "Write specifications with requirements and scenarios",
			},
		},
		Folders: []FolderConfig{
			{Path: "cmd/server", Purpose: "Main application entry point"},
			{Path: "internal/handlers", Purpose: "HTTP handlers"},
			{Path: "internal/middleware", Purpose: "Middleware (auth, logging, etc.)"},
			{Path: "internal/models", Purpose: "Domain models and entities"},
			{Path: "internal/repository", Purpose: "Data access layer"},
			{Path: "internal/service", Purpose: "Business logic layer"},
			{Path: "internal/router", Purpose: "Router configuration"},
			{Path: "pkg/errors", Purpose: "Shared error types"},
			{Path: "pkg/logger", Purpose: "Logging utilities"},
			{Path: "pkg/validator", Purpose: "Input validation"},
			{Path: "config", Purpose: "Configuration files"},
			{Path: "migrations", Purpose: "Database migrations"},
			{Path: "tests", Purpose: "Integration tests"},
			{Path: "docs", Purpose: "API documentation"},
			{Path: "scripts", Purpose: "Build and deployment scripts"},
		},
		Files: []FileConfig{
			{
				Path:         "AGENTS.md",
				TemplateName: "agents.md",
				SkipIfExists: true,
			},
			{
				Path:         "cmd/server/main.go",
				Content:      "package main\n\nimport (\n\t\"context\"\n\t\"log\"\n\t\"net/http\"\n\t\"os\"\n\t\"os/signal\"\n\t\"syscall\"\n\t\"time\"\n\n\t\"{{ .Module }}/internal/handlers\"\n\t\"{{ .Module }}/internal/middleware\"\n\t\"{{ .Module }}/internal/router\"\n\n\t\"github.com/go-chi/chi/v5\"\n)\n\nfunc main() {\n\t// Initialize router\n\tr := router.New()\n\n\t// Setup middleware\n\tr.Use(middleware.Logger())\n\tr.Use(middleware.Recover())\n\tr.Use(middleware.CORS())\n\n\t// Setup routes\n\thandlers.SetupRoutes(r)\n\n\t// Create server\n\tsrv := &http.Server{\n\t\tAddr:         \":8080\",\n\t\tHandler:      r,\n\t\tReadTimeout:  15 * time.Second,\n\t\tWriteTimeout: 15 * time.Second,\n\t\tIdleTimeout:  60 * time.Second,\n\t}\n\n\t// Start server in goroutine\n\tgo func() {\n\t\tlog.Printf(\"Server starting on %s\", srv.Addr)\n\t\tif err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {\n\t\t\tlog.Fatalf(\"Server error: %v\", err)\n\t\t}\n\t}()\n\n\t// Graceful shutdown\n\tquit := make(chan os.Signal, 1)\n\tsignal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)\n\t<-quit\n\n\tctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)\n\tdefer cancel()\n\n\tlog.Println(\"Shutting down server...\")\n\tif err := srv.Shutdown(ctx); err != nil {\n\t\tlog.Fatalf(\"Server forced to shutdown: %v\", err)\n\t}\n\n\tlog.Println(\"Server stopped\")\n}\n",
				SkipIfExists: true,
			},
			{
				Path:         "go.mod",
				Content:      "module {{ .Module }}\n\ngo 1.23\n\nrequire (\n\tgithub.com/go-chi/chi/v5 v5.1.0\n\tgithub.com/go-chi/cors v1.2.1\n\tgithub.com/jackc/pgx/v5 v5.5.0\n\tgithub.com/stretchr/testify v1.8.4\n)\n\nrequire (\n\tgithub.com/davecgh/go-spew v1.1.1 // indirect\n\tgithub.com/jackc/pgpassfile v1.0.0 // indirect\n\tgithub.com/jackc/pgservicefile v0.0.0-20231201235250-de7065d80cb9 // indirect\n\tgithub.com/jackc/puddle/v2 v2.2.1 // indirect\n\tgithub.com/pmezard/go-difflib v1.0.0 // indirect\n\tgolang.org/x/crypto v0.21.0 // indirect\n\tgolang.org/x/sync v0.6.0 // indirect\n\tgolang.org/x/text v0.14.0 // indirect\n)\n",
				SkipIfExists: true,
			},
			{
				Path:         ".env.example",
				Content:      "# Server Configuration\nSERVER_PORT=8080\nSERVER_READ_TIMEOUT=15s\nSERVER_WRITE_TIMEOUT=15s\n\n# Database\nDATABASE_URL=postgres://user:password@localhost:5432/dbname?sslmode=disable\n\n# Environment\nENVIRONMENT=development\n\n# CORS\nCORS_ALLOWED_ORIGINS=http://localhost:3000\n",
				SkipIfExists: true,
			},
		},
		Conventions: NamingConventions{
			Files: NamingRule{
				Pattern:     "snake_case",
				Example:     "user_repository.go, auth_handler.go",
				Description: "Archivos en snake_case",
			},
			Components: NamingRule{
				Pattern:     "PascalCase",
				Example:     "UserHandler, AuthMiddleware",
				Description: "Tipos y funciones exportadas en PascalCase",
			},
			Functions: NamingRule{
				Pattern:     "camelCase",
				Example:     "getUserByID, validateToken",
				Description: "Funciones privadas en camelCase",
			},
			Types: NamingRule{
				Pattern:     "PascalCase",
				Example:     "User, Product, ApiResponse",
				Description: "Structs e interfaces en PascalCase",
			},
			Constants: NamingRule{
				Pattern:     "PascalCase",
				Example:     "MaxRetries, DefaultTimeout",
				Description: "Constantes exportadas en PascalCase",
			},
			Tests: NamingRule{
				Pattern:     "nombre_test.go",
				Example:     "user_repository_test.go, handler_test.go",
				Description: "Tests con _test.go suffix",
			},
		},
		Testing: TestingPatterns{
			Framework:    "testing package + testify",
			Location:     "internal/**/*_test.go",
			Naming:       "{package}_test.go",
			SetupFile:    "tests/setup.go",
			Utilities:    []string{"github.com/stretchr/testify", "github.com/golang/mock"},
			CoverageTool: "go test -coverprofile=coverage.out && go tool cover",
		},
	}
}

// goEchoTemplate returns the Go + Echo template.
func goEchoTemplate() *Template {
	return &Template{
		Name:        "Go + Echo",
		Stack:       StackGoEcho,
		Description: "Go con Echo framework, arquitectura limpia, y testing",
		Skills: []TemplateSkill{
			{
				Name:        "go-testing",
				Path:        "~/.config/opencode/skills/go-testing/SKILL.md",
				Required:    true,
				Description: "Go testing patterns",
			},
		},
		Folders: []FolderConfig{
			{Path: "cmd/server", Purpose: "Main application entry point"},
			{Path: "internal/handlers", Purpose: "HTTP handlers"},
			{Path: "internal/middleware", Purpose: "Middleware"},
			{Path: "internal/models", Purpose: "Domain models"},
			{Path: "internal/repository", Purpose: "Data access layer"},
			{Path: "internal/service", Purpose: "Business logic"},
			{Path: "pkg/errors", Purpose: "Shared errors"},
			{Path: "pkg/logger", Purpose: "Logging"},
			{Path: "config", Purpose: "Configuration"},
			{Path: "migrations", Purpose: "Database migrations"},
			{Path: "tests", Purpose: "Integration tests"},
		},
		Files: []FileConfig{
			{
				Path:         "AGENTS.md",
				TemplateName: "agents.md",
				SkipIfExists: true,
			},
			{
				Path:         "cmd/server/main.go",
				Content:      "package main\n\nimport (\n\t\"context\"\n\t\"log\"\n\t\"net/http\"\n\t\"os\"\n\t\"os/signal\"\n\t\"time\"\n\n\t\"{{ .Module }}/internal/handlers\"\n\t\"{{ .Module }}/internal/middleware\"\n\n\t\"github.com/labstack/echo/v4\"\n\t\"github.com/labstack/echo/v4/middleware\"\n)\n\nfunc main() {\n\te := echo.New()\n\te.HideBanner = true\n\n\t// Middleware\n\te.Use(middleware.Logger())\n\te.Use(middleware.Recover())\n\te.Use(middleware.CORS())\n\te.Use(middleware.RequestID())\n\n\t// Routes\n\thandlers.SetupRoutes(e)\n\n\t// Start server\n\tgo func() {\n\t\tif err := e.Start(\":8080\"); err != nil && err != http.ErrServerClosed {\n\t\t\te.Logger.Fatal(\"shutting down the server\")\n\t\t}\n\t}()\n\n\t// Graceful shutdown\n\tquit := make(chan os.Signal, 1)\n\tsignal.Notify(quit, os.Interrupt)\n\t<-quit\n\n\tctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)\n\tdefer cancel()\n\n\tif err := e.Shutdown(ctx); err != nil {\n\t\tlog.Fatal(err)\n\t}\n}\n",
				SkipIfExists: true,
			},
		},
		Conventions: NamingConventions{
			Files:      NamingRule{Pattern: "snake_case", Example: "user_handler.go"},
			Components: NamingRule{Pattern: "PascalCase", Example: "UserHandler"},
			Functions:  NamingRule{Pattern: "camelCase", Example: "getUserByID"},
			Types:      NamingRule{Pattern: "PascalCase", Example: "User, ApiResponse"},
			Constants:  NamingRule{Pattern: "PascalCase", Example: "MaxRetries"},
			Tests:      NamingRule{Pattern: "nombre_test.go", Example: "handler_test.go"},
		},
		Testing: TestingPatterns{
			Framework:    "testing package + testify",
			Location:     "internal/**/*_test.go",
			Naming:       "{package}_test.go",
			Utilities:    []string{"github.com/stretchr/testify", "github.com/labstack/echo/v4"},
			CoverageTool: "go test -coverprofile=coverage.out",
		},
	}
}

// goGinTemplate returns the Go + Gin template.
func goGinTemplate() *Template {
	return &Template{
		Name:        "Go + Gin",
		Stack:       StackGoGin,
		Description: "Go con Gin framework, arquitectura limpia, y testing",
		Skills: []TemplateSkill{
			{
				Name:        "go-testing",
				Path:        "~/.config/opencode/skills/go-testing/SKILL.md",
				Required:    true,
				Description: "Go testing patterns",
			},
		},
		Folders: []FolderConfig{
			{Path: "cmd/server", Purpose: "Main application entry point"},
			{Path: "internal/handlers", Purpose: "HTTP handlers"},
			{Path: "internal/middleware", Purpose: "Middleware"},
			{Path: "internal/models", Purpose: "Domain models"},
			{Path: "internal/repository", Purpose: "Data access layer"},
			{Path: "internal/service", Purpose: "Business logic"},
			{Path: "pkg/errors", Purpose: "Shared errors"},
			{Path: "pkg/logger", Purpose: "Logging"},
			{Path: "config", Purpose: "Configuration"},
			{Path: "migrations", Purpose: "Database migrations"},
			{Path: "tests", Purpose: "Integration tests"},
		},
		Files: []FileConfig{
			{
				Path:         "AGENTS.md",
				TemplateName: "agents.md",
				SkipIfExists: true,
			},
			{
				Path:         "cmd/server/main.go",
				Content:      "package main\n\nimport (\n\t\"context\"\n\t\"log\"\n\t\"net/http\"\n\t\"os\"\n\t\"os/signal\"\n\t\"syscall\"\n\t\"time\"\n\n\t\"{{ .Module }}/internal/handlers\"\n\t\"{{ .Module }}/internal/middleware\"\n\n\t\"github.com/gin-gonic/gin\"\n)\n\nfunc main() {\n\t// Set Gin mode\n\tif os.Getenv(\"ENVIRONMENT\") == \"production\" {\n\t\tgin.SetMode(gin.ReleaseMode)\n\t}\n\n\tr := gin.New()\n\n\t// Middleware\n\tr.Use(gin.Logger())\n\tr.Use(gin.Recovery())\n\tr.Use(middleware.CORS())\n\n\t// Routes\n\thandlers.SetupRoutes(r)\n\n\t// Create server\n\tsrv := &http.Server{\n\t\tAddr:    \":8080\",\n\t\tHandler: r,\n\t}\n\n\t// Start server\n\tgo func() {\n\t\tlog.Printf(\"Server starting on %s\", srv.Addr)\n\t\tif err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {\n\t\t\tlog.Fatalf(\"Server error: %v\", err)\n\t\t}\n\t}()\n\n\t// Graceful shutdown\n\tquit := make(chan os.Signal, 1)\n\tsignal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)\n\t<-quit\n\n\tlog.Println(\"Shutting down server...\")\n\tctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)\n\tdefer cancel()\n\n\tif err := srv.Shutdown(ctx); err != nil {\n\t\tlog.Fatalf(\"Server forced to shutdown: %v\", err)\n\t}\n\n\tlog.Println(\"Server stopped\")\n}\n",
				SkipIfExists: true,
			},
		},
		Conventions: NamingConventions{
			Files:      NamingRule{Pattern: "snake_case", Example: "user_handler.go"},
			Components: NamingRule{Pattern: "PascalCase", Example: "UserHandler"},
			Functions:  NamingRule{Pattern: "camelCase", Example: "getUserByID"},
			Types:      NamingRule{Pattern: "PascalCase", Example: "User, ApiResponse"},
			Constants:  NamingRule{Pattern: "PascalCase", Example: "MaxRetries"},
			Tests:      NamingRule{Pattern: "nombre_test.go", Example: "handler_test.go"},
		},
		Testing: TestingPatterns{
			Framework:    "testing package + testify/gin",
			Location:     "internal/**/*_test.go",
			Naming:       "{package}_test.go",
			Utilities:    []string{"github.com/stretchr/testify", "github.com/gin-gonic/gin"},
			CoverageTool: "go test -coverprofile=coverage.out",
		},
	}
}

// pythonFastAPITemplate returns the Python + FastAPI template.
func pythonFastAPITemplate() *Template {
	return &Template{
		Name:        "Python + FastAPI",
		Stack:       StackPythonFastAPI,
		Description: "Python con FastAPI, Pydantic, SQLAlchemy, y testing con pytest",
		Skills: []TemplateSkill{
			{
				Name:        "sdd-init",
				Path:        "~/.config/opencode/skills/sdd-init/SKILL.md",
				Required:    false,
				Description: "Initialize Spec-Driven Development context",
			},
			{
				Name:        "sdd-spec",
				Path:        "~/.config/opencode/skills/sdd-spec/SKILL.md",
				Required:    false,
				Description: "Write specifications with requirements and scenarios",
			},
		},
		Folders: []FolderConfig{
			{Path: "app", Purpose: "Application source code"},
			{Path: "app/api", Purpose: "API routes and endpoints"},
			{Path: "app/api/v1", Purpose: "API v1 routes"},
			{Path: "app/core", Purpose: "Core configuration and security"},
			{Path: "app/models", Purpose: "SQLAlchemy models"},
			{Path: "app/schemas", Purpose: "Pydantic schemas"},
			{Path: "app/services", Purpose: "Business logic services"},
			{Path: "app/repositories", Purpose: "Data access layer"},
			{Path: "app/db", Purpose: "Database connection and session"},
			{Path: "app/middleware", Purpose: "Custom middleware"},
			{Path: "app/utils", Purpose: "Utility functions"},
			{Path: "tests", Purpose: "Test files"},
			{Path: "tests/unit", Purpose: "Unit tests"},
			{Path: "tests/integration", Purpose: "Integration tests"},
			{Path: "tests/fixtures", Purpose: "Test fixtures"},
			{Path: "alembic", Purpose: "Database migrations"},
			{Path: "docs", Purpose: "API documentation"},
		},
		Files: []FileConfig{
			{
				Path:         "AGENTS.md",
				TemplateName: "agents.md",
				SkipIfExists: true,
			},
			{
				Path:         "pyproject.toml",
				Content:      "[project]\nname = \"{{ .ProjectName }}\"\nversion = \"0.1.0\"\ndescription = \"{{ .Description }}\"\nrequires-python = \">=3.11\"\n\n[tool.poetry.dependencies]\npython = \"^3.11\"\nfastapi = \"^0.110.0\"\nuvicorn = { extras = [\"standard\"], version = \"^0.27.0\" }\npydantic = \"^2.6.0\"\npydantic-settings = \"^2.2.0\"\nsqlalchemy = \"^2.0.0\"\nalembic = \"^1.13.0\"\npsycopg2-binary = \"^2.9.9\"\npython-jose = { extras = [\"cryptography\"], version = \"^3.3.0\" }\npasslib = { extras = [\"bcrypt\"], version = \"^1.7.4\" }\n\n[tool.poetry.group.dev.dependencies]\npytest = \"^8.0.0\"\npytest-asyncio = \"^0.23.0\"\npytest-cov = \"^4.1.0\"\nhttpx = \"^0.27.0\"\n\n[tool.pytest.ini_options]\nasyncio_mode = \"auto\"\ntestpaths = [\"tests\"]\n\n[build-system]\nrequires = [\"poetry-core\"]\nbuild-backend = \"poetry.core.masonry.api\"\n",
				SkipIfExists: true,
			},
			{
				Path:         ".env.example",
				Content:      "# Application\nAPP_ENV=development\nDEBUG=true\nSECRET_KEY=your-secret-key-here\n\n# Database\nDATABASE_URL=postgresql://user:password@localhost:5432/dbname\n\n# CORS\nCORS_ORIGINS=http://localhost:3000\n\n# JWT\nJWT_ALGORITHM=HS256\nACCESS_TOKEN_EXPIRE_MINUTES=30\n",
				SkipIfExists: true,
			},
			{
				Path:         "app/schemas/__init__.py",
				Content:      "# Schemas package\nfrom app.schemas.base import BaseSchema\nfrom app.schemas.user import UserCreate, UserUpdate, UserResponse\n\n__all__ = [\n    \"BaseSchema\",\n    \"UserCreate\",\n    \"UserUpdate\",\n    \"UserResponse\",\n]\n",
				SkipIfExists: true,
			},
		},
		Conventions: NamingConventions{
			Files: NamingRule{
				Pattern:     "snake_case",
				Example:     "user_repository.py, auth_handler.py",
				Description: "Archivos Python en snake_case",
			},
			Components: NamingRule{
				Pattern:     "PascalCase",
				Example:     "UserRepository, AuthHandler",
				Description: "Clases en PascalCase",
			},
			Functions: NamingRule{
				Pattern:     "snake_case",
				Example:     "get_user_by_id, validate_token",
				Description: "Funciones en snake_case",
			},
			Types: NamingRule{
				Pattern:     "PascalCase",
				Example:     "UserSchema, TokenPayload",
				Description: "Schemas y tipos en PascalCase",
			},
			Constants: NamingRule{
				Pattern:     "SCREAMING_SNAKE_CASE",
				Example:     "MAX_RETRY_COUNT, DEFAULT_TIMEOUT",
				Description: "Constantes en SCREAMING_SNAKE_CASE",
			},
			Tests: NamingRule{
				Pattern:     "test_nombre.py",
				Example:     "test_user_repository.py, test_auth.py",
				Description: "Tests con prefijo test_",
			},
		},
		Testing: TestingPatterns{
			Framework:    "pytest + pytest-asyncio",
			Location:     "tests/",
			Naming:       "test_{module}.py",
			SetupFile:    "tests/conftest.py",
			Utilities:    []string{"pytest", "pytest-asyncio", "httpx", "pytest-cov"},
			CoverageTool: "pytest --cov=app --cov-report=html",
		},
	}
}

// pythonDjangoTemplate returns the Python + Django template.
func pythonDjangoTemplate() *Template {
	return &Template{
		Name:        "Python + Django",
		Stack:       StackPythonDjango,
		Description: "Django con Django REST Framework, Poetry, y testing",
		Skills: []TemplateSkill{
			{
				Name:        "sdd-init",
				Path:        "~/.config/opencode/skills/sdd-init/SKILL.md",
				Required:    false,
				Description: "Initialize Spec-Driven Development context",
			},
		},
		Folders: []FolderConfig{
			{Path: "config", Purpose: "Django project settings"},
			{Path: "apps", Purpose: "Django applications"},
			{Path: "apps/core", Purpose: "Core app (users, etc)"},
			{Path: "apps/api", Purpose: "REST API app"},
			{Path: "apps/api/v1", Purpose: "API v1 endpoints"},
			{Path: "tests", Purpose: "Test files"},
			{Path: "tests/unit", Purpose: "Unit tests"},
			{Path: "tests/integration", Purpose: "Integration tests"},
			{Path: "fixtures", Purpose: "Data fixtures"},
			{Path: "docs", Purpose: "Documentation"},
		},
		Files: []FileConfig{
			{
				Path:         "AGENTS.md",
				TemplateName: "agents.md",
				SkipIfExists: true,
			},
		},
		Conventions: NamingConventions{
			Files: NamingRule{
				Pattern:     "snake_case",
				Example:     "user_repository.py",
				Description: "Archivos en snake_case",
			},
			Components: NamingRule{
				Pattern:     "PascalCase",
				Example:     "UserSerializer, UserViewSet",
				Description: "Clases en PascalCase",
			},
			Functions: NamingRule{
				Pattern:     "snake_case",
				Example:     "get_user_by_id",
				Description: "Funciones en snake_case",
			},
			Types: NamingRule{
				Pattern:     "PascalCase",
				Example:     "UserSerializer",
				Description: "Serializers y modelos en PascalCase",
			},
			Constants: NamingRule{
				Pattern:     "SCREAMING_SNAKE_CASE",
				Example:     "MAX_USERS",
				Description: "Constantes en SCREAMING_SNAKE_CASE",
			},
			Tests: NamingRule{
				Pattern:     "test_nombre.py",
				Example:     "test_views.py",
				Description: "Tests con prefijo test_",
			},
		},
		Testing: TestingPatterns{
			Framework:    "pytest + pytest-django",
			Location:     "tests/",
			Naming:       "test_{module}.py",
			SetupFile:    "tests/conftest.py",
			Utilities:    []string{"pytest", "pytest-django", "factory-boy"},
			CoverageTool: "pytest --cov=apps --cov-report=html",
		},
	}
}

// nodeExpressTemplate returns the Node.js + Express template.
func nodeExpressTemplate() *Template {
	return &Template{
		Name:        "Node.js + Express",
		Stack:       StackNodeExpress,
		Description: "Node.js con Express, TypeScript, y testing",
		Skills: []TemplateSkill{
			{
				Name:        "sdd-init",
				Path:        "~/.config/opencode/skills/sdd-init/SKILL.md",
				Required:    false,
				Description: "Initialize Spec-Driven Development context",
			},
			{
				Name:        "sdd-spec",
				Path:        "~/.config/opencode/skills/sdd-spec/SKILL.md",
				Required:    false,
				Description: "Write specifications with requirements and scenarios",
			},
		},
		Folders: []FolderConfig{
			{Path: "src", Purpose: "Application source code"},
			{Path: "src/routes", Purpose: "Express routes"},
			{Path: "src/routes/v1", Purpose: "API v1 routes"},
			{Path: "src/controllers", Purpose: "Request controllers"},
			{Path: "src/middleware", Purpose: "Express middleware"},
			{Path: "src/services", Purpose: "Business logic services"},
			{Path: "src/repositories", Purpose: "Data access layer"},
			{Path: "src/models", Purpose: "Data models/types"},
			{Path: "src/utils", Purpose: "Utility functions"},
			{Path: "src/config", Purpose: "Configuration"},
			{Path: "src/types", Purpose: "TypeScript types"},
			{Path: "tests", Purpose: "Test files"},
			{Path: "tests/unit", Purpose: "Unit tests"},
			{Path: "tests/integration", Purpose: "Integration tests"},
			{Path: "tests/fixtures", Purpose: "Test fixtures"},
			{Path: "migrations", Purpose: "Database migrations"},
			{Path: "docs", Purpose: "API documentation"},
		},
		Files: []FileConfig{
			{
				Path:         "AGENTS.md",
				TemplateName: "agents.md",
				SkipIfExists: true,
			},
			{
				Path:         "package.json",
				Content:      "{\n  \"name\": \"{{ .ProjectName }}\",\n  \"version\": \"1.0.0\",\n  \"description\": \"{{ .Description }}\",\n  \"main\": \"dist/index.js\",\n  \"scripts\": {\n    \"dev\": \"tsx watch src/index.ts\",\n    \"build\": \"tsc\",\n    \"start\": \"node dist/index.js\",\n    \"test\": \"vitest\",\n    \"test:coverage\": \"vitest run --coverage\",\n    \"lint\": \"eslint src --ext .ts\",\n    \"migrate\": \"sequelize-cli db:migrate\"\n  },\n  \"dependencies\": {\n    \"express\": \"^4.19.0\",\n    \"cors\": \"^2.8.5\",\n    \"helmet\": \"^7.1.0\",\n    \"compression\": \"^1.7.4\",\n    \"dotenv\": \"^16.4.0\",\n    \"zod\": \"^3.22.0\",\n    \"jsonwebtoken\": \"^9.0.2\",\n    \"bcryptjs\": \"^2.4.3\",\n    \"pg\": \"^8.11.0\",\n    \"sequelize\": \"^6.37.0\"\n  },\n  \"devDependencies\": {\n    \"@types/express\": \"^4.17.21\",\n    \"@types/cors\": \"^2.8.17\",\n    \"@types/compression\": \"^1.7.5\",\n    \"@types/node\": \"^20.11.0\",\n    \"@types/jsonwebtoken\": \"^9.0.5\",\n    \"@types/bcryptjs\": \"^2.4.6\",\n    \"typescript\": \"^5.3.0\",\n    \"tsx\": \"^4.7.0\",\n    \"vitest\": \"^1.3.0\",\n    \"@vitest/coverage-v8\": \"^1.3.0\",\n    \"eslint\": \"^8.56.0\",\n    \"@typescript-eslint/eslint-plugin\": \"^7.0.0\",\n    \"@typescript-eslint/parser\": \"^7.0.0\"\n  }\n}\n",
				SkipIfExists: true,
			},
			{
				Path:         "tsconfig.json",
				Content:      "{\n  \"compilerOptions\": {\n    \"target\": \"ES2022\",\n    \"module\": \"NodeNext\",\n    \"moduleResolution\": \"NodeNext\",\n    \"lib\": [\"ES2022\"],\n    \"outDir\": \"./dist\",\n    \"rootDir\": \"./src\",\n    \"strict\": true,\n    \"esModuleInterop\": true,\n    \"skipLibCheck\": true,\n    \"forceConsistentCasingInFileNames\": true,\n    \"resolveJsonModule\": true,\n    \"declaration\": true,\n    \"declarationMap\": true,\n    \"sourceMap\": true,\n    \"baseUrl\": \".\",\n    \"paths\": {\n      \"@/*\": [\"src/*\"],\n      \"@routes/*\": [\"src/routes/*\"],\n      \"@controllers/*\": [\"src/controllers/*\"],\n      \"@services/*\": [\"src/services/*\"],\n      \"@middleware/*\": [\"src/middleware/*\"],\n      \"@utils/*\": [\"src/utils/*\"],\n      \"@config/*\": [\"src/config/*\"],\n      \"@types/*\": [\"src/types/*\"]\n    }\n  },\n  \"include\": [\"src/**/*\"],\n  \"exclude\": [\"node_modules\", \"dist\", \"tests\"]\n}\n",
				SkipIfExists: true,
			},
			{
				Path:         ".env.example",
				Content:      "# Server\nPORT=3000\nNODE_ENV=development\n\n# Database\nDATABASE_URL=postgresql://user:password@localhost:5432/dbname\n\n# JWT\nJWT_SECRET=your-secret-key\nJWT_EXPIRES_IN=7d\n\n# CORS\nCORS_ORIGINS=http://localhost:3001\n",
				SkipIfExists: true,
			},
			{
				Path:         "src/types/index.ts",
				Content:      "// Core types for the application\n\nimport type { Request, Response, NextFunction } from 'express';\n\nexport interface AppConfig {\n  port: number;\n  environment: 'development' | 'staging' | 'production';\n  jwtSecret: string;\n}\n\nexport interface ApiResponse<T> {\n  data?: T;\n  message?: string;\n  status: number;\n}\n\nexport type AsyncRequestHandler = (\n  req: Request,\n  res: Response,\n  next: NextFunction\n) => Promise<void>;\n\nexport interface PaginationParams {\n  page: number;\n  limit: number;\n}\n\nexport interface PaginatedResponse<T> {\n  data: T[];\n  total: number;\n  page: number;\n  limit: number;\n  totalPages: number;\n}\n",
				SkipIfExists: true,
			},
		},
		Conventions: NamingConventions{
			Files: NamingRule{
				Pattern:     "kebab-case",
				Example:     "user-repository.ts, auth-middleware.ts",
				Description: "Archivos en kebab-case",
			},
			Components: NamingRule{
				Pattern:     "camelCase",
				Example:     "userRepository, authMiddleware",
				Description: "Funciones y objetos en camelCase",
			},
			Functions: NamingRule{
				Pattern:     "camelCase",
				Example:     "getUserById, createUser",
				Description: "Funciones en camelCase",
			},
			Types: NamingRule{
				Pattern:     "PascalCase",
				Example:     "User, ApiResponse, CreateUserDto",
				Description: "Types e interfaces en PascalCase",
			},
			Constants: NamingRule{
				Pattern:     "SCREAMING_SNAKE_CASE",
				Example:     "MAX_RETRY_COUNT, API_TIMEOUT",
				Description: "Constantes en SCREAMING_SNAKE_CASE",
			},
			Tests: NamingRule{
				Pattern:     "nombre.test.ts / nombre.spec.ts",
				Example:     "user-repository.test.ts, auth.test.ts",
				Description: "Tests con .test.ts o .spec.ts",
			},
		},
		Testing: TestingPatterns{
			Framework:    "Vitest",
			Location:     "tests/",
			Naming:       "{module}.test.ts",
			SetupFile:    "tests/setup.ts",
			Utilities:    []string{"vitest", "@vitest/coverage-v8", "@types/node"},
			CoverageTool: "vitest run --coverage",
		},
	}
}
