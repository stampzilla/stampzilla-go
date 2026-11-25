package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestScheduleDelay(t *testing.T) {
	regular := 50 * time.Minute
	assert.Equal(t, regular, scheduleDelay(0, regular))
	assert.Equal(t, 1*time.Minute, scheduleDelay(1, regular))
	assert.Equal(t, 2*time.Minute, scheduleDelay(2, regular))
	assert.Equal(t, 32*time.Minute, scheduleDelay(6, regular))
	assert.Equal(t, regular, scheduleDelay(7, regular))
	assert.Equal(t, regular, scheduleDelay(1000, regular))
	assert.Equal(t, 30*time.Second, scheduleDelay(1, 30*time.Second))
}

func TestGetWsURLErrors(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		permanent bool
		wantURL   string
	}{
		{"graphql error with null data", `{"errors":[{"message":"internal error"}],"data":null}`, false, ""},
		{"home missing", `{"data":{"viewer":{"home":null}}}`, false, ""},
		{"realtime disabled", `{"data":{"viewer":{"websocketSubscriptionUrl":"wss://x","home":{"features":{"realTimeConsumptionEnabled":false}}}}}`, true, ""},
		{"ok", `{"data":{"viewer":{"websocketSubscriptionUrl":"wss://x","home":{"features":{"realTimeConsumptionEnabled":true}}}}}`, false, "wss://x"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(tt.body))
			}))
			defer srv.Close()
			old := httpClient
			httpClient = &http.Client{Transport: rewriteTransport{srv.URL}}
			defer func() { httpClient = old }()

			u, err := getWsURL(context.Background(), "token", "home")
			if tt.wantURL != "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantURL, u)
				return
			}
			assert.Error(t, err)
			assert.Equal(t, tt.permanent, errors.Is(err, errRealTimeNotEnabled))
		})
	}
}

type rewriteTransport struct{ target string }

func (rt rewriteTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	r.URL.Scheme = "http"
	r.URL.Host = strings.TrimPrefix(rt.target, "http://")
	return http.DefaultTransport.RoundTrip(r)
}
