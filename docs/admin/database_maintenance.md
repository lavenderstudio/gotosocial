# Database Maintenance

Regardless of whether you choose to run GoToSocial with SQLite or Postgres, you may need to occasionally take maintenance steps to keep your database running well.

!!! tip
    
    Though the maintenance tips provided here are intended to be non-destructive, you should backup your database before manually performing maintenance. That way if you mistype something or accidentally run a bad command, you can restore your backup and try again.

!!! danger
    
    Manually creating, deleting, or updating entries in your GoToSocial database is **heavily discouraged**, and such commands are not provided here. Even if you think you know what you are doing, running `DELETE` statements etc. may introduce issues that are very difficult to debug. The maintenance tips below are designed to help with the smooth running of your instance; they will not save your ass if you have manually gone into your database and hacked at entries, tables, and indexes.

## SQLite

### SQLite Vacuum

To minimize fragmentation, GoToSocial does not currently enable auto-vacuum for SQLite. To defragment the database file and repack it to an optimal size you may want to run a `VACUUM` command on your SQLite database periodically (eg., every few months).

You can see lots of information about the `VACUUM` command [here](https://sqlite.org/lang_vacuum.html).

To vacuum your GoToSocial database file, you must meet the following requirements:

- The SQLite command line tool `sqlite3` must be installed on the same machine that your GoToSocial `sqlite.db` file is stored on. For more details, see the [`sqlite3` cli docs](https://sqlite.org/cli.html).
- You must have free disk space roughly equivalent to 2x the size of your `sqlite.db` file. Eg., if you have a 10GB database file, you should make sure you have ~20GB of free space. This space is used during the vacuum to create a temporary copy of your database file, and to populate a wal file.

Once you've met these requirements, do the following:

1. Stop GoToSocial.
2. In the same directory as your GoToSocial `sqlite.db` file, run the command: `sqlite3 sqlite.db "VACUUM;"`
  This may take quite a few minutes depending on the size of your database. DO NOT INTERRUPT IT.
3. When the command has finished running, start GoToSocial again.

### SQLite Analyze / Optimize

GoToSocial runs a [full analyze command](https://sqlite.org/lang_analyze.html) after each set of database migrations (eg., when starting an updated version of GoToSocial), to ensure that any indexes added or removed by migrations are taken into account correctly by the query planner.

Following SQLite best practice, GoToSocial also runs the [`optimize` SQLite pragma](https://sqlite.org/pragma.html#pragma_optimize) when closing database connections, to help keep index information up to date. The `optimize` pragma runs `analyze` only if it's deemed to be beneficial, given the queries that the closed database connection handled.

Because of the above automated steps, in normal circumstances you should not need to run a manual `analyze` against your SQLite database file.

However, if you notice that queries are running very slowly, it could be the case that the index metadata stored in SQLite's internal tables has become out of date, or has been removed or otherwise undesirably altered, leading the query planner to make poor choices.

This is particularly prone to happening if a large cleanup operation has just occured, eg., you've just [cleaned up a lot of old statuses](./post_caching.md).

If you notice lots of timeouts, for example when trying to view your timelines or profile page, you can use the GoToSocial binary to manually run a full `analyze`.

!!! important "Use the GoToSocial binary to run `analyze`!"
    It's very important that when doing an analyze, you use the GoToSocial binary, and **not** the `sqlite3` command line tool.
    
    This is because the version of SQLite used in the GoToSocial binary has the compile-time flag `SQLITE_ENABLE_STAT4` enabled, which leads to much more thorough analysis and better query plans.
    
    If you run `analyze` using the `sqlite3` command line tool, you risk severely degrading performance for certain timeline and profile page queries!

    For more info on `SQLITE_ENABLE_STAT4`, see [the SQLite compilation flag docs](https://sqlite.org/compile.html#enable_stat4).

For a **binary** install, run the following command to do an `analyze` using the GoToSocial binary, replacing `/path/to/config.yaml` with the actual path to your `config.yaml` file:

```sh
./gotosocial --config-path /path/to/config.yaml \
    database sqlite analyze
```

When running GoToSocial from a **container**, you'll need to execute the above command inside the container instead. How to do this varies based on your container runtime, but for Docker it should look like:

```sh
docker exec -it CONTAINER_NAME_OR_ID \
    /gotosocial/gotosocial \
    database sqlite analyze
```

The command may take up to 15 minutes to finish running, depending on the size of your database file, and the specs of your machine. **DO NOT INTERRUPT IT**.

You will likely notice degraded performance of GoToSocial while the `analyze` is running, this is normal. If you prefer, you can stop GoToSocial before running the command, and start it again after running the command.

### SQLite Replication

It's a common practice to set up safeguards for your database like replication. SQLite can be replicated using external software. The basic steps are described on the [Replicating SQLite](../advanced/replicating-sqlite.md) page.

## Postgres

The commands in this section rely on having the Postgres [`psql` command line tool](https://www.postgresql.org/docs/current/app-psql.html) installed either on the same machine that is running Postgres, or inside the Postgres docker container.

### Postgres Vacuum

By default, [autovacuum is enabled for Postgres](https://www.postgresql.org/docs/current/runtime-config-vacuum.html#GUC-AUTOVACUUM). This means that unless you've configured your Postgres deployment differently from the default, you should not need to run a manual vacuum operation on your Postgres GoToSocial database. 

However, if you have autovacuum disabled, or have recently [removed a lot of entries from the database](./post_caching.md), you may wish to run a vacuum manually.

1. Ensure you have some spare disk space, about 1.5x the size of the GoToSocial database.
2. Stop GoToSocial.
3. With `psql` connected to the GoToSocial database, run the command `VACUUM FULL;`.
  This may take quite a few minutes depending on the size of your database. DO NOT INTERRUPT IT. If you want more feedback you can run `VACUUM FULL VERBOSE;` instead.
4. When the command has finished running, start GoToSocial again.

For more info, see the [Postgres docs for the `VACUUM` command](https://www.postgresql.org/docs/current/sql-vacuum.html).

### Postgres Analyze

If you notice degradation in the performance of GoToSocial on your Postgres database, it is possible that you need to use `psql` to force Postgres to rebuild statistics that the query planner uses to make decisions on which indexes to use.

To do this:

1. Ensure you have some spare disk space, about 1.5x the size of the GoToSocial database.
2. Stop GoToSocial.
3. With `psql` connected to the GoToSocial database, run the command `VACUUM FULL ANALYZE;`.
  This may take quite a few minutes depending on the size of your database. DO NOT INTERRUPT IT.
4. DO NOT SKIP THIS STEP: When the previous command has finished running, run the command `VACUUM (DISABLE_PAGE_SKIPPING ON);` to force Postgres to rebuild its [visibility map](https://www.postgresql.org/docs/current/routine-vacuuming.html#VACUUM-FOR-VISIBILITY-MAP).
  This may take quite a few minutes depending on the size of your database. DO NOT INTERRUPT IT.
5. When the second command has finished running, start GoToSocial again.

For more info on this process, see [this issue about Postgres performance](https://codeberg.org/superseriousbusiness/gotosocial/issues/4757#issuecomment-17348090).

!!! tip "Verbose"
    If you want more feedback when running either of the above commands, you can use [the `VERBOSE` option](https://www.postgresql.org/docs/current/sql-vacuum.html).

!!! warning "Postgres in Docker"
    If you're running Postgres in Docker and, while running the `VACUUM` command, you get a warning like `could not resize shared memory segment "/PostgreSQL.3206998826" to 67128672 bytes: No space left on device`, then you will need to increase the Docker parameter `shm-size` to a larger amount than the default. Try something like `--shm-size=128M` for example.
