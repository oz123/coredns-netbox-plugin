# coredns-netbox-plugin

This plugin gets records from NetBox[1] either native or netbox-plugin-dns[2].

Supported records with legacy API are: A, AAAA, PTR

Supported records with [Netbox DNS Plugin](https://github.com/peteeckel/netbox-plugin-dns)
currently are: A, AAAA, PTR, NS, SOA, MX, TXT, CNAME

It uses the REST API of netbox to ask for a an IP address of a hostname:

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

```
curl https://netbox.example.org/api/plugins/netbox-dns/records/?zone=example.org&qfdn=netbox.example.org&active=true

{
    "count": 1,
    "next": null,
    "previous": null,
    "results": [
        {
            "id": 1,
            "url": "https://netbox.example.org/api/plugins/netbox-dns/records/1/",
            "zone": {
                "id": 1,
                "url": "https://netbox.example.org/api/plugins/netbox-dns/zones/1/",
                "display": "decp.nepgroup.io",
                "name": "decp.nepgroup.io",
                "view": {
                    "id": 1,
                    "url": "https://netbox.example.org/api/plugins/netbox-dns/views/1/",
                    "display": "_default_",
                    "name": "_default_",
                    "default_view": true,
                    "description": "Default View"
                },
                "status": "active",
                "active": true,
                "rfc2317_prefix": null
            },
            "display": "netbox.example.org [A]",
            "type": "A",
            "name": "netbox",
            "fqdn": "netbox.example.org.",
            "value": "192.168.0.1",
            "status": "active",
            "ttl": null,
            "description": "",
            "tags": [],
            "created": "2025-03-21T08:55:12.590626Z",
            "last_updated": "2025-03-21T08:55:12.590629Z",
            "managed": true,
            "disable_ptr": false,
            "ptr_record": {
                "id": 41,
                "url": "https://netbox.example.org/api/plugins/netbox-dns/records/2/",
                "display": "1.0.168.192.in-addr.arpa [PTR]",
                "type": "PTR",
                "name": "1",
                "value": "netbox.example.org.",
                "status": "active",
                "ttl": null,
                "zone": {
                    "id": 3,
                    "url": "https://netbox.example.org/api/plugins/netbox-dns/zones/2/",
                    "display": "168.192.in-addr.arpa",
                    "name": "168.192.in-addr.arpa",
                    "view": {
                        "id": 1,
                        "url": "https://netbox.example.org/api/plugins/netbox-dns/views/1/",
                        "display": "_default_",
                        "name": "_default_",
                        "default_view": true,
                        "description": "Default View"
                    },
                    "status": "active",
                    "active": true,
                    "rfc2317_prefix": null
                },
                "managed": true,
                "active": null
            },
            "address_record": null,
            "active": true,
            "custom_fields": {},
            "tenant": null,
            "ipam_ip_address": {
                "id": 21,
                "url": "https://netbox.example.org/api/ipam/ip-addresses/1/",
                "display": "192.168.0.1/24",
                "family": {
                    "value": 4,
                    "label": "IPv4"
                },
                "address": "192.168.0.1/24",
                "description": ""
            },
            "absolute_value": "192.168.0.1"
        }
    ]
}
```

## Enabling

To activate the _netbox_ plugin you need to compile CoreDNS with the plugin added
to `plugin.cfg`

```
netbox:github.com/oz123/coredns-netbox-plugin
```

### Ordering in plugin.cfg

The ordering of plugins in the `plugin.cfg` file is important to ensure you
get the behaviour you expect when using multiple plugins in a
[Corefile server block][3].

For example, in order to utilise the native cache plugin, ensure that you add
the _netbox_ plugin _after_ `cache:cache` but _before_ any plugins you want to
be able to fall-through to (eg `file:file` or `forward:forward`).

## Syntax

```
netbox [ZONES...] {
  token TOKEN
  url URL
  tls CERT KEY CACERT
  fallthrough [ZONES...]
}
```

- **ZONES** zones that the _netbox_ should be authoritative for.
  If you use DNS Plugin for Netbox you MUST specify a zone
- `token` **TOKEN** sets the API token used to authenticate against NetBox
  (**REQUIRED**).
- `url` **URL** defines the URL _netbox_ should query. This URL must be
  specified as `SCHEME://HOST` (**REQUIRED**).
- `tls` is followed by:

  - no arguments, if the server certificate is signed by a system-installed
    CA and no client cert is needed (this is the default if HTTPS is used).
  - a single argument that is the CA PEM file, if the server cert is not
    signed by a system CA and no client cert is needed.
  - two arguments - path to cert PEM file, the path to private key PEM file -
    if the server certificate is signed by a system-installed CA and a client
    certificate is needed.
  - three arguments - path to cert PEM file, path to client private key PEM
    file, path to CA PEM file - if the server certificate is not signed by a
    system-installed CA and client certificate is needed.

  These options set certificate verification method for the NetBox server if
  HTTPS is used to access the API.

- `ttl` **DURATION** defines the TTL of records returned from _netbox_. Default
  is 1h (3600s).
- `timeout` **DURATION** defines the HTTP timeout for API requests against
  NetBox. Default is 5s.
- `fallthrough` If a zone matches but no record can be generated, pass request
  to the next plugin. If **[ZONES…]** is omitted, then fallthrough happens for
  all zones for which the plugin is authoritative. If specific zones are listed
  then only queries for those zones will be subject to fallthrough.

The config parameters `token`, `url` and `localCacheDuration` are required.

## Examples

### LEGACY

Send all requests to NetBox:

```
. {
    netbox {
        token SuperSecretNetBoxAPIToken
        url https://netbox.example.org
    }
}
```

Send requests within `example.org` to NetBox and fall-through to the `file`
plugin in order to respond to unsupported record types (ie `SOA`, `NS` etc):

```
. {
    netbox example.org {
        token SuperSecretNetBoxAPIToken
        url https://netbox.example.org
        fallthrough
    }
    file db.example.org
}

```

Handle all requests with _netbox_ and fall-through to the `forward`
plugin for requests within `example.org` with caching via the `cache` plugin:

```
. {
    netbox {
        token SuperSecretNetBoxAPIToken
        url https://netbox.example.org
        fallthrough example.org
    }
    forward . 1.1.1.1 1.0.0.1
    cache
}
```

### WITH NETBOX DNS PLUGIN

Send requests within `example.org` to NetBox and fall-through to the `file`
plugin in order to respond to unsupported record types:

```
. {
    netbox example.org {
        token SuperSecretNetBoxAPIToken
        url https://netbox.example.org
        fallthrough
    }
    file db.example.org
}

```

Resolve requests within `example.org` and PTR for `0.168.192.in-addr.arpa` with NetBox

```
. {
    netbox example.org 0.168.192.in-addr.arpa {
        token SuperSecretNetBoxAPIToken
        url https://netbox.example.org
    }
}

```

## Changelog

0.2 - Cleanup add IPv6 support

- Refactor query.go
- Add tests for IPv6
- Enable IPv6 in `query.go`

  0.1 - Initial Naive release

- Got it somehow working
- Gather feedback

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
[2]: https://github.com/peteeckel/netbox-plugin-dns
[3]: https://coredns.io/manual/toc/#server-blocks
