package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"distributed_finalproject/watch"

	api "go.etcd.io/etcd/api/etcdserverpb"
	client "go.etcd.io/etcd/client/v3"
)

type domain_IP struct {
	domain string
	ip     string
}

func main() {
	//parameter set and initial
	cache_size := 20
	endpoint_set := []string{"127.0.0.1:32380", "127.0.0.1:12380", "127.0.0.1:22380"}
	cache_table := make([]domain_IP, cache_size)
	//cache_table = append(*cache_table, &domain_IP{"test","123"})
	//etcd connection setting
	ctx := context.Background()

	configspec := &client.ConfigSpec{
		Endpoints:        endpoint_set,
		DialTimeout:      2 * time.Second,
		KeepAliveTime:    3 * time.Second,
		KeepAliveTimeout: 5 * time.Second,
	}
	config, _ := client.NewClientConfig(configspec, nil)

	c, _ := client.New(*config)
	defer c.Close()

	//watcher build and start
	watcher := c.Watcher
	defer watcher.Close()

	watches := watch.New_watchs(watcher, cache_size)

	//httpserver build and run
	go httpserver(watches, &cache_table, c, ctx)
	watch.Add_watch_key("test", watches, ctx)
	watch.Add_watch_key("foo", watches, ctx)

	//listen for channel of watch channel
	for {
		select {
		case val := <-watches.Watch_chaninfo:
			if check_watch_table(*watches, val.C_name) {
				select {
				case vall := <-val.Cn:
					watches.Watch_chaninfo <- val
					update_cache_table_value(domain_IP{domain: val.C_name, ip: watchResponse_process(vall)}, &cache_table)
					fmt.Println(watchResponse_process(vall))
				default:
					watches.Watch_chaninfo <- val
				}
			}
		}
	}

}

func getResponse_process(r *client.GetResponse) string { //wait to edit
	rr := api.RangeResponse(*r)
	res := rr.Kvs[0].Value
	return string(res)
}

func watchResponse_process(r client.WatchResponse) string { //wait to edit
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

func httpserver(watches *watch.Watchs, cache_table *[]domain_IP, cc *client.Client, ctx context.Context) {
	http.HandleFunc("/", a_HandleFunc_getid("foo", *watches, cache_table, cc, ctx)) //設定存取的路由
	http.ListenAndServe(":9090", nil)                                               //設定監聽的埠
}

func get_ip(domain_name string, watches watch.Watchs, cache_table *[]domain_IP, cc *client.Client, ctx context.Context) string {
	cache := *cache_table
	//search cache
	for i := 0; i < len(cache); i++ {
		if domain_name == cache[i].domain {
			fmt.Println("found!")
			return cache[i].ip
		}
	}
	//not in cache, get from etcd
	r, err := cc.Get(cc.Ctx(), domain_name)
	if err != nil {
		return "domain not found"
	}
	res := getResponse_process(r)
	update_cache_table(domain_IP{domain: domain_name, ip: res}, cache_table)
	watch.Add_watch_key(domain_name, &watches, ctx)
	return res
}

func update_cache_table_value(di domain_IP, cache_table *[]domain_IP) {
	cache := *cache_table
	clen := len(cache)
	for i := 0; i < clen; i++ {
		if cache[i].domain == di.domain {
			cache[i] = di
		}
	}
	*cache_table = cache
}

func update_cache_table(di domain_IP, cache_table *[]domain_IP) {
	cache := *cache_table
	clen := len(cache)
	if (cache[clen-1] != domain_IP{}) {
		for i := 0; i < clen-1; i++ {
			cache[i] = cache[i+1]
		}
	}
	cache = append(cache, di)
	*cache_table = cache
}

func a_HandleFunc_getid(domain_name string, watches watch.Watchs, cache_table *[]domain_IP, cc *client.Client, ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res := get_ip(domain_name, watches, cache_table, cc, ctx)
		fmt.Println(res)
	}
}
