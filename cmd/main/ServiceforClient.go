package main

import (
	"context"
	"fmt"
	"time"

	"distributed_finalproject/watch"

	api "go.etcd.io/etcd/api/etcdserverpb"
	client "go.etcd.io/etcd/client/v3"
)

func main() {
	//connection setting

	ctx := context.Background()

	configspec := &client.ConfigSpec{
		Endpoints:        []string{"127.0.0.1:32380", "127.0.0.1:12380", "127.0.0.1:22380"},
		DialTimeout:      2 * time.Second,
		KeepAliveTime:    3 * time.Second,
		KeepAliveTimeout: 5 * time.Second,
	}

	config, _ := client.NewClientConfig(configspec, nil)

	c, _ := client.New(*config)

	defer c.Close()

	//translate to RangeResponse which have method to access value (write a independent func later)
	r, _ := c.Get(c.Ctx(), "test")
	fmt.Println(getResponse_process(r))

	watcher := c.Watcher
	defer watcher.Close()

	watches := watch.New_watchs(watcher, 5)

	watch.Add_watch_key("test", watches, ctx)
	watch.Add_watch_key("foo", watches, ctx)
	//test for channel of channel
	for {
		select {
		case val := <-watches.Watch_chaninfo:
			if check_watch_table(*watches, val.C_name) {
				select {
				case vall := <-val.Cn:
					watches.Watch_chaninfo <- val
					fmt.Println(watchResponse_process(vall))
				default:
					watches.Watch_chaninfo <- val
				}
			}
		}
	}

}

func getResponse_process(r *client.GetResponse) string {
	rr := api.RangeResponse(*r)
	res := rr.Kvs[0].Value
	return string(res)
}

func watchResponse_process(r client.WatchResponse) string {
	fmt.Println(r)
	return string(r.Events[0].Kv.Value)
}

func watch_key_c(key string, ctx context.Context, watcher client.Watcher, c chan client.WatchChan) {
	c <- watcher.Watch(ctx, key)
}

func check_watch_table(watches watch.Watchs, key string) bool {
	for i := 0; i < len(watches.Cache_t); i++ {
		if watches.Cache_t[i] == key {
			return true
		}
	}
	return false
}
