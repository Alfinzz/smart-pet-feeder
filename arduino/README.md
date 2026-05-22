# SmartPetFeeder ESP32 Setup

## Windows Arduino IDE

1. Install Arduino IDE 2.x.
2. Add ESP32 board support from Boards Manager and select `ESP32 Dev Module` or the matching ESP32 DevKit board.
3. Install these libraries from Library Manager:
   - `ArduinoJson`
   - `ESP32Servo`
   - `HX711`
4. Open `SmartPetFeeder_IoT/SmartPetFeeder_IoT.ino`.
5. Fill or adjust `WIFI_SSID`, `WIFI_PASSWORD`, `BACKEND_URL`, `DEVICE_API_KEY`, `DEVICE_ID`, and `RELAY_ACTIVE_LOW` in the configuration section.
6. Choose the ESP32 port, then upload.

## Wokwi VS Code Simulation

The simulator profile is selected with `-DWOKWI_SIMULATION=1`. It uses:

- WiFi SSID: `Wokwi-GUEST`
- Backend: `http://103.47.224.190:8001`
- Device ID: `ESP32-001`
- Virtual sensor values that increase when feed or water commands run
- Automation is OFF by default so the feeder only runs from manual commands

Install Arduino CLI, the ESP32 core, and the required libraries, then build from the `arduino/SmartPetFeeder_IoT/` directory:

```sh
arduino-cli compile \
  --fqbn esp32:esp32:esp32doit-devkit-v1 \
  --output-dir .build \
  --build-property build.extra_flags="-DWOKWI_SIMULATION=1" \
  .
```

Open `arduino/SmartPetFeeder_IoT/` in VS Code with the Wokwi extension installed, then start the simulation. The extension reads `wokwi.toml` and `diagram.json`.

After starting the simulator, wait about 20 seconds before pressing buttons in the Flutter app. Commands that were already queued before the device booted are reported as failed with `stale startup command ignored`, so they do not move the servo or pump.

Use the Flutter app against `http://103.47.224.190:8001/api/v1`, log in as the demo owner, then press `Feed Now` or `Refill Water`. The serial monitor should show the new command being polled, executed, and reported as `completed`.

## Real ESP32 Upload

Build or upload without `-DWOKWI_SIMULATION=1`. Before uploading at your friend's house, update `WIFI_SSID` and `WIFI_PASSWORD`, then test these serial commands first:

- `p`: open the feed servo
- `a`: turn the water pump relay on until the virtual/real target is reached
- `t`: tare both load cells
- `w`: reconnect WiFi
- `o`: toggle automatic feed/water mode

Keep automatic mode OFF until the real load cells are calibrated and verified. If a load cell reads near zero after dispensing, automatic mode can repeatedly feed or refill every cooldown cycle.
