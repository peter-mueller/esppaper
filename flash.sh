. ./private.env
tinygo flash \
-target=xiao-esp32s3 \
-baudrate=115200 \
-serial=uart \
-ldflags="-X 'main.wlanSSID=$ESPPAPER_WLAN_SSID' \
          -X 'main.wlanPassphrase=$ESPPAPER_WLAN_PASSPHRASE'
          -X 'main.hostname=$ESPPAPER_HOSTNAME'" \
-monitor \
.
