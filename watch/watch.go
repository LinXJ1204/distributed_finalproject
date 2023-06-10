package watch

import (
	"context"
	"fmt"

	client "go.etcd.io/etcd/client/v3"
)

type watchchan_info struct {
	C_name string
	Cn     client.WatchChan
}

type Watchs struct {
	Watcher        client.Watcher
	Cache_size     int
	Cache_t        []string
	Watch_chaninfo chan watchchan_info
}

func New_watchs(watcher client.Watcher, cache_size int) *Watchs {
	nwatch := &Watchs{
		Watcher:        watcher,
		Cache_size:     cache_size,
		Watch_chaninfo: make(chan watchchan_info, cache_size),
	}
	nwatch.Cache_t = make([]string, nwatch.Cache_size)
	return nwatch
}

func add2table(key string, watch *Watchs) {
	if (watch.Cache_t[watch.Cache_size-1]) != "" {
		for i := 0; i < watch.Cache_size-1; i++ {
			watch.Cache_t[i] = watch.Cache_t[i+1]
		}
		watch.Cache_t[watch.Cache_size-1] = key
	} else {
		for i := 0; i < watch.Cache_size; i++ {
			if watch.Cache_t[i] == "" {
				watch.Cache_t[i] = key
				break
			}
		}
	}

	fmt.Println(watch.Cache_t)
}

func Add_watch_key(key string, watch *Watchs, ctx context.Context) {
	add2table(key, watch)
	watch_key_c(key, ctx, watch.Watcher, watch.Watch_chaninfo)
}

func watch_key_c(key string, ctx context.Context, watcher client.Watcher, c chan watchchan_info) {
	c <- watchchan_info{
		C_name: key,
		Cn:     watcher.Watch(ctx, key),
	}
	fmt.Println("key add")
}

func Remove_table(key string, watch *Watchs) {
	for i := 0; i < len(watch.Cache_t); i++ {
		if watch.Cache_t[i] == key {
			for t := i; t < len(watch.Cache_t); t++ {
				if t < len(watch.Cache_t)-1 {
					watch.Cache_t[t] = watch.Cache_t[t+1]
				} else {
					watch.Cache_t[t] = ""
				}

			}
			break
		}

	}
	fmt.Println("cache_watch")
	fmt.Println(&watch.Cache_t)
}
