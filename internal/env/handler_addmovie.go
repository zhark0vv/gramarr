package env

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/zhark0vv/gim/internal/radarr"
	"github.com/zhark0vv/gim/internal/util"
	"gopkg.in/telebot.v3"
	tb "gopkg.in/telebot.v3"
)

func (e *Env) HandleAddMovie(m *telebot.Message) {
	e.CM.StartConversation(NewAddMovieConversation(e), m)
}

func NewAddMovieConversation(e *Env) *AddMovieConversation {
	return &AddMovieConversation{env: e}
}

type AddMovieConversation struct {
	currentStep            func(*tb.Message)
	movieQuery             string
	movieResults           []radarr.Movie
	folderResults          []radarr.Folder
	selectedMovie          *radarr.Movie
	selectedQualityProfile *radarr.Profile
	selectedFolder         *radarr.Folder
	releaseResults         []radarr.Release
	selectedRelease        *radarr.Release
	env                    *Env
	movieID                int
}

func (c *AddMovieConversation) Run(m *tb.Message) {
	c.currentStep = c.AskMovie(m)
}

func (c *AddMovieConversation) Name() string {
	return "addmovie"
}

func (c *AddMovieConversation) CurrentStep() func(*telebot.Message) {
	return c.currentStep
}

func (c *AddMovieConversation) AskMovie(m *tb.Message) func(*telebot.Message) {
	util.Send(c.env.Bot, m.Sender, "Какой фильм смотрим сегодня?")

	return func(m *tb.Message) {
		c.movieQuery = m.Text

		movies, err := c.env.Radarr.SearchMovies(c.movieQuery)
		c.movieResults = movies

		// Search Service Failed
		if err != nil {
			util.SendError(c.env.Bot, m.Sender, "Поиск фильма не удался.")
			c.env.CM.StopConversation(c)
			return
		}

		// No Results
		if len(movies) == 0 {
			msg := fmt.Sprintf("Фильма с указанным названием - '%s', не нашлось", util.EscapeMarkdown(c.movieQuery))
			util.Send(c.env.Bot, m.Sender, msg)
			c.env.CM.StopConversation(c)
			return
		}

		// Found some movies! Yay!
		var msg []string
		msg = append(msg, fmt.Sprintf("*Нашлось %d фильмов:*", len(movies)))
		for i, movie := range movies {
			msg = append(msg, fmt.Sprintf("%d) %s", i+1, util.EscapeMarkdown(movie.String())))
		}
		util.Send(c.env.Bot, m.Sender, strings.Join(msg, "\n"))
		c.currentStep = c.AskPickMovie(m)
	}
}

func (c *AddMovieConversation) AskPickMovie(m *tb.Message) func(*telebot.Message) {

	// Send custom reply keyboard
	var options []string
	for _, movie := range c.movieResults {
		options = append(options, fmt.Sprintf("%s", movie))
	}
	options = append(options, "/cancel")
	util.SendKeyboardList(c.env.Bot, m.Sender, "Какой из них скачать?", options)

	return func(m *tb.Message) {

		// Set the selected movie
		for i, opt := range options {
			if m.Text == opt {
				c.selectedMovie = &c.movieResults[i]
				break
			}
		}

		// Not a valid movie selection
		if c.selectedMovie == nil {
			util.SendError(c.env.Bot, m.Sender, "Неправильный выбор.")
			c.currentStep = c.AskPickMovie(m)
			return
		}

		if c.selectedMovie.PosterURL != "" {
			photo := &tb.Photo{File: tb.FromURL(c.selectedMovie.PosterURL)}
			c.env.Bot.Send(m.Sender, photo)
		}

		c.currentStep = c.AskPickMovieQuality(m)
	}
}

func (c *AddMovieConversation) AskPickMovieQuality(m *tb.Message) func(*telebot.Message) {
	profiles, err := c.env.Radarr.GetProfile()

	// GetProfile Service Failed
	if err != nil {
		util.SendError(c.env.Bot, m.Sender, "Не удалось получить профили качества.")
		c.env.CM.StopConversation(c)
		return nil
	}

	// Send custom reply keyboard
	var options []string
	for _, qualityProfile := range profiles {
		options = append(options, fmt.Sprintf("%v", qualityProfile.Name))
	}
	options = append(options, "/cancel")
	util.SendKeyboardList(c.env.Bot, m.Sender, "В каком качестве искать фильм?", options)

	return func(m *tb.Message) {
		// Set the selected option
		for i := range options {
			if m.Text == options[i] {
				c.selectedQualityProfile = &profiles[i]
				break
			}
		}

		// Not a valid selection
		if c.selectedQualityProfile == nil {
			util.SendError(c.env.Bot, m.Sender, "Неправильный выбор.")
			c.currentStep = c.AskPickMovieQuality(m)
			return
		}

		c.currentStep = c.AskFolder(m)
	}
}

func (c *AddMovieConversation) AskFolder(m *tb.Message) func(*telebot.Message) {

	folders, err := c.env.Radarr.GetFolders()
	c.folderResults = folders

	// GetFolders Service Failed
	if err != nil {
		util.SendError(c.env.Bot, m.Sender, "Не удалось найти папки.")
		c.env.CM.StopConversation(c)
		return nil
	}

	// No Results
	if len(folders) == 0 {
		util.SendError(c.env.Bot, m.Sender, "Не удалось найти папки назначения.")
		c.env.CM.StopConversation(c)
		return nil
	}

	// Found folders!

	// Send the results
	var msg []string
	msg = append(msg, fmt.Sprintf("*Нашел %d папок:*", len(folders)))
	for i, folder := range folders {
		msg = append(msg, fmt.Sprintf("%d) %s", i+1, util.EscapeMarkdown(filepath.Base(folder.Path))))
	}
	util.Send(c.env.Bot, m.Sender, strings.Join(msg, "\n"))

	// Send the custom reply keyboard
	var options []string
	for _, folder := range folders {
		options = append(options, fmt.Sprintf("%s", filepath.Base(folder.Path)))
	}
	options = append(options, "/cancel")
	util.SendKeyboardList(c.env.Bot, m.Sender, "В какую папку скачать фильм?", options)

	return func(m *tb.Message) {
		// Set the selected folder
		for i, opt := range options {
			if m.Text == opt {
				c.selectedFolder = &c.folderResults[i]
				break
			}
		}

		// Not a valid folder selection
		if c.selectedMovie == nil {
			util.SendError(c.env.Bot, m.Sender, "Не правильный выбор.")
			c.currentStep = c.AskFolder(m)
			return
		}

		c.AddMovie(m)
	}
}

func (c *AddMovieConversation) AddMovie(m *tb.Message) {
	movie, err := c.env.Radarr.AddMovie(*c.selectedMovie, c.selectedQualityProfile.ID, c.selectedFolder.Path)

	// Failed to add movie
	if err != nil {
		util.SendError(c.env.Bot, m.Sender, "Не удалось добавить фильм.")
		c.env.CM.StopConversation(c)
		return
	}

	if movie.ID == 0 {
		util.SendError(c.env.Bot, m.Sender, "Фильм был добавлен ранее, можно удалить и добавить повторно")
		c.currentStep = c.DeleteMovie(m)
		return
	}

	c.movieID = movie.ID

	// Notify Admin
	adminMsg := fmt.Sprintf("%s добавил фильм '%s'", util.DisplayName(m.Sender), util.EscapeMarkdown(c.selectedMovie.String()))
	util.SendAdmin(c.env.Bot, c.env.Users.Admins(), adminMsg)

	c.currentStep = c.Releases(m)
}

func (c *AddMovieConversation) DeleteMovie(m *tb.Message) func(*tb.Message) {
	const (
		agree    = "Удаляем и добавляем повторно"
		disagree = "Не нужно, посмотрю в Radarr"
	)

	options := []string{
		agree, disagree, "/cancel",
	}
	util.SendKeyboardList(c.env.Bot, m.Sender, "Что делаем с фильмом?", options)

	return func(m *tb.Message) {
		for _, opt := range options {
			if opt == agree {
				movies, err := c.env.Radarr.GetMovies(c.selectedMovie.TMDBID)
				if err != nil {
					util.SendError(c.env.Bot, m.Sender, "Не удалось найти фильм для удаления!")
					return
				}
				err = c.env.Radarr.DeleteMovie(movies[0].ID)
				if err != nil {
					util.SendError(c.env.Bot, m.Sender, "Не удалось удалить фильм!")
					return
				}
				c.AddMovie(m)
				return
			}
		}
		util.Send(c.env.Bot, m.Sender, "Не удаляем фильм, проверь Radarr самостоятельно!")
		c.env.CM.StopConversation(c)
	}
}

func (c *AddMovieConversation) DownloadRelease(m *tb.Message) func(*tb.Message) {
	const (
		torrserver = "Добавить как ссылку на Torrserver и смотреть сейчас"
		radarr     = "Скачать в папку через клиент Radarr"
	)

	options := []string{
		torrserver, radarr, "/cancel",
	}
	util.SendKeyboardList(c.env.Bot, m.Sender, "Как скачиваем релиз?", options)

	return func(m *tb.Message) {
		for _, opt := range options {
			if m.Text == opt {
				switch opt {
				case torrserver:
				case radarr:
					_, err := c.env.Radarr.DownloadRelease(c.selectedRelease.GUID)
					if err != nil {
						util.SendError(c.env.Bot, m.Sender, "Не удалось скачать релиз через Radarr!")
						return
					}
					util.Send(
						c.env.Bot, m.Sender,
						fmt.Sprintf("Релиз %s успешно скачан через Radarr!", c.selectedRelease.Info()))
					c.env.CM.StopConversation(c)
				}
				break
			}
		}

		// Not a valid folder selection
		if c.selectedRelease == nil {
			util.SendError(c.env.Bot, m.Sender, "Неправильный выбор, выбери корретно")
			c.currentStep = c.DownloadRelease(m)
			return
		}
	}
}

func (c *AddMovieConversation) Releases(m *tb.Message) func(*tb.Message) {
	releases, err := c.env.Radarr.GetReleases(c.movieID)
	if err != nil {
		util.SendError(c.env.Bot, m.Sender, "Не удалось получить релизы")
	}

	c.releaseResults = releases
	var options []string
	for _, r := range releases {
		options = append(options, fmt.Sprintf("%s", filepath.Base(r.Info())))
	}
	options = append(options, "/cancel")
	util.SendKeyboardList(c.env.Bot, m.Sender, "Список релизов  ", options)

	return func(m *tb.Message) {
		// Set the selected folder
		for i, opt := range options {
			if m.Text == opt {
				c.selectedRelease = &c.releaseResults[i]
				break
			}
		}

		// Not a valid folder selection
		if c.selectedRelease == nil {
			util.SendError(c.env.Bot, m.Sender, "Неправильный выбор, выбери корретно")
			c.currentStep = c.Releases(m)
			return
		}

		c.currentStep = c.DownloadRelease(m)
	}
}
