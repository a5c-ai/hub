 resource "helm_release" "controller" {
   name             = "actions-runner-controller"
   repository       = "https://actions-runner-controller.github.io/actions-runner-controller"
   chart            = "actions-runner-controller"
   version          = var.chart_version
   namespace        = var.namespace
   create_namespace = true

   set {
     name  = "controller.github.token"
     value = var.github_token
   }
   set {
     name  = "controller.github.repository"
     value = "${var.github_owner}/${var.github_repository}"
   }
   set {
     name  = "controller.runner.replicas"
     value = var.runner_replicas
   }

   dynamic "set" {
     for_each = var.runner_labels
     content {
       name  = "controller.runner.labels.${each.key}"
       value = each.value
     }
   }
 }

resource "kubernetes_manifest" "runner_deployment" {
  # Skip schema validation to allow CRDs to be installed by Helm before applying this manifest
  skip_schema_validation = true
  # RunnerDeployment manifest for actions-runner-controller
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
   depends_on = [helm_release.controller]
 }
