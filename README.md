# coredns-netbox-plugin

This plugin gets an A record from NetBox[1]. It uses the REST API of netbox
to ask for a an IP address of a hostname:

```
curl https://netbox.example.org/api/ipam/ip-addresses/?dns_name=example-vm-host

{
    "count": 1,
    "next": null,
    "previous": null,
    "results": [
        {
            "family": {
                "value": 4,
                "label": "IPv4"
            },
            "address": "192.168.1.101/25",
            "interface": {
                "id": 452,
                "url": "https://netbox.example.org/api/virtualization/interfaces/452/",
                "virtual_machine": {
                    "url": "https://netbox.example.org/api/virtualization/virtual-machines/10/",
                },
            },
        }
    ]
}
```

## Usage

To activate the plugin you need to compile CoreDNS with the plugin added
to `plugin.cfg`

```
netbox:github.com/oz123/coredns-netbox-plugin
```

Then add it to Corefile:

```
. {
   netbox {
      token <YOU-NETBOX-API-TOKEN>
      url <https://netbox.example.org>
      localCacheDuration <The duration to keep each entry locally before querying netbox again. Use go `time.Duration` notation>
   }
}
```

The config parameters are mandatory.

## Changelog

0.2 - Cleanup add IPv6 support
 
 * Refactor query.go
 * Add tests for IPv6
 * Enable IPv6 in ``query.go``

0.1 - Initial Naive release

 * Got it somehow working 
 * Gather feedback
## Developing locally

You can test the plugin functionallity with CoreDNS by adding the following to
`go.mod` in the source code directory of coredns.

```
replace github.com/oz123/coredns-netbox-plugin => <path-to-you-local-copy>/coredns-netbox-plugin
```

Testing against a remote instance of netbox is possible with SSH port forwarding:

```
Host YourHost
   Hostname 10.0.0.91
   ProxyJump YourJumpHost
   LocalForward 18443 192.168.1.128:8443
```

## Credits

This plugin is heavily based on the code of the redis-plugin for CoreDNS.


[1]: https://netbox.readthedocs.io/en/stable/
