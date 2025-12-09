package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestParse_ExampleFiles tests parsing all example D2 files
func TestParse_ExampleFiles(t *testing.T) {
	p := NewD2Parser()

	// Get the project root directory
	// Tests run from the package directory, so we need to go up
	examplesDir := "../../examples"

	files, err := filepath.Glob(filepath.Join(examplesDir, "*.d2"))
	if err != nil {
		t.Fatalf("Failed to glob examples: %v", err)
	}

	if len(files) == 0 {
		t.Skip("No example files found")
	}

	for _, file := range files {
		// Skip macOS metadata files
		if strings.HasPrefix(filepath.Base(file), "._") {
			continue
		}
		t.Run(filepath.Base(file), func(t *testing.T) {
			content, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("Failed to read file: %v", err)
			}

			diagram, err := p.Parse(string(content))
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", file, err)
			}

			// Basic sanity checks
			if diagram == nil {
				t.Error("Diagram is nil")
			}

			t.Logf("Parsed %s: %d nodes, %d edges",
				filepath.Base(file), len(diagram.Nodes), len(diagram.Edges))
		})
	}
}

// TestParse_TestDataFiles tests parsing all test fixture files
func TestParse_TestDataFiles(t *testing.T) {
	p := NewD2Parser()

	testdataDir := "../../testdata"

	files, err := filepath.Glob(filepath.Join(testdataDir, "*.d2"))
	if err != nil {
		t.Fatalf("Failed to glob testdata: %v", err)
	}

	if len(files) == 0 {
		t.Skip("No testdata files found")
	}

	for _, file := range files {
		// Skip macOS metadata files
		if strings.HasPrefix(filepath.Base(file), "._") {
			continue
		}
		t.Run(filepath.Base(file), func(t *testing.T) {
			content, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("Failed to read file: %v", err)
			}

			diagram, err := p.Parse(string(content))
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", file, err)
			}

			if diagram == nil {
				t.Error("Diagram is nil")
			}

			t.Logf("Parsed %s: %d nodes, %d edges",
				filepath.Base(file), len(diagram.Nodes), len(diagram.Edges))
		})
	}
}

// TestParse_BasicShapesExample verifies parsing of example 01
func TestParse_BasicShapesExample(t *testing.T) {
	p := NewD2Parser()
	source := `
# Example 1: Basic Shapes
server: Web Server
database: Database
cache: Cache
frontend: Frontend App
backend: Backend API

frontend -> backend
backend -> database
backend -> cache
`
	diagram, err := p.Parse(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(diagram.Nodes) != 5 {
		t.Errorf("Expected 5 nodes, got %d", len(diagram.Nodes))
	}
	if len(diagram.Edges) != 3 {
		t.Errorf("Expected 3 edges, got %d", len(diagram.Edges))
	}
}

// TestParse_MicroservicesExample verifies parsing of example 07
func TestParse_MicroservicesExample(t *testing.T) {
	p := NewD2Parser()
	source := `
# Microservices Architecture
web: Web Clients { shape: person }
mobile: Mobile Clients { shape: person }

gateway: API Gateway {
  style: {
    fill: "#6366f1"
    stroke: "#4f46e5"
    font-color: white
    bold: true
  }
}

services: Microservices {
  auth: Auth Service { shape: hexagon }
  users: User Service { shape: hexagon }
  orders: Order Service { shape: hexagon }
}

data: Data Layer {
  userdb: User DB { shape: cylinder }
  orderdb: Order DB { shape: cylinder }
  cache: Redis Cache { shape: circle }
}

web -> gateway: HTTPS
mobile -> gateway: HTTPS
gateway -> services.auth: Authenticate
gateway -> services.users: User API
services.users -> data.userdb
services.orders -> data.orderdb
services.auth -> data.cache: Session Cache
`
	diagram, err := p.Parse(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Count expected nodes:
	// web, mobile, gateway = 3 top-level
	// services container + auth, users, orders = 4
	// data container + userdb, orderdb, cache = 4
	// Total = 11
	if len(diagram.Nodes) < 10 {
		t.Errorf("Expected at least 10 nodes, got %d", len(diagram.Nodes))
	}

	// Count edges
	if len(diagram.Edges) != 7 {
		t.Errorf("Expected 7 edges, got %d", len(diagram.Edges))
	}

	// Verify gateway styling
	var gateway *struct {
		fill      string
		stroke    string
		fontColor string
		bold      bool
	}
	for _, node := range diagram.Nodes {
		if node.ID == "gateway" {
			gateway = &struct {
				fill      string
				stroke    string
				fontColor string
				bold      bool
			}{
				fill:      node.Style.Fill,
				stroke:    node.Style.Stroke,
				fontColor: node.Style.FontColor,
				bold:      node.Style.Bold,
			}
			break
		}
	}

	if gateway == nil {
		t.Fatal("Gateway node not found")
	}

	if gateway.fill != "#6366f1" {
		t.Errorf("Expected gateway fill '#6366f1', got '%s'", gateway.fill)
	}
	if gateway.stroke != "#4f46e5" {
		t.Errorf("Expected gateway stroke '#4f46e5', got '%s'", gateway.stroke)
	}
	if gateway.fontColor != "white" {
		t.Errorf("Expected gateway font-color 'white', got '%s'", gateway.fontColor)
	}
	if !gateway.bold {
		t.Error("Expected gateway bold to be true")
	}
}
