# SmartPetFeeder ESP32 Setup

## Windows Arduino IDE

1. Install Arduino IDE 2.x.
2. Add ESP32 board support from Boards Manager and select `ESP32 Dev Module` or the matching ESP32 DevKit board.
3. Install these libraries from Library Manager:
   - `ArduinoJson`
   - `ESP32Servo`
   - `HX711`
4. Open `SmartPetFeeder_IoT.ino`.
5. Fill or adjust `WIFI_SSID`, `WIFI_PASSWORD`, `BACKEND_URL`, `DEVICE_API_KEY`, `DEVICE_ID`, and `RELAY_ACTIVE_LOW` in the configuration section.
6. Choose the ESP32 port, then upload.
