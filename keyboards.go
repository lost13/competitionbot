package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

var mainKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("🎲 Создать конкурс"),
		tgbotapi.NewKeyboardButton("📜 Текущие конкурсы"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("⚙️ Управление каналами"),
	),
)

var cancelBoard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Отменить создание конкурса"),
	),
)

var cancelContinueboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Продолжить без изменений"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Отменить создание конкурса"),
	),
)

var confirmcancelBoard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Изменить название конкурса"),
		tgbotapi.NewKeyboardButton("Изменить текст конкурса"),
	),
	/*tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Изменить текст кнопки"),
		tgbotapi.NewKeyboardButton("Изменить картинку конкурса"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Изменить количество участников"),
		tgbotapi.NewKeyboardButton("Изменить дату окончания"),
	),*/
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Подтвердить"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Отменить создание конкурса"),
	),
)

var channelsBoard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Мои каналы"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Добавить канал"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Главное меню"),
	),
)

var publicationBoard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Опубликовать сейчас"),
	),
	/*tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Запланировать"),
	),*/
)

var competitionButton interface{}
