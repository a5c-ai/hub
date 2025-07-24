package git

import (
	"context"
	"encoding/base64"
	"errors"
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
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/sirupsen/logrus"
)

// Common Git errors
var (
	ErrRepositoryNotFound   = errors.New("repository not found")
	ErrRepositoryCorrupted  = errors.New("repository is corrupted")
	ErrReferenceNotFound    = errors.New("reference not found")
	ErrCommitNotFound       = errors.New("commit not found")
	ErrBranchNotFound       = errors.New("branch not found")
	ErrTagNotFound          = errors.New("tag not found")
	ErrFileNotFound         = errors.New("file not found")
	ErrPathNotFound         = errors.New("path not found")
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
		s.logger.WithError(err).WithField("path", repoPath).Error("Failed to open repository")
		if os.IsNotExist(err) {
			return nil, ErrRepositoryNotFound
		}
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
	repo, err := s.openRepository(repoPath)
	if err != nil {
		return nil, err
	}

	fromHash := plumbing.NewHash(fromSHA)
	toHash := plumbing.NewHash(toSHA)

	fromCommit, err := repo.CommitObject(fromHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get from commit %s: %w", fromSHA, err)
	}

	toCommit, err := repo.CommitObject(toHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get to commit %s: %w", toSHA, err)
	}

	fromTree, err := fromCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get from tree: %w", err)
	}

	toTree, err := toCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get to tree: %w", err)
	}

	changes, err := fromTree.Diff(toTree)
	if err != nil {
		return nil, fmt.Errorf("failed to compute diff: %w", err)
	}

	var files []*DiffFile
	stats := DiffStats{}

	for _, change := range changes {
		diffFile := &DiffFile{
			Path:     change.To.Name,
			PrevPath: change.From.Name,
		}

		switch {
		case change.From.Name == "" && change.To.Name != "":
			diffFile.Status = "added"
		case change.From.Name != "" && change.To.Name == "":
			diffFile.Status = "deleted"
			diffFile.Path = change.From.Name
		case change.From.Name != change.To.Name:
			diffFile.Status = "renamed"
		default:
			diffFile.Status = "modified"
		}

		// Get patch for the file (simplified)
		patch, err := change.Patch()
		if err == nil {
			diffFile.Patch = patch.String()
			// Parse patch for stats (simplified)
			lines := strings.Split(patch.String(), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
					diffFile.Additions++
				} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
					diffFile.Deletions++
				}
			}
			diffFile.Changes = diffFile.Additions + diffFile.Deletions
		}

		files = append(files, diffFile)
		stats.Files++
		stats.Additions += diffFile.Additions
		stats.Deletions += diffFile.Deletions
	}

	stats.Total = stats.Additions + stats.Deletions

	return &Diff{
		FromSHA: fromSHA,
		ToSHA:   toSHA,
		Files:   files,
		Stats:   stats,
	}, nil
}

func (s *gitService) GetTree(ctx context.Context, repoPath, ref, path string) (*Tree, error) {
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

	tree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get tree: %w", err)
	}

	// Navigate to the specified path if provided
	if path != "" && path != "/" {
		tree, err = tree.Tree(path)
		if err != nil {
			return nil, fmt.Errorf("failed to get tree at path %s: %w", path, err)
		}
	}

	var entries []*TreeEntry
	for _, entry := range tree.Entries {
		treeEntry := &TreeEntry{
			Name: entry.Name,
			Path: filepath.Join(path, entry.Name),
			SHA:  entry.Hash.String(),
			Mode: entry.Mode.String(),
		}

		switch entry.Mode {
		case filemode.Regular, filemode.Executable:
			treeEntry.Type = "blob"
			// Get file size
			if file, err := tree.File(entry.Name); err == nil {
				treeEntry.Size = file.Size
			}
		case filemode.Dir:
			treeEntry.Type = "tree"
		case filemode.Symlink:
			treeEntry.Type = "blob"
		case filemode.Submodule:
			treeEntry.Type = "commit"
		default:
			treeEntry.Type = "blob"
		}

		entries = append(entries, treeEntry)
	}

	// Sort entries: directories first, then files, both alphabetically
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Type == "tree" && entries[j].Type != "tree" {
			return true
		}
		if entries[i].Type != "tree" && entries[j].Type == "tree" {
			return false
		}
		return entries[i].Name < entries[j].Name
	})

	return &Tree{
		SHA:     tree.Hash.String(),
		Path:    path,
		Entries: entries,
	}, nil
}

func (s *gitService) GetBlob(ctx context.Context, repoPath, sha string) (*Blob, error) {
	repo, err := s.openRepository(repoPath)
	if err != nil {
		return nil, err
	}

	hash := plumbing.NewHash(sha)
	blob, err := repo.BlobObject(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get blob %s: %w", sha, err)
	}

	reader, err := blob.Reader()
	if err != nil {
		return nil, fmt.Errorf("failed to get blob reader: %w", err)
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read blob content: %w", err)
	}

	// Determine encoding
	encoding := ""
	if !utf8.ValidString(string(content)) {
		encoding = "base64"
	}

	return &Blob{
		SHA:      sha,
		Size:     blob.Size,
		Content:  content,
		Encoding: encoding,
	}, nil
}

func (s *gitService) CreateFile(ctx context.Context, repoPath string, req CreateFileRequest) (*Commit, error) {
	repo, err := s.openRepository(repoPath)
	if err != nil {
		return nil, err
	}

	// Get or create worktree for bare repositories
	workTree, err := repo.Worktree()
	if err != nil {
		// For bare repositories, we need to work directly with the object database
		return s.createFileInBareRepo(ctx, repo, req)
	}

	// Write file to filesystem
	filePath := filepath.Join(workTree.Filesystem.Root(), req.Path)
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Decode content if base64
	content := []byte(req.Content)
	if req.Encoding == "base64" {
		content, err = base64.StdEncoding.DecodeString(req.Content)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64 content: %w", err)
		}
	}

	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// Add file to index
	if _, err := workTree.Add(req.Path); err != nil {
		return nil, fmt.Errorf("failed to add file to index: %w", err)
	}

	// Create commit
	commitHash, err := workTree.Commit(req.Message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  req.Author.Name,
			Email: req.Author.Email,
			When:  req.Author.Date,
		},
		Committer: &object.Signature{
			Name:  req.Committer.Name,
			Email: req.Committer.Email,
			When:  req.Committer.Date,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create commit: %w", err)
	}

	// Get the created commit
	commitObj, err := repo.CommitObject(commitHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit object: %w", err)
	}

	return s.convertCommit(commitObj), nil
}

func (s *gitService) UpdateFile(ctx context.Context, repoPath string, req UpdateFileRequest) (*Commit, error) {
	repo, err := s.openRepository(repoPath)
	if err != nil {
		return nil, err
	}

	// Get worktree
	workTree, err := repo.Worktree()
	if err != nil {
		// For bare repositories, we need different handling
		return s.updateFileInBareRepo(ctx, repo, req)
	}

	// Check if file exists and verify SHA
	filePath := filepath.Join(workTree.Filesystem.Root(), req.Path)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file %s does not exist", req.Path)
	}

	// TODO: Verify SHA matches current file SHA for conflict detection
	// This would require computing the current file's Git SHA

	// Decode content if base64
	content := []byte(req.Content)
	if req.Encoding == "base64" {
		content, err = base64.StdEncoding.DecodeString(req.Content)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64 content: %w", err)
		}
	}

	// Write updated content
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// Add file to index
	if _, err := workTree.Add(req.Path); err != nil {
		return nil, fmt.Errorf("failed to add file to index: %w", err)
	}

	// Create commit
	commitHash, err := workTree.Commit(req.Message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  req.Author.Name,
			Email: req.Author.Email,
			When:  req.Author.Date,
		},
		Committer: &object.Signature{
			Name:  req.Committer.Name,
			Email: req.Committer.Email,
			When:  req.Committer.Date,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create commit: %w", err)
	}

	// Get the created commit
	commitObj, err := repo.CommitObject(commitHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit object: %w", err)
	}

	return s.convertCommit(commitObj), nil
}

func (s *gitService) DeleteFile(ctx context.Context, repoPath string, req DeleteFileRequest) (*Commit, error) {
	repo, err := s.openRepository(repoPath)
	if err != nil {
		return nil, err
	}

	// Get worktree
	workTree, err := repo.Worktree()
	if err != nil {
		// For bare repositories, we need different handling
		return s.deleteFileInBareRepo(ctx, repo, req)
	}

	// Check if file exists
	filePath := filepath.Join(workTree.Filesystem.Root(), req.Path)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file %s does not exist", req.Path)
	}

	// TODO: Verify SHA matches current file SHA for conflict detection

	// Remove file from filesystem
	if err := os.Remove(filePath); err != nil {
		return nil, fmt.Errorf("failed to remove file: %w", err)
	}

	// Remove file from index
	if _, err := workTree.Remove(req.Path); err != nil {
		return nil, fmt.Errorf("failed to remove file from index: %w", err)
	}

	// Create commit
	commitHash, err := workTree.Commit(req.Message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  req.Author.Name,
			Email: req.Author.Email,
			When:  req.Author.Date,
		},
		Committer: &object.Signature{
			Name:  req.Committer.Name,
			Email: req.Committer.Email,
			When:  req.Committer.Date,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create commit: %w", err)
	}

	// Get the created commit
	commitObj, err := repo.CommitObject(commitHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit object: %w", err)
	}

	return s.convertCommit(commitObj), nil
}

func (s *gitService) GetRepositoryInfo(ctx context.Context, repoPath string) (*RepositoryInfo, error) {
	repo, err := s.openRepository(repoPath)
	if err != nil {
		return nil, err
	}

	// Check if repository is bare
	isBare := true
	if _, err := repo.Worktree(); err == nil {
		isBare = false
	}

	// Get default branch
	defaultBranch := "main"
	isEmpty := false
	var lastCommit *Commit

	if head, err := repo.Head(); err == nil {
		if head.Name().IsBranch() {
			defaultBranch = head.Name().Short()
		}
		
		// Get last commit
		if commitObj, err := repo.CommitObject(head.Hash()); err == nil {
			lastCommit = s.convertCommit(commitObj)
		}
	} else {
		// Repository is empty
		isEmpty = true
	}

	// Get repository creation/modification times
	createdAt := time.Now()
	updatedAt := time.Now()

	// Try to get filesystem stats
	if stat, err := os.Stat(repoPath); err == nil {
		createdAt = stat.ModTime()
		updatedAt = stat.ModTime()
	}

	return &RepositoryInfo{
		Path:          repoPath,
		DefaultBranch: defaultBranch,
		IsBare:        isBare,
		IsEmpty:       isEmpty,
		LastCommit:    lastCommit,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}, nil
}

func (s *gitService) GetRepositoryStats(ctx context.Context, repoPath string) (*RepositoryStats, error) {
	repo, err := s.openRepository(repoPath)
	if err != nil {
		return nil, err
	}

	// Initialize stats
	stats := &RepositoryStats{
		Languages: make(map[string]LanguageStats),
	}

	// Count commits
	if head, err := repo.Head(); err == nil {
		commitIter, err := repo.Log(&git.LogOptions{From: head.Hash()})
		if err == nil {
			commitIter.ForEach(func(c *object.Commit) error {
				stats.CommitCount++
				if c.Author.When.After(stats.LastActivity) {
					stats.LastActivity = c.Author.When
				}
				return nil
			})
			commitIter.Close()
		}
	}

	// Count branches
	if refs, err := repo.References(); err == nil {
		refs.ForEach(func(ref *plumbing.Reference) error {
			if ref.Name().IsBranch() {
				stats.BranchCount++
			}
			return nil
		})
		refs.Close()
	}

	// Count tags
	if tagRefs, err := repo.Tags(); err == nil {
		tagRefs.ForEach(func(ref *plumbing.Reference) error {
			stats.TagCount++
			return nil
		})
		tagRefs.Close()
	}

	// Calculate repository size (simplified)
	if stat, err := os.Stat(repoPath); err == nil {
		if stat.IsDir() {
			// Walk directory to calculate size
			filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
				if err == nil && !info.IsDir() {
					stats.Size += info.Size()
				}
				return nil
			})
		}
	}

	// TODO: Implement language detection and contributor counting
	// This would require analyzing file extensions and commit authors
	stats.Contributors = 1 // Placeholder

	return stats, nil
}