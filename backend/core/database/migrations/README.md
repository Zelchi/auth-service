# Migrations

As migrations são aplicadas em ordem lexicográfica e registradas na tabela
`schema_migrations`. Depois de aplicada em qualquer ambiente compartilhado,
uma migration não deve ser editada; crie o próximo arquivo numerado.

Cada arquivo deve ser idempotente quando possível (`IF NOT EXISTS`) e conter
apenas mudanças compatíveis com a versão que está sendo implantada. Mudanças
destrutivas exigem uma migration separada e um backup do volume antes da
implantação.
