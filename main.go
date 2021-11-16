package main

import (
	"fmt"

	"github.com/emwp/coinbase-dca/pkg/utils"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Errorf("Found Error: %w", err)
	}

	utils.StartCronBuy()
}
