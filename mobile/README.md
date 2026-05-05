# Smart Pet Monitoring Mobile

Flutter mobile client for the Smart Pet Monitoring backend.

## Run

Android emulator default:

```powershell
flutter run
```

Physical device on the same Wi-Fi network:

```powershell
flutter run --dart-define=API_BASE_URL=http://<backend-ip>:8080/api/v1
```

The app stores the JWT in `flutter_secure_storage` and attaches it to API requests through a Dio interceptor.
