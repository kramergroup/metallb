# MetalLB

MetalLB is a load-balancer implementation for bare
metal [Kubernetes](https://kubernetes.io) clusters, using standard
routing protocols.

[![Project maturity: alpha](https://img.shields.io/badge/maturity-alpha-yellow.svg)](https://metallb.universe.tf/concepts/maturity/) [![license](https://img.shields.io/github/license/google/metallb.svg?maxAge=2592000)](https://github.com/google/metallb/blob/master/LICENSE) [![CircleCI](https://img.shields.io/circleci/project/github/google/metallb.svg)](https://circleci.com/gh/google/metallb) [![Containers](https://img.shields.io/badge/containers-ready-green.svg)](https://hub.docker.com/u/metallb) [![Go report card](https://goreportcard.com/badge/github.com/google/metallb)](https://goreportcard.com/report/github.com/google/metallb)

Check out [MetalLB's website](https://metallb.universe.tf) for more
information.

# Fork

This fork implements a new feature to source IPs. Metallb has the ability to source IPs from a predefined list. This creates problems, if the environment does not allow for reserved IP ranges (e.g., due to coorporate policy) and only provides for dynamically assigned IPs (e.g., via DHCP). This fork implements capabilities to define `address-services` (dynamic) in addition to `address-pools` (static).

## Address services

Address services are restful HTTP endpoints that implement the following API:

| URL          | Method | Body                               | Description                                                                 |
| ------------ | ------ | ---------------------------------- | --------------------------------------------------------------------------- |
| `/v1/config` | GET    |                                    | Obtain service configuration (incl. the CIDR ranges covered by the service) |
| `/v1/status` | GET    |                                    | Obtain service status (not used in metallb)                                 |
| `/v1/ip`     | POST   | `{"service":"namespace/svc-name"}` | Request a new IP for `service`                                              |
|              | DELETE | `{"ip":"xxx.xxx.xxx.xxx"}`         | Return an IP                                                                |
| `/v1/mac`    | POST   | `{"macs":["xx.xx.xx.xx.xx.xx"]`    | Provide a list of hardware addresses to use                                 |
|              | DELETE | `{"macs":["xx.xx.xx.xx.xx.xx"]`    | Remove a list of hardware addresses from usage                              |  

A reference implementation of a DHCP-enabled Address Service can be found here: [DHCPManager](https://github.com/kramergroup/dhcpmanager)

### Configuration

Address services are configured using the config file. A minimal configuration provides the API base-URL and routing protocol:

```yaml
address-services:
- name: example
  url: http://dhcpmanager
  protocol: layer2
```

# Contributing

We welcome contributions in all forms. Please check out
the
[hacking and contributing guide](https://metallb.universe.tf/community/#contributing)
for more information.

Participation in this project is subject to
a [code of conduct](https://metallb.universe.tf/community/code-of-conduct/).

One lightweight way you can contribute is
to
[tell us that you're using MetalLB](https://github.com/google/metallb/issues/5),
which will give us warm fuzzy feelings :).

# Disclaimer

This is not an official Google project, it is just code that happens
to be owned by Google.
