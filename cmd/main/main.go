package main

import (
	"context"
	"fmt"
	"time"

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

	//translate to RangeResponse which have method to access value (write a independent func later)
	r, _ := c.Get(c.Ctx(), "test")
	fmt.Println(getResponse_process(r))

	watcher := client.NewWatcher(c) //return a channel
	wac := <-watcher.Watch(ctx, "test")
	fmt.Println(watchResponse_process(wac))

}

func getResponse_process(r *client.GetResponse) string {
	rr := api.RangeResponse(*r)
	res := rr.Kvs[0].Value
	return string(res)
}

func watchResponse_process(r client.WatchResponse) string {
	return string(r.Events[0].Kv.Value)
}
