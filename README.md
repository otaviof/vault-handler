# `vault-handler`

Is a command-line tool to download and manipulate data obtained from Vault. The primary-use case
of this application is to be used as a init-container, where it's given a manifest file, and based
in this manifest, `vault-handler` will create files accordingly.

## Installing

``` bash
go get -u github.com/otaviof/vault-handler/cmd/vault-handler
```

## Manifest

The following snippet is a manifest example.

``` yaml
secrets:
  name:
    path: secret/data/dir1/dir2
    data:
      - name: foo
        extension: txt
        unzip: false
        nameAsSubPath: false
```

Description of the options used in manifest:

- `secrets`: root of the manifest;
- `name`: arbitrary group "name". This group-name is also employed to name final files;
- `name.path`: path in Vault. When using V2 key-value store, you may need to inform
  `/secret/data`, while in V1 API it would be directly `/secret`;
- `name.data.name`: file name;
- `name.data.extension`: file extension;
- `name.data.unzip`: file contents is GZIP, needs to be decompressed;
- `name.data.nameAsSubPath`: employ name as final part of the Vault path `name.path`;

### File Naming Convention

On downloading files from Vault, the following name convention applies:

``` bash
${GROUP_NAME}.${FILE_NAME}.${EXTENSION}
```

Therefore, if you consider the example manifest, it would produce a file named `name.foo.txt` in
the output directory, defined as command line parameter.






