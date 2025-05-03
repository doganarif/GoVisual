package dashboard

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"time"

	"github.com/doganarif/govisual/internal/store"
)

// Handler is the HTTP handler for the dashboard
type Handler struct {
	store store.Store
}

// NewHandler creates a new dashboard handler
func NewHandler(store store.Store) *Handler {
	return &Handler{
		store: store,
	}
}

// ServeHTTP implements the http.Handler interface
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// API endpoints
	switch path.Clean(r.URL.Path) {
	case "/api/requests":
		h.handleAPIRequests(w, r)
		return
	case "/api/events":
		h.handleSSE(w, r)
		return
	case "/":
		h.handleDashboard(w, r)
		return
	default:
		// Serve a simple 404 page
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 - Not Found"))
	}
}

// handleDashboard serves the dashboard HTML
func (h *Handler) handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
    <title>HTTP Request Visualizer</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; }
        h1 { color: #333; }
        #requests { border-collapse: collapse; width: 100%; }
        #requests th, #requests td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        #requests tr:nth-child(even) { background-color: #f2f2f2; }
        #requests th { padding-top: 12px; padding-bottom: 12px; background-color: #4CAF50; color: white; }
    </style>
</head>
<body>
    <h1>HTTP Request Visualizer</h1>
    <table id="requests">
        <thead>
            <tr>
                <th>Time</th>
                <th>Method</th>
                <th>Path</th>
                <th>Status</th>
                <th>Duration</th>
            </tr>
        </thead>
        <tbody id="requestsBody">
            <!-- Will be populated by JavaScript -->
        </tbody>
    </table>

    <script>
        // Fetch requests initially
        fetch('/__viz/api/requests')
            .then(response => response.json())
            .then(data => {
                updateTable(data);
            })
            .catch(error => console.error('Error fetching initial data:', error));
        
        // Set up SSE for live updates
        const evtSource = new EventSource('/__viz/api/events');
        evtSource.onmessage = function(event) {
            try {
                const data = JSON.parse(event.data);
                updateTable(data);
            } catch (error) {
                console.error('Error parsing SSE data:', error);
            }
        };
        
        function updateTable(requests) {
            const tbody = document.getElementById('requestsBody');
            tbody.innerHTML = '';
            
            if (!requests || requests.length === 0) {
                const row = document.createElement('tr');
                const cell = document.createElement('td');
                cell.colSpan = 5;
                cell.textContent = 'No requests logged yet';
                cell.style.textAlign = 'center';
                row.appendChild(cell);
                tbody.appendChild(row);
                return;
            }
            
            requests.forEach(req => {
                const row = document.createElement('tr');
                
                const timeCell = document.createElement('td');
                timeCell.textContent = new Date(req.Timestamp).toLocaleTimeString();
                row.appendChild(timeCell);
                
                const methodCell = document.createElement('td');
                methodCell.textContent = req.Method;
                row.appendChild(methodCell);
                
                const pathCell = document.createElement('td');
                pathCell.textContent = req.Path + (req.Query ? '?' + req.Query : '');
                row.appendChild(pathCell);
                
                const statusCell = document.createElement('td');
                statusCell.textContent = req.StatusCode;
                row.appendChild(statusCell);
                
				const durationCell = document.createElement('td');
				if (req.Duration !== undefined && req.Duration !== null) {
                    durationCell.textContent = req.Duration + ' ms';
                } else {
                    durationCell.textContent = 'N/A';
                }
				row.appendChild(durationCell);
                
                tbody.appendChild(row);
            });
        }
    </script>
</body>
</html>
	`))
}

// handleAPIRequests serves the JSON API for requests
func (h *Handler) handleAPIRequests(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	requests := h.store.GetAll()
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.Encode(requests)
}

// handleSSE handles Server-Sent Events for live updates
func (h *Handler) handleSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	requests := h.store.GetAll()
	data, _ := json.Marshal(requests)
	fmt.Fprintf(w, "data: %s\n\n", data)
	flusher.Flush()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			requests := h.store.GetAll()
			data, _ := json.Marshal(requests)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
