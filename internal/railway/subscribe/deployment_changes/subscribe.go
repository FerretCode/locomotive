package deployment_changes

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"

	"github.com/ferretcode/locomotive/internal/railway/subscribe/environment_invalidation"

	"github.com/ferretcode/locomotive/internal/logger"
	"github.com/ferretcode/locomotive/internal/railway"
	"github.com/ferretcode/locomotive/internal/railway/gql/queries"
	"github.com/ferretcode/locomotive/internal/slice"
)

func SubscribeToDeploymentIdChanges(ctx context.Context, g *railway.GraphQLClient, deploymentIdSlice *slice.Sync[string], changeDetected chan<- struct{}, environmentId string, serviceIds []string) error {
	environment := &queries.EnvironmentData{}

	variables := map[string]any{
		"id": environmentId,
	}

	if err := g.Client.Exec(context.Background(), queries.EnvironmentQuery, &environment, variables); err != nil {
		return err
	}

	deploymentIdSlice.AppendMany(findSuccessfulDeploymentsIdsForWantedServiceIds(environment, serviceIds))

	changeDetected <- struct{}{}

	environmentHashTrack := make(chan string)
	errorChan := make(chan error, 1)

	go func() {
		if err := environment_invalidation.SubscribeToInvalidationRequests(ctx, g, environmentHashTrack, environmentId); err != nil {
			if errors.Is(err, context.Canceled) {
				errorChan <- ctx.Err()
				return
			}

			logger.Stderr.Error("error subscribing to invalidation requests", logger.ErrAttr(err))

			errorChan <- err
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errorChan:
			return err
		case <-environmentHashTrack:
			// logger.Stdout.Debug("invalidation request received", slog.String("hash", environmentHash))

			environment := &queries.EnvironmentData{}

			if err := g.Client.Exec(context.Background(), queries.EnvironmentQuery, &environment, variables); err != nil {
				return fmt.Errorf("error getting environment data for new environment hash: %w", err)
			}

			latestSuccessfulDeploymentIdsForWantedServiceIds := findSuccessfulDeploymentsIdsForWantedServiceIds(environment, serviceIds)

			if len(latestSuccessfulDeploymentIdsForWantedServiceIds) == 0 {
				// logger.Stdout.Debug("no new deployment id(s) for wanted service id(s)", slog.String("hash", environmentHash))
				continue
			}

			deploymentsChanged := false

			// Add new deployments that aren't currently tracked
			for _, deploymentId := range latestSuccessfulDeploymentIdsForWantedServiceIds {
				if !deploymentIdSlice.Contains(deploymentId) {
					deploymentIdSlice.Append(deploymentId)
					deploymentsChanged = true
				}
			}

			// Remove deployments that are no longer in the latest environment data
			for _, deploymentId := range deploymentIdSlice.Get() {
				if !slices.Contains(latestSuccessfulDeploymentIdsForWantedServiceIds, deploymentId) {
					deploymentIdSlice.Delete(deploymentId)
					deploymentsChanged = true
				}
			}

			if deploymentsChanged {
				logger.Stdout.Debug("deployment id(s) changed for wanted service id(s)", slog.Any("deployment_ids", deploymentIdSlice.Get()))
				changeDetected <- struct{}{}
			}
		}
	}
}
