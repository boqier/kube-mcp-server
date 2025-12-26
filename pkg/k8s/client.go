package k8s

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
	"sigs.k8s.io/yaml"
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
			return nil, fmt.Errorf("构建config失败 %w", err)
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
		return nil, fmt.Errorf("构建config失败 %w", err)
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
		return nil, fmt.Errorf("构建clientset失败 %w", err)
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("构建dynamicClient失败 %w", err)
	}
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("构建discoveryClient失败 %w", err)
	}
	metricsClient, err := metricsclientset.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("构建metricsClient失败 %w", err)
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

func (c *Client) ListResources(ctx context.Context, kind, namespace, labelSelector, fieldSelector string) ([]map[string]interface{}, error) {
	gvr, err := c.getCachedGVR(kind)
	if err != nil {
		return nil, err
	}
	options := metav1.ListOptions{
		LabelSelector: labelSelector,
		FieldSelector: fieldSelector,
	}
	var list *unstructured.UnstructuredList
	if namespace != "" {
		list, err = c.dynamicClient.Resource(*gvr).Namespace(namespace).List(ctx, options)
	} else {
		list, err = c.dynamicClient.Resource(*gvr).List(ctx, options)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to list resources%w", err)
	}
	var resources []map[string]interface{}
	for _, item := range list.Items {
		metadata := item.GetLabels()
		resources = append(resources, map[string]interface{}{
			"name":      item.GetName(),
			"kind":      item.GetKind(),
			"namespace": item.GetNamespace(),
			"lables":    metadata,
		})
	}
	return resources, err

}

// 通过manifest的方式创建或者更新一个资源，创建成功会返回对应资源的结构
func (c *Client) CreateOrUpdateResoureceJSON(ctx context.Context, namespace, manifestJSON, kind string) (map[string]interface{}, error) {
	obj := &unstructured.Unstructured{}
	if err := json.Unmarshal([]byte(manifestJSON), &obj.Object); err != nil {
		return nil, fmt.Errorf("failed to parse resourfce manifest JSON %w", err)
	}
	//获取资源gvr
	gvr, err := c.getCachedGVR(kind)
	if err == nil {
		return nil, err
	}
	//看对应的ns是否存在
	_, err = c.Clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err == nil {
		fmt.Printf("namespace %s exists\n", namespace)
	}
	if errors.IsNotFound(err) {
		fmt.Printf("Namespace %s does not exist,creating one\n", namespace)
		_, err = c.Clientset.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"kubernetes.io/metadata.name": namespace,
				},
				Name: namespace,
			},
			Spec: corev1.NamespaceSpec{
				Finalizers: []corev1.FinalizerName{
					corev1.FinalizerKubernetes,
				},
			},
			Status: corev1.NamespaceStatus{
				Phase:      corev1.NamespaceActive,
				Conditions: nil,
			},
		}, metav1.CreateOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create namespace %s:%w", namespace, err)
		}
	}

	obj.SetNamespace(namespace)
	if obj.GetName() == "" {
		return nil, fmt.Errorf("resource name is requird")
	}
	resource := c.dynamicClient.Resource(*gvr).Namespace(obj.GetNamespace())
	rawJSON := []byte(manifestJSON)
	//直接尝试更新
	result, err := resource.Patch(
		ctx,
		obj.GetName(),
		types.MergePatchType,
		rawJSON,
		metav1.PatchOptions{},
	)
	//说明没有资源需要创建
	if errors.IsNotFound(err) {
		result, err := resource.Create(ctx, obj, metav1.CreateOptions{})
		if err != nil {
			return nil, fmt.Errorf("falied to create or patch resopurce:%w", err)
		}
		return result.UnstructuredContent(), nil
	}
	return result.UnstructuredContent(), nil
}

// CreateOrUpdateResourceYAML 用创建一个新资源
// 先将yaml转换为json，然后使用CreateOrUpdateJSON
func (c *Client) CreateOrUpdateResourceYAML(ctx context.Context, namespace, yamlManifest, kind string) (map[string]interface{}, error) {
	jsonData, err := yaml.YAMLToJSON([]byte(yamlManifest))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve yaml manifest:%w", err)
	}
	//将json转换为 unstructured object
	obj := &unstructured.Unstructured{}
	if err := json.Unmarshal(jsonData, &obj.Object); err != nil {
		return nil, fmt.Errorf("failed to parse converted JSON From manifest:%w", err)
	}
	resourceKind := kind
	if resourceKind == "" {
		resourceKind = obj.GetKind()
		if resourceKind == "" {
			return nil, fmt.Errorf("resources is required ,either provide it as a parameter or include it in the YAML manifest")
		}
	}
	gvr, err := c.getCachedGVR(resourceKind)
	if err != nil {
		return nil, err
	}
	//看对应的ns是否存在
	_, err = c.Clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err == nil {
		fmt.Printf("namespace %s exists\n", namespace)
	}
	if errors.IsNotFound(err) {
		fmt.Printf("Namespace %s does not exist,creating one\n", namespace)
		_, err = c.Clientset.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"kubernetes.io/metadata.name": namespace,
				},
				Name: namespace,
			},
			Spec: corev1.NamespaceSpec{
				Finalizers: []corev1.FinalizerName{
					corev1.FinalizerKubernetes,
				},
			},
			Status: corev1.NamespaceStatus{
				Phase:      corev1.NamespaceActive,
				Conditions: nil,
			},
		}, metav1.CreateOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create namespace %s:%w", namespace, err)
		}
	}

	if namespace != "" {
		obj.SetNamespace(namespace)
	}
	if obj.GetName() == "" {
		return nil, fmt.Errorf("resource name is required in YAML manifest")
	}
	resource := c.dynamicClient.Resource(*gvr).Namespace(obj.GetNamespace())
	result, err := resource.Patch(
		ctx,
		obj.GetName(),
		types.MergePatchType,
		jsonData,
		metav1.PatchOptions{},
	)
	if errors.IsNotFound(err) {
		result, err = resource.Create(ctx, obj, metav1.CreateOptions{})
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create or patch resource from YAML manifest: %w", err)
	}

	return result.UnstructuredContent(), nil
}

func (c *Client) DeleteResource(ctx context.Context, kind, name, namespace string) error {
	gvr, err := c.getCachedGVR(kind)
	if err != nil {
		return err
	}
	var deleteErr error
	if namespace != "" {
		deleteErr = c.dynamicClient.Resource(*gvr).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	} else {
		deleteErr = c.dynamicClient.Resource(*gvr).Delete(ctx, name, metav1.DeleteOptions{})
	}
	if deleteErr != nil {
		return fmt.Errorf("failed to delete resource: %w", deleteErr)
	}
	return nil
}

// 使用dynamic client来获取资源的describe，传入kind,name,namespace参数
// 返回资源的unstructured content通过map[string]interface{}返回
// 其实和getresource一样
func (c *Client) DescribeResource(ctx context.Context, kind, name, namespace string) (map[string]interface{}, error) {
	gvr, err := c.getCachedGVR(kind)
	if err != nil {
		return nil, err
	}
	var obj *unstructured.Unstructured
	if namespace == "" {
		obj, err = c.dynamicClient.Resource(*gvr).Get(ctx, name, metav1.GetOptions{})
	} else {
		obj, err = c.dynamicClient.Resource(*gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve resource: %w", err)
		}
	}

	return obj.UnstructuredContent(), nil
}

// 使用clientset客户端获取日志，传入命名空间，pod名，容器名，以及行数参数
// 返回日志字符串
// 后面会加上从loki获取日志，支持更复杂的日志过滤策略
func (c *Client) GetPodsLogs(ctx context.Context, namespace, containerName, podName string, LogstailLines int) (string, error) {
	if LogstailLines > 300 {
		LogstailLines = 300
	}
	tailLines := int64(LogstailLines)
	podLogOptions := &corev1.PodLogOptions{
		TailLines: &tailLines,
	}
	//如果制定了container的name
	if containerName != "" {
		podLogOptions.Container = containerName
		req := c.Clientset.CoreV1().Pods(namespace).GetLogs(podName, podLogOptions)
		logs, err := req.Stream(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to get logs for container %s:%w", containerName, err)
		}
		defer logs.Close()
		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, logs); err != nil {
			return "", fmt.Errorf("failed to copy logs to buffer:%w", err)
		}
		return buf.String(), nil
	}

	//如果没有传递conmtainer name的话：
	pod, err := c.Clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get pod details%w", err)
	}
	//如果只有一个container的话
	if len(pod.Spec.Containers) == 1 {
		req := c.Clientset.CoreV1().Pods(namespace).GetLogs(podName, podLogOptions)
		logs, err := req.Stream(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to get logs: %w", err)
		}
		defer logs.Close()

		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, logs); err != nil {
			return "", fmt.Errorf("failed to read logs: %w", err)
		}
		return buf.String(), nil
	}
	//如果有多个容器的话：
	var allLogs strings.Builder
	for _, container := range pod.Spec.Containers {
		containerLogOptions := podLogOptions.DeepCopy()
		containerLogOptions.Container = container.Name

		req := c.Clientset.CoreV1().Pods(namespace).GetLogs(podName, containerLogOptions)
		logs, err := req.Stream(ctx)
		if err != nil {
			allLogs.WriteString(fmt.Sprintf("\n--- Error getting logs for container %s: %v ---\n", container.Name, err))
			continue
		}

		allLogs.WriteString(fmt.Sprintf("\n--- Logs for container %s ---\n", container.Name))
		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, logs)
		logs.Close()

		if err != nil {
			allLogs.WriteString(fmt.Sprintf("Error reading logs: %v\n", err))
		} else {
			allLogs.WriteString(buf.String())
		}
	}

	return allLogs.String(), nil
}

// 获取pod的mertic信息，包括cpu和内存使用率
// 使用mertci clinet来实现
// 返回一个map,存储pod的元数据以及mertic
func (c *Client) GetPodMetrics(ctx context.Context, namespace, podName string) (map[string]interface{}, error) {
	podMertics, err := c.metricsClient.MetricsV1beta1().PodMetricses(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get mertics for pod %s in namespace %s :%w", podName, namespace, err)
	}
	//构建map
	merticRestlt := map[string]interface{}{
		"podName":    podName,
		"namespace":  namespace,
		"timestamp":  podMertics.Timestamp.Time,
		"window":     podMertics.Window.Duration.String(),
		"containers": []map[string]interface{}{},
	}
	containerMerticsList := []map[string]interface{}{}
	for _, container := range podMertics.Containers {
		containerMertics := map[string]interface{}{
			"name":   container.Name,
			"cpu":    container.Usage.Cpu(),
			"memory": container.Usage.Memory(),
		}
		containerMerticsList = append(containerMerticsList, containerMertics)
	}
	merticRestlt["containers"] = containerMerticsList
	return merticRestlt, nil
}

// 获取node节点的资源使用情况
func (c *Client) GetNodeMetrics(ctx context.Context, nodeName string) (map[string]interface{}, error) {
	nodeMetrics, err := c.metricsClient.MetricsV1beta1().NodeMetricses().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get mertics for node %s:%w", nodeName, err)
	}
	metricsResult := map[string]interface{}{
		"nodeName":  nodeName,
		"timestamp": nodeMetrics.Timestamp.Time,
		"window":    nodeMetrics.Window.Duration.String(),
		"usage": map[string]string{
			"cpu":    nodeMetrics.Usage.Cpu().String(),
			"memory": nodeMetrics.Usage.Memory().String(),
		},
	}
	return metricsResult, nil
}
