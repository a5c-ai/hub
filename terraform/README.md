# Hub Infrastructure as Code

This directory contains Terraform infrastructure as code (IaC) for deploying the Hub git hosting service to Azure. The infrastructure follows Azure best practices with a focus on security, scalability, and cost optimization.

## Architecture Overview

The Hub infrastructure is designed as a microservices-based architecture optimized for Azure deployment:

### Core Services
- **Azure Kubernetes Service (AKS)**: Container orchestration platform
- **Azure Database for PostgreSQL**: Managed database service with high availability
- **Azure Blob Storage**: Object storage for repositories, artifacts, and packages
- **Azure Key Vault**: Secret and key management with private endpoints
- **Azure Virtual Network**: Network isolation and segmentation
- **Azure Application Gateway**: Load balancing with Web Application Firewall
- **Azure Monitor & Log Analytics**: Comprehensive monitoring and logging

### Security Features
- Private endpoints for all PaaS services
- Network security groups with least-privilege access
- Web Application Firewall with OWASP rules
- Azure Active Directory integration
- Key Vault for secret management
- Role-based access control (RBAC)

## Directory Structure

```
terraform/
├── modules/                    # Reusable Terraform modules
│   ├── aks/                   # Azure Kubernetes Service
│   ├── keyvault/              # Azure Key Vault
│   ├── monitoring/            # Azure Monitor & Application Insights
│   ├── networking/            # Virtual Network & Security Groups
│   ├── postgresql/            # PostgreSQL Flexible Server
│   ├── resource_group/        # Resource Group
│   ├── security/              # Application Gateway & WAF
│   └── storage/               # Azure Storage Account
├── environments/              # Environment-specific configurations
│   ├── development/           # Development environment
│   ├── staging/               # Staging environment
│   └── production/            # Production environment
├── scripts/                   # Deployment and utility scripts
│   ├── deploy.sh             # Main deployment script
│   ├── destroy.sh            # Safe destruction script
│   └── plan.sh               # Planning and analysis script
└── README.md                 # This file
```

## Prerequisites

### Required Tools
- [Terraform](https://terraform.io/) 1.5 or later
- [Azure CLI](https://docs.microsoft.com/en-us/cli/azure/) 2.50 or later
- [jq](https://stedolan.github.io/jq/) (for JSON processing in scripts)

### Azure Setup
1. **Azure Subscription**: Ensure you have an active Azure subscription
2. **Azure CLI Login**: Run `az login` to authenticate
3. **Permissions**: Ensure your account has Contributor access to the subscription
4. **Service Principal** (optional): For automated deployments, create a service principal

### Terraform Backend (Recommended)
For production deployments, configure remote state storage:

```bash
# Create storage account for Terraform state
az group create --name tfstate --location "East US"
az storage account create --name tfstateXXXXX --resource-group tfstate --location "East US" --sku Standard_LRS
az storage container create --name tfstate --account-name tfstateXXXXX
```

## Quick Start

### 1. Choose Environment
Navigate to the desired environment:
```bash
cd terraform/environments/development  # or staging/production
```

### 2. Configure Variables
Copy and customize the variables file:
```bash
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with your specific values
```

### 3. Configure Backend (Optional)
Create a backend configuration file:
```bash
cat > backend.conf << EOF
resource_group_name  = "tfstate"
storage_account_name = "tfstateXXXXX"
container_name       = "tfstate"
key                  = "development.terraform.tfstate"
EOF
```

### 4. Deploy Infrastructure
Use the deployment script for a guided deployment:
```bash
cd ../../  # Back to terraform root
./scripts/deploy.sh development
```

Or deploy manually:
```bash
cd environments/development
terraform init -backend-config=backend.conf
terraform plan
terraform apply
```

## Environment Configurations

### Development Environment
- **Purpose**: Development and testing
- **Cost**: Optimized for minimal cost
- **Scale**: 2-3 AKS nodes, single availability zone
- **Database**: Basic tier with minimal storage
- **Features**: Simplified monitoring, no Grafana

**Key Characteristics:**
- Single availability zone deployment
- Basic PostgreSQL tier (B_Standard_B1ms)
- Standard storage with LRS replication
- Minimal node count for AKS
- WAF in Detection mode
- 30-day log retention

### Staging Environment
- **Purpose**: Production-like testing and validation
- **Cost**: Balanced cost and functionality
- **Scale**: 3-6 AKS nodes, two availability zones
- **Database**: General Purpose tier with moderate storage
- **Features**: Full monitoring, optional Grafana

**Key Characteristics:**
- Two availability zone deployment
- General Purpose PostgreSQL (GP_Standard_D2s_v3)
- Standard storage with LRS replication
- Moderate auto-scaling ranges
- WAF in Detection mode (less restrictive)
- 60-day log retention

### Production Environment
- **Purpose**: Live production workloads
- **Cost**: Optimized for performance and reliability
- **Scale**: 5-20 AKS nodes, three availability zones
- **Database**: High availability with geo-redundant backups
- **Features**: Full monitoring, Grafana, enhanced security

**Key Characteristics:**
- Three availability zone deployment
- High-performance PostgreSQL (GP_Standard_D4s_v3)
- Zone-redundant storage with ZRS replication
- High availability database with standby
- WAF in Prevention mode
- 90-day log retention
- Geo-redundant backups
- Additional worker node pools
- Enhanced monitoring and alerting

## Deployment Scripts

### deploy.sh
Main deployment script with safety checks and validation:
```bash
./scripts/deploy.sh [environment] [action] [auto_approve]

# Examples:
./scripts/deploy.sh development          # Deploy to development
./scripts/deploy.sh staging apply        # Deploy to staging
./scripts/deploy.sh production plan      # Plan production deployment
```

**Features:**
- Prerequisites validation
- Azure login verification
- Environment-specific safety checks
- Production deployment protection
- Terraform state backup
- Comprehensive error handling

### plan.sh
Planning and analysis script:
```bash
./scripts/plan.sh [environment] [output_format]

# Examples:
./scripts/plan.sh development           # Plan with human-readable output
./scripts/plan.sh staging json         # Plan with JSON output
```

**Features:**
- Plan generation and analysis
- Resource change summary
- Cost estimation reminders
- Multiple output formats
- Configuration validation

### destroy.sh
Safe destruction script with safeguards:
```bash
./scripts/destroy.sh [environment] DESTROY

# Examples:
./scripts/destroy.sh development DESTROY    # Destroy development
./scripts/destroy.sh staging DESTROY        # Destroy staging
```

**Safety Features:**
- Production destruction prevention
- Multiple confirmation steps
- State backup before destruction
- Resource enumeration
- Post-destruction cleanup

## Module Documentation

### AKS Module
Deploys Azure Kubernetes Service with:
- System-assigned managed identity
- Auto-scaling node pools
- Azure CNI networking
- Log Analytics integration
- Key Vault secrets provider
- Azure Policy integration

### PostgreSQL Module
Deploys PostgreSQL Flexible Server with:
- Private DNS zone integration
- High availability options
- Automated backups with geo-replication
- Custom parameter configurations
- Diagnostic logging
- Network isolation

### Storage Module
Deploys Azure Storage Account with:
- Private endpoint connectivity
- Lifecycle management policies
- Multiple container types
- Network access controls
- Diagnostic monitoring
- Backup retention policies

### Key Vault Module
Deploys Azure Key Vault with:
- RBAC or access policy authorization
- Private endpoint integration
- Secret and key management
- Rotation policies
- Audit logging
- Network restrictions

### Networking Module
Deploys networking infrastructure:
- Virtual Network with multiple subnets
- Network Security Groups
- Private DNS zones
- Service endpoints
- Subnet delegations

### Security Module
Deploys security components:
- Application Gateway with WAF
- SSL termination
- Custom WAF rules
- Rate limiting
- Health probes
- Diagnostic logging

### Monitoring Module
Deploys monitoring infrastructure:
- Log Analytics workspace
- Application Insights
- Action groups for alerting
- Default metric alerts
- Optional Grafana dashboard
- Data collection rules

## Security Considerations

### Network Security
- All PaaS services use private endpoints
- Network Security Groups restrict traffic
- Virtual network isolation
- Service endpoints for enhanced security

### Identity and Access
- Managed identities for AKS
- RBAC for resource access
- Key Vault for secret management
- Least privilege access principles

### Data Protection
- Encryption at rest for all services
- TLS 1.3 for data in transit
- Backup encryption with separate keys
- Geo-redundant backups for production

### Monitoring and Compliance
- Comprehensive audit logging
- Security alerts and monitoring
- Compliance with Azure best practices
- Regular security assessments

## Cost Optimization

### Development Environment
- Single availability zone
- Basic tier services
- Minimal node counts
- Shorter retention periods
- Local redundancy storage

### Production Environment
- Right-sized instances
- Auto-scaling configurations
- Lifecycle management
- Reserved instance planning
- Cost monitoring and alerts

## Troubleshooting

### Common Issues

#### Terraform Init Fails
```bash
# Clear Terraform cache
rm -rf .terraform .terraform.lock.hcl
terraform init
```

#### Resource Already Exists
```bash
# Import existing resource
terraform import azurerm_resource_group.example /subscriptions/xxx/resourceGroups/xxx
```

#### Authentication Issues
```bash
# Re-authenticate with Azure
az login
az account set --subscription "your-subscription-id"
```

#### State Lock Issues
```bash
# Force unlock (use with caution)
terraform force-unlock LOCK_ID
```

### Debugging

#### Enable Terraform Logging
```bash
export TF_LOG=DEBUG
export TF_LOG_PATH=terraform.log
terraform plan
```

#### Azure Resource Debugging
```bash
# Check resource group
az group show --name "rg-hub-dev"

# Check AKS cluster
az aks show --resource-group "rg-hub-dev" --name "aks-hub-dev"

# Check Key Vault access
az keyvault list --resource-group "rg-hub-dev"
```

## Maintenance

### Regular Tasks
1. **Update Terraform providers** monthly
2. **Review and rotate secrets** quarterly
3. **Update Kubernetes versions** as available
4. **Review monitoring alerts** weekly
5. **Cost optimization review** monthly

### Backup Strategy
1. **Terraform state** backup before changes
2. **Database backups** automated daily
3. **Storage account** geo-replication
4. **Configuration files** version controlled

## Contributing

### Module Development
1. Follow Terraform best practices
2. Include comprehensive variable validation
3. Provide detailed outputs
4. Document all resources
5. Test with multiple environments

### Environment Updates
1. Test changes in development first
2. Validate with staging environment
3. Use gradual rollout for production
4. Monitor after deployments

## Support

### Documentation
- [Terraform Azure Provider](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs)
- [Azure Documentation](https://docs.microsoft.com/en-us/azure/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)

### Getting Help
1. Check this README and module documentation
2. Review Terraform and Azure documentation
3. Search existing issues in the repository
4. Contact the infrastructure team

## License

This infrastructure code is part of the Hub project and follows the same licensing terms as defined in the main project repository.