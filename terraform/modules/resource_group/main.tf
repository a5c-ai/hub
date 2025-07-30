resource "azurerm_resource_group" "main" {
  name     = var.name
  location = var.location

  tags = var.tags

  # Note: lifecycle blocks cannot be dynamic in Terraform
  # Resource groups are foundational infrastructure and should be protected
  lifecycle {
    prevent_destroy = var.prevent_destroy
  }
}
