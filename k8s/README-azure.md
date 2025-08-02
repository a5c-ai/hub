# Azure Application Gateway Integration

This directory contains the necessary configurations to enable Azure Application Gateway Ingress Controller (AGIC) support for the Hub application.

## Files Added for Azure Support

### Core Azure Ingress Configuration
- `ingress-azure.yaml` - Basic Azure Application Gateway ingress configuration
- `azure-ssl-config.yaml` - Enhanced ingress with SSL and security configurations
- `azure-lb-service.yaml` - Azure Load Balancer service configuration

### Supporting Configurations
- `azure-application-gateway-config.yaml` - Azure-specific settings and health probe service
- `services.yaml` - Updated with Azure Load Balancer annotations

## Prerequisites

1. **Azure Application Gateway**: You need an existing Azure Application Gateway with:
   - Application Gateway Ingress Controller (AGIC) installed
   - Proper networking configuration
   - SSL certificates (if using HTTPS)

2. **Azure Resources**: Ensure you have:
   - Azure Kubernetes Service (AKS) cluster
   - Azure Application Gateway
   - Azure Virtual Network with proper subnets
   - Azure DNS zone (for external-dns)

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

## Configuration Steps

### 1. Install Azure Application Gateway Ingress Controller

```bash
# Using Helm
helm repo add application-gateway-kubernetes-ingress https://appgwingress.blob.core.windows.net/ingress-azure-helm-package/
helm repo update

# Or enable as AKS add-on
az aks enable-addons -n myCluster -g myResourceGroup -a ingress-appgw --appgw-id "/subscriptions/{subscription-id}/resourceGroups/{rg}/providers/Microsoft.Network/applicationGateways/{appgw-name}"
```

### 2. Configure SSL Certificates via cert-manager

Ensure cert-manager and ClusterIssuer are in place before applying Azure SSL configuration:

```bash
kubectl apply -f k8s/azure-ssl-config.yaml
kubectl apply -f k8s/cert-manager-issuers.yaml
```

### 3. Update WAF Policy (Optional)

If using Web Application Firewall, update the WAF policy resource ID in the ingress annotations:

```yaml
appgw.ingress.kubernetes.io/waf-policy-for-path: "/subscriptions/{subscription-id}/resourceGroups/{rg}/providers/Microsoft.Network/applicationGatewayWebApplicationFirewallPolicies/{policy-name}"
```

### 4. Configure Azure Resources

Update `azure-application-gateway-config.yaml` with your Azure environment values:

```bash
export AZURE_TENANT_ID="your-tenant-id"
export AZURE_SUBSCRIPTION_ID="your-subscription-id"
export AZURE_RESOURCE_GROUP="your-resource-group"
export AZURE_LOCATION="your-location"
# ... other variables

envsubst < azure-application-gateway-config.yaml | kubectl apply -f -
```

## Deployment Options

### Option 1: Use Existing NGINX Ingress + Azure Load Balancer
Deploy the services with Azure annotations and keep the existing NGINX ingress:

```bash
kubectl apply -f services.yaml
kubectl apply -f azure-lb-service.yaml
```

### Option 2: Use Azure Application Gateway Ingress Controller
Deploy the Azure-specific ingress configurations:

```bash
kubectl apply -f services.yaml
kubectl apply -f ingress-azure.yaml
kubectl apply -f azure-application-gateway-config.yaml
```

### Option 3: Enhanced Azure with SSL and Security
For production environments with enhanced security:

```bash
kubectl apply -f services.yaml
kubectl apply -f azure-ssl-config.yaml
kubectl apply -f azure-application-gateway-config.yaml
```

## Key Azure Annotations Explained

### Required Annotations
- `kubernetes.io/ingress.class: azure/application-gateway` - Marks ingress for AGIC

### SSL/Security Annotations
- `appgw.ingress.kubernetes.io/ssl-redirect: "true"` - Forces HTTPS
- `appgw.ingress.kubernetes.io/appgw-ssl-certificate` - References SSL certificate
- `appgw.ingress.kubernetes.io/waf-policy-for-path` - WAF policy assignment

### Health Probe Annotations
- `appgw.ingress.kubernetes.io/health-probe-*` - Custom health probe configuration

### Performance Annotations
- `appgw.ingress.kubernetes.io/connection-draining: "true"` - Graceful connection handling
- `appgw.ingress.kubernetes.io/request-timeout: "30"` - Request timeout settings

### Load Balancer Annotations (Services)
- `service.beta.kubernetes.io/azure-load-balancer-health-probe-*` - Azure LB health probes
- `service.beta.kubernetes.io/azure-dns-label-name` - Azure DNS label

## Monitoring and Troubleshooting

### Check AGIC Status
```bash
kubectl get pods -n kube-system | grep ingress-appgw
kubectl logs -n kube-system deployment/ingress-appgw-deployment
```

### Verify Ingress Configuration
```bash
kubectl describe ingress hub-ingress-azure -n hub
kubectl get ingress -n hub
```

### Check Service Endpoints
```bash
kubectl get endpoints -n hub
kubectl describe service hub-backend-service -n hub
```

## Security Considerations

1. **Use HTTPS**: Always enable SSL redirect in production
2. **WAF Policy**: Configure appropriate WAF rules for your application
3. **Private IP**: Consider using private IP for internal applications
4. **Network Security Groups**: Configure NSGs to restrict access
5. **Certificate Management**: Use Azure Key Vault for certificate storage

## Performance Tuning

1. **Connection Draining**: Enable for graceful deployments
2. **Health Probes**: Configure appropriate intervals and thresholds
3. **Request Timeouts**: Set based on your application requirements
4. **Cookie Affinity**: Disable unless specifically needed

## Cost Optimization

1. **Standard_v2 SKU**: Use for autoscaling capabilities
2. **Health Probe Configuration**: Optimize probe frequency
3. **Connection Limits**: Configure appropriate limits
4. **Static IP**: Use only if necessary

For more information, see the [Azure Application Gateway Ingress Controller documentation](https://docs.microsoft.com/en-us/azure/application-gateway/ingress-controller-overview).
