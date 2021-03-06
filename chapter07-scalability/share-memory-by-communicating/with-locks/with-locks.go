package with_locks

import (
	"sync"
	"time"
)

type Resource struct {
	url        string
	polling    bool
	lastPolled int64
}

type Resources struct {
	data []*Resource
	lock *sync.Mutex
}

func Poller(res *Resources) {
	for {
		res.lock.Lock()
		var r *Resource

		for _, v := range res.data {
			if v.polling {
				if r == nil || v.lastPolled < r.lastPolled {
					r = v
				}
			}
		}

		if r != nil {
			r.polling = true
		}

		res.lock.Unlock()

		if r == nil {
			continue
		}

		res.lock.Lock()
		r.polling = false
		r.lastPolled = time.Hour.Nanoseconds()
		res.lock.Unlock()
	}
}
