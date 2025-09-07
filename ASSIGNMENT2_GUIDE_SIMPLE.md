# UECS2363 — Assignment 2: Simple CI/CD Guide (Team of 4)

**Goal**: Fork Taskcafe, add CI/CD automation, and implement 4 simple enhancements.

**Tech Stack**: Go backend + React frontend (keep it simple!)

---

## Quick Setup Checklist
- [ ] Fork Taskcafe repository to your team account
- [ ] Each student creates their own feature branch
- [ ] Add simple automation scripts (Makefile)
- [ ] Add basic CI pipeline (.github/workflows)
- [ ] Each student implements 1 small feature + tests
- [ ] Submit 4 PRs + team report (≤15 pages)

---

## Team Split (4 Students)

### Student 1: Build Automation + Auth Enhancement
**What to do:**
- Add `Makefile` for build automation
- Enhance existing login with input validation
- Add simple password reset endpoint

**Files to create/edit:**
- `Makefile` (build commands)
- `internal/handler/auth.go` (enhance existing)
- `internal/handler/auth_test.go` (simple test)

### Student 2: Task Categories + Test Automation  
**What to do:**
- Add task priority system (High/Medium/Low)
- Set up automated testing
- Write tests for the priority feature

**Files to create/edit:**
- `internal/handler/priority.go` (new)
- `internal/handler/priority_test.go` (tests)
- Update existing task creation to include priority

### Student 3: Simple Reports + Docker
**What to do:**
- Add basic project progress report
- Create production Docker setup
- Simple deployment script

**Files to create/edit:**
- `internal/handler/reports.go` (basic project stats)
- `Dockerfile` (production build)
- `docker-compose.yml` (simple setup)

### Student 4: CI/CD Pipeline
**What to do:**
- Create GitHub Actions workflow
- Automated testing and building
- Simple deployment to staging

**Files to create/edit:**
- `.github/workflows/ci.yml` (main pipeline)
- `.github/dependabot.yml` (dependency updates)

---

## Student 1: Build + Auth (Simple Version)

**Makefile** (project root):
```makefile
.PHONY: build test clean

build:
	@echo "Building Taskcafe..."
	cd frontend && npm ci && npm run build
	go mod download
	go build -o build/taskcafe ./cmd/taskcafe

test:
	@echo "Running tests..."
	go test ./...
	cd frontend && npm test -- --watchAll=false

clean:
	rm -rf build/
	rm -rf frontend/build/

dev:
	docker-compose up --build
```

**Simple auth enhancement** (`internal/handler/auth.go` - add to existing file):
```go
// Add this to existing auth.go file
func (h *AuthHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Email string `json:"email"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    // Simple validation
    if req.Email == "" {
        http.Error(w, "Email required", http.StatusBadRequest)
        return
    }
    
    // In real implementation, you'd send email
    // For demo, just return success
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "message": "Reset instructions sent if email exists",
    })
}
```

**Simple test** (`internal/handler/auth_test.go`):
```go
package handler

import (
    "bytes"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestPasswordReset(t *testing.T) {
    handler := &AuthHandler{}
    
    body := bytes.NewBufferString(`{"email":"test@example.com"}`)
    req := httptest.NewRequest("POST", "/auth/reset", body)
    w := httptest.NewRecorder()
    
    handler.RequestPasswordReset(w, req)
    
    if w.Code != http.StatusOK {
        t.Errorf("Expected status 200, got %d", w.Code)
    }
}
```

---

## Student 2: Task Priorities + Tests

**Simple priority handler** (`internal/handler/priority.go`):
```go
package handler

import (
    "encoding/json"
    "net/http"
    "strconv"
    
    "github.com/go-chi/chi/v5"
)

type PriorityHandler struct {
    // Use existing repo
}

func (h *PriorityHandler) SetTaskPriority(w http.ResponseWriter, r *http.Request) {
    taskID := chi.URLParam(r, "taskID")
    
    var req struct {
        Priority string `json:"priority"` // "high", "medium", "low"
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    // Validate priority
    validPriorities := map[string]bool{
        "high": true, "medium": true, "low": true,
    }
    
    if !validPriorities[req.Priority] {
        http.Error(w, "Invalid priority. Use: high, medium, low", http.StatusBadRequest)
        return
    }
    
    // For demo - in real app you'd update database
    response := map[string]string{
        "task_id": taskID,
        "priority": req.Priority,
        "message": "Priority updated successfully",
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
```

**Simple test** (`internal/handler/priority_test.go`):
```go
package handler

import (
    "bytes"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/go-chi/chi/v5"
)

func TestSetTaskPriority(t *testing.T) {
    handler := &PriorityHandler{}
    
    router := chi.NewRouter()
    router.Put("/tasks/{taskID}/priority", handler.SetTaskPriority)
    
    body := bytes.NewBufferString(`{"priority":"high"}`)
    req := httptest.NewRequest("PUT", "/tasks/123/priority", body)
    w := httptest.NewRecorder()
    
    router.ServeHTTP(w, req)
    
    if w.Code != http.StatusOK {
        t.Errorf("Expected status 200, got %d", w.Code)
    }
}
```

---

## Student 3: Simple Reports + Docker

**Basic reports** (`internal/handler/reports.go`):
```go
package handler

import (
    "encoding/json"
    "net/http"
    
    "github.com/go-chi/chi/v5"
)

type ReportsHandler struct {
    // Use existing repo
}

type ProjectStats struct {
    ProjectID     string `json:"project_id"`
    TotalTasks    int    `json:"total_tasks"`
    CompletedTasks int   `json:"completed_tasks"`
    Progress      float64 `json:"progress"`
}

func (h *ReportsHandler) GetProjectStats(w http.ResponseWriter, r *http.Request) {
    projectID := chi.URLParam(r, "projectID")
    
    // For demo - mock data (in real app, query database)
    stats := ProjectStats{
        ProjectID:      projectID,
        TotalTasks:     10,
        CompletedTasks: 7,
        Progress:       70.0,
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(stats)
}
```

**Simple Dockerfile**:
```dockerfile
# Simple production build
FROM node:18-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

FROM golang:1.19-alpine AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o taskcafe ./cmd/taskcafe

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=backend /app/taskcafe .
COPY --from=frontend /app/frontend/build ./web/
CMD ["./taskcafe"]
```

**Simple docker-compose.yml**:
```yaml
version: '3.8'
services:
  app:
    build: .
    ports:
      - "3333:3333"
    environment:
      - DATABASE_URL=postgres://user:pass@db:5432/taskcafe
    depends_on:
      - db
      
  db:
    image: postgres:14
    environment:
      - POSTGRES_DB=taskcafe
      - POSTGRES_USER=user
      - POSTGRES_PASSWORD=pass
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

---

## Student 4: Simple CI/CD

**GitHub Actions CI** (`.github/workflows/ci.yml`):
```yaml
name: Simple CI/CD

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.19'
          
      - name: Setup Node
        uses: actions/setup-node@v4
        with:
          node-version: '18'
          
      - name: Install dependencies
        run: |
          go mod download
          cd frontend && npm ci
          
      - name: Run tests
        run: make test
        
      - name: Build application
        run: make build

  build-docker:
    needs: test
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v4
      
      - name: Build Docker image
        run: docker build -t taskcafe:latest .
        
      - name: Test Docker image
        run: |
          docker run -d --name test-app taskcafe:latest
          sleep 5
          docker logs test-app
          docker stop test-app
```

**Dependabot** (`.github/dependabot.yml`):
```yaml
version: 2
updates:
  - package-ecosystem: "gomod"
    directory: "/"
    schedule:
      interval: "weekly"
      
  - package-ecosystem: "npm"
    directory: "/frontend"
    schedule:
      interval: "weekly"
```

---

## Simple Report Template (≤15 pages)

### Structure:
1. **Cover + Team Info** (1 page)
2. **Project Overview** (1 page) - what you built
3. **Git Workflow** (2 pages) - screenshots of PRs and branches
4. **Build Automation** (2 pages) - Student 1's Makefile + auth
5. **Testing** (2 pages) - Student 2's tests + priority feature  
6. **Deployment** (2 pages) - Student 3's Docker + reports
7. **CI/CD** (2 pages) - Student 4's pipeline screenshots
8. **Team Contributions** (1 page) - who did what
9. **Challenges & Solutions** (1 page)
10. **Appendix** (1 page) - commands to run locally

---

## Local Testing Commands

```powershell
# Clone and setup
git clone https://github.com/YourTeam/taskcafe.git
cd taskcafe

# Build everything
make build

# Run tests
make test

# Start with Docker
docker-compose up --build

# Test the API
curl http://localhost:3333/health
```

---

## What Each Student Delivers

**Student 1**: Makefile + password reset endpoint + test
**Student 2**: Task priority system + comprehensive tests  
**Student 3**: Project stats API + Docker setup
**Student 4**: GitHub Actions pipeline + dependency management

**Team**: 4 PRs merged + working CI pipeline + team report

---

## Key Simplifications Made:
- ✅ Removed complex authentication system (just add password reset)
- ✅ Simplified task management (just add priority field)
- ✅ Basic reports instead of complex analytics
- ✅ Simple Docker setup instead of multi-stage optimization
- ✅ Basic CI pipeline instead of comprehensive security scanning
- ✅ Shorter report (15 pages vs 20)
- ✅ Mock data instead of complex database integration
- ✅ Focus on automation concepts rather than production complexity

This simplified version still demonstrates:
- **CO2 (Automation)**: Build scripts, testing, CI/CD, deployment
- **CO4 (Teamwork)**: Git workflow, PRs, collaborative development
- **Technical Skills**: Go, React, Docker, GitHub Actions

But with much less complexity and more achievable scope for students!
