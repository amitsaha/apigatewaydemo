// This is the API gateway which demonstrates the following:
// 1. Handle incoming requests which it talks to two other services: one over REST,
// the other via gRPC. These services are load balanced via using the "Service registry"
// pattern
// 2. Rate limiting handled by the gateway
// 3. Auth handling handled by the gateway
// 4. Tracking
// 5. Plugging in another service should be easy and 2, 3  and 4
// should be for free

// Used with https://github.com/go-kit/kit/blob/master/examples/apigateway/main.go as
// the reference
package main

import (
	"flag"
	"github.com/go-kit/kit/log"
	consulsd "github.com/go-kit/kit/sd/consul"
	"github.com/gorilla/mux"
	"github.com/hashicorp/consul/api"
	"golang.org/x/net/context"
    "net/http"
    "encoding/json"
    "github.com/go-kit/kit/sd"
    "github.com/go-kit/kit/endpoint"
    "github.com/go-kit/kit/sd/lb"
    "io"
    httptransport "github.com/go-kit/kit/transport/http"
    "bytes"
	"io/ioutil"
    "net/url"
	"fmt"
    "strings"
	"syscall"
    "os"
	"os/signal"
    "time"

)

func encodeJSONResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func encodeJSONRequest(_ context.Context, req *http.Request, request interface{}) error {
	var buf bytes.Buffer
    if err := json.NewEncoder(&buf).Encode(request); err != nil {
		return err
	}
    req.Header.Set("Content-Type", "application/json")
	req.Body = ioutil.NopCloser(&buf)

    fmt.Printf("%v", req)

	return nil
}

func decodeCreateRequest(ctx context.Context, req *http.Request) (interface{}, error) {
	var request struct {
		Title string `json:"title"`
	}
    fmt.Printf("Decoding create request")
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func decodeProjectsResponse(ctx context.Context, resp *http.Response) (interface{}, error) {
	var response struct {
        Id int `json:"id"`
		Title   string `json:"title"`
		Err string `json:"err,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return response, nil
}

func projectsFactory(ctx context.Context, method, path string ) sd.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		if !strings.HasPrefix(instance, "http") {
			instance = "http://" + instance
		}
		u, err := url.Parse(instance)
		if err != nil {
			panic(err)
		}
		if u.Path == "" && method == "POST" {
			u.Path = "/create"
		}
		endpoint := httptransport.NewClient(
			method,
			u,
			encodeJSONRequest,
			decodeProjectsResponse,
		).Endpoint()

		return endpoint, nil, nil
	}
}

func main() {

	var (
		httpAddr     = flag.String("http.addr", ":8000", "Address for HTTP server")
		consulAddr   = flag.String("consul.addr", "", "Consul agent address")
		retryMax     = flag.Int("retry.max", 3, "per-request retries to different instances")
		retryTimeout = flag.Duration("retry.timeout", 500*time.Millisecond, "per-request timeout, including retries")
	)
	flag.Parse()

	// Logging domain.
	var logger log.Logger
	logger = log.NewLogfmtLogger(os.Stderr)
	logger = log.NewContext(logger).With("ts", log.DefaultTimestampUTC)
	logger = log.NewContext(logger).With("caller", log.DefaultCaller)

	// Service discovery domain. In this example we use Consul.
	var client consulsd.Client
	{
		consulConfig := api.DefaultConfig()
		if len(*consulAddr) > 0 {
			consulConfig.Address = *consulAddr
		}
		consulClient, err := api.NewClient(consulConfig)
		if err != nil {
			logger.Log("err", err)
			os.Exit(1)
		}
		client = consulsd.NewClient(consulClient)
	}

	// Transport domain.
	// Learn more about contexts: https://blog.golang.org/context
	ctx := context.Background()
	r := mux.NewRouter()

	var (
		tags        = []string{}
		passingOnly = true
		create      endpoint.Endpoint
	)

	factory := projectsFactory(ctx, "POST", "/create")
	subscriber := consulsd.NewSubscriber(client, factory, logger, "projects", tags, passingOnly)
	balancer := lb.NewRoundRobin(subscriber)
	retry := lb.Retry(*retryMax, *retryTimeout, balancer)
	create = retry

	// Routes
	// Handle /api/: This is another HTTP service which is where our
	// Projects API is running
	r.Handle("/projects/", httptransport.NewServer(ctx, create, decodeCreateRequest, encodeJSONResponse))

	// Interrupt handler.
	errc := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	// HTTP transport.
	go func() {
		logger.Log("transport", "HTTP", "addr", *httpAddr)
		errc <- http.ListenAndServe(*httpAddr, r)
	}()

	// Run!
	logger.Log("exit", <-errc)

}
