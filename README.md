magicdb
---

magicdb.DB wraps sql.DB objects in a way that can simplify certain operations.
First, instead of returning errors after every function call, DB and all
objects it creates (Tx, Stmt, etc) silently fail on any sql.DB error.  This
error is stored internally, and prevents any of the objects from performing any
further tasks, allowing the caller to handle the error wherever it makes sense.

Additionally, there are optional types which utilize reflection to simplify the
most common operations so that you can tag your structures and get easy read,
update, and insert operations.  These are extremely limited in scope as I don't
want to try and make magic-hell, and should really only be used for table-level
abstractions.  At this time, there are no plans to deal with magicking
multi-table select/insert/update.

This is still a work in progress and I don't claim this package will be most
people's preferred approach.

Versions
---

This project will follow the Semantic Versioning specification, v2.0.0, and all
tags will be prefixed with "v" to allow [gb](https://getgb.io/) to pull this
package as an unvendored dependency.

LICENSE
---

This is licensed under CC0, a very permissive public domain license
