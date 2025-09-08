#!/bin/bash

echo "🔨 Building workflow service Docker image..."
docker build -t workflow-service:latest .

echo "🚀 Starting services with Docker Compose..."
docker-compose up -d

echo "✅ Services started!"
echo "📊 Workflow service: http://localhost:8080"
echo "🗄️  PostgreSQL: localhost:5433"
echo ""
echo "📝 To view logs: docker-compose logs -f"
echo "🛑 To stop: docker-compose down" 