#!/bin/bash

output=(
"video/brightnessdown BRTDN 00000087 00000000"
"video/brightnessup BRTUP 00000086 00000000"
"video/brightnessdown BRTDN 00000087 00000000"
"video/brightnessup BRTUP 00000086 00000000"
"button/lid LID close"
"button/lid LID open"
"video/brightnessdown BRTDN 00000087 00000000"
"wmi PNP0C14:05 000000d0 00000000"
"video/brightnessup BRTUP 00000086 00000000"
"wmi PNP0C14:05 000000d0 00000000"
"video/brightnessdown BRTDN 00000087 00000000"
"wmi PNP0C14:05 000000d0 00000000"
)

for line in "${output[@]}"
do
  sleep 0.1
  echo "$line"
done

sleep 10000000000000