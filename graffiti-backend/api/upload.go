package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	db "github.com/vittotedja/graffiti/graffiti-backend/db/sqlc"
)

type PresignRequest struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	FileSize    int64  `json:"file_size"`
	UploadType  string `json:"upload_type"`
}

func (s *Server) presignHandler(ctx *gin.Context) {

	var req PresignRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request"})
		return
	}

	user := ctx.MustGet("currentUser").(db.User)

	// Validate file type (optional)
	ext := getFileExtension(req.Filename)
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
	}

	if !allowedExts[ext] {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid file type. Only images are allowed.",
		})
		return
	}

	// Generate unique filename
	filename := uuid.New().String() + ext
	key := "uploads/" + filename

	if req.UploadType == "profile" {
		key = "profiles/" + user.ID.String() + ext
	}

	if req.UploadType == "background_image" {
		key = "bg/" + user.ID.String() + ext
	}

	// Get presigned URL
	presignedURL, err := s.generatePresignedURL(key, req.ContentType)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to generate presigned URL: " + err.Error(),
		})
		return
	}

	// Public URL that will be accessible after upload
	cloudFrontDomain := s.config.CloudfrontDomain
	publicURL := fmt.Sprintf("https://%s/%s", cloudFrontDomain, key)

	ctx.JSON(http.StatusOK, gin.H{
		"presignedUrl": presignedURL,
		"publicUrl":    publicURL,
		"key":          key,
	})
}

func getFileExtension(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) < 2 {
		return ""
	}
	return "." + strings.ToLower(parts[len(parts)-1])
}

func (s *Server) generatePresignedURL(key, contentType string) (string, error) {
	// Get AWS config
	cfg, err := s.getAWSConfig()
	if err != nil {
		return "", err
	}

	// Create S3 client
	s3Client := s3.NewFromConfig(cfg)

	// Create presign client
	presignClient := s3.NewPresignClient(s3Client)

	bucketName := s.config.AWSS3Bucket

	// Set up the presign parameters
	putObjectInput := &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}

	// Generate the presigned URL with expiration
	presignResult, err := presignClient.PresignPutObject(context.TODO(), putObjectInput,
		s3.WithPresignExpires(15*time.Minute)) // URL expires in 15 minutes
	if err != nil {
		return "", fmt.Errorf("failed to presign: %w", err)
	}

	go func() {
		cfClient := cloudfront.NewFromConfig(cfg)
		callerReference := fmt.Sprintf("invalidate-%d", time.Now().UnixNano())

		_, err := cfClient.CreateInvalidation(context.TODO(), &cloudfront.CreateInvalidationInput{
			DistributionId: aws.String(s.config.CloudfrontDistributionID),
			InvalidationBatch: &types.InvalidationBatch{
				CallerReference: aws.String(callerReference),
				Paths: &types.Paths{
					Quantity: aws.Int32(1),
					Items:    []string{"/" + key},
				},
			},
		})
		if err != nil {
			fmt.Println("Failed to invalidate CloudFront:", err)
		}

		log.Println("Invalidation request sent to CloudFront")
	}()

	return presignResult.URL, nil
}

func (s *Server) getAWSConfig() (aws.Config, error) {
	awsRegion := s.getAWSRegion()

	return config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(awsRegion),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			s.config.AWSAccessKeyID,
			s.config.AWSSecretKey,
			"",
		)),
	)
}

func (s *Server) getAWSRegion() string {
	// awsRegion := os.Getenv("AWS_REGION")
	awsRegion := s.config.AWSRegion
	if awsRegion == "" {
		return "us-east-1" // Default region
	}
	return awsRegion
}

func (s *Server) DeleteFile(ctx context.Context, key string) error {

	if s.config.Env == "unit-test" {
		return nil
	}

	cfg, err := s.getAWSConfig()
	if err != nil {
		return fmt.Errorf("failed to get AWS config: %w", err)
	}

	s3Client := s3.NewFromConfig(cfg)

	_, err = s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.config.AWSS3Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object from S3: %w", err)
	}

	// Optionally, you might also want to invalidate CloudFront cache
	go func() {
		cfClient := cloudfront.NewFromConfig(cfg)
		callerReference := fmt.Sprintf("invalidate-delete-%d", time.Now().UnixNano())

		_, err := cfClient.CreateInvalidation(context.TODO(), &cloudfront.CreateInvalidationInput{
			DistributionId: aws.String(s.config.CloudfrontDistributionID),
			InvalidationBatch: &types.InvalidationBatch{
				CallerReference: aws.String(callerReference),
				Paths: &types.Paths{
					Quantity: aws.Int32(1),
					Items:    []string{"/" + key},
				},
			},
		})
		if err != nil {
			fmt.Println("Failed to invalidate CloudFront after delete:", err)
		} else {
			log.Println("Invalidation request sent to CloudFront after delete")
		}
	}()

	return nil
}
