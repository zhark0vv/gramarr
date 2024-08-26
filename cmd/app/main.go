package main

import (
	"flag"
	"github.com/zhark0vv/gim/internal/torrserver"
	"log"
	"path/filepath"
	"time"

	"github.com/zhark0vv/gim/internal/config"
	"github.com/zhark0vv/gim/internal/conversation"
	"github.com/zhark0vv/gim/internal/env"
	"github.com/zhark0vv/gim/internal/radarr"
	"github.com/zhark0vv/gim/internal/router"
	"github.com/zhark0vv/gim/internal/sonarr"
	"github.com/zhark0vv/gim/internal/users"
	tb "gopkg.in/telebot.v3"
)

// Flags
var configDir = flag.String("configDir", "./config", "config dir for settings and logs")

func main() {
	flag.Parse()

	conf, err := config.LoadConfig(*configDir)
	if err != nil {
		log.Fatalf("failed to load config file: %s", err.Error())
	}

	userPath := filepath.Join(*configDir, "users.json")
	users, err := users.NewUserDB(userPath)
	if err != nil {
		log.Fatalf("failed to load the users db %v", err)
	}

	var rc *radarr.Client
	if conf.Radarr != nil {
		rc, err = radarr.New(*conf.Radarr)
		if err != nil {
			log.Fatalf("failed to create radarr client: %v", err)
		}
	}

	var sn *sonarr.Client
	if conf.Sonarr != nil {
		sn, err = sonarr.New(*conf.Sonarr)
		if err != nil {
			log.Fatalf("failed to create sonarr client: %v", err)
		}
	}

	var tc *torrserver.Client
	if conf.Torrserver != nil {
		tc, err = torrserver.New(*conf.Torrserver)
		if err != nil {
			log.Fatalf("failed to create torrserver client: %v", err)
		}
	}

	cm := conversation.NewConversationManager()
	r := router.NewRouter(cm)

	poller := tb.LongPoller{Timeout: 15 * time.Second}
	bot, err := tb.NewBot(tb.Settings{
		Token:  conf.Telegram.BotToken,
		Poller: &poller,
	})
	if err != nil {
		log.Fatalf("failed to create telegram bot client: %v", err)
	}

	e := &env.Env{
		Config:     conf,
		Bot:        bot,
		Users:      users,
		CM:         cm,
		Radarr:     rc,
		Sonarr:     sn,
		Torrserver: tc,
	}

	setupHandlers(r, e)
	log.Print("Gramarr is up and running. Go call your bot!")
	bot.Start()
}

func setupHandlers(r *router.Router, e *env.Env) {
	// Send all telegram messages to our custom router
	e.Bot.Handle(tb.OnText, r.Route)

	// Commands
	r.HandleFunc("/auth", e.RequirePrivate(e.RequireAuth(users.UANone, e.HandleAuth)))
	r.HandleFunc("/start", e.RequirePrivate(e.RequireAuth(users.UANone, e.HandleStart)))
	r.HandleFunc("/help", e.RequirePrivate(e.RequireAuth(users.UANone, e.HandleStart)))
	r.HandleFunc("/cancel", e.RequirePrivate(e.RequireAuth(users.UANone, e.HandleCancel)))
	r.HandleFunc("/addmovie", e.RequirePrivate(e.RequireAuth(users.UAMember, e.HandleAddMovie)))
	r.HandleFunc("/addtv", e.RequirePrivate(e.RequireAuth(users.UAMember, e.HandleAddTVShow)))
	r.HandleFunc("/users", e.RequirePrivate(e.RequireAuth(users.UAAdmin, e.HandleUsers)))

	// Catchall Command
	r.HandleFallback(e.RequirePrivate(e.RequireAuth(users.UANone, e.HandleFallback)))

	// Conversation Commands
	r.HandleConvoFunc("/cancel", e.HandleConvoCancel)
}
