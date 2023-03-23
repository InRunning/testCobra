/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
)

// func main() {
// 	rootCmd := cobra.Command{Use: "axyomctl"}
// 	rootCmd.AddCommand(pcap.NewCmdPcap())
// 	rootCmd.Execute()
// }

func main() {
	GetInstanceItems()
}

func GetInstanceItems() (*string, error) {
	namespace := "default"
	ctx := context.Background()
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
	return nil, nil
}
