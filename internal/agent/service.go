package agent

import (
	"context"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"
)

type Task struct {
	Value   interface{}
	BaseURL string
}

type ClientService struct {
	mutex          sync.Mutex
	memStat        *runtime.MemStats
	client         http.Client
	tasks          []Task
	PollInterval   int
	ReportInterval int
	//extra metrics
	pollCount   *uint64
	randomValue *float64
}

func Init(pollInterval int, reportInterval int) *ClientService {
	stats := runtime.MemStats{}
	var count uint64
	var random float64
	return &ClientService{memStat: &stats,
		tasks: []Task{
			{Value: &stats.Alloc, BaseURL: "http://localhost:8080/update/gauge/Alloc"},
			{Value: &stats.BuckHashSys, BaseURL: "http://localhost:8080/update/gauge/BuckHashSys"},
			{Value: &stats.Frees, BaseURL: "http://localhost:8080/update/gauge/Frees"},
			{Value: &stats.GCCPUFraction, BaseURL: "http://localhost:8080/update/gauge/GCCPUFraction"},
			{Value: &stats.GCSys, BaseURL: "http://localhost:8080/update/gauge/GCSys"},
			{Value: &stats.HeapAlloc, BaseURL: "http://localhost:8080/update/gauge/HeapAlloc"},
			{Value: &stats.HeapIdle, BaseURL: "http://localhost:8080/update/gauge/HeapIdle"},
			{Value: &stats.HeapInuse, BaseURL: "http://localhost:8080/update/gauge/HeapInuse"},
			{Value: &stats.HeapObjects, BaseURL: "http://localhost:8080/update/gauge/HeapObjects"},
			{Value: &stats.HeapReleased, BaseURL: "http://localhost:8080/update/gauge/HeapReleased"},
			{Value: &stats.HeapSys, BaseURL: "http://localhost:8080/update/gauge/HeapSys"},
			{Value: &stats.LastGC, BaseURL: "http://localhost:8080/update/gauge/LastGC"},
			{Value: &stats.Lookups, BaseURL: "http://localhost:8080/update/gauge/Lookups"},
			{Value: &stats.MCacheInuse, BaseURL: "http://localhost:8080/update/gauge/MCacheInuse"},
			{Value: &stats.MCacheSys, BaseURL: "http://localhost:8080/update/gauge/MCacheSys"},
			{Value: &stats.MSpanInuse, BaseURL: "http://localhost:8080/update/gauge/MSpanInuse"},
			{Value: &stats.MSpanSys, BaseURL: "http://localhost:8080/update/gauge/MSpanSys"},
			{Value: &stats.Mallocs, BaseURL: "http://localhost:8080/update/gauge/Mallocs"},
			{Value: &stats.NextGC, BaseURL: "http://localhost:8080/update/gauge/NextGC"},
			{Value: &stats.NumForcedGC, BaseURL: "http://localhost:8080/update/gauge/NumForcedGC"},
			{Value: &stats.NumGC, BaseURL: "http://localhost:8080/update/gauge/NumGC"},
			{Value: &stats.OtherSys, BaseURL: "http://localhost:8080/update/gauge/OtherSys"},
			{Value: &stats.PauseTotalNs, BaseURL: "http://localhost:8080/update/gauge/PauseTotalNs"},
			{Value: &stats.StackInuse, BaseURL: "http://localhost:8080/update/gauge/StackInuse"},
			{Value: &stats.StackSys, BaseURL: "http://localhost:8080/update/gauge/StackSys"},
			{Value: &stats.Sys, BaseURL: "http://localhost:8080/update/gauge/Sys"},
			{Value: &stats.TotalAlloc, BaseURL: "http://localhost:8080/update/gauge/TotalAlloc"},
			{Value: &count, BaseURL: "http://localhost:8080/update/gauge/PollCount"},
			{Value: &random, BaseURL: "http://localhost:8080/update/gauge/RandomValue"},
		},
		client: http.Client{},
		//mutex:          sync.Mutex{},
		pollCount:      &count,
		randomValue:    &random,
		PollInterval:   pollInterval,
		ReportInterval: reportInterval,
	}
}

func (c *ClientService) Update(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			c.mutex.Lock()
			runtime.ReadMemStats(c.memStat)
			*c.pollCount += 1
			*c.randomValue = rand.Float64()
			fmt.Println("metrics updated")
			c.mutex.Unlock()

			time.Sleep(time.Duration(c.PollInterval) * time.Second)
		}
	}

}

func (t *Task) sendTask(client http.Client) {
	var target string
	if v, ok := t.Value.(*uint64); ok {
		target = fmt.Sprintf("%s/%d", t.BaseURL, *v)
	} else if v, ok := t.Value.(*float64); ok {
		target = fmt.Sprintf("%s/%f", t.BaseURL, *v)
	} else {
		target = fmt.Sprintf("%s/unknown", t.BaseURL)
	}

	request, err := http.NewRequest(http.MethodPost, target, nil)
	if err != nil {
		panic(err)
	}
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	io.Copy(os.Stdout, response.Body) // вывод ответа в консоль
	response.Body.Close()
	fmt.Printf("\nposted %s\n", target)
}

func (c *ClientService) Post(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			c.mutex.Lock()
			for _, t := range c.tasks {
				t.sendTask(c.client)
			}
			c.mutex.Unlock()
			time.Sleep(time.Duration(c.ReportInterval) * time.Second)
		}
	}

}
