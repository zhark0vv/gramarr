package radarr

import (
	"fmt"
	"time"
)

type Movie struct {
	ID        int          `json:"id"`
	Title     string       `json:"title"`
	TitleSlug string       `json:"titleSlug"`
	Year      int          `json:"year"`
	PosterURL string       `json:"remotePoster"`
	TMDBID    int          `json:"tmdbId"`
	Ratings   Ratings      `json:"ratings"`
	Images    []MovieImage `json:"images"`
	Path      string       `json:"path,omitempty"`
	Monitored bool         `json:"monitored,omitempty"`
}

type Ratings struct {
	IMDB RatingValue `json:"imdb"`
	TMDB RatingValue `json:"tmdb"`
}

type RatingValue struct {
	Value float64 `json:"value"`
}

func (m Movie) Rating() float64 {
	return (m.Ratings.IMDB.Value + m.Ratings.TMDB.Value) / 2
}

func (m Movie) String() string {
	if m.Year != 0 {
		return fmt.Sprintf("(Фильм: %.2f) %s (%d)",
			m.Rating(),
			m.Title,
			m.Year)
	} else {
		return m.Title
	}
}

type MovieImage struct {
	CoverType string `json:"coverType"`
	URL       string `json:"url"`
}

type Profile struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

type Folder struct {
	Path      string `json:"path"`
	FreeSpace int64  `json:"freeSpace"`
	ID        int    `json:"id"`
}

type AddMovieRequest struct {
	Title             string          `json:"title"`
	TitleSlug         string          `json:"titleSlug"`
	Images            []MovieImage    `json:"images"`
	QualityProfileID  int             `json:"qualityProfileId"`
	LanguageProfileID int             `json:"languageProfileId"`
	TMDBID            int             `json:"tmdbId"`
	RootFolderPath    string          `json:"rootFolderPath"`
	Monitored         bool            `json:"monitored"`
	AddOptions        AddMovieOptions `json:"addOptions"`
	Year              int             `json:"year"`
}

type AddMovieOptions struct {
	SearchForMovie bool `json:"searchForMovie"`
}

type RadarrQueue struct {
	Movie Movie `json:"movie,omitempty"`
	//Quality          			Quality   	    `json:"quality,omitempty"`
	Size                    int64     `json:"sizeOnDisk,omitempty"`
	sizeleft                int64     `json:"sizeleft,omitempty"`
	timeleft                time.Time `json:"timeleft,omitempty"`
	estimatedCompletionTime time.Time `json:"estimatedCompletionTime,omitempty"`
	status                  string    `json:"status,omitempty"`
	trackedDownloadStatus   string    `json:"trackedDownloadStatus,omitempty"`
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
