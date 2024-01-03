package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/hellofresh/health-go/v5"
	"github.com/joho/godotenv"
	"github.com/tailscale/tailscale-client-go/tailscale"
	"tailscale.com/tsnet"
	"tailscale.com/types/logger"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using env vars")
	}
	debug := os.Getenv("DEBUG")
	apiKey := os.Getenv("TAILSCALE_API_KEY")
	tailnet := os.Getenv("TAILSCALE_TAILNET")
	selectedMachineName := os.Getenv("HEALTHCHECK_MACHINE_NAME")
	selectedPort := os.Getenv("HEALTHCHECK_PORT")

	client, err := tailscale.NewClient(apiKey, tailnet)
	if err != nil {
		log.Fatalln(err)
	}
	authKey, err := client.CreateKey(context.Background(), tailscale.KeyCapabilities{})
	if err != nil {
		log.Fatalln(err)
	}

	devices, err := client.Devices(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
	selectedDevice, err := findSelectedMachine(selectedMachineName, devices)
	if err != nil {
		log.Fatalln(err)
	}

	h, err := health.New(health.WithChecks(health.Config{
		Name:    fmt.Sprintf("health check of device: %s", selectedMachineName),
		Timeout: 30 * time.Second,
		Check: func(ctx context.Context) error {
			var loggerOutput io.Writer
			if debug != "" {
				loggerOutput = os.Stdout
			} else {
				loggerOutput = io.Discard
			}
			logger := log.New(loggerOutput, "[tsnet]", log.Default().Flags())

			return makeTailscaleHTTPRequest(
				authKey.Key,
				fmt.Sprintf("%s:%s", selectedDevice.Addresses[0], selectedPort),
				logger.Printf,
			)
		},
	}))
	if err != nil {
		log.Fatalln(err)
	}
	result := h.Measure(context.Background())
	if len(result.Failures) > 0 {
		log.Fatalf("Healthcheck failed with error(s): %#v", result.Failures)
	}
	log.Printf("Healthcheck result: %s\n", result.Status)
}

func findSelectedMachine(selectedMachineName string, devices []tailscale.Device) (tailscale.Device, error) {
	for _, d := range devices {
		if d.Hostname == selectedMachineName {
			log.Printf("Found machine with hostname %s, has IP addresses: %s", d.Hostname, d.Addresses)
			return d, nil
		}
	}
	return tailscale.Device{}, errors.New("unable to find selected machine")
}

func makeTailscaleHTTPRequest(authKey string, address string, logger logger.Logf) error {
	srv := &tsnet.Server{
		AuthKey: authKey,
		Logf:    logger,
	}

	status, err := srv.Up(context.Background())
	defer srv.Close()
	if err != nil {
		log.Fatal(err)
	}
	cli := srv.HTTPClient()
	url := fmt.Sprintf("http://%s", address)

	log.Printf("Created HTTP client with IP %s", status.TailscaleIPs)
	log.Printf("Making request to: %s\n", url)

	resp, err := cli.Get(url)

	if err != nil {
		return err
	}
	if resp == nil {
		return errors.New("response to HTTP request was unexpectedly nil")
	}
	if resp.StatusCode > 299 {
		return fmt.Errorf("unexpected HTTP response status: %s", resp.Status)
	}
	log.Printf("HTTP request successful, status code: %s", resp.Status)
	return nil
}
