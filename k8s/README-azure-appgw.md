# Azure Application Gateway Integration for Hub

This document describes the Azure Application Gateway integration for the Hub application, providing enterprise-grade ingress capabilities with Web Application Firewall (WAF) protection.

## Overview

The Hub application now supports Azure Application Gateway as the primary ingress controller, replacing the default NGINX ingress. This provides:

- **Enterprise Security**: Built-in Web Application Firewall (WAF) with OWASP rule sets
- **SSL/TLS Termination**: Automatic certificate management and SSL offloading
- **High Availability**: Multi-zone deployment with Azure's native load balancing
- **Azure Integration**: Native integration with Azure DNS, Key Vault, and monitoring
- **Performance**: Optimized for Azure infrastructure with connection draining and health probes

## Architecture

```
Internet → Azure Application Gateway → AKS Cluster → Hub Services
                    ↓
               WAF Protection
               SSL Termination
               Health Probes
               Load Balancing
```

## Components

### 1. Terraform Infrastructure

The Azure Application Gateway is provisioned through the `security` Terraform module with:

- **Public IP**: Static IP with Standard SKU for the Application Gateway
- **WAF Policy**: OWASP 3.2 ruleset with configurable exclusions
- **SSL Configuration**: Certificate management for HTTPS
- **Health Probes**: Custom health checks for backend services
- **Network Security Groups**: Proper firewall rules for Application Gateway

### 2. AKS Integration

The AKS cluster is configured with:

- **AGIC (Application Gateway Ingress Controller)**: Azure addon that manages the Application Gateway
- **Role Assignments**: Proper permissions for AGIC to manage the Application Gateway
- **Identity Management**: User-assigned identity for secure authentication

### 3. Kubernetes Resources

#### Primary Ingress (`ingress.yaml`)
The main ingress resource configured for Azure Application Gateway with:
- Azure-specific annotations for health probes, SSL redirect, and connection management
- External DNS integration for automatic DNS record creation
- Proper routing for API, health, and frontend endpoints

#### Azure-Specific Configurations
- `ingress-azure.yaml`: Alternative Azure-specific ingress configuration
- `azure-ssl-config.yaml`: SSL-specific configuration for Azure environments
- `azure-application-gateway-config.yaml`: ConfigMap with Azure-specific settings

## Configuration Files

### Ingress Annotations

Key annotations for Azure Application Gateway:

```yaml
annotations:
  kubernetes.io/ingress.class: azure/application-gateway
  appgw.ingress.kubernetes.io/ssl-redirect: "true"
  appgw.ingress.kubernetes.io/health-probe-hostname: hub.a5c.ai
  appgw.ingress.kubernetes.io/health-probe-path: "/health"
  appgw.ingress.kubernetes.io/backend-protocol: "http"
  appgw.ingress.kubernetes.io/rule-priority: "100"
  external-dns.alpha.kubernetes.io/hostname: hub.a5c.ai
```

### Environment-Specific Configuration

Each environment (development, staging, production) is configured with:

- **Development**: Minimal capacity (1 instance), Detection mode WAF
- **Staging**: Production-like configuration with Detection mode WAF
- **Production**: Full capacity, Prevention mode WAF with strict security rules

## Deployment

### Prerequisites

1. Azure subscription with proper permissions
2. AKS cluster deployed
3. Application Gateway subnet in the VNet
4. DNS zone configured for external DNS

### Let's Encrypt Certificate Provisioning

We use cert-manager to automate TLS certificates from Let's Encrypt. To set this up:

1. **Install cert-manager CRDs**:
   ```bash
   kubectl apply --validate=false -f https://github.com/jetstack/cert-manager/releases/download/v1.9.1/cert-manager.crds.yaml
   ```
2. **Install cert-manager via Helm**:
   ```bash
   helm repo add jetstack https://charts.jetstack.io
   helm repo update
   helm install cert-manager jetstack/cert-manager --namespace cert-manager --create-namespace --version v1.9.1
   ```
3. **Apply ClusterIssuer and Certificate**:
   ```bash
   kubectl apply -f k8s/cert-manager-issuers.yaml
   ```

cert-manager will automatically request and renew certificates for `hub.a5c.ai`, populating the `hub-azure-ssl-certificate` secret.

### Terraform Deployment

1. **Plan the deployment**:
   ```bash
   cd terraform/environments/[environment]
   terraform plan
   ```

2. **Apply the infrastructure**:
   ```bash
   terraform apply
   ```

3. **Verify AGIC installation**:
   ```bash
   kubectl get pods -n kube-system | grep ingress-appgw
   ```

### Kubernetes Deployment

1. **Apply the ingress configuration**:
   ```bash
   kubectl apply -f k8s/ingress.yaml
   ```

2. **Verify ingress status**:
   ```bash
   kubectl get ingress hub-ingress -n hub
   ```

3. **Check Application Gateway configuration**:
   ```bash
   kubectl describe ingress hub-ingress -n hub
   ```

## Monitoring and Troubleshooting

### Health Checks

The Application Gateway performs health checks on:
- **Path**: `/health`
- **Port**: `8080`
- **Protocol**: `HTTP`
- **Interval**: `30 seconds`
- **Timeout**: `30 seconds`
- **Unhealthy Threshold**: `3 failures`

### Common Issues

1. **AGIC Pod Failures**:
   - Check role assignments: `kubectl logs -n kube-system -l app=ingress-appgw`
   - Verify Application Gateway permissions

2. **SSL Certificate Issues**:
   - Ensure proper certificate configuration in Terraform
   - Check Key Vault access permissions

3. **Health Probe Failures**:
   - Verify backend services are responding on `/health`
   - Check security group rules

### Monitoring

Monitor the Application Gateway through:
- **Azure Monitor**: Built-in metrics and logs
- **Log Analytics**: Centralized logging with the AKS cluster
- **Azure Application Insights**: Application performance monitoring

## Security Features

### Web Application Firewall (WAF)

- **OWASP Rule Set 3.2**: Protection against common web vulnerabilities
- **Custom Rules**: Environment-specific security rules
- **Rate Limiting**: Protection against DDoS attacks
- **Request Filtering**: Content-based filtering and blocking

### SSL/TLS

- **TLS 1.2+**: Modern encryption standards
- **Certificate Management**: Automated certificate rotation
- **HSTS**: HTTP Strict Transport Security headers

### Network Security

- **Private Endpoints**: Optional private connectivity
- **Network Security Groups**: Layered firewall protection
- **VNet Integration**: Secure communication within Azure VNet

## Comparison with NGINX Ingress

| Feature | Azure Application Gateway | NGINX Ingress |
|---------|--------------------------|---------------|
| WAF | Built-in OWASP rules | Requires additional setup |
| SSL Termination | Native Azure integration | cert-manager required |
| Health Probes | Advanced Azure-native | Basic HTTP checks |
| Monitoring | Azure Monitor integration | Requires external setup |
| Cost | Azure service pricing | Open source (infrastructure cost only) |
| Performance | Azure-optimized | General purpose |

## Migration from NGINX

If migrating from NGINX ingress:

1. **Backup current configuration**:
   ```bash
   kubectl get ingress hub-ingress -o yaml > nginx-ingress-backup.yaml
   ```

2. **Update ingress class**:
   - Change `kubernetes.io/ingress.class` to `azure/application-gateway`
   - Replace NGINX annotations with Azure Application Gateway annotations

3. **Update SSL certificates**:
   - Use Azure-managed certificates or configure in Key Vault
   - Update `secretName` to match Azure certificate configuration

4. **Test thoroughly**:
   - Verify all endpoints are accessible
   - Check health probe functionality
   - Validate SSL certificate installation

## Support

For issues related to:
- **Infrastructure**: Check Terraform configuration and Azure resources
- **AGIC**: Review AKS addon status and role assignments
- **Application**: Verify backend service health and connectivity

Useful commands:
```bash
# Check AGIC logs
kubectl logs -n kube-system -l app=ingress-appgw

# View ingress events
kubectl describe ingress hub-ingress -n hub

# Check Application Gateway status
az network application-gateway show --resource-group <rg> --name <appgw-name>
```
