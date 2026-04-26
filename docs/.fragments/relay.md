<!--what-is-a-relay-start-->
## What is a relay?

An ActivityPub relay is a software that exposes one or more ActivityPub-compatible actors, which accept posts in one or more inboxes, and then forward (ie., "Announce") those posts to actors who are subscribed to (ie., "Follow") the relay actor(s).
<!--what-is-a-relay-end-->

<!--relay-actor-uri-start-->
### Relay actor URI

A relay actor URI usually looks something like `https://relay.example.org/actor`, but for the precise URI you should check the documentation or homepage of the relay you would like to connect to.
<!--relay-actor-uri-end-->

<!--relay-matchers-start-->
## Relay matchers

Using relay matchers, you can either include or exclude posts from relaying by matching keywords (whole words or partial) in the post.

To create a relay matcher, open the detailed view of a relay connection. Below the relay update form, you will see a "Matchers" section. Here, you can view existing matchers, create a new matcher, and/or delete an existing matcher.

The matcher keyword is the phrase to be matched. For example, "sloth". Matcher keywords are case insensitive, so "Sloth", "sloth", and "SLOTH" etc are all equivalent.

If you want to match the keyword as a whole word, tick the "Match whole word" checkbox when creating a matcher. With this checkbox ticked, the matcher "sloth" will match a post with content "I saw a cool sloth today", but will *not* match with the post "I saw some cool sloths today". By contrast, with the box unticked, any post containing a word with the fragment "sloth" in it will match. So "I saw some cool sloths today" will match.

Keyword matches are made against a post's content *and* its content warning (if present).

!!! tip "Matching hashtags"
    You can also match hashtags by prefacing a keyword with a hash symbol, so you could match all posts that use the hashtag "GoToSocial" by creating a whole-word matcher with keyword `#GoToSocial`.

To create an exclude matcher, tick the checkbox "Exclude posts matched by this matcher" when creating a new matcher. When a post matches an exclude matcher, it is *never* relayed, even if other matchers would include it, or the relay connection matches all posts by default.
<!--relay-matchers-end-->

<!--relay-not-private-start-->
!!! warning "Relay connections are not always private"
    When you create a relay connection, the home page for the relay may show that your instance subscribes to that relay. Before creating a relay connection, think carefully if you want your instance to have that kind of exposure on a public page.
<!--relay-not-private-end-->