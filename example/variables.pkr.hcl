variable "access_key" {
  type = string
  sensitive = true
  default = env("AWS_ACCESS_KEY")
}

variable "secret_key" {
  type = string
  sensitive = true
  default = env("AWS_SECRET_KEY")
}

variable "regions" {
  default = ["us-east-1a"]
}

variable "snapshot_name" {
  type = string
  default = "example-snapshot"
}

variable "bundle_id" {
  type = string
  default = "nano_2_0"
}

variable "blueprint_id" {
  type = string
  default = "ubuntu_20_04"
}

variable "timeout" {
  type = string
  default = "15m"
}
