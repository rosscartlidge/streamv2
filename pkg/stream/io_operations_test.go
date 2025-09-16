package stream

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
)

// TestNewCSVSource tests CSV source creation and basic configuration
func TestNewCSVSource(t *testing.T) {
	t.Run("BasicCSVSource", func(t *testing.T) {
		csvData := "name,age,city\nAlice,30,NYC\nBob,25,LA"
		reader := strings.NewReader(csvData)
		
		source := NewCSVSource(reader)
		if source == nil {
			t.Fatal("Expected non-nil CSV source")
		}
		
		if source.Separator != ',' {
			t.Errorf("Expected comma separator, got %c", source.Separator)
		}
		
		if !source.HasHeader {
			t.Error("Expected HasHeader to be true by default")
		}
	})
}

// TestCSVSourceWithHeaders tests custom headers configuration
func TestCSVSourceWithHeaders(t *testing.T) {
	t.Run("CustomHeaders", func(t *testing.T) {
		csvData := "Alice,30,NYC\nBob,25,LA"
		reader := strings.NewReader(csvData)
		
		headers := []string{"name", "age", "location"}
		source := NewCSVSource(reader).WithHeaders(headers)
		
		if source.HasHeader {
			t.Error("Expected HasHeader to be false when using custom headers")
		}
		
		if len(source.Headers) != 3 {
			t.Errorf("Expected 3 headers, got %d", len(source.Headers))
		}
		
		if source.Headers[0] != "name" {
			t.Errorf("Expected first header 'name', got '%s'", source.Headers[0])
		}
	})
}

// TestCSVSourceWithoutHeaders tests no headers configuration
func TestCSVSourceWithoutHeaders(t *testing.T) {
	t.Run("NoHeaders", func(t *testing.T) {
		csvData := "Alice,30,NYC\nBob,25,LA"
		reader := strings.NewReader(csvData)
		
		source := NewCSVSource(reader).WithoutHeaders()
		
		if source.HasHeader {
			t.Error("Expected HasHeader to be false")
		}
	})
}

// TestCSVToStream tests CSV to stream conversion
func TestCSVToStream(t *testing.T) {
	t.Run("WithHeaders", func(t *testing.T) {
		csvData := "name,age,active\nAlice,30,true\nBob,25,false"
		reader := strings.NewReader(csvData)
		
		stream := CSVToStream(reader)
		records, err := Collect(stream)
		if err != nil {
			t.Fatalf("Failed to collect CSV stream: %v", err)
		}
		
		if len(records) != 2 {
			t.Fatalf("Expected 2 records, got %d", len(records))
		}
		
		// Check first record
		if records[0]["name"] != "Alice" {
			t.Errorf("Expected name 'Alice', got %v", records[0]["name"])
		}
		
		if records[0]["age"] != int64(30) {
			t.Errorf("Expected age 30, got %v", records[0]["age"])
		}
		
		if records[0]["active"] != true {
			t.Errorf("Expected active true, got %v", records[0]["active"])
		}
	})
	
	t.Run("WithoutHeaders", func(t *testing.T) {
		csvData := "Alice,30,NYC\nBob,25,LA"
		reader := strings.NewReader(csvData)
		
		stream := NewCSVSource(reader).WithoutHeaders().ToStream()
		records, err := Collect(stream)
		if err != nil {
			t.Fatalf("Failed to collect CSV stream: %v", err)
		}
		
		if len(records) != 2 {
			t.Fatalf("Expected 2 records, got %d", len(records))
		}
		
		// Check that default column names are generated
		if _, exists := records[0]["col0"]; !exists {
			t.Error("Expected default column name 'col0'")
		}
	})
}

// TestNewTSVSource tests TSV source creation
func TestNewTSVSource(t *testing.T) {
	t.Run("BasicTSVSource", func(t *testing.T) {
		tsvData := "name\tage\tcity\nAlice\t30\tNYC\nBob\t25\tLA"
		reader := strings.NewReader(tsvData)
		
		source := NewTSVSource(reader)
		if source == nil {
			t.Fatal("Expected non-nil TSV source")
		}
		
		if source.Separator != '\t' {
			t.Errorf("Expected tab separator, got %c", source.Separator)
		}
	})
}

// TestTSVToStream tests TSV to stream conversion
func TestTSVToStream(t *testing.T) {
	t.Run("BasicTSV", func(t *testing.T) {
		tsvData := "name\tage\tactive\nAlice\t30\ttrue\nBob\t25\tfalse"
		reader := strings.NewReader(tsvData)
		
		stream := TSVToStream(reader)
		records, err := Collect(stream)
		if err != nil {
			t.Fatalf("Failed to collect TSV stream: %v", err)
		}
		
		if len(records) != 2 {
			t.Fatalf("Expected 2 records, got %d", len(records))
		}
		
		if records[0]["name"] != "Alice" {
			t.Errorf("Expected name 'Alice', got %v", records[0]["name"])
		}
	})
}

// TestFastTSVSource tests the fast TSV implementation
func TestFastTSVSource(t *testing.T) {
	t.Run("BasicFastTSV", func(t *testing.T) {
		tsvData := "name\tage\tsalary\nAlice\t30\t75000\nBob\t25\t65000"
		reader := strings.NewReader(tsvData)
		
		source := NewFastTSVSource(reader)
		stream := source.ToStream()
		records, err := Collect(stream)
		if err != nil {
			t.Fatalf("Failed to collect fast TSV stream: %v", err)
		}
		
		if len(records) != 2 {
			t.Fatalf("Expected 2 records, got %d", len(records))
		}
		
		// Check first record
		if GetOr(records[0], "name", "") != "Alice" {
			t.Errorf("Expected name 'Alice', got %v", records[0]["name"])
		}
		if GetOr(records[0], "age", int64(0)) != int64(30) {
			t.Errorf("Expected age 30, got %v", records[0]["age"])
		}
		if GetOr(records[0], "salary", int64(0)) != int64(75000) {
			t.Errorf("Expected salary 75000, got %v", records[0]["salary"])
		}
	})
	
	t.Run("CustomSeparator", func(t *testing.T) {
		pipeData := "name|age|city\nAlice|30|NYC\nBob|25|LA"
		reader := strings.NewReader(pipeData)
		
		source := NewFastTSVSourceWithSeparator(reader, "|")
		stream := source.ToStream()
		records, err := Collect(stream)
		if err != nil {
			t.Fatalf("Failed to collect pipe-delimited stream: %v", err)
		}
		
		if len(records) != 2 {
			t.Fatalf("Expected 2 records, got %d", len(records))
		}
		
		if GetOr(records[0], "city", "") != "NYC" {
			t.Errorf("Expected city 'NYC', got %v", records[0]["city"])
		}
	})
	
	t.Run("CustomHeaders", func(t *testing.T) {
		// Data without header row
		tsvData := "Alice\t30\tNYC\nBob\t25\tLA"
		reader := strings.NewReader(tsvData)
		
		source := NewFastTSVSource(reader).WithHeaders([]string{"name", "age", "city"})
		stream := source.ToStream()
		records, err := Collect(stream)
		if err != nil {
			t.Fatalf("Failed to collect TSV with custom headers: %v", err)
		}
		
		if len(records) != 2 {
			t.Fatalf("Expected 2 records, got %d", len(records))
		}
		
		if GetOr(records[0], "name", "") != "Alice" {
			t.Errorf("Expected name 'Alice', got %v", records[0]["name"])
		}
	})
	
	t.Run("WithoutHeader", func(t *testing.T) {
		// Data without header row
		tsvData := "Alice\t30\nBob\t25"
		reader := strings.NewReader(tsvData)
		
		source := NewFastTSVSource(reader).WithoutHeader()
		stream := source.ToStream()
		records, err := Collect(stream)
		if err != nil {
			t.Fatalf("Failed to collect headerless TSV: %v", err)
		}
		
		if len(records) != 2 {
			t.Fatalf("Expected 2 records, got %d", len(records))
		}
		
		// Should have generated field names
		if GetOr(records[0], "field_0", "") != "Alice" {
			t.Errorf("Expected field_0 'Alice', got %v", records[0]["field_0"])
		}
		if GetOr(records[0], "field_1", int64(0)) != int64(30) {
			t.Errorf("Expected field_1 30, got %v", records[0]["field_1"])
		}
	})
}

// TestFastTSVConvenienceFunctions tests the convenience functions
func TestFastTSVConvenienceFunctions(t *testing.T) {
	t.Run("FastTSVToStream", func(t *testing.T) {
		tsvData := "name\tage\nAlice\t30\nBob\t25"
		reader := strings.NewReader(tsvData)
		
		stream := FastTSVToStream(reader)
		records, err := Collect(stream)
		if err != nil {
			t.Fatalf("FastTSVToStream failed: %v", err)
		}
		
		if len(records) != 2 {
			t.Fatalf("Expected 2 records, got %d", len(records))
		}
		
		if GetOr(records[0], "name", "") != "Alice" {
			t.Errorf("Expected name 'Alice', got %v", records[0]["name"])
		}
	})
	
	t.Run("FastTSVToStreamWithSeparator", func(t *testing.T) {
		data := "name;age;city\nAlice;30;NYC\nBob;25;LA"
		reader := strings.NewReader(data)
		
		stream := FastTSVToStreamWithSeparator(reader, ";")
		records, err := Collect(stream)
		if err != nil {
			t.Fatalf("FastTSVToStreamWithSeparator failed: %v", err)
		}
		
		if len(records) != 2 {
			t.Fatalf("Expected 2 records, got %d", len(records))
		}
		
		if GetOr(records[0], "city", "") != "NYC" {
			t.Errorf("Expected city 'NYC', got %v", records[0]["city"])
		}
	})
}

// TestNewCSVSink tests CSV sink creation
func TestNewCSVSink(t *testing.T) {
	t.Run("BasicCSVSink", func(t *testing.T) {
		var buffer bytes.Buffer
		sink := NewCSVSink(&buffer)
		
		if sink == nil {
			t.Fatal("Expected non-nil CSV sink")
		}
		
		if sink.Separator != ',' {
			t.Errorf("Expected comma separator, got %c", sink.Separator)
		}
	})
}

// TestCSVSinkWithHeaders tests CSV sink with custom headers
func TestCSVSinkWithHeaders(t *testing.T) {
	t.Run("CustomHeaders", func(t *testing.T) {
		var buffer bytes.Buffer
		headers := []string{"name", "age", "city"}
		sink := NewCSVSink(&buffer).WithHeaders(headers)
		
		if len(sink.Headers) != 3 {
			t.Errorf("Expected 3 headers, got %d", len(sink.Headers))
		}
		
		if sink.Headers[0] != "name" {
			t.Errorf("Expected first header 'name', got '%s'", sink.Headers[0])
		}
	})
}

// TestStreamToCSV tests stream to CSV conversion
func TestStreamToCSV(t *testing.T) {
	t.Run("BasicConversion", func(t *testing.T) {
		records := []Record{
			{"name": "Alice", "age": int64(30), "active": true},
			{"name": "Bob", "age": int64(25), "active": false},
		}
		stream := FromSlice(records)
		
		var buffer bytes.Buffer
		err := StreamToCSV(stream, &buffer)
		if err != nil {
			t.Fatalf("Failed to write stream to CSV: %v", err)
		}
		
		csvOutput := buffer.String()
		if !strings.Contains(csvOutput, "Alice") {
			t.Error("Expected CSV output to contain 'Alice'")
		}
		
		if !strings.Contains(csvOutput, "30") {
			t.Error("Expected CSV output to contain '30'")
		}
	})
}

// TestNewTSVSink tests TSV sink creation
func TestNewTSVSink(t *testing.T) {
	t.Run("BasicTSVSink", func(t *testing.T) {
		var buffer bytes.Buffer
		sink := NewTSVSink(&buffer)
		
		if sink == nil {
			t.Fatal("Expected non-nil TSV sink")
		}
		
		if sink.Separator != '\t' {
			t.Errorf("Expected tab separator, got %c", sink.Separator)
		}
	})
}

// TestStreamToTSV tests stream to TSV conversion
func TestStreamToTSV(t *testing.T) {
	t.Run("BasicConversion", func(t *testing.T) {
		records := []Record{
			{"name": "Alice", "age": int64(30)},
			{"name": "Bob", "age": int64(25)},
		}
		stream := FromSlice(records)
		
		var buffer bytes.Buffer
		err := StreamToTSV(stream, &buffer)
		if err != nil {
			t.Fatalf("Failed to write stream to TSV: %v", err)
		}
		
		tsvOutput := buffer.String()
		if !strings.Contains(tsvOutput, "Alice") {
			t.Error("Expected TSV output to contain 'Alice'")
		}
		
		// Check for tab separation
		lines := strings.Split(strings.TrimSpace(tsvOutput), "\n")
		if len(lines) < 2 {
			t.Fatal("Expected at least 2 lines in TSV output")
		}
		
		if !strings.Contains(lines[1], "\t") {
			t.Error("Expected tab separation in TSV data line")
		}
	})
}

// TestNewStreamingCSVWriter tests streaming CSV writer
func TestNewStreamingCSVWriter(t *testing.T) {
	t.Run("StreamingWrite", func(t *testing.T) {
		var buffer bytes.Buffer
		headers := []string{"name", "age"}
		writer := NewStreamingCSVWriter(&buffer, headers)
		
		// Write first record
		record1 := Record{"name": "Alice", "age": int64(30)}
		err := writer.WriteRecord(record1)
		if err != nil {
			t.Fatalf("Failed to write first record: %v", err)
		}
		
		// Write second record
		record2 := Record{"name": "Bob", "age": int64(25)}
		err = writer.WriteRecord(record2)
		if err != nil {
			t.Fatalf("Failed to write second record: %v", err)
		}
		
		err = writer.Close()
		if err != nil {
			t.Fatalf("Failed to close writer: %v", err)
		}
		
		output := buffer.String()
		if !strings.Contains(output, "name,age") {
			t.Error("Expected header line in output")
		}
		
		if !strings.Contains(output, "Alice,30") {
			t.Error("Expected Alice's data in output")
		}
	})
}

// TestNewJSONSource tests JSON source creation
func TestNewJSONSource(t *testing.T) {
	t.Run("BasicJSONSource", func(t *testing.T) {
		jsonData := `{"name": "Alice", "age": 30}`
		reader := strings.NewReader(jsonData)
		
		source := NewJSONSource(reader)
		if source == nil {
			t.Fatal("Expected non-nil JSON source")
		}
		
		if source.Format != JSONLines {
			t.Error("Expected default format to be JSONLines")
		}
	})
}

// TestJSONSourceWithFormat tests JSON format configuration
func TestJSONSourceWithFormat(t *testing.T) {
	t.Run("JSONArrayFormat", func(t *testing.T) {
		jsonData := `[{"name": "Alice", "age": 30}, {"name": "Bob", "age": 25}]`
		reader := strings.NewReader(jsonData)
		
		source := NewJSONSource(reader).WithFormat(JSONArray)
		if source.Format != JSONArray {
			t.Error("Expected format to be JSONArray")
		}
	})
}

// TestJSONToStream tests JSON to stream conversion
func TestJSONToStream(t *testing.T) {
	t.Run("JSONLines", func(t *testing.T) {
		jsonData := `{"name": "Alice", "age": 30, "active": true}
{"name": "Bob", "age": 25, "active": false}`
		reader := strings.NewReader(jsonData)
		
		stream := JSONToStream(reader)
		records, err := Collect(stream)
		if err != nil {
			t.Fatalf("Failed to collect JSON stream: %v", err)
		}
		
		if len(records) != 2 {
			t.Fatalf("Expected 2 records, got %d", len(records))
		}
		
		if records[0]["name"] != "Alice" {
			t.Errorf("Expected name 'Alice', got %v", records[0]["name"])
		}
		
		if records[0]["age"] != int64(30) {
			t.Errorf("Expected age 30, got %v", records[0]["age"])
		}
	})
	
	t.Run("JSONArray", func(t *testing.T) {
		jsonData := `[{"name": "Alice", "age": 30}, {"name": "Bob", "age": 25}]`
		reader := strings.NewReader(jsonData)
		
		stream := NewJSONSource(reader).WithFormat(JSONArray).ToStream()
		records, err := Collect(stream)
		if err != nil {
			t.Fatalf("Failed to collect JSON array stream: %v", err)
		}
		
		if len(records) != 2 {
			t.Fatalf("Expected 2 records, got %d", len(records))
		}
		
		if records[0]["name"] != "Alice" {
			t.Errorf("Expected name 'Alice', got %v", records[0]["name"])
		}
	})
}

// TestNewJSONSink tests JSON sink creation
func TestNewJSONSink(t *testing.T) {
	t.Run("BasicJSONSink", func(t *testing.T) {
		var buffer bytes.Buffer
		sink := NewJSONSink(&buffer)
		
		if sink == nil {
			t.Fatal("Expected non-nil JSON sink")
		}
		
		if sink.Format != JSONLines {
			t.Error("Expected default format to be JSONLines")
		}
		
		if sink.Pretty {
			t.Error("Expected Pretty to be false by default")
		}
	})
}

// TestJSONSinkWithFormat tests JSON sink format configuration
func TestJSONSinkWithFormat(t *testing.T) {
	t.Run("JSONArrayFormat", func(t *testing.T) {
		var buffer bytes.Buffer
		sink := NewJSONSink(&buffer).WithFormat(JSONArray)
		
		if sink.Format != JSONArray {
			t.Error("Expected format to be JSONArray")
		}
	})
}

// TestJSONSinkWithPrettyPrint tests pretty printing configuration
func TestJSONSinkWithPrettyPrint(t *testing.T) {
	t.Run("PrettyPrint", func(t *testing.T) {
		var buffer bytes.Buffer
		sink := NewJSONSink(&buffer).WithPrettyPrint()
		
		if !sink.Pretty {
			t.Error("Expected Pretty to be true")
		}
	})
}

// TestStreamToJSON tests stream to JSON conversion
func TestStreamToJSON(t *testing.T) {
	t.Run("JSONLines", func(t *testing.T) {
		records := []Record{
			{"name": "Alice", "age": int64(30), "active": true},
			{"name": "Bob", "age": int64(25), "active": false},
		}
		stream := FromSlice(records)
		
		var buffer bytes.Buffer
		err := StreamToJSON(stream, &buffer)
		if err != nil {
			t.Fatalf("Failed to write stream to JSON: %v", err)
		}
		
		jsonOutput := buffer.String()
		lines := strings.Split(strings.TrimSpace(jsonOutput), "\n")
		
		if len(lines) != 2 {
			t.Fatalf("Expected 2 JSON lines, got %d", len(lines))
		}
		
		// Parse first line
		var record1 map[string]any
		err = json.Unmarshal([]byte(lines[0]), &record1)
		if err != nil {
			t.Fatalf("Failed to parse first JSON line: %v", err)
		}
		
		if record1["name"] != "Alice" {
			t.Errorf("Expected name 'Alice', got %v", record1["name"])
		}
	})
	
	t.Run("JSONArray", func(t *testing.T) {
		records := []Record{
			{"name": "Alice", "age": int64(30)},
			{"name": "Bob", "age": int64(25)},
		}
		stream := FromSlice(records)
		
		var buffer bytes.Buffer
		sink := NewJSONSink(&buffer).WithFormat(JSONArray)
		err := sink.WriteStream(stream)
		if err != nil {
			t.Fatalf("Failed to write stream to JSON array: %v", err)
		}
		
		jsonOutput := buffer.String()
		
		var jsonArray []map[string]any
		err = json.Unmarshal([]byte(jsonOutput), &jsonArray)
		if err != nil {
			t.Fatalf("Failed to parse JSON array: %v", err)
		}
		
		if len(jsonArray) != 2 {
			t.Fatalf("Expected 2 records in JSON array, got %d", len(jsonArray))
		}
		
		if jsonArray[0]["name"] != "Alice" {
			t.Errorf("Expected name 'Alice', got %v", jsonArray[0]["name"])
		}
	})
}

// TestIsStreamType tests stream type detection
func TestIsStreamType(t *testing.T) {
	t.Run("ValidStreamType", func(t *testing.T) {
		stream := FromSlice([]int64{1, 2, 3})
		
		if !IsStreamType(stream) {
			t.Error("Expected IsStreamType to return true for stream")
		}
	})
	
	t.Run("InvalidStreamType", func(t *testing.T) {
		notAStream := "not a stream"
		
		if IsStreamType(notAStream) {
			t.Error("Expected IsStreamType to return false for non-stream")
		}
	})
}

// TestParseCSVValue tests CSV value parsing
func TestParseCSVValue(t *testing.T) {
	tests := []struct {
		input    string
		expected any
	}{
		{"", ""},
		{"123", int64(123)},
		{"123.45", 123.45},
		{"true", true},
		{"false", false},
		{"yes", true},
		{"no", false},
		{"hello", "hello"},
		{"  spaced  ", "spaced"},
	}
	
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := parseCSVValue(test.input)
			if result != test.expected {
				t.Errorf("parseCSVValue(%q) = %v (%T), expected %v (%T)", 
					test.input, result, result, test.expected, test.expected)
			}
		})
	}
}

// TestFormatCSVValue tests CSV value formatting
func TestFormatCSVValue(t *testing.T) {
	tests := []struct {
		input    any
		expected string
	}{
		{"hello", "hello"},
		{int64(123), "123"},
		{123.45, "123.45"},
		{true, "true"},
		{false, "false"},
		{nil, ""},
		{time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC), "2023-01-01T12:00:00Z"},
	}
	
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			result := formatCSVValue(test.input)
			if result != test.expected {
				t.Errorf("formatCSVValue(%v) = %q, expected %q", test.input, result, test.expected)
			}
		})
	}
}

// TestFileOperations tests file-based operations (using temp files)
func TestFileOperations(t *testing.T) {
	t.Run("CSVFileOperations", func(t *testing.T) {
		// Create temporary file
		tmpfile, err := os.CreateTemp("", "test*.csv")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpfile.Name())
		defer tmpfile.Close()
		
		// Write test data
		csvData := "name,age\nAlice,30\nBob,25"
		if _, err := tmpfile.WriteString(csvData); err != nil {
			t.Fatalf("Failed to write test data: %v", err)
		}
		tmpfile.Close()
		
		// Test reading from file
		stream, err := CSVToStreamFromFile(tmpfile.Name())
		if err != nil {
			t.Fatalf("Failed to create stream from CSV file: %v", err)
		}
		
		records, err := Collect(stream)
		if err != nil {
			t.Fatalf("Failed to collect records from file stream: %v", err)
		}
		
		if len(records) != 2 {
			t.Fatalf("Expected 2 records from file, got %d", len(records))
		}
		
		if records[0]["name"] != "Alice" {
			t.Errorf("Expected name 'Alice' from file, got %v", records[0]["name"])
		}
	})
	
	t.Run("JSONFileOperations", func(t *testing.T) {
		// Create temporary file
		tmpfile, err := os.CreateTemp("", "test*.json")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpfile.Name())
		defer tmpfile.Close()
		
		// Write test data
		jsonData := `{"name": "Alice", "age": 30}
{"name": "Bob", "age": 25}`
		if _, err := tmpfile.WriteString(jsonData); err != nil {
			t.Fatalf("Failed to write test data: %v", err)
		}
		tmpfile.Close()
		
		// Test reading from file
		stream, err := JSONToStreamFromFile(tmpfile.Name())
		if err != nil {
			t.Fatalf("Failed to create stream from JSON file: %v", err)
		}
		
		records, err := Collect(stream)
		if err != nil {
			t.Fatalf("Failed to collect records from file stream: %v", err)
		}
		
		if len(records) != 2 {
			t.Fatalf("Expected 2 records from file, got %d", len(records))
		}
		
		if records[0]["name"] != "Alice" {
			t.Errorf("Expected name 'Alice' from file, got %v", records[0]["name"])
		}
	})
}