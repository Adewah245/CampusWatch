import express, { Request, Response, NextFunction } from 'express';
import path from 'path';
import { fileURLToPath } from 'url';
import crypto from 'crypto';
import bcrypt from 'bcryptjs';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// ============================================================================
// Types & Models
// ============================================================================

export interface Institution {
  id: string;
  name: string;
  slug: string;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface Campus {
  id: string;
  institution_id: string;
  name: string;
  slug: string;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface Cluster {
  id: string;
  campus_id: string;
  name: string;
  slug: string;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface Location {
  id: string;
  cluster_id: string;
  name: string;
  slug: string;
  location_type: string;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface System {
  id: string;
  location_id: string;
  hostname: string;
  device_name: string;
  operating_system: string;
  os_version: string;
  system_status: string;
  monitor_status: string;
  last_seen_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface Agent {
  id: string;
  system_id: string;
  agent_code: string;
  agent_version: string;
  status: string;
  credential_hash?: string;
  last_heartbeat_at?: string | null;
  created_at: string;
  updated_at: string;
}

export interface HealthReport {
  system_id: string;
  recorded_at: string;
  cpu_percent: number;
  memory_percent: number;
  disk_percent: number;
  battery_percent: number | null;
  network_connected: boolean;
  operating_system: string;
  agent_health: string;
  uptime_seconds: number;
  created_at?: string;
}

export interface UserSession {
  id: string;
  system_id: string;
  username: string;
  state: 'logged_in_active' | 'logged_in_idle' | 'logged_out';
  login_at: string;
  logout_at: string | null;
  last_activity_at: string;
  duration_seconds: number | null;
  created_at: string;
}

export interface Event {
  id: string;
  system_id: string;
  agent_id: string | null;
  event_type: string;
  payload: Record<string, unknown>;
  occurred_at: string;
  created_at: string;
}

export interface Issue {
  id: string;
  system_id: string;
  title: string;
  description: string;
  priority: 'low' | 'medium' | 'high' | 'critical';
  status: 'open' | 'acknowledged' | 'in_progress' | 'resolved' | 'closed';
  reported_by?: string;
  created_at: string;
  updated_at: string;
}

export interface IssueUpdate {
  id: string;
  issue_id: string;
  status?: string;
  comment: string;
  updated_by?: string;
  created_at: string;
}

export interface Schedule {
  id: string;
  institution_id: string;
  day_of_week: number;
  opening_time: string;
  break_start?: string;
  break_end?: string;
  closing_time: string;
  created_at: string;
  updated_at: string;
}

export interface User {
  id: string;
  institution_id: string;
  email: string;
  password_hash: string;
  first_name: string;
  last_name: string;
  role: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface Session {
  token: string;
  user_id: string;
  institution_id: string;
  role: string;
  expires_at: string;
  created_at: string;
}

export interface AuditLog {
  id: string;
  user_id?: string;
  institution_id?: string;
  method: string;
  path: string;
  created_at: string;
}

// ============================================================================
// In-Memory Database Store & Seed Data
// ============================================================================

class InMemoryStore {
  institutions: Map<string, Institution> = new Map();
  campuses: Map<string, Campus> = new Map();
  clusters: Map<string, Cluster> = new Map();
  locations: Map<string, Location> = new Map();
  systems: Map<string, System> = new Map();
  agents: Map<string, Agent> = new Map();
  healthSnapshots: Map<string, HealthReport> = new Map(); // system_id -> latest
  userSessions: Map<string, UserSession> = new Map();
  events: Event[] = [];
  issues: Map<string, Issue> = new Map();
  issueUpdates: IssueUpdate[] = [];
  schedules: Map<string, Schedule> = new Map();
  users: Map<string, User> = new Map();
  sessions: Map<string, Session> = new Map();
  auditLogs: AuditLog[] = [];

  constructor() {
    this.seed();
  }

  seed() {
    const now = new Date();
    const isoNow = now.toISOString();

    // 1. Institution
    const instId = 'inst-adewah-01';
    this.institutions.set(instId, {
      id: instId,
      name: 'Adewah University',
      slug: 'adewah-university',
      status: 'active',
      created_at: new Date(now.getTime() - 90 * 86400000).toISOString(),
      updated_at: isoNow,
    });

    // 2. Admin User
    const adminId = 'usr-admin-01';
    const passwordHash = bcrypt.hashSync('AdminPassword123!', 10);
    this.users.set(adminId, {
      id: adminId,
      institution_id: instId,
      email: 'admin@campuswatch.edu',
      password_hash: passwordHash,
      first_name: 'CampusWatch',
      last_name: 'Administrator',
      role: 'admin',
      is_active: true,
      created_at: new Date(now.getTime() - 90 * 86400000).toISOString(),
      updated_at: isoNow,
    });

    // 3. Campuses
    const campus1 = 'cmp-main-01';
    const campus2 = 'cmp-tech-02';
    this.campuses.set(campus1, {
      id: campus1,
      institution_id: instId,
      name: 'Main Campus',
      slug: 'main-campus',
      status: 'active',
      created_at: isoNow,
      updated_at: isoNow,
    });
    this.campuses.set(campus2, {
      id: campus2,
      institution_id: instId,
      name: 'Science & Innovation Park',
      slug: 'science-park',
      status: 'active',
      created_at: isoNow,
      updated_at: isoNow,
    });

    // 4. Clusters
    const cluster1 = 'cls-cs-01';
    const cluster2 = 'cls-lib-02';
    const cluster3 = 'cls-eng-03';
    this.clusters.set(cluster1, {
      id: cluster1,
      campus_id: campus1,
      name: 'Computer Science Building',
      slug: 'cs-building',
      status: 'active',
      created_at: isoNow,
      updated_at: isoNow,
    });
    this.clusters.set(cluster2, {
      id: cluster2,
      campus_id: campus1,
      name: 'University Library',
      slug: 'library',
      status: 'active',
      created_at: isoNow,
      updated_at: isoNow,
    });
    this.clusters.set(cluster3, {
      id: cluster3,
      campus_id: campus2,
      name: 'Engineering & Robotics Wing',
      slug: 'engineering-wing',
      status: 'active',
      created_at: isoNow,
      updated_at: isoNow,
    });

    // 5. Locations
    const loc1 = 'loc-lab101';
    const loc2 = 'loc-lab102';
    const loc3 = 'loc-libcommons';
    const loc4 = 'loc-robotics';
    this.locations.set(loc1, {
      id: loc1,
      cluster_id: cluster1,
      name: 'CS Lab 101 (Systems)',
      slug: 'cs-lab-101',
      location_type: 'laboratory',
      status: 'active',
      created_at: isoNow,
      updated_at: isoNow,
    });
    this.locations.set(loc2, {
      id: loc2,
      cluster_id: cluster1,
      name: 'CS Lab 102 (AI / Data)',
      slug: 'cs-lab-102',
      location_type: 'laboratory',
      status: 'active',
      created_at: isoNow,
      updated_at: isoNow,
    });
    this.locations.set(loc3, {
      id: loc3,
      cluster_id: cluster2,
      name: 'Library Commons Level 2',
      slug: 'library-commons',
      location_type: 'study_area',
      status: 'active',
      created_at: isoNow,
      updated_at: isoNow,
    });
    this.locations.set(loc4, {
      id: loc4,
      cluster_id: cluster3,
      name: 'Robotics Workshop B',
      slug: 'robotics-workshop',
      location_type: 'workshop',
      status: 'active',
      created_at: isoNow,
      updated_at: isoNow,
    });

    // 6. Monitored Systems
    const initialSystems: Array<Omit<System, 'created_at' | 'updated_at'> & { health?: Partial<HealthReport>; activeUser?: string }> = [
      {
        id: 'sys-dell-01',
        location_id: loc1,
        hostname: 'cw-cs101-01',
        device_name: 'Dell OptiPlex 7090 Micro',
        operating_system: 'Ubuntu',
        os_version: '24.04 LTS',
        system_status: 'operational',
        monitor_status: 'ONLINE',
        last_seen_at: new Date(now.getTime() - 25000).toISOString(),
        health: { cpu_percent: 24.5, memory_percent: 54.2, disk_percent: 41.0, agent_health: 'healthy', uptime_seconds: 345600 },
        activeUser: 'a.turing',
      },
      {
        id: 'sys-dell-02',
        location_id: loc1,
        hostname: 'cw-cs101-02',
        device_name: 'Dell OptiPlex 7090 Micro',
        operating_system: 'Ubuntu',
        os_version: '24.04 LTS',
        system_status: 'operational',
        monitor_status: 'ONLINE',
        last_seen_at: new Date(now.getTime() - 15000).toISOString(),
        health: { cpu_percent: 42.1, memory_percent: 68.7, disk_percent: 44.5, agent_health: 'healthy', uptime_seconds: 345600 },
        activeUser: 'm.curie',
      },
      {
        id: 'sys-dell-03',
        location_id: loc1,
        hostname: 'cw-cs101-03',
        device_name: 'Dell OptiPlex 7090 Micro',
        operating_system: 'Ubuntu',
        os_version: '24.04 LTS',
        system_status: 'operational',
        monitor_status: 'INACTIVE',
        last_seen_at: new Date(now.getTime() - 40000).toISOString(),
        health: { cpu_percent: 4.2, memory_percent: 28.1, disk_percent: 40.2, agent_health: 'healthy', uptime_seconds: 180000 },
      },
      {
        id: 'sys-hp-01',
        location_id: loc2,
        hostname: 'cw-cs102-01',
        device_name: 'HP EliteDesk 800 G6',
        operating_system: 'Windows 11',
        os_version: 'Pro 23H2',
        system_status: 'operational',
        monitor_status: 'ONLINE',
        last_seen_at: new Date(now.getTime() - 10000).toISOString(),
        health: { cpu_percent: 31.8, memory_percent: 62.4, disk_percent: 58.0, agent_health: 'healthy', uptime_seconds: 520000 },
        activeUser: 'j.vonneumann',
      },
      {
        id: 'sys-hp-02',
        location_id: loc2,
        hostname: 'cw-cs102-02',
        device_name: 'HP EliteDesk 800 G6',
        operating_system: 'Windows 11',
        os_version: 'Pro 23H2',
        system_status: 'operational',
        monitor_status: 'OFFLINE',
        last_seen_at: new Date(now.getTime() - 3600000 * 5).toISOString(),
        health: { cpu_percent: 0, memory_percent: 0, disk_percent: 58.0, agent_health: 'unhealthy', uptime_seconds: 0 },
      },
      {
        id: 'sys-mac-01',
        location_id: loc3,
        hostname: 'cw-lib-01',
        device_name: 'Apple iMac 24" M3',
        operating_system: 'macOS',
        os_version: 'Sonoma 14.5',
        system_status: 'operational',
        monitor_status: 'ONLINE',
        last_seen_at: new Date(now.getTime() - 30000).toISOString(),
        health: { cpu_percent: 18.2, memory_percent: 48.0, disk_percent: 32.1, agent_health: 'healthy', uptime_seconds: 680000 },
        activeUser: 'e.lovelace',
      },
      {
        id: 'sys-mac-02',
        location_id: loc3,
        hostname: 'cw-lib-02',
        device_name: 'Apple iMac 24" M3',
        operating_system: 'macOS',
        os_version: 'Sonoma 14.5',
        system_status: 'operational',
        monitor_status: 'ONLINE',
        last_seen_at: new Date(now.getTime() - 20000).toISOString(),
        health: { cpu_percent: 14.5, memory_percent: 45.2, disk_percent: 31.9, agent_health: 'healthy', uptime_seconds: 680000 },
      },
      {
        id: 'sys-mac-03',
        location_id: loc3,
        hostname: 'cw-lib-03',
        device_name: 'Apple iMac 24" M3',
        operating_system: 'macOS',
        os_version: 'Sonoma 14.5',
        system_status: 'pending_setup',
        monitor_status: 'PENDING',
        last_seen_at: null,
      },
      {
        id: 'sys-lenovo-01',
        location_id: loc4,
        hostname: 'cw-robotics-01',
        device_name: 'Lenovo ThinkStation P360',
        operating_system: 'Ubuntu',
        os_version: '22.04 LTS',
        system_status: 'maintenance',
        monitor_status: 'MAINTENANCE',
        last_seen_at: new Date(now.getTime() - 120000).toISOString(),
        health: { cpu_percent: 12.0, memory_percent: 35.0, disk_percent: 60.0, agent_health: 'warning', uptime_seconds: 90000 },
      },
    ];

    for (const sys of initialSystems) {
      const systemId = sys.id;
      const createdAt = new Date(now.getTime() - 30 * 86400000).toISOString();
      const sRecord: System = {
        id: systemId,
        location_id: sys.location_id,
        hostname: sys.hostname,
        device_name: sys.device_name,
        operating_system: sys.operating_system,
        os_version: sys.os_version,
        system_status: sys.system_status,
        monitor_status: sys.monitor_status,
        last_seen_at: sys.last_seen_at,
        created_at: createdAt,
        updated_at: isoNow,
      };
      this.systems.set(systemId, sRecord);

      // Create Agent
      const agentId = `agt-${systemId}`;
      const agent: Agent = {
        id: agentId,
        system_id: systemId,
        agent_code: `CW-AGENT-${sys.hostname.toUpperCase()}`,
        agent_version: '1.4.2',
        status: sys.monitor_status === 'PENDING' ? 'pending' : 'approved',
        last_heartbeat_at: sys.last_seen_at,
        created_at: createdAt,
        updated_at: isoNow,
      };
      this.agents.set(agentId, agent);

      // Health Snapshot
      if (sys.health) {
        this.healthSnapshots.set(systemId, {
          system_id: systemId,
          recorded_at: sys.last_seen_at || isoNow,
          cpu_percent: sys.health.cpu_percent ?? 20.0,
          memory_percent: sys.health.memory_percent ?? 50.0,
          disk_percent: sys.health.disk_percent ?? 45.0,
          battery_percent: null,
          network_connected: sys.monitor_status !== 'OFFLINE',
          operating_system: `${sys.operating_system} ${sys.os_version}`,
          agent_health: sys.health.agent_health ?? 'healthy',
          uptime_seconds: sys.health.uptime_seconds ?? 86400,
        });
      }

      // User Session
      if (sys.activeUser) {
        const sessId = `sess-${systemId}`;
        this.userSessions.set(sessId, {
          id: sessId,
          system_id: systemId,
          username: sys.activeUser,
          state: 'logged_in_active',
          login_at: new Date(now.getTime() - 45 * 60000).toISOString(),
          logout_at: null,
          last_activity_at: new Date(now.getTime() - 2 * 60000).toISOString(),
          duration_seconds: 45 * 60,
          created_at: new Date(now.getTime() - 45 * 60000).toISOString(),
        });
      }

      // System Events
      this.events.push({
        id: `evt-${systemId}-boot`,
        system_id: systemId,
        agent_id: agentId,
        event_type: 'SYSTEM_ONLINE',
        payload: { hostname: sys.hostname, ip: '10.0.12.' + (10 + Math.floor(Math.random() * 80)) },
        occurred_at: new Date(now.getTime() - 6 * 3600000).toISOString(),
        created_at: isoNow,
      });
    }

    // After hours event
    this.events.push({
      id: 'evt-afterhours-1',
      system_id: 'sys-dell-02',
      agent_id: 'agt-sys-dell-02',
      event_type: 'AFTER_HOURS_USAGE',
      payload: { user: 'm.curie', reason: 'Late lab research' },
      occurred_at: new Date(now.getTime() - 14 * 3600000).toISOString(),
      created_at: isoNow,
    });

    // 7. Issues
    const issue1: Issue = {
      id: 'iss-001',
      system_id: 'sys-hp-02',
      title: 'Workstation unresponsive and offline for >5 hours',
      description: 'System cw-cs102-02 stopped reporting heartbeats. Please check physical power cable and Ethernet port in CS Lab 102.',
      priority: 'high',
      status: 'open',
      reported_by: 'Automated Health Monitor',
      created_at: new Date(now.getTime() - 5 * 3600000).toISOString(),
      updated_at: new Date(now.getTime() - 5 * 3600000).toISOString(),
    };
    const issue2: Issue = {
      id: 'iss-002',
      system_id: 'sys-dell-03',
      title: 'Mechanical keyboard spacebar sticking',
      description: 'Students reported spacebar sticking repeatedly during C++ lab exercises.',
      priority: 'medium',
      status: 'in_progress',
      reported_by: 'Prof. Adeyemi',
      created_at: new Date(now.getTime() - 24 * 3600000).toISOString(),
      updated_at: new Date(now.getTime() - 2 * 3600000).toISOString(),
    };
    const issue3: Issue = {
      id: 'iss-003',
      system_id: 'sys-lenovo-01',
      title: 'Scheduled GPU driver update & maintenance',
      description: 'NVIDIA CUDA 12.4 update and sensor calibration for robotics vision system.',
      priority: 'medium',
      status: 'acknowledged',
      reported_by: 'Robotics Lab Technician',
      created_at: new Date(now.getTime() - 48 * 3600000).toISOString(),
      updated_at: new Date(now.getTime() - 12 * 3600000).toISOString(),
    };

    this.issues.set(issue1.id, issue1);
    this.issues.set(issue2.id, issue2);
    this.issues.set(issue3.id, issue3);

    this.issueUpdates.push({
      id: 'upd-iss-002-1',
      issue_id: 'iss-002',
      status: 'in_progress',
      comment: 'Technician dispatched with replacement key switches.',
      updated_by: 'CampusWatch Administrator',
      created_at: new Date(now.getTime() - 2 * 3600000).toISOString(),
    });

    // 8. Schedules (Mon-Fri 08:00 - 20:00, Sat 09:00 - 15:00)
    for (let day = 1; day <= 5; day++) {
      const schId = `sch-${instId}-day-${day}`;
      this.schedules.set(schId, {
        id: schId,
        institution_id: instId,
        day_of_week: day,
        opening_time: '08:00:00',
        break_start: '12:00:00',
        break_end: '13:00:00',
        closing_time: '20:00:00',
        created_at: isoNow,
        updated_at: isoNow,
      });
    }
    const satId = `sch-${instId}-day-6`;
    this.schedules.set(satId, {
      id: satId,
      institution_id: instId,
      day_of_week: 6,
      opening_time: '09:00:00',
      closing_time: '15:00:00',
      created_at: isoNow,
      updated_at: isoNow,
    });
  }

  // Periodic heartbeat tick to simulate dynamic activity
  simulateHeartbeat() {
    const now = new Date();
    for (const [id, sys] of this.systems.entries()) {
      if (sys.monitor_status === 'ONLINE') {
        sys.last_seen_at = now.toISOString();
        sys.updated_at = now.toISOString();

        const health = this.healthSnapshots.get(id);
        if (health) {
          health.recorded_at = now.toISOString();
          // slight realistic CPU and memory fluctuation
          health.cpu_percent = Math.min(99, Math.max(5, health.cpu_percent + (Math.random() * 6 - 3)));
          health.memory_percent = Math.min(95, Math.max(25, health.memory_percent + (Math.random() * 2 - 1)));
          health.uptime_seconds += 15;
        }
      }
    }
  }
}

const store = new InMemoryStore();

// Start periodic heartbeat update every 15 seconds
setInterval(() => {
  store.simulateHeartbeat();
}, 15000);

// ============================================================================
// Helpers
// ============================================================================

function slugify(name: string): string {
  return name
    .toLowerCase()
    .trim()
    .replace(/[^\w\s-]/g, '')
    .replace(/[\s_-]+/g, '-')
    .replace(/^-+|-+$/g, '');
}

function generateToken(): string {
  return crypto.randomBytes(32).toString('hex');
}

function parseBearerToken(req: Request): string | null {
  const authHeader = req.headers.authorization;
  if (!authHeader) return null;
  const match = authHeader.match(/^Bearer\s+(.+)$/i);
  return match ? match[1].trim() : null;
}

// Auth Middleware: requires valid token and optionally role
function requireAuth(roles?: string[]) {
  return (req: Request, res: Response, next: NextFunction) => {
    const token = parseBearerToken(req);
    if (!token) {
      return res.status(401).json({ error: 'authentication required' });
    }

    const session = store.sessions.get(token);
    if (!session) {
      return res.status(401).json({ error: 'invalid or expired session' });
    }

    if (new Date(session.expires_at).getTime() <= Date.now()) {
      store.sessions.delete(token);
      return res.status(401).json({ error: 'session expired' });
    }

    const user = store.users.get(session.user_id);
    if (!user || !user.is_active) {
      return res.status(401).json({ error: 'user account inactive' });
    }

    if (roles && roles.length > 0) {
      const userRole = (user.role || '').toLowerCase();
      const hasRole = roles.some((r) => r.toLowerCase() === userRole);
      if (!hasRole) {
        return res.status(403).json({
          error: `access denied: role '${user.role}' lacks required permissions`,
        });
      }
    }

    // Attach to request
    (req as any).user = user;
    (req as any).session = session;

    // Log audit record
    store.auditLogs.push({
      id: 'aud-' + Date.now() + '-' + Math.random().toString(36).substring(2, 6),
      user_id: user.id,
      institution_id: user.institution_id,
      method: req.method,
      path: req.path,
      created_at: new Date().toISOString(),
    });

    next();
  };
}

// Guard for admin or manager routes (which powers the CampusWatch dashboard)
const adminOnly = requireAuth(['admin', 'manager']);

// ============================================================================
// API Router Setup
// ============================================================================

export function registerApiRoutes(app: express.Express) {
  const router = express.Router();

  // --- Health Endpoint -----------------------------------------------------
  app.get('/health', (req, res) => {
    res.json({
      status: 'ok',
      timestamp: new Date().toISOString(),
      fleet_count: store.systems.size,
    });
  });

  // --- Auth Endpoints ------------------------------------------------------
  router.post('/auth/login', (req, res) => {
    const { email, password } = req.body || {};
    if (!email || !password) {
      return res.status(400).json({ error: 'email and password are required' });
    }

    const cleanEmail = String(email).trim().toLowerCase();
    let foundUser: User | null = null;
    for (const u of store.users.values()) {
      if (u.email.toLowerCase() === cleanEmail) {
        foundUser = u;
        break;
      }
    }

    if (!foundUser || !foundUser.is_active) {
      return res.status(401).json({ error: 'invalid credentials' });
    }

    const validPass = bcrypt.compareSync(String(password), foundUser.password_hash);
    if (!validPass) {
      return res.status(401).json({ error: 'invalid credentials' });
    }

    const token = generateToken();
    const expiresAt = new Date(Date.now() + 24 * 3600000).toISOString();
    const sessionRecord: Session = {
      token,
      user_id: foundUser.id,
      institution_id: foundUser.institution_id,
      role: foundUser.role,
      expires_at: expiresAt,
      created_at: new Date().toISOString(),
    };

    store.sessions.set(token, sessionRecord);

    res.json({
      token,
      user_id: foundUser.id,
      institution_id: foundUser.institution_id,
      role: foundUser.role,
      expires_at: expiresAt,
    });
  });

  router.post('/auth/register', (req, res) => {
    const { institution, first_name, last_name, email, password } = req.body || {};
    if (!institution || !first_name || !last_name || !email || !password) {
      return res.status(400).json({ error: 'all registration fields are required' });
    }

    if (String(password).length < 12) {
      return res.status(400).json({ error: 'password must be at least 12 characters' });
    }

    const cleanEmail = String(email).trim().toLowerCase();
    for (const u of store.users.values()) {
      if (u.email.toLowerCase() === cleanEmail) {
        return res.status(409).json({ error: 'registration is closed: this account already exists' });
      }
    }

    const now = new Date().toISOString();
    // Create Institution if needed
    let instId = 'inst-' + Date.now();
    let instSlug = slugify(institution);
    for (const inst of store.institutions.values()) {
      if (inst.name.toLowerCase() === institution.trim().toLowerCase()) {
        instId = inst.id;
        instSlug = inst.slug;
        break;
      }
    }
    if (!store.institutions.has(instId)) {
      store.institutions.set(instId, {
        id: instId,
        name: institution.trim(),
        slug: instSlug,
        status: 'active',
        created_at: now,
        updated_at: now,
      });
    }

    // Create User
    const userId = 'usr-' + Date.now();
    const hash = bcrypt.hashSync(String(password), 10);
    const newUser: User = {
      id: userId,
      institution_id: instId,
      email: cleanEmail,
      password_hash: hash,
      first_name: first_name.trim(),
      last_name: last_name.trim(),
      role: 'admin',
      is_active: true,
      created_at: now,
      updated_at: now,
    };
    store.users.set(userId, newUser);

    res.status(201).json({
      user_id: userId,
      institution_id: instId,
      role: 'admin',
    });
  });

  router.get('/auth/me', requireAuth(), (req, res) => {
    res.json((req as any).session);
  });

  router.post('/auth/logout', requireAuth(), (req, res) => {
    const token = (req as any).session?.token;
    if (token) {
      store.sessions.delete(token);
    }
    res.status(204).send();
  });

  // --- Dashboard Endpoints -------------------------------------------------
  router.get('/dashboard/summary', adminOnly, (req, res) => {
    const systems = Array.from(store.systems.values());
    const totalSystems = systems.length;

    let onlineSystems = 0;
    let offlineSystems = 0;
    let inactiveSystems = 0;
    let maintenanceSystems = 0;

    for (const s of systems) {
      const status = (s.monitor_status || '').toUpperCase();
      if (status === 'ONLINE') onlineSystems++;
      else if (status === 'OFFLINE') offlineSystems++;
      else if (status === 'INACTIVE') inactiveSystems++;
      else if (status === 'MAINTENANCE') maintenanceSystems++;
    }

    let activeUsers = 0;
    for (const us of store.userSessions.values()) {
      if (us.state === 'logged_in_active') activeUsers++;
    }

    let openIssues = 0;
    for (const iss of store.issues.values()) {
      if (['open', 'acknowledged', 'in_progress'].includes(iss.status.toLowerCase())) {
        openIssues++;
      }
    }

    let afterHoursUsage = 0;
    for (const evt of store.events) {
      if (evt.event_type === 'AFTER_HOURS_USAGE') {
        afterHoursUsage++;
      }
    }

    res.json({
      total_systems: totalSystems,
      online_systems: onlineSystems,
      offline_systems: offlineSystems,
      inactive_systems: inactiveSystems,
      maintenance_systems: maintenanceSystems,
      active_users: activeUsers,
      open_issues: openIssues,
      after_hours_usage: afterHoursUsage,
    });
  });

  router.get('/dashboard/systems', adminOnly, (req, res) => {
    const systems = Array.from(store.systems.values());
    systems.sort((a, b) => a.hostname.localeCompare(b.hostname));
    res.json(systems);
  });

  router.get('/dashboard/issues', adminOnly, (req, res) => {
    const openIssues = Array.from(store.issues.values()).filter((i) =>
      ['open', 'acknowledged', 'in_progress'].includes(i.status.toLowerCase()),
    );
    res.json(openIssues);
  });

  // --- Systems Endpoints ---------------------------------------------------
  router.get('/systems', adminOnly, (req, res) => {
    let result = Array.from(store.systems.values());
    const { status, location_id } = req.query;
    if (status && typeof status === 'string') {
      result = result.filter((s) => s.monitor_status.toUpperCase() === status.toUpperCase());
    }
    if (location_id && typeof location_id === 'string') {
      result = result.filter((s) => s.location_id === location_id);
    }
    result.sort((a, b) => a.hostname.localeCompare(b.hostname));
    res.json(result);
  });

  router.post('/systems', adminOnly, (req, res) => {
    const { location_id, hostname, device_name, operating_system, os_version } = req.body || {};
    if (!location_id || !hostname) {
      return res.status(400).json({ error: 'location_id and hostname are required' });
    }

    const now = new Date().toISOString();
    const id = 'sys-' + Date.now();
    const system: System = {
      id,
      location_id,
      hostname: String(hostname).trim(),
      device_name: device_name ? String(device_name).trim() : 'Generic Desktop',
      operating_system: operating_system ? String(operating_system).trim() : 'Linux',
      os_version: os_version ? String(os_version).trim() : '',
      system_status: 'pending_setup',
      monitor_status: 'PENDING',
      last_seen_at: null,
      created_at: now,
      updated_at: now,
    };
    store.systems.set(id, system);

    // Create an agent record for it
    const agentId = `agt-${id}`;
    store.agents.set(agentId, {
      id: agentId,
      system_id: id,
      agent_code: `CW-AGENT-${system.hostname.toUpperCase()}`,
      agent_version: '1.4.2',
      status: 'pending',
      created_at: now,
      updated_at: now,
    });

    res.status(201).json(system);
  });

  router.get('/systems/:id', adminOnly, (req, res) => {
    const system = store.systems.get(req.params.id);
    if (!system) {
      return res.status(404).json({ error: 'system not found' });
    }
    res.json(system);
  });

  router.patch('/systems/:id', adminOnly, (req, res) => {
    const system = store.systems.get(req.params.id);
    if (!system) {
      return res.status(404).json({ error: 'system not found' });
    }

    const {
      location_id,
      hostname,
      device_name,
      operating_system,
      os_version,
      system_status,
      monitor_status,
    } = req.body || {};

    if (location_id !== undefined) system.location_id = location_id;
    if (hostname !== undefined) system.hostname = hostname;
    if (device_name !== undefined) system.device_name = device_name;
    if (operating_system !== undefined) system.operating_system = operating_system;
    if (os_version !== undefined) system.os_version = os_version;
    if (system_status !== undefined) system.system_status = system_status;
    if (monitor_status !== undefined) {
      system.monitor_status = String(monitor_status).toUpperCase();
      if (system.monitor_status === 'ONLINE' && !system.last_seen_at) {
        system.last_seen_at = new Date().toISOString();
      }
    }
    system.updated_at = new Date().toISOString();

    res.json(system);
  });

  router.post('/systems/:id/approve', adminOnly, (req, res) => {
    const system = store.systems.get(req.params.id);
    if (!system) {
      return res.status(404).json({ error: 'system not found' });
    }

    const now = new Date().toISOString();
    system.monitor_status = 'ONLINE';
    system.system_status = 'operational';
    system.last_seen_at = now;
    system.updated_at = now;

    // Approve agent
    for (const ag of store.agents.values()) {
      if (ag.system_id === system.id) {
        ag.status = 'approved';
        ag.last_heartbeat_at = now;
        ag.updated_at = now;
      }
    }

    // Create initial health snapshot if absent
    if (!store.healthSnapshots.has(system.id)) {
      store.healthSnapshots.set(system.id, {
        system_id: system.id,
        recorded_at: now,
        cpu_percent: 18.5,
        memory_percent: 42.0,
        disk_percent: 36.5,
        battery_percent: null,
        network_connected: true,
        operating_system: `${system.operating_system} ${system.os_version}`.trim(),
        agent_health: 'healthy',
        uptime_seconds: 3600,
      });
    }

    res.json(system);
  });

  router.get('/systems/:id/health', adminOnly, (req, res) => {
    const health = store.healthSnapshots.get(req.params.id);
    if (!health) {
      return res.status(404).json({ error: 'system health not found' });
    }
    res.json(health);
  });

  router.get('/systems/:id/sessions', adminOnly, (req, res) => {
    const systemId = req.params.id;
    const sessions = Array.from(store.userSessions.values()).filter(
      (s) => s.system_id === systemId,
    );
    sessions.sort((a, b) => new Date(b.login_at).getTime() - new Date(a.login_at).getTime());
    res.json(sessions);
  });

  router.get('/systems/:id/events', adminOnly, (req, res) => {
    const systemId = req.params.id;
    const events = store.events.filter((e) => e.system_id === systemId);
    events.sort((a, b) => new Date(b.occurred_at).getTime() - new Date(a.occurred_at).getTime());
    res.json(events);
  });

  router.get('/systems/:id/agents', adminOnly, (req, res) => {
    const systemId = req.params.id;
    const agents = Array.from(store.agents.values()).filter((a) => a.system_id === systemId);
    res.json(agents);
  });

  // --- Agents Endpoints ----------------------------------------------------
  router.get('/agents', adminOnly, (req, res) => {
    res.json(Array.from(store.agents.values()));
  });

  router.post('/agents', adminOnly, (req, res) => {
    const { system_id, agent_code, agent_version } = req.body || {};
    const id = 'agt-' + Date.now();
    const credential = crypto.randomBytes(24).toString('hex');
    const agent: Agent = {
      id,
      system_id,
      agent_code: agent_code || `AGENT-${id}`,
      agent_version: agent_version || '1.4.2',
      status: 'pending',
      credential_hash: bcrypt.hashSync(credential, 10),
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };
    store.agents.set(id, agent);
    res.status(201).json({ agent, credential });
  });

  router.post('/agents/:id/approve', adminOnly, (req, res) => {
    const agent = store.agents.get(req.params.id);
    if (!agent) {
      return res.status(404).json({ error: 'agent not found' });
    }
    const credential = crypto.randomBytes(24).toString('hex');
    agent.status = 'approved';
    agent.credential_hash = bcrypt.hashSync(credential, 10);
    agent.updated_at = new Date().toISOString();
    res.json({ agent, credential });
  });

  router.post('/agents/heartbeat', (req, res) => {
    const { agent_id, system_id, health } = req.body || {};
    const now = new Date().toISOString();

    if (system_id && store.systems.has(system_id)) {
      const sys = store.systems.get(system_id)!;
      sys.last_seen_at = now;
      sys.monitor_status = 'ONLINE';
      sys.updated_at = now;

      if (health) {
        store.healthSnapshots.set(system_id, {
          system_id,
          recorded_at: now,
          cpu_percent: health.cpu_percent ?? 15,
          memory_percent: health.memory_percent ?? 45,
          disk_percent: health.disk_percent ?? 50,
          battery_percent: health.battery_percent ?? null,
          network_connected: true,
          operating_system: health.operating_system || `${sys.operating_system} ${sys.os_version}`,
          agent_health: health.agent_health || 'healthy',
          uptime_seconds: health.uptime_seconds ?? 3600,
        });
      }
    }

    if (agent_id && store.agents.has(agent_id)) {
      const ag = store.agents.get(agent_id)!;
      ag.last_heartbeat_at = now;
    }

    res.json({
      agent_id: agent_id || 'unknown',
      system_id: system_id || 'unknown',
      status: 'ok',
      received_at: now,
    });
  });

  router.post('/agents/session', (req, res) => {
    const { system_id, username, event, timestamp } = req.body || {};
    const now = timestamp || new Date().toISOString();

    if (system_id && username) {
      const sessId = `sess-${Date.now()}-${Math.random().toString(36).substring(2, 6)}`;
      store.userSessions.set(sessId, {
        id: sessId,
        system_id,
        username,
        state: event === 'logout' ? 'logged_out' : 'logged_in_active',
        login_at: now,
        logout_at: event === 'logout' ? now : null,
        last_activity_at: now,
        duration_seconds: event === 'logout' ? 1800 : null,
        created_at: now,
      });
    }

    res.json({ status: 'recorded' });
  });

  router.post('/agents/event', (req, res) => {
    const { system_id, agent_id, event_type, payload, occurred_at } = req.body || {};
    if (system_id && event_type) {
      store.events.push({
        id: 'evt-' + Date.now(),
        system_id,
        agent_id: agent_id || null,
        event_type,
        payload: payload || {},
        occurred_at: occurred_at || new Date().toISOString(),
        created_at: new Date().toISOString(),
      });
    }
    res.json({ status: 'recorded' });
  });

  // --- Issues Endpoints ----------------------------------------------------
  router.get('/issues', adminOnly, (req, res) => {
    let result = Array.from(store.issues.values());
    const { status, system_id } = req.query;
    if (status && typeof status === 'string') {
      result = result.filter((i) => i.status.toLowerCase() === status.toLowerCase());
    }
    if (system_id && typeof system_id === 'string') {
      result = result.filter((i) => i.system_id === system_id);
    }
    result.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());
    res.json(result);
  });

  router.post('/issues', adminOnly, (req, res) => {
    const { system_id, title, description, priority, reported_by } = req.body || {};
    if (!system_id || !title || !description) {
      return res.status(400).json({ error: 'system_id, title, and description are required' });
    }

    const now = new Date().toISOString();
    const id = 'iss-' + Date.now();
    const issue: Issue = {
      id,
      system_id,
      title: String(title).trim(),
      description: String(description).trim(),
      priority: priority ? String(priority).toLowerCase() as any : 'medium',
      status: 'open',
      reported_by: reported_by ? String(reported_by).trim() : ((req as any).user?.email || 'Administrator'),
      created_at: now,
      updated_at: now,
    };
    store.issues.set(id, issue);

    res.status(201).json(issue);
  });

  router.get('/issues/:id', adminOnly, (req, res) => {
    const issue = store.issues.get(req.params.id);
    if (!issue) {
      return res.status(404).json({ error: 'issue not found' });
    }
    res.json(issue);
  });

  router.patch('/issues/:id', adminOnly, (req, res) => {
    const issue = store.issues.get(req.params.id);
    if (!issue) {
      return res.status(404).json({ error: 'issue not found' });
    }

    const { title, description, priority, status } = req.body || {};
    if (title !== undefined) issue.title = String(title).trim();
    if (description !== undefined) issue.description = String(description).trim();
    if (priority !== undefined) issue.priority = String(priority).toLowerCase() as any;
    if (status !== undefined) issue.status = String(status).toLowerCase() as any;
    issue.updated_at = new Date().toISOString();

    res.json(issue);
  });

  router.post('/issues/:id/updates', adminOnly, (req, res) => {
    const issue = store.issues.get(req.params.id);
    if (!issue) {
      return res.status(404).json({ error: 'issue not found' });
    }

    const { status, comment, updated_by } = req.body || {};
    if (!comment) {
      return res.status(400).json({ error: 'comment is required' });
    }

    const now = new Date().toISOString();
    if (status) {
      issue.status = String(status).toLowerCase() as any;
      issue.updated_at = now;
    }

    const update: IssueUpdate = {
      id: 'upd-' + Date.now(),
      issue_id: issue.id,
      status: status || undefined,
      comment: String(comment).trim(),
      updated_by: updated_by || ((req as any).user?.first_name || 'Operator'),
      created_at: now,
    };
    store.issueUpdates.push(update);

    res.status(201).json(update);
  });

  // --- Hierarchy Endpoints -------------------------------------------------
  // Institutions
  router.get('/institutions', adminOnly, (req, res) => {
    res.json(Array.from(store.institutions.values()));
  });

  router.post('/institutions', adminOnly, (req, res) => {
    const { name, slug, status } = req.body || {};
    if (!name) return res.status(400).json({ error: 'name is required' });
    const id = 'inst-' + Date.now();
    const now = new Date().toISOString();
    const inst: Institution = {
      id,
      name: name.trim(),
      slug: slug ? slugify(slug) : slugify(name),
      status: status || 'active',
      created_at: now,
      updated_at: now,
    };
    store.institutions.set(id, inst);
    res.status(201).json(inst);
  });

  router.get('/institutions/:id', adminOnly, (req, res) => {
    const inst = store.institutions.get(req.params.id);
    if (!inst) return res.status(404).json({ error: 'institution not found' });
    res.json(inst);
  });

  router.patch('/institutions/:id', adminOnly, (req, res) => {
    const inst = store.institutions.get(req.params.id);
    if (!inst) return res.status(404).json({ error: 'institution not found' });
    const { name, slug, status } = req.body || {};
    if (name !== undefined) inst.name = name.trim();
    if (slug !== undefined) inst.slug = slugify(slug);
    if (status !== undefined) inst.status = status;
    inst.updated_at = new Date().toISOString();
    res.json(inst);
  });

  router.delete('/institutions/:id', adminOnly, (req, res) => {
    if (!store.institutions.has(req.params.id)) {
      return res.status(404).json({ error: 'institution not found' });
    }
    store.institutions.delete(req.params.id);
    res.status(204).send();
  });

  // Campuses
  router.get('/campuses', adminOnly, (req, res) => {
    let result = Array.from(store.campuses.values());
    if (req.query.institution_id) {
      result = result.filter((c) => c.institution_id === req.query.institution_id);
    }
    res.json(result);
  });

  router.post('/campuses', adminOnly, (req, res) => {
    const { institution_id, name, slug, status } = req.body || {};
    if (!institution_id || !name) {
      return res.status(400).json({ error: 'institution_id and name are required' });
    }
    const id = 'cmp-' + Date.now();
    const now = new Date().toISOString();
    const campus: Campus = {
      id,
      institution_id,
      name: name.trim(),
      slug: slug ? slugify(slug) : slugify(name),
      status: status || 'active',
      created_at: now,
      updated_at: now,
    };
    store.campuses.set(id, campus);
    res.status(201).json(campus);
  });

  router.get('/campuses/:id', adminOnly, (req, res) => {
    const campus = store.campuses.get(req.params.id);
    if (!campus) return res.status(404).json({ error: 'campus not found' });
    res.json(campus);
  });

  router.patch('/campuses/:id', adminOnly, (req, res) => {
    const campus = store.campuses.get(req.params.id);
    if (!campus) return res.status(404).json({ error: 'campus not found' });
    const { institution_id, name, slug, status } = req.body || {};
    if (institution_id !== undefined) campus.institution_id = institution_id;
    if (name !== undefined) campus.name = name.trim();
    if (slug !== undefined) campus.slug = slugify(slug);
    if (status !== undefined) campus.status = status;
    campus.updated_at = new Date().toISOString();
    res.json(campus);
  });

  router.delete('/campuses/:id', adminOnly, (req, res) => {
    if (!store.campuses.has(req.params.id)) {
      return res.status(404).json({ error: 'campus not found' });
    }
    store.campuses.delete(req.params.id);
    res.status(204).send();
  });

  // Clusters
  router.get('/clusters', adminOnly, (req, res) => {
    let result = Array.from(store.clusters.values());
    if (req.query.campus_id) {
      result = result.filter((c) => c.campus_id === req.query.campus_id);
    }
    res.json(result);
  });

  router.post('/clusters', adminOnly, (req, res) => {
    const { campus_id, name, slug, status } = req.body || {};
    if (!campus_id || !name) {
      return res.status(400).json({ error: 'campus_id and name are required' });
    }
    const id = 'cls-' + Date.now();
    const now = new Date().toISOString();
    const cluster: Cluster = {
      id,
      campus_id,
      name: name.trim(),
      slug: slug ? slugify(slug) : slugify(name),
      status: status || 'active',
      created_at: now,
      updated_at: now,
    };
    store.clusters.set(id, cluster);
    res.status(201).json(cluster);
  });

  router.get('/clusters/:id', adminOnly, (req, res) => {
    const cluster = store.clusters.get(req.params.id);
    if (!cluster) return res.status(404).json({ error: 'cluster not found' });
    res.json(cluster);
  });

  router.patch('/clusters/:id', adminOnly, (req, res) => {
    const cluster = store.clusters.get(req.params.id);
    if (!cluster) return res.status(404).json({ error: 'cluster not found' });
    const { campus_id, name, slug, status } = req.body || {};
    if (campus_id !== undefined) cluster.campus_id = campus_id;
    if (name !== undefined) cluster.name = name.trim();
    if (slug !== undefined) cluster.slug = slugify(slug);
    if (status !== undefined) cluster.status = status;
    cluster.updated_at = new Date().toISOString();
    res.json(cluster);
  });

  router.delete('/clusters/:id', adminOnly, (req, res) => {
    if (!store.clusters.has(req.params.id)) {
      return res.status(404).json({ error: 'cluster not found' });
    }
    store.clusters.delete(req.params.id);
    res.status(204).send();
  });

  // Locations
  router.get('/locations', adminOnly, (req, res) => {
    let result = Array.from(store.locations.values());
    if (req.query.cluster_id) {
      result = result.filter((l) => l.cluster_id === req.query.cluster_id);
    }
    res.json(result);
  });

  router.post('/locations', adminOnly, (req, res) => {
    const { cluster_id, name, slug, location_type, status } = req.body || {};
    if (!cluster_id || !name) {
      return res.status(400).json({ error: 'cluster_id and name are required' });
    }
    const id = 'loc-' + Date.now();
    const now = new Date().toISOString();
    const location: Location = {
      id,
      cluster_id,
      name: name.trim(),
      slug: slug ? slugify(slug) : slugify(name),
      location_type: location_type || 'room',
      status: status || 'active',
      created_at: now,
      updated_at: now,
    };
    store.locations.set(id, location);
    res.status(201).json(location);
  });

  router.get('/locations/:id', adminOnly, (req, res) => {
    const loc = store.locations.get(req.params.id);
    if (!loc) return res.status(404).json({ error: 'location not found' });
    res.json(loc);
  });

  router.patch('/locations/:id', adminOnly, (req, res) => {
    const loc = store.locations.get(req.params.id);
    if (!loc) return res.status(404).json({ error: 'location not found' });
    const { cluster_id, name, slug, location_type, status } = req.body || {};
    if (cluster_id !== undefined) loc.cluster_id = cluster_id;
    if (name !== undefined) loc.name = name.trim();
    if (slug !== undefined) loc.slug = slugify(slug);
    if (location_type !== undefined) loc.location_type = location_type;
    if (status !== undefined) loc.status = status;
    loc.updated_at = new Date().toISOString();
    res.json(loc);
  });

  router.delete('/locations/:id', adminOnly, (req, res) => {
    if (!store.locations.has(req.params.id)) {
      return res.status(404).json({ error: 'location not found' });
    }
    store.locations.delete(req.params.id);
    res.status(204).send();
  });

  // --- Reports Endpoints ---------------------------------------------------
  router.get('/reports/:kind', adminOnly, (req, res) => {
    const kind = req.params.kind.toLowerCase();
    const period = (req.query.period as string) || 'daily';

    const validKinds = ['usage', 'uptime', 'sessions', 'issues', 'after-hours'];
    if (!validKinds.includes(kind)) {
      return res.status(400).json({ error: 'invalid report kind' });
    }

    const now = new Date();
    let days = 1;
    if (period === 'weekly') days = 7;
    else if (period === 'monthly') days = 30;

    const from = new Date(now.getTime() - days * 86400000);

    const report = {
      kind,
      period,
      from: from.toISOString(),
      to: now.toISOString(),
      user_sessions: store.userSessions.size * days,
      usage_seconds: (store.userSessions.size * 3600 + 7200) * days,
      issues_created: store.issues.size,
      after_hours_events: 1,
      uptime_seconds: 86400 * days - 3600,
      downtime_seconds: 3600,
    };

    res.json(report);
  });

  // --- Schedules Endpoints -------------------------------------------------
  router.get('/schedules', adminOnly, (req, res) => {
    let result = Array.from(store.schedules.values());
    if (req.query.institution_id) {
      result = result.filter((s) => s.institution_id === req.query.institution_id);
    }
    res.json(result);
  });

  router.post('/schedules', adminOnly, (req, res) => {
    const { institution_id, day_of_week, opening_time, break_start, break_end, closing_time } =
      req.body || {};
    if (!institution_id || day_of_week === undefined || !opening_time || !closing_time) {
      return res.status(400).json({ error: 'institution_id, day_of_week, opening_time and closing_time are required' });
    }

    const id = 'sch-' + Date.now();
    const now = new Date().toISOString();
    const schedule: Schedule = {
      id,
      institution_id,
      day_of_week: Number(day_of_week),
      opening_time,
      break_start: break_start || undefined,
      break_end: break_end || undefined,
      closing_time,
      created_at: now,
      updated_at: now,
    };
    store.schedules.set(id, schedule);
    res.status(201).json(schedule);
  });

  router.get('/schedules/check', adminOnly, (req, res) => {
    const institutionId = req.query.institution_id as string;
    if (!institutionId) {
      return res.status(400).json({ error: 'institution id is required' });
    }

    const date = req.query.timestamp ? new Date(req.query.timestamp as string) : new Date();
    const day = date.getDay(); // 0 is Sunday, 1 is Monday

    for (const schedule of store.schedules.values()) {
      if (schedule.institution_id === institutionId && schedule.day_of_week === day) {
        return res.json({
          allowed: true,
          state: 'open',
          day_of_week: day,
        });
      }
    }

    res.json({
      allowed: false,
      state: 'closed',
      day_of_week: day,
    });
  });

  // Mount under API prefix
  app.use('/api/v1', router);
}

// ============================================================================
// Server Entrypoint
// ============================================================================

export async function startServer() {
  const app = express();
  const PORT = process.env.PORT ? parseInt(process.env.PORT, 10) : 3000;

  // JSON and URL-encoded body parsing
  app.use(express.json());
  app.use(express.urlencoded({ extended: true }));

  // Security Headers
  app.use((req, res, next) => {
    res.setHeader('X-Content-Type-Options', 'nosniff');
    res.setHeader('X-Frame-Options', 'SAMEORIGIN');
    next();
  });

  // Mount API & health routes first
  registerApiRoutes(app);

  // Serve static assets from frontend
  const frontendDir = path.resolve(__dirname, 'frontend');
  app.use(express.static(frontendDir));

  // Redirect root '/' to '/pages/dashboard.html'
  app.get('/', (req, res) => {
    res.redirect('/pages/dashboard.html');
  });

  // Return 404 for missing /api routes
  app.all('/api*', (req, res) => {
    res.status(404).json({ error: `no such API route: ${req.path}` });
  });

  // Fallback for html pages
  app.use((req, res) => {
    if (req.path.endsWith('.html') || req.path.startsWith('/pages/')) {
      const pagePath = path.resolve(frontendDir, req.path.replace(/^\//, ''));
      res.sendFile(pagePath, (err) => {
        if (err) res.status(404).send('Page not found');
      });
      return;
    }
    res.redirect('/pages/dashboard.html');
  });

  app.listen(PORT, '0.0.0.0', () => {
    console.log(`CampusWatch server running on http://0.0.0.0:${PORT}`);
  });
}

// Auto-run if executed directly
startServer().catch((err) => {
  console.error('Failed to start server:', err);
  process.exit(1);
});
