# DCA Coinbase

Automated dollar cost averaging for BTC, LTC, BCH and ETH on Coinbase.
Origonal from https://github.com/sberserker/dcagdax
Added the following features
- support for coinbase with new advanced api


```
--type limit
--spread % to increase ask price to accommodate possible price fluctuation when order is placed. Default: 1
--fee % for exchange commission typically 0.1-0.5 for most exchanges depends on fee exchange and fee tier. Default 0.5
```
Limit order may spend a little less every purchase to accommodate spread and fee.
Unused portion will be left on exchange and included into a next order.
## Setup

You will need to set up environment variables for your API keys.

This can done using an .env file or a shell script, for example:
```agsl
#!/bin/bash

export COINBASE_SECRET="-----BEGIN EC PRIVATE KEY-----
<YOUR EC PRIVATE KEY>
-----END EC PRIVATE KEY-----"

export COINBASE_KEY="<YOUR EC ID"

export PORTFOLIO_ID="<YOUR PORTFOLIO ID>" // UUID that can be found in network request in developer tools when you clck on the profile in the left menu GET
	https://login.coinbase.com/api/v2/profile-menu

/<folder where dcagdax executable is found>/dcagdax "$@"
```


You will need to also sign into
[Coinbase](https://pro.coinbase.com/). Make sure you have a bank account linked to one of these for
ACH transfers.

Procure a Coinbase API key for yourself by visiting
[https://pro.coinbase.com/profile/api](https://pro.coinbase.com/profile/api). **Do not share
this API key with third parties!**

## Usage

Build the binary:

```
$ go build -o .  ./...
```

Then run it:

```
./dcagdax --help
usage: dcagdax --every=EVERY [<flags>]

Flags:
  --help                 Show context-sensitive help (also try --help-long and--help-man).
  --exchange="coinbase"  Exchange coinbase, gemini, ftx, ftxus. Default: coinbase
  --coin=BTC             Which coin you want to buy: BTC, LTC, BCH or ETH : percentage amount. Can be split between multipe coins. Total must be 100%. Example --coin BTC:70 --coin ETH:30
  --every=EVERY          How often to make purchases, e.g. 1h, 7d, 3w.
  --usd=USD              How much USD to spend on each purchase. If unspecified, the
                         minimum purchase amount allowed will be used.
  --currency="USD"       USD, EUR etc
  --until=UNTIL          Stop executing trades after this date, e.g. 2017-12-31.
  --after=AFTER          Start executing trades after this date, e.g. 2017-12-31.
  --trade                Actually execute trades.
  --autofund             Automatically initiate ACH deposits.
  --force                Force trade despite trading windows, will ask for user confirmation
  --type="market"        Order type market, limit. Default: market
  --spread=1.0           Percentage to add above ask price to get limit order executed. Default: 1.0
  --fee=0.5              Fee level to exclude from limit order amount. Default: 0.5
  --version              Show application version.
```

Be aware that if you set your purchase amount near 0.01 BTC (the minimum trade
amount) then an upswing in price might prevent you from trading.

## Run in Docker
The application can run in docker with cron.
Create env file with the following format
```
COINBASE_SECRET=secret
COINBASE_KEY=key

```
Adjust cron.conf as you wish. Note this will run the cointainer in foreground. To detach: Ctrl+P+Q
Timezone is optionalal -e TZ=... and added for convenience by default cron will run in UTC timezone
```
docker build -t dcagdax .
docker run -t -i --name dcagdax -e TZ=America/Los_Angeles  --env-file .env dcagdax
```

Run docker with automatic start
```
docker run -d --name dcagdax -e TZ=America/Los_Angeles  --env-file .env --restart unless-stopped dcagdax
```

Follow container output
```
docker logs dcagdax --follow
```


To stop and and remove
```
docker stop dcagdax
docker rm -f dcagdax
```

## FAQ

**Q:** Why do I not see any trading activity from the bot?

**A:** If you have other BTC trades on your account, the bot will detect that as a
cost-averaged purchase and wait until the next purchase window. This is for
people who want to "set it and forget it," not day traders!

**Q:** Why would I use this instead of Coinbase's recurring purchase feature?

**A:** The [fees on recurring
purchases](https://support.coinbase.com/customer/portal/articles/2109597)
(currently a minimum of $2.99 per purchase!) can add up quickly. This
side-steps those costs by automating free ACH deposits into your exchange
account & submitting market orders to exchange with BTC.

**Q:** How should I deploy this?

**A:** You could run this as a periodic cronjob on your workstation or run inside docker container or in the
cloud. Just be sure your API key & secret are not made available to anyone else
as part of your deployment!

**Q:** Which coins can I purchase?

**A:** We support all of Coinbase's products: BTC, LTC, BCH and ETH.


## Development references

### Unit tests
Generate mocks
```
go generate ./...
```
Run unit tests and get coverage
```
go test github.com/sberserker/dcagdax github.com/sberserker/dcagdax/exchanges  -coverprofile coverage.out
go tool cover -html=coverage.out -o coverage.html
```

