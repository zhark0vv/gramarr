package sonarr

import (
	"fmt"
	"net/url"
)

type TVShow struct {
	ID         int             `json:"ID"`
	Title      string          `json:"title"`
	TitleSlug  string          `json:"titleSlug"`
	Year       int             `json:"year"`
	PosterURL  string          `json:"remotePoster"`
	TVDBID     int             `json:"tvdbId"`
	Ratings    RatingValue     `json:"ratings"`
	Images     []TVShowImage   `json:"images"`
	Seasons    []*TVShowSeason `json:"seasons"`
	SeriesType string          `json:"seriesType"`
}

type RatingValue struct {
	Value float64 `json:"value"`
}

func (s TVShow) String() string {
	if s.Year != 0 {
		return fmt.Sprintf("(Сериал: %.2f) %s (%d)",
			s.Rating(),
			s.Title,
			s.Year)
	} else {
		return s.Title
	}
}

func (s TVShow) Rating() float64 {
	return s.Ratings.Value
}

type TVShowImage struct {
	CoverType string `json:"coverType"`
	URL       string `json:"url"`
}

type TVShowSeason struct {
	SeasonNumber int  `json:"seasonNumber"`
	Monitored    bool `json:"monitored"`
}

type Folder struct {
	Path      string `json:"path"`
	FreeSpace int64  `json:"freeSpace"`
	ID        int    `json:"id"`
}

type Profile struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

type AddTVShowRequest struct {
	Title             string           `json:"title"`
	TitleSlug         string           `json:"titleSlug"`
	Images            []TVShowImage    `json:"images"`
	QualityProfileID  int              `json:"qualityProfileId"`
	LanguageProfileID int              `json:"languageProfileId"`
	TVDBID            int              `json:"tvdbId"`
	RootFolderPath    string           `json:"rootFolderPath"`
	Monitored         bool             `json:"monitored"`
	AddOptions        AddTVShowOptions `json:"addOptions"`
	Year              int              `json:"year"`
	Seasons           []*TVShowSeason  `json:"seasons"`
	SeriesType        string           `json:"seriesType"`
}

type AddTVShowOptions struct {
	SearchForMissingEpisodes   bool `json:"searchForMissingEpisodes"`
	IgnoreEpisodesWithFiles    bool `json:"ignoreEpisodesWithFiles"`
	IgnoreEpisodesWithoutFiles bool `json:"ignoreEpisodesWithoutFiles"`
}

type Release struct {
	GUID        string `json:"guid"`
	Title       string `json:"title"`
	DownloadURL string `json:"downloadUrl"`
	InfoURL     string `json:"infoUrl"`
	Size        int64  `json:"size"`
}

func (r Release) Info() string {
	if len(r.Title) > 100 {
		r.Title = fmt.Sprintf("%s...", r.Title[:100])
	}
	return fmt.Sprintf("%s (%s)", r.Title, r.size())
}

func (r Release) size() string {
	sizeMB := float64(r.Size) / 1024 / 1024
	if sizeMB >= 1024 {
		return fmt.Sprintf("%.1f GB", sizeMB/1024)
	}
	return fmt.Sprintf("%.1f MB", sizeMB)
}

func (r Release) GetDownloadURL(replaceURL *string, replacePort *int) string {
	u, err := url.Parse(r.DownloadURL)
	if err != nil {
		return r.DownloadURL
	}

	if replaceURL != nil && replacePort != nil {
		u.Host = fmt.Sprintf("%s:%d", *replaceURL, *replacePort)
	}

	return u.String()
}
