package env

import (
	"fmt"
	"strings"

	"github.com/zhark0vv/gramarr/internal/conversation"
	"github.com/zhark0vv/gramarr/internal/util"
	tb "gopkg.in/telebot.v3"
)

func (e *Env) HandleCancel(m *tb.Message) {
	util.Send(e.Bot, m.Sender, "Не нашлось активной операци для отмены")
}

func (e *Env) HandleConvoCancel(c conversation.Conversation, m *tb.Message) {
	var cancelkeyboard []string
	cancelkeyboard = append(cancelkeyboard, "/help")
	util.SendKeyboardList(e.Bot, m.Sender, "", cancelkeyboard)

	var msg []string
	msg = append(msg, fmt.Sprintf("Команда '*%s*' была отменена.", c.Name()))
	msg = append(msg, "")
	msg = append(msg, "Отправь /help чтобы прочитать список команд.")
	util.Send(e.Bot, m.Sender, strings.Join(msg, "\n"))

	e.CM.StopConversation(c)
}
