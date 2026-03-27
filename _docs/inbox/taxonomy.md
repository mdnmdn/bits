### **Crypto Exchange API Support Taxonomy**

#### **1. REST API (Request-Response)**

* **Base / System (Public)**
    * `Time`: Synchronizing local clock with exchange (vital for signed requests).
    * `Exchange Info`: The "Source of Truth."
        * **Symbols**: Pairs (BTCUSDT), Status (Trading/Break/Halt).
        * **Filters**: Min/Max price, Lot sizes, Step size (precision).
        * **Commissions**: Maker/Taker default tiers.
* **Market Data (Public)**
    * `Order Book Snapshots`: Static view of the $L1$ (Top), $L2$ (Depth), or $L3$ (Orders) book.
    * `Candlesticks (OHLCV)`: Historical bars from 1m to 1M intervals.
    * `Recent Trades`: The last $n$ filled orders on the tape.
    * `Ticker 24h`: Rolling 24-hour statistics (Volume, High, Low).
* **Account & Trading (Private - Requires API Keys)**
    * `Balances`: Spot, Margin, or Futures wallet snapshots.
    * `Order Management`: Create, Cancel, and Batch-cancel orders.
    * `Trade History`: Personal fills and fee audits.

#### **2. WebSocket (Event-Driven / Streaming)**
*Used for high-frequency updates and real-time state management.*

* **Market Streams (Public)**
    * `Aggregated Trade`: Every single fill on the market.
    * `Diff. Depth (Order Book)`: Incremental updates (deltas) to keep a local book synced without re-downloading the whole thing.
    * `Book Ticker`: Real-time Best Bid and Offer (BBO). The fastest way to track price.
    * `Mini Ticker`: High-frequency $24h$ stats (usually updated every 1s).
* **User Data Streams (Private - Requires Listen Key)**
    * `Account Update`: Instant notification of balance changes (e.g., after a trade or deposit).
    * `Order Update`: Real-time status changes (New -> Partially Filled -> Filled/Canceled).

---
