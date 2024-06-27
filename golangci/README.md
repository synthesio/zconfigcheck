### Golangci-lint plugin

In order to use `zconfigcheck` with [`golangci-lint`](https://golangci-lint.run/), you will need
to use the `golangci-lint custom` command as explained [here](https://golangci-lint.run/plugins/module-plugins/#the-automatic-way).

You can use the reference [custom modules configuration file](.custom-gcl.yml) for your integration.

We also provide an example [linter settings file](golangci.zconfigcheck.yaml) with some suggested configuration parameters.

Once all configuration files are in place, you can run the command: 
```console
$ golangci-lint custom
```

This will produce a new `custom-gcl` binary which supports the `zconfigcheck` custom linter. 
You can verify this by running:
```console
$ ./custom-gcl linters | grep zconfigcheck 
```
