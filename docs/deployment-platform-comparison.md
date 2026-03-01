# Deployment Platform Comparison

Comparison of deployment platforms for a Go/htmx application with requirements:
- Affordable pricing
- Good developer experience
- Managed database support
- European hosting (ideally Germany)

*Last updated: March 2026*

---

## Summary Table

| Feature | Fly.io | Render | Railway |
|---------|--------|--------|---------|
| **Germany Region** | Yes (Frankfurt) | Yes (Frankfurt) | No (Amsterdam closest) |
| **Managed Postgres** | Yes ($38+/mo) | Yes ($6+/mo) | No (self-managed templates) |
| **Min Monthly Cost** | ~$0 (pay-as-you-go) | $0-7/mo | $5/mo |
| **Pricing Model** | Per-second usage | Fixed instance pricing | Per-second usage |
| **Free Tier** | Trial credits | Limited free instances | 30-day trial ($5 credit) |
| **CLI Tool** | `flyctl` | `render` CLI | `railway` CLI |
| **Docker Support** | Yes (native) | Yes | Yes |
| **Auto-deploy from Git** | Yes | Yes | Yes |
| **Hard Spending Limits** | No | No | Yes |

---

## 1. Fly.io

### Overview
Fly.io runs applications on their own hardware in datacenters worldwide. Apps are deployed as "Machines" (lightweight VMs) that can auto-stop when idle to save costs.

### Regions
- **Frankfurt, Germany (`fra`)** - Available
- Amsterdam (`ams`), Paris (`cdg`), London (`lhr`), Stockholm (`arn`)
- Plus US, Asia-Pacific, South America regions

### Pricing

#### Compute (Frankfurt region)
| Size | vCPU | RAM | Monthly (24/7) |
|------|------|-----|----------------|
| shared-cpu-1x | 1 shared | 256MB | ~$2.24 |
| shared-cpu-1x | 1 shared | 1GB | ~$7.12 |
| shared-cpu-2x | 2 shared | 2GB | ~$14.24 |
| performance-1x | 1 dedicated | 2GB | ~$38.75 |

- Billed per second when running
- Stopped machines: $0.15/GB/month (rootfs only)
- **Cost optimization**: Machines can auto-stop when no traffic

#### Managed Postgres
| Plan | CPU | Memory | Monthly |
|------|-----|--------|---------|
| Basic | Shared-2x | 1GB | $38.00 |
| Starter | Shared-2x | 2GB | $72.00 |
| Launch | Performance-2x | 8GB | $282.00 |

- Storage: $0.28/GB/month
- Includes: HA, backups, connection pooling, encryption
- Frankfurt region available for Managed Postgres

#### Other Costs
- Volumes: $0.15/GB/month
- Dedicated IPv4: $2/month
- SSL certificates: $0.10/month (first 10 free)
- Egress (Europe): $0.02/GB

### Developer Experience

**Deployment workflow:**
```bash
# Install CLI
brew install flyctl

# Login
fly auth login

# Launch app (auto-detects Go/Dockerfile)
fly launch

# Deploy updates
fly deploy

# View logs
fly logs

# SSH into machine
fly ssh console
```

**Configuration (`fly.toml`):**
```toml
app = "my-app"
primary_region = "fra"

[build]
  dockerfile = "Dockerfile"

[http_service]
  internal_port = 8080
  force_https = true
  auto_stop_machines = true
  auto_start_machines = true
  min_machines_running = 0

[[vm]]
  memory = "512mb"
  cpu_kind = "shared"
  cpus = 1
```

### Pros
- Frankfurt region available
- Machines can auto-stop (pay only when used)
- Fully managed Postgres with HA
- Good CLI tooling
- Per-second billing
- Private networking between services
- WireGuard-based VPN access

### Cons
- Managed Postgres starts at $38/month (expensive for hobby projects)
- No hard spending limits
- Credit card required
- Unmanaged Postgres option requires self-management

### Estimated Monthly Cost (Small App)
- 1x shared-cpu-1x (1GB RAM, auto-stop): ~$3-7/month
- Managed Postgres Basic: $38/month
- 10GB storage: $2.80/month
- **Total: ~$44-48/month** with managed DB
- **Or: ~$5-10/month** with self-managed Postgres on a volume

---

## 2. Render

### Overview
Render is a cloud platform focused on simplicity. Git-based deployments with automatic builds and deploys.

### Regions
- **Frankfurt, Germany** - Available
- Oregon, Ohio, Virginia (USA)
- Singapore

### Pricing

#### Web Services
| Instance | RAM | CPU | Monthly |
|----------|-----|-----|---------|
| Free | 512MB | 0.1 | $0 (with limitations) |
| Starter | 512MB | 0.5 | $7 |
| Standard | 2GB | 1 | $25 |
| Pro | 4GB | 2 | $85 |

- Fixed monthly pricing (no per-second billing)
- Free tier: spins down after 15 min inactivity, limited hours

#### Render Postgres
| Tier | RAM | CPU | Monthly |
|------|-----|-----|---------|
| Free | 256MB | 0.1 | $0 (30-day limit) |
| Basic-256mb | 256MB | 0.1 | $6 |
| Basic-1gb | 1GB | 0.5 | $19 |
| Pro-4gb | 4GB | 1 | $55 |

- Storage: $0.30/GB (expandable)
- Includes: Point-in-time recovery (paid), logical backups
- High availability: Pro tier and above

#### Other Costs
- Persistent disk: $0.25/GB/month
- Bandwidth: 100GB included (Hobby), then varies by plan

### Developer Experience

**Deployment workflow:**
```bash
# Connect GitHub/GitLab repo via dashboard
# Or use CLI:
render login
render deploy
```

**Configuration (`render.yaml` - Blueprint):**
```yaml
services:
  - type: web
    name: my-app
    runtime: go
    region: frankfurt
    plan: starter
    buildCommand: go build -o app .
    startCommand: ./app
    envVars:
      - key: DATABASE_URL
        fromDatabase:
          name: my-db
          property: connectionString

databases:
  - name: my-db
    region: frankfurt
    plan: basic-256mb
```

### Pros
- Frankfurt region available
- Simpler mental model (fixed pricing)
- Managed Postgres starting at $6/month
- Free tier available (with limitations)
- Good dashboard UI
- Zero-downtime deploys
- Preview environments for PRs

### Cons
- No per-second billing (pay for full month)
- Free tier has severe limitations (spin-down, limited hours)
- No hard spending limits
- Less flexibility than Fly.io

### Estimated Monthly Cost (Small App)
- 1x Starter web service: $7/month
- Basic Postgres (256MB): $6/month
- **Total: ~$13/month**

---

## 3. Railway

### Overview
Railway offers a visual canvas for infrastructure with excellent DX. Usage-based pricing with included credits.

### Regions
- US West (California)
- US East (Virginia)
- **EU West (Amsterdam, Netherlands)** - Closest to Germany (~400km)
- Southeast Asia (Singapore)

**Note: No Frankfurt/Germany region available**

### Pricing

#### Plans
| Plan | Min Monthly | Included Credits | Limits |
|------|-------------|------------------|--------|
| Free (Trial) | $0 | $5 one-time | 30 days |
| Hobby | $5 | $5/month | 48 vCPU, 48GB RAM |
| Pro | $20 | $20/month | 1000 vCPU, 1TB RAM |

#### Usage Rates
- Memory: $0.000386/GB/min (~$10/GB/month)
- CPU: $0.000772/vCPU/min (~$20/vCPU/month)
- Volumes: $0.00000006/GB/sec (~$0.15/GB/month)
- Egress: $0.05/GB

### Developer Experience

**Deployment workflow:**
```bash
# Install CLI
npm install -g @railway/cli

# Login
railway login

# Initialize project
railway init

# Deploy
railway up

# View logs
railway logs

# Open dashboard
railway open
```

**Configuration (`railway.toml`):**
```toml
[build]
builder = "dockerfile"
dockerfilePath = "Dockerfile"

[deploy]
startCommand = "./app"
healthcheckPath = "/health"
restartPolicyType = "ON_FAILURE"
```

### Database Options
Railway provides database **templates** (not fully managed):
- PostgreSQL, MySQL, Redis, MongoDB
- You deploy and manage them yourself
- Includes volumes for persistence
- No automatic backups (you configure them)

```bash
# Add Postgres from template
railway add -t postgres
```

### Pros
- Excellent developer experience
- Visual infrastructure canvas
- **Hard spending limits** (unique feature)
- Usage-based pricing with included credits
- One-click database templates
- Preview environments
- Simple CLI

### Cons
- **No Germany region** (Amsterdam is closest)
- Databases are unmanaged (you handle backups, HA)
- $5/month minimum (Hobby plan)
- Amsterdam adds ~5-10ms latency vs Frankfurt for German users

### Estimated Monthly Cost (Small App)
- Hobby plan: $5/month (includes $5 credits)
- Small app + Postgres template: Usually within $5-15/month
- **Total: ~$5-15/month**

---

## Recommendation

### For Germany/EU with Managed Database: **Fly.io**

Best choice if you need:
- Frankfurt region (lowest latency for German users)
- Fully managed Postgres with HA and backups
- Pay-per-use model with auto-stop capability

**Estimated cost: $44-50/month** with managed Postgres, or **$5-10/month** with self-managed.

### For Simplicity and Lower Cost: **Render**

Best choice if you want:
- Simple, predictable pricing
- Frankfurt region
- Cheap managed Postgres ($6/month)
- Less operational complexity

**Estimated cost: ~$13/month**

### For Best DX (if EU-West is acceptable): **Railway**

Best choice if you:
- Prioritize developer experience
- Can accept Amsterdam region (~400km from Germany)
- Want hard spending limits
- Prefer visual infrastructure management

**Estimated cost: ~$5-15/month**

---

## Quick Start Commands

### Fly.io
```bash
brew install flyctl
fly auth signup
cd your-app
fly launch --region fra
fly postgres create --region fra
fly deploy
```

### Render
```bash
# Via dashboard: Connect repo, select Frankfurt, deploy
# Or with render.yaml blueprint in repo
```

### Railway
```bash
npm install -g @railway/cli
railway login
railway init
railway add -t postgres
railway up
```

---

## Additional Considerations

### GDPR Compliance
All three platforms offer GDPR-compliant hosting in EU regions:
- **Fly.io**: Frankfurt region, DPA available
- **Render**: Frankfurt region, GDPR DPA included
- **Railway**: Amsterdam region (EU), GDPR compliant

### For Very Low Budget
If cost is the primary concern and you're comfortable with more ops work:
- Consider **Hetzner Cloud** (German company, ~$4/month for VPS)
- Run your own Postgres on the same server
- Use Coolify or CapRover for deployment automation
