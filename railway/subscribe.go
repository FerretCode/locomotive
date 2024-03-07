package railway

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/ferretcode/locomotive/config"
	"github.com/ferretcode/locomotive/logger"
	"github.com/ferretcode/locomotive/util"
	"github.com/google/uuid"
	"nhooyr.io/websocket"
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

type operationMessage struct {
	Id      string  `json:"id"`
	Type    string  `json:"type"`
	Payload payload `json:"payload"`
}

type payload struct {
	Query     string     `json:"query"`
	Variables *variables `json:"variables"`
}

type variables struct {
	EnvironmentId string `json:"environmentId"`
	Filter        string `json:"filter"`
	BeforeLimit   int64  `json:"beforeLimit"`
	BeforeDate    string `json:"beforeDate"`
}

var (
	connectionInit = []byte(`{"type":"connection_init"}`)
	connectionAck  = []byte(`{"type":"connection_ack"}`)
)

func (g *GraphQLClient) createSubscription(ctx context.Context, cfg *config.Config) (*websocket.Conn, error) {
	payload := &payload{
		Query: streamEnvironmentLogsQuery,
		Variables: &variables{
			EnvironmentId: cfg.EnvironmentId,
			Filter:        buildServiceFilter(cfg.Train),

			// needed for seamless subscription resuming
			BeforeDate:  time.Now().UTC().Add(-5 * time.Minute).Format(time.RFC3339Nano),
			BeforeLimit: 500,
		},
	}

	subPayload := operationMessage{
		Id:      uuid.Must(uuid.NewUUID()).String(),
		Type:    "subscribe",
		Payload: *payload,
	}

	payloadBytes, err := json.Marshal(&subPayload)
	if err != nil {
		return nil, err
	}

	opts := &websocket.DialOptions{
		HTTPHeader: http.Header{
			"Authorization": []string{"Bearer " + g.AuthToken},
			"Content-Type":  []string{"application/json"},
		},
		Subprotocols: []string{"graphql-transport-ws"},
	}

	ctxTimeout, cancel := context.WithTimeout(context.Background(), (10 * time.Second))
	defer cancel()

	c, _, err := websocket.Dial(ctxTimeout, g.BaseSubscriptionURL, opts)
	if err != nil {
		return nil, err
	}

	c.SetReadLimit(-1)

	if err := c.Write(ctx, websocket.MessageText, connectionInit); err != nil {
		return nil, err
	}

	_, ackMessage, err := c.Read(ctx)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(ackMessage, connectionAck) {
		return nil, errors.New("did not receive connection ack from server")
	}

	if err := c.Write(ctx, websocket.MessageText, payloadBytes); err != nil {
		return nil, err
	}

	return c, nil
}

func (g *GraphQLClient) SubscribeToLogs(logTrack chan<- []EnvironmentLog, trackError chan<- error, cfg *config.Config) error {
	metadataMap, err := g.buildMetadataMap(cfg)
	if err != nil {
		return fmt.Errorf("error building metadata map: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn, err := g.createSubscription(ctx, cfg)
	if err != nil {
		return err
	}

	defer conn.CloseNow()

	LogTime := time.Now().UTC()

	var errAccumulation int

	for {
		_, logPayload, err := safeConnRead(conn, ctx)
		if err != nil {
			if errAccumulation > cfg.MaxErrAccumulations {
				return err
			}

			safeConnCloseNow(conn)

			logger.Stdout.Debug("resubscribing", slog.Any("reason", err))

			conn, err = g.createSubscription(ctx, cfg)
			if err != nil {
				errAccumulation++

				if errAccumulation > cfg.MaxErrAccumulations {
					return err
				}

				continue
			}

			continue
		}

		errAccumulation = 0

		logs := &LogPayloadResponse{}

		if err := json.Unmarshal(logPayload, &logs); err != nil {
			trackError <- err
			continue
		}

		filteredLogs := []EnvironmentLog{}

		for i := range logs.Payload.Data.EnvironmentLogs {
			// skip logs with empty messages
			if logs.Payload.Data.EnvironmentLogs[i].Message == "" {
				logger.Stdout.Debug("skipping blank log message")
				continue
			}

			// skip build logs, build logs don't have deployment ids
			if logs.Payload.Data.EnvironmentLogs[i].Tags.DeploymentID == "" {
				logger.Stdout.Debug("skipping build log message")
				continue
			}

			// on first subscription skip logs if they where logged before the first subscription, on resubscription skip logs if they where already processed
			if logs.Payload.Data.EnvironmentLogs[i].Timestamp.Before(LogTime) || LogTime == logs.Payload.Data.EnvironmentLogs[i].Timestamp {
				continue
			}

			// skip logs that don't match our desired global filter(s)
			if !util.IsWantedLevel(cfg.LogsFilterGlobal, logs.Payload.Data.EnvironmentLogs[i].Severity) {
				logger.Stdout.Debug("skipping undesired global log level", slog.String("level", logs.Payload.Data.EnvironmentLogs[i].Severity), slog.Any("wanted", cfg.LogsFilterGlobal))
				continue
			}

			LogTime = logs.Payload.Data.EnvironmentLogs[i].Timestamp

			serviceName, ok := metadataMap[logs.Payload.Data.EnvironmentLogs[i].Tags.ServiceID]
			if !ok {
				logger.Stdout.Warn("service name could not be found")
				serviceName = "undefined"
			}

			logs.Payload.Data.EnvironmentLogs[i].Tags.ServiceName = serviceName

			environmentName, ok := metadataMap[logs.Payload.Data.EnvironmentLogs[i].Tags.EnvironmentID]
			if !ok {
				logger.Stdout.Warn("environment name could not be found")
				environmentName = "undefined"
			}

			logs.Payload.Data.EnvironmentLogs[i].Tags.EnvironmentName = environmentName

			projectName, ok := metadataMap[logs.Payload.Data.EnvironmentLogs[i].Tags.ProjectID]
			if !ok {
				logger.Stdout.Warn("project name could not be found")
				projectName = "undefined"
			}

			logs.Payload.Data.EnvironmentLogs[i].Tags.ProjectName = projectName

			filteredLogs = append(filteredLogs, logs.Payload.Data.EnvironmentLogs[i])
		}

		if len(filteredLogs) == 0 {
			continue
		}

		logTrack <- filteredLogs
	}
}

// helper function to build a service filter string from provided service ids
func buildServiceFilter(serviceIds []string) string {
	var filterString string

	for i, serviceId := range serviceIds {
		filterString += "@service:" + serviceId
		if i < len(serviceIds)-1 {
			filterString += " OR "
		}
	}

	return filterString
}

// Railway tends to close the connection abruptly, this is needed to prevent any panics caused by reading from an abruptly closed connection
func safeConnRead(conn *websocket.Conn, ctx context.Context) (mT websocket.MessageType, b []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered from panic: %v", r)
		}
	}()

	return conn.Read(ctx)
}

func safeConnCloseNow(conn *websocket.Conn) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered from panic: %v", r)
		}
	}()

	return conn.CloseNow()
}
