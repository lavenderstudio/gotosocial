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

import { AnyAction } from "redux";
import { Account } from "../../../types/account";
import { RelayActor, RelayActorCreateRequest, RelayActorUpdateRequest, RelayMatcherCreateUpdateRequest } from "../../../types/relay";
import { gtsApi } from "../../gts-api";
import { ThunkDispatch } from "@reduxjs/toolkit";
import { MaybeDrafted, PatchCollection } from "@reduxjs/toolkit/dist/query/core/buildThunks";
import { NoArg } from "../../../types/query";

const extended = gtsApi.injectEndpoints({
	endpoints: (build) => ({
		relayActors: build.query<RelayActor[], void>({
			query: () => ({
				url: `/api/v1/admin/relay_actors`
			}),
			providesTags: (res, _error, _arg) =>
				res
					? [
						...res.map((relayActor) => ({ type: "RelayActor" as const, id: relayActor.id })),
						{ type: "RelayActor", id: "LIST" }
					]
					: [{ type: "RelayActor", id: "LIST" }]
		}),

		relayActor: build.query<RelayActor, string>({
			query: (id) => ({
				url: `/api/v1/admin/relay_actors/${id}`
			}),
			providesTags: (_res, _error, id) => [{ type: "RelayActor", id }]
		}),

		createRelayActor: build.mutation<RelayActor, RelayActorCreateRequest>({
			query: (formData) => {
				return {
					method: "POST",
					url: `/api/v1/admin/relay_actors`,
					asForm: true,
					body: formData,
					discardEmpty: true,
				};
			},
			invalidatesTags: (_res) => [{ type: "RelayActor", id: "LIST" }],
		}),

		updateRelayActor: build.mutation<RelayActor, {id: string} & RelayActorUpdateRequest>({
			query: ({ id, ...formData}) => {
				return {
					method: "PUT",
					url: `/api/v1/admin/relay_actors/${id}`,
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
						{ type: "RelayActor", id: "LIST" },
						{ type: "RelayActor", id: res.id },
					]
					: [{ type: "RelayActor", id: "LIST" }]
		}),

		deleteRelayActorHeader: build.mutation<RelayActor, string>({
			query: (id) => ({
				method: "DELETE",
				url: `/api/v1/admin/relay_actors/${id}/profile/header`,
			}),
			invalidatesTags: (_res, _error, id) => [
				{ type: "RelayActor", id: "LIST" },
				{ type: "RelayActor", id }
			]
		}),

		deleteRelayActorAvatar: build.mutation<RelayActor, string>({
			query: (id) => ({
				method: "DELETE",
				url: `/api/v1/admin/relay_actors/${id}/profile/avatar`,
			}),
			invalidatesTags: (_res, _error, id) => [
				{ type: "RelayActor", id: "LIST" },
				{ type: "RelayActor", id }
			]
		}),

		deleteRelayActor: build.mutation<RelayActor, string>({
			query: (id) => ({
				method: "DELETE",
				url: `/api/v1/admin/relay_actors/${id}`
			}),
			invalidatesTags: (_res, _error, id) => [
				{ type: "RelayActor", id: "LIST" },
				{ type: "RelayActor", id }
			]
		}),

		relayActorFollowers: build.query<Account[], string>({
			query: (id) => ({
				url: `/api/v1/admin/relay_actors/${id}/followers`
			}),
			providesTags: (_res, _error, id) => [
				{ type: "Account", id: "FOLLOWING-" + id }
			]
		}),

		relayActorRemoveFromFollowers: build.mutation<any, {relayActorID: string, accountID: string}>({
			query: ({ relayActorID, accountID }) => {
				return {
					method: "POST",
					url: `/api/v1/admin/relay_actors/${relayActorID}/accounts/${accountID}/remove_from_followers`,
				};
			},
			async onQueryStarted({ relayActorID, accountID }, { dispatch, queryFulfilled }) {
				// Prepare patches.
				const patches: PatchCollection[] = [];

				// Remove the account from relay's followers.
				const removeBlockPatch = removeAccountFromCollection(
					"relayActorFollowers",
					relayActorID,
					accountID,
					dispatch,
				);
				patches.push(removeBlockPatch);

				// Update stats of individual
				// relay actor account.
				const statsPatch1 = dispatch(
					extended.util.updateQueryData(
						"relayActor",
						relayActorID,
						draft => {
							if (draft.account.source) {
								// Decrement followers.
								draft.account.followers_count--;
							}
						}
					)
				);
				patches.push(statsPatch1);

				// Update stats of relay
				// actor account within list.
				const statsPatch2 = dispatch(
					extended.util.updateQueryData(
						"relayActors",
						NoArg,
						draft => {
							const relayActor = draft.find(ra => ra.id == relayActorID);
							if (relayActor) {
								if (relayActor.account.source) {
									// Decrement followers.
									relayActor.account.followers_count--;
								}
							}
						}
					)
				);
				patches.push(statsPatch2);

				// Revert optimistic update
				// patches if query fails.
				try {
					await queryFulfilled;
				} catch {
					patches.forEach(p => p.undo());
				}
			},
		}),

		relayActorFollowRequests: build.query<Account[], string>({
			query: (id) => ({
				url: `/api/v1/admin/relay_actors/${id}/follow_requests`
			}),
			providesTags: (_res, _error, id) => [
				{ type: "Account", id: "FOLLOW-REQUESTING-" + id }
			]
		}),

		acceptRelayActorFollowRequest: build.mutation<any, {relayActorID: string, accountID: string}>({
			query: ({ relayActorID, accountID }) => {
				return {
					method: "POST",
					url: `/api/v1/admin/relay_actors/${relayActorID}/follow_requests/${accountID}/authorize`,
				};
			},
			async onQueryStarted({ relayActorID, accountID }, { dispatch, queryFulfilled, getState }) {
				// Prepare patches.
				const patches: PatchCollection[] = [];

				// Remove the accepted account from relay's follow requests.
				const removeFollowRequestPatch = removeAccountFromCollection(
					"relayActorFollowRequests",
					relayActorID,
					accountID,
					dispatch,
				);
				patches.push(removeFollowRequestPatch);

				// Get account from follow requests state if possible.
				const stateKey = `relayActorFollowRequests("${relayActorID}")`;
				const account = accountFromQueriesState(getState(), stateKey, accountID);

				if (account) {
					// Account was cached in follow requests,
					// add accepted account to followers directly
					const addToFollowersPatch = prependAccountToCollection(
						"relayActorFollowers",
						relayActorID,
						account,
						dispatch,
					);
					patches.push(addToFollowersPatch);

				} else {
					// We didn't have account cached,
					// just invalidate followers instead.
					dispatch(extended.util.invalidateTags([
						{type: "Account", id: "FOLLOWING-" + relayActorID},
					]));
				}

				// Update stats of individual
				// relay actor account.
				const statsPatch1 = dispatch(
					extended.util.updateQueryData(
						"relayActor",
						relayActorID,
						draft => {
							if (draft.account.source) {
								// Decrement follow requests.
								draft.account.source.follow_requests_count--;
							}
							// Increment followers.
							draft.account.followers_count++;
						}
					)
				);
				patches.push(statsPatch1);

				// Update stats of relay
				// actor account within list.
				const statsPatch2 = dispatch(
					extended.util.updateQueryData(
						"relayActors",
						NoArg,
						draft => {
							const relayActor = draft.find(ra => ra.id == relayActorID);
							if (relayActor) {
								// Decrement follow requests.
								if (relayActor.account.source) {
									relayActor.account.source.follow_requests_count--;
								}
								// Increment followers.
								relayActor.account.followers_count++;
							}
						}
					)
				);
				patches.push(statsPatch2);

				// Revert optimistic update
				// patches if query fails.
				try {
					await queryFulfilled;
				} catch {
					patches.forEach(p => p.undo());
				}
			},
		}),

		rejectRelayActorFollowRequest: build.mutation<any, {relayActorID: string, accountID: string}>({
			query: ({ relayActorID, accountID }) => {
				return {
					method: "POST",
					url: `/api/v1/admin/relay_actors/${relayActorID}/follow_requests/${accountID}/reject`,
				};
			},
			async onQueryStarted({ relayActorID, accountID }, { dispatch, queryFulfilled }) {
				// Prepare patches.
				const patches: PatchCollection[] = [];

				// Remove the rejected account from relay's follow requests.
				const removeFollowRequestPatch = removeAccountFromCollection(
					"relayActorFollowRequests",
					relayActorID,
					accountID,
					dispatch,
				);
				patches.push(removeFollowRequestPatch);

				// Update stats of individual
				// relay actor account.
				const statsPatch1 = dispatch(
					extended.util.updateQueryData(
						"relayActor",
						relayActorID,
						draft => {
							if (draft.account.source) {
								// Decrement follow requests.
								draft.account.source.follow_requests_count--;
							}
						}
					)
				);
				patches.push(statsPatch1);

				// Update stats of relay
				// actor account within list.
				const statsPatch2 = dispatch(
					extended.util.updateQueryData(
						"relayActors",
						NoArg,
						draft => {
							const relayActor = draft.find(ra => ra.id == relayActorID);
							if (relayActor) {
								// Decrement follow requests.
								if (relayActor.account.source) {
									relayActor.account.source.follow_requests_count--;
								}
							}
						}
					)
				);
				patches.push(statsPatch2);

				// Revert optimistic update
				// patches if query fails.
				try {
					await queryFulfilled;
				} catch {
					patches.forEach(p => p.undo());
				}
			},
		}),

		relayActorBlocks: build.query<Account[], string>({
			query: (id) => ({
				url: `/api/v1/admin/relay_actors/${id}/blocks`
			}),
			providesTags: (_res, _error, id) => [
				{ type: "Account", id: "BLOCKED-BY-" + id }
			]
		}),

		relayActorCreateBlock: build.mutation<any, {relayActorID: string, accountID: string}>({
			query: ({ relayActorID, accountID }) => {
				return {
					method: "POST",
					url: `/api/v1/admin/relay_actors/${relayActorID}/accounts/${accountID}/block`,
				};
			},
			async onQueryStarted({ relayActorID, accountID }, { dispatch, queryFulfilled, getState }) {
				// Prepare patches.
				const patches: PatchCollection[] = [];

				// Try to get account from follow requests state if possible.
				var stateKey = `relayActorFollowRequests("${relayActorID}")`;
				const followRequester = accountFromQueriesState(getState(), stateKey, accountID);

				// Remove the blocked account from relay's follow requests.
				const removeFollowRequestPatch = removeAccountFromCollection(
					"relayActorFollowRequests",
					relayActorID,
					accountID,
					dispatch,
				);
				patches.push(removeFollowRequestPatch);

				// Try to get account from followers state if possible.
				stateKey = `relayActorFollowers("${relayActorID}")`;
				const follower = accountFromQueriesState(getState(), stateKey, accountID);

				// Remove the blocked account from relay's followers.
				const removeFollowerPatch = removeAccountFromCollection(
					"relayActorFollowers",
					relayActorID,
					accountID,
					dispatch,
				);
				patches.push(removeFollowerPatch);
				
				// See if we got an account.
				const account = followRequester ? followRequester : follower;
				if (account) {
					// Account was cached, add it to blocks directly.
					const addToBlocksPatch = prependAccountToCollection(
						"relayActorBlocks",
						relayActorID,
						account,
						dispatch,
					);
					patches.push(addToBlocksPatch);

				} else {
					// We didn't have account cached,
					// just invalidate blocks instead.
					dispatch(extended.util.invalidateTags([
						{type: "Account", id: "BLOCKED-BY-" + relayActorID},
					]));
				}

				// Invalidate relay actor so stats have to
				// be refetched, as we don't necessarily
				// know if the blocked account was a follow
				// requester / follower; it depends on whether
				// calls to those endpoints had recently been
				// cached, which we can't predict.
				dispatch(extended.util.invalidateTags([
					{type: "RelayActor", id: relayActorID},
					{type: "RelayActor", id: "LIST"},
				]));

				// Revert optimistic update
				// patches if query fails.
				try {
					await queryFulfilled;
				} catch {
					patches.forEach(p => p.undo());
				}
			},
		}),

		relayActorRemoveBlock: build.mutation<any, {relayActorID: string, accountID: string}>({
			query: ({ relayActorID, accountID }) => {
				return {
					method: "POST",
					url: `/api/v1/admin/relay_actors/${relayActorID}/accounts/${accountID}/unblock`,
				};
			},
			async onQueryStarted({ relayActorID, accountID }, { dispatch, queryFulfilled }) {
				// Prepare patches.
				const patches: PatchCollection[] = [];

				// Remove the unblocked account from relay's blocks.
				const removeBlockPatch = removeAccountFromCollection(
					"relayActorBlocks",
					relayActorID,
					accountID,
					dispatch,
				);
				patches.push(removeBlockPatch);

				// Update stats of individual
				// relay actor account.
				const statsPatch1 = dispatch(
					extended.util.updateQueryData(
						"relayActor",
						relayActorID,
						draft => {
							if (draft.account.source) {
								// Decrement blocks.
								draft.account.source.blocks_count--;
							}
						}
					)
				);
				patches.push(statsPatch1);

				// Update stats of relay
				// actor account within list.
				const statsPatch2 = dispatch(
					extended.util.updateQueryData(
						"relayActors",
						NoArg,
						draft => {
							const relayActor = draft.find(ra => ra.id == relayActorID);
							if (relayActor) {
								if (relayActor.account.source) {
									// Decrement blocks.
									relayActor.account.source.blocks_count--;
								}
							}
						}
					)
				);
				patches.push(statsPatch2);

				// Revert optimistic update
				// patches if query fails.
				try {
					await queryFulfilled;
				} catch {
					patches.forEach(p => p.undo());
				}
			},
		}),

		createRelayActorMatcher: build.mutation<RelayActor, {id: string} & RelayMatcherCreateUpdateRequest>({
			query: ({ id, ...formData}) => {
				return {
					method: "POST",
					url: `/api/v1/admin/relay_actors/${id}/matchers`,
					asForm: true,
					body: formData,
					discardEmpty: true,
				};
			},
			invalidatesTags: (res) =>
				res
					? [
						{ type: "RelayActor", id: "LIST" },
						{ type: "RelayActor", id: res.id },
					]
					: [{ type: "RelayActor", id: "LIST" }]
		}),

		deleteRelayActorMatcher: build.mutation<RelayActor, { id: string, matcherID: string }>({
			query: ({ id, matcherID }) => ({
				method: "DELETE",
				url: `/api/v1/admin/relay_actors/${id}/matchers/${matcherID}`
			}),
			invalidatesTags: (res) =>
				res
					? [
						{ type: "RelayActor", id: "LIST" },
						{ type: "RelayActor", id: res.id },
					]
					: [{ type: "RelayActor", id: "LIST" }]
		}),
	}),
});

function accountFromQueriesState(
	state,
	stateKey: string,
	accountID: string,
) {
	let account: Account | undefined;
	const accounts = state.api.queries[stateKey]?.data as Account[] | undefined;
	if (accounts) {
		account = accounts.find(a => a.id == accountID );
	}
	return account;
}

function removeAccountFromCollection(
	endpointName: any,
	collectionOwnerID: string,
	accountID: string,
	dispatch: ThunkDispatch<any, any, AnyAction>,
): PatchCollection {
	return dispatch(
		extended.util.updateQueryData(
			endpointName,
			collectionOwnerID,
			(draft: MaybeDrafted<Account[]>) => {
				const i = draft.findIndex((a) => a.id === accountID);
				if (i !== -1) {
					// Remove from
					// collection.
					draft.splice(i, 1);
				}
			},
		)
	);
}

function prependAccountToCollection(
	endpointName: any,
	collectionOwnerID: string,
	account: Account,
	dispatch: ThunkDispatch<any, any, AnyAction>,
): PatchCollection {
	return dispatch(
		extended.util.updateQueryData(
			endpointName,
			collectionOwnerID,
			(draft: MaybeDrafted<Account[]>) => {
				draft.unshift(account);
			}),
	);
	
}

/**
 * Get all relay actors.
 */
const useRelayActorsQuery = extended.useRelayActorsQuery;

/**
 * Get a single relay actor.
 */
const useRelayActorQuery = extended.useRelayActorQuery;

/**
 * Create a relay actor.
 */
const useCreateRelayActorMutation = extended.useCreateRelayActorMutation;

/**
 * Update a relay actor (and its account).
 */
const useUpdateRelayActorMutation = extended.useUpdateRelayActorMutation;

/**
 * Delete the header of a relay actor account.
 */
const useDeleteRelayActorHeaderMutation = extended.useDeleteRelayActorHeaderMutation;

/**
 * Delete the avatar of a relay actor account.
 */
const useDeleteRelayActorAvatarMutation = extended.useDeleteRelayActorAvatarMutation;

/**
 * Delete a relay actor.
 */
const useDeleteRelayActorMutation = extended.useDeleteRelayActorMutation;

/**
 * Get followers of relay actor.
 */
const useRelayActorFollowersQuery = extended.useRelayActorFollowersQuery;

/**
 * Remove an account from followers on behalf of relay actor.
 */
const useRelayActorRemoveFromFollowersMutation = extended.useRelayActorRemoveFromFollowersMutation;

/**
 * Get follow requests of relay actor.
 */
const useRelayActorFollowRequestsQuery = extended.useRelayActorFollowRequestsQuery;

/**
 * Accept a follow request on behalf of relay actor.
 */
const useAcceptRelayActorFollowRequestMutation = extended.useAcceptRelayActorFollowRequestMutation;

/**
 * Reject a follow request on behalf of relay actor.
 */
const useRejectRelayActorFollowRequestMutation = extended.useRejectRelayActorFollowRequestMutation;

/**
 * Get blocks of relay actor.
 */
const useRelayActorBlocksQuery = extended.useRelayActorBlocksQuery;

/**
 * Create block on behalf of relay actor.
 */
const useRelayActorCreateBlockMutation = extended.useRelayActorCreateBlockMutation;

/**
 * Remove block on behalf of relay actor.
 */
const useRelayActorRemoveBlockMutation = extended.useRelayActorRemoveBlockMutation;

/**
 * Create a matcher of a relay actor.
 */
const useCreateRelayActorMatcherMutation = extended.useCreateRelayActorMatcherMutation;

/**
 * Delete a matcher of a relay actor.
 */
const useDeleteRelayActorMatcherMutation = extended.useDeleteRelayActorMatcherMutation;

export {
	useRelayActorsQuery,
	useRelayActorQuery,
	useCreateRelayActorMutation,
	useUpdateRelayActorMutation,
	useDeleteRelayActorHeaderMutation,
	useDeleteRelayActorAvatarMutation,
	useDeleteRelayActorMutation,
	useRelayActorFollowersQuery,
	useRelayActorRemoveFromFollowersMutation,
	useRelayActorFollowRequestsQuery,
	useAcceptRelayActorFollowRequestMutation,
	useRejectRelayActorFollowRequestMutation,
	useRelayActorBlocksQuery,
	useRelayActorCreateBlockMutation,
	useRelayActorRemoveBlockMutation,
	useCreateRelayActorMatcherMutation,
	useDeleteRelayActorMatcherMutation,
};
