/*
 * ============================================================
 * SMART PET FEEDER - WOKWI SIMULATION ESP32 DevKit V1
 * Posisi Servo Default: Buka = 25, Tutup = 55
 *
 * Integrasi Backend:
 *  - POST /api/v1/sensors/feed-weight  -> kirim data berat & stok pakan
 *  - POST /api/v1/sensors/status       -> kirim status air
 *  - GET  /api/v1/devices/:id/commands/next       -> polling perintah manual
 *  - PATCH /api/v1/devices/:id/commands/:cid/status -> lapor hasil perintah
 *
 * Header Autentikasi: X-Device-Key: <DEVICE_API_KEY>
 * ============================================================
 */

// ===================== LIBRARY =====================
#include <WiFi.h>
#include <HTTPClient.h>
#include <WiFiClient.h>
#include <WiFiClientSecure.h>   // Diperlukan untuk HTTPS
#include <ArduinoJson.h>
#include <ESP32Servo.h>
#include <Preferences.h>

// Build Wokwi dengan:
// arduino-cli compile --fqbn esp32:esp32:esp32doit-devkit-v1 --output-dir .build .

// ===================== KONFIGURASI WiFi =====================
const char* WIFI_SSID     = "Wokwi-GUEST";
const char* WIFI_PASSWORD = "";

// ===================== KONFIGURASI BACKEND =====================
const char* BACKEND_URL    = "http://103.47.224.190:8001";
const char* DEVICE_API_KEY = "fY-XGzWSxyPe4a9IpMtWT5H1Ddb0tdpcuRkcirkuqa8";
const char* DEVICE_ID      = "ESP32-001";

// Simulasi mengikuti relay fisik Songle 1-channel: aktif LOW.
// OFF memakai mode INPUT agar GPIO tidak menarik optocoupler relay.
const bool RELAY_ACTIVE_LOW = true;

// ===================== PIN DEFINITION =====================
#define SERVO_PIN        18
#define RELAY_PIN        23
#define TRIG_PIN         26
#define ECHO_PIN         27
// ===================== PARAMETER OPERASIONAL =====================
const float PORSI_PAKAN_GRAM_DEFAULT = 50.0;
const float TARGET_AIR_GRAM       = 200.0;
const float TINGGI_WADAH_CM       = 13.1;

const float BATAS_MINIMAL_STOK_PAKAN_PERSEN = 10.0;

const unsigned long US_TIMEOUT          = 30000;
const unsigned long INTERVAL_CETAK      = 2000;   // cetak serial setiap 2 detik
const unsigned long INTERVAL_KIRIM_DATA = 10000;  // kirim ke backend setiap 10 detik
const unsigned long INTERVAL_POLLING    = 5000;   // polling perintah setiap 5 detik
const unsigned long INTERVAL_CONFIG     = 15000;  // ambil config device setiap 15 detik

const unsigned long FEED_ACTUATION_MS  = 1500;
const unsigned long WATER_ACTUATION_MS = 3000;
const unsigned long JEDA_OTOMATIS      = 60000;   // cooldown 1 menit
const unsigned long STARTUP_COMMAND_DRAIN_MS = 20000;
const bool AUTOMATION_ENABLED_DEFAULT = false;
const int SERVO_OPEN_DEG_DEFAULT = 25;
const int SERVO_CLOSED_DEG_DEFAULT = 55;
const int AUTOMATION_MAX_FAILURES = 3;

// ===================== OBJEK GLOBAL =====================
Servo   servoPakan;
Preferences preferences;

unsigned long lastPrint     = 0;
unsigned long lastKirimData = 0;
unsigned long lastPolling   = 0;
unsigned long lastConfig    = 0;

unsigned long waktuTerakhirPakan = 0;
long lastCompletedCommandID = 0;
bool automationEnabled = AUTOMATION_ENABLED_DEFAULT;
int autoFeedFailures = 0;
bool autoFeedLocked = false;
float porsiPakanGram = PORSI_PAKAN_GRAM_DEFAULT;
int servoOpenDeg = SERVO_OPEN_DEG_DEFAULT;
int servoClosedDeg = SERVO_CLOSED_DEG_DEFAULT;
String configuredSSID = WIFI_SSID;
String configuredPassword = WIFI_PASSWORD;

float simBeratPakan = 80.0;
float simBeratMinum = 120.0;
float simStokPakanPersen = 75.0;

// ===================== SETUP =====================
void setup() {
  Serial.begin(115200);
  delay(500);
  Serial.println();
  Serial.println(F("========================================"));

  preferences.begin("spfeeder", false);
  lastCompletedCommandID = preferences.getLong("last_ok", 0);
  muatConfigTersimpan();
  Serial.print(F("[INIT] Last completed command ID: "));
  Serial.println(lastCompletedCommandID);
  Serial.println(F("   SMART PET FEEDER - MANUAL SAFE MODE  "));
  Serial.println(F("   + Koneksi Backend                    "));
  Serial.println(F("========================================"));
  Serial.print(F("[INIT] Otomatisasi: "));
  Serial.println(automationEnabled ? F("ON") : F("OFF"));
  cetakConfigDevice();

  // Inisialisasi Servo
  servoPakan.attach(SERVO_PIN);
  servoPakan.write(servoClosedDeg);
  Serial.print(F("[INIT] Servo Pakan  -> GPIO 18 (TUTUP: "));
  Serial.print(servoClosedDeg);
  Serial.println(F(" deg)"));

  // Inisialisasi Relay (Pompa)
  matikanPompa();
  Serial.println(F("[INIT] Relay Pompa  -> GPIO 23 (MATI)"));

  // Inisialisasi Ultrasonik
  pinMode(TRIG_PIN, OUTPUT);
  pinMode(ECHO_PIN, INPUT);
  Serial.println(F("[INIT] Ultrasonik   -> Trig 26, Echo 27"));

  Serial.println(F("[INIT] Mode Wokwi: sensor memakai nilai virtual."));

  // Koneksi WiFi
  connectWiFi();

  // Bypass jeda awal agar sistem langsung aktif
  waktuTerakhirPakan = millis() - JEDA_OTOMATIS;

  Serial.println();
  Serial.println(F("[READY] Sistem siap & Berjalan."));
  Serial.println(F("========================================"));
}

void simpanCommandSelesai(long commandID) {
  if (commandID > lastCompletedCommandID) {
    lastCompletedCommandID = commandID;
    preferences.putLong("last_ok", lastCompletedCommandID);
  }
}

int clampServoDeg(int value, int fallback) {
  if (value < 0 || value > 180) return fallback;
  return value;
}

void muatConfigTersimpan() {
  configuredSSID = WIFI_SSID;
  configuredPassword = WIFI_PASSWORD;
  porsiPakanGram = preferences.getFloat("portion_g", PORSI_PAKAN_GRAM_DEFAULT);
  if (porsiPakanGram <= 0) porsiPakanGram = PORSI_PAKAN_GRAM_DEFAULT;
  servoOpenDeg = clampServoDeg(preferences.getInt("servo_open", SERVO_OPEN_DEG_DEFAULT), SERVO_OPEN_DEG_DEFAULT);
  servoClosedDeg = clampServoDeg(preferences.getInt("servo_closed", SERVO_CLOSED_DEG_DEFAULT), SERVO_CLOSED_DEG_DEFAULT);
  if (servoOpenDeg == servoClosedDeg) {
    servoOpenDeg = SERVO_OPEN_DEG_DEFAULT;
    servoClosedDeg = SERVO_CLOSED_DEG_DEFAULT;
  }
  automationEnabled = preferences.getBool("auto", AUTOMATION_ENABLED_DEFAULT);
}

void simpanConfigDevice() {
  preferences.putFloat("portion_g", porsiPakanGram);
  preferences.putInt("servo_open", servoOpenDeg);
  preferences.putInt("servo_closed", servoClosedDeg);
  preferences.putBool("auto", automationEnabled);
}

void resetAutomationFaults() {
  autoFeedFailures = 0;
  autoFeedLocked = false;
  Serial.println(F("[AUTO] Counter kegagalan otomatisasi direset."));
}

void catatHasilAutoPakan(bool berhasil) {
  if (berhasil) {
    autoFeedFailures = 0;
    return;
  }
  autoFeedFailures++;
  Serial.print(F("[AUTO] Gagal isi pakan otomatis berturut-turut: "));
  Serial.println(autoFeedFailures);
  if (autoFeedFailures >= AUTOMATION_MAX_FAILURES) {
    autoFeedLocked = true;
    Serial.println(F("[AUTO] Otomatisasi pakan dikunci. Matikan lalu nyalakan otomatisasi untuk mencoba lagi."));
  }
}

void cetakConfigDevice() {
  Serial.print(F("[CONFIG] Portion: "));
  Serial.print(porsiPakanGram, 1);
  Serial.print(F(" g, servo open: "));
  Serial.print(servoOpenDeg);
  Serial.print(F(" deg, closed: "));
  Serial.print(servoClosedDeg);
  Serial.print(F(" deg, auto: "));
  Serial.println(automationEnabled ? F("ON") : F("OFF"));
}

void nyalakanPompa() {
  if (RELAY_ACTIVE_LOW) {
    digitalWrite(RELAY_PIN, LOW);
    pinMode(RELAY_PIN, OUTPUT);
    return;
  }
  digitalWrite(RELAY_PIN, HIGH);
  pinMode(RELAY_PIN, OUTPUT);
}

void matikanPompa() {
  if (RELAY_ACTIVE_LOW) {
    pinMode(RELAY_PIN, INPUT);
    return;
  }
  digitalWrite(RELAY_PIN, LOW);
  pinMode(RELAY_PIN, OUTPUT);
}

// ===================== KONEKSI WiFi =====================
void connectWiFi() {
  Serial.print(F("[WiFi] Menghubungkan ke "));
  Serial.print(configuredSSID);
  WiFi.begin(configuredSSID.c_str(), configuredPassword.c_str(), 6);
  int retry = 0;
  while (WiFi.status() != WL_CONNECTED && retry < 30) {
    delay(500);
    Serial.print(F("."));
    retry++;
  }
  if (WiFi.status() == WL_CONNECTED) {
    Serial.println();
    Serial.print(F("[WiFi] Terhubung! IP: "));
    Serial.println(WiFi.localIP());
  } else {
    Serial.println();
    Serial.println(F("[WiFi] GAGAL terhubung! Mode offline aktif."));
  }
}

void pastikanWiFi() {
  if (WiFi.status() != WL_CONNECTED) {
    Serial.println(F("[WiFi] Koneksi terputus, mencoba ulang..."));
    connectWiFi();
  }
}

bool gunakanHTTPS() {
  return String(BACKEND_URL).startsWith("https://");
}

bool mulaiRequestHTTP(HTTPClient& http, WiFiClient& plainClient, WiFiClientSecure& secureClient, const String& url) {
  if (gunakanHTTPS()) {
    secureClient.setInsecure();  // Skip SSL certificate verification for development deployments.
    return http.begin(secureClient, url);
  }
  return http.begin(plainClient, url);
}

void pollingConfigDevice() {
  if (WiFi.status() != WL_CONNECTED) return;

  WiFiClient client;
  WiFiClientSecure secureClient;
  HTTPClient http;
  String url = String(BACKEND_URL) + "/api/v1/devices/" + String(DEVICE_ID) + "/config";
  if (!mulaiRequestHTTP(http, client, secureClient, url)) {
    Serial.println(F("[CONFIG] Gagal membuat koneksi config."));
    return;
  }
  http.addHeader("X-Device-Key", DEVICE_API_KEY);

  int statusCode = http.GET();
  if (statusCode != 200) {
    cetakErrorHTTP(F("[CONFIG] Polling config gagal"), statusCode, http);
    http.end();
    return;
  }

  String respBody = http.getString();
  http.end();

  StaticJsonDocument<768> doc;
  DeserializationError err = deserializeJson(doc, respBody);
  if (err) {
    Serial.println(F("[CONFIG] JSON parse error"));
    return;
  }

  float nextPortion = doc["manual_feed_portion_grams"] | porsiPakanGram;
  int nextOpen = doc["servo_open_degrees"] | servoOpenDeg;
  int nextClosed = doc["servo_closed_degrees"] | servoClosedDeg;
  bool nextAuto = doc["automation_enabled"] | automationEnabled;

  if (nextPortion <= 0) nextPortion = porsiPakanGram;
  nextOpen = clampServoDeg(nextOpen, servoOpenDeg);
  nextClosed = clampServoDeg(nextClosed, servoClosedDeg);
  if (nextOpen == nextClosed) {
    Serial.println(F("[CONFIG] Abaikan config servo invalid."));
    return;
  }

  bool changed = abs(nextPortion - porsiPakanGram) > 0.01 ||
                 nextOpen != servoOpenDeg ||
                 nextClosed != servoClosedDeg ||
                 nextAuto != automationEnabled;

  porsiPakanGram = nextPortion;
  servoOpenDeg = nextOpen;
  servoClosedDeg = nextClosed;
  bool automationChanged = nextAuto != automationEnabled;
  automationEnabled = nextAuto;
  if (automationChanged) {
    resetAutomationFaults();
  }
  simpanConfigDevice();
  servoPakan.write(servoClosedDeg);

  if (changed) {
    Serial.println(F("[CONFIG] Config device diperbarui dari backend."));
    cetakConfigDevice();
  }
}

// ===================== FUNGSI SENSOR ULTRASONIK =====================
float bacaJarakUltrasonik() {
  return TINGGI_WADAH_CM * (1.0 - (simStokPakanPersen / 100.0));
}

float hitungPersentaseStok() {
  float jarak = bacaJarakUltrasonik();
  if (jarak < 0) return -1.0;
  if (jarak > TINGGI_WADAH_CM) jarak = TINGGI_WADAH_CM;
  if (jarak < 0) jarak = 0;
  float persen = ((TINGGI_WADAH_CM - jarak) / TINGGI_WADAH_CM) * 100.0;
  return persen;
}

// ===================== FUNGSI BACA LOAD CELL =====================
float bacaBeratPakan() {
  return simBeratPakan;
}

float bacaBeratMinum() {
  return simBeratMinum;
}

bool tareSensor(String& errorMessage) {
  (void)errorMessage;
  simBeratPakan = 0.0;
  simBeratMinum = 0.0;
  Serial.println(F("[TARE] Sensor virtual direset."));
  return true;
}

// ===================== FUNGSI AKTUATOR =====================
bool bukaPakan(String& errorMessage) {
  (void)errorMessage;
  Serial.println(F("\n[ACT] MEMBERI PAKAN - MULAI"));
  servoPakan.write(servoOpenDeg);
  Serial.print(F("[ACT] Servo BUKA ("));
  Serial.print(servoOpenDeg);
  Serial.print(F(" deg) selama "));
  Serial.print(FEED_ACTUATION_MS);
  Serial.println(F(" ms"));
  delay(FEED_ACTUATION_MS);

  simBeratPakan += porsiPakanGram;
  float stokTurun = (porsiPakanGram / PORSI_PAKAN_GRAM_DEFAULT) * 16.0;
  if (simStokPakanPersen > stokTurun) {
    simStokPakanPersen -= stokTurun;
  } else {
    simStokPakanPersen = 0.0;
  }

  servoPakan.write(servoClosedDeg);
  Serial.print(F("[ACT] Servo TUTUP ("));
  Serial.print(servoClosedDeg);
  Serial.println(F(" deg)"));
  Serial.println(F("[DONE] Pengisian pakan selesai.\n"));
  return true;
}

bool testServoPakan(String& errorMessage) {
  (void)errorMessage;
  Serial.println(F("\n[ACT] TEST SERVO PAKAN"));
  servoPakan.write(servoOpenDeg);
  delay(700);
  servoPakan.write(servoClosedDeg);
  Serial.println(F("[DONE] Test servo selesai.\n"));
  return true;
}

bool isiAir(String& errorMessage) {
  (void)errorMessage;
  Serial.println(F("\n[ACT] MENGISI AIR - MULAI"));
  nyalakanPompa();
  Serial.print(F("[ACT] Pompa NYALA selama "));
  Serial.print(WATER_ACTUATION_MS);
  Serial.println(F(" ms"));
  delay(WATER_ACTUATION_MS);

  simBeratMinum += 60.0;
  if (simBeratMinum > TARGET_AIR_GRAM) {
    simBeratMinum = TARGET_AIR_GRAM;
  }

  matikanPompa();
  Serial.println(F("[ACT] Pompa MATI"));
  Serial.println(F("[DONE] Pengisian air selesai.\n"));
  return true;
}

void cetakErrorHTTP(const __FlashStringHelper* prefix, int statusCode, HTTPClient& http) {
  Serial.print(prefix);
  Serial.print(F(", kode: "));
  Serial.println(statusCode);

  String response = http.getString();
  response.trim();
  if (response.length() > 0) {
    if (response.length() > 180) response = response.substring(0, 180);
    Serial.print(F("[HTTP] Response: "));
    Serial.println(response);
  }
}

// ===================== KIRIM DATA SENSOR KE BACKEND =====================
/*
 * POST /api/v1/sensors/feed-weight
 * Header: X-Device-Key: <key>
 * Body: {
 *   "device_id": "ESP32-001",
 *   "weight_grams": 125.5,
 *   "food_stock_percent": 73.2,
 *   "water_available": true,
 *   "water_status": "available"
 * }
 */
void kirimDataSensor(float beratPakan, float beratMinum, float stokPersen) {
  (void)beratMinum;
  if (WiFi.status() != WL_CONNECTED) return;

  WiFiClient client;
  WiFiClientSecure secureClient;
  HTTPClient http;
  String url = String(BACKEND_URL) + "/api/v1/sensors/feed-weight";
  if (!mulaiRequestHTTP(http, client, secureClient, url)) {
    Serial.println(F("[HTTP] Gagal membuat koneksi sensor."));
    return;
  }
  http.addHeader("Content-Type", "application/json");
  http.addHeader("X-Device-Key", DEVICE_API_KEY);

  // Build JSON body
  StaticJsonDocument<768> doc;
  doc["device_id"]    = DEVICE_ID;
  doc["weight_grams"] = (beratPakan >= 0) ? beratPakan : 0.0;
  if (stokPersen >= 0) {
    doc["food_stock_percent"] = stokPersen;
  }
  doc["water_available"] = true;
  doc["water_status"]    = "available";

  String body;
  serializeJson(doc, body);

  int statusCode = http.POST(body);
  if (statusCode == 201) {
    Serial.println(F("[HTTP] Data sensor terkirim ke backend OK"));
  } else {
    cetakErrorHTTP(F("[HTTP] Gagal kirim sensor"), statusCode, http);
  }
  http.end();
}

// ===================== POLLING PERINTAH DARI BACKEND =====================
/*
 * GET /api/v1/devices/:deviceID/commands/next
 * Header: X-Device-Key: <key>
 * Response: { "data": { "id": 1, "action": "feed", "status": "sent" } }
 *           { "data": null }  <- tidak ada perintah baru
 */
void pollingPerintah() {
  if (WiFi.status() != WL_CONNECTED) return;

  WiFiClient client;
  WiFiClientSecure secureClient;
  HTTPClient http;
  String url = String(BACKEND_URL) + "/api/v1/devices/" + String(DEVICE_ID) + "/commands/next";
  if (!mulaiRequestHTTP(http, client, secureClient, url)) {
    Serial.println(F("[CMD] Gagal membuat koneksi polling."));
    return;
  }
  http.addHeader("X-Device-Key", DEVICE_API_KEY);

  int statusCode = http.GET();
  if (statusCode != 200) {
    cetakErrorHTTP(F("[CMD] Polling gagal"), statusCode, http);
    http.end();
    return;
  }

  String respBody = http.getString();
  http.end();

  StaticJsonDocument<768> doc;
  DeserializationError err = deserializeJson(doc, respBody);
  if (err) {
    Serial.println(F("[CMD] JSON parse error"));
    return;
  }

  // Cek apakah ada perintah
  JsonVariant data = doc["data"];
  if (data.isNull()) {
    // Tidak ada perintah baru
    return;
  }

  long   commandID = data["id"] | 0;
  String action    = data["action"] | "";

  if (commandID <= 0 || action.length() == 0) {
    Serial.println(F("[CMD] Perintah dari backend tidak valid."));
    return;
  }

  if (millis() < STARTUP_COMMAND_DRAIN_MS) {
    Serial.print(F("[CMD] Abaikan command startup yang sudah antre: "));
    Serial.print(action);
    Serial.print(F(" (ID: "));
    Serial.print(commandID);
    Serial.println(F(")"));
    laporkanHasilPerintah(commandID, false, "stale startup command ignored");
    return;
  }

  Serial.print(F("[CMD] Perintah diterima: "));
  Serial.print(action);
  Serial.print(F(" (ID: "));
  Serial.print(commandID);
  Serial.println(F(")"));

  // Eksekusi perintah
  bool berhasil = false;
  String errorMessage = "";
  if (action == "feed") {
    if (commandID == lastCompletedCommandID) {
      Serial.println(F("[CMD] Command sudah pernah selesai, kirim ulang status saja."));
      berhasil = true;
    } else {
      berhasil = bukaPakan(errorMessage);
      if (berhasil) {
        waktuTerakhirPakan = millis();
        simpanCommandSelesai(commandID);
      }
    }
  } else if (action == "servo_test") {
    berhasil = testServoPakan(errorMessage);
  } else if (action == "tare") {
    berhasil = tareSensor(errorMessage);
    if (berhasil) {
      simpanCommandSelesai(commandID);
    }
  } else if (action == "drink") {
    if (commandID == lastCompletedCommandID) {
      Serial.println(F("[CMD] Command sudah pernah selesai, kirim ulang status saja."));
      berhasil = true;
    } else {
      berhasil = isiAir(errorMessage);
      if (berhasil) {
        simpanCommandSelesai(commandID);
      }
    }
  } else {
    Serial.print(F("[CMD] Aksi tidak dikenal: "));
    Serial.println(action);
    errorMessage = "aksi tidak dikenal";
  }

  // Laporkan status ke backend
  laporkanHasilPerintah(commandID, berhasil, errorMessage);
}

// ===================== LAPOR HASIL PERINTAH =====================
/*
 * PATCH /api/v1/devices/:deviceID/commands/:commandID/status
 * Header: X-Device-Key: <key>
 * Body: { "status": "completed" }  atau  { "status": "failed" }
 */
void laporkanHasilPerintah(long commandID, bool berhasil, const String& errorMessage) {
  if (WiFi.status() != WL_CONNECTED) return;

  WiFiClient client;
  WiFiClientSecure secureClient;
  HTTPClient http;
  String url = String(BACKEND_URL) + "/api/v1/devices/" + String(DEVICE_ID)
               + "/commands/" + String(commandID) + "/status";
  if (!mulaiRequestHTTP(http, client, secureClient, url)) {
    Serial.println(F("[CMD] Gagal membuat koneksi laporan status."));
    return;
  }
  http.addHeader("Content-Type", "application/json");
  http.addHeader("X-Device-Key", DEVICE_API_KEY);

  StaticJsonDocument<192> doc;
  doc["status"] = berhasil ? "completed" : "failed";
  if (!berhasil && errorMessage.length() > 0) {
    doc["error"] = errorMessage;
  }
  String body;
  serializeJson(doc, body);

  int statusCode = http.PATCH(body);
  if (statusCode == 200) {
    Serial.print(F("[CMD] Status perintah dilaporkan: "));
    Serial.println(berhasil ? F("completed OK") : F("failed"));
  } else {
    cetakErrorHTTP(F("[CMD] Gagal laporkan status"), statusCode, http);
  }
  http.end();
}

// ===================== CETAK DATA SERIAL =====================
void cetakDataSensor(float bPakan, float bAir, float sPersen) {
  Serial.println(F("+----------------------------------------+"));
  Serial.print(F("| Berat Pakan : ")); Serial.print(bPakan >= 0 ? bPakan : 0.0, 1); Serial.println(F(" g"));
  Serial.print(F("| Berat Air   : ")); Serial.print(bAir   >= 0 ? bAir   : 0.0, 1); Serial.println(F(" g"));
  Serial.print(F("| Stok Pakan  : ")); Serial.print(sPersen >= 0 ? sPersen : 0.0, 1); Serial.println(F(" %"));
  Serial.print(F("| WiFi        : ")); Serial.println(WiFi.status() == WL_CONNECTED ? F("Terhubung") : F("Terputus"));
  Serial.print(F("| Otomatisasi : ")); Serial.println(automationEnabled ? F("ON") : F("OFF"));
  Serial.println(F("+----------------------------------------+\n"));
}

// ===================== LOOP UTAMA =====================
void loop() {
  // Kontrol Manual via Serial Monitor
  if (Serial.available() > 0) {
    char cmd = Serial.read();
    String errorMessage = "";
    switch (cmd) {
      case 'P': case 'p':
        if (bukaPakan(errorMessage)) {
          waktuTerakhirPakan = millis();
        }
        break;
      case 'A': case 'a':
        isiAir(errorMessage);
        break;
      case 'T': case 't':
        tareSensor(errorMessage);
        break;
      case 'W': case 'w':
        pastikanWiFi();
        break;
      case 'O': case 'o':
        automationEnabled = !automationEnabled;
        resetAutomationFaults();
        simpanConfigDevice();
        Serial.print(F("[AUTO] Otomatisasi sekarang: "));
        Serial.println(automationEnabled ? F("ON") : F("OFF"));
        break;
      case 'C': case 'c':
        cetakConfigDevice();
        break;
      case '[':
        servoOpenDeg = max(0, servoOpenDeg - 1);
        cetakConfigDevice();
        break;
      case ']':
        servoOpenDeg = min(180, servoOpenDeg + 1);
        cetakConfigDevice();
        break;
      case 'V': case 'v':
        testServoPakan(errorMessage);
        break;
      case 'S': case 's':
        simpanConfigDevice();
        Serial.println(F("[CONFIG] Config lokal disimpan."));
        break;
      case 'R': case 'r':
        porsiPakanGram = PORSI_PAKAN_GRAM_DEFAULT;
        servoOpenDeg = SERVO_OPEN_DEG_DEFAULT;
        servoClosedDeg = SERVO_CLOSED_DEG_DEFAULT;
        automationEnabled = AUTOMATION_ENABLED_DEFAULT;
        resetAutomationFaults();
        simpanConfigDevice();
        servoPakan.write(servoClosedDeg);
        Serial.println(F("[CONFIG] Config lokal direset ke default."));
        cetakConfigDevice();
        break;
    }
  }

  // Polling Perintah dari Backend diprioritaskan sebelum otomatisasi
  if (millis() - lastPolling >= INTERVAL_POLLING) {
    lastPolling = millis();
    pastikanWiFi();
    pollingPerintah();
  }

  if (millis() - lastConfig >= INTERVAL_CONFIG) {
    lastConfig = millis();
    pollingConfigDevice();
  }

  // Baca Sensor
  float beratPakan = bacaBeratPakan();
  float beratMinum = bacaBeratMinum();
  float stokPersen = hitungPersentaseStok();

  // LOGIKA OTOMATIS PAKAN BERDASARKAN ULTRASONIK
  if (automationEnabled && !autoFeedLocked &&
      stokPersen >= 0 && stokPersen <= BATAS_MINIMAL_STOK_PAKAN_PERSEN) {
    if (millis() - waktuTerakhirPakan >= JEDA_OTOMATIS) {
      Serial.println(F("[AUTO] Stok pakan rendah berdasarkan ultrasonik!"));
      String errorMessage = "";
      waktuTerakhirPakan = millis();
      bool berhasil = bukaPakan(errorMessage);
      catatHasilAutoPakan(berhasil);
    }
  }

  // Kirim Data ke Backend (setiap INTERVAL_KIRIM_DATA ms)
  if (millis() - lastKirimData >= INTERVAL_KIRIM_DATA) {
    lastKirimData = millis();
    pastikanWiFi();
    kirimDataSensor(beratPakan, beratMinum, stokPersen);
  }

  // Cetak Log Serial (setiap INTERVAL_CETAK ms)
  if (millis() - lastPrint >= INTERVAL_CETAK) {
    lastPrint = millis();
    cetakDataSensor(beratPakan, beratMinum, stokPersen);
  }
}
