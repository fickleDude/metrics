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
}

func (c *Config) ParseFlags() {
	//both
	flag.Func("a", "адрес и порт на котором запущен сервер", func(flagAddr string) error {
		err := checkRunAddr(flagAddr)
		if err == nil {
			c.runAddr = flagAddr

		}
		return nil
	})
	//agent
	flag.IntVar(&c.reportInterval, "ri", 10, "частота отправки метрик в секундах")
	flag.IntVar(&c.pollInterval, "p", 2, "частота опроса метрик в секундах")
	//server
	flag.IntVar(&c.storeInterval, "i", 300, "частота сохранения показаний метрик в файл в секундах")
	flag.StringVar(&c.fileStoragePath, "f", "metrics.txt", "путь до файла для сохранения показаний метрик")
	flag.BoolVar(&c.restore, "r", false, "определяет, нужно ли загружать значения из файла при старте сервера")

	flag.Parse()
}
