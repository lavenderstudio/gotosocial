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
import RelayNew from "../../../components/relaynew";
import { useCreateRelayActorMutation } from "../../../lib/query/admin/relay-actors";

export default function RelayActorNew() {
	return (
		<RelayNew
			relayType="Actor"
			verb="relay"
			helpBlurb={
				<>
					<p>
						You can use this form to create a new relay actor on this instance, which can be used to relay posts to and from (remote) instances.
						<br/>For help with this form and its various flags, see the documentation section <a
							href="https://docs.gotosocial.org/en/stable/admin/relay_actors/#create-a-relay-actor"
							target="_blank"
							rel="noreferrer"
						>
							create a relay actor (opens in a new tab)
						</a>.
					</p>
					<div className="info">
						<i className="fa fa-fw fa-info-circle" aria-hidden="true"></i>
						<b>
							Relay actor usernames are unique.
							<br/>Once you've created a relay actor with a certain username, you will never be able
							to use that username again for another relay actor, even if you delete the first one.
						</b>
					</div>
				</>
			}
			createHook={useCreateRelayActorMutation}
		/>
	);
}

