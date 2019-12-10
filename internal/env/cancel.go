package env

import (
	"fmt"
	"strings"

	"github.com/memodota/gramarr/internal/conversation"
	"github.com/memodota/gramarr/internal/util"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (e *Env) HandleCancel(m *tb.Message) {
	util.Send(e.Bot, m.Sender, "Не нашлось активной операци для отмены, кусок мяса. Я не собирался ничего делать. Аста ла виста...")
}

func (e *Env) HandleConvoCancel(c conversation.Conversation, m *tb.Message) {
	e.CM.StopConversation(c)

	var msg []string
	msg = append(msg, fmt.Sprintf("Команда '*%s*' была отменена. Что тебе еще кожаный у***ок?", c.Name()))
	msg = append(msg, "")
	msg = append(msg, "Отправь /help чтобы прочитать список команд.")
	util.Send(e.Bot, m.Sender, strings.Join(msg, "\n"))
}
