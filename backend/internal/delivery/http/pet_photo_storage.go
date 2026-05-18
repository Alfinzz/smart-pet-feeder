package http

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type savedPetPhoto struct {
	PublicPath string
	LocalPath  string
}

func savePetPhoto(c *gin.Context, uploadDir string, ownerID int64, fileHeader *multipart.FileHeader, extension string) (savedPetPhoto, error) {
	destinationDir := filepath.Join(uploadDir, "pets")
	if err := os.MkdirAll(destinationDir, 0o755); err != nil {
		return savedPetPhoto{}, err
	}

	fileName := fmt.Sprintf("%d-%d%s", ownerID, time.Now().UnixNano(), extension)
	destinationPath := filepath.Join(destinationDir, fileName)
	if err := c.SaveUploadedFile(fileHeader, destinationPath); err != nil {
		return savedPetPhoto{}, err
	}

	return savedPetPhoto{
		PublicPath: "/uploads/pets/" + fileName,
		LocalPath:  destinationPath,
	}, nil
}

func detectPetPhotoExtension(fileHeader *multipart.FileHeader) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("photo file cannot be read")
	}
	defer file.Close()

	return detectImageExtension(file)
}

func detectImageExtension(reader io.Reader) (string, error) {
	buffer := make([]byte, 512)
	n, err := reader.Read(buffer)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("photo file cannot be read")
	}
	if n == 0 {
		return "", fmt.Errorf("photo file is empty")
	}

	if isWebP(buffer[:n]) {
		return ".webp", nil
	}

	switch http.DetectContentType(buffer[:n]) {
	case "image/jpeg":
		return ".jpg", nil
	case "image/png":
		return ".png", nil
	default:
		return "", fmt.Errorf("photo must be a jpeg, png, or webp image")
	}
}

func isWebP(buffer []byte) bool {
	return len(buffer) >= 12 &&
		string(buffer[0:4]) == "RIFF" &&
		string(buffer[8:12]) == "WEBP"
}

func removeLocalPetPhoto(uploadDir, oldPhotoPath, newPhotoPath string) {
	oldPhotoPath = strings.TrimSpace(oldPhotoPath)
	if oldPhotoPath == "" || oldPhotoPath == newPhotoPath {
		return
	}

	localPath, ok := localUploadPath(uploadDir, oldPhotoPath)
	if !ok {
		return
	}
	_ = os.Remove(localPath)
}

func localUploadPath(uploadDir, publicPath string) (string, bool) {
	cleaned := path.Clean("/" + strings.TrimSpace(publicPath))
	if !strings.HasPrefix(cleaned, "/uploads/pets/") {
		return "", false
	}

	relativePath := strings.TrimPrefix(cleaned, "/uploads/")
	localPath := filepath.Join(uploadDir, filepath.FromSlash(relativePath))

	uploadRoot, err := filepath.Abs(uploadDir)
	if err != nil {
		return "", false
	}
	target, err := filepath.Abs(localPath)
	if err != nil {
		return "", false
	}
	if target == uploadRoot || !strings.HasPrefix(target, uploadRoot+string(os.PathSeparator)) {
		return "", false
	}
	return target, true
}
