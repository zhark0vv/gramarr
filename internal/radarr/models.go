package radarr

import (
	"fmt"
	"time"
)

type Movie struct {
	Title     string       `json:"title"`
	TitleSlug string       `json:"titleSlug"`
	Year      int          `json:"year"`
	PosterURL string       `json:"remotePoster"`
	TMDBID    int          `json:"tmdbId"`
	Images    []MovieImage `json:"images"`
	Path                  string             `json:"path,omitempty"`
	Monitored             bool               `json:"monitored,omitempty"`
}

func (m Movie) String() string {
	if m.Year != 0 {
		return fmt.Sprintf("%s (%d)", m.Title, m.Year)
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
	Movie             			Movie          `json:"movie,omitempty"`
	//Quality          			Quality   	    `json:"quality,omitempty"`
	Size						int64           `json:"sizeOnDisk,omitempty"`
	sizeleft 					int64 		    `json:"sizeleft,omitempty"`
	timeleft 					time.Time 		`json:"timeleft,omitempty"`
	estimatedCompletionTime  time.Time 		`json:"estimatedCompletionTime,omitempty"`
	status 						string 			`json:"status,omitempty"`
	trackedDownloadStatus 	string 			`json:"trackedDownloadStatus,omitempty"`
}
