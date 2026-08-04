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
import { FileInput, TextInput } from "./form/inputs";
import MutationButton from "./form/mutation-button";
import { FileFormInputHook, TextFormInputHook } from "../lib/form/types";
import { useCapitalize } from "../lib/util";

export interface ProfileImageUploadProps {
	headerOrAvatar: "header" | "avatar";
	imageField: FileFormInputHook;
	descriptionField: TextFormInputHook;
	deleteRes;
	deleteOnClick;
	unset: boolean;
}

export default function ProfileImageUpload(props: ProfileImageUploadProps) {
	const {
		headerOrAvatar,
		imageField,
		descriptionField,
		deleteRes,
		deleteOnClick,
		unset,
	} = props;

	const placeholder = useMemo(() => {
		return headerOrAvatar === "header"
			? "A green field with pink flowers."
			: "A cute drawing of a smiling sloth.";
	}, [headerOrAvatar]);

	return (
		<fieldset className="file-input-with-image-description">
			<legend>{useCapitalize(headerOrAvatar)}</legend>
			<FileInput
				label="Upload file"
				field={imageField}
				accept="image/png, image/jpeg, image/webp, image/gif"
			/>
			<TextInput
				field={descriptionField}
				label={`Image description; only settable if not using default ${headerOrAvatar}`}
				placeholder={placeholder}
				autoCapitalize="sentences"
				disabled={unset && !imageField.value}
			/>
			<MutationButton
				className={`delete-${headerOrAvatar}-button`}
				label={`Delete ${headerOrAvatar}`} 
				tabIndex={0}
				disabled={unset}
				onClick={deleteOnClick}
				result={deleteRes}
				submit={false}
			/>
		</fieldset>
	);
}
