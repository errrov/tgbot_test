package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var numericKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("отзывы"),
		tgbotapi.NewKeyboardButton("вопросы"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("сотрудничество"),
	),
)

var returnButton = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("назад"),
	),
)

var (
	feedbackMessage = []string{"Спасибо! Для нас важно ваше мнение - именно так мы понимаем, что стоит улучшить. Пишите чаще, xo xo",
		"Спасибо, что поделились с нами мнением о продукте ORELE. Мы рады, что вы выбрали нас. Обещаем, что не подведем!",
		"Спасибо за выбор ORELE. Надеемся и другие наши продукты придутся вам по душе, ведь все они сделаны для вас. Улыбайтесь, вы прекрасны",
		"Благодарим вас за то, что вы выбрали наш продукт и нашли время для обратной связи. Передадим ваш отзыв в отдел качества для дальнейшего совершенствования нашей продукции. Shine bright ✨",
		"Благодарим за доверие и выбор ORELE. Мы вкладываем всю свою любовь в наш продукт, и нам всегда приятно получать частичку тепла в ответ. Пишите нам почаще, это делает нас счастливее🥰",
		"Спасибо! ORELE создан для вас, поэтому мы читаем все отзывы и прислушиваемся к каждому пожеланию ❤",
		"Спасибо, что выбрали ORELE для заботы о вашей естественной красоте. Мы не подведем!",
		"Спасибо! Мы обожаем ваши отзывы и читаем каждый. Stay true, stay you! А мы будем лучше для вас)",
		"Спасибо за отзыв и выбор нашего продукта ❤️ Будем рады видеть вас чаще в наших соцсетях, там мы рассказываем о расширении ассортимента и делимся вдохновением",
		"Спасибо, что нашли время для обратной связи. Прекрасного вам дня и хорошей погоды в душе, даже если вы в дождливом Питере❤️",
	}
	handlingMenu        = 0
	handlingFeedback    = 1
	handlingCooperation = 2
	handlingQuestions   = 3
	handlingDialog      = 4
	masterChat          int64
)

func main() {
	master_chat, err := strconv.Atoi(os.Getenv("master_chat"))
	if err != nil {
		log.Println("error getting masterchat ID")
		os.Exit(1)
	}
	masterChat = int64(master_chat)
	botToken := os.Getenv("bot_token")
	chatStates := map[int64]int{}
	openChats := []int64{}
	LoadChatIds("chatIds.txt", &openChats, chatStates)
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil {
			continue
		}
		if (strings.Contains(update.Message.Text, "[массовая рассылка]") || strings.Contains(update.Message.Caption, "[массовая рассылка]")) && update.Message.Chat.ID == masterChat {
			update.Message.Text = strings.ReplaceAll(update.Message.Text, "[массовая рассылка]", "")
			massiveMessage(update.Message, openChats, bot)
			continue
		}
		if (update.Message.Chat.ID == masterChat) && (update.Message.ReplyToMessage != nil) && (strings.Contains(update.Message.ReplyToMessage.Text, "[вопросы]")) {
			strings := strings.Fields(update.Message.ReplyToMessage.Text)
			responseChatId := strings[len(strings)-3]
			responseMessage := strings[len(strings)-1]
			respChat, err := strconv.Atoi(responseChatId)
			if err != nil {
				log.Println(err)
				return
			}
			messageToResponse, err := strconv.Atoi(responseMessage)
			if err != nil {
				log.Println(err)
				return
			}
			msg := tgbotapi.NewMessage(int64(respChat), update.Message.Text)
			msg.ReplyToMessageID = messageToResponse
			bot.Send(msg)

		}

		switch update.Message.Text {
		case "/start":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Привет! Это бот бренда ORELE, который создан для оперативной связи с нами по любым вопросам.\nВыберите, что вы хотите сделать:\n(написать отзыв, задать вопрос, предложить сотрудничество)")
			if _, ok := chatStates[update.Message.Chat.ID]; !ok {
				chatStates[update.Message.Chat.ID] = handlingMenu
				openChats = append(openChats, update.Message.Chat.ID)
				AddOrCreateChatIDS(update.Message.Chat.ID)
			}
			msg.ReplyMarkup = numericKeyboard
			bot.Send(msg)
		case "назад":
			chatStates[update.Message.Chat.ID] = handlingMenu
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "выберите по какому именно поводу вы здесь")
			msg.ReplyMarkup = numericKeyboard
			bot.Send(msg)
		case "сотрудничество":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Мы уже заинтригованы! Пожалуйста, заполните короткую форму для предложений о сотрудничестве https://docs.google.com/forms/d/1ChSWUfvXzgYb3WQ5hLV1m-QwaicOcLrstPEHPsolaSU/viewform?edit_requested=true. Также вы можете написать на нашу почту pr.orelecosmetics@mail.ru")
			chatStates[update.Message.Chat.ID] = handlingCooperation
			msg.ReplyMarkup = returnButton
			bot.Send(msg)
		case "отзывы":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Сообщите нам о ваших мыслях по поводу продукции ORELE. Подробный отзыв позволит нам улучшать качество работы каждый день")
			chatStates[update.Message.Chat.ID] = handlingFeedback
			msg.ReplyMarkup = returnButton
			bot.Send(msg)
		case "вопросы":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы можете задать любой вопрос о продукте и бренде или написать нам, какой продукт вы хотите увидеть в линейке ORELE следующим")
			chatStates[update.Message.Chat.ID] = handlingQuestions
			msg.ReplyToMessageID = update.Message.MessageID
			msg.ReplyMarkup = returnButton
			bot.Send(msg)
		default:
			if chatStates[update.Message.Chat.ID] == handlingFeedback {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Если вы хотите оставить отзыв, то оставьте его здесь")
				r := rand.New(rand.NewSource(time.Now().Unix()))
				randomMessageIdx := r.Intn(len(feedbackMessage))
				msg.ReplyToMessageID = update.Message.MessageID
				msg.Text = feedbackMessage[randomMessageIdx]
				msg.ReplyMarkup = numericKeyboard
				bot.Send(msg)
				chatStates[update.Message.Chat.ID] = handlingMenu
				resp := craftMessage("[отзывы]", update.Message)
				msgMaster := tgbotapi.NewMessage(masterChat, resp)
				bot.Send(msgMaster)
				continue
			}
			if chatStates[update.Message.Chat.ID] == handlingCooperation {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Спасибо за интерес к бренду ORELE! Мы обязательно прочтем ваше предложение и вернемся с ответом в течение 2-3 рабочих дней.")
				msg.ReplyToMessageID = update.Message.MessageID
				msg.Text = "С вами свяжутся, как появится возможность"
				msg.ReplyMarkup = numericKeyboard
				bot.Send(msg)
				chatStates[update.Message.Chat.ID] = handlingMenu
				resp := craftMessage("[сотрудничество]", update.Message)
				msgMaster := tgbotapi.NewMessage(masterChat, resp)
				bot.Send(msgMaster)
				continue
			}
			if chatStates[update.Message.Chat.ID] == handlingQuestions || chatStates[update.Message.Chat.ID] == handlingDialog {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
				if chatStates[update.Message.Chat.ID] == handlingQuestions {
					msg.Text = "Спасибо! Мы ответим вам лично в ближайшее время ❤️. Для возврата в меню нажмите назад"
					chatStates[update.Message.Chat.ID] = handlingDialog
				}
				msg.ReplyMarkup = returnButton
				bot.Send(msg)
				resp := craftMessage("[вопросы]", update.Message)
				msgMaster := tgbotapi.NewMessage(masterChat, resp)
				bot.Send(msgMaster)
				continue
			}
			if chatStates[update.Message.Chat.ID] == handlingMenu && update.Message.Chat.ID != masterChat {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
				msg.ReplyMarkup = numericKeyboard
				bot.Send(msg)
				resp := craftMessage("[вопросы]", update.Message)
				msgMaster := tgbotapi.NewMessage(masterChat, resp)
				bot.Send(msgMaster)
			}
		}
	}
}

func craftMessage(tag string, message *tgbotapi.Message) string {
	var answer strings.Builder
	answer.Write([]byte(tag))
	answer.Write([]byte("\nMessage: "))
	answer.Write([]byte(message.Text))
	answer.Write([]byte("\nFrom: @"))
	answer.Write([]byte(message.From.UserName))
	answer.Write([]byte("\nChat_id: "))
	answer.Write([]byte(fmt.Sprint(message.Chat.ID)))
	answer.Write([]byte("\nMessage_id: "))
	answer.Write([]byte(fmt.Sprint(message.MessageID)))
	return answer.String()
}

func massiveMessage(message *tgbotapi.Message, chats []int64, bot *tgbotapi.BotAPI) {
	var hasImage bool
	var files []interface{}
	hasImage = false
	log.Println("message len:", len(message.Photo))
	log.Println(message.Photo)
	if len(message.Photo) != 0 {
		message.Caption = strings.ReplaceAll(message.Caption, "[массовая рассылка]", "")
		fileId := message.Photo[0].FileID
		var FileID tgbotapi.FileID = tgbotapi.FileID(fileId)
		newPhotoMsg := tgbotapi.InputMediaPhoto{}
		newPhotoMsg.Type = "photo"
		newPhotoMsg.Media = FileID
		newPhotoMsg.Caption = message.Caption
		hasImage = true
		files = append(files, newPhotoMsg)
	}
	for _, v := range chats {
		if !hasImage {
			msg := tgbotapi.NewMessage(v, message.Text)
			bot.Send(msg)
		} else {
			msg := tgbotapi.NewMediaGroup(v, files)
			_, err := bot.SendMediaGroup(msg)
			log.Println(err)
		}
	}
}

func AddOrCreateChatIDS(chatId int64) {
	f, err := os.OpenFile("chatIds.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Error openning file:", err)
		return
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Println("Error closing file:", err)
			return
		}
	}()

	_, err = f.WriteString(fmt.Sprint(chatId) + " ")
	if err != nil {
		log.Println("Error writting string:", err)
		return
	}
}

func LoadChatIds(filename string, chats *[]int64, chatStates map[int64]int) {
	b, err := os.ReadFile(filename)
	if err != nil {
		log.Println("Error loading ids:", err)
		return
	}
	chatIds := strings.Fields(string(b))
	for _, v := range chatIds {
		val, err := strconv.Atoi(v)
		if err != nil {
			log.Println("Error converting string to int64:", err)
			continue
		}
		*chats = append(*chats, int64(val))
		chatStates[int64(val)] = 0
	}
}
