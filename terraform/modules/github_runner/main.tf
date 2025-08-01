 resource "helm_release" "controller" {
  name             = "actions-runner-controller"
  repository       = "https://actions-runner-controller.github.io/actions-runner-controller"
  chart            = "actions-runner-controller"
  version          = var.chart_version
  namespace        = var.namespace
  create_namespace = true


  set {
    name  = "authSecret.github_token"
    value = var.github_token
  }

  # Disable cert-manager webhook since we don't have cert-manager installed
  set {
    name  = "certManagerEnabled"
    value = "false"
  }

  set {
    name  = "admissionWebHooks.enabled"
    value = "false"
  }
}

# Wait for CRDs to be installed and ready
resource "time_sleep" "wait_for_crds" {
  depends_on = [helm_release.controller]
  create_duration = "30s"
}

# Use kubectl to apply the RunnerDeployment manifest
resource "null_resource" "runner_deployment" {
  triggers = {
    manifest_content = jsonencode({
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
    })
    runner_deployment_name = var.runner_deployment_name
    namespace = var.namespace
  }

  provisioner "local-exec" {
    command = <<-EOT
      cat <<EOF | kubectl apply -f -
      apiVersion: actions.summerwind.dev/v1alpha1
      kind: RunnerDeployment
      metadata:
        name: ${var.runner_deployment_name}
        namespace: ${var.namespace}
      spec:
        replicas: ${var.runner_replicas}
        template:
          spec:
            repository: ${var.github_owner}/${var.github_repository}
            labels: ${jsonencode(var.runner_labels)}
      EOF
    EOT
  }

  provisioner "local-exec" {
    when    = destroy
    command = "kubectl delete runnerdeployment ${self.triggers.runner_deployment_name} -n ${self.triggers.namespace} --ignore-not-found=true"
  }

  depends_on = [time_sleep.wait_for_crds]
}
