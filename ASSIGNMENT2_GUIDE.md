# UECS2363 — Assignment 2: Taskcafe CI/CD Enhancement Guide (Team of 4)

Purpose: a practical, team-split guide that satisfies Task 1 (CO4) and Task 2 (CO2) by enhancing the existing Taskcafe project management system. Work is split into 4 students, respecting the existing Go backend + React frontend architecture.

**Project Context**: Taskcafe is a mature project management tool with Go backend, React TypeScript frontend, and existing authentication system. We'll enhance it rather than rebuild it.

---

## Quick plan and checklist
- Goal: Fork Taskcafe and implement CI + automation + new project management features.
- Tech Stack: Go backend, React/TypeScript frontend, PostgreSQL database, Docker deployment
- Deliverables: Makefile/build scripts, test automation, enhanced Docker setup, GitHub Actions CI/CD, feature enhancements

Checklist (must be demonstrated in report & PRs):
- [ ] Fork & clone repo (evidence: GitHub screenshots + git log)
- [ ] Per-member feature branch + PRs (4 PRs)
- [ ] Build automation (Makefile + Go modules)
- [ ] Test automation (Go tests + React tests)
- [ ] Deployment automation (multi-stage Docker + docker-compose)
- [ ] CI pipeline (.github/workflows)
- [ ] Report covering repository management, automation, CI setup (<=20 pages)

---

## Team split (4 students)
Overview: each student enhances a specific area of Taskcafe and must produce code, tests, and a PR.

- Student 1 (Authentication Enhancement + Build Automation)
- Student 2 (Project Templates & Task Categories + Test Automation) 
- Student 3 (Project Analytics & Reports + Deployment Automation)
- Student 4 (CI/CD Pipeline Integration)

Each student: create a feature branch: `feature/<area>` and open PR into `develop` or `main` as per group agreement.

---

## Student 1 — Authentication Enhancement & Build Automation (CO2)
Objectives:
- Enhance existing Go authentication with password reset and session management features.
- Create comprehensive build automation using Makefile and Go modules.

**Important**: Taskcafe already has JWT authentication implemented in Go. We'll enhance it, not rebuild it.

Files to add/edit:
- `internal/route/auth.go` (enhance existing)
- `internal/handler/password_reset.go` (new)
- `internal/commands/migration.go` (enhance)
- `Makefile` (project root)
- `scripts/build.sh`

Code: `internal/handler/password_reset.go` (new Go handler)

```go
package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jordanknott/taskcafe/internal/db"
	log "github.com/sirupsen/logrus"
)

type PasswordResetHandler struct {
	repo db.Repository
}

type PasswordResetRequest struct {
	Email string `json:"email"`
}

type PasswordResetResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

func NewPasswordResetHandler(repo db.Repository) *PasswordResetHandler {
	return &PasswordResetHandler{repo: repo}
}

func (h *PasswordResetHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req PasswordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.WithError(err).Error("Failed to decode password reset request")
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Check if user exists
	user, err := h.repo.GetUserAccountByEmail(r.Context(), req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			// Don't reveal if email exists - security best practice
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(PasswordResetResponse{
				Message: "If email exists, reset instructions have been sent",
				Success: true,
			})
			return
		}
		log.WithError(err).Error("Database error during password reset")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Generate reset token
	resetToken := uuid.New().String()
	expiresAt := time.Now().Add(time.Hour * 24) // 24 hour expiry

	// Store reset token (you'll need to add this table via migration)
	_, err = h.repo.CreatePasswordResetToken(r.Context(), db.CreatePasswordResetTokenParams{
		UserID:    user.UserID,
		Token:     resetToken,
		ExpiresAt: expiresAt,
	})

	if err != nil {
		log.WithError(err).Error("Failed to create password reset token")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// TODO: Send email with reset link (integration with email service)
	log.WithFields(log.Fields{
		"user_id": user.UserID,
		"email":   req.Email,
		"token":   resetToken, // Remove in production
	}).Info("Password reset token generated")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(PasswordResetResponse{
		Message: "If email exists, reset instructions have been sent",
		Success: true,
	})
}
```

Build automation: `Makefile` (project root)

```makefile
# Taskcafe Build Automation
.PHONY: help build test clean dev deps docker lint migration

# Default target
help: ## Show this help message
	@echo "Taskcafe Build Automation"
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Environment setup
BINARY_NAME=taskcafe
BUILD_DIR=./build
FRONTEND_DIR=./frontend
MIGRATION_DIR=./migrations

deps: ## Install dependencies
	@echo "Installing Go dependencies..."
	go mod download
	go mod tidy
	@echo "Installing frontend dependencies..."
	cd $(FRONTEND_DIR) && npm ci

build-backend: ## Build Go backend
	@echo "Building backend..."
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/taskcafe

build-frontend: ## Build React frontend  
	@echo "Building frontend..."
	cd $(FRONTEND_DIR) && npm run build

build: deps build-backend build-frontend ## Build entire application

test-backend: ## Run Go tests
	@echo "Running Go tests..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-frontend: ## Run React tests
	@echo "Running frontend tests..."
	cd $(FRONTEND_DIR) && npm test -- --coverage --watchAll=false

test: test-backend test-frontend ## Run all tests

lint: ## Run linters
	@echo "Running Go linter..."
	golangci-lint run
	@echo "Running frontend linter..."
	cd $(FRONTEND_DIR) && npm run lint

clean: ## Clean build artifacts
	rm -rf $(BUILD_DIR)
	rm -rf $(FRONTEND_DIR)/build
	rm -f coverage.out coverage.html

dev: ## Start development environment
	@echo "Starting development environment..."
	docker-compose -f docker-compose.dev.yml up --build

migration-up: ## Run database migrations
	@echo "Running database migrations..."
	go run ./cmd/migrate up

migration-down: ## Rollback database migrations  
	@echo "Rolling back database migrations..."
	go run ./cmd/migrate down

docker: ## Build Docker image
	docker build -t taskcafe:latest .

docker-dev: ## Build and run development Docker environment
	docker-compose -f docker-compose.dev.yml up --build

# CI/CD targets
ci-build: deps test build ## CI build pipeline
	@echo "CI build completed successfully"

ci-test: deps test ## CI test pipeline  
	@echo "CI tests completed successfully"

ci-deploy: build docker ## CI deploy pipeline
	@echo "CI deploy completed successfully"
```

Build script: `scripts/build.sh`

```bash
#!/bin/bash
set -e

echo "=== Taskcafe Build Script ==="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check dependencies
command -v go >/dev/null 2>&1 || { print_error "Go is required but not installed. Aborting."; exit 1; }
command -v node >/dev/null 2>&1 || { print_error "Node.js is required but not installed. Aborting."; exit 1; }

print_status "Starting Taskcafe build process..."

# Clean previous builds
print_status "Cleaning previous builds..."
make clean

# Install dependencies
print_status "Installing dependencies..."
make deps

# Run tests
print_status "Running tests..."
make test || { print_error "Tests failed. Build aborted."; exit 1; }

# Build application
print_status "Building application..."
make build

print_status "Build completed successfully!"
print_status "Binary location: ./build/taskcafe"
print_status "Frontend build: ./frontend/build"
```

Student 1 deliverables: Enhanced auth handlers, build automation (Makefile), build scripts, database migration for reset tokens, PR with evidence.

---

## Student 2 — Project Templates & Task Categories + Test Automation (CO2)
Objectives:
- Implement project template system and enhanced task categorization for Taskcafe.
- Ensure comprehensive test coverage for Go backend and React frontend components.

**Context**: Build upon Taskcafe's existing project and task management to add templating capabilities.

Files to add/edit:
- `internal/db/query/project_templates.sql` (new)
- `internal/handler/project_template.go` (new)
- `frontend/src/components/ProjectTemplates/` (new React components)
- `internal/handler/project_template_test.go` (new)
- `frontend/src/components/ProjectTemplates/ProjectTemplates.test.tsx`

Database schema: `internal/db/query/project_templates.sql`

```sql
-- name: CreateProjectTemplate :one
INSERT INTO project_templates (
  project_template_id, name, description, created_by, created_at, updated_at
) VALUES (
  $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetProjectTemplateByID :one
SELECT * FROM project_templates WHERE project_template_id = $1;

-- name: GetProjectTemplatesByUser :many
SELECT * FROM project_templates 
WHERE created_by = $1 OR is_public = true 
ORDER BY created_at DESC;

-- name: UpdateProjectTemplate :one
UPDATE project_templates 
SET name = $2, description = $3, updated_at = $4
WHERE project_template_id = $1 AND created_by = $5
RETURNING *;

-- name: DeleteProjectTemplate :exec
DELETE FROM project_templates 
WHERE project_template_id = $1 AND created_by = $2;

-- name: CreateProjectFromTemplate :one
INSERT INTO projects (
  project_id, team_id, name, description, created_at, updated_at
) VALUES (
  $1, $2, $3, $4, $5, $6
) RETURNING *;
```

Go handler: `internal/handler/project_template.go`

```go
package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jordanknott/taskcafe/internal/db"
	log "github.com/sirupsen/logrus"
)

type ProjectTemplateHandler struct {
	repo db.Repository
}

type CreateProjectTemplateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
}

type ProjectTemplateResponse struct {
	ProjectTemplateID uuid.UUID `json:"project_template_id"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	CreatedBy        uuid.UUID `json:"created_by"`
	IsPublic         bool      `json:"is_public"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func NewProjectTemplateHandler(repo db.Repository) *ProjectTemplateHandler {
	return &ProjectTemplateHandler{repo: repo}
}

func (h *ProjectTemplateHandler) CreateProjectTemplate(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r) // Assume this exists from existing auth
	if userID == uuid.Nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateProjectTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.WithError(err).Error("Failed to decode create template request")
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Name == "" {
		http.Error(w, "Template name is required", http.StatusBadRequest)
		return
	}

	templateID := uuid.New()
	now := time.Now()

	template, err := h.repo.CreateProjectTemplate(r.Context(), db.CreateProjectTemplateParams{
		ProjectTemplateID: templateID,
		Name:             req.Name,
		Description:      req.Description,
		CreatedBy:        userID,
		CreatedAt:        now,
		UpdatedAt:        now,
	})

	if err != nil {
		log.WithError(err).Error("Failed to create project template")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := ProjectTemplateResponse{
		ProjectTemplateID: template.ProjectTemplateID,
		Name:             template.Name,
		Description:      template.Description,
		CreatedBy:        template.CreatedBy,
		IsPublic:         template.IsPublic,
		CreatedAt:        template.CreatedAt,
		UpdatedAt:        template.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *ProjectTemplateHandler) GetProjectTemplates(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	if userID == uuid.Nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	templates, err := h.repo.GetProjectTemplatesByUser(r.Context(), userID)
	if err != nil {
		log.WithError(err).Error("Failed to get project templates")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := make([]ProjectTemplateResponse, len(templates))
	for i, template := range templates {
		response[i] = ProjectTemplateResponse{
			ProjectTemplateID: template.ProjectTemplateID,
			Name:             template.Name,
			Description:      template.Description,
			CreatedBy:        template.CreatedBy,
			IsPublic:         template.IsPublic,
			CreatedAt:        template.CreatedAt,
			UpdatedAt:        template.UpdatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
```

React component: `frontend/src/components/ProjectTemplates/ProjectTemplateList.tsx`

```typescript
import React, { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from 'react-query';
import styled from 'styled-components';

interface ProjectTemplate {
  project_template_id: string;
  name: string;
  description: string;
  created_by: string;
  is_public: boolean;
  created_at: string;
  updated_at: string;
}

const TemplateCard = styled.div`
  background: white;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  transition: box-shadow 0.2s ease;

  &:hover {
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.15);
  }
`;

const TemplateTitle = styled.h3`
  margin: 0 0 8px 0;
  color: #333;
  font-size: 18px;
`;

const TemplateDescription = styled.p`
  margin: 0 0 12px 0;
  color: #666;
  font-size: 14px;
`;

const ActionButton = styled.button`
  background: #7367f0;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 4px;
  cursor: pointer;
  margin-right: 8px;
  font-size: 14px;

  &:hover {
    background: #5e4ec2;
  }

  &:disabled {
    background: #ccc;
    cursor: not-allowed;
  }
`;

const ProjectTemplateList: React.FC = () => {
  const queryClient = useQueryClient();

  const { data: templates, isLoading, error } = useQuery<ProjectTemplate[]>(
    'project-templates',
    async () => {
      const response = await fetch('/api/project-templates', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('accessToken')}`,
        },
      });
      if (!response.ok) {
        throw new Error('Failed to fetch templates');
      }
      return response.json();
    }
  );

  const createProjectMutation = useMutation(
    async (templateId: string) => {
      const response = await fetch(`/api/project-templates/${templateId}/create-project`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('accessToken')}`,
          'Content-Type': 'application/json',
        },
      });
      if (!response.ok) {
        throw new Error('Failed to create project from template');
      }
      return response.json();
    },
    {
      onSuccess: () => {
        queryClient.invalidateQueries('projects');
        // Show success notification
      },
    }
  );

  const handleCreateProject = (templateId: string) => {
    createProjectMutation.mutate(templateId);
  };

  if (isLoading) return <div>Loading templates...</div>;
  if (error) return <div>Error loading templates</div>;

  return (
    <div>
      <h2>Project Templates</h2>
      {templates?.map((template) => (
        <TemplateCard key={template.project_template_id}>
          <TemplateTitle>{template.name}</TemplateTitle>
          <TemplateDescription>{template.description}</TemplateDescription>
          <ActionButton
            onClick={() => handleCreateProject(template.project_template_id)}
            disabled={createProjectMutation.isLoading}
          >
            {createProjectMutation.isLoading ? 'Creating...' : 'Use Template'}
          </ActionButton>
        </TemplateCard>
      ))}
    </div>
  );
};

export default ProjectTemplateList;
```

Test file: `internal/handler/project_template_test.go`

```go
package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jordanknott/taskcafe/internal/db"
	"github.com/jordanknott/taskcafe/internal/handler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateProjectTemplate(ctx context.Context, params db.CreateProjectTemplateParams) (db.ProjectTemplate, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(db.ProjectTemplate), args.Error(1)
}

func (m *MockRepository) GetProjectTemplatesByUser(ctx context.Context, userID uuid.UUID) ([]db.ProjectTemplate, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]db.ProjectTemplate), args.Error(1)
}

func TestCreateProjectTemplate(t *testing.T) {
	mockRepo := new(MockRepository)
	h := handler.NewProjectTemplateHandler(mockRepo)

	templateID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	reqBody := handler.CreateProjectTemplateRequest{
		Name:        "Test Template",
		Description: "A test project template",
		IsPublic:    false,
	}

	expectedTemplate := db.ProjectTemplate{
		ProjectTemplateID: templateID,
		Name:             reqBody.Name,
		Description:      reqBody.Description,
		CreatedBy:        userID,
		IsPublic:         reqBody.IsPublic,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	mockRepo.On("CreateProjectTemplate", mock.Anything, mock.MatchedBy(func(params db.CreateProjectTemplateParams) bool {
		return params.Name == reqBody.Name && params.Description == reqBody.Description
	})).Return(expectedTemplate, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/project-templates", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	// Mock user authentication
	ctx := context.WithValue(req.Context(), "user_id", userID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	h.CreateProjectTemplate(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response handler.ProjectTemplateResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedTemplate.Name, response.Name)
	assert.Equal(t, expectedTemplate.Description, response.Description)

	mockRepo.AssertExpectations(t)
}

func TestGetProjectTemplates(t *testing.T) {
	mockRepo := new(MockRepository)
	h := handler.NewProjectTemplateHandler(mockRepo)

	userID := uuid.New()
	templates := []db.ProjectTemplate{
		{
			ProjectTemplateID: uuid.New(),
			Name:             "Template 1",
			Description:      "First template",
			CreatedBy:        userID,
			IsPublic:         false,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
	}

	mockRepo.On("GetProjectTemplatesByUser", mock.Anything, userID).Return(templates, nil)

	req := httptest.NewRequest("GET", "/api/project-templates", nil)
	ctx := context.WithValue(req.Context(), "user_id", userID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	h.GetProjectTemplates(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []handler.ProjectTemplateResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 1)
	assert.Equal(t, templates[0].Name, response[0].Name)

	mockRepo.AssertExpectations(t)
}
```

Frontend test: `frontend/src/components/ProjectTemplates/ProjectTemplates.test.tsx`

```typescript
import React from 'react';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from 'react-query';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import ProjectTemplateList from './ProjectTemplateList';

// Mock server for API calls
const server = setupServer(
  rest.get('/api/project-templates', (req, res, ctx) => {
    return res(
      ctx.json([
        {
          project_template_id: '1',
          name: 'Test Template',
          description: 'A test template',
          created_by: 'user1',
          is_public: false,
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
      ])
    );
  }),
  rest.post('/api/project-templates/:id/create-project', (req, res, ctx) => {
    return res(ctx.json({ success: true }));
  })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

// Mock localStorage
const mockLocalStorage = {
  getItem: jest.fn(() => 'mock-token'),
  setItem: jest.fn(),
  removeItem: jest.fn(),
};
Object.defineProperty(window, 'localStorage', { value: mockLocalStorage });

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
};

describe('ProjectTemplateList', () => {
  test('renders project templates', async () => {
    render(<ProjectTemplateList />, { wrapper: createWrapper() });

    expect(screen.getByText('Loading templates...')).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByText('Test Template')).toBeInTheDocument();
      expect(screen.getByText('A test template')).toBeInTheDocument();
      expect(screen.getByText('Use Template')).toBeInTheDocument();
    });
  });

  test('creates project from template', async () => {
    render(<ProjectTemplateList />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.getByText('Use Template')).toBeInTheDocument();
    });

    const useTemplateButton = screen.getByText('Use Template');
    fireEvent.click(useTemplateButton);

    await waitFor(() => {
      expect(screen.getByText('Creating...')).toBeInTheDocument();
    });
  });
});
```

Student 2 deliverables: Project template system (Go handlers + React components), comprehensive tests, database migrations, PR with test coverage report.

---

## Student 3 — Project Analytics & Reports + Deployment Automation (CO2)
Objectives:
- Implement project analytics and reporting endpoints for Taskcafe's project management domain.
- Create production-ready deployment automation with multi-stage Docker and orchestration.

**Context**: Build analytics for project completion rates, team productivity, and task metrics instead of financial reports.

Files to add/edit:
- `internal/handler/analytics.go` (new)
- `internal/db/query/analytics.sql` (new)
- `frontend/src/components/Analytics/` (new React components)
- `Dockerfile.production` (multi-stage optimized)
- `docker-compose.production.yml`
- `scripts/deploy.sh`

Analytics SQL queries: `internal/db/query/analytics.sql`

```sql
-- name: GetProjectMetrics :one
SELECT 
  COUNT(CASE WHEN complete = true THEN 1 END) as completed_tasks,
  COUNT(CASE WHEN complete = false THEN 1 END) as pending_tasks,
  COUNT(CASE WHEN due_date < NOW() AND complete = false THEN 1 END) as overdue_tasks,
  AVG(CASE WHEN complete = true AND completed_at IS NOT NULL 
      THEN EXTRACT(EPOCH FROM (completed_at - created_at))/86400 
      END) as avg_completion_days
FROM tasks 
WHERE project_id = $1;

-- name: GetTeamProductivityReport :many
SELECT 
  u.full_name,
  u.username,
  COUNT(t.task_id) as total_tasks,
  COUNT(CASE WHEN t.complete = true THEN 1 END) as completed_tasks,
  AVG(CASE WHEN t.complete = true AND t.completed_at IS NOT NULL 
      THEN EXTRACT(EPOCH FROM (t.completed_at - t.created_at))/86400 
      END) as avg_completion_days
FROM user_accounts u
LEFT JOIN task_assigned ta ON u.user_id = ta.user_id  
LEFT JOIN tasks t ON ta.task_id = t.task_id
WHERE t.project_id = $1
GROUP BY u.user_id, u.full_name, u.username
ORDER BY completed_tasks DESC;

-- name: GetProjectCompletionTrend :many
SELECT 
  DATE_TRUNC('week', completed_at) as week_start,
  COUNT(*) as tasks_completed
FROM tasks 
WHERE project_id = $1 
  AND complete = true 
  AND completed_at >= $2
GROUP BY DATE_TRUNC('week', completed_at)
ORDER BY week_start;

-- name: GetTaskCategoryBreakdown :many
SELECT 
  tg.name as task_group_name,
  COUNT(t.task_id) as total_tasks,
  COUNT(CASE WHEN t.complete = true THEN 1 END) as completed_tasks,
  ROUND(COUNT(CASE WHEN t.complete = true THEN 1 END) * 100.0 / COUNT(t.task_id), 2) as completion_percentage
FROM task_groups tg
LEFT JOIN tasks t ON tg.task_group_id = t.task_group_id
WHERE tg.project_id = $1
GROUP BY tg.task_group_id, tg.name
ORDER BY completion_percentage DESC;
```

Go analytics handler: `internal/handler/analytics.go`

```go
package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jordanknott/taskcafe/internal/db"
	log "github.com/sirupsen/logrus"
)

type AnalyticsHandler struct {
	repo db.Repository
}

type ProjectMetricsResponse struct {
	CompletedTasks     int     `json:"completed_tasks"`
	PendingTasks       int     `json:"pending_tasks"`
	OverdueTasks       int     `json:"overdue_tasks"`
	AvgCompletionDays  float64 `json:"avg_completion_days"`
	CompletionRate     float64 `json:"completion_rate"`
}

type TeamProductivityMember struct {
	FullName          string  `json:"full_name"`
	Username          string  `json:"username"`
	TotalTasks        int     `json:"total_tasks"`
	CompletedTasks    int     `json:"completed_tasks"`
	AvgCompletionDays float64 `json:"avg_completion_days"`
	CompletionRate    float64 `json:"completion_rate"`
}

type ProjectCompletionTrend struct {
	WeekStart      time.Time `json:"week_start"`
	TasksCompleted int       `json:"tasks_completed"`
}

type TaskCategoryBreakdown struct {
	TaskGroupName        string  `json:"task_group_name"`
	TotalTasks          int     `json:"total_tasks"`
	CompletedTasks      int     `json:"completed_tasks"`
	CompletionPercentage float64 `json:"completion_percentage"`
}

type AnalyticsReport struct {
	ProjectID            uuid.UUID                `json:"project_id"`
	ProjectName          string                   `json:"project_name"`
	GeneratedAt          time.Time                `json:"generated_at"`
	Metrics              ProjectMetricsResponse   `json:"metrics"`
	TeamProductivity     []TeamProductivityMember `json:"team_productivity"`
	CompletionTrend      []ProjectCompletionTrend `json:"completion_trend"`
	CategoryBreakdown    []TaskCategoryBreakdown  `json:"category_breakdown"`
}

func NewAnalyticsHandler(repo db.Repository) *AnalyticsHandler {
	return &AnalyticsHandler{repo: repo}
}

func (h *AnalyticsHandler) GetProjectAnalytics(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "projectID")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	// Verify user has access to project (implement your auth logic)
	userID := GetUserID(r)
	hasAccess, err := h.repo.CheckProjectAccess(r.Context(), userID, projectID)
	if err != nil || !hasAccess {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get project metrics
	metrics, err := h.repo.GetProjectMetrics(r.Context(), projectID)
	if err != nil {
		log.WithError(err).Error("Failed to get project metrics")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Calculate completion rate
	totalTasks := metrics.CompletedTasks + metrics.PendingTasks
	var completionRate float64
	if totalTasks > 0 {
		completionRate = float64(metrics.CompletedTasks) / float64(totalTasks) * 100
	}

	metricsResponse := ProjectMetricsResponse{
		CompletedTasks:    metrics.CompletedTasks,
		PendingTasks:      metrics.PendingTasks,
		OverdueTasks:      metrics.OverdueTasks,
		AvgCompletionDays: metrics.AvgCompletionDays,
		CompletionRate:    completionRate,
	}

	// Get team productivity
	teamData, err := h.repo.GetTeamProductivityReport(r.Context(), projectID)
	if err != nil {
		log.WithError(err).Error("Failed to get team productivity")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	teamProductivity := make([]TeamProductivityMember, len(teamData))
	for i, member := range teamData {
		var memberCompletionRate float64
		if member.TotalTasks > 0 {
			memberCompletionRate = float64(member.CompletedTasks) / float64(member.TotalTasks) * 100
		}

		teamProductivity[i] = TeamProductivityMember{
			FullName:          member.FullName,
			Username:          member.Username,
			TotalTasks:        member.TotalTasks,
			CompletedTasks:    member.CompletedTasks,
			AvgCompletionDays: member.AvgCompletionDays,
			CompletionRate:    memberCompletionRate,
		}
	}

	// Get completion trend (last 12 weeks)
	twelveWeeksAgo := time.Now().AddDate(0, 0, -84)
	trendData, err := h.repo.GetProjectCompletionTrend(r.Context(), projectID, twelveWeeksAgo)
	if err != nil {
		log.WithError(err).Error("Failed to get completion trend")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	completionTrend := make([]ProjectCompletionTrend, len(trendData))
	for i, trend := range trendData {
		completionTrend[i] = ProjectCompletionTrend{
			WeekStart:      trend.WeekStart,
			TasksCompleted: trend.TasksCompleted,
		}
	}

	// Get category breakdown
	categoryData, err := h.repo.GetTaskCategoryBreakdown(r.Context(), projectID)
	if err != nil {
		log.WithError(err).Error("Failed to get category breakdown")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	categoryBreakdown := make([]TaskCategoryBreakdown, len(categoryData))
	for i, category := range categoryData {
		categoryBreakdown[i] = TaskCategoryBreakdown{
			TaskGroupName:        category.TaskGroupName,
			TotalTasks:          category.TotalTasks,
			CompletedTasks:      category.CompletedTasks,
			CompletionPercentage: category.CompletionPercentage,
		}
	}

	// Get project name
	project, err := h.repo.GetProjectByID(r.Context(), projectID)
	if err != nil {
		log.WithError(err).Error("Failed to get project")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	report := AnalyticsReport{
		ProjectID:         projectID,
		ProjectName:       project.Name,
		GeneratedAt:       time.Now(),
		Metrics:           metricsResponse,
		TeamProductivity:  teamProductivity,
		CompletionTrend:   completionTrend,
		CategoryBreakdown: categoryBreakdown,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

func (h *AnalyticsHandler) ExportProjectReport(w http.ResponseWriter, r *http.Request) {
	// Implementation for CSV/PDF export would go here
	projectIDStr := chi.URLParam(r, "projectID")
	format := r.URL.Query().Get("format") // csv, pdf

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Export functionality would be implemented here",
		"project_id": projectIDStr,
		"format": format,
	})
}
```

React Analytics Dashboard: `frontend/src/components/Analytics/ProjectAnalytics.tsx`

```typescript
import React from 'react';
import { useQuery } from 'react-query';
import styled from 'styled-components';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer, LineChart, Line, PieChart, Pie, Cell } from 'recharts';

interface AnalyticsReport {
  project_id: string;
  project_name: string;
  generated_at: string;
  metrics: {
    completed_tasks: number;
    pending_tasks: number;
    overdue_tasks: number;
    avg_completion_days: number;
    completion_rate: number;
  };
  team_productivity: Array<{
    full_name: string;
    username: string;
    total_tasks: number;
    completed_tasks: number;
    completion_rate: number;
  }>;
  completion_trend: Array<{
    week_start: string;
    tasks_completed: number;
  }>;
  category_breakdown: Array<{
    task_group_name: string;
    total_tasks: number;
    completed_tasks: number;
    completion_percentage: number;
  }>;
}

const DashboardContainer = styled.div`
  padding: 24px;
  background-color: #f8f9fa;
  min-height: 100vh;
`;

const MetricsGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
  margin-bottom: 32px;
`;

const MetricCard = styled.div`
  background: white;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  text-align: center;
`;

const MetricValue = styled.div`
  font-size: 32px;
  font-weight: bold;
  color: #7367f0;
  margin-bottom: 8px;
`;

const MetricLabel = styled.div`
  color: #666;
  font-size: 14px;
`;

const ChartContainer = styled.div`
  background: white;
  padding: 24px;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  margin-bottom: 24px;
`;

const ChartTitle = styled.h3`
  margin: 0 0 20px 0;
  color: #333;
`;

interface ProjectAnalyticsProps {
  projectId: string;
}

const ProjectAnalytics: React.FC<ProjectAnalyticsProps> = ({ projectId }) => {
  const { data: analytics, isLoading, error } = useQuery<AnalyticsReport>(
    ['project-analytics', projectId],
    async () => {
      const response = await fetch(`/api/projects/${projectId}/analytics`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('accessToken')}`,
        },
      });
      if (!response.ok) {
        throw new Error('Failed to fetch analytics');
      }
      return response.json();
    }
  );

  if (isLoading) return <div>Loading analytics...</div>;
  if (error) return <div>Error loading analytics</div>;
  if (!analytics) return <div>No analytics data available</div>;

  const COLORS = ['#7367f0', '#28c76f', '#ea5455', '#ff9f43'];

  return (
    <DashboardContainer>
      <h1>Project Analytics: {analytics.project_name}</h1>
      
      {/* Key Metrics */}
      <MetricsGrid>
        <MetricCard>
          <MetricValue>{analytics.metrics.completed_tasks}</MetricValue>
          <MetricLabel>Completed Tasks</MetricLabel>
        </MetricCard>
        <MetricCard>
          <MetricValue>{analytics.metrics.pending_tasks}</MetricValue>
          <MetricLabel>Pending Tasks</MetricLabel>
        </MetricCard>
        <MetricCard>
          <MetricValue>{analytics.metrics.overdue_tasks}</MetricValue>
          <MetricLabel>Overdue Tasks</MetricLabel>
        </MetricCard>
        <MetricCard>
          <MetricValue>{analytics.metrics.completion_rate.toFixed(1)}%</MetricValue>
          <MetricLabel>Completion Rate</MetricLabel>
        </MetricCard>
      </MetricsGrid>

      {/* Team Productivity Chart */}
      <ChartContainer>
        <ChartTitle>Team Productivity</ChartTitle>
        <ResponsiveContainer width="100%" height={300}>
          <BarChart data={analytics.team_productivity}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="username" />
            <YAxis />
            <Tooltip />
            <Legend />
            <Bar dataKey="completed_tasks" fill="#7367f0" name="Completed Tasks" />
            <Bar dataKey="total_tasks" fill="#e3e1fc" name="Total Tasks" />
          </BarChart>
        </ResponsiveContainer>
      </ChartContainer>

      {/* Completion Trend */}
      <ChartContainer>
        <ChartTitle>Weekly Completion Trend</ChartTitle>
        <ResponsiveContainer width="100%" height={300}>
          <LineChart data={analytics.completion_trend}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="week_start" />
            <YAxis />
            <Tooltip />
            <Legend />
            <Line type="monotone" dataKey="tasks_completed" stroke="#7367f0" strokeWidth={2} />
          </LineChart>
        </ResponsiveContainer>
      </ChartContainer>

      {/* Category Breakdown */}
      <ChartContainer>
        <ChartTitle>Task Category Breakdown</ChartTitle>
        <ResponsiveContainer width="100%" height={300}>
          <PieChart>
            <Pie
              data={analytics.category_breakdown}
              cx="50%"
              cy="50%"
              labelLine={false}
              label={(entry) => `${entry.task_group_name}: ${entry.completion_percentage}%`}
              outerRadius={80}
              fill="#8884d8"
              dataKey="completed_tasks"
            >
              {analytics.category_breakdown.map((entry, index) => (
                <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
              ))}
            </Pie>
            <Tooltip />
          </PieChart>
        </ResponsiveContainer>
      </ChartContainer>
    </DashboardContainer>
  );
};

export default ProjectAnalytics;
```

Production Dockerfile: `Dockerfile.production`

```dockerfile
# Multi-stage production build for Taskcafe
FROM node:18-alpine AS frontend-builder

WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci --only=production

COPY frontend/ .
RUN npm run build

# Go backend builder
FROM golang:1.19-alpine AS backend-builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev sqlite-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o taskcafe ./cmd/taskcafe

# Final production image
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy binary from builder stage
COPY --from=backend-builder /app/taskcafe .

# Copy frontend build from frontend builder
COPY --from=frontend-builder /app/frontend/build ./web/

# Copy migrations
COPY migrations ./migrations/

# Create non-root user for security
RUN addgroup -g 1001 -S taskcafe && \
    adduser -S taskcafe -u 1001 -G taskcafe

# Change ownership
RUN chown -R taskcafe:taskcafe /root/

USER taskcafe

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3333/health || exit 1

EXPOSE 3333

CMD ["./taskcafe"]
```

Production docker-compose: `docker-compose.production.yml`

```yaml
version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile.production
    environment:
      - NODE_ENV=production
      - DATABASE_URL=postgres://taskcafe:${DB_PASSWORD}@db:5432/taskcafe?sslmode=disable
      - JWT_SECRET=${JWT_SECRET}
      - CORS_ALLOWED_ORIGINS=${CORS_ALLOWED_ORIGINS}
    ports:
      - "3333:3333"
    depends_on:
      db:
        condition: service_healthy
    restart: unless-stopped
    networks:
      - taskcafe-network
    volumes:
      - app-uploads:/root/uploads

  db:
    image: postgres:14-alpine
    environment:
      - POSTGRES_DB=taskcafe
      - POSTGRES_USER=taskcafe
      - POSTGRES_PASSWORD=${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/init-db.sql:/docker-entrypoint-initdb.d/init-db.sql
    restart: unless-stopped
    networks:
      - taskcafe-network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U taskcafe"]
      interval: 10s
      timeout: 5s
      retries: 5

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf
      - ./nginx/ssl:/etc/nginx/ssl
    depends_on:
      - app
    restart: unless-stopped
    networks:
      - taskcafe-network

volumes:
  postgres_data:
  app-uploads:

networks:
  taskcafe-network:
    driver: bridge
```

Deployment script: `scripts/deploy.sh`

```bash
#!/bin/bash
set -e

echo "=== Taskcafe Production Deployment ==="

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

print_status() { echo -e "${GREEN}[DEPLOY]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[WARN]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Check if required environment file exists
if [ ! -f .env.production ]; then
    print_error ".env.production file not found"
    exit 1
fi

# Load environment variables
export $(cat .env.production | xargs)

# Validate required environment variables
required_vars=("DB_PASSWORD" "JWT_SECRET" "CORS_ALLOWED_ORIGINS")
for var in "${required_vars[@]}"; do
    if [ -z "${!var}" ]; then
        print_error "Required environment variable $var is not set"
        exit 1
    fi
done

print_status "Starting production deployment..."

# Stop existing containers
print_status "Stopping existing containers..."
docker-compose -f docker-compose.production.yml down

# Pull latest images
print_status "Pulling latest base images..."
docker-compose -f docker-compose.production.yml pull

# Build application
print_status "Building production images..."
docker-compose -f docker-compose.production.yml build --no-cache

# Run database migrations
print_status "Running database migrations..."
docker-compose -f docker-compose.production.yml run --rm app ./taskcafe migrate up

# Start services
print_status "Starting production services..."
docker-compose -f docker-compose.production.yml up -d

# Wait for services to be healthy
print_status "Waiting for services to be healthy..."
sleep 30

# Health check
print_status "Performing health check..."
if curl -f http://localhost:3333/health; then
    print_status "Deployment completed successfully!"
    print_status "Taskcafe is now running at http://localhost"
else
    print_error "Health check failed. Check logs with: docker-compose -f docker-compose.production.yml logs"
    exit 1
fi

# Display running containers
docker-compose -f docker-compose.production.yml ps
```

Student 3 deliverables: Analytics system (Go handlers + React dashboard), production Docker setup, deployment automation, performance monitoring, PR with deployment documentation.

---

## Student 4 — CI/CD Pipeline Integration (CO4 + CO2)
Objectives:
- Create comprehensive GitHub Actions CI/CD pipeline for Go backend + React frontend.
- Add automated testing, security scanning, and deployment workflows.

**Context**: Design CI/CD for Taskcafe's actual tech stack (Go + React + PostgreSQL).

Files to add/edit:
- `.github/workflows/ci.yml`
- `.github/workflows/security.yml`
- `.github/workflows/deploy.yml`
- `.github/dependabot.yml`
- `scripts/test-integration.sh`

Main CI Pipeline: `.github/workflows/ci.yml`

```yaml
name: Taskcafe CI/CD Pipeline

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

env:
  GO_VERSION: '1.19'
  NODE_VERSION: '18'

jobs:
  lint-and-format:
    name: Code Quality
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: ${{ env.GO_VERSION }}

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}
          cache: 'npm'
          cache-dependency-path: frontend/package-lock.json

      - name: Install Go linter
        run: |
          curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.54.2
          echo "$(go env GOPATH)/bin" >> $GITHUB_PATH

      - name: Run Go linter
        run: golangci-lint run --timeout=10m

      - name: Check Go formatting
        run: |
          if [ "$(gofmt -s -l . | wc -l)" -gt 0 ]; then
            echo "Go files are not formatted:"
            gofmt -s -l .
            exit 1
          fi

      - name: Install frontend dependencies
        working-directory: frontend
        run: npm ci

      - name: Run frontend linter
        working-directory: frontend
        run: npm run lint

      - name: Check frontend formatting
        working-directory: frontend
        run: npm run format:check

  test-backend:
    name: Backend Tests
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:14
        env:
          POSTGRES_PASSWORD: test_password
          POSTGRES_USER: test_user
          POSTGRES_DB: taskcafe_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: ${{ env.GO_VERSION }}

      - name: Cache Go modules
        uses: actions/cache@v3
        with:
          path: ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
          restore-keys: |
            ${{ runner.os }}-go-

      - name: Install dependencies
        run: go mod download

      - name: Run database migrations
        env:
          DATABASE_URL: postgres://test_user:test_password@localhost:5432/taskcafe_test?sslmode=disable
        run: |
          go run ./cmd/migrate up

      - name: Run tests with coverage
        env:
          DATABASE_URL: postgres://test_user:test_password@localhost:5432/taskcafe_test?sslmode=disable
        run: |
          go test -v -race -covermode=atomic -coverprofile=coverage.out ./...

      - name: Generate coverage report
        run: go tool cover -html=coverage.out -o coverage.html

      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out
          flags: backend
          name: backend-coverage

      - name: Upload coverage artifacts
        uses: actions/upload-artifact@v3
        with:
          name: backend-coverage
          path: |
            coverage.out
            coverage.html

  test-frontend:
    name: Frontend Tests
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}
          cache: 'npm'
          cache-dependency-path: frontend/package-lock.json

      - name: Install dependencies
        working-directory: frontend
        run: npm ci

      - name: Run tests with coverage
        working-directory: frontend
        run: npm run test:ci

      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v3
        with:
          file: ./frontend/coverage/lcov.info
          flags: frontend
          name: frontend-coverage

      - name: Upload test artifacts
        uses: actions/upload-artifact@v3
        with:
          name: frontend-test-results
          path: |
            frontend/coverage/
            frontend/test-results/

  integration-tests:
    name: Integration Tests
    runs-on: ubuntu-latest
    needs: [test-backend, test-frontend]
    services:
      postgres:
        image: postgres:14
        env:
          POSTGRES_PASSWORD: test_password
          POSTGRES_USER: test_user
          POSTGRES_DB: taskcafe_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: ${{ env.GO_VERSION }}

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}

      - name: Build application
        run: make build

      - name: Run integration tests
        env:
          DATABASE_URL: postgres://test_user:test_password@localhost:5432/taskcafe_test?sslmode=disable
        run: ./scripts/test-integration.sh

  build-and-package:
    name: Build & Package
    runs-on: ubuntu-latest
    needs: [lint-and-format, test-backend, test-frontend]
    if: github.event_name == 'push'
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: ${{ env.GO_VERSION }}

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}

      - name: Build application
        run: make build

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Login to GitHub Container Registry
        uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ghcr.io/${{ github.repository }}
          tags: |
            type=ref,event=branch
            type=ref,event=pr
            type=sha
            type=raw,value=latest,enable={{is_default_branch}}

      - name: Build and push Docker image
        uses: docker/build-push-action@v5
        with:
          context: .
          file: ./Dockerfile.production
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

      - name: Upload build artifacts
        uses: actions/upload-artifact@v3
        with:
          name: build-artifacts
          path: |
            build/
            frontend/build/

  deploy-staging:
    name: Deploy to Staging
    runs-on: ubuntu-latest
    needs: [integration-tests, build-and-package]
    if: github.ref == 'refs/heads/develop'
    environment:
      name: staging
      url: https://staging.taskcafe.example.com
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Deploy to staging
        env:
          DEPLOY_HOST: ${{ secrets.STAGING_HOST }}
          DEPLOY_USER: ${{ secrets.STAGING_USER }}
          DEPLOY_KEY: ${{ secrets.STAGING_SSH_KEY }}
          DB_PASSWORD: ${{ secrets.STAGING_DB_PASSWORD }}
          JWT_SECRET: ${{ secrets.STAGING_JWT_SECRET }}
        run: |
          echo "$DEPLOY_KEY" > deploy_key
          chmod 600 deploy_key
          
          scp -i deploy_key -o StrictHostKeyChecking=no \
            docker-compose.production.yml \
            scripts/deploy.sh \
            $DEPLOY_USER@$DEPLOY_HOST:/app/
          
          ssh -i deploy_key -o StrictHostKeyChecking=no $DEPLOY_USER@$DEPLOY_HOST \
            "cd /app && \
             echo 'DB_PASSWORD=$DB_PASSWORD' > .env.production && \
             echo 'JWT_SECRET=$JWT_SECRET' >> .env.production && \
             echo 'CORS_ALLOWED_ORIGINS=https://staging.taskcafe.example.com' >> .env.production && \
             ./deploy.sh"

  deploy-production:
    name: Deploy to Production
    runs-on: ubuntu-latest
    needs: [integration-tests, build-and-package]
    if: github.ref == 'refs/heads/main'
    environment:
      name: production
      url: https://taskcafe.example.com
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Deploy to production
        env:
          DEPLOY_HOST: ${{ secrets.PRODUCTION_HOST }}
          DEPLOY_USER: ${{ secrets.PRODUCTION_USER }}
          DEPLOY_KEY: ${{ secrets.PRODUCTION_SSH_KEY }}
          DB_PASSWORD: ${{ secrets.PRODUCTION_DB_PASSWORD }}
          JWT_SECRET: ${{ secrets.PRODUCTION_JWT_SECRET }}
        run: |
          echo "$DEPLOY_KEY" > deploy_key
          chmod 600 deploy_key
          
          scp -i deploy_key -o StrictHostKeyChecking=no \
            docker-compose.production.yml \
            scripts/deploy.sh \
            $DEPLOY_USER@$DEPLOY_HOST:/app/
          
          ssh -i deploy_key -o StrictHostKeyChecking=no $DEPLOY_USER@$DEPLOY_HOST \
            "cd /app && \
             echo 'DB_PASSWORD=$DB_PASSWORD' > .env.production && \
             echo 'JWT_SECRET=$JWT_SECRET' >> .env.production && \
             echo 'CORS_ALLOWED_ORIGINS=https://taskcafe.example.com' >> .env.production && \
             ./deploy.sh"
```

Security scanning: `.github/workflows/security.yml`

```yaml
name: Security Scanning

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]
  schedule:
    - cron: '0 0 * * 1' # Weekly on Monday

jobs:
  security-scan:
    name: Security Analysis
    runs-on: ubuntu-latest
    permissions:
      actions: read
      contents: read
      security-events: write

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Run Trivy vulnerability scanner
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          scan-ref: '.'
          format: 'sarif'
          output: 'trivy-results.sarif'

      - name: Upload Trivy scan results to GitHub Security tab
        uses: github/codeql-action/upload-sarif@v2
        with:
          sarif_file: 'trivy-results.sarif'

      - name: Run Gosec Security Scanner
        uses: securecodewarrior/github-action-gosec@master
        with:
          args: '-fmt sarif -out gosec-results.sarif ./...'

      - name: Upload Gosec scan results
        uses: github/codeql-action/upload-sarif@v2
        with:
          sarif_file: 'gosec-results.sarif'

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '18'

      - name: Install frontend dependencies
        working-directory: frontend
        run: npm ci

      - name: Run npm audit
        working-directory: frontend
        run: npm audit --audit-level high

      - name: Run Snyk security scan
        uses: snyk/actions/node@master
        env:
          SNYK_TOKEN: ${{ secrets.SNYK_TOKEN }}
        with:
          args: --file=frontend/package.json
```

Dependabot configuration: `.github/dependabot.yml`

```yaml
version: 2
updates:
  # Go modules
  - package-ecosystem: "gomod"
    directory: "/"
    schedule:
      interval: "weekly"
      day: "monday"
      time: "09:00"
    assignees:
      - "your-team"
    reviewers:
      - "your-team"
    commit-message:
      prefix: "deps"
      include: "scope"

  # Frontend npm packages
  - package-ecosystem: "npm"
    directory: "/frontend"
    schedule:
      interval: "weekly"
      day: "monday"
      time: "09:00"
    assignees:
      - "your-team"
    reviewers:
      - "your-team"
    commit-message:
      prefix: "deps"
      include: "scope"

  # GitHub Actions
  - package-ecosystem: "github-actions"
    directory: "/"
    schedule:
      interval: "weekly"
      day: "monday"
      time: "09:00"
    assignees:
      - "your-team"
    reviewers:
      - "your-team"
    commit-message:
      prefix: "ci"
      include: "scope"

  # Docker
  - package-ecosystem: "docker"
    directory: "/"
    schedule:
      interval: "weekly"
      day: "monday"
      time: "09:00"
    assignees:
      - "your-team"
    reviewers:
      - "your-team"
    commit-message:
      prefix: "docker"
      include: "scope"
```

Integration test script: `scripts/test-integration.sh`

```bash
#!/bin/bash
set -e

echo "=== Running Integration Tests ==="

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

print_status() { echo -e "${GREEN}[TEST]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[WARN]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Check if binaries exist
if [ ! -f "./build/taskcafe" ]; then
    print_error "Backend binary not found. Run 'make build' first."
    exit 1
fi

if [ ! -d "./frontend/build" ]; then
    print_error "Frontend build not found. Run 'make build' first."
    exit 1
fi

# Start the application in background
print_status "Starting Taskcafe server..."
export DATABASE_URL=${DATABASE_URL:-"postgres://test_user:test_password@localhost:5432/taskcafe_test?sslmode=disable"}
export JWT_SECRET="test_jwt_secret_for_integration_tests"
export PORT=3334

./build/taskcafe &
APP_PID=$!

# Wait for server to start
print_status "Waiting for server to start..."
sleep 5

# Function to cleanup
cleanup() {
    print_status "Cleaning up..."
    kill $APP_PID 2>/dev/null || true
    wait $APP_PID 2>/dev/null || true
}

# Set trap to cleanup on exit
trap cleanup EXIT

# Health check
print_status "Performing health check..."
if ! curl -f http://localhost:3334/health; then
    print_error "Health check failed"
    exit 1
fi

# Test user registration
print_status "Testing user registration..."
REGISTER_RESPONSE=$(curl -s -X POST http://localhost:3334/auth/register \
    -H "Content-Type: application/json" \
    -d '{
        "username": "testuser",
        "email": "test@example.com",
        "full_name": "Test User",
        "password": "testpassword123"
    }')

if ! echo "$REGISTER_RESPONSE" | grep -q "success"; then
    print_error "User registration failed: $REGISTER_RESPONSE"
    exit 1
fi

# Test user login
print_status "Testing user login..."
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:3334/auth/login \
    -H "Content-Type: application/json" \
    -d '{
        "username": "testuser",
        "password": "testpassword123"
    }')

ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"accessToken":"[^"]*"' | cut -d'"' -f4)

if [ -z "$ACCESS_TOKEN" ]; then
    print_error "Login failed: $LOGIN_RESPONSE"
    exit 1
fi

# Test creating a project
print_status "Testing project creation..."
PROJECT_RESPONSE=$(curl -s -X POST http://localhost:3334/projects \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -d '{
        "name": "Test Project",
        "team_id": "00000000-0000-0000-0000-000000000000"
    }')

PROJECT_ID=$(echo "$PROJECT_RESPONSE" | grep -o '"projectID":"[^"]*"' | cut -d'"' -f4)

if [ -z "$PROJECT_ID" ]; then
    print_error "Project creation failed: $PROJECT_RESPONSE"
    exit 1
fi

# Test creating a task
print_status "Testing task creation..."
TASK_RESPONSE=$(curl -s -X POST http://localhost:3334/projects/$PROJECT_ID/task-groups/default/tasks \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -d '{
        "name": "Test Task",
        "description": "A test task for integration testing"
    }')

if ! echo "$TASK_RESPONSE" | grep -q "Test Task"; then
    print_error "Task creation failed: $TASK_RESPONSE"
    exit 1
fi

# Test analytics endpoint
print_status "Testing analytics endpoint..."
ANALYTICS_RESPONSE=$(curl -s -X GET http://localhost:3334/projects/$PROJECT_ID/analytics \
    -H "Authorization: Bearer $ACCESS_TOKEN")

if ! echo "$ANALYTICS_RESPONSE" | grep -q "project_id"; then
    print_error "Analytics endpoint failed: $ANALYTICS_RESPONSE"
    exit 1
fi

print_status "All integration tests passed successfully!"
```

Student 4 deliverables: Complete CI/CD pipeline, security scanning, automated deployment, dependency management, integration tests, documentation on secrets management and pipeline monitoring.

---

## Report template & what to include (≤20 pages)
Structure your team report with these sections (map to grading rubric):

1. Cover + marking rubric (first page)
2. Executive summary (1/2 page)
3. Project overview + repo link
4. Team workflow (forks, branches, PRs) — include screenshots and commands used:
   - git clone ...
   - git checkout -b feature/...
   - git add/commit/push
   - PRs and merge strategy
5. Build automation (Student 1): include `build.gradle`, sample `./gradlew build` output and analysis
6. Test automation (Student 2): include `npm run test:ci` output, coverage report
7. Deployment automation (Student 3): include Dockerfile, docker-compose, staging run logs
8. CI setup (Student 4): `.github/workflows/ci-cd.yml` explanation and successful run screenshots
9. Challenges and how the team resolved them (merge conflicts, test failures)
10. Individual contributions (who did what) with commit ranges/PR links
11. Appendix: commands, environment variables, secrets to set, run instructions

---

## Required commands / try it locally
Run tests and build locally (from project root):

```powershell
# Install Go dependencies
go mod download
go mod tidy

# Install frontend dependencies  
cd frontend; npm ci; cd ..

# Build application using Makefile
make build

# Run backend tests
make test-backend

# Run frontend tests
make test-frontend

# Run all tests
make test

# Start development environment
make dev

# Build Docker image
make docker

# Run production environment
docker-compose -f docker-compose.production.yml up --build -d

# Run integration tests
./scripts/test-integration.sh

# Deploy to production (with proper environment)
./scripts/deploy.sh
```

## Database setup for development
```powershell
# Start PostgreSQL with Docker
docker run --name taskcafe-postgres -e POSTGRES_PASSWORD=password -e POSTGRES_DB=taskcafe -p 5432:5432 -d postgres:14

# Run migrations
make migration-up

# Or manually with Go
go run ./cmd/migrate up
```

---

## Report template & what to include (≤20 pages)
Structure your team report with these sections (map to grading rubric):

1. **Cover + Marking Rubric** (1 page)
   - Group number, member names, student IDs
   - Copy of marking rubric with self-assessment

2. **Executive Summary** (0.5 page)
   - Project overview: enhancing Taskcafe with CI/CD automation
   - Key achievements and technologies used

3. **Project Overview** (1 page)
   - Repository link and fork evidence
   - Technology stack: Go backend, React frontend, PostgreSQL
   - Architecture overview with diagrams

4. **Team Workflow & Repository Management** (2-3 pages) — include screenshots and commands:
   - Fork and clone process: `git clone https://github.com/YourTeam/taskcafe.git`
   - Branch strategy: `git checkout -b feature/authentication-enhancement`
   - PR workflow with merge strategy evidence
   - Git log showing individual contributions
   - Merge conflict resolution examples

5. **Build Automation** (Student 1) (2-3 pages):
   - Makefile implementation and usage
   - Build script automation: `./scripts/build.sh`
   - Go modules management
   - Sample output: `make build` execution logs
   - Authentication enhancements implemented

6. **Test Automation** (Student 2) (2-3 pages):
   - Go backend testing with coverage reports
   - React frontend testing with Jest
   - Sample output: `make test` execution
   - Project template system implementation
   - Coverage reports and analysis

7. **Deployment Automation** (Student 3) (2-3 pages):
   - Multi-stage Dockerfile analysis
   - Production docker-compose setup
   - Deployment script walkthrough: `./scripts/deploy.sh`
   - Analytics system implementation
   - Environment configuration management

8. **CI/CD Pipeline** (Student 4) (3-4 pages):
   - GitHub Actions workflow explanation
   - Pipeline stages: lint → test → build → deploy
   - Security scanning integration
   - Successful run screenshots with status badges
   - Artifacts and deployment evidence

9. **Challenges and Team Resolution** (1-2 pages):
   - Technical challenges faced (e.g., Go/React integration)
   - Merge conflicts and resolution strategies
   - Testing challenges with PostgreSQL
   - Team coordination and communication

10. **Individual Contributions** (1-2 pages):
    - Who did what with commit ranges and PR links
    - Code review participation
    - Documentation contributions
    - Problem-solving contributions

11. **Appendix** (2-3 pages):
    - Environment variables and secrets documentation
    - Command reference for running locally
    - Deployment instructions
    - Architecture diagrams
    - Code snippets and configurations

---

## Required secrets to set in GitHub repository

Navigate to Repository Settings → Secrets and variables → Actions:

**Staging Environment:**
- `STAGING_HOST` - Staging server hostname/IP
- `STAGING_USER` - SSH username for staging deployment
- `STAGING_SSH_KEY` - Private SSH key for staging server
- `STAGING_DB_PASSWORD` - PostgreSQL password for staging
- `STAGING_JWT_SECRET` - JWT secret for staging (generate with PowerShell command shown earlier)

**Production Environment:**
- `PRODUCTION_HOST` - Production server hostname/IP  
- `PRODUCTION_USER` - SSH username for production deployment
- `PRODUCTION_SSH_KEY` - Private SSH key for production server
- `PRODUCTION_DB_PASSWORD` - PostgreSQL password for production
- `PRODUCTION_JWT_SECRET` - JWT secret for production

**Security Scanning:**
- `SNYK_TOKEN` - Snyk security scanning token (optional)
- `CODECOV_TOKEN` - Code coverage reporting token (optional)

**Container Registry:**
- `GITHUB_TOKEN` - Automatically provided by GitHub for container registry

---

## Notes & tips
- Keep branch names descriptive: `feature/auth-enhancement`, `feature/project-templates`, `feature/analytics-dashboard`, `feature/ci-cd-pipeline`
- Each student must open at least one PR with meaningful commits and comprehensive tests
- Use Go's built-in testing for backend, Jest/React Testing Library for frontend
- All secrets must be properly documented but never committed to repository
- Include evidence of successful pipeline runs with green checkmarks
- Docker images should be optimized for production with multi-stage builds
- Database migrations must be automated and reversible

---

## Final deliverables (per group)
1. **GitHub Repository** with 4 feature PRs merged to main/develop branch
2. **Build Automation**: `Makefile` and build scripts in repository root
3. **Docker Setup**: `Dockerfile.production` and `docker-compose.production.yml`
4. **CI/CD Pipeline**: `.github/workflows/` with comprehensive automation
5. **Test Reports**: Coverage reports and test artifacts from pipeline runs
6. **Documentation**: Updated README with setup and deployment instructions
7. **Group Report PDF**: Named `UECS2363A2_GROUP<NUMBER>.pdf` (submit to WBLE)

---

## Comprehensive project structure after implementation:
```
taskcafe/
├── .github/
│   ├── workflows/
│   │   ├── ci.yml
│   │   ├── security.yml
│   │   └── deploy.yml
│   └── dependabot.yml
├── scripts/
│   ├── build.sh
│   ├── deploy.sh
│   └── test-integration.sh
├── internal/
│   ├── handler/
│   │   ├── password_reset.go (Student 1)
│   │   ├── project_template.go (Student 2)
│   │   ├── analytics.go (Student 3)
│   │   └── *_test.go files
│   └── db/query/
│       ├── project_templates.sql (Student 2)
│       └── analytics.sql (Student 3)
├── frontend/
│   └── src/components/
│       ├── ProjectTemplates/ (Student 2)
│       └── Analytics/ (Student 3)
├── Dockerfile.production (Student 3)
├── docker-compose.production.yml (Student 3)
├── Makefile (Student 1)
└── README.md (updated by all)
```

---

If you want, I can now:
- create the files in this repo (one per student) as drafts, or
- produce the report skeleton PDF template and sample screenshots commands.

Which do you want me to do next?  
