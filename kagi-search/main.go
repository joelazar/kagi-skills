package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	readability "codeberg.org/readeck/go-readability/v2"
)

var version = "dev" // injected via -ldflags "-X main.version=..."

const (
	kagiSearchURL    = "https://kagi.com/api/v0/search"
	defaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

type apiMeta struct {
	ID         string   `json:"id,omitempty"`
	Node       string   `json:"node,omitempty"`
	MS         int      `json:"ms,omitempty"`
	APIBalance *float64 `json:"api_balance,omitempty"`
}

type apiThumbnail struct {
	URL    string `json:"url,omitempty"`
	Width  *int   `json:"width,omitempty"`
	Height *int   `json:"height,omitempty"`
}

type apiItem struct {
	T         int           `json:"t"`
	URL       string        `json:"url,omitempty"`
	Title     string        `json:"title,omitempty"`
	Snippet   string        `json:"snippet,omitempty"`
	Published string        `json:"published,omitempty"`
	Thumbnail *apiThumbnail `json:"thumbnail,omitempty"`
	List      []string      `json:"list,omitempty"`
}

type kagiSearchResponse struct {
	Meta apiMeta   `json:"meta"`
	Data []apiItem `json:"data"`
}

type searchResult struct {
	Title        string        `json:"title"`
	Link         string        `json:"link"`
	Snippet      string        `json:"snippet"`
	Published    string        `json:"published,omitempty"`
	Thumbnail    *apiThumbnail `json:"thumbnail,omitempty"`
	Content      string        `json:"content,omitempty"`
	ContentError string        `json:"content_error,omitempty"`
}

type searchOutput struct {
	Query           string         `json:"query"`
	Meta            apiMeta        `json:"meta"`
	Results         []searchResult `json:"results"`
	RelatedSearches []string       `json:"related_searches,omitempty"`
}

type contentOutput struct {
	URL     string `json:"url"`
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
	Error   string `json:"error,omitempty"`
}

var (
	reComments = regexp.MustCompile(`(?is)<!--.*?-->`)
	reNoise    = regexp.MustCompile(`(?is)<(?:script|style|noscript|svg|iframe|nav|header|footer|aside)[^>]*>.*?</(?:script|style|noscript|svg|iframe|nav|header|footer|aside)>`)
	reBlocks   = regexp.MustCompile(`(?is)</?(p|div|section|article|main|h[1-6]|li|ul|ol|blockquote|pre|tr|table|hr|br)[^>]*>`)
	reTags     = regexp.MustCompile(`(?is)<[^>]+>`)
	reMultiNL  = regexp.MustCompile(`\n{3,}`)
	reTitle    = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		printGeneralUsage()
		os.Exit(1)
	}

	var err error
	switch args[0] {
	case "--version", "-v":
		fmt.Printf("kagi-search %s\n", version)
	case "search":
		err = runSearch(args[1:])
	case "content":
		err = runContent(args[1:])
	default:
		// Convenience: allow calling binary directly without subcommand.
		err = runSearch(args)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func printGeneralUsage() {
	fmt.Println("Usage:")
	fmt.Println("  kagi-search search <query> [-n <num>] [--content] [--json]")
	fmt.Println("  kagi-search content <url> [--json]")
}

func runSearch(args []string) error {
	limit := 10
	fetchContent := false
	jsonOut := false
	timeoutSec := 15
	maxContentChars := 5000

	queryParts := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-h", "--help":
			printSearchUsage()
			return nil
		case "--":
			queryParts = append(queryParts, args[i+1:]...)
			i = len(args)
		case "-n":
			if i+1 >= len(args) {
				printSearchUsage()
				return errors.New("missing value for -n")
			}
			i++
			n, err := strconv.Atoi(args[i])
			if err != nil {
				printSearchUsage()
				return fmt.Errorf("invalid value for -n: %s", args[i])
			}
			limit = n
		case "--content":
			fetchContent = true
		case "--json":
			jsonOut = true
		case "--timeout":
			if i+1 >= len(args) {
				printSearchUsage()
				return errors.New("missing value for --timeout")
			}
			i++
			n, err := strconv.Atoi(args[i])
			if err != nil {
				printSearchUsage()
				return fmt.Errorf("invalid value for --timeout: %s", args[i])
			}
			timeoutSec = n
		case "--max-content-chars":
			if i+1 >= len(args) {
				printSearchUsage()
				return errors.New("missing value for --max-content-chars")
			}
			i++
			n, err := strconv.Atoi(args[i])
			if err != nil {
				printSearchUsage()
				return fmt.Errorf("invalid value for --max-content-chars: %s", args[i])
			}
			maxContentChars = n
		default:
			if strings.HasPrefix(arg, "-") {
				printSearchUsage()
				return fmt.Errorf("unknown option: %s", arg)
			}
			queryParts = append(queryParts, arg)
		}
	}

	query := strings.TrimSpace(strings.Join(queryParts, " "))
	if query == "" {
		printSearchUsage()
		return errors.New("query is required")
	}

	apiKey := strings.TrimSpace(os.Getenv("KAGI_API_KEY"))
	if apiKey == "" {
		return errors.New("KAGI_API_KEY environment variable is required (https://kagi.com/settings/api)")
	}

	if limit < 1 {
		limit = 1
	}
	if limit > 100 {
		limit = 100
	}
	if timeoutSec < 1 {
		timeoutSec = 1
	}
	if maxContentChars < 0 {
		maxContentChars = 0
	}

	client := newHTTPClient(time.Duration(timeoutSec) * time.Second)
	resp, err := fetchSearch(client, apiKey, query, limit)
	if err != nil {
		return err
	}

	out := searchOutput{
		Query:   query,
		Meta:    resp.Meta,
		Results: make([]searchResult, 0, len(resp.Data)),
	}

	for _, item := range resp.Data {
		switch item.T {
		case 0:
			out.Results = append(out.Results, searchResult{
				Title:     item.Title,
				Link:      item.URL,
				Snippet:   item.Snippet,
				Published: item.Published,
				Thumbnail: item.Thumbnail,
			})
		case 1:
			out.RelatedSearches = append(out.RelatedSearches, item.List...)
		}
	}

	if fetchContent {
		contentClient := newSafeContentClient(client.Timeout)
		for i := range out.Results {
			title, content, fetchErr := fetchPageContent(contentClient, out.Results[i].Link, maxContentChars)
			if out.Results[i].Title == "" && title != "" {
				out.Results[i].Title = title
			}
			if fetchErr != nil {
				out.Results[i].ContentError = fetchErr.Error()
				continue
			}
			out.Results[i].Content = content
		}
	}

	if jsonOut {
		return writeJSON(out)
	}

	if len(out.Results) == 0 {
		fmt.Fprintln(os.Stderr, "No results found.")
		if out.Meta.APIBalance != nil {
			fmt.Fprintf(os.Stderr, "[API Balance: $%.2f]\n", *out.Meta.APIBalance)
		}
		return nil
	}

	for i, r := range out.Results {
		fmt.Printf("--- Result %d ---\n", i+1)
		fmt.Printf("Title: %s\n", r.Title)
		fmt.Printf("Link: %s\n", r.Link)
		if r.Published != "" {
			fmt.Printf("Published: %s\n", r.Published)
		}
		fmt.Printf("Snippet: %s\n", r.Snippet)
		if fetchContent {
			if r.Content != "" {
				fmt.Printf("Content:\n%s\n", r.Content)
			} else if r.ContentError != "" {
				fmt.Printf("Content: (Error: %s)\n", r.ContentError)
			}
		}
		fmt.Println()
	}

	if len(out.RelatedSearches) > 0 {
		sorted := append([]string(nil), out.RelatedSearches...)
		sort.Strings(sorted)
		fmt.Println("--- Related Searches ---")
		for _, term := range sorted {
			fmt.Printf("- %s\n", term)
		}
	}

	if out.Meta.APIBalance != nil {
		fmt.Fprintf(os.Stderr, "[API Balance: $%.2f]\n", *out.Meta.APIBalance)
	}

	return nil
}

func runContent(args []string) error {
	jsonOut := false
	timeoutSec := 20
	maxChars := 20000

	positionals := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-h", "--help":
			printContentUsage()
			return nil
		case "--":
			positionals = append(positionals, args[i+1:]...)
			i = len(args)
		case "--json":
			jsonOut = true
		case "--timeout":
			if i+1 >= len(args) {
				printContentUsage()
				return errors.New("missing value for --timeout")
			}
			i++
			n, err := strconv.Atoi(args[i])
			if err != nil {
				printContentUsage()
				return fmt.Errorf("invalid value for --timeout: %s", args[i])
			}
			timeoutSec = n
		case "--max-chars":
			if i+1 >= len(args) {
				printContentUsage()
				return errors.New("missing value for --max-chars")
			}
			i++
			n, err := strconv.Atoi(args[i])
			if err != nil {
				printContentUsage()
				return fmt.Errorf("invalid value for --max-chars: %s", args[i])
			}
			maxChars = n
		default:
			if strings.HasPrefix(arg, "-") {
				printContentUsage()
				return fmt.Errorf("unknown option: %s", arg)
			}
			positionals = append(positionals, arg)
		}
	}

	if len(positionals) == 0 {
		printContentUsage()
		return errors.New("url is required")
	}
	if len(positionals) > 1 {
		printContentUsage()
		return errors.New("content accepts exactly one URL")
	}

	targetURL := strings.TrimSpace(positionals[0])
	parsedURL, err := validateRemoteFetchURL(targetURL)
	if err != nil {
		return err
	}
	targetURL = parsedURL.String()
	if timeoutSec < 1 {
		timeoutSec = 1
	}
	if maxChars < 0 {
		maxChars = 0
	}

	client := newSafeContentClient(time.Duration(timeoutSec) * time.Second)
	title, content, err := fetchPageContent(client, targetURL, maxChars)

	if jsonOut {
		out := contentOutput{
			URL:     targetURL,
			Title:   title,
			Content: content,
		}
		if err != nil {
			out.Error = err.Error()
		}
		return writeJSON(out)
	}

	if err != nil {
		return err
	}

	if title != "" {
		fmt.Printf("# %s\n\n", title)
	}
	fmt.Println(content)
	return nil
}

func printSearchUsage() {
	fmt.Println("Usage: kagi-search search <query> [-n <num>] [--content] [--json]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -n <num>              Number of results (default: 10, max: 100)")
	fmt.Println("  --content             Fetch readable page content")
	fmt.Println("  --json                Emit JSON output")
	fmt.Println("  --timeout <sec>       HTTP timeout in seconds (default: 15)")
	fmt.Println("  --max-content-chars   Max chars per fetched content (default: 5000)")
	fmt.Println()
	fmt.Println("Environment:")
	fmt.Println("  KAGI_API_KEY          Required. Your Kagi Search API key.")
}

func printContentUsage() {
	fmt.Println("Usage: kagi-search content <url> [--json]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --json                Emit JSON output")
	fmt.Println("  --timeout <sec>       HTTP timeout in seconds (default: 20)")
	fmt.Println("  --max-chars <num>     Max chars to output (default: 20000)")
}

func newHTTPClient(timeout time.Duration) *http.Client {
	t, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return &http.Client{Timeout: timeout}
	}
	transport := t.Clone()
	transport.Proxy = http.ProxyFromEnvironment
	transport.ForceAttemptHTTP2 = true
	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}

func newSafeContentClient(timeout time.Duration) *http.Client {
	var transport *http.Transport
	if base, ok := http.DefaultTransport.(*http.Transport); ok {
		transport = base.Clone()
	} else {
		transport = &http.Transport{}
	}
	// Security: ignore proxy env vars for untrusted URL fetches to avoid bypassing
	// local-IP protections through a forward proxy.
	transport.Proxy = nil
	transport.ForceAttemptHTTP2 = true

	dialer := &net.Dialer{Timeout: 15 * time.Second, KeepAlive: 30 * time.Second}
	transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}

		if ip := net.ParseIP(host); ip != nil {
			if isBlockedIP(ip) {
				return nil, fmt.Errorf("blocked private or local IP address: %s", ip)
			}
			return dialer.DialContext(ctx, network, address)
		}

		ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
		if err != nil {
			return nil, err
		}

		allowed := make([]string, 0, len(ips))
		for _, ipAddr := range ips {
			if !isBlockedIP(ipAddr.IP) {
				allowed = append(allowed, ipAddr.IP.String())
			}
		}
		if len(allowed) == 0 {
			return nil, fmt.Errorf("blocked host %q: resolves to private or local IP", host)
		}

		var lastErr error
		for _, ip := range allowed {
			conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip, port))
			if err == nil {
				return conn, nil
			}
			lastErr = err
		}
		if lastErr != nil {
			return nil, lastErr
		}
		return nil, fmt.Errorf("failed to dial host %q", host)
	}

	client := &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return errors.New("stopped after 10 redirects")
		}
		_, err := validateRemoteFetchURL(req.URL.String())
		return err
	}
	return client
}

func validateRemoteFetchURL(rawURL string) (*url.URL, error) {
	u, err := url.ParseRequestURI(strings.TrimSpace(rawURL))
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("invalid URL scheme %q (only http/https are allowed)", u.Scheme)
	}
	host := u.Hostname()
	if host == "" {
		return nil, errors.New("invalid URL: missing hostname")
	}
	if ip := net.ParseIP(host); ip != nil && isBlockedIP(ip) {
		return nil, fmt.Errorf("blocked private or local IP address: %s", ip)
	}
	return u, nil
}

func isBlockedIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if ip.IsLoopback() || ip.IsUnspecified() || ip.IsMulticast() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsPrivate() {
		return true
	}
	if ip4 := ip.To4(); ip4 != nil {
		// Shared/reserved IPv4 ranges that should never be fetched from untrusted input.
		if ip4[0] == 0 || ip4[0] >= 224 {
			return true
		}
		if ip4[0] == 100 && ip4[1] >= 64 && ip4[1] <= 127 {
			return true // 100.64.0.0/10
		}
	}
	return false
}

func fetchSearch(client *http.Client, apiKey, query string, limit int) (*kagiSearchResponse, error) {
	params := url.Values{}
	params.Set("q", query)
	params.Set("limit", strconv.Itoa(limit))

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, kagiSearchURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bot "+apiKey)
	req.Header.Set("User-Agent", defaultUserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		text := strings.TrimSpace(string(body))
		if len(text) > 500 {
			text = text[:500] + "..."
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, text)
	}

	var out kagiSearchResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func fetchPageContent(client *http.Client, targetURL string, maxChars int) (title string, content string, err error) {
	parsedURL, err := validateRemoteFetchURL(targetURL)
	if err != nil {
		return "", "", err
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", defaultUserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return "", "", err
	}

	htmlDoc := string(body)

	title, content = tryReadability(htmlDoc, parsedURL.String())
	if title == "" {
		title = extractTitle(htmlDoc)
	}
	if content == "" {
		content = extractReadableText(htmlDoc)
	}

	if strings.TrimSpace(content) == "" {
		return title, "", errors.New("could not extract readable content")
	}

	if maxChars > 0 {
		content = truncateRunes(content, maxChars)
	}
	return title, content, nil
}

// tryReadability attempts to extract title and content using the readability
// algorithm. Returns empty strings if parsing fails at any step.
func tryReadability(htmlDoc, targetURL string) (title, content string) {
	pageURL, err := url.Parse(targetURL)
	if err != nil {
		return
	}
	article, err := readability.FromReader(strings.NewReader(htmlDoc), pageURL)
	if err != nil {
		return
	}
	if t := cleanLine(article.Title()); t != "" {
		title = t
	}
	var sb strings.Builder
	if err := article.RenderText(&sb); err != nil {
		return
	}
	content = strings.TrimSpace(sb.String())
	return
}

func extractTitle(htmlDoc string) string {
	matches := reTitle.FindStringSubmatch(htmlDoc)
	if len(matches) < 2 {
		return ""
	}
	title := cleanLine(matches[1])
	return title
}

func extractReadableText(htmlDoc string) string {
	s := reComments.ReplaceAllString(htmlDoc, " ")
	s = reNoise.ReplaceAllString(s, "\n")
	s = reBlocks.ReplaceAllString(s, "\n")
	s = reTags.ReplaceAllString(s, " ")
	s = html.UnescapeString(s)
	s = strings.ReplaceAll(s, "\r", "")

	lines := strings.Split(s, "\n")
	cleaned := make([]string, 0, len(lines))
	for _, line := range lines {
		line = cleanLine(line)
		if line == "" {
			continue
		}
		cleaned = append(cleaned, line)
	}

	if len(cleaned) == 0 {
		return ""
	}

	joined := strings.Join(cleaned, "\n\n")
	joined = reMultiNL.ReplaceAllString(joined, "\n\n")
	return strings.TrimSpace(joined)
}

func cleanLine(s string) string {
	fields := strings.Fields(strings.TrimSpace(s))
	return strings.Join(fields, " ")
}

func truncateRunes(s string, limit int) string {
	if limit <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= limit {
		return s
	}
	return string(r[:limit])
}

func writeJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
