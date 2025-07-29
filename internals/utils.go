package internals

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"
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

func (a *Aggregator) Add(agg Aggregator) {
	a.reqsTotal += agg.reqsTotal

	a.sucessReqs += agg.sucessReqs

	a.failureReqs += agg.failureReqs

	if agg.longestReq > a.longestReq {
		a.longestReq = agg.longestReq
	}

	if agg.shortestReq < a.shortestReq || a.shortestReq == 0 {
		a.shortestReq = agg.shortestReq
	}

	a.sucessRate = float32((a.sucessReqs * 100) / a.reqsTotal)

	a.failureRate = float32((a.failureReqs * 100) / a.reqsTotal)

	a.avgReq = (a.longestReq + a.shortestReq) / 2
}

func (a *Aggregator) Rows() []string {
	return []string{fmt.Sprintf("%d", a.reqsTotal), fmt.Sprintf("%d", a.sucessReqs), fmt.Sprintf("%d", a.failureReqs),
		fmt.Sprintf("%f", a.sucessRate), fmt.Sprintf("%f", a.failureRate), a.shortestReq.String(),
		a.longestReq.String(), a.avgReq.String()}
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

func Render(aggs map[string]*Aggregator) {

	data := [][]string{
		{"URL", "Total Reqs", "Sucess Reqs", "Failure Reqs", "Sucess %", "Failure %", "Fastest Req", "Slowest Req", "Avg Req"},
	}

	for key, value := range aggs {
		row := []string{key}
		row = append(row, value.Rows()...)
		data = append(data, row)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.Header(data[0])
	table.Bulk(data[1:])
	table.Render()
}
