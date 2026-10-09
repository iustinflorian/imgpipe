package processor

import (
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"imgpipe/pkg/models"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/draw"
)

func downloadImage(ctx context.Context, url string) (image.Image, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("invalid image URL request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("failed to download image, status: %d", resp.StatusCode)
	}

	img, format, err := image.Decode(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode image format: %w", err)
	}

	return img, format, nil
}

func resizeImage(src image.Image, targetWidth, targetHeight int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	draw.NearestNeighbor.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
	return dst
}

func ProcessImage(ctx context.Context, job *models.Job) (string, error) {
	img, format, err := downloadImage(ctx, job.URL)
	if err != nil {
		return "", err
	}

	resized := resizeImage(img, job.Width, job.Height)

	outputFilename := fmt.Sprintf("%s_%dx%d.%s", job.ID, job.Width, job.Height, format)
	outputPath := filepath.Join("renders", outputFilename)

	outFile, err := os.Create(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		err = jpeg.Encode(outFile, resized, &jpeg.Options{Quality: 85})
	case "png":
		err = png.Encode(outFile, resized)
	default:
		err = jpeg.Encode(outFile, resized, &jpeg.Options{Quality: 85})
	}

	if err != nil {
		return "", fmt.Errorf("failed to encode resized image: %w", err)
	}

	return "/" + outputPath, nil
}
