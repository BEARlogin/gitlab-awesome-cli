package gitlab

import (
	"fmt"
	"log"
	"net/http"
	"time"

	gogitlab "github.com/xanzy/go-gitlab"
)

func NewClient(baseURL, token string) (*gogitlab.Client, error) {
	log.Printf("[gitlab] creating client: url=%s", baseURL)
	httpClient := &http.Client{Timeout: 15 * time.Second}
	client, err := gogitlab.NewClient(token,
		gogitlab.WithBaseURL(baseURL+"/api/v4"),
		gogitlab.WithHTTPClient(httpClient),
		gogitlab.WithCustomRetryMax(2),
	)
	if err != nil {
		return nil, fmt.Errorf("creating gitlab client: %w", err)
	}
	return client, nil
}
