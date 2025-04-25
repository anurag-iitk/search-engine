package server

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"backend/pkg/index"
	"backend/pkg/parquet"
)

type SearchRequest struct {
	Query string `json:"query"`
	Field string `json:"field,omitempty"`
}

type SearchResponse struct {
	Results    []index.SearchResult `json:"results"`
	TotalHits  int                  `json:"totalHits"`
	SearchTime string               `json:"searchTime"`
}

type StatsResponse struct {
	TotalDocuments int    `json:"totalDocuments"`
	SearchCount    int    `json:"searchCount"`
	AvgSearchTime  string `json:"avgSearchTime"`
}

type UploadResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	NewDocuments int    `json:"newDocuments,omitempty"`
}

func withCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow all origins (you can restrict this in production)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		h.ServeHTTP(w, r)
	})
}

func StartServer(index *index.SearchIndex, port string, dataDir string) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse multipart form
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		files := r.MultipartForm.File["files"]
		if len(files) == 0 {
			response := UploadResponse{
				Success: false,
				Message: "No files uploaded",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		var totalNew int
		for _, fileHeader := range files {
			file, err := fileHeader.Open()
			if err != nil {
				log.Printf("Error opening uploaded file: %v", err)
				continue
			}
			defer file.Close()

			// Save the file
			filePath := filepath.Join(dataDir, fileHeader.Filename)
			dst, err := os.Create(filePath)
			if err != nil {
				log.Printf("Error creating file: %v", err)
				continue
			}
			defer dst.Close()

			if _, err := io.Copy(dst, file); err != nil {
				log.Printf("Error copying file: %v", err)
				continue
			}

			// Process the file
			records, err := parquet.ReadParquetFile(filePath)
			if err != nil {
				log.Printf("Error reading Parquet file: %v", err)
				continue
			}

			index.AddDocuments(records)
			totalNew += len(records)
		}

		response := UploadResponse{
			Success:      true,
			Message:      "Files processed successfully",
			NewDocuments: totalNew,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	mux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req SearchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		results, duration := index.Search(req.Query, req.Field)
		response := SearchResponse{
			Results:    results,
			TotalHits:  len(results),
			SearchTime: duration.String(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	mux.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		searchCount, avgTime := index.GetStats()
		response := StatsResponse{
			TotalDocuments: index.GetDocumentCount(),
			SearchCount:    searchCount,
			AvgSearchTime:  avgTime.String(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	return http.ListenAndServe(":"+port, withCORS(mux))
}
