# Health Check Configuration Fix

## ✅ Issue Resolved

**Problem**: Kubernetes startup probe failing with HTTP 404
- App was starting successfully (`✓ Ready in 507ms`)
- Listening on correct port (3000)
- But health probes were hitting non-existent `/health` endpoint

## Root Cause

Health probes were configured to check `/health` but Next.js API route is at `/api/health`:

- **Kubernetes probes**: `/health` ❌
- **Actual API route**: `/api/health` ✅

## Files Updated

### 1. Frontend Deployment (`k8s/frontend-deployment.yaml`)
```yaml
# Before: All probes used /health
livenessProbe:
  httpGet:
    path: /health  # ❌

# After: All probes use /api/health  
livenessProbe:
  httpGet:
    path: /api/health  # ✅
```

**Updated all three probe types:**
- ✅ `livenessProbe: /api/health`
- ✅ `readinessProbe: /api/health` 
- ✅ `startupProbe: /api/health`

### 2. Azure Load Balancer Service (`k8s/services.yaml`)
```yaml
# Before
service.beta.kubernetes.io/azure-load-balancer-health-probe-request-path: "/"

# After  
service.beta.kubernetes.io/azure-load-balancer-health-probe-request-path: "/api/health"
```

### 3. Ingress Routes (Multiple files)
Updated health check routes in:
- `k8s/ingress.yaml`
- `k8s/ingress-azure.yaml` 
- `k8s/azure-ssl-config.yaml`

```yaml
# Now routes /api/health properly to backend service
- path: /api/health
  pathType: Prefix
  backend:
    service:
      name: hub-backend-service  # Correctly routes to backend
```

## Expected Results

After redeployment:
- ✅ **Startup probe will succeed** (HTTP 200 from `/api/health`)
- ✅ **Readiness probe will succeed** 
- ✅ **Liveness probe will succeed**
- ✅ **Azure Load Balancer health checks will pass**

## Health Endpoint Details

The existing health API at `frontend/src/app/api/health/route.ts`:
```typescript
export async function GET() {
  return NextResponse.json({ status: 'ok' });
}
```

Routes to: **`/api/health`** (not `/health`)

## Container Logs Verification

After this fix, you should see:
```
✓ Starting...
✓ Ready in 507ms
- Local: http://localhost:3000
- Network: http://0.0.0.0:3000
```

**No more**: `Startup probe failed: HTTP probe failed with statuscode: 404`

## Metadata Warnings (Non-Critical)

The warnings about `viewport` and `themeColor` are just Next.js recommendations:
```
⚠ Unsupported metadata themeColor is configured in metadata export
⚠ Unsupported metadata viewport is configured in metadata export  
```

These are **warnings only** and don't affect functionality. They can be addressed separately if needed.