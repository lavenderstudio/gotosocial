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
} from "../../types/relay";
import { gtsApi } from "../gts-api";

const extended = gtsApi.injectEndpoints({
	endpoints: (build) => ({
		relayPushes: build.query<RelayConnection[], void>({
			query: () => ({
				url: `/api/v1/relay_pushes`
			}),
			providesTags: (res, _error, _arg) =>
				res
					? [
						...res.map((relayPush) => ({ type: "RelayPush" as const, id: relayPush.id })),
						{ type: "RelayPush", id: "LIST" }
					]
					: [{ type: "RelayPush", id: "LIST" }]
		}),

		relayPush: build.query<RelayConnection, string>({
			query: (id) => ({
				url: `/api/v1/relay_pushes/${id}`
			}),
			providesTags: (_res, _error, id) => [{ type: "RelayPush", id }]
		}),

		createRelayPush: build.mutation<RelayConnection, RelayConnectionCreateRequest>({
			query: (formData) => {
				return {
					method: "POST",
					url: `/api/v1/relay_pushes`,
					asForm: true,
					body: formData,
					discardEmpty: true,
				};
			},
			invalidatesTags: (_res) => [{ type: "RelayPush", id: "LIST" }],
		}),

		updateRelayPush: build.mutation<RelayConnection, {id: string} & RelayConnectionUpdateRequest>({
			query: ({ id, ...formData}) => {
				return {
					method: "PUT",
					url: `/api/v1/relay_pushes/${id}`,
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
						{ type: "RelayPush", id: "LIST" },
						{ type: "RelayPush", id: res.id },
					]
					: [{ type: "RelayPush", id: "LIST" }]
		}),

		deleteRelayPush: build.mutation<RelayConnection, string>({
			query: (id) => ({
				method: "DELETE",
				url: `/api/v1/relay_pushes/${id}`
			}),
			invalidatesTags: (_res, _error, id) => [
				{ type: "RelayPush", id: "LIST" },
				{ type: "RelayPush", id }
			]
		}),

		createRelayPushMatcher: build.mutation<RelayConnection, {id: string} & RelayMatcherCreateUpdateRequest>({
			query: ({ id, ...formData}) => {
				return {
					method: "POST",
					url: `/api/v1/relay_pushes/${id}/matchers`,
					asForm: true,
					body: formData,
					discardEmpty: true,
				};
			},
			invalidatesTags: (res) =>
				res
					? [
						{ type: "RelayPush", id: "LIST" },
						{ type: "RelayPush", id: res.id },
					]
					: [{ type: "RelayPush", id: "LIST" }]
		}),

		deleteRelayPushMatcher: build.mutation<RelayConnection, { id: string, matcherID: string }>({
			query: ({ id, matcherID }) => ({
				method: "DELETE",
				url: `/api/v1/relay_pushes/${id}/matchers/${matcherID}`
			}),
			invalidatesTags: (res) =>
				res
					? [
						{ type: "RelayPush", id: "LIST" },
						{ type: "RelayPush", id: res.id },
					]
					: [{ type: "RelayPush", id: "LIST" }]
		}),
	}),
});

/**
 * Get all relay pushes.
 */
const useRelayPushesQuery = extended.useRelayPushesQuery;

/**
 * Get one relay push. 
 */
const useRelayPushQuery = extended.useRelayPushQuery;

/**
 * Create a relay push.
 */
const useCreateRelayPushMutation = extended.useCreateRelayPushMutation;

/**
 * Update a relay push.
 */
const useUpdateRelayPushMutation = extended.useUpdateRelayPushMutation;

/**
 * Delete a relay push.
 */
const useDeleteRelayPushMutation = extended.useDeleteRelayPushMutation;

/**
 * Create a relay push matcher.
 */
const useCreateRelayPushMatcherMutation = extended.useCreateRelayPushMatcherMutation;

/**
 * Delete a relay push matcher.
 */
const useDeleteRelayPushMatcherMutation = extended.useDeleteRelayPushMatcherMutation;

export {
	useRelayPushesQuery,
	useRelayPushQuery,
	useCreateRelayPushMutation,
	useUpdateRelayPushMutation,
	useDeleteRelayPushMutation,
	useCreateRelayPushMatcherMutation,
	useDeleteRelayPushMatcherMutation,
};
