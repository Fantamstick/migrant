# Migrant

SQL File based migrations for databases. Only tested for mysql currently. Databases and other settings are defined in a yaml config file (config.yaml by default). The config file can either be in the current directory, or stored in `/etc/migrant/`.

## Config

### Describing databases

A minimal configuration file looks like this:

```yaml
migrations: ./migrations
databases:
    hamburgers:
        driver: mysql
        default: true
        uri: "admin:radpassword@tcp(hamburgers.net:3306)/hamburgers?charset=utf8&parseTime=True&multiStatements=true"
```

This would allow you to run the migrations stored in `./migrations/hamburger` on the hamburgers database, which is described by the uri.

It is possible to describe more than one database and specify which one you will connect with using `-d`. You can also describe databases using credentials instead of a uri, which can be useful if you need to split things up.

```yaml
# this is equivilant to the above config, just larger
databases:
    hamburgers:
        driver: mysql
        default: true
        user: "admin"
        pass: "radpassword"
        host: "hamburgers.net"
        port: "3306"
        prms: "charset=utf8&parseTime=True&multiStatements=true"
```

### Using a jump host (bastion)

If you set the `port_forward` setting to true for a database, you can tell migrant to use port forwarding to connect to your database. This is useful if you keep your services behind a jump host and cannot connect to them directly. The details for the port forwarding must be described in an `ssh` block.

```yaml
# migrant will connec to this databse using port forwarding 
databases:
    hamburgers:
        driver: mysql
        default: true
        uri: "admin:radpassword@tcp(localhost:33061)/hamburgers?charset=utf8&parseTime=True&multiStatements=true"
        port_forward: true
        ssh:
            username: ec2-user
            identity: "./identity-file.pem"
            local_uri: localhost:33061
            jump_uri: bastion.hamburgers.net:22
            remote_uri: 10.10.10.10

            # note - you can also use uri components to describe a any of these uris
            remote_host: 10.10.10.10
            remote_port: 3306
```

The above config will use the jump host to create a tunnel between the remote host and the local host. Note that in these cases you must tell migrant to connect to the localhost in your databse uri.

### Retriving secrets

You can set parameter variables for connections as secrets, which are not saved in your config. The only drivers that currentky work for secret retreival are `aws-secretsmanager` and `json`.

To use secrets, put a secrets block in your config, and refer to the values you want to keep secret as `SECRET://` uris.

```yaml
migrations: ./migrations
databases:
    hamburgers:
        driver: mysql
        default: true
        uri: "SECRET://aws/hamburger_db_uri"

secrets:
    aws:
        driver: "aws-secretsmanager"
        uri: "secret1?region=ap-northeast-1"
```

For aws-secretsmanager, the uri is the name of the secret, then the region where the secret is stored as a parameter. The uri for each secret is the name of the secret block that holds the secret, and then the name of the key inside of the secret. It is important to note that **all secrets must be stored as strings**.

## Commands

### Gen

```bash
# generate new migration for default database
migrant gen "add pickles to hamburgers"
```

Generate a new migration file. You can specify a data base or the default database if none is specified.


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

### Seed

```bash
# seed with a specific file
migrant seed "seeds/dev_only.yaml"

# seed using multiple files
migrant seed "seeds/always_include.yaml" "seeds/dev_only.yaml"
```

Seeds the database using values from one or more yaml files. See the section on Seed Files below for more information how to write seed files.

### Reset

```bash
# drop all tables and reapply all migrations
migrate reset
```

Drops all tables and reapplies all migrations. This has the obvious consequence that it will **destroy all data** so seriously don't do it in production. Not even as a joke.

### Truncate

```bash
# truncate all data from database
migrant truncate
```

truncates all tables in the database except for the migration table. Just like resetting the database, this **destroys all your data**, obviously, so be careful.

## Config File

The config file must live in the current directory or `/etc/migrant`. It should look like this:

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

Use a connection string for the uri. The only driver that works right now is mysql. Sorry. The config file is pretty straight forward. You can set a database as the default by adding a `default` key set to true.

For all commands you can `-c config_file_name.yaml` to specify which config to use.

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

As rows values get added to the database, if they have a primary id, that gets added to a list behind the scenes, which you can access via the `id` helper. Pass it the name of the table and the index of the object whose ID you want.

## Testing

Because there's lot of touching the database testing asks for a database to play with. There's a docker-compose.yaml file that will create a container with mysql on it. Make sure it's running before you try testing anything.

```bash
go test ./...
```

