package deployment_changes

import (
	"slices"

	"github.com/ferretcode/locomotive/internal/railway/gql/queries"
)

func findSuccessfulDeploymentsIdsForWantedServiceIds(environment *queries.EnvironmentData, wantedServiceIds []string) []DeploymentIdWithInfo {
	successfulDeploymentsIdsForWantedServiceIds := []DeploymentIdWithInfo{}

	for _, deployment := range environment.Environment.Deployments.Edges {
		// Only consider successful deployments
		if deployment.Node.Status != "SUCCESS" {
			continue
		}

		// Only consider deployments for the specified trains
		if !slices.Contains(wantedServiceIds, deployment.Node.ServiceID) {
			continue
		}

		successfulDeploymentsIdsForWantedServiceIds = append(successfulDeploymentsIdsForWantedServiceIds, DeploymentIdWithInfo{
			ID:        deployment.Node.ID,
			CreatedAt: deployment.Node.CreatedAt,
		})
	}

	return successfulDeploymentsIdsForWantedServiceIds
}
