package internals

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type RequestMeta struct {
	RType     string `json:"type"`
	Path      string `json:"path"`
	Duration  string `json:"duration"`
	SpawnRate int    `json:"spawnRatePerSecond"`
	Vus       int    `json:"vuMax"`
}

type Configuration struct {
	Requests []RequestMeta `json:"requests"`
	Url      string        `json:"url"`
}

func (c *Configuration) ReadConfig(path string) error {
	file, err := os.ReadFile(path)

	if err != nil {
		return err
	}

	json.Unmarshal(file, c)
	return nil
}

type Worker struct {
	urlMain    string
	start      bool
	currentVus int
	cond       sync.Cond
	request    *RequestMeta
	timer      time.Timer

	handle chan AggregatorMeta
	Agg    Aggregator
}

func (w *Worker) Job(mainChan *chan Job) {

	w.cond.L.Lock()
	defer w.cond.L.Unlock()

	// wait for our turn
	for !w.start {
		w.cond.Wait()
	}

	quit := make(chan bool)
	// listening for the aggregators
	go func() {
		for {
			select {
			case <-quit:
				return
			case meta := <-w.handle:
				w.Agg.Consume(meta)
			}
		}
	}()

	url := fmt.Sprintf("%s%s", w.urlMain, w.request.Path)
	fmt.Printf("Starting the requests to: %s\n", url)
	for {
		select {
		case <-w.timer.C:
			quit <- true
			j := Job{url: url, agg: w.Agg}
			*(mainChan) <- j
			fmt.Println("DONE")
			return
		default:
			if w.currentVus < w.request.Vus {
				for range w.request.SpawnRate {
					switch w.request.RType {
					case "GET":
						go GET(url, w.handle, quit)
					}
					w.currentVus++
				}
			}
			time.Sleep(time.Second)
		}
	}

}

type Dult struct {
	NWorkers int
	Worker   []*Worker
	Handle   chan Job
	Aggs     map[string]*Aggregator
}

func (d *Dult) FromConfig(conf Configuration) {

	d.NWorkers = len(conf.Requests)
	d.Handle = make(chan Job)
	d.Aggs = make(map[string]*Aggregator)

	for i := range d.NWorkers {
		r := conf.Requests[i]
		// TODO: Parse correctly duration
		duration, _ := time.ParseDuration(conf.Requests[i].Duration)

		w := &Worker{
			urlMain: conf.Url,
			cond:    *sync.NewCond(&sync.Mutex{}),
			request: &r,
			timer:   *time.NewTimer(duration),
			handle:  make(chan AggregatorMeta),
		}
		d.Worker = append(d.Worker, w)

		go w.Job(&d.Handle)
	}
}

func (d *Dult) Start() {

	wait := sync.WaitGroup{}

	wait.Add(d.NWorkers)
	go func() {
		for job := range d.Handle {
			if d.Aggs[job.url] == nil {
				d.Aggs[job.url] = &Aggregator{}
			}

			d.Aggs[job.url].Add(job.agg)
			wait.Done()
		}
	}()

	for _, w := range d.Worker {
		w.cond.L.Lock()
		w.start = true
		w.cond.L.Unlock()

		w.cond.Signal()
	}
	wait.Wait()
	Render(d.Aggs)
}
