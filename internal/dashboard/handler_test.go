package dashboard

import (
	"bufio"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/doganarif/govisual/v2/store"
)

func TestSSEPushesOnAdd(t *testing.T) {
	ns := store.WithNotify(store.NewMemory(10))
	srv := httptest.NewServer(NewHandler(ns, nil, HandlerOptions{}))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/events")
	if err != nil {
		t.Fatalf("connect SSE: %v", err)
	}
	defer resp.Body.Close()

	events := make(chan string, 16)
	go func() {
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "event: ") {
				events <- strings.TrimPrefix(line, "event: ")
			}
		}
	}()

	select {
	case ev := <-events:
		if ev != "snapshot" {
			t.Fatalf("first event = %q, want snapshot", ev)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no snapshot event")
	}

	ns.Add(&store.RequestLog{ID: "x1", Timestamp: time.Now(), Method: "GET", Path: "/p"})

	// The push must arrive well under the 15s heartbeat tick.
	select {
	case ev := <-events:
		if ev != "append" {
			t.Fatalf("second event = %q, want append", ev)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("append was not pushed; SSE still waiting on the ticker")
	}
}
