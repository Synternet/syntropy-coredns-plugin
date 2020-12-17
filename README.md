# CoreDNS (Syntropy enabled)

This repository contains a Syntropy CoreDNS plugin that will automatically resolve domain names (service.endpoint_hostname) to their respective services.
You can either compile it manually or use the provided Docker image (syntropynet/coredns)

## Using Docker

The default cache duration is set as 300 seconds, however it is configurable by the env variable ```LOCAL_CACHE_DURATION```.
The command needed to run a default instance of Syntropy CoreDNS is
```
docker run -d -e SYNTROPY_CONTROLLER_URL="https://controller-prod-server.noia.network" -e SYNTROPY_USERNAME="<YOUR_USERNAME>" -e SYNTROPY_PASSWORD="<YOUR_PASSWORD>" syntropy/coredns 
```

## Manually compile CoreDNS

- Add the line ```syntropy:syntropy``` to coredns/plugin.cfg, preferably after the ```acl/acl``` line (priority of plugins matters).
- Move all of the directory contents of ```src/``` to ```coredns/plugin/syntropy/```.
- Run make in the root CoreDNS directory to get a binary
- Follow the example configuration in ```Corefile.example```
