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

/**
 * RelayConnection models a relay push or relay subscription targeting a relay actor.
 */
export interface RelayConnection {
	/**
	 * ID of this item.
	 */
	id: string;

	/**
	 * The date when this relay connection was created (ISO 8601 Datetime).
	 */
	created_at: string;

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
	 * Matchers that apply to this relay connection.
	 */
	matchers: RelayMatcher[];

	/**
	 * True if this relay connection has been approved by the relay actor.
	 */
	approved: boolean;

	/**
	 * Include public posts when relaying via this connection.
	 */
	public: boolean;

	/**
	 * Include unlisted/unlocked posts when relaying via this connection.
	 */
	unlisted: boolean;

	
	match_by_default: boolean;

	/**
	 * Exclude sensitive posts when relaying via this connection.
	 */
	ignore_sensitive: boolean;

	/**
	 * Exclude posts with media attachments when relaying via this connection.
	 */
	ignore_media: boolean;

	/**
	 * Exclude replies to other accounts when relaying via this connection.
	 */
	ignore_replies: boolean;
}

/**
 * RelayConnectionUpdateRequest models an update request for a relay push or relay subscription.
 */
export interface RelayConnectionUpdateRequest {
	/**
	 * Include public posts when relaying via this connection.
	 */
	include_public?: boolean;

	/**
	 * Include unlisted/unlocked posts when relaying via this connection.
	 */
	include_unlisted?: boolean;

	/**
	 * Exclude sensitive posts when relaying via this connection.
	 */
	exclude_sensitive?: boolean;

	/**
	 * Exclude posts with media attachments when relaying via this connection.
	 */
	exclude_media?: boolean;

	/**
	 * Exclude replies to other accounts when relaying via this connection.
	 */
	exclude_replies?: boolean;
}

/**
 * RelayConnectionCreateRequest models an create request for a relay push or relay subscription.
 */
export interface RelayConnectionCreateRequest extends RelayConnectionUpdateRequest {
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
