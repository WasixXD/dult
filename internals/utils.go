package internals

import (
	"fmt"
	"net/http"
	"time"
)

type AggregatorMeta struct {
	sucess  int
	failure int

	reqTime time.Duration
}

type Aggregator struct {
	reqsTotal int

	sucessReqs int
	sucessRate float32

	failureReqs int
	failureRate float32

	longestReq  time.Duration
	shortestReq time.Duration
	avgReq      time.Duration
}

func (a *Aggregator) Consume(meta AggregatorMeta) {

	a.reqsTotal++

	a.sucessReqs += meta.sucess
	a.failureReqs += meta.failure

	if meta.reqTime > a.longestReq {
		a.longestReq = meta.reqTime
	} else if meta.reqTime < a.shortestReq || a.shortestReq == 0 {
		a.shortestReq = meta.reqTime
	}

}

type Job struct {
	url string
	agg Aggregator
}

func GET(url string, handle chan AggregatorMeta, quit chan bool) {
	for {
		select {
		case <-quit:
			return
		default:
			now := time.Now()
			res, err := http.Get(url)

			if err != nil {
				fmt.Println("Something went wrong")
			}

			end := time.Since(now)

			meta := AggregatorMeta{reqTime: end}

			if res.StatusCode == http.StatusOK {
				meta.sucess = 1
			} else {
				meta.failure = 1
			}

			res.Body.Close()
			handle <- meta

		}
	}
}
