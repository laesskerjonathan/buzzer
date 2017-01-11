#/usr/bin/env bash

# choose display
export DISPLAY=:0

# disable screensaver
xset s noblank
xset s off
xset s -dpms

# start buzzer
export PBUZZER_KEYPAD_DEVICE="HID 04d9:1203"
export PBUZZER_PITCH_URL="https://buzzer-ws.appspot.com/" 
export PBUZZER_PITCH_CHECK_INTERVAL=60

exec $(dirname $0)/pbuzzer &
