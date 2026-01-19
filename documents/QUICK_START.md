# Brain Sentry - Quick Start Guide

**Get Started in 30 Minutes** ⚡  
**Version:** 1.0  
**Date:** January 2025  

---

## 🎯 Goal

Build a running Brain Sentry instance with:
- ✅ Backend API responding
- ✅ Frontend UI accessible
- ✅ FalkorDB connected
- ✅ First memory created

---

## ⚡ 5-Step Quick Start

### Step 1: Prerequisites (5 min)

```bash
# Check versions
java --version        # Need: Java 17+
node --version        # Need: Node 18+
docker --version      # Need: Docker 24+
mvn --version         # Need: Maven 3.8+

# If missing, install:
# - Java: https://adoptium.net/
# - Node: https://nodejs.org/
# - Docker: https://docker.com/
# - Maven: https://maven.apache.org/
```

---

### Step 2: Clone & Setup (5 min)

```bash
# Create project structure
mkdir brain-sentry
cd brain-sentry

mkdir backend frontend

# Backend: Initialize Maven project
cd backend
mvn archetype:generate \
  -DgroupId=com.integraltech \
  -DartifactId=brain-sentry \
  -DarchetypeArtifactId=maven-archetype-quickstart \
  -DinteractiveMode=false

# Frontend: Initialize Next.js
cd ../frontend
npx create-next-app@latest . \
  --typescript \
  --tailwind \
  --app \
  --src-dir \
  --no-git
```

---

### Step 3: Start FalkorDB (2 min)

```bash
# Create docker-compose.yml
cd ..
mkdir docker
cd docker

cat > docker-compose.yml << 'EOF'
version: '3.8'

services:
  falkordb:
    image: falkordb/falkordb:latest
    container_name: brain-sentry-db
    ports:
      - "6379:6379"
    volumes:
      - falkordb-data:/data
    environment:
      - REDIS_PASSWORD=
    restart: unless-stopped

volumes:
  falkordb-data:
EOF

# Start database
docker-compose up -d

# Verify
docker-compose ps
# Should show: brain-sentry-db (running)
```

**✅ Checkpoint:** FalkorDB is running on port 6379

---

### Step 4: Backend "Hello World" (8 min)

```bash
cd ../backend

# Update pom.xml (add Spring Boot parent)
cat > pom.xml << 'EOF'
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
    
    <properties>
        <java.version>17</java.version>
    </properties>
    
    <dependencies>
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-web</artifactId>
        </dependency>
        
        <dependency>
            <groupId>redis.clients</groupId>
            <artifactId>jedis</artifactId>
            <version>5.1.0</version>
        </dependency>
        
        <dependency>
            <groupId>org.projectlombok</groupId>
            <artifactId>lombok</artifactId>
            <scope>provided</scope>
        </dependency>
    </dependencies>
    
    <build>
        <plugins>
            <plugin>
                <groupId>org.springframework.boot</groupId>
                <artifactId>spring-boot-maven-plugin</artifactId>
            </plugin>
        </plugins>
    </build>
</project>
EOF

# Create main application class
mkdir -p src/main/java/com/integraltech/brainsentry
cat > src/main/java/com/integraltech/brainsentry/BrainSentryApplication.java << 'EOF'
package com.integraltech.brainsentry;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.web.bind.annotation.*;

@SpringBootApplication
public class BrainSentryApplication {
    public static void main(String[] args) {
        SpringApplication.run(BrainSentryApplication.class, args);
    }
}

@RestController
@RequestMapping("/api")
class HelloController {
    
    @GetMapping("/health")
    public String health() {
        return "Brain Sentry is alive! 🧠";
    }
    
    @GetMapping("/version")
    public String version() {
        return "v0.1.0";
    }
}
EOF

# Create application.yml
mkdir -p src/main/resources
cat > src/main/resources/application.yml << 'EOF'
server:
  port: 8080

spring:
  application:
    name: brain-sentry
EOF

# Build and run
mvn clean install
mvn spring-boot:run
```

**Open new terminal and test:**
```bash
curl http://localhost:8080/api/health
# Expected: Brain Sentry is alive! 🧠
```

**✅ Checkpoint:** Backend responding on http://localhost:8080

---

### Step 5: Frontend "Hello World" (10 min)

```bash
cd ../frontend

# Install dependencies
npm install

# Install Radix UI basics
npm install @radix-ui/react-dialog \
            @radix-ui/react-slot \
            class-variance-authority \
            clsx \
            tailwind-merge \
            lucide-react

# Create basic layout
cat > src/app/page.tsx << 'EOF'
export default function Home() {
  return (
    <main className="min-h-screen p-8">
      <div className="max-w-4xl mx-auto">
        <h1 className="text-4xl font-bold mb-4">
          🧠 Brain Sentry
        </h1>
        <p className="text-xl text-gray-600 mb-8">
          Intelligent Context Management System
        </p>
        
        <div className="grid grid-cols-3 gap-4">
          <div className="p-6 border rounded-lg">
            <h2 className="text-2xl font-bold mb-2">0</h2>
            <p className="text-gray-600">Memories</p>
          </div>
          
          <div className="p-6 border rounded-lg">
            <h2 className="text-2xl font-bold mb-2">0</h2>
            <p className="text-gray-600">Injections</p>
          </div>
          
          <div className="p-6 border rounded-lg">
            <h2 className="text-2xl font-bold mb-2">Ready</h2>
            <p className="text-gray-600">Status</p>
          </div>
        </div>
      </div>
    </main>
  );
}
EOF

# Run frontend
npm run dev
```

**Open browser:** http://localhost:3000

**✅ Checkpoint:** You should see Brain Sentry homepage!

---

## 🎉 Success!

You now have:
- ✅ FalkorDB running (port 6379)
- ✅ Backend API running (port 8080)
- ✅ Frontend running (port 3000)

---

## 🚀 Next Steps

### Immediate (Next 30 min)

1. **Create First Memory** - Add CRUD endpoint
2. **Connect to FalkorDB** - Store memory in graph
3. **Display in Frontend** - Show memories in UI

### This Week

1. Read **00-PROJECT-OVERVIEW.md** (30 min)
2. Follow **SETUP_GUIDE.md** for full setup (2 hours)
3. Implement Phase 1 Week 1 tasks from **DEVELOPMENT_PHASES.md**

### This Month

Complete **Phase 1: Foundation** (3 weeks)
- Domain models
- CRUD operations
- Basic UI
- FalkorDB integration

---

## 📚 Documentation Reference

```
Quick Reference:
├── README.md                    ← Start here
├── 00-PROJECT-OVERVIEW.md       ← Architecture & roadmap
├── SETUP_GUIDE.md               ← Detailed setup
├── BACKEND_SPECIFICATION.md     ← Backend implementation
├── FRONTEND_SPECIFICATION.md    ← Frontend implementation
├── GRAPH_VISUALIZATION.md       ← Cytoscape.js guide
└── DEVELOPMENT_PHASES.md        ← Week-by-week plan
```

---

## 🐛 Troubleshooting

### Backend won't start
```bash
# Check if port 8080 is free
lsof -i :8080

# Check Java version
java --version  # Must be 17+

# Clean and rebuild
mvn clean install
```

### Frontend won't start
```bash
# Delete node_modules and reinstall
rm -rf node_modules package-lock.json
npm install

# Check if port 3000 is free
lsof -i :3000
```

### FalkorDB won't start
```bash
# Check Docker is running
docker ps

# Restart container
cd docker
docker-compose down
docker-compose up -d

# Check logs
docker-compose logs falkordb
```

---

## 💡 Pro Tips

### Development Workflow

```bash
# Terminal 1: Database
cd docker && docker-compose up

# Terminal 2: Backend
cd backend && mvn spring-boot:run

# Terminal 3: Frontend  
cd frontend && npm run dev

# Terminal 4: Commands
# Use for git, testing, etc
```

### Hot Reload

**Backend:** Spring Boot DevTools (auto-reload on save)
**Frontend:** Next.js Fast Refresh (instant updates)

### Recommended IDE

**Backend:** IntelliJ IDEA Community (free)
**Frontend:** VS Code with extensions:
- ESLint
- Prettier
- Tailwind CSS IntelliSense

---

## ✅ Validation Checklist

Before moving to full development, verify:

- [ ] Java 17+ installed and working
- [ ] Node 18+ installed and working
- [ ] Docker running FalkorDB
- [ ] Maven can build backend
- [ ] Backend responds to `/api/health`
- [ ] Frontend loads in browser
- [ ] No console errors
- [ ] Git repository initialized
- [ ] All documentation downloaded

---

## 🎯 Your 30-Day Roadmap

```
Week 1: Setup + Domain Models
└── Complete: Basic CRUD, Database connected

Week 2: Core Features
└── Complete: Memory management, Search

Week 3: Integration
└── Complete: Backend + Frontend working together

Week 4: Polish
└── Complete: Testing, Documentation, First demo
```

---

**You're ready to build! 🚀**

For detailed implementation, see:
- **BACKEND_SPECIFICATION.md** - Complete backend guide
- **FRONTEND_SPECIFICATION.md** - Complete frontend guide
- **GRAPH_VISUALIZATION.md** - Cytoscape.js implementation

**Questions?** Check **SETUP_GUIDE.md** troubleshooting section.

---

**Status:** ✅ Ready to Start Development  
**Estimated Time:** 30 minutes completed, 18 weeks to V1.0  
**Next Document:** 00-PROJECT-OVERVIEW.md
