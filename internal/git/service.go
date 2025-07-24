package git

import (
	"context"
	"time"
)

// GitService provides Git operations for repositories
type GitService interface {
	// Repository operations
	InitRepository(ctx context.Context, repoPath string, bare bool) error
	CloneRepository(ctx context.Context, sourceURL, destPath string, options CloneOptions) error
	DeleteRepository(ctx context.Context, repoPath string) error
	
	// Commit operations
	GetCommits(ctx context.Context, repoPath string, opts CommitOptions) ([]*Commit, error)
	GetCommit(ctx context.Context, repoPath, sha string) (*Commit, error)
	GetCommitDiff(ctx context.Context, repoPath, fromSHA, toSHA string) (*Diff, error)
	
	// Branch operations
	GetBranches(ctx context.Context, repoPath string) ([]*Branch, error)
	GetBranch(ctx context.Context, repoPath, branchName string) (*Branch, error)
	CreateBranch(ctx context.Context, repoPath, branchName, fromRef string) error
	DeleteBranch(ctx context.Context, repoPath, branchName string) error
	
	// Tag operations
	GetTags(ctx context.Context, repoPath string) ([]*Tag, error)
	GetTag(ctx context.Context, repoPath, tagName string) (*Tag, error)
	CreateTag(ctx context.Context, repoPath, tagName, ref, message string) error
	DeleteTag(ctx context.Context, repoPath, tagName string) error
	
	// File operations
	GetTree(ctx context.Context, repoPath, ref, path string) (*Tree, error)
	GetBlob(ctx context.Context, repoPath, sha string) (*Blob, error)
	GetFile(ctx context.Context, repoPath, ref, path string) (*File, error)
	CreateFile(ctx context.Context, repoPath string, req CreateFileRequest) (*Commit, error)
	UpdateFile(ctx context.Context, repoPath string, req UpdateFileRequest) (*Commit, error)
	DeleteFile(ctx context.Context, repoPath string, req DeleteFileRequest) (*Commit, error)
	
	// Repository info
	GetRepositoryInfo(ctx context.Context, repoPath string) (*RepositoryInfo, error)
	GetRepositoryStats(ctx context.Context, repoPath string) (*RepositoryStats, error)
}

// CloneOptions represents options for cloning a repository
type CloneOptions struct {
	Depth    int
	Branch   string
	Mirror   bool
	Bare     bool
	Username string
	Password string
	SSHKey   string
}

// CommitOptions represents options for retrieving commits
type CommitOptions struct {
	Branch    string
	Since     *time.Time
	Until     *time.Time
	Author    string
	Message   string
	Path      string
	Page      int
	PerPage   int
}

// Commit represents a Git commit
type Commit struct {
	SHA         string         `json:"sha"`
	Message     string         `json:"message"`
	Author      CommitAuthor   `json:"author"`
	Committer   CommitAuthor   `json:"committer"`
	Parents     []string       `json:"parents"`
	Tree        string         `json:"tree"`
	Stats       *CommitStats   `json:"stats,omitempty"`
	Files       []*CommitFile  `json:"files,omitempty"`
}

// CommitAuthor represents the author or committer of a commit
type CommitAuthor struct {
	Name  string    `json:"name"`
	Email string    `json:"email"`
	Date  time.Time `json:"date"`
}

// CommitStats represents statistics about a commit
type CommitStats struct {
	Additions int `json:"additions"`
	Deletions int `json:"deletions"`
	Total     int `json:"total"`
}

// CommitFile represents a file changed in a commit
type CommitFile struct {
	Path       string `json:"path"`
	Additions  int    `json:"additions"`
	Deletions  int    `json:"deletions"`
	Changes    int    `json:"changes"`
	Status     string `json:"status"` // added, modified, deleted, renamed
	PrevPath   string `json:"prev_path,omitempty"`
}

// Branch represents a Git branch
type Branch struct {
	Name      string    `json:"name"`
	SHA       string    `json:"sha"`
	Protected bool      `json:"protected"`
	IsDefault bool      `json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Tag represents a Git tag
type Tag struct {
	Name       string       `json:"name"`
	SHA        string       `json:"sha"`
	Message    string       `json:"message,omitempty"`
	Tagger     *CommitAuthor `json:"tagger,omitempty"`
	CreatedAt  time.Time    `json:"created_at"`
}

// Tree represents a Git tree (directory)
type Tree struct {
	SHA     string      `json:"sha"`
	Path    string      `json:"path"`
	Entries []*TreeEntry `json:"entries"`
}

// TreeEntry represents an entry in a Git tree
type TreeEntry struct {
	Name string `json:"name"`
	Path string `json:"path"`
	SHA  string `json:"sha"`
	Size int64  `json:"size"`
	Type string `json:"type"` // blob, tree, commit (submodule)
	Mode string `json:"mode"`
}

// Blob represents a Git blob (file content)
type Blob struct {
	SHA      string `json:"sha"`
	Size     int64  `json:"size"`
	Content  []byte `json:"content"`
	Encoding string `json:"encoding"` // base64 for binary files
}

// File represents a file in the repository
type File struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	SHA      string `json:"sha"`
	Size     int64  `json:"size"`
	Type     string `json:"type"`
	Content  string `json:"content,omitempty"`
	Encoding string `json:"encoding,omitempty"`
}

// Diff represents differences between commits
type Diff struct {
	FromSHA string      `json:"from_sha"`
	ToSHA   string      `json:"to_sha"`
	Files   []*DiffFile `json:"files"`
	Stats   DiffStats   `json:"stats"`
}

// DiffFile represents a file in a diff
type DiffFile struct {
	Path         string `json:"path"`
	PrevPath     string `json:"prev_path,omitempty"`
	Status       string `json:"status"`
	Additions    int    `json:"additions"`
	Deletions    int    `json:"deletions"`
	Changes      int    `json:"changes"`
	Patch        string `json:"patch,omitempty"`
}

// DiffStats represents statistics about a diff
type DiffStats struct {
	Files     int `json:"files"`
	Additions int `json:"additions"`
	Deletions int `json:"deletions"`
	Total     int `json:"total"`
}

// CreateFileRequest represents a request to create a file
type CreateFileRequest struct {
	Path      string `json:"path"`
	Content   string `json:"content"`
	Encoding  string `json:"encoding,omitempty"` // base64 for binary files
	Message   string `json:"message"`
	Branch    string `json:"branch"`
	Author    CommitAuthor `json:"author"`
	Committer CommitAuthor `json:"committer,omitempty"`
}

// UpdateFileRequest represents a request to update a file
type UpdateFileRequest struct {
	Path      string `json:"path"`
	Content   string `json:"content"`
	Encoding  string `json:"encoding,omitempty"`
	Message   string `json:"message"`
	Branch    string `json:"branch"`
	SHA       string `json:"sha"` // Current file SHA for conflict detection
	Author    CommitAuthor `json:"author"`
	Committer CommitAuthor `json:"committer,omitempty"`
}

// DeleteFileRequest represents a request to delete a file
type DeleteFileRequest struct {
	Path      string `json:"path"`
	Message   string `json:"message"`
	Branch    string `json:"branch"`
	SHA       string `json:"sha"` // Current file SHA for conflict detection
	Author    CommitAuthor `json:"author"`
	Committer CommitAuthor `json:"committer,omitempty"`
}

// RepositoryInfo represents basic information about a repository
type RepositoryInfo struct {
	Path          string    `json:"path"`
	DefaultBranch string    `json:"default_branch"`
	IsBare        bool      `json:"is_bare"`
	IsEmpty       bool      `json:"is_empty"`
	LastCommit    *Commit   `json:"last_commit,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// RepositoryStats represents statistics about a repository
type RepositoryStats struct {
	Size           int64                    `json:"size"`           // in bytes
	CommitCount    int                      `json:"commit_count"`
	BranchCount    int                      `json:"branch_count"`
	TagCount       int                      `json:"tag_count"`
	Contributors   int                      `json:"contributors"`
	Languages      map[string]LanguageStats `json:"languages"`
	LastActivity   time.Time                `json:"last_activity"`
}

// LanguageStats represents statistics about a programming language in the repository
type LanguageStats struct {
	Bytes      int64   `json:"bytes"`
	Percentage float64 `json:"percentage"`
}