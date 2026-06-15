# Media Caching and Pruning

GoToSocial uses the configured [storage backend](https://docs.gotosocial.org/en/latest/configuration/storage/) in order to store media (images, videos, etc) uploaded to the instance by local users, as well as to cache media attached to posts and profiles federated in from remote instances.

Media uploaded by local instance users will be kept in storage forever (unless the post or profile it's attached to is deleted), so that it's always available to be served in response to requests coming from remote instances.

Remote media, on the other hand, is cached only temporarily. After a certain amount of time (see below), it will be removed from storage to help alleviate storage space usage. Remote media uncached this way will be re-fetched automatically from the remote instance if it's needed again.

!!! info "Why cache?"
    There is an argument to be made for not caching remote media at all, since it's always available on the origin server. Why not just forego caching entirely, and rely on the remote instance to serve everything on demand?
    
    While this is a straightforward approach to saving storage space, it can cause other problems and is generally considered to be rather impolite.
    
    For example, say someone from a small instance makes a funny post with an image attached. The post gets boosted by an account that's followed by 1,000 people across 5 different instances (200 on each instance). Each of those 1,000 people then have the image put in their timeline at once.
    
    With no remote media caching in place, this may cause up to 1,000 requests to hit the small instance simultaneously, as the browser of each recipient of the post must go and make a unique request to fetch the image from the small instance. This causes a large traffic spike for the small instance. In extreme scenarios, this can cause the instance to become unresponsive or crash, essentially DDOS'ing it.
    
    With remote media caching in place, however, boosting a post to 1,000 people across 5 different instances will cause only 5 requests to the small instance: 1 request for each instance. Each instance will then serve 200 requests to its local users from the cached version of the remote image, effectively spreading the load and sparing the smaller instance.

## Cleanup

Cleanup of the remote media cache occurs as a scheduled background process, and no manual intervention is required by admins. Cleanup takes somewhere between 5-30 minutes depending on the speed of the server, the speed of the configured storage, and the amount of media to work through.

GoToSocial exposes two variables that let you, the admin, tune when and how this work is performed: `media-cleanup-cron` (accepts a cron expression), and `media-remote-cache-duration` (accepts a human-language duration string).

!!! info "Cron expressions"
    A ["cron expression"](https://en.wikipedia.org/wiki/Cron#Cron_expression) is a string that allows a user or computer administrator to specify when a background task should be run. Cron expressions are typically used in programming and system maintenance to schedule jobs using the program ["cron"](https://en.wikipedia.org/wiki/Cron), which is a time-based job scheduler. Because of their ubiquity, however, cron expressions are also accepted by some other programs -- such as GoToSocial! -- to allow users to customize the scheduling of background tasks.

    For more information on cron expressions and for help writing them, see the following resources:

    - [wikipedia page for cron expressions](https://en.wikipedia.org/wiki/Cron#Cron_expression)
    - [cron expression helper website](https://crontab.guru)

By default, these variables are set to the following values:

| Variable name                 | Default      | Meaning                                        |
|-------------------------------|--------------|------------------------------------------------|
| `media-cleanup-cron`          | `0 0 * * *`  | Cron expression meaning every night @ midnight |
| `media-remote-cache-duration` | `7 days`     | 7 days                                         |

In other words, the default settings mean that every night at midnight, remote media older than seven days will be uncached and removed from storage.

You can achieve different results by tuning these variables. For example, say you wanted to prune at 4.30am instead of midnight, you could change `media-cleanup-cron` to `30 4 * * *`.

If you only want to prune every two days instead of every night, you could set `media-cleanup-cron` to something like `0 0 */2 * *`

If you wanted to adopt an aggressive cleanup strategy to minimize storage usage, you could set the following values:

| Variable name                 | Setting       | Meaning     |
|-------------------------------|---------------|-------------|
| `media-cleanup-cron`          | `0 */8 * * *` | every 8 hrs |
| `media-remote-cache-duration` | `1`           | 1 day       |

The above settings would mean that every 8 hours, GoToSocial would prune any media older than 1 day (24hrs). With this configuration, the longest amount of time you could possibly keep remote media in your storage would be about 32 hours.

!!! tip
    Setting `media-remote-cache-duration` to 0 or less means that remote media will never be uncached. However, cleanup jobs for orphaned local media and other consistency checks will still be run using the schedule defined by the other variables.

!!! tip
    You can also run cleanup manually as a one-off action through the admin panel, if you so wish ([see docs](./settings.md#media)).

!!! warning
    Setting `media-cleanup-cron` to a very small value like every hour or less will probably cause your instance to just constantly iterate through attachments, causing high database use for very little benefit. We don't recommend setting this value to less than about every eight hours, and even that is probably overkill.
