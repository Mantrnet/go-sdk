# Mantr Go SDK

> **Deterministic Semantic Memory for Go Applications**

[![Go Reference](https://pkg.go.dev/badge/github.com/Mantrnet/go-sdk.svg)](https://pkg.go.dev/github.com/Mantrnet/go-sdk)
[![Go Report Card](https://goreportcard.com/badge/github.com/Mantrnet/go-sdk)](https://goreportcard.com/report/github.com/Mantrnet/go-sdk)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

---

## Installation

```bash
go get github.com/Mantrnet/go-sdk
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/Mantrnet/go-sdk"
)

func main() {
    client, err := mantr.NewClient("vak_live_...")
    if err != nil {
        log.Fatal(err)
    }
    
    result, err := client.Walk(&mantr.WalkRequest{
        Phonemes: []string{"dharma", "karma"},
        Depth:    3,
        Limit:    100,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Found %d paths\n", len(result.Paths))
}
```

---

## AI & Agentic Patterns

### 1. RAG Middleware (Gin/Echo)

```go
package middleware

import (
    "github.com/gin-gonic/gin"
    "github.com/Mantrnet/go-sdk"
)

func MantrContext(apiKey string) gin.HandlerFunc {
    client, _ := mantr.NewClient(apiKey)
    
    return func(c *gin.Context) {
        query := c.Query("q")
        if query == "" {
            c.Next()
            return
        }
        
        // Get deterministic context
        result, err := client.Walk(&mantr.WalkRequest{
            Phonemes: extractConcepts(query),
            Depth:    4,
            Limit:    20,
        })
        
        if err == nil {
            c.Set("mantr_context", result)
        }
        
        c.Next()
    }
}

// Usage
r := gin.Default()
r.Use(MantrContext("vak_live_..."))

r.GET("/search", func(c *gin.Context) {
    context, _ := c.Get("mantr_context")
    c.JSON(200, gin.H{"context": context})
})
```

### 2. Concurrent Agent Pool

```go
type AgentPool struct {
    client *mantr.Client
    jobs   chan string
    results chan *mantr.WalkResponse
}

func NewAgentPool(apiKey string, workers int) *AgentPool {
    client, _ := mantr.NewClient(apiKey)
    
    pool := &AgentPool{
        client:  client,
        jobs:    make(chan string, 100),
        results: make(chan *mantr.WalkResponse, 100),
    }
    
    // Start workers
    for i := 0; i < workers; i++ {
        go pool.worker()
    }
    
    return pool
}

func (p *AgentPool) worker() {
    for concept := range p.jobs {
        result, err := p.client.Walk(&mantr.WalkRequest{
            Phonemes: []string{concept},
            Depth:    3,
        })
        
        if err == nil {
            p.results <- result
        }
    }
}

// Process 1000 concepts concurrently
pool := NewAgentPool("vak_live_...", 10)
for _, concept := range concepts {
    pool.jobs <- concept
}
```

### 3. gRPC Service with Context

```go
type SearchServer struct {
    mantr *mantr.Client
    pb.UnimplementedSearchServiceServer
}

func (s *SearchServer) Search(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
    // Get Mantr context
    context, err := s.mantr.Walk(&mantr.WalkRequest{
        Phonemes: extractConcepts(req.Query),
        Depth:    4,
    })
    if err != nil {
        return nil, err
    }
    
    // Use context for downstream processing
    results := s.processWithContext(req.Query, context)
    
    return &pb.SearchResponse{Results: results}, nil
}
```

### 4. CLI Tool with Caching

```go
type CLIAgent struct {
    client *mantr.Client
    cache  map[string]*mantr.WalkResponse
}

func (a *CLIAgent) Query(concept string) (*mantr.WalkResponse, error) {
    // Check cache
    if cached, ok := a.cache[concept]; ok {
        return cached, nil
    }
    
    // Walk
    result, err := a.client.Walk(&mantr.WalkRequest{
        Phonemes: []string{concept},
    })
    if err != nil {
        return nil, err
    }
    
    // Cache
    a.cache[concept] = result
    return result, nil
}
```

---

## API Reference

```go
// Client initialization
client, err := mantr.NewClient("vak_live_...", options...)

// Walk the graph
result, err := client.Walk(&mantr.WalkRequest{
    Phonemes: []string{"concept1", "concept2"},
    Pod:      "optional_pod",
    Depth:    3,
    Limit:    100,
})

// Response
type WalkResponse struct {
    Paths       []PathResult
    LatencyUS   int
    CreditsUsed int
}

type PathResult struct {
    Nodes []string
    Score float64
    Depth int
}
```

---

## License

MIT
