#/usr/bin/env bash

# start ticker
export TICKER_DEVICE="/dev/ttyAMA0"
export TICKER_PITCH_URL="https://buzzer-ws.appspot.com/" 
export TICKER_PITCH_CHECK_INTERVAL=60

exec $(dirname $0)/ticker &
