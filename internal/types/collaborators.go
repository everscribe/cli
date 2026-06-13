package types

import "time"

// Collaborator is one entry in the caller's collaborator list. UserID/Username
// are populated only when the collaborator has an account (status "active").
type Collaborator struct {
	ID        string                `json:"id" yaml:"id"`
	Email     string                `json:"email" yaml:"email"`
	UserID    string                `json:"user_id,omitempty" yaml:"user_id,omitempty"`
	Username  string                `json:"username,omitempty" yaml:"username,omitempty"`
	Status    string                `json:"status" yaml:"status"` // "active" | "pending"
	Projects  []CollaboratorProject `json:"projects,omitempty" yaml:"projects,omitempty"`
	CreatedAt time.Time             `json:"created_at" yaml:"created_at"`
}

// CollaboratorProject is one of the caller's projects a collaborator can access.
type CollaboratorProject struct {
	ID   string `json:"id" yaml:"id"`
	Name string `json:"name" yaml:"name"`
	Role string `json:"role" yaml:"role"`
}

type AddCollaboratorRequest struct {
	Email string `json:"email"`
}

type ListCollaboratorsResponse struct {
	Collaborators []Collaborator `json:"collaborators"`
}

type AddCollaboratorResponse struct {
	Collaborator Collaborator `json:"collaborator"`
}

// ProjectMemberView is one row of a project's member list: a person with access
// plus their effective project role.
type ProjectMemberView struct {
	UserID      string `json:"user_id" yaml:"user_id"`
	Username    string `json:"username" yaml:"username"`
	Email       string `json:"email" yaml:"email"`
	AccountRole string `json:"account_role" yaml:"account_role"` // owner / admin / "" (grantee)
	ProjectRole string `json:"project_role" yaml:"project_role"` // admin / editor / viewer / ""
}

type ListProjectMembersResponse struct {
	Members    []ProjectMemberView `json:"members"`
	IsPersonal bool                `json:"is_personal"`
	CanManage  bool                `json:"can_manage"`
}

type SetProjectRoleRequest struct {
	Role string `json:"role"`
}
