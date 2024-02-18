package graphql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/ferretcode/locomotive/config"
)

func (g *GraphQLClient) buildMetadataMap(cfg *config.Config) (map[string]string, error) {
	if g.client == nil {
		return nil, errors.New("client is nil")
	}

	environment := &Environment{}

	variables := map[string]any{
		"id": cfg.EnvironmentId,
	}

	if err := g.client.Exec(context.Background(), environmentQuery, &environment, variables); err != nil {
		return nil, err
	}

	project := &Project{}

	variables = map[string]any{
		"id": environment.Environment.ProjectID,
	}

	if err := g.client.Exec(context.Background(), projectQuery, &project, variables); err != nil {
		return nil, err
	}

	idNameMap := make(map[string]string)

	for _, e := range project.Project.Environments.Edges {
		idNameMap[e.Node.ID] = e.Node.Name
	}

	for _, s := range project.Project.Services.Edges {
		idNameMap[s.Node.ID] = s.Node.Name
	}

	idNameMap[project.Project.ID] = project.Project.Name

	return idNameMap, nil
}

func (g *GraphQLClient) SubscribeToLogs(logTrack chan<- *EnvironmentLog, trackError chan<- error, cfg *config.Config) error {
	if g.subscriptionClient == nil {
		return errors.New("subscriptionClient is nil")
	}

	metadataMap, err := g.buildMetadataMap(cfg)
	if err != nil {
		return fmt.Errorf("error building metadata map: %w", err)
	}

	variables := map[string]any{
		"environmentId": cfg.EnvironmentId,
		"beforeDate":    time.Now().Format(time.RFC3339Nano),
		"beforeLimit":   0,
	}

	if _, err := g.subscriptionClient.Exec(streamEnvironmentLogsQuery, variables, func(message []byte, err error) error {
		if err != nil {
			trackError <- err
			return err
		}

		data := &SubscriptionLogResponse{}

		if err := json.Unmarshal(message, &data); err != nil {
			trackError <- err
			return err
		}

		if len(data.EnvironmentLogs) == 0 {
			return nil
		}

		for i := range data.EnvironmentLogs {
			if len(data.EnvironmentLogs[i].MessageRaw) == 0 {
				continue
			}

			if !slices.Contains(cfg.Train, data.EnvironmentLogs[i].Tags.ServiceId) {
				continue
			}

			if data.EnvironmentLogs[i].Severity, err = strconv.Unquote(
				unsafe.String(unsafe.SliceData(data.EnvironmentLogs[i].SeverityRaw), len(data.EnvironmentLogs[i].SeverityRaw)),
			); err != nil {
				trackError <- err
				return err
			}

			if len(cfg.LogsFilter) > 0 && !slices.Contains(cfg.LogsFilter, "all") && !slices.Contains(cfg.LogsFilter, strings.ToLower(data.EnvironmentLogs[i].Severity)) {
				continue
			}

			if data.EnvironmentLogs[i].Message, err = strconv.Unquote(
				unsafe.String(unsafe.SliceData(data.EnvironmentLogs[i].MessageRaw), len(data.EnvironmentLogs[i].MessageRaw)),
			); err != nil {
				trackError <- err
				return err
			}

			serviceName, ok := metadataMap[data.EnvironmentLogs[i].Tags.ServiceId]
			if !ok {
				trackError <- fmt.Errorf("service name could not be found")
				serviceName = "undefined"
			}

			environmentName, ok := metadataMap[cfg.EnvironmentId]
			if !ok {
				trackError <- fmt.Errorf("environment name could not be found")
				environmentName = "undefined"
			}

			projectName, ok := metadataMap[data.EnvironmentLogs[i].Tags.ProjectId]
			if !ok {
				trackError <- fmt.Errorf("project name could not be found")
				projectName = "undefined"
			}

			data.EnvironmentLogs[i].Metadata = &Metadata{
				ProjectId:   data.EnvironmentLogs[i].Tags.ProjectId,
				ProjectName: projectName,

				EnvironmentId:   cfg.EnvironmentId,
				EnvironmentName: environmentName,

				ServiceId:   data.EnvironmentLogs[i].Tags.ServiceId,
				ServiceName: serviceName,
			}

			logTrack <- &data.EnvironmentLogs[i]
		}

		return nil
	}); err != nil {
		return err
	}

	defer g.subscriptionClient.Close()

	if err := g.subscriptionClient.Run(); err != nil {
		return err
	}

	return nil
}
