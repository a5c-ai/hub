resource "helm_release" "controller" {
  count            = var.enabled ? 1 : 0
  name             = var.release_name
  repository       = var.chart_repository
  chart            = var.chart_name
  version          = var.chart_version
  namespace        = var.namespace
  create_namespace = var.create_namespace

  dynamic "set" {
    for_each = var.values
    content {
      name  = set.key
      value = set.value
    }
  }

  timeout = var.timeout
}

resource "time_sleep" "wait_for_controller" {
  count           = var.enabled ? 1 : 0
  depends_on      = [helm_release.controller]
  create_duration = var.wait_duration
}
