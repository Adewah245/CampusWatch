with all the steps from
```md
# 1. Backend foundation

backend/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── config/
│   ├── database/
│   ├── server/
│   └── health/
├── migrations/
├── go.mod
└── .env.example

We implement
1. Go project initialization
2. Environment configuration
3. PostgreSQL connection
4. HTTP server
5. Basic routing
6. Logging
7. Graceful shutdown
8. /health endpoint
End of day
```
# 2. PostgreSQL
```md

001_users.sql
002_roles.sql
003_institutions.sql
004_campuses.sql
005_clusters.sql
006_locations.sql
007_systems.sql
008_agents.sql

***day2** We create*
1. Tables
2. Primary keys
4. Foreign keys
5. Unique constraints
7. Indexes
8. Created/updated timestamps
9. Status fields
# 3. Authentication/RBAC

```md

internal/
└── auth/
    ├── handler.go
    ├── service.go
    ├── repository.go
    ├── middleware.go
    └── models.go
**API
POST /api/v1/auth/login
POST /api/v1/auth/logout
GET  /api/v1/auth/me

**We implement
1. User login
2. Password hashing
3. Session/token handling
4. Authentication middleware
5. Role checking
6. Institution access checking
End of day
```

# 4. Institution hierarchy
```md

internal/
├── institutions/
├── campuses/
├── clusters/
└── locations/
*APIs
/institutions
/campuses
/clusters
/locations
*ADMIN
Create Institution
       ↓
Create Campus
       ↓
Create Cluster
       ↓
Create Location/Table
```

# 5. System registration
```md

internal/systems/
├── handler.go
├── service.go
├── repository.go
├── models.go
└── validation.go

### system registration
PC-001
Hostname
Operating System
Device Type
Location
Agent Version
### system registration status
PENDING
ACTIVE
SUSPENDED
RETIRED
### monitor status
ONLINE
OFFLINE
INACTIVE
MAINTENANCE
SUSPENDED
RETIRED

### APIs
GET   /api/v1/systems
GET   /api/v1/systems/{id}
PATCH /api/v1/systems/{id}
POST  /api/v1/systems/{id}/approve
```
# Secure agent registration
```md
internal/agents/
├── handler.go
├── service.go
├── repository.go
├── models.go
└── credentials.go
Agent
 ↓
Registration
 ↓
Unique agent identity
 ↓
System association
 ↓
PENDING
 ↓
Admin approval
 ↓
Credential
```

# Heartbeat

```md
internal/heartbeats/
### API
POST /api/v1/agents/heartbeat
agent send:
Agent ID
System ID
Timestamp
Agent Version
Health information
backend stores
last_seen_at
Heartbeat received
       ↓
Update last_seen_at
       ↓
ONLINE
No heartbeat
       ↓
Threshold exceeded
       ↓
OFFLINE
example: PC-001
ONLINE
Last seen: 10 seconds ago
```
# User sessions
```md
internal/sessions/
### session state
LOGGED_OUT
LOGGED_IN_ACTIVE
LOGGED_IN_IDLE

We implement
User identification
Login event
Logout event
Active state
Idle state
Session duration
Session history
Example
PC-001

09:15 James logged in
11:40 James logged out

11:42 John logged in
13:05 John logged out
Important

We will first agree on how CampusWatch identifies the actual human when the OS account is simply student.
```

# System health
```md
We create
internal/health/
Agent reports
CPU
Memory
Disk
Battery
Network
OS
Agent health
System uptime
Example
PC-001

CPU:       45%
Memory:    62%
Disk:      81%
Battery:   74%
Network:   Connected
Agent:     Healthy
```
# Events
```md
We create
internal/events/
internal/schedules/
Events
SYSTEM_REGISTERED
SYSTEM_APPROVED
SYSTEM_ONLINE
SYSTEM_OFFLINE
USER_LOGIN
USER_LOGOUT
USER_IDLE
USER_ACTIVE
AFTER_HOURS_USAGE
HEALTH_WARNING
FAULT_DETECTED
ISSUE_REPORTED
Schedule

We support:

Opening time
Working hours
Break
Closing time
Example
08:00 → Open
10:30 → Break
11:00 → Resume
17:00 → Close
End of day

CampusWatch can determine whether usage happened during or outside allowed hours.
```

# After-hours monitoring
# Issues
```md
We create
internal/issues/
Issue states
OPEN
ACKNOWLEDGED
IN_PROGRESS
RESOLVED
CLOSED
API
POST  /api/v1/issues
GET   /api/v1/issues
GET   /api/v1/issues/{id}
PATCH /api/v1/issues/{id}
POST  /api/v1/issues/{id}/updates
Example
PC-004
Issue:
High disk usage

Priority:
HIGH

Status:
OPEN
End of day
```

# Dashboard APIs
```md
Now we give the frontend team the data they need.

We create
internal/dashboard/
internal/reports/
APIs
GET /api/v1/dashboard/summary
GET /api/v1/dashboard/systems
GET /api/v1/dashboard/issues
Dashboard can request
Total systems
Online
Offline
Inactive
Maintenance
Active users
Open issues
After-hours usage
Reports
Daily usage
Weekly usage
Monthly usage
Uptime
Downtime
User sessions
Issues
After-hours activity
```

# Reports
# Windows installer
# Linux installer
```md
Now we build the deployment system.

installer/
├── windows/
│   ├── install.ps1
│   └── uninstall.ps1
│
└── linux/
    ├── install.sh
    └── uninstall.sh
Windows installer

It should:

Run installer
 ↓
Install Agent
 ↓
Create configuration
 ↓
Register system
 ↓
Store credential securely
 ↓
Create service/startup
 ↓
Start Agent
 ↓
Test connection
Linux installer

Same concept:

Install
 ↓
Configure
 ↓
Register
 ↓
Secure credential
 ↓
Create service
 ↓
Start
 ↓
Verify
```
# ##Security
#### Testing
#### Integration
#### Final MVP checklist
```md
This is our final MVP testing day.

We test the complete chain:

Windows/Linux Computer
        ↓
Installer
        ↓
Agent
        ↓
Agent Registration
        ↓
Backend
        ↓
PostgreSQL
        ↓
Heartbeat
        ↓
ONLINE
        ↓
User Session
        ↓
Health
        ↓
Events
        ↓
Issues
        ↓
Dashboard API
Security review

Check:

Authentication
Authorization
Institution isolation
Agent credentials
Password hashing
Input validation
Rate limiting
Secrets
Audit logs
HTTPS
Error handling
```
```md
 WEEK 3 — Buffer / Real-World Testing

I don't want us to pretend everything will work perfectly in 14 days.

Week 3 is for:

Bug fixing
Cross-platform testing
Installer testing
Security fixes
Database improvements
API improvements
Agent/backend problems
Documentation
Deployment testing

This gives us breathing room.
```