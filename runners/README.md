# GitHub Actions Custom Runner

This directory contains the configuration for building a custom GitHub Actions runner image with pre-installed prerequisites.

## 🐳 What's Included

The custom runner image includes:

### **Development Tools**
- Go 1.21.5
- Node.js & npm (latest)
- Python 3 & pip
- Build tools (gcc, g++, make)

### **Cloud Tools**
- Azure CLI
- kubectl
- Helm
- Terraform

### **CI/CD Tools**
- Docker Compose
- GitHub CLI
- golangci-lint

### **Package Managers**
- yarn, pnpm
- pipenv, poetry
- npm global packages (TypeScript, ESLint, Prettier)

## 🚀 Building the Runner Image

```bash
# Build and push to your ACR using existing scripts
./scripts/build-images.sh --runner-only

# Or use your existing deploy script which handles everything
./scripts/deploy.sh
```

## ⚙️ Using the Custom Runner

### **Step 1: Update ACR URL in Terraform**

After the image is built, update your ACR URL in the Terraform configuration:

```bash
# Edit terraform/environments/development/terraform.tfvars
# Replace "youracr.azurecr.io" with your actual ACR URL
runner_image = "youracr.azurecr.io/hub/github-runner:latest"
```

### **Step 2: Deploy via Terraform**

```bash
cd terraform/environments/development
terraform plan
terraform apply
```

### **Step 3: Use in Workflows**

```yaml
jobs:
  build:
    runs-on: hub-dev-runners  # Your runner scale set name
    steps:
      - uses: actions/checkout@v4
      
      # Tools are pre-installed, ready to use!
      - name: Run tests
        run: |
          go version      # Go is ready
          node --version  # Node.js is ready
          az --version    # Azure CLI is ready
          kubectl version --client  # kubectl is ready
```

## 🔧 Customizing the Runner

### **Adding Custom Tools**

1. **Edit the Dockerfile** (`runners/Dockerfile`):
   ```dockerfile
   # Add your custom tools
   RUN apt-get update && apt-get install -y your-tool
   ```

2. **Add custom scripts** to `runners/scripts/`:
   ```bash
   # Create executable scripts
   echo '#!/bin/bash\necho "Custom script"' > runners/scripts/my-script.sh
   chmod +x runners/scripts/my-script.sh
   ```

3. **Rebuild the image**:
   ```bash
   ./scripts/build-images.sh --runner-only
   ```

### **Version Management**

```bash
# Build with specific version
./scripts/build-images.sh --runner-only -v v1.2.3

# Update terraform.tfvars with specific version
runner_image = "youracr.azurecr.io/hub/github-runner:v1.2.3"
```

## 📋 Troubleshooting

### **Image Build Fails**
```bash
# Check Docker daemon
docker version

# Build with verbose output
docker build -f runners/Dockerfile . --progress=plain
```

### **Image Not Found in ACR**
```bash
# Check if image exists
az acr repository list --name youracr --output table

# Check image tags
az acr repository show-tags --name youracr --repository hub/github-runner
```

### **Runner Pods Fail to Start**
```bash
# Check runner pod logs
kubectl logs -n arc-runners $(kubectl get pods -n arc-runners -l app.kubernetes.io/name=hub-dev-runners -o jsonpath='{.items[0].metadata.name}')

# Check if image can be pulled
kubectl run test-runner --image=youracr.azurecr.io/hub/github-runner:latest --rm -it --restart=Never -- /bin/bash
```

### **Disable Custom Runner**
```bash
# Comment out runner_image in terraform.tfvars
# runner_image = "youracr.azurecr.io/hub/github-runner:latest"

# Then redeploy
terraform apply
```

## 📦 Available Scripts

- `runners/scripts/` - Custom scripts included in the runner image
- `scripts/build-images.sh --runner-only` - Build only the runner image

## 🔄 CI/CD Integration

The runner image uses your existing CI/CD infrastructure:
1. **Built** using existing `scripts/build-images.sh --runner-only`
2. **Pushed** to ACR using existing credentials and scripts
3. **Deployed** via existing infrastructure pipeline

**Scripts:**
- `scripts/build-images.sh --runner-only` - Build runner image
- `scripts/deploy.sh` - Full deployment including images

## 💡 Best Practices

1. **Pin versions** for production: Use specific version tags instead of `latest`
2. **Test changes** in development environment first
3. **Monitor image size** - current image is ~2GB with all tools
4. **Update regularly** - rebuild monthly for security updates
5. **Cache layers** - organize Dockerfile for optimal layer caching