package graphql

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
)

type GraphQLClient struct {
	BaseURL string
}

type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

func (g *GraphQLClient) DoQuery(query string, variables map[string]interface{}, to interface{}) error {
	graphQlRequest := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	body, err := json.Marshal(graphQlRequest)

	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		"POST",
		g.BaseURL,
		bytes.NewReader(body),
	)

	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+os.Getenv("RAILWAY_API_KEY"))

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)

	if err != nil {
		return err
	}

	err = json.Unmarshal(b, to)

	if err != nil {
		return err
	}

	return nil
}
