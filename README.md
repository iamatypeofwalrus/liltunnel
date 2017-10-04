# Lil' HTTP Tunnel
Just a simple, little TCO / hTTP proxy over an SSH tunnel

## So you need to setup paswordless access to your server
### Generate an SSH Key for liltunnel
```sh
ssh-keygen -t rsa -b 4096 -C "your_email@example.com" -P "" -f ~/.ssh/liltunnel_rsa
```
### Copy your public key to your server
In order to SSH with a public key you'll need to add `liltunnel_rsa.pub` to
`~/.ssh/authorized_keys` on your server

```sh
cat ~/.ssh/liltunnel_rsa.pub | pbcopy

# SSH into your server
cd ~/.ssh

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

## CLI API
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