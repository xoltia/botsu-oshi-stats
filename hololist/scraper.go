package hololist

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/time/rate"
)

type UnknownError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e UnknownError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

var (
	ErrExhaustedPosts = errors.New("posts exhausted")
)

const postsEndpoint = "https://hololist.net/wp-json/wp/v2/posts"

type Scraper struct {
	limiter *rate.Limiter
	client  *http.Client
	offset  int
}

func NewScraper(client *http.Client, limiter *rate.Limiter) *Scraper {
	return &Scraper{
		client:  client,
		limiter: limiter,
		offset:  0,
	}
}

// Resets post pagination.
func (s *Scraper) Reset() {
	s.offset = 0
}

func (s *Scraper) getWithBackoff(ctx context.Context, url string, initial time.Duration, maxAttempts int) (*http.Response, error) {
	return s.getWithBackoffAttempt(ctx, url, initial, 0, maxAttempts)
}

func (s *Scraper) getWithBackoffAttempt(ctx context.Context, url string, initial time.Duration, attempt, maxAttempts int) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	err = s.limiter.Wait(req.Context())
	if err != nil {
		return nil, err
	}
	res, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusForbidden {
		return res, nil
	}
	res.Body.Close()
	if attempt+1 >= maxAttempts {
		return nil, ErrBackoff
	}

	log.Printf("Retrying [%d]: %s\n", attempt+1, url)

	if attempt > 0 {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(initial << attempt):
		}
	}

	return s.getWithBackoffAttempt(ctx, url, initial, attempt+1, maxAttempts)
}

// Get the next page of posts. Returns ErrExhaustedPosts when there are no more to fetch.
// Not safe for concurrent access.
func (s *Scraper) NextPosts(ctx context.Context, limit int) ([]VTuberMeta, error) {
	url := fmt.Sprintf("%s?type=216&per_page=%d&offset=%d", postsEndpoint, limit, s.offset)
	res, err := s.getWithBackoff(ctx, url, time.Second, 5)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(res.Body)

	if res.Header.Get("X-WP-Total") == "0" {
		return nil, ErrExhaustedPosts
	}
	if res.StatusCode == http.StatusForbidden {
		return nil, ErrBackoff
	}

	if res.StatusCode != http.StatusOK {
		var apiError UnknownError
		err := decoder.Decode(&apiError)
		if err != nil {
			return nil, fmt.Errorf("decode api error: %w", err)
		}

		if apiError.Code == "rest_post_invalid_page_number" {
			return nil, ErrExhaustedPosts
		} else {
			return nil, apiError
		}
	}

	result := make([]VTuberMeta, 0, limit)
	if err = decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("decode posts: %w", err)
	}
	s.offset += len(result)
	return result, nil
}

var (
	handleRegex = regexp.MustCompile(`(?:youtube.com/)(@[A-Za-z0-9-_]+)`)
)

// Get a rendered post from the webpage URL. Can be obtained from `VTuberMeta.URL`.
// Safe to use concurrently.
func (s *Scraper) GetRenderedPost(ctx context.Context, url string) (v VTuberRendered, err error) {
	res, err := s.getWithBackoff(ctx, url, time.Second, 5)
	if err != nil {
		return
	}

	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("status not ok: %s", res.Status)
		return
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return
	}

	v.OriginalName = doc.Find("#original-name").First().Text()
	v.OriginalName = lastLineStripped(v.OriginalName)
	v.EnglishName = doc.Find("h1").First().Text()
	v.EnglishName = lastLineStripped(v.EnglishName)
	v.OshiMark = doc.Find("#oshi-mark").First().Text()
	v.OshiMark = lastLineStripped(v.OshiMark)
	v.Zodiac = doc.Find("#zodiac").First().Text()
	v.Zodiac = lastLineStripped(v.Zodiac)
	v.Affiliation = doc.Find("#affiliation").First().Text()
	v.Affiliation = lastLineStripped(v.Affiliation)
	v.Birthday = doc.Find("#birthday").First().Text()
	v.Birthday = formatDate(removeFromParen(lastLineStripped(v.Birthday)))
	v.DebutDate = doc.Find("#debut").First().Text()
	v.DebutDate = formatDate(removeFromParen(lastLineStripped(v.DebutDate)))
	v.Gender = doc.Find("#gender").First().Text()
	v.Gender = lastLineStripped(v.Gender)
	v.Height = doc.Find("#height").First().Text()
	v.Height = lastLineStripped(v.Height)
	v.Fanbase = doc.Find("#fanbase").First().Text()
	v.Fanbase = lastLineStripped(v.Fanbase)
	v.Status = doc.Find("#status").First().Text()
	v.Status = lastLineStripped(v.Status)

	img := doc.Find("#left").First().Find("a").First().AttrOr("href", "")
	v.PictureURL = img

	prefix := "https://www.youtube.com/channel/"
	links := doc.Find("#links").Find("a").EachIter()
	for _, link := range links {
		href := link.AttrOr("href", "")
		text := link.Text()
		if strings.HasPrefix(href, prefix) {
			pathQuery := strings.TrimPrefix(href, prefix)
			parts := strings.SplitN(pathQuery, "?", 2)
			if len(parts) > 0 {
				v.YouTubeID = parts[0]
			}
			handleMatch := handleRegex.FindStringSubmatch(text)
			if len(handleMatch) == 2 {
				v.YouTubeHandle = handleMatch[1]
			}
			break
		}
	}

	return
}

func filterEmpty(text string) string {
	if text == "...." {
		return ""
	}
	return text
}

func removeFromParen(text string) string {
	parenIndex := strings.IndexByte(text, '(')
	if parenIndex == -1 {
		return text
	}
	return text[:parenIndex]
}

func lastLineStripped(text string) string {
	text = strings.TrimSpace(text)
	lastLineStart := strings.LastIndexByte(text, '\n')
	if lastLineStart == -1 {
		return filterEmpty(text)
	}
	lastLine := text[lastLineStart:]
	return filterEmpty(strings.TrimSpace(lastLine))
}

// Takes a date in format: January 2[, 1970]
// Returns in format: [1970-]01-02
func formatDate(text string) string {
	parts := strings.Fields(text)
	if len(parts) < 2 {
		return ""
	}

	var (
		year  string
		month string
		day   string
	)

	months := map[string]string{
		"January":   "01",
		"February":  "02",
		"March":     "03",
		"April":     "04",
		"May":       "05",
		"June":      "06",
		"July":      "07",
		"August":    "08",
		"September": "09",
		"October":   "10",
		"November":  "11",
		"December":  "12",
	}

	month = months[parts[0]]
	if month == "" {
		return ""
	}

	day = parts[1]
	hasYear := day[len(day)-1] == ','
	if hasYear {
		day = day[0 : len(day)-1]
		month = "-" + month
	}

	if len(day) > 2 || len(day) == 0 {
		return ""
	} else if len(day) == 1 {
		day = "0" + day
	}
	day = "-" + day

	if hasYear {
		if len(parts) != 3 {
			return ""
		}
		year = parts[2]
	}

	return fmt.Sprintf("%s%s%s", year, month, day)
}
