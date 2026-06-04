package main

import (
	"fmt"
	"github/sanjay-khandelwal/internal/shared/config"
)

func main() {

	config := config.Get()
	fmt.Printf("App running on port: %s in %s environment\n", config.App.Port, config.App.Env)
}
