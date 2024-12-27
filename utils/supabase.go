package utils

import (
	"fmt"
	"log"
	"os"

	"github.com/go-resty/resty/v2"
)

// UploadImageToSupabase uploads an image to Supabase Storage and returns its public URL
func UploadImageToSupabase(filePath, fileName string) (string, error) {
	client := resty.New()

	// Read environment variables
	supabaseURL := os.Getenv("SUPABASE_URL")
	apiKey := os.Getenv("SUPABASE_API_KEY")
	bucket := os.Getenv("SUPABASE_BUCKET")

	// Upload file to Supabase
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s/%s", supabaseURL, bucket, "uploads", fileName)
	resp, err := client.R().
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", apiKey)).
		SetHeader("Content-Type", "multipart/form-data").
		SetFile("file", filePath).Post(url)

	if err != nil {
		return "", fmt.Errorf("failed to upload image: %v", err)
	}

	log.Println(resp)

	if resp.StatusCode() != 200 && resp.StatusCode() != 201 {
		return "", fmt.Errorf("failed to upload image, status code: %d, response: %s", resp.StatusCode(), resp.String())
	}

	// Construct public URL
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s/%s", supabaseURL, bucket, "uploads", fileName)
	fmt.Println("publicURL", publicURL)
	return publicURL, nil
}
