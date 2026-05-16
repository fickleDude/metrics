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

	"github.com/DataDog/gopsutil/cpu"
	"github.com/shirou/gopsutil/v4/mem"

	"github.com/fickleDude/metrics.git/internal/helpers"
	"github.com/fickleDude/metrics.git/internal/logger"
	models "github.com/fickleDude/metrics.git/internal/model"
)

type Agent struct {
	//concurrent
	tasks          []*models.Metrics
	jobs           chan []*models.Metrics
	memStat        *runtime.MemStats
	gopsutilStat   *mem.VirtualMemoryStat
	pollInterval   int
	reportInterval int
	rateLimit      int
	//request
	signer  *helpers.Signer
	client  http.Client
	baseURL string
}

func Init(serverAddress string, pollInterval int, reportInterval int, key string, rateLimit int) *Agent {
	v, _ := mem.VirtualMemory()
	agent := &Agent{
		//concurrent
		tasks: []*models.Metrics{
			{ID: "Alloc", MType: "gauge", Value: nil},
			{ID: "BuckHashSys", MType: "gauge", Value: nil},
			{ID: "Frees", MType: "gauge", Value: nil},
			{ID: "GCCPUFraction", MType: "gauge", Value: nil},
			{ID: "GCSys", MType: "gauge", Value: nil},
			{ID: "HeapAlloc", MType: "gauge", Value: nil},
			{ID: "HeapIdle", MType: "gauge", Value: nil},
			{ID: "HeapInuse", MType: "gauge", Value: nil},
			{ID: "HeapObjects", MType: "gauge", Value: nil},
			{ID: "HeapReleased", MType: "gauge", Value: nil},
			{ID: "HeapSys", MType: "gauge", Value: nil},
			{ID: "LastGC", MType: "gauge", Value: nil},
			{ID: "Lookups", MType: "gauge", Value: nil},
			{ID: "MCacheInuse", MType: "gauge", Value: nil},
			{ID: "MCacheSys", MType: "gauge", Value: nil},
			{ID: "MSpanInuse", MType: "gauge", Value: nil},
			{ID: "MSpanSys", MType: "gauge", Value: nil},
			{ID: "Mallocs", MType: "gauge", Value: nil},
			{ID: "NextGC", MType: "gauge", Value: nil},
			{ID: "NumForcedGC", MType: "gauge", Value: nil},
			{ID: "NumGC", MType: "gauge", Value: nil},
			{ID: "OtherSys", MType: "gauge", Value: nil},
			{ID: "PauseTotalNs", MType: "gauge", Value: nil},
			{ID: "StackInuse", MType: "gauge", Value: nil},
			{ID: "StackSys", MType: "gauge", Value: nil},
			{ID: "Sys", MType: "gauge", Value: nil},
			{ID: "TotalAlloc", MType: "gauge", Value: nil},
			{ID: "PollCount", MType: "counter", Delta: nil},
			{ID: "RandomValue", MType: "gauge", Value: nil},
			{ID: "TotalMemory", MType: "gauge", Value: nil},
			{ID: "FreeMemory", MType: "gauge", Delta: nil},
			{ID: "CPUutilization1", MType: "gauge", Value: nil},
		},
		jobs:           make(chan []*models.Metrics, rateLimit),
		memStat:        &runtime.MemStats{},
		gopsutilStat:   v,
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
		//request
		client:  http.Client{},
		baseURL: fmt.Sprintf("http://%s/updates/", serverAddress),
		signer:  helpers.NewSigner(key),
	}
	if rateLimit == 0 {
		rateLimit = len(agent.tasks)
	}
	return agent
}
func (a *Agent) getMemStatValue(ID string) interface{} {
	var value interface{}
	switch ID {
	// case "PollCount":
	// 	if v, ok := a.tasks[ID].Value.(int); ok {
	// 		value = v + 1
	// 	} else {
	// 		value = 1
	// 	}
	case "RandomValue":
		value = rand.Float64() //c.randomValue
	case "Alloc":
		value = a.memStat.Alloc
	case "BuckHashSys":
		value = a.memStat.BuckHashSys
	case "Frees":
		value = a.memStat.Frees
	case "GCCPUFraction":
		value = a.memStat.GCCPUFraction
	case "GCSys":
		value = a.memStat.GCSys
	case "HeapAlloc":
		value = a.memStat.HeapAlloc
	case "HeapIdle":
		value = a.memStat.HeapIdle
	case "HeapInuse":
		value = a.memStat.HeapInuse
	case "HeapObjects":
		value = a.memStat.HeapObjects
	case "HeapReleased":
		value = a.memStat.HeapReleased
	case "HeapSys":
		value = a.memStat.HeapSys
	case "LastGC":
		value = a.memStat.LastGC
	case "Lookups":
		value = a.memStat.Lookups
	case "MCacheInuse":
		value = a.memStat.MCacheInuse
	case "MCacheSys":
		value = a.memStat.MCacheSys
	case "MSpanInuse":
		value = a.memStat.MSpanInuse
	case "MSpanSys":
		value = a.memStat.MSpanSys
	case "Mallocs":
		value = a.memStat.Mallocs
	case "NextGC":
		value = a.memStat.NextGC
	case "NumForcedGC":
		value = a.memStat.NumForcedGC
	case "NumGC":
		value = a.memStat.NumGC
	case "OtherSys":
		value = a.memStat.OtherSys
	case "PauseTotalNs":
		value = a.memStat.PauseTotalNs
	case "StackInuse":
		value = a.memStat.StackInuse
	case "StackSys":
		value = a.memStat.StackSys
	case "Sys":
		value = a.memStat.Sys
	case "TotalAlloc":
		value = a.memStat.TotalAlloc
	case "TotalMemory":
		value = a.gopsutilStat.Total
	case "FreeMemory":
		value = a.gopsutilStat.Free
	case "CPUutilization1":
		value, _ = cpu.Counts(true)

	default:
		value = nil
	}
	return value
}

func (a *Agent) Update(ctx context.Context, wg *sync.WaitGroup) {
	ticker := time.NewTicker(time.Duration(a.pollInterval) * time.Second)
	defer ticker.Stop()
	defer close(a.jobs)
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fmt.Println("updater ready")

			runtime.ReadMemStats(a.memStat)
			for _, t := range a.tasks {
				if t.ID == "PollCount" {

					if t.Delta == nil {
						value := int64(1)
						t.Delta = &value
					} else {
						value := *t.Delta + 1
						t.Delta = &value
					}
				} else {
					value := a.getMemStatValue(t.ID)
					t.SetValue(value)
				}
			}

			iter := (len(a.tasks) / a.rateLimit)
			if len(a.tasks)%a.rateLimit != 0 {
				iter++
			}
			for j := 0; j < min(a.rateLimit, len(a.tasks)); j++ {
				fmt.Println("[", j*iter, ",", min(j*iter+iter, len(a.tasks)), "]")
				a.jobs <- a.tasks[j*iter : min(j*iter+iter, len(a.tasks))]
			}
		}
	}
}

func isRetriable(err error) bool {
	var netErr *net.OpError
	if errors.As(err, &netErr) {
		return netErr.Temporary() || netErr.Timeout()
	}
	return false
}

func (a *Agent) sendTask(metric []*models.Metrics) error {
	for _, m := range metric {
		fmt.Println(m.ID, " updated")
	}

	//encode response
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(metric); err != nil {
		logger.Log.Error(err.Error())
		return err
	}
	//gzip
	var gzBuf bytes.Buffer
	gz := gzip.NewWriter(&gzBuf)
	_, err := gz.Write(buf.Bytes())
	if err != nil {
		logger.Log.Error(err.Error())
		return err
	}
	gz.Close()

	//create request
	request, err := http.NewRequest(http.MethodPost, a.baseURL, &gzBuf)
	if err != nil {
		logger.Log.Error(err.Error())
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Content-Encoding", "gzip")
	a.signer.SignRequest(buf.Bytes(), request)
	//get response
	maxRetries := 3
	retryDelay := []int{1, 3, 5}
	var response *http.Response
	for attempt := 0; attempt < maxRetries; attempt++ {
		response, err = a.client.Do(request)
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
		return err
	}
	response.Body.Close()
	return nil
}

func (a *Agent) Post(ctx context.Context, id int, wg *sync.WaitGroup) error {
	ticker := time.NewTicker(time.Duration(a.reportInterval) * time.Second)
	defer ticker.Stop()
	defer wg.Done()
	for j := range a.jobs {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			err := a.sendTask(j)
			if err != nil {
				return err
			} else {
				fmt.Println("рабочий", id, "запущен задача", j)
			}
		default:
			fmt.Println("рабочий", id, "ждет")
		}

	}
	return nil
}
