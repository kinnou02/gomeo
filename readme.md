Gomeo
=====

Basic implementation of the "compagnon" api of timeo dedicated to Navitia.

How to build
============
You must have the latest version of [go](https://golang.org/) installed.
In the root directory of the project do:
```
make setup
```

You can then build the project by doing:
```
make
```

The (non existent) test can be run with the following command:
```
make ci
```

TODO
====
 - parse query with the original format
 - handle error: no departure isn't the same thing than an invalid route point
 - handle multiple database
