path "secret/*" {
  capabilities = [ "read" ]
}

path "auth/approle/login" {
  capabilities = [ "create", "read" ]
}