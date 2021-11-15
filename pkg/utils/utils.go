package utils

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/preichenberger/go-coinbasepro/v2"
)

type currencyConfig struct {
	Symbol     string
	Percentage int
}

type Config struct {
	BaseUrl        string `env:"BASE_URL"`
	Key            string `env:"KEY"`
	Secret         string `env:"SECRET"`
	Passphrase     string `env:"PASSPHRASE"`
	Cron           string `env:"CRON"`
	CurrencyConfig []currencyConfig
}

func GetEnvConfig() Config {
	config := Config{
		BaseUrl:    os.Getenv("BASE_URL"),
		Key:        os.Getenv("KEY"),
		Secret:     os.Getenv("SECRET"),
		Passphrase: os.Getenv("PASSPHRASE"),
		Cron:       os.Getenv("CRON"),
	}
	currencyString := os.Getenv("CURRENCY_CONFIG")

	if len(currencyString) > 0 {
		var currencies []currencyConfig
		currencyList := strings.Split(os.Getenv("CURRENCY_CONFIG"), ";")
		for _, v := range currencyList {
			currConfig := strings.Split(v, ":")
			percentage, err := strconv.Atoi(currConfig[1])

			if err != nil {
				log.Fatal("Error: ", err)
			}

			currencies = append(currencies, currencyConfig{
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

func getCoinbaseClient(c Config) *coinbasepro.Client {
	client := coinbasepro.NewClient()
	client.UpdateConfig(&coinbasepro.ClientConfig{
		BaseURL:    c.BaseUrl,
		Key:        c.Key,
		Passphrase: c.Passphrase,
		Secret:     c.Secret,
	})

	return client
}

type SubscribeOptions struct {
	Base   string
	Target string
}

func SubscribeToCurrency(opts SubscribeOptions) {
	var wsDialer = websocket.DefaultDialer
	wsConn, _, err := wsDialer.Dial("wss://ws-feed.pro.coinbase.com", nil)

	if err != nil {
		log.Fatal(err)
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
			println("PRODUCT ID:", message.ProductID)
			println("LAST:", message.LastSize)
			println("PRICE: ", message.Price)
			println("BEST BID: ", message.BestBid)
			println()
			time.Sleep(time.Second * 1)
		}
	}
}
