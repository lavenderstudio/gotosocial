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

import React, { ReactNode, useMemo } from "react";
import { useInstanceV1Query } from "../lib/query/gts-api";
import { Account } from "../lib/types/account";

interface AccountCardProps {
	account: Account;
	children?: ReactNode;
}

export default function AccountCard({
	account,
	children,
}: AccountCardProps) {
	const { data: instance } = useInstanceV1Query();
	const username = useMemo(() => {
		if (account.acct.includes("@")) {
			// Remote account.
			return "@" + account.acct;
		} else {
			// Local account.
			return "@" + account.acct + "@" + instance?.account_domain;
		}
	}, [account.acct, instance?.account_domain]);

	return (
		<div className="account-card">
			<img className="avatar" src={account.avatar} alt="" />
			<h3 className="text-cutoff">{account.display_name?.length > 0 ? account.display_name : account.acct}</h3>
			<span className="text-cutoff">{username}</span>
			{ children }
		</div>
	);
}