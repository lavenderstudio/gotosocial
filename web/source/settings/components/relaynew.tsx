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

import React, { ReactNode } from "react";
import useFormSubmit from "../lib/form/submit";
import { useBoolInput, useTextInput } from "../lib/form";
import MutationButton from "./form/mutation-button";
import { TextInput } from "./form/inputs";
import { useLocation } from "wouter";
import RelayFlagsForm from "./relayflags";

interface RelayNewProps {
	relayType: string,
	verb: string,
	helpBlurb: ReactNode,
	createHook,
}

export default function RelayNew({
	relayType,
	verb,
	helpBlurb,
	createHook
}: RelayNewProps) {
	const [ _location, setLocation ] = useLocation();
	const ourActor = relayType === "Actor";

	const form = {
		// When creating an actor on our instance, don't submit relay_actor_uri.
		relay_actor_uri: useTextInput("relay_actor_uri", { nosubmit: ourActor }),
		// When creating a connection to a remote actor, don't submit username.
		username: useTextInput("username", { nosubmit: !ourActor }),
		public: useBoolInput("public", { defaultValue: true }),
		unlisted: useBoolInput("unlisted", { defaultValue: false }),
		match_by_default: useBoolInput("match_by_default", { defaultValue: false }),
		ignore_sensitive: useBoolInput("ignore_sensitive", { defaultValue: false }),
		ignore_media: useBoolInput("ignore_media", { defaultValue: false }),
		ignore_replies: useBoolInput("ignore_replies", { defaultValue: false }),
	};

	const [formSubmit, result] = useFormSubmit(
		form,
		createHook(),
		{
			changedOnly: false,
			onFinish: (res) => {
				if (res.data) {
					// Creation successful,
					// redirect to overview.
					setLocation(`/`);
				}
			},
		});

	return (
		<form
			onSubmit={formSubmit}
			// Prevent password managers
			// trying to fill in fields.
			autoComplete="off"
		>
			<div className="form-section-docs">
				<h2>New Relay {relayType}</h2>
				{helpBlurb}
			</div>

			{ !ourActor
				? <TextInput
					field={form.relay_actor_uri}
					label="Relay actor URI"
					placeholder="https://relay.example.org/actor"
					type="url"
					pattern="https?://.*"
					spellCheck="false"
					autoCapitalize="none"
					required={true}
				/>
				: <TextInput
					field={form.username}
					label="Relay actor username (lowercase a-z, numbers, and underscores; max 64 characters)"
					placeholder="some_username_or_whatever"
					type="text"
					pattern="^[a-z0-9_]{1,64}$"
					spellCheck="false"
					autoCapitalize="off"
					required={true}
					title="lowercase a-z, numbers, and underscores; max 64 characters"
				/>
			}

			<RelayFlagsForm
				verb={verb}
				form_field_public={form.public}
				form_field_unlisted={form.unlisted}
				form_field_ignore_sensitive={form.ignore_sensitive}
				form_field_ignore_media={form.ignore_media}
				form_field_ignore_replies={form.ignore_replies}
				form_field_match_by_default={form.match_by_default}
			/>

			<MutationButton
				label="Save"
				result={result}
				disabled={
					ourActor
						? !form.username.value || !form.username.valid
						: !form.relay_actor_uri.value || !form.relay_actor_uri.valid
				}
			/>
		</form>
	);
}
