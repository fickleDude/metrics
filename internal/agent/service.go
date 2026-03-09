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
	Value interface{}
	Url   string
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

func Init(serverAddress string, pollInterval int, reportInterval int) *ClientService {
	stats := runtime.MemStats{}
	var count uint64
	var random float64
	baseURL := fmt.Sprintf("http://%s/update", serverAddress)
	return &ClientService{memStat: &stats,
		tasks: []Task{
			{Value: &stats.Alloc, Url: baseURL + "/gauge/Alloc"},
			{Value: &stats.BuckHashSys, Url: baseURL + "/gauge/BuckHashSys"},
			{Value: &stats.Frees, Url: baseURL + "/gauge/Frees"},
			{Value: &stats.GCCPUFraction, Url: baseURL + "/gauge/GCCPUFraction"},
			{Value: &stats.GCSys, Url: baseURL + "/gauge/GCSys"},
			{Value: &stats.HeapAlloc, Url: baseURL + "/gauge/HeapAlloc"},
			{Value: &stats.HeapIdle, Url: baseURL + "/gauge/HeapIdle"},
			{Value: &stats.HeapInuse, Url: baseURL + "/gauge/HeapInuse"},
			{Value: &stats.HeapObjects, Url: baseURL + "/gauge/HeapObjects"},
			{Value: &stats.HeapReleased, Url: baseURL + "/gauge/HeapReleased"},
			{Value: &stats.HeapSys, Url: baseURL + "/gauge/HeapSys"},
			{Value: &stats.LastGC, Url: baseURL + "/gauge/LastGC"},
			{Value: &stats.Lookups, Url: baseURL + "/gauge/Lookups"},
			{Value: &stats.MCacheInuse, Url: baseURL + "/gauge/MCacheInuse"},
			{Value: &stats.MCacheSys, Url: baseURL + "/gauge/MCacheSys"},
			{Value: &stats.MSpanInuse, Url: baseURL + "/gauge/MSpanInuse"},
			{Value: &stats.MSpanSys, Url: baseURL + "/gauge/MSpanSys"},
			{Value: &stats.Mallocs, Url: baseURL + "/gauge/Mallocs"},
			{Value: &stats.NextGC, Url: baseURL + "/gauge/NextGC"},
			{Value: &stats.NumForcedGC, Url: baseURL + "/gauge/NumForcedGC"},
			{Value: &stats.NumGC, Url: baseURL + "/gauge/NumGC"},
			{Value: &stats.OtherSys, Url: baseURL + "/gauge/OtherSys"},
			{Value: &stats.PauseTotalNs, Url: baseURL + "/gauge/PauseTotalNs"},
			{Value: &stats.StackInuse, Url: baseURL + "/gauge/StackInuse"},
			{Value: &stats.StackSys, Url: baseURL + "/gauge/StackSys"},
			{Value: &stats.Sys, Url: baseURL + "/gauge/Sys"},
			{Value: &stats.TotalAlloc, Url: baseURL + "/gauge/TotalAlloc"},
			{Value: &count, Url: baseURL + "/counter/PollCount"},
			{Value: &random, Url: baseURL + "/gauge/RandomValue"},
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
		target = fmt.Sprintf("%s/%d", t.Url, *v)
	} else if v, ok := t.Value.(*float64); ok {
		target = fmt.Sprintf("%s/%f", t.Url, *v)
	} else {
		target = fmt.Sprintf("%s/unknown", t.Url)
	}

	request, err := http.NewRequest(http.MethodPost, target, nil)
	if err != nil {
		return
	}
	response, err := client.Do(request)
	if err != nil {
		return
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
