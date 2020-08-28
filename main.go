package main

import (
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"log"
	"regexp"
)

func main() {
	// CLI flags
	elbPtr := flag.String("elb", "", "the elb to take over - myelb-1234.elb.amazonaws.com")
	// Parse flags
	flag.Parse()
	if *elbPtr == "" {
		log.Fatal("You have to Specify the ELB")
	}
	// We're using `MustCompile` so we fail hard if something's wrong with the regex
	/*
	   Define regex to match the "random" part of the ELB
	   This will match any random string between 6 and 10 digits enclosed between `-` and `.`
	*/
	reNum := regexp.MustCompile(`-(\d{6,10})\.`)
	// Get the "random" number from the elb
	NumToFind := reNum.FindStringSubmatch(*elbPtr)

	/* All AWS Regions we're matching
	us-east-2
	us-east-1
	us-west-1
	us-west-2
	ap-east-1
	ap-south-1
	ap-northeast-2
	ap-southeast-1
	ap-southeast-2
	ap-northeast-1
	ca-central-1
	cn-north-1
	cn-northwest-1
	eu-central-1
	eu-west-1
	eu-west-2
	eu-west-3
	eu-north-1
	sa-east-1
	us-gov-east-1
	*/
	// Define the regex to match the region
	reReg := regexp.MustCompile(`(us(-gov)?|ap|ca|cn|eu|sa)-(central|(north|south)?(east|west)?)-\d`)
	// Get the region of the ELB
	elbReg := reReg.FindString(*elbPtr)

}
