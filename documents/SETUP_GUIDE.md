# Brain Sentry - Setup & Development Guide

**Version:** 1.0  
**Last Updated:** Janeiro 2025  

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Backend Setup](#backend-setup)
3. [Frontend Setup](#frontend-setup)
4. [Database Setup (FalkorDB)](#database-setup)
5. [LLM Setup](#llm-setup)
6. [Running the Application](#running-the-application)
7. [Development Workflow](#development-workflow)
8. [Testing](#testing)
9. [Deployment](#deployment)
10. [Troubleshooting](#troubleshooting)

---

## Prerequisites

### Required Software

```bash
# Java
Java 17 or higher
Maven 3.8+ or Gradle 8+

# Node.js
Node.js 18+ (recommend 20 LTS)
npm 9+ or pnpm 8+

# Docker
Docker 24+
Docker Compose 2.20+

# Git
Git 2.40+

# IDE (Optional but recommended)
IntelliJ IDEA (for backend)
VS Code (for frontend)
```

### Hardware Requirements

**Minimum (Development):**
- CPU: 4 cores
- RAM: 16GB
- Storage: 50GB free
- GPU: Optional (CPU-only mode available)

**Recommended (With LLM):**
- CPU: 8 cores
- RAM: 32GB
- Storage: 100GB free (50GB for models)
- GPU: NVIDIA RTX 3060 (12GB VRAM) or better

---

## Backend Setup

### 1. Clone Repository

```bash
git clone https://github.com/your-org/brain-sentry.git
cd brain-sentry/backend
```

### 2. Project Structure

```bash
# Create directory structure
mkdir -p src/main/java/com/integraltech/brainsentry
mkdir -p src/main/resources
mkdir -p src/test/java
mkdir -p docker
```

### 3. Maven Setup

Create `pom.xml`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 
         http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>
    
    <parent>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-starter-parent</artifactId>
        <version>3.2.1</version>
    </parent>
    
    <groupId>com.integraltech</groupId>
    <artifactId>brain-sentry</artifactId>
    <version>0.1.0</version>
    <name>Brain Sentry</name>
    
    <properties>
        <java.version>17</java.version>
        <lombok.version>1.18.30</lombok.version>
        <mapstruct.version>1.5.5.Final</mapstruct.version>
    </properties>
    
    <dependencies>
        <!-- Spring Boot -->
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-web</artifactId>
        </dependency>
        
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-data-redis</artifactId>
        </dependency>
        
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-actuator</artifactId>
        </dependency>
        
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-validation</artifactId>
        </dependency>
        
        <!-- Redis / FalkorDB -->
        <dependency>
            <groupId>redis.clients</groupId>
            <artifactId>jedis</artifactId>
            <version>5.1.0</version>
        </dependency>
        
        <!-- Lombok -->
        <dependency>
            <groupId>org.projectlombok</groupId>
            <artifactId>lombok</artifactId>
            <version>${lombok.version}</version>
            <scope>provided</scope>
        </dependency>
        
        <!-- MapStruct -->
        <dependency>
            <groupId>org.mapstruct</groupId>
            <artifactId>mapstruct</artifactId>
            <version>${mapstruct.version}</version>
        </dependency>
        
        <!-- Testing -->
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-test</artifactId>
            <scope>test</scope>
        </dependency>
        
        <dependency>
            <groupId>org.testcontainers</groupId>
            <artifactId>testcontainers</artifactId>
            <version>1.19.3</version>
            <scope>test</scope>
        </dependency>
    </dependencies>
    
    <build>
        <plugins>
            <plugin>
                <groupId>org.springframework.boot</groupId>
                <artifactId>spring-boot-maven-plugin</artifactId>
            </plugin>
            
            <plugin>
                <groupId>org.apache.maven.plugins</groupId>
                <artifactId>maven-compiler-plugin</artifactId>
                <configuration>
                    <source>17</source>
                    <target>17</target>
                    <annotationProcessorPaths>
                        <path>
                            <groupId>org.projectlombok</groupId>
                            <artifactId>lombok</artifactId>
                            <version>${lombok.version}</version>
                        </path>
                        <path>
                            <groupId>org.mapstruct</groupId>
                            <artifactId>mapstruct-processor</artifactId>
                            <version>${mapstruct.version}</version>
                        </path>
                    </annotationProcessorPaths>
                </configuration>
            </plugin>
        </plugins>
    </build>
</project>
```

### 4. Application Configuration

Create `src/main/resources/application.yml`:

```yaml
server:
  port: 8080

spring:
  application:
    name: brain-sentry
  
  redis:
    host: localhost
    port: 6379
    password: 
    timeout: 2000ms

brain-sentry:
  graph:
    name: brain_sentry
  
  llm:
    model-path: ${LLM_MODEL_PATH:./models/qwen2.5-7b.gguf}
    context-size: 4096
    gpu-layers: 35
  
  embedding:
    model: all-MiniLM-L6-v2
    dimensions: 384

logging:
  level:
    com.integraltech.brainsentry: DEBUG
    org.springframework: INFO
```

### 5. Environment Variables

Create `.env`:

```bash
# Redis / FalkorDB
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# LLM
LLM_MODEL_PATH=./models/qwen2.5-7b.gguf
GPU_LAYERS=35

# Application
SERVER_PORT=8080
```

### 6. Build & Run

```bash
# Install dependencies
mvn clean install

# Run application
mvn spring-boot:run

# Or with custom profile
mvn spring-boot:run -Dspring-boot.run.profiles=dev

# Build JAR
mvn clean package
java -jar target/brain-sentry-0.1.0.jar
```

### 7. Verify Backend

```bash
# Health check
curl http://localhost:8080/actuator/health

# Expected response:
# {"status":"UP"}
```

---

## Frontend Setup

### 1. Navigate to Frontend

```bash
cd ../frontend
```

### 2. Initialize Next.js Project

```bash
# Create Next.js app
npx create-next-app@latest . --typescript --tailwind --app --src-dir

# Install dependencies
npm install @radix-ui/react-dialog \
            @radix-ui/react-dropdown-menu \
            @radix-ui/react-tabs \
            @radix-ui/react-select \
            @radix-ui/react-label \
            @radix-ui/react-slot \
            @radix-ui/react-toast \
            class-variance-authority \
            clsx \
            tailwind-merge \
            lucide-react \
            zustand \
            @tanstack/react-query \
            react-hook-form \
            @hookform/resolvers \
            zod \
            recharts \
            reactflow \
            axios \
            date-fns

# Install dev dependencies
npm install -D @types/node \
               eslint \
               prettier \
               prettier-plugin-tailwindcss
```

### 3. Project Configuration

Create `tsconfig.json`:

```json
{
  "compilerOptions": {
    "target": "ES2020",
    "lib": ["dom", "dom.iterable", "esnext"],
    "allowJs": true,
    "skipLibCheck": true,
    "strict": true,
    "forceConsistentCasingInFileNames": true,
    "noEmit": true,
    "esModuleInterop": true,
    "module": "esnext",
    "moduleResolution": "bundler",
    "resolveJsonModule": true,
    "isolatedModules": true,
    "jsx": "preserve",
    "incremental": true,
    "plugins": [
      {
        "name": "next"
      }
    ],
    "paths": {
      "@/*": ["./src/*"]
    }
  },
  "include": ["next-env.d.ts", "**/*.ts", "**/*.tsx", ".next/types/**/*.ts"],
  "exclude": ["node_modules"]
}
```

Create `.env.local`:

```bash
NEXT_PUBLIC_API_URL=http://localhost:8080/api
```

### 4. Run Frontend

```bash
# Development mode
npm run dev

# Production build
npm run build
npm run start
```

### 5. Verify Frontend

Open browser: `http://localhost:3000`

---

## Database Setup (FalkorDB)

### Option 1: Docker Compose (Recommended)

Create `docker/docker-compose.yml`:

```yaml
version: '3.8'

services:
  falkordb:
    image: falkordb/falkordb:latest
    container_name: brain-sentry-falkordb
    ports:
      - "6379:6379"
    volumes:
      - falkordb-data:/data
    environment:
      - REDIS_PASSWORD=
    command: >
      redis-server
      --loadmodule /usr/lib/redis/modules/falkordb.so
      --save 60 1
      --loglevel notice
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 3s
      retries: 5

volumes:
  falkordb-data:
    driver: local
```

### Start FalkorDB

```bash
cd docker
docker-compose up -d

# Verify
docker-compose ps
docker-compose logs falkordb

# Test connection
redis-cli ping
# Expected: PONG
```

### Option 2: Local Installation

```bash
# Ubuntu/Debian
wget https://github.com/FalkorDB/FalkorDB/releases/latest/download/falkordb.deb
sudo dpkg -i falkordb.deb

# Start service
sudo systemctl start falkordb
sudo systemctl enable falkordb

# Verify
redis-cli ping
```

### Initialize Database

```bash
# Connect to Redis
redis-cli

# Create graph
127.0.0.1:6379> GRAPH.QUERY brain_sentry "CREATE ()"
# Expected: (empty array)

# Verify
127.0.0.1:6379> GRAPH.LIST
# Expected: 1) "brain_sentry"
```

---

## LLM Setup

### 1. Download Model

```bash
# Create models directory
mkdir -p models
cd models

# Download Qwen 2.5-7B GGUF
# Option 1: HuggingFace
huggingface-cli download \
  Qwen/Qwen2.5-7B-Instruct-GGUF \
  qwen2_5-7b-instruct-q4_k_m.gguf \
  --local-dir .

# Option 2: Manual download
wget https://huggingface.co/Qwen/Qwen2.5-7B-Instruct-GGUF/resolve/main/qwen2_5-7b-instruct-q4_k_m.gguf
```

### 2. Test Model (Optional)

```bash
# Using llama.cpp
git clone https://github.com/ggerganov/llama.cpp
cd llama.cpp
make

# Test inference
./main -m ../models/qwen2_5-7b-instruct-q4_k_m.gguf \
       -n 128 \
       -p "Hello, how are you?"
```

### 3. Configure Backend

Update `.env`:

```bash
LLM_MODEL_PATH=./models/qwen2_5-7b-instruct-q4_k_m.gguf
GPU_LAYERS=35  # Adjust based on your GPU
```

---

## Running the Application

### Full Stack Development

#### Terminal 1: FalkorDB
```bash
cd docker
docker-compose up
```

#### Terminal 2: Backend
```bash
cd backend
mvn spring-boot:run
```

#### Terminal 3: Frontend
```bash
cd frontend
npm run dev
```

### Access Application

- **Frontend:** http://localhost:3000
- **Backend API:** http://localhost:8080/api
- **Health Check:** http://localhost:8080/actuator/health
- **FalkorDB:** redis://localhost:6379

---

## Development Workflow

### Backend Development

#### Hot Reload
```bash
# Spring Boot DevTools (add to pom.xml)
<dependency>
    <groupId>org.springframework.boot</groupId>
    <artifactId>spring-boot-devtools</artifactId>
    <scope>runtime</scope>
    <optional>true</optional>
</dependency>

# Run with DevTools
mvn spring-boot:run
```

#### Code Style
```bash
# Google Java Format
mvn com.spotify.fmt:fmt-maven-plugin:format

# Checkstyle
mvn checkstyle:check
```

#### Database Migrations
```bash
# Connect to FalkorDB
redis-cli

# Run Cypher scripts
127.0.0.1:6379> GRAPH.QUERY brain_sentry "CREATE INDEX..."
```

### Frontend Development

#### Linting
```bash
# ESLint
npm run lint

# Fix issues
npm run lint -- --fix
```

#### Formatting
```bash
# Prettier
npm run format

# Check formatting
npm run format:check
```

#### Type Checking
```bash
npm run type-check
```

---

## Testing

### Backend Tests

```bash
# Run all tests
mvn test

# Run specific test
mvn test -Dtest=MemoryServiceTest

# Integration tests
mvn verify

# With coverage
mvn test jacoco:report
# Report: target/site/jacoco/index.html
```

### Frontend Tests

```bash
# Install testing libraries
npm install -D @testing-library/react \
               @testing-library/jest-dom \
               @testing-library/user-event \
               vitest

# Run tests
npm run test

# Watch mode
npm run test:watch

# Coverage
npm run test:coverage
```

### End-to-End Tests

```bash
# Install Playwright
npm install -D @playwright/test

# Run E2E tests
npm run test:e2e

# With UI
npm run test:e2e -- --ui
```

---

## Deployment

### Docker Build

#### Backend
```dockerfile
# backend/Dockerfile
FROM eclipse-temurin:17-jre-alpine
WORKDIR /app
COPY target/brain-sentry-0.1.0.jar app.jar
EXPOSE 8080
ENTRYPOINT ["java", "-jar", "app.jar"]
```

```bash
# Build
docker build -t brain-sentry-backend:latest .

# Run
docker run -p 8080:8080 \
  -e REDIS_HOST=host.docker.internal \
  brain-sentry-backend:latest
```

#### Frontend
```dockerfile
# frontend/Dockerfile
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM node:20-alpine AS runner
WORKDIR /app
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static
COPY --from=builder /app/public ./public
EXPOSE 3000
CMD ["node", "server.js"]
```

```bash
# Build
docker build -t brain-sentry-frontend:latest .

# Run
docker run -p 3000:3000 \
  -e NEXT_PUBLIC_API_URL=http://localhost:8080/api \
  brain-sentry-frontend:latest
```

### Production Docker Compose

```yaml
version: '3.8'

services:
  falkordb:
    image: falkordb/falkordb:latest
    volumes:
      - falkordb-data:/data
    environment:
      - REDIS_PASSWORD=${REDIS_PASSWORD}
    restart: unless-stopped
  
  backend:
    image: brain-sentry-backend:latest
    ports:
      - "8080:8080"
    environment:
      - REDIS_HOST=falkordb
      - REDIS_PORT=6379
      - REDIS_PASSWORD=${REDIS_PASSWORD}
      - LLM_MODEL_PATH=/models/qwen2.5-7b.gguf
    volumes:
      - ./models:/models:ro
    depends_on:
      - falkordb
    restart: unless-stopped
  
  frontend:
    image: brain-sentry-frontend:latest
    ports:
      - "3000:3000"
    environment:
      - NEXT_PUBLIC_API_URL=http://localhost:8080/api
    depends_on:
      - backend
    restart: unless-stopped

volumes:
  falkordb-data:
```

---

## Troubleshooting

### Backend Issues

#### Port Already in Use
```bash
# Find process using port 8080
lsof -i :8080
# Or
netstat -ano | grep :8080

# Kill process
kill -9 <PID>
```

#### Redis Connection Failed
```bash
# Check if Redis is running
docker-compose ps falkordb

# Check logs
docker-compose logs falkordb

# Test connection
redis-cli ping
```

#### Out of Memory (LLM)
```bash
# Reduce GPU layers
export GPU_LAYERS=0  # CPU-only mode

# Or use smaller model
# qwen2_5-3b instead of qwen2_5-7b
```

### Frontend Issues

#### Module Not Found
```bash
# Clear node_modules and reinstall
rm -rf node_modules package-lock.json
npm install
```

#### Port 3000 in Use
```bash
# Use different port
PORT=3001 npm run dev
```

#### API Connection Error
```bash
# Check backend is running
curl http://localhost:8080/actuator/health

# Verify NEXT_PUBLIC_API_URL in .env.local
cat .env.local
```

### Database Issues

#### FalkorDB Won't Start
```bash
# Check Docker logs
docker-compose logs falkordb

# Remove and recreate
docker-compose down -v
docker-compose up -d
```

#### Graph Query Errors
```bash
# Verify graph exists
redis-cli
127.0.0.1:6379> GRAPH.LIST

# Recreate graph
127.0.0.1:6379> GRAPH.DELETE brain_sentry
127.0.0.1:6379> GRAPH.QUERY brain_sentry "CREATE ()"
```

---

## Useful Commands

### Development

```bash
# Clean and rebuild everything
make clean && make build

# Run tests
make test

# Start all services
make dev

# Stop all services
make stop

# View logs
make logs
```

### Create Makefile

```makefile
.PHONY: help dev stop clean build test

help:
	@echo "Brain Sentry Development Commands"
	@echo "  make dev    - Start all services"
	@echo "  make stop   - Stop all services"
	@echo "  make clean  - Clean build artifacts"
	@echo "  make build  - Build all services"
	@echo "  make test   - Run all tests"

dev:
	docker-compose -f docker/docker-compose.yml up -d
	cd backend && mvn spring-boot:run &
	cd frontend && npm run dev &

stop:
	docker-compose -f docker/docker-compose.yml down
	pkill -f "spring-boot:run"
	pkill -f "next dev"

clean:
	cd backend && mvn clean
	cd frontend && rm -rf .next node_modules

build:
	cd backend && mvn clean package
	cd frontend && npm run build

test:
	cd backend && mvn test
	cd frontend && npm run test
```

---

## IDE Setup

### IntelliJ IDEA (Backend)

1. Open `backend` folder as Maven project
2. Enable annotation processing:
   - Settings → Build → Compiler → Annotation Processors
   - ✅ Enable annotation processing
3. Install plugins:
   - Lombok
   - MapStruct Support
4. Set JDK to Java 17
5. Import code style (Google Java Format)

### VS Code (Frontend)

Install extensions:
```json
{
  "recommendations": [
    "dbaeumer.vscode-eslint",
    "esbenp.prettier-vscode",
    "bradlc.vscode-tailwindcss",
    "formulahendry.auto-rename-tag",
    "PKief.material-icon-theme"
  ]
}
```

Create `.vscode/settings.json`:
```json
{
  "editor.defaultFormatter": "esbenp.prettier-vscode",
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.fixAll.eslint": true
  },
  "typescript.tsdk": "node_modules/typescript/lib"
}
```

---

## Next Steps

1. ✅ Verify all services are running
2. ✅ Access frontend at http://localhost:3000
3. ✅ Create your first memory via UI
4. ✅ Test interception feature
5. ⏭️ Start Phase 1 development

---

**Document Status:** ✅ Complete  
**Support:** For issues, check [Troubleshooting](#troubleshooting) or open an issue on GitHub
