package cryptocom

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/mdnmdn/bits/model"
)

// fakeCandleServer serves 1m candles in [first, last] (ms) the way the venue
// does: start_ts inclusive, end_ts exclusive, at most count rows, oldest first.
func fakeCandleServer(t *testing.T, first, last int64, requests *[]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		*requests = append(*requests, r.URL.RawQuery)
		start, _ := strconv.ParseInt(q.Get("start_ts"), 10, 64)
		end, _ := strconv.ParseInt(q.Get("end_ts"), 10, 64)
		count, _ := strconv.Atoi(q.Get("count"))
		var rows []string
		for ts := first; ts <= last && len(rows) < count; ts += 60000 {
			if ts < start || ts >= end {
				continue
			}
			rows = append(rows, fmt.Sprintf(`{"o":"1","h":"2","l":"0.5","c":"1.5","v":"3","t":%d}`, ts))
		}
		fmt.Fprintf(w, `{"id":-1,"method":"public/get-candlestick","code":0,"result":{"interval":"1m","data":[%s]}}`, strings.Join(rows, ","))
	}))
}

func TestCandles_RangePaging(t *testing.T) {
	base := time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC)
	const total = 700
	var reqs []string
	srv := fakeCandleServer(t, base.UnixMilli(), base.Add((total-1)*time.Minute).UnixMilli(), &reqs)
	defer srv.Close()
	c := newTestClient(srv.URL)

	to := base.Add(total * time.Minute)
	limit := 1000 // above the venue cap: the provider must clamp it to 300
	var got []model.Candle
	for cursor := base; cursor.Before(to); cursor = cursor.Add(300 * time.Minute) {
		end := cursor.Add(300 * time.Minute)
		if end.After(to) {
			end = to
		}
		resp, err := c.Candles(context.Background(), "BTC_USDT", model.MarketSpot, "1m", model.CandleOpts{From: &cursor, To: &end, Limit: &limit})
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
	}
	if len(reqs) != 3 {
		t.Fatalf("got %d requests, want 3", len(reqs))
	}
	for _, r := range reqs {
		if !strings.Contains(r, "start_ts=") || !strings.Contains(r, "end_ts=") || !strings.Contains(r, "count=300") {
			t.Errorf("request lacks range params: %s", r)
		}
	}
}

func TestCandles_NoOlderHistory(t *testing.T) {
	base := time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC)
	var reqs []string
	srv := fakeCandleServer(t, base.UnixMilli(), base.Add(10*time.Minute).UnixMilli(), &reqs)
	defer srv.Close()
	c := newTestClient(srv.URL)
	from, to := base.Add(-time.Hour), base.Add(-30*time.Minute)
	resp, err := c.Candles(context.Background(), "BTC_USDT", model.MarketSpot, "1m", model.CandleOpts{From: &from, To: &to})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Data) != 0 {
		t.Fatalf("got %d candles, want none", len(resp.Data))
	}
}

func TestCandles_NoRangeKeepsQuery(t *testing.T) {
	var reqs []string
	srv := fakeCandleServer(t, 0, 0, &reqs)
	defer srv.Close()
	if _, err := newTestClient(srv.URL).Candles(context.Background(), "BTC_USDT", model.MarketSpot, "1m", model.CandleOpts{}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(reqs[0], "start_ts") || strings.Contains(reqs[0], "end_ts") {
		t.Errorf("unexpected range params: %s", reqs[0])
	}
}
