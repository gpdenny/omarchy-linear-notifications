package main

import "time"

func ptr(s string) *string { return &s }

func newDemoModel() model {
	now := time.Now().UTC()
	readAt := now.Add(-2 * time.Hour).Format(time.RFC3339)

	notifications := []Notification{
		{
			ID:        "demo-1",
			Type:      "issueAssignedToYou",
			CreatedAt: now.Add(-25 * time.Minute).Format(time.RFC3339),
			Title:     "Migrate auth service to OIDC",
			Subtitle:  "Alex Chen assigned ENG-482 to you",
			URL:       "https://linear.app",
			Actor:     &Actor{Name: "Alex Chen"},
			Issue: &Issue{
				Identifier: "ENG-482",
				Title:      "Migrate auth service to OIDC",
				URL:        "https://linear.app",
				Team:       &Team{Name: "Engineering"},
				State:      &IssueState{Name: "In Progress"},
				Assignee:   &Actor{Name: "Jamie Rivera"},
				Priority:   1,
				Labels:     &LabelConn{Nodes: []Label{{Name: "Backend"}, {Name: "Security"}}},
			},
		},
		{
			ID:        "demo-2",
			Type:      "issueComment",
			CreatedAt: now.Add(-2 * time.Hour).Format(time.RFC3339),
			Title:     "Rate limiter dropping websocket connections",
			Subtitle:  "Morgan Wu commented on ENG-351",
			URL:       "https://linear.app",
			Actor:     &Actor{Name: "Morgan Wu"},
			Issue: &Issue{
				Identifier: "ENG-351",
				Title:      "Rate limiter dropping websocket connections",
				URL:        "https://linear.app",
				Team:       &Team{Name: "Engineering"},
				State:      &IssueState{Name: "In Review"},
				Assignee:   &Actor{Name: "Jamie Rivera"},
				Labels:     &LabelConn{Nodes: []Label{{Name: "Bug"}, {Name: "Urgent"}}},
			},
			Comment: &Comment{
				User: &Actor{Name: "Morgan Wu"},
				Body: "I traced this to the token bucket refill logic. When the bucket is empty and a burst of WS frames arrives, the limiter rejects them before the refill tick fires.\n\nThe fix should be straightforward — we need to pre-fill on first connection rather than waiting for the ticker.",
			},
		},
		{
			ID:        "demo-3",
			Type:      "pullRequestCommented",
			CreatedAt: now.Add(-4 * time.Hour).Format(time.RFC3339),
			Title:     "feat: ENG-510 Add request signing middleware",
			Subtitle:  "Sam Patel commented: Looks good overall, but the HMAC comparison on line 84 should use constant-time compare to avoid timing attacks.",
			URL:       "https://linear.app",
			Actor:     &Actor{Name: "Sam Patel"},
			PullRequest: &PullRequest{
				Title:        "feat: ENG-510 Add request signing middleware",
				URL:          "https://linear.app",
				SourceBranch: "sam/eng-510-request-signing",
				TargetBranch: "main",
				Status:       "open",
			},
		},
		{
			ID:        "demo-4",
			Type:      "issueStatusChanged",
			ReadAt:    &readAt,
			CreatedAt: now.Add(-6 * time.Hour).Format(time.RFC3339),
			Title:     "Webhook retry backoff not respecting config",
			Subtitle:  "Status changed to Done",
			URL:       "https://linear.app",
			Actor:     &Actor{Name: "Taylor Brooks"},
			Issue: &Issue{
				Identifier: "BEN-894",
				Title:      "Webhook retry backoff not respecting config",
				URL:        "https://linear.app",
				Team:       &Team{Name: "Backend"},
				State:      &IssueState{Name: "Done"},
				Assignee:   &Actor{Name: "Taylor Brooks"},
				Labels:     &LabelConn{Nodes: []Label{{Name: "Bug"}}},
			},
		},
		{
			ID:        "demo-5",
			Type:      "projectUpdate",
			ReadAt:    &readAt,
			CreatedAt: now.Add(-1 * 24 * time.Hour).Format(time.RFC3339),
			Title:     "Q3 Platform reliability roadmap",
			Subtitle:  "Jordan Lee posted an update",
			URL:       "https://linear.app",
			Actor:     &Actor{Name: "Jordan Lee"},
			Project: &Project{
				Name: "Q3 Platform reliability roadmap",
				URL:  "https://linear.app",
			},
		},
		{
			ID:        "demo-6",
			Type:      "issueMention",
			ReadAt:    &readAt,
			CreatedAt: now.Add(-2 * 24 * time.Hour).Format(time.RFC3339),
			Title:     "Add OpenTelemetry tracing to payment flow",
			Subtitle:  "Casey Kim mentioned you in ENG-621",
			URL:       "https://linear.app",
			Actor:     &Actor{Name: "Casey Kim"},
			Issue: &Issue{
				Identifier: "ENG-621",
				Title:      "Add OpenTelemetry tracing to payment flow",
				URL:        "https://linear.app",
				Team:       &Team{Name: "Engineering"},
				State:      &IssueState{Name: "Todo"},
				Assignee:   &Actor{Name: "Casey Kim"},
				Labels:     &LabelConn{Nodes: []Label{{Name: "Observability"}}},
			},
			Comment: &Comment{
				User: &Actor{Name: "Casey Kim"},
				Body: "@jamie can you review the span naming conventions here? Want to make sure we're consistent with the auth service traces you set up.",
			},
		},
		{
			ID:        "demo-7",
			Type:      "issueComment",
			ReadAt:    &readAt,
			CreatedAt: now.Add(-3 * 24 * time.Hour).Format(time.RFC3339),
			Title:     "DB connection pool exhaustion under load",
			Subtitle:  "Riley Chen commented on PLA-178",
			URL:       "https://linear.app",
			Actor:     &Actor{Name: "Riley Chen"},
			Issue: &Issue{
				Identifier: "PLA-178",
				Title:      "DB connection pool exhaustion under load",
				URL:        "https://linear.app",
				Team:       &Team{Name: "Platform"},
				State:      &IssueState{Name: "In Progress"},
				Assignee:   &Actor{Name: "Riley Chen"},
				Labels:     &LabelConn{Nodes: []Label{{Name: "Performance"}, {Name: "Infrastructure"}}},
			},
			Comment: &Comment{
				User: &Actor{Name: "Riley Chen"},
				Body: "Bumping max_connections to 200 and adding PgBouncer in transaction mode. Load test looks stable at 5k req/s now.",
			},
		},
		{
			ID:        "demo-8",
			Type:      "issueCreated",
			ReadAt:    &readAt,
			CreatedAt: now.Add(-4 * 24 * time.Hour).Format(time.RFC3339),
			Title:     "Audit log export to S3",
			Subtitle:  "New issue created by Alex Chen",
			URL:       "https://linear.app",
			Actor:     &Actor{Name: "Alex Chen"},
			Issue: &Issue{
				Identifier: "SEC-42",
				Title:      "Audit log export to S3",
				URL:        "https://linear.app",
				Team:       &Team{Name: "Security"},
				State:      &IssueState{Name: "Backlog"},
				Labels:     &LabelConn{Nodes: []Label{{Name: "Compliance"}}},
			},
		},
	}

	return model{
		notifications: notifications,
		unreadCount:   3,
	}
}
