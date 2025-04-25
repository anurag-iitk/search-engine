package parquet

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/parquet"
	"github.com/xitongsys/parquet-go/reader"
)

type Record struct {
	MsgId          string `parquet:"name=MsgId, type=BYTE_ARRAY, convertedtype=UTF8"`
	PartitionId    uint64 `parquet:"name=PartitionId, type=INT64, convertedtype=UINT_64"`
	Timestamp      string `parquet:"name=Timestamp, type=BYTE_ARRAY, convertedtype=UTF8"`
	Hostname       string `parquet:"name=Hostname, type=BYTE_ARRAY, convertedtype=UTF8"`
	Priority       int32  `parquet:"name=Priority, type=INT32"`
	Facility       int32  `parquet:"name=Facility, type=INT32"`
	FacilityString string `parquet:"name=FacilityString, type=BYTE_ARRAY, convertedtype=UTF8"`
	Severity       int32  `parquet:"name=Severity, type=INT32"`
	SeverityString string `parquet:"name=SeverityString, type=BYTE_ARRAY, convertedtype=UTF8"`
	AppName        string `parquet:"name=AppName, type=BYTE_ARRAY, convertedtype=UTF8"`
	ProcId         string `parquet:"name=ProcId, type=BYTE_ARRAY, convertedtype=UTF8"`
	Message        string `parquet:"name=Message, type=BYTE_ARRAY, convertedtype=UTF8"`
	MessageRaw     string `parquet:"name=MessageRaw, type=BYTE_ARRAY, convertedtype=UTF8"`
	StructuredData string `parquet:"name=StructuredData, type=BYTE_ARRAY, convertedtype=UTF8"`
	Tag            string `parquet:"name=Tag, type=BYTE_ARRAY, convertedtype=UTF8"`
	Sender         string `parquet:"name=Sender, type=BYTE_ARRAY, convertedtype=UTF8"`
	Groupings      string `parquet:"name=Groupings, type=BYTE_ARRAY, convertedtype=UTF8"`
	Event          string `parquet:"name=Event, type=BYTE_ARRAY, convertedtype=UTF8"`
	EventId        string `parquet:"name=EventId, type=BYTE_ARRAY, convertedtype=UTF8"`
	NanoTimeStamp  string `parquet:"name=NanoTimeStamp, type=BYTE_ARRAY, convertedtype=UTF8"`
	Namespace      string `parquet:"name=namespace, type=BYTE_ARRAY, convertedtype=UTF8"`
}

func isParquetFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Check PAR1 magic number
	magic := make([]byte, 4)
	if _, err := file.Read(magic); err != nil {
		return fmt.Errorf("could not read magic number: %v", err)
	}
	if string(magic) != "PAR1" {
		return fmt.Errorf("invalid magic number: %v", magic)
	}

	// Check footer
	stat, err := file.Stat()
	if err != nil {
		return err
	}
	if _, err := file.Seek(stat.Size()-4, io.SeekStart); err != nil {
		return err
	}
	if _, err := file.Read(magic); err != nil {
		return fmt.Errorf("could not read footer: %v", err)
	}
	if string(magic) != "PAR1" {
		return fmt.Errorf("invalid footer: %v", magic)
	}

	return nil
}

func ReadParquetFile(filePath string) ([]Record, error) {
	if err := isParquetFile(filePath); err != nil {
		return nil, fmt.Errorf("invalid Parquet file %s: %v", filePath, err)
	}

	var records []Record

	fr, err := local.NewLocalFileReader(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer fr.Close()

	// Use a concrete sample of the struct for schema binding
	pr, err := reader.NewParquetReader(fr, new(Record), 4)
	if err != nil {
		return nil, fmt.Errorf("failed to create parquet reader: %w", err)
	}
	defer pr.ReadStop()

	// Validate schema if necessary
	if err := validateSchema(pr); err != nil {
		return nil, fmt.Errorf("schema validation failed: %w", err)
	}

	numRows := int(pr.GetNumRows())
	if numRows <= 0 {
		return records, nil
	}

	log.Printf("Reading %d rows from Parquet file: %s", numRows, filePath)

	batchSize := 1000
	for i := 0; i < numRows; i += batchSize {
		count := batchSize
		if i+batchSize > numRows {
			count = numRows - i
		}

		// Use *[]Record because Read expects a pointer to a slice of the correct type
		recs := make([]Record, count) // not zero-length!
		if err := pr.Read(&recs); err != nil {
			log.Printf("Read error at batch %d: %v", i/batchSize, err)
			return nil, fmt.Errorf("failed to read batch at row %d: %w", i, err)
		}
		records = append(records, recs...)
	}

	return records, nil
}

func validateSchema(pr *reader.ParquetReader) error {
	// Expected fields from your Record struct
	expectedFields := map[string]parquet.Type{
		"MsgId":          parquet.Type_BYTE_ARRAY,
		"PartitionId":    parquet.Type_INT64,
		"Timestamp":      parquet.Type_BYTE_ARRAY,
		"Hostname":       parquet.Type_BYTE_ARRAY,
		"Priority":       parquet.Type_INT32,
		"Facility":       parquet.Type_INT32,
		"FacilityString": parquet.Type_BYTE_ARRAY,
		"Severity":       parquet.Type_INT32,
		"SeverityString": parquet.Type_BYTE_ARRAY,
		"AppName":        parquet.Type_BYTE_ARRAY,
		"ProcId":         parquet.Type_BYTE_ARRAY,
		"Message":        parquet.Type_BYTE_ARRAY,
		"MessageRaw":     parquet.Type_BYTE_ARRAY,
		"StructuredData": parquet.Type_BYTE_ARRAY,
		"Tag":            parquet.Type_BYTE_ARRAY,
		"Sender":         parquet.Type_BYTE_ARRAY,
		"Groupings":      parquet.Type_BYTE_ARRAY,
		"Event":          parquet.Type_BYTE_ARRAY,
		"EventId":        parquet.Type_BYTE_ARRAY,
		"NanoTimeStamp":  parquet.Type_BYTE_ARRAY,
		"Namespace":      parquet.Type_BYTE_ARRAY,
	}

	availableFields := make(map[string]parquet.Type)
	for _, se := range pr.SchemaHandler.SchemaElements {
		if se.Name != "" && se.Type != nil {
			availableFields[se.Name] = *se.Type
		}
	}

	for fieldName, expectedType := range expectedFields {
		actualType, exists := availableFields[fieldName]
		if !exists {
			return fmt.Errorf("missing required field: %s", fieldName)
		}
		if actualType != expectedType {
			return fmt.Errorf("type mismatch for field %s: expected %s, got %s",
				fieldName, expectedType, actualType)
		}
	}

	return nil
}
