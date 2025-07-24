#!/bin/bash

# Hub Infrastructure Planning Script
# This script creates and analyzes Terraform plans for different environments

set -e

# Default values
ENVIRONMENT="${1:-development}"
OUTPUT_FORMAT="${2:-human}"

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
    echo "Usage: $0 [ENVIRONMENT] [OUTPUT_FORMAT]"
    echo
    echo "Arguments:"
    echo "  ENVIRONMENT     Environment to plan (development|staging|production)"
    echo "                  Default: development"
    echo "  OUTPUT_FORMAT   Output format (human|json)"
    echo "                  Default: human"
    echo
    echo "Examples:"
    echo "  $0                        # Plan development environment"
    echo "  $0 staging               # Plan staging environment"
    echo "  $0 production json       # Plan production with JSON output"
    echo
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

# Function to validate output format
validate_output_format() {
    if [[ ! "$OUTPUT_FORMAT" =~ ^(human|json)$ ]]; then
        print_error "Invalid output format: $OUTPUT_FORMAT"
        print_info "Valid formats: human, json"
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

# Function to initialize Terraform
terraform_init() {
    print_info "Initializing Terraform..."
    cd "environments/$ENVIRONMENT"
    
    # Check if backend configuration exists
    BACKEND_CONFIG=""
    if [[ -f "backend.conf" ]]; then
        BACKEND_CONFIG="-backend-config=backend.conf"
        print_info "Using backend configuration: backend.conf"
    fi
    
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

# Function to create and analyze plan
create_plan() {
    local timestamp=$(date +%Y%m%d-%H%M%S)
    local plan_file="tfplan-${timestamp}"
    local plan_output_file="tfplan-${timestamp}.txt"
    local plan_json_file="tfplan-${timestamp}.json"
    
    print_info "Creating Terraform plan..."
    
    # Create the plan
    if terraform plan -out="$plan_file" -detailed-exitcode; then
        local exit_code=$?
        case $exit_code in
            0)
                print_success "No changes needed - infrastructure is up-to-date"
                ;;
            2)
                print_success "Plan created successfully with changes: $plan_file"
                ;;
        esac
    else
        local exit_code=$?
        case $exit_code in
            1)
                print_error "Terraform plan failed due to an error"
                exit 1
                ;;
        esac
    fi
    
    # Generate human-readable output
    if [[ "$OUTPUT_FORMAT" == "human" || "$OUTPUT_FORMAT" == "both" ]]; then
        print_info "Generating human-readable plan output..."
        terraform show -no-color "$plan_file" > "$plan_output_file"
        
        # Display plan summary
        echo
        print_info "Plan Summary:"
        echo "============="
        terraform show "$plan_file" | grep -E "Plan:|No changes\." || true
        
        # Show resource changes summary
        echo
        print_info "Resource Changes:"
        echo "=================="
        terraform show "$plan_file" | grep -E "^\s*[#~+-]" | head -20 || true
        
        local total_changes=$(terraform show "$plan_file" | grep -E "^\s*[#~+-]" | wc -l)
        if [[ $total_changes -gt 20 ]]; then
            echo "  ... and $((total_changes - 20)) more changes"
        fi
        
        print_success "Human-readable output saved to: $plan_output_file"
    fi
    
    # Generate JSON output
    if [[ "$OUTPUT_FORMAT" == "json" || "$OUTPUT_FORMAT" == "both" ]]; then
        print_info "Generating JSON plan output..."
        terraform show -json "$plan_file" > "$plan_json_file"
        print_success "JSON output saved to: $plan_json_file"
        
        # Show JSON summary if not in human mode
        if [[ "$OUTPUT_FORMAT" == "json" ]]; then
            echo
            print_info "Plan Summary (from JSON):"
            echo "========================="
            local changes=$(jq -r '.planned_changes | length' "$plan_json_file")
            local resource_changes=$(jq -r '.resource_changes | length' "$plan_json_file")
            echo "  Total planned changes: $changes"
            echo "  Resource changes: $resource_changes"
            
            # Show resource change breakdown
            local to_add=$(jq -r '.resource_changes[] | select(.change.actions[] == "create") | .address' "$plan_json_file" | wc -l)
            local to_change=$(jq -r '.resource_changes[] | select(.change.actions[] == "update") | .address' "$plan_json_file" | wc -l)
            local to_destroy=$(jq -r '.resource_changes[] | select(.change.actions[] == "delete") | .address' "$plan_json_file" | wc -l)
            
            echo "  Resources to add: $to_add"
            echo "  Resources to change: $to_change"
            echo "  Resources to destroy: $to_destroy"
        fi
    fi
    
    # Show cost estimation reminder
    echo
    print_info "Cost Estimation:"
    echo "================"
    print_warning "Remember to review the cost implications of these changes."
    print_info "Consider using 'az consumption' commands or third-party tools like Infracost."
    
    # Clean up plan file (keep outputs)
    rm -f "$plan_file"
    
    echo
    print_success "Plan analysis completed!"
    print_info "Plan files saved in: $(pwd)"
}

# Function to show environment info
show_environment_info() {
    echo
    print_info "Environment Information:"
    echo "========================"
    echo "  Environment: $ENVIRONMENT"
    echo "  Working Directory: environments/$ENVIRONMENT"
    echo "  Output Format: $OUTPUT_FORMAT"
    
    # Show Azure subscription info
    local subscription_info=$(az account show --query '{name:name, id:id}' -o tsv)
    echo "  Azure Subscription: $subscription_info"
    
    # Show Terraform version
    local tf_version=$(terraform version -json | jq -r '.terraform_version')
    echo "  Terraform Version: $tf_version"
    
    echo
}

# Function to show configuration summary
show_configuration_summary() {
    print_info "Configuration Summary:"
    echo "======================"
    
    # Count terraform files
    local tf_files=$(find . -name "*.tf" | wc -l)
    echo "  Terraform files: $tf_files"
    
    # Show variable files
    if [[ -f "terraform.tfvars" ]]; then
        echo "  Variables file: terraform.tfvars (exists)"
    else
        echo "  Variables file: terraform.tfvars (not found - using defaults)"
        if [[ -f "terraform.tfvars.example" ]]; then
            print_info "  Example file available: terraform.tfvars.example"
        fi
    fi
    
    # Show backend configuration
    if [[ -f "backend.conf" ]]; then
        echo "  Backend config: backend.conf (exists)"
    else
        echo "  Backend config: backend.conf (not found - using default)"
    fi
    
    echo
}

# Main execution
main() {
    print_info "Hub Infrastructure Planning"
    print_info "=========================="
    
    # Parse arguments
    if [[ "$1" == "-h" || "$1" == "--help" ]]; then
        show_usage
        exit 0
    fi
    
    # Validate inputs
    validate_environment
    validate_output_format
    
    # Check prerequisites
    check_prerequisites
    
    # Show environment information
    show_environment_info
    
    # Initialize and validate
    terraform_init
    terraform_validate
    
    # Show configuration summary
    show_configuration_summary
    
    # Create and analyze plan
    create_plan
    
    echo
    print_success "Planning completed successfully!"
    print_info "Review the plan output above and the generated files."
    print_info "Run './deploy.sh $ENVIRONMENT apply' to apply the changes."
}

# Trap to ensure we're in the right directory
trap 'cd "$(dirname "$0")/.."' EXIT

# Change to terraform root directory
cd "$(dirname "$0")/.."

# Run main function
main "$@"