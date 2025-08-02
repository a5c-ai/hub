# Terraform Import Commands for cert-manager

## Step 1: Check existing Helm releases
```bash
# Connect to your AKS cluster first
az aks get-credentials --resource-group <your-rg> --name <your-aks-cluster>

# Check existing cert-manager release
helm list -n cert-manager
```

## Step 2: Import the existing release
```bash
# Navigate to your terraform environment
cd terraform/environments/development  # or staging/production

# Import the existing cert-manager Helm release
terraform import module.cert_manager.helm_release.cert_manager cert-manager/cert-manager
```

## Step 3: Continue with terraform apply
```bash
terraform plan
terraform apply
```

## If import fails, you can force replace:
```bash
terraform apply -replace="module.cert_manager.helm_release.cert_manager"
```