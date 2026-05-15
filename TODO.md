# GoToSocial TODOs

things noted down by the maintainers that could do with being done!

## database
- [chore] tidy up db.GetAccountStatuses() (separate functions perhaps?)
- [chore] update remaining database queries / API endpoints to use paging.Page{}
- [space] add migrations to convert edits, accounts (, etc) flags to bitsets
- [space] drop unnecessary 'updated_at' columns
- [space] replace 'created_at' columns with time parsed from ULIDs (where possible)
- [performance] replace ULID columns with binary representation
- [performance] replace "upsert" queries with more performant alternatives (these do a lot of runtime logic which could be determined ahead of time)
- [performance] add multi-delete queries for mentions, statuses (+boosts), etc

## miscellaneous
- [chore] finish code commenting where missing (search for '// (\w+\b)?...')
- [chore] move away from using Gin, they're all-in on "AI", blegh
- [chore] deinterface the database somehow (where possible given dependency cycling 😭), have a single DB type so all bundb/*.go can access all other internal funcs
- kim: [chore/docs] update support matrix given go-sqlite3 dropped wazero usage, (potentially) remove modernc.org/sqlite dependency
- kim: [supported platforms] update ffmpreg to use ncruces/wasm2go (?)
- kim: [performance] write alternative bundb serializing backend (?)
- kim: [performance] fork bundb to support not always going through "database/sql" (?)
