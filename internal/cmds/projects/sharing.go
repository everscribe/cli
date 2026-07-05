package projects

import (
	"context"
	"fmt"
	"strings"

	"github.com/everscribe/cli/internal/client"
)

func validProjectRole(role string) bool {
	switch role {
	case "viewer", "editor", "admin":
		return true
	}
	return false
}

// activeCollaboratorUserID resolves an email to the user id of an active
// collaborator. Sharing is strict: the person must already be an active
// collaborator (the API enforces this too), so this surfaces a helpful error
// rather than a raw 403 when they aren't.
func activeCollaboratorUserID(ctx context.Context, c *client.Client, email string) (string, error) {
	cols, err := c.ListCollaborators(ctx)
	if err != nil {
		return "", err
	}
	for _, col := range cols {
		if strings.EqualFold(col.Email, email) {
			if col.Status != "active" || col.UserID == "" {
				return "", fmt.Errorf("%s hasn't created an Everscribe account yet - you can only share with active collaborators", email)
			}
			return col.UserID, nil
		}
	}
	return "", fmt.Errorf("%s is not one of your collaborators - add them first with `es collaborators add %s`", email, email)
}
