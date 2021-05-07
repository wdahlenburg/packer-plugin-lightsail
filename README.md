# AWS Lightsail Packer Plugin

This project creates a custom plugin for HashiCorp packer AWS Lightsail instances. The core is based off of https://github.com/YafimK/packer-lightsail/, which does not support Packer 1.7+ changes. The major upgrade change is that the HCL2 format is now required instead of JSON.

## Building

Download the repo and build the binary from source:

```
$ mkdir -p $GOPATH/src/github.com/wdahlenburg/
$ cd $GOPATH/src/github.com/wdahlenburg/
$ git clone https://github.com/wdahlenburg/packer-plugin-lightsail.git
$ cd $GOPATH/src/github.com/wdahlenburg/packer-plugin-lightsail
$ go build
```

Place the plugin in the proper directory along with the SHA256 hash:

```
$ mkdir -p ~/.packer.d/plugins/github.com/wdahlenburg/lightsail/
$ cp packer-plugin-lightsail ~/.packer.d/plugins/github.com/wdahlenburg/lightsail/packer-plugin-lightsail_v0.0.4_x5.0_darwin_amd64
$ sha256sum ~/.packer.d/plugins/github.com/wdahlenburg/lightsail/packer-plugin-lightsail_v0.0.4_x5.0_darwin_amd64 | awk '{print $1}' | tee ~/.packer.d/plugins/github.com/wdahlenburg/lightsail/packer-plugin-lightsail_v0.0.4_x5.0_darwin_amd64_SHA256SUM
```

## Installing

You can install the package right from GitHub. Use an HCL2 example like the one in the [example](example) folder.

Packer will download the plugin based off the criteria in the required_plugins section.

```
$ packer init -upgrade example
```

## Example

Make sure you already have the plugin installed. See Building or Installing first.

```
$ export AWS_ACCESS_KEY=AKIA.....
$ export AWS_SECRET_KEY=ABCD.....
$ packer build example
```

## Requirements

-	[packer-plugin-sdk](https://github.com/hashicorp/packer-plugin-sdk) >= v0.1.0
-	[Go](https://golang.org/doc/install) >= 1.16

## Packer Compatibility
This scaffolding template is compatible with Packer >= v1.7.0


## Todo

- [ ] Delete Key Pair After Build
- [ ] Identify and Support Additional Variables
- [ ] Improve Docs
