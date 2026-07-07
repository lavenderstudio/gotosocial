#!/bin/bash
set -exo pipefail

OUTPUT='./web/assets/all_licenses.txt'

# stable sort
export LC_ALL=C

# open (temporary) output file
exec >"$OUTPUT~"

# write header
cat <<-'EOD'
	The GoToSocial software uses the following dependencies,
	whose individual licenses are reproduced in full:

EOD

# Copy over misc. licences, as well as these from our golang (vendor)
# and javascript (web/source) dependencies but sort them; those tr
# calls are needed because sort -z is not in POSIX yet.
find ./LICENSE ./vendor ./web/source -iname 'licen[CcSs]e*' -print0 | \
    tr '\n\0' '\0\n' | sort | tr '\n\0' '\0\n' | \
    while IFS= read -d '' -r file; do
	cat <<-EOD
		----------------------------------------------------------

		${file}:

		$(<"$file")

	EOD
done
# double line as EOF marker
echo '=========================================================='

# close output file and rename to final location if nothing went wrong
exec >&2
mv "$OUTPUT~" "$OUTPUT"
