# Migrant

SQL File based migrations for databases. Only tested for mysql currently. Databases and other 
settings are defined in a yaml config file (config.yaml by default). The config file can either
be in the current directory, or stored in `/etc/migrant/`.

## Commands

### Up

```bash
# migrate default database
migrant up

# migrate specific database
migrant up -d chorizo

# specify a config file
migrant up -d chorizo -c besto_configo.yaml
```

Apply all unapplied migrations to the database.

### Gen

```bash
# generate new migration for default database
migrant gen "add pickles to hamburgers"
```

Generate a new migration file. You can specify a data base or the default database if none 
is specified.

## Config File

The config file must live in the current directory or `/etc/migrant`. It should look like 
this:

```yaml
migrations: ./migrations
databases:
  hamburgers:
    driver: "mysql"
    uri: "root:secret@tcp(localhost:3306)/hamburgers?charset=utf8&parseTime=True"
    default: true
  employee:
    driver: "mysql"
    uri: "root:secret@tcp(localhost:3306)/employee?charset=utf8&parseTime=True"
```

Use a connection string for the uri. The only driver that works right now is mysql. Sorry.
The config file is pretty straight forward. You can set a database as the default by adding
a `default` key set to true.

For all commands you can `-c config_gile_name.yaml` to specify which config to use.

## Seed Files

Seeds are handled by creating yaml files that contains the desired seed information. 
The general format for the file is:

```yaml
vars:
  foobar: "hoge"

seeds:
  - table: "test_table_1"
    insert:
      - name: "test 1"

  - table: "link_table_1"
    insert:
      - test_table_id: '{{ id "test_table_1" 0 }}'
        foo: "test 1"

      - test_table_id: '{{ id "test_table_1" 0 }}'
        foo: '{{ var "foobar" }}'
```

This is pretty self explanatory, but here are the charm points: 

Each value that gets inserted into the database is essentially a small evaluated go template. 

If you define a section called `vars`, those values will be available via a `var` helper method.

As rows values get added to the database, if they have a primary id, that gets added to a list 
behind the scenes, which you can access via the `id` helper. Pass it the name of the table and 
the index of the object whose ID you want.

## Testing

Because there's lot of touching the database testing asks for a database to play with. 
There's a docker-compose.yaml file that will create a container with mysql on it. Make
sure it's running before you try testing anything.

```bash
go test ./...
```

