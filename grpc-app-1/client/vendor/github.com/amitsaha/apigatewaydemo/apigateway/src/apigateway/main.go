// This is the API gateway which demonstrates the following:
// 1. Handle incoming requests which it talks to two other services: one over REST,
// the other via gRPC. These services are load balanced via using the "Service registry"
// pattern
// 2. Rate limiting handled by the gateway
// 3. Auth handling handled by the gateway
// 4. Export /metrics endpoint for Prometheus of API Gateway Stats
// 5. Export /debug/vars (expvar) listening over :9000
// 6.

// Used with https://github.com/go-kit/kit/blob/master/examples/ as the starting
// point
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	_ "expvar"
	"flag"
	"fmt"
	pb "github.com/amitsaha/apigatewaydemo/grpc-app-1/verify"
	"github.com/codegangsta/negroni"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/ratelimit"
	"github.com/go-kit/kit/sd"
	consulsd "github.com/go-kit/kit/sd/consul"
	"github.com/go-kit/kit/sd/lb"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/hashicorp/consul/api"
	jujuratelimit "github.com/juju/ratelimit"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/zbindenren/negroni-prometheus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/pprof"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
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
	// This is required for Python web applications to process
	// the request correctly
	req.ContentLength = int64(len(buf.Bytes()))

	req.Body = ioutil.NopCloser(&buf)
	return nil
}

func decodeCreateRequest(ctx context.Context, req *http.Request) (interface{}, error) {
	var request struct {
		Title string `json:"title"`
	}
	// Check if we have the Auth-Header-V1 set for Header based authentication
	// TODO: Could be a middleware like rate limiter
	if req.Header.Get("Auth-Header-V1") == "" {
		return nil, errors.New("Auth-Header-V1 missing")
	}
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func decodeVerifyRequest(ctx context.Context, req *http.Request) (interface{}, error) {
	// Check if we have the Auth-Header-V1 set for Header based authentication
	// TODO: Could be a middleware like rate limiter
	var request pb.VerifyRequest
	if req.Header.Get("Auth-Header-V1") == "" {
		return nil, errors.New("Auth-Header-V1 missing")
	}
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func decodeProjectsResponse(ctx context.Context, resp *http.Response) (interface{}, error) {
	var response struct {
		Id  int    `json:"id,omitempty"`
		Url string `json:"url,omitempty"`
		Err string `json:"err,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return response, nil
}

type verifyResponse struct {
	Message string
	Err     error
}

func DecodeGRPCVerifyResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(*pb.VerifyReply)
	return verifyResponse{Message: resp.Message, Err: nil}, nil
}

//type verifyRequest struct{ Name string }

func EncodeGRPCVerifyRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(pb.VerifyRequest)
	return &pb.VerifyRequest{Id: req.Id, Token: req.Token}, nil
}

func makeVerifyEndpoint(conn *grpc.ClientConn) endpoint.Endpoint {
	return grpctransport.NewClient(
		conn,
		"userverify.UserVerify", // Service name, packagename.ServiceName
		"VerifyUser",            // Function
		EncodeGRPCVerifyRequest,
		DecodeGRPCVerifyResponse,
		pb.VerifyReply{},
		//grpctransport.ClientBefore(opentracing.ToGRPCRequest(tracer, logger)),
	).Endpoint()
}

func verifyFactory(ctx context.Context) sd.Factory {

	var (
		qps = 1 //1 queries per second
	)
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		// We can set functions to be called before and after the request
		// as well
		conn, err := grpc.Dial(instance, grpc.WithInsecure())
		if err != nil {
			return nil, nil, err
		}
		// Our gRPC client
		endpoint := makeVerifyEndpoint(conn)
		// Add rate limiting for this endpoint
		endpoint = ratelimit.NewTokenBucketLimiter(jujuratelimit.NewBucketWithRate(float64(qps), int64(qps)))(endpoint)
		return endpoint, nil, nil
	}
}

func projectsFactory(ctx context.Context, method, path string) sd.Factory {

	var (
		qps = 1 //1 queries per second
	)
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
		// We can set functions to be called before and after the request
		// as well
		endpoint := httptransport.NewClient(
			method,
			u,
			encodeJSONRequest,
			decodeProjectsResponse,
		).Endpoint()
		// Add rate limiting for this endpoint
		endpoint = ratelimit.NewTokenBucketLimiter(jujuratelimit.NewBucketWithRate(float64(qps), int64(qps)))(endpoint)
		return endpoint, nil, nil
	}
}

func main() {

	var (
		httpAddr            = flag.String("http.addr", ":8000", "Address for API Gateway")
		healthcheckHttpAddr = flag.String("healthcheck.addr", ":9000", "Address for Healthcheck")
		consulAddr          = flag.String("consul.addr", "", "Consul agent address")
		// Retry upon a non-200 response (TODO: investigate)
		retryMax     = flag.Int("retry.max", 1, "per-request retries to different instances")
		retryTimeout = flag.Duration("retry.timeout", 5000*time.Millisecond, "per-request timeout, including retries")
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
	// API gateway setup
	n := negroni.New()
	// Prometheus middleware
	m := negroniprometheus.NewMiddleware("apigateway")
	n.Use(m)
	r := mux.NewRouter()
	r.Handle("/metrics", prometheus.Handler())
	n.UseHandler(r)

	// Handlers for /debug/* (net/http/pprof)
	r.HandleFunc("/debug/pprof/", pprof.Index)
	r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/debug/pprof/profile", pprof.Profile)
	r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)

	var (
		tags        = []string{}
		passingOnly = true
		create      endpoint.Endpoint
	)

	// Route to a HTTP service
	// Handle /projects/: This is another HTTP service which is where our
	// Projects API is running
	{
		factory := projectsFactory(ctx, "POST", "/create")
		subscriber := consulsd.NewSubscriber(client, factory, logger, "projects", tags, passingOnly)
		balancer := lb.NewRoundRobin(subscriber)
		retry := lb.Retry(*retryMax, *retryTimeout, balancer)
		create = retry
		r.Handle("/projects/", httptransport.NewServer(ctx, create, decodeCreateRequest, encodeJSONResponse))
	}
	// Route to a gRPC service
	// Handle /verify/
	{
		factory := verifyFactory(ctx)
		subscriber := consulsd.NewSubscriber(client, factory, logger, "verification", tags, passingOnly)
		balancer := lb.NewRoundRobin(subscriber)
		retry := lb.Retry(*retryMax, *retryTimeout, balancer)
		create = retry
		r.Handle("/verify/", httptransport.NewServer(ctx, create, decodeVerifyRequest, encodeJSONResponse))
	}

	// Interrupt handler.
	errc := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	// Start our API gateway
	go func() {
		logger.Log("transport", "HTTP", "addr", *httpAddr)
		errc <- http.ListenAndServe(*httpAddr, n)
	}()

	// Start another service for debugging, possibly healthchecks
	// Right now, we just expose expvar data
    // Use with expvarmon -ports=9000 (https://github.com/divan/expvarmon)
	// Note that we also have /metrics exporting prometheus metrics
	go func() {
		logger.Log("healthcheck", "HTTP", "addr", *healthcheckHttpAddr)
		errc <- http.ListenAndServe(*healthcheckHttpAddr, nil)
	}()

	// Run!
	logger.Log("exit", <-errc)

}
