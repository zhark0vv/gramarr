package env

import (
	"fmt"
	"strings"

	"github.com/tommy647/gramarr/internal/util"

	tb "gopkg.in/tucnak/telebot.v2"
)

func (e *Env) HandleStart(m *tb.Message) {

	user, exists := e.Users.User(m.Sender.ID)

	var msg []string
	msg = append(msg, fmt.Sprintf("Привет, я %s! Используйте эти команды чтобы контролировать меня:", e.Bot.Me.FirstName))

	if !exists {
		msg = append(msg, "")
		msg = append(msg, "/auth [password] - введите пароль в указанном виде, где "[password]" - пароль")
	}

	if exists && user.IsAdmin() {
		msg = append(msg, "")
		msg = append(msg, "*Admin*")
		msg = append(msg, "/users - Список всех пользователей")
	}

	if exists && (user.IsMember() || user.IsAdmin()) {
		msg = append(msg, "")
		msg = append(msg, "*Media*")
		msg = append(msg, "/addmovie - добавить фильм")
		msg = append(msg, "/addtv - добавить ТВ-шоу")
		msg = append(msg, "")
		msg = append(msg, "/cancel - отмена текущей операции")
	}

	util.Send(e.Bot, m.Sender, strings.Join(msg, "\n"))
}
