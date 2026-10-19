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
import { useTextInput } from "../../../../lib/form";
import { useLazyApURLQuery } from "../../../../lib/query/admin/debug";
import { TextInput } from "../../../../components/form/inputs";
import MutationButton from "../../../../components/form/mutation-button";
import { ApURLResponse } from "../../../../lib/types/debug";
import Loading from "../../../../components/loading";
import { HighlightedCode } from "../../../../components/highlightedcode";

export default function ApURL() {
	const urlField = useTextInput("url");

	const [apURL, apURLResult] = useLazyApURLQuery();
	function submit(e) {
		e.preventDefault();
		apURL(urlField.value ?? "");
	}
	
	return (
		<div className="admin-debug-apurl">
			<form onSubmit={submit}>
				<div className="form-section-docs">
					<h2>AP URL</h2>
					<p>
						Call the provided URL with a signed request, and return the
						raw ActivityPub response in the <code>response_body</code> field.
						<br/>
						Useful for debugging communication between instances, and
						checking the shape of remote objects and actors.
					</p>
				</div>
				<TextInput
					field={urlField}
					label="URL"
					type="url"
					pattern="(http|https):\/\/.+"
					placeholder="https://example.org/users/someone"
				/>
				<MutationButton
					disabled={!urlField.value}
					label="Send request"
					result={apURLResult}
				/>
			</form>
			<ApURLResult {...apURLResult} />
		</div>
	);
}

interface ApURLResultProps {
	isLoading: boolean;
	isFetching: boolean;
	isSuccess: boolean;
	data?: ApURLResponse,
	isError: boolean;
}

function ApURLResult({
	isLoading,
	isFetching,
	isSuccess,
	data,
	isError,
}: ApURLResultProps) {	
	if (!(isSuccess || isError)) {
		// Hasn't been called yet.
		return null;
	}

	if (isLoading || isFetching) {
		return <Loading />;
	}

	if (!data) {
		return "No data";
	}

	// If response body was provided from remote
	// endpoint, try to format it nicely as JSON
	// to make it readable for GtS admin end user.
	let response_body = data.response_body;
	if (response_body.length > 0) {
		try {
			response_body = JSON.parse(response_body);
		} catch {
			// Response body wasn't valid JSON
			// (probably HTML), leave it alone.
		}
	}

	const jsonObj = {
		...data,
		response_body,
	};

	const jsonStr = JSON.stringify(jsonObj, null, 2);
	return <HighlightedCode code={jsonStr} lang="json" />;
}
