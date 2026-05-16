package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

const (
	Server = "server"
	Agent  = "agent"
)

type Config struct {
	runAddr         string
	reportInterval  int
	pollInterval    int
	storeInterval   int
	fileStoragePath string
	restore         bool
	databaseDNS     string
	key             string
	rateLimit       int
}

var (
	agentConfig  *Config
	serverConfig *Config
	initServer   sync.Once
	initAgent    sync.Once
)

func GetConfig(binary string) *Config {
	switch binary {
	case Server:
		initServer.Do(func() {
			serverConfig = &Config{
				runAddr:        "localhost:8080",
				reportInterval: 10,
				pollInterval:   2,
				storeInterval:  300,
				restore:        false,
			}
			serverConfig.parseFlags(binary)
			serverConfig.parseEnv(binary)
		})
		return serverConfig
	case Agent:
		initAgent.Do(func() {
			agentConfig = &Config{
				runAddr:        "localhost:8080",
				reportInterval: 10,
				pollInterval:   2,
				storeInterval:  300,
				restore:        false,
				rateLimit:      32,
			}
			agentConfig.parseFlags(binary)
			agentConfig.parseEnv(binary)
		})
		return agentConfig
	default:
		return &Config{}
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

func (c *Config) DatabaseDNS() string {
	return c.databaseDNS
}

func (c *Config) Key() string {
	return c.key
}

func (c *Config) RateLimit() int {
	return c.rateLimit
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

func (c *Config) parseEnv(binary string) {
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

		databaseDNS := os.Getenv("DATABASE_DSN")
		if databaseDNS != "" {
			c.databaseDNS = databaseDNS
		}

		envKey := os.Getenv("KEY")
		if envKey != "" {
			c.key = envKey
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

		envKey := os.Getenv("KEY")
		if envKey != "" {
			c.key = envKey
		}

		rateLimit := os.Getenv("RATE_LIMIT")
		rateLimitInt, err := strconv.Atoi(rateLimit)
		if err == nil {
			c.rateLimit = rateLimitInt
		}
	}

}

func (c *Config) parseFlags(binary string) {
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
		server.StringVar(&c.databaseDNS, "d", "", "строка подключения к СУБД")
		server.StringVar(&c.key, "k", "", "ключ подписи")
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
		agent.StringVar(&c.key, "k", "", "ключ подписи")
		agent.IntVar(&c.rateLimit, "l", 32, "количество одновременно исходящих на сервер запросов")
		agent.Parse(os.Args[1:])
	}
}
