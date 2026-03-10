package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
)

var runAddr string

func parseFlags() {
	//set default value
	runAddr = "localhost:8080"
	flag.Func("a", "адрес и порт на котором запущен сервер", func(flagValue string) error {
		params := strings.Split(flagValue, ":")
		if len(params) < 2 {
			return fmt.Errorf("формат флага адрес:порт")
		}
		_, err := strconv.Atoi(params[1])
		if err != nil {
			return fmt.Errorf("порт указан некорректно")
		}
		runAddr = flagValue
		return nil
	})

	flag.Parse()

}
