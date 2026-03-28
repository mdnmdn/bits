# Renderers

## Overview

Every `bits` command renders a `model.Response[T]` value. The output format is selected with the global `-o` flag (default: `table`). All formats include **provenance** — the provider and market that actually served the response.

Available formats: `table`, `json`, `yaml`, `toon`, `markdown`.

---

## Provenance envelope

All structured formats (json, yaml, toon, markdown) wrap the data in an envelope:

| Field          | Key            | Present when         |
|----------------|----------------|----------------------|
| Kind           | `kind`         | always               |
| Data           | `data`         | always               |
| Provider       | `provider`     | always               |
| Market         | `mkt`          | always (omitted if empty in yaml/toon) |
| Fallback flag  | `fallback`     | only when true       |
| Orig. provider | `req_provider` | only when fallback   |
| Orig. market   | `req_mkt`      | only when fallback   |
| Item errors    | `errors`       | only when non-empty  |
| Metadata       | `metadata`     | only when non-nil    |

`errors` entries: `{ sym, err }` — per-symbol failures from batch/fan-out calls.

The table format shows provenance as a `provider: <label>` header line above the table body, with an optional fallback footnote below.

---

## Field name conventions

All serialized field names are **lowercase and compact**. Key shorthands:

| Concept             | Key        |
|---------------------|------------|
| market type         | `mkt`      |
| symbol              | `sym`      |
| currency            | `cur`      |
| open / high / low / close | `o` / `h` / `l` / `c` |
| volume              | `v` / `vol` |
| open time / close time | `ot` / `ct` |
| price change %      | `chg_pct`  |
| weighted avg price  | `vwap`     |
| market cap          | `mcap`     |
| market cap rank     | `rank`     |
| price/qty precision | `pp` / `qp` |
| min/max price       | `min_p` / `max_p` |
| min/max qty         | `min_q` / `max_q` |
| step size           | `step`     |
| latency / skew      | `lat` / `skew` |

Optional/pointer fields use `omitempty` and are omitted when absent.

---

## Format specifications

### `table` (default)

Human-readable aligned text via `text/tabwriter`.

**Structure:**
```
provider: <provider>[/<market>]

<column headers>
<rows...>

† served by <provider> (requested: <req_provider>)   ← only on fallback
```

**Per-command columns:**

| Command    | Columns |
|------------|---------|
| `price`    | ID, SYMBOL, PRICE, CURRENCY, CHANGE 24H |
| `ticker`   | SYMBOL, MARKET, LAST, CHANGE%, HIGH, LOW, VOLUME |
| `book`     | BIDS (PRICE, QTY) / ASKS (PRICE, QTY) side-by-side |
| `candles`  | OPEN TIME, OPEN, HIGH, LOW, CLOSE, VOLUME |
| `markets`  | #, ID, SYMBOL, NAME, PRICE, CHANGE 24H, MKT CAP, VOLUME 24H |
| `info`     | Exchange, Market, Time, Symbols count + SYMBOL, BASE, QUOTE, STATUS |
| `time`     | Server Time, Local Time, Latency, Clock Skew (key-value) |

Missing optional values render as `-`.

---

### `json`

Pretty-printed JSON (2-space indent) with the full provenance envelope.

```json
{
  "kind": "ticker",
  "data": { ... },
  "provider": "binance",
  "mkt": "spot",
  "fallback": true,
  "req_provider": "coingecko",
  "req_mkt": "spot",
  "errors": [{ "sym": "XYZUSDT", "err": "symbol not found" }]
}
```

---

### `yaml`

YAML (2-space indent) with the full provenance envelope.

```yaml
kind: ticker
data:
  ...
provider: binance
mkt: spot
# fallback, req_provider, req_mkt, errors, metadata omitted when absent
```

---

### `toon`

[TOON](https://github.com/toon-format/toon-go) (Token-Oriented Object Notation) with the full provenance envelope. Compact format optimised for LLM consumption.

```
kind: ticker
data{...}:
  ...
provider: binance
mkt: spot
```

---

### `markdown`

Markdown document with a level-1 heading showing `provider[/market]`, an optional fallback blockquote, and a fenced `yaml` code block containing the full provenance envelope.

```markdown
# binance/spot

> † served by binance (requested: coingecko)   ← only on fallback

​```yaml
kind: ticker
data:
  ...
provider: binance
mkt: spot
​```
```

---

## Implementation

| Format    | Package                         | Entry point              |
|-----------|---------------------------------|--------------------------|
| `table`   | `internal/render/table/`        | `Render<Type>(w, res)`   |
| `json`    | `internal/render/json/`         | `Render[T](w, res)`      |
| `yaml`    | `internal/render/yaml/`         | `Render[T](w, res)`      |
| `toon`    | `internal/render/toon/`         | `Render[T](w, res)`      |
| `markdown`| `internal/render/markdown/`     | `Render[T](w, res)`      |

Shared helpers in `internal/render/`:
- `ProviderLabel(res)` — returns `"provider"` or `"provider/market"`
- `FallbackFootnote(res)` — returns fallback note string, or `""`

Table-specific helpers in `internal/render/table/provenance.go`:
- `printHeader(w, res)` — writes `provider: <label>\n\n`
- `printFooter(w, res)` — writes fallback footnote if present
