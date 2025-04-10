-- Create notifications table
CREATE TABLE IF NOT EXISTS notifications (
    "id" uuid PRIMARY KEY DEFAULT gen_random_uuid (),
    "recipient_id" uuid NOT NULL,
    "sender_id" uuid NOT NULL,
    "type" varchar NOT NULL,
    "entity_id" uuid NOT NULL,
    "message" varchar NOT NULL,
    "is_read" boolean DEFAULT false,
    "created_at" timestamp DEFAULT (now ()),
    
    CONSTRAINT "notifications_recipient_fk" FOREIGN KEY ("recipient_id") REFERENCES "users"("id") ON DELETE CASCADE,
    CONSTRAINT "notifications_sender_fk" FOREIGN KEY ("sender_id") REFERENCES "users"("id") ON DELETE CASCADE
);

-- Add indexes
CREATE INDEX idx_notifications_recipient_id ON "notifications"("recipient_id");
CREATE INDEX idx_notifications_created_at ON "notifications"("created_at");