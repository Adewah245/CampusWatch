```md
<p align="center">
  <img src="Adewahlog.png" width="180" alt="Adewah Logo">
</p>
```
# CampusWatch

> **Monitor. Understand. Protect Every System.**

CampusWatch is a multi-institution computer monitoring and management platform designed for schools, colleges, training centres, laboratories, and other institutions.

It allows an institution to monitor its computer systems, know when systems are online or offline, identify who is using a system, monitor system health, record login/logout activity, manage issues, and generate useful reports.

CampusWatch is built to support **multiple institutions from one platform**, while keeping each institution's data isolated and protected.

---

# 0. Quick Start

**Prerequisites:** Go 1.26+, Python 3.9+ (for the development dashboard server).

Every command below runs from the **repository root**.

```bash
# 1. Configure the database. Copy the example and fill in your connection string.
cp backend/.env.example backend/.env
#    then edit backend/.env
```

```bash
# 2. Create the first operator account. This also creates the institution, and
#    applies the database migrations, so it is safe to run against a fresh
#    database.
CAMPUSWATCH_PASSWORD='choose-a-strong-password' \
go run ./backend/cmd/createuser \
    --email admin@adewah.edu \
    --institution 'Adewah University'
```

```bash
# 3. Start the API and the dashboard, in two terminals.
go run ./backend/cmd/server      # terminal 1 - API on :8080
python3 frontend/serve.py        # terminal 2 - dashboard on :8000
```

Then open <http://localhost:8000/> and sign in with the account from step 2.

`serve.py` serves the dashboard and forwards `/api/*` to the backend, so the browser sees a
single origin and CORS never applies. It is a development convenience — see
[frontend/README.md](frontend/README.md) for the production arrangement.

> **Sign-in returns 401?** That means no account exists yet, not that the password is wrong.
> Run step 2 and try again. The dashboard needs an account with the `admin` or `manager`
> role; every route is restricted to those two.

> **Backend exits with "migrations directory not found"?** Run it from the repository root,
> as shown, or set `MIGRATIONS_DIR` explicitly.

---

# 1. What is CampusWatch?

CampusWatch connects three major parts:

```text
CampusWatch
     │
     ├── Monitoring Agent
     │       │
     │       └── Installed on computers
     │
     ├── Go Backend/API
     │       │
     │       ├── Authentication
     │       ├── Monitoring
     │       ├── Users
     │       ├── Sessions
     │       ├── Issues
     │       └── Reports
     │
     ├── PostgreSQL Database
     │
     └── Web Dashboard
```

The monitoring agent collects approved operational information from a computer and communicates with the CampusWatch backend.

The backend processes and stores the information.

The dashboard allows authorized administrators and operators to understand what is happening across their institution.

---

# 2. Main Goals

CampusWatch should help an institution answer questions such as:

* Which computers are currently online?
* Which computers are offline?
* Which computers are currently being used?
* Who is using a particular computer?
* When did a user log in?
* When did a user log out?
* How long was the system used?
* Which computers are idle?
* Which systems have health problems?
* Which systems need maintenance?
* Which systems were used after working hours?
* What happened before a system went offline?
* What issues have been reported?
* Which issues have been resolved?
* How much uptime and downtime has a system had?
* What systems require attention?

---

# 3. Product Hierarchy

CampusWatch uses a clear physical hierarchy.

```text
CampusWatch
    ↓
Institution
    ↓
Campus
    ↓
Cluster
    ↓
Location / Table
    ↓
System
    ↓
Agent
```

Example:

```text
CampusWatch
│
├── Institution A
│   │
│   ├── Main Campus
│   │   │
│   │   ├── Cluster 1
│   │   │   │
│   │   │   ├── Table 1
│   │   │   │   ├── PC-001
│   │   │   │   └── PC-002
│   │   │   │
│   │   │   └── Table 2
│   │   │       ├── PC-003
│   │   │       └── PC-004
│   │
│   └── Other Campus
│
├── Institution B
│   └── Main Campus
│       └── Computer Laboratory
│
└── Institution C
    └── Main Campus
        └── Computer Laboratory
```

### Important

A table/location is a **physical grouping or location**, not a computer.

The final database naming for this concept will be decided during schema design.

---

# 4. Multi-Institution Support

CampusWatch is not designed for only one school.

It should support:

```text
Institution A
Institution B
Institution C
Institution D
...
```

Each institution must have isolated data.

For example:

```text
Institution A
    ├── Users
    ├── Campuses
    ├── Systems
    ├── Sessions
    └── Issues

Institution B
    ├── Users
    ├── Campuses
    ├── Systems
    ├── Sessions
    └── Issues
```

Institution A must not be able to access Institution B's data unless an authorized platform-level role explicitly permits it.

Multi-tenancy must therefore be considered in:

* Database design
* Authentication
* Authorization
* API
* Dashboard
* Agent registration
* Reporting
* Audit logs

---

# 5. Core Components

## 5.1 Monitoring Agent

The CampusWatch Agent runs on each monitored computer.

Responsibilities include:

* Identify the system
* Report operating system information
* Report agent version
* Monitor CPU
* Monitor memory
* Monitor disk
* Monitor battery where available
* Monitor network connectivity
* Detect system activity/idle state
* Detect login/logout events
* Report user session information
* Send heartbeat
* Report agent health
* Report approved system events
* Communicate securely with the backend

The agent should operate as a managed monitoring component.

---

# 6. Privacy Boundary

CampusWatch is intended for legitimate institutional system monitoring.

The agent should collect only information required for the platform's approved purpose.

### Information CampusWatch may collect

```text
System ID
Hostname
Operating System
OS Version
Agent Version
CPU Usage
Memory Usage
Disk Usage
Battery Status
Network Status
Current Session Identity
Login Events
Logout Events
Active/Idle State
Heartbeat
Agent Health
System Health Events
```

### CampusWatch must NOT collect

```text
Passwords
Keystrokes
Private messages
Browser passwords
Browser cookies
Secret credentials
Private files
Unnecessary browsing history
Screenshots by default
```

The exact information collected should always be clearly defined by the institution's monitoring policy.

---

# 7. User Identity

One important problem CampusWatch must solve is identifying the actual person using a computer.

For example, a laboratory may have:

```text
PC-001
OS username: student
```

If every student uses the same OS account, the operating-system username cannot tell CampusWatch who the actual person is.

Therefore CampusWatch uses two separate concepts:

```text
Operating System Identity
        +
CampusWatch / Institution Identity
```

Example:

```text
Computer:
PC-001

OS user:
student

CampusWatch user:
James

Session:
LOGGED_IN_ACTIVE

Login:
09:15
```

This allows CampusWatch to maintain history such as:

```text
PC-001

09:15 — James logged in
11:40 — James logged out

11:42 — John logged in
13:05 — John logged out
```

Possible identity mechanisms include:

* CampusWatch account
* Student/staff account
* School email
* Student ID
* Institutional SSO
* AD/LDAP
* Google/Microsoft school identity

The final identity mechanism will be decided during authentication architecture design.

CampusWatch must not secretly inspect browser cookies, saved passwords, or private profiles to identify users.

---

# 8. System Status

CampusWatch separates **system status** from **user session status**.

System status:

```text
PENDING
ONLINE
OFFLINE
INACTIVE
MAINTENANCE
SUSPENDED
RETIRED
```

### PENDING

The system has been registered but has not yet been approved.

### ONLINE

The agent is communicating normally with CampusWatch.

### OFFLINE

The backend has not received communication within the configured offline threshold.

### INACTIVE

The computer is available but is not currently being used.

### MAINTENANCE

The computer has intentionally been placed into maintenance mode.

### SUSPENDED

Normal monitoring/use has been administratively suspended.

### RETIRED

The computer has permanently been removed from service.

---

# 9. User Session Status

User session status should remain separate from system status.

Recommended session states:

```text
LOGGED_OUT
LOGGED_IN_ACTIVE
LOGGED_IN_IDLE
```

Example:

```text
System Status:
ONLINE

User Session:
LOGGED_OUT
```

This means:

> The computer is online, but nobody is currently using it.

Another example:

```text
System Status:
ONLINE

User Session:
LOGGED_IN_ACTIVE
```

This means:

> The computer is online and someone is actively using it.

---

# 10. Registration and Approval

CampusWatch should not automatically trust every computer that contacts the server.

The intended flow is:

```text
Install CampusWatch Agent
        ↓
Agent contacts backend
        ↓
New system created as PENDING
        ↓
Administrator reviews system
        ↓
Administrator approves system
        ↓
Assign Campus
        ↓
Assign Cluster
        ↓
Assign Location / Table
        ↓
System becomes ACTIVE
        ↓
Monitoring begins
```

System registration states:

```text
PENDING
ACTIVE
SUSPENDED
RETIRED
```

---

# 11. System Identity

Each computer needs a stable CampusWatch identity.

Example:

```text
System ID:
SYS-000001

System Code:
LAB-A-PC-001

Hostname:
LAB-A-01

Device Type:
Desktop

Operating System:
Linux

Campus:
Main Campus

Cluster:
Cluster A

Location:
Table 1
```

CampusWatch should not depend entirely on:

* IP address
* Hostname
* MAC address

because these values can change.

The backend should maintain a stable internal system identity.

---

# 12. Agent Authentication

Agents must authenticate when communicating with the backend.

CampusWatch should not accept random unauthenticated heartbeat requests.

Conceptual flow:

```text
Agent Installation
        ↓
Agent Registration
        ↓
System Approval
        ↓
Server provides agent identity/credential
        ↓
Agent securely stores credential
        ↓
Authenticated requests
        ↓
Heartbeat / Sessions / Health / Events
```

Agent credentials must:

* Be unique per installation
* Not be hard-coded into the source code
* Be stored securely
* Be revocable
* Be replaceable/rotatable when necessary

The exact provisioning mechanism will be finalized during the agent security architecture phase.

---

# 13. Heartbeat System

The monitoring agent periodically sends a heartbeat to the backend.

Initial proposed configuration:

```text
Heartbeat interval:
30 seconds

Offline threshold:
2 minutes
```

These values should be configurable.

The backend determines whether a system is:

```text
ONLINE
```

or

```text
OFFLINE
```

rather than trusting the agent to declare its own status.

Initial communication can use normal HTTP requests.

Real-time technologies such as WebSockets can be introduced later if required.

---

# 14. Operating Schedule

CampusWatch should support institutional working schedules.

A schedule may contain:

```text
Opening Time
Working Hours
Break Time
Closing Time
```

Example:

```text
Opening:
08:00

Break:
12:00 - 13:00

Closing:
17:00
```

The exact scope of schedules will be decided during database architecture:

```text
Institution
    ↓
Campus
    ↓
Cluster
    ↓
Location
```

depending on the institution's requirements.

---

# 15. After-Hours Monitoring

CampusWatch should be able to identify system use outside approved working hours.

Example:

```text
Closing time:
17:00

System:
PC-001

User:
James

Activity:
LOGIN

Time:
19:32
```

CampusWatch can record:

```text
AFTER_HOURS_USAGE
```

This allows administrators to determine:

* Who used a system after closing?
* Which system was used?
* When was it used?
* Where was the system located?
* What happened before/after the event?

---

# 16. Events

CampusWatch maintains an event history.

Possible events include:

```text
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
ISSUE_UPDATED
ISSUE_RESOLVED
```

Events provide historical context.

Example:

```text
17:00 — Working hours ended
17:05 — PC-001 remained online
17:20 — User logged in
17:20 — After-hours usage detected
18:10 — User logged out
18:11 — System remained online
```

---

# 17. System Health Monitoring

CampusWatch should monitor operational system health.

Initial metrics:

```text
CPU
Memory
Disk
Battery
Network
Agent Health
```

Example:

```text
PC-001

CPU:
72%

Memory:
81%

Disk:
91%

Battery:
38%

Network:
CONNECTED

Agent:
HEALTHY
```

CampusWatch may generate warnings when configured thresholds are reached.

Examples:

```text
Disk usage is critically high
Battery is critically low
Agent stopped communicating
Network connection lost
System health threshold exceeded
```

CampusWatch should distinguish between a **health warning** and a confirmed hardware failure.

---

# 18. Issue Management

Administrators/operators should be able to report problems directly against a system.

Example:

```text
System:
PC-001

Issue Type:
Keyboard

Title:
Keyboard not responding

Description:
Several keys are not working.

Priority:
HIGH
```

The backend should automatically attach:

```text
System
Location
Reporter
Timestamp
```

Issues may be manually reported or automatically generated from serious system conditions.

---

# 19. Issue Lifecycle

Issues use the following lifecycle:

```text
OPEN
   ↓
ACKNOWLEDGED
   ↓
IN_PROGRESS
   ↓
RESOLVED
   ↓
CLOSED
```

Issue history should be preserved.

Example:

```text
10:15 — Issue reported
10:20 — Administrator acknowledged
10:30 — Technician started work
11:05 — Problem fixed
11:10 — Issue closed
```

---

# 20. Availability and Uptime

CampusWatch should eventually calculate:

* Current availability
* Daily uptime
* Weekly uptime
* Monthly uptime
* Downtime duration
* Number of outages
* Maintenance periods
* Offline periods

Example:

```text
PC-001

Daily Uptime:
7h 42m

Downtime:
18m

Availability:
96.2%
```

Maintenance and intentionally powered-off systems should have explicit rules so that they are not incorrectly counted as failures.

---

# 21. Installation and Deployment

CampusWatch must support large-scale deployment.

An institution should not have to manually configure every computer one by one.

The project therefore includes an **Agent Installation and Deployment system**.

Target operating systems:

```text
Windows
Linux
```

Initial installer structure:

```text
installer/
│
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

---

# 22. Agent Installation Flow

The intended deployment process is:

```text
Administrator gets CampusWatch installer
        ↓
Runs installer
        ↓
Installer installs Agent
        ↓
Creates Agent directory
        ↓
Configures Agent
        ↓
Registers system
        ↓
Connects to CampusWatch backend
        ↓
System appears as PENDING
        ↓
Administrator approves system
        ↓
Agent receives/uses its unique credential
        ↓
Agent starts monitoring
```

The installer should eventually handle:

1. Installing the agent
2. Creating the required directories
3. Creating configuration
4. Registering the system
5. Establishing agent identity
6. Securely storing credentials
7. Configuring automatic startup
8. Starting the agent
9. Checking that the agent is running
10. Verifying communication with CampusWatch

---

# 23. Windows Deployment

Windows deployment will initially use PowerShell.

```text
installer/windows/
├── install.ps1
└── uninstall.ps1
```

The installer should eventually be able to:

```text
Install Agent
     ↓
Configure Agent
     ↓
Register System
     ↓
Configure Windows Service
     ↓
Start Service
     ↓
Verify Connection
```

The exact Windows service and credential-provisioning implementation will be finalized during agent architecture.

---

# 24. Linux Deployment

Linux deployment will initially use shell scripts.

```text
installer/linux/
├── install.sh
└── uninstall.sh
```

The installer should eventually be able to:

```text
Install Agent
     ↓
Configure Agent
     ↓
Register System
     ↓
Configure Linux Service
     ↓
Start Service
     ↓
Verify Connection
```

A service manager such as systemd may be used where supported.

---

# 25. Why Deployment is Part of the Architecture

Deployment is not an afterthought.

CampusWatch may eventually monitor:

```text
10 computers
100 computers
1,000 computers
10,000+ computers
```

Therefore installation must be designed for scale.

The goal is:

```text
One approved deployment process
        ↓
Many systems
        ↓
Each system gets its own identity
        ↓
Each system communicates securely
        ↓
All systems appear in the correct institution
```

---

# 26. Backend/API

The backend is written in Go.

Initial API version:

```text
/api/v1/
```

The backend is responsible for:

* Authentication
* Authorization
* Institution management
* Campus management
* Cluster management
* Location management
* System registration
* Agent authentication
* Heartbeats
* User sessions
* Health monitoring
* Events
* Issues
* Dashboard data
* Reports

---

# 27. Authentication API

```text
POST /api/v1/auth/login
POST /api/v1/auth/logout
GET  /api/v1/auth/me
```

Authentication architecture will support appropriate institutional roles.

---

# 28. Institution API

```text
GET  /api/v1/institutions
POST /api/v1/institutions
GET  /api/v1/institutions/{id}
```

---

# 29. Campus API

```text
GET  /api/v1/campuses
POST /api/v1/campuses
GET  /api/v1/campuses/{id}
```

---

# 30. Cluster API

```text
GET /api/v1/clusters
GET /api/v1/clusters/{id}
```

---

# 31. System API

```text
GET   /api/v1/systems
GET   /api/v1/systems/{id}
PATCH /api/v1/systems/{id}

POST /api/v1/systems/{id}/approve
```

---

# 32. Agent API

```text
POST /api/v1/agents/register
POST /api/v1/agents/heartbeat
POST /api/v1/agents/session
```

Additional agent endpoints may be added as the agent architecture becomes more complete.

---

# 33. Dashboard API

```text
GET /api/v1/dashboard/summary
GET /api/v1/dashboard/systems
GET /api/v1/dashboard/issues
```

The dashboard should be able to show:

```text
Total Systems
Online Systems
Offline Systems
Inactive Systems
Systems in Maintenance
Active Users
Open Issues
Critical Health Warnings
After-Hours Activity
```

---

# 34. Issue API

```text
POST  /api/v1/issues
GET   /api/v1/issues
GET   /api/v1/issues/{id}
PATCH /api/v1/issues/{id}

POST /api/v1/issues/{id}/updates
```

---

# 35. API Contract

The backend, frontend, and agent communicate through agreed API contracts.

Each contract should define:

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

Example:

```text
POST /api/v1/agents/heartbeat
```

Request:

```text
System ID
Agent ID
Timestamp
Health information
```

Response:

```text
Success
Server time
Next expected heartbeat
Configuration information
```

The implementation language does not matter as long as the contract is respected.

---

# 36. Database

CampusWatch will use PostgreSQL as the primary database.

Conceptual structure:

```text
institutions
    ↓
campuses
    ↓
clusters
    ↓
locations
    ↓
systems
    ↓
agents
    ├── heartbeats
    ├── sessions
    ├── health events
    ├── system events
    └── issues

users
    ├── sessions
    ├── issues
    ├── issue updates
    └── administrative actions
```

This is a conceptual model.

Actual migrations will be designed only after the team agrees on:

* Tenant boundaries
* Roles
* User identity
* System identity
* Agent registration
* Schedule scope
* Session rules
* Event retention
* Issue history
* Audit requirements

---

# 37. Security

CampusWatch must be designed with security from the beginning.

Minimum security requirements:

```text
HTTPS in production
Secure password hashing
User authentication
Agent authentication
Role-based access control
Input validation
Rate limiting where appropriate
Audit logs
Secure secret management
Institution-level data isolation
Credential revocation
Credential rotation where required
```

Never commit secrets into Git.

Examples of secrets that must not be committed:

```text
Database passwords
API keys
Agent credentials
JWT secrets
Private keys
Production configuration
```

---

# 38. Repository Structure

```text
campuswatch/
│
├── backend/
│   ├── cmd/
│   │
│   ├── internal/
│   │   ├── auth/
│   │   ├── institutions/
│   │   ├── campuses/
│   │   ├── clusters/
│   │   ├── locations/
│   │   ├── systems/
│   │   ├── agents/
│   │   ├── heartbeats/
│   │   ├── sessions/
│   │   ├── health/
│   │   ├── events/
│   │   ├── issues/
│   │   └── dashboard/
│   │
│   ├── migrations/
│   └── go.mod
│
├── agent/
│   ├── cmd/
│   │
│   ├── internal/
│   │   ├── identity/
│   │   ├── system/
│   │   ├── session/
│   │   ├── activity/
│   │   ├── health/
│   │   └── client/
│   │
│   └── README.md
│
├── installer/
│   ├── windows/
│   │   ├── install.ps1
│   │   └── uninstall.ps1
│   │
│   ├── linux/
│   │   ├── install.sh
│   │   └── uninstall.sh
│   │
│   └── README.md
│
├── frontend/
│   ├── pages/
│   ├── assets/
│   ├── css/
│   └── js/
│
├── docs/
│   ├── ARCHITECTURE.md
│   ├── API.md
│   ├── DATABASE.md
│   ├── AGENT.md
│   ├── SECURITY.md
│   └── CONTRIBUTING.md
│
├── .env.example
├── .gitignore
└── README.md
```

---

# 39. Team Responsibilities

CampusWatch can be developed by three major engineering roles.

## Backend Developer

Responsible for:

* Go backend
* Authentication
* Authorization
* Institutions
* Campuses
* Clusters
* Locations
* Systems
* Agent registration
* Agent authentication
* Heartbeats
* Sessions
* Events
* Issues
* Reports
* REST API
* Database integration

---

## Monitoring Agent Developer

Responsible for:

* Agent application
* System identity
* OS information
* Current session
* Login/logout detection
* Active/idle detection
* CPU monitoring
* Memory monitoring
* Disk monitoring
* Battery monitoring
* Network monitoring
* Agent health
* Heartbeats
* Backend communication

The agent can be written in Go or another suitable language as long as it follows the CampusWatch API contract.

---

## Frontend Developer

Responsible for:

* Login
* Dashboard
* Institution view
* Campus view
* Cluster view
* Location/system view
* System status
* User sessions
* System health
* Issue management
* Reports
* Notifications
* Responsive UI

---

# 40. Communication Between Team Members

The team should communicate through **API contracts**, not implementation details.

For example:

```text
Frontend
    ↓
POST /api/v1/auth/login
    ↓
Backend
```

and:

```text
Agent
    ↓
POST /api/v1/agents/heartbeat
    ↓
Backend
```

The frontend does not need to know how the backend stores data.

The agent does not need to know how the dashboard is implemented.

The backend does not need to know how the frontend renders a system card.

Everyone follows the agreed contract.

---

# 41. Development Phases

## Phase 1 — Foundation

Build:

* Project structure
* Database foundation
* Authentication
* Institution management
* Campus management
* Cluster management
* Location management
* System registration
* Basic dashboard

---

## Phase 2 — Agent and Heartbeat

Build:

* Monitoring agent
* Agent identity
* Agent authentication
* Heartbeat
* Last-seen tracking
* Online/offline detection
* Installation scripts

---

## Phase 3 — User Sessions

Build:

* User identity
* Login detection
* Logout detection
* Active/idle detection
* Session history
* User/system relationship

---

## Phase 4 — System Health

Build:

* CPU monitoring
* Memory monitoring
* Disk monitoring
* Battery monitoring
* Network monitoring
* Agent health
* Health events
* Configurable thresholds

---

## Phase 5 — Issue Management

Build:

* Issue reporting
* Issue assignment
* Issue updates
* Issue history
* Issue lifecycle
* Automatic issue creation where appropriate

---

## Phase 6 — Dashboard and Reports

Build:

* Statistics
* Uptime
* Downtime
* Usage history
* After-hours activity
* User history
* Issue reports
* Health reports

---

## Phase 7 — Advanced Features

Possible future features:

* Automatic issue detection
* Alerts
* Notifications
* Configurable monitoring thresholds
* Real-time dashboard updates
* WebSockets
* Agent updates
* Advanced analytics
* Large-scale deployment tools

---

# 42. MVP Vertical Slice

The first working version should prove the complete core flow.

```text
Institution registered
        ↓
Campus created
        ↓
Cluster created
        ↓
Location/Table created
        ↓
System registered
        ↓
Agent installed
        ↓
Agent connects
        ↓
System appears as PENDING
        ↓
Administrator approves
        ↓
System becomes ACTIVE
        ↓
Heartbeat begins
        ↓
System becomes ONLINE
        ↓
User identifies
        ↓
Session recorded
        ↓
System health monitored
        ↓
Issue reported
        ↓
Issue handled
        ↓
Issue resolved
```

This is more important than building many disconnected features.

---

# 43. Future Scalability

CampusWatch should be designed so that it can grow from:

```text
One institution
```

to:

```text
Many institutions
```

and from:

```text
10 computers
```

to:

```text
Thousands of computers
```

without requiring a complete redesign.

Important scalability considerations include:

* Multi-tenancy
* Database indexing
* Efficient heartbeat processing
* Event storage
* Session history
* API rate limiting
* Agent authentication
* Background processing
* Monitoring thresholds
* Data retention
* Deployment automation

---

# 44. Architecture Decision Process

Before making major architectural or database changes, the team should follow this process:

```text
1. Current Structure
        ↓
2. Proposed Change
        ↓
3. Reason
        ↓
4. Alternatives
        ↓
5. Trade-offs
        ↓
6. Impact
        ↓
7. Team Agreement
        ↓
8. Implementation
```

This prevents unnecessary technical debt.

We should not create tables, endpoints, services, or features simply because they sound useful.

Every major component should have a clear responsibility.

---

# 45. Current Important Architecture Decisions

The following decisions are currently established:

### Product

```text
CampusWatch
```

### Backend

```text
Go
```

### Database

```text
PostgreSQL
```

### API

```text
REST
/api/v1/
```

### Monitoring

```text
Managed Monitoring Agent
```

### System hierarchy

```text
Institution
    ↓
Campus
    ↓
Cluster
    ↓
Location / Table
    ↓
System
```

### System status

```text
PENDING
ONLINE
OFFLINE
INACTIVE
MAINTENANCE
SUSPENDED
RETIRED
```

### Session status

```text
LOGGED_OUT
LOGGED_IN_ACTIVE
LOGGED_IN_IDLE
```

### Issue status

```text
OPEN
ACKNOWLEDGED
IN_PROGRESS
RESOLVED
CLOSED
```

### Deployment

```text
Windows installer
Linux installer
```

---

# 46. Important Decisions Still To Be Finalized

These areas require architecture discussion before implementation:

1. Exact user identity mechanism
2. Exact agent credential provisioning
3. Final database schema
4. Final name for physical location/table
5. Exact role and permission model
6. Schedule ownership/scope
7. Session timeout rules
8. Health thresholds
9. Event retention
10. Uptime calculation rules
11. Agent update mechanism
12. Large-scale deployment strategy

These decisions should be made collaboratively before they become difficult to change.

---

# 47. Project Vision

CampusWatch aims to become a reliable institutional monitoring platform that gives administrators a clear picture of their computing environment.

The long-term vision is:

```text
Install
   ↓
Register
   ↓
Approve
   ↓
Monitor
   ↓
Understand
   ↓
Detect Problems
   ↓
Respond
   ↓
Report
```

Instead of administrators asking:

> "What is happening with the computers?"

CampusWatch should provide the answer.

---

# 48. Project Principle

> **Know the system. Know the activity. Know the problem. Take action.**

CampusWatch is designed to provide useful operational visibility while respecting privacy, security, and institutional boundaries.
