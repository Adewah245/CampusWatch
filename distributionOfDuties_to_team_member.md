# CampusWatch — Team Work Distribution

## 1. Overview

CampusWatch is divided into three major engineering areas:

```text
                         CAMPUSWATCH
                              │
              ┌───────────────┼───────────────┐
              ↓               ↓               ↓
          BACKEND           AGENT          FRONTEND
           TEAM              TEAM             TEAM
              │               │               │
              ↓               ↓               ↓
         Go + API         Computer         Dashboard
         PostgreSQL       Monitoring          UI
              │               │               │
              └───────────────┼───────────────┘
                              ↓
                       WORKING SYSTEM
```

The three teams work independently but communicate through agreed API contracts.

---

# 2. Team 1 — Backend Developer

## Role

The Backend Developer is responsible for the **brain of CampusWatch**.

The backend controls the business logic, database, authentication, APIs, monitoring data, and communication between the agent and frontend.

## Main Directory

```text
backend/
├── cmd/
├── internal/
│   ├── auth/
│   ├── institutions/
│   ├── campuses/
│   ├── clusters/
│   ├── locations/
│   ├── systems/
│   ├── agents/
│   ├── heartbeats/
│   ├── sessions/
│   ├── health/
│   ├── events/
│   ├── issues/
│   └── dashboard/
│
├── migrations/
└── go.mod
```

## Responsibilities

### Authentication

* User login
* User logout
* User sessions
* Roles
* Permissions
* Authorization

### Institution Management

* Institutions
* Campuses
* Clusters
* Locations/Tables
* Systems

### Agent Management

* Agent registration
* Agent authentication
* Agent approval
* Agent credentials
* Agent status

### Monitoring

* Heartbeats
* Online/offline detection
* Last seen
* System status

### User Sessions

* Login
* Logout
* Active state
* Idle state
* Session history

### System Health

* CPU
* Memory
* Disk
* Battery
* Network
* Agent health

### Events

* System online
* System offline
* User login
* User logout
* User idle
* User active
* After-hours activity
* Health warnings
* Fault events

### Issues

* Create issues
* Update issues
* Assign issues
* Resolve issues
* Close issues
* Issue history

### Dashboard API

The backend provides the data required by the frontend.

Example:

```text
GET /api/v1/dashboard/summary
```

Possible response:

```text
Total Systems: 100
Online: 87
Offline: 8
Inactive: 5
Open Issues: 12
Active Users: 63
```

## Main Deliverables

```text
Go Backend
      +
REST API
      +
PostgreSQL Database
      +
Authentication
      +
Business Logic
```

---

# 3. Team 2 — Monitoring Agent Developer

## Role

The Agent Developer is responsible for the software installed on each monitored computer.

The agent is the **eyes and ears of CampusWatch**.

## Main Directories

```text
agent/
├── cmd/
├── internal/
│   ├── identity/
│   ├── system/
│   ├── session/
│   ├── activity/
│   ├── health/
│   └── client/
│
└── README.md
```

The agent developer also owns the deployment scripts:

```text
installer/
├── windows/
│   ├── install.ps1
│   └── uninstall.ps1
│
├── linux/
│   ├── install.sh
│   └── uninstall.sh
│
└── README.md
```

## Responsibilities

### System Identity

The agent identifies the computer.

It can collect:

```text
Hostname
Operating System
OS Version
Device Type
```

### System Health

The agent monitors:

```text
CPU
Memory
Disk
Battery
Network
```

### User Session Detection

The agent detects:

```text
USER_LOGIN
USER_LOGOUT
ACTIVE
IDLE
```

The agent must not secretly collect passwords, private messages, browser passwords, cookies, or other unnecessary private information.

### Heartbeat

The agent periodically tells the backend:

```text
"I am still alive."
```

Example:

```text
Agent
   ↓
Heartbeat
   ↓
Backend
```

### Event Reporting

The agent reports approved events such as:

```text
USER_LOGIN
USER_LOGOUT
USER_IDLE
USER_ACTIVE
HEALTH_WARNING
```

### Agent Authentication

Each installed agent must have its own identity and secure credential.

The agent must not contain a permanent shared secret that is identical on every computer.

### Installation

The agent developer is also responsible for making deployment easy.

Instead of:

```text
Computer 1 → Manual configuration
Computer 2 → Manual configuration
Computer 3 → Manual configuration
Computer 4 → Manual configuration
```

CampusWatch should provide:

```text
CampusWatch Installer
        ↓
Install Agent
        ↓
Configure Agent
        ↓
Register System
        ↓
Start Agent
        ↓
Connect to Backend
```

## Windows

Initial deployment:

```text
installer/windows/
├── install.ps1
└── uninstall.ps1
```

## Linux

Initial deployment:

```text
installer/linux/
├── install.sh
└── uninstall.sh
```

## Main Deliverables

```text
Monitoring Agent
       +
System Monitoring
       +
Session Detection
       +
Heartbeat
       +
Secure Agent Communication
       +
Windows Installer
       +
Linux Installer
```

---

# 4. Team 3 — Frontend Developer

## Role

The Frontend Developer builds the **face of CampusWatch**.

This is what administrators and authorized users interact with.

## Main Directory

```text
frontend/
├── pages/
├── assets/
├── css/
└── js/
```

## Responsibilities

### Authentication

Build:

```text
Login
Logout
Session
```

### Dashboard

Display:

```text
Total Systems
Online Systems
Offline Systems
Inactive Systems
Maintenance Systems
Active Users
Open Issues
Health Warnings
After-Hours Activity
```

Example:

```text
┌────────────────────────────────────┐
│       CampusWatch Dashboard        │
├────────────────────────────────────┤
│                                    │
│  100 Systems       87 Online       │
│                                    │
│  8 Offline         5 Inactive      │
│                                    │
│  12 Issues         63 Users       │
│                                    │
└────────────────────────────────────┘
```

### Institution Management

Build interfaces for:

```text
Institutions
Campuses
Clusters
Locations
Systems
```

### System Monitoring

Example:

```text
PC-001

Status: ONLINE
User: James

CPU:       42%
Memory:    67%
Disk:      72%
Battery:   38%
Network:   Connected

Last Heartbeat:
15 seconds ago
```

### User Activity

Display session history:

```text
PC-001

09:15 — James logged in
11:40 — James logged out

11:42 — John logged in
13:05 — John logged out
```

### Issue Management

The frontend should allow administrators/operators to:

```text
Report Issue
View Issue
Assign Issue
Update Issue
Resolve Issue
Close Issue
```

### Reports

Display:

```text
System Uptime
System Downtime
User Activity
After-Hours Usage
Health Problems
Issue History
```

## Main Deliverables

```text
Web Dashboard
      +
Authentication UI
      +
System Monitoring UI
      +
User Activity UI
      +
Issue Management UI
      +
Reports
```

---

# 5. How the Three Teams Connect

The backend is the central communication point.

```text
                         FRONTEND
                            │
                            │
                         REST API
                            │
                            ↓
                     ┌─────────────┐
                     │   BACKEND   │
                     │     Go      │
                     └──────┬──────┘
                            │
                       PostgreSQL
                            │
                            ↑
                         REST API
                            │
                     ┌──────┴──────┐
                     │    AGENT    │
                     │  Computer   │
                     └─────────────┘
```

The three teams should not directly depend on each other's implementation.

They depend on **contracts**.

---

# 6. API Contract

The backend defines how the frontend and agent communicate with CampusWatch.

Each API contract should define:

```text
HTTP Method
URL
Authentication
Request
Response
Required Fields
Optional Fields
Errors
Examples
```

For example:

```text
POST /api/v1/agents/heartbeat
```

The agent developer knows:

> "This is how I send heartbeat information."

The backend developer knows:

> "This is how I receive and process heartbeat information."

The frontend developer knows:

> "I can use the resulting system status to display ONLINE/OFFLINE."

---

# 7. Example — System Goes Offline

### Step 1 — Agent

The agent stops sending heartbeats.

```text
Agent
   X
No heartbeat
```

### Step 2 — Backend

The backend checks the last heartbeat.

```text
Last heartbeat:
> 2 minutes ago
```

The backend changes:

```text
ONLINE → OFFLINE
```

and records the event.

### Step 3 — Frontend

The dashboard receives the updated information.

```text
PC-001

🔴 OFFLINE

Last Seen:
15:32
```

One feature has now involved all three teams.

---

# 8. Example — User Logs In

### Step 1 — Agent

The agent detects:

```text
USER_LOGIN
```

It sends the event:

```text
Agent
   ↓
POST /api/v1/agents/session
   ↓
Backend
```

### Step 2 — Backend

The backend validates and records:

```text
User: James
System: PC-001
Login: 15:40
```

### Step 3 — Frontend

The dashboard displays:

```text
PC-001

Status:
ONLINE

Current User:
James

Session:
Active
```

---

# 9. Work Distribution by Development Phase

## Phase 1 — Foundation

### Backend

```text
Project structure
Database foundation
Authentication
Institution
Campus
Cluster
Location
System registration
Basic API
```

### Agent

```text
Agent project foundation
System identity
Basic system information
```

### Frontend

```text
Frontend foundation
Login page
Basic dashboard
Institution view
System view
```

---

# 10. Phase 2 — Agent and Heartbeat

### Backend

```text
Agent registration
Agent authentication
Heartbeat endpoint
Last-seen tracking
Online/offline detection
```

### Agent

```text
Agent registration
Agent authentication
Heartbeat
Backend communication
```

### Frontend

```text
Online systems
Offline systems
Last seen
System status
```

---

# 11. Phase 3 — User Sessions

### Backend

```text
User identity
Session creation
Login events
Logout events
Active/idle state
Session history
```

### Agent

```text
Login detection
Logout detection
Active detection
Idle detection
Session reporting
```

### Frontend

```text
Current user
Session status
Session history
Usage duration
```

---

# 12. Phase 4 — System Health

### Backend

```text
Health data
Health events
Thresholds
Health warnings
```

### Agent

```text
CPU monitoring
Memory monitoring
Disk monitoring
Battery monitoring
Network monitoring
Agent health
```

### Frontend

```text
CPU display
Memory display
Disk display
Battery display
Network status
Health warnings
```

---

# 13. Phase 5 — Issues

### Backend

```text
Issue creation
Issue updates
Issue assignment
Issue status
Issue history
```

### Agent

```text
Health-triggered events
Fault detection
System warnings
```

### Frontend

```text
Report issue
View issue
Assign issue
Update issue
Resolve issue
Close issue
```

---

# 14. Phase 6 — Reports

### Backend

```text
Usage statistics
Uptime
Downtime
User activity
After-hours activity
Issue statistics
Health statistics
```

### Agent

Provides the raw monitoring data required for reports.

### Frontend

Displays:

```text
Charts
Tables
Statistics
Reports
History
```

---

# 15. Phase 7 — Advanced Features

Possible future work:

```text
Automatic issue detection
Alerts
Notifications
Configurable thresholds
Real-time dashboard updates
WebSockets
Agent updates
Large-scale deployment
Advanced analytics
```

---

# 16. First Team Milestone

The team should not attempt to build every feature immediately.

The first goal is to create one complete working flow.

```text
Institution
     ↓
Campus
     ↓
Cluster
     ↓
Location
     ↓
System
     ↓
Agent
     ↓
Backend
     ↓
Dashboard
```

The first successful result should be:

> **A real computer appears on the CampusWatch dashboard and shows that it is ONLINE.**

---

# 17. First Vertical Slice

```text
                    CAMPUSWATCH
                         │
        ┌────────────────┼────────────────┐
        ↓                ↓                ↓
    FRONTEND          BACKEND           AGENT
        │                │                │
      Login          Register          Start
        │             System           Agent
        │                │                │
        │                ←───────────────┤
        │             Register           │
        │                                │
        │              Approve           │
        │                System          │
        │                                │
        │                ←───────────────┤
        │             Heartbeat          │
        │                                │
        └───────────────→                 │
             Display ONLINE               │
```

At this point, the three teams have successfully connected their work.

---

# 18. Team Rule

The three developers should follow this principle:

```text
Backend = Brain
Agent   = Eyes and Ears
Frontend = Face
```

But none of the three should work in isolation.

```text
Backend ↔ Agent
Backend ↔ Frontend
Agent   ↔ Backend
```

The backend is the central integration point.

---

# 19. Definition of Done

A feature should not be considered complete simply because one developer finished their code.

For example, **heartbeat** is complete only when:

```text
Agent
  ↓
Sends heartbeat
  ↓
Backend
  ↓
Stores/processes heartbeat
  ↓
Determines system status
  ↓
Frontend
  ↓
Displays system status
```

All three parts must work together.

---

# 20. Team Ownership Summary

| Area                | Backend | Agent         | Frontend |
| ------------------- | ------- | ------------- | -------- |
| Authentication      | ✅       | —             | ✅        |
| Institution         | ✅       | —             | ✅        |
| Campus              | ✅       | —             | ✅        |
| Cluster             | ✅       | —             | ✅        |
| Location            | ✅       | —             | ✅        |
| System Registration | ✅       | ✅             | ✅        |
| Agent Registration  | ✅       | ✅             | —        |
| Heartbeat           | ✅       | ✅             | ✅        |
| User Sessions       | ✅       | ✅             | ✅        |
| System Health       | ✅       | ✅             | ✅        |
| Events              | ✅       | ✅             | ✅        |
| Issues              | ✅       | Partial       | ✅        |
| Reports             | ✅       | Provides data | ✅        |
| Installer           | —       | ✅             | —        |
| Database            | ✅       | —             | —        |
| REST API            | ✅       | Consumes      | Consumes |

---

# 21. Final Architecture

```text
                         CAMPUSWATCH
                              │
             ┌────────────────┼────────────────┐
             │                │                │
             ↓                ↓                ↓
          BACKEND           AGENT           FRONTEND
             │                │                │
             │                │                │
             │          Installed on           │
             │           computers             │
             │                │                │
             │                │                │
             └────────── API CONTRACTS ────────┘
                              │
                              ↓
                         POSTGRESQL
```

The objective is to build these three pieces into **one complete product**, not three disconnected projects.

---

# 22. Recommended First Work

The team should start with:

```text
1. Finalize architecture
        ↓
2. Finalize API contracts
        ↓
3. Finalize database foundation
        ↓
4. Backend foundation
        ↓
5. Agent foundation
        ↓
6. Frontend foundation
        ↓
7. Connect first vertical slice
        ↓
8. Test with a real computer
```

Only after this foundation works should we move into the more advanced monitoring features.
