package http_logs

import (
	"context"
	"errors"
	"fmt"
	"time"

	cache "github.com/Code-Hex/go-generics-cache"

	"github.com/ferretcode/locomotive/internal/railway"
	"github.com/ferretcode/locomotive/internal/railway/gql/queries"
	"github.com/ferretcode/locomotive/internal/railway/gql/subscriptions"
)

var metadataDeploymentCache = cache.New[string, DeploymentHttpLogMetadata]()

func getMetadataForDeployment(ctx context.Context, g *railway.GraphQLClient, deploymentId string) (DeploymentHttpLogMetadata, error) {
	if cached, ok := metadataDeploymentCache.Get(deploymentId); ok {
		return cached, nil
	}

	if g.Client == nil {
		return DeploymentHttpLogMetadata{}, errors.New("client is nil")
	}

	deployment := &queries.Deployment{}

	variables := map[string]any{
		"id": deploymentId,
	}

	if err := g.Client.Exec(ctx, queries.DeploymentQuery, &deployment, variables); err != nil {
		return DeploymentHttpLogMetadata{}, err
	}

	metadata := DeploymentHttpLogMetadata{}

	metadata.ServiceName = deployment.Deployment.Service.Name
	metadata.ServiceID = deployment.Deployment.Service.ID

	metadata.EnvironmentName = deployment.Deployment.Environment.Name
	metadata.EnvironmentID = deployment.Deployment.Environment.ID

	metadata.ProjectName = deployment.Deployment.Service.Project.Name
	metadata.ProjectID = deployment.Deployment.Service.Project.ID

	metadata.DeploymentID = deploymentId

	metadataDeploymentCache.Set(deploymentId, metadata, cache.WithExpiration((10 * time.Minute)))

	return metadata, nil
}

// getTimeStampFromHttpLog is a helper function to get the timestamp from an HttpLog since we use the `any` type to keep it flexible
func getTimeStampAttributeFromHttpLog(h subscriptions.HttpLog) (time.Time, error) {
	log, ok := h.(map[string]any)
	if !ok {
		return time.Time{}, fmt.Errorf("HttpLog is not a map[string]any, got %T", h)
	}

	timestampStr, exists := log["timestamp"]
	if !exists {
		return time.Time{}, fmt.Errorf("timestamp field not found in HttpLog")
	}

	_, ok = timestampStr.(string)
	if !ok {
		return time.Time{}, fmt.Errorf("timestamp field is not a string, got %T", timestampStr)
	}

	t, err := time.Parse(time.RFC3339, timestampStr.(string))
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse timestamp '%s' as RFC3339 format: %w", timestampStr, err)
	}

	return t, nil
}

func getStringAttributeFromHttpLog(h subscriptions.HttpLog, attribute string) (string, error) {
	log, ok := h.(map[string]any)
	if !ok {
		return "", fmt.Errorf("HttpLog is not a map[string]any, got %T", h)
	}

	attributeValue, exists := log[attribute]
	if !exists {
		return "", fmt.Errorf("attribute %s not found in HttpLog", attribute)
	}

	_, ok = attributeValue.(string)
	if !ok {
		return "", fmt.Errorf("attribute %s is not a string, got %T", attribute, attributeValue)
	}

	return attributeValue.(string), nil
}
