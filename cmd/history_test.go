package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coingecko/coingecko-cli/internal/api"
	"github.com/coingecko/coingecko-cli/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withInstantRetrySleep overrides retrySleep to return immediately during tests.
// Returns a cleanup function to restore the original.
func withInstantRetrySleep(t *testing.T) {
	t.Helper()
	orig := retrySleep
	retrySleep = func(d time.Duration) <-chan time.Time {
		ch := make(chan time.Time, 1)
		ch <- time.Now()
		return ch
	}
	t.Cleanup(func() { retrySleep = orig })
}

func testAPIClient(handler http.HandlerFunc) (*api.Client, *httptest.Server) {
	srv := httptest.NewServer(handler)
	cfg := &config.Config{APIKey: "test-key", Tier: config.TierDemo}
	c := api.NewClient(cfg)
	c.SetBaseURL(srv.URL)
	return c, srv
}

func testPaidAPIClient(handler http.HandlerFunc) (*api.Client, *httptest.Server) {
	srv := httptest.NewServer(handler)
	cfg := &config.Config{APIKey: "test-key", Tier: config.TierPaid}
	c := api.NewClient(cfg)
	c.SetBaseURL(srv.URL)
	return c, srv
}

// --- Constants ---

func TestConstants(t *testing.T) {
	assert.Equal(t, 90, hourlyChunkDays)
	assert.Equal(t, 91, minDailyRangeDays)
}

func TestOHLCRangeChunkDays(t *testing.T) {
	tests := []struct {
		interval string
		want     int
	}{
		{"daily", 170},
		{"hourly", 30},
		{"", 0},
		{"weekly", 0},
	}
	for _, tt := range tests {
		t.Run(tt.interval, func(t *testing.T) {
			assert.Equal(t, tt.want, ohlcRangeChunkDays(tt.interval))
		})
	}
}

// --- checkHourlyAvailability ---

func TestCheckHourlyAvailability_OK(t *testing.T) {
	// Date after cutoff should pass.
	fromUnix := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	assert.NoError(t, checkHourlyAvailability(fromUnix))
}

func TestCheckHourlyAvailability_ExactCutoff(t *testing.T) {
	// Exactly the cutoff date should pass.
	assert.NoError(t, checkHourlyAvailability(hourlyAvailableFrom.Unix()))
}

func TestCheckHourlyAvailability_BeforeCutoff(t *testing.T) {
	// Date before cutoff should fail.
	fromUnix := time.Date(2017, 12, 1, 0, 0, 0, 0, time.UTC).Unix()
	err := checkHourlyAvailability(fromUnix)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "hourly data is only available from 2018-01-30")
}

// --- trimTimeseries ---

func TestTrimTimeseries(t *testing.T) {
	data := [][]float64{
		{1000, 100}, {2000, 200}, {3000, 300}, {4000, 400}, {5000, 500},
	}

	// Trim everything before 3000
	result := trimTimeseries(data, 3000)
	assert.Len(t, result, 3)
	assert.Equal(t, float64(3000), result[0][0])
	assert.Equal(t, float64(5000), result[2][0])

	// Trim nothing
	result = trimTimeseries(data, 1000)
	assert.Len(t, result, 5)

	// Trim everything
	result = trimTimeseries(data, 9000)
	assert.Len(t, result, 0)

	// Empty input
	result = trimTimeseries(nil, 1000)
	assert.Len(t, result, 0)
}

// --- fetchMarketChartBatched ---

func TestFetchMarketChartBatched_SingleChunk(t *testing.T) {
	var calls int32
	client, srv := testAPIClient(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		resp := api.MarketChartResponse{
			Prices:       [][]float64{{1000, 50000}, {2000, 51000}},
			MarketCaps:   [][]float64{{1000, 900e9}, {2000, 910e9}},
			TotalVolumes: [][]float64{{1000, 30e9}, {2000, 31e9}},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()

	from := int64(1700000000)
	to := from + 5*86400
	data, err := fetchMarketChartBatched(context.Background(), client, "bitcoin", "usd", from, to, "", 90)
	require.NoError(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&calls))
	assert.Len(t, data.Prices, 2)
	assert.Len(t, data.MarketCaps, 2)
	assert.Len(t, data.TotalVolumes, 2)
}

func TestFetchMarketChartBatched_HourlyOmitsInterval(t *testing.T) {
	var gotIntervals []string
	client, srv := testAPIClient(func(w http.ResponseWriter, r *http.Request) {
		gotIntervals = append(gotIntervals, r.URL.Query().Get("interval"))
		_ = json.NewEncoder(w).Encode(api.MarketChartResponse{
			Prices: [][]float64{{1000, 50000}},
		})
	})
	defer srv.Close()

	from := int64(1700000000)
	to := from + 200*86400 // 200 days → 3 chunks
	_, err := fetchMarketChartBatched(context.Background(), client, "bitcoin", "usd", from, to, "", hourlyChunkDays)
	require.NoError(t, err)
	require.Len(t, gotIntervals, 3)
	for i, iv := range gotIntervals {
		assert.Empty(t, iv, "chunk %d should have no interval param", i+1)
	}
}

func TestFetchMarketChartBatched_MultipleChunks(t *testing.T) {
	var calls int32
	client, srv := testAPIClient(func(w http.ResponseWriter, r *http.Request) {
		callNum := atomic.AddInt32(&calls, 1)
		base := float64(callNum) * 1e6
		resp := api.MarketChartResponse{
			Prices:       [][]float64{{base + 1, 50000}, {base + 2, 51000}},
			MarketCaps:   [][]float64{{base + 1, 900e9}, {base + 2, 910e9}},
			TotalVolumes: [][]float64{{base + 1, 30e9}, {base + 2, 31e9}},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()

	from := int64(1700000000)
	to := from + 200*86400 // 200 days / 90-day chunks = 3 chunks
	data, err := fetchMarketChartBatched(context.Background(), client, "bitcoin", "usd", from, to, "", hourlyChunkDays)
	require.NoError(t, err)
	assert.Equal(t, int32(3), atomic.LoadInt32(&calls))
	assert.Len(t, data.Prices, 6)
	assert.Len(t, data.MarketCaps, 6)
	assert.Len(t, data.TotalVolumes, 6)
}

func TestFetchMarketChartBatched_Deduplication(t *testing.T) {
	callCount := 0
	client, srv := testAPIClient(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		var resp api.MarketChartResponse
		if callCount == 1 {
			resp = api.MarketChartResponse{
				Prices:       [][]float64{{1000, 50000}, {2000, 51000}, {3000, 52000}},
				MarketCaps:   [][]float64{{1000, 900e9}, {2000, 910e9}, {3000, 920e9}},
				TotalVolumes: [][]float64{{1000, 30e9}, {2000, 31e9}, {3000, 32e9}},
			}
		} else {
			// Overlapping: timestamp 3000 appears in both chunks.
			resp = api.MarketChartResponse{
				Prices:       [][]float64{{3000, 52000}, {4000, 53000}},
				MarketCaps:   [][]float64{{3000, 920e9}, {4000, 930e9}},
				TotalVolumes: [][]float64{{3000, 32e9}, {4000, 33e9}},
			}
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()

	from := int64(1700000000)
	to := from + 150*86400 // 2 chunks of 90 days
	data, err := fetchMarketChartBatched(context.Background(), client, "bitcoin", "usd", from, to, "", hourlyChunkDays)
	require.NoError(t, err)
	assert.Equal(t, 2, callCount)
	// 3 unique from chunk1 + 1 new from chunk2 = 4
	assert.Len(t, data.Prices, 4)
	assert.Len(t, data.MarketCaps, 4)
	assert.Len(t, data.TotalVolumes, 4)
}

func TestFetchMarketChartBatched_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var calls int32
	client, srv := testAPIClient(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			cancel()
		}
		resp := api.MarketChartResponse{
			Prices: [][]float64{{float64(n) * 1000, 50000}},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()

	from := int64(1700000000)
	to := from + 200*86400 // 3 chunks
	_, err := fetchMarketChartBatched(ctx, client, "bitcoin", "usd", from, to, "", hourlyChunkDays)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, int32(1), atomic.LoadInt32(&calls))
}

func TestFetchMarketChartBatched_APIError(t *testing.T) {
	callCount := 0
	client, srv := testAPIClient(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 2 {
			w.WriteHeader(500)
			_, _ = fmt.Fprint(w, `{"status":{"error_code":500,"error_message":"Internal error"}}`)
			return
		}
		resp := api.MarketChartResponse{
			Prices: [][]float64{{float64(callCount) * 1000, 50000}},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()

	from := int64(1700000000)
	to := from + 200*86400
	_, err := fetchMarketChartBatched(context.Background(), client, "bitcoin", "usd", from, to, "", hourlyChunkDays)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "chunk 2/3 failed")
}

func TestFetchMarketChartBatched_EmptyChunk(t *testing.T) {
	callCount := 0
	client, srv := testAPIClient(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		var resp api.MarketChartResponse
		if callCount == 1 {
			resp.Prices = [][]float64{{1000, 50000}}
			resp.MarketCaps = [][]float64{{1000, 900e9}}
			resp.TotalVolumes = [][]float64{{1000, 30e9}}
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()

	from := int64(1700000000)
	to := from + 150*86400 // 2 chunks
	data, err := fetchMarketChartBatched(context.Background(), client, "bitcoin", "usd", from, to, "", hourlyChunkDays)
	require.NoError(t, err)
	assert.Equal(t, 2, callCount)
	assert.Len(t, data.Prices, 1)
	assert.Len(t, data.MarketCaps, 1)
	assert.Len(t, data.TotalVolumes, 1)
}

func TestFetchMarketChartBatched_ChunkBoundaries(t *testing.T) {
	var requestParams []map[string]string
	client, srv := testAPIClient(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		requestParams = append(requestParams, map[string]string{
			"from": q.Get("from"),
			"to":   q.Get("to"),
		})
		_ = json.NewEncoder(w).Encode(api.MarketChartResponse{})
	})
	defer srv.Close()

	from := int64(1700000000)
	chunkSec := int64(hourlyChunkDays) * 86400
	to := from + 200*86400 // 3 chunks

	_, err := fetchMarketChartBatched(context.Background(), client, "bitcoin", "usd", from, to, "", hourlyChunkDays)
	require.NoError(t, err)
	require.Len(t, requestParams, 3)

	assert.Equal(t, fmt.Sprintf("%d", from), requestParams[0]["from"])
	assert.Equal(t, fmt.Sprintf("%d", from+chunkSec), requestParams[0]["to"])

	assert.Equal(t, fmt.Sprintf("%d", from+chunkSec), requestParams[1]["from"])
	assert.Equal(t, fmt.Sprintf("%d", from+2*chunkSec), requestParams[1]["to"])

	assert.Equal(t, fmt.Sprintf("%d", from+2*chunkSec), requestParams[2]["from"])
	assert.Equal(t, fmt.Sprintf("%d", to), requestParams[2]["to"])
}

// --- future toUnix capping ---

func TestFetchOHLCRangeBatched_FutureToUnixCapped(t *testing.T) {
	// Regression: endOfDayUnix(today) produces a future timestamp.
	// The OHLC range API rejects future dates with 400.
	// Verify that fetchOHLCRangeBatched caps toUnix at now, so no chunk sends a future "to".
	var requestTos []string
	client, srv := testPaidAPIClient(func(w http.ResponseWriter, r *http.Request) {
		requestTos = append(requestTos, r.URL.Query().Get("to"))
		_ = json.NewEncoder(w).Encode(api.OHLCData{{1000, 1, 2, 3, 4}})
	})
	defer srv.Close()

	now := time.Now().Unix()
	from := now - 40*86400        // 40 days ago
	futureToUnix := now + 12*3600 // 12 hours in the future (endOfDayUnix scenario)

	_, err := fetchOHLCRangeBatched(context.Background(), client, "bitcoin", "usd", from, futureToUnix, "hourly", 30)
	require.NoError(t, err)

	// The last "to" param sent to the API must not exceed current time.
	for i, toStr := range requestTos {
		var toVal int64
		_, _ = fmt.Sscanf(toStr, "%d", &toVal)
		assert.LessOrEqual(t, toVal, time.Now().Unix()+1, "chunk %d sent future 'to' to API", i+1)
	}
}

func TestFetchMarketChartBatched_FutureToUnixCapped(t *testing.T) {
	var requestTos []string
	client, srv := testAPIClient(func(w http.ResponseWriter, r *http.Request) {
		requestTos = append(requestTos, r.URL.Query().Get("to"))
		_ = json.NewEncoder(w).Encode(api.MarketChartResponse{
			Prices: [][]float64{{1000, 50000}},
		})
	})
	defer srv.Close()

	now := time.Now().Unix()
	from := now - 100*86400
	futureToUnix := now + 12*3600

	_, err := fetchMarketChartBatched(context.Background(), client, "bitcoin", "usd", from, futureToUnix, "", hourlyChunkDays)
	require.NoError(t, err)

	for i, toStr := range requestTos {
		var toVal int64
		_, _ = fmt.Sscanf(toStr, "%d", &toVal)
		assert.LessOrEqual(t, toVal, time.Now().Unix()+1, "chunk %d sent future 'to' to API", i+1)
	}
}

// --- fetchOHLCRangeBatched ---

func TestFetchOHLCRangeBatched_MultipleChunks(t *testing.T) {
	var calls int32
	client, srv := testPaidAPIClient(func(w http.ResponseWriter, r *http.Request) {
		callNum := atomic.AddInt32(&calls, 1)
		base := float64(callNum) * 1e9
		data := api.OHLCData{
			{base + 1, 100, 110, 90, 105},
			{base + 2, 105, 115, 95, 110},
		}
		_ = json.NewEncoder(w).Encode(data)
	})
	defer srv.Close()

	from := int64(1700000000)
	to := from + 400*86400 // 400 days / 170-day chunks = 3 chunks
	data, err := fetchOHLCRangeBatched(context.Background(), client, "bitcoin", "usd", from, to, "daily", 170)
	require.NoError(t, err)
	assert.Equal(t, int32(3), atomic.LoadInt32(&calls))
	assert.Len(t, data, 6)
}

func TestFetchOHLCRangeBatched_Deduplication(t *testing.T) {
	callCount := 0
	client, srv := testPaidAPIClient(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		var data api.OHLCData
		if callCount == 1 {
			data = api.OHLCData{
				{1000, 100, 110, 90, 105},
				{2000, 105, 115, 95, 110},
			}
		} else {
			// Overlapping timestamp 2000
			data = api.OHLCData{
				{2000, 105, 115, 95, 110},
				{3000, 110, 120, 100, 115},
			}
		}
		_ = json.NewEncoder(w).Encode(data)
	})
	defer srv.Close()

	from := int64(1700000000)
	to := from + 50*86400 // 2 chunks of 30
	data, err := fetchOHLCRangeBatched(context.Background(), client, "bitcoin", "usd", from, to, "hourly", 30)
	require.NoError(t, err)
	assert.Equal(t, 2, callCount)
	assert.Len(t, data, 3) // 2 + 1 new (ts 2000 deduped)
}

// --- withRateLimitRetry ---

func TestWithRateLimitRetry_Success(t *testing.T) {
	calls := 0
	err := withRateLimitRetry(context.Background(), "test", func() error {
		calls++
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, 1, calls)
}

func TestWithRateLimitRetry_TransientRateLimit(t *testing.T) {
	withInstantRetrySleep(t)
	calls := 0
	err := withRateLimitRetry(context.Background(), "test", func() error {
		calls++
		if calls <= 2 {
			return &api.RateLimitError{RetryAfter: 1}
		}
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, 3, calls) // 2 retries then success
}

func TestWithRateLimitRetry_ExhaustedRetries(t *testing.T) {
	withInstantRetrySleep(t)
	calls := 0
	err := withRateLimitRetry(context.Background(), "test", func() error {
		calls++
		return &api.RateLimitError{RetryAfter: 1}
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, api.ErrRateLimited)
	assert.Equal(t, maxChunkRetries+1, calls) // initial + 3 retries
}

func TestWithRateLimitRetry_NonRateLimitError(t *testing.T) {
	calls := 0
	err := withRateLimitRetry(context.Background(), "test", func() error {
		calls++
		return fmt.Errorf("some other error")
	})
	require.Error(t, err)
	assert.Equal(t, 1, calls) // no retry on non-429
	assert.Contains(t, err.Error(), "some other error")
}

func TestWithRateLimitRetry_ContextCancelled(t *testing.T) {
	withInstantRetrySleep(t)
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	err := withRateLimitRetry(ctx, "test", func() error {
		calls++
		cancel() // cancel during first rate limit
		return &api.RateLimitError{RetryAfter: 0}
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestFetchMarketChartBatched_RateLimitRetry(t *testing.T) {
	withInstantRetrySleep(t)
	var calls int32
	client, srv := testAPIClient(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(429)
			return
		}
		_ = json.NewEncoder(w).Encode(api.MarketChartResponse{
			Prices: [][]float64{{1000, 50000}},
		})
	})
	defer srv.Close()

	from := int64(1700000000)
	to := from + 5*86400 // single chunk
	data, err := fetchMarketChartBatched(context.Background(), client, "bitcoin", "usd", from, to, "", 90)
	require.NoError(t, err)
	assert.Equal(t, int32(2), atomic.LoadInt32(&calls)) // 1 retry
	assert.Len(t, data.Prices, 1)
}

func TestFetchMarketChartBatched_RateLimitResetRetry(t *testing.T) {
	withInstantRetrySleep(t)

	// Track the durations passed to retrySleep to verify ResetAt branch is used.
	var sleepDurations []time.Duration
	orig := retrySleep
	retrySleep = func(d time.Duration) <-chan time.Time {
		sleepDurations = append(sleepDurations, d)
		ch := make(chan time.Time, 1)
		ch <- time.Now()
		return ch
	}
	t.Cleanup(func() { retrySleep = orig })

	var calls int32
	client, srv := testAPIClient(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			// Return 429 with only x-ratelimit-reset (no Retry-After) — the production pattern.
			resetTime := time.Now().Add(5 * time.Second).UTC().Format("2006-01-02 15:04:05 -0700")
			w.Header().Set("x-ratelimit-reset", resetTime)
			w.WriteHeader(429)
			return
		}
		_ = json.NewEncoder(w).Encode(api.MarketChartResponse{
			Prices: [][]float64{{1000, 50000}},
		})
	})
	defer srv.Close()

	from := int64(1700000000)
	to := from + 5*86400 // single chunk
	data, err := fetchMarketChartBatched(context.Background(), client, "bitcoin", "usd", from, to, "", 90)
	require.NoError(t, err)
	assert.Equal(t, int32(2), atomic.LoadInt32(&calls)) // 1 retry then success
	assert.Len(t, data.Prices, 1)

	// Verify the ResetAt branch was exercised: sleep duration should be ~5s (time.Until(resetAt)).
	require.Len(t, sleepDurations, 1, "expected exactly one retry sleep")
	assert.GreaterOrEqual(t, sleepDurations[0], 3*time.Second, "sleep should reflect time.Until(resetAt)")
	assert.LessOrEqual(t, sleepDurations[0], 7*time.Second, "sleep should be roughly 5s")
}

func TestWithRateLimitRetry_ResetAtBranch(t *testing.T) {
	// Unit-level test: verify the ResetAt path computes wait from time.Until(resetAt).
	var sleepDuration time.Duration
	orig := retrySleep
	retrySleep = func(d time.Duration) <-chan time.Time {
		sleepDuration = d
		ch := make(chan time.Time, 1)
		ch <- time.Now()
		return ch
	}
	t.Cleanup(func() { retrySleep = orig })

	calls := 0
	err := withRateLimitRetry(context.Background(), "test", func() error {
		calls++
		if calls == 1 {
			return &api.RateLimitError{
				ResetAt: time.Now().Add(10 * time.Second),
			}
		}
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, 2, calls)
	// The wait should be ~10s (time.Until(resetAt)), not the backoff default.
	assert.GreaterOrEqual(t, sleepDuration, 8*time.Second)
	assert.LessOrEqual(t, sleepDuration, 12*time.Second)
}

func TestFetchOHLCRangeBatched_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var calls int32
	client, srv := testPaidAPIClient(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			cancel()
		}
		_ = json.NewEncoder(w).Encode(api.OHLCData{{float64(n) * 1000, 1, 2, 3, 4}})
	})
	defer srv.Close()

	from := int64(1700000000)
	to := from + 400*86400
	_, err := fetchOHLCRangeBatched(ctx, client, "bitcoin", "usd", from, to, "daily", 170)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, int32(1), atomic.LoadInt32(&calls))
}
