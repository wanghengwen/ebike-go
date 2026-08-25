package es

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	"go.uber.org/zap"
)

// deleteByQueryResponse mirrors the relevant parts of ES delete_by_query JSON response.
type deleteByQueryResponse struct {
	Deleted int64 `json:"deleted"`
}

// DeleteByQuery deletes documents matching the query.
// Returns the number of deleted documents.
func DeleteByQuery(ctx context.Context, indices []string, query M) (int64, error) {
	body := M{"query": query}

	reader, err := encodeBody(body)
	if err != nil {
		return 0, err
	}

	opts := []func(*esapi.DeleteByQueryRequest){
		Client.DeleteByQuery.WithContext(ctx),
	}

	res, err := Client.DeleteByQuery(indices, reader, opts...)
	if err != nil {
		return 0, fmt.Errorf("ES delete_by_query request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return 0, fmt.Errorf("ES delete_by_query error: %s", res.String())
	}

	var resp deleteByQueryResponse
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return 0, fmt.Errorf("failed to decode delete_by_query response: %w", err)
	}

	zap.L().Info("DeleteByQuery completed",
		zap.Strings("indices", indices),
		zap.Int64("deleted", resp.Deleted),
	)
	return resp.Deleted, nil
}
