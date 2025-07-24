variable "cluster_name" {
  description = "The name of the AKS cluster"
  type        = string
}

variable "location" {
  description = "The Azure location where the AKS cluster should be created"
  type        = string
}

variable "resource_group_name" {
  description = "The name of the resource group"
  type        = string
}

variable "dns_prefix" {
  description = "DNS prefix for the cluster"
  type        = string
}

variable "kubernetes_version" {
  description = "The version of Kubernetes to use"
  type        = string
  default     = "1.28"
}

variable "node_count" {
  description = "The number of nodes in the default node pool"
  type        = number
  default     = 3
}

variable "vm_size" {
  description = "The size of the virtual machines in the default node pool"
  type        = string
  default     = "Standard_D2s_v5"
}

variable "availability_zones" {
  description = "The availability zones to use for the node pool"
  type        = list(string)
  default     = ["1", "2", "3"]
}

variable "enable_auto_scaling" {
  description = "Enable auto scaling for the node pool"
  type        = bool
  default     = true
}

variable "min_node_count" {
  description = "The minimum number of nodes in the default node pool"
  type        = number
  default     = 1
}

variable "max_node_count" {
  description = "The maximum number of nodes in the default node pool"
  type        = number
  default     = 10
}

variable "os_disk_size_gb" {
  description = "The size of the OS disk in GB"
  type        = number
  default     = 128
}

variable "subnet_id" {
  description = "The ID of the subnet where the cluster should be deployed"
  type        = string
}

variable "vnet_id" {
  description = "The ID of the virtual network"
  type        = string
}

variable "service_cidr" {
  description = "The CIDR block for Kubernetes services"
  type        = string
  default     = "10.1.0.0/16"
}

variable "dns_service_ip" {
  description = "The IP address for the DNS service"
  type        = string
  default     = "10.1.0.10"
}

variable "max_surge" {
  description = "The maximum number or percentage of nodes that can be created during an upgrade"
  type        = string
  default     = "10%"
}

variable "log_retention_days" {
  description = "The number of days to retain logs"
  type        = number
  default     = 30
}

variable "create_worker_node_pool" {
  description = "Create an additional worker node pool"
  type        = bool
  default     = false
}

variable "worker_vm_size" {
  description = "The size of the virtual machines in the worker node pool"
  type        = string
  default     = "Standard_D4s_v5"
}

variable "worker_node_count" {
  description = "The number of nodes in the worker node pool"
  type        = number
  default     = 2
}

variable "worker_min_node_count" {
  description = "The minimum number of nodes in the worker node pool"
  type        = number
  default     = 1
}

variable "worker_max_node_count" {
  description = "The maximum number of nodes in the worker node pool"
  type        = number
  default     = 5
}

variable "worker_node_taints" {
  description = "Taints to apply to worker nodes"
  type        = list(string)
  default     = []
}

variable "environment" {
  description = "The environment name"
  type        = string
  default     = "development"
}

variable "tags" {
  description = "A mapping of tags to assign to the resources"
  type        = map(string)
  default     = {}
}