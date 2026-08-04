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

import { useLocation } from "wouter";
import { useBaseUrl } from "../lib/navigation/util";
import { RelayConnection } from "../lib/types/relay";
import MutationButton from "./form/mutation-button";
import UsernameLozenge from "./username-lozenge";
import { useBoolInput, useValue } from "../lib/form";
import useFormSubmit from "../lib/form/submit";
import RelayFlagsForm from "./relayflags";
import { DateTimeMinute } from "./datetime";
import RelayMatchers from "./relaymatchers";

interface RelayDetailFormProps {
	data: RelayConnection,
	verb: string,
	updateHook,
	deleteHook,
	createMatcherHook,
	deleteMatcherHook,
}

export default function RelayDetailForm({
	data: conn,
	verb,
	updateHook,
	deleteHook,
	createMatcherHook,
	deleteMatcherHook,
}: RelayDetailFormProps) {
	const [ _location, setLocation ] = useLocation();
	const baseUrl = useBaseUrl();
	const form = {
		id: useValue("id", conn.id),
		public: useBoolInput("public", { source: conn }),
		unlisted: useBoolInput("unlisted", { source: conn }),
		match_by_default: useBoolInput("match_by_default", { source: conn }),
		ignore_sensitive: useBoolInput("ignore_sensitive", { source: conn }),
		ignore_media: useBoolInput("ignore_media", { source: conn }),
		ignore_replies: useBoolInput("ignore_replies", { source: conn }),
	};
	const [ submit, result ] = useFormSubmit(form, updateHook());
	const [ removeTrigger, removeResult ] = deleteHook();

	return (
		<>
			<dl className="info-list">
				<div className="info-list-entry">
					<dt>Actor URI:</dt>
					<dd>{conn.relay_actor_uri}</dd>
				</div>
				<div className="info-list-entry">
					<dt>Created at:</dt>
					<dd><DateTimeMinute iso8601={conn.created_at}/></dd>
				</div>
				{
					conn.account_id &&
						<div className="info-list-entry">
							<dt>Created by:</dt>
							<dd><UsernameLozenge account={conn.account_id}/></dd>
						</div>
				}
				<div className="info-list-entry">
					<dt>Approved:</dt>
					<dd className={`text-cutoff ${conn.approved ? "relay-connection-approved" : "relay-connection-not-approved"}`}>{conn.approved ? "yes" : "not yet"}</dd>
				</div>
			</dl>
			<form onSubmit={submit}>
				<RelayFlagsForm
					verb={verb}
					form_field_public={form.public}
					form_field_unlisted={form.unlisted}
					form_field_ignore_sensitive={form.ignore_sensitive}
					form_field_ignore_media={form.ignore_media}
					form_field_ignore_replies={form.ignore_replies}
					form_field_match_by_default={form.match_by_default}
				/>

				<div className="action-buttons row">
					<MutationButton
						label="Update"
						result={result}
						disabled={
							!form.public.hasChanged() &&
							!form.unlisted.hasChanged() &&
							!form.match_by_default.hasChanged() &&
							!form.ignore_sensitive.hasChanged() &&
							!form.ignore_media.hasChanged() &&
							!form.ignore_replies.hasChanged()
						}
					/>

					<MutationButton
						type="button"
						onClick={() => {
							removeTrigger(conn.id);
							setLocation(`~${baseUrl}/overview`);
						}}
						label="Delete"
						result={removeResult}
						className="button danger"
						showError={false}
						disabled={false}
					/>
				</div>
			</form>
			<RelayMatchers
				entity={conn}
				createMatcherHook={createMatcherHook}
				deleteMatcherHook={deleteMatcherHook}
			/>
		</>
	);
}
