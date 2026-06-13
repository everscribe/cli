package client

import (
	"context"
	"net/http"

	"github.com/everscribe/cli/internal/types"
)

// ListCollaborators returns the caller's collaborator list.
func (c *Client) ListCollaborators(ctx context.Context) ([]types.Collaborator, error) {
	var out types.ListCollaboratorsResponse
	if err := c.Do(ctx, http.MethodGet, "/v1/collaborators", nil, nil, &out); err != nil {
		return nil, err
	}
	return out.Collaborators, nil
}

// AddCollaborator adds someone to the caller's collaborator list by email.
func (c *Client) AddCollaborator(ctx context.Context, email string) (*types.Collaborator, error) {
	var out types.AddCollaboratorResponse
	body := types.AddCollaboratorRequest{Email: email}
	if err := c.Do(ctx, http.MethodPost, "/v1/collaborators", nil, body, &out); err != nil {
		return nil, err
	}
	return &out.Collaborator, nil
}

// RemoveCollaborator removes a collaborator (and revokes their grants on the
// caller's projects).
func (c *Client) RemoveCollaborator(ctx context.Context, id string) error {
	return c.Do(ctx, http.MethodDelete, "/v1/collaborators/"+id, nil, nil, nil)
}

// ListProjectMembers returns everyone with access to a project + their role.
func (c *Client) ListProjectMembers(ctx context.Context, projectID string) (*types.ListProjectMembersResponse, error) {
	var out types.ListProjectMembersResponse
	if err := c.Do(ctx, http.MethodGet, "/v1/projects/"+projectID+"/members", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SetProjectRole grants (or changes) a member's role on a project. The API
// requires the target to be an active collaborator of the caller.
func (c *Client) SetProjectRole(ctx context.Context, projectID, userID, role string) error {
	body := types.SetProjectRoleRequest{Role: role}
	return c.Do(ctx, http.MethodPut, "/v1/projects/"+projectID+"/members/"+userID, nil, body, nil)
}

// RemoveProjectMember revokes a member's access to a project.
func (c *Client) RemoveProjectMember(ctx context.Context, projectID, userID string) error {
	return c.Do(ctx, http.MethodDelete, "/v1/projects/"+projectID+"/members/"+userID, nil, nil, nil)
}
