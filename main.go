package main

import (
	"context"
	"fmt"

	"github.com/boqier/kube-mcp-server/pkg/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	client, err := k8s.NewClient("")
	if err != nil {
		panic(err)
	}
	pods, err := client.Clientset.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, pod := range pods.Items {
		fmt.Println(pod.Name)
	}
}
