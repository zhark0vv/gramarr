package sonarr

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"golang.org/x/sync/errgroup"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var apiRgx = regexp.MustCompile(`[a-z0-9]{32}`)

type Client struct {
	apiKey     string
	username   string
	password   string
	baseURL    string
	maxResults int
	client     *resty.Client
}

func New(c Config) (*Client, error) {
	if c.Hostname == "" {
		return nil, fmt.Errorf("hostname is empty")
	}

	if match := apiRgx.MatchString(c.APIKey); !match {
		return nil, fmt.Errorf("api key is invalid format: %s", c.APIKey)
	}

	baseURL := createApiURL(c)

	r := resty.New()
	r.SetHostURL(baseURL)
	r.SetHeader("Accept", "application/json")
	r.SetQueryParam("apikey", c.APIKey)
	if c.Username != "" && c.Password != "" {
		r.SetBasicAuth(c.Username, c.Password)
	}

	client := &Client{
		apiKey:     c.APIKey,
		maxResults: c.MaxResults,
		username:   c.Username,
		password:   c.Password,
		baseURL:    baseURL,
		client:     r,
	}
	return client, nil
}

func createApiURL(c Config) string {
	c.Hostname = strings.TrimPrefix(c.Hostname, "http://")
	c.Hostname = strings.TrimPrefix(c.Hostname, "https://")
	c.URLBase = strings.TrimPrefix(c.URLBase, "/")

	u := url.URL{}

	u.Scheme = "http"

	if c.SSL {
		u.Scheme = "https"
	}

	u.Host = c.Hostname
	if c.Port != 80 {
		u.Host = fmt.Sprintf("%s:%d", c.Hostname, c.Port)
	}
	u.Path = "/api/v3"
	if c.URLBase != "" {
		u.Path = fmt.Sprintf("%s/api", c.URLBase)
	}

	return u.String()
}

func (c *Client) SearchTVShows(term string) ([]TVShow, error) {
	resp, err := c.client.R().SetQueryParam("term", term).SetResult([]TVShow{}).Get("series/lookup")
	if err != nil {
		return nil, err
	}

	TVShows := *resp.Result().(*[]TVShow)
	if len(TVShows) > c.maxResults {
		TVShows = TVShows[:c.maxResults]
	}
	return TVShows, nil
}

func (c *Client) GetFolders() ([]Folder, error) {
	resp, err := c.client.R().SetResult([]Folder{}).Get("rootfolder")
	if err != nil {
		return nil, err
	}

	folders := *resp.Result().(*[]Folder)
	return folders, nil
}

func (c *Client) GetProfile(p string) ([]Profile, error) {
	resp, err := c.client.R().SetResult([]Profile{}).Get(p)
	if err != nil {
		return nil, err
	}
	profile := *resp.Result().(*[]Profile)

	return profile, nil

}

func (c *Client) AddTVShow(m TVShow, qualityProfile int, path string, seriestype string) (tvShow TVShow, err error) {

	request := AddTVShowRequest{
		Title:            m.Title,
		TitleSlug:        m.TitleSlug,
		Images:           m.Images,
		QualityProfileID: qualityProfile,
		TVDBID:           m.TVDBID,
		RootFolderPath:   path,
		Monitored:        true,
		Year:             m.Year,
		Seasons:          m.Seasons,
		AddOptions:       AddTVShowOptions{SearchForMissingEpisodes: true},
		SeriesType:       seriestype,
	}

	resp, err := c.client.R().SetBody(request).SetResult(TVShow{}).Post("series")
	if err != nil {
		fmt.Println(err)
		return
	}

	tvShow = *resp.Result().(*TVShow)
	return
}

func (c *Client) GetTVShows(tmdbID int) (movie []TVShow, err error) {
	resp, err := c.client.R().
		SetResult([]TVShow{}).
		SetQueryParam("tmdbId", strconv.Itoa(tmdbID)).
		Get("series")
	if err != nil {
		return
	}

	movie = *resp.Result().(*[]TVShow)
	return
}

func (c *Client) GetReleases(seriesID int, seasons []TVShowSeason) ([]Release, error) {
	eg := errgroup.Group{}
	mu := sync.Mutex{}

	// Use a map to track unique releases
	releaseMap := make(map[string]Release)

	for _, season := range seasons {
		season := season

		eg.Go(func() error {
			resp, err := c.client.R().
				SetResult([]Release{}).
				SetQueryParams(
					map[string]string{
						"seriesId":     strconv.Itoa(seriesID),
						"seasonNumber": strconv.Itoa(season.SeasonNumber),
					}).
				Get("release")

			if err != nil {
				return fmt.Errorf("failed to get releases: %w", err)
			}

			mu.Lock()
			defer mu.Unlock()

			for _, release := range *resp.Result().(*[]Release) {
				if _, exists := releaseMap[release.Title]; !exists {
					releaseMap[release.Title] = release
				}
			}
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	// Convert the map back to a slice
	releasesBatch := make([]Release, 0, len(releaseMap))
	for _, release := range releaseMap {
		releasesBatch = append(releasesBatch, release)
	}

	return releasesBatch, nil
}

func (c *Client) DeleteTVShow(seriesID int) (err error) {
	_, err = c.client.R().
		SetQueryParam("deleteFiles", "true").
		Delete("series/" + strconv.Itoa(seriesID))
	return
}

func (c *Client) DownloadRelease(guid string) (r Release, err error) {
	resp, err := c.client.R().
		SetBody(map[string]string{
			"guid": guid,
			// TODO: parametrize id
			"indexerId": "1",
		}).
		SetResult(Release{}).
		Post("release")

	if err != nil {
		return Release{}, fmt.Errorf("failed to download release: %w", err)
	}

	r = *resp.Result().(*Release)
	return
}
