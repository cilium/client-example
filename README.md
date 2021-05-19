# Cilium client API example

Simple example illustrating use of the Cilium API.

See [Cilium API reference](https://docs.cilium.io/en/stable/api/) for further
documentation on use of the API.

Repository that contains a "Hello World" source code for all Cilium minor
versions.

Each directory has its own `vendor/` directory, and the clients can be compiled
for each Cilium version.

## Example Client Output

```bash
$ cd latest
$ go build ./main.go
$ ./main 
EP ID 10 has IP addresses: 10.17.138.46
EP ID 387 has IP addresses: 10.17.165.167
EP ID 2170 has IP addresses: 10.17.145.34
EP ID 2374 has IP addresses: 10.17.111.212
EP ID 2399 has IP addresses: 10.17.200.251
EP ID 3400 does not have an IP address
```
