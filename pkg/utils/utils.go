package utils

import (
	"log"
	"os"
	"strconv"
	"strings"

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
