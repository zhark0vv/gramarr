package radarr

import (
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"

	"github.com/go-resty/resty/v2"
)

var (
	apiRgx = regexp.MustCompile(`[a-z0-9]{32}`)
)

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

	u.Path = "/api"
	if c.URLBase != "" {
		u.Path = fmt.Sprintf("%s/api", c.URLBase)
	}

	return u.String()
}

func (c *Client) SearchMovies(term string) ([]Movie, error) {
	// Prepare the request
	req := c.client.R().
		SetQueryParam("term", term).
		SetResult([]Movie{})

	// Log the full request URL
	fullURL := req.URL + "?term=" + term
	log.Printf("Sending request to: %s", fullURL)

	// Execute the request
	resp, err := req.Get("movie/lookup")
	if err != nil {
		log.Printf("Failed to get response: %v", err)
		return nil, err
	}

	// Log the response status and body
	log.Printf("Received response with status: %s", resp.Status())
	log.Printf("Response body: %s", resp.String())

	// Process the response
	movies := *resp.Result().(*[]Movie)
	if len(movies) > c.maxResults {
		movies = movies[:c.maxResults]
	}
	return movies, nil
}

func (c *Client) GetProfile(prfl string) ([]Profile, error) {

	resp, err := c.client.R().SetResult([]Profile{}).Get(prfl)
	if err != nil {
		return nil, err
	}
	profile := *resp.Result().(*[]Profile)

	return profile, nil

}

func (c *Client) GetFolders() ([]Folder, error) {
	resp, err := c.client.R().SetResult([]Folder{}).Get("rootfolder")
	if err != nil {
		return nil, err
	}

	folders := *resp.Result().(*[]Folder)
	return folders, nil
}

func (c *Client) AddMovie(m Movie, qualityProfile int, path string) (movie Movie, err error) {

	request := AddMovieRequest{
		Title:            m.Title,
		TitleSlug:        m.TitleSlug,
		Images:           m.Images,
		QualityProfileID: qualityProfile,
		TMDBID:           m.TMDBID,
		RootFolderPath:   path,
		Monitored:        true,
		Year:             m.Year,
		AddOptions:       AddMovieOptions{SearchForMovie: true},
	}

	resp, err := c.client.R().SetBody(request).SetResult(Movie{}).Post("movie")
	if err != nil {
		return
	}

	movie = *resp.Result().(*Movie)
	return
}

//func (c *Client) DeleteMovie(movieId int) (err error) {
//	_, err = c.client.R().SetQueryParam("deleteFiles", "true").Delete("movie/" + strconv.Itoa(movieId))
//	return
//}

//func (c *Client) UpdateMovie(m Movie) (movie Movie, err error) {
//	resp, err := c.client.R().SetBody(m).SetResult(Movie{}).Put("movie")
//	if err != nil {
//		return
//	}
//	movie = *resp.Result().(*Movie)
//	return
//}

//func (c *Client) GetMoviesByFolder(folder Folder) (movies []Movie, err error) {
//	allMovies, err := c.GetMovies()
//	if err != nil {
//		return
//	}
//	for _, movie := range allMovies {
//		if strings.HasPrefix(movie.Path, folder.Path) {
//			movies = append(movies, movie)
//		}
//	}
//	return
//}

//func (c *Client) GetMovies() (movies []Movie, err error) {
//	resp, err := c.client.R().SetResult([]Movie{}).Get("movie")
//	if err != nil {
//		return
//	}
//	allMovies := *resp.Result().(*[]Movie)
//	for _, movie := range allMovies {
//		if movie.Monitored {
//			movies = append(movies, movie)
//		}
//	}
//	return
//}

func (c *Client) GetRadarrQueue() ([]RadarrQueue, error) {
	resp, err := c.client.R().SetResult([]RadarrQueue{}).Get("queue")
	if err != nil {
		return nil, err
	}

	queuemovies := *resp.Result().(*[]RadarrQueue)
	return queuemovies, nil
}
