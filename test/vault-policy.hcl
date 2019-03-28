path "secret/*" {
  capabilities = [
    "create",
    "delete",
    "read",
    "update",
  ]
}

path "auth/approle/login" {
  capabilities = [
    "create",
    "read",
  ]
}
