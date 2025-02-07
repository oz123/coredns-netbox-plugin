CoreDNS plugin that reads a simple rest API
-------------------------------------------

How to build and test.

1. Clone this repository and check out the `general-rest` branch

```
$ git clone git@github.com:oz123/coredns-rest-plugin.git 
$ git checkout general-rest
```
2. Fetch and unpack CoreDNS sources.
```
$ make coredns-fetch
$ make coredns-unzip
```
3. Patch go.mod:
```
$ make coredns-patch-go.mod
```

4. Add the plugin to the config
```
$ make coredns-add-rest-plugin
```

5. Build CoreDNS with the plugin
```
$ make coredns-build
```

6. Test the plugin was compiled properly
```
$ ./coredns-1.12.0/coredns -plugins | grep rest
```
It should show `rest` if succeeded.

7. Run the dummy API:
```
$ make test-api 
```

8. Run coredns with the config enabling the plugin:
```
$ sudo ./coredns-1.12.0/coredns -conf examples/Corefile.rest.example
```

9. Query coredns using dig
```
$ dig @127.0.0.1 example.com
```

It should output:
```
; <<>> DiG 9.18.29 <<>> @127.0.0.1 example.com
; (1 server found)
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 34118
;; flags: qr aa rd; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 1
;; WARNING: recursion requested but not available

;; OPT PSEUDOSECTION:
; EDNS: version: 0, flags:; udp: 1232
; COOKIE: 3d7162cca245c873 (echoed)
;; QUESTION SECTION:
;example.com.			IN	A

;; ANSWER SECTION:
example.com.		3555	IN	A	1.2.3.4

;; Query time: 0 msec
;; SERVER: 127.0.0.1#53(127.0.0.1) (UDP)
;; WHEN: Fri Feb 07 14:13:52 CET 2025
;; MSG SIZE  rcvd: 79
```
