package env

import (
	"fmt"
	"strings"

	"github.com/zhark0vv/gramarr/internal/users"

	"github.com/zhark0vv/gramarr/internal/util"

	tb "gopkg.in/telebot.v3"
)

func (e *Env) HandleAuth(m *tb.Message) {
	var msg []string
	pass := m.Payload
	user, exists := e.Users.User(m.Sender.ID)

	// Empty Password?
	if pass == "" {
		util.Send(e.Bot, m.Sender, "Необходимо ввести: `/auth [*ваш пароль*]`")
		return
	}

	// Is User Already Admin?
	if exists && user.IsAdmin() {
		// Notify User
		msg = append(msg, "Че те нада ты уже авторизован.")
		msg = append(msg, "Пиши /start чтобы начать.")
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
		msg = append(msg, "Хозяин ты вернулся. А я че, я ниче")
		msg = append(msg, "Напиши пожалуйста /start чтобы начать.")
		util.Send(e.Bot, m.Sender, strings.Join(msg, "\n"))

		// Notify Admin
		adminMsg := fmt.Sprintf("%s предоставили права админа.", util.DisplayName(m.Sender))
		util.SendAdmin(e.Bot, e.Users.Admins(), adminMsg)

		return
	}

	// Check if pass is User Password
	if pass == e.Config.Bot.Password {
		if exists {
			// Notify User
			msg = append(msg, "Че те нада ты уже авторизован.")
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
		msg = append(msg, "Вы были авторизованы.")
		msg = append(msg, "Напишите /start чтобы начала.")
		util.Send(e.Bot, m.Sender, strings.Join(msg, "\n"))

		// Notify Admin
		adminMsg := fmt.Sprintf("%s был предоставлен доступ.", util.DisplayName(m.Sender))
		util.SendAdmin(e.Bot, e.Users.Admins(), adminMsg)
		return
	}

	// Notify User
	util.SendError(e.Bot, m.Sender, "Твой пароль не верен. Начинаю обратный отсчет до активации режима охраны.")

	// Notify Admin
	adminMsg := "%s сделал не верную попытку войти: %s"
	adminMsg = fmt.Sprintf(adminMsg, util.DisplayName(m.Sender), util.EscapeMarkdown(m.Payload))
	util.SendAdmin(e.Bot, e.Users.Admins(), adminMsg)
}
