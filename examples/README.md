To use these files use the netbox stack using
the compose file found in `testing/`

```
$ ./coredns -p 1053 -conf Corefile.example &
$ dig -p 1053 @127.0.0.1 somehost.company
```

To get AAAA records:

```
dig -p 1053 -6 @127.0.0.1 mail2.foo.com AAAA
```

