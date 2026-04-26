# Relay Subscriptions

![Admin relay subscription details, showing a relay subscription and matchers.](../public/admin-settings-relay-subscription.png)

GoToSocial allows you (as admin) to create subscriptions to ActivityPub relays in order to populate your instance with posts that you may not otherwise have seen by just following other accounts across the fediverse.

!!! warning
    Relay subscriptions are useful for populating your instance with posts, but bear in mind that more posts stored on your instance means more storage space used. If you subscribe your instance to a busy relay, expect your storage requirements to increase significantly!

{%
  include "../.fragments/relay.md"
  start='<!--what-is-a-relay-start-->'
  end='<!--what-is-a-relay-end-->'
%}

## What happens when you subscribe to a relay?

When you create a relay subscription (see below), the service actor for your instance will send a follow request to the relay actor. When the follow request is accepted, either automatically, or pending a manual action by the administrator of the relay, your instance service actor will begin to receive posts from the relay in its inbox. Posts received in this way will be ingested into your instance and shown in the federated timeline (if applicable).

{%
  include "../.fragments/relay.md"
  start='<!--relay-not-private-start-->'
  end='<!--relay-not-private-end-->'
%}

## Create a relay subscription

You can subscribe to a relay by using the [admin settings panel](./settings.md#relay-subscriptions) to input the relay actor URI, and select flags that you would like to apply to the relay connection.

!!! tip "Relay connection must be approved by relay owner"
    After a relay subscription is created, you must wait for approval from the relay owner before the subscription will become active. This approval may be instantaneous + automatic, or may never happen at all! Some relay admins require that you message or email them *before* sending a subscription request, so make sure you take account of this.

{%
  include "../.fragments/relay.md"
  start='<!--relay-actor-uri-start-->'
  end='<!--relay-actor-uri-end-->'
%}

### Flags

The flag checkboxes allow you to customize which posts should be ingested by the relay subscription.

<dl>

    <dt><strong>Ingest public visibility posts:</strong></dt>
    <dd>By checking this flag, you instruct GoToSocial to ingest Public posts via this relay subscription. If the box is not checked, then posts with Public visibility will never be ingested by the relay subscription.</dd>

    <dt><strong>Ingest unlisted visibility posts:</strong></dt>
    <dd>By checking this flag, you instruct GoToSocial to ingest Unlisted (aka Unlocked, aka Quiet Public) posts via this relay subscription. If the box is not checked, then posts with Unlisted visibility will never be ingested by the relay subscription.</dd>

    <dt><strong>Never ingest posts marked as sensitive:</strong></dt>
    <dd>With this flag checked, this relay subscription will <em>never</em> ingest a post that was marked as sensitive by the author, even if such a post would normally be matched and ingested.</dd>

    <dt><strong>Never ingest posts with media:</strong></dt>
    <dd>With this flag checked, this relay subscription will <em>never</em> ingest a post that has media attachments, even if such a post would normally be matched and ingested. This can be useful to avoid ballooning an instance's storage size when it subscribes to a relay that has lots of media posts.</dd>

    <dt><strong>Never ingest replies/comments:</strong></dt>
    <dd>With this flag checked, this relay subscription will <em>never</em> ingest replies or comments that are forwarded to it by the relay. In this context, a reply or comment is a post that replies to the post of another author. Note that even if you check this flag, GoToSocial will still try to dereference comments on top-level posts or self-reply thread posts that are sent to it by the relay.</dd>

    <dt><strong>Match posts by default:</strong></dt>
    <dd>With this flag checked, you tell GoToSocial that all posts sent to it via this relay subscription should be matched by default. In other words, any posts (of appropriate visibilities) that are not ignored because of other flags will be ingested, unless their content is matched by an exclude matcher. With the flag unchecked, posts will only be ingested if their content is matched by one or more matchers.</dd>

</dl>

Any posts that should not be ingested by the relay subscription, according to the above flags, will be dropped.

{%
  include "../.fragments/relay.md"
  start='<!--relay-matchers-start-->'
  end='<!--relay-matchers-end-->'
%}
