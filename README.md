<p align="center">
    <img src ="https://raw.githubusercontent.com/otaviof/vault-handler/master/assets/logo/vault-handler.png"/>
</p>
<p align="center">
    <a alt="GoReport" href="https://goreportcard.com/report/github.com/otaviof/vault-handler">
        <img alt="GoReport" src="https://goreportcard.com/badge/github.com/otaviof/vault-handler">
    </a>
    <a alt="Code Coverage" href="https://codecov.io/gh/otaviof/vault-handler">
        <img alt="Code Coverage" src="https://codecov.io/gh/otaviof/vault-handler/branch/master/graph/badge.svg">
    </a>
    <a href="https://godoc.org/github.com/otaviof/vault-handler/pkg/vault-handler">
        <img alt="GoDoc Reference" src="https://godoc.org/github.com/otaviof/vault-handler/pkg/vault-handler?status.svg">
    </a>
    <a alt="CI Status" href="https://travis-ci.com/otaviof/vault-handler">
        <img alt="CI Status" src="https://travis-ci.com/otaviof/vault-handler.svg?branch=master">
    </a>
    <a alt="Docker-Cloud Build Status" href="https://hub.docker.com/r/otaviof/vault-handler">
        <img alt="Docker-Cloud Build Status" src="https://img.shields.io/docker/cloud/build/otaviof/vault-handler.svg">
    </a>
</p>

# `vault-handler` (WIP)

Is a manifest based application to upload and download secrets from
[Hashicorp-Vault](https://www.vaultproject.io/). Therefore, the manifest file is promoted as
"source-of-authority" over secrets that a given application may consume, where engineers can use
the manifest to upload secrets, and later on use it as configuration input for downloading secrets
on applications' behalf.

Once a set of secrets is uploaded to Vault, you can `copy` them over to Kubernetes, and keep secrets
in sync by running `copy` command at later times.

You can employ `vault-handler` as a Kubernetes
[init-container](https://kubernetes.io/docs/concepts/workloads/pods/init-containers) in order to
download secrets, and have it as a command-line application to upload.

## Installing

To install the command-line interface via `go get` use:

``` bash
go get -u github.com/otaviof/vault-handler/cmd/vault-handler
```

### Docker Images

Docker images can be bound at [Docker-Hub](https://hub.docker.com/r/otaviof/vault-handler). And for
example, to show the command line help use

``` bash
docker run --interactive --tty otaviof/vault-handler:latest --help
```

Downloading data from Vault would be, for instance:

``` bash
docker run otaviof/vault-handler:latest download
```

Consider exporting configuration as environment variables to change the behavior of `vault-handler`.

## HashiCorp Vault

This application is a client of Vault. You can choose either
[AppRole](https://www.vaultproject.io/docs/auth/approle.html) type of authentication, which will
request a runtime token, or directly use a
[existing token](https://www.vaultproject.io/docs/auth/token.html). Both ways are possible, but
AppRole would be the recommended method.

## Configuration

All the options in command-line can be set via environment variables. The convention of environment
variable names to add a prefix, `VAULT_HANDLER`, and the option name in capitals, replacing dashes
(`-`) by underscore (`_`). For instance, `--vault-addr` would become `VAULT_HANDLER_VAULT_ADDR`
in environment.

Additionally, command-line arguments overwrite what's informed via environment variables.

## Usage

The command-line interface is organized as follows:

```
vault-handler [command] [arguments] [manifest-files]
```

Where as `command` you can use `upload` or `download`. Please check `--help` output in command-line
to see all possible arguments.

A `upload` command example:

``` bash
vault-handler upload --input-dir /var/tmp --dry-run /path/to/manifest.yaml
```

And then to `download`, use for instance:

``` bash
vault-handler download --output-dir /tmp --dry-run /path/to/manifest.yaml
```

Afterwards you can `copy` secrets to Kubernetes:

``` bash
vault-handler copy --namespace default --dry-run /path/to/manifest.yaml
```

### Manifest

The following snippet is a manifest example, the actual secrets can be found
[here](./test/mock/kube-secrets).

``` yaml
---
secrets:
  ingress:
    path: secret/data/kube/tls
    type: kubernetes.io/tls
    data:
      - name: tls.crt
        extension: secret
      - name: tls.key
        extension: secret
```

Description of the options used in manifest:

- `secrets`: root of the manifest;
- `name`: arbitrary group "name". This group-name is also employed to name final files;
- `name.path`: path in Vault. When using V2 key-value store, you may need to inform
  `/secret/data`, while in V1 API it would be directly `/secret`;
- `name.type`: Kubernetes secret type, used by `copy` sub-command;
- `name.data.name`: file name;
- `name.data.extension`: file extension;
- `name.data.zip`: file contents is GZIP, needs to be compressed/decompressed;
- `name.data.nameAsSubPath`: employ name as final part of the Vault path `name.path`;
- `name.data.key`: employ a alternative key name on Vault;
- `name.data.fromEnv`: instead of reading file payload from file-system, you can use a environment
  variable instead. The value informed for this option is the environment variable to be read;

### File Naming Convention

On downloading files from Vault, the following name convention applies:

``` bash
${GROUP_NAME}.${FILE_NAME}.${EXTENSION}
```

Therefore, if you consider the example manifest, it would produce a file named `name.foo.txt` in
the output directory, defined as command line parameter.

## Contributing

In order to build and test `vault-hander` you will need the following:

- [Docker-Compose](https://docs.docker.com/compose/): to run Vault in the background;
- [GNU/Make](https://www.gnu.org/software/make/): to run building and testing commands;
- [Dep](https://github.com/golang/dep): to manage `vendor` folder (use `make dep` to install);
- [Vault-CLI](https://www.vaultproject.io/docs/commands/): to apply initial configuration to Vault;

### Building and Testing

Before running tests, you will need to spin up Vault in the background, and apply initial
configuration to enable [AppRole](https://www.vaultproject.io/docs/auth/approle.html) authentication,
and [secrets K/V store](https://www.vaultproject.io/docs/secrets/kv/index.html).

Additionally you need a Kubernetes cluster available, please consider
[minikube project](https://kubernetes.io/docs/setup/minikube/). During [CI](./.travis.yml),
[KinD](https://github.com/kubernetes-sigs/kind) project is used.

``` bash
docker-compose -d           # run vault in development mode
.ci/bootstrap-vault.sh      # bootstrap instance
source .env                 # import environment variables to running shell
```

And then you can:

``` bash
make bootstrap              # populate vendor
make                        # build application
make build-docker           # build docker image
make test                   # run unit-tests
make integration            # run integration-tests
```

Vault related bootstrap is needed only once, since data is kept over `.data` directory, so when you
turn on development Vault instance again, it's not mandatory anymore to boostrap.

To clean it up, you can:

``` bash
docker-compose down --volumes
rm -rf ./data
```
