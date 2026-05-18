package http

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"smart-pet-monitoring/backend/internal/domain"
)

const petPhotoFormField = "photo"

func (h *Handler) updatePetPhoto(c *gin.Context) {
	ownerID, ok := ownerIDFromContext(c)
	if !ok {
		respondUsecaseError(c, domain.ErrUnauthorized)
		return
	}
	if h.maxUploadSize > 0 {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.maxUploadSize+1024*1024)
	}

	fileHeader, err := c.FormFile(petPhotoFormField)
	if err != nil {
		respondError(c, http.StatusBadRequest, "photo file is required")
		return
	}
	if fileHeader.Size <= 0 {
		respondError(c, http.StatusBadRequest, "photo file is empty")
		return
	}
	if h.maxUploadSize > 0 && fileHeader.Size > h.maxUploadSize {
		respondError(c, http.StatusBadRequest, "photo file is too large")
		return
	}

	extension, err := detectPetPhotoExtension(fileHeader)
	if err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	photo, err := savePetPhoto(c, h.uploadDir, ownerID, fileHeader, extension)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to save photo")
		return
	}

	pet, oldPhotoPath, err := h.dashboard.UpdatePetPhoto(c.Request.Context(), ownerID, photo.PublicPath)
	if err != nil {
		_ = os.Remove(photo.LocalPath)
		respondUsecaseError(c, err)
		return
	}

	removeLocalPetPhoto(h.uploadDir, oldPhotoPath, photo.PublicPath)
	c.JSON(http.StatusOK, h.petProfileResponse(pet))
}
