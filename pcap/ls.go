package pcap

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/cobra"
	"log"
)

var (
	Namespace      string
	TracesessionId string
)

var cmdLs = &cobra.Command{
	Use:   "ls",
	Short: "list the pcap files",
	Run: func(cmd *cobra.Command, args []string) {
		// list pcap file
		list()
	},
}

func NewCmdLs() *cobra.Command {

	cmd := cmdLs
	f := cmd.Flags()

	f.StringVar(&Namespace, "namespace", "", " Specify which namespace to list pcap files, if not set, it will retrieve all namespaces")
	f.StringVar(&TracesessionId, "TracesessionId", "", "Specify which tracesession to list pcap ifles, if not set, it will retrieve all tracession")
	return cmdLs
}

func list() {

	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               "http://172.0.3.66:9000",
			HostnameImmutable: true,
		}, nil
	})

	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithEndpointResolverWithOptions(customResolver))
	if err != nil {
		log.Fatal(err)
	}

	// Create an Amazon S3 service client
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		// o.Region = "us-west-2"
		// o.UseAccelerate = true
	})

	buckets := []string{"test1", "test2", "test3"} // define a list of bucket names

	for _, bucket := range buckets {
		// Get the first page of results for ListObjectsV2 for a bucket
		output, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
			Bucket: aws.String(bucket),
		})
		if err != nil {
			log.Fatalf("Failed to list objects in bucket %s: %v", bucket, err)
		}

		log.Printf("Objects in bucket %s:", bucket)
		for _, object := range output.Contents {
			log.Printf("\tkey=%s size=%d", aws.ToString(object.Key), object.Size)
		}
	}
}
