package env

import (
	"fmt"
	"github.com/zhark0vv/gim/internal/radarr"
	"github.com/zhark0vv/gim/internal/sonarr"
	"golang.org/x/sync/errgroup"
	tb "gopkg.in/telebot.v3"
)

func (e *Env) Inline() func(ctx tb.Context) error {
	return func(ctx tb.Context) error {
		eg := errgroup.Group{}

		var (
			movies    []radarr.Movie
			radarrErr error
		)

		eg.Go(func() error {
			movies, radarrErr = e.Radarr.SearchMovies(ctx.Query().Text)
			if radarrErr != nil {
				return radarrErr
			}
			e.SetGlobalState(ctx.Sender().ID, "movies", movies)
			return nil
		})

		var (
			series   []sonarr.TVShow
			sonarErr error
		)
		eg.Go(func() error {
			series, sonarErr = e.Sonarr.SearchTVShows(ctx.Query().Text)
			if sonarErr != nil {
				return sonarErr
			}
			e.SetGlobalState(ctx.Sender().ID, "series", series)
			return nil
		})

		err := eg.Wait()
		if err != nil {
			return err
		}

		var results []tb.Result
		for i, movie := range movies {
			movieButton := &tb.ArticleResult{
				Title:       movie.String(),
				Text:        fmt.Sprintf("/addmovie %d", i),
				ThumbURL:    movie.PosterURL,
				ThumbHeight: 1200,
				ThumbWidth:  1200,
			}

			results = append(results, movieButton)
		}

		for i, show := range series {
			seriesButton := &tb.ArticleResult{
				Title:       show.String(),
				Text:        fmt.Sprintf("/addtv %d", i),
				ThumbURL:    show.PosterURL,
				ThumbHeight: 1200,
				ThumbWidth:  1200,
			}

			results = append(results, seriesButton)

		}

		return ctx.Answer(&tb.QueryResponse{
			Results:   results,
			CacheTime: 60,
		})
	}
}
