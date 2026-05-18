# Smart Pet Monitoring Mobile

Flutter mobile client for the Smart Pet Monitoring backend.

## Run

Copy the example environment file and adjust it for your target backend:

```powershell
cp .env.example .env
```

Android emulator default:

```powershell
flutter run
```

Physical device on the same Wi-Fi network:

```powershell
$env:API_BASE_URL="http://<backend-ip>:8080/api/v1"
$env:DEVICE_ID="ESP32-001"
flutter run --dart-define=API_BASE_URL="$env:API_BASE_URL" --dart-define=DEVICE_ID="$env:DEVICE_ID"
```

The app stores the JWT in `flutter_secure_storage` and attaches it to API requests through a Dio interceptor.
