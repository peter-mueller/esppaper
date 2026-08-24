# CrowPanel ESP32 5.79inch E-paper with TinyGo

Set up your environment variables by editing the configuration file:

```bash
cp private_example.env private.env
# Inside private.env, uncomment and fill in your credentials
nano private.env
```

Flash the firmware onto your device and send a test image string:

```bash
sh flash.sh
curl -X POST http://crowpanel579:80/epaper --data-binary @demo.txt
```