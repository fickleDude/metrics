package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	runAddr         string
	reportInterval  int
	pollInterval    int
	storeInterval   int
	fileStoragePath string
	restore         bool
}

func NewConfig() *Config {
	return &Config{
		runAddr:         "localhost:8080", //default value
		reportInterval:  10,
		pollInterval:    2,
		storeInterval:   300,
		fileStoragePath: "metrics.txt",
		restore:         false,
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

func (c *Config) StoreInterval() int {
	return c.storeInterval
}

func (c *Config) FileStoragePath() string {
	return c.fileStoragePath
}

func (c *Config) Restore() bool {
	return c.restore
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

func (c *Config) ParseEnv(binary string) {
	envAddr := os.Getenv("ADDRESS")
	err := checkRunAddr(envAddr)
	if err == nil {
		c.runAddr = envAddr
	}
	switch binary {
	case "server":
		envStore := os.Getenv("STORE_INTERVAL")
		envStoreInt, err := strconv.Atoi(envStore)
		if err == nil {
			c.storeInterval = envStoreInt
		}

		envStorePath := os.Getenv("FILE_STORAGE_PATH")
		if envStorePath != "" {
			c.fileStoragePath = envStorePath
		}

		envRestore := os.Getenv("RESTORE")
		envRestoreBool, err := strconv.ParseBool(envRestore)
		if err == nil {
			c.restore = envRestoreBool
		}

	case "agent":
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

}

func (c *Config) ParseFlags(binary string) {
	switch binary {
	case "server":
		server := flag.NewFlagSet("server", flag.ExitOnError)
		server.Func("a", "адрес и порт на котором запущен сервер", func(flagAddr string) error {
			err := checkRunAddr(flagAddr)
			if err == nil {
				c.runAddr = flagAddr

			}
			return nil
		})
		server.IntVar(&c.storeInterval, "i", 300, "частота сохранения показаний метрик в файл в секундах")
		server.StringVar(&c.fileStoragePath, "f", "metrics.txt", "путь до файла для сохранения показаний метрик")
		server.BoolVar(&c.restore, "r", false, "определяет, нужно ли загружать значения из файла при старте сервера")
		server.Parse(os.Args[1:])
	case "agent":
		agent := flag.NewFlagSet("agent", flag.ExitOnError)
		agent.Func("a", "адрес и порт на котором запущен сервер", func(flagAddr string) error {
			err := checkRunAddr(flagAddr)
			if err == nil {
				c.runAddr = flagAddr

			}
			return nil
		})
		agent.IntVar(&c.reportInterval, "r", 10, "частота отправки метрик в секундах")
		agent.IntVar(&c.pollInterval, "p", 2, "частота опроса метрик в секундах")
		agent.Parse(os.Args[1:])
	}
}
