 # Create namespaces
resource "kubernetes_namespace" "controller" {
  metadata {
    name = var.controller_namespace
  }
}

resource "kubernetes_namespace" "runners" {
  metadata {
    name = var.runners_namespace
  }
}

# Install the GitHub Actions Runner Controller
resource "helm_release" "arc_controller" {
  name             = "arc-controller"
  repository       = "oci://ghcr.io/actions/actions-runner-controller-charts"
  chart            = "gha-runner-scale-set-controller"
  version          = var.controller_chart_version
  namespace        = var.controller_namespace
  create_namespace = false
  timeout          = 600

  depends_on = [kubernetes_namespace.controller]
}

# Create a secret for GitHub authentication (App or Token)
resource "kubernetes_secret" "github_secret" {
  metadata {
    name      = "github-secret"
    namespace = var.runners_namespace
  }

  data = var.auth_method == "token" ? {
    github_token = var.github_token
  } : {
    github_app_id              = var.github_app_id
    github_app_installation_id = var.github_app_installation_id
    github_app_private_key     = var.github_app_private_key
  }

  type = "Opaque"
  
  depends_on = [kubernetes_namespace.runners]
}

# Wait for the controller to be ready
resource "time_sleep" "wait_for_controller" {
  depends_on = [helm_release.arc_controller]
  create_duration = "30s"
}

# Clean up existing problematic AutoscalingRunnerSet resource
resource "null_resource" "cleanup_existing_runner_set" {
  triggers = {
    # Trigger cleanup when runner set name changes
    runner_set_name = var.runner_scale_set_name
    namespace = var.runners_namespace
  }

  provisioner "local-exec" {
    command = <<-EOT
      # Delete existing AutoscalingRunnerSet if it exists
      kubectl delete autoscalingrunnersets.actions.github.com -n ${var.runners_namespace} ${var.runner_scale_set_name} --ignore-not-found=true --timeout=60s
      
      # Wait a moment for cleanup
      sleep 5
      
      # Also clean up any orphaned ephemeral runners
      kubectl delete ephemeralrunners.actions.github.com -n ${var.runners_namespace} --all --timeout=60s --ignore-not-found=true
    EOT
    
    on_failure = continue
  }

  depends_on = [
    kubernetes_namespace.runners,
    time_sleep.wait_for_controller
  ]
}

# Define volume spec configuration
locals {
  # Base volume spec without storage class
  base_volume_spec = {
    accessModes = ["ReadWriteOnce"]
    resources = {
      requests = {
        storage = var.ephemeral_storage_size
      }
    }
  }
  
  # Volume spec with optional storage class
  # Always merge base spec with optional storageClassName for consistent types
  volume_spec = merge(
    local.base_volume_spec,
    var.storage_class_name != "" ? { storageClassName = var.storage_class_name } : {}
  )
}

# Create the runner scale set
resource "helm_release" "arc_runner_set" {
  name             = var.runner_scale_set_name
  repository       = "oci://ghcr.io/actions/actions-runner-controller-charts"
  chart            = "gha-runner-scale-set"
  version          = var.runner_set_chart_version
  namespace        = var.runners_namespace
  create_namespace = false
  timeout          = 600
  
  # Add lifecycle rule to force recreation when needed
  lifecycle {
    create_before_destroy = false
  }
  
  # Temporarily disable atomic to get better error details
  # atomic           = true
  # force_update     = true

  values = [
    yamlencode({
      githubConfigUrl    = var.github_config_url
      githubConfigSecret = kubernetes_secret.github_secret.metadata[0].name
      
      runnerGroup = var.runner_group
      runnerScaleSetName    = var.runner_scale_set_name
      
      minRunners = var.min_runners
      maxRunners = var.max_runners
      
      # Container mode - use variable (dind or kubernetes)
      # Keep object shape consistent for Terraform conditional typing rules by
      # always including kubernetesModeWorkVolumeClaim and setting it to null when not in use.
      containerMode = {
        type = var.container_mode
        kubernetesModeWorkVolumeClaim = var.container_mode == "kubernetes" ? local.volume_spec : null
      }
      
      # Use custom runner image or init container overrides, merged into template map for consistent typing
      template = tomap({
        spec = merge(
          # { nodeSelector = var.runner_node_selector },
          var.runner_image != null ? {
            containers = [{
              name  = "runner"
              image = var.runner_image
            }]
          } : {},
        )
      })
    })
  ]

  depends_on = [
    kubernetes_secret.github_secret,
    time_sleep.wait_for_controller,
    null_resource.cleanup_existing_runner_set
  ]
}
