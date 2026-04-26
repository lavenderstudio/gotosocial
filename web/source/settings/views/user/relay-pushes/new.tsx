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
import { useCreateRelayPushMutation } from "../../../lib/query/user/relay-pushes";
import RelayNew from "../../../components/relaynew";

export default function RelayPushNew() {
	return (
		<RelayNew
			relayType="Push"
			verb="send"
			helpBlurb={
				<>
					On this page you can create a new relay push connection by specifying an ActivityPub relay actor to forward your posts to.
					<br/><br/>To determine which of your posts should and shouldn't be sent to the relay, you can use flags on the relay form below,
					combined with relay matchers (once you've created the relay connection) to match keywords in posts in order to either include
					or exclude them in relaying.
					<br/><br/>If you don't check "Send all allowed posts by default" then nothing will be relayed unless you create relay matchers.
					<br/><br/>If you 
				</>
			}
			createHook={useCreateRelayPushMutation}
		/>);
}
