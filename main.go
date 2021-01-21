package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jmoiron/sqlx"
	"log"
	"strconv"
	"strings"
	"time"
)

func main() {

	bot, err := getAPI()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	// u - структура с конфигом для получения апдейтов
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// используя конфиг u создаем канал в который будут прилетать новые сообщения
	updates := bot.GetUpdatesChan(u)

	db, err := sqlx.Connect("mysql", Config.Db.User+":"+Config.Db.Password+"@tcp("+Config.Db.Host+":"+Config.Db.Port+")/"+Config.Db.Name+"?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		log.Panic(err)
	}

	log.Printf("MySQL %v", db.Stats())

	if Config.Db.Drop == true {
		_, err = db.Exec("DROP TABLE IF EXISTS `cbotchannels`")
		_, err = db.Exec("DROP TABLE IF EXISTS `competitions`")
		_, err = db.Exec("DROP TABLE IF EXISTS `participants`")
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS `cbotchannels` (`id` int(11) AUTO_INCREMENT NOT NULL PRIMARY KEY, `owner` int(22) NOT NULL, `channelid` bigint(22) NOT NULL, `channelname` varchar(128) NOT NULL, `channeltitle` varchar(128) NOT NULL)")
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS `competitions` (`id` int(11) AUTO_INCREMENT NOT NULL PRIMARY KEY, `owner` int(22) NOT NULL, `channel` bigint(22) NOT NULL, `name` varchar(512) NOT NULL, `photo` varchar(512) NOT NULL, `text` varchar(512) NOT NULL, `button` varchar(512) NOT NULL, `date` varchar(512) NOT NULL, `members` int(11) NOT NULL, `wintext` varchar(512) NOT NULL)")
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS `participants` (`id` int(11) AUTO_INCREMENT NOT NULL PRIMARY KEY, `username` varchar(512) NOT NULL, `chatid` int(22) NOT NULL, `competid` int(22) NOT NULL)")
	if err != nil {
		log.Panic(err)
	}

	go checktimedate(time.Second*30, bot, db)

	for update := range updates {

		if update.Message != nil {
			// логируем от кого какое сообщение пришло
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			if update.Message.Text != "" && update.Message.Text != "Отменить создание конкурса" {

				update.Message.Text = strings.Replace(update.Message.Text, "``", "`", -1)

				if ismakecompetition[update.Message.Chat.ID] == 1 {
					Competition.Name = update.Message.Text
					ismakecompetition[update.Message.Chat.ID] = 0
					if iscompetitionedit[update.Message.Chat.ID] == 0 {
						update.Message.Text = "Название конкурса сохранено, отправьте картинку:\n\nЕсли картинка не нужна, отправьте \"-\"."
						ismakecomptext[update.Message.Chat.ID] = 1
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
						msg.ReplyMarkup = cancelBoard
						bot.Send(msg)
					} else {
						iscompetitionedit[update.Message.Chat.ID] = 0
						update.Message.Text = "Новое название сохранено."
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
						msg.ReplyMarkup = confirmcancelBoard
						bot.Send(msg)

						competitionButton = tgbotapi.NewInlineKeyboardMarkup(
							tgbotapi.NewInlineKeyboardRow(
								tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%s", Competition.Button), fmt.Sprintf("button:%s", Competition.Button)),
							),
						)

						if Competition.Photo != "-" {
							msg := tgbotapi.NewPhotoShare(update.Message.Chat.ID, Competition.Photo)
							msg.Caption = Competition.Text
							msg.ReplyMarkup = competitionButton
							msg.ParseMode = "Markdown"
							bot.Send(msg)
						} else {
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, Competition.Text)
							msg.ReplyMarkup = competitionButton
							msg.ParseMode = "Markdown"
							bot.Send(msg)
						}
					}
					continue
				}

				if ismakecomptext[update.Message.Chat.ID] == 1 && update.Message.Text == "-" {
					Competition.Photo = update.Message.Text
					update.Message.Text = "Отправьте текст конкурса:\n\nНе бойтесь допустить ошибку, Вы всегда сможете отредактировать конкурс.\n\nДля размети используйте:\n\n*Жирный*\n_Курсив_\n``Код``\n[Ссылка](t.me)"
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
					msg.ReplyMarkup = cancelBoard
					ismakecomptext[update.Message.Chat.ID] = 0
					ismakecompbutton[update.Message.Chat.ID] = 1
					bot.Send(msg)
					continue
				}

				if ismakecompbutton[update.Message.Chat.ID] == 1 {
					Competition.Text = update.Message.Text
					ismakecompbutton[update.Message.Chat.ID] = 0
					if iscompetitionedit[update.Message.Chat.ID] == 0 {
						update.Message.Text = "Текст конкурса сохранен, введите текст кнопки:"
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
						msg.ReplyMarkup = cancelBoard
						ismakecompdate[update.Message.Chat.ID] = 1
						bot.Send(msg)
					} else {
						iscompetitionedit[update.Message.Chat.ID] = 0
						update.Message.Text = "Новый текст сохранен."
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
						msg.ReplyMarkup = confirmcancelBoard
						bot.Send(msg)

						competitionButton = tgbotapi.NewInlineKeyboardMarkup(
							tgbotapi.NewInlineKeyboardRow(
								tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%s", Competition.Button), fmt.Sprintf("button:%s", Competition.Button)),
							),
						)

						if Competition.Photo != "-" {
							msg := tgbotapi.NewPhotoShare(update.Message.Chat.ID, Competition.Photo)
							msg.Caption = Competition.Text
							msg.ReplyMarkup = competitionButton
							msg.ParseMode = "Markdown"
							bot.Send(msg)
						} else {
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, Competition.Text)
							msg.ReplyMarkup = competitionButton
							msg.ParseMode = "Markdown"
							bot.Send(msg)
						}
					}
					continue
				}

				if ismakecompdate[update.Message.Chat.ID] == 1 {
					Competition.Button = update.Message.Text
					update.Message.Text = "Текст кнопки сохранен, введите текст итогов:\n\nИспользуйте тег \"[winners]\", в том месте, где хотите видеть список победителей."
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
					msg.ReplyMarkup = cancelBoard
					ismakecompdate[update.Message.Chat.ID] = 0
					ismakecompetitionwintext[update.Message.Chat.ID] = 1
					bot.Send(msg)
					continue
				}

				if ismakecompetitionwintext[update.Message.Chat.ID] == 1 {
					Competition.Wintext = update.Message.Text
					update.Message.Text = "Текст итогов сохранен, введите количество победителей:"
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
					msg.ReplyMarkup = cancelBoard
					ismakecompetitionwintext[update.Message.Chat.ID] = 0
					ismakecompmembers[update.Message.Chat.ID] = 1
					bot.Send(msg)
					continue
				}

				if ismakecompmembers[update.Message.Chat.ID] == 1 {
					if isInt(update.Message.Text) == true {
						Competition.Members = update.Message.Text
						update.Message.Text = "Количество победителей сохранено, введите дату окончания конкурса в формате \"01.01.2001 17:00\":\n\nПРИ ВВОДЕ ДАННЫХ БУДЬТЕ ПРЕДЕЛЬНО ВНИМАТЕЛЬНЫ!!!"
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
						msg.ReplyMarkup = cancelBoard
						ismakecompmembers[update.Message.Chat.ID] = 0
						ismakecompend[update.Message.Chat.ID] = 1
						bot.Send(msg)
					} else {
						update.Message.Text = "Количество победителей не сохранено, введите цифровое значение:"
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
						msg.ReplyMarkup = cancelBoard
						bot.Send(msg)
					}
					continue
				}

				if ismakecompend[update.Message.Chat.ID] == 1 {
					Competition.Date = update.Message.Text
					var date string
					var mtime string
					var day string
					var month string
					var year string
					var hour string
					var minute string

					datetime := strings.Split(Competition.Date, " ")
					if len(datetime) == 2 {
						date = datetime[0]
						mtime = datetime[1]
					} else {
						update.Message.Text = "Данные введены неверно, вводите в формате \"01.01.2001 17:00\"."
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
						bot.Send(msg)
						continue
					}

					splitdate := strings.Split(date, ".")
					if len(splitdate) == 3 {
						day = splitdate[0]
						month = splitdate[1]
						year = splitdate[2]
					} else {
						update.Message.Text = "Данные введены неверно, вводите в формате \"01.01.2001 17:00\"."
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
						bot.Send(msg)
						continue
					}

					splittime := strings.Split(mtime, ":")
					if len(splittime) == 2 {
						hour = splittime[0]
						minute = splittime[1]
					} else {
						update.Message.Text = "Данные введены неверно, вводите в формате \"01.01.2001 17:00\"."
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
						bot.Send(msg)
						continue
					}

					Competition.Date = fmt.Sprintf("%s-%s-%sT%s:%s:00+03:00", year, month, day, hour, minute)

					update.Message.Text = "Конкурс создан, если все впорядке - нажмите \"Подтвердить\"."
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
					msg.ReplyMarkup = confirmcancelBoard
					ismakecompend[update.Message.Chat.ID] = 0
					ismakecompcreated[update.Message.Chat.ID] = 1
					bot.Send(msg)

					competitionButton = tgbotapi.NewInlineKeyboardMarkup(
						tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%s", Competition.Button), fmt.Sprintf("button:%s", Competition.Button)),
						),
					)

					if Competition.Photo != "-" {
						msg := tgbotapi.NewPhotoShare(update.Message.Chat.ID, Competition.Photo)
						msg.Caption = fmt.Sprintf("%s", Competition.Text)
						msg.ReplyMarkup = competitionButton
						msg.ParseMode = "Markdown"
						bot.Send(msg)
					} else {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, Competition.Text)
						msg.ReplyMarkup = competitionButton
						msg.ParseMode = "Markdown"
						bot.Send(msg)
					}

					continue
				}

			}

			if update.Message.Photo != nil && ismakecomptext[update.Message.Chat.ID] == 1 {
				photoArray := update.Message.Photo
				photoLastIndex := len(photoArray) - 1
				photo := photoArray[photoLastIndex] // Получаем последний элемент массива (самую большую картинку)
				Competition.Photo = photo.FileID
				update.Message.Text = "Отправьте текст конкурса:\n\nНе бойтесь допустить ошибку, Вы всегда сможете отредактировать конкурс.\n\nДля размети используйте:\n\n*Жирный*\n_Курсив_\n``Код``\n[Ссылка](t.me)"
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
				msg.ReplyMarkup = cancelBoard
				ismakecomptext[update.Message.Chat.ID] = 0
				ismakecompbutton[update.Message.Chat.ID] = 1
				bot.Send(msg)
				continue
			}

			switch update.Message.Text {
			case "🎲 Создать конкурс":
				//update.Message.Text = "Вы попали в мастер быстрого создания конкурса, для начала введите название конкурса:"
				/*update.Message.Text = "Вы попали в мастер быстрого создания конкурса, для начала выберите канал для проведения конкурса:"
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
				msg.ReplyMarkup = cancelBoard
				//ismakecompetition[update.Message.Chat.ID] = 1
				bot.Send(msg)*/
				update.Message.Text = "Вы попали в мастер быстрого создания конкурса, для начала выберите канал для проведения конкурса:"
				ismakecompetitionchan[update.Message.Chat.ID] = 1
				rowscount := 0
				cmpts := make([]*Channels, 0)
				var channelslist string

				rows, err := db.Query(fmt.Sprintf("SELECT * FROM `cbotchannels` WHERE `owner` = %d", update.Message.Chat.ID))
				if err != nil {
					log.Panic(err)
				}

				for rows.Next() {
					cmpt := new(Channels)
					err := rows.Scan(&cmpt.Id, &cmpt.Owner, &cmpt.Channelid, &cmpt.Channelname, &cmpt.Channeltitle)
					if err != nil {
						log.Panic(err)
					}

					cmpts = append(cmpts, cmpt)

					channelslist += fmt.Sprintf("[%s](t.me/%s)\n\n", cmpt.Channeltitle, cmpt.Channelname)
					println(cmpt.Channeltitle, cmpt.Channelid)
					rowscount += 1
				}

				if rowscount == 0 {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Не найдено ни одного канала.")
					msg.ReplyMarkup = mainKeyboard
					bot.Send(msg)
					continue
				}

				var chosecompetitionBoard = tgbotapi.NewInlineKeyboardMarkup(
					makeRows(makeButtons2("chanid", cmpts))...,
				)

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
				msg.ReplyMarkup = &chosecompetitionBoard
				msg.DisableWebPagePreview = true
				msg.ParseMode = "Markdown"
				bot.Send(msg)

			case "⚙️ Управление каналами":
				update.Message.Text = "Выберите нужный пункт:"
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
				msg.ReplyMarkup = channelsBoard
				bot.Send(msg)

			case "Подтвердить":
				if ismakecompcreated[update.Message.Chat.ID] == 1 {
					update.Message.Text = fmt.Sprintf("Конкурс сохранен, итоги будут известны: %s", Competition.Date)
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
					msg.ReplyMarkup = publicationBoard
					bot.Send(msg)
					ismakecompcreated[update.Message.Chat.ID] = 0
				}

			case "Опубликовать сейчас":
				chanid, _ := strconv.Atoi(Competition.Channel)

				update.Message.Text = "Конкурс успешно опубликован."
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
				msg.ReplyMarkup = mainKeyboard
				bot.Send(msg)

				println(Competition.Text)

				result, err := db.Exec("INSERT INTO `competitions` (`owner`, `channel`, `name`, `photo`, `text`, `button`, `date`, `members`, `wintext`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)", update.Message.Chat.ID, Competition.Channel, Competition.Name, Competition.Photo, Competition.Text, Competition.Button, Competition.Date, Competition.Members, Competition.Wintext)
				if err != nil {
					log.Panic(err)
				}

				cmptid, _ := result.LastInsertId()

				competitionButton = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%s", Competition.Button), fmt.Sprintf("btn:%d", cmptid)),
					),
				)

				if Competition.Photo != "-" {
					msg := tgbotapi.NewPhotoShare(int64(chanid), Competition.Photo)
					msg.Caption = fmt.Sprintf("%s", Competition.Text)
					msg.ReplyMarkup = competitionButton
					msg.ParseMode = "Markdown"
					bot.Send(msg)
				} else {
					msg := tgbotapi.NewMessage(int64(chanid), Competition.Text)
					println(int64(chanid))
					msg.ReplyMarkup = competitionButton
					msg.ParseMode = "Markdown"
					bot.Send(msg)
				}

				ismakecompetition[update.Message.Chat.ID] = 0
				ismakecomptext[update.Message.Chat.ID] = 0
				ismakecompbutton[update.Message.Chat.ID] = 0
				ismakecompdate[update.Message.Chat.ID] = 0
				ismakecompend[update.Message.Chat.ID] = 0
				ismakecompcreated[update.Message.Chat.ID] = 0
				ismakecompetition[update.Message.Chat.ID] = 0
				addchannel[update.Message.Chat.ID] = 0
				iscompetitionedit[update.Message.Chat.ID] = 0

			case "Предпросмотр":
				msg := tgbotapi.NewPhotoShare(update.Message.Chat.ID, Competition.Photo)
				msg.Caption = fmt.Sprintf("%s", Competition.Text)
				competitionButton = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%s", Competition.Button), fmt.Sprintf("button:%s", Competition.Button)),
					),
				)
				msg.ReplyMarkup = competitionButton
				msg.ParseMode = "Markdown"
				bot.Send(msg)

			case "Назад":

			case "Отменить создание конкурса", "Главное меню":
				update.Message.Text = "Вы вернулись в главное меню."
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
				msg.ReplyMarkup = mainKeyboard
				bot.Send(msg)

				ismakecompetition[update.Message.Chat.ID] = 0
				ismakecomptext[update.Message.Chat.ID] = 0
				ismakecompbutton[update.Message.Chat.ID] = 0
				ismakecompdate[update.Message.Chat.ID] = 0
				ismakecompend[update.Message.Chat.ID] = 0
				ismakecompcreated[update.Message.Chat.ID] = 0
				iscompetitionedit[update.Message.Chat.ID] = 0

			case "📜 Текущие конкурсы":
				rowscount := 0
				cmpts := make([]*Competitions, 0)

				rows, err := db.Query(fmt.Sprintf("SELECT * FROM `competitions` WHERE `owner` = %d", update.Message.Chat.ID))
				if err != nil {
					log.Panic(err)
				}

				for rows.Next() {
					cmpt := new(Competitions)
					err := rows.Scan(&cmpt.Id, &cmpt.Owner, &cmpt.Channel, &cmpt.Name, &cmpt.Photo, &cmpt.Text, &cmpt.Button, &cmpt.Date, &cmpt.Members, &cmpt.Wintext)
					if err != nil {
						log.Panic(err)
					}
					println(cmpt.Photo)
					cmpts = append(cmpts, cmpt)
					rowscount += 1
				}

				if rowscount == 0 {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Активных конкурсов не найдено.")
					//msg.ReplyMarkup = &chosecompetitionBoard
					bot.Send(msg)
					continue
				}

				var chosecompetitionBoard = tgbotapi.NewInlineKeyboardMarkup(
					makeRows(makeButtons("cmptid", cmpts))...,
				)

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите нужный конкурс для просмотра и редактирования:")
				msg.ReplyMarkup = &chosecompetitionBoard
				bot.Send(msg)

			case "Добавить канал":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Перешлите любое сообщение из канала, который хотите добавить. Не забудьте добавить бота в администраторы канала, если Вы этого не сделаете, бот не сможет публиковать конкурсы.")
				addchannel[update.Message.Chat.ID] = 1
				bot.Send(msg)

			case "Мои каналы":
				rowscount := 0
				cmpts := make([]*Channels, 0)
				var channelslist string

				rows, err := db.Query(fmt.Sprintf("SELECT * FROM `cbotchannels` WHERE `owner` = %d", update.Message.Chat.ID))
				if err != nil {
					log.Panic(err)
				}

				for rows.Next() {
					cmpt := new(Channels)
					err := rows.Scan(&cmpt.Id, &cmpt.Owner, &cmpt.Channelid, &cmpt.Channelname, &cmpt.Channeltitle)
					if err != nil {
						log.Panic(err)
					}

					cmpts = append(cmpts, cmpt)

					channelslist += fmt.Sprintf("[%s](t.me/%s)\n\n", cmpt.Channeltitle, cmpt.Channelname)
					println(cmpt.Channeltitle, cmpt.Channelid)
					rowscount += 1
				}

				if rowscount == 0 {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Не найдено ни одного канала.")
					bot.Send(msg)
					continue
				}

				/*var chosecompetitionBoard = tgbotapi.NewInlineKeyboardMarkup(
					makeRows(makeButtons("cmptid", cmpts))...
				)*/

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Список добавленных каналов:\n\n%s", channelslist))
				//msg.ReplyMarkup = &chosecompetitionBoard
				msg.DisableWebPagePreview = true
				msg.ParseMode = "Markdown"
				bot.Send(msg)

			case "Изменить название конкурса":
				iscompetitionedit[update.Message.Chat.ID] = 1
				update.Message.Text = "Введите новое название конкурса:"
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
				msg.ReplyMarkup = cancelContinueboard
				ismakecompetition[update.Message.Chat.ID] = 1
				bot.Send(msg)

			case "Изменить текст конкурса":
				iscompetitionedit[update.Message.Chat.ID] = 1
				update.Message.Text = "Введите новый текст конкурса:"
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
				msg.ReplyMarkup = cancelContinueboard
				ismakecompbutton[update.Message.Chat.ID] = 1
				bot.Send(msg)

			}

			if update.Message.ForwardFromChat != nil && addchannel[update.Message.Chat.ID] == 1 {
				if update.Message.ForwardFromChat.UserName != "" {
					_, err := db.Exec("INSERT INTO `cbotchannels` (`channelid`, `owner`, `channelname`, `channeltitle`) VALUES (?, ?, ?, ?)", update.Message.ForwardFromChat.ID, update.Message.Chat.ID, update.Message.ForwardFromChat.UserName, update.Message.ForwardFromChat.Title)
					if err != nil {
						log.Fatal(err)
					}
					println(update.Message.ForwardFromChat.ID)

					msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("✅ Канал [%s](t.me/%s) успешно добавлен.", update.Message.ForwardFromChat.Title, update.Message.ForwardFromChat.UserName))
					msg.ParseMode = "Markdown"
					msg.DisableWebPagePreview = true
					bot.Send(msg)
				} else {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "❌ Канал не добавлен, т.к. отсутствует username")
					bot.Send(msg)
				}

				addchannel[update.Message.Chat.ID] = 0
				continue
			}

			switch update.Message.Command() {
			case "start":
				update.Message.Text = "Привет, здесь ты сможешь провести конкурс на своем канале!"
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
				msg.ReplyMarkup = mainKeyboard
				bot.Send(msg)
			}
		}

		if update.CallbackQuery != nil {
			fmt.Println("CALLBACK DATA", update.CallbackQuery.Data)

			splistr := strings.Split(update.CallbackQuery.Data, ":")

			if splistr[0] == "btn" {
				var dbchatid int64
				var username string

				username = update.CallbackQuery.From.FirstName

				if update.CallbackQuery.From.LastName != "" {
					username += " " + update.CallbackQuery.From.LastName
				}

				_ = db.Get(&dbchatid, "SELECT `chatid` FROM `participants` WHERE `chatid` = ? AND `competid` = ?", update.CallbackQuery.From.ID, splistr[1])

				if dbchatid == int64(update.CallbackQuery.From.ID) {
					bot.Send(tgbotapi.NewCallbackWithAlert(update.CallbackQuery.ID, "Ты уже в деле!"))
					continue
				}

				println("ID IN CALLBACK", update.CallbackQuery.From.ID)

				bot.Send(tgbotapi.NewCallbackWithAlert(update.CallbackQuery.ID, "Отлично, ты в деле!"))
				_, err := db.Exec("INSERT INTO `participants` (`username`, `chatid`, `competid`) VALUES (?, ?, ?)", username, update.CallbackQuery.From.ID, splistr[1])
				if err != nil {
					log.Panic(err)
				}
			}

			if splistr[0] == "cmptid" {
				println(splistr[1])

				var cmpt Competitions

				_ = db.Get(&cmpt, fmt.Sprintf("SELECT * FROM `competitions` WHERE `id` = %s LIMIT 1", splistr[1]))
				/*if err != nil {
					log.Panic(err)
				}*/

				competitionButton = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%s", cmpt.Button), fmt.Sprintf("button:%s", cmpt.Button)),
					),
				)

				if cmpt.Photo != "-" {
					msg := tgbotapi.NewPhotoShare(update.CallbackQuery.Message.Chat.ID, cmpt.Photo)
					println(cmpt.Text)
					msg.Caption = fmt.Sprintf("%s", cmpt.Text)
					msg.ReplyMarkup = competitionButton
					msg.ParseMode = "Markdown"
					bot.Send(msg)
				} else {
					println(cmpt.Text)
					msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, cmpt.Text)
					msg.ReplyMarkup = competitionButton
					msg.ParseMode = "Markdown"
					bot.Send(msg)
				}
			}

			if splistr[0] == "chanid" && ismakecompetitionchan[update.CallbackQuery.Message.Chat.ID] == 1 {
				println(splistr[1])
				Competition.Channel = splistr[1]
				update.CallbackQuery.Message.Text = "Введите название конкурса:"
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.Text)
				msg.ReplyMarkup = cancelBoard
				bot.Send(msg)

				ismakecompetitionchan[update.CallbackQuery.Message.Chat.ID] = 0
				ismakecompetition[update.CallbackQuery.Message.Chat.ID] = 1
			}
		}
	}
}
