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
import { Checkbox } from "./form/inputs";
import { useCapitalize } from "../lib/util";
import { BoolFormInputHook } from "../lib/form/types";

interface RelayFlagsProps {
	verb: string;
	form_field_public: BoolFormInputHook;
	form_field_unlisted: BoolFormInputHook;
	form_field_ignore_sensitive: BoolFormInputHook;
	form_field_ignore_media: BoolFormInputHook;
	form_field_ignore_replies: BoolFormInputHook;
	form_field_match_by_default: BoolFormInputHook;
}

export default function RelayFlags(props: RelayFlagsProps) {
	const {
		verb,
		form_field_public,
		form_field_unlisted,
		form_field_ignore_sensitive,
		form_field_ignore_media,
		form_field_ignore_replies,
		form_field_match_by_default,
	} = props;
	
	const verbUpper = useCapitalize(verb);
	return (
		<>
			<Checkbox
				label={`${verbUpper} public visibility posts`}
				field={form_field_public}
			/>

			<Checkbox
				label={`${verbUpper} unlisted visibility posts`}
				field={form_field_unlisted}
			/>

			<Checkbox
				label={`Never ${verb} posts marked as sensitive`}
				field={form_field_ignore_sensitive}
			/>

			<Checkbox
				label={`Never ${verb} posts with media`}
				field={form_field_ignore_media}
			/>

			<Checkbox
				label={`Never ${verb} replies/comments`}
				field={form_field_ignore_replies}
			/>

			<Checkbox
				label={`Match posts by default (use exclude matchers to exclude specific posts)`}
				field={form_field_match_by_default}
			/>
		</>
	);
}
