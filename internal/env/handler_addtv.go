package env

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/memodota/gramarr/internal/sonarr"
	"github.com/memodota/gramarr/internal/util"
	"gopkg.in/tucnak/telebot.v2"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (e *Env) HandleAddTVShow(m *telebot.Message) {
	e.CM.StartConversation(NewAddTVShowConversation(e), m)
}

func NewAddTVShowConversation(e *Env) *AddTVShowConversation {
	return &AddTVShowConversation{env: e}
}

type AddTVShowConversation struct {
	currentStep             func(*tb.Message)
	TVQuery                 string
	TVShowResults           []sonarr.TVShow
	folderResults           []sonarr.Folder
	selectedTVShow          *sonarr.TVShow
	selectedTVShowSeasons   []sonarr.TVShowSeason
	selectedQualityProfile  *sonarr.Profile
	selectedLanguageProfile *sonarr.Profile
	selectedFolder          *sonarr.Folder
	env                     *Env
	selectedType            string
}

func (c *AddTVShowConversation) Run(m *telebot.Message) {
	c.currentStep = c.AskTVShow(m)
}

func (c *AddTVShowConversation) Name() string {
	return "addtv"
}

func (c *AddTVShowConversation) CurrentStep() func(*tb.Message) {
	return c.currentStep
}

func (c *AddTVShowConversation) AskTVShow(m *telebot.Message) func(*tb.Message) {
	util.Send(c.env.Bot, m.Sender, "Введите название Сериала")

	return func(m *telebot.Message) {
		c.TVQuery = m.Text

		TVShows, err := c.env.Sonarr.SearchTVShows(c.TVQuery)
		c.TVShowResults = TVShows

		// Search Service Failed
		if err != nil {
			util.SendError(c.env.Bot, m.Sender, "Не удалось организовать поиск.")
			c.env.CM.StopConversation(c)
			return
		}

		// No Results
		if len(TVShows) == 0 {
			msg := fmt.Sprintf("Не нашлось сериала с названием - '%s'", util.EscapeMarkdown(c.TVQuery))
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
	for _, TVShow := range c.TVShowResults {
		options = append(options, fmt.Sprintf("%s", TVShow))
	}
	options = append(options, "/cancel")
	util.SendKeyboardList(c.env.Bot, m.Sender, "Какой из них скачать?", options)

	return func(m *telebot.Message) {

		// Set the selected TVShow
		for i := range options {
			if m.Text == options[i] {
				c.selectedTVShow = &c.TVShowResults[i]
				break
			}
		}

		// Not a valid TV selection
		if c.selectedTVShow == nil {
			util.SendError(c.env.Bot, m.Sender, "Не правильный выбор.")
			c.currentStep = c.AskPickTVShow(m)
			return
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
				if TVShowSeason.SeasonNumber == Season.SeasonNumber {
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

	profiles, err := c.env.Sonarr.GetProfile("profile")

	// GetProfile Service Failed
	if err != nil {
		util.SendError(c.env.Bot, m.Sender, "Не удалось запросить профили качества.")
		c.env.CM.StopConversation(c)
		return nil
	}

	// Send custom reply keyboard
	var options []string
	for _, QualityProfile := range profiles {
		options = append(options, fmt.Sprintf("%v", QualityProfile.Name))
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
			util.SendError(c.env.Bot, m.Sender, "Не правильный выбор.")
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
	_, err := c.env.Sonarr.AddTVShow(*c.selectedTVShow, c.selectedQualityProfile.ID, c.selectedFolder.Path, c.selectedType)

	// Failed to add TV
	if err != nil {
		util.SendError(c.env.Bot, m.Sender, "Не удалось добавить TV.")
		c.env.CM.StopConversation(c)
		return
	}

	if c.selectedTVShow.PosterURL != "" {
		photo := &telebot.Photo{File: telebot.FromURL(c.selectedTVShow.PosterURL)}
		c.env.Bot.Send(m.Sender, photo)
	}

	// Notify User
	util.Send(c.env.Bot, m.Sender, "TV Шоу добавлено!")

	// Notify Admin
	adminMsg := fmt.Sprintf("%s добавил ТВ Шоу '%s'", util.DisplayName(m.Sender), util.EscapeMarkdown(c.selectedTVShow.String()))
	util.SendAdmin(c.env.Bot, c.env.Users.Admins(), adminMsg)

	c.env.CM.StopConversation(c)
}
