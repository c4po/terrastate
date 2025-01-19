package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

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
		bucketName := os.Getenv("S3_BUCKET_NAME")
		if bucketName == "" {
			log.Fatal("S3_BUCKET_NAME environment variable is required for S3 storage")
		}

		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			return nil, err
		}

		s3Client := s3.NewFromConfig(cfg)
		return s3storage.NewS3Storage(s3Client, bucketName), nil

	case "local":
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

func main() {
	// Initialize storage backend
	storage, err := initializeStorage()
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Initialize handlers
	stateHandler := handlers.NewStateHandler(storage)
	discoveryHandler := handlers.NewDiscoveryHandler()

	// Setup router
	r := mux.NewRouter()

	// Discovery endpoint
	r.HandleFunc("/.well-known/terraform.json", discoveryHandler.GetDiscovery).Methods("GET")

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
