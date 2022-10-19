package main

import (
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var (
	periodKeyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Минута"),
			tgbotapi.NewKeyboardButton("Час"),
			tgbotapi.NewKeyboardButton("День"),
			tgbotapi.NewKeyboardButton("Неделя"),
		))
	materialKeyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Видеоматериалы"),
			tgbotapi.NewKeyboardButton("Доп. материалы"),
		),
	)
)

const (
	SPREADSHEET_ID string = ""
	GOOGLE_API_KEY string = ""
	TELEGRAM_TOKEN string = ""
)

type User struct {
	Period         time.Duration
	SelectedPeriod bool
}

type Material struct {
	Values [][]string `json:"values"`
}

type Queue struct {
	Id   int64
	Time time.Time
}

var users = make(map[int64]User)

func ChoosePeriod(id int64, text string) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(id, text)
	msg.ReplyMarkup = periodKeyboard
	return msg
}

func ChooseMaterial(id int64, text string) tgbotapi.MessageConfig {
	textMaterial := "\nКакой материал ты хочешь получить?"
	msg := tgbotapi.NewMessage(id, text+textMaterial)
	msg.ReplyMarkup = materialKeyboard
	return msg
}

func GetInfoGoogleSheet(variant string) string {
	var rang string
	if variant == "video" {
		rang = "A2:A10"
	} else {
		rang = "L2:L10"
	}

	request := "https://sheets.googleapis.com/v4/spreadsheets/" + string(SPREADSHEET_ID) + "/values/" + string(rang) +
		"?key=" + GOOGLE_API_KEY
	resp, err := http.Get(request)
	if err != nil {
		log.Fatalln(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	var text string
	var dictBody Material
	fmt.Println(body)
	fmt.Println(string(body))
	if err = json.Unmarshal(body, &dictBody); err != nil {
		log.Panic(err)
	}
	for _, elem := range dictBody.Values {
		text += elem[0] + "\n"
	}
	return text
}

func AddToQueue(queue []Queue, current Queue) []Queue {
	var index int = 0
	flag := false
	for i := 0; i < len(queue); i++ {
		if current.Time.Before(queue[i].Time) {
			index = i
			flag = true
			break
		}
	}
	if !flag {
		queue = append(queue, current)
	} else {
		if index == 0 {
			queue = append([]Queue{current}, queue...)
		} else {
			queue = append(queue[:index+1], queue[index:]...)
			queue[index] = current
		}
	}
	return queue
}

func main() {
	bot, err := tgbotapi.NewBotAPI(TELEGRAM_TOKEN)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 20

	updates := bot.GetUpdatesChan(u)

	queue := []Queue{}

	go func() {
		for {
			currentTime := time.Now()
			fmt.Print("----очередь----:", queue, "\n")
			fmt.Print("----нынешнее время---:", currentTime, "\n")
			if len(queue) != 0 {
				currentDay := currentTime.Day()
				currentHour := currentTime.Hour()
				currentMinute := currentTime.Minute()
				queTime := queue[0].Time
				queDay := queTime.Day()
				queHour := queTime.Hour()
				queMinute := queTime.Minute()
				if queDay == currentDay && queHour == currentHour && queMinute == currentMinute {
					textNotification := "Пора возвращаться к учёбе.\n" +
						"C любовью, твой чат-бот :)"
					msg := tgbotapi.NewMessage(queue[0].Id, textNotification)
					current := Queue{queue[0].Id, queue[0].Time.Add(users[queue[0].Id].Period)}
					queue = queue[1:]
					queue = AddToQueue(queue, current)
					if _, err = bot.Send(msg); err != nil {
						panic(err)
					}
				}
			}
			time.Sleep(2 * time.Second)
		}
	}()
	// Loop through each update.
	for update := range updates {
		currentTime := time.Now()
		if update.Message != nil {
			var text string = "такого я не ожидал"
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
			id := update.Message.Chat.ID

			switch update.Message.Text {
			case "/start":
				if _, ok := users[id]; ok {
					for i, val := range queue {
						if val.Id == id {
							if i == 0 {
								queue = queue[1:]
							} else if i == len(queue)-1 {
								queue = queue[:len(queue)-1]
							} else {
								queue = append(queue[:i], queue[i+1:]...)
							}
						}
					}
				}
				users[id] = User{Period: time.Minute, SelectedPeriod: false}
				text := "Ты зарегистрировался в обучающем чат-боте. \n" +
					"Выбери периодичность, с которой хочешь получать " +
					"напоминание об учёбе"
				msg = ChoosePeriod(id, text)
			case "/help":
				text := "Что умеет данный бот:\n" +
					"1. Отправлять ссылки на материалы\n" +
					"2. С какой-то периодичностью напоминать об учёбе\n" +
					"Если что-то пошло не так, используй команду /start"

				msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
			case "Минута":
				if val, ok := users[id]; ok {
					if users[id].SelectedPeriod == false {
						users[id] = User{Period: time.Minute, SelectedPeriod: true}
						current := Queue{id, currentTime.Add(val.Period)}
						queue = AddToQueue(queue, current)
						text = "Ты будешь получать напоминание каждую минуту."
						msg = ChooseMaterial(id, text)
					} else {
						text = "Ты уже выбрал период для напоминаний. \n" +
							"Если хочешь его изменить, используй команду /start"
						msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
					}
				}

			case "Час":
				if val, ok := users[id]; ok {
					if val.SelectedPeriod == false {
						users[id] = User{Period: time.Hour, SelectedPeriod: true}
						current := Queue{id, currentTime.Add(val.Period)}
						queue = AddToQueue(queue, current)
						text = "Ты будешь получать напоминание каждый час."
						msg = ChooseMaterial(id, text)
					} else {
						text = "Ты уже выбрал период для напоминаний. \n" +
							"Если хочешь его изменить, используй команду /start"
						msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
					}
				}
			case "День":
				if val, ok := users[id]; ok {
					if val.SelectedPeriod == false {
						users[id] = User{Period: time.Hour * 24, SelectedPeriod: true}
						current := Queue{id, currentTime.Add(val.Period)}
						queue = AddToQueue(queue, current)
						text = "Ты будешь получать напоминание каждый день."
						msg = ChooseMaterial(id, text)
					} else {
						text = "Ты уже выбрал период для напоминаний. \n" +
							"Если хочешь его изменить, используй команду /start"
						msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
					}
				}
			case "Неделя":
				if val, ok := users[id]; ok {
					if val.SelectedPeriod == false {
						users[id] = User{Period: time.Hour * 24 * 7, SelectedPeriod: true}
						current := Queue{id, currentTime.Add(val.Period)}
						queue = AddToQueue(queue, current)
						text = "Ты будешь получать напоминание каждую неделю"
						msg = ChooseMaterial(id, text)
					} else {
						text = "Ты уже выбрал период для напоминаний. \n" +
							"Если хочешь его изменить, используй команду /start"
						msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
					}
				}
			case "Видеоматериалы":
				text = GetInfoGoogleSheet("video")
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
			case "Доп. материалы":
				text = GetInfoGoogleSheet("additional")
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
			}
			if _, err = bot.Send(msg); err != nil {
				panic(err)
			}
		}
	}
}
