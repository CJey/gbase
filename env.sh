#! /bin/sh

# _ROOT: 工作目录
# ROOT: 总是正确指向build脚本所在目录
_ROOT="$(pwd)" && cd "$(dirname "$0")" && ROOT="$(pwd)"
PJROOT="$ROOT"

# 检查golang环境
__check_golang() {
    GO_DEFAULT=/usr/local/go/bin/go
    GO=go

    if ! which $GO >/dev/null ; then
        if [ -x $GO_DEFAULT ]; then
            GO=$GO_DEFAULT
        else
            echo "[Error] go environment not found" >&2
            exit 1
        fi
    fi

    if $GO mod 2>&1 | grep -q -i 'unknown command'; then
        echo "[Error] low golang version(should be >=1.11), do not support go mod command"
        exit 1
    fi

    if [ ! -r $PJROOT/go.mod ]; then
        echo "[Error] go.mod not found or not readable"
        exit 1
    fi

    MODULE="$(cat $PJROOT/go.mod | grep ^module | head -n1 | awk '{print $NF}')"
}

__check_git() {
    cd "$PJROOT"

    if which git 2>/dev/null > /dev/null && git status 2>/dev/null >/dev/null; then
        _upstream="$(git rev-parse --abbrev-ref @{upstream} 2>/dev/null | cut -d/ -f1)"
        [ -z "$GIT_UPSTREAM" ] && _upstream="origin"
        GIT_REPO="$(git config --get remote.$_upstream.url 2>/dev/null)"
        GIT_BRANCH="$(git rev-parse --abbrev-ref HEAD 2>/dev/null)"

        GIT_HASH="$(git log -n1 --pretty=format:%H 2>/dev/null)"
        TAG_NAME="$(git describe --tags --long --match v[0-9]* 2>/dev/null | sed -nE 's/(.*)-[0-9]+-g.{7,}/\1/p')"
        if [ -n "$TAG_NAME" ]; then
            TAG_HASH="$(git rev-list -n1 "$TAG_NAME")"
        fi
    fi
}

# 搜集待注入的编译环境信息
__version_info() {
    cd "$PJROOT"

    VERSION="0.0.1"
    RELEASE="1"

    if [ -n "$GIT_REPO" ]; then
        GIT_TIME="$(git log -n1 --pretty=format:%at 2>/dev/null)"
        GIT_NUMBER="$(git rev-list --count HEAD 2>/dev/null)"

        if [ -n "$TAG_NAME" ]; then
            TAG_TIME="$(git log -n1 $TAG_HASH --pretty=format:%at)"
            TAG_NUMBER="$(git rev-list --count $TAG_HASH)"

            TAG_DIFF="$(git rev-list --count HEAD ^$TAG_HASH)"
            TAG_MESSAGE="$(git tag -l v[0-9]* -n100 --sort=-v:refname | sed -n "/^$TAG_NAME /,\$p" | sed -E 's/^(v[0-9][^ \t]+)[ \t]{6}/\1\
/')"

            VERSION="$(echo $TAG_NAME | cut -c2-)"
            RELEASE="$((1+TAG_DIFF))"
        fi
    fi
}

__check_golang
__check_git
__version_info

echo "Module      $MODULE"
echo "Version     $VERSION-$RELEASE"
echo
echo "GitTrace    $GIT_NUMBER.$(echo $GIT_HASH | cut -c1-7)"
echo "GitBranch   $GIT_BRANCH"
echo "GitRepo     $GIT_REPO"
echo "GitHash     $GIT_HASH @ $(date --date=@$GIT_TIME '+%Y-%m-%d %H:%M:%S %Z')"

if [ -n "$TAG_NAME" -a "$GIT_HASH" != "$TAG_HASH" ]; then
echo
echo "TagTrace    $TAG_NUMBER.$(echo $TAG_HASH | cut -c1-7)"
echo "TagName     $TAG_NAME"
echo "TagDiff     $TAG_DIFF"
echo "TagHash     $TAG_HASH @ $(date --date=@$TAG_TIME '+%Y-%m-%d %H:%M:%S %Z')"
fi

if [ -n "$TAG_NAME" ]; then
MAXSHOW=10
echo
echo "### BEGIN CHANGELOG"
echo "$TAG_MESSAGE" | head -n$MAXSHOW | sed 's/^/# /'
if [ $(echo "$TAG_MESSAGE" | wc -l) -gt $MAXSHOW ]; then
echo "# "
echo "# + You can execute changelog.sh to see more changelog!"
fi
echo "### END CHANGELOG"
fi
