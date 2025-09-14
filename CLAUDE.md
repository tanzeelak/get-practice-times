# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based web scraper that extracts rehearsal room availability from the San Francisco Community Music Center (CMC) website. The scraper fetches studio availability data from Acuity Scheduling and provides a JSON output with chronologically sorted dates and time slots.

## Architecture

The application is a single-file Go program (`scraper.go`) with these key components:

### Core Data Flow
1. **Data Source**: Scrapes from CMC's Acuity Scheduling system at `schedule.php/app.acuity.scheduling.com/schedule.php?owner=30525417`
2. **HTTP Requests**: Uses Colly to make POST requests for each studio type/calendar combination
3. **HTML Parsing**: Extracts date and time information using goquery selectors
4. **Caching**: Redis-based caching with 6-hour expiration to reduce API load
5. **Output**: Chronologically sorted JSON with dates and available time slots

### Key Data Structures
- `Schedule`: `map[string]map[string][]string` - Maps dates to time slots to available studios
- `Calendar`: Struct containing studio ID and name
- `typeToCalendars`: Maps studio type IDs to calendar information

### Studio Mapping
The scraper queries 14 different studios using hardcoded type/calendar ID pairs. Some studios are intentionally omitted (Studio 7/Drumset, Recital Hall, Concert Hall) as noted in the README.

## Commands

### Prerequisites
- Redis server must be running: `redis-server`
- Go 1.22.0+ required

### Running the Scraper
```bash
go run scraper.go
```

### Cache Management
- Clear cache: `redis-cli del schedule`
- Check cache: `redis-cli get schedule`

### Development
- Run scraper: `go run scraper.go`
- Dependencies are managed via `go.mod` - run `go mod tidy` if needed

## Key Implementation Details

### Date Sorting Logic
The application manually builds JSON output to preserve chronological order (Go maps are unordered). It:
1. Parses dates with current year for proper chronological comparison
2. Sorts dates chronologically using time.Parse with format "Monday, January _2, 2006"
3. Sorts time slots within each date using "3:04PM" format
4. Builds JSON manually using strings.Builder to maintain order

### Cache Behavior
- **Cache Hit**: Returns stored JSON directly from Redis
- **Cache Miss**: Scrapes all studios, sorts data, stores in Redis, outputs result
- Cache expiration: 21600 seconds (6 hours)

### Error Handling
- Falls back to string comparison if date parsing fails
- Graceful handling of Redis connection errors
- Colly handles HTTP request failures

## Common Issues

### Date Ordering
The most critical functionality is maintaining chronological order of dates. The sorting logic appends the current year to date strings for proper time.Parse comparison, then manually builds JSON to preserve order.

### Studio Configuration
Studio mappings are hardcoded in `typeToCalendars`. When adding/removing studios, update both the map and consider whether they should be included or omitted based on the use case.

## Future TODOs (from README)
- Extract roomID without function-scoped variable updates
- Add studio booking links to Calendar objects  
- Flatten schedule formatting
- Accept input time and date ranges