package es

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	"go.uber.org/zap"
)

// SearchResult holds parsed ES search response.
type SearchResult struct {
	TotalHits int64
	Hits      []Hit
}

// Hit represents a single search hit from ES.
type Hit struct {
	ID         string                 `json:"_id"`
	Source     map[string]interface{} `json:"_source"`
	SortValues []interface{}          `json:"sort"`
}

// encodeBody encodes the given map as JSON and returns an io.Reader.
func encodeBody(body M) (io.Reader, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, fmt.Errorf("failed to encode request body: %w", err)
	}
	return &buf, nil
}

// esSearchResponse mirrors the relevant parts of an ES search JSON response.
type esSearchResponse struct {
	Hits struct {
		Total struct {
			Value int64 `json:"value"`
		} `json:"total"`
		Hits []struct {
			ID     string                 `json:"_id"`
			Source map[string]interface{} `json:"_source"`
			Sort   []interface{}          `json:"sort"`
		} `json:"hits"`
	} `json:"hits"`
}

// parseSearchResponse reads and parses the ES search response body.
func parseSearchResponse(body io.ReadCloser) (*SearchResult, error) {
	defer body.Close()

	var raw esSearchResponse
	if err := json.NewDecoder(body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("failed to decode ES response: %w", err)
	}

	result := &SearchResult{
		TotalHits: raw.Hits.Total.Value,
		Hits:      make([]Hit, 0, len(raw.Hits.Hits)),
	}
	for _, h := range raw.Hits.Hits {
		result.Hits = append(result.Hits, Hit{
			ID:         h.ID,
			Source:     h.Source,
			SortValues: h.Sort,
		})
	}
	return result, nil
}

// SearchCount returns the total number of documents matching the query.
// Aligns with Java ElasticsearchServiceImpl.searchCount
func SearchCount(ctx context.Context, indices []string, query M, trackTotalHitsUpTo int) (int64, error) {
	body := M{
		"query": query,
		"size":  1,
	}
	if trackTotalHitsUpTo > 0 {
		body["track_total_hits"] = trackTotalHitsUpTo
	}

	reader, err := encodeBody(body)
	if err != nil {
		return 0, err
	}

	opts := []func(*esapi.SearchRequest){
		Client.Search.WithContext(ctx),
		Client.Search.WithIndex(indices...),
		Client.Search.WithBody(reader),
	}

	res, err := Client.Search(opts...)
	if err != nil {
		return 0, fmt.Errorf("ES search request failed: %w", err)
	}
	if res.IsError() {
		defer res.Body.Close()
		return 0, fmt.Errorf("ES search error: %s", res.String())
	}

	result, err := parseSearchResponse(res.Body)
	if err != nil {
		return 0, err
	}

	zap.L().Debug("SearchCount completed",
		zap.Strings("indices", indices),
		zap.Int64("totalHits", result.TotalHits),
	)
	return result.TotalHits, nil
}

// SearchLocation retrieves location data using search_after pagination.
// Each batch returns up to 5000 hits, sorted by _id DESC.
// Aligns with Java ElasticsearchServiceImpl.searchLocation
func SearchLocation(ctx context.Context, indices []string, query M, searchAfter []interface{}) (*SearchResult, error) {
	body := M{
		"query":            query,
		"size":             5000,
		"track_total_hits": false,
		"sort":             []M{{"_id": M{"order": "desc"}}},
		"_source":          []string{"endLat", "endLng", "endParkingId"},
	}
	if searchAfter != nil {
		body["search_after"] = searchAfter
	}

	reader, err := encodeBody(body)
	if err != nil {
		return nil, err
	}

	opts := []func(*esapi.SearchRequest){
		Client.Search.WithContext(ctx),
		Client.Search.WithIndex(indices...),
		Client.Search.WithBody(reader),
	}

	res, err := Client.Search(opts...)
	if err != nil {
		return nil, fmt.Errorf("ES search request failed: %w", err)
	}
	if res.IsError() {
		defer res.Body.Close()
		return nil, fmt.Errorf("ES search error: %s", res.String())
	}

	result, err := parseSearchResponse(res.Body)
	if err != nil {
		return nil, err
	}

	zap.L().Debug("SearchLocation completed",
		zap.Strings("indices", indices),
		zap.Int("hitCount", len(result.Hits)),
	)
	return result, nil
}

// SearchAll performs a paginated search with custom sort.
// Aligns with feature/Bin201Decode-20240531: from = (pageNum-1)*pageSize.
func SearchAll(ctx context.Context, indices []string, query M, pageNum, pageSize int, sort []M, trackTotalHitsUpTo int) (*SearchResult, error) {
	body := buildSearchAllBody(query, pageNum, pageSize, sort, trackTotalHitsUpTo)
	from, _ := body["from"].(int)

	reader, err := encodeBody(body)
	if err != nil {
		return nil, err
	}

	opts := []func(*esapi.SearchRequest){
		Client.Search.WithContext(ctx),
		Client.Search.WithIndex(indices...),
		Client.Search.WithBody(reader),
	}

	res, err := Client.Search(opts...)
	if err != nil {
		return nil, fmt.Errorf("ES search request failed: %w", err)
	}
	if res.IsError() {
		defer res.Body.Close()
		return nil, fmt.Errorf("ES search error: %s", res.String())
	}

	result, err := parseSearchResponse(res.Body)
	if err != nil {
		return nil, err
	}

	zap.L().Debug("SearchAll completed",
		zap.Strings("indices", indices),
		zap.Int64("totalHits", result.TotalHits),
		zap.Int("hitCount", len(result.Hits)),
		zap.Int("from", from),
		zap.Int("size", pageSize),
	)
	return result, nil
}

func buildSearchAllBody(query M, pageNum, pageSize int, sort []M, trackTotalHitsUpTo int) M {
	from := 0
	if pageNum > 0 && pageSize > 0 {
		from = (pageNum - 1) * pageSize
	}
	body := M{
		"query": query,
		"from":  from,
		"size":  pageSize,
	}
	if len(sort) > 0 {
		body["sort"] = sort
	}
	if trackTotalHitsUpTo > 0 {
		body["track_total_hits"] = trackTotalHitsUpTo
	}
	return body
}
