 resource "helm_release" "controller" {
  name             = "actions-runner-controller"
  repository       = "https://actions-runner-controller.github.io/actions-runner-controller"
  chart            = "actions-runner-controller"
  version          = var.chart_version
  namespace        = var.namespace
  create_namespace = true
  install_crds     = true

  set {
    name  = "authSecret.github_token"
    value = var.github_token
  }
}

# Wait for CRDs to be installed and ready
resource "time_sleep" "wait_for_crds" {
  depends_on = [helm_release.controller]
  create_duration = "30s"
}

resource "kubernetes_manifest" "runner_deployment" {
  manifest = {
    apiVersion = "actions.summerwind.dev/v1alpha1"
    kind       = "RunnerDeployment"
    metadata = {
      name      = var.runner_deployment_name
      namespace = var.namespace
    }
    spec = {
      replicas = var.runner_replicas
      template = {
        spec = {
          repository = "${var.github_owner}/${var.github_repository}"
          labels     = var.runner_labels
        }
      }
    }
  }
  depends_on = [time_sleep.wait_for_crds]
}
