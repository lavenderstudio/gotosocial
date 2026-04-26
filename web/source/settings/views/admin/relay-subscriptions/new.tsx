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
import { useCreateRelaySubscriptionMutation } from "../../../lib/query/admin/relay-subscriptions";
import RelayNew from "../../../components/relaynew";

export default function RelaySubscriptionNew() {
	return (
		<RelayNew
			relayType="Subscription"
			verb="ingest"
			helpBlurb={
				<>
					<p>
						You can use this form to create a new relay subscription targeting a remote actor URI.
						<br/>For help with this form and its various flags, see the documentation section <a
							href="https://docs.gotosocial.org/en/stable/admin/relay_subscriptions/#create-a-relay-subscription"
							target="_blank"
							rel="noreferrer"
						>
							create a relay subscription (opens in a new tab)
						</a>.
					</p>
				</>
			}
			createHook={useCreateRelaySubscriptionMutation}
		/>);
}
