# Post Caching and Pruning

GoToSocial stores posts (both local and remote) in whatever [database backend](../configuration/database.md) the instance is configured to use.

Typically, the `statuses` table that GoToSocial uses to store posts is by far the largest table in the database, and will continue to grow over time the longer you use your GoToSocial instance for, as your instance receives new posts from people you follow, and dereferences boosted posts and replied-to posts, etc.

To avoid the issue of ever-increasing database sizes, GoToSocial provides a mechanism whereby posts can be regularly pruned from the database by a background job, freeing up space.

## Which posts will be pruned

When selecting posts to clean up, GoToSocial considers posts in the context of the thread that they're part of (if applicable). For the purposes of cleanup, single posts are considered to be in a thread of length 1.

To be eligible for pruning, a thread must meet the following criteria:

1. **All posts in the thread are remote**, ie., created by a remote account. If a local account created or has participated in a thread, that thread will never be removed using this pruning method.
2. **No posts in the thread have been interacted with by a local account**. If a local user has faved, boosted, or bookmarked any post in a thread, that thread will not be pruned.
3. **All posts in the thread have been created or fetched less recently than the duration `statuses-cleanup-remote-older-than`**. Ie., only threads that haven't been dereferenced or added to since `statuses-cleanup-remote-older-than` will be pruned. (See below for more info on this config setting.)

If a thread meets the above criteria, then the cleanup process will mark that thread for removal, and all posts in the thread (and any attached media) will be removed from the database.

!!! info "Posts will be refetched on demand"
    Much like with [media caching + cleanup](./media_caching.md), if a post is removed from post pruning, it can always be fetched again by the instance later, for example if a local user looks up one of the posts in the thread by its URI/URL, or if the instance sees a post in that thread due to someone replying to it, etc.

## How to enable remote post pruning

By default, the remote post pruning task is not enabled. To enable it, update your config to set `statuses-cleanup-remote-older-than` to a valid duration string, for example "1 year", and restart your instance. Posts last created/fetched/updated before that duration will then become [eligible for cleanup](#which-posts-will-be-pruned).

## When will post pruning occur

The scheduling of the post cleanup operation is controlled by the config variable `statuses-cleanup-cron`, which accepts a cron expression as a value. By default, this is `"0 1 * * 0`, which means every Sunday at 1am (ie., weekly).

!!! info "Cron expressions"
    A ["cron expression"](https://en.wikipedia.org/wiki/Cron#Cron_expression) is a string that allows a user or computer administrator to specify when a background task should be run. Cron expressions are typically used in programming and system maintenance to schedule jobs using the program ["cron"](https://en.wikipedia.org/wiki/Cron), which is a time-based job scheduler. Because of their ubiquity, however, cron expressions are also accepted by some other programs -- such as GoToSocial! -- to allow users to customize the scheduling of background tasks.

    For more information on cron expressions and for help writing them, see the following resources:

    - [wikipedia page for cron expressions](https://en.wikipedia.org/wiki/Cron#Cron_expression)
    - [cron expression helper website](https://crontab.guru)

If you wish to this cleanup operation to run more or less frequently, depending on your needs and resources, you can adjust the config value to your liking. For example, to run the job once per month instead of once per week, you could change `statuses-cleanup-cron` to `0 1 1 * *`, which will cause the job to run at 1am on the first day of every month.

!!! tip "First cleanup can take a long time"
    If you've enabled remote post pruning for the first time on a large database (eg., larger than a few GB), be aware that the first cleanup operation can take quite a long time, as there will be a significant number of threads that must be iterated through, depending on how you've set `statuses-cleanup-remote-older-than`.
    
    For example, on a 17GB SQLite database with `statuses-cleanup-remote-older-than` set to "1 year", on a 1CPU, 1GB ram VPS, the first remote post prune operation was observed to take about 48+ hours in total.

    While the cleanup is running, depending on the specs of the machine that GoToSocial is running on, you may notice high CPU usage and some degradation of performance (eg., home timeline queries taking longer to respond, that sort of thing). This is normal and does not indicate a problem with the operation.

    After you've run a cleanup for the first time, subsequent cleanups should be much faster, as there will be fewer threads remaining that must be iterated through.

## Considerations for after cleanup

### SQLite

With SQLite, when you clean up posts from the database the database file will not shrink in size automatically, as GoToSocial does not enable [autovacuum](https://sqlite.org/pragma.html#pragma_auto_vacuum) in SQLite. Instead, freed space will be held by the database file and written into when new posts arrive.

If you want to actually shrink the size of your database file after cleanup, you should [run a `vacuum` command against it using the `sqlite3` CLI tool](./database_maintenance.md#sqlite-vacuum). To ensure indexes remain performant after removal of a large amount of statuses, you may also wish to [run `analyze` on your database file using the GoToSocial binary](./database_maintenance.md#sqlite-analyze--optimize)

### Postgres

By default, [autovacuum is enabled for Postgres](https://www.postgresql.org/docs/current/runtime-config-vacuum.html#GUC-AUTOVACUUM). This means that any space freed up by the removal of remote posts should be made available to the operating system again once the autovacuum daemon has done its thing, and you should not need to manually intervene in order to shrink your database volume.

However, it may be that you wish to run vacuum manually, and/or reanalyze the database after the removal of a large amount of posts, to ensure that indexes remain performant. For steps on how to do this, check the [Postgres maintenance docs](./database_maintenance.md#postgres).
