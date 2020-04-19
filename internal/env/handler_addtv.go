package env

import (
	"fmt"
	"path/filepath"
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
	util.Send(c.env.Bot, m.Sender, "What TV Show do you want to search for?")

	return func(m *telebot.Message) {
		c.TVQuery = m.Text

		TVShows, err := c.env.Sonarr.SearchTVShows(c.TVQuery)
		c.TVShowResults = TVShows

		// Search Service Failed
		if err != nil {
			util.SendError(c.env.Bot, m.Sender, "Failed to search TV Show.")
			c.env.CM.StopConversation(c)
			return
		}

		// No Results
		if len(TVShows) == 0 {
			msg := fmt.Sprintf("No TV Show found with the title '%s'", util.EscapeMarkdown(c.TVQuery))
			util.Send(c.env.Bot, m.Sender, msg)
			c.env.CM.StopConversation(c)
			return
		}

		// Found some TVShows! Yay!
		var msg []string
		msg = append(msg, fmt.Sprintf("*Found %d TV Shows:*", len(TVShows)))
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
	util.SendKeyboardList(c.env.Bot, m.Sender, "Which one would you like to download?", options)

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
			util.SendError(c.env.Bot, m.Sender, "Invalid selection.")
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
		options = append(options, "Nope. I'm done!")
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
		util.SendKeyboardList(c.env.Bot, m.Sender, "Any other season?", options)
	} else {
		util.SendKeyboardList(c.env.Bot, m.Sender, "Which season would you like to download?", options)
	}

	return func(m *telebot.Message) {

		if m.Text == "Nope. I'm done!" {
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
		for i := range options {
			if m.Text == options[i] {
				c.selectedTVShowSeasons = append(c.selectedTVShowSeasons, *c.selectedTVShow.Seasons[i])
				break
			}
		}

		// Not a valid TV selection
		if c.selectedTVShowSeasons == nil {
			util.SendError(c.env.Bot, m.Sender, "Invalid selection.")
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
		util.SendError(c.env.Bot, m.Sender, "Failed to get quality profiles.")
		c.env.CM.StopConversation(c)
		return nil
	}

	// Send custom reply keyboard
	var options []string
	for _, QualityProfile := range profiles {
		options = append(options, fmt.Sprintf("%v", QualityProfile.Name))
	}
	options = append(options, "/cancel")
	util.SendKeyboardList(c.env.Bot, m.Sender, "Which quality shall I look for?", options)

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
			util.SendError(c.env.Bot, m.Sender, "Invalid selection.")
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
		util.SendError(c.env.Bot, m.Sender, "Failed to get language profiles.")
		c.env.CM.StopConversation(c)
		return nil
	}

	// Send custom reply keyboard
	var options []string
	for _, LanguageProfile := range languages {
		options = append(options, fmt.Sprintf("%v", LanguageProfile.Name))
	}
	options = append(options, "/cancel")
	util.SendKeyboardList(c.env.Bot, m.Sender, "Which language shall I look for?", options)

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
			util.SendError(c.env.Bot, m.Sender, "Invalid selection.")
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
		util.SendError(c.env.Bot, m.Sender, "Failed to get folders.")
		c.env.CM.StopConversation(c)
		return nil
	}

	// No Results
	if len(folders) == 0 {
		util.SendError(c.env.Bot, m.Sender, "No destination folders found.")
		c.env.CM.StopConversation(c)
		return nil
	}

	// Found folders!

	// Send the results
	var msg []string
	msg = append(msg, fmt.Sprintf("*Found %d folders:*", len(folders)))
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
	util.SendKeyboardList(c.env.Bot, m.Sender, "Which folder should it download to?", options)

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
			util.SendError(c.env.Bot, m.Sender, "Invalid selection.")
			c.currentStep = c.AskFolder(m)
			return
		}

		c.AddTVShow(m)
	}
}

func (c *AddTVShowConversation) AddTVShow(m *telebot.Message) {
	_, err := c.env.Sonarr.AddTVShow(*c.selectedTVShow, c.selectedQualityProfile.ID, c.selectedFolder.Path)

	// Failed to add TV
	if err != nil {
		util.SendError(c.env.Bot, m.Sender, "Failed to add TV.")
		c.env.CM.StopConversation(c)
		return
	}

	if c.selectedTVShow.PosterURL != "" {
		photo := &telebot.Photo{File: telebot.FromURL(c.selectedTVShow.PosterURL)}
		c.env.Bot.Send(m.Sender, photo)
	}

	// Notify User
	util.Send(c.env.Bot, m.Sender, "TV Show has been added!")

	// Notify Admin
	adminMsg := fmt.Sprintf("%s added TV Show '%s'", util.DisplayName(m.Sender), util.EscapeMarkdown(c.selectedTVShow.String()))
	util.SendAdmin(c.env.Bot, c.env.Users.Admins(), adminMsg)

	c.env.CM.StopConversation(c)
}
