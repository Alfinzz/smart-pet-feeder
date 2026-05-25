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

### Physical power and relay notes

The physical sketch assumes a 1-channel Songle SRD-05VDC-SL-C relay module that is active LOW on GPIO 23. To keep the pump off, the firmware releases GPIO 23 as `INPUT`; to turn it on, the firmware drives GPIO 23 LOW.

Use the ESP32 only as the controller signal source. Do not power the Tower-Pro MG996R servo or the 3V-5V mini submersible pump from an ESP32 GPIO, 3V3 pin, or weak USB rail. Use a stable external 5V supply or LM2596 step-down output sized for the servo, relay module, and pump, then connect the external supply GND to ESP32 GND.

Wire the pump through the relay contact as normally-off: supply positive -> relay `COM`, relay `NO` -> pump positive, pump negative -> supply negative. If the pump is wired through `NC`, it will run at power-on until the relay is energized.

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

Manual feed and water refill commands drive the servo and pump by fixed time, so they do not require load cells to be ready. Load cells are only used for telemetry and `tare` calibration. Keep automatic mode OFF until the ultrasonic food-stock reading is stable. The dashboard food stock comes from the ultrasonic sensor percentage. The dashboard water status is always reported as available for demo stability.

## Wokwi VS Code Simulation

Open `SmartPetFeeder_Wokwi/` in VS Code with the Wokwi extension installed. The folder contains its own `diagram.json` and `wokwi.toml`.

The Wokwi sketch uses:

- WiFi SSID: `Wokwi-GUEST`
- Backend: `http://103.47.224.190:8001`
- Device ID: `ESP32-001`
- Relay simulation: active LOW, matching the physical relay control logic
- Virtual food stock percentage driven by the simulated ultrasonic reading
- Water status always reported as available
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
