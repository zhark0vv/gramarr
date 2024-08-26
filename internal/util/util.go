package util

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/zhark0vv/gim/internal/users"
	tb "gopkg.in/telebot.v3"
)

func Send(bot *tb.Bot, to tb.Recipient, msg string) {
	bot.Send(to, msg, tb.ModeMarkdown)
}

func SendError(bot *tb.Bot, to tb.Recipient, msg string) {
	bot.Send(to, msg, tb.ModeMarkdown)
}

func SendAdmin(bot *tb.Bot, to []users.User, msg string) {
	SendMany(bot, to, fmt.Sprintf("*[Admin]* %s", msg))
}

func SendKeyboardList(bot *tb.Bot, to tb.Recipient, msg string, list []string) {
	var buttons []tb.ReplyButton
	for _, item := range list {
		buttons = append(buttons, tb.ReplyButton{Text: item})
	}

	var replyKeys [][]tb.ReplyButton
	for _, b := range buttons {
		replyKeys = append(replyKeys, []tb.ReplyButton{b})
	}

	bot.Send(to, msg, &tb.ReplyMarkup{
		ReplyKeyboard:   replyKeys,
		OneTimeKeyboard: true,
	})
}

func SendMany(bot *tb.Bot, to []users.User, msg string) {
	for _, user := range to {
		bot.Send(user, msg, tb.ModeMarkdown)
	}
}

func DisplayName(u *tb.User) string {
	if u.FirstName != "" && u.LastName != "" {
		return EscapeMarkdown(fmt.Sprintf("%s %s", u.FirstName, u.LastName))
	}

	return EscapeMarkdown(u.FirstName)
}

func EscapeMarkdown(s string) string {
	s = strings.Replace(s, "[", "\\[", -1)
	s = strings.Replace(s, "]", "\\]", -1)
	s = strings.Replace(s, "_", "\\_", -1)
	return s
}

func BoolToYesOrNo(condition bool) string {
	if condition {
		return "Yes"
	}
	return "No"
}

func FormatDate(t time.Time) string {
	if t.IsZero() {
		return "Unknown"
	}
	return t.Format("02.01.2006")
}

func FormatDateTime(t time.Time) string {
	if t.IsZero() {
		return "Unknown"
	}
	return t.Format("02.01.2006 15:04:05")
}

func GetRootFolderFromPath(path string) string {
	return strings.Title(filepath.Base(filepath.Dir(path)))
}
