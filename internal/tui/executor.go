package tui

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joelazar/kagi/internal/api"
	"github.com/joelazar/kagi/internal/config"
	"github.com/joelazar/kagi/internal/version"
	"gopkg.in/yaml.v3"
)

const (
	authCheckSearchURL = "https://kagi.com/api/v0/search"
	tuiCommandTimeout  = 2 * time.Minute
)

// NewExecutor creates a CommandExecutor backed by the current kagi CLI binary.
// @lat: [[architecture#Interactive TUI reuse]]
func NewExecutor(cfg *config.Config) CommandExecutor {
	cliPath, cliErr := os.Executable()

	return func(command string, inputs map[string]string) ([]ResultItem, error) {
		if cliErr != nil {
			return nil, cliErr
		}

		switch command {
		case "search":
			var out searchOutputJSON
			if err := runCLIJSON(cliPath, buildSearchArgs(inputs), &out); err != nil {
				return nil, err
			}
			return searchItems(out), nil
		case "search content":
			var out contentOutputJSON
			if err := runCLIJSON(cliPath, buildSearchContentArgs(inputs), &out); err != nil {
				return nil, err
			}
			return searchContentItems(out), nil
		case "fastgpt":
			var out fastGPTOutputJSON
			if err := runCLIJSON(cliPath, buildFastGPTArgs(inputs), &out); err != nil {
				return nil, err
			}
			return fastGPTItems(out), nil
		case "summarize":
			var out summarizeOutputJSON
			if err := runCLIJSON(cliPath, buildSummarizeArgs(inputs), &out); err != nil {
				return nil, err
			}
			return summarizeItems(out), nil
		case "enrich":
			var out enrichOutputJSON
			if err := runCLIJSON(cliPath, buildEnrichArgs(inputs), &out); err != nil {
				return nil, err
			}
			return enrichItems(out), nil
		case "quick":
			var out quickOutputJSON
			if err := runCLIJSON(cliPath, buildQuickArgs(inputs), &out); err != nil {
				return nil, err
			}
			return quickItems(out), nil
		case "translate":
			var out translateOutputJSON
			if err := runCLIJSON(cliPath, buildTranslateArgs(inputs), &out); err != nil {
				return nil, err
			}
			return translateItems(out), nil
		case "news":
			var out newsOutputJSON
			if err := runCLIJSON(cliPath, buildNewsArgs(inputs), &out); err != nil {
				return nil, err
			}
			return newsItems(out), nil
		case "smallweb":
			var out smallWebOutputJSON
			if err := runCLIJSON(cliPath, buildSmallWebArgs(inputs), &out); err != nil {
				return nil, err
			}
			return smallWebItems(out), nil
		case "askpage":
			var out askPageOutputJSON
			if err := runCLIJSON(cliPath, buildAskPageArgs(inputs), &out); err != nil {
				return nil, err
			}
			return askPageItems(out), nil
		case "assistant":
			var out assistantOutputJSON
			if err := runCLIJSON(cliPath, buildAssistantArgs(inputs), &out); err != nil {
				return nil, err
			}
			return assistantItems(out), nil
		case "assistant threads":
			return executeAssistantThreads(cliPath, inputs)
		case "assistant delete thread":
			return executeAssistantDeleteThread(cliPath, inputs)
		case "balance":
			var out api.BalanceCache
			if err := runCLIJSON(cliPath, []string{"balance"}, &out); err != nil {
				return nil, err
			}
			return balanceItems(out), nil
		case "auth":
			return executeAuthCheck(cfg)
		case "config":
			return executeConfig(cfg, inputs)
		case "completion":
			return executeCompletionHelp(), nil
		case "version":
			return executeVersion(), nil
		default:
			return nil, fmt.Errorf("command %q not yet supported in TUI mode", command)
		}
	}
}

type searchOutputJSON struct {
	Query   string `json:"query"`
	Results []struct {
		Title        string `json:"title"`
		Link         string `json:"link"`
		Snippet      string `json:"snippet"`
		Published    string `json:"published,omitempty"`
		Content      string `json:"content,omitempty"`
		ContentError string `json:"content_error,omitempty"`
	} `json:"results"`
	RelatedSearches []string `json:"related_searches,omitempty"`
}

type contentOutputJSON struct {
	URL     string `json:"url"`
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
	Error   string `json:"error,omitempty"`
}

type fastGPTOutputJSON struct {
	Query      string `json:"query"`
	Output     string `json:"output"`
	Tokens     int    `json:"tokens"`
	References []struct {
		Title   string `json:"title"`
		Snippet string `json:"snippet"`
		URL     string `json:"url"`
	} `json:"references,omitempty"`
}

type summarizeOutputJSON struct {
	Input  string `json:"input"`
	Output string `json:"output"`
	Tokens int    `json:"tokens"`
	Engine string `json:"engine,omitempty"`
	Type   string `json:"type,omitempty"`
}

type enrichOutputJSON struct {
	Query   string `json:"query"`
	Index   string `json:"index"`
	Results []struct {
		Rank      int    `json:"rank"`
		Title     string `json:"title"`
		URL       string `json:"url"`
		Snippet   string `json:"snippet,omitempty"`
		Published string `json:"published,omitempty"`
	} `json:"results"`
}

type quickOutputJSON struct {
	Query      string `json:"query"`
	Answer     string `json:"answer"`
	References string `json:"references,omitempty"`
}

type translateOutputJSON struct {
	Text        string `json:"text"`
	Translation string `json:"translation"`
	SourceLang  string `json:"source_lang,omitempty"`
	TargetLang  string `json:"target_lang"`
}

type newsOutputJSON struct {
	Category string `json:"category,omitempty"`
	Items    []struct {
		Title    string `json:"title"`
		Summary  string `json:"summary,omitempty"`
		Category string `json:"category,omitempty"`
		URL      string `json:"url,omitempty"`
		Domain   string `json:"domain,omitempty"`
	} `json:"items"`
	Count int `json:"count"`
}

type smallWebOutputJSON struct {
	Items []struct {
		Title     string `json:"title"`
		URL       string `json:"url"`
		Author    string `json:"author,omitempty"`
		Published string `json:"published,omitempty"`
		Summary   string `json:"summary,omitempty"`
	} `json:"items"`
	Count int `json:"count"`
}

type askPageOutputJSON struct {
	URL        string `json:"url"`
	Question   string `json:"question"`
	Answer     string `json:"answer"`
	ThreadID   string `json:"thread_id,omitempty"`
	References string `json:"references,omitempty"`
}

type assistantOutputJSON struct {
	Query      string `json:"query"`
	ThreadID   string `json:"thread_id"`
	Title      string `json:"title,omitempty"`
	Output     string `json:"output"`
	References string `json:"references,omitempty"`
}

type threadListOutputJSON struct {
	Threads []threadSummaryJSON `json:"threads"`
}

type threadSummaryJSON struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	CreatedAt string `json:"created_at,omitempty"`
	Saved     bool   `json:"saved"`
}

type threadDetailJSON struct {
	ID       string `json:"id"`
	Title    string `json:"title,omitempty"`
	Messages []struct {
		ID        string `json:"id"`
		State     string `json:"state"`
		Prompt    string `json:"prompt,omitempty"`
		Reply     string `json:"reply,omitempty"`
		CreatedAt string `json:"created_at,omitempty"`
	} `json:"messages"`
}

type cliResult struct {
	stdout string
	stderr string
	err    error
}

func runCLI(path string, args []string) cliResult {
	ctx, cancel := context.WithTimeout(context.Background(), tuiCommandTimeout)
	defer cancel()

	fullArgs := append([]string{"--no-tui"}, args...)
	cmd := exec.CommandContext(ctx, path, fullArgs...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		err = fmt.Errorf("timed out after %s", tuiCommandTimeout)
	}

	return cliResult{
		stdout: strings.TrimSpace(stdout.String()),
		stderr: strings.TrimSpace(stderr.String()),
		err:    err,
	}
}

func runCLIJSON(path string, args []string, target any) error {
	res := runCLI(path, append([]string{"--format", "json"}, args...))
	if res.err != nil {
		return cliError(res)
	}
	if res.stdout == "" {
		return errors.New("command returned no JSON output")
	}
	if err := json.Unmarshal([]byte(res.stdout), target); err != nil {
		return fmt.Errorf("parse command JSON: %w", err)
	}
	return nil
}

func cliError(res cliResult) error {
	parts := make([]string, 0, 2)
	if res.stderr != "" {
		parts = append(parts, res.stderr)
	}
	if res.stdout != "" {
		parts = append(parts, res.stdout)
	}
	if len(parts) == 0 {
		if res.err != nil {
			return res.err
		}
		return errors.New("command failed")
	}
	return errors.New(strings.Join(parts, "\n"))
}

func buildSearchArgs(inputs map[string]string) []string {
	args := []string{"search", requiredInput(inputs, "query")}
	if limit := cleanIntString(inputs["limit"]); limit != "" {
		args = append(args, "-n", limit)
	}
	if parseBoolInput(inputs["include_content"]) {
		args = append(args, "--content")
	}
	return args
}

func buildSearchContentArgs(inputs map[string]string) []string {
	return []string{"search", "content", requiredInput(inputs, "url")}
}

func buildFastGPTArgs(inputs map[string]string) []string {
	return []string{"fastgpt", requiredInput(inputs, "query")}
}

func buildSummarizeArgs(inputs map[string]string) []string {
	input := requiredInput(inputs, "input")
	mode := strings.ToLower(strings.TrimSpace(inputs["input_mode"]))
	args := []string{"summarize"}
	if mode == "text" {
		args = append(args, "--text", input)
	} else {
		args = append(args, input)
	}
	if engine := strings.TrimSpace(inputs["engine"]); engine != "" {
		args = append(args, "--engine", engine)
	}
	if summaryType := strings.TrimSpace(inputs["summary_type"]); summaryType != "" {
		args = append(args, "--type", summaryType)
	}
	if lang := strings.TrimSpace(inputs["lang"]); lang != "" {
		args = append(args, "--lang", lang)
	}
	return args
}

func buildEnrichArgs(inputs map[string]string) []string {
	index := strings.ToLower(strings.TrimSpace(inputs["index"]))
	if index == "" {
		index = "web"
	}
	args := []string{"enrich", index, requiredInput(inputs, "query")}
	if limit := cleanIntString(inputs["limit"]); limit != "" {
		args = append(args, "-n", limit)
	}
	return args
}

func buildQuickArgs(inputs map[string]string) []string {
	return []string{"quick", requiredInput(inputs, "query")}
}

func buildTranslateArgs(inputs map[string]string) []string {
	args := []string{"translate", "--to", requiredInput(inputs, "target_lang")}
	if source := strings.TrimSpace(inputs["source_lang"]); source != "" && !strings.EqualFold(source, "auto") {
		args = append(args, "--from", source)
	}
	if formality := strings.TrimSpace(inputs["formality"]); formality != "" {
		args = append(args, "--formality", formality)
	}
	args = append(args, requiredInput(inputs, "text"))
	return args
}

func buildNewsArgs(inputs map[string]string) []string {
	args := []string{"news"}
	if category := strings.TrimSpace(inputs["category"]); category != "" {
		args = append(args, "--category", category)
	}
	if limit := cleanIntString(inputs["limit"]); limit != "" {
		args = append(args, "-n", limit)
	}
	if lang := strings.TrimSpace(inputs["lang"]); lang != "" {
		args = append(args, "--lang", lang)
	}
	return args
}

func buildSmallWebArgs(inputs map[string]string) []string {
	args := []string{"smallweb"}
	if limit := cleanIntString(inputs["limit"]); limit != "" {
		args = append(args, "-n", limit)
	}
	return args
}

func buildAskPageArgs(inputs map[string]string) []string {
	return []string{"askpage", requiredInput(inputs, "url"), requiredInput(inputs, "query")}
}

func buildAssistantArgs(inputs map[string]string) []string {
	args := []string{"assistant", requiredInput(inputs, "query")}
	if threadID := strings.TrimSpace(inputs["thread_id"]); threadID != "" {
		args = append(args, "--thread", threadID)
	}
	return args
}

func requiredInput(inputs map[string]string, key string) string {
	return strings.TrimSpace(inputs[key])
}

func parseBoolInput(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "y", "yes", "true", "on":
		return true
	default:
		return false
	}
}

func cleanIntString(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return ""
	}
	return strconv.Itoa(n)
}

func searchItems(out searchOutputJSON) []ResultItem {
	items := make([]ResultItem, 0, len(out.Results)+len(out.RelatedSearches))
	for _, result := range out.Results {
		detail := markdownDocument(result.Title, []string{
			"**URL:** " + result.Link,
			optionalLine("**Published:** ", result.Published),
			optionalSection("Snippet", result.Snippet),
			optionalSection("Content", result.Content),
			optionalSection("Content error", result.ContentError),
		})
		items = append(items, ResultItem{
			Title:       fallbackTitle(result.Title, result.Link),
			URL:         result.Link,
			Description: trimToLen(strings.TrimSpace(result.Snippet), 120),
			Detail:      detail,
		})
	}
	for _, related := range out.RelatedSearches {
		items = append(items, ResultItem{
			Title:       related,
			Description: "Related search suggestion",
			Detail:      markdownDocument(related, []string{"Use this query in `search` for a follow-up lookup."}),
		})
	}
	return items
}

func searchContentItems(out contentOutputJSON) []ResultItem {
	title := fallbackTitle(out.Title, out.URL)
	detailParts := []string{"**URL:** " + out.URL, optionalSection("Content", out.Content)}
	if out.Error != "" {
		detailParts = append(detailParts, optionalSection("Extraction error", out.Error))
	}
	return []ResultItem{{
		Title:       title,
		URL:         out.URL,
		Description: trimToLen(out.Content, 140),
		Detail:      markdownDocument(title, detailParts),
	}}
}

func fastGPTItems(out fastGPTOutputJSON) []ResultItem {
	answerDetail := markdownDocument("Answer", []string{
		out.Output,
		fmt.Sprintf("**Tokens:** %d", out.Tokens),
		optionalSection("References", referenceMarkdown(out.References)),
	})
	items := []ResultItem{{
		Title:       "Answer",
		Description: trimToLen(out.Output, 140),
		Detail:      answerDetail,
	}}
	for _, ref := range out.References {
		items = append(items, ResultItem{
			Title:       fallbackTitle(ref.Title, ref.URL),
			URL:         ref.URL,
			Description: trimToLen(ref.Snippet, 120),
			Detail:      markdownDocument(ref.Title, []string{"**URL:** " + ref.URL, optionalSection("Snippet", ref.Snippet)}),
		})
	}
	return items
}

func summarizeItems(out summarizeOutputJSON) []ResultItem {
	detailParts := []string{out.Output, fmt.Sprintf("**Tokens:** %d", out.Tokens)}
	if out.Engine != "" {
		detailParts = append(detailParts, fmt.Sprintf("**Engine:** %s", out.Engine))
	}
	if out.Type != "" {
		detailParts = append(detailParts, fmt.Sprintf("**Type:** %s", out.Type))
	}
	return []ResultItem{{
		Title:       "Summary",
		Description: trimToLen(out.Output, 140),
		Detail:      markdownDocument(out.Input, detailParts),
	}}
}

func enrichItems(out enrichOutputJSON) []ResultItem {
	items := make([]ResultItem, 0, len(out.Results))
	for _, result := range out.Results {
		items = append(items, ResultItem{
			Title:       fallbackTitle(result.Title, result.URL),
			URL:         result.URL,
			Description: trimToLen(result.Snippet, 120),
			Detail: markdownDocument(result.Title, []string{
				fmt.Sprintf("**Index:** %s", out.Index),
				fmt.Sprintf("**Rank:** %d", result.Rank),
				"**URL:** " + result.URL,
				optionalLine("**Published:** ", result.Published),
				optionalSection("Snippet", result.Snippet),
			}),
		})
	}
	return items
}

func quickItems(out quickOutputJSON) []ResultItem {
	answer := htmlToMarkdown(out.Answer)
	refs := htmlToMarkdown(out.References)
	return []ResultItem{{
		Title:       "Quick answer",
		Description: trimToLen(answer, 140),
		Detail:      markdownDocument(out.Query, []string{answer, optionalSection("References", refs)}),
	}}
}

func translateItems(out translateOutputJSON) []ResultItem {
	detailParts := []string{
		fmt.Sprintf("**Target language:** %s", out.TargetLang),
		optionalLine("**Detected source:** ", out.SourceLang),
		optionalSection("Original text", out.Text),
		optionalSection("Translation", out.Translation),
	}
	return []ResultItem{{
		Title:       "Translation",
		Description: trimToLen(out.Translation, 140),
		Detail:      markdownDocument("Translation", detailParts),
	}}
}

func newsItems(out newsOutputJSON) []ResultItem {
	items := make([]ResultItem, 0, len(out.Items))
	for _, item := range out.Items {
		items = append(items, ResultItem{
			Title:       item.Title,
			URL:         item.URL,
			Description: trimToLen(item.Summary, 120),
			Detail: markdownDocument(item.Title, []string{
				optionalLine("**Category:** ", item.Category),
				optionalLine("**Source:** ", item.Domain),
				optionalLine("**URL:** ", item.URL),
				optionalSection("Summary", item.Summary),
			}),
		})
	}
	return items
}

func smallWebItems(out smallWebOutputJSON) []ResultItem {
	items := make([]ResultItem, 0, len(out.Items))
	for _, item := range out.Items {
		items = append(items, ResultItem{
			Title:       item.Title,
			URL:         item.URL,
			Description: trimToLen(item.Summary, 120),
			Detail: markdownDocument(item.Title, []string{
				optionalLine("**Author:** ", item.Author),
				optionalLine("**Published:** ", item.Published),
				optionalLine("**URL:** ", item.URL),
				optionalSection("Summary", item.Summary),
			}),
		})
	}
	return items
}

func askPageItems(out askPageOutputJSON) []ResultItem {
	answer := htmlToMarkdown(out.Answer)
	refs := htmlToMarkdown(out.References)
	parts := []string{"**URL:** " + out.URL, answer, optionalSection("References", refs)}
	if out.ThreadID != "" {
		parts = append(parts, fmt.Sprintf("**Thread ID:** %s", out.ThreadID))
	}
	return []ResultItem{{
		Title:       "Page answer",
		URL:         out.URL,
		Description: trimToLen(answer, 140),
		Detail:      markdownDocument(out.Question, parts),
	}}
}

func assistantItems(out assistantOutputJSON) []ResultItem {
	answer := htmlToMarkdown(out.Output)
	refs := htmlToMarkdown(out.References)
	title := out.Title
	if title == "" {
		title = "Assistant response"
	}
	parts := []string{answer, optionalSection("References", refs)}
	if out.ThreadID != "" {
		parts = append(parts, fmt.Sprintf("**Thread ID:** %s", out.ThreadID))
	}
	return []ResultItem{{
		Title:       title,
		Description: trimToLen(answer, 140),
		Detail:      markdownDocument(out.Query, parts),
	}}
}

func executeAssistantThreads(cliPath string, inputs map[string]string) ([]ResultItem, error) {
	var out threadListOutputJSON
	if err := runCLIJSON(cliPath, []string{"assistant", "thread", "list"}, &out); err != nil {
		return nil, err
	}

	limit := 10
	if raw := cleanIntString(inputs["limit"]); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			limit = n
		}
	}
	if limit > len(out.Threads) {
		limit = len(out.Threads)
	}

	items := make([]ResultItem, 0, limit)
	for i := 0; i < limit; i++ {
		thread := out.Threads[i]
		var detail threadDetailJSON
		if err := runCLIJSON(cliPath, []string{"assistant", "thread", "get", thread.ID}, &detail); err != nil {
			items = append(items, ResultItem{
				Title:       fallbackTitle(thread.Title, thread.ID),
				Description: trimToLen("Could not load full thread details", 120),
				Detail: markdownDocument(thread.Title, []string{
					fmt.Sprintf("**Thread ID:** %s", thread.ID),
					optionalLine("**Created:** ", thread.CreatedAt),
					optionalSection("Load error", err.Error()),
				}),
			})
			continue
		}
		items = append(items, ResultItem{
			Title:       fallbackTitle(thread.Title, thread.ID),
			Description: trimToLen(threadConversationSummary(detail), 120),
			Detail:      threadDetailMarkdown(thread, detail),
		})
	}
	return items, nil
}

func executeAssistantDeleteThread(cliPath string, inputs map[string]string) ([]ResultItem, error) {
	threadID := requiredInput(inputs, "thread_id")
	if threadID == "" {
		return nil, errors.New("thread ID is required")
	}

	res := runCLI(cliPath, []string{"assistant", "thread", "delete", threadID})
	if res.err != nil {
		return nil, cliError(res)
	}
	body := strings.TrimSpace(strings.Join([]string{res.stdout, res.stderr}, "\n"))
	if body == "" {
		body = fmt.Sprintf("Thread %s deleted.", threadID)
	}
	return []ResultItem{{
		Title:       "Assistant thread deleted",
		Description: body,
		Detail:      markdownDocument("Thread deleted", []string{fmt.Sprintf("**Thread ID:** %s", threadID), body}),
	}}, nil
}

func balanceItems(cached api.BalanceCache) []ResultItem {
	detail := markdownDocument("API balance", []string{
		fmt.Sprintf("**Balance:** $%.4f", cached.APIBalance),
		optionalLine("**Updated:** ", cached.UpdatedAt),
		optionalLine("**Source:** ", cached.Source),
	})
	return []ResultItem{{
		Title:       "API balance",
		Description: fmt.Sprintf("$%.4f", cached.APIBalance),
		Detail:      detail,
	}}
}

func executeAuthCheck(cfg *config.Config) ([]ResultItem, error) {
	apiKey, apiErr := api.ResolveAPIKey(cfg)
	sessionToken, sessionErr := api.ResolveSessionToken(cfg)

	status := []string{}
	details := []string{}

	if apiErr != nil {
		status = append(status, "API key missing")
		details = append(details, "- **API key:** not configured")
		details = append(details, "- Set `KAGI_API_KEY` or save `api_key` in `~/.config/kagi/config.yaml`.")
	} else {
		client := api.NewHTTPClient(10 * time.Second)
		if err := validateAPIKey(client, apiKey); err != nil {
			status = append(status, "API key invalid")
			details = append(details, fmt.Sprintf("- **API key:** invalid (`%s…`) — %s", maskKey(apiKey), err))
		} else {
			status = append(status, "API key valid")
			details = append(details, fmt.Sprintf("- **API key:** valid (`%s…`)", maskKey(apiKey)))
		}
	}

	if sessionErr != nil {
		status = append(status, "session token missing")
		details = append(details, "- **Session token:** not configured (needed for subscriber features like assistant, quick, askpage, and translate).")
	} else {
		status = append(status, "session token configured")
		details = append(details, fmt.Sprintf("- **Session token:** configured (`%s…`)", maskKey(sessionToken)))
	}

	details = append(details, "\n**Note:** interactive mode validates the API key and reports session-token presence; subscriber-token validation remains an explicit live-command workflow.")

	return []ResultItem{{
		Title:       "Authentication",
		Description: strings.Join(status, " • "),
		Detail:      markdownDocument("Authentication status", details),
	}}, nil
}

func validateAPIKey(client *http.Client, apiKey string) error {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, authCheckSearchURL+"?q=test&limit=1", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bot "+apiKey)
	req.Header.Set("User-Agent", api.DefaultUserAgent)
	_, err = api.DoAPIRequest(client, req)
	return err
}

func executeConfig(cfg *config.Config, inputs map[string]string) ([]ResultItem, error) {
	configPath, err := userConfigPath()
	if err != nil {
		return nil, err
	}

	current := config.Config{}
	if cfg != nil {
		current = *cfg
	}

	updated := mergeConfigInputs(current, inputs)
	changed := configChanged(current, updated)
	if changed {
		if err := writeConfigFile(configPath, updated); err != nil {
			return nil, err
		}
	}

	status := "Showing current configuration"
	if changed {
		status = "Saved configuration updates"
	}

	detailLines := []string{
		fmt.Sprintf("**Config path:** `%s`", configPath),
		fmt.Sprintf("**API key:** %s", maskedOrMissing(updated.APIKey)),
		fmt.Sprintf("**Session token:** %s", maskedOrMissing(updated.SessionToken)),
		fmt.Sprintf("**Default format:** %s", fallbackText(updated.Defaults.Format, "not set")),
		fmt.Sprintf("**Default search region:** %s", fallbackText(updated.Defaults.Search.Region, "not set")),
		"",
		"Blank fields in the TUI preserve the saved values above. Clearing values is still a direct file-editing task.",
	}
	if changed {
		detailLines = append(detailLines, "", "Config file updated successfully.")
	} else {
		detailLines = append(detailLines, "", "No changes were written.")
	}

	return []ResultItem{{
		Title:       "Configuration",
		Description: status,
		Detail:      markdownDocument("Configuration", detailLines),
	}}, nil
}

func mergeConfigInputs(current config.Config, inputs map[string]string) config.Config {
	updated := current
	if apiKey := strings.TrimSpace(inputs["api_key"]); apiKey != "" {
		updated.APIKey = apiKey
	}
	if sessionToken := strings.TrimSpace(inputs["session_token"]); sessionToken != "" {
		updated.SessionToken = sessionToken
	}
	if defaultFormat := strings.TrimSpace(inputs["default_format"]); defaultFormat != "" {
		updated.Defaults.Format = defaultFormat
	}
	if searchRegion := strings.TrimSpace(inputs["search_region"]); searchRegion != "" {
		updated.Defaults.Search.Region = searchRegion
	}
	return updated
}

func configChanged(a, b config.Config) bool {
	return a.APIKey != b.APIKey ||
		a.SessionToken != b.SessionToken ||
		a.Defaults.Format != b.Defaults.Format ||
		a.Defaults.Search.Region != b.Defaults.Search.Region
}

func userConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("could not determine config directory: %w", err)
	}
	return filepath.Join(configDir, "kagi", "config.yaml"), nil
}

func writeConfigFile(path string, cfg config.Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

func executeCompletionHelp() []ResultItem {
	detail := markdownDocument("Shell completions", []string{
		"Interactive mode explains the setup flow, but generating the actual completion script remains a CLI export step.",
		"",
		"```bash\n# Bash\nkagi completion bash > /etc/bash_completion.d/kagi\n\n# Zsh\nkagi completion zsh > \"${fpath[1]}/_kagi\"\n\n# Fish\nkagi completion fish > ~/.config/fish/completions/kagi.fish\n\n# PowerShell\nkagi completion powershell >> $PROFILE\n```",
	})
	return []ResultItem{{
		Title:       "Shell completions",
		Description: "CLI-generated completion scripts for bash, zsh, fish, and PowerShell",
		Detail:      detail,
	}}
}

func executeVersion() []ResultItem {
	info := version.Info()
	return []ResultItem{{
		Title:       "Version",
		Description: strings.Split(info, "\n")[0],
		Detail:      markdownDocument("Version", []string{"```text\n" + info + "\n```"}),
	}}
}

func threadConversationSummary(detail threadDetailJSON) string {
	for _, msg := range detail.Messages {
		if strings.TrimSpace(msg.Prompt) != "" {
			return msg.Prompt
		}
	}
	return detail.Title
}

func threadDetailMarkdown(summary threadSummaryJSON, detail threadDetailJSON) string {
	parts := []string{fmt.Sprintf("**Thread ID:** %s", detail.ID)}
	if summary.CreatedAt != "" {
		parts = append(parts, fmt.Sprintf("**Created:** %s", summary.CreatedAt))
	}
	if summary.Saved {
		parts = append(parts, "**Saved:** yes")
	}

	for _, msg := range detail.Messages {
		if msg.Prompt != "" {
			parts = append(parts, fmt.Sprintf("### User\n\n%s", msg.Prompt))
		}
		if msg.Reply != "" {
			parts = append(parts, fmt.Sprintf("### Assistant\n\n%s", htmlToMarkdown(msg.Reply)))
		}
	}

	return markdownDocument(fallbackTitle(detail.Title, detail.ID), parts)
}

func markdownDocument(title string, sections []string) string {
	filtered := make([]string, 0, len(sections))
	for _, section := range sections {
		section = strings.TrimSpace(section)
		if section != "" {
			filtered = append(filtered, section)
		}
	}
	if len(filtered) == 0 {
		return "# " + title
	}
	return "# " + title + "\n\n" + strings.Join(filtered, "\n\n")
}

func optionalSection(title, body string) string {
	body = strings.TrimSpace(body)
	if body == "" {
		return ""
	}
	return "## " + title + "\n\n" + body
}

func optionalLine(prefix, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return prefix + value
}

func referenceMarkdown(refs []struct {
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
	URL     string `json:"url"`
},
) string {
	if len(refs) == 0 {
		return ""
	}
	lines := make([]string, 0, len(refs))
	for _, ref := range refs {
		line := fmt.Sprintf("- [%s](%s)", fallbackTitle(ref.Title, ref.URL), ref.URL)
		if snippet := strings.TrimSpace(ref.Snippet); snippet != "" {
			line += " — " + snippet
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func fallbackTitle(title, fallback string) string {
	title = strings.TrimSpace(title)
	if title != "" {
		return title
	}
	fallback = strings.TrimSpace(fallback)
	if fallback != "" {
		return fallback
	}
	return "Untitled"
}

func maskKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "…" + key[len(key)-2:]
}

func maskedOrMissing(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "not set"
	}
	return maskKey(value)
}

func fallbackText(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
