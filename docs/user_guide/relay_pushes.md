# Relay Pushes

GoToSocial allows you to create relay push connections, in order to forward your posts to relays which will then broadcast them to all instances that are connected (subscribed) to the relay. By opting in to sending your posts to one or more relays, you make it more likely that your posts will be seen by ActivityPub users who don't follow you directly and might not have encountered you otherwise.

{%
  include "../.fragments/relay.md"
  start='<!--what-is-a-relay-start-->'
  end='<!--what-is-a-relay-end-->'
%}

## What happens when you create a relay push connection?

When you create a relay push connection (see below), the service actor for your instance will send a follow request to the relay actor to indicate that you wish to take part in the relay. When the follow request is accepted, either automatically, or pending a manual action by the administrator of the relay, your instance's service actor will begin to send your posts to the inbox of the relay. The relay actor will then distribute your posts by forwarding them to relay subscribers.

{%
  include "../.fragments/relay.md"
  start='<!--relay-not-private-start-->'
  end='<!--relay-not-private-end-->'
%}

## Create a relay push connection

You can configure pushing your posts to a relay by using the [user settings panel](./settings.md) to input a relay actor URI, and then select flags that you would like to apply to the relay connection.

!!! tip "Relay connection must be approved by relay owner"
    After a relay push connection is created, you must wait for approval from the relay owner before the connection will become active. This approval may be instantaneous + automatic, or may never happen at all! Some relay admins require that you message or email them *before* creating a connection, so make sure you take account of this.

{%
  include "../.fragments/relay.md"
  start='<!--relay-actor-uri-start-->'
  end='<!--relay-actor-uri-end-->'
%}

### Flags

The flag checkboxes allow you to customize which of your posts should be pushed to a relay.

<dl>

    <dt><strong>Send public visibility posts:</strong></dt>
    <dd>By checking this flag, you instruct GoToSocial to send your Public posts to the relay. If the box is not checked, then posts you create with Public visibility will never be pushed to the relay.</dd>

    <dt><strong>Send unlisted visibility posts:</strong></dt>
    <dd>By checking this flag, you instruct GoToSocial to send your Unlisted (aka Unlocked, aka Quiet Public) posts to the relay. If the box is not checked, then posts with Unlisted visibility will never be pushed to the relay.</dd>

    <dt><strong>Never send posts marked as sensitive:</strong></dt>
    <dd>With this flag checked, GoToSocial will never push a post of yours to the relay connection if that post is marked "sensitive" by a content warning or inclusion of sensitive media attachments.</dd>

    <dt><strong>Never send posts with media:</strong></dt>
    <dd>With this flag checked, GoToSocial will never push a post of yours to the relay connection if that post include media attachments.</dd>

    <dt><strong>Never send replies/comments:</strong></dt>
    <dd>With this flag checked, GoToSocial will never send your replies to other accounts' posts to the relay, it will only send top-level posts and self-replies within a thread.</dd>

    <dt><strong>Match posts by default:</strong></dt>
    <dd>With this flag checked, you tell GoToSocial that all of your posts should be matched by default. In other words, all of your posts (of appropriate visibilities) that are not ignored because of other flags will be sent to the relay, unless their content is matched by an exclude matcher. With the flag unchecked, posts of yours will only be pushed to the relay if their content is matched by one or more matchers.</dd>

</dl>

{%
  include "../.fragments/relay.md"
  start='<!--relay-matchers-start-->'
  end='<!--relay-matchers-end-->'
%}
