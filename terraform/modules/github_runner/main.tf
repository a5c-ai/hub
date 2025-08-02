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

# Create the runner scale set
resource "helm_release" "arc_runner_set" {
  name             = var.runner_scale_set_name
  repository       = "oci://ghcr.io/actions/actions-runner-controller-charts"
  chart            = "gha-runner-scale-set"
  version          = var.runner_set_chart_version
  namespace        = var.runners_namespace
  create_namespace = false
  timeout          = 600

  values = [
    yamlencode({
      githubConfigUrl    = var.github_config_url
      githubConfigSecret = kubernetes_secret.github_secret.metadata[0].name
      
      runnerGroup = var.runner_group
      runnerScaleSetName = var.runner_scale_set_name
      
      minRunners = var.min_runners
      maxRunners = var.max_runners
      
      containerMode = {
        type = var.container_mode
      }
      
      template = {
        spec = {
          containers = [{
            name  = "runner"
            image = var.runner_image != null ? var.runner_image : "ghcr.io/actions/actions-runner:latest"
          }]
        }
      }
    })
  ]

  depends_on = [
    kubernetes_secret.github_secret,
    time_sleep.wait_for_controller
  ]
}
