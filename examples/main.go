package main

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	_ "expvar"

	"github.com/nicolai86/collector"
)

func main() {
	var a = collector.NewAgent(time.Millisecond*150, time.Second*5)
	a.HighResolutionInterval = time.Second * 2
	a.DownSamplingInterval = time.Millisecond * 500

	a.Add("TotalAlloc", &collector.IntCollector{
		Run: func(ic *collector.IntCollector) int64 {
			var ms runtime.MemStats
			runtime.ReadMemStats(&ms)
			runtime.GC()
			fmt.Printf("%d: total %d\n", a.Metrics["TotalAlloc"].Len(), ms.Alloc)

			return int64(ms.Alloc)
		}}, collector.Int64MaxSampler{})
	go a.Run()
	go func() {
		for {
			time.Sleep(time.Second)
			var e = a.Metrics["TotalAlloc"].Front()
			for e != nil {
				fmt.Printf("%v\n", e.Value)
				e = e.Next()
			}
		}
	}()

	http.ListenAndServe(":8080", http.DefaultServeMux)
}
