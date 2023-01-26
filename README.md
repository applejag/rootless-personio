<!--
SPDX-FileCopyrightText: 2023 Kalle Fagerberg

SPDX-License-Identifier: CC-BY-4.0
-->

# "Rootless" Personio API client

Accessing [Personio's API](https://developer.personio.de/docs)
requires API credentials [which does not scope to the employee level](https://developer.personio.de/discuss/634e4b08a3f8d80051c52cfe),
meaning you can only get official API access as an admin user,
where you get access to the sensitive information of all the employees in your
company.

Instead, this package uses a different API: the same API as your web browser.

This is done by pretending to be a browser and logging in normally using
email and password.

## License

This repository was created by [@jorie1234](https://github.com/jorie1234),
but has been forked and is now maintained by Kalle Fagerberg
([@jilleJr](https://github.com/jilleJr)).

The code in this project is licensed under GNU General Public License v3.0
or later ([LICENSES/GPL-3.0-or-later.txt](LICENSES/GPL-3.0-or-later.txt)),
and documentation is licensed under Creative Commons Attribution 4.0
International ([LICENSES/CC-BY-4.0.txt](LICENSES/CC-BY-4.0.txt)).