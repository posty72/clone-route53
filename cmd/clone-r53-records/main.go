package main

import (
	"flag"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
)

// var name string
// var target string
// var TTL int64
var weight = int64(1)
var sourceZoneID string
var destZoneID string
var destAccountID string

func init() {
	// flag.StringVar(&name, "d", "", "domain name")
	// flag.StringVar(&target, "t", "", "target of domain name")
	flag.StringVar(&sourceZoneID, "f", "", "AWS Zone Id for domain")
	// flag.Int64Var(&TTL, "ttl", int64(60), "TTL for DNS Cache")

}

func main() {
	flag.Parse()

	if sourceZoneID == "" {
		fmt.Println(fmt.Errorf("incomplete arguments: f: %s", sourceZoneID))
		flag.PrintDefaults()
		return
	}

	sess, err := session.NewSession(&aws.Config{Region: aws.String("ap-southeast-2")})

	if err != nil {
		panic(err.Error)
	}

	r53 := route53.New(sess)

	records := listCnames(r53)

	for i := 0; i < len(records.ResourceRecordSets); i++ {
		println(*records.ResourceRecordSets[i].Name)
		println(*records.ResourceRecordSets[i].Type)

		for k := 0; k < len(records.ResourceRecordSets[i].ResourceRecords); k++ {

			println(*records.ResourceRecordSets[i].ResourceRecords[k].Value)
		}

		println("\n")
	}

	arn := fmt.Sprintf(
		"arn:aws:iam::%v:role/OrganizationAccountAccessRole",
		destAccountID,
	)
	assumedCreds := stscreds.NewCredentials(sess, arn)

	assumedSession := session.NewSession(&aws.Config{
		Credentials: *assumedCreds.Get()
	})

	r53Dest := route53.New(assumedCreds)

}

func listCnames(r53 *route53.Route53) *route53.ListResourceRecordSetsOutput {
	listParams := &route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String(sourceZoneID),
	}
	records, err := r53.ListResourceRecordSets(listParams)

	if err != nil {
		panic(err.Error())
	}

	return records
}
