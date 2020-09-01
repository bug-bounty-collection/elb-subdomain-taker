# elb-subdomain-taker
Utility to overtake unclaimed subomdains pointing to AWS ELBs

![Go](https://github.com/RiRa12621/elb-subdomain-taker/workflows/Go/badge.svg)
![Docker Image CI](https://github.com/RiRa12621/elb-subdomain-taker/workflows/Docker%20Image%20CI/badge.svg)
[![Docker Repository on Quay](https://quay.io/repository/rira12621/elb-subdomain-taker/status "Docker Repository on Quay")](https://quay.io/repository/rira12621/elb-subdomain-taker)

## How to run

### Locally with go installed
To get the elb-subdomain-taker run:
```
go get github.com/rira12621/elb-subdomain-taker
```

To execute it:
```
elb-subdomain-taker --elb mycoolelb-1234123.us-east-1.elb.amazonaws.com
```
That's it.

### Binary without go
Pre-built binaries will be available soon.

### Docker
The images are automatically build via [Quay](https://quay.io/repository/rira12621/elb-subdomain-taker?tab=builds).

You can run the taker via

```
docker run --rm quay.io/rira12621/elb-subdomain-taker:latest --elb mycoolelb-1234123.us-east-1.elb.amazonaws.com
```

## How does this work?
You pass in an elb that you believe to be a vulnerable target for subdomain takeover.

This means you have an A record or a CNAME pointing to it but the ELB itself doesn't have any records.

Against regular assumptions, the number in the ELB domain is not for security measures so we can just enumarete it.
What happens here is the following:
* Create ELB
* Check if it matches the given ELB
* If not, destroy it and start again
* If yes, success we took over that subdomain

## What do I need?

* Valid AWS credentials set up
* A vulnerable subdomain

## Metrics

elb-subdomain-taker exposes prometheus style metrics on port 2112 and `/metrics`

See the table below for a list of implemented custom metrics in addition to default go metrics.


| Metric                                 | Explanation                      |
|----------------------------------------|----------------------------------|
| elb_subdomain_taker_elbs_created_total | The total number of ELBs created |
| elb_subdomain_taker_elbs_deleted_total | The total number of ELBs deleted |
| elb_subdomain_taker_elb_random_number  | Random number in the new ELB     |

You can use those metrics to monitor the behaviour of the elb-subdomain-taker.
Under regular circumstances the number of created and deleted records should be the same.
If they're not, you're piling up ELBs.

You can also use those metrics to monitor the rate at which ELBs are created.

## Warning

**THIS CAN BE VERY PRICEY**

Worst case, we're creating a lot of ELBs here, so better be sure you have your budgets set up.
Also be extra sure the target is vulnerable. Otherwise you may spend a lot of money for nothing.
Please refer to [AWS prices](https://aws.amazon.com/elasticloadbalancing/pricing/?nc1=h_ls) for more information.
