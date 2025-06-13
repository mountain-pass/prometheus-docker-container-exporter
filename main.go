package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/metrics", metricsHandler)
	
	log.Printf("Starting server on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	
	// Create Docker client using the Unix socket
	cli, err := client.NewClientWithOpts(
		client.WithHost("unix:///var/run/docker.sock"),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create Docker client: %v", err), http.StatusInternalServerError)
		return
	}
	defer cli.Close()

	// List all containers
	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list containers: %v", err), http.StatusInternalServerError)
		return
	}

	// Get Docker server version info
	version, err := cli.ServerVersion(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get Docker version: %v", err), http.StatusInternalServerError)
		return
	}

	// Set content type for Prometheus metrics
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

	// Docker API version metric
	fmt.Fprintln(w, "# HELP docker_api_version Docker API version information")
	fmt.Fprintln(w, "# TYPE docker_api_version gauge")
	fmt.Fprintf(w, `docker_api_version %s`+"\n", version.APIVersion)

	// Generate Prometheus metrics
	fmt.Fprintln(w, "# HELP docker_container_info Information about Docker containers")
	fmt.Fprintln(w, "# TYPE docker_container_info gauge")
	
	fmt.Fprintln(w, "# HELP docker_container_state State of Docker containers (1=running, 0=stopped)")
	fmt.Fprintln(w, "# TYPE docker_container_state gauge")

	fmt.Fprintln(w, "# HELP docker_container_health Health status of Docker containers (1=healthy, 0=unhealthy, -1=no_healthcheck)")
	fmt.Fprintln(w, "# TYPE docker_container_health gauge")

	// Initialize counters for aggregate metrics
	var runningCount, notRunningCount, runningUnhealthyCount int

	for _, container := range containers {
		// Get container name (remove leading slash)
		name := strings.TrimPrefix(container.Names[0], "/")
		
		// Container info metric
		fmt.Fprintf(w, `docker_container_info{id="%s",name="%s",image="%s"} "%s"`+"\n",
			container.ID[:12], name, container.Image, container.Status)
		
		// Container state metric (1 for running, 0 for not running)
		stateValue := 0
		if container.State == "running" {
			stateValue = 1
		}
		fmt.Fprintf(w, `docker_container_state{id="%s",name="%s",image="%s"} %d`+"\n",
			container.ID[:12], name, container.Image, stateValue)

		// Container health metric (1=healthy, 0=unhealthy, -1=no healthcheck)
		healthValue := -1 // Default to no healthcheck
		switch container.Status {
		case "Up (healthy)":
			healthValue = 1
		case "Up (unhealthy)":
			healthValue = 0
		default:
			// Check if status contains health information
			if strings.Contains(container.Status, "(healthy)") {
				healthValue = 1
			} else if strings.Contains(container.Status, "(unhealthy)") {
				healthValue = 0
			}
		}
		fmt.Fprintf(w, `docker_container_health{id="%s",name="%s",image="%s"} %d`+"\n",
			container.ID[:12], name, container.Image, healthValue)

		// Update counters for aggregate metrics
		if container.State == "running" {
			runningCount++
			if healthValue == 0 {
				runningUnhealthyCount++
			}
		} else {
			notRunningCount++
		}
	}

	// Add a metric for total container count
	fmt.Fprintln(w, "# HELP docker_containers_total Total number of Docker containers")
	fmt.Fprintln(w, "# TYPE docker_containers_total gauge")
	fmt.Fprintf(w, "docker_containers_total %d\n", len(containers))

	// Add new aggregate metrics
	fmt.Fprintln(w, "# HELP docker_containers_total_running Total number of running Docker containers")
	fmt.Fprintln(w, "# TYPE docker_containers_total_running gauge")
	fmt.Fprintf(w, "docker_containers_total_running %d\n", runningCount)

	fmt.Fprintln(w, "# HELP docker_containers_total_not_running Total number of not running Docker containers")
	fmt.Fprintln(w, "# TYPE docker_containers_total_not_running gauge")
	fmt.Fprintf(w, "docker_containers_total_not_running %d\n", notRunningCount)

	fmt.Fprintln(w, "# HELP docker_containers_total_running_unhealthy Total number of running unhealthy Docker containers")
	fmt.Fprintln(w, "# TYPE docker_containers_total_running_unhealthy gauge")
	fmt.Fprintf(w, "docker_containers_total_running_unhealthy %d\n", runningUnhealthyCount)
}