package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	pb "github.com/amitsaha/apigatewaydemo/grpc-app-1/verify"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/zbindenren/negroni-prometheus"
	"google.golang.org/grpc"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
)

func verificationHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := grpc.Dial("rpc-app-1:6000", grpc.WithInsecure())
	// TODO If we cannot connect, return a proper HTTP response here
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	c := pb.NewUserVerifyClient(conn)

	var resp *pb.VerifyReply

	//TODO remove the harcoding
	// Marshal incoming JSON into a pb.VerifyRequest
	resp, err = c.VerifyUser(context.Background(), &pb.VerifyRequest{Id: 12321, Token: "$kasdasa"})
	// TODO raise a proper HTTP error response here
	if err != nil {
		log.Fatal("Could not verify: %v", err)
	}

	// Serialize resp.Message into JSON
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(resp)
}

func projectsHandler(w http.ResponseWriter, r *http.Request) {
	u, e := url.Parse("http://webapp-1/create")
	if e != nil {
		log.Fatal("Error parsing service URL")
	}
	// TODO raise a proper HTTP error response here
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	//log.Println(string(body))
	req, err := http.NewRequest(r.Method, u.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	// TODO raise a proper HTTP error response here
	if err != nil {
		log.Fatal(err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	// TODO raise a proper HTTP error response here
	if err != nil {
		log.Fatal(err)
	}
	// We get back JSON from the upstream service
	// so we can just send the body back as it is
	_, err = io.Copy(w, resp.Body)
}

func main() {

	var (
		httpAddr            = flag.String("http.addr", ":8000", "Address for API Gateway")
		healthcheckHttpAddr = flag.String("healthcheck.addr", ":9000", "Address for Healthcheck")
	)
	flag.Parse()

	// API gateway setup
	n := negroni.New()
	m := negroniprometheus.NewMiddleware("apigateway")
	n.Use(m)
	r := mux.NewRouter()
	r.Handle("/metrics", prometheus.Handler())
	n.UseHandler(r)

	r.HandleFunc("/projects/", projectsHandler)
	r.HandleFunc("/verify/", verificationHandler)

	// Interrupt handler.
	errc := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	// Start our API gateway
	go func() {
		errc <- http.ListenAndServe(*httpAddr, n)
	}()

	// Start another service for debugging, possibly healthchecks
	// Note that we also have /metrics exporting prometheus metrics
	go func() {
		errc <- http.ListenAndServe(*healthcheckHttpAddr, nil)
	}()
	// Run!
	<-errc
}
