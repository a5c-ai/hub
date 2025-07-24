#!/bin/bash

# Hub Infrastructure Validation Script
# This script validates Terraform configurations and performs syntax checks

set -e

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

# Counters for validation results
TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0
WARNING_CHECKS=0

# Function to increment counters
increment_total() {
    ((TOTAL_CHECKS++))
}

increment_passed() {
    ((PASSED_CHECKS++))
    increment_total
}

increment_failed() {
    ((FAILED_CHECKS++))
    increment_total
}

increment_warning() {
    ((WARNING_CHECKS++))
    increment_total
}

# Function to check if required tools are installed
check_prerequisites() {
    print_info "Checking prerequisites..."
    
    local tools=("terraform" "tflint" "checkov")
    local missing_tools=()
    
    for tool in "${tools[@]}"; do
        if ! command -v "$tool" &> /dev/null; then
            missing_tools+=("$tool")
        fi
    done
    
    if [[ ${#missing_tools[@]} -eq 0 ]]; then
        print_success "All required tools are installed"
        increment_passed
    else
        print_warning "Some optional tools are missing: ${missing_tools[*]}"
        print_info "Install missing tools for enhanced validation:"
        for tool in "${missing_tools[@]}"; do
            case $tool in
                "tflint")
                    echo "  - TFLint: https://github.com/terraform-linters/tflint"
                    ;;
                "checkov")
                    echo "  - Checkov: pip install checkov"
                    ;;
            esac
        done
        increment_warning
    fi
}

# Function to validate Terraform syntax
validate_terraform_syntax() {
    print_info "Validating Terraform syntax..."
    
    local failed_modules=()
    
    # Validate each module
    for module_dir in modules/*/; do
        if [[ -d "$module_dir" ]]; then
            module_name=$(basename "$module_dir")
            print_info "Validating module: $module_name"
            
            cd "$module_dir"
            
            if terraform init -backend=false &> /dev/null && terraform validate &> /dev/null; then
                print_success "Module $module_name: Syntax valid"
            else
                print_error "Module $module_name: Syntax validation failed"
                failed_modules+=("$module_name")
            fi
            
            cd - > /dev/null
        fi
    done
    
    # Validate each environment
    for env_dir in environments/*/; do
        if [[ -d "$env_dir" ]]; then
            env_name=$(basename "$env_dir")
            print_info "Validating environment: $env_name"
            
            cd "$env_dir"
            
            if terraform init -backend=false &> /dev/null && terraform validate &> /dev/null; then
                print_success "Environment $env_name: Syntax valid"
            else
                print_error "Environment $env_name: Syntax validation failed"
                failed_modules+=("$env_name")
            fi
            
            cd - > /dev/null
        fi
    done
    
    if [[ ${#failed_modules[@]} -eq 0 ]]; then
        print_success "All Terraform configurations have valid syntax"
        increment_passed
    else
        print_error "Syntax validation failed for: ${failed_modules[*]}"
        increment_failed
    fi
}

# Function to run TFLint if available
run_tflint() {
    if ! command -v tflint &> /dev/null; then
        print_warning "TFLint not available, skipping linting checks"
        increment_warning
        return
    fi
    
    print_info "Running TFLint checks..."
    
    local failed_linting=()
    
    # Initialize TFLint
    if ! tflint --init &> /dev/null; then
        print_warning "TFLint initialization failed"
        increment_warning
        return
    fi
    
    # Lint each module
    for module_dir in modules/*/; do
        if [[ -d "$module_dir" ]]; then
            module_name=$(basename "$module_dir")
            print_info "Linting module: $module_name"
            
            if tflint "$module_dir" &> /dev/null; then
                print_success "Module $module_name: Linting passed"
            else
                print_warning "Module $module_name: Linting issues found"
                failed_linting+=("$module_name")
            fi
        fi
    done
    
    # Lint each environment
    for env_dir in environments/*/; do
        if [[ -d "$env_dir" ]]; then
            env_name=$(basename "$env_dir")
            print_info "Linting environment: $env_name"
            
            if tflint "$env_dir" &> /dev/null; then
                print_success "Environment $env_name: Linting passed"
            else
                print_warning "Environment $env_name: Linting issues found"
                failed_linting+=("$env_name")
            fi
        fi
    done
    
    if [[ ${#failed_linting[@]} -eq 0 ]]; then
        print_success "All configurations passed TFLint checks"
        increment_passed
    else
        print_warning "TFLint found issues in: ${failed_linting[*]}"
        increment_warning
    fi
}

# Function to run security checks with Checkov
run_security_checks() {
    if ! command -v checkov &> /dev/null; then
        print_warning "Checkov not available, skipping security checks"
        increment_warning
        return
    fi
    
    print_info "Running security checks with Checkov..."
    
    # Run Checkov on all Terraform files
    local checkov_output
    checkov_output=$(checkov -d . --framework terraform --quiet --compact 2>/dev/null || true)
    
    if [[ -z "$checkov_output" ]]; then
        print_success "No security issues found by Checkov"
        increment_passed
    else
        local failed_count=$(echo "$checkov_output" | grep -c "FAILED" || true)
        local passed_count=$(echo "$checkov_output" | grep -c "PASSED" || true)
        
        if [[ $failed_count -gt 0 ]]; then
            print_warning "Checkov found $failed_count security issues and $passed_count passed checks"
            print_info "Run 'checkov -d . --framework terraform' for detailed output"
            increment_warning
        else
            print_success "All security checks passed ($passed_count checks)"
            increment_passed
        fi
    fi
}

# Function to validate file structure
validate_file_structure() {
    print_info "Validating file structure..."
    
    local missing_files=()
    local required_files=(
        "modules/aks/main.tf"
        "modules/aks/variables.tf"
        "modules/aks/outputs.tf"
        "modules/postgresql/main.tf"
        "modules/postgresql/variables.tf"
        "modules/postgresql/outputs.tf"
        "modules/storage/main.tf"
        "modules/storage/variables.tf"
        "modules/storage/outputs.tf"
        "modules/keyvault/main.tf"
        "modules/keyvault/variables.tf"
        "modules/keyvault/outputs.tf"
        "modules/networking/main.tf"
        "modules/networking/variables.tf"
        "modules/networking/outputs.tf"
        "modules/monitoring/main.tf"
        "modules/monitoring/variables.tf"
        "modules/monitoring/outputs.tf"
        "modules/security/main.tf"
        "modules/security/variables.tf"
        "modules/security/outputs.tf"
        "environments/development/main.tf"
        "environments/development/variables.tf"
        "environments/development/outputs.tf"
        "environments/staging/main.tf"
        "environments/staging/variables.tf"
        "environments/staging/outputs.tf"
        "environments/production/main.tf"
        "environments/production/variables.tf"
        "environments/production/outputs.tf"
    )
    
    for file in "${required_files[@]}"; do
        if [[ ! -f "$file" ]]; then
            missing_files+=("$file")
        fi
    done
    
    if [[ ${#missing_files[@]} -eq 0 ]]; then
        print_success "All required files are present"
        increment_passed
    else
        print_error "Missing required files:"
        for file in "${missing_files[@]}"; do
            echo "  - $file"
        done
        increment_failed
    fi
}

# Function to validate variable files
validate_variable_files() {
    print_info "Validating variable configurations..."
    
    local issues=()
    
    # Check for example variable files
    for env_dir in environments/*/; do
        if [[ -d "$env_dir" ]]; then
            env_name=$(basename "$env_dir")
            example_file="$env_dir/terraform.tfvars.example"
            
            if [[ ! -f "$example_file" ]]; then
                issues+=("Missing example file: $example_file")
            else
                print_success "Environment $env_name: Example variables file exists"
            fi
        fi
    done
    
    if [[ ${#issues[@]} -eq 0 ]]; then
        print_success "All variable configurations are valid"
        increment_passed
    else
        print_warning "Variable configuration issues:"
        for issue in "${issues[@]}"; do
            echo "  - $issue"
        done
        increment_warning
    fi
}

# Function to check Terraform version compatibility
check_terraform_version() {
    print_info "Checking Terraform version compatibility..."
    
    local tf_version
    tf_version=$(terraform version -json | jq -r '.terraform_version' 2>/dev/null || terraform version | head -1 | cut -d' ' -f2 | sed 's/v//')
    
    local required_version="1.5.0"
    
    if command -v dpkg &> /dev/null; then
        # Use dpkg for version comparison on Debian/Ubuntu
        if dpkg --compare-versions "$tf_version" "ge" "$required_version"; then
            print_success "Terraform version $tf_version meets requirements (>= $required_version)"
            increment_passed
        else
            print_error "Terraform version $tf_version is too old (required: >= $required_version)"
            increment_failed
        fi
    else
        # Fallback to string comparison
        if [[ "$tf_version" == "$required_version" ]]; then
            print_success "Terraform version $tf_version meets requirements"
            increment_passed
        else
            print_warning "Cannot verify Terraform version compatibility: $tf_version"
            increment_warning
        fi
    fi
}

# Function to validate provider versions
validate_provider_versions() {
    print_info "Validating provider version constraints..."
    
    local version_issues=()
    
    # Check if all modules have versions.tf
    for module_dir in modules/*/; do
        if [[ -d "$module_dir" ]]; then
            module_name=$(basename "$module_dir")
            versions_file="$module_dir/versions.tf"
            
            if [[ ! -f "$versions_file" ]]; then
                version_issues+=("Module $module_name: Missing versions.tf")
            elif ! grep -q "required_providers" "$versions_file"; then
                version_issues+=("Module $module_name: No required_providers block")
            elif ! grep -q "azurerm" "$versions_file"; then
                version_issues+=("Module $module_name: Missing azurerm provider constraint")
            fi
        fi
    done
    
    if [[ ${#version_issues[@]} -eq 0 ]]; then
        print_success "All provider version constraints are properly defined"
        increment_passed
    else
        print_warning "Provider version issues:"
        for issue in "${version_issues[@]}"; do
            echo "  - $issue"
        done
        increment_warning
    fi
}

# Function to check for common Terraform best practices
check_best_practices() {
    print_info "Checking Terraform best practices..."
    
    local practice_issues=()
    
    # Check for resource naming conventions
    local naming_violations=0
    while IFS= read -r -d '' file; do
        if grep -q 'resource "azurerm_' "$file"; then
            # Check for hardcoded resource names (basic check)
            if grep -q 'name.*=.*"[^$]' "$file" && ! grep -q 'var\.' "$file"; then
                ((naming_violations++))
            fi
        fi
    done < <(find . -name "*.tf" -print0)
    
    if [[ $naming_violations -gt 0 ]]; then
        practice_issues+=("Found $naming_violations potential hardcoded resource names")
    fi
    
    # Check for missing tags
    local files_without_tags=0
    while IFS= read -r -d '' file; do
        if grep -q 'resource "azurerm_' "$file" && ! grep -q 'tags.*=' "$file"; then
            ((files_without_tags++))
        fi
    done < <(find . -name "*.tf" -print0)
    
    if [[ $files_without_tags -gt 0 ]]; then
        practice_issues+=("Found $files_without_tags files potentially missing tags")
    fi
    
    if [[ ${#practice_issues[@]} -eq 0 ]]; then
        print_success "No obvious best practice violations found"
        increment_passed
    else
        print_warning "Best practice recommendations:"
        for issue in "${practice_issues[@]}"; do
            echo "  - $issue"
        done
        increment_warning
    fi
}

# Function to show validation summary
show_validation_summary() {
    echo
    print_info "Validation Summary"
    print_info "=================="
    echo "  Total checks: $TOTAL_CHECKS"
    echo "  Passed: $PASSED_CHECKS"
    echo "  Warnings: $WARNING_CHECKS"
    echo "  Failed: $FAILED_CHECKS"
    echo
    
    if [[ $FAILED_CHECKS -eq 0 ]]; then
        if [[ $WARNING_CHECKS -eq 0 ]]; then
            print_success "All validations passed! ✅"
            echo "Your Terraform configuration is ready for deployment."
        else
            print_warning "Validations passed with warnings ⚠️"
            echo "Address warnings for optimal configuration."
        fi
    else
        print_error "Some validations failed ❌"
        echo "Please fix the failed checks before deploying."
        exit 1
    fi
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo
    echo "Options:"
    echo "  -h, --help     Show this help message"
    echo "  --no-security  Skip security checks (faster)"
    echo "  --no-lint      Skip linting checks"
    echo
    echo "This script validates Terraform configurations for the Hub infrastructure."
    echo "It performs syntax validation, linting, security checks, and best practice verification."
}

# Main execution
main() {
    local skip_security=false
    local skip_lint=false
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_usage
                exit 0
                ;;
            --no-security)
                skip_security=true
                shift
                ;;
            --no-lint)
                skip_lint=true
                shift
                ;;
            *)
                print_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
    
    print_info "Hub Infrastructure Validation"
    print_info "============================="
    echo
    
    # Run validation checks
    check_prerequisites
    check_terraform_version
    validate_file_structure
    validate_terraform_syntax
    validate_provider_versions
    validate_variable_files
    
    if [[ "$skip_lint" != true ]]; then
        run_tflint
    else
        print_info "Skipping linting checks"
    fi
    
    if [[ "$skip_security" != true ]]; then
        run_security_checks
    else
        print_info "Skipping security checks"
    fi
    
    check_best_practices
    
    # Show summary
    show_validation_summary
}

# Trap to ensure we're in the right directory
trap 'cd "$(dirname "$0")/.."' EXIT

# Change to terraform root directory
cd "$(dirname "$0")/.."

# Run main function
main "$@"