package handler

import (
	"fmt"
	"regexp"
	"strings"
)

// Channel scheme of the AppSync Event API. It mirrors pennsieve-go-core
// pkg/realtime (docs/realtime-appsync-design.md there):
//
//	/datasets/<datasetUuid>
//	/runs/org-<orgUuid>/<runUuid>    /runs/org-<orgUuid>/*
//	/runs/user-<userUuid>/<runUuid>  /runs/user-<userUuid>/*
//	/applications/<appUuid>
const (
	namespaceDatasets     = "datasets"
	namespaceRuns         = "runs"
	namespaceApplications = "applications"

	runScopeOrg  = "org"
	runScopeUser = "user"
	runWildcard  = "*"
)

var (
	uuidPattern  = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	runIdPattern = regexp.MustCompile(`^[A-Za-z0-9-]{1,50}$`)
)

// eventsChannel is a parsed subscription channel.
type eventsChannel struct {
	namespace string
	// id is the dataset or application UUID, or for runs the run ID or "*".
	id string
	// scopeKind and scopeId are set for runs: "org" or "user", and its UUID.
	scopeKind string
	scopeId   string
}

// parseEventsChannel parses a channel AppSync passed for EVENT_SUBSCRIBE.
// AppSync passes wildcard subscriptions through literally, so anything that
// isn't exactly one of the shapes above, including any other use of "*", is
// rejected.
func parseEventsChannel(namespace, channel string) (eventsChannel, error) {
	if !strings.HasPrefix(channel, "/") {
		return eventsChannel{}, fmt.Errorf("channel %q does not start with /", channel)
	}
	parts := strings.Split(channel[1:], "/")
	if parts[0] != namespace {
		return eventsChannel{}, fmt.Errorf("channel %q is not in namespace %q", channel, namespace)
	}
	segments := parts[1:]

	switch parts[0] {
	case namespaceDatasets, namespaceApplications:
		if len(segments) != 1 || !uuidPattern.MatchString(segments[0]) {
			return eventsChannel{}, fmt.Errorf("channel %q must be /%s/<uuid>", channel, parts[0])
		}
		return eventsChannel{namespace: parts[0], id: segments[0]}, nil

	case namespaceRuns:
		if len(segments) != 2 {
			return eventsChannel{}, fmt.Errorf("channel %q must be /runs/<scope>/<runId|*>", channel)
		}
		kind, id, _ := strings.Cut(segments[0], "-")
		if (kind != runScopeOrg && kind != runScopeUser) || !uuidPattern.MatchString(id) {
			return eventsChannel{}, fmt.Errorf("channel %q has an invalid run scope", channel)
		}
		if segments[1] != runWildcard && !runIdPattern.MatchString(segments[1]) {
			return eventsChannel{}, fmt.Errorf("channel %q has an invalid run ID", channel)
		}
		return eventsChannel{namespace: namespaceRuns, id: segments[1], scopeKind: kind, scopeId: id}, nil
	}
	return eventsChannel{}, fmt.Errorf("unknown namespace in channel %q", channel)
}
