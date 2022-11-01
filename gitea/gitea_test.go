package gitea

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	path = "/webhooks"
)

var hook *Webhook

func TestMain(m *testing.M) {

	// setup
	var err error
	hook, err = New(Options.Secret("IsWishesWereHorsesWedAllBeEatingSteak!"))
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(m.Run())
	// teardown
}

func newServer(handler http.HandlerFunc) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc(path, handler)
	return httptest.NewServer(mux)
}

func TestBadRequests(t *testing.T) {
	assert := require.New(t)
	tests := []struct {
		name    string
		event   HookEventType
		payload io.Reader
		headers http.Header
	}{
		{
			name:    "BadNoEventHeader",
			event:   HookEventPush,
			payload: bytes.NewBuffer([]byte("{}")),
			headers: http.Header{},
		},
		{
			name:    "UnsubscribedEvent",
			event:   HookEventPush,
			payload: bytes.NewBuffer([]byte("{}")),
			headers: http.Header{
				"X-Gitea-Event": []string{"noneexistant_event"},
			},
		},
		{
			name:    "BadBody",
			event:   HookEventPush,
			payload: bytes.NewBuffer([]byte("")),
			headers: http.Header{
				"X-Gitea-Event": []string{"push"},
			},
		},
		{
			name:    "BadSignatureLength",
			event:   HookEventPush,
			payload: bytes.NewBuffer([]byte("{}")),
			headers: http.Header{
				"X-Gitea-Event":     []string{"push"},
				"X-Gitea-Signature": []string{""},
			},
		},
		{
			name:    "BadSignatureMatch",
			event:   HookEventPush,
			payload: bytes.NewBuffer([]byte("{}")),
			headers: http.Header{
				"X-Gitea-Event":     []string{"push"},
				"X-Gitea-Signature": []string{"111"},
			},
		},
	}

	for _, tt := range tests {
		tc := tt
		client := &http.Client{}
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var parseError error
			server := newServer(func(w http.ResponseWriter, r *http.Request) {
				_, parseError = hook.Parse(r, tc.event)
			})
			defer server.Close()
			req, err := http.NewRequest(http.MethodPost, server.URL+path, tc.payload)
			assert.NoError(err)
			req.Header = tc.headers
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			assert.NoError(err)
			assert.Equal(http.StatusOK, resp.StatusCode)
			assert.Error(parseError)
		})
	}
}
