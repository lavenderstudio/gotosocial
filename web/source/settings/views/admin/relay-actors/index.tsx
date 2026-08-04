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
import { useRelayActorsQuery } from "../../../lib/query/admin/relay-actors";
import Loading from "../../../components/loading";
import { Error } from "../../../components/error";
import { PageableList } from "../../../components/pageable-list";
import { RelayActor } from "../../../lib/types/relay";
import { useLocation } from "wouter";
import { RelayFlagsInfoListEntry } from "../../../components/relaylistentry";
import AccountCard from "../../../components/account-card";

export default function RelayActorsOverview() {
	
	const {
		data,
		isLoading,
		isFetching,
		isSuccess,
		isError,
		error,
	} = useRelayActorsQuery();
	
	if (isLoading || isFetching) {
		return <Loading />;
	}

	if (isError) {
		return <Error error={error} />;
	}

	if (data === undefined) {
		throw "undefined data";
	}

	const itemToEntry = (actor: RelayActor) => {
		return <RelayActorListEntry actor={actor} key={actor.id} />;
	};
	
	return (
		<div className="relay-actors">
			<div className="form-section-docs">
				<h1>Relay Actors</h1>
				<p>
					On this page you can see an overview of relay actors that have been created
					on this instance in order to relay posts to and from (remote) instances.
				</p>
				<a
					href="https://docs.gotosocial.org/en/stable/admin/relay_actors/"
					target="_blank"
					className="docslink"
					rel="noreferrer"
				>
					Learn more about relay actors (opens in a new tab)
				</a>
			</div>
			<PageableList
				isLoading={isLoading}
				isFetching={isFetching}
				isSuccess={isSuccess}
				isError={isError}
				error={error}
				items={data}
				itemToEntry={itemToEntry}
				emptyMessage="There are no relay actors yet."
			/>
		</div>
	);
}

function RelayActorListEntry({ actor }: { actor: RelayActor }) {
	const [ location, setLocation ] = useLocation();
	
	const onClick = (e) => {
		e.preventDefault();
		// When clicking on a relay actor,
		// go to the detail view for it.
		setLocation(`/${actor.id}`, {
			// Store the back location in
			// history so the detail view
			// can use it to return here.
			state: { backLocation: location }
		});
	};

	return (
		<span
			className="pseudolink relay-actor entry"
			aria-label={actor.account.username}
			title={actor.account.username}
			onClick={onClick}
			onKeyDown={(e) => {
				if (e.key === "Enter") {
					e.preventDefault();
					onClick(e);
				}
			}}
			role="link"
			tabIndex={0}
		>
			<AccountCard account={actor.account} />
			<dl className="info-list">
				<div className="info-list-entry">
					<dt>Followers:</dt>
					<dd>{actor.account.followers_count}</dd>
				</div>
				<div className="info-list-entry">
					<dt>Follow requests:</dt>
					<dd>{actor.account.source?.follow_requests_count}</dd>
				</div>
				<div className="info-list-entry">
					<dt>Blocks:</dt>
					<dd>{actor.account.source?.blocks_count}</dd>
				</div>
				<div className="info-list-entry">
					<dt>Matchers:</dt>
					<dd>{actor.matchers.length}</dd>
				</div>
				<RelayFlagsInfoListEntry
					public={actor.public}
					unlisted={actor.unlisted}
					match_by_default={actor.match_by_default}
					ignore_sensitive={actor.ignore_sensitive}
					ignore_media={actor.ignore_media}
					ignore_replies={actor.ignore_replies}
				/>
			</dl>
		</span>
	);
}
