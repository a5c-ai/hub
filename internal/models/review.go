package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReviewState string

const (
	ReviewStatePending          ReviewState = "pending"
	ReviewStateApproved         ReviewState = "approved"
	ReviewStateRequestChanges   ReviewState = "request_changes"
	ReviewStateCommented        ReviewState = "commented"
	ReviewStateDismissed        ReviewState = "dismissed"
)

type Review struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	PullRequestID uuid.UUID   `json:"pull_request_id" gorm:"type:uuid;not null;index"`
	UserID        *uuid.UUID  `json:"user_id" gorm:"type:uuid;index"`
	CommitSHA     string      `json:"commit_sha" gorm:"not null;size:40"`
	State         ReviewState `json:"state" gorm:"type:varchar(50);not null;check:state IN ('pending','approved','request_changes','commented','dismissed')"`
	Body          string      `json:"body" gorm:"type:text"`
	SubmittedAt   *time.Time  `json:"submitted_at"`

	// Relationships
	PullRequest    PullRequest      `json:"pull_request,omitempty" gorm:"foreignKey:PullRequestID"`
	User           *User            `json:"user,omitempty" gorm:"foreignKey:UserID"`
	ReviewComments []ReviewComment  `json:"review_comments,omitempty" gorm:"foreignKey:ReviewID"`
}

func (r *Review) TableName() string {
	return "reviews"
}

type ReviewComment struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	ReviewID         *uuid.UUID `json:"review_id" gorm:"type:uuid;index"`
	PullRequestID    uuid.UUID  `json:"pull_request_id" gorm:"type:uuid;not null;index"`
	UserID           *uuid.UUID `json:"user_id" gorm:"type:uuid;index"`
	CommitSHA        string     `json:"commit_sha" gorm:"not null;size:40"`
	Path             string     `json:"path" gorm:"not null;size:4096"`
	Position         *int       `json:"position"`          // Position in the diff
	OriginalPosition *int       `json:"original_position"` // Original position in the diff
	Line             *int       `json:"line"`              // Line number in the file
	OriginalLine     *int       `json:"original_line"`     // Original line number in the file
	Side             string     `json:"side" gorm:"size:10;check:side IN ('LEFT','RIGHT')"`
	StartLine        *int       `json:"start_line"`        // First line of a multi-line comment
	StartSide        string     `json:"start_side" gorm:"size:10;check:start_side IN ('LEFT','RIGHT')"`
	Body             string     `json:"body" gorm:"not null;type:text"`
	InReplyToID      *uuid.UUID `json:"in_reply_to_id" gorm:"type:uuid;index"`

	// Relationships
	Review       *Review       `json:"review,omitempty" gorm:"foreignKey:ReviewID"`
	PullRequest  PullRequest   `json:"pull_request,omitempty" gorm:"foreignKey:PullRequestID"`
	User         *User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
	InReplyTo    *ReviewComment `json:"in_reply_to,omitempty" gorm:"foreignKey:InReplyToID"`
	Replies      []ReviewComment `json:"replies,omitempty" gorm:"foreignKey:InReplyToID"`
}

func (rc *ReviewComment) TableName() string {
	return "review_comments"
}

type PullRequestFile struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	PullRequestID uuid.UUID `json:"pull_request_id" gorm:"type:uuid;not null;index"`
	Filename      string    `json:"filename" gorm:"not null;size:4096"`
	Status        string    `json:"status" gorm:"not null;size:20;check:status IN ('added','deleted','modified','renamed','copied')"`
	Additions     int       `json:"additions" gorm:"default:0"`
	Deletions     int       `json:"deletions" gorm:"default:0"`
	Changes       int       `json:"changes" gorm:"default:0"`
	Patch         string    `json:"patch" gorm:"type:text"`
	PreviousFilename *string `json:"previous_filename" gorm:"size:4096"`

	// Relationships
	PullRequest PullRequest `json:"pull_request,omitempty" gorm:"foreignKey:PullRequestID"`
}

func (prf *PullRequestFile) TableName() string {
	return "pull_request_files"
}

type MergeMethod string

const (
	MergeMethodMerge  MergeMethod = "merge"
	MergeMethodSquash MergeMethod = "squash"
	MergeMethodRebase MergeMethod = "rebase"
)

type PullRequestMerge struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	PullRequestID uuid.UUID   `json:"pull_request_id" gorm:"type:uuid;not null;uniqueIndex"`
	MergeMethod   MergeMethod `json:"merge_method" gorm:"type:varchar(20);not null;check:merge_method IN ('merge','squash','rebase')"`
	CommitTitle   string      `json:"commit_title" gorm:"size:255"`
	CommitMessage string      `json:"commit_message" gorm:"type:text"`
	MergedAt      time.Time   `json:"merged_at"`
	MergedByID    *uuid.UUID  `json:"merged_by_id" gorm:"type:uuid;index"`

	// Relationships
	PullRequest PullRequest `json:"pull_request,omitempty" gorm:"foreignKey:PullRequestID"`
	MergedBy    *User       `json:"merged_by,omitempty" gorm:"foreignKey:MergedByID"`
}

func (prm *PullRequestMerge) TableName() string {
	return "pull_request_merges"
}