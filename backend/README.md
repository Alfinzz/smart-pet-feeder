# Smart Pet Monitoring Backend

REST API backend for the Smart Pet Monitoring IoT system.

## Project Structure

```text
cmd/api/main.go                         # application bootstrap
internal/config                         # env config
internal/domain                         # business entities and errors
internal/usecase                        # application logic
internal/repository/postgres            # PostgreSQL persistence
internal/delivery/http                  # Gin routes, middleware, handlers
internal/security                       # password hashing and JWT
internal/bootstrap                      # demo seed data
migrations                              # PostgreSQL schema
```

## Requirements

- Go 1.22 or newer
- PostgreSQL running on `localhost:5432`

## Setup

1. Create a database named `smart_pet_monitoring`.
2. Apply the migrations in order:

```powershell
psql -U postgres -d smart_pet_monitoring -f migrations/001_init.sql
psql -U postgres -d smart_pet_monitoring -f migrations/002_dashboard_design.sql
psql -U postgres -d smart_pet_monitoring -f migrations/003_health_profile.sql
psql -U postgres -d smart_pet_monitoring -f migrations/004_normalize_device_id.sql
```

3. Copy `.env.example` to `.env` and adjust credentials/secrets.
4. Run the API:

```powershell
go mod tidy
go run ./cmd/api
```

The demo owner is created on startup when `SEED_DEMO_OWNER=true`:

- Email: `owner@gmail.com`
- Password: `password123`

## Endpoints

Public auth:

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`

Device API, secured with `X-Device-Key`:

- `POST /api/v1/sensors/feed-weight`
- `POST /api/v1/sensors/status`
- `GET /api/v1/devices/{device_id}/commands/next`
- `PATCH /api/v1/devices/{device_id}/commands/{command_id}/status`

Owner API, secured with `Authorization: Bearer <token>`:

- `GET /api/v1/dashboard/overview`
- `GET /api/v1/analytics/dashboard`
- `GET /api/v1/analytics/weekly-nutrition?days=7`
- `GET /api/v1/analytics/feed-logs?limit=50`
- `GET /api/v1/feed/history?limit=50`
- `GET /api/v1/feed/weekly-consumption?days=7`
- `POST /api/v1/control/manual`
- `GET /api/v1/health/summary`
- `GET /api/v1/health/overview?days=7`
- `POST /api/v1/health/vitals`
- `PUT /api/v1/health/vitals`
- `GET /api/v1/health/tasks`
- `POST /api/v1/health/tasks`
- `PUT /api/v1/health/tasks/{task_id}`
- `DELETE /api/v1/health/tasks/{task_id}`
- `GET /api/v1/profile`
- `GET /api/v1/profile/pet-details`
- `POST /api/v1/profile/pet-details`
- `PUT /api/v1/profile/pet-details`
- `DELETE /api/v1/profile/pet-details`
- `GET /api/v1/profile/device-settings`
- `PATCH /api/v1/profile/device-settings`
- `POST /api/v1/device/calibrate`
- `POST /api/device/calibrate`
- `GET /api/v1/profile/notification-preferences`
- `PUT /api/v1/profile/notification-preferences`
- `PUT /api/v1/profile/pet`
- `PATCH /api/v1/profile/device`

## Example Payloads

Register:

```json
{
  "name": "Demo Owner",
  "email": "owner@gmail.com",
  "password": "password123"
}
```

Sensor update:

```json
{
  "device_id": "ESP32-001",
  "food_stock_percent": 72,
  "water_available": true,
  "water_status": "Clean & Fresh"
}
```

Manual command:

```json
{
  "device_id": "ESP32-001",
  "action": "feed"
}
```

Care task:

```json
{
  "category": "vaccination",
  "title": "Vaccination",
  "subtitle": "Annual Rabies Booster",
  "due_label": "Due in 5 days",
  "due_at": "2026-05-10",
  "priority": "high",
  "sort_order": 1
}
```

Vital signs:

```json
{
  "weight_kg": 25.4,
  "activity_minutes": 45,
  "sleep_hours": 9.5
}
```

Pet details:

```json
{
  "device_id": "ESP32-001",
  "name": "Fluffy",
  "species": "Dog",
  "breed": "Golden Retriever",
  "gender": "female",
  "birth_date": "2023-05-06",
  "daily_feed_target_grams": 150
}
```

Device settings:

```json
{
  "name": "Kitchen Smart Feeder",
  "manual_feed_portion_grams": 20
}
```

Notification preferences:

```json
{
  "low_food_alert": true,
  "empty_water_alert": true,
  "feeding_success_report": true
}
```
