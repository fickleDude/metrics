package main

import (
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

func post(id int, jobs <-chan []int, doneCh chan struct{}, wg *sync.WaitGroup) error {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for j := range jobs {
		select {
		case <-doneCh:
			wg.Done()
			if id == 3 {
				return fmt.Errorf("check")
			}
			return nil
		case <-ticker.C:
			fmt.Println("рабочий", id, "запущен задача", j)
		default:
			fmt.Println("рабочий", id, "ждет")
		}

	}
	if id == 3 {
		return fmt.Errorf("check")
	}
	return nil
}
func update(doneCh chan struct{}, jobs chan<- []int, rateLimits int, values [13]int, wg *sync.WaitGroup) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	defer close(jobs)
	for {
		select {
		case <-doneCh:
			wg.Done()
			return
		case <-ticker.C:
			fmt.Println("updater ready")
			iter := (len(values) / rateLimits)
			if len(values)%rateLimits != 0 {
				iter++
			}
			for j := 0; j < rateLimits; j++ {
				jobs <- values[j*iter : min(j*iter+iter, len(values))]
			}
		}
	}
}

func main() {
	const rateLimit = 5
	values := [13]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13}
	doneCh := make(chan struct{}, 1)
	g := new(errgroup.Group)

	var wg sync.WaitGroup
	jobs := make(chan []int, rateLimit)
	for w := 1; w <= rateLimit; w++ {
		wg.Add(1)
		g.Go(func() error {
			return post(w, jobs, doneCh, &wg)
		})
	}

	wg.Add(1)
	go update(doneCh, jobs, rateLimit, values, &wg)

	<-doneCh
	if err := g.Wait(); err != nil {
		fmt.Println(err)
	}
	wg.Wait()
}
