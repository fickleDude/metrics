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
	BaseUrl string
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
			{Value: &stats.Alloc, BaseUrl: "http://localhost:8080/update/gauge/Alloc"},
			{Value: &stats.BuckHashSys, BaseUrl: "http://localhost:8080/update/gauge/BuckHashSys"},
			{Value: &stats.Frees, BaseUrl: "http://localhost:8080/update/gauge/Frees"},
			{Value: &stats.GCCPUFraction, BaseUrl: "http://localhost:8080/update/gauge/GCCPUFraction"},
			{Value: &stats.GCSys, BaseUrl: "http://localhost:8080/update/gauge/GCSys"},
			{Value: &stats.HeapAlloc, BaseUrl: "http://localhost:8080/update/gauge/HeapAlloc"},
			{Value: &stats.HeapIdle, BaseUrl: "http://localhost:8080/update/gauge/HeapIdle"},
			{Value: &stats.HeapInuse, BaseUrl: "http://localhost:8080/update/gauge/HeapInuse"},
			{Value: &stats.HeapObjects, BaseUrl: "http://localhost:8080/update/gauge/HeapObjects"},
			{Value: &stats.HeapReleased, BaseUrl: "http://localhost:8080/update/gauge/HeapReleased"},
			{Value: &stats.HeapSys, BaseUrl: "http://localhost:8080/update/gauge/HeapSys"},
			{Value: &stats.LastGC, BaseUrl: "http://localhost:8080/update/gauge/LastGC"},
			{Value: &stats.Lookups, BaseUrl: "http://localhost:8080/update/gauge/Lookups"},
			{Value: &stats.MCacheInuse, BaseUrl: "http://localhost:8080/update/gauge/MCacheInuse"},
			{Value: &stats.MCacheSys, BaseUrl: "http://localhost:8080/update/gauge/MCacheSys"},
			{Value: &stats.MSpanInuse, BaseUrl: "http://localhost:8080/update/gauge/MSpanInuse"},
			{Value: &stats.MSpanSys, BaseUrl: "http://localhost:8080/update/gauge/MSpanSys"},
			{Value: &stats.Mallocs, BaseUrl: "http://localhost:8080/update/gauge/Mallocs"},
			{Value: &stats.NextGC, BaseUrl: "http://localhost:8080/update/gauge/NextGC"},
			{Value: &stats.NumForcedGC, BaseUrl: "http://localhost:8080/update/gauge/NumForcedGC"},
			{Value: &stats.NumGC, BaseUrl: "http://localhost:8080/update/gauge/NumGC"},
			{Value: &stats.OtherSys, BaseUrl: "http://localhost:8080/update/gauge/OtherSys"},
			{Value: &stats.PauseTotalNs, BaseUrl: "http://localhost:8080/update/gauge/PauseTotalNs"},
			{Value: &stats.StackInuse, BaseUrl: "http://localhost:8080/update/gauge/StackInuse"},
			{Value: &stats.StackSys, BaseUrl: "http://localhost:8080/update/gauge/StackSys"},
			{Value: &stats.Sys, BaseUrl: "http://localhost:8080/update/gauge/Sys"},
			{Value: &stats.TotalAlloc, BaseUrl: "http://localhost:8080/update/gauge/TotalAlloc"},
			{Value: &count, BaseUrl: "http://localhost:8080/update/gauge/PollCount"},
			{Value: &random, BaseUrl: "http://localhost:8080/update/gauge/RandomValue"},
		},
		client:         http.Client{},
		mutex:          sync.Mutex{},
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
		target = fmt.Sprintf("%s/%d", t.BaseUrl, *v)
	} else if v, ok := t.Value.(*float64); ok {
		target = fmt.Sprintf("%s/%f", t.BaseUrl, *v)
	} else {
		target = fmt.Sprintf("%s/unknown", t.BaseUrl)
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
