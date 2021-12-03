# robot-gitee-openeuler-review
[中文README](README_zh_CN.md)
### Overview

The bot provides Code Review-related functionality for the openEuler community. Provides `lgtm`, `approved` labels, PR merge command and tracking PR source code changes to automatically remove obsolete `lgtm`, `approved` labels and automatically merge PR when PR merge conditions are met.

### Features

- **Command**

  The following  command are provided:

  | command           | example                      | description                                                  | who can use                                                  |
  | ----------------- | ---------------------------- | ------------------------------------------------------------ | ------------------------------------------------------------ |
  | /lgtm [cancel]    | /lgtm<br/>/lgtm cancel       | Add or remove the `lgtm` label for a Pull Request, this label will be used for Pull Request merge determination. | Collaborators of this repository.<br/>Pull Request authors can use the `/lgtm cancel` command, but cannot use the `/lgtm` command. |
  | /approve [cancel] | /approve<br/>/approve cancel | Add or remove the `approved` label for a Pull Request, this label will be used for Pull Request merge determination. | Collaborators of this repository.                            |
  | /check-pr         | /check-pr                    | Check whether the current PR's tag meets the condition, if it does, it is merged into the PR. | Anyone can trigger such a command on a Pull Request.         |

- **Specify the number of lgtm labels**

  The [configuration item](#configuration) provides a setting for the number of PR `lgtm` tags. When this configuration item is greater than 1, the contents of the `lgtm` tags consist of `lgtm-user`. ps：the `user` is the login id of the user using /lgtm command in the gitee platform.

- **Automatic cleaning of lgtm labels**

  We will remove the existing `lgtm` labels when a new commit is submitted for the PR.

- **Merge PR**

  1. Auto-merge: automatically detects the conditions for PR merge, and automatically merges in when the merge conditions are met.
  2. Manual check-trigger merge-in: Use the **/check-pr** command to trigger the robot to check the current merge-in condition of the PR, and give the corresponding prompt when the merge-in condition is not met, otherwise the PR is merged in.

- **Automatically add `/retest` comments**

  When a PR has a new commit, it will automatically add `/retest` comments to trigger the test task

- **Check whether the PR author has designated a reviewer**

  According to the configuration item, when the check reviewer function is turned on, after the PR is created, it will check whether the author has designated a reviewer. If not, it will give corresponding prompts.

### Configuration<a id="configuration"/>

example:

```yaml
#no additional description of the configuration items are not required
config_items:
  - repos:  #list of warehouses to be managed by robot (required)
     -  owner/repo
     -  owner1
    excluded_repos: #robot manages the list of repositories to be excluded
     - owner1/repo1
    lgtm_counts_required: 1 #lgtm label threshold
    labels_for_merge: #labels required for PR merging
      - ci-pipline-success
    missing_labels_for_merge: #labels that cannot exist when PR is merged in
      - ci-pipline-failed
    # specify it should check the devepler's permission besed on the owners file in sig directory when the developer comment /lgtm or /approve command.
    check_permission_based_on_sig_owners: true
    # is the directory of Sig. It must be set when CheckPermissionBasedOnSigOwners is true.
    sigs_dir: sig
    # merge_method is the method to merge PR.The default method of merge. valid options are squash and merge.
    merge_method: merge
    unable_checking_reviewer_for_pr: true #Whether to check the reviewer
```



  

