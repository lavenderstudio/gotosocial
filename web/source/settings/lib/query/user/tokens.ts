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

import {
	SearchTokenInfoParams,
	SearchTokenInfoResp,
	TokenInfo,
	TokenInfoUpdateParams,
} from "../../types/tokeninfo";
import { gtsApi } from "../gts-api";
import parse from "parse-link-header";

const extended = gtsApi.injectEndpoints({
	endpoints: (build) => ({
		searchTokenInfo: build.query<SearchTokenInfoResp, SearchTokenInfoParams>({
			query: (form) => {
				const params = new(URLSearchParams);
				Object.entries(form).forEach(([k, v]) => {
					if (v !== undefined) {
						params.append(k, v);
					}
				});

				let query = "";
				if (params.size !== 0) {
					query = `?${params.toString()}`;
				}

				return {
					url: `/api/v1/tokens${query}`
				};
			},
			// Headers required for paging.
			transformResponse: (apiResp: TokenInfo[], meta) => {
				const tokens = apiResp;
				const linksStr = meta?.response?.headers.get("Link");
				const links = parse(linksStr);
				return { tokens, links };
			},
			providesTags: [{ type: "TokenInfo", id: "TRANSFORMED" }]
		}),
		getToken: build.query<TokenInfo, string>({
			query: (id) => ({
				url: `/api/v1/tokens/${id}`,
			}),
			providesTags: (_result, _error, id) => [
				{ type: "TokenInfo", id }
			],
		}),
		updateToken: build.mutation<TokenInfo, {id: string} & TokenInfoUpdateParams>({
			query: ({ id, ...formData}) => {
				return {
					method: "PUT",
					url: `/api/v1/tokens/${id}`,
					asForm: true,
					body: formData,
					// Don't discardEmpty when updating, as we
					// want to be able to set name to empty string.
					discardEmpty: false,
				};
			},
			invalidatesTags: (res) =>
				res
					? [
						{ type: "TokenInfo", id: "TRANSFORMED" },
						{ type: "TokenInfo", id: res.id },
					]
					: [{ type: "TokenInfo", id: "TRANSFORMED" }]
		}),
		invalidateToken: build.mutation<any, string>({
			query: (id) => ({
				method: "POST",
				url: `/api/v1/tokens/${id}/invalidate`,
			}),
			invalidatesTags: (res) =>
				res
					? [
						{ type: "TokenInfo", id: "TRANSFORMED" },
						{ type: "TokenInfo", id: res.id },
					]
					: [{ type: "TokenInfo", id: "TRANSFORMED" }]
		}),
	})
});

export const {
	useLazySearchTokenInfoQuery,
	useInvalidateTokenMutation,
	useGetTokenQuery,
	useUpdateTokenMutation,
} = extended;
