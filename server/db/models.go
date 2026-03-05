package db

import (
	"time"
)

// Repo represents a model or dataset repository
type Repo struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	RepoID    string    `gorm:"uniqueIndex;not null" json:"id"`       // HfApi expects "id" for repo_id
	Namespace string    `gorm:"index;not null" json:"namespace"`
	Name      string    `gorm:"not null" json:"modelId"`              // HfApi expects "modelId" for name
	Type      string    `gorm:"not null;default:'model'" json:"type"` // model, dataset
	Private   bool      `gorm:"default:false" json:"private"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`     // HfApi expects "createdAt"
	UpdatedAt time.Time `json:"-" gorm:"autoUpdateTime"`              // Hidden from JSON
}

// MarshalJSON customizes JSON output for HfApi compatibility
func (r Repo) MarshalJSON() ([]byte, error) {
	type Alias Repo
	return []byte(`{"id":"` + r.RepoID + `","modelId":"` + r.Name + `","namespace":"` + r.Namespace + `","type":"` + r.Type + `","private":` + boolStr(r.Private) + `,"createdAt":"` + r.CreatedAt.UTC().Format("2006-01-02T15:04:05.000000Z") + `"}`), nil
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

type Commit struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	RepoID    string    `gorm:"index;not null" json:"repo_id"`
	CommitID  string    `gorm:"uniqueIndex;not null" json:"commit_id"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type FileIndex struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	RepoID    string    `gorm:"index;not null" json:"repo_id"`
	CommitID  string    `gorm:"index;not null" json:"commit_id"`
	Path      string    `gorm:"index;not null" json:"path"`
	Size      int64     `json:"size"`
	LFS       bool      `gorm:"default:false" json:"lfs"`
	SHA256    string    `json:"sha256"`
	CreatedAt time.Time `json:"created_at"`
}

// OAuthState stores OAuth state tokens for CSRF protection
type OAuthState struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	State     string    `gorm:"uniqueIndex;not null" json:"state"`
	Provider  string    `gorm:"index;not null" json:"provider"` // "hf", "github", etc.
	Status    string    `gorm:"default:'pending'" json:"status"` // "pending", "used", "expired"
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// IsExpired checks if the OAuth state has expired
func (s *OAuthState) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}
