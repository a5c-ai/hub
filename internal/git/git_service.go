package git

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/sirupsen/logrus"
)

// gitService implements the GitService interface using go-git
type gitService struct {
	logger *logrus.Logger
}

// NewGitService creates a new Git service instance
func NewGitService(logger *logrus.Logger) GitService {
	return &gitService{
		logger: logger,
	}
}

// InitRepository initializes a new Git repository
func (s *gitService) InitRepository(ctx context.Context, repoPath string, bare bool) error {
	s.logger.WithFields(logrus.Fields{
		"path": repoPath,
		"bare": bare,
	}).Info("Initializing Git repository")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		return fmt.Errorf("failed to create repository directory: %w", err)
	}

	// Initialize repository
	_, err := git.PlainInit(repoPath, bare)
	if err != nil {
		return fmt.Errorf("failed to initialize Git repository: %w", err)
	}

	return nil
}

// CloneRepository clones a repository from a remote URL
func (s *gitService) CloneRepository(ctx context.Context, sourceURL, destPath string, options CloneOptions) error {
	s.logger.WithFields(logrus.Fields{
		"source": sourceURL,
		"dest":   destPath,
		"bare":   options.Bare,
		"mirror": options.Mirror,
	}).Info("Cloning Git repository")

	cloneOptions := &git.CloneOptions{
		URL:   sourceURL,
		Depth: options.Depth,
	}

	if options.Branch != "" {
		cloneOptions.ReferenceName = plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", options.Branch))
		cloneOptions.SingleBranch = true
	}

	if options.Mirror {
		cloneOptions.Mirror = true
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(destPath, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	_, err := git.PlainCloneContext(ctx, destPath, options.Bare, cloneOptions)
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	return nil
}

// DeleteRepository removes a repository from disk
func (s *gitService) DeleteRepository(ctx context.Context, repoPath string) error {
	s.logger.WithField("path", repoPath).Info("Deleting Git repository")

	if err := os.RemoveAll(repoPath); err != nil {
		return fmt.Errorf("failed to delete repository: %w", err)
	}

	return nil
}

// GetCommits retrieves commits from a repository
func (s *gitService) GetCommits(ctx context.Context, repoPath string, opts CommitOptions) ([]*Commit, error) {
	repo, err := s.openRepository(repoPath)
	if err != nil {
		return nil, err
	}

	ref, err := s.resolveReference(repo, opts.Branch)
	if err != nil {
		return nil, err
	}

	commitIter, err := repo.Log(&git.LogOptions{
		From: ref,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get commit log: %w", err)
	}
	defer commitIter.Close()

	var commits []*Commit
	count := 0
	perPage := opts.PerPage
	if perPage <= 0 {
		perPage = 30
	}
	skip := opts.Page * perPage

	err = commitIter.ForEach(func(c *object.Commit) error {
		if count < skip {
			count++
			return nil
		}

		if len(commits) >= perPage {
			return io.EOF
		}

		// Apply filters
		if opts.Since != nil && c.Author.When.Before(*opts.Since) {
			return nil
		}
		if opts.Until != nil && c.Author.When.After(*opts.Until) {
			return nil
		}
		if opts.Author != "" && !strings.Contains(c.Author.Name, opts.Author) {
			return nil
		}
		if opts.Message != "" && !strings.Contains(c.Message, opts.Message) {
			return nil
		}

		commit := s.convertCommit(c)
		commits = append(commits, commit)
		count++

		return nil
	})

	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to iterate commits: %w", err)
	}

	return commits, nil
}

// GetCommit retrieves a single commit by SHA
func (s *gitService) GetCommit(ctx context.Context, repoPath, sha string) (*Commit, error) {
	repo, err := s.openRepository(repoPath)
	if err != nil {
		return nil, err
	}

	hash := plumbing.NewHash(sha)
	commit, err := repo.CommitObject(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit %s: %w", sha, err)
	}

	return s.convertCommit(commit), nil
}

// GetBranches retrieves all branches from a repository
func (s *gitService) GetBranches(ctx context.Context, repoPath string) ([]*Branch, error) {
	repo, err := s.openRepository(repoPath)
	if err != nil {
		return nil, err
	}

	refs, err := repo.References()
	if err != nil {
		return nil, fmt.Errorf("failed to get references: %w", err)
	}
	defer refs.Close()

	var branches []*Branch
	defaultBranch := "main" // Default assumption

	// Try to get the actual default branch
	if head, err := repo.Head(); err == nil {
		if head.Name().IsBranch() {
			defaultBranch = head.Name().Short()
		}
	}

	err = refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().IsBranch() {
			branchName := ref.Name().Short()
			branch := &Branch{
				Name:      branchName,
				SHA:       ref.Hash().String(),
				IsDefault: branchName == defaultBranch,
				CreatedAt: time.Now(), // go-git doesn't provide branch creation time
				UpdatedAt: time.Now(),
			}
			branches = append(branches, branch)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to iterate branches: %w", err)
	}

	// Sort branches by name
	sort.Slice(branches, func(i, j int) bool {
		return branches[i].Name < branches[j].Name
	})

	return branches, nil
}

// GetBranch retrieves a single branch by name
func (s *gitService) GetBranch(ctx context.Context, repoPath, branchName string) (*Branch, error) {
	repo, err := s.openRepository(repoPath)
	if err != nil {
		return nil, err
	}

	ref, err := repo.Reference(plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branchName)), true)
	if err != nil {
		return nil, fmt.Errorf("failed to get branch %s: %w", branchName, err)
	}

	// Check if this is the default branch
	isDefault := false
	if head, err := repo.Head(); err == nil && head.Name().Short() == branchName {
		isDefault = true
	}

	return &Branch{
		Name:      branchName,
		SHA:       ref.Hash().String(),
		IsDefault: isDefault,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

// CreateBranch creates a new branch
func (s *gitService) CreateBranch(ctx context.Context, repoPath, branchName, fromRef string) error {
	repo, err := s.openRepository(repoPath)
	if err != nil {
		return err
	}

	// Resolve the reference to create branch from
	var hash plumbing.Hash
	if fromRef == "" {
		// Use HEAD if no reference specified
		if head, err := repo.Head(); err == nil {
			hash = head.Hash()
		} else {
			return fmt.Errorf("failed to get HEAD reference: %w", err)
		}
	} else {
		ref, err := s.resolveReference(repo, fromRef)
		if err != nil {
			return fmt.Errorf("failed to resolve reference %s: %w", fromRef, err)
		}
		hash = ref
	}

	// Create the new branch reference
	refName := plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branchName))
	ref := plumbing.NewHashReference(refName, hash)

	err = repo.Storer.SetReference(ref)
	if err != nil {
		return fmt.Errorf("failed to create branch %s: %w", branchName, err)
	}

	return nil
}

// DeleteBranch deletes a branch
func (s *gitService) DeleteBranch(ctx context.Context, repoPath, branchName string) error {
	repo, err := s.openRepository(repoPath)
	if err != nil {
		return err
	}

	refName := plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branchName))
	err = repo.Storer.RemoveReference(refName)
	if err != nil {
		return fmt.Errorf("failed to delete branch %s: %w", branchName, err)
	}

	return nil
}

// GetTags retrieves all tags from a repository
func (s *gitService) GetTags(ctx context.Context, repoPath string) ([]*Tag, error) {
	repo, err := s.openRepository(repoPath)
	if err != nil {
		return nil, err
	}

	tagRefs, err := repo.Tags()
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}
	defer tagRefs.Close()

	var tags []*Tag
	err = tagRefs.ForEach(func(ref *plumbing.Reference) error {
		tag := &Tag{
			Name:      ref.Name().Short(),
			SHA:       ref.Hash().String(),
			CreatedAt: time.Now(),
		}

		// Try to get tag object for additional information
		if tagObj, err := repo.TagObject(ref.Hash()); err == nil {
			tag.Message = tagObj.Message
			tag.Tagger = &CommitAuthor{
				Name:  tagObj.Tagger.Name,
				Email: tagObj.Tagger.Email,
				Date:  tagObj.Tagger.When,
			}
		}

		tags = append(tags, tag)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to iterate tags: %w", err)
	}

	// Sort tags by name
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Name < tags[j].Name
	})

	return tags, nil
}

// GetTag retrieves a single tag by name
func (s *gitService) GetTag(ctx context.Context, repoPath, tagName string) (*Tag, error) {
	repo, err := s.openRepository(repoPath)
	if err != nil {
		return nil, err
	}

	ref, err := repo.Tag(tagName)
	if err != nil {
		return nil, fmt.Errorf("failed to get tag %s: %w", tagName, err)
	}

	tag := &Tag{
		Name:      tagName,
		SHA:       ref.Hash().String(),
		CreatedAt: time.Now(),
	}

	// Try to get tag object for additional information
	if tagObj, err := repo.TagObject(ref.Hash()); err == nil {
		tag.Message = tagObj.Message
		tag.Tagger = &CommitAuthor{
			Name:  tagObj.Tagger.Name,
			Email: tagObj.Tagger.Email,
			Date:  tagObj.Tagger.When,
		}
	}

	return tag, nil
}

// CreateTag creates a new tag
func (s *gitService) CreateTag(ctx context.Context, repoPath, tagName, ref, message string) error {
	repo, err := s.openRepository(repoPath)
	if err != nil {
		return err
	}

	// Resolve the reference to tag
	hash, err := s.resolveReference(repo, ref)
	if err != nil {
		return fmt.Errorf("failed to resolve reference %s: %w", ref, err)
	}

	// Create the tag reference
	refName := plumbing.ReferenceName(fmt.Sprintf("refs/tags/%s", tagName))
	tagRef := plumbing.NewHashReference(refName, hash)

	err = repo.Storer.SetReference(tagRef)
	if err != nil {
		return fmt.Errorf("failed to create tag %s: %w", tagName, err)
	}

	return nil
}

// DeleteTag deletes a tag
func (s *gitService) DeleteTag(ctx context.Context, repoPath, tagName string) error {
	repo, err := s.openRepository(repoPath)
	if err != nil {
		return err
	}

	refName := plumbing.ReferenceName(fmt.Sprintf("refs/tags/%s", tagName))
	err = repo.Storer.RemoveReference(refName)
	if err != nil {
		return fmt.Errorf("failed to delete tag %s: %w", tagName, err)
	}

	return nil
}

// GetFile retrieves a file from the repository
func (s *gitService) GetFile(ctx context.Context, repoPath, ref, path string) (*File, error) {
	repo, err := s.openRepository(repoPath)
	if err != nil {
		return nil, err
	}

	hash, err := s.resolveReference(repo, ref)
	if err != nil {
		return nil, err
	}

	commit, err := repo.CommitObject(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit: %w", err)
	}

	file, err := commit.File(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get file %s: %w", path, err)
	}

	content, err := file.Contents()
	if err != nil {
		return nil, fmt.Errorf("failed to get file contents: %w", err)
	}

	// Determine encoding
	encoding := ""
	if !utf8.ValidString(content) {
		encoding = "base64"
		content = base64.StdEncoding.EncodeToString([]byte(content))
	}

	return &File{
		Name:     filepath.Base(path),
		Path:     path,
		SHA:      file.Hash.String(),
		Size:     file.Size,
		Type:     "file",
		Content:  content,
		Encoding: encoding,
	}, nil
}

// Helper methods

func (s *gitService) openRepository(repoPath string) (*git.Repository, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository at %s: %w", repoPath, err)
	}
	return repo, nil
}

func (s *gitService) resolveReference(repo *git.Repository, ref string) (plumbing.Hash, error) {
	if ref == "" {
		// Use HEAD if no reference specified
		head, err := repo.Head()
		if err != nil {
			return plumbing.ZeroHash, fmt.Errorf("failed to get HEAD: %w", err)
		}
		return head.Hash(), nil
	}

	// Try to parse as hash first
	if hash := plumbing.NewHash(ref); !hash.IsZero() {
		return hash, nil
	}

	// Try to resolve as reference
	reference, err := repo.Reference(plumbing.ReferenceName(ref), true)
	if err != nil {
		// Try as branch reference
		branchRef := plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", ref))
		reference, err = repo.Reference(branchRef, true)
		if err != nil {
			// Try as tag reference
			tagRef := plumbing.ReferenceName(fmt.Sprintf("refs/tags/%s", ref))
			reference, err = repo.Reference(tagRef, true)
			if err != nil {
				return plumbing.ZeroHash, fmt.Errorf("failed to resolve reference %s: %w", ref, err)
			}
		}
	}

	return reference.Hash(), nil
}

func (s *gitService) convertCommit(c *object.Commit) *Commit {
	var parents []string
	for _, parent := range c.ParentHashes {
		parents = append(parents, parent.String())
	}

	return &Commit{
		SHA:     c.Hash.String(),
		Message: strings.TrimSpace(c.Message),
		Author: CommitAuthor{
			Name:  c.Author.Name,
			Email: c.Author.Email,
			Date:  c.Author.When,
		},
		Committer: CommitAuthor{
			Name:  c.Committer.Name,
			Email: c.Committer.Email,
			Date:  c.Committer.When,
		},
		Parents: parents,
		Tree:    c.TreeHash.String(),
	}
}

// Placeholder implementations for methods that need more complex logic

func (s *gitService) GetCommitDiff(ctx context.Context, repoPath, fromSHA, toSHA string) (*Diff, error) {
	// TODO: Implement diff functionality
	return nil, fmt.Errorf("GetCommitDiff not yet implemented")
}

func (s *gitService) GetTree(ctx context.Context, repoPath, ref, path string) (*Tree, error) {
	// TODO: Implement tree functionality
	return nil, fmt.Errorf("GetTree not yet implemented")
}

func (s *gitService) GetBlob(ctx context.Context, repoPath, sha string) (*Blob, error) {
	// TODO: Implement blob functionality
	return nil, fmt.Errorf("GetBlob not yet implemented")
}

func (s *gitService) CreateFile(ctx context.Context, repoPath string, req CreateFileRequest) (*Commit, error) {
	// TODO: Implement file creation
	return nil, fmt.Errorf("CreateFile not yet implemented")
}

func (s *gitService) UpdateFile(ctx context.Context, repoPath string, req UpdateFileRequest) (*Commit, error) {
	// TODO: Implement file update
	return nil, fmt.Errorf("UpdateFile not yet implemented")
}

func (s *gitService) DeleteFile(ctx context.Context, repoPath string, req DeleteFileRequest) (*Commit, error) {
	// TODO: Implement file deletion
	return nil, fmt.Errorf("DeleteFile not yet implemented")
}

func (s *gitService) GetRepositoryInfo(ctx context.Context, repoPath string) (*RepositoryInfo, error) {
	// TODO: Implement repository info
	return nil, fmt.Errorf("GetRepositoryInfo not yet implemented")
}

func (s *gitService) GetRepositoryStats(ctx context.Context, repoPath string) (*RepositoryStats, error) {
	// TODO: Implement repository statistics
	return nil, fmt.Errorf("GetRepositoryStats not yet implemented")
}