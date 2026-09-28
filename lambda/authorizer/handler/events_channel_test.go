package handler

import (
	"strings"
	"testing"

	"github.com/pennsieve/pennsieve-go-core/pkg/realtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testUUID  = "12345678-1234-1234-1234-123456789abc"
	testUUID2 = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
)

func TestParseEventsChannel(t *testing.T) {
	cases := map[string]struct {
		namespace, channel string
		want               eventsChannel
	}{
		"dataset":     {"datasets", "/datasets/" + testUUID, eventsChannel{namespace: "datasets", id: testUUID}},
		"application": {"applications", "/applications/" + testUUID, eventsChannel{namespace: "applications", id: testUUID}},
		"org run":     {"runs", "/runs/org-" + testUUID + "/" + testUUID2, eventsChannel{namespace: "runs", id: testUUID2, scopeKind: "org", scopeId: testUUID}},
		"org runs":    {"runs", "/runs/org-" + testUUID + "/*", eventsChannel{namespace: "runs", id: "*", scopeKind: "org", scopeId: testUUID}},
		"user run":    {"runs", "/runs/user-" + testUUID + "/run-1", eventsChannel{namespace: "runs", id: "run-1", scopeKind: "user", scopeId: testUUID}},
		"user runs":   {"runs", "/runs/user-" + testUUID + "/*", eventsChannel{namespace: "runs", id: "*", scopeKind: "user", scopeId: testUUID}},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := parseEventsChannel(c.namespace, c.channel)
			require.NoError(t, err)
			assert.Equal(t, c.want, got)
		})
	}
}

func TestParseEventsChannelRejects(t *testing.T) {
	cases := map[string][2]string{
		"no leading slash":       {"datasets", "datasets/" + testUUID},
		"namespace mismatch":     {"runs", "/datasets/" + testUUID},
		"unknown namespace":      {"orgs", "/orgs/" + testUUID},
		"dataset wildcard":       {"datasets", "/datasets/*"},
		"dataset node id":        {"datasets", "/datasets/N:dataset:" + testUUID},
		"dataset not a uuid":     {"datasets", "/datasets/abc"},
		"dataset extra segment":  {"datasets", "/datasets/" + testUUID + "/x"},
		"dataset trailing slash": {"datasets", "/datasets/" + testUUID + "/"},
		"namespace only":         {"datasets", "/datasets"},
		"application wildcard":   {"applications", "/applications/*"},
		"runs wildcard":          {"runs", "/runs/*"},
		"runs scope wildcard":    {"runs", "/runs/*/*"},
		"runs unknown scope":     {"runs", "/runs/team-" + testUUID + "/*"},
		"runs scope not a uuid":  {"runs", "/runs/org-abc/*"},
		"runs no run segment":    {"runs", "/runs/org-" + testUUID},
		"runs too deep":          {"runs", "/runs/org-" + testUUID + "/" + testUUID2 + "/*"},
		"runs partial wildcard":  {"runs", "/runs/org-" + testUUID + "/abc*"},
		"runs long run id":       {"runs", "/runs/org-" + testUUID + "/" + strings.Repeat("a", 51)},
		"empty":                  {"", ""},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := parseEventsChannel(c[0], c[1])
			assert.Error(t, err)
		})
	}
}

// The channels go-core's publishers (and subscriber helpers) build must parse.
func TestParseEventsChannelAcceptsRealtimeChannels(t *testing.T) {
	cases := map[string]struct {
		ch   realtime.Channel
		want eventsChannel
	}{
		"dataset":   {realtime.Dataset("N:dataset:" + testUUID), eventsChannel{namespace: "datasets", id: testUUID}},
		"org run":   {realtime.Run("N:organization:"+testUUID, "", testUUID2), eventsChannel{namespace: "runs", id: testUUID2, scopeKind: "org", scopeId: testUUID}},
		"user run":  {realtime.Run("", "N:user:"+testUUID, testUUID2), eventsChannel{namespace: "runs", id: testUUID2, scopeKind: "user", scopeId: testUUID}},
		"org scope": {realtime.RunScope("N:organization:"+testUUID, ""), eventsChannel{namespace: "runs", id: "*", scopeKind: "org", scopeId: testUUID}},
		"app":       {realtime.Application(testUUID), eventsChannel{namespace: "applications", id: testUUID}},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := parseEventsChannel(c.ch.Namespace, c.ch.Path())
			require.NoError(t, err)
			assert.Equal(t, c.want, got)
		})
	}
}
