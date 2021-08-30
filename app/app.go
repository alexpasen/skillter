package app

import (
	"cloud.google.com/go/translate"
	"context"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/rs/zerolog"
	"golang.org/x/text/language"
	"google.golang.org/api/option"
	"io"
	"net/http"
	"os"
	"skillter/model"
)

var Logger zerolog.Logger

func init() {
	//consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}

	Logger = zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()
}

type App struct {
	translateApi *translate.Client
	telegramApi  *tgbotapi.BotAPI
	quitChannel  chan int
}

func NewApp() (*model.AppError, *App) {
	Logger.Info().Msg("Starting skillter ... ")

	ctx := context.Background()
	app := &App{
		quitChannel: make(chan int, 1),
	}

	webPort, ok := os.LookupEnv("PORT")
	if !ok {
		return model.NewAppError("setup_http_port", "app.new", errors.New("Invalid PORT value.")), nil
	}

	googleApiKey, ok :=os.LookupEnv(model.GOOGLE_TRANSLATE_API_KEY)
	if !ok {
		return model.NewAppError("setup_translate_api_key", "app.new", errors.New("Invalid GOOGLE_TRANSLATE_API_KEY value.")), nil
	}

	telegramApiKey, ok :=os.LookupEnv(model.TELEGRAM_API_KEY)
	if !ok {
		return model.NewAppError("setup_telegram_api_key", "app.new", errors.New("Invalid TELEGRAM_API_KEY value.")), nil
	}

	if translateApi, err := translate.NewClient(ctx, option.WithAPIKey(googleApiKey));
	err != nil {
		return model.NewAppError("Connect google translate", "NewApp", err), nil
	} else {
		app.translateApi = translateApi
	}

	if telegramApi, err := tgbotapi.NewBotAPI(telegramApiKey); err != nil {
		return model.NewAppError("connect_telegram_bot", "app.new", err), nil
	} else {
		app.telegramApi = telegramApi
	}

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		io.WriteString(writer, "Skillter backend.")
	})

	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%s", webPort), nil); err != nil {
			Logger.Error().Msg(model.NewAppError("setup_web_server", "app.new", err).String())
		}

	}()

	return nil, app
}

func (a *App) Start() *model.AppError {

	ctx := context.Background()
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := a.telegramApi.GetUpdatesChan(u)
	if err != nil {
		return model.NewAppError("get_bot_update_chan", "app.start", err)
	}

	for {
		select {
		case update := <-updates:
			{
				if update.Message == nil { // ignore any non-Message Updates
					continue
				}

				Logger.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

				tr, err := a.translateApi.Translate(ctx, []string{update.Message.Text}, language.Russian, &translate.Options{})
				if err != nil {
					Logger.Error().Msg(err.Error())
				}

				var variants string
				for i, t := range tr {
					if len(variants) == 0 {
						variants = fmt.Sprintf("\n%d. %s", i+1, t.Text)
					} else {
						variants += fmt.Sprintf("\n%d. %s", i+1, t.Text)
					}
				}

				translated := fmt.Sprintf("You want to translate: %s\nProbably the translation is:%s", update.Message.Text, variants)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, translated)
				//msg.ReplyToMessageID = update.Message.MessageID

				a.telegramApi.Send(msg)
			}
		case <-a.quitChannel:
			{
				a.translateApi.Close()
				os.Exit(0)
			}

		}

	}
}

func (a *App) Stop() {
	a.quitChannel <- 0
}
