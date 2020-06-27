# coredns-netbox-plugin

This plugin gets an A record from NetBox[1]. It uses the REST API of netxbox
to ask for a an IP address of a hostname:

https://netbox.example.org/api/ipam/ip-addresses/?dns_name=example-vm-host


```
{
    "count": 1,
    "next": null,
    "previous": null,
    "results": [
        {
            "id": 426,
            "family": {
                "value": 4,
                "label": "IPv4"
            },
            "address": "192.168.1.101/25",
            "vrf": null,
            "tenant": null,
            "status": {
                "value": 1,
                "label": "Active"
            },
            "role": null,
            "interface": {
                "id": 452,
                "url": "https://netbox.example.org/api/virtualization/interfaces/452/",
                "device": null,
                "virtual_machine": {
                    "id": 10,
                    "url": "https://netbox.example.org/api/virtualization/virtual-machines/10/",
                    "name": "VM22"
                },
                "name": "ens12"
            },
            "nat_inside": null,
            "nat_outside": null,
            "dns_name": "vm22-jump-host",
            "description": "",
            "tags": [],
            "custom_fields": {
                "dhcp_route": null
            },
            "created": "2020-03-04",
            "last_updated": "2020-03-04T16:57:07.375937Z"
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
   }
}
```


## Credits

This plugin is heavily based on the code of the redis-plugin for CoreDNS.


[1]: https://netbox.readthedocs.io/en/stable/
