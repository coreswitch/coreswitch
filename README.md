# coreswitch

----

coreswitch is an open soruce project for EPC (Evolved Packet Core) of LTE and 5G
infrastructure. Right now we are implementing MME (Mobility Management Entity).
Other component will be implemented later on.

----

## Supported Platform

Right now only Ubuntu 18.04 is supported.

## Build

To build the system.  ASN1 handling C library needs to be built.

``` shell
$ cd coreswitch/pkg/s1ap/asn1
$ make lib
$ sudo make install
```

After this

``` shell
$ go get github.com/coreswitch/coreswitch/cmd/mmed
```

will build mmed.

## Run

Just simply execute `mmed` will start MME handling on all of interfaces.

``` shell
$ mmed
2019/08/28 14:21:42 Listen on 127.0.0.1/[::1%lo]/10.211.55.26/[fe80::38e7:5a76:4355:51d7%enp0s5]/[fe80::79d1:506d:7682:9ee5%enp0s6]/[fe80::cd97:63f4:fcad:b2c5%enp0s7]/172.18.0.1/172.17.0.1:36412

```
