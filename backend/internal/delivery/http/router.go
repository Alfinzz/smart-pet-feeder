package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"smart-pet-monitoring/backend/internal/config"
	"smart-pet-monitoring/backend/internal/security"
	"smart-pet-monitoring/backend/internal/usecase"
)

type Handler struct {
	auth            *usecase.AuthUsecase
	feed            *usecase.FeedUsecase
	control         *usecase.ControlUsecase
	dashboard       *usecase.DashboardUsecase
	health          *usecase.HealthUsecase
	profile         *usecase.ProfileUsecase
	jwt             *security.JWTService
	deviceAPIKey    string
	corsAllowOrigin string
	uploadDir       string
	maxUploadSize   int64
	publicBaseURL   string
}

func NewHandler(
	auth *usecase.AuthUsecase,
	feed *usecase.FeedUsecase,
	control *usecase.ControlUsecase,
	dashboard *usecase.DashboardUsecase,
	health *usecase.HealthUsecase,
	profile *usecase.ProfileUsecase,
	jwt *security.JWTService,
	cfg config.Config,
) *Handler {
	return &Handler{
		auth:            auth,
		feed:            feed,
		control:         control,
		dashboard:       dashboard,
		health:          health,
		profile:         profile,
		jwt:             jwt,
		deviceAPIKey:    cfg.DeviceAPIKey,
		corsAllowOrigin: cfg.CORSAllowOrigin,
		uploadDir:       cfg.UploadDir,
		maxUploadSize:   cfg.MaxUploadSize,
		publicBaseURL:   cfg.PublicBaseURL,
	}
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	router.Use(corsMiddleware(h.corsAllowOrigin))
	router.OPTIONS("/*path", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	healthHandler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
	router.GET("/health", healthHandler)
	router.HEAD("/health", healthHandler)
	router.StaticFS("/uploads", gin.Dir(h.uploadDir, false))

	v1 := router.Group("/api/v1")
	v1.POST("/auth/register", h.register)
	v1.POST("/auth/login", h.login)

	sensors := v1.Group("/sensors")
	sensors.Use(h.deviceAPIKeyMiddleware())
	sensors.POST("/feed-weight", h.createFeedReading)
	sensors.POST("/status", h.updateDeviceStatus)

	devices := v1.Group("/devices")
	devices.Use(h.deviceAPIKeyMiddleware())
	devices.GET("/:deviceID/config", h.getDeviceConfig)
	devices.GET("/:deviceID/commands/next", h.getNextDeviceCommand)
	devices.PATCH("/:deviceID/commands/:commandID/status", h.updateDeviceCommandStatus)

	protected := v1.Group("")
	protected.Use(h.authMiddleware())
	protected.GET("/dashboard/overview", h.getDashboardOverview)
	protected.GET("/analytics/dashboard", h.getDashboardOverview)
	protected.GET("/analytics/feed-logs", h.getFeedHistory)
	protected.GET("/analytics/weekly-nutrition", h.getWeeklyConsumption)
	protected.GET("/feed/history", h.getFeedHistory)
	protected.GET("/feed/weekly-consumption", h.getWeeklyConsumption)
	protected.POST("/control/manual", h.createManualCommand)
	protected.GET("/control/manual/:commandID", h.getManualCommand)
	protected.GET("/health/summary", h.getHealthSummary)
	protected.GET("/health/overview", h.getHealthOverview)
	protected.POST("/health/vitals", h.updateVitalSigns)
	protected.PUT("/health/vitals", h.updateVitalSigns)
	protected.POST("/health/vital-signs", h.updateVitalSigns)
	protected.PUT("/health/vital-signs", h.updateVitalSigns)
	protected.GET("/health/tasks", h.listCareTasks)
	protected.POST("/health/tasks", h.createCareTask)
	protected.PUT("/health/tasks/:taskID", h.updateCareTask)
	protected.DELETE("/health/tasks/:taskID", h.deleteCareTask)
	protected.GET("/profile", h.getProfile)
	protected.GET("/profile/pet-details", h.getPetDetails)
	protected.POST("/profile/pet-details", h.createPetDetails)
	protected.PUT("/profile/pet-details", h.updatePetDetails)
	protected.DELETE("/profile/pet-details", h.deletePetDetails)
	protected.POST("/profile/pet/photo", h.updatePetPhoto)
	protected.PUT("/profile/pet/photo", h.updatePetPhoto)
	protected.GET("/profile/device-settings", h.getDeviceSettings)
	protected.PATCH("/profile/device-settings", h.updateProfileDeviceSettings)
	protected.GET("/profile/notification-preferences", h.getNotificationPreferences)
	protected.PUT("/profile/notification-preferences", h.updateNotificationPreferences)
	protected.PUT("/profile/pet", h.updatePetProfile)
	protected.PATCH("/profile/device", h.updateDeviceSettings)
	protected.POST("/device/calibrate", h.calibrateDevice)

	legacyAPI := router.Group("/api")
	legacyAPI.Use(h.authMiddleware())
	legacyAPI.POST("/device/calibrate", h.calibrateDevice)
}
