package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/c4po/terrastate/internal/api/handlers"
	"github.com/c4po/terrastate/internal/storage"
	"github.com/c4po/terrastate/internal/storage/disk"
	s3storage "github.com/c4po/terrastate/internal/storage/s3"
	"github.com/gorilla/mux"
)

var (
	Version   string = "dev"
	GitSha    string = "unknown"
	BuildTime string = "unknown"
)

func initializeStorage() (storage.StateStorage, error) {
	storageType := os.Getenv("STORAGE_TYPE")

	switch storageType {
	case "s3":
		log.Println("Using S3 storage")
		bucketName := os.Getenv("S3_BUCKET_NAME")
		if bucketName == "" {
			log.Fatal("S3_BUCKET_NAME environment variable is required for S3 storage")
		}

		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			return nil, err
		}

		s3Client := s3.NewFromConfig(cfg)
		log.Println("S3 client initialized, connecting to bucket", bucketName)
		return s3storage.NewS3Storage(s3Client, bucketName), nil

	case "local":
		log.Println("Using local storage")
		basePath := os.Getenv("STORAGE_PATH")
		if basePath == "" {
			basePath = "data"
		}
		return disk.NewDiskStorage(basePath), nil

	default:
		log.Fatalf("Unsupported storage type: %s", storageType)
		return nil, nil
	}
}

func versionHandler(w http.ResponseWriter, r *http.Request) {
	versionInfo := map[string]string{
		"version":    Version,
		"git_sha":    GitSha,
		"build_time": BuildTime,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(versionInfo)
}

// Logging Middleware to log each HTTP request method and details
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		// Log Request Details
		log.Printf("[DEBUG] %s %s %s", r.Method, r.RequestURI, r.RemoteAddr)

		// Call the next handler
		next.ServeHTTP(w, r)

		// Log Response Time
		duration := time.Since(startTime)
		log.Printf("[DEBUG] Completed in %v", duration)
	})
}

func main() {
	// Initialize storage backend
	storage, err := initializeStorage()
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Initialize handlers
	stateHandler := handlers.NewStateHandler(storage)
	discoveryHandler := handlers.NewDiscoveryHandler()
	loginHandler := handlers.NewLoginHandler()

	// Setup router
	r := mux.NewRouter()

	// Logging middleware
	r.Use(loggingMiddleware)

	// Discovery endpoint
	r.HandleFunc("/.well-known/terraform.json", discoveryHandler.GetDiscovery).Methods("GET")

	r.HandleFunc("/login", loginHandler.TerraformLogin).Methods("GET")
	r.HandleFunc("/app/settings/tokens", loginHandler.Tokens).Methods("GET")
	r.HandleFunc("/app/settings/tokens/create", loginHandler.CreateToken).Methods("POST")

	// State endpoints
	r.HandleFunc("/state/{workspace}/{id}", stateHandler.GetState).Methods("GET")
	r.HandleFunc("/state/{workspace}/{id}", stateHandler.PutState).Methods("PUT")
	r.HandleFunc("/state/{workspace}/{id}", stateHandler.DeleteState).Methods("DELETE")
	r.HandleFunc("/state/{workspace}", stateHandler.ListStates).Methods("GET")

	// Lock endpoints
	r.HandleFunc("/lock/{workspace}/{id}", stateHandler.Lock).Methods("POST")
	r.HandleFunc("/lock/{workspace}/{id}", stateHandler.Unlock).Methods("DELETE")

	// Version endpoint
	r.HandleFunc("/version", versionHandler).Methods("GET")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
