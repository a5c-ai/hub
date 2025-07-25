package services

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/a5c-ai/hub/internal/git"
	"github.com/a5c-ai/hub/internal/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type PullRequestService struct {
	db             *gorm.DB
	gitService     git.GitService
	repoService    RepositoryService
	logger         *logrus.Logger
	repoBasePath   string
}

func NewPullRequestService(db *gorm.DB, gitService git.GitService, repoService RepositoryService, logger *logrus.Logger, repoBasePath string) *PullRequestService {
	return &PullRequestService{
		db:           db,
		gitService:   gitService,
		repoService:  repoService,
		logger:       logger,
		repoBasePath: repoBasePath,
	}
}

type CreatePullRequestRequest struct {
	Title                string    `json:"title" binding:"required"`
	Body                 string    `json:"body"`
	Head                 string    `json:"head" binding:"required"`
	Base                 string    `json:"base" binding:"required"`
	HeadRepositoryID     *uuid.UUID `json:"head_repository_id"`
	Draft                bool      `json:"draft"`
	MaintainerCanModify  bool      `json:"maintainer_can_modify"`
}

type PullRequestListOptions struct {
	State     string
	Head      string
	Base      string
	Sort      string
	Direction string
	Page      int
	PerPage   int
}

func (s *PullRequestService) CreatePullRequest(repoID uuid.UUID, userID uuid.UUID, req CreatePullRequestRequest) (*models.PullRequest, error) {
	var repo models.Repository
	if err := s.db.First(&repo, "id = ?", repoID).Error; err != nil {
		return nil, err
	}

	// Get the next issue number
	nextNumber, err := s.getNextIssueNumber(repoID)
	if err != nil {
		return nil, err
	}

	// Create the issue first
	issue := models.Issue{
		RepositoryID:  repoID,
		Number:        nextNumber,
		Title:         req.Title,
		Body:          req.Body,
		UserID:        &userID,
		State:         models.IssueStateOpen,
		CommentsCount: 0,
	}

	if err := s.db.Create(&issue).Error; err != nil {
		return nil, err
	}

	// Set head repository ID
	headRepoID := repoID
	if req.HeadRepositoryID != nil {
		headRepoID = *req.HeadRepositoryID
	}

	// Get branch comparison data
	ownerName, err := s.getRepositoryOwnerName(&repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository owner: %w", err)
	}
	
	repoPath := filepath.Join(s.repoBasePath, ownerName, repo.Name)
	comparison, err := s.gitService.CompareRefs(repoPath, req.Base, req.Head)
	if err != nil {
		s.logger.WithError(err).Error("Failed to compare branches")
		return nil, fmt.Errorf("failed to compare branches: %w", err)
	}

	// Create the pull request
	pr := models.PullRequest{
		IssueID:            issue.ID,
		HeadRepositoryID:   &headRepoID,
		HeadRef:            req.Head,
		BaseRepositoryID:   repoID,
		BaseRef:            req.Base,
		Draft:              req.Draft,
		Additions:          comparison.Additions,
		Deletions:          comparison.Deletions,
		ChangedFiles:       len(comparison.Files),
		MergeableState:     "unknown",
	}

	if err := s.db.Create(&pr).Error; err != nil {
		return nil, err
	}

	// Store file changes
	for _, file := range comparison.Files {
		prFile := models.PullRequestFile{
			PullRequestID: pr.ID,
			Filename:      file.Path,
			Status:        file.Status,
			Additions:     file.Additions,
			Deletions:     file.Deletions,
			Changes:       file.Additions + file.Deletions,
			Patch:         file.Patch,
		}
		if file.PrevPath != "" && file.PrevPath != file.Path {
			prFile.PreviousFilename = &file.PrevPath
		}
		s.db.Create(&prFile)
	}

	// Check mergeability
	mergeable, err := s.checkMergeability(repoPath, req.Base, req.Head)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to check mergeability")
	} else {
		pr.Mergeable = &mergeable
		if mergeable {
			pr.MergeableState = "clean"
		} else {
			pr.MergeableState = "dirty"
		}
		s.db.Save(&pr)
	}

	// Load relationships
	s.db.Preload("Issue").Preload("Issue.User").Preload("HeadRepository").Preload("BaseRepository").First(&pr, pr.ID)

	return &pr, nil
}

func (s *PullRequestService) GetPullRequest(repoID uuid.UUID, number int) (*models.PullRequest, error) {
	var pr models.PullRequest
	err := s.db.Joins("JOIN issues ON pull_requests.issue_id = issues.id").
		Preload("Issue").
		Preload("Issue.User").
		Preload("HeadRepository").
		Preload("BaseRepository").
		Where("issues.repository_id = ? AND issues.number = ?", repoID, number).
		First(&pr).Error
	
	if err != nil {
		return nil, err
	}

	return &pr, nil
}

func (s *PullRequestService) ListPullRequests(repoID uuid.UUID, opts PullRequestListOptions) ([]models.PullRequest, int64, error) {
	query := s.db.Joins("JOIN issues ON pull_requests.issue_id = issues.id").
		Preload("Issue").
		Preload("Issue.User").
		Preload("HeadRepository").
		Preload("BaseRepository").
		Where("issues.repository_id = ?", repoID)

	// Apply filters
	if opts.State != "" {
		query = query.Where("issues.state = ?", opts.State)
	}
	if opts.Head != "" {
		query = query.Where("pull_requests.head_ref = ?", opts.Head)
	}
	if opts.Base != "" {
		query = query.Where("pull_requests.base_ref = ?", opts.Base)
	}

	// Count total
	var total int64
	query.Model(&models.PullRequest{}).Count(&total)

	// Apply sorting
	orderBy := "issues.created_at DESC"
	if opts.Sort != "" {
		direction := "DESC"
		if opts.Direction == "asc" {
			direction = "ASC"
		}
		switch opts.Sort {
		case "created":
			orderBy = fmt.Sprintf("issues.created_at %s", direction)
		case "updated":
			orderBy = fmt.Sprintf("issues.updated_at %s", direction)
		}
	}
	query = query.Order(orderBy)

	// Apply pagination
	if opts.Page > 0 && opts.PerPage > 0 {
		offset := (opts.Page - 1) * opts.PerPage
		query = query.Offset(offset).Limit(opts.PerPage)
	}

	var prs []models.PullRequest
	err := query.Find(&prs).Error
	return prs, total, err
}

func (s *PullRequestService) UpdatePullRequest(repoID uuid.UUID, number int, updates map[string]interface{}) (*models.PullRequest, error) {
	pr, err := s.GetPullRequest(repoID, number)
	if err != nil {
		return nil, err
	}

	// Update the issue
	issueUpdates := make(map[string]interface{})
	if title, ok := updates["title"]; ok {
		issueUpdates["title"] = title
	}
	if body, ok := updates["body"]; ok {
		issueUpdates["body"] = body
	}
	if state, ok := updates["state"]; ok {
		issueUpdates["state"] = state
		if state == "closed" {
			now := time.Now()
			issueUpdates["closed_at"] = &now
		}
	}

	if len(issueUpdates) > 0 {
		if err := s.db.Model(&pr.Issue).Updates(issueUpdates).Error; err != nil {
			return nil, err
		}
	}

	// Update the pull request
	prUpdates := make(map[string]interface{})
	if draft, ok := updates["draft"]; ok {
		prUpdates["draft"] = draft
	}

	if len(prUpdates) > 0 {
		if err := s.db.Model(pr).Updates(prUpdates).Error; err != nil {
			return nil, err
		}
	}

	// Reload with relationships
	return s.GetPullRequest(repoID, number)
}

func (s *PullRequestService) MergePullRequest(repoID uuid.UUID, number int, userID uuid.UUID, mergeMethod models.MergeMethod, commitTitle, commitMessage string) (*models.PullRequest, error) {
	pr, err := s.GetPullRequest(repoID, number)
	if err != nil {
		return nil, err
	}

	if pr.Issue.State == models.IssueStateClosed {
		return nil, errors.New("pull request is already closed")
	}

	if pr.Merged {
		return nil, errors.New("pull request is already merged")
	}

	// Check if mergeable
	if pr.Mergeable != nil && !*pr.Mergeable {
		return nil, errors.New("pull request has conflicts and cannot be merged")
	}

	var repo models.Repository
	if err := s.db.First(&repo, "id = ?", repoID).Error; err != nil {
		return nil, err
	}

	ownerName, err := s.getRepositoryOwnerName(&repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository owner: %w", err)
	}
	
	repoPath := filepath.Join(s.repoBasePath, ownerName, repo.Name)

	// Perform the merge
	mergeCommitSHA, err := s.gitService.MergeBranches(repoPath, pr.BaseRef, pr.HeadRef, string(mergeMethod), commitTitle, commitMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to merge branches: %w", err)
	}

	// Update pull request
	now := time.Now()
	updates := map[string]interface{}{
		"merged":           true,
		"merged_at":        &now,
		"merged_by_id":     &userID,
		"merge_commit_sha": mergeCommitSHA,
	}
	if err := s.db.Model(pr).Updates(updates).Error; err != nil {
		return nil, err
	}

	// Close the issue
	if err := s.db.Model(&pr.Issue).Updates(map[string]interface{}{
		"state":     models.IssueStateClosed,
		"closed_at": &now,
	}).Error; err != nil {
		return nil, err
	}

	// Create merge record
	merge := models.PullRequestMerge{
		PullRequestID: pr.ID,
		MergeMethod:   mergeMethod,
		CommitTitle:   commitTitle,
		CommitMessage: commitMessage,
		MergedAt:      now,
		MergedByID:    &userID,
	}
	s.db.Create(&merge)

	return s.GetPullRequest(repoID, number)
}

func (s *PullRequestService) GetPullRequestFiles(repoID uuid.UUID, number int) ([]models.PullRequestFile, error) {
	pr, err := s.GetPullRequest(repoID, number)
	if err != nil {
		return nil, err
	}

	var files []models.PullRequestFile
	err = s.db.Where("pull_request_id = ?", pr.ID).Find(&files).Error
	return files, err
}

func (s *PullRequestService) checkMergeability(repoPath, base, head string) (bool, error) {
	return s.gitService.CanMerge(repoPath, base, head)
}

func (s *PullRequestService) getNextIssueNumber(repoID uuid.UUID) (int, error) {
	var maxNumber int
	err := s.db.Model(&models.Issue{}).
		Where("repository_id = ?", repoID).
		Select("COALESCE(MAX(number), 0)").
		Scan(&maxNumber).Error
	return maxNumber + 1, err
}

// Review-related methods

func (s *PullRequestService) CreateReview(repoID uuid.UUID, number int, userID uuid.UUID, commitSHA string, body string, state models.ReviewState, comments []CreateReviewCommentRequest) (*models.Review, error) {
	pr, err := s.GetPullRequest(repoID, number)
	if err != nil {
		return nil, err
	}

	review := models.Review{
		PullRequestID: pr.ID,
		UserID:        &userID,
		CommitSHA:     commitSHA,
		State:         state,
		Body:          body,
	}

	if state != models.ReviewStatePending {
		now := time.Now()
		review.SubmittedAt = &now
	}

	if err := s.db.Create(&review).Error; err != nil {
		return nil, err
	}

	// Create review comments
	for _, commentReq := range comments {
		comment := models.ReviewComment{
			ReviewID:         &review.ID,
			PullRequestID:    pr.ID,
			UserID:           &userID,
			CommitSHA:        commitSHA,
			Path:             commentReq.Path,
			Position:         commentReq.Position,
			OriginalPosition: commentReq.OriginalPosition,
			Line:             commentReq.Line,
			OriginalLine:     commentReq.OriginalLine,
			Side:             commentReq.Side,
			StartLine:        commentReq.StartLine,
			StartSide:        commentReq.StartSide,
			Body:             commentReq.Body,
		}
		s.db.Create(&comment)
	}

	// Load relationships
	s.db.Preload("User").Preload("ReviewComments").First(&review, review.ID)
	return &review, nil
}

type CreateReviewCommentRequest struct {
	Path             string `json:"path" binding:"required"`
	Position         *int   `json:"position"`
	OriginalPosition *int   `json:"original_position"`
	Line             *int   `json:"line"`
	OriginalLine     *int   `json:"original_line"`
	Side             string `json:"side"`
	StartLine        *int   `json:"start_line"`
	StartSide        string `json:"start_side"`
	Body             string `json:"body" binding:"required"`
}

func (s *PullRequestService) ListReviews(repoID uuid.UUID, number int) ([]models.Review, error) {
	pr, err := s.GetPullRequest(repoID, number)
	if err != nil {
		return nil, err
	}

	var reviews []models.Review
	err = s.db.Where("pull_request_id = ?", pr.ID).
		Preload("User").
		Preload("ReviewComments").
		Find(&reviews).Error
	return reviews, err
}

func (s *PullRequestService) CreateReviewComment(repoID uuid.UUID, number int, userID uuid.UUID, req CreateReviewCommentRequest) (*models.ReviewComment, error) {
	pr, err := s.GetPullRequest(repoID, number)
	if err != nil {
		return nil, err
	}

	// Get the latest commit SHA
	var repo models.Repository
	if err := s.db.First(&repo, "id = ?", repoID).Error; err != nil {
		return nil, err
	}

	ownerName, err := s.getRepositoryOwnerName(&repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository owner: %w", err)
	}
	
	repoPath := filepath.Join(s.repoBasePath, ownerName, repo.Name)
	commitSHA, err := s.gitService.GetBranchCommit(repoPath, pr.HeadRef)
	if err != nil {
		return nil, fmt.Errorf("failed to get head commit: %w", err)
	}

	comment := models.ReviewComment{
		PullRequestID:    pr.ID,
		UserID:           &userID,
		CommitSHA:        commitSHA,
		Path:             req.Path,
		Position:         req.Position,
		OriginalPosition: req.OriginalPosition,
		Line:             req.Line,
		OriginalLine:     req.OriginalLine,
		Side:             req.Side,
		StartLine:        req.StartLine,
		StartSide:        req.StartSide,
		Body:             req.Body,
	}

	if err := s.db.Create(&comment).Error; err != nil {
		return nil, err
	}

	s.db.Preload("User").First(&comment, comment.ID)
	return &comment, nil
}

func (s *PullRequestService) ListReviewComments(repoID uuid.UUID, number int) ([]models.ReviewComment, error) {
	pr, err := s.GetPullRequest(repoID, number)
	if err != nil {
		return nil, err
	}

	var comments []models.ReviewComment
	err = s.db.Where("pull_request_id = ?", pr.ID).
		Preload("User").
		Preload("Review").
		Find(&comments).Error
	return comments, err
}

// Helper function to get repository owner name
func (s *PullRequestService) getRepositoryOwnerName(repo *models.Repository) (string, error) {
	if repo.OwnerType == models.OwnerTypeUser {
		var user models.User
		if err := s.db.First(&user, "id = ?", repo.OwnerID).Error; err != nil {
			return "", err
		}
		return user.Username, nil
	} else {
		var org models.Organization
		if err := s.db.First(&org, "id = ?", repo.OwnerID).Error; err != nil {
			return "", err
		}
		return org.Name, nil
	}
}