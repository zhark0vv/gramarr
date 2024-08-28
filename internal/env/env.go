package env

import (
	"fmt"
	"github.com/zhark0vv/gim/internal/torrserver"
	"strings"
	"sync"

	"github.com/zhark0vv/gim/internal/config"
	"github.com/zhark0vv/gim/internal/conversation"
	"github.com/zhark0vv/gim/internal/radarr"
	"github.com/zhark0vv/gim/internal/sonarr"
	"github.com/zhark0vv/gim/internal/users"
	"github.com/zhark0vv/gim/internal/util"
	tb "gopkg.in/telebot.v3"
)

type Env struct {
	Config      *config.Config
	Users       *users.UserDB
	Bot         *tb.Bot
	CM          *conversation.ConversationManager
	Radarr      *radarr.Client
	Sonarr      *sonarr.Client
	Torrserver  *torrserver.Client
	mu          sync.Mutex
	globalState map[string]interface{}
}

func (e *Env) Init() {
	e.globalState = make(map[string]interface{})
}

func (e *Env) RequirePrivate(h func(m *tb.Message)) func(m *tb.Message) {
	return func(m *tb.Message) {
		if !m.Private() {
			return
		}
		h(m)
	}
}

func (e *Env) RequireAuth(access users.UserAccess, h func(m *tb.Message)) func(m *tb.Message) {
	return func(m *tb.Message) {
		user, _ := e.Users.User(m.Sender.ID)
		var msg []string

		// Is Revoked?
		if user.IsRevoked() {
			// Notify User
			msg = append(msg, "Your access has been revoked and you cannot reauthorize.")
			msg = append(msg, "Please reach out to the bot owner for support.")
			util.SendError(e.Bot, m.Sender, strings.Join(msg, "\n"))

			// Notify Admins
			msg = append(msg, fmt.Sprintf("Revoked users %s attempted the following command:", util.DisplayName(m.Sender)))
			msg = append(msg, fmt.Sprintf("`%s`", m.Text))
			util.SendAdmin(e.Bot, e.Users.Admins(), strings.Join(msg, "\n"))
			return
		}

		// Is Not Member?
		isAuthorized := user.IsAdmin() || user.IsMember()
		if !isAuthorized && access != users.UANone {
			// Notify User
			util.SendError(e.Bot, m.Sender, "You are not authorized to use this bot.\n`/auth [password]` to authorize.")

			// Notify Admins
			msg = append(msg, fmt.Sprintf("Unauthorized users %s attempted the following command:", util.DisplayName(m.Sender)))
			msg = append(msg, fmt.Sprintf("`%s`", m.Text))
			util.SendAdmin(e.Bot, e.Users.Admins(), strings.Join(msg, "\n"))
			return
		}

		// Is Non-Admin and requires Admin?
		if !user.IsAdmin() && access == users.UAAdmin {
			// Notify User
			util.SendError(e.Bot, m.Sender, "Only admins can use this command.")

			// Notify Admins
			msg = append(msg, fmt.Sprintf("User %s attempted the following admin command:", util.DisplayName(m.Sender)))
			msg = append(msg, fmt.Sprintf("`%s`", m.Text))
			util.SendAdmin(e.Bot, e.Users.Admins(), strings.Join(msg, "\n"))
			return
		}

		h(m)
	}
}

func (e *Env) SetGlobalState(userID int64, key string, value any) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.globalState[fmt.Sprintf("%s_%d", key, userID)] = value
}

func (e *Env) GetGlobalState(userID int64, key string) any {
	e.mu.Lock()
	defer e.mu.Unlock()
	st := e.globalState[fmt.Sprintf("%s_%d", key, userID)]
	delete(e.globalState, fmt.Sprintf("%s_%d", key, userID))
	return st
}
