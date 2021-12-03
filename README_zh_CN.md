# robot-gitee-openeuler-review

### 概述

该机器人为openEuler社区提供了Code Review 相关功能。提供了`lgtm`、`approved`标签、PR合入指令以及跟踪PR源代码改变自动清除过时的`lgtm`、`approved`标签，并在满足PR合入的条件自动合入PR。

### 功能

- **命令**

  提供如下指令：

  | 命令              | 示例                         | 描述                                                         | 谁能使用                                                     |
  | ----------------- | ---------------------------- | ------------------------------------------------------------ | ------------------------------------------------------------ |
  | /lgtm [cancel]    | /lgtm<br/>/lgtm cancel       | 为一个Pull Request添加或者删除`lgtm`标签，这个标签将用于Pull Request合入判断。 | 这个仓库的协作者。Pull Request作者能使用`/lgtm cancel`命令，但是不能使用`/lgtm`命令。 |
  | /approve [cancel] | /approve<br/>/approve cancel | 为一个Pull Request添加或者删除`approved`标签，这个标签将用于Pull Request合入判断。 | 这个仓库的协作者。                                           |
  | /check-pr         | /check-pr                    | 检测当前PR的标签是否满足条件，如果满足即合入PR。             | 任何人都能在一个Pull Request上触发这种命令。                 |

- **指定lgtm标签个数**

  [配置项](#configuration)提供了PR `lgtm`标签的个数设置，当该配置项大于1时，`lgtm`标签的内容以`lgtm-user`组成。ps： user为使用/lgtm命令的用户在码云平台的login id。

- **自动清理lgtm标签**

  当PR有新的commit提交时我们将会移除已存在的`lgtm`标签。

- **PR合入**

  1. 自动合入：自动检测PR合入的条件，满足合入条件即自动合入。
  2. 手动检查触发合入：使用**/check-pr**指令可以触发机器人检查PR当前的合入条件，不满足合入条件时给与相应提示，否则PR合入。

- **自动添加`/retest`评论**

  当PR有新的commit提交时自动加`/retest`评论以触发测试任务
  
- **检查PR作者是否指定审查者**

  根据配置项当开启检查审查者功能时，PR创建后会检查作者是否指定审查者如果未指定，给予相应提示。
  
### 配置<a id="configuration"/>

例子：

```yaml
#无额外说明配置项为非必须项
config_items:
  - repos:  #robot需管理的仓库列表(必需)
     -  owner/repo
     -  owner1
    excluded_repos: #robot 管理列表中需排除的仓库
     - owner1/repo1
    lgtm_counts_required: 1 #lgtm标签阈值
    labels_for_merge: #PR合入需要的标签
      - ci-pipline-success
    missing_labels_for_merge: #PR合入时不能存在的标签
      - ci-pipline-failed
    # 指定在开发者评论/lgtm 或/approve 命令时根据sig 目录下的owners 文件检查开发者的权限。
    check_permission_based_on_sig_owners: true
    # Sig 的目录。当 CheckPermissionBasedOnSigOwners 为真时必须设置它。
    sigs_dir: sig
     merge_method: merge #PR合入时使用的方式，可选项：merge、squash.默认merge.
     unable_checking_reviewer_for_pr: true #是否检查审核人
```

