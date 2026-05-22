/*
 * ============================================================
 * SMART PET FEEDER — ESP32 DevKit V1 (FULL OTOMATIS + BACKEND)
 * Posisi Servo: Buka = 0, Tutup = 55
 *
 * Integrasi Backend:
 *  - POST /api/v1/sensors/feed-weight  → kirim data berat & stok pakan
 *  - POST /api/v1/sensors/status       → kirim status air
 *  - GET  /api/v1/devices/:id/commands/next       → polling perintah manual
 *  - PATCH /api/v1/devices/:id/commands/:cid/status → lapor hasil perintah
 *
 * Header Autentikasi: X-Device-Key: <DEVICE_API_KEY>
 * ============================================================
 */

// ===================== LIBRARY =====================
#include <WiFi.h>
#include <HTTPClient.h>
#include <WiFiClientSecure.h>   // Diperlukan untuk HTTPS
#include <ArduinoJson.h>
#include <ESP32Servo.h>
#include <HX711.h>
#include <Preferences.h>

// ===================== KONFIGURASI WiFi =====================
const char* WIFI_SSID     = "Finzz";
const char* WIFI_PASSWORD = "veteranpride";

// ===================== KONFIGURASI BACKEND =====================
const char* BACKEND_URL    = "https://smart-pet-feeder.alfian-gading.my.id";
const char* DEVICE_API_KEY = "fY-XGzWSxyPe4a9IpMtWT5H1Ddb0tdpcuRkcirkuqa8";
const char* DEVICE_ID      = "ESP32-001";

// Kebanyakan relay module aktif saat pin LOW. Ubah ke false jika relay aktif saat HIGH.
const bool RELAY_ACTIVE_LOW = true;

// ===================== PIN DEFINITION =====================
#define SERVO_PIN        18
#define RELAY_PIN        23
#define TRIG_PIN         26
#define ECHO_PIN         27
#define LC_PAKAN_DT      13
#define LC_PAKAN_SCK     14
#define LC_MINUM_DT       4
#define LC_MINUM_SCK     16

// ===================== PARAMETER KALIBRASI =====================
float calibration_factor_pakan = 1043.13;
float calibration_factor_minum = 223.41;

// ===================== PARAMETER OPERASIONAL =====================
const float PORSI_PAKAN_GRAM      = 50.0;
const float TARGET_AIR_GRAM       = 200.0;
const float TINGGI_WADAH_CM       = 13.1;

const float BATAS_MINIMAL_PAKAN   = 10.0;
const float BATAS_MINIMAL_AIR     = 50.0;

const unsigned long US_TIMEOUT          = 30000;
const unsigned long INTERVAL_CETAK      = 2000;   // cetak serial setiap 2 detik
const unsigned long INTERVAL_KIRIM_DATA = 10000;  // kirim ke backend setiap 10 detik
const unsigned long INTERVAL_POLLING    = 5000;   // polling perintah setiap 5 detik

const unsigned long TIMEOUT_BUKA_PAKAN = 15000;
const unsigned long TIMEOUT_ISI_AIR    = 30000;
const unsigned long JEDA_OTOMATIS      = 60000;   // cooldown 1 menit

// ===================== OBJEK GLOBAL =====================
Servo   servoPakan;
HX711   scalePakan;
HX711   scaleMinum;
Preferences preferences;

unsigned long lastPrint     = 0;
unsigned long lastKirimData = 0;
unsigned long lastPolling   = 0;

unsigned long waktuTerakhirPakan = 0;
unsigned long waktuTerakhirAir   = 0;
long lastCompletedCommandID = 0;

// ===================== SETUP =====================
void setup() {
  Serial.begin(115200);
  delay(500);
  Serial.println();
  Serial.println(F("========================================"));

  preferences.begin("cmd", false);
  lastCompletedCommandID = preferences.getLong("last_ok", 0);
  Serial.print(F("[INIT] Last completed command ID: "));
  Serial.println(lastCompletedCommandID);
  Serial.println(F("   SMART PET FEEDER — FULL OTOMATIS     "));
  Serial.println(F("   + Koneksi Backend                    "));
  Serial.println(F("========================================"));

  // Inisialisasi Servo
  servoPakan.attach(SERVO_PIN);
  servoPakan.write(55);
  Serial.println(F("[INIT] Servo Pakan  → GPIO 18 (TUTUP: 55°)"));

  // Inisialisasi Relay (Pompa)
  pinMode(RELAY_PIN, OUTPUT);
  digitalWrite(RELAY_PIN, RELAY_ACTIVE_LOW ? HIGH : LOW);
  Serial.println(F("[INIT] Relay Pompa  → GPIO 23 (MATI)"));

  // Inisialisasi Ultrasonik
  pinMode(TRIG_PIN, OUTPUT);
  pinMode(ECHO_PIN, INPUT);
  Serial.println(F("[INIT] Ultrasonik   → Trig 26, Echo 27"));

  // Inisialisasi Load Cell Pakan
  scalePakan.begin(LC_PAKAN_DT, LC_PAKAN_SCK);
  scalePakan.set_scale(calibration_factor_pakan);
  scalePakan.tare();
  Serial.println(F("[INIT] Load Cell Pakan → Ready"));

  // Inisialisasi Load Cell Minum
  scaleMinum.begin(LC_MINUM_DT, LC_MINUM_SCK);
  scaleMinum.set_scale(calibration_factor_minum);
  scaleMinum.tare();
  Serial.println(F("[INIT] Load Cell Minum → Ready"));

  // Koneksi WiFi
  connectWiFi();

  // Bypass jeda awal agar sistem langsung aktif
  waktuTerakhirPakan = millis() - JEDA_OTOMATIS;
  waktuTerakhirAir   = millis() - JEDA_OTOMATIS;

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

void nyalakanPompa() {
  digitalWrite(RELAY_PIN, RELAY_ACTIVE_LOW ? LOW : HIGH);
}

void matikanPompa() {
  digitalWrite(RELAY_PIN, RELAY_ACTIVE_LOW ? HIGH : LOW);
}

// ===================== KONEKSI WiFi =====================
void connectWiFi() {
  Serial.print(F("[WiFi] Menghubungkan ke "));
  Serial.print(WIFI_SSID);
  WiFi.begin(WIFI_SSID, WIFI_PASSWORD);
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

// ===================== FUNGSI SENSOR ULTRASONIK =====================
float bacaJarakUltrasonik() {
  digitalWrite(TRIG_PIN, LOW);
  delayMicroseconds(2);
  digitalWrite(TRIG_PIN, HIGH);
  delayMicroseconds(10);
  digitalWrite(TRIG_PIN, LOW);

  long durasi = pulseIn(ECHO_PIN, HIGH, US_TIMEOUT);
  if (durasi == 0) return -1.0;
  float jarak = (durasi * 0.0343) / 2.0;
  return jarak;
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
  if (scalePakan.is_ready()) {
    float berat = scalePakan.get_units(10);
    if (abs(berat) < 1.0) berat = 0.0;
    return berat;
  }
  return -1.0;
}

float bacaBeratMinum() {
  if (scaleMinum.is_ready()) {
    float berat = scaleMinum.get_units(10);
    if (abs(berat) < 1.0) berat = 0.0;
    return berat;
  }
  return -1.0;
}

// ===================== FUNGSI AKTUATOR =====================
bool bukaPakan(String& errorMessage) {
  Serial.println(F("\n[ACT] MEMBERI PAKAN — MULAI"));
  float beratAwal = bacaBeratPakan();
  if (beratAwal < 0) {
    errorMessage = "load cell pakan tidak siap";
    Serial.println(F("[FAIL] Load cell pakan tidak siap."));
    return false;
  }
  float beratTarget = beratAwal + PORSI_PAKAN_GRAM;

  servoPakan.write(0);
  Serial.println(F("[ACT] Servo BUKA (0°) — pakan mengalir..."));

  bool targetTercapai = false;
  unsigned long mulai = millis();
  while (millis() - mulai < TIMEOUT_BUKA_PAKAN) {
    float beratSekarang = bacaBeratPakan();
    if (beratSekarang >= 0 && beratSekarang >= beratTarget) {
      targetTercapai = true;
      break;
    }
    delay(200);
  }

  servoPakan.write(55);
  Serial.println(F("[ACT] Servo TUTUP (55°)"));
  if (!targetTercapai) {
    errorMessage = "target porsi pakan tidak tercapai";
    Serial.println(F("[FAIL] Target porsi pakan tidak tercapai."));
    return false;
  }
  Serial.println(F("[DONE] Pengisian pakan selesai.\n"));
  return true;
}

bool isiAir(String& errorMessage) {
  Serial.println(F("\n[ACT] MENGISI AIR — MULAI"));
  float beratAwal = bacaBeratMinum();
  if (beratAwal < 0) {
    errorMessage = "load cell minum tidak siap";
    Serial.println(F("[FAIL] Load cell minum tidak siap."));
    return false;
  }
  if (beratAwal >= TARGET_AIR_GRAM) {
    Serial.println(F("[SKIP] Air sudah penuh."));
    return true;
  }

  nyalakanPompa();
  Serial.println(F("[ACT] Pompa NYALA — mengisi air..."));

  bool targetTercapai = false;
  unsigned long mulai = millis();
  while (millis() - mulai < TIMEOUT_ISI_AIR) {
    float beratSekarang = bacaBeratMinum();
    if (beratSekarang >= 0 && beratSekarang >= TARGET_AIR_GRAM) {
      targetTercapai = true;
      break;
    }
    delay(200);
  }

  matikanPompa();
  Serial.println(F("[ACT] Pompa MATI"));
  if (!targetTercapai) {
    errorMessage = "target air tidak tercapai";
    Serial.println(F("[FAIL] Target air tidak tercapai."));
    return false;
  }
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
  if (WiFi.status() != WL_CONNECTED) return;

  WiFiClientSecure client;
  client.setInsecure();  // Skip SSL certificate verification
  HTTPClient http;
  String url = String(BACKEND_URL) + "/api/v1/sensors/feed-weight";
  http.begin(client, url);
  http.addHeader("Content-Type", "application/json");
  http.addHeader("X-Device-Key", DEVICE_API_KEY);

  // Tentukan status air
  bool  airTersedia = (beratMinum >= 0 && beratMinum >= BATAS_MINIMAL_AIR);
  String statusAir  = airTersedia ? "available" : "low";

  // Build JSON body
  StaticJsonDocument<768> doc;
  doc["device_id"]    = DEVICE_ID;
  doc["weight_grams"] = (beratPakan >= 0) ? beratPakan : 0.0;
  if (stokPersen >= 0) {
    doc["food_stock_percent"] = stokPersen;
  }
  doc["water_available"] = airTersedia;
  doc["water_status"]    = statusAir;

  String body;
  serializeJson(doc, body);

  int statusCode = http.POST(body);
  if (statusCode == 201) {
    Serial.println(F("[HTTP] Data sensor terkirim ke backend ✓"));
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
 *           { "data": null }  ← tidak ada perintah baru
 */
void pollingPerintah() {
  if (WiFi.status() != WL_CONNECTED) return;

  WiFiClientSecure client;
  client.setInsecure();  // Skip SSL certificate verification
  HTTPClient http;
  String url = String(BACKEND_URL) + "/api/v1/devices/" + String(DEVICE_ID) + "/commands/next";
  http.begin(client, url);
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
  } else if (action == "drink") {
    if (commandID == lastCompletedCommandID) {
      Serial.println(F("[CMD] Command sudah pernah selesai, kirim ulang status saja."));
      berhasil = true;
    } else {
      berhasil = isiAir(errorMessage);
      if (berhasil) {
        waktuTerakhirAir = millis();
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

  WiFiClientSecure client;
  client.setInsecure();  // Skip SSL certificate verification
  HTTPClient http;
  String url = String(BACKEND_URL) + "/api/v1/devices/" + String(DEVICE_ID)
               + "/commands/" + String(commandID) + "/status";
  http.begin(client, url);
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
    Serial.println(berhasil ? F("completed ✓") : F("failed ✗"));
  } else {
    cetakErrorHTTP(F("[CMD] Gagal laporkan status"), statusCode, http);
  }
  http.end();
}

// ===================== CETAK DATA SERIAL =====================
void cetakDataSensor(float bPakan, float bAir, float sPersen) {
  Serial.println(F("┌────────────────────────────────────────┐"));
  Serial.print(F("│ Berat Pakan : ")); Serial.print(bPakan >= 0 ? bPakan : 0.0, 1); Serial.println(F(" g"));
  Serial.print(F("│ Berat Air   : ")); Serial.print(bAir   >= 0 ? bAir   : 0.0, 1); Serial.println(F(" g"));
  Serial.print(F("│ Stok Pakan  : ")); Serial.print(sPersen >= 0 ? sPersen : 0.0, 1); Serial.println(F(" %"));
  Serial.print(F("│ WiFi        : ")); Serial.println(WiFi.status() == WL_CONNECTED ? F("Terhubung ✓") : F("Terputus ✗"));
  Serial.println(F("└────────────────────────────────────────┘\n"));
}

// ===================== LOOP UTAMA =====================
void loop() {
  // ── Kontrol Manual via Serial Monitor ──────────────────────────────────
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
        if (isiAir(errorMessage)) {
          waktuTerakhirAir = millis();
        }
        break;
      case 'T': case 't':
        scalePakan.tare();
        scaleMinum.tare();
        Serial.println(F("[TARE] Kedua load cell di-tare."));
        break;
      case 'W': case 'w':
        pastikanWiFi();
        break;
    }
  }

  // ── Polling Perintah dari Backend diprioritaskan sebelum otomatisasi ──
  if (millis() - lastPolling >= INTERVAL_POLLING) {
    lastPolling = millis();
    pastikanWiFi();
    pollingPerintah();
  }

  // ── Baca Sensor ────────────────────────────────────────────────────────
  float beratPakan = bacaBeratPakan();
  float beratMinum = bacaBeratMinum();

  // ── LOGIKA OTOMATIS PAKAN ───────────────────────────────────────────────
  if (beratPakan >= 0 && beratPakan <= BATAS_MINIMAL_PAKAN) {
    if (millis() - waktuTerakhirPakan >= JEDA_OTOMATIS) {
      Serial.println(F("[AUTO] Mangkok pakan kosong!"));
      String errorMessage = "";
      if (bukaPakan(errorMessage)) {
        waktuTerakhirPakan = millis();
      }
    }
  }

  // ── LOGIKA OTOMATIS AIR ─────────────────────────────────────────────────
  if (beratMinum >= 0 && beratMinum <= BATAS_MINIMAL_AIR) {
    if (millis() - waktuTerakhirAir >= JEDA_OTOMATIS) {
      Serial.println(F("[AUTO] Mangkok air kosong!"));
      String errorMessage = "";
      if (isiAir(errorMessage)) {
        waktuTerakhirAir = millis();
      }
    }
  }

  // ── Kirim Data ke Backend (setiap INTERVAL_KIRIM_DATA ms) ──────────────
  if (millis() - lastKirimData >= INTERVAL_KIRIM_DATA) {
    lastKirimData = millis();
    pastikanWiFi();
    float stokPersen = hitungPersentaseStok();
    kirimDataSensor(beratPakan, beratMinum, stokPersen);
  }

  // ── Cetak Log Serial (setiap INTERVAL_CETAK ms) ─────────────────────────
  if (millis() - lastPrint >= INTERVAL_CETAK) {
    lastPrint = millis();
    float stokPersen = hitungPersentaseStok();
    cetakDataSensor(beratPakan, beratMinum, stokPersen);
  }
}
