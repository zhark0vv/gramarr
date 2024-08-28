package env

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/zhark0vv/gim/internal/sonarr"
	"github.com/zhark0vv/gim/internal/util"
	"gopkg.in/telebot.v3"
	tb "gopkg.in/telebot.v3"
)

func (e *Env) HandleAddTVShow(m *telebot.Message) {
	e.CM.StartConversation(NewAddTVShowConversation(e), m)
}

func NewAddTVShowConversation(e *Env) *AddTVShowConversation {
	return &AddTVShowConversation{env: e}
}

type AddTVShowConversation struct {
	currentStep             func(*tb.Message)
	tvQuery                 string
	tvShowResults           []sonarr.TVShow
	folderResults           []sonarr.Folder
	releaseResults          []sonarr.Release
	selectedTVShow          *sonarr.TVShow
	selectedTVShowSeasons   []sonarr.TVShowSeason
	selectedQualityProfile  *sonarr.Profile
	selectedLanguageProfile *sonarr.Profile
	selectedFolder          *sonarr.Folder
	selectedRelease         *sonarr.Release
	env                     *Env
	selectedType            string
	tvShowID                int
}

func (c *AddTVShowConversation) Run(m *telebot.Message) {
	payload := util.PayloadFromText(m.Payload)

	i, ok := payload.AsIndex()
	if ok {
		c.tvShowResults, ok = c.env.GetGlobalState(m.Sender.ID, "series").([]sonarr.TVShow)
		if !ok {
			util.SendError(c.env.Bot, m.Sender, "Нет контекста по выбранным сериалам, использование команды невозможно!")
			c.currentStep = c.AskTVShow(m)
			return
		}

		c.selectedTVShow = &c.tvShowResults[i]
		c.currentStep = c.AskPickTVShowSeason(m)
		return
	}

	c.currentStep = c.AskTVShow(m)
}

func (c *AddTVShowConversation) Name() string {
	return "addtv"
}

func (c *AddTVShowConversation) CurrentStep() func(*tb.Message) {
	return c.currentStep
}

func (c *AddTVShowConversation) AskTVShow(m *telebot.Message) func(*tb.Message) {
	util.Send(c.env.Bot, m.Sender, "Введите название сериала")

	return func(m *telebot.Message) {
		c.tvQuery = m.Text

		TVShows, err := c.env.Sonarr.SearchTVShows(c.tvQuery)
		c.tvShowResults = TVShows

		// Search Service Failed
		if err != nil {
			util.SendError(c.env.Bot, m.Sender, "Не удалось организовать поиск.")
			c.env.CM.StopConversation(c)
			return
		}

		// No Results
		if len(TVShows) == 0 {
			msg := fmt.Sprintf("Не нашлось сериала с названием - '%s'", util.EscapeMarkdown(c.tvQuery))
			util.Send(c.env.Bot, m.Sender, msg)
			c.env.CM.StopConversation(c)
			return
		}

		// Found some TVShows! Yay!
		var msg []string
		msg = append(msg, fmt.Sprintf("*найдено %d Сериалов:*", len(TVShows)))
		for i, TV := range TVShows {
			msg = append(msg, fmt.Sprintf("%d) %s", i+1, util.EscapeMarkdown(TV.String())))
		}
		util.Send(c.env.Bot, m.Sender, strings.Join(msg, "\n"))
		c.currentStep = c.AskPickTVShow(m)
	}
}

func (c *AddTVShowConversation) AskPickTVShow(m *telebot.Message) func(*tb.Message) {

	// Send custom reply keyboard
	var options []string
	for _, TVShow := range c.tvShowResults {
		options = append(options, fmt.Sprintf("%s", TVShow))
	}
	options = append(options, "/cancel")
	util.SendKeyboardList(c.env.Bot, m.Sender, "Какой из них скачать?", options)

	return func(m *telebot.Message) {

		// Set the selected TVShow
		for i := range options {
			if m.Text == options[i] {
				c.selectedTVShow = &c.tvShowResults[i]
				break
			}
		}

		// Not a valid TV selection
		if c.selectedTVShow == nil {
			util.SendError(c.env.Bot, m.Sender, "Не правильный выбор.")
			c.currentStep = c.AskPickTVShow(m)
			return
		}

		if c.selectedTVShow.PosterURL != "" {
			photo := &telebot.Photo{File: telebot.FromURL(c.selectedTVShow.PosterURL)}
			c.env.Bot.Send(m.Sender, photo)
		}

		c.currentStep = c.AskPickTVShowSeason(m)
	}
}

func (c *AddTVShowConversation) AskPickTVShowSeason(m *telebot.Message) func(*tb.Message) {
	// Send custom reply keyboard
	var options []string
	if len(c.selectedTVShowSeasons) > 0 {
		options = append(options, "Нет, это все.")
	}
	for _, Season := range c.selectedTVShow.Seasons {
		if len(c.selectedTVShowSeasons) > 0 {
			show := true
			for _, TVShowSeason := range c.selectedTVShowSeasons {
				if TVShowSeason.SeasonNumber == Season.SeasonNumber || TVShowSeason.SeasonNumber == 0 {
					show = false
				}
			}
			if show {
				options = append(options, fmt.Sprintf("%v", Season.SeasonNumber))
			}
		} else {
			options = append(options, fmt.Sprintf("%v", Season.SeasonNumber))
		}
	}
	options = append(options, "/cancel")
	if len(c.selectedTVShowSeasons) > 0 {
		util.SendKeyboardList(c.env.Bot, m.Sender, "Еще?", options)
	} else {
		util.SendKeyboardList(c.env.Bot, m.Sender, "Какой сезон скачать?", options)
	}

	return func(m *telebot.Message) {

		if m.Text == "Нет, это все." {
			for _, selectedTVShowSeason := range c.selectedTVShow.Seasons {
				selectedTVShowSeason.Monitored = false
				for _, chosenSeason := range c.selectedTVShowSeasons {
					if chosenSeason.SeasonNumber == selectedTVShowSeason.SeasonNumber {
						selectedTVShowSeason.Monitored = true
					}
				}
			}
			c.currentStep = c.AskPickTVShowQuality(m)
			return
		}

		// Set the selected TV
		for _, v := range c.selectedTVShow.Seasons {
			if m.Text == strconv.Itoa(v.SeasonNumber) {
				c.selectedTVShowSeasons = append(c.selectedTVShowSeasons, *v)
				break
			}
		}

		// Not a valid TV selection
		if c.selectedTVShowSeasons == nil {
			util.SendError(c.env.Bot, m.Sender, "Не правильный выбор.")
			c.currentStep = c.AskPickTVShowSeason(m)
			return
		}

		c.currentStep = c.AskPickTVShowSeason(m)
	}
}

func (c *AddTVShowConversation) AskPickTVShowQuality(m *telebot.Message) func(*tb.Message) {

	profiles, err := c.env.Sonarr.GetProfile("qualityprofile")

	// GetProfile Service Failed
	if err != nil {
		util.SendError(c.env.Bot, m.Sender, "Не удалось запросить профили качества.")
		c.env.CM.StopConversation(c)
		return nil
	}

	// Send custom reply keyboard
	var options []string
	for _, qualityProfile := range profiles {
		options = append(options, fmt.Sprintf("%v", qualityProfile.Name))
	}
	options = append(options, "/cancel")
	util.SendKeyboardList(c.env.Bot, m.Sender, "С каким качеством искать?", options)

	return func(m *telebot.Message) {
		// Set the selected option
		for i := range options {
			if m.Text == options[i] {
				c.selectedQualityProfile = &profiles[i]
				break
			}
		}

		// Not a valid selection
		if c.selectedQualityProfile == nil {
			util.SendError(c.env.Bot, m.Sender, "Не правильный выбор.")
			c.currentStep = c.AskPickTVShowQuality(m)
			return
		}

		//c.currentStep = c.AskPickTVShowLanguage(m)
		c.currentStep = c.AskFolder(m)
	}
}

func (c *AddTVShowConversation) AskPickTVShowLanguage(m *telebot.Message) func(*tb.Message) {
	languages, err := c.env.Sonarr.GetProfile("languageprofile")

	// GetProfile Service Failed
	if err != nil {
		util.SendError(c.env.Bot, m.Sender, "Не удалось запросить профили языков.")
		c.env.CM.StopConversation(c)
		return nil
	}

	// Send custom reply keyboard
	var options []string
	for _, LanguageProfile := range languages {
		options = append(options, fmt.Sprintf("%v", LanguageProfile.Name))
	}
	options = append(options, "/cancel")
	util.SendKeyboardList(c.env.Bot, m.Sender, "Какой язык искать?", options)

	return func(m *telebot.Message) {
		// Set the selected option
		for i, opt := range options {
			if m.Text == opt {
				c.selectedLanguageProfile = &languages[i]
				break
			}
		}

		// Not a valid selection
		if c.selectedLanguageProfile == nil {
			util.SendError(c.env.Bot, m.Sender, "Не правильный выбор.")
			c.currentStep = c.AskPickTVShowLanguage(m)
			return
		}

		c.currentStep = c.AskFolder(m)
	}
}

func (c *AddTVShowConversation) AskFolder(m *telebot.Message) func(*tb.Message) {

	folders, err := c.env.Sonarr.GetFolders()
	c.folderResults = folders

	// GetFolders Service Failed
	if err != nil {
		util.SendError(c.env.Bot, m.Sender, "Не удалось запросить папки.")
		c.env.CM.StopConversation(c)
		return nil
	}

	// No Results
	if len(folders) == 0 {
		util.SendError(c.env.Bot, m.Sender, "Не нашлось ни одной папки.")
		c.env.CM.StopConversation(c)
		return nil
	}

	// Found folders!
	if len(folders) == 1 {
		c.selectedFolder = &c.folderResults[0]
		util.SendKeyboardList(c.env.Bot, m.Sender,
			fmt.Sprintf("Найдена всего одна папка - %s, выбрал ее", c.selectedFolder.Path),
			[]string{"OK"})
		return func(m *tb.Message) {
			c.currentStep = c.AskSeriesType(m)
		}
	}

	// Send the results
	var msg []string
	msg = append(msg, fmt.Sprintf("*Найдено %d папок:*", len(folders)))
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
	util.SendKeyboardList(c.env.Bot, m.Sender, "В какую папку скачать?", options)

	return func(m *telebot.Message) {
		// Set the selected folder
		for i, opt := range options {
			if m.Text == opt {
				c.selectedFolder = &c.folderResults[i]
				break
			}
		}

		// Not a valid folder selection
		if c.selectedTVShow == nil {
			util.SendError(c.env.Bot, m.Sender, "Неправильный выбор.")
			c.currentStep = c.AskFolder(m)
			return
		}

		c.currentStep = c.AskSeriesType(m)
	}
}

func (c *AddTVShowConversation) AskSeriesType(m *telebot.Message) func(*tb.Message) {
	var options []string
	options = append(options, "anime")
	options = append(options, "standard")
	options = append(options, "daily")
	options = append(options, "/cancel")
	util.SendKeyboardList(c.env.Bot, m.Sender, "Какой тип сериала?", options)

	return func(m *telebot.Message) {
		for i, opt := range options {
			if m.Text == opt {
				c.selectedType = options[i]
				break
			}
		}
		c.AddTVShow(m)
	}
}

func (c *AddTVShowConversation) AddTVShow(m *telebot.Message) {
	show, err := c.env.Sonarr.AddTVShow(*c.selectedTVShow, c.selectedQualityProfile.ID, c.selectedFolder.Path, c.selectedType)

	// Failed to add TV
	if err != nil {
		util.SendError(c.env.Bot, m.Sender, "Не удалось добавить TV.")
		c.env.CM.StopConversation(c)
		return
	}

	if show.ID == 0 {
		util.SendError(c.env.Bot, m.Sender, "Сериал был добавлен ранее, можно удалить и добавить повторно")
		c.currentStep = c.DeleteTVShow(m)
		return
	}

	c.tvShowID = show.ID

	// Notify User
	util.Send(c.env.Bot, m.Sender, "Сериал добавлен!")
	util.Send(c.env.Bot, m.Sender, "Ищу релизы...")

	// Notify Admin
	adminMsg := fmt.Sprintf("%s добавил сериал '%s'", util.DisplayName(m.Sender), util.EscapeMarkdown(c.selectedTVShow.String()))
	util.SendAdmin(c.env.Bot, c.env.Users.Admins(), adminMsg)

	c.currentStep = c.Releases(m)
}

func (c *AddTVShowConversation) DeleteTVShow(m *tb.Message) func(*tb.Message) {
	const (
		agree    = "Удаляем и добавляем повторно"
		disagree = "Не нужно, посмотрю в Sonarr"
	)

	options := []string{
		agree, disagree, "/cancel",
	}
	util.SendKeyboardList(c.env.Bot, m.Sender, "Что делаем с сериалом?", options)

	return func(m *tb.Message) {
		for _, opt := range options {
			if opt == agree {
				tvShows, err := c.env.Sonarr.GetTVShows(c.selectedTVShow.TVDBID)
				if err != nil {
					util.SendError(c.env.Bot, m.Sender, "Не удалось найти фильм для удаления!")
					return
				}
				err = c.env.Sonarr.DeleteTVShow(tvShows[0].ID)
				if err != nil {
					util.SendError(c.env.Bot, m.Sender, "Не удалось удалить фильм!")
					return
				}
				c.AddTVShow(m)
				return
			}
		}
		util.Send(c.env.Bot, m.Sender, "Не удаляем фильм, проверь Sonarr самостоятельно!")
		c.env.CM.StopConversation(c)
	}
}

func (c *AddTVShowConversation) DownloadRelease(m *tb.Message) func(*tb.Message) {
	const (
		torrserver = "Добавить как ссылку на Torrserver и смотреть сейчас"
		sonarr     = "Скачать в папку через клиент Sonarr"
	)

	options := []string{
		torrserver, sonarr, "/cancel",
	}
	util.SendKeyboardList(c.env.Bot, m.Sender, "Как скачиваем релиз?", options)

	return func(m *tb.Message) {
		for _, opt := range options {
			if m.Text == opt {
				switch opt {
				case torrserver:
					err := c.env.Torrserver.AddTorrent(
						c.selectedRelease.GetDownloadURL(
							c.env.Config.Torrserver.TrackerHost,
							c.env.Config.Torrserver.TrackerPort),
						c.selectedTVShow.PosterURL)
					if err != nil {
						util.SendError(c.env.Bot, m.Sender, "Не удалось добавить релиз в Torrserver!")
						return
					}
					util.Send(
						c.env.Bot, m.Sender,
						fmt.Sprintf("Релиз %s успешно добавлен в Torrserver!", c.selectedRelease.Info()))
					c.env.CM.StopConversation(c)
				case sonarr:
					_, err := c.env.Radarr.DownloadRelease(c.selectedRelease.GUID)
					if err != nil {
						util.SendError(c.env.Bot, m.Sender, "Не удалось скачать релиз через Sonarr!")
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
	}
}

func (c *AddTVShowConversation) Releases(m *tb.Message) func(*tb.Message) {
	releases, err := c.env.Sonarr.GetReleases(c.tvShowID, c.selectedTVShowSeasons)
	if err != nil {
		util.SendError(c.env.Bot, m.Sender, "Не удалось получить релизы")
	}

	c.releaseResults = releases
	var options []string
	for _, r := range releases {
		options = append(options, r.Info())
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
			util.SendError(c.env.Bot, m.Sender, "Неправильный выбор релиза, выбери корретно")
			c.currentStep = c.Releases(m)
			return
		}

		c.currentStep = c.DownloadRelease(m)
	}
}
