#!/bin/bash

# Hub Infrastructure Destruction Script
# This script safely destroys Hub infrastructure with proper safeguards

set -e

# Default values
ENVIRONMENT="${1:-}"
CONFIRMATION="${2:-}"

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

# Function to show usage
show_usage() {
    echo "Usage: $0 [ENVIRONMENT] [CONFIRMATION]"
    echo
    echo "Arguments:"
    echo "  ENVIRONMENT    Environment to destroy (development|staging)"
    echo "  CONFIRMATION   Type 'DESTROY' to confirm (case-sensitive)"
    echo
    echo "Examples:"
    echo "  $0 development DESTROY    # Destroy development environment"
    echo "  $0 staging DESTROY        # Destroy staging environment"
    echo
    echo "Note: Production environment cannot be destroyed using this script."
    echo "      This is a safety measure to prevent accidental destruction."
    echo
}

# Function to validate inputs
validate_inputs() {
    if [[ -z "$ENVIRONMENT" ]]; then
        print_error "Environment is required"
        show_usage
        exit 1
    fi
    
    if [[ ! "$ENVIRONMENT" =~ ^(development|staging)$ ]]; then
        print_error "Invalid environment: $ENVIRONMENT"
        print_info "Only development and staging environments can be destroyed using this script."
        print_info "Valid environments: development, staging"
        exit 1
    fi
    
    if [[ "$ENVIRONMENT" == "production" ]]; then
        print_error "Production environment cannot be destroyed using this script!"
        print_info "This is a safety measure. If you need to destroy production:"
        print_info "1. Use the Azure Portal for manual deletion"
        print_info "2. Or run 'terraform destroy' manually in the production directory"
        print_info "3. Ensure you have proper approvals and backups"
        exit 1
    fi
    
    if [[ -z "$CONFIRMATION" ]]; then
        print_error "Confirmation is required"
        show_usage
        exit 1
    fi
    
    if [[ "$CONFIRMATION" != "DESTROY" ]]; then
        print_error "Invalid confirmation. Must be exactly 'DESTROY' (case-sensitive)"
        show_usage
        exit 1
    fi
    
    if [[ ! -d "environments/$ENVIRONMENT" ]]; then
        print_error "Environment directory not found: environments/$ENVIRONMENT"
        exit 1
    fi
}

# Function to check prerequisites
check_prerequisites() {
    print_info "Checking prerequisites..."
    
    if ! command -v terraform &> /dev/null; then
        print_error "Terraform is not installed"
        exit 1
    fi
    
    if ! command -v az &> /dev/null; then
        print_error "Azure CLI is not installed"
        exit 1
    fi
    
    if ! az account show &> /dev/null; then
        print_error "Not logged in to Azure. Please run 'az login' first."
        exit 1
    fi
    
    print_success "Prerequisites check passed"
}

# Function to show what will be destroyed
show_destruction_plan() {
    print_warning "DESTRUCTIVE OPERATION WARNING"
    print_warning "============================="
    echo
    print_warning "This script will PERMANENTLY DESTROY the following:"
    echo "  • Environment: $ENVIRONMENT"
    echo "  • All AKS clusters and workloads"
    echo "  • All databases and data"
    echo "  • All storage accounts and files"
    echo "  • All Key Vaults and secrets"
    echo "  • All networking components"
    echo "  • All monitoring and logging data"
    echo
    print_error "THIS ACTION CANNOT BE UNDONE!"
    echo
    
    # Show current resources (if terraform state exists)
    cd "environments/$ENVIRONMENT"
    if [[ -f ".terraform/terraform.tfstate" ]] || terraform state list &> /dev/null; then
        print_info "Current resources that will be destroyed:"
        terraform state list 2>/dev/null | head -20
        
        local resource_count=$(terraform state list 2>/dev/null | wc -l)
        if [[ $resource_count -gt 20 ]]; then
            echo "  ... and $((resource_count - 20)) more resources"
        fi
        echo
    fi
}

# Function to get final confirmation
get_final_confirmation() {
    print_warning "Final Confirmation Required"
    print_warning "=========================="
    echo
    print_info "You are about to destroy the '$ENVIRONMENT' environment."
    print_info "Please confirm the following information:"
    echo
    
    # Show Azure subscription info
    local subscription_info=$(az account show --query '{name:name, id:id}' -o tsv)
    print_info "Azure Subscription: $subscription_info"
    
    # Show current directory
    print_info "Working Directory: $(pwd)"
    
    echo
    print_warning "Type the environment name to confirm: $ENVIRONMENT"
    read -p "Confirmation: " -r env_confirm
    
    if [[ "$env_confirm" != "$ENVIRONMENT" ]]; then
        print_info "Environment name mismatch. Destruction cancelled."
        exit 0
    fi
    
    echo
    print_warning "Type 'YES I UNDERSTAND' to proceed with destruction:"
    read -p "Final confirmation: " -r final_confirm
    
    if [[ "$final_confirm" != "YES I UNDERSTAND" ]]; then
        print_info "Final confirmation not provided. Destruction cancelled."
        exit 0
    fi
}

# Function to create backup of terraform state
backup_state() {
    print_info "Creating backup of Terraform state..."
    
    local backup_dir="../../../backups/terraform-state"
    local timestamp=$(date +%Y%m%d-%H%M%S)
    local backup_file="$backup_dir/${ENVIRONMENT}-pre-destroy-${timestamp}.backup"
    
    mkdir -p "$backup_dir"
    
    if terraform state pull > "$backup_file" 2>/dev/null; then
        print_success "State backup created: $backup_file"
    else
        print_warning "Could not create state backup (state may not exist)"
    fi
}

# Function to destroy infrastructure
destroy_infrastructure() {
    print_info "Initializing Terraform..."
    if ! terraform init; then
        print_error "Terraform initialization failed"
        exit 1
    fi
    
    print_info "Starting infrastructure destruction..."
    echo
    
    # Use a longer timeout for destroy operations
    export TF_CLI_ARGS_destroy="-parallelism=3"
    
    if terraform destroy -auto-approve; then
        print_success "Infrastructure destroyed successfully!"
    else
        print_error "Infrastructure destruction failed!"
        print_info "Some resources may still exist. Please check the Azure portal."
        print_info "You may need to manually delete resources or run terraform destroy again."
        exit 1
    fi
}

# Function to clean up local files
cleanup_local_files() {
    print_info "Cleaning up local Terraform files..."
    
    # Remove terraform files but keep the configuration
    rm -rf .terraform/
    rm -f .terraform.lock.hcl
    rm -f terraform.tfstate*
    rm -f tfplan*
    
    print_success "Local cleanup completed"
}

# Function to show post-destruction summary
show_summary() {
    echo
    print_success "Destruction Summary"
    print_success "=================="
    echo
    print_info "Environment '$ENVIRONMENT' has been destroyed."
    print_info "All Azure resources have been removed."
    print_info "Local Terraform state files have been cleaned up."
    echo
    print_warning "Important Notes:"
    echo "  • Verify resource deletion in the Azure portal"
    echo "  • Check for any orphaned resources that may incur costs"
    echo "  • Update any external systems that referenced this environment"
    echo "  • Inform team members about the environment destruction"
    echo
    
    if [[ -f "../../../backups/terraform-state/${ENVIRONMENT}-pre-destroy-"* ]]; then
        print_info "State backup is available in case you need to reference the previous configuration."
    fi
}

# Main execution
main() {
    print_info "Hub Infrastructure Destruction Script"
    print_info "===================================="
    echo
    
    # Parse arguments and show usage if needed
    if [[ "$1" == "-h" || "$1" == "--help" ]]; then
        show_usage
        exit 0
    fi
    
    # Validate inputs
    validate_inputs
    
    # Check prerequisites
    check_prerequisites
    
    # Show what will be destroyed
    show_destruction_plan
    
    # Get final confirmation
    get_final_confirmation
    
    # Change to environment directory
    cd "environments/$ENVIRONMENT"
    
    # Create backup
    backup_state
    
    # Destroy infrastructure
    destroy_infrastructure
    
    # Clean up local files
    cleanup_local_files
    
    # Show summary
    show_summary
    
    print_success "Destruction script completed successfully!"
}

# Trap to ensure we're in the right directory
trap 'cd "$(dirname "$0")/.."' EXIT

# Change to terraform root directory
cd "$(dirname "$0")/.."

# Run main function
main "$@"