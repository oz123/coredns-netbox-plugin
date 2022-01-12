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

## Enabling

To activate the *netbox* plugin you need to compile CoreDNS with the plugin added
to `plugin.cfg`

```
netbox:github.com/oz123/coredns-netbox-plugin
```

### Ordering in plugin.cfg

The ordering of plugins in the `plugin.cfg` file is important to ensure you
get the behaviour you expect when using multiple plugins in a config section.

For example, in order to utilise the native cache plugin, ensure that you add
the *netbox* plugin _after_ `cache:cache` but _before_ any plugins you want to
be able to fall-through to (eg `file:file` or `forward:forward`).

## Syntax

```
netbox [ZONES...] {
    token TOKEN
    url URL
    localCacheDuration DURATION
    timeout DURATION
    tls CERT KEY CACERT 
    ttl DURATION
    fallthrough [ZONES...]
}
```

* **ZONES** zones that *netbox* should be authoritative for.
* `token` **TOKEN** sets the API token used to authenticate against NetBox
  (**REQUIRED**).
* `url` **URL** defines the URL *netbox* should query. This URL must be
  specified in full as `SCHEME://HOST/api/ipam/ip-addresses` (**REQUIRED**).
* `localCacheDuration` **DURATION** sets the time to cache responses from
  NetBox. Default is 0s (disabled). *NOTE*: It is recommended to use
  the built-in cache plugin instead.
* `timeout` **DURATION** sets the HTTP timeout for requests to NetBox. Default
  is 10s.
* `tls` followed by:
    * no arguments, if the server certificate is signed by a system-installed
      CA and no client cert is needed (this is the default if HTTPS is used).
    * a single argument that is the CA PEM file, if the server cert is not
      signed by a system CA and no client cert is needed.
    * two arguments - path to cert PEM file, the path to private key PEM file -
      if the server certificate is signed by a system-installed CA and a client
      certificate is needed.
    * three arguments - path to cert PEM file, path to client private key PEM
      file, path to CA PEM file - if the server certificate is not signed by a
      system-installed CA and client certificate is needed.
* `ttl` **DURATION** defines the TTL of records returned from *netbox*. Default
  is 1h (3600s).
* `fallthrough` If a zone matches but no record can be generated, pass request
  to the next plugin. If **[ZONES…]** is omitted, then fallthrough happens for
  all zones for which the plugin is authoritative. If specific zones are
  listed then only queries for those zones will be subject to fallthrough.

## Examples

Send all requests to NetBox:

```
. {
    netbox {
        token SuperSecretNetBoxAPIToken
        url https://netbox.example.org/api/ipam/ip-addresses
    }
}
```

Send requests within `example.org` to NetBox and fall-through to the `file`
plugin in order to respond to unsupported record types (ie `SOA`, `NS` etc):

```
example.org {
    netbox {
        token SuperSecretNetBoxAPIToken
        url https://netbox.example.org/api/ipam/ip-addresses
        fallthrough
    }
    file db.example.org
}

```

Send requests within `example.org` to NetBox and fall-through to the `forward`
plugin to send requests upstream and `cache` responses:

```
example.org {
    netbox {
        token SuperSecretNetBoxAPIToken
        url https://netbox.example.org/api/ipam/ip-addresses
        fallthrough
    }
    forward 1.1.1.1 1.0.0.1
    cache
}
```

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
