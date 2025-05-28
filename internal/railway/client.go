package railway

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/Khan/genqlient/graphql"
	"github.com/buger/jsonparser"

	_ "github.com/Khan/genqlient/generate"
)

type authedTransport struct {
	token   string
	wrapped http.RoundTripper
}

type RailwayClient struct {
	graphql.Client
}

func (t *authedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.token)
	req.Header.Set("Content-Type", "application/json")

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	operationName, err := jsonparser.GetString(body, "operationName")
	if err != nil {
		return nil, err
	}

	if len(operationName) > 0 {
		operationName = strings.ToLower(string(operationName[0])) + operationName[1:]

		params := req.URL.Query()
		params.Set("q", operationName)
		req.URL.RawQuery = params.Encode()
	}

	req.Body = io.NopCloser(bytes.NewBuffer(body))

	return t.wrapped.RoundTrip(req)
}

func NewAuthedClient(token string) *RailwayClient {
	httpClient := http.Client{
		Transport: &authedTransport{
			token:   token,
			wrapped: http.DefaultTransport,
		},
	}

	return &RailwayClient{
		graphql.NewClient("https://backboard.railway.com/graphql/v2", &httpClient),
	}
}
