# GoToSocial TODOs

things noted down by the maintainers that could do with being done!

## database
- [chore] tidy up db.GetAccountStatuses() (separate functions perhaps?)
- [chore] update remaining database queries / API endpoints to use paging.Page{}
- [chore] update more of the database helpers in util.go to use the WhereGroupQuery() style generic parameterisation
- [space] add migrations to convert edits, accounts, ~~follows~~, ~~follow requests~~ (, etc) flags to bitsets
- [space] drop unnecessary 'updated_at' columns
- [space] replace 'created_at' columns with time parsed from ULIDs (where possible)
- [performance] replace ULID columns with binary representation
- [performance] replace "upsert" queries with more performant alternatives (these do a lot of runtime logic which could be determined ahead of time)
- [performance] add multi-delete queries for mentions, statuses (+boosts), etc
- [performance] replacing packaged pageable response with returned pageable response builder, so we can reuse incoming request query instead of allocating a new map

## miscellaneous
- [chore] update our libraries list on the README
- [chore] move dereferencer hooks to Ingester{} type that handles calling appropriate hooks for any ingested model
- [chore] deinterface the database somehow (where possible given dependency cycling 😭), have a single DB type so all bundb/*.go can access all other internal funcs
- kim: [chore/docs] update support matrix given go-sqlite3 dropped wazero usage
- kim: [supported platforms] update ffmpreg to use ncruces/wasm2go (?)
- kim: [performance] write alternative bundb serializing backend (?)
- kim: [performance] fork bundb to support not always going through "database/sql" (?)
