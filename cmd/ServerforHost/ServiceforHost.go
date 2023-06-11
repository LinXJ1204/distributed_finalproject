package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	client "go.etcd.io/etcd/client/v3"
)

type Domain_IP struct {
	Domain string
	Ip     string
}

func main() {
	endpoint_set := []string{"127.0.0.1:32380", "127.0.0.1:12380", "127.0.0.1:22380"}

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

	http.HandleFunc("/", a_HandleFunc_s(c, ctx))
	http.ListenAndServe("172.20.10.5:50000", nil) //設定監聽的埠

}

func a_HandleFunc_s(cc *client.Client, ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r)
		var req Domain_IP
		var res string
		body, _ := ioutil.ReadAll(r.Body)
		json.Unmarshal(body, &req)
		fmt.Println(&req)
		err := put_domain_ip(req, cc, ctx)
		if err != nil {
			res = "Error"
		} else {
			res = "ok"
		}
		w.Header().Set("Content-Type", "text/plain")
		// Write the response body
		fmt.Fprint(w, res)
	}
}

func put_domain_ip(di Domain_IP, cc *client.Client, ctx context.Context) error {
	les, _ := cc.Grant(ctx, 15)
	_, err := cc.Put(ctx, di.Domain, di.Ip, client.WithLease(les.ID))
	return err
}
