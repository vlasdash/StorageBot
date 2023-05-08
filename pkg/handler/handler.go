package handler

import (
	"fmt"
	"github.com/vlasdash/passwordbot/config"
	"log"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/vlasdash/passwordbot/internal/credential"
)

const (
	getCommand   = "get"
	setCommand   = "set"
	delCommand   = "del"
	startCommand = "start"
)

type Handler struct {
	Bot             *tgbotapi.BotAPI
	CredentialRepo  credential.CredentialRepo
	Coder           credential.PasswordCoder
	commandHandlers map[string]func(tgbotapi.Update)
}

func NewHandler(bot *tgbotapi.BotAPI, cr credential.CredentialRepo, coder credential.PasswordCoder) *Handler {
	h := &Handler{
		Bot:            bot,
		CredentialRepo: cr,
		Coder:          coder,
	}

	h.commandHandlers = map[string]func(tgbotapi.Update){
		getCommand:   h.GetLoginCredentials,
		setCommand:   h.SetLoginCredentials,
		delCommand:   h.DeleteLoginCredentials,
		startCommand: h.StartCommand,
	}

	return h
}

func (h *Handler) HandleCommand(update tgbotapi.Update) {
	handler, ok := h.commandHandlers[update.Message.Command()]

	if !ok {
		h.UnknownCommand(update)
		return
	}

	handler(update)

}

func (h *Handler) StartCommand(update tgbotapi.Update) {
	_, err := h.Bot.Send(tgbotapi.NewMessage(
		update.Message.Chat.ID,
		"Добро пожаловать! Я сохраню все Ваши логины и пароли от всех сервисов!",
	))
	if err != nil {
		log.Printf("error at unknown command in send message to bot: %v\n", err)
	}
}

func (h *Handler) UnknownCommand(update tgbotapi.Update) {
	_, err := h.Bot.Send(tgbotapi.NewMessage(
		update.Message.Chat.ID,
		"Неизвестная команда",
	))
	if err != nil {
		log.Printf("error at unknown command in send message to bot: %v\n", err)
	}
}

func (h *Handler) SetLoginCredentials(update tgbotapi.Update) {
	arg := strings.Fields(update.Message.CommandArguments())
	if len(arg) < 3 {
		_, err := h.Bot.Send(tgbotapi.NewMessage(
			update.Message.Chat.ID,
			"Введите, пожалуйста, название сервера, логин и пароль через пробел.",
		))
		if err != nil {
			log.Printf("error at save credential in send message to bot: %v\n", err)
		}
		return
	}

	passwordEncrypt, err := h.Coder.Encrypt(arg[2])
	if err != nil {
		log.Printf("error at save credential in encrypt password: %v\n", err)
		_, err = h.Bot.Send(tgbotapi.NewMessage(
			update.Message.Chat.ID,
			"Не удалось сохранить данные. Повторите запрос позже",
		))
		if err != nil {
			log.Printf("error at save credential in send message to bot: %v\n", err)
		}
		return
	}
	err = h.CredentialRepo.Add(
		update.Message.From.ID,
		arg[0],
		arg[1],
		passwordEncrypt,
	)
	if err == credential.ErrAlreadyExist {
		_, err = h.Bot.Send(tgbotapi.NewMessage(
			update.Message.Chat.ID,
			fmt.Sprintf("Данные для %s уже сохранены в хранилище", arg[0]),
		))
		if err != nil {
			log.Printf("error at save credential in send message to bot: %v\n", err)
		}
		return
	}
	if err != nil {
		log.Printf("error at save credential: %v\n", err)
		_, err = h.Bot.Send(tgbotapi.NewMessage(
			update.Message.Chat.ID,
			"Не удалось сохранить данные. Повторите запрос позже",
		))
		if err != nil {
			log.Printf("error at save credential in send message to bot: %v\n", err)
		}
		return
	}

	response := fmt.Sprintf("Ваши учётные данные с логином %s для сервера %s успешно сохранены", arg[1], arg[0])
	_, err = h.Bot.Send(tgbotapi.NewMessage(
		update.Message.Chat.ID,
		response,
	))
	if err != nil {
		log.Printf("error at save credential in send message to bot: %v\n", err)
		return
	}

	go func(chatID int64, messageID int) {
		time.Sleep(time.Duration(config.C.App.PasswordRetentionMinute) * time.Minute)

		params := tgbotapi.Params{}
		params.AddNonZero("chat_id", int(chatID))
		params.AddNonZero("message_id", messageID)

		_, err = h.Bot.MakeRequest("deleteMessage", params)
		if err != nil {
			log.Printf("error at delete message with password: %v\n", err)
		}
	}(update.Message.Chat.ID, update.Message.MessageID)
}

func (h *Handler) GetLoginCredentials(update tgbotapi.Update) {
	if len(strings.TrimSpace(update.Message.CommandArguments())) == 0 {
		_, err := h.Bot.Send(tgbotapi.NewMessage(
			update.Message.Chat.ID,
			"Введите, пожалуйста, название сервера.",
		))
		if err != nil {
			log.Printf("error at get credential in send message to bot: %v\n", err)
		}
		return
	}

	credit, err := h.CredentialRepo.GetByServerName(
		update.Message.From.ID,
		update.Message.CommandArguments(),
	)
	if err == credential.ErrNoExist || err == credential.ErrNoAccess {
		_, err = h.Bot.Send(tgbotapi.NewMessage(
			update.Message.Chat.ID,
			fmt.Sprintf(
				"В системе не сохранены учетные данные для сервера %s",
				update.Message.CommandArguments(),
			),
		))
		if err != nil {
			log.Printf("error at get credentials by server name in send message to bot: %v\n", err)
		}
		return
	}
	if err != nil {
		log.Printf("error at get credentials by server name: %v\n", err)
		_, err = h.Bot.Send(tgbotapi.NewMessage(
			update.Message.Chat.ID,
			"Не удалось получить данные. Повторите запрос позже",
		))
		if err != nil {
			log.Printf("error at get credentials by server name in send message to bot: %v\n", err)
		}
		return
	}
	credit.Password, err = h.Coder.Decrypt(credit.Password)
	if err != nil {
		log.Printf("error at get credential in decrypt password: %v\n", err)
		_, err = h.Bot.Send(tgbotapi.NewMessage(
			update.Message.Chat.ID,
			"Не удалось получить данные. Повторите запрос позже",
		))
		if err != nil {
			log.Printf("error at get credential in send message to bot: %v\n", err)
		}
		return
	}

	message := fmt.Sprintf("Логин: %s\nПароль: %s", credit.Login, credit.Password)
	response, err := h.Bot.Send(tgbotapi.NewMessage(
		update.Message.Chat.ID,
		message,
	))
	if err != nil {
		log.Printf("error at get credentials by server name in send message to bot: %v\n", err)
		return
	}

	newText := fmt.Sprintf("Логин: %s\n", credit.Login)
	go func(newText string, chatID int64, messageID int) {
		time.Sleep(time.Duration(config.C.App.PasswordRetentionMinute) * time.Minute)

		msg := tgbotapi.NewEditMessageText(chatID, messageID, newText)
		_, err = h.Bot.Send(msg)
		if err != nil {
			log.Printf("error at delete message with password: %v\n", err)
		}
	}(newText, update.Message.Chat.ID, response.MessageID)
}

func (h *Handler) DeleteLoginCredentials(update tgbotapi.Update) {
	if len(strings.TrimSpace(update.Message.CommandArguments())) == 0 {
		_, err := h.Bot.Send(tgbotapi.NewMessage(
			update.Message.Chat.ID,
			"Введите, пожалуйста, название сервера.",
		))
		if err != nil {
			log.Printf("error at get credential in send message to bot: %v\n", err)
		}
		return
	}

	_, err := h.CredentialRepo.Delete(
		update.Message.From.ID,
		update.Message.CommandArguments(),
	)

	if err == credential.ErrNoExist || err == credential.ErrNoAccess {
		_, err = h.Bot.Send(tgbotapi.NewMessage(
			update.Message.Chat.ID,
			fmt.Sprintf(
				"В системе не сохранены учетные данные для сервера %s",
				update.Message.CommandArguments(),
			),
		))
		if err != nil {
			log.Printf("error at delete credentials by server name in send message to bot: %v\n", err)
		}
		return
	}
	if err != nil {
		log.Printf("error at get credentials by server name: %v\n", err)
		_, err = h.Bot.Send(tgbotapi.NewMessage(
			update.Message.Chat.ID,
			"Произошла ошибка при попытке получения доступа к данным. Повторите запрос позже",
		))
		if err != nil {
			log.Printf("error at delete credentials by server name in send message to bot: %v\n", err)
		}
		return
	}

	_, err = h.Bot.Send(tgbotapi.NewMessage(
		update.Message.Chat.ID,
		"Учетные данные успешно удалены из хранилища",
	))
	if err != nil {
		log.Printf("error at get credentials by server name in send message to bot: %v\n", err)
	}
}
