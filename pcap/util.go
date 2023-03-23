package pcap

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type S3Config struct {
	AccessKeyId     string
	SecretAccessKey string
	Endpoint        string
}

func GetK8sClient() (client.Client, error) {
	log.Info("get k8s client")
	k8sClient, err := client.New(ctrl.GetConfigOrDie(), client.Options{Scheme: runtime.NewScheme()})
	if err != nil {
		log.Error("unable to get k8s client, error: ", err)
		return nil, err
	}
	return k8sClient, nil
}

func GetS3Config() (*S3Config, error) {
	log.Error("获取config")
	k8sClient, err := client.New(ctrl.GetConfigOrDie(), client.Options{Scheme: runtime.NewScheme()})
	if err != nil {
		log.Panic(err)
	}
	// get CRD object by name
	// crdName := "smf.axyom.casa-systems.io/v1alpha1.SMF"
	ctx := context.Background()

	crd := &unstructured.Unstructured{}
	// create API resource reference
	crd.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "smf.axyom.casa-systems.io",
		Version: "v1alpha1",
		Kind:    "SMF",
	})
	err = k8sClient.Get(ctx, client.ObjectKey{
		Namespace: "default",
		Name:      "smf1",
	}, crd)
	if err != nil {
		log.Panic(err)
		return nil, err
	}
	s3Storage, _, err := unstructured.NestedMap(crd.Object, "spec", "tracing", "pcapConfig", "s3Storage")
	if err != nil {
		log.Panic(err)
		return nil, err
	}
	s3Config := &S3Config{
		AccessKeyId:     s3Storage["accessKeyId"].(string),
		SecretAccessKey: s3Storage["secretAccessKey"].(string),
		Endpoint:        s3Storage["endpoint"].(string),
	}
	log.Error("获取config成功")
	return s3Config, nil
}

func GetS3Client() (*s3.Client, error) {
	s3Config, err := GetS3Config()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	customCredentialProvider := credentials.NewStaticCredentialsProvider(s3Config.AccessKeyId, s3Config.SecretAccessKey, "")

	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               s3Config.Endpoint,
			HostnameImmutable: true,
		}, nil
	})

	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(customCredentialProvider),
	)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// Create an Amazon S3 service client
	s3client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		// o.Region = "us-west-2"
		// o.UseAccelerate = true
	})
	return s3client, nil
}
