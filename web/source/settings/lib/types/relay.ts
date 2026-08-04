/*
	GoToSocial
	Copyright (C) GoToSocial Authors admin@gotosocial.org
	SPDX-License-Identifier: AGPL-3.0-or-later

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU Affero General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	You should have received a copy of the GNU Affero General Public License
	along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

import { Account } from "./account";

/**
 * RelayEntityCommon models fields that are common
 * across relay pushes, subscriptions, and actors.
 */
interface RelayEntityCommon {
	/**
	 * ID of this item.
	 */
	id: string;

	/**
	 * The date when this entity was created (ISO 8601 Datetime).
	 */
	created_at: string;

	/**
	 * Matchers that apply to this relay entity.
	 */
	matchers: RelayMatcher[];

	/**
	 * Include public posts when relaying.
	 */
	public: boolean;

	/**
	 * Include unlisted/unlocked posts when relaying.
	 */
	unlisted: boolean;

	/**
	 * Match all posts by default.
	 */
	match_by_default: boolean;

	/**
	 * Ignore sensitive posts when relaying.
	 */
	ignore_sensitive: boolean;

	/**
	 * Ignore posts with media attachments when relaying.
	 */
	ignore_media: boolean;

	/**
	 * Ignore replies to other accounts when relaying.
	 */
	ignore_replies: boolean;
}

/**
 * RelayConnection models a relay push or relay subscription targeting a relay actor.
 */
export interface RelayConnection extends RelayEntityCommon {
	/**
	 * ID of the account that created this relay connection.
	 * Will only be set for relay subscriptions, not relay pushes.
	 */
	account_id?: string;

	/**
	 * ActivityPub URI of the relay service actor.
	 */
	relay_actor_uri: string;

	/**
	 * True if this relay connection has been approved by the relay actor.
	 */
	approved: boolean;
}

/**
 * RelayFlagsUpdateRequest models an update request for modifying a relay entity's flags.
 */
export interface RelayFlagsUpdateRequest {
	/**
	 * Include public posts when relaying.
	 */
	public?: boolean;

	/**
	 * Include unlisted/unlocked posts when relaying.
	 */
	unlisted?: boolean;

	/**
	 * Ignore sensitive posts when relaying.
	 */
	ignore_sensitive?: boolean;

	/**
	 * Ignore posts with media attachments when relaying.
	 */
	ignore_media?: boolean;

	/**
	 * Ignore replies to other accounts when relaying.
	 */
	ignore_replies?: boolean;
}

/**
 * RelayConnectionCreateRequest models an create request for a relay push or relay subscription.
 */
export interface RelayConnectionCreateRequest extends RelayFlagsUpdateRequest {
	/**
	 * ActivityPub URI of the relay service actor.
	 */
	relay_actor_uri: string;
}

/**
 * RelayMatcher models a relay matcher used to filter what is + isn't pushed / subscribed to by a relay connection.
 */
export interface RelayMatcher {
	/**
	 * ID of this item.
	 */
	id: string;

	/**
	 * The text to be matched.
	 */
	keyword: string;

	/**
	 * Consider word boundaries when matching.
	 */
	whole_word: boolean;

	/**
	 * If true, this relay matcher will cause matches to be EXCLUDED from relaying rather than INCLUDED in relaying.
	 */
	exclude: boolean;
}

/**
 * RelayMatcherCreateUpdateRequest models a request to create or update a relay matcher for a relay connection.
 */
export interface RelayMatcherCreateUpdateRequest {
	/**
	 * The text to be matched.
	 */
	keyword?: string;

	/**
	 * Consider word boundaries when matching.
	 */
	whole_word?: boolean;

	/**
	 * If true, this relay matcher will cause matches to be EXCLUDED from relaying rather than INCLUDED in relaying.
	 */
	exclude?: boolean;
} 

/**
 * RelayActor models a local relay actor created by an admin on this instance.
 */
export interface RelayActor extends RelayConnection {
	/**
	 * Relay actor account model.
	 */
	account: Account;

	/**
	 * ID of the admin account that
	 * created this relay actor.
	 */
	created_by_account_id: string;
}

/**
 * RelayActorCreateRequest models a request to create a relay actor.
 */
export interface RelayActorCreateRequest extends RelayActorUpdateRequest {
	/**
	 * The desired username for the relay actor account.
	 * Will be prefixed with "relay." to form the final
	 * username. Eg., "username=example" results in
	 * username "relay.example".
	 */
	username: string;
}

/**
 * RelayActorUpdateRequest models a request to update a relay actor.
 */
export interface RelayActorUpdateRequest extends RelayFlagsUpdateRequest {
	/**
	 * Relay actor account should be made discoverable
	 * and shown in the profile directory (if enabled).
	 */
	discoverable?: boolean;

	/**
	 * The display name to use for the account.
	 */
	display_name?: string;

	/**
	 * Bio/description of this account.
	 */
	note?: string;

	/**
	 * Relay account avatar.
	 */
	avatar;

	/**
	 * Description of avatar image, for alt-text.
	 */
	avatar_description?: string;

	/**
	 * Relay account header.
	 */
	header;

	/**
	 * Description of header image, for alt-text.
	 */
	header_description?: string;

	/**
	 * Require manual approval of follow requests.
	 */
	locked?: boolean;

	fields_attributes;
}
