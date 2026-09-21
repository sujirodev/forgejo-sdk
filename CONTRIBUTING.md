# Contribution Guidelines

## Bug reports

Please search the issues on the issue tracker with a variety of keywords to ensure your bug is not already reported.

If unique, [open an issue](https://codeberg.org/MatheusAlves96/forgejo-sdk/issues/new) and answer the questions so we can understand and reproduce the problematic behavior.

To show us that the issue you are having is in Forgejo SDK itself, please write clear, concise instructions so we can reproduce the behavior (even if it seems obvious). The more detailed and specific you are, the faster we can fix the issue.

Please be kind, remember that Forgejo SDK comes at no cost to you, and you're getting free help.

## Discuss your design

The project welcomes submissions but please let everyone know what you're working on if you want to change or add something to the Forgejo SDK repository.

Before starting to write something new for the Forgejo SDK project, please [file an issue](https://codeberg.org/MatheusAlves96/forgejo-sdk/issues/new).

## Testing redux

Before sending code out for review, run all the tests using `make test`, to make sure the changes don't break other usage. In order to run the test, you'll need a test instance. You can create one using `make test-instance`.

Every exported `*Client` method counts as a route and needs a test:
`make check-route-coverage` says which ones are still missing one.

## Adding or changing a route

[`ROUTES.md`](ROUTES.md) and the counters in [`README.md`](README.md) are
generated, never edited by hand. CI regenerates them from what the suite
actually did against each Forgejo version, so a route added without a test
that exercises it fails the `route-matrix` job rather than quietly landing
as a supported route.

Two consequences for your pull request:

* If `ROUTES.md` conflicts on merge, do not resolve the conflict line by
  line. Take either side and run `make route-matrix` again.
* If your route genuinely cannot be exercised on some version, declare it in
  [`route-exceptions.json`](route-exceptions.json) with one of the accepted
  reasons and a link to the issue that tracks it. `todo` expires after 90
  days.

## Code review

Changes must be reviewed before they are accepted, no matter who makes the change even if it is an owner or a maintainer.

Please try to make your pull request easy to review for us. Some of the key points:

* Make small pull requests. The smaller, the faster to review and the more likely it will be merged soon.
* Don't make changes unrelated to your PR. Maybe there are typos on some comments, maybe refactoring would be welcome on a function... but if that is not related to your PR, please make *another* PR for that.
* Split big pull requests into multiple small ones. An incremental change will be faster to review than a huge PR.

## Sign your work

The sign-off is a simple line at the end of the explanation for the patch. Your signature certifies that you wrote the patch or otherwise have the right to pass it on as an open-source patch. The rules are pretty simple: If you can certify [DCO](DCO), then you just add a line to every git commit message:

```
Signed-off-by: Joe Smith <joe.smith@email.com>
```

Please use your real name. If you set your `user.name` and `user.email` git configs, you can sign your commit automatically with `git commit -s`.

## Copyright

Code that you contribute should use the standard copyright header:

```
// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
```

Files in the repository contain copyright from the year they are added to the year they are last changed. If the copyright author is changed, just paste the header below the old one.
