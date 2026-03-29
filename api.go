package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const graphqlURL = "https://api.linear.app/graphql"

type Client struct {
	apiKey string
	http   *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		http:   &http.Client{Timeout: 25 * time.Second},
	}
}

type Notification struct {
	ID        string   `json:"id"`
	Type      string   `json:"type"`
	ReadAt    *string  `json:"readAt"`
	CreatedAt string   `json:"createdAt"`
	Title     string   `json:"title"`
	Subtitle  string   `json:"subtitle"`
	URL       string   `json:"url"`
	Actor     *Actor        `json:"actor"`
	Issue     *Issue        `json:"issue"`
	Comment   *Comment      `json:"comment"`
	Project   *Project      `json:"project"`
	PullRequest *PullRequest `json:"pullRequest"`
}

type Actor struct {
	Name string `json:"name"`
}

type Issue struct {
	Identifier string      `json:"identifier"`
	Title      string      `json:"title"`
	URL        string      `json:"url"`
	State      *IssueState `json:"state"`
	Assignee   *Actor      `json:"assignee"`
	Team       *Team       `json:"team"`
	Priority   int         `json:"priority"`
	Labels     *LabelConn  `json:"labels"`
}

type IssueState struct {
	Name string `json:"name"`
}

type Team struct {
	Name string `json:"name"`
}

type Comment struct {
	Body string `json:"body"`
	User *Actor `json:"user"`
}

type Project struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type PullRequest struct {
	Title        string `json:"title"`
	URL          string `json:"url"`
	SourceBranch string `json:"sourceBranch"`
	TargetBranch string `json:"targetBranch"`
	Status       string `json:"status"`
}

type LabelConn struct {
	Nodes []Label `json:"nodes"`
}

type Label struct {
	Name string `json:"name"`
}

type PageInfo struct {
	HasNextPage bool   `json:"hasNextPage"`
	EndCursor   string `json:"endCursor"`
}

type gqlRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

type gqlResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []gqlError      `json:"errors"`
}

type gqlError struct {
	Message string `json:"message"`
}

func (c *Client) do(req gqlRequest) (json.RawMessage, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequest("POST", graphqlURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", c.apiKey)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	var gqlResp gqlResponse
	if err := json.NewDecoder(resp.Body).Decode(&gqlResp); err != nil {
		return nil, fmt.Errorf("invalid response: %w", err)
	}
	if len(gqlResp.Errors) > 0 {
		return gqlResp.Data, fmt.Errorf("graphql: %s", gqlResp.Errors[0].Message)
	}
	return gqlResp.Data, nil
}

const notificationsQuery = `query($after: String) {
  notifications(first: 50, after: $after) {
    nodes {
      id type readAt createdAt
      title subtitle url
      actor { name }
      ... on IssueNotification {
        issue {
          identifier title url
          state { name }
          assignee { name }
          team { name }
          priority
          labels { nodes { name } }
        }
        comment { body user { name } }
      }
      ... on ProjectNotification {
        project { name url }
      }
      ... on PullRequestNotification {
        pullRequest {
          title url
          sourceBranch targetBranch status
        }
      }
    }
    pageInfo { hasNextPage endCursor }
  }
}`

func (c *Client) FetchNotifications(after string) ([]Notification, *PageInfo, error) {
	vars := map[string]any{}
	if after != "" {
		vars["after"] = after
	}
	data, err := c.do(gqlRequest{Query: notificationsQuery, Variables: vars})
	if err != nil {
		return nil, nil, err
	}
	var result struct {
		Notifications struct {
			Nodes    []Notification `json:"nodes"`
			PageInfo PageInfo       `json:"pageInfo"`
		} `json:"notifications"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, nil, err
	}
	return result.Notifications.Nodes, &result.Notifications.PageInfo, nil
}

const unreadCountQuery = `query { notificationsUnreadCount }`

// Light query used only for the unread-count fallback.
const unreadCountFallbackQuery = `query($after: String) {
  notifications(first: 50, after: $after) {
    nodes { readAt }
    pageInfo { hasNextPage endCursor }
  }
}`

// FetchUnreadCount tries the fast notificationsUnreadCount field first.
// If the API key doesn't support it, falls back to paginating notifications.
func (c *Client) FetchUnreadCount() (int, error) {
	data, err := c.do(gqlRequest{Query: unreadCountQuery})
	if err == nil {
		var result struct {
			Count int `json:"notificationsUnreadCount"`
		}
		if json.Unmarshal(data, &result) == nil && result.Count >= 0 {
			return result.Count, nil
		}
	}
	return c.countUnreadFallback()
}

func (c *Client) countUnreadFallback() (int, error) {
	total := 0
	after := ""
	for page := 0; page < 20; page++ {
		vars := map[string]any{}
		if after != "" {
			vars["after"] = after
		}
		data, err := c.do(gqlRequest{Query: unreadCountFallbackQuery, Variables: vars})
		if err != nil {
			return 0, err
		}
		var result struct {
			Notifications struct {
				Nodes    []struct{ ReadAt *string `json:"readAt"` } `json:"nodes"`
				PageInfo PageInfo                                   `json:"pageInfo"`
			} `json:"notifications"`
		}
		if err := json.Unmarshal(data, &result); err != nil {
			return 0, err
		}
		for _, n := range result.Notifications.Nodes {
			if n.ReadAt == nil {
				total++
			}
		}
		if !result.Notifications.PageInfo.HasNextPage || result.Notifications.PageInfo.EndCursor == "" {
			break
		}
		after = result.Notifications.PageInfo.EndCursor
	}
	return total, nil
}

func (c *Client) MarkAsRead(id string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	query := `mutation($id: String!, $readAt: DateTime!) {
		notificationUpdate(id: $id, input: { readAt: $readAt }) { success }
	}`
	_, err := c.do(gqlRequest{
		Query:     query,
		Variables: map[string]any{"id": id, "readAt": now},
	})
	return err
}
