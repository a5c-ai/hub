# cert-manager Terraform module for AKS

# cert-manager Helm release (only if managing via Terraform)
resource "helm_release" "cert_manager" {
  count = var.manage_cert_manager ? 1 : 0
  name       = "cert-manager"
  repository = "https://charts.jetstack.io"
  chart      = "cert-manager"
  version    = var.cert_manager_version
  namespace        = "cert-manager"
  create_namespace = true

  set {
    name  = "installCRDs"
    value = "true"  # Let Helm install CRDs automatically
  }

  set {
    name  = "global.leaderElection.namespace"
    value = "cert-manager"
  }

  # Azure-specific configuration for Application Gateway
  set {
    name  = "nodeSelector.kubernetes\\.io/os"
    value = "linux"
  }

  # Additional security settings
  set {
    name  = "securityContext.runAsNonRoot"
    value = "true"
  }


  timeout = 300
}

# Wait for cert-manager to be ready before creating issuers
resource "time_sleep" "wait_for_cert_manager" {
  count           = var.manage_cert_manager ? 1 : 0
  depends_on      = [helm_release.cert_manager[0]]
  create_duration = "60s"
}

# Let's Encrypt Staging ClusterIssuer
resource "kubernetes_manifest" "letsencrypt_staging" {
count   = var.manage_cert_manager ? 1 : 0
manifest = {
    apiVersion = "cert-manager.io/v1"
    kind       = "ClusterIssuer"
    metadata = {
      name = var.staging_cluster_issuer_name
    }
    spec = {
      acme = {
        server = "https://acme-staging-v02.api.letsencrypt.org/directory"
        email  = var.email
        privateKeySecretRef = {
          name = "letsencrypt-staging"
        }
        solvers = [
          {
            http01 = {
              ingress = {
                class = "azure/application-gateway"
              }
            }
          }
        ]
      }
    }
  }

  depends_on = [time_sleep.wait_for_cert_manager]
}

# Let's Encrypt Production ClusterIssuer
resource "kubernetes_manifest" "letsencrypt_production" {
count   = var.manage_cert_manager ? 1 : 0
manifest = {
    apiVersion = "cert-manager.io/v1"
    kind       = "ClusterIssuer"
    metadata = {
      name = var.cluster_issuer_name
    }
    spec = {
      acme = {
        server = "https://acme-v02.api.letsencrypt.org/directory"
        email  = var.email
        privateKeySecretRef = {
          name = "letsencrypt-production"
        }
        solvers = [
          {
            http01 = {
              ingress = {
                class = "azure/application-gateway"
              }
            }
          }
        ]
      }
    }
  }

  depends_on = [time_sleep.wait_for_cert_manager]
}
