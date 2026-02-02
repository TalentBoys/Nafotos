package services

import (
	"crypto/md5"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	_ "image/gif"
	"os"
	"path/filepath"

	"github.com/disintegration/imaging"
	_ "golang.org/x/image/tiff" // TIFF format support
	_ "golang.org/x/image/bmp"  // BMP format support
	_ "golang.org/x/image/webp" // WebP format support
)

// ThumbnailSize defines the size variants for thumbnails
type ThumbnailSize struct {
	Name   string
	Width  int
	Height int
}

var (
	// ThumbnailSizes defines available thumbnail sizes
	ThumbnailSizes = map[string]ThumbnailSize{
		"small":  {Name: "small", Width: 300, Height: 300},
		"medium": {Name: "medium", Width: 800, Height: 800},
		"large":  {Name: "large", Width: 1920, Height: 1920},
	}
)

type ThumbnailService struct {
	thumbsDir string
}

func NewThumbnailService(thumbsDir string) *ThumbnailService {
	return &ThumbnailService{
		thumbsDir: thumbsDir,
	}
}

// GetThumbnail returns the path to a thumbnail, generating it if necessary
// sizeType can be "small", "medium", or "large". Defaults to "small" if empty.
func (ts *ThumbnailService) GetThumbnail(originalPath string, fileID int64, sizeType string) (string, error) {
	// Default to small size if not specified
	if sizeType == "" {
		sizeType = "small"
	}

	size, ok := ThumbnailSizes[sizeType]
	if !ok {
		sizeType = "small"
		size = ThumbnailSizes["small"]
	}

	// Generate thumbnail filename based on file ID, hash, and size
	hash := fmt.Sprintf("%x", md5.Sum([]byte(originalPath)))
	thumbFilename := fmt.Sprintf("%d_%s_%s.jpg", fileID, hash[:8], sizeType)
	thumbPath := filepath.Join(ts.thumbsDir, thumbFilename)

	// Check if thumbnail already exists
	if _, err := os.Stat(thumbPath); err == nil {
		return thumbPath, nil
	}

	// Generate thumbnail
	if err := ts.generateThumbnail(originalPath, thumbPath, size.Width, size.Height); err != nil {
		return "", err
	}

	return thumbPath, nil
}

// generateThumbnail creates a thumbnail from an image
func (ts *ThumbnailService) generateThumbnail(srcPath, dstPath string, width, height int) error {
	// Open source image
	src, err := imaging.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}

	// Resize image to thumbnail size while maintaining aspect ratio
	thumb := imaging.Fit(src, width, height, imaging.Lanczos)

	// Save thumbnail
	if err := imaging.Save(thumb, dstPath, imaging.JPEGQuality(85)); err != nil {
		return fmt.Errorf("failed to save thumbnail: %w", err)
	}

	return nil
}

// GetDimensions returns the dimensions of an image
func GetDimensions(imagePath string) (int, int, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	img, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, err
	}

	return img.Width, img.Height, nil
}
