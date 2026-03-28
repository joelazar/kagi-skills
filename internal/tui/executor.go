package tui

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/joelazar/kagi/internal/api"
	"github.com/joelazar/kagi/internal/config"
)

const (
	kagiSearchURL    = "https://kagi.com/api/v0/search"
	kagiFastGPTURL   = "https://kagi.com/api/v0/fastgpt"
	kagiSummarizeURL = "https://kagi.com/api/v0/summarize"
	kagiEnrichURL    = "https://kagi.com/api/v0/enrich/"
	defaultTimeout   = 30 * time.Second
)

// NewExecutor creates a CommandExecutor that calls Kagi APIs.
func NewExecutor(cfg *config.Config) CommandExecutor {
	return func(command string, inputs map[string]string) ([]ResultItem, error) {
		apiKey, err := api.ResolveAPIKey(cfg)
		if err != nil {
			return nil, err
		}
		client := api.NewHTTPClient(defaultTimeout)

		switch command {
		case "search":
			return executeSearch(client, apiKey, inputs)
		case "fastgpt":
			return executeFastGPT(client, apiKey, inputs)
		case "summarize":
			return executeSummarize(client, apiKey, inputs)
		case "enrich":
			return executeEnrich(client, apiKey, inputs)
		case "balance":
			return executeBalance(client, apiKey)
		default:
			return nil, fmt.Errorf("command %q not yet supported in TUI mode", command)
		}
	}
}

func executeSearch(client *http.Client, apiKey string, inputs map[string]string) ([]ResultItem, error) {
	query := inputs["query"]
	if query == "" {
		return nil, errors.New("query is required")
	}

	limit := 10
	if l, ok := inputs["limit"]; ok && l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	params := url.Values{}
	params.Set("q", query)
	params.Set("limit", strconv.Itoa(limit))

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, kagiSearchURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bot "+apiKey)
	req.Header.Set("Accept", "application/json")

	body, err := api.DoAPIRequest(client, req)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Data []struct {
			T       int    `json:"t"`
			Title   string `json:"title"`
			URL     string `json:"url"`
			Snippet string `json:"snippet"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	var items []ResultItem
	for _, d := range resp.Data {
		if d.T != 0 {
			continue
		}
		items = append(items, ResultItem{
			Title:       d.Title,
			URL:         d.URL,
			Description: d.Snippet,
			Detail:      fmt.Sprintf("# %s\n\n%s\n\n[Open](%s)", d.Title, d.Snippet, d.URL),
		})
	}
	return items, nil
}

func executeFastGPT(client *http.Client, apiKey string, inputs map[string]string) ([]ResultItem, error) {
	query := inputs["query"]
	if query == "" {
		return nil, errors.New("query is required")
	}

	reqBody, err := json.Marshal(map[string]any{
		"query":      query,
		"cache":      true,
		"web_search": true,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, kagiFastGPTURL, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bot "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	body, err := api.DoAPIRequest(client, req)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Data struct {
			Output     string `json:"output"`
			References []struct {
				Title   string `json:"title"`
				URL     string `json:"url"`
				Snippet string `json:"snippet"`
			} `json:"references"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	// Main answer as first result.
	items := []ResultItem{{
		Title:       "Answer",
		Description: trimToLen(resp.Data.Output, 100),
		Detail:      resp.Data.Output,
	}}

	// References as additional results.
	for _, ref := range resp.Data.References {
		items = append(items, ResultItem{
			Title:       ref.Title,
			URL:         ref.URL,
			Description: ref.Snippet,
			Detail:      fmt.Sprintf("# %s\n\n%s\n\n[Open](%s)", ref.Title, ref.Snippet, ref.URL),
		})
	}

	return items, nil
}

func executeSummarize(client *http.Client, apiKey string, inputs map[string]string) ([]ResultItem, error) {
	targetURL := inputs["url"]
	if targetURL == "" {
		return nil, errors.New("URL is required")
	}

	reqBody := map[string]any{"url": targetURL}
	if engine := inputs["engine"]; engine != "" {
		reqBody["engine"] = engine
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, kagiSummarizeURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bot "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	body, err := api.DoAPIRequest(client, req)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Data struct {
			Output string `json:"output"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return []ResultItem{{
		Title:       "Summary",
		URL:         targetURL,
		Description: trimToLen(resp.Data.Output, 100),
		Detail:      resp.Data.Output,
	}}, nil
}

func executeEnrich(client *http.Client, apiKey string, inputs map[string]string) ([]ResultItem, error) {
	query := inputs["query"]
	if query == "" {
		return nil, errors.New("query is required")
	}

	params := url.Values{}
	params.Set("q", query)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, kagiEnrichURL+"web?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bot "+apiKey)
	req.Header.Set("Accept", "application/json")

	body, err := api.DoAPIRequest(client, req)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Data []struct {
			T       int    `json:"t"`
			Title   string `json:"title"`
			URL     string `json:"url"`
			Snippet string `json:"snippet"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	var items []ResultItem
	for _, d := range resp.Data {
		if d.T != 0 {
			continue
		}
		items = append(items, ResultItem{
			Title:       d.Title,
			URL:         d.URL,
			Description: d.Snippet,
			Detail:      fmt.Sprintf("# %s\n\n%s\n\n[Open](%s)", d.Title, d.Snippet, d.URL),
		})
	}
	return items, nil
}

func executeBalance(client *http.Client, apiKey string) ([]ResultItem, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://kagi.com/api/v0/search?q=test&limit=1", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bot "+apiKey)
	req.Header.Set("Accept", "application/json")

	body, err := api.DoAPIRequest(client, req)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Meta struct {
			APIBalance *float64 `json:"api_balance"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	balanceStr := "unknown"
	if resp.Meta.APIBalance != nil {
		balanceStr = fmt.Sprintf("$%.4f", *resp.Meta.APIBalance)
	}

	detail := fmt.Sprintf("# API Balance\n\nCurrent balance: **%s**", balanceStr)
	parts := strings.SplitN(detail, "\n", 2)
	desc := balanceStr
	if len(parts) > 1 {
		desc = "Balance: " + balanceStr
	}

	return []ResultItem{{
		Title:       "API Balance",
		Description: desc,
		Detail:      detail,
	}}, nil
}
