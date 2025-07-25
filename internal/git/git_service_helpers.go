package git

import (
	"context"
	"encoding/base64"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// Helper methods for bare repository operations

func (s *gitService) createFileInBareRepo(ctx context.Context, repo *git.Repository, req CreateFileRequest) (*Commit, error) {
	return s.modifyFileInBareRepo(ctx, repo, req.Path, req.Content, req.Encoding, req.Message, req.Branch, req.Author, req.Committer, false)
}

func (s *gitService) updateFileInBareRepo(ctx context.Context, repo *git.Repository, req UpdateFileRequest) (*Commit, error) {
	return s.modifyFileInBareRepo(ctx, repo, req.Path, req.Content, req.Encoding, req.Message, req.Branch, req.Author, req.Committer, true)
}

func (s *gitService) deleteFileInBareRepo(ctx context.Context, repo *git.Repository, req DeleteFileRequest) (*Commit, error) {
	return s.modifyFileInBareRepo(ctx, repo, req.Path, "", "", req.Message, req.Branch, req.Author, req.Committer, true)
}

// modifyFileInBareRepo handles create, update, and delete operations for files in bare repositories
func (s *gitService) modifyFileInBareRepo(ctx context.Context, repo *git.Repository, path, content, encoding, message, branchName string, author, committer CommitAuthor, isUpdate bool) (*Commit, error) {
	// Get the branch reference
	branchRef := fmt.Sprintf("refs/heads/%s", branchName)
	ref, err := repo.Reference(plumbing.ReferenceName(branchRef), true)
	if err != nil {
		return nil, fmt.Errorf("failed to get branch reference %s: %w", branchName, err)
	}

	// Get the current commit
	currentCommit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get current commit: %w", err)
	}

	// Get the current tree
	currentTree, err := currentCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get current tree: %w", err)
	}

	// Prepare content
	var fileContent []byte
	if content != "" {
		if encoding == "base64" {
			fileContent, err = base64.StdEncoding.DecodeString(content)
			if err != nil {
				return nil, fmt.Errorf("failed to decode base64 content: %w", err)
			}
		} else {
			fileContent = []byte(content)
		}
	}

	// Create or update the tree with the new/modified file
	newTreeHash, err := s.updateTreeWithFile(repo, currentTree, path, fileContent, isUpdate && content == "")
	if err != nil {
		return nil, fmt.Errorf("failed to update tree: %w", err)
	}

	// Set default committer if not provided
	if committer.Name == "" {
		committer = author
	}
	if committer.Date.IsZero() {
		committer.Date = time.Now()
	}
	if author.Date.IsZero() {
		author.Date = time.Now()
	}

	// Create a new commit
	newCommit := &object.Commit{
		Author: object.Signature{
			Name:  author.Name,
			Email: author.Email,
			When:  author.Date,
		},
		Committer: object.Signature{
			Name:  committer.Name,
			Email: committer.Email,
			When:  committer.Date,
		},
		Message:      message,
		TreeHash:     newTreeHash,
		ParentHashes: []plumbing.Hash{currentCommit.Hash},
	}

	// Store the commit object
	commitObj := repo.Storer.NewEncodedObject()
	if err := newCommit.Encode(commitObj); err != nil {
		return nil, fmt.Errorf("failed to encode commit: %w", err)
	}

	commitHash, err := repo.Storer.SetEncodedObject(commitObj)
	if err != nil {
		return nil, fmt.Errorf("failed to store commit: %w", err)
	}

	// Update the branch reference
	newRef := plumbing.NewHashReference(plumbing.ReferenceName(branchRef), commitHash)
	if err := repo.Storer.SetReference(newRef); err != nil {
		return nil, fmt.Errorf("failed to update branch reference: %w", err)
	}

	// Return the commit information
	return &Commit{
		SHA:     commitHash.String(),
		Message: message,
		Author: CommitAuthor{
			Name:  author.Name,
			Email: author.Email,
			Date:  author.Date,
		},
		Committer: CommitAuthor{
			Name:  committer.Name,
			Email: committer.Email,
			Date:  committer.Date,
		},
		Parents: []string{currentCommit.Hash.String()},
	}, nil
}

// updateTreeWithFile creates a new tree with the specified file added, updated, or removed
func (s *gitService) updateTreeWithFile(repo *git.Repository, baseTree *object.Tree, filePath string, content []byte, delete bool) (plumbing.Hash, error) {
	// Split the path into directory components
	pathParts := strings.Split(strings.Trim(filePath, "/"), "/")

	return s.updateTreeRecursive(repo, baseTree, pathParts, content, delete)
}

// updateTreeRecursive recursively updates tree objects to modify a file at the given path
func (s *gitService) updateTreeRecursive(repo *git.Repository, tree *object.Tree, pathParts []string, content []byte, delete bool) (plumbing.Hash, error) {
	if len(pathParts) == 0 {
		return tree.Hash, nil
	}

	fileName := pathParts[0]
	isLastPart := len(pathParts) == 1

	// Start with a map to ensure uniqueness and easier manipulation
	entryMap := make(map[string]object.TreeEntry)

	// Copy existing entries, excluding the one we're modifying
	for _, entry := range tree.Entries {
		if entry.Name != fileName {
			entryMap[entry.Name] = entry
		}
	}

	if isLastPart {
		// This is the file we want to modify
		if !delete && content != nil {
			// Create a blob for the file content
			blob := repo.Storer.NewEncodedObject()
			blob.SetType(plumbing.BlobObject)
			writer, err := blob.Writer()
			if err != nil {
				return plumbing.Hash{}, fmt.Errorf("failed to create blob writer: %w", err)
			}

			if _, err := writer.Write(content); err != nil {
				writer.Close()
				return plumbing.Hash{}, fmt.Errorf("failed to write blob content: %w", err)
			}
			writer.Close()

			blobHash, err := repo.Storer.SetEncodedObject(blob)
			if err != nil {
				return plumbing.Hash{}, fmt.Errorf("failed to store blob: %w", err)
			}

			// Add the new file entry
			entryMap[fileName] = object.TreeEntry{
				Name: fileName,
				Mode: filemode.Regular,
				Hash: blobHash,
			}
		}
		// If delete == true, we simply don't add the entry (effectively deleting it)
	} else {
		// This is a directory, we need to recurse
		var subTree *object.Tree

		// Find the existing subtree if it exists
		for _, entry := range tree.Entries {
			if entry.Name == fileName && entry.Mode == filemode.Dir {
				var err error
				subTree, err = repo.TreeObject(entry.Hash)
				if err != nil {
					return plumbing.Hash{}, fmt.Errorf("failed to get subtree: %w", err)
				}
				break
			}
		}

		// If subtree doesn't exist, create an empty one
		if subTree == nil {
			subTree = &object.Tree{}
		}

		// Recursively update the subtree
		newSubTreeHash, err := s.updateTreeRecursive(repo, subTree, pathParts[1:], content, delete)
		if err != nil {
			return plumbing.Hash{}, fmt.Errorf("failed to update subtree: %w", err)
		}

		// Add the updated subtree entry
		entryMap[fileName] = object.TreeEntry{
			Name: fileName,
			Mode: filemode.Dir,
			Hash: newSubTreeHash,
		}
	}

	// Convert map to slice and sort properly
	var entries []object.TreeEntry
	for _, entry := range entryMap {
		entries = append(entries, entry)
	}

	// Use a more robust sorting approach
	sort.Slice(entries, func(i, j int) bool {
		return s.compareTreeEntries(entries[i], entries[j])
	})

	// Create the new tree object using go-git's constructor
	newTree := &object.Tree{Entries: entries}

	// Encode the tree
	encoded := &plumbing.MemoryObject{}
	if err := newTree.Encode(encoded); err != nil {
		return plumbing.Hash{}, fmt.Errorf("failed to encode tree: %w", err)
	}

	// Store the tree object
	return repo.Storer.SetEncodedObject(encoded)
}

// compareTreeEntries implements Git's tree entry comparison logic
// Git sorts tree entries by name, but treats directories as if they have a trailing "/"
func (s *gitService) compareTreeEntries(a, b object.TreeEntry) bool {
	nameA := a.Name
	nameB := b.Name

	// For directories, append "/" for comparison (Git's tree sorting rule)
	if a.Mode == filemode.Dir {
		nameA += "/"
	}
	if b.Mode == filemode.Dir {
		nameB += "/"
	}

	// Git uses byte-wise lexicographic comparison
	return nameA < nameB
}

// gitTreeEntryLess implements Git's tree entry comparison logic
// Git sorts tree entries by name, but treats directories as if they have a trailing "/"
func gitTreeEntryLess(a, b object.TreeEntry) bool {
	nameA := a.Name
	nameB := b.Name

	// For directories, append "/" for comparison
	if a.Mode == filemode.Dir {
		nameA += "/"
	}
	if b.Mode == filemode.Dir {
		nameB += "/"
	}

	// Git uses byte-wise comparison
	return nameA < nameB
}
