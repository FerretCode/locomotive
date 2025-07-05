package environment_logs

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	cache "github.com/Code-Hex/go-generics-cache"

	"github.com/ferretcode/locomotive/internal/railway/gql/queries"
	"github.com/ferretcode/locomotive/internal/railway/gql/subscriptions"
	"github.com/ferretcode/locomotive/internal/util"
	"github.com/hasura/go-graphql-client"
)

var metadataEnvironmentCache = cache.New[string, map[string]string]()

func getMetadataMapForEnvironment(ctx context.Context, g *graphql.Client, environmentId string) (map[string]string, error) {
	metadataMap, ok := metadataEnvironmentCache.Get(environmentId)
	if ok {
		return metadataMap, nil
	}

	if g == nil {
		return nil, errors.New("client is nil")
	}

	environment := &queries.EnvironmentData{}

	variables := map[string]any{
		"id": environmentId,
	}

	if err := g.Exec(ctx, queries.EnvironmentQuery, &environment, variables); err != nil {
		return nil, err
	}

	project := &queries.ProjectData{}

	variables = map[string]any{
		"id": environment.Environment.ProjectID,
	}

	if err := g.Exec(ctx, queries.ProjectQuery, &project, variables); err != nil {
		return nil, err
	}

	idToNameMap := make(map[string]string)

	for _, e := range project.Project.Environments.Edges {
		idToNameMap[e.Node.ID] = e.Node.Name
	}

	for _, s := range project.Project.Services.Edges {
		idToNameMap[s.Node.ID] = s.Node.Name
	}

	idToNameMap[project.Project.ID] = project.Project.Name

	metadataEnvironmentCache.Set(environmentId, idToNameMap, cache.WithExpiration((10 * time.Minute)))

	return idToNameMap, nil
}

// searches for the given key and returns the corresponding value (and true) if found, or an empty string (and false)
func AttributesHasKeys(attributes []subscriptions.EnvironmentLogAttributes, keys []string) (string, bool) {
	for i := range attributes {
		for j := range keys {
			if keys[j] == attributes[i].Key {
				return attributes[i].Value, true
			}
		}
	}

	return "", false
}

func FilterLogs(logs []EnvironmentLogWithMetadata, wantedLevel []string, contentFilter string) []EnvironmentLogWithMetadata {
	if len(wantedLevel) == 0 && contentFilter == "" {
		return logs
	}

	filteredLogs := []EnvironmentLogWithMetadata{}

	for i := range logs {
		if !util.IsWantedLevel(wantedLevel, logs[i].Log.Severity) {
			continue
		}

		// Convert log to JSON string for content filtering
		logJSON, _ := json.Marshal(logs[i])
		if !util.MatchesContentFilter(contentFilter, string(logJSON)) {
			continue
		}

		filteredLogs = append(filteredLogs, logs[i])
	}

	return filteredLogs
}
