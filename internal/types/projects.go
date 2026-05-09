// Package types mirrors the Everscribe API's wire types. Kept in sync
// by hand with github.com/everscribe/monorepo/internal/types — only
// the subset the CLI actually uses lives here, since the monorepo is a
// separate Go module that the CLI doesn't import directly.
package types

import "time"

type Project struct {
	ID        string    `json:"id" yaml:"id"`
	UserID    string    `json:"user_id" yaml:"user_id"`
	Name      string    `json:"name" yaml:"name"`
	CreatedAt time.Time `json:"created_at" yaml:"created_at"`
	UpdatedAt time.Time `json:"updated_at" yaml:"updated_at"`
}

type CreateProjectRequest struct {
	Name string `json:"name"`
}

type UpdateProjectRequest struct {
	Name string `json:"name"`
}

type ProjectResponse struct {
	Project Project `json:"project"`
}

type ListProjectsResponse struct {
	Projects []Project `json:"projects"`
}
