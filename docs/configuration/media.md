# Media

## Settings

```yaml
########################
##### MEDIA CONFIG #####
########################

# Config pertaining to media uploads (media, image descriptions, emoji).

# Size. Max size in bytes of media uploads via API.
#
# Raising this limit may cause other servers to not fetch media
# attached to a post.
#
# Examples: [2097152, 10485760, 40MB, 40MiB]
# Default: 40MiB (41943040 bytes)
media-local-max-size: 40MiB

# Size. Size in bytes of max image size referred to on /api/v_/instance endpoints,
# used by applications like Tusky to automatically scale locally uploaded media.
#
# Leaving this unset will default to media-local-max-size.
#
# Examples: [64, 500, 5MiB, 5MB, 50M]
# Default: unset
media-image-size-hint: 5MiB

# Size. Size in bytes of max video size referred to on /api/v_/instance endpoints,
# used by applications like Tusky to automatically scale locally uploaded media.
#
# Leaving this unset will default to media-local-max-size.
#
# Examples: [64, 4096, 4MiB, 4MB, 40M]
# Default: unset
media-video-size-hint: 40MiB

# Size. Max size in bytes of media to download from other instances.
#
# Lowering this limit may cause your instance not to fetch post media.
#
# Examples: [2097152, 10485760, 40MB, 40MiB]
# Default: 40MiB (41943040 bytes)
media-remote-max-size: 40MiB

# Int. Minimum amount of characters required as an image or video description.
# Examples: [500, 1000, 1500]
# Default: 0 (not required)
media-description-min-chars: 0

# Int. Maximum amount of characters permitted in an image or video description.
# Examples: [1000, 5000, 10000]
# Default: 5000
media-description-max-chars: 5000

# Size. Max size in bytes of emojis uploaded to this instance via the admin API.
#
# The default is the same as the Mastodon size limit for emojis (50kb), which allows
# for good interoperability. Raising this limit may cause issues with federation
# of your emojis to other instances, so beware.
#
# Examples: [51200, 102400, 50KB, 50KiB]
# Default: 50KiB (51200 bytes)
media-emoji-local-max-size: 50KiB

# Size. Max size in bytes of emojis to download from other instances.
#
# By default this is 100kb, or twice the size of the default for media-emoji-local-max-size.
# This strikes a good balance between decent interoperability with instances that have
# higher emoji size limits, and not taking up too much space in storage.
#
# Examples: [51200, 102400, 100KB, 100KiB]
# Default: 100KiB (102400 bytes)
media-emoji-remote-max-size: 100KiB

# Int. Number of instances of ffmpeg+ffprobe to add to the media processing pool.
#
# Increasing this number will lead to faster concurrent media processing,
# but at the cost of up to about 250MB of (spiking) memory usage per increment.
#
# You'll want to increase this number if you have RAM to spare, and/or if you're
# hosting an instance for more than 50 or so people who post/view lots of media,
# but you should leave it at 1 for single-user instances or when running GoToSocial
# in a constrained (low-memory) environment.
#
# If you set this number to 0 or less, then instead of a fixed number of instances,
# it will scale with GOMAXPROCS x 1, yielding (usually) one ffmpeg instance and one
# ffprobe instance per CPU core on the host machine.
#
# Examples: [1, 2, -1, 8]
# Default: 1
media-ffmpeg-pool-size: 1

# The below media cleanup settings allow admins to customize when and
# how often media cleanup + prune jobs run, while being set to a fairly
# sensible default (every night @ midnight). For more information on exactly
# what these settings do, with some customization examples, see the docs:
# https://docs.gotosocial.org/en/latest/admin/media_caching#cleanup

# Integer duration.
# 
# Examples: ["7 days", "1 week", "1 month"]
# Default: "7 days"
media-remote-cache-duration: "7 days"

# Cron expression (see https://crontab.guru/ for help).
# 
# Examples: ["0 0 * * *", "30 0 * * *", "0 0 * * 0"]
# Default: "0 0 * * *" (at 00:00am, every day)
media-cleanup-cron: "0 0 * * *"
```
