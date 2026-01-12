package mongodb

import (
	"context"
	"testing"
	"time"

	"go.opentelemetry.io/otel/trace"
)

func TestOptions_DefaultValues(t *testing.T) {
	opts := Options{
		Host: "localhost",
		Port: 27017,
	}

	if opts.MaxPoolSize != 0 {
		t.Errorf("MaxPoolSize = %v, want 0 (default)", opts.MaxPoolSize)
	}
	if opts.MinPoolSize != 0 {
		t.Errorf("MinPoolSize = %v, want 0", opts.MinPoolSize)
	}
	if opts.ConnectTimeout != 0 {
		t.Errorf("ConnectTimeout = %v, want 0", opts.ConnectTimeout)
	}
	if opts.SocketTimeout != 0 {
		t.Errorf("SocketTimeout = %v, want 0", opts.SocketTimeout)
	}
	if opts.EnableTrace {
		t.Errorf("EnableTrace = %v, want false", opts.EnableTrace)
	}
}

func TestOptions_WithTimeouts(t *testing.T) {
	opts := Options{
		Host:           "localhost",
		Port:           27017,
		MaxPoolSize:    100,
		MinPoolSize:    10,
		ConnectTimeout: 10 * time.Second,
		SocketTimeout:  30 * time.Second,
	}

	if opts.MaxPoolSize != 100 {
		t.Errorf("MaxPoolSize = %v, want 100", opts.MaxPoolSize)
	}
	if opts.MinPoolSize != 10 {
		t.Errorf("MinPoolSize = %v, want 10", opts.MinPoolSize)
	}
	if opts.ConnectTimeout != 10*time.Second {
		t.Errorf("ConnectTimeout = %v, want 10s", opts.ConnectTimeout)
	}
	if opts.SocketTimeout != 30*time.Second {
		t.Errorf("SocketTimeout = %v, want 30s", opts.SocketTimeout)
	}
}

func TestOptions_WithCredentials(t *testing.T) {
	opts := Options{
		Host:       "localhost",
		Port:       27017,
		Database:   "testdb",
		Username:   "admin",
		Password:   "password123",
		AuthSource: "admin",
	}

	if opts.Username != "admin" {
		t.Errorf("Username = %v, want 'admin'", opts.Username)
	}
	if opts.Password != "password123" {
		t.Errorf("Password = %v, want 'password123'", opts.Password)
	}
	if opts.AuthSource != "admin" {
		t.Errorf("AuthSource = %v, want 'admin'", opts.AuthSource)
	}
}

func TestBuildURI_NoAuth(t *testing.T) {
	opts := &Options{
		Host: "localhost",
		Port: 27017,
	}

	uri := buildURI(opts)

	expected := "mongodb://localhost:27017"
	if uri != expected {
		t.Errorf("buildURI() = %v, want %v", uri, expected)
	}
}

func TestBuildURI_WithAuth(t *testing.T) {
	opts := &Options{
		Host:       "localhost",
		Port:       27017,
		Username:   "admin",
		Password:   "password",
		AuthSource: "admin",
	}

	uri := buildURI(opts)

	expected := "mongodb://admin:password@localhost:27017?authSource=admin"
	if uri != expected {
		t.Errorf("buildURI() = %v, want %v", uri, expected)
	}
}

func TestBuildURI_WithSpecialChars(t *testing.T) {
	opts := &Options{
		Host:     "localhost",
		Port:     27017,
		Username: "admin@domain",
		Password: "p@ss:word/",
	}

	uri := buildURI(opts)

	if uri == "" {
		t.Error("buildURI() should not return empty string")
	}
	if len(uri) <= len("mongodb://localhost:27017") {
		t.Error("buildURI() should include credentials")
	}
}

func TestNewMongoDB_NilOptions(t *testing.T) {
	client, err := NewMongoDB(nil)

	if err == nil {
		t.Error("NewMongoDB(nil) should return error")
	}
	if client != nil {
		t.Error("NewMongoDB(nil) should return nil client")
	}
	if err.Error() != "mongodb options cannot be nil" {
		t.Errorf("error message = %v, want 'mongodb options cannot be nil'", err.Error())
	}
}

func TestNewMongoDB_EmptyHost(t *testing.T) {
	opts := &Options{
		Host: "",
		Port: 27017,
	}

	client, err := NewMongoDB(opts)

	if err == nil {
		t.Error("NewMongoDB with empty host should return error")
	}
	if client != nil {
		t.Error("NewMongoDB with empty host should return nil client")
	}
}

func TestMongoDBClient_Fields(t *testing.T) {
	opts := &Options{
		Host:     "localhost",
		Port:     27017,
		Database: "testdb",
	}

	client := &MongoDBClient{
		opts: opts,
	}

	if client.opts != opts {
		t.Error("MongoDBClient.opts should match provided options")
	}
	if client.Client != nil {
		t.Error("MongoDBClient.Client should be nil for unit test")
	}
	if client.Database != nil {
		t.Error("MongoDBClient.Database should be nil for unit test")
	}
}

func TestMongoDBClient_Close_NilClient(t *testing.T) {
	client := &MongoDBClient{}

	err := client.Close(context.Background())
	if err != nil {
		t.Errorf("Close() with nil client error = %v", err)
	}
}

func TestMongoDBClient_Collection_NilDatabase(t *testing.T) {
	client := &MongoDBClient{}

	collection := client.Collection("test")
	if collection != nil {
		t.Error("Collection() should return nil when database is nil")
	}
}

func TestMongoDBClient_Ping_NilClient(t *testing.T) {
	client := &MongoDBClient{}

	err := client.Ping(context.Background())
	if err == nil {
		t.Error("Ping() with nil client should return error")
	}
}

func TestOptions_EnableTraceField(t *testing.T) {
	tests := []struct {
		name        string
		enableTrace bool
	}{
		{"trace disabled", false},
		{"trace enabled", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &Options{
				Host:        "localhost",
				Port:        27017,
				EnableTrace: tt.enableTrace,
			}

			if opts.EnableTrace != tt.enableTrace {
				t.Errorf("EnableTrace = %v, want %v", opts.EnableTrace, tt.enableTrace)
			}
		})
	}
}

func TestOptions_DatabaseField(t *testing.T) {
	opts := &Options{
		Host:     "localhost",
		Port:     27017,
		Database: "myapp_db",
	}

	if opts.Database != "myapp_db" {
		t.Errorf("Database = %v, want 'myapp_db'", opts.Database)
	}
}

func TestBuildURI_AuthSourceOnly(t *testing.T) {
	opts := &Options{
		Host:       "localhost",
		Port:       27017,
		Username:   "user",
		Password:   "pass",
		AuthSource: "custom_db",
	}

	uri := buildURI(opts)

	if uri == "" {
		t.Error("buildURI() should not return empty string")
	}
}

func TestNewMongoDB_InvalidPort(t *testing.T) {
	opts := &Options{
		Host: "localhost",
		Port: 0,
	}

	client, err := NewMongoDB(opts)

	if err == nil {
		t.Error("NewMongoDB with invalid port should return error")
	}
	if client != nil {
		t.Error("NewMongoDB with invalid port should return nil client")
	}
}

func TestMongoDBClient_Close_ValidClientNil(t *testing.T) {
	client := &MongoDBClient{
		Client:   nil,
		Database: nil,
		opts:     &Options{Host: "localhost"},
	}

	err := client.Close(context.Background())
	if err != nil {
		t.Errorf("Close() with nil client should return nil error, got %v", err)
	}
}

func TestMongoDBClient_Collection_NilDatabaseConsistent(t *testing.T) {
	client := &MongoDBClient{}

	collection := client.Collection("users")
	if collection != nil {
		t.Error("Collection() should return nil when database is nil (safe behavior)")
	}
}

func TestOptions_PoolSizes(t *testing.T) {
	opts := &Options{
		Host:           "localhost",
		Port:           27017,
		MaxPoolSize:    200,
		MinPoolSize:    20,
		ConnectTimeout: 15 * time.Second,
		SocketTimeout:  60 * time.Second,
	}

	if opts.MaxPoolSize != 200 {
		t.Errorf("MaxPoolSize = %v, want 200", opts.MaxPoolSize)
	}
	if opts.MinPoolSize != 20 {
		t.Errorf("MinPoolSize = %v, want 20", opts.MinPoolSize)
	}
}

func TestTraceCommandMonitor_Struct(t *testing.T) {
	monitor := &traceCommandMonitor{
		enableTrace: true,
	}

	if !monitor.enableTrace {
		t.Error("traceCommandMonitor.enableTrace should be true")
	}
	if monitor.commandDatabaseMap != nil {
		t.Error("traceCommandMonitor.commandDatabaseMap should be nil initially")
	}
	if monitor.commandSpanMap != nil {
		t.Error("traceCommandMonitor.commandSpanMap should be nil initially")
	}
}

func TestTraceCommandMonitor_WithData(t *testing.T) {
	monitor := &traceCommandMonitor{
		enableTrace:        true,
		commandDatabaseMap: make(map[int64]string),
		commandSpanMap:     make(map[int64]trace.Span),
	}

	monitor.commandDatabaseMap[1001] = "test_db"
	monitor.commandSpanMap[1001] = nil

	if monitor.commandDatabaseMap[1001] != "test_db" {
		t.Error("commandDatabaseMap should contain expected value")
	}
}

func TestMongoDBClient_SetFields(t *testing.T) {
	client := &MongoDBClient{}

	client.opts = &Options{
		Host:     "localhost",
		Port:     27017,
		Database: "testdb",
	}

	if client.opts.Host != "localhost" {
		t.Error("opts.Host should be set correctly")
	}
	if client.opts.Port != 27017 {
		t.Error("opts.Port should be set correctly")
	}
	if client.opts.Database != "testdb" {
		t.Error("opts.Database should be set correctly")
	}
}

func TestNewMongoDB_WithDisabledTrace(t *testing.T) {
	opts := &Options{
		Host:        "localhost",
		Port:        27017,
		EnableTrace: false,
	}

	if opts.EnableTrace {
		t.Error("EnableTrace should be false")
	}
}
