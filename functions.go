package main

import (
	_ "database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jinzhu/configor"
	"github.com/jmoiron/sqlx"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func init() {
	_ = configor.Load(&Config, "config.yaml")
	fmt.Printf("config: %#v\n", Config)
	// без него не запускаемся
	if Config.Bot.Token == "" {
		log.Print("telegrambottoken is required")
		os.Exit(1)
	}
}

func getAPI() (*tgbotapi.BotAPI, error) {
	/*host := "93.170.76.34:1080"

	auth := proxy.Auth {
		User:     "proxyroot",
		Password: "CqbpSqDS48Vr",
	}

	dialer, err := proxy.SOCKS5("tcp", host, &auth, proxy.Direct)

	if err != nil {
		// handle err
		log.Panic()
	}

	client := &http.Client {
		Transport: &http.Transport {
			Dial: dialer.Dial,
		},
	}*/

	conn, err := tgbotapi.NewBotAPI(Config.Bot.Token /*, client*/)

	return conn, err
}

func makeRows(buttons []tgbotapi.InlineKeyboardButton) [][]tgbotapi.InlineKeyboardButton {
	var rows [][]tgbotapi.InlineKeyboardButton

	for _, item := range buttons {
		rows = append(rows,
			tgbotapi.NewInlineKeyboardRow(item),
		)
	}

	return rows
}

func makeButtons(command string, competitions []*Competitions) []tgbotapi.InlineKeyboardButton {
	var buttons []tgbotapi.InlineKeyboardButton

	for _, item := range competitions {

		button := makeInlineButton(
			command,
			strconv.Itoa(item.Id),
		)

		println(item.Name)

		buttons = append(buttons,
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf(item.Name), button),
		)
	}

	return buttons
}

func makeButtons2(command string, competitions []*Channels) []tgbotapi.InlineKeyboardButton {
	var buttons []tgbotapi.InlineKeyboardButton

	for _, item := range competitions {

		button := makeInlineButton(
			command,
			fmt.Sprintf("%d", item.Channelid),
		)

		println(item.Channeltitle)

		buttons = append(buttons,
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf(item.Channeltitle), button),
		)
	}

	return buttons
}

func makeInlineButton(args ...string) string {
	return /*"#" +*/ strings.Join(args, ":")
}

func checktimedate(d time.Duration, bot *tgbotapi.BotAPI /*, message *tgbotapi.Message*/, db *sqlx.DB) {
	for range time.Tick(d) {
		tunix := time.Now().Unix()
		var winners string
		//var channel int64
		//var wintext string

		//fmt.Printf("%s в UNIX: %d\n", maindate, tTime.Unix())
		//fmt.Printf("%v: Hello, Timestamp!\n", tunix)

		rowscount := 0
		cmpts := make([]*Competitions, 0)

		rows, _ := db.Query("SELECT * FROM `competitions`")
		/*if err != nil {
			log.Panic(err)
		}*/

		for rows.Next() {
			cmpt := new(Competitions)
			_ = rows.Scan(&cmpt.Id, &cmpt.Owner, &cmpt.Channel, &cmpt.Name, &cmpt.Photo, &cmpt.Text, &cmpt.Button, &cmpt.Date, &cmpt.Members, &cmpt.Wintext)
			/*if err != nil {
				log.Panic(err)
			}*/
			//channel = cmpt.Channel
			cmpts = append(cmpts, cmpt)
			rowscount += 1

			//var maindate = cmpt.Date
			tTime, _ := time.Parse(time.RFC3339, cmpt.Date)
			//fmt.Printf("%s в UNIX: %d\n", maindate, tTime.Unix())

			if tunix >= tTime.Unix() && tTime.Unix() > 0 {
				go win(bot, db, winners, cmpt.Channel, cmpt.Wintext, cmpts)
				_, _ = db.Query(fmt.Sprintf("DELETE FROM `competitions` WHERE `id` = %d", cmpt.Id))
			}
		}
	}
}

func win(bot *tgbotapi.BotAPI, db *sqlx.DB, winners string, channel int64, wintext string, cmpt []*Competitions) {
	for _, cmpt := range cmpt {
		fmt.Printf("ПОБЕДА!!!\n")

		rowscount := 0
		prtns := make([]*Participans, 0)

		rows, _ := db.Query("SELECT * FROM `participants` WHERE `competid` = ? ORDER BY RAND() LIMIT ?", cmpt.Id, cmpt.Members)
		println("Победителей: ", cmpt.Members)

		for rows.Next() {
			prtn := new(Participans)
			err := rows.Scan(&prtn.Id, &prtn.Username, &prtn.Chatid, &prtn.Cmptid)
			if err != nil {
				log.Panic(err)
			}

			println(channel, prtn.Chatid, prtn.Username)

			var results map[string]map[string]interface{}
			var resp *http.Response

			var httpsend = fmt.Sprintf("https://api.telegram.org/bot%s/getChatMember?chat_id=%d&user_id=%d", Config.Bot.Token, channel, prtn.Chatid)

			//fmt.Printf("%s Статус в проверке на подписку", results["result"]["status"])

			resp, _ = http.Get(httpsend)
			json.NewDecoder(resp.Body).Decode(&results)

			if results["result"]["status"] == "creator" || results["result"]["status"] == "member" /*|| results["result"]["status"] == "administrator"*/ {
				winners += fmt.Sprintf("[%s](tg://user?id=%d), ", prtn.Username, prtn.Chatid)
			}
			prtns = append(prtns, prtn)
			rowscount += 1
		}

		winners = strings.TrimRight(winners, "!, ")

		//chanid, _ := strconv.Atoi()
		var ownerid int64
		_ = db.Get(&ownerid, "SELECT `owner` FROM `cbotchannels` WHERE `channelid` = ? LIMIT 1", channel)

		msg := tgbotapi.NewMessage(ownerid, fmt.Sprintf("В конкурсе \"%s\", победили: %s", cmpt.Name, winners))
		if winners == "" {
			msg = tgbotapi.NewMessage(ownerid, fmt.Sprintf("В конкурсе \"%s\", никто не победил.", cmpt.Name))
		}
		msg.DisableWebPagePreview = true
		msg.ParseMode = "Markdown"
		bot.Send(msg)

		wintext = strings.Replace(wintext, "[winners]", winners, -1)

		msg = tgbotapi.NewMessage(channel, fmt.Sprintf("%s", wintext))
		if winners == "" {
			msg = tgbotapi.NewMessage(channel, fmt.Sprintf("В конкурсе \"%s\", никто не победил.", cmpt.Name))
		}
		msg.DisableWebPagePreview = true
		msg.ParseMode = "Markdown"
		bot.Send(msg)
		//println(cmpt.Id)
	}
}

func isInt(s string) bool {
	for _, c := range s {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}
