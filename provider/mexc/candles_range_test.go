package mexc

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/mdnmdn/bits/config"
	"github.com/mdnmdn/bits/model"
)

// fakeKlines serves 1m klines in [first, last] (ms) like the live API with
// startTime and endTime set: that window, at most limit rows (cap 1000).
func fakeKlines(first, last int64, requests *[]string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		*requests = append(*requests, r.URL.RawQuery)
		start, _ := strconv.ParseInt(q.Get("startTime"), 10, 64)
		end, _ := strconv.ParseInt(q.Get("endTime"), 10, 64)
		limit, _ := strconv.Atoi(q.Get("limit"))
		if limit == 0 {
			limit = 500
		}
		rows := []string{}
		for ts := first; ts <= last && len(rows) < limit; ts += 60000 {
			if ts < start || ts >= end {
				continue
			}
			rows = append(rows, fmt.Sprintf(`[%d,"1","2","0.5","1.5","3",%d,"4"]`, ts, ts+59999))
		}
		fmt.Fprintf(w, "[%s]", strings.Join(rows, ","))
	}))
}

func TestSpotCandles_RangePaging(t *testing.T) {
	base := time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC)
	const total = 1200
	var reqs []string
	srv := fakeKlines(base.UnixMilli(), base.Add((total-1)*time.Minute).UnixMilli(), &reqs)
	defer srv.Close()
	c := NewClient(config.MEXCConfig{BaseURL: srv.URL})

	to := base.Add(total * time.Minute)
	limit := 500
	var got []model.Candle
	for cursor := base; cursor.Before(to); cursor = cursor.Add(500 * time.Minute) {
		end := cursor.Add(500 * time.Minute)
		if end.After(to) {
			end = to
		}
		resp, err := c.Candles(context.Background(), "BTCUSDT", model.MarketSpot, "1m", model.CandleOpts{From: &cursor, To: &end, Limit: &limit})
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, resp.Data...)
	}
	if len(got) != total {
		t.Fatalf("got %d candles, want %d", len(got), total)
	}
	for i, cd := range got {
		if want := base.Add(time.Duration(i) * time.Minute); !cd.OpenTime.Equal(want) {
			t.Fatalf("candle %d at %s, want %s", i, cd.OpenTime, want)
		}
		if cd.Volume == nil {
			t.Fatalf("candle %d has no volume", i)
		}
	}
	if len(reqs) != 3 {
		t.Fatalf("got %d requests, want 3", len(reqs))
	}
	for _, r := range reqs {
		if !strings.Contains(r, "startTime=") || !strings.Contains(r, "endTime=") {
			t.Errorf("request lacks range params: %s", r)
		}
	}
}

func TestSpotCandles_NoOlderHistory(t *testing.T) {
	base := time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC)
	var reqs []string
	srv := fakeKlines(base.UnixMilli(), base.Add(10*time.Minute).UnixMilli(), &reqs)
	defer srv.Close()
	c := NewClient(config.MEXCConfig{BaseURL: srv.URL})
	from, to := base.Add(-time.Hour), base.Add(-30*time.Minute)
	resp, err := c.Candles(context.Background(), "BTCUSDT", model.MarketSpot, "1m", model.CandleOpts{From: &from, To: &to})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Data) != 0 {
		t.Fatalf("got %d candles, want none", len(resp.Data))
	}
}
