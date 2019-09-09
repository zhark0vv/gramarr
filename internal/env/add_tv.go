package env

import (
	"gopkg.in/tucnak/telebot.v2"
)

func (e *Env) HandleAddTVShow(m *telebot.Message) {
	e.CM.StartConversation(NewAddTVShowConversation(e), m)
}
