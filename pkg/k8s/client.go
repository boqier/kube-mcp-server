package k8s

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
)

// k8s有多种客户端工具，主要包括以下几种：
// - clientset：类型安全的客户端，提供针对每种Kubernetes资源的强类型API，使用最广泛
// - dynamic：动态客户端，无需预定义资源结构，适合处理自定义资源或未知资源类型，在运行时动态发现资源类型
// 提高代码灵活性
// - discovery：发现客户端，用于探测API服务器支持的资源类型和版本信息。
// - metrics：指标客户端，用于访问集群中资源的监控指标数据
// 同时，加入可以缓存gvr资源的功能，减少对api server的调用次数
type Client struct {
	Clientset        *kubernetes.Clientset
	dynamicClient    dynamic.Interface
	discoveryClient  *discovery.DiscoveryClient
	metricsClient    *metricsclientset.Clientset
	restConfig       *rest.Config
	apiResourceCache map[string]*schema.GroupVersionResource
	cacheLock        sync.RWMutex
}

// 构建客户端的 rest config,使用不同的方式：按次序分为：
// - 从KUBECONFIG_DATA环境变量中加载
// - 从service account token中加载:KUBERNETES_SERVER和KUBERNETES_TOKEN
// - 从in-cluster config中加载:/var/run/secrets/kubernetes.io/serviceaccount
// - 从kubeconfig文件中加载:~/.kube/config
func BuildRestConfig(kubeconfigPath string) (*rest.Config, error) {
	//1环境变量加载，通过byte的方式构建
	if kubeconfigData := os.Getenv("KUBECONFIG_DATA"); kubeconfigData != "" {
		config, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfigData))
		if err != nil {
			return nil, errors.New("构建config失败 " + err.Error())
		}
		return config, nil
	}
	//从service account token中加载:KUBERNETES_SERVER和KUBERNETES_TOKEN
	if serverURL := os.Getenv("KUBERNETES_SERVER"); serverURL != "" {
		token := os.Getenv("KUBERNETES_TOKEN")
		if token == "" {
			return nil, fmt.Errorf("KUBERNETES_TOKEN environment variable is required when KUBERNETES_SERVER is set")
		}

		config := &rest.Config{
			Host:        serverURL,
			BearerToken: token,
			TLSClientConfig: rest.TLSClientConfig{
				Insecure: os.Getenv("KUBERNETES_INSECURE") == "true",
			},
		}

		// Set CA certificate if provided
		if caCert := os.Getenv("KUBERNETES_CA_CERT"); caCert != "" {
			config.TLSClientConfig.CAData = []byte(caCert)
		} else if caCertPath := os.Getenv("KUBERNETES_CA_CERT_PATH"); caCertPath != "" {
			caCertData, err := os.ReadFile(caCertPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read CA certificate from %s: %w", caCertPath, err)
			}
			config.TLSClientConfig.CAData = caCertData
		}

		return config, nil
	}
	//3从in-cluster config中加载:/var/run/secrets/kubernetes.io/serviceaccount
	if config, err := rest.InClusterConfig(); err == nil {
		return config, nil
	}
	//4从kubeconfig文件中加载:~/.kube/config
	var kubeconfig string
	if kubeconfigPath != "" {
		kubeconfig = kubeconfigPath
	} else if kubeconfigEnv := os.Getenv("KUBECONFIG"); kubeconfigEnv != "" {
		kubeconfig = kubeconfigEnv
	} else {
		kubeconfig = os.ExpandEnv("${HOME}/.kube/config")
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, errors.New("构建config失败 " + err.Error())
	}
	return config, nil
}

// 通过restconfig构建客户端
func NewClient(kubeconfigPath string) (*Client, error) {
	config, err := BuildRestConfig(kubeconfigPath)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.New("构建clientset失败 " + err.Error())
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, errors.New("构建dynamicClient失败 " + err.Error())
	}
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, errors.New("构建discoveryClient失败 " + err.Error())
	}
	metricsClient, err := metricsclientset.NewForConfig(config)
	if err != nil {
		return nil, errors.New("构建metricsClient失败 " + err.Error())
	}
	return &Client{
		Clientset:        clientset,
		dynamicClient:    dynamicClient,
		discoveryClient:  discoveryClient,
		metricsClient:    metricsClient,
		restConfig:       config,
		apiResourceCache: make(map[string]*schema.GroupVersionResource),
		cacheLock:        sync.RWMutex{},
	}, nil

}

// 列出所有的在集群中的资源类型
// 使用discovery client 来获取集群中的所有resource
// 分为includeNamespace和includecluster两种情况: 类似执行kubectl api-resources --namespaced=true
// includeNamespace 比如pod deployment configmap等，cluster-scoped如node namespace等
// 返回一个map slice,每个元素都是一种API 资源
func (c *Client) GetAPIResources(ctx context.Context, includeNamespaceScoped, includeClusterScoped bool) ([]map[string]interface{}, error) {
	resourcesList, err := c.discoveryClient.ServerPreferredResources()
	if err != nil && discovery.IsGroupDiscoveryFailedError(err) {
		return nil, fmt.Errorf("failed to retrieve api resource:%w", err)
	}
	var resources []map[string]interface{}
	for _, resourcesList := range resourcesList {
		for _, resource := range resourcesList.APIResources {
			if (resource.Namespaced && !includeClusterScoped) || (!resource.Namespaced && !includeClusterScoped) {
				continue
			}
			resources = append(resources, map[string]interface{}{
				"name":         resource.Name,
				"singularName": resource.SingularName,
				"namespaced":   resource.Namespaced,
				"kind":         resource.Kind,
				"group":        resource.Group,
				"version":      resource.Version,
				"verbs":        resource.Verbs,
			})
		}
	}
	return resources, nil
}

// getCachedGVR 用来获取GVR ，通过输入kind的方式，如果catch中有就直接从中获取，如果没有就先写入在获取
// 通过gvr可以方便的使用dynamicClient来进行资源的增删改等
func (c *Client) getCachedGVR(kind string) (*schema.GroupVersionResource, error) {
	c.cacheLock.RLock()
	if gvr, exists := c.apiResourceCache[kind]; exists {
		c.cacheLock.RUnlock()
		return gvr, nil
	}
	c.cacheLock.RUnlock()
	//cache miss,从discovery client获取
	resourceLists, err := c.discoveryClient.ServerPreferredResources()
	if err != nil && !discovery.IsGroupDiscoveryFailedError(err) {
		return nil, fmt.Errorf("failed to retrieve api resource:%w", err)
	}
	for _, resourceList := range resourceLists {
		gv, err := schema.ParseGroupVersion(resourceList.GroupVersion)
		if err != nil {
			continue
		}
		for _, resource := range resourceList.APIResources {
			if resource.Kind == kind {
				gvr := &schema.GroupVersionResource{
					Group:    gv.Group,
					Version:  gv.Version,
					Resource: resource.Name,
				}
				c.cacheLock.Lock()
				c.apiResourceCache[kind] = gvr
				c.cacheLock.Unlock()
				return gvr, nil
			}
		}
	}

	return nil, fmt.Errorf("resource type %s not found", kind)
}

// GetResource retrieves detailed information about a specific resource.
// It uses the dynamic client to fetch the resource by kind, name, and namespace.
// It utilizes a cached GroupVersionResource (GVR) for efficiency.
// Returns the unstructured content of the resource as a map, or an error.
func (c *Client) GetResource(ctx context.Context, kind, name, namespace string) (map[string]interface{}, error) {
	gvr, err := c.getCachedGVR(kind)
	if err != nil {
		return nil, err
	}
	//通过gvr获取资源清单
	var obj *unstructured.Unstructured
	if namespace != "" {
		obj, err = c.dynamicClient.Resource(*gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	} else {
		obj, err = c.dynamicClient.Resource(*gvr).Get(ctx, name, metav1.GetOptions{})
	}
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve resource:%w", err)
	}
	return obj.UnstructuredContent(), nil
}
