resource "yandex_vpc_network" "vpc" {
  name = "vct-vpc"
}

resource "yandex_vpc_gateway" "nat" {
  name     = "nat-gateway"
  folder_id = var.yc_folder_id
  shared_egress_gateway {}
}

resource "yandex_vpc_route_table" "nat" {
  network_id = yandex_vpc_network.vpc.id
  name       = "nat-route"
  
  static_route {
    destination_prefix = "0.0.0.0/0"
    gateway_id         = yandex_vpc_gateway.nat.id
  }
}

resource "yandex_vpc_subnet" "subnet-a" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.vpc.id
  v4_cidr_blocks = ["10.0.1.0/24"]
  route_table_id = yandex_vpc_route_table.nat.id
}

resource "yandex_iam_service_account" "k8s-master-sa" {
  folder_id = var.yc_folder_id
  name      = "vct-k8s-master-sa"
}

resource "yandex_iam_service_account" "k8s-node-sa" {
  folder_id = var.yc_folder_id
  name      = "vct-k8s-node-sa"
}

resource "yandex_resourcemanager_folder_iam_member" "k8s-master-editor" {
  folder_id = var.yc_folder_id
  role      = "editor"
  member    = "serviceAccount:${yandex_iam_service_account.k8s-master-sa.id}"
  depends_on = [yandex_iam_service_account.k8s-master-sa]
}

resource "yandex_resourcemanager_folder_iam_member" "k8s-node-puller" {
  folder_id = var.yc_folder_id
  role      = "container-registry.images.puller"
  member    = "serviceAccount:${yandex_iam_service_account.k8s-node-sa.id}"
  depends_on = [yandex_iam_service_account.k8s-node-sa]
}

resource "yandex_kubernetes_cluster" "vct-cluster" {
  name                     = var.cluster_name
  network_id               = yandex_vpc_network.vpc.id
  service_account_id       = yandex_iam_service_account.k8s-master-sa.id
  node_service_account_id  = yandex_iam_service_account.k8s-node-sa.id
  
  master {
    zonal {
      zone      = "ru-central1-a"
      subnet_id = yandex_vpc_subnet.subnet-a.id
    }
    version   = "1.31"
    public_ip = true
  }

  depends_on = [
    yandex_resourcemanager_folder_iam_member.k8s-master-editor,
    yandex_resourcemanager_folder_iam_member.k8s-node-puller
  ]
}

resource "yandex_kubernetes_node_group" "vct-nodes" {
  cluster_id = yandex_kubernetes_cluster.vct-cluster.id
  name       = "vct-nodes"
  version    = "1.31"
  
  scale_policy {
    fixed_scale {
      size = 1
    }
  }
  
  instance_template {
    platform_id = "standard-v3"
    
    resources {
      memory        = 2
      core_fraction = 50
    }
    
    boot_disk {
      type = "network-ssd"
      size = 30
    }
  }
}