./_tools/test_capabilities.py
\Testing provider: binance
  - server_time (spot)... OK
  - exchange_info (spot)... OK
  - exchange_info (futures)... OK
  - price (spot)... OK
  - price (futures)... OK
  - price (margin)... OK
  - candles (spot)... OK
  - candles (futures)... OK
  - candles (margin)... OK
  - ticker_24h (spot)... OK
  - ticker_24h (futures)... OK
  - ticker_24h (margin)... OK
  - order_book (spot)... OK
  - order_book (futures)... OK
  - stream_price (spot)... FAILED: No stream data received
  - stream_order_book (spot)... OK
  - stream_order_book (futures)... OK
Testing provider: bitget
  - server_time (spot)... OK
  - exchange_info (spot)... OK
  - exchange_info (futures)... OK
  - price (spot)... OK
  - candles (spot)... OK
  - candles (futures)... OK
  - ticker_24h (spot)... OK
  - ticker_24h (futures)... OK
  - order_book (spot)... OK
  - order_book (futures)... FAILED: API error: Request URL NOT FOUND
  - stream_price (spot)... OK
  - stream_price (futures)... OK
  - stream_order_book (spot)... OK
  - stream_order_book (futures)... OK
Testing provider: coingecko
  - price (spot)... OK
  - candles (spot)... FAILED: API error 400: {"error":"Invalid days parameter"}
  - markets_list (spot)... OK
  - stream_price (spot)... FAILED: SKIPPED (plan restricted: requires paid CoinGecko API key)
Testing provider: cryptocom
  - server_time (spot)... OK
  - server_time (futures)... OK
  - exchange_info (spot)... OK
  - exchange_info (futures)... OK
  - exchange_info (margin)... OK
  - price (spot)... OK
  - price (futures)... OK
  - price (margin)... OK
  - ticker_24h (spot)... OK
  - ticker_24h (futures)... OK
  - ticker_24h (margin)... OK
  - order_book (spot)... OK
  - order_book (futures)... OK
  - order_book (margin)... OK
  - stream_price (spot)... FAILED: No stream data received
  - stream_order_book (spot)... FAILED: No stream data received
Testing provider: mexc
  - server_time (spot)... OK
  - server_time (futures)... OK
  - server_time (margin)... OK
  - exchange_info (spot)... FAILED: json: cannot unmarshal string into Go struct field mexcSpotSymbol.symbols.baseSizePrecision of type int
  - exchange_info (futures)... OK
  - exchange_info (margin)... FAILED: json: cannot unmarshal string into Go struct field mexcSpotSymbol.symbols.baseSizePrecision of type int
  - price (spot)... OK
  - price (futures)... OK
  - price (margin)... OK
  - candles (spot)... OK
  - candles (futures)... OK
  - candles (margin)... OK
  - ticker_24h (spot)... OK
  - ticker_24h (futures)... OK
  - ticker_24h (margin)... OK
  - order_book (spot)... OK
  - order_book (futures)... OK
  - order_book (margin)... OK
Testing provider: whitebit
  - server_time (spot)... OK
  - exchange_info (spot)... OK
  - exchange_info (futures)... OK
  - price (spot)... OK
  - candles (spot)... OK
  - candles (futures)... OK
  - ticker_24h (spot)... OK
  - ticker_24h (futures)... OK
  - order_book (spot)... OK
  - order_book (futures)... OK
  - stream_price (spot)... OK
  - stream_order_book (spot)... OK
  - stream_order_book (futures)... OK

========================================
SUMMARY
========================================
binance         | OK: 16 | FAILED:  1 | SKIPPED:  0
bitget          | OK: 13 | FAILED:  1 | SKIPPED:  0
coingecko       | OK:  2 | FAILED:  1 | SKIPPED:  1
cryptocom       | OK: 14 | FAILED:  2 | SKIPPED:  0
mexc            | OK: 16 | FAILED:  2 | SKIPPED:  0
whitebit        | OK: 13 | FAILED:  0 | SKIPPED:  0
----------------------------------------
TOTAL           | OK: 74 | FAILED:  7 | SKIPPED:  1
