resource "azurerm_resource_group" "main" {
  name     = var.name
  location = var.location

  tags = var.tags

  # Note: Lifecycle protection is handled per environment
  # Development allows recreation, production should be protected
}
