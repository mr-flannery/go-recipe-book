# Notes

## agent workflows

- ask, plan, implement, review

## go

- having to specifically use `call` is a thing when debugging go, apparently
  ```
  call "html/template".ParseGlob("templates/*.gohtml")
  ```

## local postgres

```
docker run --name recipes-postgres -e POSTGRES_USER=local-recipe-user -e POSTGRES_PASSWORD=local-recipe-password -e POSTGRES_DB=recipe-book -p 5432:5432 -d postgres
```