package gitea

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	api "code.gitea.io/gitea/modules/structs"
)

// parse errors
var (
	ErrEventNotSpecifiedToParse    = errors.New("no Event specified to parse")
	ErrInvalidHTTPMethod           = errors.New("invalid HTTP Method")
	ErrMissingGiteaEventHeader     = errors.New("missing X-Gitea-Event Header")
	ErrMissingGiteaSignatureHeader = errors.New("missing X-Gitea-Signature Header")
	ErrEventNotFound               = errors.New("event not defined to be parsed")
	ErrParsingPayload              = errors.New("error parsing payload")
	ErrHMACVerificationFailed      = errors.New("HMAC verification failed")
)

type HookEventType string

// Types of hook events
const (
	HookEventCreate                    HookEventType = "create"
	HookEventDelete                    HookEventType = "delete"
	HookEventFork                      HookEventType = "fork"
	HookEventPush                      HookEventType = "push"
	HookEventIssues                    HookEventType = "issues"
	HookEventIssueAssign               HookEventType = "issue_assign"
	HookEventIssueLabel                HookEventType = "issue_label"
	HookEventIssueMilestone            HookEventType = "issue_milestone"
	HookEventIssueComment              HookEventType = "issue_comment"
	HookEventPullRequest               HookEventType = "pull_request"
	HookEventPullRequestAssign         HookEventType = "pull_request_assign"
	HookEventPullRequestLabel          HookEventType = "pull_request_label"
	HookEventPullRequestMilestone      HookEventType = "pull_request_milestone"
	HookEventPullRequestComment        HookEventType = "pull_request_comment"
	HookEventPullRequestReviewApproved HookEventType = "pull_request_review_approved"
	HookEventPullRequestReviewRejected HookEventType = "pull_request_review_rejected"
	HookEventPullRequestReviewComment  HookEventType = "pull_request_review_comment"
	HookEventPullRequestSync           HookEventType = "pull_request_sync"
	HookEventWiki                      HookEventType = "wiki"
	HookEventRepository                HookEventType = "repository"
	HookEventRelease                   HookEventType = "release"
	HookEventPackage                   HookEventType = "package"
)

// Option is a configuration option for the webhook
type Option func(*Webhook) error

// Options is a namespace var for configuration options
var Options = WebhookOptions{}

// WebhookOptions is a namespace for configuration option methods
type WebhookOptions struct{}

// Secret registers the GitLab secret
func (WebhookOptions) Secret(secret string) Option {
	return func(hook *Webhook) error {
		hook.secret = secret
		return nil
	}
}

// Webhook instance contains all methods needed to process events
type Webhook struct {
	secret string
}

// New creates and returns a WebHook instance denoted by the Provider type
func New(options ...Option) (*Webhook, error) {
	hook := new(Webhook)
	for _, opt := range options {
		if err := opt(hook); err != nil {
			return nil, errors.New("Error applying Option")
		}
	}
	return hook, nil
}

// Parse verifies and parses the events specified and returns the payload object or an error
func (hook Webhook) Parse(r *http.Request, events ...HookEventType) (interface{}, error) {
	defer func() {
		_, _ = io.Copy(io.Discard, r.Body)
		_ = r.Body.Close()
	}()

	if len(events) == 0 {
		return nil, ErrEventNotSpecifiedToParse
	}
	if r.Method != http.MethodPost {
		return nil, ErrInvalidHTTPMethod
	}

	event := r.Header.Get("X-Gitea-Event")
	if len(event) == 0 {
		return nil, ErrMissingGiteaEventHeader
	}

	giteaEvent := HookEventType(event)

	var found bool
	for _, evt := range events {
		if evt == giteaEvent {
			found = true
			break
		}
	}
	// event not defined to be parsed
	if !found {
		return nil, ErrEventNotFound
	}

	payload, err := io.ReadAll(r.Body)
	if err != nil || len(payload) == 0 {
		return nil, ErrParsingPayload
	}

	// If we have a Secret set, we should check the MAC
	if len(hook.secret) > 0 {
		signature := r.Header.Get("X-Gitea-Signature")
		if len(signature) == 0 {
			return nil, ErrMissingGiteaSignatureHeader
		}
		sig256 := hmac.New(sha256.New, []byte(hook.secret))
		_, _ = io.Writer(sig256).Write([]byte(payload))
		expectedMAC := hex.EncodeToString(sig256.Sum(nil))

		if !hmac.Equal([]byte(signature), []byte(expectedMAC)) {
			return nil, ErrHMACVerificationFailed
		}
	}
	// https://github.com/go-gitea/gitea/blob/main/services/webhook/payloader.go
	// https://github.com/go-gitea/gitea/blob/33fca2b537d36cf998dd27425b2bb8ed5b0965f3/services/webhook/payloader.go#L27
	switch giteaEvent {
	case HookEventCreate:
		var pl api.CreatePayload
		err = json.Unmarshal([]byte(payload), &pl)
		return pl, err
	case HookEventDelete:
		var pl api.DeletePayload
		err = json.Unmarshal([]byte(payload), &pl)
		return pl, err
	case HookEventFork:
		var pl api.ForkPayload
		err = json.Unmarshal([]byte(payload), &pl)
		return pl, err
	case HookEventPush:
		var pl api.PushPayload
		err = json.Unmarshal([]byte(payload), &pl)
		return pl, err
	case HookEventIssues, HookEventIssueAssign, HookEventIssueLabel, HookEventIssueMilestone:
		var pl api.IssuePayload
		err = json.Unmarshal([]byte(payload), &pl)
		return pl, err
	case HookEventIssueComment:
		var pl api.IssueCommentPayload
		err = json.Unmarshal([]byte(payload), &pl)
		return pl, err
	case HookEventPullRequest, HookEventPullRequestAssign, HookEventPullRequestLabel,
		HookEventPullRequestMilestone, HookEventPullRequestSync, HookEventPullRequestReviewApproved,
		HookEventPullRequestReviewRejected, HookEventPullRequestReviewComment, HookEventPullRequestComment:
		var pl api.PullRequestPayload
		err = json.Unmarshal([]byte(payload), &pl)
		return pl, err
	case HookEventRepository:
		var pl api.RepositoryPayload
		err = json.Unmarshal([]byte(payload), &pl)
		return pl, err
	case HookEventRelease:
		var pl api.ReleasePayload
		err = json.Unmarshal([]byte(payload), &pl)
		return pl, err
	case HookEventWiki:
		var pl api.WikiPayload
		err = json.Unmarshal([]byte(payload), &pl)
		return pl, err
	default:
		return nil, fmt.Errorf("unknown event %s", giteaEvent)
	}
}
