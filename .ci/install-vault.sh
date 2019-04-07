#!/bin/bash

VAULT_VERSION="${VAULT_VERSOIN:-1.1.0}"
VAULT_TARGET_DIR="${VAULT_TARGET_DIR:-/home/travis/bin}"

VAULT_URL="https://releases.hashicorp.com"
VAULT_ZIP_FILE="vault_${VAULT_VERSION}_linux_amd64.zip"
VAULT_URL_PATH="vault/${VAULT_VERSION}/${VAULT_ZIP_FILE}"

VAULT_BIN="vault"

function die () {
    echo "[ERROR] ${*}" 1>&2
    exit 1
}

[ ! -d ${VAULT_TARGET_DIR} ] && die "Can't find target directory at '${VAULT_TARGET_DIR}'!"

if ! curl --location --output ${VAULT_ZIP_FILE} "${VAULT_URL}/${VAULT_URL_PATH}" ; then
    die "Can't download Vault!"
fi

if ! unzip ${VAULT_ZIP_FILE} ; then
    die "Can't unzip '${VAULT_ZIP_FILE}'"
fi

[ ! -f ${VAULT_BIN} ] && die "Can't find vault binary at './${VAULT_BIN}'"

if ! mv -v "${VAULT_BIN}" "${VAULT_TARGET_DIR}" ; then
    die "Can't move '${VAULT_BIN}' to '${VAULT_TARGET_DIR}'!"
fi

rm -vf "${VAULT_ZIP_FILE}" > /dev/null 2>&1
