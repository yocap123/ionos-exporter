package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

type ValkeyV2Cluster struct {
	ID      string
	Name    string
	Version string
}

type ValkeyV2ItemProperties struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ValkeyV2Item struct {
	ID         string                 `json:"id"`
	Properties ValkeyV2ItemProperties `json:"properties"`
}

type ValkeyV2Collection struct {
	Items []ValkeyV2Item `json:"items"`
}

var (
	valkeyV2Clusters   []ValkeyV2Cluster
	valkeyV2FetchError float64
	valkeyV2Mutex      sync.RWMutex
)

func ValkeyV2CollectResources(cycletime int32) {
	for {
		scrapeValkeyV2()
		time.Sleep(time.Duration(cycletime) * time.Second)
	}
}

func scrapeValkeyV2() {
	token := os.Getenv("IONOS_TOKEN")
	if token == "" {
		fmt.Fprintln(os.Stderr, "TOKEN environment variable is not set")
		os.Exit(1)
	}

	endpoints := []string{
		"https://in-memory-db.de-txl.ionos.com/v2/clusters",
		"https://in-memory-db.de-fra.ionos.com/v2/clusters",
	}

	clusters := make([]ValkeyV2Cluster, 0)
	fetchError := 0.0

	for _, endpoint := range endpoints {
		collection, err := fetchValkeyV2Clusters(token, endpoint)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching %s: %v\n", endpoint, err)
			fetchError = 1.0
			continue
		}
		for idx := range collection.Items {
			item := collection.Items[idx]

			clusterID := "UNKNOWN"
			if item.ID != "" {
				clusterID = item.ID
			}
			clusterName := "UNKNOWN"
			if item.Properties.Name != "" {
				clusterName = item.Properties.Name
			}
			clusterVersion := "UNKNOWN"
			if item.Properties.Version != "" {
				clusterVersion = item.Properties.Version
			}

			clusters = append(clusters, ValkeyV2Cluster{
				ID:      clusterID,
				Name:    clusterName,
				Version: clusterVersion,
			})
		}
	}

	valkeyV2Mutex.Lock()
	valkeyV2Clusters = clusters
	valkeyV2FetchError = fetchError
	valkeyV2Mutex.Unlock()
}

func fetchValkeyV2Clusters(token, url string) (*ValkeyV2Collection, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status code %d from %s", resp.StatusCode, url)
	}

	var collection ValkeyV2Collection
	if err := json.NewDecoder(resp.Body).Decode(&collection); err != nil {
		return nil, err
	}

	return &collection, nil
}
