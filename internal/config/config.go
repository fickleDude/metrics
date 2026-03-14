package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	runAddr        string
	reportInterval int
	pollInterval   int
}

func NewConfig() *Config {
	return &Config{
		runAddr:        "localhost:8080", //default value
		reportInterval: 10,
		pollInterval:   2,
	}
}

func (c *Config) RunAddr() string {
	return c.runAddr
}

func (c *Config) ReportInterval() int {
	return c.reportInterval
}

func (c *Config) PollInterval() int {
	return c.pollInterval
}

func checkRunAddr(addr string) error {
	params := strings.Split(addr, ":")
	if len(params) < 2 {
		return fmt.Errorf("формат флага адрес:порт")
	}
	_, err := strconv.Atoi(params[1])
	if err != nil {
		return fmt.Errorf("порт указан некорректно")
	}
	return nil
}

func (c *Config) ParseEnv() {
	envAddr := os.Getenv("ADDRESS")
	err := checkRunAddr(envAddr)
	if err == nil {
		c.runAddr = envAddr
	}

	envReport := os.Getenv("REPORT_INTERVAL")
	envReportInt, err := strconv.Atoi(envReport)
	if err == nil {
		c.reportInterval = envReportInt
	}

	envPoll := os.Getenv("POLL_INTERVAL")
	envPollInt, err := strconv.Atoi(envPoll)
	if err == nil {
		c.pollInterval = envPollInt
	}

}

func (c *Config) ParseFlags() {
	flag.Func("a", "адрес и порт на котором запущен сервер", func(flagAddr string) error {
		err := checkRunAddr(flagAddr)
		if err == nil {
			c.runAddr = flagAddr

		}
		return nil
	})
	flag.IntVar(&c.reportInterval, "r", 10, "частота отправки метрик в секундах")
	flag.IntVar(&c.pollInterval, "p", 2, "частота опроса метрик в секундах")

	flag.Parse()
}
