package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"net"

	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/fickleDude/metrics.git/internal/helpers"
	"github.com/fickleDude/metrics.git/internal/logger"
	models "github.com/fickleDude/metrics.git/internal/model"
)

type Task struct {
	Type  string
	Value interface{}
}

type ClientService struct {
	mutex        sync.RWMutex
	memStat      *runtime.MemStats
	client       http.Client
	baseURL      string
	tasks        map[string]*Task
	pollTicker   *time.Ticker
	reportTicker *time.Ticker
	signer       *helpers.Signer
}

func Init(serverAddress string, pollInterval int, reportInterval int, key string) *ClientService {
	//get initial stat
	memStat := runtime.MemStats{}

	return &ClientService{
		mutex:   sync.RWMutex{},
		memStat: &memStat,
		client:  http.Client{},
		baseURL: fmt.Sprintf("http://%s/updates/", serverAddress),
		tasks: map[string]*Task{
			"Alloc":         {Value: nil, Type: "gauge"},
			"BuckHashSys":   {Value: nil, Type: "gauge"},
			"Frees":         {Value: nil, Type: "gauge"},
			"GCCPUFraction": {Value: nil, Type: "gauge"},
			"GCSys":         {Value: nil, Type: "gauge"},
			"HeapAlloc":     {Value: nil, Type: "gauge"},
			"HeapIdle":      {Value: nil, Type: "gauge"},
			"HeapInuse":     {Value: nil, Type: "gauge"},
			"HeapObjects":   {Value: nil, Type: "gauge"},
			"HeapReleased":  {Value: nil, Type: "gauge"},
			"HeapSys":       {Value: nil, Type: "gauge"},
			"LastGC":        {Value: nil, Type: "gauge"},
			"Lookups":       {Value: nil, Type: "gauge"},
			"MCacheInuse":   {Value: nil, Type: "gauge"},
			"MCacheSys":     {Value: nil, Type: "gauge"},
			"MSpanInuse":    {Value: nil, Type: "gauge"},
			"MSpanSys":      {Value: nil, Type: "gauge"},
			"Mallocs":       {Value: nil, Type: "gauge"},
			"NextGC":        {Value: nil, Type: "gauge"},
			"NumForcedGC":   {Value: nil, Type: "gauge"},
			"NumGC":         {Value: nil, Type: "gauge"},
			"OtherSys":      {Value: nil, Type: "gauge"},
			"PauseTotalNs":  {Value: nil, Type: "gauge"},
			"StackInuse":    {Value: nil, Type: "gauge"},
			"StackSys":      {Value: nil, Type: "gauge"},
			"Sys":           {Value: nil, Type: "gauge"},
			"TotalAlloc":    {Value: nil, Type: "gauge"},
			"PollCount":     {Value: nil, Type: "counter"},
			"RandomValue":   {Value: nil, Type: "gauge"},
		},
		pollTicker:   time.NewTicker(time.Duration(pollInterval) * time.Second),
		reportTicker: time.NewTicker(time.Duration(reportInterval) * time.Second),
		signer:       helpers.NewSigner(key),
	}
}

func (c *ClientService) Update(ctx context.Context, wg *sync.WaitGroup) {
	for {
		select {
		case <-ctx.Done():
			wg.Done()
			return
		default:
			//read stat
			runtime.ReadMemStats(c.memStat)
			//update metric
			for k := range c.tasks {
				c.setValue(k)
			}
			//sleep
			<-c.pollTicker.C
		}
	}

}

func (c *ClientService) setValue(name string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	var value interface{}
	switch name {
	case "PollCount":
		if v, ok := c.tasks[name].Value.(int); ok {
			value = v + 1
		} else {
			value = 1
		}
	case "RandomValue":
		value = rand.Float64() //c.randomValue
	case "Alloc":
		value = c.memStat.Alloc
	case "BuckHashSys":
		value = c.memStat.BuckHashSys
	case "Frees":
		value = c.memStat.Frees
	case "GCCPUFraction":
		value = c.memStat.GCCPUFraction
	case "GCSys":
		value = c.memStat.GCSys
	case "HeapAlloc":
		value = c.memStat.HeapAlloc
	case "HeapIdle":
		value = c.memStat.HeapIdle
	case "HeapInuse":
		value = c.memStat.HeapInuse
	case "HeapObjects":
		value = c.memStat.HeapObjects
	case "HeapReleased":
		value = c.memStat.HeapReleased
	case "HeapSys":
		value = c.memStat.HeapSys
	case "LastGC":
		value = c.memStat.LastGC
	case "Lookups":
		value = c.memStat.Lookups
	case "MCacheInuse":
		value = c.memStat.MCacheInuse
	case "MCacheSys":
		value = c.memStat.MCacheSys
	case "MSpanInuse":
		value = c.memStat.MSpanInuse
	case "MSpanSys":
		value = c.memStat.MSpanSys
	case "Mallocs":
		value = c.memStat.Mallocs
	case "NextGC":
		value = c.memStat.NextGC
	case "NumForcedGC":
		value = c.memStat.NumForcedGC
	case "NumGC":
		value = c.memStat.NumGC
	case "OtherSys":
		value = c.memStat.OtherSys
	case "PauseTotalNs":
		value = c.memStat.PauseTotalNs
	case "StackInuse":
		value = c.memStat.StackInuse
	case "StackSys":
		value = c.memStat.StackSys
	case "Sys":
		value = c.memStat.Sys
	case "TotalAlloc":
		value = c.memStat.TotalAlloc
	default:
		value = nil
	}
	c.tasks[name].Value = value
}

func (c *ClientService) getValue(name string) interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.tasks[name].Value
}

func isRetriable(err error) bool {
	var netErr *net.OpError
	if errors.As(err, &netErr) {
		return netErr.Temporary() || netErr.Timeout()
	}
	return false
}

func (c *ClientService) sendTask(metric []models.Metrics) {
	//encode response
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(metric); err != nil {
		logger.Log.Error(err.Error())
		return
	}
	//gzip
	var gzBuf bytes.Buffer
	gz := gzip.NewWriter(&gzBuf)
	_, err := gz.Write(buf.Bytes())
	if err != nil {
		logger.Log.Error(err.Error())
		return
	}
	gz.Close()
	//create request
	request, err := http.NewRequest(http.MethodPost, c.baseURL, &gzBuf)
	if err != nil {
		logger.Log.Error(err.Error())
		return
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Content-Encoding", "gzip")
	c.signer.SignRequest(buf.Bytes(), request)
	//get response
	maxRetries := 3
	retryDelay := []int{1, 3, 5}
	var response *http.Response
	for attempt := 0; attempt < maxRetries; attempt++ {
		response, err = c.client.Do(request)
		if err == nil {
			break
		}
		if isRetriable(err) {
			time.Sleep(time.Duration(retryDelay[attempt]) * time.Second)
			continue
		} else {
			break
		}
	}
	if err != nil {
		logger.Log.Error(fmt.Sprintf("Failed after %d attempts: %v", maxRetries, err))
		return
	}
	response.Body.Close()
}

func (c *ClientService) Post(ctx context.Context, wg *sync.WaitGroup) {
	for {
		select {
		case <-ctx.Done():
			wg.Done()
			return
		default:
			var metrics []models.Metrics
			for k, v := range c.tasks {
				//create metric
				var metric models.Metrics
				metric.ID = k
				metric.MType = v.Type
				value := c.getValue(k)
				err := metric.SetValue(value)
				if err != nil {
					logger.Log.Error(err.Error())
					continue
				}
				//add to request
				metrics = append(metrics, metric)
			}
			if len(metrics) > 0 {
				c.sendTask(metrics)
			}

			<-c.reportTicker.C
		}
	}

}
