package api

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "time"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/sqs"
    "github.com/google/uuid"
    db "github.com/vittotedja/graffiti/graffiti-backend/db/sqlc"
    "github.com/jackc/pgx/v5/pgtype"
	logutil "github.com/vittotedja/graffiti/graffiti-backend/util/logger"
)

// NotificationMessage represents a notification message to be sent to SQS
type NotificationMessage struct {
    RecipientID string    `json:"recipient_id"`
    SenderID    string    `json:"sender_id"`
    Type        string    `json:"type"`
    EntityID    string    `json:"entity_id"`
    Message     string    `json:"message"`
    IsRead      bool      `json:"is_read"`
    CreatedAt   time.Time `json:"created_at"`
}

// getSQSClient creates and returns an SQS client using the server's AWS configuration
func (s *Server) getSQSClient() (*sqs.Client, error) {
	logger := logutil.Global()
    logger.Info("Initializing SQS client")

    cfg, err := s.getAWSConfig()
    if err != nil {
        return nil, fmt.Errorf("failed to get AWS config: %w", err)
    }
    
    sqsClient := sqs.NewFromConfig(cfg)
	logger.Info("SQS client initialized successfully")
    return sqsClient, nil
}

func (s *Server) SendNotification(ctx context.Context, recipientID, senderID, notificationType, entityID, message string) error {
	logger := logutil.Global()
    logger.Info("Preparing to send notification to SQS: type=%s, recipient=%s", notificationType, recipientID)
    notification := NotificationMessage{
        RecipientID: recipientID,
        SenderID:    senderID,
        Type:        notificationType,
        EntityID:    entityID,
        Message:     message,
        IsRead:      false,
        CreatedAt:   time.Now(),
    }
    
    messageBody, err := json.Marshal(notification)
    if err != nil {
		logger.Error("Failed to marshal notification: %v", err)
        return fmt.Errorf("failed to marshal notification: %w", err)
    }
    
    client, err := s.getSQSClient()
    if err != nil {
		logger.Error("Failed to get SQS client: %v", err)
        return fmt.Errorf("failed to get SQS client: %w", err)
    }

	logger.Info("Sending message to SQS queue: %s", s.config.SQSQueueURL)
    
    result, err := client.SendMessage(ctx, &sqs.SendMessageInput{
        QueueUrl:    aws.String(s.config.SQSQueueURL),
        MessageBody: aws.String(string(messageBody)),
    })
    if err != nil {
		logger.Error("Failed to send message to SQS: %v", err)
        return fmt.Errorf("failed to send message to SQS: %w", err)
    }

	logger.Info("Successfully sent message to SQS: MessageId=%s", *result.MessageId)
    
    return nil
}

// ListQueues lists all available SQS queues
func (s *Server) ListQueues(ctx context.Context) ([]string, error) {
    client, err := s.getSQSClient()
    if err != nil {
        return nil, err
    }
    
    var queueUrls []string
    paginator := sqs.NewListQueuesPaginator(client, &sqs.ListQueuesInput{})
    
    for paginator.HasMorePages() {
        output, err := paginator.NextPage(ctx)
        if err != nil {
            return nil, fmt.Errorf("failed to list SQS queues: %w", err)
        }
        queueUrls = append(queueUrls, output.QueueUrls...)
    }
    
    return queueUrls, nil
}

// ProcessNotifications polls the SQS queue for notifications and processes them
func (s *Server) ProcessNotifications(ctx context.Context) error {
	logger := logutil.Global()
    logger.Info("Starting to poll for messages from SQS queue: %s", s.config.SQSQueueURL)
    
    client, err := s.getSQSClient()
    if err != nil {
		logger.Error("Failed to get SQS client for receiving messages: %v", err)
        return err
    }
    
	logger.Info("Receiving messages with long polling (wait time: 20s)")
    result, err := client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
        QueueUrl:            aws.String(s.config.SQSQueueURL),
        MaxNumberOfMessages: 10, // Receive up to 10 messages at once
        WaitTimeSeconds:     20, // Long polling (wait up to 20 seconds for messages)
    })
    if err != nil {
		logger.Error("Failed to receive messages from SQS: %v", err)
        return fmt.Errorf("failed to receive messages from SQS: %w", err)
    }

	messageCount := len(result.Messages)
    logger.Info("Received %d messages from SQS queue", messageCount)
    
    for i, message := range result.Messages {
		logger.Info("Processing message %d/%d: MessageId=%s", i+1, messageCount, *message.MessageId)

        var notification NotificationMessage
        err := json.Unmarshal([]byte(*message.Body), &notification)
        if err != nil {
			errorMessage := fmt.Sprintf("Failed to unmarshal notification (MessageId=%s): ", *message.MessageId)
			logger.Error(errorMessage, err)
            log.Printf("Failed to unmarshal notification: %v\n", err)
            continue
        }

		logger.Info("Notification details: Type=%s, Recipient=%s, Sender=%s", 
            notification.Type, notification.RecipientID, notification.SenderID)
        
        err = s.storeNotificationInDB(ctx, notification)
        if err != nil {
			errorMessage := fmt.Sprintf("Failed to store notification in DB (MessageId=%s): ", *message.MessageId)
			logger.Error(errorMessage, err)
            log.Printf("Failed to store notification in DB: %v\n", err)
            continue
        }

		logger.Info("Successfully stored notification in database")
        
		logger.Info("Deleting message from queue: MessageId=%s", *message.MessageId)
        _, err = client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
            QueueUrl:      aws.String(s.config.SQSQueueURL),
            ReceiptHandle: message.ReceiptHandle,
        })
        if err != nil {
			errorMessage := fmt.Sprintf("Failed to delete message from SQS (MessageId=%s): ", *message.MessageId)
			logger.Error(errorMessage, err)
            log.Printf("Failed to delete message from SQS: %v\n", err)
        } else {
			logger.Info("Successfully deleted message from SQS queue: MessageId=%s", *message.MessageId)
		}
    }
    
    return nil
}

// storeNotificationInDB stores a notification in the database
func (s *Server) storeNotificationInDB(ctx context.Context, notification NotificationMessage) error {
    recipientID, err := uuid.Parse(notification.RecipientID)
    if err != nil {
        return fmt.Errorf("invalid recipient ID: %w", err)
    }
    
    senderID, err := uuid.Parse(notification.SenderID)
    if err != nil {
        return fmt.Errorf("invalid sender ID: %w", err)
    }
    
    entityID, err := uuid.Parse(notification.EntityID)
    if err != nil {
        return fmt.Errorf("invalid entity ID: %w", err)
    }
    
    params := db.CreateNotificationParams{
        RecipientID: pgtype.UUID{Bytes: recipientID, Valid: true},
        SenderID:    pgtype.UUID{Bytes: senderID, Valid: true},
        Type:        notification.Type,
        EntityID:    pgtype.UUID{Bytes: entityID, Valid: true},
        Message:     notification.Message,
        IsRead:      pgtype.Bool{Bool: notification.IsRead, Valid: true},
        CreatedAt:   pgtype.Timestamp{Time: notification.CreatedAt, Valid: true},
    }
    
    _, err = s.hub.CreateNotification(ctx, params)
    return err
}
