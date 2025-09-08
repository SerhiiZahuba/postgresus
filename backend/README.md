# Before run

Keep in mind: you need to use dev-db from docker-compose.yml in this folder
instead of postgresus-db from docker-compose.yml in the root folder.

> Copy .env.example to .env
> Copy docker-compose.yml.example to docker-compose.yml (for development only)
> Go to tools folder and install Postgres versions

# Run

To run:

> make run

To run tests:

> make test

Before commit (make sure `golangci-lint` is installed):

> make lint

# Migrations

To create migration:

> make migration-create name=MIGRATION_NAME

To run migrations:

> make migration-up

If latest migration failed:

To rollback on migration:

> make migration-down

# Swagger

To generate swagger docs:

> make swagger

Swagger URL is:

> http://localhost:4005/api/v1/docs/swagger/index.html#/

# Project structure

Default endpoint structure is:

/feature
/feature/controller.go
/feature/service.go
/feature/repository.go
/feature/model.go
/feature/dto.go

If there are couple of models:
/feature/models/model1.go
/feature/models/model2.go
...

# Project rules

Read .cursor/rules folder, it contains all the rules for the project.
