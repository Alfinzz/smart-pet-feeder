INSERT INTO devices (
    id,
    name,
    food_stock_percent,
    water_available,
    water_status,
    last_seen_at,
    created_at
)
SELECT
    'ESP32-001',
    name,
    food_stock_percent,
    water_available,
    water_status,
    last_seen_at,
    created_at
FROM devices
WHERE id = 'esp32-001'
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    food_stock_percent = EXCLUDED.food_stock_percent,
    water_available = EXCLUDED.water_available,
    water_status = EXCLUDED.water_status,
    last_seen_at = GREATEST(devices.last_seen_at, EXCLUDED.last_seen_at);

INSERT INTO devices (id, name, food_stock_percent, water_available, water_status, last_seen_at)
VALUES ('ESP32-001', 'Kitchen Smart Feeder', 75, TRUE, 'Clean & Fresh', NOW())
ON CONFLICT (id) DO NOTHING;

UPDATE pets
SET device_id = 'ESP32-001'
WHERE device_id = 'esp32-001';

UPDATE feed_readings
SET device_id = 'ESP32-001'
WHERE device_id = 'esp32-001';

UPDATE manual_commands
SET device_id = 'ESP32-001'
WHERE device_id = 'esp32-001';

DELETE FROM devices
WHERE id = 'esp32-001'
  AND NOT EXISTS (
      SELECT 1
      FROM pets
      WHERE pets.device_id = devices.id
  );
