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
import { RelayConnection } from "../lib/types/relay";
import { useLocation } from "wouter";

export default function RelayListEntry({ conn }: { conn: RelayConnection }) {
	const [ location, setLocation ] = useLocation();
	
	const onClick = (e) => {
		e.preventDefault();
		// When clicking on a relay conn,
		// go to the detail view for it.
		setLocation(`/${conn.id}`, {
			// Store the back location in
			// history so the detail view
			// can use it to return here.
			state: { backLocation: location }
		});
	};

	return (
		<span
			className="pseudolink relay-connection entry"
			aria-label=""
			title=""
			onClick={onClick}
			onKeyDown={(e) => {
				if (e.key === "Enter") {
					e.preventDefault();
					onClick(e);
				}
			}}
			role="link"
			tabIndex={0}
			key={conn.id}
		>
			<dl className="info-list">
				<div className="info-list-entry">
					<dt>Actor URI:</dt>
					<dd className="text-cutoff">{conn.relay_actor_uri}</dd>
				</div>
				<div className="info-list-entry">
					<dt>Approved:</dt>
					<dd className={`text-cutoff ${conn.approved ? "relay-connection-approved" : "relay-connection-not-approved"}`}>{conn.approved ? "yes" : "not yet"}</dd>
				</div>
				<div className="info-list-entry">
					<dt>Matchers:</dt>
					<dd>{conn.matchers.length}</dd>
				</div>
				<div className="info-list-entry relay-flags">
					<dt>Flags:</dt>
					<div className="relay-flags-icons">
						{ conn.public && 
							<dd title="public">
								<i className="fa fa-fw fa-globe" aria-hidden="true"></i>
								<span className="sr-only">public</span>
							</dd>
						}
						{ conn.unlisted && 
							<dd title="unlisted">
								<i className="fa fa-fw fa-unlock" aria-hidden="true"></i>
								<span className="sr-only">unlisted</span>
							</dd>
						}
						{ conn.match_by_default && 
							<dd title="match by default">
								<i className="fa fa-fw fa-check-square-o" aria-hidden="true"></i>
								<span className="sr-only">match by default</span>
							</dd>
						}
						{ conn.ignore_sensitive && 
							<dd title="ignore sensitive">
								<span className="fa-stack fa-fw">
									<i className="fa fa-fw fa-eye-slash fa-stack-1x"></i>
									<i className="fa fa-fw fa-ban fa-stack-1x crossed-out"></i>
								</span>
								<span className="sr-only">ignore sensitive</span>
							</dd>
						}
						{ conn.ignore_media && 
							<dd title="ignore media">
								<span className="fa-stack fa-fw">
									<i className="fa fa-fw fa-image fa-stack-1x"></i>
									<i className="fa fa-fw fa-ban fa-stack-1x crossed-out"></i>
								</span>
								<span className="sr-only">ignore media</span>
							</dd>
						}
						{ conn.ignore_replies && 
							<dd title="ignore replies">
								<span className="fa-stack fa-fw">
									<i className="fa fa-fw fa-reply-all fa-stack-1x"></i>
									<i className="fa fa-fw fa-ban fa-stack-1x crossed-out"></i>
								</span>
								<span className="sr-only">ignore replies</span>
							</dd>
						}
					</div>
				</div>
			</dl>
		</span>
	);
}
