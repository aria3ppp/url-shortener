variable "postgres_username" {
  default = "postgres"
  type = string
}

variable "postgres_password" {
    sensitive = true
    type = string
}

variable "postgres_user_password" {
    sensitive = true
    type = string
}

variable "postgres_database" {
  default = "custom-db"
  type = string
}

variable "postgres_repmgr_username" {
    default = "repmgr"
    type = string
}

variable "postgres_repmgr_password" {
    sensitive = true
    type = string
}

variable "postgres_repmgr_database" {
    default = "repmgr"
    type = string
}

variable "postgres_pgpool_admin_username" {
    default = "admin"
    type = string
}

variable "postgres_pgpool_admin_password" {
    sensitive = true
    type = string
}

variable "postgres_replica_count" {
    default = 3
    type = number
    validation {
        condition = var.postgres_replica_count >= 3 && var.postgres_replica_count % 2 == 1
        error_message = "postgres replica count must be an odd number with minimum value of 3"
    }
}

variable "postgres_storage_size_in_gb" {
    default = 0.5
    type = number
    validation {
      condition = var.postgres_storage_size_in_gb >= 0.5
      error_message = "minimum postgres storage size in giga byte is 0.5"
    }
}

##################################################################################

variable "kubernetes_namespace" {
    default = "url-shortener-ns"
    type = string
}

variable "kubernetes_postgres_secret__password_key" {
    default = "password"
    type = string
}

variable "kubernetes_postgres_secret__postgres_password_key" {
    default = "postgres-password"
    type = string
}

variable "kubernetes_postgres_secret__repmgr_password_key" {
    default = "repmgr-password"
    type = string
}

variable "kubernetes_pgpool_secret__admin_password_key" {
    default = "admin-password"
    type = string
}

##################################################################################

variable "url_shortener_replica_count" {
  default = 2
  type = number
  validation {
    condition = var.url_shortener_replica_count % 1 == 0 && var.url_shortener_replica_count >= 1
    error_message = "url-shortener replica count is a whole number with minimum value of 1"
  }
}