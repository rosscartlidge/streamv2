package stream

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
	
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

// ============================================================================
// CSV/TSV SOURCES - EXTERNAL DATA INPUT
// ============================================================================

// CSVSource configuration for reading CSV data
type CSVSource struct {
	Reader    io.Reader
	HasHeader bool
	Separator rune
	Headers   []string
}

// NewCSVSource creates a CSV source from a reader
func NewCSVSource(reader io.Reader) *CSVSource {
	return &CSVSource{
		Reader:    reader,
		HasHeader: true,
		Separator: ',',
	}
}

// NewCSVSourceFromFile creates a CSV source from a file
func NewCSVSourceFromFile(filename string) (*CSVSource, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file %s: %w", filename, err)
	}
	
	return NewCSVSource(file), nil
}

// WithHeaders sets custom headers for the CSV
func (cs *CSVSource) WithHeaders(headers []string) *CSVSource {
	cs.Headers = headers
	cs.HasHeader = false
	return cs
}

// WithoutHeaders configures CSV to not expect headers
func (cs *CSVSource) WithoutHeaders() *CSVSource {
	cs.HasHeader = false
	return cs
}

// ToStream converts CSV data to a Record stream
func (cs *CSVSource) ToStream() Stream[Record] {
	reader := csv.NewReader(cs.Reader)
	reader.Comma = cs.Separator
	
	var headers []string
	var headerRead bool = false
	
	return func() (Record, error) {
		// Read headers on first call if needed
		if !headerRead {
			if cs.HasHeader {
				headerRow, err := reader.Read()
				if err != nil {
					return nil, err
				}
				headers = headerRow
			} else if len(cs.Headers) > 0 {
				headers = cs.Headers
			} else {
				// Generate default headers
				firstRow, err := reader.Read()
				if err != nil {
					return nil, err
				}
				
				headers = make([]string, len(firstRow))
				for i := range headers {
					headers[i] = fmt.Sprintf("col%d", i)
				}
				
				// Process first row as data
				record := make(Record)
				for i, value := range firstRow {
					if i < len(headers) {
						record[headers[i]] = parseCSVValue(value)
					}
				}
				headerRead = true
				return record, nil
			}
			headerRead = true
		}
		
		// Read data row
		row, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				return nil, EOS
			}
			return nil, err
		}
		
		// Convert to Record
		record := make(Record)
		for i, value := range row {
			if i < len(headers) {
				record[headers[i]] = parseCSVValue(value)
			} else {
				// Handle extra columns
				record[fmt.Sprintf("extra_col%d", i)] = parseCSVValue(value)
			}
		}
		
		return record, nil
	}
}

// TSVSource creates a TSV source (Tab-Separated Values)
func NewTSVSource(reader io.Reader) *CSVSource {
	return &CSVSource{
		Reader:    reader,
		HasHeader: true,
		Separator: '\t',
	}
}

// NewTSVSourceFromFile creates a TSV source from a file
func NewTSVSourceFromFile(filename string) (*CSVSource, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open TSV file %s: %w", filename, err)
	}
	
	return NewTSVSource(file), nil
}

// parseCSVValue attempts to parse CSV string values into appropriate types
func parseCSVValue(value string) any {
	value = strings.TrimSpace(value)
	
	// Empty string
	if value == "" {
		return ""
	}
	
	// Boolean values
	lowerValue := strings.ToLower(value)
	if lowerValue == "true" || lowerValue == "t" || lowerValue == "yes" || lowerValue == "y" {
		return true
	}
	if lowerValue == "false" || lowerValue == "f" || lowerValue == "no" || lowerValue == "n" {
		return false
	}
	
	// Integer values
	if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
		return intValue
	}
	
	// Float values
	if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
		return floatValue
	}
	
	// Time values (common formats)
	timeFormats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02",
		"15:04:05",
	}
	
	for _, format := range timeFormats {
		if timeValue, err := time.Parse(format, value); err == nil {
			return timeValue
		}
	}
	
	// Default to string
	return value
}

// ============================================================================
// CSV/TSV SINKS - DATA OUTPUT
// ============================================================================

// CSVSink configuration for writing CSV data
type CSVSink struct {
	Writer    io.Writer
	Separator rune
	Headers   []string
	headerWritten bool
}

// NewCSVSink creates a CSV sink to a writer
func NewCSVSink(writer io.Writer) *CSVSink {
	return &CSVSink{
		Writer:    writer,
		Separator: ',',
		headerWritten: false,
	}
}

// NewCSVSinkToFile creates a CSV sink to a file
func NewCSVSinkToFile(filename string) (*CSVSink, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create CSV file %s: %w", filename, err)
	}
	
	return NewCSVSink(file), nil
}

// WithHeaders sets the headers for CSV output
func (sink *CSVSink) WithHeaders(headers []string) *CSVSink {
	sink.Headers = headers
	return sink
}

// WriteStream writes a Record stream to CSV format
func (sink *CSVSink) WriteStream(stream Stream[Record]) error {
	writer := csv.NewWriter(sink.Writer)
	writer.Comma = sink.Separator
	defer writer.Flush()
	
	var headers []string
	
	for {
		record, err := stream()
		if err != nil {
			if err == EOS {
				break
			}
			return err
		}
		
		// Write headers on first record
		if !sink.headerWritten {
			if len(sink.Headers) > 0 {
				headers = sink.Headers
			} else {
				// Extract headers from first record
				headers = make([]string, 0, len(record))
				for key := range record {
					headers = append(headers, key)
				}
			}
			
			if err := writer.Write(headers); err != nil {
				return fmt.Errorf("failed to write CSV headers: %w", err)
			}
			sink.headerWritten = true
		}
		
		// Write record data
		row := make([]string, len(headers))
		for i, header := range headers {
			if value, exists := record[header]; exists {
				row[i] = formatCSVValue(value)
			} else {
				row[i] = ""
			}
		}
		
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}
	
	return nil
}

// WriteRecords writes a slice of Records to CSV format
func (sink *CSVSink) WriteRecords(records []Record) error {
	return sink.WriteStream(FromSlice(records))
}

// NewTSVSink creates a TSV sink to a writer
func NewTSVSink(writer io.Writer) *CSVSink {
	return &CSVSink{
		Writer:    writer,
		Separator: '\t',
		headerWritten: false,
	}
}

// NewTSVSinkToFile creates a TSV sink to a file
func NewTSVSinkToFile(filename string) (*CSVSink, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create TSV file %s: %w", filename, err)
	}
	
	return NewTSVSink(file), nil
}

// formatCSVValue converts any value to a string for CSV output
func formatCSVValue(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%g", v)
	case bool:
		if v {
			return "true"
		}
		return "false"
	case time.Time:
		return v.Format(time.RFC3339)
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}

// ============================================================================
// STREAMING FILE I/O UTILITIES
// ============================================================================

// StreamingCSVWriter allows writing Records to CSV as they arrive (for infinite streams)
type StreamingCSVWriter struct {
	writer   *csv.Writer
	headers  []string
	headerWritten bool
}

// NewStreamingCSVWriter creates a streaming CSV writer
func NewStreamingCSVWriter(w io.Writer, headers []string) *StreamingCSVWriter {
	writer := csv.NewWriter(w)
	return &StreamingCSVWriter{
		writer:  writer,
		headers: headers,
		headerWritten: false,
	}
}

// WriteRecord writes a single record to the CSV stream
func (scw *StreamingCSVWriter) WriteRecord(record Record) error {
	// Write headers on first record
	if !scw.headerWritten {
		if err := scw.writer.Write(scw.headers); err != nil {
			return fmt.Errorf("failed to write CSV headers: %w", err)
		}
		scw.headerWritten = true
	}
	
	// Write record data
	row := make([]string, len(scw.headers))
	for i, header := range scw.headers {
		if value, exists := record[header]; exists {
			row[i] = formatCSVValue(value)
		}
	}
	
	if err := scw.writer.Write(row); err != nil {
		return fmt.Errorf("failed to write CSV row: %w", err)
	}
	
	// Flush immediately for streaming
	scw.writer.Flush()
	return scw.writer.Error()
}

// Close flushes and closes the writer
func (scw *StreamingCSVWriter) Close() error {
	scw.writer.Flush()
	return scw.writer.Error()
}

// ============================================================================
// JSON SOURCES AND SINKS - STRUCTURED DATA SUPPORT
// ============================================================================

// JSONSource configuration for reading JSON data
type JSONSource struct {
	Reader io.Reader
	Format JSONFormat
}

// JSONFormat specifies how JSON data is structured
type JSONFormat int

const (
	JSONLines JSONFormat = iota // Each line is a separate JSON object
	JSONArray                   // Single JSON array containing objects
)

// NewJSONSource creates a JSON source from a reader (defaults to JSON Lines)
func NewJSONSource(reader io.Reader) *JSONSource {
	return &JSONSource{
		Reader: reader,
		Format: JSONLines,
	}
}

// WithFormat sets the JSON format
func (js *JSONSource) WithFormat(format JSONFormat) *JSONSource {
	js.Format = format
	return js
}

// ToStream converts JSON data to a Record stream
func (js *JSONSource) ToStream() Stream[Record] {
	switch js.Format {
	case JSONArray:
		return js.arrayToStream()
	default: // JSONLines
		return js.linesToStream()
	}
}

// linesToStream handles JSON Lines format (one JSON object per line)
func (js *JSONSource) linesToStream() Stream[Record] {
	scanner := bufio.NewScanner(js.Reader)
	
	return func() (Record, error) {
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return nil, err
			}
			return nil, EOS
		}
		
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" {
			// Skip empty lines, try next
			return js.linesToStream()()
		}
		
		var jsonObj map[string]any
		if err := json.Unmarshal([]byte(line), &jsonObj); err != nil {
			return nil, fmt.Errorf("failed to parse JSON line: %w", err)
		}
		
		return convertJSONToRecord(jsonObj), nil
	}
}

// arrayToStream handles JSON Array format (single array of objects)
func (js *JSONSource) arrayToStream() Stream[Record] {
	// Read entire input
	data, err := io.ReadAll(js.Reader)
	if err != nil {
		return func() (Record, error) {
			return nil, fmt.Errorf("failed to read JSON array: %w", err)
		}
	}
	
	var jsonArray []map[string]any
	if err := json.Unmarshal(data, &jsonArray); err != nil {
		return func() (Record, error) {
			return nil, fmt.Errorf("failed to parse JSON array: %w", err)
		}
	}
	
	// Convert to Record slice and use FromSlice
	records := make([]Record, len(jsonArray))
	for i, obj := range jsonArray {
		records[i] = convertJSONToRecord(obj)
	}
	
	return FromSlice(records)
}

// convertJSONToRecord converts a JSON object to a Record, preserving structure
func convertJSONToRecord(jsonObj map[string]any) Record {
	record := make(Record)
	for key, value := range jsonObj {
		record[key] = convertJSONValue(value)
	}
	return record
}

// convertJSONValue converts JSON values to appropriate Record field types
func convertJSONValue(value any) any {
	switch v := value.(type) {
	case nil:
		return nil
	case bool:
		return v
	case float64:
		// JSON numbers are always float64, convert to int64 if it's a whole number
		if v == float64(int64(v)) {
			return int64(v)
		}
		return v
	case string:
		return v
	case []any:
		// Convert array to Stream[any] for nested processing capability
		return FromSliceAny(v)
	case map[string]any:
		// Convert nested object to nested Record
		return convertJSONToRecord(v)
	default:
		// Fallback to string representation
		return fmt.Sprintf("%v", v)
	}
}

// JSONSink configuration for writing JSON data
type JSONSink struct {
	Writer io.Writer
	Format JSONFormat
	Pretty bool
}

// NewJSONSink creates a JSON sink to a writer (defaults to JSON Lines)
func NewJSONSink(writer io.Writer) *JSONSink {
	return &JSONSink{
		Writer: writer,
		Format: JSONLines,
		Pretty: false,
	}
}

// WithFormat sets the JSON format
func (sink *JSONSink) WithFormat(format JSONFormat) *JSONSink {
	sink.Format = format
	return sink
}

// WithPrettyPrint enables pretty printing
func (sink *JSONSink) WithPrettyPrint() *JSONSink {
	sink.Pretty = true
	return sink
}

// WriteStream writes a Record stream to JSON format
func (sink *JSONSink) WriteStream(stream Stream[Record]) error {
	switch sink.Format {
	case JSONArray:
		return sink.writeAsArray(stream)
	default: // JSONLines
		return sink.writeAsLines(stream)
	}
}

// writeAsLines writes each record as a separate JSON line
func (sink *JSONSink) writeAsLines(stream Stream[Record]) error {
	encoder := json.NewEncoder(sink.Writer)
	if !sink.Pretty {
		encoder.SetIndent("", "")
	}
	
	for {
		record, err := stream()
		if err != nil {
			if err == EOS {
				break
			}
			return err
		}
		
		jsonObj := convertRecordToJSON(record)
		if err := encoder.Encode(jsonObj); err != nil {
			return fmt.Errorf("failed to write JSON line: %w", err)
		}
	}
	
	return nil
}

// writeAsArray writes all records as a single JSON array
func (sink *JSONSink) writeAsArray(stream Stream[Record]) error {
	// Collect all records first
	var jsonArray []map[string]any
	
	for {
		record, err := stream()
		if err != nil {
			if err == EOS {
				break
			}
			return err
		}
		
		jsonObj := convertRecordToJSON(record)
		jsonArray = append(jsonArray, jsonObj)
	}
	
	// Write as single JSON array
	var data []byte
	var err error
	
	if sink.Pretty {
		data, err = json.MarshalIndent(jsonArray, "", "  ")
	} else {
		data, err = json.Marshal(jsonArray)
	}
	
	if err != nil {
		return fmt.Errorf("failed to marshal JSON array: %w", err)
	}
	
	_, err = sink.Writer.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write JSON array: %w", err)
	}
	
	return nil
}

// WriteRecords writes a slice of Records to JSON format
func (sink *JSONSink) WriteRecords(records []Record) error {
	return sink.WriteStream(FromSlice(records))
}

// convertRecordToJSON converts a Record to a JSON-serializable map
func convertRecordToJSON(record Record) map[string]any {
	jsonObj := make(map[string]any)
	for key, value := range record {
		jsonObj[key] = convertRecordValueToJSON(value)
	}
	return jsonObj
}

// convertRecordValueToJSON converts Record field values to JSON-serializable types
func convertRecordValueToJSON(value any) any {
	switch v := value.(type) {
	case nil:
		return nil
	case bool, int64, float64, string:
		return v
	case int:
		return int64(v)
	case float32:
		return float64(v)
	case time.Time:
		return v.Format(time.RFC3339)
	case Record:
		// Nested Record
		return convertRecordToJSON(v)
	case Stream[any]:
		// Convert stream to array for JSON serialization
		items, _ := Collect(v)
		jsonArray := make([]any, len(items))
		for i, item := range items {
			jsonArray[i] = convertRecordValueToJSON(item)
		}
		return jsonArray
	default:
		// Handle all other Stream types by attempting to collect them
		if IsStreamType(value) {
			// Try to collect the stream and convert to JSON array
			if collected := collectAnyStream(value); collected != nil {
				jsonArray := make([]any, len(collected))
				for i, item := range collected {
					jsonArray[i] = convertRecordValueToJSON(item)
				}
				return jsonArray
			}
			// If collection fails, return placeholder
			return fmt.Sprintf("<Stream[%T]>", value)
		}
		// Fallback to string representation
		return fmt.Sprintf("%v", value)
	}
}

// IsStreamType checks if a value is a Stream[T] type
func IsStreamType(value any) bool {
	// This is a simplified check - in practice you might want more sophisticated type checking
	return strings.Contains(fmt.Sprintf("%T", value), "func()")
}

// collectAnyStream attempts to collect items from any stream type using reflection
func collectAnyStream(value any) []any {
	// Use reflection to call the stream function repeatedly
	streamFunc := reflect.ValueOf(value)
	if streamFunc.Kind() != reflect.Func {
		return nil
	}
	
	// Check if it matches Stream[T] signature: func() (T, error)
	streamType := streamFunc.Type()
	if streamType.NumIn() != 0 || streamType.NumOut() != 2 {
		return nil
	}
	
	// Second return must be error
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	if !streamType.Out(1).Implements(errorType) {
		return nil
	}
	
	var collected []any
	for {
		// Call the stream function
		results := streamFunc.Call(nil)
		if len(results) != 2 {
			break
		}
		
		// Check for error
		errorVal := results[1]
		if !errorVal.IsNil() {
			// Check if it's EOS error
			if err, ok := errorVal.Interface().(error); ok && err == EOS {
				break
			}
			// Other errors, stop collection
			break
		}
		
		// Add the value to collection
		collected = append(collected, results[0].Interface())
	}
	
	return collected
}

// File-based JSON functions
func NewJSONSourceFromFile(filename string) (*JSONSource, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open JSON file %s: %w", filename, err)
	}
	
	return NewJSONSource(file), nil
}

func NewJSONSinkToFile(filename string) (*JSONSink, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create JSON file %s: %w", filename, err)
	}
	
	return NewJSONSink(file), nil
}

// ============================================================================
// PROTOCOL BUFFER SOURCES AND SINKS - HIGH-PERFORMANCE BINARY DATA
// ============================================================================

// ProtobufSource configuration for reading Protocol Buffer data
type ProtobufSource struct {
	Reader      io.Reader
	MessageDesc protoreflect.MessageDescriptor
	Format      ProtobufFormat
}

// ProtobufFormat specifies how protobuf data is structured
type ProtobufFormat int

const (
	ProtobufDelimited ProtobufFormat = iota // Length-delimited messages (streaming)
	ProtobufJSON                            // JSON representation of protobuf
)

// NewProtobufSource creates a protobuf source from a reader with a message descriptor
func NewProtobufSource(reader io.Reader, messageDesc protoreflect.MessageDescriptor) *ProtobufSource {
	return &ProtobufSource{
		Reader:      reader,
		MessageDesc: messageDesc,
		Format:      ProtobufDelimited,
	}
}

// WithFormat sets the protobuf format
func (ps *ProtobufSource) WithFormat(format ProtobufFormat) *ProtobufSource {
	ps.Format = format
	return ps
}

// ToStream converts protobuf data to a Record stream
func (ps *ProtobufSource) ToStream() Stream[Record] {
	switch ps.Format {
	case ProtobufJSON:
		return ps.jsonToStream()
	default: // ProtobufDelimited
		return ps.delimitedToStream()
	}
}

// delimitedToStream handles length-delimited protobuf messages
func (ps *ProtobufSource) delimitedToStream() Stream[Record] {
	return func() (Record, error) {
		// Read length prefix (varint)
		length, err := readVarint(ps.Reader)
		if err != nil {
			if err == io.EOF {
				return nil, EOS
			}
			return nil, fmt.Errorf("failed to read message length: %w", err)
		}
		
		// Read message data
		msgData := make([]byte, length)
		if _, err := io.ReadFull(ps.Reader, msgData); err != nil {
			return nil, fmt.Errorf("failed to read message data: %w", err)
		}
		
		// Parse protobuf message
		msg := dynamicpb.NewMessage(ps.MessageDesc)
		if err := proto.Unmarshal(msgData, msg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal protobuf message: %w", err)
		}
		
		return convertProtobufToRecord(msg), nil
	}
}

// jsonToStream handles JSON representation of protobuf messages
func (ps *ProtobufSource) jsonToStream() Stream[Record] {
	scanner := bufio.NewScanner(ps.Reader)
	
	return func() (Record, error) {
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return nil, err
			}
			return nil, EOS
		}
		
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			// Skip empty lines, try next
			return ps.jsonToStream()()
		}
		
		// Parse JSON into protobuf message
		msg := dynamicpb.NewMessage(ps.MessageDesc)
		if err := protojson.Unmarshal([]byte(line), msg); err != nil {
			return nil, fmt.Errorf("failed to parse protobuf JSON: %w", err)
		}
		
		return convertProtobufToRecord(msg), nil
	}
}

// convertProtobufToRecord converts a protobuf message to a Record
func convertProtobufToRecord(msg protoreflect.ProtoMessage) Record {
	record := make(Record)
	msgReflect := msg.ProtoReflect()
	
	msgReflect.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		fieldName := string(fd.Name())
		record[fieldName] = convertProtobufValue(fd, v)
		return true
	})
	
	return record
}

// convertProtobufValue converts protobuf field values to Record field types
func convertProtobufValue(fd protoreflect.FieldDescriptor, v protoreflect.Value) any {
	if fd.IsList() {
		// Handle repeated fields as streams
		list := v.List()
		items := make([]any, list.Len())
		for i := 0; i < list.Len(); i++ {
			items[i] = convertProtobufScalarValue(fd, list.Get(i))
		}
		return FromSliceAny(items)
	}
	
	if fd.IsMap() {
		// Handle map fields as nested Records
		mapVal := v.Map()
		mapRecord := make(Record)
		mapVal.Range(func(k protoreflect.MapKey, v protoreflect.Value) bool {
			keyStr := fmt.Sprintf("%v", k.Interface())
			mapRecord[keyStr] = convertProtobufScalarValue(fd.MapValue(), v)
			return true
		})
		return mapRecord
	}
	
	return convertProtobufScalarValue(fd, v)
}

// convertProtobufScalarValue converts scalar protobuf values
func convertProtobufScalarValue(fd protoreflect.FieldDescriptor, v protoreflect.Value) any {
	switch fd.Kind() {
	case protoreflect.BoolKind:
		return v.Bool()
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return int64(v.Int())
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return v.Int()
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return int64(v.Uint())
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return int64(v.Uint())
	case protoreflect.FloatKind:
		return float64(v.Float())
	case protoreflect.DoubleKind:
		return v.Float()
	case protoreflect.StringKind:
		return v.String()
	case protoreflect.BytesKind:
		return string(v.Bytes())
	case protoreflect.MessageKind:
		// Nested message
		return convertProtobufToRecord(v.Message().Interface())
	case protoreflect.EnumKind:
		return string(fd.Enum().Values().ByNumber(v.Enum()).Name())
	default:
		return fmt.Sprintf("%v", v.Interface())
	}
}

// ProtobufSink configuration for writing protobuf data
type ProtobufSink struct {
	Writer      io.Writer
	MessageDesc protoreflect.MessageDescriptor
	Format      ProtobufFormat
}

// NewProtobufSink creates a protobuf sink to a writer
func NewProtobufSink(writer io.Writer, messageDesc protoreflect.MessageDescriptor) *ProtobufSink {
	return &ProtobufSink{
		Writer:      writer,
		MessageDesc: messageDesc,
		Format:      ProtobufDelimited,
	}
}

// WithFormat sets the protobuf format
func (sink *ProtobufSink) WithFormat(format ProtobufFormat) *ProtobufSink {
	sink.Format = format
	return sink
}

// WriteStream writes a Record stream to protobuf format
func (sink *ProtobufSink) WriteStream(stream Stream[Record]) error {
	switch sink.Format {
	case ProtobufJSON:
		return sink.writeAsJSON(stream)
	default: // ProtobufDelimited
		return sink.writeAsDelimited(stream)
	}
}

// writeAsDelimited writes length-delimited protobuf messages
func (sink *ProtobufSink) writeAsDelimited(stream Stream[Record]) error {
	for {
		record, err := stream()
		if err != nil {
			if err == EOS {
				break
			}
			return err
		}
		
		// Convert Record to protobuf message
		msg := dynamicpb.NewMessage(sink.MessageDesc)
		if err := convertRecordToProtobuf(record, msg); err != nil {
			return fmt.Errorf("failed to convert record to protobuf: %w", err)
		}
		
		// Marshal message
		data, err := proto.Marshal(msg)
		if err != nil {
			return fmt.Errorf("failed to marshal protobuf message: %w", err)
		}
		
		// Write length prefix + message
		if err := writeVarint(sink.Writer, uint64(len(data))); err != nil {
			return fmt.Errorf("failed to write message length: %w", err)
		}
		
		if _, err := sink.Writer.Write(data); err != nil {
			return fmt.Errorf("failed to write message data: %w", err)
		}
	}
	
	return nil
}

// writeAsJSON writes protobuf messages as JSON lines
func (sink *ProtobufSink) writeAsJSON(stream Stream[Record]) error {
	for {
		record, err := stream()
		if err != nil {
			if err == EOS {
				break
			}
			return err
		}
		
		// Convert Record to protobuf message
		msg := dynamicpb.NewMessage(sink.MessageDesc)
		if err := convertRecordToProtobuf(record, msg); err != nil {
			return fmt.Errorf("failed to convert record to protobuf: %w", err)
		}
		
		// Marshal as JSON
		jsonData, err := protojson.Marshal(msg)
		if err != nil {
			return fmt.Errorf("failed to marshal protobuf as JSON: %w", err)
		}
		
		// Write JSON line
		if _, err := sink.Writer.Write(jsonData); err != nil {
			return fmt.Errorf("failed to write JSON data: %w", err)
		}
		if _, err := sink.Writer.Write([]byte("\n")); err != nil {
			return fmt.Errorf("failed to write newline: %w", err)
		}
	}
	
	return nil
}

// WriteRecords writes a slice of Records to protobuf format
func (sink *ProtobufSink) WriteRecords(records []Record) error {
	return sink.WriteStream(FromSlice(records))
}

// convertRecordToProtobuf converts a Record to a protobuf message
func convertRecordToProtobuf(record Record, msg *dynamicpb.Message) error {
	msgReflect := msg.ProtoReflect()
	msgDesc := msgReflect.Descriptor()
	
	for fieldName, value := range record {
		// Find field descriptor by name
		fd := msgDesc.Fields().ByName(protoreflect.Name(fieldName))
		if fd == nil {
			// Skip unknown fields
			continue
		}
		
		protobufValue, err := convertRecordValueToProtobuf(msgReflect, fd, value)
		if err != nil {
			return fmt.Errorf("failed to convert field '%s': %w", fieldName, err)
		}
		
		msgReflect.Set(fd, protobufValue)
	}
	
	return nil
}

// convertRecordValueToProtobuf converts Record field values to protobuf values
func convertRecordValueToProtobuf(msgReflect protoreflect.Message, fd protoreflect.FieldDescriptor, value any) (protoreflect.Value, error) {
	if fd.IsList() {
		// Handle repeated fields
		list := msgReflect.Mutable(fd).List()
		
		// Handle Stream[T] values
		if streamValue, ok := value.(Stream[any]); ok {
			items, _ := Collect(streamValue)
			for _, item := range items {
				itemValue, err := convertRecordScalarToProtobuf(fd, item)
				if err != nil {
					return protoreflect.Value{}, err
				}
				list.Append(itemValue)
			}
		} else if slice, ok := value.([]any); ok {
			for _, item := range slice {
				itemValue, err := convertRecordScalarToProtobuf(fd, item)
				if err != nil {
					return protoreflect.Value{}, err
				}
				list.Append(itemValue)
			}
		}
		
		return protoreflect.ValueOfList(list), nil
	}
	
	if fd.IsMap() {
		// Handle map fields
		mapVal := msgReflect.Mutable(fd).Map()
		
		if recordValue, ok := value.(Record); ok {
			for k, v := range recordValue {
				key := protoreflect.ValueOfString(k).MapKey()
				val, err := convertRecordScalarToProtobuf(fd.MapValue(), v)
				if err != nil {
					return protoreflect.Value{}, err
				}
				mapVal.Set(key, val)
			}
		}
		
		return protoreflect.ValueOfMap(mapVal), nil
	}
	
	return convertRecordScalarToProtobuf(fd, value)
}

// convertRecordScalarToProtobuf converts scalar Record values to protobuf values
func convertRecordScalarToProtobuf(fd protoreflect.FieldDescriptor, value any) (protoreflect.Value, error) {
	switch fd.Kind() {
	case protoreflect.BoolKind:
		if b, ok := value.(bool); ok {
			return protoreflect.ValueOfBool(b), nil
		}
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		if i, ok := convertToInt64(value); ok {
			return protoreflect.ValueOfInt32(int32(i)), nil
		}
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		if i, ok := convertToInt64(value); ok {
			return protoreflect.ValueOfInt64(i), nil
		}
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		if i, ok := convertToInt64(value); ok {
			return protoreflect.ValueOfUint32(uint32(i)), nil
		}
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		if i, ok := convertToInt64(value); ok {
			return protoreflect.ValueOfUint64(uint64(i)), nil
		}
	case protoreflect.FloatKind:
		if f, ok := convertToFloat64(value); ok {
			return protoreflect.ValueOfFloat32(float32(f)), nil
		}
	case protoreflect.DoubleKind:
		if f, ok := convertToFloat64(value); ok {
			return protoreflect.ValueOfFloat64(f), nil
		}
	case protoreflect.StringKind:
		if s, ok := convertToString(value); ok {
			return protoreflect.ValueOfString(s), nil
		}
	case protoreflect.BytesKind:
		if s, ok := convertToString(value); ok {
			return protoreflect.ValueOfBytes([]byte(s)), nil
		}
	case protoreflect.MessageKind:
		if record, ok := value.(Record); ok {
			// Create nested message
			nestedMsg := dynamicpb.NewMessage(fd.Message())
			if err := convertRecordToProtobuf(record, nestedMsg); err != nil {
				return protoreflect.Value{}, err
			}
			return protoreflect.ValueOfMessage(nestedMsg.ProtoReflect()), nil
		}
	}
	
	return protoreflect.Value{}, fmt.Errorf("cannot convert %T to protobuf %s", value, fd.Kind())
}

// Utility functions for varint encoding/decoding
func readVarint(r io.Reader) (uint64, error) {
	var result uint64
	var shift uint
	
	for {
		var b [1]byte
		if _, err := r.Read(b[:]); err != nil {
			return 0, err
		}
		
		result |= uint64(b[0]&0x7F) << shift
		if b[0]&0x80 == 0 {
			break
		}
		shift += 7
		if shift >= 64 {
			return 0, fmt.Errorf("varint too long")
		}
	}
	
	return result, nil
}

func writeVarint(w io.Writer, value uint64) error {
	for value >= 0x80 {
		if _, err := w.Write([]byte{byte(value) | 0x80}); err != nil {
			return err
		}
		value >>= 7
	}
	_, err := w.Write([]byte{byte(value)})
	return err
}

// File-based protobuf functions
func NewProtobufSourceFromFile(filename string, messageDesc protoreflect.MessageDescriptor) (*ProtobufSource, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open protobuf file %s: %w", filename, err)
	}
	
	return NewProtobufSource(file, messageDesc), nil
}

func NewProtobufSinkToFile(filename string, messageDesc protoreflect.MessageDescriptor) (*ProtobufSink, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create protobuf file %s: %w", filename, err)
	}
	
	return NewProtobufSink(file, messageDesc), nil
}

// ============================================================================
// CONVENIENCE FUNCTIONS
// ============================================================================

// CSVToStream reads CSV from a reader and returns a Record stream
func CSVToStream(reader io.Reader) Stream[Record] {
	source := NewCSVSource(reader)
	return source.ToStream()
}

// TSVToStream reads TSV from a reader and returns a Record stream
func TSVToStream(reader io.Reader) Stream[Record] {
	source := NewTSVSource(reader)
	return source.ToStream()
}

// StreamToCSV writes a Record stream to a CSV writer
func StreamToCSV(stream Stream[Record], writer io.Writer) error {
	sink := NewCSVSink(writer)
	return sink.WriteStream(stream)
}

// StreamToTSV writes a Record stream to a TSV writer
func StreamToTSV(stream Stream[Record], writer io.Writer) error {
	sink := NewTSVSink(writer)
	return sink.WriteStream(stream)
}

// JSONToStream reads JSON from a reader and returns a Record stream (defaults to JSON Lines)
func JSONToStream(reader io.Reader) Stream[Record] {
	source := NewJSONSource(reader)
	return source.ToStream()
}

// StreamToJSON writes a Record stream to a JSON writer (defaults to JSON Lines)
func StreamToJSON(stream Stream[Record], writer io.Writer) error {
	sink := NewJSONSink(writer)
	return sink.WriteStream(stream)
}

// File-based convenience functions for backward compatibility
func CSVToStreamFromFile(filename string) (Stream[Record], error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file %s: %w", filename, err)
	}
	return CSVToStream(file), nil
}

func TSVToStreamFromFile(filename string) (Stream[Record], error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open TSV file %s: %w", filename, err)
	}
	return TSVToStream(file), nil
}

func StreamToCSVFile(stream Stream[Record], filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create CSV file %s: %w", filename, err)
	}
	defer file.Close()
	return StreamToCSV(stream, file)
}

func StreamToTSVFile(stream Stream[Record], filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create TSV file %s: %w", filename, err)
	}
	defer file.Close()
	return StreamToTSV(stream, file)
}

func JSONToStreamFromFile(filename string) (Stream[Record], error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open JSON file %s: %w", filename, err)
	}
	return JSONToStream(file), nil
}

func StreamToJSONFile(stream Stream[Record], filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create JSON file %s: %w", filename, err)
	}
	defer file.Close()
	return StreamToJSON(stream, file)
}

// ProtobufToStream reads protobuf from a reader and returns a Record stream  
func ProtobufToStream(reader io.Reader, messageDesc protoreflect.MessageDescriptor) Stream[Record] {
	source := NewProtobufSource(reader, messageDesc)
	return source.ToStream()
}

// StreamToProtobuf writes a Record stream to a protobuf writer
func StreamToProtobuf(stream Stream[Record], writer io.Writer, messageDesc protoreflect.MessageDescriptor) error {
	sink := NewProtobufSink(writer, messageDesc)
	return sink.WriteStream(stream)
}

func ProtobufToStreamFromFile(filename string, messageDesc protoreflect.MessageDescriptor) (Stream[Record], error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open protobuf file %s: %w", filename, err)
	}
	return ProtobufToStream(file, messageDesc), nil
}

func StreamToProtobufFile(stream Stream[Record], filename string, messageDesc protoreflect.MessageDescriptor) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create protobuf file %s: %w", filename, err)
	}
	defer file.Close()
	return StreamToProtobuf(stream, file, messageDesc)
}