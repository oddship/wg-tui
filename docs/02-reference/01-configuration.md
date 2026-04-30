# Configuration

`wgt` stores config in HUML.

Example:

```huml
%HUML v0.2.0

server:
  url: "https://warpgate.example.com"
  token: "<warpgate-api-token>"
  insecure_skip_tls_verify: false

ssh:
  username: "user@example.com"
  host: "warpgate.example.com"
  port: 2222
  binary: "ssh"
  extra_args::
    - "-o"
    - "ServerAliveInterval=30"

cache:
  dir: "~/.cache/wgt"
  ttl: "10m"
  max_age: "168h"
  use_stale_on_error: true
```
