# Lil' TCP / HTTP Tunnel over SSH
Lil' Tunnel is a simple CLI application for proxying TCP / HTTP requests over SSH.

## TCP AND HTTP?!?
Yes. You and I both know that HTTP runs on top of TCP. By interacting at the `HTTP`
level `liltunnel` can cache HTTP responses to disk. I find this pretty
handy when I need a quasi-offline mode. If you don't need this feature `liltunnel` happily defaults to `tcp`
like any normal SSH tunnel.

## CLI Args
```
Usage:
  liltunnel --port-mapping 80:8080 --remote me@remote.example.com --identity ~/.ssh/liltunnel_rsa

Options:
  -p, --port-mapping=           local:remote or port. If remote is not specified local port is used
  -r, --remote=                 username@remote.example.com or remote.example.com. If username is not specified the current $USER is used
  -i, --identity=               private key to be used when establishing a connection to the remote (default: ~/.ssh/id_rsa)
  -o, --known-hosts=            known hosts file (default: ~/.ssh/known_hosts)
  -n, --protocol=[http|tcp]     network protocol to use when tunneling (default: tcp)
  -c, --http-cache              HTTP only. Cache all succesful responses to GET requests to disk
  -t, --http-cache-ttl=         HTTP only. Expressed in seconds. Length of time to keep successful responses in cache. Defaults to 12 hours
  -s, --http-cache-serve-stale  HTTP only. Always return return a stale read from the cache. Handy if you need an offline mode
  -v, --verbose

Help Options:
  -h, --help                    Show this help message
```

## Setup passwordless access to your server
### Generate an SSH Key for liltunnel
```sh
MY_EMAIL=...
ssh-keygen -t rsa -b 4096 -C "$MY_EMAIL" -P "" -f ~/.ssh/liltunnel_rsa
```
### Copy your public key to your server
In order to SSH with a public key you'll need to add `~/.ssh/liltunnel_rsa.pub` to
`~/.ssh/authorized_keys` on your server

```sh
cat ~/.ssh/liltunnel_rsa.pub | pbcopy


# SSH into your server
ssh ...

# From your remote machine
mkdir -p ~/.ssh

# If authorized_keys does not exist run the following
touch authorized_keys 
chmod 600 authorized_keys

# open authorized_keys with your favorite editor and paste your public key
# into the file.
```

### Verify with liltunnel
Run the following command:

```sh
liltunnel --port 2009 --host your.host.example.com --ssh-key ~/.ssh/liltunnel_rsa
```

As a smoke test you can run something like the following, or run something else
that makes an HTTP request to 2009
```
curl -v http://localhost:2009
```

## Usages
```
liltunnel --local-host-port 1080 \
          --remote-host-port 1081 \
          --remote-host-user-name root \
          --remote-host-name the.best.example.com \
          --ssh-key ~/.ssh/liltunnel_rsa \
          --protocol-http \
          --http-cache-responses

liltunnel -p 1080 -h example.com --protocol-http --http-cache-responses
liltunnel --port-mapping 1080:1080 --remotehost root@example.com
liltunnel -p 1080:1080 -h root@example.com -k ~/.ssh/liltunnel_rsa
```
