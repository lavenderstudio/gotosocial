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

import React, { useState } from "react";
import { useBaseUrl } from "../../../lib/navigation/util";
import { useLocation, useParams } from "wouter";
import BackButton from "../../../components/back-button";
import FormWithData from "../../../lib/form/form-with-data";
import { useCreateRelayActorMatcherMutation, useDeleteRelayActorAvatarMutation, useDeleteRelayActorHeaderMutation, useDeleteRelayActorMatcherMutation, useDeleteRelayActorMutation, useRelayActorQuery, useUpdateRelayActorMutation } from "../../../lib/query/admin/relay-actors";
import { RelayActor } from "../../../lib/types/relay";
import RelayFlagsForm from "../../../components/relayflags";
import { useBoolInput, useFieldArrayInput, useFileInput, useTextInput, useValue } from "../../../lib/form";
import useFormSubmit from "../../../lib/form/submit";
import MutationButton from "../../../components/form/mutation-button";
import { useInstanceV1Query } from "../../../lib/query/gts-api";
import RelayMatchers from "../../../components/relaymatchers";
import FakeProfile from "../../../components/profile";
import ProfileImageUpload from "../../../components/profileimage";
import { Checkbox, TextArea, TextInput } from "../../../components/form/inputs";
import ProfileFields from "../../../components/fields";

export default function RelayActorDetail() {
	const params: { relayActorId: string } = useParams();
	const baseUrl = useBaseUrl();
	const backLocation: String = history.state?.backLocation ?? `~${baseUrl}`;

	return (
		<div className="relay-actor-detail">
			<h1><BackButton to={backLocation} /> Relay Actor Details</h1>
			<div className="form-section-docs">
				<p>
					On this page you can customize various settings relating to the relay actor's appearance and behavior, and manage the relay actor's relationships.
					<br/>After changing settings and/or uploading a new avatar or header, be sure to scroll to the bottom of this page and click "Update actor" to confirm your changes.
				</p>
				<a
					href="https://docs.gotosocial.org/en/stable/admin/relay_actors"
					target="_blank"
					className="docslink"
					rel="noreferrer"
				>
					Learn more about relay actors (opens in a new tab)
				</a>
			</div>
			<FormWithData
				dataQuery={useRelayActorQuery}
				queryArg={params.relayActorId}
				DataForm={RelayActorDetailForm}
			/>
		</div>
	);
}

interface RelayActorDetailFormProps {
	data: RelayActor;
}

function RelayActorDetailForm({data: actor}: RelayActorDetailFormProps) {
	const { data: instance } = useInstanceV1Query();
	const instanceConfig = React.useMemo(() => {
		return {
			allowCustomCSS: instance?.configuration?.accounts?.allow_custom_css === true,
			maxPinnedFields: instance?.configuration?.accounts?.max_profile_fields ?? 6
		};
	}, [instance]);
	
	const [ _location, setLocation ] = useLocation();
	const baseUrl = useBaseUrl();
	const form = {
		id: useValue("id", actor.id),
		public: useBoolInput("public", { source: actor }),
		unlisted: useBoolInput("unlisted", { source: actor }),
		match_by_default: useBoolInput("match_by_default", { source: actor }),
		ignore_sensitive: useBoolInput("ignore_sensitive", { source: actor }),
		ignore_media: useBoolInput("ignore_media", { source: actor }),
		ignore_replies: useBoolInput("ignore_replies", { source: actor }),
		discoverable: useBoolInput("discoverable", { source: actor.account }),
		display_name: useTextInput("display_name", { source: actor.account }),
		note: useTextInput("note", { source: actor.account, valueSelector: (a) => a.source?.note }),
		avatar: useFileInput("avatar", { withPreview: true }),
		avatar_description: useTextInput("avatar_description", { source: actor.account }),
		header: useFileInput("header", { withPreview: true }),
		header_description: useTextInput("header_description", { source: actor.account }),
		locked: useBoolInput("locked", { source: actor.account }),
		fields: useFieldArrayInput("fields_attributes", {
			defaultValue: actor.account.source?.fields,
			length: instanceConfig.maxPinnedFields
		}),
	};

	const [ noHeader, setNoHeader ] = useState(!actor.account.header_media_id);
	const [ deleteHeader, deleteHeaderRes ] = useDeleteRelayActorHeaderMutation();
	const [ noAvatar, setNoAvatar ] = useState(!actor.account.avatar_media_id);
	const [ deleteAvatar, deleteAvatarRes ] = useDeleteRelayActorAvatarMutation();

	const [submit, result] = useFormSubmit(form, useUpdateRelayActorMutation(), {
		changedOnly: true,
		onFinish: (res) => {
			if ('data' in res) {
				form.avatar.reset();
				form.header.reset();
				setNoAvatar(!res.data.account.avatar_media_id);
				setNoHeader(!res.data.account.header_media_id);
			}
		}
	});
	const [ removeTrigger, removeResult ] = useDeleteRelayActorMutation();
	
	return (
		<>
			<ManageRelationships actor={actor} />
			<form className="user-profile" onSubmit={submit}>
				<div className="overview">
					<div className="form-section-docs">
						<h3>Profile</h3>
					</div>

					<a href={actor.account.url} target="_blank">Open relay actor account profile page.</a>

					<FakeProfile
						avatar={form.avatar.previewValue ?? actor.account.avatar}
						header={form.header.previewValue ?? actor.account.header}
						display_name={form.display_name.value ?? actor.account.username}
						bot={actor.account.bot}
						username={actor.account.username}
						role={actor.account.role}
					/>

					<ProfileImageUpload
						headerOrAvatar="header"
						imageField={form.header}
						descriptionField={form.header_description}
						deleteRes={deleteHeaderRes}
						deleteOnClick={(e) => {
							e.preventDefault();
							deleteHeader(actor.id).then(res => {
								if ('data' in res) {
									setNoHeader(true);
								}
							});
						}}
						unset={noHeader}
					/>

					<ProfileImageUpload
						headerOrAvatar="avatar"
						imageField={form.avatar}
						descriptionField={form.avatar_description}
						deleteRes={deleteAvatarRes}
						deleteOnClick={(e) => {
							e.preventDefault();
							deleteAvatar(actor.id).then(res => {
								if ('data' in res) {
									setNoAvatar(true);
								}
							});
						}}
						unset={noAvatar}
					/>
				</div>

				<div className="form-section-docs">
					<h3>Basic Information</h3>
				</div>

				<TextInput
					field={form.display_name}
					label="Display name"
					placeholder="Some Kickass Relay"
					autoCapitalize="words"
					spellCheck="false"
				/>
				<TextArea
					field={form.note}
					label="Bio"
					placeholder="This relay relays posts with tags #boobies and #FOSS!"
					autoCapitalize="sentences"
					rows={8}
				/>
				<fieldset>
					<legend>Profile fields</legend>
					<ProfileFields
						field={form.fields}
					/>
				</fieldset>

				<div className="form-section-docs">
					<h3>Visibility and privacy</h3>
				</div>

				<Checkbox
					field={form.locked}
					label="Manually approve requests to connect to the relay"
				/>
				<Checkbox
					field={form.discoverable}
					label="Mark the relay as discoverable by search engines"
				/>

				<div className="form-section-docs">
					<h3>Relay Behavior</h3>
				</div>

				<RelayFlagsForm
					verb="relay"
					form_field_public={form.public}
					form_field_unlisted={form.unlisted}
					form_field_ignore_sensitive={form.ignore_sensitive}
					form_field_ignore_media={form.ignore_media}
					form_field_ignore_replies={form.ignore_replies}
					form_field_match_by_default={form.match_by_default}
				/>

				<div className="action-buttons row">
					<MutationButton
						label="Update actor"
						result={result}
						disabled={
							!form.discoverable.hasChanged() &&
							!form.display_name.hasChanged() &&
							!form.note.hasChanged() &&
							!form.avatar.hasChanged() &&
							!form.avatar_description.hasChanged() &&
							!form.header.hasChanged() &&
							!form.header_description.hasChanged() &&
							!form.locked.hasChanged() &&
							!form.fields.hasChanged() &&
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
							removeTrigger(actor.id);
							setLocation(`~${baseUrl}/overview`);
						}}
						label="Delete actor"
						result={removeResult}
						className="button danger"
						showError={false}
						disabled={false}
					/>
				</div>
			</form>
			<RelayMatchers
				entity={actor}
				createMatcherHook={useCreateRelayActorMatcherMutation}
				deleteMatcherHook={useDeleteRelayActorMatcherMutation}
			/>
		</>
	);
}

function ManageRelationships({ actor }: { actor: RelayActor }) {
	const [ location, setLocation ] = useLocation();

	// For each button, set the appropriate
	// location, and store backLocation in state.
	const followersCount = actor.account.followers_count;
	const onClickFollowers = (e) => {
		e.preventDefault();
		setLocation(`${location}/followers`, {
			state: { backLocation: location }
		});
	};
	
	const followRequestsCount = actor.account.source?.follow_requests_count;
	const onClickFollowRequests = (e) => {
		e.preventDefault();
		setLocation(`${location}/follow_requests`, {
			state: { backLocation: location }
		});
	};

	const blocksCount = actor.account.source?.blocks_count;
	const onClickBlocks = (e) => {
		e.preventDefault();
		setLocation(`${location}/blocks`, {
			state: { backLocation: location }
		});
	};
	
	return (
		<>
			<div className="form-section-docs">
				<h3>Relationships</h3>
			</div>
			<div className="stats-and-buttons">
				<div className="stats-and-button">
					<span className="text-cutoff">
						{ followersCount } follower{ followersCount !== 1 && "s" }
					</span>
					<span
						className="button"
						title={"Manage followers"}
						onClick={onClickFollowers}
						onKeyDown={(e) => {
							if (e.key === "Enter") {
								e.preventDefault();
								onClickFollowers(e);
							}
						}}
						role="link"
						tabIndex={0}
					>
						Manage followers
					</span>
				</div>
				<div className="stats-and-button">
					<span className="text-cutoff">
						{ followRequestsCount } follow request{ followRequestsCount !== 1 && "s" }
					</span>
					<span
						className="button"
						title={"Manage follow requests"}
						onClick={onClickFollowRequests}
						onKeyDown={(e) => {
							if (e.key === "Enter") {
								e.preventDefault();
								onClickFollowRequests(e);
							}
						}}
						role="link"
						tabIndex={0}
					>
						Manage follow requests
					</span>
				</div>
				<div className="stats-and-button">
					<span className="text-cutoff">
						{ blocksCount } block{ blocksCount !== 1 && "s" }
					</span>
					<span
						className="button"
						title={"Manage blocks"}
						onClick={onClickBlocks}
						onKeyDown={(e) => {
							if (e.key === "Enter") {
								e.preventDefault();
								onClickBlocks(e);
							}
						}}
						role="link"
						tabIndex={0}
					>
						Manage blocks
					</span>
				</div>
			</div>
		</>
	);
}
