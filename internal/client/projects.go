package client

import (
	"context"
	"net/http"

	"github.com/everscribe/cli/internal/types"
)

// ListProjects returns all projects owned by the authenticated user.
// The API doesn't paginate this endpoint, so a single call is enough.
func (c *Client) ListProjects(ctx context.Context) ([]types.Project, error) {
	var out types.ListProjectsResponse
	if err := c.Do(ctx, http.MethodGet, "/v1/projects", nil, nil, &out); err != nil {
		return nil, err
	}
	return out.Projects, nil
}

func (c *Client) GetProject(ctx context.Context, id string) (*types.Project, error) {
	var out types.ProjectResponse
	if err := c.Do(ctx, http.MethodGet, "/v1/projects/"+id, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out.Project, nil
}

func (c *Client) CreateProject(ctx context.Context, name string) (*types.Project, error) {
	var out types.ProjectResponse
	body := types.CreateProjectRequest{Name: name}
	if err := c.Do(ctx, http.MethodPost, "/v1/projects", nil, body, &out); err != nil {
		return nil, err
	}
	return &out.Project, nil
}

func (c *Client) UpdateProject(ctx context.Context, id, name string) (*types.Project, error) {
	var out types.ProjectResponse
	body := types.UpdateProjectRequest{Name: name}
	if err := c.Do(ctx, http.MethodPut, "/v1/projects/"+id, nil, body, &out); err != nil {
		return nil, err
	}
	return &out.Project, nil
}

func (c *Client) DeleteProject(ctx context.Context, id string) error {
	return c.Do(ctx, http.MethodDelete, "/v1/projects/"+id, nil, nil, nil)
}
