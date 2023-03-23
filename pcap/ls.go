package pcap

import (
	"context"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"regexp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

var (
	namespace      string
	tracesessionId string
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

	var namespace string
	var tracesessionId string
	f.StringVar(&namespace, "namespace", "", " Specify which namespace to list pcap files, if not set, it will retrieve all namespace")
	f.StringVar(&tracesessionId, "tracesessionId", "", "Specify which tracesession to list pcap ifles, if not set, it will retrieve all tracession")
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
	config, err := ctrl.GetConfig()
	if err != nil {
		log.Error("can't get config, error: ", err)
		return nil, err
	}
	dynamicClient := dynamic.NewForConfigOrDie(config)

	// Specify the group, version and resource name
	smf := schema.GroupVersionResource{
		Group:    "smf.axyom.casa-systems.io",
		Version:  "v1alpha1",
		Resource: "SMF",
	}
	upf := schema.GroupVersionResource{
		Group:    "nfs.axyom.casa-systems.io",
		Version:  "v1alpha1",
		Resource: "UPF",
	}

	// Retrieve CRD instances
	smfInstanceList, err := dynamicClient.Resource(smf).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Error("there is no smf instance, error: ", err)
		return nil, err
	}
	upfInstanceList, err := dynamicClient.Resource(upf).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Error("there is no smf instance, error: ", err)
		return nil, err
	}

	for _, item := range smfInstanceList.Items {
		fmt.Println("smf instance item: ")
		fmt.Println("%+v\n", item)
	}
	for _, item := range upfInstanceList.Items {
		fmt.Println("upf instance item: ")
		fmt.Println("%+v\n", item)
	}

	// Get the CRDs and print their names
	list, err := resource.Do().Get().List(v1.ListOptions{})
	if err != nil {
		fmt.Println("Could not list CRDs")
		panic(err.Error())
	}
	for _, item := range list.Items {
		fmt.Printf("CRD name: %s\n", item.GetName())
	}
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
