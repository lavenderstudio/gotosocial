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

import React, { useMemo } from "react";

import { RelayActor, RelayConnection } from "../lib/types/relay";
import MutationButton from "./form/mutation-button";
import { useBoolInput, useTextInput, useValue } from "../lib/form";
import { Checkbox, TextInput } from "./form/inputs";
import useFormSubmit from "../lib/form/submit";

interface RelayMatchersProps {
	entity: RelayConnection | RelayActor,
	createMatcherHook,
	deleteMatcherHook,
}

export default function RelayMatchers({
	entity,
	createMatcherHook,
	deleteMatcherHook,
}: RelayMatchersProps) {
	const form = {
		id: useValue("id", entity.id),
		keyword: useTextInput("keyword"),
		whole_word: useBoolInput("whole_word"),
		exclude: useBoolInput("exclude"),
	};
	const [ formSubmit, result ] = useFormSubmit(form, createMatcherHook());
	const [ remove, removeResult ] = deleteMatcherHook();

	// Link to appropriate docs page depending on if
	// this is a relay actor, subscription, or push.
	const [ noun, docsLink ] = useMemo(() => {
		switch (true) {
			case (entity.account_id !== undefined && entity.account_id.length != 0):
				// This is a relay subscription.
				return [
					"subscription",
					"https://docs.gotosocial.org/en/stable/admin/relay_subscriptions/#relay-matchers",
				];
			
			case ('created_by_account_id' in entity):
				// Must be a relay actor.
				return [
					"actor",
					"https://docs.gotosocial.org/en/stable/admin/relay_actors/#relay-matchers",
				];
			
			default:
				// Must be a relay push.
				return [
					"push",
					"https://docs.gotosocial.org/en/stable/user_guide/relay_pushes/#relay-matchers",
				];
		}
	}, [entity]);

	return (
		<>
			<div className="form-section-docs">
				<h3>Matchers</h3>
				<p>
					You can add relay matchers to this {noun} to give granular control over which posts are relayed.
					<br/><br/>If the relay {noun} <em>does not</em> match posts by default, posts will only be relayed if their content is matched by a matcher. If you create no matchers, nothing will be relayed by the {noun}.
					<br/><br/>Conversely, if the relay {noun} <em>does</em> match posts by default, you can use exclude matchers to <em>prevent</em> posts from being relayed, based on their content. If you create no exclude matchers, everything will be relayed.
					<br/><br/>Regardless of whether the relay {noun} does or does not match posts by default, exclude matchers will prevent posts from being relayed, even if they would otherwise be matched (ie., exclude matchers take priority).
				</p>
				<a
					href={docsLink}
					target="_blank"
					className="docslink"
					rel="noreferrer"
				>
					Learn more about relay matchers (opens in a new tab)
				</a>
			</div>
			<form
				onSubmit={formSubmit}
				// Prevent password managers
				// trying to fill in fields.
				autoComplete="off"
			>
				<TextInput
					field={form.keyword}
					label="Keyword (case insensitive)"
					placeholder="#SomeHashtag"
					spellCheck="false"
					autoCapitalize="none"
				/>

				<Checkbox
					label={"Match whole word; if unchecked, allow matching word fragments"}
					field={form.whole_word}
				/>

				<Checkbox
					label={"Exclude posts matched by this matcher, instead of including them"}
					field={form.exclude}
				/>

				<MutationButton
					label="Create matcher"
					result={result}
					disabled={form.keyword.value == ""}
				/>
			</form>
			{ entity.matchers.length !== 0 &&
				<>
					<h4>Active matchers</h4>
					<ol className="matchers list">
						{ entity.matchers.map(matcher => {
							const label = `"${matcher.keyword}"; ${matcher.whole_word ? "whole word match" : "partial match"}; ${matcher.exclude ? "exclude matches" : "include matches"}`;
							return (
								<li
									className="entry"
									id={matcher.id}
									key={matcher.id}
									aria-label={label}
									title={label}
								>
									<div className="relay-flags-icons">
										{ matcher.whole_word
											? <>
												<div title="whole word match">
													<i className="fa fa-fw fa-text-width" aria-hidden="true"></i>
													<span className="sr-only">whole word match</span>
												</div>
											</>
											: <>
												<div title="partial word match">
													<i className="fa fa-fw fa-i-cursor" aria-hidden="true"></i>
													<span className="sr-only">partial word match</span>
												</div>
											</>
										}
										{ matcher.exclude
											? <>
												<div title="exclude matches">
													<i className="fa fa-fw fa-close" aria-hidden="true"></i>
													<span className="sr-only">exclude matches</span>
												</div>
											</>
											: <>
												<div title="include matches">
													<i className="fa fa-fw fa-check" aria-hidden="true"></i>
													<span className="sr-only">include matches</span>
												</div>
											</>
										}
									</div>
									<div className="relay-matcher-keyword">{matcher.keyword}</div>
									<MutationButton
										label="Delete"
										type="button"
										className="button danger"
										onClick={(e) => {
											e.preventDefault();
											remove({id: entity.id, matcherID: matcher.id});
										}}
										disabled={false}
										showError={false}
										result={removeResult}
									/>
								</li>
							);
						}) }
					</ol>
				</>
			}
		</>
	);
}
