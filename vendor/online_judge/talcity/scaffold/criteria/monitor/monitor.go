package monitor

import (
	"net/http"
	"time"

	"online_judge/talcity/scaffold/criteria/monitor/caller"
)

func HttpHandlerWrapper(api string, handler func(w http.ResponseWriter, r *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer func() {
			cost := time.Now().Sub(start)
			code := "0" // TODO: get code
			caller := caller.CallerNameFromContext(r.Context())
			if caller == "" {
				caller = "unknown"
			}

			if counter, _ := Monitor.Counter(caller, api, code); counter != nil { // TODO: caller
				counter.Inc()
			}

			if timer, _ := Monitor.Timer(caller, api, code); timer != nil {
				timer.Observe(float64(cost / time.Millisecond))
			}
		}()

		handler(w, r)
		return
	}
}
