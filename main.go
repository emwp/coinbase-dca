package main

import (
	"fmt"
	"log"

	"github.com/emwp/coinbase-dca/pkg/utils"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	config := utils.GetEnvConfig()
	fmt.Println(config)

	// client := getCoinbaseClient(config)

	// accounts, err := client.GetAccounts()
	// if err != nil {
	// 	println("Error: ", err)
	// }

	// for _, a := range accounts {

	// 	if i, err := strconv.ParseFloat(a.Balance, 64); err == nil {
	// 		if i > 0 {
	// 			fmt.Printf("[%v] - %v\n", a.Currency, a.Balance)
	// 		}
	// 	} else {
	// 		fmt.Println("Error: ", err)
	// 		return
	// 	}
	// }
}
