# Graffiti - Social Wall Application

Graffiti is a modern social platform that allows users to create and share content on customizable walls. Built with Next.js for the frontend and Go for the backend, this application features real-time notifications, friend requests, and interactive content sharing.

## Features

- **User Authentication**: Secure login and registration system
- **Customizable Walls**: Create and personalize your own walls
- **Friend System**: Send/accept friend requests and discover mutual connections
- **Real-time Notifications**: Get notified when someone interacts with your content
- **Content Sharing**: Post media and links to your walls or friends' walls
- **Like System**: Interact with posts through likes
- **Responsive Design**: Optimized for both desktop and mobile experiences

## Tech Stack

### Frontend
- **Framework**: [Next.js](https://nextjs.org/) (React)
- **Styling**: Tailwind CSS
- **State Management**: React Hooks
- **Routing**: Next.js App Router
- **Deployment**: [Vercel](https://vercel.com)

### Backend
- **Language**: Go (Golang)
- **Database**: PostgreSQL
- **API**: RESTful API with Gin framework
- **Storage**: AWS S3 for media uploads
- **Notifications**: AWS SQS for asynchronous notification delivery

## Getting Started

### Prerequisites
- Node.js (v18 or later)
- Go (v1.19 or later)
- PostgreSQL
- AWS account (for S3 and SQS)

### Frontend Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/vittotedja/graffiti.git
   cd graffiti/graffiti-frontend
   ```

2. Install dependencies:
   ```bash
   npm install
   # or
   yarn install
   # or
   pnpm install
   # or
   bun install
   ```

3. Create a `.env` file in the graffiti-frontend directory with the following variables:
   ```
   NEXT_PUBLIC_API_URL=http://localhost:8080
   ```

4. Run the development server:
   ```bash
   npm run dev
   # or
   yarn dev
   # or
   pnpm dev
   # or
   bun dev
   ```

5. Open [http://localhost:3000](http://localhost:3000) with your browser to see the result.

### Backend Setup

1. Navigate to the backend directory:
   ```bash
   cd graffiti/graffiti-backend
   ```

2. Create a `.env` file with your configuration:
   ```
   ENV=devlocal
   DB_DRIVER=postgres
   DB_SOURCE=postgresql://username:password@localhost:5432/graffiti?sslmode=disable
   DB_SOURCE_DOCKER=postgresql://username:password@localhost:5432/graffiti?sslmode=disable
   SERVER_ADDRESS=0.0.0.0:8080
   TOKEN_SYMMETRIC_KEY=your_secret_key_here
   
   # AWS Configuration
   AWS_REGION=your_aws_region
   AWS_ACCESS_KEY_ID=your_access_key
   AWS_SECRET_ACCESS_KEY=your_secret_key
   AWS_S3_BUCKET=your_bucket_name
   CLOUDFRONT_DOMAIN=your_cloudfront_domain
   CLOUDFRONT_DISTRIBUTION_ID=your_cloudfront_distribution_id
   FRONTEND_URL=http://localhost:3000
   IS_PRODUCTION=false
   REDIS_HOST=localhost:6379
   SQS_QUEUE_URL=your_sqs_queue_url
   SQS_DLQ_URL=your_dlq_url
   ```

3. Run the backend server:
   ```bash
   go run main.go
   ```

4. The backend API will be available at [http://localhost:8080](http://localhost:8080)

## Project Structure

```
graffiti/
├── graffiti-frontend/     # Next.js frontend
│   ├── app/               # App router pages and layouts
│   ├── components/        # Reusable UI components
│   ├── hooks/             # Custom React hooks
│   ├── lib/               # Utility functions
│   ├── public/            # Static assets
│   ├── services/          # API service functions
│   └── types/             # TypeScript type definitions
│
├── graffiti-backend/      # Go backend
│   ├── api/               # API handlers and routes
│   ├── db/                # Database models and queries
│   ├── token/             # Authentication logic
│   └── util/              # Utility functions
```

## Notification System

Graffiti uses AWS SQS for asynchronous notification delivery. The system handles various notification types:

### Notification Types
- **friend_request**: When someone sends you a friend request
- **friend_request_accepted**: When someone accepts your friend request
- **post_like**: When someone likes your post
- **wall_post**: When someone posts on your wall

## Deployment

### Frontend
The frontend is deployed on [Vercel](https://example.com). Any push to the main branch will trigger a new deployment.

### Backend
The backend is deployed on a cloud provider with the following steps:
1. Build the Go binary
2. Set up environment variables
3. Run the binary with proper configuration
