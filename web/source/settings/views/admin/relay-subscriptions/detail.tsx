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

import React from "react";

import FormWithData from "../../../lib/form/form-with-data";
import { useParams } from "wouter";
import { useBaseUrl } from "../../../lib/navigation/util";
import BackButton from "../../../components/back-button";
import {
	useCreateRelaySubscriptionMatcherMutation,
	useDeleteRelaySubscriptionMatcherMutation,
	useDeleteRelaySubscriptionMutation,
	useRelaySubscriptionQuery,
	useUpdateRelaySubscriptionMutation
} from "../../../lib/query/admin/relay-subscriptions";
import RelayDetailForm from "../../../components/relaydetailform";

export default function RelaySubscriptionDetail() {
	const params: { relaySubscriptionId: string } = useParams();
	const baseUrl = useBaseUrl();
	const backLocation: String = history.state?.backLocation ?? `~${baseUrl}`;

	return (
		<div className="relay-connection-detail">
			<h1><BackButton to={backLocation} /> Relay Subscription Details</h1>
			<FormWithData
				dataQuery={useRelaySubscriptionQuery}
				queryArg={params.relaySubscriptionId}
				DataForm={RelayDetailForm}
				{...{
					verb: "ingest",
					updateHook: useUpdateRelaySubscriptionMutation,
					deleteHook: useDeleteRelaySubscriptionMutation,
					createMatcherHook: useCreateRelaySubscriptionMatcherMutation,
					deleteMatcherHook: useDeleteRelaySubscriptionMatcherMutation,
				}}
			/>
		</div>
	);
}
