package telemetry

import (
	"context"
	"testing"
	"time"
)

func TestInitTracer_NoopExporter(t *testing.T) {
	ctx := context.Background()
	cfg := Config{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Exporter:       ExporterNoop,
	}

	shutdown, err := InitTracer(ctx, cfg)
	if err != nil {
		t.Fatalf("InitTracer with noop exporter failed: %v", err)
	}
	if shutdown == nil {
		t.Fatal("expected shutdown function, got nil")
	}

	// Test shutdown works
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := shutdown(ctx); err != nil {
		t.Errorf("shutdown failed: %v", err)
	}
}

func TestInitTracer_StdoutExporter(t *testing.T) {
	ctx := context.Background()
	cfg := Config{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Exporter:       ExporterStdout,
	}

	shutdown, err := InitTracer(ctx, cfg)
	if err != nil {
		t.Fatalf("InitTracer with stdout exporter failed: %v", err)
	}
	if shutdown == nil {
		t.Fatal("expected shutdown function, got nil")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := shutdown(ctx); err != nil {
		t.Errorf("shutdown failed: %v", err)
	}
}

func TestInitTracer_UnknownExporter(t *testing.T) {
	ctx := context.Background()
	cfg := Config{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Exporter:       "unknown-exporter",
	}

	shutdown, err := InitTracer(ctx, cfg)
	if err == nil {
		t.Fatal("expected error for unknown exporter, got nil")
	}
	if shutdown != nil {
		t.Fatal("expected nil shutdown function for error case")
	}
}

func TestNoopExporter(t *testing.T) {
	exporter := &noopExporter{}

	// Test ExportSpans does nothing and returns nil
	err := exporter.ExportSpans(context.Background(), nil)
	if err != nil {
		t.Errorf("ExportSpans should return nil, got: %v", err)
	}

	// Test Shutdown does nothing and returns nil
	err = exporter.Shutdown(context.Background())
	if err != nil {
		t.Errorf("Shutdown should return nil, got: %v", err)
	}
}

func TestExporterConstants(t *testing.T) {
	// Ensure constants have expected values
	if ExporterOTLP != "otlp" {
		t.Errorf("ExporterOTLP = %q, want %q", ExporterOTLP, "otlp")
	}
	if ExporterStdout != "stdout" {
		t.Errorf("ExporterStdout = %q, want %q", ExporterStdout, "stdout")
	}
	if ExporterNoop != "noop" {
		t.Errorf("ExporterNoop = %q, want %q", ExporterNoop, "noop")
	}
}
