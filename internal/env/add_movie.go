package env

import (
	"gopkg.in/tucnak/telebot.v2"
)

func (e *Env) HandleAddMovie(m *telebot.Message) {
	e.CM.StartConversation(NewAddMovieConversation(e), m)
}
