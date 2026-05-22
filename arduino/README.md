# SmartPetFeeder ESP32 Setup

Arduino code is split into two independent sketches:

- `SmartPetFeeder_Physical/`: upload this to the real ESP32 device.
- `SmartPetFeeder_Wokwi/`: use this only for Wokwi simulation.

## Required Libraries

Install ESP32 board support and these libraries from Arduino IDE Library Manager:

- `ArduinoJson`
- `ESP32Servo`
- `HX711` for `SmartPetFeeder_Physical`

## Physical ESP32

Open `SmartPetFeeder_Physical/SmartPetFeeder_Physical.ino` in Arduino IDE.

Before upload, check these values in the configuration section:

- `WIFI_SSID`
- `WIFI_PASSWORD`
- `BACKEND_URL`
- `DEVICE_API_KEY`
- `DEVICE_ID`
- `RELAY_ACTIVE_LOW`

Compile with Arduino CLI:

```sh
arduino-cli compile \
  --fqbn esp32:esp32:esp32doit-devkit-v1 \
  arduino/SmartPetFeeder_Physical
```

Initial physical-device test sequence:

- `v`: test feed servo open/close using current configured degrees.
- `p`: run feed dispensing.
- `a`: run water refill.
- `t`: tare both load cells.
- `w`: reconnect WiFi.
- `o`: toggle automatic feed/water mode.
- `[` and `]`: adjust servo open degree by 1 degree.
- `s`: save local config.
- `r`: reset local config to defaults.

Keep automatic mode OFF until the real load cells are calibrated and stable. If a load cell reads near zero after dispensing, automatic mode can repeatedly feed or refill every cooldown cycle.

## Wokwi VS Code Simulation

Open `SmartPetFeeder_Wokwi/` in VS Code with the Wokwi extension installed. The folder contains its own `diagram.json` and `wokwi.toml`.

The Wokwi sketch uses:

- WiFi SSID: `Wokwi-GUEST`
- Backend: `http://103.47.224.190:8001`
- Device ID: `ESP32-001`
- Virtual feed and water values that increase when commands run
- Automation OFF by default

Compile Wokwi firmware with Arduino CLI:

```sh
arduino-cli compile \
  --fqbn esp32:esp32:esp32doit-devkit-v1 \
  --output-dir arduino/SmartPetFeeder_Wokwi/.build \
  arduino/SmartPetFeeder_Wokwi
```

After starting the simulator, wait about 20 seconds before pressing buttons in the Flutter app. Commands queued before boot are reported as failed with `stale startup command ignored`, so they do not move the servo or pump.

Use the Flutter app against `http://103.47.224.190:8001/api/v1`, log in as the demo owner, then press `Feed Now`, `Refill Water`, or `Save & Test Servo`. The serial monitor should show the command being polled, executed, and reported as `completed`.
