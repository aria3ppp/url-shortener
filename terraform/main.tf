terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "2.19.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "2.9.0"
    }
  }
}

provider "kubernetes" {
  config_path = "~/.kube/config"
  config_context = "minikube"
}

provider "helm" {
  kubernetes {
    config_path = "~/.kube/config"
    config_context = "minikube"
  }
}

# create kubernetes namespace
resource "kubernetes_namespace" "namespace" {
  metadata {
    name = var.kubernetes_namespace
  }
}

# create kubernetes postgres secret
resource "kubernetes_secret" "postgres_secret" {
  depends_on = [
    kubernetes_namespace.namespace,
  ]

  metadata {
    name = "postgres-secret"
    namespace = kubernetes_namespace.namespace.metadata.0.name
  }

  type = "Opaque"

  # TODO: values will be encoded base64 internally by `kubernetes_secret` resource???
  data = {
    (var.kubernetes_postgres_secret__password_key) = var.postgres_password
    (var.kubernetes_postgres_secret__postgres_password_key) = var.postgres_user_password
    (var.kubernetes_postgres_secret__repmgr_password_key) = var.postgres_repmgr_password
  }
}

# create kubernetes pgpool secret
resource "kubernetes_secret" "pgpool_secret" {
  depends_on = [
    kubernetes_namespace.namespace,
  ]

  metadata {
    name = "pgpool-secret"
    namespace = kubernetes_namespace.namespace.metadata.0.name
  }

  type = "Opaque"

  # TODO: values will be encoded base64 internally by `kubernetes_secret` resource???
  data = {
    (var.kubernetes_pgpool_secret__admin_password_key) = var.postgres_pgpool_admin_password
  }
}

# create kubernetes storage class
resource "kubernetes_storage_class" "storage_class" {
  metadata {
    name = "sc"
  }

  # hardcode minikube storage provisioner
  storage_provisioner = "k8s.io/minikube-hostpath"
  volume_binding_mode = "Immediate"
  reclaim_policy = "Delete"
}

# deploy postgres helm chart with replicas
resource "helm_release" "postgres_deployment" {
  depends_on = [
    kubernetes_namespace.namespace,
    kubernetes_secret.postgres_secret,
    kubernetes_secret.pgpool_secret,
    kubernetes_storage_class.storage_class,
  ]

  name       = "postgres-release"
  namespace  = kubernetes_namespace.namespace.metadata.0.name
  repository = "https://charts.bitnami.com/bitnami"

  chart   = "postgresql-ha"
  version = "11.2.1"

  # override default values
  values = [yamlencode({
    postgresql = {
      replicaCount = var.postgres_replica_count
      existingSecret = kubernetes_secret.postgres_secret.metadata.0.name
      username = var.postgres_username
      database = var.postgres_database
      repmgrUsername = var.postgres_repmgr_username
      repmgrDatabase = var.postgres_repmgr_database
    }
    pgpool = {
      adminUsername = var.postgres_pgpool_admin_username
      existingSecret = kubernetes_secret.pgpool_secret.metadata.0.name
    }
    persistence = {
      enabled = true
      storageClass = kubernetes_storage_class.storage_class.metadata.0.name
      size = "${var.postgres_storage_size_in_gb}Gi"
      accessModes = [
        "ReadWriteMany",
      ]
    }
  })]
}

locals {
  url_shortener_app = "url-shortener"
}

# deploy url-shortener with replicas
resource "kubernetes_deployment" "url_shortener_deployment" {
    depends_on = [
        kubernetes_namespace.namespace,
        kubernetes_secret.postgres_secret,
        helm_release.postgres_deployment,
    ]

    metadata {
        name = "url-shortener-deployment"
        namespace = kubernetes_namespace.namespace.metadata.0.name
        labels = {
          app = local.url_shortener_app
        }
    }

    spec {
        replicas = var.url_shortener_replica_count

        selector {
          match_labels = {
            app = local.url_shortener_app
          }
        }

      template {
        metadata{
            labels = {
                app = local.url_shortener_app
            }
        }

        spec{
            container {
                name = "url-shortener-xxx"
                image = "aria3ppp/url-shortener"

                port {
                  container_port = 5432
                }

                # envs
                env {
                  name = "SERVER_PORT"
                  value = 8080
                }
                env {
                  name = "POSTGRES_USER"
                  value = var.postgres_username
                }
                env {
                    name = "POSTGRES_PASSWORD"
                    value_from {
                      secret_key_ref {
                        name = kubernetes_secret.postgres_secret.metadata.0.name
                        key = var.kubernetes_postgres_secret__password_key
                      }
                    }
                }
                env {
                    name = "POSTGRES_HOST"
                    value = "${helm_release.postgres_deployment.name}-${helm_release.postgres_deployment.chart}-pgpool"
                    # value = [for km in [for c in split("---", helm_release.postgres_deployment.manifest) : yamldecode(c) if length(trimspace(c))>0] : km.metadata.name if contains(values(km.metadata.labels), local.pgpool_service_label_value)][0]
                }
                env {
                  name = "POSTGRES_PORT"
                  value = 5432
                }
                env {
                  name = "POSTGRES_DB"
                  value = var.postgres_database
                }

                # healthcheck
                liveness_probe {
                  http_get {
                    path = "/test/redirection-destination"
                    port = 8080
                  }
                  initial_delay_seconds = 20
                  period_seconds = 10
                  timeout_seconds = 5
                  success_threshold = 1
                  failure_threshold = 5
                }

                readiness_probe {
                  http_get {
                    path = "/test/redirection-destination"
                    port = 8080
                  }
                  initial_delay_seconds = 5
                  period_seconds = 5
                  timeout_seconds = 5
                  success_threshold = 1
                  failure_threshold = 5
                }

                # startup_probe {
                  
                # }
            }

        }
      }
    }
}

# create url-shortener service
resource "kubernetes_service" "url_shortener_service" {
    depends_on = [
      kubernetes_namespace.namespace,
      kubernetes_deployment.url_shortener_deployment,
    ]

    metadata {
        name = "url-shortener-service"
        namespace = kubernetes_namespace.namespace.metadata.0.name
    }

    spec {
        selector = {
          app = local.url_shortener_app
        }

        type = "NodePort"
        port {
          port = 8080
          target_port = 8080
          node_port = 30000
        }
    }
}

output "url-shortener_server_address" {
  value = "$(minikube ip):${kubernetes_service.url_shortener_service.spec.0.port.0.node_port}"
}