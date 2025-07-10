package queries

import _ "embed"

//go:embed project.graphql
var ProjectQuery string

//go:embed environment.graphql
var EnvironmentQuery string

//go:embed deployment.graphql
var DeploymentQuery string
