
# CMC Rehearsal Room Scraper

A web application that scrapes and displays available rehearsal room slots from the San Francisco Community Music Center (CMC).

## Project Structure

```
├── backend/          # Go API server
│   ├── server.go     # Main server with scraping logic
│   ├── go.mod        # Go dependencies
│   └── go.sum        # Go dependency checksums
├── frontend/         # Static web frontend
│   ├── index.html    # Main HTML page
│   ├── styles.css    # CSS styling
│   └── script.js     # JavaScript for API calls
├── CLAUDE.md         # Development guidance
└── README.md         # This file
```

## Quick Start

### Prerequisites
- Go 1.22.0+
- Redis server
- Modern web browser

### Development Setup

1. **Start Redis:**
   ```bash
   redis-server
   ```

2. **Start the backend API:**
   ```bash
   cd backend
   go run server.go
   ```
   Server will start at `http://localhost:8080`

3. **Open the frontend:**
   Open `frontend/index.html` in your browser, or serve it with a local server:
   ```bash
   cd frontend
   python3 -m http.server 3000
   # Then open http://localhost:3000
   ```

## API Endpoints

- `GET /api/rehearsals` - Returns available rehearsal slots as JSON
- `GET /health` - Health check endpoint

## Data Source

Scrapes from [CMC's Acuity Scheduling system](https://sfcmc.org/events/event-space-rentals/) at `schedule.php/app.acuity.scheduling.com/schedule.php?owner=30525417`

Omitted Rooms
- 58324961: {9703524, "Studio 7"} -> Drumset
- 54535652: {9651030, "Recital Hall"},
- 58324504: {9650981, "Concert Hall"},

TODOs
- Figure out how to extract roomID instead of updating a function-scoped value
- Investigate if I can get extract a link for each studio. If possible, add to Calendar object
- Accept an input time range
- Accept an input date range
- Deploy to Vercel
- Make mobile friendly
