#! /bin/sh

# _ROOT: 工作目录
# ROOT: 总是正确指向build脚本所在目录
_ROOT="$(pwd)" && cd "$(dirname "$0")" && ROOT="$(pwd)"
PJROOT="$ROOT"
DARWIN="$([ "$(uname -s)" = "Darwin" ] && echo true || echo false)"

__tag_message() {
    cd "$PJROOT"

    if which git 2>/dev/null > /dev/null && git status 2>/dev/null >/dev/null; then
        _upstream="$(git rev-parse --abbrev-ref @{upstream} 2>/dev/null | cut -d/ -f1)"
        [ -z "$GIT_UPSTREAM" ] && _upstream="origin"
        GIT_REPO="$(git config --get remote.$_upstream.url 2>/dev/null)"
        TAG_NAME="$(git describe --tags --long --match v[0-9]* 2>/dev/null | sed -nE 's/(.*)-[0-9]+-g.{7,}/\1/p')"
        if [ -n "$TAG_NAME" ]; then
            TAG_MESSAGE="$(git tag -l v[0-9]* -n100 --sort=-v:refname | sed -n "/^$TAG_NAME /,\$p" | sed -E 's/^(v[0-9][^ \t]+)[ \t]{6}/\1\
/')"
        fi
    fi
}

__tag_message

if [ -n "$TAG_MESSAGE" ]; then
    echo "$TAG_MESSAGE"
else
    echo "[WARN] There is no changelog" >&2
fi
