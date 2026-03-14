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
	URL   string
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
			{Value: &stats.Alloc, URL: baseURL + "/gauge/Alloc"},
			{Value: &stats.BuckHashSys, URL: baseURL + "/gauge/BuckHashSys"},
			{Value: &stats.Frees, URL: baseURL + "/gauge/Frees"},
			{Value: &stats.GCCPUFraction, URL: baseURL + "/gauge/GCCPUFraction"},
			{Value: &stats.GCSys, URL: baseURL + "/gauge/GCSys"},
			{Value: &stats.HeapAlloc, URL: baseURL + "/gauge/HeapAlloc"},
			{Value: &stats.HeapIdle, URL: baseURL + "/gauge/HeapIdle"},
			{Value: &stats.HeapInuse, URL: baseURL + "/gauge/HeapInuse"},
			{Value: &stats.HeapObjects, URL: baseURL + "/gauge/HeapObjects"},
			{Value: &stats.HeapReleased, URL: baseURL + "/gauge/HeapReleased"},
			{Value: &stats.HeapSys, URL: baseURL + "/gauge/HeapSys"},
			{Value: &stats.LastGC, URL: baseURL + "/gauge/LastGC"},
			{Value: &stats.Lookups, URL: baseURL + "/gauge/Lookups"},
			{Value: &stats.MCacheInuse, URL: baseURL + "/gauge/MCacheInuse"},
			{Value: &stats.MCacheSys, URL: baseURL + "/gauge/MCacheSys"},
			{Value: &stats.MSpanInuse, URL: baseURL + "/gauge/MSpanInuse"},
			{Value: &stats.MSpanSys, URL: baseURL + "/gauge/MSpanSys"},
			{Value: &stats.Mallocs, URL: baseURL + "/gauge/Mallocs"},
			{Value: &stats.NextGC, URL: baseURL + "/gauge/NextGC"},
			{Value: &stats.NumForcedGC, URL: baseURL + "/gauge/NumForcedGC"},
			{Value: &stats.NumGC, URL: baseURL + "/gauge/NumGC"},
			{Value: &stats.OtherSys, URL: baseURL + "/gauge/OtherSys"},
			{Value: &stats.PauseTotalNs, URL: baseURL + "/gauge/PauseTotalNs"},
			{Value: &stats.StackInuse, URL: baseURL + "/gauge/StackInuse"},
			{Value: &stats.StackSys, URL: baseURL + "/gauge/StackSys"},
			{Value: &stats.Sys, URL: baseURL + "/gauge/Sys"},
			{Value: &stats.TotalAlloc, URL: baseURL + "/gauge/TotalAlloc"},
			{Value: &count, URL: baseURL + "/counter/PollCount"},
			{Value: &random, URL: baseURL + "/gauge/RandomValue"},
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
		target = fmt.Sprintf("%s/%d", t.URL, *v)
	} else if v, ok := t.Value.(*float64); ok {
		target = fmt.Sprintf("%s/%f", t.URL, *v)
	} else {
		target = fmt.Sprintf("%s/unknown", t.URL)
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
