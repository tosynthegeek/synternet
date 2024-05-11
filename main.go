package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/SyntropyNet/pubsub-go/pubsub"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
)

// Configuration variables
const (
  natsUrl  = "nats://127.0.0.1:4222" // Default NATS server address
  subject  = "staking.osmosis.mempool" // The subject we want to subscribe to
)

// Access token retrieved from environment variable
var accessToken string

// PrintData function processes and logs received message data
func PrintData(ctx context.Context, service *pubsub.NatsService, data []byte) error {
  log.Println("Received message on subject:", subject)
  fmt.Println(string(data)) // Print the message data as a string
  return nil
}

func main() {
  // Load environment variables from .env
  err := godotenv.Load()
  if err != nil {
    log.Fatal("Error loading .env file: ", err)
  }

  // Read access token from the environment variable "ACCESS_TOKEN"
  accessToken = os.Getenv("ACCESS_TOKEN")
  fmt.Println("Access Token:", accessToken) // For debugging, remove for security issues

  // Attempt to create a JWT token using the access token from the environment
  jwt, err := pubsub.CreateAppJwt(accessToken)
  if err != nil {
    log.Println("Error creating JWT token:", err)
  } else {
    fmt.Println("Generated JWT token:", jwt) // Log the generated JWT token (for debugging)
  }

  // NATS connection options with User JWT and Seed (access token) for authentication
  opts := []nats.Option{
    nats.UserJWTAndSeed(jwt, accessToken),
  }

  // Connect to the NATS server using the configured URL and options
  service := pubsub.MustConnect(pubsub.Config{
    URI:  natsUrl,
    Opts: opts,
  })
  log.Println("Connected to NATS server successfully.")

  // Create a context with cancellation functionality
  ctx, cancel := context.WithCancel(context.Background())
  defer cancel() // Ensure cancellation on exit

  // Informational message
  fmt.Println("Waiting for messages...")

  // Subscribe to the specified subject using a handler function
  service.AddHandler(subject, func(data []byte) error {
    fmt.Println("Received message data:")
    return PrintData(ctx, service, data) // Process data using PrintData function
  })

  signalChan := make(chan os.Signal, 1)
  signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
  go func() {
    <-signalChan
    cancel() // Cancel the context on receiving SIGINT or SIGTERM
  }()

  // Start serving messages and handle them using the registered handler
  service.Serve(ctx)

  fmt.Println("Exiting application...")
}
