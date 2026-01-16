package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"path/filepath"
)

type SecretResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func getSecretValue(namespace, secretName, key string) (string, error) {
	// Use in-cluster config if running inside a Kubernetes cluster
	if inClusterConfig, err := rest.InClusterConfig(); err == nil {
		config := inClusterConfig
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			return "", fmt.Errorf("failed to create clientset: %v", err)
		}
		return getSecretValueFromClientset(clientset, namespace, secretName, key)
	}

	// Use kubeconfig if running outside a Kubernetes cluster
	kubeconfig := filepath.Join(
		homeDir(), ".kube", "config",
	)
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return "", fmt.Errorf("failed to build config from kubeconfig: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", fmt.Errorf("failed to create clientset: %v", err)
	}
	return getSecretValueFromClientset(clientset, namespace, secretName, key)
}

func getSecretValueFromClientset(clientset *kubernetes.Clientset, namespace, secretName, key string) (string, error) {
	secret, err := clientset.CoreV1().Secrets(namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get secret: %v", err)
	}

	value, exists := secret.Data[key]
	if !exists {
		return "", fmt.Errorf("key %s not found in secret %s", key, secretName)
	}

	return string(value), nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func getSecretHandler(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	secretName := r.URL.Query().Get("secretName")
	key := r.URL.Query().Get("key")

	if namespace == "" || secretName == "" || key == "" {
		http.Error(w, "Missing namespace, secretName, or key parameter", http.StatusBadRequest)
		return
	}

	value, err := getSecretValue(namespace, secretName, key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := SecretResponse{
		Key:   key,
		Value: value,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/", getSecretHandler)
	log.Println("Starting server on :9090")
	log.Fatal(http.ListenAndServe(":9090", nil))
}
