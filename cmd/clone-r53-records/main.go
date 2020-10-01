package main

import (
	"flag"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
)

var weight = int64(1)
var sourceZoneID string
var destZoneID string
var destAccountID string
var mode string

func init() {
	flag.StringVar(&sourceZoneID, "s", "", "AWS Zone Id for domain the existing zone")
	flag.StringVar(&destZoneID, "d", "", "AWS Zone Id for domain the destination zone")
	flag.StringVar(&destAccountID, "a", "", "AWS Account Id for the destination zone")
	flag.StringVar(&mode, "m", "UPSERT", "AWS Account Id for the destination zone")

}

func main() {
	flag.Parse()

	fmt.Println(mode)

	if sourceZoneID == "" || destZoneID == "" || destAccountID == "" {
		fmt.Println(fmt.Errorf("incomplete arguments: s: %s, d: %s: a: %s", sourceZoneID, destZoneID, destAccountID))
		flag.PrintDefaults()
		return
	}

	sess, err := session.NewSession(&aws.Config{Region: aws.String("ap-southeast-2")})

	if err != nil {
		panic(err.Error)
	}

	r53 := route53.New(sess)

	records := listRecords(r53, sourceZoneID)

	arn := fmt.Sprintf(
		"arn:aws:iam::%v:role/OrganizationAccountAccessRole",
		destAccountID,
	)
	assumedSession, err := session.NewSession(&aws.Config{
		Credentials: stscreds.NewCredentials(sess, arn),
	})

	if err != nil {
		println(err.Error())
		panic(err)
	}

	r53Dest := route53.New(assumedSession)

	createRecords(r53Dest, records)
}

func listRecords(r53 *route53.Route53, zoneID string) *route53.ListResourceRecordSetsOutput {
	listParams := &route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String(zoneID),
	}
	records, err := r53.ListResourceRecordSets(listParams)

	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Retrived records for %s", zoneID)

	return records
}

func createRecords(r53 *route53.Route53, records *route53.ListResourceRecordSetsOutput) {
	params := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: formatRecordsToChanges(records),
			Comment: aws.String("Inserted new records"),
		},
		HostedZoneId: &destZoneID,
	}

	response, err := r53.ChangeResourceRecordSets(params)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(response)
}

func printRecords(records *route53.ListResourceRecordSetsOutput) {
	for i := 0; i < len(records.ResourceRecordSets); i++ {
		println(*records.ResourceRecordSets[i].Name)
		println(*records.ResourceRecordSets[i].Type)

		for k := 0; k < len(records.ResourceRecordSets[i].ResourceRecords); k++ {

			println(*records.ResourceRecordSets[i].ResourceRecords[k].Value)
		}

		println("\n")
	}
}

func formatRecordsToChanges(records *route53.ListResourceRecordSetsOutput) []*route53.Change {
	changes := make([]*route53.Change, len(records.ResourceRecordSets))

	for i := 0; i < len(records.ResourceRecordSets); i++ {
		if *records.ResourceRecordSets[i].Type != "NS" || *records.ResourceRecordSets[i].Name != "ngatirangi.com." {
			changes[i] = &route53.Change{
				Action:            aws.String("UPSERT"),
				ResourceRecordSet: records.ResourceRecordSets[i],
			}
		}
	}

	filteredChanges := make([]*route53.Change, 0)

	for _, item := range changes {
		if item != nil {
			filteredChanges = append(filteredChanges, item)
		}
	}

	fmt.Println(filteredChanges)

	return filteredChanges
}
