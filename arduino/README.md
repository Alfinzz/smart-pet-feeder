# SmartPetFeeder ESP32 Setup

## Windows Arduino IDE

1. Install Arduino IDE 2.x.
2. Add ESP32 board support from Boards Manager and select `ESP32 Dev Module` or the matching ESP32 DevKit board.
3. Install these libraries from Library Manager:
   - `ArduinoJson`
   - `ESP32Servo`
   - `HX711`
4. Copy `arduino_secrets.h.example` to `arduino_secrets.h` in this same folder.
5. Fill `WIFI_SSID`, `WIFI_PASSWORD`, `BACKEND_URL`, `DEVICE_API_KEY`, and `DEVICE_ID`.
6. Open `SmartPetFeeder_IoT.ino`, choose the ESP32 port, then upload.

`arduino_secrets.h` is ignored by git so WiFi credentials and the device API key are not committed.
