package env

import (
	"fmt"
	"strings"

	"github.com/tommy647/gramarr/internal/users"

	"github.com/tommy647/gramarr/internal/util"

	tb "gopkg.in/tucnak/telebot.v2"
)

func (e *Env) HandleAuth(m *tb.Message) {
	var msg []string
	pass := m.Payload
	user, exists := e.Users.User(m.Sender.ID)

	// Empty Password?
	if pass == "" {
		util.Send(e.Bot, m.Sender, "Usage: `/auth [password]`")
		return
	}

	// Is User Already Admin?
	if exists && user.IsAdmin() {
		// Notify User
		msg = append(msg, "You're already authorized.")
		msg = append(msg, "Type /start to begin.")
		util.Send(e.Bot, m.Sender, strings.Join(msg, "\n"))
		return
	}

	// Check if pass is Admin Password
	if pass == e.Config.Bot.AdminPassword {
		if exists {
			user.Access = users.UAAdmin
			e.Users.Update(user)
		} else {
			newUser := users.User{
				ID:        m.Sender.ID,
				FirstName: m.Sender.FirstName,
				LastName:  m.Sender.LastName,
				Username:  m.Sender.Username,
				Access:    users.UAAdmin,
			}
			e.Users.Create(newUser)
		}

		// Notify User
		msg = append(msg, "You have been authorized as an *admin*.")
		msg = append(msg, "Type /start to begin.")
		util.Send(e.Bot, m.Sender, strings.Join(msg, "\n"))

		// Notify Admin
		adminMsg := fmt.Sprintf("%s has been granted admin access.", util.DisplayName(m.Sender))
		util.SendAdmin(e.Bot, e.Users.Admins(), adminMsg)

		return
	}

	// Check if pass is User Password
	if pass == e.Config.Bot.Password {
		if exists {
			// Notify User
			msg = append(msg, "You're already authorized.")
			msg = append(msg, "Type /start to begin.")
			util.Send(e.Bot, m.Sender, strings.Join(msg, "\n"))
			return
		}
		newUser := users.User{
			ID:        m.Sender.ID,
			Username:  m.Sender.Username,
			FirstName: m.Sender.FirstName,
			LastName:  m.Sender.LastName,
			Access:    users.UAMember,
		}
		e.Users.Create(newUser)

		// Notify User
		msg = append(msg, "You have been authorized.")
		msg = append(msg, "Type /start to begin.")
		util.Send(e.Bot, m.Sender, strings.Join(msg, "\n"))

		// Notify Admin
		adminMsg := fmt.Sprintf("%s has been granted acccess.", util.DisplayName(m.Sender))
		util.SendAdmin(e.Bot, e.Users.Admins(), adminMsg)
		return
	}

	// Notify User
	util.SendError(e.Bot, m.Sender, "Your password is invalid.")

	// Notify Admin
	adminMsg := "%s made an invalid auth request with password: %s"
	adminMsg = fmt.Sprintf(adminMsg, util.DisplayName(m.Sender), util.EscapeMarkdown(m.Payload))
	util.SendAdmin(e.Bot, e.Users.Admins(), adminMsg)
}
