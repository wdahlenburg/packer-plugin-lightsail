packer {
  required_plugins {
    scaffolding = {
      version = ">=v0.1.0"
      source  = "github.com/wdahlenburg/lightsail"
    }
  }
}

source "scaffolding-my-builder" "foo-example" {
  mock = local.foo
}

source "scaffolding-my-builder" "bar-example" {
  mock = local.bar
}

build {
  sources = [
    "source.scaffolding-my-builder.foo-example",
  ]

  source "source.scaffolding-my-builder.bar-example" {
    name = "bar"
  }
}
