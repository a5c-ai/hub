#!/bin/bash

# Hub Infrastructure Deployment Script
# This script deploys the Hub infrastructure to Azure using Terraform

set -e

# Default values
ENVIRONMENT="${1:-development}"
ACTION="${2:-apply}"
AUTO_APPROVE="${3:-false}"
BACKEND_CONFIG=""

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if required tools are installed
check_prerequisites() {
    print_info "Checking prerequisites..."
    
    if ! command -v terraform &> /dev/null; then
        print_error "Terraform is not installed. Please install Terraform 1.5 or later."
        exit 1
    fi
    
    if ! command -v az &> /dev/null; then
        print_error "Azure CLI is not installed. Please install Azure CLI."
        exit 1
    fi
    
    # Check Terraform version
    TERRAFORM_VERSION=$(terraform version -json | jq -r '.terraform_version')
    REQUIRED_VERSION="1.5.0"
    
    if ! printf '%s\n' "$REQUIRED_VERSION" "$TERRAFORM_VERSION" | sort -V -C; then
        print_error "Terraform version $TERRAFORM_VERSION is too old. Required: $REQUIRED_VERSION or later."
        exit 1
    fi
    
    print_success "Prerequisites check passed"
}

# Function to check Azure login status
check_azure_login() {
    print_info "Checking Azure login status..."
    
    if ! az account show &> /dev/null; then
        print_error "Not logged in to Azure. Please run 'az login' first."
        exit 1
    fi
    
    ACCOUNT_INFO=$(az account show --query '{name:name, id:id, tenantId:tenantId}' -o table)
    print_info "Currently logged in to Azure:"
    echo "$ACCOUNT_INFO"
    
    read -p "Continue with this account? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "Please login to the correct Azure account and try again."
        exit 1
    fi
}

# Function to validate environment
validate_environment() {
    if [[ ! "$ENVIRONMENT" =~ ^(development|staging|production)$ ]]; then
        print_error "Invalid environment: $ENVIRONMENT"
        print_info "Valid environments: development, staging, production"
        exit 1
    fi
    
    if [[ ! -d "environments/$ENVIRONMENT" ]]; then
        print_error "Environment directory not found: environments/$ENVIRONMENT"
        exit 1
    fi
    
    print_success "Environment validation passed: $ENVIRONMENT"
}

# Function to set up backend configuration
setup_backend() {
    print_info "Setting up Terraform backend..."
    
    # Check if backend configuration file exists
    BACKEND_FILE="environments/$ENVIRONMENT/backend.conf"
    if [[ -f "$BACKEND_FILE" ]]; then
        BACKEND_CONFIG="-backend-config=$BACKEND_FILE"
        print_success "Using backend configuration: $BACKEND_FILE"
    else
        print_warning "Backend configuration file not found: $BACKEND_FILE"
        print_info "You may need to provide backend configuration during init"
    fi
}

# Function to show deployment plan
show_plan() {
    print_info "Deployment Configuration:"
    echo "  Environment: $ENVIRONMENT"
    echo "  Action: $ACTION"
    echo "  Auto-approve: $AUTO_APPROVE"
    echo "  Working directory: environments/$ENVIRONMENT"
    echo
    
    if [[ "$ENVIRONMENT" == "production" ]]; then
        print_warning "You are about to deploy to PRODUCTION!"
        print_warning "This will create/modify production resources."
        echo
        read -p "Are you absolutely sure you want to continue? (type 'yes' to confirm): " -r
        if [[ $REPLY != "yes" ]]; then
            print_info "Deployment cancelled."
            exit 0
        fi
    fi
}

# Function to initialize Terraform
terraform_init() {
    print_info "Initializing Terraform..."
    cd "environments/$ENVIRONMENT"
    
    if terraform init $BACKEND_CONFIG -upgrade; then
        print_success "Terraform initialization completed"
    else
        print_error "Terraform initialization failed"
        exit 1
    fi
}

# Function to validate Terraform configuration
terraform_validate() {
    print_info "Validating Terraform configuration..."
    
    if terraform validate; then
        print_success "Terraform validation passed"
    else
        print_error "Terraform validation failed"
        exit 1
    fi
}

# Function to plan Terraform deployment
terraform_plan() {
    print_info "Creating Terraform plan..."
    
    PLAN_FILE="tfplan-$(date +%Y%m%d-%H%M%S)"
    
    if terraform plan -out="$PLAN_FILE"; then
        print_success "Terraform plan created: $PLAN_FILE"
        echo
        print_info "Plan summary:"
        terraform show -no-color "$PLAN_FILE" | grep -E "Plan:|No changes"
        return 0
    else
        print_error "Terraform plan failed"
        exit 1
    fi
}

# Function to apply Terraform deployment
terraform_apply() {
    if [[ "$AUTO_APPROVE" == "true" ]]; then
        APPROVE_FLAG="-auto-approve"
        print_warning "Auto-approve is enabled"
    else
        APPROVE_FLAG=""
        echo
        print_info "Review the plan above and confirm deployment."
    fi
    
    print_info "Applying Terraform plan..."
    
    if terraform apply $APPROVE_FLAG "$PLAN_FILE"; then
        print_success "Terraform deployment completed successfully!"
        
        # Show important outputs
        print_info "Deployment outputs:"
        terraform output
        
        # Clean up plan file
        rm -f "$PLAN_FILE"
        
    else
        print_error "Terraform deployment failed"
        exit 1
    fi
}

# Function to destroy infrastructure
terraform_destroy() {
    print_warning "DESTRUCTIVE OPERATION: This will destroy all infrastructure!"
    
    if [[ "$ENVIRONMENT" == "production" ]]; then
        print_error "Destroying production infrastructure is not allowed via this script."
        print_info "If you really need to destroy production, do it manually with extreme caution."
        exit 1
    fi
    
    echo
    print_warning "This will destroy the following environment: $ENVIRONMENT"
    read -p "Type 'destroy' to confirm: " -r
    if [[ $REPLY != "destroy" ]]; then
        print_info "Destruction cancelled."
        exit 0
    fi
    
    if [[ "$AUTO_APPROVE" == "true" ]]; then
        APPROVE_FLAG="-auto-approve"
    else
        APPROVE_FLAG=""
    fi
    
    print_info "Destroying infrastructure..."
    
    if terraform destroy $APPROVE_FLAG; then
        print_success "Infrastructure destroyed successfully"
    else
        print_error "Infrastructure destruction failed"
        exit 1
    fi
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [ENVIRONMENT] [ACTION] [AUTO_APPROVE]"
    echo
    echo "Arguments:"
    echo "  ENVIRONMENT    Environment to deploy (development|staging|production)"
    echo "                 Default: development"
    echo "  ACTION         Action to perform (plan|apply|destroy)"
    echo "                 Default: apply"
    echo "  AUTO_APPROVE   Skip interactive approval (true|false)"
    echo "                 Default: false"
    echo
    echo "Examples:"
    echo "  $0                                    # Deploy to development"
    echo "  $0 staging                           # Deploy to staging"
    echo "  $0 production apply false           # Deploy to production with confirmation"
    echo "  $0 development plan                 # Plan deployment to development"
    echo "  $0 development destroy              # Destroy development infrastructure"
    echo
}

# Main execution
main() {
    print_info "Hub Infrastructure Deployment"
    print_info "=============================="
    
    # Parse arguments
    if [[ "$1" == "-h" || "$1" == "--help" ]]; then
        show_usage
        exit 0
    fi
    
    # Validate action
    if [[ ! "$ACTION" =~ ^(plan|apply|destroy)$ ]]; then
        print_error "Invalid action: $ACTION"
        print_info "Valid actions: plan, apply, destroy"
        exit 1
    fi
    
    # Check prerequisites
    check_prerequisites
    check_azure_login
    validate_environment
    setup_backend
    show_plan
    
    # Execute Terraform commands
    terraform_init
    terraform_validate
    
    case "$ACTION" in
        "plan")
            terraform_plan
            print_success "Plan completed. Review the output above."
            ;;
        "apply")
            terraform_plan
            terraform_apply
            ;;
        "destroy")
            terraform_destroy
            ;;
    esac
    
    print_success "Script execution completed!"
}

# Run main function
main "$@"