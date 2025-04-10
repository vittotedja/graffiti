package api

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/vittotedja/graffiti/graffiti-backend/util/logger"
	"github.com/aws/aws-sdk-go/aws/credentials"
)

// SQSNotificationMessage represents a message to be sent to SQS
type SQSNotificationMessage struct {
	RecipientID string    `json:"recipient_id"`
	SenderID    string    `json:"sender_id"`
	Type        string    `json:"type"`
	EntityID    string    `json:"entity_id,omitempty"`
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
}

// SQSProducer handles sending notifications to SQS
type SQSProducer struct {
	sqsClient *sqs.SQS
	queueURL  string
}

// NewSQSProducer creates a new SQSProducer
func NewSQSProducer(awsRegion string, queueURL string, accessKeyId string, secretKey string, sessionToken string) (*SQSProducer, error) {
	// Create a new AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(awsRegion),
		Credentials: credentials.NewStaticCredentials(
			accessKeyId,
			secretKey,
			sessionToken,
		),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %v", err)
	}

	// Create SQS client
	sqsClient := sqs.New(sess)

	return &SQSProducer{
		sqsClient: sqsClient,
		queueURL:  queueURL,
	}, nil
}

// SendNotification sends a notification to SQS
func (p *SQSProducer) SendNotification(ctx context.Context, message SQSNotificationMessage) error {
	meta := logger.GetMetadata(ctx)
	log := meta.GetLogger()

	// Convert notification to JSON
	messageBody, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %v", err)
	}

	// Send message to SQS
	_, err = p.sqsClient.SendMessageWithContext(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(p.queueURL),
		MessageBody: aws.String(string(messageBody)),
	})

	if err != nil {
		log.Error("Failed to send notification to SQS", err)
		return fmt.Errorf("failed to send message to SQS: %v", err)
	}

	log.Info(fmt.Sprintf("Sent notification to SQS: %s", string(messageBody)))
	return nil
}

// SendLikeNotification is a helper function to send like notifications
func (p *SQSProducer) SendLikeNotification(ctx context.Context, recipientID, senderID, postID, senderName string) error {
	notification := SQSNotificationMessage{
		RecipientID: recipientID,
		SenderID:    senderID,
		Type:        "like",
		EntityID:    postID,
		Message:     fmt.Sprintf("%s liked your post", senderName),
		Timestamp:   time.Now(),
	}

	return p.SendNotification(ctx, notification)
}

// SendFriendRequestNotification is a helper function to send friend request notifications
func (p *SQSProducer) SendFriendRequestNotification(ctx context.Context, recipientID, senderID, senderName string) error {
	notification := SQSNotificationMessage{
		RecipientID: recipientID,
		SenderID:    senderID,
		Type:        "friend_request",
		EntityID:    senderID, // The entity is the sender in this case
		Message:     fmt.Sprintf("%s sent you a friend request", senderName),
		Timestamp:   time.Now(),
	}

	return p.SendNotification(ctx, notification)
}


// SendWallPostNotification sends a notification when someone posts on a user's wall
func (p *SQSProducer) SendWallPostNotification(ctx context.Context, recipientID, senderID, postID, senderName string) error {
	notification := SQSNotificationMessage{
		RecipientID: recipientID,
		SenderID:    senderID,
		Type:        "wall_post",
		EntityID:    postID,
		Message:     fmt.Sprintf("%s posted on your wall", senderName),
		Timestamp:   time.Now(),
	}

	return p.SendNotification(ctx, notification)
}

