package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/fickleDude/metrics.git/internal/logger"
	models "github.com/fickleDude/metrics.git/internal/model"
)

type Task struct {
	Type  string
	Value interface{}
}

type ClientService struct {
	mutex   sync.Mutex
	memStat *runtime.MemStats
	client  http.Client
	baseURL string
	tasks   map[string]*Task
	//extra
	pollCount      *int64
	randomValue    *float64
	pollInterval   int
	reportInterval int
}

func Init(serverAddress string, pollInterval int, reportInterval int) *ClientService {
	//get initial stat
	memStat := runtime.MemStats{}
	var count int64
	var random float64

	return &ClientService{
		mutex:   sync.Mutex{},
		memStat: &memStat,
		client:  http.Client{},
		baseURL: fmt.Sprintf("http://%s/update/", serverAddress),
		tasks: map[string]*Task{
			"Alloc":         {Value: &memStat.Alloc, Type: "gauge"},
			"BuckHashSys":   {Value: &memStat.BuckHashSys, Type: "gauge"},
			"Frees":         {Value: &memStat.Frees, Type: "gauge"},
			"GCCPUFraction": {Value: &memStat.GCCPUFraction, Type: "gauge"},
			"GCSys":         {Value: &memStat.GCSys, Type: "gauge"},
			"HeapAlloc":     {Value: &memStat.HeapAlloc, Type: "gauge"},
			"HeapIdle":      {Value: &memStat.HeapIdle, Type: "gauge"},
			"HeapInuse":     {Value: &memStat.HeapInuse, Type: "gauge"},
			"HeapObjects":   {Value: &memStat.HeapObjects, Type: "gauge"},
			"HeapReleased":  {Value: &memStat.HeapReleased, Type: "gauge"},
			"HeapSys":       {Value: &memStat.HeapSys, Type: "gauge"},
			"LastGC":        {Value: &memStat.LastGC, Type: "gauge"},
			"Lookups":       {Value: &memStat.Lookups, Type: "gauge"},
			"MCacheInuse":   {Value: &memStat.MCacheInuse, Type: "gauge"},
			"MCacheSys":     {Value: &memStat.MCacheSys, Type: "gauge"},
			"MSpanInuse":    {Value: &memStat.MSpanInuse, Type: "gauge"},
			"MSpanSys":      {Value: &memStat.MSpanSys, Type: "gauge"},
			"Mallocs":       {Value: &memStat.Mallocs, Type: "gauge"},
			"NextGC":        {Value: &memStat.NextGC, Type: "gauge"},
			"NumForcedGC":   {Value: &memStat.NumForcedGC, Type: "gauge"},
			"NumGC":         {Value: &memStat.NumGC, Type: "gauge"},
			"OtherSys":      {Value: &memStat.OtherSys, Type: "gauge"},
			"PauseTotalNs":  {Value: &memStat.PauseTotalNs, Type: "gauge"},
			"StackInuse":    {Value: &memStat.StackInuse, Type: "gauge"},
			"StackSys":      {Value: &memStat.StackSys, Type: "gauge"},
			"Sys":           {Value: &memStat.Sys, Type: "gauge"},
			"TotalAlloc":    {Value: &memStat.TotalAlloc, Type: "gauge"},
			"PollCount":     {Value: &count, Type: "counter"},
			"RandomValue":   {Value: &random, Type: "gauge"},
		},
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
		pollCount:      &count,
		randomValue:    &random,
	}
}

func (c *ClientService) Update(ctx context.Context, wg *sync.WaitGroup) {
	for {
		select {
		case <-ctx.Done():
			wg.Done()
			return
		default:
			c.mutex.Lock()
			//read stat
			runtime.ReadMemStats(c.memStat)
			//update metric
			*c.pollCount += 1
			*c.randomValue = rand.Float64()

			c.mutex.Unlock()
			//sleep
			time.Sleep(time.Duration(c.pollInterval) * time.Second)
		}
	}

}

func sendTask(client http.Client, target string, metric models.Metrics) {
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
	request, err := http.NewRequest(http.MethodPost, target, &gzBuf)
	if err != nil {
		logger.Log.Error(err.Error())
		return
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Content-Encoding", "gzip")

	//get response
	response, err := client.Do(request)
	if err != nil {
		logger.Log.Error(err.Error())
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
			c.mutex.Lock()
			snap := c.tasks
			c.mutex.Unlock()
			for k, v := range snap {
				//create metric
				var metric models.Metrics
				metric.ID = k
				metric.MType = v.Type
				err := metric.SetValue(v.Value)
				if err != nil {
					logger.Log.Error(err.Error())
					continue
				}
				//make request
				sendTask(c.client, c.baseURL, metric)
			}
			time.Sleep(time.Duration(c.reportInterval) * time.Second)
		}
	}

}
