# Changelog

## [v3.2.6](https://github.com/sujirodev/forgejo-sdk/releases/tag/forgejo/v3.2.6) - 2026-09-29

* GENERAL
  * SDK: exercitar ActivityPubFollow de verdade contra o par federado (#57) ([#78](https://codeberg.org/sujirodev/forgejo-sdk/pulls/78))
  * ci(release): English release notes in upstream's format ([#77](https://codeberg.org/sujirodev/forgejo-sdk/pulls/77))
  * chore(deps): corrigir os lookups do Renovate e resolver as duas deps abandonadas (#14) ([#72](https://codeberg.org/sujirodev/forgejo-sdk/pulls/72))
  * ci: check the Codeberg token before pushing ([#3](https://github.com/sujirodev/forgejo-sdk/pull/3))
  * ci: migrate the repository and its pipeline to GitHub ([#1](https://github.com/sujirodev/forgejo-sdk/pull/1))
* FIXES
  * fix: never return an empty error from statusCodeToErr ([#75](https://codeberg.org/sujirodev/forgejo-sdk/pulls/75))
  * fix: make the release train work across the move to GitHub ([#4](https://github.com/sujirodev/forgejo-sdk/pull/4))
  * fix: strip whitespace from the Codeberg token in mirror.yml ([#2](https://github.com/sujirodev/forgejo-sdk/pull/2))


## [v3.2.5](https://codeberg.org/MatheusAlves96/forgejo-sdk/releases/tag/forgejo/v3.2.5) - 2026-09-22

* MISC
  * SDK: cobrir rotas de organização — runners, quota, avatar e block/unblock (#47)
  * SDK: implementar attachments e dependências de Issue (21 rotas, Forgejo 16.0.5) (#48)
  * chore(deps): update ghcr.io/renovatebot/renovate docker tag to v44.106.0 (#61)
  * ci: release train (integration.yml, compute-release-version.sh, release.yml, docs) (#64)
  * ci(release): publicar o que um release de biblioteca Go deveria carregar (#66)
  * test: close the coverage gap PR #49 reopened (fase 8) (#67)
  * test: run the integration suite in parallel (291s -> ~105s) (#68)


## [v3.2.4](https://codeberg.org/MatheusAlves96/forgejo-sdk/releases/tag/forgejo/v3.2.4) - 2026-09-22

* MISC
  * routes: implementar 48 das 62 rotas de repositório faltando (Forgejo 16.0.5) (#49)


## [v3.2.3](https://codeberg.org/MatheusAlves96/forgejo-sdk/releases/tag/forgejo/v3.2.3) - 2026-09-22

* MISC
  * SDK: cobrir as 37 rotas de admin faltando (Forgejo 16.0.5) (#46)


## [v3.2.2](https://codeberg.org/MatheusAlves96/forgejo-sdk/releases/tag/forgejo/v3.2.2) - 2026-09-22

* MISC
  * SDK: implementar as 11 rotas de ActivityPub (Forgejo 16.0.5) (#45)


## [v3.2.1](https://codeberg.org/MatheusAlves96/forgejo-sdk/releases/tag/forgejo/v3.2.1) - 2026-09-22

* MISC
  * SDK: cobrir rotas de usuário — avatar, bloqueio, GPG e runners pessoais (#44)


## [v3.2.0](https://codeberg.org/MatheusAlves96/forgejo-sdk/releases/tag/forgejo/v3.2.0) - 2026-09-22

* MISC
  * feat(misc): add 13 miscellaneous instance-level routes (#43)


## [v3.1.3](https://codeberg.org/MatheusAlves96/forgejo-sdk/releases/tag/forgejo/v3.1.3) - 2026-09-21

* MISC
  * test: close the 12 routes the evidence did not support, and turn the gate on (#30)


## [v3.1.2](https://codeberg.org/MatheusAlves96/forgejo-sdk/releases/tag/forgejo/v3.1.2) - 2026-09-21

* MISC
  * docs: generate the per-version route contract from what the suite actually did (#29)


## [v3.1.1](https://codeberg.org/MatheusAlves96/forgejo-sdk/releases/tag/forgejo/v3.1.1) - 2026-09-21

* MISC
  * test: a test for every route, and a check that keeps it that way (#28)


## [v3.1.0](https://codeberg.org/MatheusAlves96/forgejo-sdk/releases/tag/forgejo/v3.1.0) - 2026-09-21

* MISC
  * feat(packages): add LinkPackage and UnlinkPackage (#42)


## [v3.0.1](https://codeberg.org/MatheusAlves96/forgejo-sdk/releases/tag/forgejo/v3.0.1) - 2026-09-19

* MISC
  * ci: run the integration suite against Forgejo's LTS lines too (#21)


## [v3.0.0](https://codeberg.org/mvdkleijn/forgejo-sdk/releases/tag/forgejo/v3.0.0) - 2026-03-04

* GENERAL
  * chore: bump minimum go version to 1.25 (#112)
  * chore: update renovate and integration workflows

* FEATURES
  * feat: Initial quota support (#111)
  * feat(actions): Add comprehensive Actions API client (#103) (thanks @redbeard)
  * add repo tag protection (#106) (thanks @kfkonrad)
  * allow access token operations on arbitrary users (#104) (thanks @kfkonrad)
  * feat: add fast-forward-only merge style (#101) (thanks @kernald)
  * Add HTTP accept header into request (#100) (thanks @MartinBasti)
  * feat: Support external_tracker_regexp_pattern in ExternalTracker (#97) (thanks @qaiser42)
  * feat: improve action secret validation and test infrastructure (#91) (thanks @pavel_hushcha)
  * feat: add UnitsMap support for per-unit team permissions (#83) (thanks @kfkonrad)
  * feat: add endpoint GetCommitPullRequest (#79) (thanks @apricote)

* FIXES
  * fix: add missing fields to EditRepoOption (#114)
  * fix: removed unneeded ListReleaseAttachementsOption (#113)
  * fix: add missing RunNumber and IsRefDeleted fields to ActionRun (#110) (thanks @redbeard)
  * fix: mandatory units in EditTeamOption (#108)
  * fix: change Body in EditPullRequestOption to *string (#99) (thanks @OFHansen)
  * test: fix race condition in pull request merge test (#90) (thanks @pavel_hushcha)


## [v2.2.0](https://codeberg.org/mvdkleijn/forgejo-sdk/releases/tag/forgejo/v2.2.0) - 2025-07-17

* GENERAL

  * Support API endpoint repoGetPullRequestByBaseHead (#73) (thanks @wandhydrant)

* DEPENDENCIES

  * fix(deps): update module golang.org/x/crypto to v0.39.0 (#54)


## [v2.1.0](https://codeberg.org/mvdkleijn/forgejo-sdk/releases/tag/forgejo/v2.1.0) - 2025-05-12

* add User.HTMLURL field (#57) thanks @infinoid
* Bumped some dependencies
* Improved CI and Renovate setup
* Fixed some linting issues


## [v2.0.0](https://codeberg.org/mvdkleijn/forgejo-sdk/releases/tag/forgejo/v2.0.0) - 2025-02-10

This release of the SDK is mostly the same as previous ones, but introduces a
couple of breaking changes, hence the major version upgrade.

* BREAKING

  * feat!: Add GetTreesOptions argument (#21)
    This feature introduces pagination and a new API for the GetTrees function
    which brings it in line with other paginated SDK functions.

  * chore!: update AccessTokenScope* constants for 9.0.3 (#26)
    The AccessTokenScope* constants in the SDK were outdated and brought into
    line with current use within Forgejo. Tested against Forgejo v9.0.3.

* MAINTENANCE

  * chore: Bump SDK to go module v2 (#33)
  * chore: update readme for Forgejo v9.0.3 support (#32)
  * chore: update readme with new badges (#25)

* CI CHANGES

  * ci: move-releasedrafter-config (#31)
  * ci: attempt to fix release drafter workflow (#30)
  * ci: correct release-drafter url (#29)
  * fix(ci): release drafter config (#28)
  * ci: try out release drafter on forgejo actions (#27)
  * ci: allow manual trigger of workflow (#24)
  * chore(ci): bump forgejo version to 9.0.3 for integration workflow (#23)
  * chore(ci): add actions based integration workflow (#22)


## [v1.2.0](https://codeberg.org/mvdkleijn/forgejo-sdk/releases/tag/forgejo/v1.2.0) - 2024-10-21

* GENERAL

  * chore: bump minimal go version to 1.22 as well as deps
  * chore: bump gofumpt and golangci_lint
  * fix: correct package name

* TESTING CHANGES

  * test: bump integration workflow to 8.0.3
  * chore(test): add version test for 8.0.3
  * fix(test): ensure stat and verification are returned
  * fix(test): make sure to check for nil
  * test: bump forgejo release to 8.0.3
  * test: bump forgejo version to use in testing to 8.0.3
  * chore(test): rename unused param to underscore

* CHERRY PICKED from upstream

  * Update httpsig dependency (#667)
  * (feat) support more query params for /repos/{owner}/{repo}/commits (#668)
  * Allow team names of up to 255 characters (#670)
  * feat: implement Gitea Repo Action Secrets Management (#662)


## [v1.1.1](https://codeberg.org/mvdkleijn/forgejo-sdk/releases/tag/forgejo/v1.1.1) - 2024-06-20

This is a security patch release that prevents the Forgejo token from being logged
when logging an "unknown api" error.

Upgrading is advised to prevent accidentally leaking the Forgejo token.


## [v1.1.0](https://codeberg.org/mvdkleijn/forgejo-sdk/releases/tag/forgejo/v1.1.0) - 2024-06-16

* Updates CI testing to use Forgejo 7.0.4;
* Uses Codeberg's Woodpecker CI for testing;
* Cherry picked the latest changes to the Gitea SDK;
* Bumps dependencies;
* Fixes silly typo on copyright header;

A couple of points to take into account:

* This is a HARD fork;
* This SDK intends to follow Semver 2;
* The last pre-fork commit was January 29th 2024 (see: 5d0143e4e7);


## [v0.15.1](https://gitea.com/gitea/go-sdk/releases/tag/gitea/v0.15.1) - 2022-01-04

* FEATURES
  * Add ignoreVersion & manuall version set option (#560) (#562)
* BUGFIXES
  * Fix version string for next release (#559)


## [v0.15.0](https://gitea.com/gitea/go-sdk/releases/tag/gitea/v0.15.0) - 2021-08-13

* BREAKING
  * Introduce NotifySubjectState (#520)
  * Drop deprecations (#503)
* FEATURES
  * Add Repo Team Management Functions (#537)
  * Add CreateRepoFromTemplate (#536)
  * Add GetReviewers & GetAssignees (#534)
  * Add GetTag, GetAnnotatedTag & CreateTag (#533)
  * Add GetUserSettings & UpdateUserSettings (#531)
  * Add ListPullRequestCommits (#530)
  * Add GetUserByID (#513)
  * Add GetRepoByID (#511)
* ENHANCEMENTS
  * Update List Options (#527)
  * Update Structs (#524)
  * ListFunctions: option to disable pagination (#509)


## [v0.14.1](https://gitea.com/gitea/go-sdk/releases/tag/gitea/v0.14.1) - 2021-06-30

* BUGFIXES
  * Fix setDefaults (#508) (#510)


## [v0.14.0](https://gitea.com/gitea/go-sdk/releases/tag/gitea/v0.14.0) - 2021-03-21

* BREAKING
  * Update Structs (#486)
  * Added repo ListContents and changed GetContents doc to talk about a single file (#485)
  * Remove & Rename  TrackedTimes list functions (#467)
  * UrlEscape Function Arguments used in UrlPath (#273)
* FEATURES
  * Add Create/Delete ReviewRequests (#493)
  * Add Un-/DismissPullReview funcs (#489)
  * Add Repo Un-Star Functions (#483)
  * introduce Client.GetArchiveReader (#476)
  * Add DeleteRepoTag function (#461)
  * Add GetReleaseByTag (#427)
* BUGFIXES
  * Handle Contents Edge-Case (#492)
  * Fix GetCombinedStatus() (#470)
  * Use Predefind Versions & Compare Function (#442)
  * Return resp on NotFound too (#428)
* ENHANCEMENTS
  * Add workaround to get head branch sha of pulls with deleted head branch (#498)
  * GetFile: Use "ref" in-query if posible (#491)
  * Add DeleteTag & Correct DeleteReleaseByTag (#488)
  * Add html_url field to Release struct (#477)
  * Add Ref to Issue structs (#466)
  * Update Issue Struct (#458)
  * Use sync.Once for loading ServerVersion (#456)
  * Add Gitea2Gitea Migration Support (#454)
  * Add Helper for Optional Values (#448)
  * Update CreateRepoOption struct (#445)
  * Update CommitMeta Struct (#434)
  * Update Struct NotificationSubject (#424)
  * Add Debug Mode (#422)
* DOCS
  * Make Client thread-safe & add docs (#495)
  * Improve PullReview docs (#469)


## [v0.13.3](https://gitea.com/gitea/go-sdk/releases/tag/gitea/v0.13.3) - 2021-03-22

* BUGFIXES
  * Fix GetCombinedStatus() (#470) (#472)
* ENHANCEMENTS
  * Add html_url field to Release struct (#477) (#478)


## [v0.13.2](https://gitea.com/gitea/go-sdk/releases/tag/gitea/v0.13.2) - 2020-12-07

* BUGFIXES
  * Use Predefind Versions & Compare Function (#442) (#446)
* ENHANCEMENTS
  * Add Gitea2Gitea Migration Support (#454) (#455)
  * Update CreateRepoOption struct (#445) (#447)
  * Update CommitMeta Struct (#434) (#437)


## [v0.13.1](https://gitea.com/gitea/go-sdk/releases/tag/gitea/v0.13.1) - 2020-09-29

* FEATURES
  * Add GetReleaseByTag (#427) (#430)
* BUGFIXES
  * Return http Response on NotFound too (#428) (#429)
* ENHANCEMENTS
  * Update Struct NotificationSubject (#424) (#425)
  * Add Debug Mode (#422) (#423)


## [v0.13.0](https://gitea.com/gitea/go-sdk/releases/tag/gitea/v0.13.0) - 2020-09-15

* BREAKING
  * Check Gitea Version Requirement (#419)
  * All Function return http responce (#416)
  * Remove opts from ListPullReviewComments (#411)
  * Use enum AccessMode for OrgTeam and Collaborator functions (#408)
  * CreateOrgOption rename UserName to Name (#386)
  * EditMilestoneOption also use StateType (#350)
  * Refactor RepoSearch to be easy usable (#346)
* FEATURES
  * Milestone Functions accept name to identify (#418)
  * Make http requests with context (#417)
  * Add GetGlobalAttachmentSettings (#414)
  * Add GetArchive (#413)
  * Add GetRepoLanguages + TESTs (#412)
  * Add CreateBranch (#407)
  * Add Admin CronTask functions (#406)
  * Add GetGlobalAPISettings Function (#404)
  * Add Get Diff and Patch endpoints for pull requests (#398)
  * Add Validate func for Create/Edit Options (#370)
  * Add Function to get GetGlobalSettings and GetSettingAllowedReactions (#359)
* ENHANCEMENTS
  * TrackedTime API >= 1.11.x needed (#415)
  * Update Milestone struct (#410)
  * Add Fallback for GetPullRequestDiff/Patch (#399)
  * DeleteToken Accept Names too (#394)
  * Update ListMilestoneOption struct (#393)
  * Migration Api Changed (#392)
  * Refactor Visibletype Orgs (#382)
  * Extend Notification Functions (#381)
  * Update GetGlobalSettings Functions (#376)
  * Allow Creating Closed Milestones (#373)
  * CreateLabel correct Color if needed for old versions (#365)
  * Issue/Pull add IsLocked Property (#357)
  * Update EditPullRequestOption Add Base (#353)
  * File Create/Update/Delete detect DefaultBranch if Branch not set for old Versions (#352)
  * Improve Error Handling (#351)

## [v0.12.2](https://gitea.com/gitea/go-sdk/releases/tag/gitea/v0.12.2) - 2020-09-05

* ENHANCEMENTS
  * Extend Notification Functions (#381) (#385)

## [v0.12.1](https://gitea.com/gitea/go-sdk/pulls?q=&type=all&state=closed&milestone=1268) - 2020-07-09

* ENHANCEMENTS
  * Improve Error Handling (#351) (#377)
  * Allow Creating Closed Milestones (#373) (#375)
  * File Create/Update/Delete detect DefaultBranch if Branch not set for old Versions (#352) (#372)
  * CreateLabel correct Color if needed for old versions (#365) (#371)
  * Update EditPullRequestOption Add Base (#353) (#363)

## [v0.12.0](https://gitea.com/gitea/go-sdk/pulls?q=&type=all&state=closed&milestone=1223) - 2020-05-21

* BREAKING
  * Support 2FA for basic auth & refactor Token functions (#335)
  * PullMerge: use enum for MergeStyle (#328)
  * Refactor List/SetRepoTopics (#276)
  * Remove ListUserIssues() ... (#262)
  * Extend SearchUsers (#248)
  * Fix & Refactor UserApp Functions (#247)
  * Add ListMilestoneOption to ListRepoMilestones (#244)
  * Add ListIssueCommentOptions for optional param (#243)
  * Refactor RepoWatch (#241)
  * Add Pagination Options for List Requests (#205)
* FEATURES
  * Add BranchProtection functions (#341)
  * Add PullReview functions (#338)
  * Add Issue Subscription Check & Fix DeleteIssueSubscription (#318)
  * Add Branch Deletion (#317)
  * Add Get/Update for oauth2 apps (#311)
  * Add Create/Get/Delete for oauth2 apps (#305)
  * Add DeleteFile() (#302)
  * Add Get/Update/Create File (#281)
  * Add List/Check/SetPublic/Delete OrgMembership functions (#275)
  * Add ListRepoCommits (#266)
  * Add TransferRepo (#264)
  * Add SearchRepo API Call (#254)
  * Add ListOptions struct (#249)
  * Add Notification functions (#226)
  * Add GetIssueComment (#216)
* BUGFIXES
  * Add missing JSON header to AddCollaborator() (#306)
  * On Internal Server Error, show request witch caused this (#296)
  * Fix MergePullRequest & extend Tests (#278)
  * Fix AddEmail (#260)
* ENHANCEMENTS
  * Check if gitea is able to squash-merge via API (#336)
  * ListIssues: add milestones filter (#327)
  * Update CreateRepoOption struct (#300)
  * Add IssueType as filter for ListIssues (#286)
  * Extend ListDeployKeys (#268)
  * Use RepositoryMeta struct on Issues (#267)
  * Use StateType (#265)
  * Extend Issue Struct (#258)
  * IssueSubscribtion: Check http Status responce (#242)

## [v0.11.3](https://gitea.com/gitea/go-sdk/pulls?q=&type=all&state=closed&milestone=1259) - 2020-04-27
* BUGFIXES
  * Fix MergePullRequest (#278) (#316)
  * Add missing JSON header to AddCollaborator() (#307)

## [v0.11.2](https://gitea.com/gitea/go-sdk/pulls?q=&type=all&state=closed&milestone=1256) - 2020-03-31
* ENHANCEMENTS
  * On Internal Server Error, show request witch caused this (#297)

## [v0.11.1](https://gitea.com/gitea/go-sdk/pulls?q=&type=all&state=closed&milestone=1235) - 2020-03-29
* BUGFIXES
  * Fix SetRepoTopics (#276) (#274)
  * Fix AddEmail (#260) (#261)
  * Fix UserApp Functions (#247) (#256)
* ENHANCEMENTS
  * Add IssueType as filter for ListIssues (#288)
  * Correct version (#259)

## [v0.11.0](https://gitea.com/gitea/go-sdk/pulls?q=&type=all&state=closed&milestone=1222) - 2020-01-27
* FEATURES
  * Add VersionCheck (#215)
  * Add Issue Un-/Subscription function (#214)
  * Add Reaction struct and functions (#213)
  * Add GetBlob (#212)
* BUGFIXES
  * Fix ListIssue Functions (#225)
  * Fix ListRepoPullRequests (#219)
* ENHANCEMENTS
  * Add some pull list options (#217)
  * Extend StopWatch struct & functions (#211)
* TESTING
  * Add Test Framework (#227)
* BUILD
  * Use golangci-lint and revive for linting (#220)
