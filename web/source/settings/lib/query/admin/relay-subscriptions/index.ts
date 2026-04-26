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

import type {
	RelayConnection,
	RelayConnectionCreateRequest,
	RelayConnectionUpdateRequest,
	RelayMatcherCreateUpdateRequest,
} from "../../../types/relay";
import { gtsApi } from "../../gts-api";

const extended = gtsApi.injectEndpoints({
	endpoints: (build) => ({
		relaySubscriptions: build.query<RelayConnection[], void>({
			query: () => ({
				url: `/api/v1/admin/relay_subscriptions`
			}),
			providesTags: (res, _error, _arg) =>
				res
					? [
						...res.map((relaySubscription) => ({ type: "RelaySubscription" as const, id: relaySubscription.id })),
						{ type: "RelaySubscription", id: "LIST" }
					]
					: [{ type: "RelaySubscription", id: "LIST" }]
		}),

		relaySubscription: build.query<RelayConnection, string>({
			query: (id) => ({
				url: `/api/v1/admin/relay_subscriptions/${id}`
			}),
			providesTags: (_res, _error, id) => [{ type: "RelaySubscription", id }]
		}),

		createRelaySubscription: build.mutation<RelayConnection, RelayConnectionCreateRequest>({
			query: (formData) => {
				return {
					method: "POST",
					url: `/api/v1/admin/relay_subscriptions`,
					asForm: true,
					body: formData,
					discardEmpty: true,
				};
			},
			invalidatesTags: (_res) => [{ type: "RelaySubscription", id: "LIST" }],
		}),

		updateRelaySubscription: build.mutation<RelayConnection, {id: string} & RelayConnectionUpdateRequest>({
			query: ({ id, ...formData}) => {
				return {
					method: "PUT",
					url: `/api/v1/admin/relay_subscriptions/${id}`,
					asForm: true,
					body: formData,
					// Don't discardEmpty when updating, as we
					// want to be able to set "false" booleans.
					discardEmpty: false,
				};
			},
			invalidatesTags: (res) =>
				res
					? [
						{ type: "RelaySubscription", id: "LIST" },
						{ type: "RelaySubscription", id: res.id },
					]
					: [{ type: "RelaySubscription", id: "LIST" }]
		}),

		deleteRelaySubscription: build.mutation<RelayConnection, string>({
			query: (id) => ({
				method: "DELETE",
				url: `/api/v1/admin/relay_subscriptions/${id}`
			}),
			invalidatesTags: (_res, _error, id) => [
				{ type: "RelaySubscription", id: "LIST" },
				{ type: "RelaySubscription", id }
			]
		}),

		createRelaySubscriptionMatcher: build.mutation<RelayConnection, {id: string} & RelayMatcherCreateUpdateRequest>({
			query: ({ id, ...formData}) => {
				return {
					method: "POST",
					url: `/api/v1/admin/relay_subscriptions/${id}/matchers`,
					asForm: true,
					body: formData,
					discardEmpty: true,
				};
			},
			invalidatesTags: (res) =>
				res
					? [
						{ type: "RelaySubscription", id: "LIST" },
						{ type: "RelaySubscription", id: res.id },
					]
					: [{ type: "RelaySubscription", id: "LIST" }]
		}),

		deleteRelaySubscriptionMatcher: build.mutation<RelayConnection, { id: string, matcherID: string }>({
			query: ({ id, matcherID }) => ({
				method: "DELETE",
				url: `/api/v1/admin/relay_subscriptions/${id}/matchers/${matcherID}`
			}),
			invalidatesTags: (res) =>
				res
					? [
						{ type: "RelaySubscription", id: "LIST" },
						{ type: "RelaySubscription", id: res.id },
					]
					: [{ type: "RelaySubscription", id: "LIST" }]
		}),
	}),
});

/**
 * Get all relay subscriptions.
 */
const useRelaySubscriptionsQuery = extended.useRelaySubscriptionsQuery;

/**
 * Get one relay subscription. 
 */
const useRelaySubscriptionQuery = extended.useRelaySubscriptionQuery;

/**
 * Create a relay subscription.
 */
const useCreateRelaySubscriptionMutation = extended.useCreateRelaySubscriptionMutation;

/**
 * Update a relay subscription.
 */
const useUpdateRelaySubscriptionMutation = extended.useUpdateRelaySubscriptionMutation;

/**
 * Delete a relay subscription.
 */
const useDeleteRelaySubscriptionMutation = extended.useDeleteRelaySubscriptionMutation;

/**
 * Create a relay subscription matcher.
 */
const useCreateRelaySubscriptionMatcherMutation = extended.useCreateRelaySubscriptionMatcherMutation;

/**
 * Delete a relay subscription matcher.
 */
const useDeleteRelaySubscriptionMatcherMutation = extended.useDeleteRelaySubscriptionMatcherMutation;

export {
	useRelaySubscriptionsQuery,
	useRelaySubscriptionQuery,
	useCreateRelaySubscriptionMutation,
	useUpdateRelaySubscriptionMutation,
	useDeleteRelaySubscriptionMutation,
	useCreateRelaySubscriptionMatcherMutation,
	useDeleteRelaySubscriptionMatcherMutation,
};
