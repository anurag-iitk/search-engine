package main

import (
	"flag"
	"log"
	"path/filepath"

	"backend/pkg/index"
	"backend/pkg/parquet"
	"backend/pkg/server"
)

func main() {
	dataDir := flag.String("data", "./data", "Directory containing Parquet files")
	port := flag.String("port", "8080", "Port to run the server on")
	flag.Parse()

	// Initialize search index
	searchIndex := index.NewSearchIndex()

	// Load all Parquet files in the data directory
	// files, err := filepath.Glob(filepath.Join(*dataDir, "*.parquet"))
	files, err := filepath.Glob(filepath.Join(*dataDir, "File*"))
	if err != nil {
		log.Fatalf("Failed to find Parquet files: %v", err)
	}

	if len(files) == 0 {
		log.Fatalf("No Parquet files found in directory: %s", *dataDir)
	}

	// Process each file
	for _, file := range files {
		log.Printf("Processing file: %s", file)
		records, err := parquet.ReadParquetFile(file)
		if err != nil {
			log.Printf("Failed to read Parquet file %s: %v", file, err)
			continue
		}

		searchIndex.AddDocuments(records)
		log.Printf("Added %d records from %s", len(records), file)
	}

	log.Printf("Indexed %d total documents", searchIndex.GetDocumentCount())

	// Start the HTTP server
	log.Printf("Starting server on port %s", *port)
	if err := server.StartServer(searchIndex, *port, *dataDir); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
