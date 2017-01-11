#/usr/bin/env bash

# start ticker
export PTICKER_DEVICE="/dev/ttyAMA0"
export PTICKER_PITCH_URL="https://buzzer-ws.appspot.com/" 
export PTICKER_PITCH_CHECK_INTERVAL=60

exec $(dirname $0)/pticker &
