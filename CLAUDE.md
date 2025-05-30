# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What is Spiders?

Spiders is a multi-threaded web crawler written in Go that collects text content from websites. It supports two main modes:
- **Crawl Mode**: Discovers links starting from a root URL
- **Batch Mode**: Processes predefined lists or patterns of URLs

## Quick Start

### Essential Commands
```bash
# Build the project
go build -o spiders

# Run with configuration (config file is required)
./spiders -c config.yaml

# Resume an interrupted crawl
./spiders -r -c config.yaml

# Run tests
go test ./crawler/
```

### Basic Configuration (`config.yaml`)
```yaml
project: "My Crawler"
project_type: "crawl"          # or "batch"
directory: "output"
web_count: 100                 # number of concurrent workers
root_url: "https://example.com"
content_selector: ".content"   # CSS selector for content to extract
```

## How It Works

### 🏗️ System Architecture

The crawler uses a multi-threaded worker pool design:

```
┌─────────────┐    ┌──────────────┐    ┌─────────────┐
│ Main Thread │───▶│ Worker Pool  │───▶│ SQLite DB   │
│             │    │ (N workers)  │    │ (URL Queue) │
└─────────────┘    └──────────────┘    └─────────────┘
       │                   │
       ▼                   ▼
┌─────────────┐    ┌──────────────┐
│ Link        │    │ Content      │
│ Processor   │    │ Extractor    │
└─────────────┘    └──────────────┘
```

### 🔧 Core Components

| Component | File | Purpose |
|-----------|------|---------|
| **Entry Point** | `main.go` | CLI handling, signal management |
| **Orchestrator** | `crawler/spider.go` | Worker coordination, shutdown logic |
| **URL Discovery** | `crawler/link-processor.go` | Find and validate new URLs |
| **Batch Processing** | `crawler/batch-link-processor.go` | Handle URL lists/patterns |
| **Configuration** | `crawler/config/` | YAML config loading |
| **Persistence** | `crawler/db/` | SQLite queue and status tracking |

### 🎯 Two Operating Modes

#### Crawl Mode (`project_type: "crawl"`)
- Starts from `root_url`
- Discovers links automatically
- Uses regex patterns to validate URLs (`link_validator`)
- Excludes unwanted URLs (`link_invalidators`)

#### Batch Mode (`project_type: "batch"`)  
- Processes predefined URL lists
- Supports URL patterns: `https://site.com/[$1-$100]/page`
- Can read from file: `batch_file: "urls.txt"`

## Configuration Guide

### Key Settings

**Worker Control**
```yaml
web_count: 50                    # Number of concurrent workers
```

**URL Processing** 
```yaml
link_validator: "^https://example.com/.*"     # Include only matching URLs
link_invalidators:                            # Exclude these patterns
  - "^https://example.com/admin"
  - ".*\.pdf$"
link_sanitizer: "(#.*)"                       # Remove URL fragments
```

**Content Extraction**
```yaml
page_validator: ".article"                    # Page must have this element
content_selector: ".content"                  # Extract this content
content_holder: "text"                        # Extract: text, attr, or html
```

**Proxy Support**
```yaml
proxy_api: "http://api.scraperapi.com?key=YOUR_KEY&url="
```

## Development Guidelines

### 🔍 Key Implementation Details

- **Structured Architecture**: Clean separation between main app, crawler orchestration, and database layers
- **Graceful Shutdown**: SIGTERM/SIGINT properly terminates all workers using coordinated channels
- **Resume Capability**: SQLite tracks processing state for interruption recovery  
- **Rate Limiting**: Random 1-5 second delays between requests to be respectful
- **Deduplication**: MD5 hashes prevent reprocessing identical content
- **Content Storage**: Individual text files numbered sequentially in organized directories
- **Error Handling**: Comprehensive error propagation with context-aware logging
- **Concurrent Processing**: Worker pool pattern with configurable concurrency

### 🧪 Testing

```bash
# Run all tests
go test ./crawler/

# Test with verbose output
go test -v ./crawler/

# Build and verify
go build -o spiders
```

### 📁 Refactored File Structure

```
spiders/
├── main.go                      # Clean CLI entry point with App struct
├── config.yaml                  # YAML configuration
├── crawler/
│   ├── spider.go               # Crawler struct and worker orchestration
│   ├── link-processor.go       # URL discovery and validation (refactored)
│   ├── batch-link-processor.go # Batch URL processing
│   ├── setup.go               # System initialization
│   ├── config/
│   │   ├── config.go          # Enhanced config loading with validation
│   │   └── structs.go         # Configuration types and constants
│   └── db/
│       ├── init.go            # Database multiplexer (unchanged)
│       ├── setup.go           # Enhanced setup with error handling
│       ├── queries.go         # Improved query functions with better naming
│       └── structs.go         # Enhanced Node struct with proper field names
└── output/                    # Generated content directory
    └── data/                  # Extracted content files
```

### 🛠️ Recent Improvements

- **App Structure**: Main entry reorganized into clean App struct with separate concerns
- **Error Handling**: Comprehensive error propagation instead of log.Fatal calls
- **Function Naming**: CamelCase naming convention throughout (pushAllBatchLinks, extractTextContent)
- **Database Layer**: Better error messages, proper field naming (ID instead of Id)
- **Logging**: Structured logging with context instead of fmt.Println
- **Validation**: Enhanced configuration validation with detailed error messages
- **Code Organization**: Separated complex functions into smaller, focused units