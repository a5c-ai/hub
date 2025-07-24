resource "azurerm_resource_group" "main" {
  name     = var.name
  location = var.location

  tags = var.tags

  lifecycle {
    prevent_destroy = var.prevent_destroy
  }
}