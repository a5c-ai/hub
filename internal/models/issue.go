package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IssueState string

const (
	IssueStateOpen   IssueState = "open"
	IssueStateClosed IssueState = "closed"
)

type Issue struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	RepositoryID  uuid.UUID  `json:"repository_id" gorm:"type:uuid;not null;index"`
	Number        int        `json:"number" gorm:"not null"`
	Title         string     `json:"title" gorm:"not null;size:255"`
	Body          string     `json:"body" gorm:"type:text"`
	UserID        *uuid.UUID `json:"user_id" gorm:"type:uuid;index"`
	AssigneeID    *uuid.UUID `json:"assignee_id" gorm:"type:uuid;index"`
	MilestoneID   *uuid.UUID `json:"milestone_id" gorm:"type:uuid;index"`
	State         IssueState `json:"state" gorm:"type:varchar(50);not null;check:state IN ('open','closed')"`
	StateReason   string     `json:"state_reason" gorm:"size:50"`
	Locked        bool       `json:"locked" gorm:"default:false"`
	CommentsCount int        `json:"comments_count" gorm:"default:0"`
	ClosedAt      *time.Time `json:"closed_at"`

	// Relationships
	Repository  Repository   `json:"repository,omitempty" gorm:"foreignKey:RepositoryID"`
	User        *User        `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Assignee    *User        `json:"assignee,omitempty" gorm:"foreignKey:AssigneeID"`
	Milestone   *Milestone   `json:"milestone,omitempty" gorm:"foreignKey:MilestoneID"`
	Comments    []Comment    `json:"comments,omitempty" gorm:"foreignKey:IssueID"`
	PullRequest *PullRequest `json:"pull_request,omitempty" gorm:"foreignKey:IssueID"`
	Labels      []Label      `json:"labels,omitempty" gorm:"many2many:issue_labels"`
}

func (i *Issue) TableName() string {
	return "issues"
}

type PullRequest struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	IssueID          uuid.UUID  `json:"issue_id" gorm:"type:uuid;not null;uniqueIndex"`
	HeadRepositoryID *uuid.UUID `json:"head_repository_id" gorm:"type:uuid;index"`
	HeadRef          string     `json:"head_ref" gorm:"not null;size:255"`
	BaseRepositoryID uuid.UUID  `json:"base_repository_id" gorm:"type:uuid;not null;index"`
	BaseRef          string     `json:"base_ref" gorm:"not null;size:255"`
	MergeCommitSHA   string     `json:"merge_commit_sha" gorm:"size:40"`
	Merged           bool       `json:"merged" gorm:"default:false"`
	MergedAt         *time.Time `json:"merged_at"`
	MergedByID       *uuid.UUID `json:"merged_by_id" gorm:"type:uuid;index"`
	Draft            bool       `json:"draft" gorm:"default:false"`
	Mergeable        *bool      `json:"mergeable"`
	MergeableState   string     `json:"mergeable_state" gorm:"size:50"`
	Additions        int        `json:"additions" gorm:"default:0"`
	Deletions        int        `json:"deletions" gorm:"default:0"`
	ChangedFiles     int        `json:"changed_files" gorm:"default:0"`

	// Relationships
	Issue          Issue       `json:"issue,omitempty" gorm:"foreignKey:IssueID"`
	HeadRepository *Repository `json:"head_repository,omitempty" gorm:"foreignKey:HeadRepositoryID"`
	BaseRepository Repository  `json:"base_repository,omitempty" gorm:"foreignKey:BaseRepositoryID"`
	MergedBy       *User       `json:"merged_by,omitempty" gorm:"foreignKey:MergedByID"`
}

func (pr *PullRequest) TableName() string {
	return "pull_requests"
}

type Comment struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	IssueID uuid.UUID  `json:"issue_id" gorm:"type:uuid;not null;index"`
	UserID  *uuid.UUID `json:"user_id" gorm:"type:uuid;index"`
	Body    string     `json:"body" gorm:"not null;type:text"`

	// Relationships
	Issue Issue `json:"issue,omitempty" gorm:"foreignKey:IssueID"`
	User  *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (c *Comment) TableName() string {
	return "comments"
}

type Milestone struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	RepositoryID uuid.UUID  `json:"repository_id" gorm:"type:uuid;not null;index"`
	Number       int        `json:"number" gorm:"not null"`
	Title        string     `json:"title" gorm:"not null;size:255"`
	Description  string     `json:"description" gorm:"type:text"`
	State        string     `json:"state" gorm:"not null;size:50;check:state IN ('open','closed')"`
	DueOn        *time.Time `json:"due_on"`
	ClosedAt     *time.Time `json:"closed_at"`

	// Relationships
	Repository Repository `json:"repository,omitempty" gorm:"foreignKey:RepositoryID"`
	Issues     []Issue    `json:"issues,omitempty" gorm:"foreignKey:MilestoneID"`
}

func (m *Milestone) TableName() string {
	return "milestones"
}

type Label struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	RepositoryID uuid.UUID `json:"repository_id" gorm:"type:uuid;not null;index"`
	Name         string    `json:"name" gorm:"not null;size:255"`
	Color        string    `json:"color" gorm:"not null;size:7"`
	Description  string    `json:"description" gorm:"type:text"`

	// Relationships
	Repository Repository `json:"repository,omitempty" gorm:"foreignKey:RepositoryID"`
	Issues     []Issue    `json:"issues,omitempty" gorm:"many2many:issue_labels"`
}

func (l *Label) TableName() string {
	return "labels"
}

// IssueLabel is the join table for many-to-many relationship between issues and labels
type IssueLabel struct {
	IssueID uuid.UUID `json:"issue_id" gorm:"type:uuid;primaryKey"`
	LabelID uuid.UUID `json:"label_id" gorm:"type:uuid;primaryKey"`
}

func (il *IssueLabel) TableName() string {
	return "issue_labels"
}
