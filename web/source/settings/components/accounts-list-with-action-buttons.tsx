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
import { Account } from "../lib/types/account";
import AccountCard from "./account-card";

interface AccountsCollectionWithButtonsProps {
	accounts: Account[];
	getButtons?: (_account: Account) => ReactNode;
}

export default function AccountsCollectionWithButtons({
	accounts,
	getButtons,
}: AccountsCollectionWithButtonsProps) {
	return (
		<div className="list pageable-list">
			{ accounts.map(account => {
				return (
					<a
						title={account.acct}
						className="entry nounderline"
						href={account.url}
						target="_blank"
						rel="noreferrer"
						key={account.id}
					>
						<AccountCard
							key={account.id}
							account={account}
						/>
						{ getButtons &&
							<div className="action-buttons">
								{getButtons(account)}
							</div>
						}
					</a>
				);
			}) }
		</div>
	);
}
