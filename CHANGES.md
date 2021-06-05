v2.3.3 - 2021-06-05
===================

* Move assets and templates out of code, include them using the `embed` module
* Require Go 1.16 (for `embed`)
* Add favicon


v2.3.2 - 2021-04-10
===================

* Change licence to EUPL-1.2
* [gh] Fix action to build release binaries


v2.3.1 - 2021-04-10
===================

* Fix flaky test
* [gh] Fix GitHub actions branch name


v2.3.0 - 2021-04-10
===================

* Support custom CSS file for listing (#10)
* Add OS/architecture info at bottom of listing page
* Update dependencies and go versions


v2.2.2 - 2020-12-19
===================

* Log request source address
* Cleanups and refactoring
* [snap] Use go modules
* [snap] switch to `core20`


v2.2.1 - 2020-02-08
===================

* Don't show directory size in HTML output (#9)
* CSS/HTML cleanups


v2.2.0 - 2020-01-13
===================

* Report symlinks as files/directories based on the type of the target (#8)
* Don't follow symlinks outside of the base directory, add option to allow it


v2.1.1 - 2020-01-11
===================

* Add command options validation for paths, return errors instead of panicking
  on invalid configurations (#7)


v2.1.0 - 2020-01-08
===================

* Suport directory listing sorting by name or size
* Serve CSS and SVG logo as separate assets
* Log handler errors on 500 responses (#6)


v2.0.0 - 2020-01-06
===================

* Support for HTTP Basic Authentication (#1)
* Hide dotfiles by default (togglable via command line option) (#2)
* Rework directory listing code, add nicer HTML+CSS output (#3)
* Support JSON output for directory listing (#4)
* Redirect directory paths without trailing slash to ones with slash (#5)


v1.2.0 - 2019-12-29
===================

* When a path without `.htm(l)` suffix is requrested, if it doesn't exist but a
  file with the suffix exists, serve that file instead


v1.1.0 - 2019-07-01
===================

* Use go modules
* Refactor code, add tests


v1.0.2 - 2019-03-30
===================

* Add logo
* Snap improvements


v1.0.1 - 2019-02-02
===================

* Add service to the snap


v1.0.0 - 2019-01-23
===================

* First release
