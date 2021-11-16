package utils

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/preichenberger/go-coinbasepro/v2"
	"github.com/robfig/cron/v3"
)

type SubscribeOptions struct {
	Base   string
	Target string
}

type CurrencyConfig struct {
	Symbol     string
	Percentage int
}

type Config struct {
	BaseUrl    string `env:"BASE_URL"`
	Key        string `env:"KEY"`
	Secret     string `env:"SECRET"`
	Passphrase string `env:"PASSPHRASE"`

	Cron           string `env:"CRON"`
	DailyLimit     string `env:"DAILY_LIMIT"`
	BaseCurrency   string `env:"BASE_CURRENCY"`
	CurrencyConfig []CurrencyConfig

	TelegramChatID string `env:"TELEGRAM_CHAT_ID"`
}

func GetEnvConfig() Config {
	config := Config{
		BaseUrl:        os.Getenv("BASE_URL"),
		Key:            os.Getenv("KEY"),
		Secret:         os.Getenv("SECRET"),
		Passphrase:     os.Getenv("PASSPHRASE"),
		Cron:           os.Getenv("CRON"),
		DailyLimit:     os.Getenv("DAILY_LIMIT"),
		BaseCurrency:   os.Getenv("BASE_CURRENCY"),
		TelegramChatID: os.Getenv("TELEGRAM_CHAT_ID"),
	}
	currencyString := os.Getenv("CURRENCY_CONFIG")

	if len(currencyString) > 0 {
		var currencies []CurrencyConfig
		currencyList := strings.Split(os.Getenv("CURRENCY_CONFIG"), ";")
		for _, v := range currencyList {
			currConfig := strings.Split(v, ":")
			percentage, err := strconv.Atoi(currConfig[1])

			if err != nil {
				log.Fatal("Error: ", err)
			}

			currencies = append(currencies, CurrencyConfig{
				Symbol:     currConfig[0],
				Percentage: percentage,
			})
		}
		if len(currencies) != 0 {
			config.CurrencyConfig = currencies
		}
	}

	return config
}

func GetCoinbaseClient(c Config) *coinbasepro.Client {
	client := coinbasepro.NewClient()
	client.UpdateConfig(&coinbasepro.ClientConfig{
		BaseURL:    c.BaseUrl,
		Key:        c.Key,
		Passphrase: c.Passphrase,
		Secret:     c.Secret,
	})

	return client
}

func SubscribeToCurrency(opts SubscribeOptions) {
	var wsDialer = websocket.DefaultDialer
	wsConn, _, err := wsDialer.Dial("wss://ws-feed.pro.coinbase.com", nil)

	if err != nil {
		fmt.Errorf("Found error:  %w", err)
	}

	subscribe := coinbasepro.Message{
		Type: "subscribe",
		Channels: []coinbasepro.MessageChannel{
			{
				Name: "ticker",
				ProductIds: []string{
					opts.Target + "-" + opts.Base,
				},
			},
		},
	}

	if err := wsConn.WriteJSON(subscribe); err != nil {
		println(err.Error())
	}

	for true {
		message := coinbasepro.Message{}
		if err := wsConn.ReadJSON(&message); err != nil {
			println(err.Error())
			break
		}
		if message.Type == "ticker" {
			// println("PRODUCT ID:", message.ProductID)
			// println("LAST:", message.LastSize)
			// println("PRICE: ", message.Price)
			// println("BEST BID: ", message.BestBid)
			// println()
			// time.Sleep(time.Second * 1)
		}
	}
}

func StartCronBuy() {
	config := GetEnvConfig()

	l, err := time.LoadLocation("Europe/Tallinn")
	if err != nil {
		fmt.Errorf("Found Error: %w\n", err)
	}

	cron.WithLocation(l)
	c := cron.New()

	c.AddFunc(config.Cron, func() {
		config := GetEnvConfig()
		client := GetCoinbaseClient(config)

		accounts, err := client.GetAccounts()
		if err != nil {
			fmt.Errorf("Found error:  %w\n", err)
		}

		dailyLimit, err := strconv.Atoi(config.DailyLimit)
		if err != nil {
			fmt.Errorf("Found error:  %w\n", err)
		}

		for _, account := range accounts {
			if account.Currency == config.BaseCurrency {
				fmt.Printf("The Balance for %v account is %v\n", account.Currency, account.Balance)
				for _, coin := range config.CurrencyConfig {
					productId := coin.Symbol + "-" + config.BaseCurrency
					order := coinbasepro.Order{
						Side:      "buy",
						Type:      "market",
						Funds:     strconv.Itoa(((dailyLimit * coin.Percentage) / 100)),
						ProductID: productId,
					}
					newOrder, err := client.CreateOrder(&order)
					if err != nil {
						fmt.Errorf("Found error:  %w\n", err)
					}

					if err == nil && len(newOrder.ID) > 0 {
						msg := buildOrderNotification(newOrder, coin, config)
						sendNotification(msg)
					}
				}
			}
		}
	})

	c.Run()
}

func buildOrderNotification(o coinbasepro.Order, coin CurrencyConfig, c Config) string {
	priceValue, _ := GetCoinbaseClient(c).GetTicker(coin.Symbol + "-" + c.BaseCurrency)
	orderFunds, _ := strconv.ParseFloat(o.Funds, 64)
	dailyLimit, _ := strconv.Atoi(c.DailyLimit)
	orderSize := float64((dailyLimit * coin.Percentage) / 100)
	fee := orderSize - orderFunds
	marketPrice, _ := strconv.ParseFloat(priceValue.Price, 64)
	amount := orderFunds / marketPrice

	msg := fmt.Sprintf(`
	💰 Order created 💰
	Date: time.Now()
	Market: %v
	Amout: %.8f %v
	Size: %.3f %v
	Fee: %.3f %v
	%v: %.3f %v
	`,
		coin.Symbol+"-"+c.BaseCurrency,
		amount,
		coin.Symbol,
		orderFunds,
		c.BaseCurrency,
		fee,
		c.BaseCurrency,
		coin.Symbol,
		marketPrice,
		c.BaseCurrency)

	return msg
}

func sendNotification(message string) (string, error) {
	telegramApi := "https://api.telegram.org/bot" + os.Getenv("TELEGRAM_BOT_TOKEN") + "/sendMessage"
	response, err := http.PostForm(
		telegramApi,
		url.Values{
			"chat_id": {os.Getenv("TELEGRAM_CHAT_ID")},
			"text":    {message},
		})

	if err != nil {
		log.Printf("error when posting text to the chat: %s", err.Error())
		return "", err
	}
	defer response.Body.Close()

	var bodyBytes, errRead = ioutil.ReadAll(response.Body)
	if errRead != nil {
		log.Printf("error in parsing telegram answer %s", errRead.Error())
		return "", err
	}
	bodyString := string(bodyBytes)

	return bodyString, nil
}
