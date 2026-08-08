tinygo flash \
-target=./crowpanel579.json \
-serial=uart \
-ldflags="-X 'main.wlanSSID=$ESPPAPER_WLAN_SSID' \
          -X 'main.wlanPassphrase=$ESPPAPER_WLAN_PASSPHRASE'" \
-monitor \
.
