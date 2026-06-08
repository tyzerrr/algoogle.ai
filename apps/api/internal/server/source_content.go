package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

const leetcodeGraphQLEndpoint = "https://leetcode.com/graphql"

type ProblemSourceProvider interface {
	Fetch(ctx context.Context, problem Problem) (*OfficialProblemContent, error)
}

type ProblemSourceFetcher struct {
	client  *http.Client
	cacheMu sync.Mutex
	cache   map[string]OfficialProblemContent
}

func NewProblemSourceFetcher() *ProblemSourceFetcher {
	return &ProblemSourceFetcher{
		client: &http.Client{Timeout: 12 * time.Second},
		cache:  map[string]OfficialProblemContent{},
	}
}

func (f *ProblemSourceFetcher) Fetch(ctx context.Context, problem Problem) (*OfficialProblemContent, error) {
	slug := leetcodeSlug(problem.SourceURL)
	if slug == "" {
		return nil, errors.New("LeetCode source URL is not available")
	}

	f.cacheMu.Lock()
	if cached, ok := f.cache[slug]; ok {
		f.cacheMu.Unlock()
		return &cached, nil
	}
	f.cacheMu.Unlock()

	content, err := f.fetchLeetCode(ctx, slug, problem.SourceURL)
	if err != nil {
		return nil, err
	}
	parsed := parseLeetCodeProblemContent(problem.Title, problem.SourceURL, content)
	if parsed.Statement == "" && len(parsed.Examples) == 0 {
		return nil, errors.New("official problem content was empty")
	}

	f.cacheMu.Lock()
	f.cache[slug] = parsed
	f.cacheMu.Unlock()
	return &parsed, nil
}

func (f *ProblemSourceFetcher) fetchLeetCode(ctx context.Context, slug, sourceURL string) (string, error) {
	payload := map[string]interface{}{
		"query": `query questionData($titleSlug: String!) {
  question(titleSlug: $titleSlug) {
    title
    content
    exampleTestcases
  }
}`,
		"variables": map[string]string{"titleSlug": slug},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, leetcodeGraphQLEndpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", sourceURL)
	req.Header.Set("User-Agent", "Mozilla/5.0 algoogle.ai problem fetcher")

	resp, err := f.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch LeetCode content: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("fetch LeetCode content: status %d", resp.StatusCode)
	}

	var decoded struct {
		Data struct {
			Question *struct {
				Title   string `json:"title"`
				Content string `json:"content"`
			} `json:"question"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return "", fmt.Errorf("decode LeetCode content: %w", err)
	}
	if len(decoded.Errors) > 0 {
		return "", fmt.Errorf("fetch LeetCode content: %s", decoded.Errors[0].Message)
	}
	if decoded.Data.Question == nil || decoded.Data.Question.Content == "" {
		return "", errors.New("LeetCode returned no problem content")
	}
	return decoded.Data.Question.Content, nil
}

func leetcodeSlug(sourceURL string) string {
	parsed, err := url.Parse(sourceURL)
	if err != nil {
		return ""
	}
	if parsed.Host != "leetcode.com" && !strings.HasSuffix(parsed.Host, ".leetcode.com") {
		return ""
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	for i, part := range parts {
		if part == "problems" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

func parseLeetCodeProblemContent(title, sourceURL, contentHTML string) OfficialProblemContent {
	withoutConstraints := cutBeforeConstraints(contentHTML)
	statementHTML := contentBeforeFirstExample(withoutConstraints)
	examples := parseLeetCodeExamples(withoutConstraints)
	images := parseLeetCodeImages(withoutConstraints, sourceURL)

	return OfficialProblemContent{
		Source:    "LeetCode",
		SourceURL: sourceURL,
		Title:     title,
		Statement: htmlToPlainText(statementHTML),
		Examples:  examples,
		Images:    images,
		FetchedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

var (
	constraintsHeadingPattern = regexp.MustCompile(`(?is)<strong[^>]*>\s*Constraints\s*:?\s*</strong>`)
	exampleHeadingPattern     = regexp.MustCompile(`(?is)<strong[^>]*>\s*Example\s+\d+\s*:?\s*</strong>`)
	exampleBlockPattern       = regexp.MustCompile(`(?is)<strong[^>]*>\s*Example\s+(\d+)\s*:?\s*</strong>.*?<pre[^>]*>(.*?)</pre>`)
	imageTagPattern           = regexp.MustCompile(`(?is)<img\b[^>]*>`)
	attributePattern          = regexp.MustCompile(`(?is)([a-zA-Z_:][-a-zA-Z0-9_:.]*)\s*=\s*("([^"]*)"|'([^']*)'|([^\s"'>]+))`)
	tagPattern                = regexp.MustCompile(`(?is)<[^>]+>`)
	spacePattern              = regexp.MustCompile(`[ \t\r\f\v]+`)
	blankLinePattern          = regexp.MustCompile(`\n{3,}`)
)

func cutBeforeConstraints(contentHTML string) string {
	match := constraintsHeadingPattern.FindStringIndex(contentHTML)
	if match == nil {
		return contentHTML
	}
	return contentHTML[:match[0]]
}

func contentBeforeFirstExample(contentHTML string) string {
	match := exampleHeadingPattern.FindStringIndex(contentHTML)
	if match == nil {
		return contentHTML
	}
	return contentHTML[:match[0]]
}

func parseLeetCodeExamples(contentHTML string) []Example {
	matches := exampleBlockPattern.FindAllStringSubmatch(contentHTML, -1)
	examples := make([]Example, 0, len(matches))
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		example := parseExamplePre(match[2])
		if example.Input == "" && example.Output == "" {
			continue
		}
		examples = append(examples, example)
	}
	return examples
}

func parseLeetCodeImages(contentHTML, sourceURL string) []ProblemImage {
	tags := imageTagPattern.FindAllString(contentHTML, -1)
	images := make([]ProblemImage, 0, len(tags))
	seen := map[string]bool{}
	for _, tag := range tags {
		attributes := parseHTMLAttributes(tag)
		rawURL := attributes["src"]
		if rawURL == "" {
			continue
		}
		imageURL := normalizeImageURL(rawURL, sourceURL)
		if imageURL == "" || seen[imageURL] {
			continue
		}
		seen[imageURL] = true
		images = append(images, ProblemImage{
			URL: imageURL,
			Alt: strings.TrimSpace(attributes["alt"]),
		})
	}
	return images
}

func parseHTMLAttributes(tag string) map[string]string {
	attributes := map[string]string{}
	for _, match := range attributePattern.FindAllStringSubmatch(tag, -1) {
		if len(match) < 6 {
			continue
		}
		value := match[3]
		if value == "" {
			value = match[4]
		}
		if value == "" {
			value = match[5]
		}
		attributes[strings.ToLower(match[1])] = strings.TrimSpace(html.UnescapeString(value))
	}
	return attributes
}

func normalizeImageURL(rawURL, sourceURL string) string {
	rawURL = strings.TrimSpace(html.UnescapeString(rawURL))
	if rawURL == "" || strings.HasPrefix(rawURL, "data:") {
		return ""
	}
	base, err := url.Parse(sourceURL)
	if err != nil {
		return ""
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	resolved := base.ResolveReference(parsed)
	if resolved.Scheme == "" && strings.HasPrefix(rawURL, "//") {
		resolved.Scheme = "https"
	}
	if resolved.Scheme != "http" && resolved.Scheme != "https" {
		return ""
	}
	return resolved.String()
}

func parseExamplePre(preHTML string) Example {
	text := htmlToPlainText(preHTML)
	lines := strings.Split(text, "\n")
	var example Example
	var explanation []string
	section := ""

	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}
		switch {
		case strings.HasPrefix(line, "Input:"):
			section = "input"
			example.Input = strings.TrimSpace(strings.TrimPrefix(line, "Input:"))
		case strings.HasPrefix(line, "Output:"):
			section = "output"
			example.Output = strings.TrimSpace(strings.TrimPrefix(line, "Output:"))
		case strings.HasPrefix(line, "Explanation:"):
			section = "explanation"
			explanation = append(explanation, strings.TrimSpace(strings.TrimPrefix(line, "Explanation:")))
		case section == "input":
			example.Input = strings.TrimSpace(example.Input + "\n" + line)
		case section == "output":
			example.Output = strings.TrimSpace(example.Output + "\n" + line)
		case section == "explanation":
			explanation = append(explanation, line)
		}
	}
	example.Explanation = strings.TrimSpace(strings.Join(explanation, "\n"))
	return example
}

func htmlToPlainText(fragment string) string {
	text := fragment
	replacements := []struct {
		from *regexp.Regexp
		to   string
	}{
		{regexp.MustCompile(`(?i)<br\s*/?>`), "\n"},
		{regexp.MustCompile(`(?i)</(p|div|li|pre|ul|ol|h\d)>`), "\n"},
		{regexp.MustCompile(`(?i)<li[^>]*>`), "\n- "},
	}
	for _, replacement := range replacements {
		text = replacement.from.ReplaceAllString(text, replacement.to)
	}
	text = tagPattern.ReplaceAllString(text, "")
	text = html.UnescapeString(text)
	text = strings.ReplaceAll(text, "\u00a0", " ")

	lines := strings.Split(text, "\n")
	cleaned := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(spacePattern.ReplaceAllString(line, " "))
		if line == "" {
			if len(cleaned) > 0 && cleaned[len(cleaned)-1] != "" {
				cleaned = append(cleaned, "")
			}
			continue
		}
		cleaned = append(cleaned, line)
	}
	return strings.TrimSpace(blankLinePattern.ReplaceAllString(strings.Join(cleaned, "\n"), "\n\n"))
}
