package gitlab

import (
	"fmt"

	gogitlab "github.com/xanzy/go-gitlab"
)

func NewClient(baseURL, token string) (*gogitlab.Client, error) {
	client, err := gogitlab.NewClient(token, gogitlab.WithBaseURL(baseURL+"/api/v4"))
	if err != nil {
		return nil, fmt.Errorf("creating gitlab client: %w", err)
	}
	return client, nil
}
