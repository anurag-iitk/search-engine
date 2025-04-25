package index

import (
	"encoding/json"
	"log"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"backend/pkg/parquet"
)

type Document struct {
	ID      int
	Record  parquet.Record
	Matches map[string]int // field -> match count
}

type SearchIndex struct {
	documents  []Document
	index      map[string][]int            // term -> document IDs
	fieldIndex map[string]map[string][]int // field -> term -> document IDs
	lock       sync.RWMutex
	stats      struct {
		searchCount     int
		totalSearchTime time.Duration
	}
}

// Add this method to the SearchIndex
func (si *SearchIndex) GetStats() (int, time.Duration) {
	si.lock.RLock()
	defer si.lock.RUnlock()

	if si.stats.searchCount == 0 {
		return 0, 0
	}

	return si.stats.searchCount, si.stats.totalSearchTime / time.Duration(si.stats.searchCount)
}

func NewSearchIndex() *SearchIndex {
	return &SearchIndex{
		index:      make(map[string][]int),
		fieldIndex: make(map[string]map[string][]int),
	}
}

func (si *SearchIndex) AddDocuments(records []parquet.Record) {
	si.lock.Lock()
	defer si.lock.Unlock()

	startID := len(si.documents)
	for i, record := range records {
		docID := startID + i
		doc := Document{
			ID:      docID,
			Record:  record,
			Matches: make(map[string]int),
		}
		si.documents = append(si.documents, doc)

		// Index all fields
		si.indexDocument(docID, record)
	}
}

func (si *SearchIndex) indexDocument(docID int, record parquet.Record) {
	// Index Message field
	si.indexField("message", record.Message, docID)

	// Index MessageRaw field
	si.indexField("messageraw", record.MessageRaw, docID)

	// Index Tag field
	si.indexField("tag", record.Tag, docID)

	// Index Sender field
	si.indexField("sender", record.Sender, docID)

	// Parse Groupings (assumed to be JSON array of strings)
	var groupings []string
	if err := json.Unmarshal([]byte(record.Groupings), &groupings); err == nil {
		for _, grouping := range groupings {
			si.indexField("groupings", grouping, docID)
		}
	} else {
		log.Printf("warning: failed to parse Groupings for doc %d: %v", docID, err)
	}

	// Index Event field
	si.indexField("event", record.Event, docID)

	// Index Namespace field
	si.indexField("namespace", record.Namespace, docID)
}

func (si *SearchIndex) indexField(field, text string, docID int) {
	if text == "" {
		return
	}

	terms := tokenize(text)
	for _, term := range terms {
		// Update global index
		si.index[term] = append(si.index[term], docID)

		// Update field-specific index
		if _, ok := si.fieldIndex[field]; !ok {
			si.fieldIndex[field] = make(map[string][]int)
		}
		si.fieldIndex[field][term] = append(si.fieldIndex[field][term], docID)
	}
}

func tokenize(text string) []string {
	// Simple tokenizer - split on whitespace and remove punctuation
	text = strings.ToLower(text)
	text = strings.ReplaceAll(text, ".", "")
	text = strings.ReplaceAll(text, ",", "")
	text = strings.ReplaceAll(text, "!", "")
	text = strings.ReplaceAll(text, "?", "")
	return strings.Fields(text)
}

type SearchResult struct {
	Document Document
	Score    float64
}

func (si *SearchIndex) Search(query string, field string) ([]SearchResult, time.Duration) {
	startTime := time.Now()
	si.lock.RLock()
	defer si.lock.RUnlock()

	terms := tokenize(query)
	if len(terms) == 0 {
		return nil, time.Since(startTime)
	}

	initialCapacity := len(terms) * 50
	scores := make(map[int]float64, initialCapacity)
	termDocIDs := make([][]int, 0, len(terms))

	// Determine index
	var indexToUse map[string][]int
	if field != "" {
		if fieldIdx, ok := si.fieldIndex[strings.ToLower(field)]; ok {
			indexToUse = fieldIdx
		} else {
			return nil, time.Since(startTime)
		}
	} else {
		indexToUse = si.index
	}

	// Compute IDF
	idfs := make([]float64, len(terms))
	totalDocs := float64(len(si.documents))

	for i, term := range terms {
		if docIDs, ok := indexToUse[term]; ok {
			df := float64(len(docIDs))
			idfs[i] = 1.0 + math.Log(totalDocs/(df+1))
			termDocIDs = append(termDocIDs, docIDs)
		} else {
			idfs[i] = 0
			termDocIDs = append(termDocIDs, nil)
		}
	}

	// TF-IDF scoring
	for i, term := range terms {
		if termDocIDs[i] == nil {
			continue
		}

		var contentFunc func(*parquet.Record) string
		if field != "" {
			switch strings.ToLower(field) {
			case "message":
				contentFunc = func(r *parquet.Record) string { return r.Message }
			case "messageraw":
				contentFunc = func(r *parquet.Record) string { return r.MessageRaw }
			case "tag":
				contentFunc = func(r *parquet.Record) string { return r.Tag }
			case "sender":
				contentFunc = func(r *parquet.Record) string { return r.Sender }
			case "groupings":
				contentFunc = func(r *parquet.Record) string {
					var groupings []string
					if err := json.Unmarshal([]byte(r.Groupings), &groupings); err != nil {
						log.Printf("warning: failed to parse groupings: %v", err)
						return ""
					}
					return strings.Join(groupings, " ")
				}
			case "event":
				contentFunc = func(r *parquet.Record) string { return r.Event }
			case "namespace":
				contentFunc = func(r *parquet.Record) string { return r.Namespace }
			}
		}

		idf := idfs[i]
		for _, docID := range termDocIDs[i] {
			if _, exists := scores[docID]; exists && i > 0 {
				continue
			}

			tfDoc := 1.0
			if contentFunc != nil {
				content := strings.ToLower(contentFunc(&si.documents[docID].Record))
				tfDoc = float64(strings.Count(content, term))
			}

			scores[docID] += 1.0 * tfDoc * idf
		}
	}

	results := make([]SearchResult, 0, len(scores))
	for docID, score := range scores {
		results = append(results, SearchResult{
			Document: si.documents[docID],
			Score:    score,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Update stats
	duration := time.Since(startTime)
	si.lock.RUnlock() // unlock read lock to take write lock
	si.lock.Lock()
	si.stats.searchCount++
	si.stats.totalSearchTime += duration
	si.lock.Unlock()
	si.lock.RLock() // re-acquire read lock for return (optional safety)

	return results, duration
}

func (si *SearchIndex) GetDocumentCount() int {
	si.lock.RLock()
	defer si.lock.RUnlock()
	return len(si.documents)
}
