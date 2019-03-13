path "secret/*" {
  capabilities = [ "create", "read", "update" ]
}

path "auth/approle/login" {
  capabilities = [ "create", "read" ]
}
