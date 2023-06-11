package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"distributed_finalproject/watch"

	api "go.etcd.io/etcd/api/etcdserverpb"
	client "go.etcd.io/etcd/client/v3"
)

type Domain_IP struct {
	Domain string
	Ip     string
}

func main() {
	//parameter set and initial
	cache_size := 4
	endpoint_set := []string{"127.0.0.1:32380", "127.0.0.1:12380", "127.0.0.1:22380"}
	cache_table := make([]Domain_IP, cache_size)
	//cache_table = append(*cache_table, &Domain_IP{"test","123"})
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

	//listen for channel of watch channel
	for {
		select {
		case val := <-watches.Watch_chaninfo:
			if check_watch_table(*watches, val.C_name) {
				select {
				case vall := <-val.Cn:
					if watchResponse_process(vall) != "" {
						watches.Watch_chaninfo <- val
						update_cache_table_value(Domain_IP{Domain: val.C_name, Ip: watchResponse_process(vall)}, &cache_table)
					} else { //key expire!
						fmt.Println("expired")
						watch.Remove_table(val.C_name, watches)
						remove_cache_table(Domain_IP{Domain: val.C_name, Ip: watchResponse_process(vall)}, &cache_table)
					}
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

func httpserver(watches *watch.Watchs, cache_table *[]Domain_IP, cc *client.Client, ctx context.Context) {
	addr := "172.20.10.5:50010"
	fmt.Println("listen on:" + addr)
	http.HandleFunc("/", a_HandleFunc_getid(*watches, cache_table, cc, ctx)) //設定存取的路由
	http.ListenAndServe(addr, nil)                                           //設定監聽的埠
}

func get_ip(domain_name string, watches watch.Watchs, cache_table *[]Domain_IP, cc *client.Client, ctx context.Context) string {
	cache := *cache_table
	//search cache
	for i := 0; i < len(cache); i++ {
		if domain_name == cache[i].Domain {
			fmt.Println("found!")
			return cache[i].Ip
		}
	}
	//not in cache, get from etcd
	r, err := cc.Get(cc.Ctx(), domain_name)
	if err != nil {
		return "Domain not found"
	}
	res := getResponse_process(r)
	//update cache
	update_cache_table(Domain_IP{Domain: domain_name, Ip: res}, cache_table)
	//watch the new domainIP
	watch.Add_watch_key(domain_name, &watches, ctx)
	return res
}

// update specific value where is already existed.
func update_cache_table_value(di Domain_IP, cache_table *[]Domain_IP) {
	cache := *cache_table
	clen := len(cache)
	for i := 0; i < clen; i++ {
		if cache[i].Domain == di.Domain {
			cache[i] = di
		}
	}
	fmt.Println(cache)
	*cache_table = cache
}

// remove the oldest Domain_IP(if full) and add new
func update_cache_table(di Domain_IP, cache_table *[]Domain_IP) {
	cache := *cache_table
	clen := len(cache)
	if (cache[clen-1] != Domain_IP{}) {
		for i := 0; i < clen-1; i++ {
			cache[i] = cache[i+1]
		}
		//edit
		cache[clen-1] = di
	} else {
		for i := 0; i < clen; i++ {
			if (cache[i] == Domain_IP{}) {
				cache[i] = di
				break
			}
		}
	}
	*cache_table = cache
}

func a_HandleFunc_getid(watches watch.Watchs, cache_table *[]Domain_IP, cc *client.Client, ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req Domain_IP
		body, _ := ioutil.ReadAll(r.Body)
		json.Unmarshal(body, &req)
		fmt.Println(&req)
		res := get_ip(req.Domain, watches, cache_table, cc, ctx)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// Write the response body
		ress := Domain_IP{Domain: req.Domain, Ip: res}
		jsonres, _ := json.Marshal(ress)
		w.Write(jsonres)
	}
}

func remove_cache_table(di Domain_IP, cache_table *[]Domain_IP) {
	cache := *cache_table
	for i := 0; i < len(cache); i++ {
		if cache[i].Domain == di.Domain {
			for t := i; t < len(cache); t++ {
				if t < len(cache)-1 {
					cache[t] = cache[t+1]
				} else {
					cache[t] = Domain_IP{}
				}
			}
			break
		}

	}
	*cache_table = cache
	fmt.Println("cache_table")
	fmt.Println(cache_table)
}
