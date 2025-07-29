resource "azurerm_resource_group" "main" {
  name     = var.name
  location = var.location

  tags = var.tags

  dynamic "lifecycle" {
    for_each = var.prevent_destroy ? [1] : []
    content {
      prevent_destroy = true
    }
  }
}
