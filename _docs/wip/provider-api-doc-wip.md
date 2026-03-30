# GOAL

Create a comphrehensive provider api documentation.

explore the current implementation and the exchange online docs then write in 

_docs/providers/apis/
  - provider-index.md - write the pro
  - bitget/
     - bitget-general.md - describe the general info as base url/ rate limits, ping time, auth info, exchange info apis
     - bitget-market.md - apis for market data as prices, orderbooks, candles, tickers
  - ...
    
  
  The documentation should have the same structure for all the providers:
  
  - general info
  - api list
  - for each apis
    - description (ed eventually quirks and notes)
    - url (complete)
    - parameters
    - return values
    - samples of parameters and return values
    
wherevere possible add links to the online documentation, for the perform the calls run with subagents (with some head limits) and write some "extract" of the values

fell free to curl to the provider, ed eventually create a small python (or shelll) program with a simple menu that invokes the api and show the result for each, maybe with different example if has different parameter.
Eventually you can create a single program (for all the provider) and a yaml/toml config file for each provider with all the sample invocations

so you could:
- explore the source code in the ./pkg/provider/*
- explore the online doc for each exchange
- try with curl to build this documentation

use the provider-index.md file to mantain the memory progress of the exploration and the target structure of the documentation in order to maintain coherence.

use subagents to delegate work and free context
