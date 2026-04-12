package service

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"ai-blog/backend/internal/model"
	"ai-blog/backend/internal/repository"
)

const (
	briefingStatusDraft     = 0
	briefingStatusPublished = 1
)

var htmlTagPattern = regexp.MustCompile(`<[^>]+>`)

type DailyBriefingFetchResult struct {
	Trigger      string    `json:"trigger"`
	Date         string    `json:"date"`
	StartedAt    time.Time `json:"started_at"`
	FinishedAt   time.Time `json:"finished_at"`
	Success      bool      `json:"success"`
	Message      string    `json:"message"`
	FetchedCount int       `json:"fetched_count"`
	SavedCount   int       `json:"saved_count"`
	RemovedCount int       `json:"removed_count"`
}

type DailyBriefingService struct {
	repo repository.DailyBriefingRepository

	mutex      sync.RWMutex
	isRunning  bool
	lastResult DailyBriefingFetchResult
}

type rssFeed struct {
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title string    `xml:"title"`
	Items []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	Source      string `xml:"source"`
}

type fetchedBriefingCandidate struct {
	Title             string
	Summary           string
	SourceName        string
	SourceURL         string
	SourcePublishedAt *time.Time
	OriginFeed        string
}

func NewDailyBriefingService(repo repository.DailyBriefingRepository) *DailyBriefingService {
	return &DailyBriefingService{
		repo: repo,
		lastResult: DailyBriefingFetchResult{
			Success: false,
			Message: "No briefing fetch has run yet.",
		},
	}
}

func (service *DailyBriefingService) ListPublic(date string) (map[string]any, error) {
	briefingDate := strings.TrimSpace(date)
	if briefingDate == "" {
		latestDate, err := service.repo.LatestPublishedDate()
		if err != nil {
			return nil, err
		}
		briefingDate = latestDate
	}

	availableDates, err := service.repo.ListPublishedDates(14)
	if err != nil {
		return nil, err
	}

	if briefingDate == "" {
		return map[string]any{
			"date":            "",
			"items":           []model.DailyBriefing{},
			"available_dates": availableDates,
		}, nil
	}

	items, err := service.repo.ListPublicByDate(briefingDate)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"date":            briefingDate,
		"items":           items,
		"available_dates": availableDates,
	}, nil
}

func (service *DailyBriefingService) ListAdmin(date string, keyword string, status *int, page int, pageSize int) (repository.DailyBriefingListResult, error) {
	return service.repo.ListAdmin(date, keyword, status, page, pageSize)
}

func (service *DailyBriefingService) SaveBriefing(id int64, briefingDate string, title string, summary string, sourceName string, sourceURL string, status int, sortOrder int, sourcePublishedAt string) (model.DailyBriefing, error) {
	briefingDate = strings.TrimSpace(briefingDate)
	if briefingDate == "" {
		briefingDate = time.Now().Format("2006-01-02")
	}
	if _, err := time.Parse("2006-01-02", briefingDate); err != nil {
		return model.DailyBriefing{}, errors.New("briefing date must use YYYY-MM-DD")
	}

	title = strings.TrimSpace(title)
	if title == "" {
		return model.DailyBriefing{}, errors.New("title is required")
	}

	parsedPublishedAt, err := parseBriefingTime(sourcePublishedAt)
	if err != nil {
		return model.DailyBriefing{}, err
	}

	if status != briefingStatusDraft && status != briefingStatusPublished {
		status = briefingStatusPublished
	}

	var item model.DailyBriefing
	if id > 0 {
		item, err = service.repo.FindByID(id)
		if err != nil {
			return model.DailyBriefing{}, err
		}
	} else {
		item = model.DailyBriefing{
			SourceType: "manual",
			Region:     "global",
			Language:   "en",
		}
	}

	item.BriefingDate = briefingDate
	item.Title = title
	item.Summary = strings.TrimSpace(summary)
	item.SourceName = strings.TrimSpace(sourceName)
	item.SourceURL = strings.TrimSpace(sourceURL)
	item.Status = status
	item.SortOrder = sortOrder
	item.SourcePublishedAt = parsedPublishedAt
	item.SourceHash = buildBriefingHash(item.BriefingDate, item.SourceURL, item.Title)

	if id > 0 {
		err = service.repo.Update(&item)
	} else {
		err = service.repo.Create(&item)
	}
	if err != nil {
		return model.DailyBriefing{}, err
	}

	return item, nil
}

func (service *DailyBriefingService) DeleteBriefing(id int64) error {
	return service.repo.Delete(id)
}

func (service *DailyBriefingService) Status() map[string]any {
	service.mutex.RLock()
	defer service.mutex.RUnlock()

	return map[string]any{
		"is_running":     service.isRunning,
		"last_result":    service.lastResult,
		"default_limit":  defaultBriefingLimit(),
		"script_command": "go run ./cmd/dailybriefing_fetcher",
	}
}

func (service *DailyBriefingService) FetchNow(ctx context.Context, date string, limit int, trigger string) (DailyBriefingFetchResult, error) {
	result := DailyBriefingFetchResult{
		Trigger:   trigger,
		Date:      normalizeBriefingDate(date),
		StartedAt: time.Now(),
		Success:   false,
		Message:   "Fetching latest AI briefings.",
	}

	finalize := func(fetchErr error) (DailyBriefingFetchResult, error) {
		result.FinishedAt = time.Now()
		service.setLastResult(result, false)
		return result, fetchErr
	}

	service.mutex.Lock()
	if service.isRunning {
		lastResult := service.lastResult
		service.mutex.Unlock()
		return lastResult, errors.New("a briefing fetch is already running")
	}
	service.isRunning = true
	service.lastResult = result
	service.mutex.Unlock()

	candidates, err := service.fetchCandidates(ctx, limit)
	if err != nil {
		result.Message = "Failed to fetch AI briefing feeds."
		return finalize(err)
	}

	result.FetchedCount = len(candidates)
	if len(candidates) == 0 {
		result.Message = "No AI news items were found from the configured feeds."
		return finalize(errors.New(result.Message))
	}

	items := make([]model.DailyBriefing, 0, len(candidates))
	for index, candidate := range candidates {
		items = append(items, model.DailyBriefing{
			BriefingDate:      result.Date,
			Title:             candidate.Title,
			Summary:           candidate.Summary,
			SourceName:        candidate.SourceName,
			SourceURL:         candidate.SourceURL,
			SourceHash:        buildBriefingHash(result.Date, candidate.SourceURL, candidate.Title),
			SourceType:        "auto",
			Status:            briefingStatusPublished,
			SortOrder:         index + 1,
			Region:            "global",
			Language:          "en",
			OriginFeed:        candidate.OriginFeed,
			SourcePublishedAt: candidate.SourcePublishedAt,
		})
	}

	if err := service.repo.DeleteAutoByDate(result.Date); err != nil {
		result.Message = "Failed to clear previous auto-fetched briefings."
		return finalize(err)
	}
	if len(items) > 0 {
		result.RemovedCount = 1
	}

	for index := range items {
		if err := service.repo.Create(&items[index]); err != nil {
			result.Message = "Failed to store fetched AI briefings."
			return finalize(err)
		}
		result.SavedCount++
	}

	result.Success = true
	result.Message = fmt.Sprintf("Fetched %d AI briefing items for %s.", result.SavedCount, result.Date)
	return finalize(nil)
}

func (service *DailyBriefingService) fetchCandidates(ctx context.Context, limit int) ([]fetchedBriefingCandidate, error) {
	if limit <= 0 {
		limit = defaultBriefingLimit()
	}

	httpClient := &http.Client{Timeout: 15 * time.Second}
	candidates := make([]fetchedBriefingCandidate, 0, limit*2)
	seen := make(map[string]struct{})
	errorMessages := make([]string, 0)

	for _, feedURL := range defaultBriefingFeeds() {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
		if err != nil {
			errorMessages = append(errorMessages, err.Error())
			continue
		}
		req.Header.Set("User-Agent", "codex-blog-daily-briefing/1.0")

		resp, err := httpClient.Do(req)
		if err != nil {
			errorMessages = append(errorMessages, err.Error())
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			errorMessages = append(errorMessages, err.Error())
			continue
		}
		if resp.StatusCode >= http.StatusBadRequest {
			errorMessages = append(errorMessages, fmt.Sprintf("%s returned status %d", feedURL, resp.StatusCode))
			continue
		}

		var feed rssFeed
		if err := xml.Unmarshal(body, &feed); err != nil {
			errorMessages = append(errorMessages, err.Error())
			continue
		}

		for _, item := range feed.Channel.Items {
			candidate := normalizeCandidate(item, feed.Channel.Title)
			if candidate.Title == "" || candidate.SourceURL == "" {
				continue
			}

			key := strings.ToLower(candidate.Title + "|" + candidate.SourceURL)
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			candidates = append(candidates, candidate)
		}
	}

	sort.SliceStable(candidates, func(i int, j int) bool {
		left := candidates[i].SourcePublishedAt
		right := candidates[j].SourcePublishedAt
		if left == nil && right == nil {
			return candidates[i].Title < candidates[j].Title
		}
		if left == nil {
			return false
		}
		if right == nil {
			return true
		}
		return left.After(*right)
	})

	if len(candidates) > limit {
		candidates = candidates[:limit]
	}

	if len(candidates) == 0 && len(errorMessages) > 0 {
		return nil, errors.New(strings.Join(errorMessages, "; "))
	}

	return candidates, nil
}

func (service *DailyBriefingService) setLastResult(result DailyBriefingFetchResult, isRunning bool) {
	service.mutex.Lock()
	defer service.mutex.Unlock()

	service.lastResult = result
	service.isRunning = isRunning
}

func defaultBriefingLimit() int {
	return 10
}

func normalizeBriefingDate(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return time.Now().Format("2006-01-02")
	}
	return trimmed
}

func defaultBriefingFeeds() []string {
	return []string{
		"https://www.artificialintelligence-news.com/feed/",
		"https://techcrunch.com/tag/artificial-intelligence/feed/",
		"https://www.marktechpost.com/feed/",
		"https://hnrss.org/frontpage?q=AI",
		"https://www.reddit.com/r/artificial/.rss",
		fmt.Sprintf(
			"https://news.google.com/rss/search?q=%s&hl=en-US&gl=US&ceid=US:en",
			url.QueryEscape(`"artificial intelligence" when:1d`),
		),
	}
}

func normalizeCandidate(item rssItem, feedTitle string) fetchedBriefingCandidate {
	title := cleanBriefingText(item.Title)
	sourceName := cleanBriefingText(item.Source)
	if sourceName != "" {
		title = strings.TrimSuffix(title, " - "+sourceName)
	}
	if sourceName == "" {
		sourceName = inferSourceFromTitle(item.Title)
	}
	if sourceName == "" {
		sourceName = cleanBriefingText(feedTitle)
	}

	summary := cleanBriefingText(item.Description)
	if len([]rune(summary)) > 180 {
		runes := []rune(summary)
		summary = string(runes[:180]) + "..."
	}
	if summary == "" && sourceName != "" {
		summary = fmt.Sprintf("From %s. Open the original article for the full report.", sourceName)
	}

	return fetchedBriefingCandidate{
		Title:             title,
		Summary:           summary,
		SourceName:        sourceName,
		SourceURL:         strings.TrimSpace(item.Link),
		SourcePublishedAt: parsePubDate(item.PubDate),
		OriginFeed:        cleanBriefingText(feedTitle),
	}
}

func inferSourceFromTitle(rawTitle string) string {
	parts := strings.Split(cleanBriefingText(rawTitle), " - ")
	if len(parts) < 2 {
		return ""
	}
	return strings.TrimSpace(parts[len(parts)-1])
}

func cleanBriefingText(value string) string {
	text := html.UnescapeString(value)
	text = htmlTagPattern.ReplaceAllString(text, " ")
	text = strings.ReplaceAll(text, "\u00a0", " ")
	text = strings.Join(strings.Fields(text), " ")
	return strings.TrimSpace(text)
}

func parsePubDate(value string) *time.Time {
	layouts := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822Z,
		time.RFC822,
		time.RFC3339,
	}

	for _, layout := range layouts {
		parsed, err := time.Parse(layout, strings.TrimSpace(value))
		if err == nil {
			return &parsed
		}
	}
	return nil
}

func parseBriefingTime(value string) (*time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, nil
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04",
		"2006-01-02 15:04",
	}
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, trimmed)
		if err == nil {
			return &parsed, nil
		}
	}

	return nil, errors.New("source published time must use RFC3339 or YYYY-MM-DDTHH:mm")
}

func buildBriefingHash(date string, sourceURL string, title string) string {
	sum := sha1.Sum([]byte(strings.TrimSpace(date) + "|" + strings.TrimSpace(sourceURL) + "|" + strings.TrimSpace(title)))
	return hex.EncodeToString(sum[:])
}
