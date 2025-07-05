package subscriptions

import _ "embed"

//go:embed environment_logs.graphql
var EnvironmentLogsSubscription string

//go:embed canvas_invalidation.graphql
var CanvasInvalidationSubscription string

//go:embed http_logs.graphql
var HttpLogsSubscription string
