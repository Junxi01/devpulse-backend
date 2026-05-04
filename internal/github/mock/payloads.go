package mock

import (
	"encoding/json"
	"time"
)

type pushPayload struct {
	Ref     string       `json:"ref"`
	Commits []pushCommit `json:"commits"`
}

type pushCommit struct {
	ID        string `json:"id"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	Author    struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Login string `json:"login"`
	} `json:"author"`
}

type pullRequestPayload struct {
	Action      string `json:"action"`
	PullRequest prBody `json:"pull_request"`
}

type prBody struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	State  string `json:"state"`
	User   struct {
		Login string `json:"login"`
	} `json:"user"`
	Base struct {
		Ref string `json:"ref"`
	} `json:"base"`
	Head struct {
		Ref string `json:"ref"`
	} `json:"head"`
	ChangedFiles int `json:"changed_files"`
	Additions    int `json:"additions"`
	Deletions    int `json:"deletions"`
}

type issuesPayload struct {
	Action string    `json:"action"`
	Issue  issueBody `json:"issue"`
}

type issueBody struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	State  string `json:"state"`
	User   struct {
		Login string `json:"login"`
	} `json:"user"`
	Labels []struct {
		Name string `json:"name"`
	} `json:"labels"`
}

func parseTime(s string, fallback time.Time) time.Time {
	if s == "" {
		return fallback
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return fallback
	}
	return t
}

func authorName(c pushCommit, fallback string) string {
	if c.Author.Name != "" {
		return c.Author.Name
	}
	if c.Author.Login != "" {
		return c.Author.Login
	}
	return fallback
}

func ensurePayloadRaw(payload json.RawMessage) json.RawMessage {
	if len(payload) == 0 {
		return json.RawMessage(`{}`)
	}
	return payload
}
