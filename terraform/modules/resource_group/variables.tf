variable "name" {
  description = "The name of the resource group"
  type        = string
}

variable "location" {
  description = "The Azure location where the resource group should be created"
  type        = string
}

variable "tags" {
  description = "A mapping of tags to assign to the resource"
  type        = map(string)
  default     = {}
}

variable "prevent_destroy" {
  description = "Whether to prevent the destruction of the resource group"
  type        = bool
  default     = true
}
