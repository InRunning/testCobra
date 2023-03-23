package pcap

import (
	"context"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"regexp"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

var (
	namespaces      string
	tracesessionIds string
)

type BucketFile struct {
	BucketName string
	FileName   string
}

var cmdLs = &cobra.Command{
	Use:   "ls",
	Short: "list the pcap files",
	Run: func(cmd *cobra.Command, args []string) {
		// list pcap file
		ctx := context.Background()
		err := listFileNames(ctx)
		if err != nil {
			log.Fatal(err)
			return
		}
	},
}

func NewCmdLs() *cobra.Command {

	cmd := cmdLs
	f := cmd.Flags()

	f.StringVar(&namespaces, "namespace", "", " Specify which namespace to list pcap files, if not set, it will retrieve all namespaces")
	f.StringVar(&tracesessionIds, "tracesessionId", "", "Specify which tracesession to list pcap ifles, if not set, it will retrieve all tracession")
	flag.Parse()
	return cmdLs
}

func GetFileList(ctx context.Context) ([]BucketFile, error) {
	log.Info("execute to get file list")
	var namespaceList []string
	var tracesessionIdList []string
	fmt.Println("namespaces: ", namespaces)
	fmt.Println("tracesessionIds: ", tracesessionIds)

	if namespaces != "" {
		// Split the string using comma separator
		parts := strings.Split(namespaces, ",")

		// Trim spaces from each part
		for _, part := range parts {
			namespaceList = append(namespaceList, strings.TrimSpace(part))
		}
	}
	if tracesessionIds != "" {
		// Split the string using comma separator
		parts := strings.Split(tracesessionIds, ",")

		// Trim spaces from each part
		for _, part := range parts {
			tracesessionIdList = append(tracesessionIdList, strings.TrimSpace(part))
		}
	}
	bucketFiles, err := GetAllFileNames(ctx, client)
}

func GetBucketFiles(ctx context.Context) ([]BucketFile, error) {

	fmt.Println("执行GetBucketFiles")
	namespaces = "default"
	client, err := GetS3Client()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	var namespaceList []string
	var tracesessionIdList []string
	fmt.Println("namespaces: ", namespaces)
	fmt.Println("tracesessionIds: ", tracesessionIds)

	if namespaces != "" {
		// Split the string using comma separator
		parts := strings.Split(namespaces, ",")

		// Trim spaces from each part
		for _, part := range parts {
			namespaceList = append(namespaceList, strings.TrimSpace(part))
		}
	}
	if tracesessionIds != "" {
		// Split the string using comma separator
		parts := strings.Split(tracesessionIds, ",")

		// Trim spaces from each part
		for _, part := range parts {
			tracesessionIdList = append(tracesessionIdList, strings.TrimSpace(part))
		}
	}

	bucketFiles, err := GetAllFileNames(ctx, client)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	fmt.Println("bucketFiles number: ", len(bucketFiles))
	if bucketFiles != nil && len(bucketFiles) != 0 {
		fmt.Println("first bucket file: ", bucketFiles[0])
		fmt.Println("last bucket file: ", bucketFiles[len(bucketFiles)-1])
	}

	var regexs []*regexp.Regexp
	if tracesessionIds != "" {
		for _, tracesessionId := range tracesessionIdList {
			if namespaces == "" {
				regexs = append(regexs, regexp.MustCompile(tracesessionId))
			} else {
				for _, namespace := range namespaceList {
					regexs = append(regexs, regexp.MustCompile(tracesessionId+namespace))
				}
			}
		}
	} else {
		if namespaces != "" {
			for _, namespace := range namespaceList {
				regexs = append(regexs, regexp.MustCompile(namespace))
			}
		}
	}

	var filterBucketFiles []BucketFile
	for _, bucketFile := range bucketFiles {
		fileName := bucketFile.FileName
		for _, regex := range regexs {
			match := regex.FindStringSubmatch(fileName)
			if match != nil {
				filterBucketFiles = append(filterBucketFiles, bucketFile)
				break
			}
		}
	}
	fmt.Println("filterBucketFiles: ", filterBucketFiles)

	return filterBucketFiles, nil
}

func listFileNames(ctx context.Context) error {
	filterBucketFiles, err := GetBucketFiles(ctx)
	if err != nil {
		return err
	}
	if filterBucketFiles == nil || len(filterBucketFiles) == 0 {
		return nil
	}
	fmt.Printf("name\t\t" + "time\t\t" + "tracesession\t\t" + "ue/interface\t\t\n")
	for _, bucketFile := range filterBucketFiles {
		fileName := bucketFile.FileName
		parts := strings.Split(fileName, "-")
		timestamp := parts[len(parts)-1]
		tracesessionId := parts[0]
		fmt.Printf(fileName + "\t\t" + timestamp + "\t\t" + tracesessionId + "\t\t" + "\t\t\n")
	}
	return nil
}

func GetAllFileNames(ctx context.Context, k8sClient client.Client) ([]BucketFile, error) {
	k8sClient, err := GetK8sClient()

	if err != nil {
		log.Error("unable to get k8s client, error: ", err)
		return nil, err
	}
	crd := &unstructured.Unstructured{}
	// create API resource reference
	crd.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "smf.axyom.casa-systems.io",
		Version: "v1alpha1",
		Kind:    "SMF",
	})

	var bucketFiles []BucketFile
	for _, bucketName := range bucketNames {
		// Get the first page of results for ListObjectsV2 for a bucket
		output, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket: aws.String(bucketName),
		})
		if err != nil {
			log.Fatalf("Failed to list objects in bucket %s: %v", bucketName, err)
			return nil, err
		}

		for _, object := range output.Contents {
			if object.Key != nil {
				bucketFiles = append(bucketFiles, BucketFile{bucketName, *object.Key})
			}
		}
	}
	return bucketFiles, nil
}
