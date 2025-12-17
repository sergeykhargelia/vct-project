terraform {
  required_providers {
    yandex = {
      source = "yandex-cloud/yandex"
    }
  }
  required_version = ">= 1.5"
}

provider "yandex" {
  token     = var.yc_oauth_token
  cloud_id  = var.yc_cloud_id
  folder_id = var.yc_folder_id
}