# Developer Guide - Hub Git Hosting Service

This guide provides comprehensive information for developers who want to contribute to Hub, integrate with its APIs, develop plugins, or extend its functionality. Hub is built with modern technologies and follows industry best practices for maintainability and extensibility.

## Table of Contents

- [Development Environment Setup](#development-environment-setup)
- [Architecture Overview](#architecture-overview)
- [API Documentation](#api-documentation)
- [Database Schema](#database-schema)
- [Frontend Development](#frontend-development)
- [Backend Development](#backend-development)
- [Plugin Development](#plugin-development)
- [Testing](#testing)
- [Deployment and CI/CD](#deployment-and-cicd)
- [Contributing Guidelines](#contributing-guidelines)

## Development Environment Setup

### Prerequisites

#### Required Software
- **Git**: 2.30+
- **Go**: 1.21+
- **Node.js**: 18+
- **Docker**: 20.10+
- **Docker Compose**: 2.0+
- **PostgreSQL**: 12+ (for local development)
- **Redis**: 6.0+ (for local development)

#### Recommended Tools
- **IDE**: VS Code, GoLand, or similar
- **Database Tool**: pgAdmin, DBeaver, or psql
- **API Testing**: Postman, Insomnia, or curl
- **Git Client**: Command line or GUI client

### Quick Start

#### 1. Clone Repository
```bash
git clone https://github.com/a5c-ai/hub.git
cd hub
```

#### 2. Environment Setup
```bash
# Copy configuration template
cp config.example.yaml config.yaml

# Install dependencies
cd frontend && npm install && cd ..
go mod download
```

#### 3. Start Development Environment
```bash
# Start infrastructure services
docker-compose up -d postgres redis

# Run database migrations
go run cmd/migrate/main.go up

# Start backend (in one terminal)
./scripts/dev-run.sh backend

# Start frontend (in another terminal)
./scripts/dev-run.sh frontend
```

#### 4. Verify Setup
- Backend: http://localhost:8080/health
- Frontend: http://localhost:3000
- API Docs: http://localhost:8080/docs

### Development Configuration

#### Environment Variables
```bash
# .env.development
APP_ENV=development
LOG_LEVEL=debug
DATABASE_URL=postgresql://hub:hub@localhost:5432/hub?sslmode=disable
REDIS_URL=redis://localhost:6379/0
JWT_SECRET=dev-secret-key
GIT_DATA_PATH=./data/repositories
FRONTEND_URL=http://localhost:3000

# Azure Blob Storage Configuration
AZURE_STORAGE_ACCOUNT_NAME=hubstorage
AZURE_STORAGE_ACCOUNT_KEY=your-account-key
AZURE_STORAGE_CONTAINER_NAME=artifacts
# Optional: custom endpoint (useful for emulator)
AZURE_STORAGE_ENDPOINT_URL=
# SMTP / Email Configuration
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=your-smtp-username
SMTP_PASSWORD=your-smtp-password
SMTP_FROM=noreply@example.com
SMTP_USE_TLS=true

# Application settings for email links
BASE_URL=http://localhost:3000
APPLICATION_NAME=Hub
```

#### IDE Configuration

##### VS Code Settings
```json
{
  "go.testFlags": ["-v"],
  "go.testTimeout": "60s",
  "eslint.workingDirectories": ["frontend"],
  "typescript.preferences.importModuleSpecifier": "relative",
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.fixAll.eslint": true
  }
}
```

##### Go Tools
```bash
# Install development tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/swaggo/swag/cmd/swag@latest
go install github.com/air-verse/air@latest
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

## Architecture Overview

### System Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Load Balancer │────│   Ingress       │────│   CDN           │
│   (NGINX)       │    │   Controller    │    │   (Optional)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   Backend       │    │   Git Storage   │
│   (Next.js)     │◄───│   (Go/Gin)      │◄───│   (Bare Repos)  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Browser       │    │   Database      │    │   Cache         │
│   (React SPA)   │    │   (PostgreSQL)  │    │   (Redis)       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Tech Stack

#### Backend
- **Language**: Go 1.21+
- **Framework**: Gin HTTP framework
- **Database**: PostgreSQL with GORM ORM
- **Cache**: Redis for sessions and caching
- **Authentication**: JWT tokens, OAuth providers
- **Storage**: Local filesystem, S3, Azure Blob
- **Testing**: Go testing, Testify, Ginkgo

#### Frontend
- **Framework**: Next.js 15 with React 19
- **Language**: TypeScript
- **Styling**: Tailwind CSS
- **State Management**: Zustand
- **API Client**: SWR with Axios
- **Testing**: Jest, React Testing Library
- **Build**: Turbopack for development

#### Infrastructure
- **Containerization**: Docker and Docker Compose
- **Orchestration**: Kubernetes with Helm
- **CI/CD**: GitHub Actions compatible
- **Monitoring**: Prometheus metrics, structured logging
- **Documentation**: OpenAPI/Swagger

### Project Structure

```
hub/
├── cmd/                    # Application entry points
│   ├── server/            # Main server application
│   └── migrate/           # Database migration tool
├── internal/              # Private application code
│   ├── api/              # HTTP handlers and routes
│   ├── auth/             # Authentication logic
│   ├── config/           # Configuration management
│   ├── db/               # Database models and migrations
│   ├── middleware/       # HTTP middleware
│   └── models/           # Domain models
├── frontend/             # Next.js frontend application
│   ├── src/
│   │   ├── app/         # Next.js app router pages
│   │   ├── components/  # React components
│   │   ├── lib/        # Utility functions
│   │   ├── store/      # State management
│   │   └── types/      # TypeScript type definitions
│   └── public/         # Static assets
├── k8s/                 # Kubernetes manifests
├── terraform/           # Infrastructure as Code
├── scripts/            # Build and deployment scripts
├── docs/               # Documentation
└── tests/              # Integration and E2E tests
```

## API Documentation

### REST API Overview

The Hub API follows RESTful principles and is documented using OpenAPI 3.0 specification.

#### Base URL
```
Development: http://localhost:8080/api/v1
Production: https://hub.yourdomain.com/api/v1
```

#### Authentication
```bash
# Using personal access token
curl -H "Authorization: Bearer YOUR_TOKEN" \
  https://hub.yourdomain.com/api/v1/user

# Using session cookie (for web interface)
curl -b "session=YOUR_SESSION_COOKIE" \
  https://hub.yourdomain.com/api/v1/user
```

### Core API Endpoints

#### User Management
```http
GET    /api/v1/user                 # Get current user
PUT    /api/v1/user                 # Update current user
GET    /api/v1/users/:username      # Get user by username
GET    /api/v1/user/repos           # Get user repositories
GET    /api/v1/user/orgs            # Get user organizations
```

#### Repository Management
```http
GET    /api/v1/repos                # List repositories
POST   /api/v1/repos                # Create repository
GET    /api/v1/repos/:owner/:repo   # Get repository
PUT    /api/v1/repos/:owner/:repo   # Update repository
DELETE /api/v1/repos/:owner/:repo   # Delete repository

GET    /api/v1/repos/:owner/:repo/branches    # List branches
GET    /api/v1/repos/:owner/:repo/commits     # List commits
GET    /api/v1/repos/:owner/:repo/contents/*  # Get file contents
```

#### Issue Management
```http
GET    /api/v1/repos/:owner/:repo/issues      # List issues
POST   /api/v1/repos/:owner/:repo/issues      # Create issue
GET    /api/v1/repos/:owner/:repo/issues/:id  # Get issue
PUT    /api/v1/repos/:owner/:repo/issues/:id  # Update issue
DELETE /api/v1/repos/:owner/:repo/issues/:id  # Delete issue

POST   /api/v1/repos/:owner/:repo/issues/:id/comments  # Add comment
```

#### Pull Request Management
```http
GET    /api/v1/repos/:owner/:repo/pulls       # List pull requests
POST   /api/v1/repos/:owner/:repo/pulls       # Create pull request
GET    /api/v1/repos/:owner/:repo/pulls/:id   # Get pull request
PUT    /api/v1/repos/:owner/:repo/pulls/:id   # Update pull request

POST   /api/v1/repos/:owner/:repo/pulls/:id/reviews    # Create review
GET    /api/v1/repos/:owner/:repo/pulls/:id/files      # Get changed files
PUT    /api/v1/repos/:owner/:repo/pulls/:id/merge      # Merge pull request
```

### API Examples

#### Create Repository
```bash
curl -X POST https://hub.yourdomain.com/api/v1/repos \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-project",
    "description": "My awesome project",
    "private": false,
    "auto_init": true,
    "gitignore_template": "Go",
    "license_template": "MIT"
  }'
```

#### Create Issue
```bash
curl -X POST https://hub.yourdomain.com/api/v1/repos/owner/repo/issues \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Bug in authentication",
    "body": "Authentication fails when...",
    "labels": ["bug", "priority-high"],
    "assignees": ["developer1", "developer2"]
  }'
```

#### Create Pull Request
```bash
curl -X POST https://hub.yourdomain.com/api/v1/repos/owner/repo/pulls \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Fix authentication bug",
    "body": "This PR fixes the authentication issue by...",
    "head": "fix/auth-bug",
    "base": "main",
    "draft": false
  }'
```

### GraphQL API

#### GraphQL Endpoint
```
POST /api/graphql
```

#### Example Query
```graphql
query GetRepository($owner: String!, $name: String!) {
  repository(owner: $owner, name: $name) {
    id
    name
    description
    isPrivate
    defaultBranch
    stargazerCount
    forkCount
    issues(first: 10) {
      nodes {
        id
        title
        state
        createdAt
        author {
          login
        }
      }
    }
    pullRequests(first: 10) {
      nodes {
        id
        title
        state
        createdAt
        author {
          login
        }
      }
    }
  }
}
```

#### Example Mutation
```graphql
mutation CreateRepository($input: CreateRepositoryInput!) {
  createRepository(input: $input) {
    repository {
      id
      name
      url
    }
    errors {
      field
      message
    }
  }
}
```

## Database Schema

### Core Tables

#### Users Table
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255),
    full_name VARCHAR(255),
    avatar_url TEXT,
    bio TEXT,
    location VARCHAR(255),
    website VARCHAR(255),
    company VARCHAR(255),
    public_repos INTEGER DEFAULT 0,
    private_repos INTEGER DEFAULT 0,
    followers_count INTEGER DEFAULT 0,
    following_count INTEGER DEFAULT 0,
    is_admin BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    email_verified BOOLEAN DEFAULT FALSE,
    two_factor_enabled BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

#### Organizations Table
```sql
CREATE TABLE organizations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    display_name VARCHAR(255),
    description TEXT,
    avatar_url TEXT,
    website VARCHAR(255),
    location VARCHAR(255),
    billing_email VARCHAR(255),
    public_repos INTEGER DEFAULT 0,
    private_repos INTEGER DEFAULT 0,
    members_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

#### Repositories Table
```sql
CREATE TABLE repositories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    owner_id INTEGER REFERENCES users(id),
    organization_id INTEGER REFERENCES organizations(id),
    is_private BOOLEAN DEFAULT FALSE,
    is_fork BOOLEAN DEFAULT FALSE,
    parent_id INTEGER REFERENCES repositories(id),
    default_branch VARCHAR(255) DEFAULT 'main',
    size INTEGER DEFAULT 0,
    stars_count INTEGER DEFAULT 0,
    forks_count INTEGER DEFAULT 0,
    watchers_count INTEGER DEFAULT 0,
    open_issues_count INTEGER DEFAULT 0,
    language VARCHAR(100),
    has_issues BOOLEAN DEFAULT TRUE,
    has_projects BOOLEAN DEFAULT TRUE,
    has_wiki BOOLEAN DEFAULT TRUE,
    archived BOOLEAN DEFAULT FALSE,
    disabled BOOLEAN DEFAULT FALSE,
    pushed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    CONSTRAINT repositories_owner_check 
        CHECK ((owner_id IS NOT NULL) != (organization_id IS NOT NULL))
);
```

#### Issues Table
```sql
CREATE TABLE issues (
    id SERIAL PRIMARY KEY,
    number INTEGER NOT NULL,
    repository_id INTEGER REFERENCES repositories(id),
    title VARCHAR(500) NOT NULL,
    body TEXT,
    state VARCHAR(50) DEFAULT 'open',
    author_id INTEGER REFERENCES users(id),
    assignee_id INTEGER REFERENCES users(id),
    milestone_id INTEGER REFERENCES milestones(id),
    comments_count INTEGER DEFAULT 0,
    locked BOOLEAN DEFAULT FALSE,
    closed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(repository_id, number)
);
```

#### Pull Requests Table
```sql
CREATE TABLE pull_requests (
    id SERIAL PRIMARY KEY,
    number INTEGER NOT NULL,
    repository_id INTEGER REFERENCES repositories(id),
    title VARCHAR(500) NOT NULL,
    body TEXT,
    state VARCHAR(50) DEFAULT 'open',
    author_id INTEGER REFERENCES users(id),
    head_branch VARCHAR(255) NOT NULL,
    base_branch VARCHAR(255) NOT NULL,
    head_sha VARCHAR(40),
    base_sha VARCHAR(40),
    merge_commit_sha VARCHAR(40),
    merged BOOLEAN DEFAULT FALSE,
    mergeable BOOLEAN,
    merged_by_id INTEGER REFERENCES users(id),
    merged_at TIMESTAMP,
    draft BOOLEAN DEFAULT FALSE,
    commits_count INTEGER DEFAULT 0,
    additions INTEGER DEFAULT 0,
    deletions INTEGER DEFAULT 0,
    changed_files INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(repository_id, number)
);
```

### Database Migrations

#### Creating Migrations
```bash
# Create new migration
migrate create -ext sql -dir internal/db/migrations -seq add_user_preferences

# This creates:
# internal/db/migrations/000003_add_user_preferences.up.sql
# internal/db/migrations/000003_add_user_preferences.down.sql
```

#### Migration Example
```sql
-- 000003_add_user_preferences.up.sql
CREATE TABLE user_preferences (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    theme VARCHAR(20) DEFAULT 'light',
    language VARCHAR(10) DEFAULT 'en',
    timezone VARCHAR(50) DEFAULT 'UTC',
    email_notifications BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(user_id)
);

-- 000003_add_user_preferences.down.sql
DROP TABLE user_preferences;
```

#### Running Migrations
```bash
# Run migrations
go run cmd/migrate/main.go up

# Rollback migration
go run cmd/migrate/main.go down 1

# Check migration status
go run cmd/migrate/main.go version
```

## Frontend Development

### Component Structure

#### Component Organization
```
src/components/
├── ui/                   # Basic UI components
│   ├── Button.tsx
│   ├── Input.tsx
│   ├── Modal.tsx
│   └── index.ts
├── forms/               # Form components
│   ├── LoginForm.tsx
│   ├── RegisterForm.tsx
│   └── RepositoryForm.tsx
├── layout/              # Layout components
│   ├── Header.tsx
│   ├── Sidebar.tsx
│   └── AppLayout.tsx
└── features/            # Feature-specific components
    ├── repositories/
    ├── issues/
    ├── pull-requests/
    └── users/
```

#### Example Component
```typescript
// src/components/ui/Button.tsx
import { forwardRef, ButtonHTMLAttributes } from 'react'
import { clsx } from 'clsx'

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'danger'
  size?: 'sm' | 'md' | 'lg'
  loading?: boolean
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = 'primary', size = 'md', loading, children, disabled, ...props }, ref) => {
    return (
      <button
        ref={ref}
        className={clsx(
          'inline-flex items-center justify-center rounded-md font-medium transition-colors',
          'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
          'disabled:pointer-events-none disabled:opacity-50',
          {
            'bg-primary text-primary-foreground hover:bg-primary/90': variant === 'primary',
            'bg-secondary text-secondary-foreground hover:bg-secondary/80': variant === 'secondary',
            'bg-destructive text-destructive-foreground hover:bg-destructive/90': variant === 'danger',
            'h-9 px-3 text-sm': size === 'sm',
            'h-10 px-4 py-2': size === 'md',
            'h-11 px-8 text-lg': size === 'lg'
          },
          className
        )}
        disabled={disabled || loading}
        {...props}
      >
        {loading && (
          <svg className="mr-2 h-4 w-4 animate-spin" viewBox="0 0 24 24">
            <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" opacity="0.25" />
            <path fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
          </svg>
        )}
        {children}
      </button>
    )
  }
)

Button.displayName = 'Button'
```

### State Management

#### Zustand Store Example
```typescript
// src/store/auth.ts
import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface User {
  id: number
  username: string
  email: string
  avatar_url?: string
  is_admin: boolean
}

interface AuthState {
  user: User | null
  token: string | null
  isAuthenticated: boolean
  login: (token: string, user: User) => void
  logout: () => void
  updateUser: (updates: Partial<User>) => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      token: null,
      isAuthenticated: false,

      login: (token: string, user: User) => {
        set({
          token,
          user,
          isAuthenticated: true
        })
      },

      logout: () => {
        set({
          token: null,
          user: null,
          isAuthenticated: false
        })
      },

      updateUser: (updates: Partial<User>) => {
        const { user } = get()
        if (user) {
          set({
            user: { ...user, ...updates }
          })
        }
      }
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        token: state.token,
        user: state.user,
        isAuthenticated: state.isAuthenticated
      })
    }
  )
)
```

### API Integration

#### API Client Setup
```typescript
// src/lib/api.ts
import axios, { AxiosRequestConfig } from 'axios'
import { useAuthStore } from '@/store/auth'

const baseURL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1'

export const api = axios.create({
  baseURL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// Request interceptor to add auth token
api.interceptors.request.use((config) => {
  const token = useAuthStore.getState().token
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Response interceptor for error handling
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      useAuthStore.getState().logout()
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// API functions
export const apiClient = {
  // User APIs
  user: {
    me: () => api.get('/user'),
    update: (data: any) => api.put('/user', data),
    repos: () => api.get('/user/repos')
  },

  // Repository APIs
  repositories: {
    list: (params?: any) => api.get('/repos', { params }),
    get: (owner: string, repo: string) => api.get(`/repos/${owner}/${repo}`),
    create: (data: any) => api.post('/repos', data),
    update: (owner: string, repo: string, data: any) => api.put(`/repos/${owner}/${repo}`, data),
    delete: (owner: string, repo: string) => api.delete(`/repos/${owner}/${repo}`)
  },

  // Issue APIs
  issues: {
    list: (owner: string, repo: string, params?: any) => 
      api.get(`/repos/${owner}/${repo}/issues`, { params }),
    get: (owner: string, repo: string, number: number) => 
      api.get(`/repos/${owner}/${repo}/issues/${number}`),
    create: (owner: string, repo: string, data: any) => 
      api.post(`/repos/${owner}/${repo}/issues`, data),
    update: (owner: string, repo: string, number: number, data: any) => 
      api.put(`/repos/${owner}/${repo}/issues/${number}`, data)
  }
}
```

#### SWR Data Fetching
```typescript
// src/hooks/useRepositories.ts
import useSWR from 'swr'
import { apiClient } from '@/lib/api'

export function useRepositories() {
  const { data, error, isLoading, mutate } = useSWR(
    '/user/repos',
    () => apiClient.user.repos().then(res => res.data),
    {
      revalidateOnFocus: false,
      dedupingInterval: 60000 // 1 minute
    }
  )

  return {
    repositories: data || [],
    isLoading,
    isError: error,
    refresh: mutate
  }
}

export function useRepository(owner: string, repo: string) {
  const { data, error, isLoading } = useSWR(
    owner && repo ? `/repos/${owner}/${repo}` : null,
    () => apiClient.repositories.get(owner, repo).then(res => res.data)
  )

  return {
    repository: data,
    isLoading,
    isError: error
  }
}
```

## Backend Development

### Project Structure

#### Handler Example
```go
// internal/api/repositories.go
package api

import (
    "net/http"
    "strconv"
    
    "github.com/gin-gonic/gin"
    "github.com/a5c-ai/hub/internal/models"
    "github.com/a5c-ai/hub/internal/db"
)

type RepositoryHandler struct {
    db *db.DB
}

func NewRepositoryHandler(database *db.DB) *RepositoryHandler {
    return &RepositoryHandler{db: database}
}

// GetRepository godoc
// @Summary Get repository
// @Description Get repository by owner and name
// @Tags repositories
// @Accept json
// @Produce json
// @Param owner path string true "Repository owner"
// @Param repo path string true "Repository name"
// @Success 200 {object} models.Repository
// @Failure 404 {object} ErrorResponse
// @Router /repos/{owner}/{repo} [get]
func (h *RepositoryHandler) GetRepository(c *gin.Context) {
    owner := c.Param("owner")
    repo := c.Param("repo")
    
    repository, err := h.db.GetRepositoryByName(owner, repo)
    if err != nil {
        c.JSON(http.StatusNotFound, ErrorResponse{
            Error:   "Repository not found",
            Message: err.Error(),
        })
        return
    }
    
    // Check if user has access to repository
    user, exists := c.Get("user")
    if !exists || !h.hasRepositoryAccess(user.(*models.User), repository) {
        c.JSON(http.StatusForbidden, ErrorResponse{
            Error: "Access denied",
        })
        return
    }
    
    c.JSON(http.StatusOK, repository)
}

// CreateRepository godoc
// @Summary Create repository
// @Description Create a new repository
// @Tags repositories
// @Accept json
// @Produce json
// @Param repository body CreateRepositoryRequest true "Repository data"
// @Success 201 {object} models.Repository
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /repos [post]
func (h *RepositoryHandler) CreateRepository(c *gin.Context) {
    var req CreateRepositoryRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{
            Error:   "Invalid request",
            Message: err.Error(),
        })
        return
    }
    
    user := c.MustGet("user").(*models.User)
    
    repository := &models.Repository{
        Name:        req.Name,
        Description: req.Description,
        IsPrivate:   req.Private,
        OwnerID:     user.ID,
        FullName:    user.Username + "/" + req.Name,
    }
    
    if err := h.db.CreateRepository(repository); err != nil {
        if errors.Is(err, db.ErrRepositoryExists) {
            c.JSON(http.StatusConflict, ErrorResponse{
                Error: "Repository already exists",
            })
            return
        }
        
        c.JSON(http.StatusInternalServerError, ErrorResponse{
            Error:   "Failed to create repository",
            Message: err.Error(),
        })
        return
    }
    
    // Initialize Git repository if requested
    if req.AutoInit {
        if err := h.initializeGitRepository(repository); err != nil {
            // Log error but don't fail the request
            log.Printf("Failed to initialize git repository: %v", err)
        }
    }
    
    c.JSON(http.StatusCreated, repository)
}

func (h *RepositoryHandler) hasRepositoryAccess(user *models.User, repo *models.Repository) bool {
    // Admin users have access to everything
    if user.IsAdmin {
        return true
    }
    
    // Owner has full access
    if repo.OwnerID == user.ID {
        return true
    }
    
    // Public repositories are accessible to everyone
    if !repo.IsPrivate {
        return true
    }
    
    // Check organization membership for private repos
    if repo.OrganizationID != nil {
        return h.db.IsOrganizationMember(user.ID, *repo.OrganizationID)
    }
    
    // Check collaborator access
    return h.db.IsRepositoryCollaborator(user.ID, repo.ID)
}
```

#### Model Example
```go
// internal/models/repository.go
package models

import (
    "time"
    "gorm.io/gorm"
)

type Repository struct {
    ID             uint           `json:"id" gorm:"primarykey"`
    Name           string         `json:"name" gorm:"not null"`
    FullName       string         `json:"full_name" gorm:"uniqueIndex;not null"`
    Description    string         `json:"description"`
    OwnerID        uint           `json:"owner_id"`
    Owner          *User          `json:"owner,omitempty" gorm:"foreignKey:OwnerID"`
    OrganizationID *uint          `json:"organization_id"`
    Organization   *Organization  `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
    IsPrivate      bool           `json:"private" gorm:"default:false"`
    IsFork         bool           `json:"fork" gorm:"default:false"`
    ParentID       *uint          `json:"parent_id"`
    Parent         *Repository    `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
    DefaultBranch  string         `json:"default_branch" gorm:"default:main"`
    Size           int64          `json:"size" gorm:"default:0"`
    StarsCount     int            `json:"stargazers_count" gorm:"default:0"`
    ForksCount     int            `json:"forks_count" gorm:"default:0"`
    WatchersCount  int            `json:"watchers_count" gorm:"default:0"`
    OpenIssuesCount int           `json:"open_issues_count" gorm:"default:0"`
    Language       string         `json:"language"`
    HasIssues      bool           `json:"has_issues" gorm:"default:true"`
    HasProjects    bool           `json:"has_projects" gorm:"default:true"`
    HasWiki        bool           `json:"has_wiki" gorm:"default:true"`
    Archived       bool           `json:"archived" gorm:"default:false"`
    Disabled       bool           `json:"disabled" gorm:"default:false"`
    PushedAt       *time.Time     `json:"pushed_at"`
    CreatedAt      time.Time      `json:"created_at"`
    UpdatedAt      time.Time      `json:"updated_at"`
    DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
    
    // Associations
    Issues        []Issue        `json:"issues,omitempty" gorm:"foreignKey:RepositoryID"`
    PullRequests  []PullRequest  `json:"pull_requests,omitempty" gorm:"foreignKey:RepositoryID"`
    Collaborators []User         `json:"collaborators,omitempty" gorm:"many2many:repository_collaborators;"`
}

type CreateRepositoryRequest struct {
    Name              string `json:"name" binding:"required,min=1,max=100"`
    Description       string `json:"description" binding:"max=500"`
    Private           bool   `json:"private"`
    AutoInit          bool   `json:"auto_init"`
    GitignoreTemplate string `json:"gitignore_template"`
    LicenseTemplate   string `json:"license_template"`
}

type UpdateRepositoryRequest struct {
    Name          *string `json:"name,omitempty" binding:"omitempty,min=1,max=100"`
    Description   *string `json:"description,omitempty" binding:"omitempty,max=500"`
    DefaultBranch *string `json:"default_branch,omitempty"`
    Private       *bool   `json:"private,omitempty"`
    HasIssues     *bool   `json:"has_issues,omitempty"`
    HasProjects   *bool   `json:"has_projects,omitempty"`
    HasWiki       *bool   `json:"has_wiki,omitempty"`
    Archived      *bool   `json:"archived,omitempty"`
}

func (Repository) TableName() string {
    return "repositories"
}

// Hooks
func (r *Repository) BeforeCreate(tx *gorm.DB) error {
    if r.FullName == "" && r.Owner != nil {
        r.FullName = r.Owner.Username + "/" + r.Name
    }
    return nil
}

func (r *Repository) BeforeUpdate(tx *gorm.DB) error {
    if tx.Statement.Changed("Name") && r.Owner != nil {
        r.FullName = r.Owner.Username + "/" + r.Name
    }
    return nil
}
```

#### Database Layer
```go
// internal/db/repositories.go
package db

import (
    "errors"
    "gorm.io/gorm"
    "github.com/a5c-ai/hub/internal/models"
)

var (
    ErrRepositoryNotFound = errors.New("repository not found")
    ErrRepositoryExists   = errors.New("repository already exists")
)

func (db *DB) GetRepositoryByName(owner, name string) (*models.Repository, error) {
    var repository models.Repository
    
    err := db.Where("full_name = ?", owner+"/"+name).
        Preload("Owner").
        Preload("Organization").
        First(&repository).Error
    
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, ErrRepositoryNotFound
    }
    
    return &repository, err
}

func (db *DB) CreateRepository(repository *models.Repository) error {
    // Check if repository already exists
    var count int64
    err := db.Model(&models.Repository{}).
        Where("full_name = ?", repository.FullName).
        Count(&count).Error
    
    if err != nil {
        return err
    }
    
    if count > 0 {
        return ErrRepositoryExists
    }
    
    return db.Create(repository).Error
}

func (db *DB) UpdateRepository(repository *models.Repository) error {
    return db.Save(repository).Error
}

func (db *DB) DeleteRepository(id uint) error {
    return db.Delete(&models.Repository{}, id).Error
}

func (db *DB) ListUserRepositories(userID uint, private bool) ([]models.Repository, error) {
    var repositories []models.Repository
    
    query := db.Where("owner_id = ?", userID)
    if !private {
        query = query.Where("is_private = ?", false)
    }
    
    err := query.Preload("Owner").Find(&repositories).Error
    return repositories, err
}

func (db *DB) IsRepositoryCollaborator(userID, repoID uint) bool {
    var count int64
    db.Table("repository_collaborators").
        Where("user_id = ? AND repository_id = ?", userID, repoID).
        Count(&count)
    return count > 0
}
```

### Testing

#### Unit Test Example
```go
// internal/api/repositories_test.go
package api

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    
    "github.com/a5c-ai/hub/internal/models"
    "github.com/a5c-ai/hub/internal/db/mocks"
)

func TestRepositoryHandler_CreateRepository(t *testing.T) {
    gin.SetMode(gin.TestMode)
    
    tests := []struct {
        name           string
        request        CreateRepositoryRequest
        user           *models.User
        mockSetup      func(*mocks.MockDB)
        expectedStatus int
        expectedError  string
    }{
        {
            name: "successful creation",
            request: CreateRepositoryRequest{
                Name:        "test-repo",
                Description: "Test repository",
                Private:     false,
                AutoInit:    true,
            },
            user: &models.User{
                ID:       1,
                Username: "testuser",
            },
            mockSetup: func(mockDB *mocks.MockDB) {
                mockDB.On("CreateRepository", mock.AnythingOfType("*models.Repository")).
                    Return(nil)
            },
            expectedStatus: http.StatusCreated,
        },
        {
            name: "repository already exists",
            request: CreateRepositoryRequest{
                Name: "existing-repo",
            },
            user: &models.User{
                ID:       1,
                Username: "testuser",
            },
            mockSetup: func(mockDB *mocks.MockDB) {
                mockDB.On("CreateRepository", mock.AnythingOfType("*models.Repository")).
                    Return(db.ErrRepositoryExists)
            },
            expectedStatus: http.StatusConflict,
            expectedError:  "Repository already exists",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockDB := new(mocks.MockDB)
            tt.mockSetup(mockDB)
            
            handler := NewRepositoryHandler(mockDB)
            
            // Create request
            body, _ := json.Marshal(tt.request)
            req, _ := http.NewRequest(http.MethodPost, "/repos", bytes.NewReader(body))
            req.Header.Set("Content-Type", "application/json")
            
            // Create response recorder
            w := httptest.NewRecorder()
            
            // Create Gin context
            c, _ := gin.CreateTestContext(w)
            c.Request = req
            c.Set("user", tt.user)
            
            // Call handler
            handler.CreateRepository(c)
            
            // Assert response
            assert.Equal(t, tt.expectedStatus, w.Code)
            
            if tt.expectedError != "" {
                var response ErrorResponse
                err := json.Unmarshal(w.Body.Bytes(), &response)
                assert.NoError(t, err)
                assert.Equal(t, tt.expectedError, response.Error)
            }
            
            mockDB.AssertExpectations(t)
        })
    }
}
```

#### Integration Test Example
```go
// tests/integration/repositories_test.go
package integration

import (
    "bytes"
    "encoding/json"
    "net/http"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
    
    "github.com/a5c-ai/hub/internal/models"
)

type RepositoryTestSuite struct {
    IntegrationTestSuite
    user *models.User
}

func (suite *RepositoryTestSuite) SetupTest() {
    suite.IntegrationTestSuite.SetupTest()
    
    // Create test user
    suite.user = &models.User{
        Username: "testuser",
        Email:    "test@example.com",
        IsActive: true,
    }
    err := suite.db.CreateUser(suite.user)
    suite.Require().NoError(err)
}

func (suite *RepositoryTestSuite) TestCreateRepository() {
    // Login and get token
    token := suite.loginUser(suite.user)
    
    // Create repository request
    request := map[string]interface{}{
        "name":        "test-repo",
        "description": "Test repository",
        "private":     false,
        "auto_init":   true,
    }
    
    body, _ := json.Marshal(request)
    req, _ := http.NewRequest(http.MethodPost, "/api/v1/repos", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+token)
    
    resp, err := suite.client.Do(req)
    suite.Require().NoError(err)
    defer resp.Body.Close()
    
    assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)
    
    var repository models.Repository
    err = json.NewDecoder(resp.Body).Decode(&repository)
    suite.Require().NoError(err)
    
    assert.Equal(suite.T(), "test-repo", repository.Name)
    assert.Equal(suite.T(), "testuser/test-repo", repository.FullName)
    assert.Equal(suite.T(), "Test repository", repository.Description)
    assert.False(suite.T(), repository.IsPrivate)
}

func TestRepositoryTestSuite(t *testing.T) {
    suite.Run(t, new(RepositoryTestSuite))
}
```

## Plugin Development

### Plugin Architecture

Hub supports plugins at both the organization and repository levels. Plugins can extend functionality, integrate with external services, and customize workflows.

#### Plugin Structure
```
my-plugin/
├── plugin.yaml           # Plugin manifest
├── main.go              # Plugin entry point
├── handlers/            # HTTP handlers
├── hooks/              # Git hooks
├── templates/          # Templates and assets
└── README.md           # Plugin documentation
```

#### Plugin Manifest
```yaml
# plugin.yaml
apiVersion: v1
kind: Plugin
metadata:
  name: my-awesome-plugin
  version: "1.0.0"
  description: "An awesome plugin for Hub"
  author: "Your Name <your.email@example.com>"
  website: "https://github.com/yourname/hub-plugin"
  license: "MIT"

spec:
  runtime: go
  entry: main.go
  
  permissions:
    - repositories:read
    - repositories:write
    - issues:read
    - issues:write
    - hooks:manage
  
  hooks:
    - name: pre-receive
      script: hooks/pre-receive.sh
    - name: post-receive
      script: hooks/post-receive.sh
  
  webhooks:
    - event: push
      handler: handlers.HandlePush
    - event: pull_request
      handler: handlers.HandlePullRequest
  
  settings:
    - name: api_key
      type: string
      required: true
      secret: true
      description: "API key for external service"
    - name: enabled_repos
      type: array
      description: "List of repositories to enable plugin for"
    - name: notification_channel
      type: string
      default: "#general"
      description: "Slack channel for notifications"
  
  dependencies:
    - name: github.com/gorilla/mux
      version: "^1.8.0"
```

#### Plugin Implementation
```go
// main.go
package main

import (
    "context"
    "log"
    
    "github.com/a5c-ai/hub/pkg/plugin"
    "github.com/a5c-ai/hub/pkg/events"
)

type MyPlugin struct {
    config *plugin.Config
    client *plugin.APIClient
}

func (p *MyPlugin) Initialize(config *plugin.Config) error {
    p.config = config
    p.client = plugin.NewAPIClient(config.APIToken)
    
    log.Printf("Plugin %s initialized", config.Name)
    return nil
}

func (p *MyPlugin) HandleWebhook(ctx context.Context, event *events.Event) error {
    switch event.Type {
    case events.PushEvent:
        return p.handlePush(ctx, event)
    case events.PullRequestEvent:
        return p.handlePullRequest(ctx, event)
    default:
        log.Printf("Unhandled event type: %s", event.Type)
    }
    return nil
}

func (p *MyPlugin) handlePush(ctx context.Context, event *events.Event) error {
    push := event.Payload.(*events.PushPayload)
    
    // Get plugin settings
    apiKey := p.config.GetSetting("api_key")
    channel := p.config.GetSetting("notification_channel")
    
    // Process push event
    message := fmt.Sprintf("New push to %s by %s: %d commits",
        push.Repository.FullName,
        push.Pusher.Name,
        len(push.Commits))
    
    // Send notification (example)
    return p.sendSlackNotification(channel, message, apiKey)
}

func (p *MyPlugin) handlePullRequest(ctx context.Context, event *events.Event) error {
    pr := event.Payload.(*events.PullRequestPayload)
    
    if pr.Action == "opened" {
        // Run custom checks on new PR
        return p.runPRChecks(ctx, pr.PullRequest)
    }
    
    return nil
}

func (p *MyPlugin) runPRChecks(ctx context.Context, pr *events.PullRequest) error {
    // Example: Check if PR has proper title format
    if !p.isValidPRTitle(pr.Title) {
        return p.client.CreatePRComment(ctx, pr.Repository.Owner, pr.Repository.Name, pr.Number,
            "⚠️ Please follow the PR title format: `type: description`")
    }
    
    // Add success status check
    return p.client.CreateStatusCheck(ctx, pr.Repository.Owner, pr.Repository.Name, pr.Head.SHA, &plugin.StatusCheck{
        State:       "success",
        Context:     "my-plugin/pr-checks",
        Description: "All checks passed",
    })
}

func main() {
    p := &MyPlugin{}
    plugin.Run(p)
}
```

### Plugin API

#### Available APIs
```go
// pkg/plugin/api.go
package plugin

import (
    "context"
    "github.com/a5c-ai/hub/internal/models"
)

type APIClient struct {
    token  string
    client *http.Client
}

// Repository APIs
func (c *APIClient) GetRepository(ctx context.Context, owner, repo string) (*models.Repository, error)
func (c *APIClient) ListRepositories(ctx context.Context, opts *ListOptions) ([]*models.Repository, error)
func (c *APIClient) UpdateRepository(ctx context.Context, owner, repo string, updates map[string]interface{}) error

// Issue APIs
func (c *APIClient) GetIssue(ctx context.Context, owner, repo string, number int) (*models.Issue, error)
func (c *APIClient) CreateIssue(ctx context.Context, owner, repo string, issue *CreateIssueRequest) (*models.Issue, error)
func (c *APIClient) UpdateIssue(ctx context.Context, owner, repo string, number int, updates map[string]interface{}) error
func (c *APIClient) AddIssueComment(ctx context.Context, owner, repo string, number int, body string) error

// Pull Request APIs
func (c *APIClient) GetPullRequest(ctx context.Context, owner, repo string, number int) (*models.PullRequest, error)
func (c *APIClient) CreatePRComment(ctx context.Context, owner, repo string, number int, body string) error
func (c *APIClient) CreatePRReview(ctx context.Context, owner, repo string, number int, review *CreateReviewRequest) error

// Status Check APIs
func (c *APIClient) CreateStatusCheck(ctx context.Context, owner, repo, sha string, check *StatusCheck) error
func (c *APIClient) ListStatusChecks(ctx context.Context, owner, repo, sha string) ([]*StatusCheck, error)

// Webhook APIs
func (c *APIClient) CreateWebhook(ctx context.Context, owner, repo string, webhook *CreateWebhookRequest) error
func (c *APIClient) UpdateWebhook(ctx context.Context, owner, repo string, id int, updates map[string]interface{}) error
func (c *APIClient) DeleteWebhook(ctx context.Context, owner, repo string, id int) error
```

### Plugin Installation

#### Installing from Marketplace
```bash
# Install plugin from Hub marketplace
hub plugin install my-awesome-plugin

# Install specific version
hub plugin install my-awesome-plugin@1.0.0

# Install from URL
hub plugin install https://github.com/yourname/hub-plugin/releases/download/v1.0.0/plugin.tar.gz
```

#### Manual Installation
```bash
# Build plugin
cd my-plugin
go build -o plugin main.go

# Package plugin
tar czf my-plugin-1.0.0.tar.gz plugin plugin.yaml templates/ hooks/

# Install via API
curl -X POST https://hub.yourdomain.com/api/v1/plugins \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@my-plugin-1.0.0.tar.gz"
```

#### Plugin Configuration
```bash
# Configure plugin settings
hub plugin config my-awesome-plugin \
  --set api_key=secret-key \
  --set notification_channel=#dev-notifications \
  --set enabled_repos=repo1,repo2

# Enable plugin for organization
hub plugin enable my-awesome-plugin --org myorg

# Enable plugin for repository
hub plugin enable my-awesome-plugin --repo myorg/myrepo
```

## Testing

### Backend Testing

#### Unit Tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detection
go test -race ./...

# Run specific test
go test ./internal/api -run TestRepositoryHandler_CreateRepository

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

#### Integration Tests
```bash
# Run integration tests (requires database)
go test -tags=integration ./tests/integration/...

# Run with test database
TEST_DATABASE_URL=postgresql://test:test@localhost:5432/hub_test go test -tags=integration ./tests/integration/...
```

### Frontend Testing

#### Unit Tests
```bash
cd frontend

# Run all tests
npm test

# Run tests in watch mode
npm run test:watch

# Run tests with coverage
npm run test:ci
```

#### E2E Tests
```typescript
// tests/e2e/repository.spec.ts
import { test, expect } from '@playwright/test'

test.describe('Repository Management', () => {
  test.beforeEach(async ({ page }) => {
    // Login
    await page.goto('/login')
    await page.fill('[data-testid=username]', 'testuser')
    await page.fill('[data-testid=password]', 'password')
    await page.click('[data-testid=login-button]')
    await expect(page).toHaveURL('/dashboard')
  })

  test('should create new repository', async ({ page }) => {
    await page.click('[data-testid=new-repo-button]')
    await expect(page).toHaveURL('/repos/new')
    
    await page.fill('[data-testid=repo-name]', 'test-repo')
    await page.fill('[data-testid=repo-description]', 'Test repository')
    await page.click('[data-testid=create-repo-button]')
    
    await expect(page).toHaveURL('/testuser/test-repo')
    await expect(page.locator('h1')).toContainText('test-repo')
  })

  test('should display repository list', async ({ page }) => {
    await page.goto('/repositories')
    
    await expect(page.locator('[data-testid=repo-list]')).toBeVisible()
    await expect(page.locator('[data-testid=repo-item]')).toHaveCount.greaterThan(0)
  })
})
```

### Test Data and Fixtures

#### Database Fixtures
```go
// tests/fixtures/users.go
package fixtures

import (
    "github.com/a5c-ai/hub/internal/models"
)

func CreateTestUser(db *gorm.DB, username string) *models.User {
    user := &models.User{
        Username:      username,
        Email:        username + "@example.com",
        PasswordHash: "$2a$10$example...", // bcrypt hash of "password"
        IsActive:     true,
    }
    
    err := db.Create(user).Error
    if err != nil {
        panic(err)
    }
    
    return user
}

func CreateTestRepository(db *gorm.DB, owner *models.User, name string) *models.Repository {
    repo := &models.Repository{
        Name:        name,
        FullName:    owner.Username + "/" + name,
        Description: "Test repository",
        OwnerID:     owner.ID,
        IsPrivate:   false,
    }
    
    err := db.Create(repo).Error
    if err != nil {
        panic(err)
    }
    
    return repo
}
```

## Deployment and CI/CD

### GitHub Actions Workflow

#### CI/CD Pipeline
```yaml
# .github/workflows/ci.yml
name: CI/CD Pipeline

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

env:
  GO_VERSION: '1.21'
  NODE_VERSION: '18'

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:14
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: hub_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
      
      redis:
        image: redis:7
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ env.GO_VERSION }}
    
    - name: Set up Node.js
      uses: actions/setup-node@v4
      with:
        node-version: ${{ env.NODE_VERSION }}
        cache: 'npm'
        cache-dependency-path: frontend/package-lock.json
    
    - name: Install Go dependencies
      run: go mod download
    
    - name: Install Node.js dependencies
      run: cd frontend && npm ci
    
    - name: Run Go tests
      run: |
        go test -race -coverprofile=coverage.out ./...
        go tool cover -html=coverage.out -o coverage.html
      env:
        DATABASE_URL: postgresql://postgres:postgres@localhost:5432/hub_test?sslmode=disable
        REDIS_URL: redis://localhost:6379/0
    
    - name: Run Node.js tests
      run: cd frontend && npm run test:ci
    
    - name: Lint Go code
      uses: golangci/golangci-lint-action@v3
      with:
        version: latest
    
    - name: Lint frontend code
      run: cd frontend && npm run lint
    
    - name: Type check frontend
      run: cd frontend && npm run type-check
    
    - name: Upload coverage reports
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
        flags: backend
    
    - name: Upload frontend coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./frontend/coverage/lcov.info
        flags: frontend

  build:
    needs: test
    runs-on: ubuntu-latest
    if: github.event_name == 'push'
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v3
    
    - name: Login to Container Registry
      uses: docker/login-action@v3
      with:
        registry: ghcr.io
        username: ${{ github.actor }}
        password: ${{ secrets.GITHUB_TOKEN }}
    
    - name: Build and push backend image
      uses: docker/build-push-action@v5
      with:
        context: .
        file: ./Dockerfile
        push: true
        tags: |
          ghcr.io/${{ github.repository }}/backend:latest
          ghcr.io/${{ github.repository }}/backend:${{ github.sha }}
        build-args: |
          VERSION=${{ github.ref_name }}
          BUILD_DATE=${{ github.event.head_commit.timestamp }}
          VCS_REF=${{ github.sha }}
    
    - name: Build and push frontend image
      uses: docker/build-push-action@v5
      with:
        context: ./frontend
        file: ./frontend/Dockerfile
        push: true
        tags: |
          ghcr.io/${{ github.repository }}/frontend:latest
          ghcr.io/${{ github.repository }}/frontend:${{ github.sha }}
        build-args: |
          VERSION=${{ github.ref_name }}
          BUILD_DATE=${{ github.event.head_commit.timestamp }}
          VCS_REF=${{ github.sha }}

  deploy-staging:
    needs: build
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/develop'
    environment: staging
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Deploy to staging
      run: |
        ./scripts/deploy.sh staging
      env:
        KUBECONFIG: ${{ secrets.KUBECONFIG_STAGING }}
        IMAGE_TAG: ${{ github.sha }}

  deploy-production:
    needs: build
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    environment: production
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Deploy to production
      run: |
        ./scripts/deploy.sh production
      env:
        KUBECONFIG: ${{ secrets.KUBECONFIG_PRODUCTION }}
        IMAGE_TAG: ${{ github.sha }}
```

### Docker Build

#### Multi-stage Dockerfile
```dockerfile
# Backend Dockerfile
FROM golang:1.21-alpine AS backend-builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/server/main.go

FROM alpine:latest AS backend-runtime

RUN apk --no-cache add ca-certificates git
WORKDIR /root/

COPY --from=backend-builder /app/main .
COPY --from=backend-builder /app/config.example.yaml ./config.yaml

EXPOSE 8080
CMD ["./main"]

# Frontend Dockerfile  
FROM node:18-alpine AS frontend-builder

WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production

COPY . .
RUN npm run build

FROM node:18-alpine AS frontend-runtime

WORKDIR /app
COPY --from=frontend-builder /app/next.config.js ./
COPY --from=frontend-builder /app/public ./public
COPY --from=frontend-builder /app/.next ./.next
COPY --from=frontend-builder /app/node_modules ./node_modules
COPY --from=frontend-builder /app/package.json ./package.json

EXPOSE 3000
CMD ["npm", "start"]
```

## Contributing Guidelines

### Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally
3. **Create a feature branch** from `develop`
4. **Make your changes** with tests
5. **Submit a pull request** to `develop` branch

### Development Workflow

#### Branch Naming
- `feature/description` - New features
- `fix/description` - Bug fixes
- `refactor/description` - Code refactoring
- `docs/description` - Documentation updates
- `test/description` - Test improvements

#### Commit Messages
Follow conventional commit format:
```
type(scope): description

[optional body]

[optional footer]
```

Examples:
```
feat(api): add repository creation endpoint
fix(auth): resolve JWT token validation issue
docs(readme): update installation instructions
test(repo): add integration tests for repository API
```

### Code Standards

#### Go Code Style
- Follow official Go style guide
- Use `gofmt` for formatting
- Run `golangci-lint` for linting
- Maintain test coverage above 80%
- Write godoc comments for public functions

#### TypeScript/React Style
- Use TypeScript strict mode
- Follow React best practices
- Use functional components with hooks
- Maintain consistent naming conventions
- Write comprehensive tests

### Pull Request Process

1. **Update documentation** if needed
2. **Add tests** for new functionality
3. **Ensure all tests pass**
4. **Update CHANGELOG.md**
5. **Request review** from maintainers

#### PR Template
```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] Tests added/updated
```

### Release Process

1. **Version bump** in appropriate files
2. **Update CHANGELOG.md** with new version
3. **Create release branch** from `develop`
4. **Merge to main** after testing
5. **Tag release** with version number
6. **Deploy to production**

---

This developer guide provides a comprehensive foundation for contributing to and extending Hub. For additional resources, see the [User Guide](user-guide.md) and [Administrator Guide](admin-guide.md).

For community support and discussions, visit our [GitHub repository](https://github.com/a5c-ai/hub) and join our community channels.
