locals {
  root_disk_size = lookup(var.providerInitConfig.masterInstanceClass, "rootDiskSizeInGb", "")
  image_name = var.providerInitConfig.masterInstanceClass.imageName
  flavor_name = var.providerInitConfig.masterInstanceClass.flavorName
  security_group_names = lookup(var.providerInitConfig.masterInstanceClass, "securityGroups", [])
  network_security = local.pod_network_mode == "DirectRoutingWithPortSecurityEnabled"
}

module "simple_master" {
  source = "../../../terraform-modules/simple-master"
  prefix = local.prefix
  root_disk_size = local.root_disk_size
  image_name = local.image_name
  flavor_name = local.flavor_name
  keypair_ssh_name = data.openstack_compute_keypair_v2.ssh.name
  master_external_port_id = local.network_security ? openstack_networking_port_v2.master_external_with_security[0].id : openstack_networking_port_v2.master_external_without_security[0].id
  config_drive = !local.external_network_dhcp
}

module "security_groups" {
  source = "../../../terraform-modules/security-groups"
  security_group_names = local.security_group_names
}

data "openstack_compute_keypair_v2" "ssh" {
  name = local.prefix
}

data "openstack_networking_network_v2" "external" {
  name = local.external_network_name
}

resource "openstack_networking_port_v2" "master_external_with_security" {
  count = local.network_security ? 1 : 0
  network_id = data.openstack_networking_network_v2.external.id
  admin_state_up = "true"
  security_group_ids = module.security_groups.security_group_ids

  allowed_address_pairs {
    ip_address = local.pod_subnet_cidr
  }
}

resource "openstack_networking_port_v2" "master_external_without_security" {
  count = local.network_security ? 0 : 1
  network_id = data.openstack_networking_network_v2.external.id
  admin_state_up = "true"
  security_group_ids = module.security_groups.security_group_ids
}
