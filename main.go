package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/imeyer/go-utilities/hostnameutils"
	//"net"
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
	"sync"
)

var interval = flag.String("interval", "15s", "Interval to process data before sending it to InfluxDB")

type HttpStatusCodes struct {
	sync.RWMutex
	counters map[int64]int64
}

func (h *HttpStatusCodes) Increment(key int64) {
	h.Lock()
	
	if _, empty := h.counters[key]; !empty {
		h.counters[key] = 1
	} else {
		h.counters[key] += 1
	}

	h.Unlock()
}

func main() {
	flag.Parse()

	hostname, err := os.Hostname()
	if err != nil {
		glog.Fatal("Can not get hostname: ", err)
	}

	interval, err := time.ParseDuration(*interval)
	if err != nil {
		glog.Fatal("Error parsing interval duration: ", err)
	}

	ticker := time.NewTicker(interval)
	reader := bufio.NewReader(os.Stdin)
	quit := make(chan struct{})

	glog.Info(hostnameutils.Reverse(hostname))
	st := &HttpStatusCodes{
		counters: map[int64]int64{},
	}

	go func() {
		for {
			c, err := reader.ReadString('\n')
			if err != nil {
				continue
			}
			c = strings.TrimSpace(c)
			num, err := strconv.ParseInt(c, 10, 32)
			st.Increment(num)
		}
	}()

	for {
		select {
		case <-ticker.C:
			st.RLock()
			glog.Info(st.counters)
			st.RUnlock()
			st.Lock()
			st.counters = map[int64]int64{}
			st.Unlock()
		case <-quit:
			ticker.Stop()
			return
		}
	}
}
